package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"

	golog "github.com/ipfs/go-log/v2"
)

var log = golog.Logger("ping")

const ProtocolVersion string = "aicn/0.0.1"
const AgentVersion string = "tool/ping/0.0.1"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listenF := flag.Int("l", 6000, "listening port waiting for incoming connections")
	pskString := flag.String("psk", "", "Pre-Shared Key")
	peerKeyString := flag.String("peerkey", "", "the string of peer private key")
	natport := flag.Bool("natport", false, "whether to enable nat port map")
	relayMode := flag.String("relay", "", "relay is client or service mode")
	holePunching := flag.Bool("hole", false, "whether to enable hole punching")
	router := flag.String("router", "", "the router type of DHT kad")
	protocolPrefix := flag.String("protocol", "", "the prefix attached to all DHT protocols")
	topicNameFlag := flag.String("topicName", "applesauce", "name of topic to join")
	target := flag.String("target", "", "target peer to dial")
	flag.Parse()

	golog.SetAllLoggers(golog.LevelInfo)

	if *target == "" {
		log.Fatalf("target peer is empty")
	}
	addr, err := multiaddr.NewMultiaddr(*target)
	if err != nil {
		log.Fatalf("target peer string to multiaddr failed: %v", err)
	}
	pai, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Fatalf("target peer info from string failed: %v", err)
	}

	var kadDHT *dht.IpfsDHT

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *listenF),
			fmt.Sprintf("/ip6/::/tcp/%d", *listenF),
		),
		libp2p.Ping(false),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.ProtocolVersion(ProtocolVersion),
		libp2p.UserAgent(AgentVersion),
	}

	if *pskString != "" {
		psk, err := hex.DecodeString(*pskString)
		if err != nil {
			log.Fatalf("Decode Pre-Shared Key failed: %v", err)
		}
		log.Info("Pre-Shared Key ", psk)
		opts = append(opts, libp2p.PrivateNetwork(psk), libp2p.DefaultPrivateTransports)
	} else {
		opts = append(opts, libp2p.DefaultTransports)
	}

	if *peerKeyString != "" {
		privKeyBytes, err := crypto.ConfigDecodeKey(*peerKeyString)
		if err != nil {
			log.Fatalf("Decode peer key failed: %v", err)
		}
		privKey, err := crypto.UnmarshalPrivateKey(privKeyBytes)
		if err != nil {
			log.Fatalf("Unmarshal Private Key failed: %v", err)
		}
		opts = append(opts, libp2p.Identity(privKey))
	}

	if *natport {
		opts = append(opts, libp2p.NATPortMap())
	}

	opts = append(opts, libp2p.EnableNATService())

	if *relayMode == "client" {
		opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays([]peer.AddrInfo{*pai}))
	} else if *relayMode == "service" {
		opts = append(opts, libp2p.EnableRelayService())
	}

	if *holePunching {
		opts = append(opts, libp2p.EnableHolePunching())
	}

	opts = append(opts, libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		dhtOpts := []dht.Option{}
		if *router == "" {
			dhtOpts = append(dhtOpts, dht.Mode(dht.ModeAuto))
		} else if *router == "dhtclient" {
			dhtOpts = append(dhtOpts, dht.Mode(dht.ModeClient))
		} else if *router == "dhtserver" {
			dhtOpts = append(dhtOpts, dht.Mode(dht.ModeServer))
		}
		if *protocolPrefix != "" {
			dhtOpts = append(dhtOpts, dht.ProtocolPrefix(protocol.ID(*protocolPrefix)))
		}
		var err error = nil
		kadDHT, err = dht.New(ctx, h, dhtOpts...)
		return kadDHT, err
	}))

	// start a libp2p node that listens on a random local TCP port,
	// but without running the built-in ping protocol
	node, err := libp2p.New(opts...)
	if err != nil {
		log.Fatalf("Create libp2p host: %v", err)
	}
	defer node.Close()

	log.Info("Listen addresses: ", node.Addrs())
	log.Info("Node id: ", node.ID())

	// print the node's PeerInfo in multiaddr format
	peerInfo := peer.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		log.Fatalf("AddrInfoToP2pAddrs failed: %v", err)
	}
	log.Info("libp2p node address: ", addrs) // addrs[0]

	if err := kadDHT.Bootstrap(ctx); err != nil {
		log.Fatalf("Bootstrap the host: %v", err)
	}

	routingDiscovery := drouting.NewRoutingDiscovery(kadDHT)
	dutil.Advertise(ctx, routingDiscovery, *topicNameFlag)

	// configure our own ping protocol
	pingService := &ping.PingService{Host: node}
	node.SetStreamHandler(ping.ID, pingService.PingHandler)

	if err := node.Connect(ctx, *pai); err != nil {
		log.Fatalf("Connect target error: %v", err)
	}
	log.Info("sending 5 ping messages to ", addr)
	ch := pingService.Ping(context.Background(), pai.ID)
	for i := 0; i < 5; i++ {
		res := <-ch
		log.Info("pinged ", addr, " in ", res.RTT)
	}
}
