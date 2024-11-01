package host

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"

	ips "github.com/libp2p/go-libp2p/p2p/host/peerstore"
)

func TestLatency(t *testing.T) {
	privKey1, pubKey1, err := crypto.GenerateKeyPair(crypto.Secp256k1, -1)
	if err != nil {
		t.Fatalf("generate key pair %v", err)
	}
	peer1, err := peer.IDFromPublicKey(pubKey1)
	if err != nil {
		t.Fatalf("parse peer id %v", err)
	}
	privKey2, pubKey2, err := crypto.GenerateKeyPair(crypto.Secp256k1, -1)
	if err != nil {
		t.Fatalf("generate key pair %v", err)
	}
	peer2, err := peer.IDFromPublicKey(pubKey2)
	if err != nil {
		t.Fatalf("parse peer id %v", err)
	}

	t.Logf("peer id 1 %v", peer1)
	t.Logf("peer id 2 %v", peer2)

	ips.LatencyEWMASmoothing = 0.2

	host1, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000",      // regular tcp connections
			"/ip4/0.0.0.0/udp/9000/quic", // a UDP endpoint for the QUIC transport
		),
		libp2p.Ping(false),
		libp2p.Identity(privKey1),
		libp2p.DefaultResourceManager,
	)
	if err != nil {
		t.Fatalf("create libp2p host %v", err)
	}
	host2, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9001",      // regular tcp connections
			"/ip4/0.0.0.0/udp/9001/quic", // a UDP endpoint for the QUIC transport
		),
		libp2p.Identity(privKey2),
		libp2p.DefaultResourceManager,
	)
	if err != nil {
		t.Fatalf("create libp2p host %v", err)
	}

	pingService := ping.NewPingService(host1)

	t.Logf("host1 connectivity of host2 %v", host1.Network().Connectedness(peer2))

	pi1 := peer.AddrInfo{
		ID:    host1.ID(),
		Addrs: host1.Addrs(),
	}
	// pi2 := peer.AddrInfo{
	// 	ID:    host2.ID(),
	// 	Addrs: host2.Addrs(),
	// }

	// t.Logf("host1 latency of host2 %v", host1.Peerstore().LatencyEWMA(peer2).String())
	// t.Logf("host2 latency of host1 %v", host2.Peerstore().LatencyEWMA(peer1).String())

	// if err := host1.Connect(context.Background(), pi2); err != nil {
	// 	t.Logf("host1 connect host2 failed %v", err)
	// } else {
	// 	t.Log("host1 connect host2 success")
	// }

	t.Logf("host1 latency of host2 %v", host1.Peerstore().LatencyEWMA(peer2).String())
	t.Logf("host2 latency of host1 %v", host2.Peerstore().LatencyEWMA(peer1).String())

	if err := host2.Connect(context.Background(), pi1); err != nil {
		t.Logf("host2 connect host1 failed %v", err)
	} else {
		t.Log("host2 connect host1 success")
	}

	t.Logf("host1 connectivity of host2 %v", host1.Network().Connectedness(peer2))

	t.Logf("host1 latency of host2 %v", host1.Peerstore().LatencyEWMA(peer2).String())
	t.Logf("host2 latency of host1 %v", host2.Peerstore().LatencyEWMA(peer1).String())

	for ii := 0; ii < 5; ii++ {
		ctx, cancel := context.WithCancel(context.Background())
		ch := pingService.Ping(ctx, peer2)
		for i := 0; i < 5; i++ {
			res := <-ch
			if res.Error != nil {
				t.Errorf("ping failed %v", res.Error)
			} else {
				t.Log("pinged", peer2.String(), "in", res.RTT)
			}
		}
		cancel()

		conns := host1.Network().Conns()
		for _, conn := range conns {
			t.Logf("host1 conn has %d streams with %s", len(conn.GetStreams()), conn.RemotePeer().String())
		}
		t.Logf("host1 latency of host2 %v", host1.Peerstore().LatencyEWMA(peer2).String())
		t.Logf("host2 latency of host1 %v", host2.Peerstore().LatencyEWMA(peer1).String())
		time.Sleep(3 * time.Second)
	}
	// for res := range ch {
	// 	if res.Error != nil {
	// 		t.Errorf("ping failed %v", res.Error)
	// 	} else {
	// 		t.Log("pinged", peer2.String(), "in", res.RTT)
	// 	}
	// }
}
