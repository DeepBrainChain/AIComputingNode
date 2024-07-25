package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/conngater"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/p2p"
	ps "AIComputingNode/pkg/pubsub"
	"AIComputingNode/pkg/serve"
	"AIComputingNode/pkg/timer"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"

	dht "github.com/libp2p/go-libp2p-kad-dht"
)

var version string

const ProtocolVersion string = "aicn/0.0.1"

type connectionResult struct {
	IsPublic bool // Is it a public network node?
	Success  bool // Is the connection successful?
}

func main() {
	configPath := flag.String("config", "", "run using the configuration file")
	versionFlag := flag.Bool("version", false, "show version number and exit")
	initFlag := flag.String("init", "", "initialize configuration in client/server mode")
	peerKeyPath := flag.String("peerkey", "", "parse or generate a key file based on the specified file path")
	pskFlag := flag.Bool("psk", false, "generate a random Pre-Shared Key")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *initFlag == "client" || *initFlag == "server" {
		config.Init(*initFlag)
		os.Exit(0)
	} else if *initFlag != "" {
		fmt.Println("only supports client or server mode")
		os.Exit(0)
	}

	if *peerKeyPath != "" {
		config.PeerKeyParse(*peerKeyPath)
		os.Exit(0)
	}

	if *pskFlag {
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			fmt.Println("Generate Pre-Shared key:", err)
		} else {
			fmt.Println("Generate Pre-Shared key:", hex.EncodeToString(key))
		}
		os.Exit(0)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Println("Failed to load JSON configuration file:", err)
		os.Exit(1)
	}
	err = cfg.Validate()
	if err != nil {
		fmt.Println("Invalid configuration item:", err)
		os.Exit(1)
	}

	err = log.InitLogging(cfg.App.LogLevel, cfg.App.LogFile, cfg.App.LogOutput)
	if err != nil {
		fmt.Println("Initialize the log module failed:", err)
		os.Exit(1)
	}

	log.Logger.Info("################################################################")
	log.Logger.Info("#                          START                               #")
	log.Logger.Info("################################################################")

	err = db.InitDb(db.InitOptions{
		Folder: cfg.App.Datastore,
		// the node is deployed on a public server && enable peers collect
		EnablePeersCollect: (cfg.Swarm.RelayService.Enabled && cfg.App.PeersCollect.Enabled),
	})
	if err != nil {
		log.Logger.Fatalf("Init database: %v", err)
	}

	PeersHistory, err := p2p.ConvertPeersFromStringMap(db.LoadPeerConnHistory())
	if err != nil {
		log.Logger.Fatalf("Load peer history: %v", err)
	}

	DefaultBootstrapPeers, err := p2p.ConvertPeersFromStringArray(cfg.Bootstrap)
	if err != nil {
		log.Logger.Fatalf("Parse bootstrap: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var kadDHT *dht.IpfsDHT

	privKey, _ := p2p.PrivKeyFromString(cfg.Identity.PrivKey)
	connGater := &conngater.ConnectionGater{}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(cfg.Addresses...),
		libp2p.Identity(privKey),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.ProtocolVersion(ProtocolVersion),
		libp2p.UserAgent(version),
		libp2p.ConnectionGater(connGater),
	}
	if cfg.Swarm.ConnMgr.Type == "basic" {
		gracePeriod, _ := time.ParseDuration(cfg.Swarm.ConnMgr.GracePeriod)
		connmgr, err := connmgr.NewConnManager(
			cfg.Swarm.ConnMgr.LowWater,
			cfg.Swarm.ConnMgr.HighWater,
			connmgr.WithGracePeriod(gracePeriod),
		)
		if err != nil {
			log.Logger.Fatalf("Create connection manager: %v", err)
		}
		opts = append(opts, libp2p.ConnectionManager(connmgr))
	}
	if cfg.App.PreSharedKey != "" {
		psk, err := hex.DecodeString(cfg.App.PreSharedKey)
		if err != nil {
			log.Logger.Fatalf("Decoding PSK: %v", err)
		}
		opts = append(opts, libp2p.PrivateNetwork(psk), libp2p.DefaultPrivateTransports)
	} else {
		opts = append(opts, libp2p.DefaultTransports)
	}
	if cfg.Routing.Type != "none" {
		opts = append(opts, libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			dhtOpts := []dht.Option{}
			if cfg.Routing.Type == "auto" {
				dhtOpts = append(dhtOpts, dht.Mode(dht.ModeAuto))
			} else if cfg.Routing.Type == "dhtclient" {
				dhtOpts = append(dhtOpts, dht.Mode(dht.ModeClient))
			} else if cfg.Routing.Type == "dhtserver" {
				dhtOpts = append(dhtOpts, dht.Mode(dht.ModeServer))
			}
			if cfg.Routing.ProtocolPrefix != "" {
				dhtOpts = append(dhtOpts, dht.ProtocolPrefix(protocol.ID(cfg.Routing.ProtocolPrefix)))
			}
			kadDHT, err = dht.New(ctx, h, dhtOpts...)
			return kadDHT, err
		}))
	}
	if cfg.Swarm.RelayClient.Enabled {
		opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(DefaultBootstrapPeers))
	}
	if cfg.Swarm.RelayService.Enabled {
		opts = append(opts, libp2p.EnableRelayService())
	}
	if !cfg.Swarm.DisableNatPortMap {
		opts = append(opts, libp2p.NATPortMap())
	}
	if cfg.Swarm.EnableAutoNATService {
		opts = append(opts, libp2p.EnableNATService())
	}
	if cfg.Swarm.EnableHolePunching {
		opts = append(opts, libp2p.EnableHolePunching())
	}
	host, err := libp2p.New(opts...)
	if err != nil {
		log.Logger.Fatalf("Create libp2p host: %v", err)
	}
	log.Logger.Info("Listen addresses:", host.Addrs())
	log.Logger.Info("Node id:", host.ID())

	// print the node's PeerInfo in multiaddr format
	peerInfo := peer.AddrInfo{
		ID:    host.ID(),
		Addrs: host.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	log.Logger.Info("libp2p node address:", addrs) // addrs[0]

	host.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, c network.Conn) {
			log.Logger.Infof("OnConnected remote multi-addr %v %v", c.RemoteMultiaddr(), c.RemotePeer())
			db.PeerConnected(c.RemotePeer().String(), c.RemoteMultiaddr().String())
		},
		DisconnectedF: func(n network.Network, c network.Conn) {
			log.Logger.Infof("OnDisconnected remote multi-addr %v %v", c.RemoteMultiaddr(), c.RemotePeer())
			// db.PeerDisconnected(c.RemotePeer().String(), c.RemoteMultiaddr().String())
		},
	})

	host.SetStreamHandler(p2p.ChatProxyProtocol, p2p.ChatProxyStreamHandler)

	if cfg.Routing.Type == "none" {
		dhtOpts := []dht.Option{
			dht.Mode(dht.ModeClient),
		}
		if cfg.Routing.ProtocolPrefix != "" {
			dhtOpts = append(dhtOpts, dht.ProtocolPrefix(protocol.ID(cfg.Routing.ProtocolPrefix)))
		}
		kadDHT, err = dht.New(ctx, host, dhtOpts...)
		if err != nil {
			log.Logger.Fatalf("Create Kademlia DHT: %v", err)
		}
	}

	psOpts := []pubsub.Option{
		// pubsub.WithDiscovery(routingDiscovery),
		pubsub.WithEventTracer(&ps.Tracer{}),
		// pubsub.WithRawTracer(&ps.RawTracer{}),
	}
	var gs *pubsub.PubSub

	if cfg.Pubsub.Router == "gossipsub" {
		psOpts = append(psOpts,
			pubsub.WithDirectPeers(DefaultBootstrapPeers),
			pubsub.WithDirectConnectTicks(30),
		)
		if cfg.Routing.Type == "dhtserver" || cfg.Swarm.RelayService.Enabled {
			psOpts = append(psOpts, pubsub.WithPeerExchange(true))
		}
		if cfg.Pubsub.FloodPublish {
			psOpts = append(psOpts, pubsub.WithFloodPublish(true))
		}
		gs, err = pubsub.NewGossipSub(ctx, host, psOpts...)
	} else {
		gs, err = pubsub.NewFloodSub(ctx, host, psOpts...)
	}
	if err != nil {
		log.Logger.Fatalf("New %s: %v", cfg.Pubsub.Router, err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	var connectionResults sync.Map
	for _, peerinfo := range PeersHistory {
		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()
			ispub := p2p.IsPublicNode(pi)
			result := connectionResult{
				IsPublic: ispub,
			}
			if ispub {
				host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
			}
			if err := host.Connect(ctx, pi); err != nil {
				log.Logger.Warnf("Connect history node %v : %v", pi, err)
				result.Success = false
			} else {
				log.Logger.Info("Connection established with history node:", pi)
				result.Success = true
			}
			connectionResults.Store(pi.ID, result)
		}(peerinfo)
	}
	wg.Wait()

	publishConns := 0
	connectionResults.Range(func(key, value any) bool {
		pi := key.(peer.ID)
		result := value.(connectionResult)
		if !result.Success {
			db.PeerConnectFailed(pi.String())
		}
		if result.IsPublic && result.Success {
			publishConns++
		}
		return true
	})

	if publishConns < 4 {
		for _, peerinfo := range DefaultBootstrapPeers {
			wg.Add(1)
			go func(pi peer.AddrInfo) {
				defer wg.Done()
				if host.Network().Connectedness(pi.ID) != network.Connected {
					host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
					if err := host.Connect(ctx, pi); err != nil {
						log.Logger.Warnf("Connect bootstrap node %v : %v", pi, err)
					} else {
						log.Logger.Info("Connection established with bootstrap node:", pi)
					}
				}
			}(peerinfo)
		}
		wg.Wait()
	}

	err = kadDHT.Bootstrap(ctx)
	if err != nil {
		log.Logger.Fatalf("Bootstrap the host: %v", err)
	}

	routingDiscovery := drouting.NewRoutingDiscovery(kadDHT)
	dutil.Advertise(ctx, routingDiscovery, cfg.App.TopicName)

	topic, err := gs.Join(cfg.App.TopicName)
	if err != nil {
		log.Logger.Fatalf("Join PubSub: %v", err)
	}
	defer topic.Close()
	sub, err := topic.Subscribe()
	if err != nil {
		log.Logger.Fatalf("Subscribe PubSub: %v", err)
	}

	p2p.Hio = &p2p.HostInfo{
		Host:            host,
		UserAgent:       version,
		ProtocolVersion: ProtocolVersion,
		PrivKey:         privKey,
		Ctx:             ctx,
		Dht:             kadDHT,
		RD:              routingDiscovery,
		Topic:           topic,
	}

	// Topic publish channel
	publishChan := make(chan []byte, 1024)
	heartbeatInterval, _ := time.ParseDuration(cfg.App.PeersCollect.HeartbeatInterval)
	go ps.PublishToTopic(ctx, topic, publishChan)
	timer.StartAITimer(heartbeatInterval, publishChan)
	go ps.PubsubHandler(ctx, sub, publishChan)
	serve.NewHttpServe(publishChan, *configPath)

	log.Logger.Info("listening for connections")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	// select {} // hang forever
	<-stop
	timer.StopAITimer()
	serve.StopHttpService()
	kadDHT.Close()
	host.Close()

	log.Logger.Info("################################################################")
	log.Logger.Info("#                          OVER                                #")
	log.Logger.Info("################################################################")
}
