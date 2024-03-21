package main

import (
	"context"
	"encoding/hex"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/SuperImageAI/AIComputingNode/pkg/config"
	"github.com/SuperImageAI/AIComputingNode/pkg/log"
	"github.com/SuperImageAI/AIComputingNode/pkg/p2p"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"

	dht "github.com/libp2p/go-libp2p-kad-dht"
)

func main() {
	configPath := flag.String("config", "", "the config file path")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		panic(err)
	}
	err = cfg.Validate()
	if err != nil {
		panic(err)
	}

	err = log.SetLogLevel(cfg.App.LogLevel)
	if err != nil {
		panic(err)
	}

	DefaultBootstrapPeers, err := p2p.ConvertPeers(cfg.Bootstrap)
	if err != nil {
		log.Logger.Fatalln("Parse bootstrap: %v", err)
	}

	ctx := context.Background()
	var kadDHT *dht.IpfsDHT

	privKey, _ := p2p.PrivKeyFromString(cfg.Identity.PrivKey)

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(cfg.Addresses...),
		libp2p.Identity(privKey),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
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
	} else {
		opts = append(opts, libp2p.DisableRelay())
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
	for _, peerinfo := range DefaultBootstrapPeers {
		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()
			host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
			if err := host.Connect(ctx, pi); err != nil {
				log.Logger.Warnf("Connect bootstrap node %v : %v", pi, err)
			} else {
				log.Logger.Info("Connection established with bootstrap node:", pi)
			}
		}(peerinfo)
	}
	wg.Wait()

	err = kadDHT.Bootstrap(ctx)
	if err != nil {
		log.Logger.Fatalf("Bootstrap the host: %v", err)
	}

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		log.Logger.Fatalf("New GossipSub: %v", err)
	}
	topic, err := ps.Join(cfg.App.TopicName)
	if err != nil {
		log.Logger.Fatalf("Join GossipSub: %v", err)
	}
	defer topic.Close()
	sub, err := topic.Subscribe()
	if err != nil {
		log.Logger.Fatalf("Subscribe GossipSub: %v", err)
	}
	go pubsubHandler(ctx, sub)

	log.Logger.Info("listening for connections")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	// select {} // hang forever
	<-stop
	host.Close()
}

func pubsubHandler(ctx context.Context, sub *pubsub.Subscription) {
	defer sub.Cancel()
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			log.Logger.Warnf("Read GossipSub: %v", err)
			continue
		}

		log.Logger.Info(msg.ReceivedFrom, ": ", string(msg.Message.Data))
	}
}
