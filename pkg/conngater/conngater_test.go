package conngater

import (
	"context"
	"fmt"
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/conngater"

	ma "github.com/multiformats/go-multiaddr"
)

type BasicConnectionGater struct {
	blockedDialPeers   map[peer.ID]struct{}
	blockedDialedPeers map[peer.ID]struct{}
}

func (cg *BasicConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	fmt.Printf("InterceptPeerDial {Peer.ID %s}\n", p.String())
	_, ok := cg.blockedDialPeers[p]
	return !ok
	// return true
}

func (cg *BasicConnectionGater) InterceptAddrDial(p peer.ID, a ma.Multiaddr) (allow bool) {
	fmt.Printf("InterceptAddrDial {Peer.ID %s, Multiaddr %s}\n", p.String(), a.String())
	return true
}

func (cg *BasicConnectionGater) InterceptAccept(cma network.ConnMultiaddrs) (allow bool) {
	fmt.Printf("InterceptAccept {LocalMultiaddr %s, RemoteMultiaddr %s}\n",
		cma.LocalMultiaddr().String(), cma.RemoteMultiaddr().String())
	return true
}

func (cg *BasicConnectionGater) InterceptSecured(dir network.Direction, p peer.ID, cma network.ConnMultiaddrs) (allow bool) {
	fmt.Printf("InterceptSecured {Direction %s, Peer.ID %s, LocalMultiaddr %s, RemoteMultiaddr %s}\n",
		dir.String(), p.String(), cma.LocalMultiaddr().String(), cma.RemoteMultiaddr().String())
	if dir == network.DirInbound {
		_, ok := cg.blockedDialedPeers[p]
		return !ok
	}
	return true
}

func (cg *BasicConnectionGater) InterceptUpgraded(conn network.Conn) (allow bool, reason control.DisconnectReason) {
	fmt.Printf("InterceptUpgraded {ID %s, LocalPeer %s, RemotePeer %s}\n",
		conn.ID(), conn.LocalPeer().String(), conn.RemotePeer().String())
	return true, 0
}

func TestMyConnectionGater(t *testing.T) {
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

	cg1 := &BasicConnectionGater{
		blockedDialPeers:   make(map[peer.ID]struct{}),
		blockedDialedPeers: make(map[peer.ID]struct{}),
	}
	// cg1.blockedDialPeers[peer2] = struct{}{}
	cg1.blockedDialedPeers[peer2] = struct{}{}

	host1, err := libp2p.New(
		libp2p.ConnectionGater(cg1),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000", // regular tcp connections
		),
		libp2p.Identity(privKey1),
	)
	if err != nil {
		t.Fatalf("create libp2p host %v", err)
	}
	host2, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9001", // regular tcp connections
		),
		libp2p.Identity(privKey2),
	)
	if err != nil {
		t.Fatalf("create libp2p host %v", err)
	}

	pi1 := peer.AddrInfo{
		ID:    host1.ID(),
		Addrs: host1.Addrs(),
	}
	// pi2 := peer.AddrInfo{
	// 	ID:    host2.ID(),
	// 	Addrs: host2.Addrs(),
	// }

	// if err := host1.Connect(context.Background(), pi2); err != nil {
	// 	t.Logf("host1 connect host2 failed %v", err)
	// } else {
	// 	t.Log("host1 connect host2 success")
	// }

	if err := host2.Connect(context.Background(), pi1); err != nil {
		t.Logf("host2 connect host1 failed %v", err)
	} else {
		t.Log("host2 connect host1 success")
	}

	for _, conn := range host1.Network().Conns() {
		t.Logf("host1.Network().Conns() item %s with %s", conn.ID(), conn.RemotePeer().String())
	}

	for _, conn := range host2.Network().Conns() {
		t.Logf("host2.Network().Conns() item %s with %s", conn.ID(), conn.RemotePeer().String())
	}

	delete(cg1.blockedDialedPeers, peer2)

	if err := host2.Connect(context.Background(), pi1); err != nil {
		t.Logf("host2 connect host1 failed %v", err)
	} else {
		t.Log("host2 connect host1 success")
	}

	for _, conn := range host1.Network().Conns() {
		t.Logf("host1.Network().Conns() item %s with %s", conn.ID(), conn.RemotePeer().String())
	}

	for _, conn := range host2.Network().Conns() {
		t.Logf("host2.Network().Conns() item %s with %s", conn.ID(), conn.RemotePeer().String())
	}
}

func TestBasicConnectionGater(t *testing.T) {
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

	cg1, err := conngater.NewBasicConnectionGater(nil)
	if err != nil {
		t.Fatalf("new basic connection gater %v", err)
	}
	cg1.BlockPeer(peer2)

	host1, err := libp2p.New(
		libp2p.ConnectionGater(cg1),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000",      // regular tcp connections
			"/ip4/0.0.0.0/udp/9000/quic", // a UDP endpoint for the QUIC transport
		),
		libp2p.Identity(privKey1),
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
	)
	if err != nil {
		t.Fatalf("create libp2p host %v", err)
	}

	pi1 := peer.AddrInfo{
		ID:    host1.ID(),
		Addrs: host1.Addrs(),
	}
	pi2 := peer.AddrInfo{
		ID:    host2.ID(),
		Addrs: host2.Addrs(),
	}

	if err := host1.Connect(context.Background(), pi2); err != nil {
		t.Logf("host1 connect host2 failed %v", err)
	} else {
		t.Log("host1 connect host2 success")
	}

	if err := host2.Connect(context.Background(), pi1); err != nil {
		t.Logf("host2 connect host1 failed %v", err)
	} else {
		t.Log("host2 connect host1 success")
	}
}
