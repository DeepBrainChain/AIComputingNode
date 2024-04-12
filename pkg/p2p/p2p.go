package p2p

import (
	"context"
	"encoding/hex"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
)

var Hio *HostInfo

var PeerList = make(map[peer.ID]multiaddr.Multiaddr)

const (
	connectionManagerTag    = "user-connect"
	connectionManagerWeight = 100
)

type HostInfo struct {
	Host            host.Host
	UserAgent       string
	ProtocolVersion string
	PrivKey         crypto.PrivKey
	Ctx             context.Context

	Topic *pubsub.Topic
	RD    *drouting.RoutingDiscovery
}

type IdentifyProtocol struct {
	ID              string
	ProtocolVersion string
	AgentVersion    string
	PublicKey       string
	Addresses       []string
	Protocols       []string
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

func (hio *HostInfo) GetIdentifyProtocol() IdentifyProtocol {
	id := IdentifyProtocol{
		ID:              hio.Host.ID().String(),
		ProtocolVersion: hio.ProtocolVersion,
		AgentVersion:    hio.UserAgent,
	}
	if pubKeyBytes, err := MarshalPubKeyFromPrivKey(hio.PrivKey); err == nil {
		id.PublicKey = hex.EncodeToString(pubKeyBytes)
	}
	for _, addr := range hio.Host.Addrs() {
		id.Addresses = append(id.Addresses, addr.String())
	}
	protos := hio.Host.Mux().Protocols()
	id.Protocols = protocol.ConvertToStrings(protos)
	return id
}

func (hio *HostInfo) FindPeers(ns string) (<-chan peer.AddrInfo, error) {
	return hio.RD.FindPeers(hio.Ctx, ns)
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

func (hio *HostInfo) SwarmConnect(addr string) error {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return err
	}
	pi, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return err
	}
	if swrm, ok := hio.Host.Network().(*swarm.Swarm); ok {
		swrm.Backoff().Clear(pi.ID)
	}
	if err := hio.Host.Connect(hio.Ctx, *pi); err != nil {
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

func (hio *HostInfo) PubsubPeers() []string {
	peers := hio.Topic.ListPeers()
	ids := make([]string, len(peers))
	for i, peer := range peers {
		ids[i] = peer.String()
	}
	return ids
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
