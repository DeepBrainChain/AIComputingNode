package ps

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/host"
	"AIComputingNode/pkg/ipfs"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/model"
	"AIComputingNode/pkg/p2p"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/serve"
	"AIComputingNode/pkg/types"

	"google.golang.org/protobuf/proto"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

var MsgNotSupported string = "Not supported"

type ModelResult struct {
	Timestamp     time.Time
	Code          int
	Message       string
	ImageFilePath string
	IpfsAddr      string
	Cid           string
}

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

		go handleBroadcastMessage(ctx, pmsg, publishChan)
	}
}

func handleBroadcastMessage(ctx context.Context, msg *protocol.Message, publishChan chan<- []byte) {
	msgBody, err := p2p.Decrypt(msg.Header.NodePubKey, msg.Body)
	if err != nil {
		res := TransformErrorResponse(msg, int32(types.ErrCodeDecrypt), types.ErrCodeDecrypt.String())
		resBytes, err := proto.Marshal(res)
		if err != nil {
			log.Logger.Errorf("Marshal ErrCodeDecrypt Response %v", err)
		}
		publishChan <- resBytes
		log.Logger.Warnf("Decrypt message body failed %v", err)
		return
	}

	switch msg.Type {
	case protocol.MessageType_PEER_IDENTITY:
		handlePeerIdentityMessage(ctx, msg, msgBody, publishChan)
	case protocol.MessageType_HOST_INFO:
		handleHostInfoMessage(ctx, msg, msgBody, publishChan)
	case protocol.MessageType_CHAT_COMPLETION:
		handleChatCompletionMessage(ctx, msg, msgBody, publishChan)
	case protocol.MessageType_IMAGE_GENERATION:
		handleImageGenerationMessage(ctx, msg, msgBody, publishChan)
	default:
		res := TransformErrorResponse(msg, int32(types.ErrCodeUnsupported), MsgNotSupported)
		resBytes, err := proto.Marshal(res)
		if err != nil {
			log.Logger.Errorf("Marshal Unsupported Response %v", err)
			break
		}
		publishChan <- resBytes
		log.Logger.Warnf("Unknowned message type", msg.Type)
	}
}

func handlePeerIdentityMessage(ctx context.Context, msg *protocol.Message, decBody []byte, publishChan chan<- []byte) {
	pi := &protocol.PeerIdentityBody{}
	if msg.ResultCode != 0 {
		res := serve.PeerResponse{
			Code:    int(msg.ResultCode),
			Message: msg.ResultMessage,
			Data:    p2p.IdentifyProtocol{},
		}
		notifyData, err := json.Marshal(res)
		if err != nil {
			log.Logger.Errorf("Marshal Identity Protocol %v", err)
			return
		}
		serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
	} else if err := proto.Unmarshal(decBody, pi); err == nil {
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
					log.Logger.Errorf("Marshal Identity Response Body %v", err)
					return
				}
				resBody, err = p2p.Encrypt(msg.Header.NodeId, resBody)
				res := protocol.Message{
					Header: &protocol.MessageHeader{
						ClientVersion: p2p.Hio.UserAgent,
						Timestamp:     time.Now().Unix(),
						Id:            msg.Header.Id,
						NodeId:        config.GC.Identity.PeerID,
						Receiver:      msg.Header.NodeId,
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
					log.Logger.Errorf("Marshal Identity Response %v", err)
					return
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
					ID:              msg.Header.NodeId,
					ProtocolVersion: piRes.ProtocolVersion,
					AgentVersion:    piRes.AgentVersion,
					Addresses:       piRes.ListenAddrs,
					Protocols:       piRes.Protocols,
				},
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal Identity Protocol %v", err)
				return
			}
			serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
		}
	} else {
		log.Logger.Warn("Message type and body do not match")
	}
}

func handleChatCompletionMessage(ctx context.Context, msg *protocol.Message, decBody []byte, publishChan chan<- []byte) {
	ccb := &protocol.ChatCompletionBody{}
	if msg.ResultCode != 0 {
		res := serve.ChatCompletionResponse{
			Code:    int(msg.ResultCode),
			Message: msg.ResultMessage,
		}
		notifyData, err := json.Marshal(res)
		if err != nil {
			log.Logger.Errorf("Marshal Chat Completion Protocol %v", err)
			return
		}
		serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
	} else if err := proto.Unmarshal(decBody, ccb); err == nil {
		if chatReq := ccb.GetReq(); chatReq != nil {
			if chatReq.GetNodeId() == config.GC.Identity.PeerID {
				code, message, chatRes := handleChatCompletionRequest(ctx, chatReq)
				igBody := &protocol.ChatCompletionBody{
					Data: &protocol.ChatCompletionBody_Res{
						Res: chatRes,
					},
				}
				resBody, err := proto.Marshal(igBody)
				if err != nil {
					log.Logger.Errorf("Marshal Chat Completion Response Body %v", err)
					return
				}
				resBody, err = p2p.Encrypt(msg.Header.NodeId, resBody)
				res := protocol.Message{
					Header: &protocol.MessageHeader{
						ClientVersion: p2p.Hio.UserAgent,
						Timestamp:     chatRes.Created,
						Id:            msg.Header.Id,
						NodeId:        config.GC.Identity.PeerID,
						Receiver:      msg.Header.NodeId,
					},
					Type:          protocol.MessageType_CHAT_COMPLETION,
					Body:          resBody,
					ResultCode:    int32(code),
					ResultMessage: message,
				}
				if err == nil {
					res.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
				}
				resBytes, err := proto.Marshal(&res)
				if err != nil {
					log.Logger.Errorf("Marshal Chat Completion Response %v", err)
					return
				}
				publishChan <- resBytes
				log.Logger.Info("Sending Chat Completion Response")
			} else {
				log.Logger.Info("Gossip Chat Completion Request of ", chatReq.GetNodeId())
			}
		} else if chatRes := ccb.GetRes(); chatRes != nil {
			res := serve.ChatCompletionResponse{
				Code:    int(msg.ResultCode),
				Message: msg.ResultMessage,
			}
			if msg.ResultCode == 0 {
				res.Data.Created = chatRes.Created
				for _, choice := range chatRes.Choices {
					ccmsg := serve.ChatCompletionMessage{
						Role:    choice.GetMessage().GetRole(),
						Content: choice.GetMessage().GetContent(),
					}
					res.Data.Choices = append(res.Data.Choices, serve.ChatResponseChoice{
						Index:        int(choice.GetIndex()),
						Message:      ccmsg,
						FinishReason: choice.GetFinishReason(),
					})
				}
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal Chat Completion Response %v", err)
				return
			}
			serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
		}
	} else {
		log.Logger.Warn("Message type and body do not match")
	}
}

func handleImageGenerationMessage(ctx context.Context, msg *protocol.Message, decBody []byte, publishChan chan<- []byte) {
	ig := &protocol.ImageGenerationBody{}
	if msg.ResultCode != 0 {
		res := serve.ImageGenerationResponse{
			Code:    int(msg.ResultCode),
			Message: msg.ResultMessage,
		}
		notifyData, err := json.Marshal(res)
		if err != nil {
			log.Logger.Errorf("Marshal Image Generation Protocol %v", err)
			return
		}
		serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
	} else if err := proto.Unmarshal(decBody, ig); err == nil {
		if igReq := ig.GetReq(); igReq != nil {
			if igReq.GetNodeId() == config.GC.Identity.PeerID {
				modelResult := handleImageGenerationRequest(ctx, igReq)
				igBody := &protocol.ImageGenerationBody{
					Data: &protocol.ImageGenerationBody_Res{
						Res: &protocol.ImageGenerationResponse{
							IpfsNode:  modelResult.IpfsAddr,
							Cid:       modelResult.Cid,
							ImageName: filepath.Base(modelResult.ImageFilePath),
						},
					},
				}
				resBody, err := proto.Marshal(igBody)
				if err != nil {
					log.Logger.Errorf("Marshal Image Generation Response Body %v", err)
					return
				}
				resBody, err = p2p.Encrypt(msg.Header.NodeId, resBody)
				res := protocol.Message{
					Header: &protocol.MessageHeader{
						ClientVersion: p2p.Hio.UserAgent,
						Timestamp:     modelResult.Timestamp.Unix(),
						Id:            msg.Header.Id,
						NodeId:        config.GC.Identity.PeerID,
						Receiver:      msg.Header.NodeId,
					},
					Type:          protocol.MessageType_IMAGE_GENERATION,
					Body:          resBody,
					ResultCode:    int32(modelResult.Code),
					ResultMessage: modelResult.Message,
				}
				if err == nil {
					res.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
				}
				resBytes, err := proto.Marshal(&res)
				if err != nil {
					log.Logger.Errorf("Marshal Image Generation Response %v", err)
					return
				}
				publishChan <- resBytes
				log.Logger.Info("Sending Image Generation Response")
			} else {
				log.Logger.Info("Gossip Image Generation Request of ", igReq.GetNodeId())
			}
		} else if igRes := ig.GetRes(); igRes != nil {
			res := serve.ImageGenerationResponse{
				Code:    int(msg.ResultCode),
				Message: msg.ResultMessage,
			}
			if msg.ResultCode == 0 {
				res.Data.IpfsNode = igRes.IpfsNode
				res.Data.CID = igRes.Cid
				res.Data.ImageName = igRes.ImageName
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal Image Generation Response %v", err)
				return
			}
			serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
		}
	} else {
		log.Logger.Warn("Message type and body do not match")
	}
}

func handleHostInfoMessage(ctx context.Context, msg *protocol.Message, decBody []byte, publishChan chan<- []byte) {
	hi := &protocol.HostInfoBody{}
	if msg.ResultCode != 0 {
		res := serve.HostInfoResponse{
			Code:    int(msg.ResultCode),
			Message: msg.ResultMessage,
		}
		notifyData, err := json.Marshal(res)
		if err != nil {
			log.Logger.Errorf("Marshal Host Info Protocol %v", err)
			return
		}
		serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
	} else if err := proto.Unmarshal(decBody, hi); err == nil {
		if hiReq := hi.GetReq(); hiReq != nil {
			if hiReq.GetNodeId() == config.GC.Identity.PeerID {
				hostInfo, err := host.GetHostInfo()
				var code int32 = 0
				var message string = "ok"
				if err != nil {
					code = int32(types.ErrCodeHostInfo)
					message = err.Error()
				}
				hiRes := HostInfo2ProtocolMessage(hostInfo)
				hiBody := &protocol.HostInfoBody{
					Data: &protocol.HostInfoBody_Res{
						Res: hiRes,
					},
				}
				resBody, err := proto.Marshal(hiBody)
				if err != nil {
					log.Logger.Warnf("Marshal HostInfo Response Body %v", err)
					return
				}
				resBody, err = p2p.Encrypt(msg.Header.NodeId, resBody)
				res := protocol.Message{
					Header: &protocol.MessageHeader{
						ClientVersion: p2p.Hio.UserAgent,
						Timestamp:     time.Now().Unix(),
						Id:            msg.Header.Id,
						NodeId:        config.GC.Identity.PeerID,
						Receiver:      msg.Header.NodeId,
					},
					Type:          protocol.MessageType_HOST_INFO,
					Body:          resBody,
					ResultCode:    code,
					ResultMessage: message,
				}
				if err == nil {
					res.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
				}
				resBytes, err := proto.Marshal(&res)
				if err != nil {
					log.Logger.Errorf("Marshal HostInfo Response %v", err)
					return
				}
				publishChan <- resBytes
				log.Logger.Info("Sending HostInfo Response")
			} else {
				log.Logger.Warnf("Invalid node id %v in request body", hiReq.GetNodeId())
			}
		} else if hiRes := hi.GetRes(); hiRes != nil {
			res := serve.HostInfoResponse{
				Code:    int(msg.ResultCode),
				Message: msg.ResultMessage,
				Data:    *ProtocolMessage2HostInfo(hiRes),
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal HostInfo Protocol %v", err)
				return
			}
			serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
		}
	} else {
		log.Logger.Warn("Message type and body do not match")
	}
}

func TransformErrorResponse(msg *protocol.Message, code int32, message string) *protocol.Message {
	res := protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: p2p.Hio.UserAgent,
			Timestamp:     time.Now().Unix(),
			Id:            msg.Header.Id,
			NodeId:        config.GC.Identity.PeerID,
			Receiver:      msg.Header.NodeId,
			NodePubKey:    nil,
			Sign:          nil,
		},
		Type:          msg.Type,
		Body:          nil,
		ResultCode:    code,
		ResultMessage: message,
	}
	return &res
}

func handleChatCompletionRequest(ctx context.Context, req *protocol.ChatCompletionRequest) (int, string, *protocol.ChatCompletionResponse) {
	chatReq := model.ChatCompletionRequest{
		Model: req.Model,
	}
	for _, ccm := range req.Messages {
		chatReq.Messages = append(chatReq.Messages, model.ChatCompletionMessage{
			Role:    ccm.Role,
			Content: ccm.Content,
		})
	}
	chatRes := model.ChatModel(config.GC.App.ModelAPI, chatReq)
	response := &protocol.ChatCompletionResponse{}
	if chatRes.Code != 0 {
		return chatRes.Code, chatRes.Message, response
	}
	response.Created = chatRes.Data.Created
	for _, choice := range chatRes.Data.Choices {
		response.Choices = append(response.Choices, &protocol.ChatCompletionResponse_ChatResponseChoice{
			Index: int32(choice.Index),
			Message: &protocol.ChatCompletionMessage{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: choice.FinishReason,
		})
	}
	return chatRes.Code, chatRes.Message, response
}

func handleImageGenerationRequest(ctx context.Context, req *protocol.ImageGenerationRequest) *ModelResult {
	// TODO: Run the model and upload the images generated by the model to the IPFS node
	// var ipfsAddr string = "/ip4/192.168.1.159/tcp/4002"
	res := &ModelResult{
		Timestamp:     time.Now(),
		Code:          0,
		Message:       "ok",
		ImageFilePath: "",
		IpfsAddr:      req.GetIpfsNode(),
		Cid:           "",
	}

	if res.IpfsAddr == "" {
		res.IpfsAddr = config.GC.App.IpfsStorageAPI
	}
	if res.IpfsAddr == "" {
		res.Code = int(types.ErrCodeModel)
		res.Message = "No ipfs storage server"
		return res
	}

	if config.ValidateIpfsServer(res.IpfsAddr) {
		res.Code, res.Message, res.ImageFilePath = model.ExecuteModel(
			config.GC.App.ModelAPI, req.GetModel(), req.GetPromptWord())
		if res.Code == 0 {
			log.Logger.Infof("Execute model %s with %q result %q",
				req.GetModel(), req.GetPromptWord(), res.ImageFilePath)
		} else {
			log.Logger.Errorf("Execute model %s with %q error %s",
				req.GetModel(), req.GetPromptWord(), res.Message)
		}
	} else {
		res.Code = int(types.ErrCodeModel)
		res.Message = "Unavailable ipfs storage node"
		return res
	}

	var err error = nil
	if res.Code == 0 {
		res.Cid, res.Code, err = ipfs.UploadFile(ctx, HttpUrl2Multiaddr(res.IpfsAddr), res.ImageFilePath)
		if err != nil {
			res.Message = err.Error()
			log.Logger.Errorf("Upload image %s to %s failed %v", res.ImageFilePath, res.IpfsAddr, err)
		} else {
			log.Logger.Infof("Upload image %s to %s with %s success", res.ImageFilePath, res.IpfsAddr, res.Cid)
			if err := ipfs.WriteMFSHistory(res.Timestamp.Unix(), res.IpfsAddr,
				req.GetModel(), req.GetPromptWord(), res.Cid, filepath.Base(res.ImageFilePath)); err != nil {
				log.Logger.Errorf("Write ipfs mfs history failed %v", err)
			} else {
				log.Logger.Info("Write ipfs mfs history success")
			}
		}
	}
	_ = db.WriteModelHistory(res.Timestamp.Unix(), res.Code, res.Message,
		req.GetModel(), req.GetPromptWord(), res.IpfsAddr, res.Cid, res.ImageFilePath)
	return res
}

func HttpUrl2Multiaddr(url string) string {
	urlParts := strings.Split(url, "//")
	if len(urlParts) != 2 {
		return ""
	}
	url = urlParts[1]

	urlParts = strings.Split(url, "/")
	if len(urlParts) < 1 {
		return ""
	}
	url = urlParts[0]

	urlParts = strings.Split(url, ":")
	if len(urlParts) != 2 {
		return ""
	}
	return fmt.Sprintf("/ip4/%s/tcp/%s", urlParts[0], urlParts[1])
}

func HostInfo2ProtocolMessage(hostInfo *host.HostInfo) *protocol.HostInfoResponse {
	res := &protocol.HostInfoResponse{
		Os: &protocol.HostInfoResponse_OSInfo{
			Os:              hostInfo.Os.OS,
			Platform:        hostInfo.Os.Platform,
			PlatformFamily:  hostInfo.Os.PlatformFamily,
			PlatformVersion: hostInfo.Os.PlatformVersion,
			KernelVersion:   hostInfo.Os.KernelVersion,
			KernelArch:      hostInfo.Os.KernelArch,
		},
		Memory: &protocol.HostInfoResponse_MemoryInfo{
			TotalPhysicalBytes: hostInfo.Memory.TotalPhysicalBytes,
			TotalUsableBytes:   hostInfo.Memory.TotalUsableBytes,
		},
	}
	for _, cpu := range hostInfo.Cpu {
		res.Cpu = append(res.Cpu, &protocol.HostInfoResponse_CpuInfo{
			ModelName:    cpu.ModelName,
			TotalCores:   cpu.Cores,
			TotalThreads: cpu.Threads,
		})
	}
	for _, disk := range hostInfo.Disk {
		res.Disk = append(res.Disk, &protocol.HostInfoResponse_DiskInfo{
			DriveType:    disk.DriveType,
			SizeBytes:    disk.SizeBytes,
			Model:        disk.Model,
			SerialNumber: disk.SerialNumber,
		})
	}
	for _, gpu := range hostInfo.Gpu {
		res.Gpu = append(res.Gpu, &protocol.HostInfoResponse_GpuInfo{
			Vendor:  gpu.Vendor,
			Product: gpu.Product,
		})
	}
	return res
}

func ProtocolMessage2HostInfo(res *protocol.HostInfoResponse) *host.HostInfo {
	hostInfo := &host.HostInfo{
		Os: host.OSInfo{
			OS:              res.Os.Os,
			Platform:        res.Os.Platform,
			PlatformFamily:  res.Os.PlatformFamily,
			PlatformVersion: res.Os.PlatformVersion,
			KernelVersion:   res.Os.KernelVersion,
			KernelArch:      res.Os.KernelArch,
		},
		Memory: host.MemoryInfo{
			TotalPhysicalBytes: res.Memory.TotalPhysicalBytes,
			TotalUsableBytes:   res.Memory.TotalUsableBytes,
		},
	}
	for _, cpu := range res.Cpu {
		hostInfo.Cpu = append(hostInfo.Cpu, host.CpuInfo{
			ModelName: cpu.ModelName,
			Cores:     cpu.TotalCores,
			Threads:   cpu.TotalThreads,
		})
	}
	for _, disk := range res.Disk {
		hostInfo.Disk = append(hostInfo.Disk, host.DiskInfo{
			DriveType:    disk.DriveType,
			SizeBytes:    disk.SizeBytes,
			Model:        disk.Model,
			SerialNumber: disk.SerialNumber,
		})
	}
	for _, gpu := range res.Gpu {
		hostInfo.Gpu = append(hostInfo.Gpu, host.GpuInfo{
			Vendor:  gpu.Vendor,
			Product: gpu.Product,
		})
	}
	return hostInfo
}
