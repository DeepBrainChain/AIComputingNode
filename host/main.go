package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/p2p"
	ps "AIComputingNode/pkg/pubsub"
	"AIComputingNode/pkg/serve"

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
	configPath := flag.String("config", "", "the config file path")
	versionFlag := flag.Bool("version", false, "show version number and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		panic(err)
	}
	err = cfg.Validate()
	if err != nil {
		panic(err)
	}

	err = log.InitLogging(cfg.App.LogLevel, cfg.App.LogFile, cfg.App.LogOutput)
	if err != nil {
		panic(err)
	}

	log.Logger.Info("################################################################")
	log.Logger.Info("#                          START                               #")
	log.Logger.Info("################################################################")

	err = db.InitDb(db.InitOptions{Folder: cfg.App.Datastore})
	if err != nil {
		log.Logger.Fatalf("Init database: %v", err)
	}

	PeersHistory, err := p2p.ConvertPeersFromStringMap(db.LoadPeers())
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

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(cfg.Addresses...),
		libp2p.Identity(privKey),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.ProtocolVersion(ProtocolVersion),
		libp2p.UserAgent(version),
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
			log.Logger.Info("OnConnected remote multi-addr ", c.RemoteMultiaddr(), c.RemotePeer())
			p2p.PeerList[c.RemotePeer()] = c.RemoteMultiaddr()
			db.PeerConnected(c.RemotePeer().String(), c.RemoteMultiaddr().String())
		},
		DisconnectedF: func(n network.Network, c network.Conn) {
			log.Logger.Info("OnDisconnected remote multi-addr ", c.RemoteMultiaddr(), c.RemotePeer())
			delete(p2p.PeerList, c.RemotePeer())
			// db.PeerDisconnected(c.RemotePeer().String(), c.RemoteMultiaddr().String())
		},
	})

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

	if publishConns < 2 {
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

	psOpts := []pubsub.Option{
		pubsub.WithDiscovery(routingDiscovery),
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
		Topic:           topic,
		RD:              routingDiscovery,
	}

	publishChan := make(chan []byte, 1024)
	go ps.PublishToTopic(ctx, topic, publishChan)
	go ps.PubsubHandler(ctx, sub, publishChan)
	go serve.NewHttpServe(publishChan)

	log.Logger.Info("listening for connections")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	// select {} // hang forever
	<-stop
	host.Close()

	log.Logger.Info("################################################################")
	log.Logger.Info("#                          OVER                                #")
	log.Logger.Info("################################################################")
}
