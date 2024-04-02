package ps

import (
	"context"
	"encoding/json"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/p2p"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/serve"

	"google.golang.org/protobuf/proto"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func PublishToTopic(ctx context.Context, topic *pubsub.Topic, messageChan <-chan []byte) {
	for message := range messageChan {
		if err := topic.Publish(ctx, message); err != nil {
			log.Logger.Errorf("Error when publish to topic %v", err)
		} else {
			log.Logger.Infof("Published %d bytes", len(message))
		}
	}
}

func PubsubHandler(ctx context.Context, sub *pubsub.Subscription, publishChan chan<- []byte) {
	defer sub.Cancel()
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			log.Logger.Warnf("Read PubSub: %v", err)
			continue
		}

		pmsg := &protocol.Message{}
		err = proto.Unmarshal(msg.Data, pmsg)
		if err != nil {
			log.Logger.Warnf("Unmarshal PubSub: %v", err)
			continue
		}

		if pmsg.Header.NodeId == config.GC.Identity.PeerID {
			log.Logger.Info("Received message type ", pmsg.Type, " from the node itself")
			continue
		} else {
			log.Logger.Info("Received message type ", pmsg.Type, " from ", pmsg.Header.NodeId)
		}

		switch pmsg.Type {
		case protocol.MesasgeType_PEER_IDENTITY_REQUEST:
			if piReq := pmsg.GetPiReq(); piReq != nil {
				if piReq.GetNodeId() == config.GC.Identity.PeerID {
					idp := p2p.Hio.GetIdentifyProtocol()
					res := protocol.Message{
						Header: &protocol.MessageHeader{
							ClientVersion: p2p.Hio.UserAgent,
							Timestamp:     time.Now().Unix(),
							Id:            pmsg.Header.Id,
							NodeId:        config.GC.Identity.PeerID,
						},
						Type: protocol.MesasgeType_PEER_IDENTITY_RESPONSE,
						Body: &protocol.Message_PiRes{
							PiRes: &protocol.PeerIdentityResponse{
								ProtocolVersion: idp.ProtocolVersion,
								AgentVersion:    idp.AgentVersion,
								PublicKey:       []byte(idp.PublicKey),
								ListenAddrs:     idp.Addresses,
								Protocols:       idp.Protocols,
							},
						},
					}
					resBytes, err := proto.Marshal(&res)
					if err != nil {
						log.Logger.Warnf("Marshal Identity Response %v", err)
						break
					}
					publishChan <- resBytes
					log.Logger.Info("Sending Peer Identity Response")
				} else {
					log.Logger.Info("Gossip Peer Identity Request of ", piReq.GetNodeId())
				}
			} else {
				log.Logger.Warn("Message type and body do not match")
			}
		case protocol.MesasgeType_PEER_IDENTITY_RESPONSE:
			if piRes := pmsg.GetPiRes(); piRes != nil {
				idp := p2p.IdentifyProtocol{
					ID:              pmsg.Header.NodeId,
					ProtocolVersion: piRes.ProtocolVersion,
					AgentVersion:    piRes.AgentVersion,
					PublicKey:       string(piRes.PublicKey),
					Addresses:       piRes.ListenAddrs,
					Protocols:       piRes.Protocols,
				}
				notifyData, err := json.Marshal(idp)
				if err != nil {
					log.Logger.Warnf("Marshal Identity Protocol %v", err)
					break
				}
				serve.QueueLock.Lock()
				for i, item := range serve.RequestQueue {
					if item.ID == pmsg.Header.Id {
						item.Notify <- notifyData
						serve.RequestQueue = append(serve.RequestQueue[:i], serve.RequestQueue[i+1:]...)
						close(item.Notify)
						break
					}
				}
				serve.QueueLock.Unlock()
			} else {
				log.Logger.Warn("Message type and body do not match")
			}
		default:
			log.Logger.Warnf("Unknowned message type", pmsg.Type)
		}
	}
}
