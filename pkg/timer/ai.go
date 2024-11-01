package timer

import (
	"context"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/model"
	"AIComputingNode/pkg/p2p"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/types"

	"google.golang.org/protobuf/proto"
)

type AITimer struct {
	PublishChan chan<- []byte
}

var AIT *AITimer

func (service AITimer) SendAIProjects() {
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
			ClientVersion: p2p.Hio.UserAgent,
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
	service.PublishChan <- message
	log.Logger.Info("Sending AI Project Heartbeat")
}

func (service AITimer) HandleBroadcastMessage(ctx context.Context, msg *protocol.Message) {
	switch msg.Type {
	case protocol.MessageType_AI_PROJECT:
		if !config.GC.Swarm.RelayService.Enabled || !config.GC.App.PeersCollect.Enabled {
			return
		}
		aip := &protocol.AIProjectBody{}
		if err := proto.Unmarshal(msg.GetBody(), aip); err != nil {
			log.Logger.Warnf("Unmarshal AI Project Heartbeat %v", err)
			return
		}
		if aiRes := aip.GetRes(); aiRes != nil {
			service.HandleAIProjectMessage(msg.Header.NodeId, types.ProtocolMessage2AIProject(aiRes), aiRes.NodeType)
		}
	default:
		log.Logger.Warnf("Unsupported heartbeat message type", msg.Type)
	}
}

func (service AITimer) HandleAIProjectMessage(node_id string, projects map[string]map[string]types.ModelInfo, nodeType uint32) {
	info := db.PeerCollectInfo{
		Timestamp:  time.Now().Unix(),
		AIProjects: projects,
		NodeType:   nodeType,
	}
	db.UpdatePeerCollect(node_id, info)
}

func StartTimer(ctx context.Context, interval time.Duration, pcn chan<- []byte) {
	AIT = &AITimer{
		PublishChan: pcn,
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Logger.Info("ai timer goroutine over")
				return
			case <-ticker.C:
				db.CleanExpiredPeerCollectInfo()
				AIT.SendAIProjects()
			}
		}
	}()
}
