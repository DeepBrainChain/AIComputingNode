package ps

import (
	"context"
	"encoding/json"
	"path/filepath"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/ipfs"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/p2p"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/serve"

	"google.golang.org/protobuf/proto"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

var MsgNotSupported string = "Not supported"

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
		if err := proto.Unmarshal(msg.Data, pmsg); err != nil {
			log.Logger.Warnf("Unmarshal PubSub: %v", err)
			continue
		}

		if pmsg.Header.NodeId == config.GC.Identity.PeerID {
			log.Logger.Infof("Received message type %s from the node itself", pmsg.Type)
			continue
		} else if pmsg.Header.Receiver != config.GC.Identity.PeerID {
			log.Logger.Infof("Gossip message type %s from %s to %s", pmsg.Type, pmsg.Header.NodeId, pmsg.Header.Receiver)
			continue
		} else {
			log.Logger.Infof("Received message type %s from %s", pmsg.Type, pmsg.Header.NodeId)
		}

		msgBody, err := p2p.Decrypt(pmsg.Header.NodePubKey, pmsg.Body)
		if err != nil {
			log.Logger.Warnf("Decrypt message body failed %v", err)
			continue
		}

		switch pmsg.Type {
		case protocol.MessageType_PEER_IDENTITY:
			pi := &protocol.PeerIdentityBody{}
			if pmsg.ResultCode != 0 {
				res := serve.PeerResponse{
					Code:    int(pmsg.ResultCode),
					Message: pmsg.ResultMessage,
					Data:    p2p.IdentifyProtocol{},
				}
				notifyData, err := json.Marshal(res)
				if err != nil {
					log.Logger.Warnf("Marshal Identity Protocol %v", err)
					break
				}
				serve.WriteAndDeleteRequestItem(pmsg.Header.Id, notifyData)
			} else if err := proto.Unmarshal(msgBody, pi); err == nil {
				if piReq := pi.GetReq(); piReq != nil {
					if piReq.GetNodeId() == config.GC.Identity.PeerID {
						idp := p2p.Hio.GetIdentifyProtocol()
						piBody := &protocol.PeerIdentityBody{
							Data: &protocol.PeerIdentityBody_Res{
								Res: &protocol.PeerIdentityResponse{
									ProtocolVersion: idp.ProtocolVersion,
									AgentVersion:    idp.AgentVersion,
									ListenAddrs:     idp.Addresses,
									Protocols:       idp.Protocols,
								},
							},
						}
						resBody, err := proto.Marshal(piBody)
						if err != nil {
							log.Logger.Warnf("Marshal Identity Response Body %v", err)
							break
						}
						resBody, err = p2p.Encrypt(pmsg.Header.NodeId, resBody)
						res := protocol.Message{
							Header: &protocol.MessageHeader{
								ClientVersion: p2p.Hio.UserAgent,
								Timestamp:     time.Now().Unix(),
								Id:            pmsg.Header.Id,
								NodeId:        config.GC.Identity.PeerID,
								Receiver:      pmsg.Header.NodeId,
							},
							Type:       protocol.MessageType_PEER_IDENTITY,
							Body:       resBody,
							ResultCode: 0,
						}
						if err == nil {
							res.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
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
				} else if piRes := pi.GetRes(); piRes != nil {
					res := serve.PeerResponse{
						Code:    0,
						Message: "ok",
						Data: p2p.IdentifyProtocol{
							ID:              pmsg.Header.NodeId,
							ProtocolVersion: piRes.ProtocolVersion,
							AgentVersion:    piRes.AgentVersion,
							Addresses:       piRes.ListenAddrs,
							Protocols:       piRes.Protocols,
						},
					}
					notifyData, err := json.Marshal(res)
					if err != nil {
						log.Logger.Warnf("Marshal Identity Protocol %v", err)
						break
					}
					serve.WriteAndDeleteRequestItem(pmsg.Header.Id, notifyData)
				}
			} else {
				log.Logger.Warn("Message type and body do not match")
			}
		case protocol.MessageType_IMAGE_GENERATION:
			ig := &protocol.ImageGenerationBody{}
			if pmsg.ResultCode != 0 {
				res := serve.ImageGenerationResponse{
					Code:    int(pmsg.ResultCode),
					Message: pmsg.ResultMessage,
				}
				notifyData, err := json.Marshal(res)
				if err != nil {
					log.Logger.Warnf("Marshal Identity Protocol %v", err)
					break
				}
				serve.WriteAndDeleteRequestItem(pmsg.Header.Id, notifyData)
			} else if err := proto.Unmarshal(msgBody, ig); err == nil {
				if igReq := ig.GetReq(); igReq != nil {
					if igReq.GetNodeId() == config.GC.Identity.PeerID {
						// TODO: Run the model and upload the images generated by the model to the IPFS node
						var ipfsAddr string = "/ip4/192.168.1.159/tcp/4002"
						var filePath string = "D:\\Code\\AIComputingNode\\tools\\ipfs\\tux.png"
						cid, code, err := ipfs.UploadImage(ctx, ipfsAddr, filePath)
						var msg string = "ok"
						if err != nil {
							msg = err.Error()
							log.Logger.Errorf("Failed to upload image %s to ipfs endpoint %s", filePath, ipfsAddr)
						}
						igBody := &protocol.ImageGenerationBody{
							Data: &protocol.ImageGenerationBody_Res{
								Res: &protocol.ImageGenerationResponse{
									IpfsNode:  ipfsAddr,
									Cid:       cid,
									ImageName: filepath.Base(filePath),
								},
							},
						}
						resBody, err := proto.Marshal(igBody)
						if err != nil {
							log.Logger.Warnf("Marshal Image Generation Response Body %v", err)
							break
						}
						resBody, err = p2p.Encrypt(pmsg.Header.NodeId, resBody)
						res := protocol.Message{
							Header: &protocol.MessageHeader{
								ClientVersion: p2p.Hio.UserAgent,
								Timestamp:     time.Now().Unix(),
								Id:            pmsg.Header.Id,
								NodeId:        config.GC.Identity.PeerID,
								Receiver:      pmsg.Header.NodeId,
							},
							Type:          protocol.MessageType_IMAGE_GENERATION,
							Body:          resBody,
							ResultCode:    int32(code),
							ResultMessage: msg,
						}
						if err == nil {
							res.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
						}
						resBytes, err := proto.Marshal(&res)
						if err != nil {
							log.Logger.Errorf("Marshal Image Generation Response %v", err)
							break
						}
						publishChan <- resBytes
						log.Logger.Info("Sending Image Generation Response")
					} else {
						log.Logger.Info("Gossip Image Generation Request of ", igReq.GetNodeId())
					}
				} else if igRes := ig.GetRes(); igRes != nil {
					res := serve.ImageGenerationResponse{
						Code:    int(pmsg.ResultCode),
						Message: pmsg.ResultMessage,
					}
					if pmsg.ResultCode == 0 {
						res.Data.IpfsNode = igRes.IpfsNode
						res.Data.CID = igRes.Cid
						res.Data.ImageName = igRes.ImageName
					}
					notifyData, err := json.Marshal(res)
					if err != nil {
						log.Logger.Warnf("Marshal Image Generation Response %v", err)
						break
					}
					serve.WriteAndDeleteRequestItem(pmsg.Header.Id, notifyData)
				}
			} else {
				log.Logger.Warn("Message type and body do not match")
			}
		default:
			res := protocol.Message{
				Header: &protocol.MessageHeader{
					ClientVersion: p2p.Hio.UserAgent,
					Timestamp:     time.Now().Unix(),
					Id:            pmsg.Header.Id,
					NodeId:        config.GC.Identity.PeerID,
					Receiver:      pmsg.Header.NodeId,
				},
				Type:          pmsg.Type,
				Body:          nil,
				ResultCode:    serve.ErrCodeUnsupported,
				ResultMessage: MsgNotSupported,
			}
			resBytes, err := proto.Marshal(&res)
			if err != nil {
				log.Logger.Errorf("Marshal Image Generation Response %v", err)
				break
			}
			publishChan <- resBytes
			log.Logger.Warnf("Unknowned message type", pmsg.Type)
		}
	}
}
