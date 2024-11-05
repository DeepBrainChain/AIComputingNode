package host

import (
	"context"
	"fmt"
	"time"

	"AIComputingNode/pkg/types"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	libp2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

var Hio *HostInfo

const (
	connectionManagerTag    = "user-connect"
	connectionManagerWeight = 100
)

type HostInfo struct {
	Host            libp2phost.Host
	UserAgent       string
	ProtocolVersion string
	PrivKey         crypto.PrivKey

	PingService *ping.PingService

	Dht   *dht.IpfsDHT
	RD    *drouting.RoutingDiscovery
	Topic *pubsub.Topic
}

type SwarmPeerInfo struct {
	ID        string `json:"id"`
	Peer      string `json:"peer"`
	Addr      string `json:"addr"`
	Latency   string `json:"latency"`
	Direction string `json:"direction"`
}

type PeerAddrInfo struct {
	Peer  string   `json:"peer"`
	Addrs []string `json:"addrs"`
}

func (hio *HostInfo) GetIdentifyProtocol() types.IdentifyProtocol {
	id := types.IdentifyProtocol{
		ID:              hio.Host.ID().String(),
		ProtocolVersion: hio.ProtocolVersion,
		AgentVersion:    hio.UserAgent,
	}
	for _, addr := range hio.Host.Addrs() {
		id.Addresses = append(id.Addresses, addr.String())
	}
	protos := hio.Host.Mux().Protocols()
	id.Protocols = protocol.ConvertToStrings(protos)
	return id
}

func (hio *HostInfo) FindPeers(ctx context.Context, ns string) (<-chan peer.AddrInfo, error) {
	return hio.RD.FindPeers(ctx, ns)
}

func (hio *HostInfo) SwarmPeers() []SwarmPeerInfo {
	conns := hio.Host.Network().Conns()
	pinfos := make([]SwarmPeerInfo, len(conns))

	for i, c := range conns {
		pinfos[i] = SwarmPeerInfo{
			ID:        c.ID(),
			Peer:      c.RemotePeer().String(),
			Addr:      c.RemoteMultiaddr().String(),
			Latency:   hio.Host.Peerstore().LatencyEWMA(c.RemotePeer()).String(),
			Direction: c.Stat().Direction.String(),
		}
	}
	return pinfos
}

func (hio *HostInfo) SwarmAddrs() []PeerAddrInfo {
	ps := hio.Host.Network().Peerstore()
	addrs := make([]PeerAddrInfo, 0)
	for _, p := range ps.Peers() {
		peer := PeerAddrInfo{
			Peer: p.String(),
		}
		maddrs := ps.Addrs(p)
		for _, addr := range maddrs {
			peer.Addrs = append(peer.Addrs, addr.String())
		}
		addrs = append(addrs, peer)
	}
	return addrs
}

func (hio *HostInfo) SwarmConnect(ctx context.Context, addr string) error {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return err
	}
	pi, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return err
	}
	if hio.Host.Network().Connectedness(pi.ID) == network.Connected {
		return nil
	}
	if swrm, ok := hio.Host.Network().(*swarm.Swarm); ok {
		swrm.Backoff().Clear(pi.ID)
	}
	if err := hio.Host.Connect(ctx, *pi); err != nil {
		return err
	}
	hio.Host.ConnManager().TagPeer(pi.ID, connectionManagerTag, connectionManagerWeight)
	return nil
}

func (hio *HostInfo) SwarmDisconnect(addr string) error {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return err
	}
	pi, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return err
	}
	net := hio.Host.Network()
	if net.Connectedness(pi.ID) != network.Connected {
		return fmt.Errorf("not connected")
	}
	if err := net.ClosePeer(pi.ID); err != nil {
		return err
	}
	return nil
}

func (hio *HostInfo) SwarmConnectBootstrap(ctx context.Context, addr string) error {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return err
	}
	pi, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return err
	}
	if hio.Host.Network().Connectedness(pi.ID) == network.Connected {
		hio.Host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
		return nil
	}
	if swrm, ok := hio.Host.Network().(*swarm.Swarm); ok {
		swrm.Backoff().Clear(pi.ID)
	}
	if err := hio.Host.Connect(ctx, *pi); err != nil {
		return err
	}
	hio.Host.ConnManager().TagPeer(pi.ID, connectionManagerTag, connectionManagerWeight)
	hio.Host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
	return nil
}

func (hio *HostInfo) SwarmDisconnectBootstrap(addr string) error {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return err
	}
	pi, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return err
	}
	net := hio.Host.Network()
	if net.Connectedness(pi.ID) != network.Connected {
		hio.Host.Peerstore().ClearAddrs(pi.ID)
		return nil
	}
	if err := net.ClosePeer(pi.ID); err != nil {
		return err
	}
	hio.Host.Peerstore().ClearAddrs(pi.ID)
	return nil
}

func (hio *HostInfo) PubsubPeers() []string {
	peers := hio.Topic.ListPeers()
	ids := make([]string, len(peers))
	for i, peer := range peers {
		ids[i] = peer.String()
	}
	return ids
}

func (hio *HostInfo) GetPublicKey(ctx context.Context, p peer.ID) (crypto.PubKey, error) {
	if pubKey := hio.Host.Peerstore().PubKey(p); pubKey != nil {
		return pubKey, nil
	}
	return hio.Dht.GetPublicKey(ctx, p)
}

func (hio *HostInfo) NewStream(ctx context.Context, nodeId string) (network.Stream, error) {
	peer, err := peer.Decode(nodeId)
	if err != nil {
		return nil, err
	}
	return hio.Host.NewStream(ctx, peer, types.ChatProxyProtocol)
}

func (hio *HostInfo) Connectedness(nodeId string) int {
	peer, err := peer.Decode(nodeId)
	if err != nil {
		return int(network.NotConnected)
	}
	return int(hio.Host.Network().Connectedness(peer))
}

func (hio *HostInfo) Latency(nodeId string) time.Duration {
	peer, err := peer.Decode(nodeId)
	if err != nil {
		return time.Duration(0)
	}
	return hio.Host.Peerstore().LatencyEWMA(peer)
}

func PrivKeyFromString(pk string) (crypto.PrivKey, error) {
	privKeyBytes, err := crypto.ConfigDecodeKey(pk)
	if err != nil {
		return nil, err
	}
	privKey, err := crypto.UnmarshalPrivateKey(privKeyBytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func PeerIDFromPrivKeyString(pk string) (peer.ID, error) {
	privKey, err := PrivKeyFromString(pk)
	if err != nil {
		return "", err
	}
	peer, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return "", err
	}
	return peer, nil
}

func MarshalPubKeyFromPrivKey(priv crypto.PrivKey) ([]byte, error) {
	pubKey := priv.GetPublic()
	pubKeyBytes, err := crypto.MarshalPublicKey(pubKey)
	return pubKeyBytes, err
}

func ConvertPeersFromStringArray(peers []string) ([]peer.AddrInfo, error) {
	pinfos := make([]peer.AddrInfo, len(peers))
	for i, addr := range peers {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return []peer.AddrInfo{}, err
		}
		p, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return []peer.AddrInfo{}, err
		}
		pinfos[i] = *p
	}
	return pinfos, nil
}

func ConvertPeersFromStringMap(peers map[string]string) ([]peer.AddrInfo, error) {
	pinfos := make([]peer.AddrInfo, 0, len(peers))
	for key, value := range peers {
		pi, err := peer.Decode(key)
		if err != nil {
			return []peer.AddrInfo{}, err
		}
		maddr, err := multiaddr.NewMultiaddr(value)
		if err != nil {
			return []peer.AddrInfo{}, err
		}
		pinfos = append(pinfos, peer.AddrInfo{
			ID:    pi,
			Addrs: []multiaddr.Multiaddr{maddr},
		})
	}
	return pinfos, nil
}

func IsPublicNode(pi peer.AddrInfo) bool {
	for _, addr := range pi.Addrs {
		if manet.IsPublicAddr(addr) {
			return true
		}
	}
	return false
}
