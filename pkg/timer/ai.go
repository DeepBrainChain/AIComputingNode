package timer

import (
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/libp2p/host"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/model"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/types"

	"google.golang.org/protobuf/proto"
)

func SendAIProjects(pcn chan<- []byte) {
	projects := model.GetAIProjects()
	var nt types.NodeType = 0x00
	if config.GC.Swarm.RelayService.Enabled {
		nt |= types.PublicIpFlag
	}
	if config.GC.App.PeersCollect.Enabled {
		nt |= types.PeersCollectFlag
	}
	for _, models := range projects {
		if len(models) > 0 {
			nt |= types.ModelFlag
			break
		}
	}
	aiBody := &protocol.AIProjectBody{
		Data: &protocol.AIProjectBody_Res{
			Res: types.AIProject2ProtocolMessage(projects, uint32(nt)),
		},
	}
	body, err := proto.Marshal(aiBody)
	if err != nil {
		log.Logger.Warnf("Marshal AI Project Heartbeat Body %v", err)
		return
	}
	protoMsg := protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: host.Hio.UserAgent,
			Timestamp:     time.Now().Unix(),
			Id:            "",
			NodeId:        config.GC.Identity.PeerID,
			Receiver:      "",
			NodePubKey:    nil,
			Sign:          nil,
		},
		Type:          protocol.MessageType_AI_PROJECT,
		Body:          body,
		ResultCode:    0,
		ResultMessage: "heartbeat",
	}
	message, err := proto.Marshal(&protoMsg)
	if err != nil {
		log.Logger.Errorf("Marshal AI Project Heartbeat %v", err)
		return
	}
	pcn <- message
	log.Logger.Info("Sending AI Project Heartbeat")
}
