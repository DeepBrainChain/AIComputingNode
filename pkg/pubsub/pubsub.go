package ps

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/hardware"
	"AIComputingNode/pkg/libp2p/host"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/model"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/serve"
	"AIComputingNode/pkg/timer"
	"AIComputingNode/pkg/types"

	"google.golang.org/protobuf/proto"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

var MsgNotSupported string = "Not supported"

type PubSub struct {
	topic       *pubsub.Topic
	sub         *pubsub.Subscription
	publishChan chan []byte
}

func NewPubSub(topic *pubsub.Topic, sub *pubsub.Subscription, pc chan []byte) *PubSub {
	return &PubSub{
		topic:       topic,
		sub:         sub,
		publishChan: pc,
	}
}

func (pst *PubSub) PublishToTopic(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Logger.Info("Publish to topic goroutine end")
			return
		case message := <-pst.publishChan:
			if err := pst.topic.Publish(ctx, message); err != nil {
				log.Logger.Errorf("Error when publish to topic %v", err)
			} else {
				log.Logger.Infof("Published %d bytes", len(message))
			}
		}
	}
}

func (pst *PubSub) ReadFromTopic(ctx context.Context) {
	defer pst.sub.Cancel()
	for {
		msg, err := pst.sub.Next(ctx)
		if err != nil {
			if errors.Is(err, ctx.Err()) {
				log.Logger.Info("Subscribe read goroutine end")
				return
			} else {
				log.Logger.Warnf("Read PubSub: %v", err)
				continue
			}
		}

		pmsg := &protocol.Message{}
		if err := proto.Unmarshal(msg.Data, pmsg); err != nil {
			log.Logger.Warnf("Unmarshal PubSub: %v", err)
			continue
		}

		if pmsg.Header.GetId() == "" && pmsg.Header.GetReceiver() == "" {
			log.Logger.Infof("Received scheduled broadcast message type %s from %s", pmsg.Type, pmsg.Header.GetNodeId())
			pst.handleScheduledBroadcastMessage(ctx, pmsg)
			continue
		} else if pmsg.Header.GetNodeId() == config.GC.Identity.PeerID {
			log.Logger.Infof("Received message type %s from the node itself", pmsg.Type)
			continue
		} else if pmsg.Header.GetReceiver() != config.GC.Identity.PeerID {
			log.Logger.Infof("Gossip message type %s from %s to %s", pmsg.Type, pmsg.Header.GetNodeId(), pmsg.Header.GetReceiver())
			continue
		} else {
			log.Logger.Infof("Received message type %s from %s with request_id %s", pmsg.Type, pmsg.Header.GetNodeId(), pmsg.Header.GetId())
		}

		go pst.handleBroadcastMessage(ctx, pmsg)
	}
}

func (pst *PubSub) handleScheduledBroadcastMessage(ctx context.Context, msg *protocol.Message) {
	switch msg.Type {
	case protocol.MessageType_AI_PROJECT:
		pst.handleScheduledAIProjectMessage(ctx, msg)
	default:
		log.Logger.Warnf("Unsupported scheduled broadcast message type", msg.Type)
	}
}

func (pst *PubSub) handleScheduledAIProjectMessage(ctx context.Context, msg *protocol.Message) {
	if !config.GC.App.PeersCollect.Enabled {
		log.Logger.Warnf("PeersCollect disabled when received %v message", msg.Type)
		return
	}
	aip := &protocol.AIProjectBody{}
	if err := proto.Unmarshal(msg.GetBody(), aip); err != nil {
		log.Logger.Warnf("Unmarshal AI Project Heartbeat %v", err)
		return
	}
	if aiRes := aip.GetRes(); aiRes != nil {
		info := db.PeerCollectInfo{
			Timestamp:  time.Now().Unix(),
			AIProjects: types.ProtocolMessage2AIProject(aiRes),
			NodeType:   aiRes.NodeType,
		}
		db.UpdatePeerCollect(msg.Header.GetNodeId(), info)
	} else {
		log.Logger.Warn("No ai project response found")
	}
}

func (pst *PubSub) handleBroadcastMessage(ctx context.Context, msg *protocol.Message) {
	existed := serve.ExistRequestItem(msg.Header.GetId())
	var code int
	var message string
	if msg.GetResultCode() == 0 {
		msgBody, err := host.Decrypt(msg.Header.GetNodePubKey(), msg.Body)
		if err != nil {
			code = int(types.ErrCodeDecrypt)
			message = types.ErrCodeDecrypt.String()
			log.Logger.Warnf("Decrypt %s message from %s failed %v", msg.Type.String(), msg.Header.GetNodeId(), err)
		} else {
			switch msg.Type {
			case protocol.MessageType_PEER_IDENTITY:
				code, message = pst.handlePeerIdentityMessage(ctx, msg, msgBody)
			case protocol.MessageType_HOST_INFO:
				code, message = pst.handleHostInfoMessage(ctx, msg, msgBody)
			case protocol.MessageType_AI_PROJECT:
				code, message = pst.handleAIProjectMessage(ctx, msg, msgBody)
			case protocol.MessageType_CHAT_COMPLETION:
				code, message = pst.handleChatCompletionMessage(ctx, msg, msgBody)
			case protocol.MessageType_IMAGE_GENERATION:
				code, message = pst.handleImageGenerationMessage(ctx, msg, msgBody)
			default:
				code = int(types.ErrCodeUnsupported)
				message = MsgNotSupported
				log.Logger.Warnf("Unknowned message type", msg.Type)
			}
			log.Logger.Infof("Handle %s message with request_id %s from %s result {code: %v, message: %v}",
				msg.Type.String(), msg.Header.GetId(), msg.Header.GetNodeId(), code, message)
		}

		if code == 0 {
			return
		}
		if existed {
			res := types.BaseHttpResponse{
				Code:    code,
				Message: message,
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal %s json %v", msg.Type.String(), err)
				return
			}
			serve.WriteAndDeleteRequestItem(msg.Header.GetId(), notifyData)
			log.Logger.Warnf("Send %s json response {code: %v, message: %v}", msg.Type.String(), code, message)
		} else {
			res := TransformErrorResponse(msg, int32(code), message)
			resBytes, err := proto.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal %s proto %v", msg.Type.String(), err)
				return
			}
			pst.publishChan <- resBytes
			log.Logger.Warnf("Send %s proto response {code: %v, message: %v}", msg.Type.String(), code, message)
		}
	} else {
		code = int(msg.GetResultCode())
		message = msg.GetResultMessage()

		if existed {
			res := types.BaseHttpResponse{
				Code:    code,
				Message: message,
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal %s json %v", msg.Type.String(), err)
				return
			}
			serve.WriteAndDeleteRequestItem(msg.Header.GetId(), notifyData)
			log.Logger.Warnf("Send %s json response {code: %v, message: %v}", msg.Type.String(), code, message)
		} else {
			log.Logger.Warnf("Unknown %s message with request_id %s and {result_code: %v, result_message: %v}, cannot be processed",
				msg.Type.String(), msg.Header.GetId(), code, message)
			// res := TransformErrorResponse(msg, int32(code), message)
			// resBytes, err := proto.Marshal(res)
			// if err != nil {
			// 	log.Logger.Errorf("Marshal %s proto %v", msg.Type.String(), err)
			// 	return
			// }
			// pst.publishChan <- resBytes
			// log.Logger.Warnf("Send %s proto response {code: %v, message: %v}", msg.Type.String(), code, message)
		}
	}
}

func (pst *PubSub) handlePeerIdentityMessage(ctx context.Context, msg *protocol.Message, decBody []byte) (int, string) {
	pi := &protocol.PeerIdentityBody{}
	if err := proto.Unmarshal(decBody, pi); err == nil {
		if piReq := pi.GetReq(); piReq != nil {
			idp := host.Hio.GetIdentifyProtocol()
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
				return int(types.ErrCodeProtobuf), types.ErrCodeProtobuf.String()
			}
			resBody, err = host.Encrypt(ctx, msg.Header.GetNodeId(), resBody)
			res := protocol.Message{
				Header: &protocol.MessageHeader{
					ClientVersion: host.Hio.UserAgent,
					Timestamp:     time.Now().Unix(),
					Id:            msg.Header.GetId(),
					NodeId:        config.GC.Identity.PeerID,
					Receiver:      msg.Header.GetNodeId(),
				},
				Type:       protocol.MessageType_PEER_IDENTITY,
				Body:       resBody,
				ResultCode: 0,
			}
			if err == nil {
				res.Header.NodePubKey, _ = host.MarshalPubKeyFromPrivKey(host.Hio.PrivKey)
			}
			resBytes, err := proto.Marshal(&res)
			if err != nil {
				log.Logger.Errorf("Marshal Identity Response %v", err)
				return int(types.ErrCodeProtobuf), types.ErrCodeProtobuf.String()
			}
			pst.publishChan <- resBytes
			log.Logger.Info("Sending Peer Identity Response")
			return 0, ""
		} else if piRes := pi.GetRes(); piRes != nil {
			res := types.PeerResponse{
				BaseHttpResponse: types.BaseHttpResponse{
					Code:    int(msg.ResultCode),
					Message: msg.ResultMessage,
				},
				IdentifyProtocol: types.IdentifyProtocol{
					ID:              msg.Header.GetNodeId(),
					ProtocolVersion: piRes.ProtocolVersion,
					AgentVersion:    piRes.AgentVersion,
					Addresses:       piRes.ListenAddrs,
					Protocols:       piRes.Protocols,
				},
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal Identity Protocol %v", err)
				return int(types.ErrCodeJson), types.ErrCodeJson.String()
			}
			serve.WriteAndDeleteRequestItem(msg.Header.GetId(), notifyData)
			return 0, ""
		} else {
			log.Logger.Error("No request or response found")
			return int(types.ErrCodeProtobuf), "No request or response found"
		}
	} else {
		log.Logger.Warn("Message type and body do not match")
		return int(types.ErrCodeProtobuf), "Message type and body do not match"
	}
}

func (pst *PubSub) handleChatCompletionMessage(ctx context.Context, msg *protocol.Message, decBody []byte) (int, string) {
	ccb := &protocol.ChatCompletionBody{}
	if err := proto.Unmarshal(decBody, ccb); err == nil {
		if chatReq := ccb.GetReq(); chatReq != nil {
			code, message, chatRes := pst.handleChatCompletionRequest(ctx, chatReq, msg.Header)
			igBody := &protocol.ChatCompletionBody{
				Data: &protocol.ChatCompletionBody_Res{
					Res: chatRes,
				},
			}
			resBody, err := proto.Marshal(igBody)
			if err != nil {
				log.Logger.Errorf("Marshal Chat Completion Response Body %v", err)
				return int(types.ErrCodeProtobuf), types.ErrCodeProtobuf.String()
			}
			resBody, err = host.Encrypt(ctx, msg.Header.GetNodeId(), resBody)
			res := protocol.Message{
				Header: &protocol.MessageHeader{
					ClientVersion: host.Hio.UserAgent,
					Timestamp:     chatRes.GetCreated(),
					Id:            msg.Header.GetId(),
					NodeId:        config.GC.Identity.PeerID,
					Receiver:      msg.Header.GetNodeId(),
				},
				Type:          protocol.MessageType_CHAT_COMPLETION,
				Body:          resBody,
				ResultCode:    int32(code),
				ResultMessage: message,
			}
			if err == nil {
				res.Header.NodePubKey, _ = host.MarshalPubKeyFromPrivKey(host.Hio.PrivKey)
			}
			resBytes, err := proto.Marshal(&res)
			if err != nil {
				log.Logger.Errorf("Marshal Chat Completion Response %v", err)
				return int(types.ErrCodeProtobuf), types.ErrCodeProtobuf.String()
			}
			pst.publishChan <- resBytes
			log.Logger.Info("Sending Chat Completion Response")
			return 0, ""
		} else if chatRes := ccb.GetRes(); chatRes != nil {
			res := types.ChatCompletionResponse{
				BaseHttpResponse: types.BaseHttpResponse{
					Code:    int(msg.ResultCode),
					Message: msg.ResultMessage,
				},
			}
			if msg.ResultCode == 0 {
				res.Created = chatRes.GetCreated()
				for _, choice := range chatRes.Choices {
					ccmsg := types.ChatCompletionResponseMessage{
						Role:    choice.GetMessage().GetRole(),
						Content: choice.GetMessage().GetContent(),
					}
					res.Choices = append(res.Choices, types.ChatResponseChoice{
						Index:        int(choice.GetIndex()),
						Message:      ccmsg,
						FinishReason: choice.GetFinishReason(),
					})
				}
				res.Usage = types.ChatResponseUsage{
					CompletionTokens: int(chatRes.GetUsage().GetCompletionTokens()),
					PromptTokens:     int(chatRes.GetUsage().GetPromptTokens()),
					TotalTokens:      int(chatRes.GetUsage().GetTotalTokens()),
				}
				res.Id = chatRes.GetId()
				res.Object = chatRes.GetObject()
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal Chat Completion Response %v", err)
				return int(types.ErrCodeJson), types.ErrCodeJson.String()
			}
			serve.WriteAndDeleteRequestItem(msg.Header.GetId(), notifyData)
			return 0, ""
		} else {
			log.Logger.Error("No request or response found")
			return int(types.ErrCodeProtobuf), "No request or response found"
		}
	} else {
		log.Logger.Warn("Message type and body do not match")
		return int(types.ErrCodeProtobuf), "Message type and body do not match"
	}
}

func (pst *PubSub) handleImageGenerationMessage(ctx context.Context, msg *protocol.Message, decBody []byte) (int, string) {
	ig := &protocol.ImageGenerationBody{}
	if err := proto.Unmarshal(decBody, ig); err == nil {
		if igReq := ig.GetReq(); igReq != nil {
			code, message, igRes := pst.handleImageGenerationRequest(ctx, igReq, msg.Header)
			igBody := &protocol.ImageGenerationBody{
				Data: &protocol.ImageGenerationBody_Res{
					Res: igRes,
				},
			}
			resBody, err := proto.Marshal(igBody)
			if err != nil {
				log.Logger.Errorf("Marshal Image Generation Response Body %v", err)
				return int(types.ErrCodeProtobuf), types.ErrCodeProtobuf.String()
			}
			resBody, err = host.Encrypt(ctx, msg.Header.GetNodeId(), resBody)
			res := protocol.Message{
				Header: &protocol.MessageHeader{
					ClientVersion: host.Hio.UserAgent,
					Timestamp:     igRes.GetCreated(),
					Id:            msg.Header.GetId(),
					NodeId:        config.GC.Identity.PeerID,
					Receiver:      msg.Header.GetNodeId(),
				},
				Type:          protocol.MessageType_IMAGE_GENERATION,
				Body:          resBody,
				ResultCode:    int32(code),
				ResultMessage: message,
			}
			if err == nil {
				res.Header.NodePubKey, _ = host.MarshalPubKeyFromPrivKey(host.Hio.PrivKey)
			}
			resBytes, err := proto.Marshal(&res)
			if err != nil {
				log.Logger.Errorf("Marshal Image Generation Response %v", err)
				return int(types.ErrCodeProtobuf), types.ErrCodeProtobuf.String()
			}
			pst.publishChan <- resBytes
			log.Logger.Info("Sending Image Generation Response")
			return 0, ""
		} else if igRes := ig.GetRes(); igRes != nil {
			res := types.ImageGenerationResponse{
				BaseHttpResponse: types.BaseHttpResponse{
					Code:    int(msg.ResultCode),
					Message: msg.ResultMessage,
				},
			}
			if msg.ResultCode == 0 {
				res.Created = igRes.GetCreated()
				for _, choice := range igRes.GetChoices() {
					res.Choices = append(res.Choices, types.ImageResponseChoice{
						Url:           choice.GetUrl(),
						B64Json:       choice.GetB64Json(),
						RevisedPrompt: choice.GetRevisedPrompt(),
					})
				}
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal Image Generation Response %v", err)
				return int(types.ErrCodeJson), types.ErrCodeJson.String()
			}
			serve.WriteAndDeleteRequestItem(msg.Header.GetId(), notifyData)
			return 0, ""
		} else {
			log.Logger.Error("No request or response found")
			return int(types.ErrCodeProtobuf), "No request or response found"
		}
	} else {
		log.Logger.Warn("Message type and body do not match")
		return int(types.ErrCodeProtobuf), "Message type and body do not match"
	}
}

func (pst *PubSub) handleHostInfoMessage(ctx context.Context, msg *protocol.Message, decBody []byte) (int, string) {
	hi := &protocol.HostInfoBody{}
	if err := proto.Unmarshal(decBody, hi); err == nil {
		if hiReq := hi.GetReq(); hiReq != nil {
			hostInfo, err := hardware.GetHostInfo()
			var code int32 = 0
			var message string = ""
			if err != nil {
				code = int32(types.ErrCodeHostInfo)
				message = err.Error()
			}
			hiRes := types.HostInfo2ProtocolMessage(hostInfo)
			hiBody := &protocol.HostInfoBody{
				Data: &protocol.HostInfoBody_Res{
					Res: hiRes,
				},
			}
			resBody, err := proto.Marshal(hiBody)
			if err != nil {
				log.Logger.Warnf("Marshal HostInfo Response Body %v", err)
				return int(types.ErrCodeProtobuf), types.ErrCodeProtobuf.String()
			}
			resBody, err = host.Encrypt(ctx, msg.Header.GetNodeId(), resBody)
			res := protocol.Message{
				Header: &protocol.MessageHeader{
					ClientVersion: host.Hio.UserAgent,
					Timestamp:     time.Now().Unix(),
					Id:            msg.Header.GetId(),
					NodeId:        config.GC.Identity.PeerID,
					Receiver:      msg.Header.GetNodeId(),
				},
				Type:          protocol.MessageType_HOST_INFO,
				Body:          resBody,
				ResultCode:    code,
				ResultMessage: message,
			}
			if err == nil {
				res.Header.NodePubKey, _ = host.MarshalPubKeyFromPrivKey(host.Hio.PrivKey)
			}
			resBytes, err := proto.Marshal(&res)
			if err != nil {
				log.Logger.Errorf("Marshal HostInfo Response %v", err)
				return int(types.ErrCodeProtobuf), types.ErrCodeProtobuf.String()
			}
			pst.publishChan <- resBytes
			log.Logger.Info("Sending HostInfo Response")
			return 0, ""
		} else if hiRes := hi.GetRes(); hiRes != nil {
			res := types.HostInfoResponse{
				BaseHttpResponse: types.BaseHttpResponse{
					Code:    int(msg.ResultCode),
					Message: msg.ResultMessage,
				},
				HostInfo: *types.ProtocolMessage2HostInfo(hiRes),
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal HostInfo Protocol %v", err)
				return int(types.ErrCodeJson), types.ErrCodeJson.String()
			}
			serve.WriteAndDeleteRequestItem(msg.Header.GetId(), notifyData)
			return 0, ""
		} else {
			log.Logger.Error("No request or response found")
			return int(types.ErrCodeProtobuf), "No request or response found"
		}
	} else {
		log.Logger.Warn("Message type and body do not match")
		return int(types.ErrCodeProtobuf), "Message type and body do not match"
	}
}

func (pst *PubSub) handleAIProjectMessage(ctx context.Context, msg *protocol.Message, decBody []byte) (int, string) {
	aip := &protocol.AIProjectBody{}
	if err := proto.Unmarshal(decBody, aip); err == nil {
		if aiReq := aip.GetReq(); aiReq != nil {
			projects := model.GetAIProjects()
			aiBody := &protocol.AIProjectBody{
				Data: &protocol.AIProjectBody_Res{
					Res: types.AIProject2ProtocolMessage(projects, 0),
				},
			}
			resBody, err := proto.Marshal(aiBody)
			if err != nil {
				log.Logger.Warnf("Marshal AI Project Response Body %v", err)
				return int(types.ErrCodeProtobuf), types.ErrCodeProtobuf.String()
			}
			resBody, err = host.Encrypt(ctx, msg.Header.GetNodeId(), resBody)
			res := protocol.Message{
				Header: &protocol.MessageHeader{
					ClientVersion: host.Hio.UserAgent,
					Timestamp:     time.Now().Unix(),
					Id:            msg.Header.GetId(),
					NodeId:        config.GC.Identity.PeerID,
					Receiver:      msg.Header.GetNodeId(),
				},
				Type:          protocol.MessageType_AI_PROJECT,
				Body:          resBody,
				ResultCode:    0,
				ResultMessage: "",
			}
			if err == nil {
				res.Header.NodePubKey, _ = host.MarshalPubKeyFromPrivKey(host.Hio.PrivKey)
			}
			resBytes, err := proto.Marshal(&res)
			if err != nil {
				log.Logger.Errorf("Marshal AI Project Response %v", err)
				return int(types.ErrCodeProtobuf), types.ErrCodeProtobuf.String()
			}
			pst.publishChan <- resBytes
			log.Logger.Info("Sending AI Project Response")
			return 0, ""
		} else if aiRes := aip.GetRes(); aiRes != nil {
			res := types.AIProjectListResponse{
				BaseHttpResponse: types.BaseHttpResponse{
					Code:    int(msg.ResultCode),
					Message: msg.ResultMessage,
				},
				Data: types.ProtocolMessage2AIProject(aiRes),
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal AI Project Protocol %v", err)
				return int(types.ErrCodeJson), types.ErrCodeJson.String()
			}
			serve.WriteAndDeleteRequestItem(msg.Header.GetId(), notifyData)
			return 0, ""
		} else {
			log.Logger.Error("No request or response found")
			return int(types.ErrCodeProtobuf), "No request or response found"
		}
	} else {
		log.Logger.Warn("Message type and body do not match")
		return int(types.ErrCodeProtobuf), "Message type and body do not match"
	}
}

func TransformErrorResponse(msg *protocol.Message, code int32, message string) *protocol.Message {
	res := protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: host.Hio.UserAgent,
			Timestamp:     time.Now().Unix(),
			Id:            msg.Header.GetId(),
			NodeId:        config.GC.Identity.PeerID,
			Receiver:      msg.Header.GetNodeId(),
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

func (pst *PubSub) handleChatCompletionRequest(ctx context.Context, req *protocol.ChatCompletionRequest, reqHeader *protocol.MessageHeader) (int, string, *protocol.ChatCompletionResponse) {
	response := &protocol.ChatCompletionResponse{}

	mi, err := model.GetModelInfo(req.GetProject(), req.GetModel(), req.GetCid())
	if err != nil {
		return int(types.ErrCodeModel), err.Error(), response
	}

	chatReq := types.ChatModelRequest{
		Model:  req.GetModel(),
		Stream: req.GetStream(),
		WalletVerification: types.WalletVerification{
			Wallet:    req.GetWallet().GetWallet(),
			Signature: req.GetWallet().GetSignature(),
			Hash:      req.GetWallet().GetHash(),
		},
		// temperature/top_p is 0 when field is omitempty
		// Temperature: req.GetTemperature(),
		// TopP:        req.GetTopP(),
	}
	for _, ccm := range req.Messages {
		chatReq.Messages = append(chatReq.Messages, types.ChatCompletionMessage{
			Role:    ccm.GetRole(),
			Content: ccm.GetContent(),
		})
	}

	model.IncRef(req.GetProject(), req.GetModel(), mi.CID)
	timer.SendAIProjects(pst.publishChan)
	defer func() {
		model.DecRef(req.GetProject(), req.GetModel(), mi.CID)
		timer.SendAIProjects(pst.publishChan)
	}()
	chatRes := model.ChatModel(mi.API, chatReq)

	log.Logger.Infof("Execute model %s in %s result {code:%d, message:%s}", req.GetProject(), req.GetModel(), chatRes.Code, chatRes.Message)
	modelHistory := &types.ModelHistory{
		TimeStamp:    chatRes.Created,
		ReqId:        reqHeader.GetId(),
		ReqNodeId:    reqHeader.GetNodeId(),
		ResNodeId:    reqHeader.GetReceiver(),
		Code:         chatRes.Code,
		Message:      chatRes.Message,
		Project:      req.GetProject(),
		Model:        req.GetModel(),
		ChatMessages: chatReq.Messages,
		ChatChoices:  chatRes.Choices,
		ChatUsage:    chatRes.Usage,
		ImagePrompt:  "",
		ImageChoices: []types.ImageResponseChoice{},
	}
	_ = db.WriteModelHistory(modelHistory)

	if chatRes.Code != 0 {
		return chatRes.Code, chatRes.Message, response
	}
	response.Created = chatRes.Created
	for _, choice := range chatRes.Choices {
		response.Choices = append(response.Choices, &protocol.ChatCompletionResponse_ChatResponseChoice{
			Index: int32(choice.Index),
			Message: &protocol.ChatCompletionResponseMessage{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: choice.FinishReason,
		})
	}
	response.Usage = &protocol.ChatCompletionResponse_ChatResponseUsage{
		CompletionTokens: int32(chatRes.Usage.CompletionTokens),
		PromptTokens:     int32(chatRes.Usage.PromptTokens),
		TotalTokens:      int32(chatRes.Usage.TotalTokens),
	}
	return chatRes.Code, chatRes.Message, response
}

func (pst *PubSub) handleImageGenerationRequest(ctx context.Context, req *protocol.ImageGenerationRequest, reqHeader *protocol.MessageHeader) (int, string, *protocol.ImageGenerationResponse) {
	response := &protocol.ImageGenerationResponse{}

	mi, err := model.GetModelInfo(req.GetProject(), req.GetModel(), req.GetCid())
	if err != nil {
		return int(types.ErrCodeModel), err.Error(), response
	}

	igReq := types.ImageGenModelRequest{
		File:           req.GetFile(),
		Model:          req.GetModel(),
		Prompt:         req.GetPrompt(),
		Number:         int(req.GetNumber()),
		Size:           req.GetSize(),
		Width:          int(req.GetWidth()),
		Height:         int(req.GetHeight()),
		ResponseFormat: req.GetResponseFormat(),
		WalletVerification: types.WalletVerification{
			Wallet:    req.GetWallet().GetWallet(),
			Signature: req.GetWallet().GetSignature(),
			Hash:      req.GetWallet().GetHash(),
		},
		// step is 0 when field is omitempty
		// Step: req.GetStep(),
	}

	model.IncRef(req.GetProject(), req.GetModel(), mi.CID)
	timer.SendAIProjects(pst.publishChan)
	defer func() {
		model.DecRef(req.GetProject(), req.GetModel(), mi.CID)
		timer.SendAIProjects(pst.publishChan)
	}()
	igRes := model.ImageGenerationModel(mi.API, igReq)

	if igRes.Code == 0 {
		log.Logger.Infof("Execute model %s with (%q, %d, %s) result %v",
			req.GetModel(), req.GetPrompt(), req.GetNumber(), req.GetSize(), igRes.Choices)
	} else {
		log.Logger.Errorf("Execute model %s with (%q, %d, %s) error {code:%d, message:%s}",
			req.GetModel(), req.GetPrompt(), req.GetNumber(), req.GetSize(), igRes.Code, igRes.Message)
	}
	modelHistory := &types.ModelHistory{
		TimeStamp:    igRes.Created,
		ReqId:        reqHeader.GetId(),
		ReqNodeId:    reqHeader.GetNodeId(),
		ResNodeId:    reqHeader.GetReceiver(),
		Code:         igRes.Code,
		Message:      igRes.Message,
		Project:      req.GetProject(),
		Model:        req.GetModel(),
		ChatMessages: []types.ChatCompletionMessage{},
		ChatChoices:  []types.ChatResponseChoice{},
		ChatUsage:    types.ChatResponseUsage{},
		ImagePrompt:  req.GetPrompt(),
		ImageChoices: igRes.Choices,
	}
	_ = db.WriteModelHistory(modelHistory)

	if igRes.Code != 0 {
		return igRes.Code, igRes.Message, response
	}
	response.Created = igRes.Created
	for _, choice := range igRes.Choices {
		response.Choices = append(response.Choices, &protocol.ImageGenerationResponse_ImageResponseChoice{
			Url:           choice.Url,
			B64Json:       choice.B64Json,
			RevisedPrompt: choice.RevisedPrompt,
		})
	}
	return igRes.Code, igRes.Message, response
}
