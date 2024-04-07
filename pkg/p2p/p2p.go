package p2p

import (
	"context"
	"encoding/hex"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

var Hio *HostInfo

var PeerList = make(map[peer.ID]multiaddr.Multiaddr)

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
