package ps

import (
	"context"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/p2p"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/types"

	"google.golang.org/protobuf/proto"
)

type HeartbeatService struct {
	Interval    time.Duration
	Timer       *time.Ticker
	PublishChan chan<- []byte
	DoneChan    chan bool
}

var hbs *HeartbeatService

func (service HeartbeatService) Heartbeat() {
	for {
		select {
		case <-service.DoneChan:
			log.Logger.Info("heartbeat goroutine over")
			return
		case <-service.Timer.C:
			service.SendAIProjects()
		}
	}
}

func (service HeartbeatService) SendAIProjects() {
	projects := config.GC.GetAIProjectsOfNode()
	aiBody := &protocol.AIProjectBody{
		Data: &protocol.AIProjectBody_Res{
			Res: AIProject2ProtocolMessage(projects),
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

func (service HeartbeatService) HandleBroadcastMessage(ctx context.Context, msg *protocol.Message) {
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
			service.HandleAIProjectMessage(msg.Header.NodeId, ProtocolMessage2AIProject(aiRes))
		}
	default:
		log.Logger.Warnf("Unsupported heartbeat message type", msg.Type)
	}
}

func (service HeartbeatService) HandleAIProjectMessage(node_id string, projects []types.AIProjectOfNode) {
	info := db.PeerCollectInfo{
		Timestamp:  time.Now().Unix(),
		AIProjects: projects,
	}
	db.UpdatePeerCollect(node_id, info)
}

func StartHeartbeatService(interval time.Duration, pcn chan<- []byte, done chan bool) {
	hbs = &HeartbeatService{
		Interval:    interval,
		Timer:       time.NewTicker(interval),
		PublishChan: pcn,
		DoneChan:    done,
	}
	go hbs.Heartbeat()
}
