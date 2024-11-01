package conngater

import (
	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/types"

	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	ma "github.com/multiformats/go-multiaddr"
)

type ConnectionGater struct {
	// blockedDialPeers   map[peer.ID]struct{}
	// blockedDialedPeers map[peer.ID]struct{}
}

func (cg *ConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	log.Logger.Infof("InterceptPeerDial {Peer.ID %s}", p.String())
	// _, ok := cg.blockedDialPeers[p]
	// return !ok
	return true
}

func (cg *ConnectionGater) InterceptAddrDial(p peer.ID, a ma.Multiaddr) (allow bool) {
	log.Logger.Infof("InterceptAddrDial {Peer.ID %s, Multiaddr %s}", p.String(), a.String())
	return true
}

func (cg *ConnectionGater) InterceptAccept(cma network.ConnMultiaddrs) (allow bool) {
	log.Logger.Infof("InterceptAccept {LocalMultiaddr %s, RemoteMultiaddr %s}",
		cma.LocalMultiaddr().String(), cma.RemoteMultiaddr().String())
	return true
}

func (cg *ConnectionGater) InterceptSecured(dir network.Direction, p peer.ID, cma network.ConnMultiaddrs) (allow bool) {
	log.Logger.Infof("InterceptSecured {Direction %s, Peer.ID %s, LocalMultiaddr %s, RemoteMultiaddr %s}",
		dir.String(), p.String(), cma.LocalMultiaddr().String(), cma.RemoteMultiaddr().String())
	if dir == network.DirInbound && config.GC.App.PeersCollect.ClientProject != "" {
		// _, ok := cg.blockedDialedPeers[p]
		// return !ok
		info := &db.PeerCollectInfo{}
		if err := db.GetAIProjectsOfNode(p.String(), info); err != nil {
			return true
		}
		nt := (types.NodeType)(info.NodeType)
		if !nt.IsModelNode() {
			return true
		}
		for pn := range info.AIProjects {
			if pn == config.GC.App.PeersCollect.ClientProject {
				return true
			}
		}
		return false
	}
	return true
}

func (cg *ConnectionGater) InterceptUpgraded(conn network.Conn) (allow bool, reason control.DisconnectReason) {
	log.Logger.Infof("InterceptUpgraded {ID %s, LocalPeer %s, RemotePeer %s}",
		conn.ID(), conn.LocalPeer().String(), conn.RemotePeer().String())
	return true, 0
}
