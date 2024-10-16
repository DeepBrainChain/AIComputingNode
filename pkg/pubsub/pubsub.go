package ps

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/host"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/model"
	"AIComputingNode/pkg/p2p"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/serve"
	"AIComputingNode/pkg/timer"
	"AIComputingNode/pkg/types"

	"google.golang.org/protobuf/proto"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

var MsgNotSupported string = "Not supported"

func PublishToTopic(ctx context.Context, topic *pubsub.Topic, messageChan <-chan []byte) {
	for {
		select {
		case <-ctx.Done():
			log.Logger.Info("Publish to topic goroutine end")
			return
		case message := <-messageChan:
			if err := topic.Publish(ctx, message); err != nil {
				log.Logger.Errorf("Error when publish to topic %v", err)
			} else {
				log.Logger.Infof("Published %d bytes", len(message))
			}
		}
	}
}

func ReadFromTopic(ctx context.Context, sub *pubsub.Subscription, publishChan chan<- []byte) {
	defer sub.Cancel()
	for {
		msg, err := sub.Next(ctx)
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
			timer.AIT.HandleBroadcastMessage(ctx, pmsg)
			log.Logger.Infof("Received heartbeat message type %s from %s", pmsg.Type, pmsg.Header.NodeId)
			continue
		} else if pmsg.Header.NodeId == config.GC.Identity.PeerID {
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
	case protocol.MessageType_AI_PROJECT:
		handleAIProjectMessage(ctx, msg, msgBody, publishChan)
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
		res := types.PeerResponse{
			BaseHttpResponse: types.BaseHttpResponse{
				Code:    int(msg.ResultCode),
				Message: msg.ResultMessage,
			},
		}
		notifyData, err := json.Marshal(res)
		if err != nil {
			log.Logger.Errorf("Marshal Identity Protocol %v", err)
			return
		}
		serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
	} else if err := proto.Unmarshal(decBody, pi); err == nil {
		if piReq := pi.GetReq(); piReq != nil {
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
			resBody, err = p2p.Encrypt(ctx, msg.Header.NodeId, resBody)
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
		} else if piRes := pi.GetRes(); piRes != nil {
			res := types.PeerResponse{
				BaseHttpResponse: types.BaseHttpResponse{
					Code:    int(msg.ResultCode),
					Message: msg.ResultMessage,
				},
				IdentifyProtocol: types.IdentifyProtocol{
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
		res := types.ChatCompletionResponse{
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
			code, message, chatRes := handleChatCompletionRequest(ctx, chatReq, msg.Header)
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
			resBody, err = p2p.Encrypt(ctx, msg.Header.NodeId, resBody)
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
		} else if chatRes := ccb.GetRes(); chatRes != nil {
			res := types.ChatCompletionResponse{
				Code:    int(msg.ResultCode),
				Message: msg.ResultMessage,
			}
			if msg.ResultCode == 0 {
				res.Data.Created = chatRes.Created
				for _, choice := range chatRes.Choices {
					ccmsg := types.ChatCompletionMessage{
						Role:    choice.GetMessage().GetRole(),
						Content: choice.GetMessage().GetContent(),
					}
					res.Data.Choices = append(res.Data.Choices, types.ChatResponseChoice{
						Index:        int(choice.GetIndex()),
						Message:      ccmsg,
						FinishReason: choice.GetFinishReason(),
					})
				}
				res.Data.Usage = types.ChatResponseUsage{
					CompletionTokens: int(chatRes.Usage.CompletionTokens),
					PromptTokens:     int(chatRes.Usage.PromptTokens),
					TotalTokens:      int(chatRes.Usage.TotalTokens),
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
		res := types.ImageGenerationResponse{
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
			code, message, igRes := handleImageGenerationRequest(ctx, igReq, msg.Header)
			igBody := &protocol.ImageGenerationBody{
				Data: &protocol.ImageGenerationBody_Res{
					Res: igRes,
				},
			}
			resBody, err := proto.Marshal(igBody)
			if err != nil {
				log.Logger.Errorf("Marshal Image Generation Response Body %v", err)
				return
			}
			resBody, err = p2p.Encrypt(ctx, msg.Header.NodeId, resBody)
			res := protocol.Message{
				Header: &protocol.MessageHeader{
					ClientVersion: p2p.Hio.UserAgent,
					Timestamp:     igRes.Created,
					Id:            msg.Header.Id,
					NodeId:        config.GC.Identity.PeerID,
					Receiver:      msg.Header.NodeId,
				},
				Type:          protocol.MessageType_IMAGE_GENERATION,
				Body:          resBody,
				ResultCode:    int32(code),
				ResultMessage: message,
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
		} else if igRes := ig.GetRes(); igRes != nil {
			res := types.ImageGenerationResponse{
				Code:    int(msg.ResultCode),
				Message: msg.ResultMessage,
			}
			if msg.ResultCode == 0 {
				res.Data.Created = igRes.Created
				for _, choice := range igRes.GetChoices() {
					res.Data.Choices = append(res.Data.Choices, types.ImageResponseChoice{
						Url:           choice.Url,
						B64Json:       choice.B64Json,
						RevisedPrompt: choice.RevisedPrompt,
					})
				}
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
		res := types.HostInfoResponse{
			BaseHttpResponse: types.BaseHttpResponse{
				Code:    int(msg.ResultCode),
				Message: msg.ResultMessage,
			},
		}
		notifyData, err := json.Marshal(res)
		if err != nil {
			log.Logger.Errorf("Marshal Host Info Protocol %v", err)
			return
		}
		serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
	} else if err := proto.Unmarshal(decBody, hi); err == nil {
		if hiReq := hi.GetReq(); hiReq != nil {
			hostInfo, err := host.GetHostInfo()
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
				return
			}
			resBody, err = p2p.Encrypt(ctx, msg.Header.NodeId, resBody)
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
				return
			}
			serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
		}
	} else {
		log.Logger.Warn("Message type and body do not match")
	}
}

func handleAIProjectMessage(ctx context.Context, msg *protocol.Message, decBody []byte, publishChan chan<- []byte) {
	aip := &protocol.AIProjectBody{}
	if msg.ResultCode != 0 {
		res := types.AIProjectListResponse{
			Code:    int(msg.ResultCode),
			Message: msg.ResultMessage,
		}
		notifyData, err := json.Marshal(res)
		if err != nil {
			log.Logger.Errorf("Marshal AI Project Protocol %v", err)
			return
		}
		serve.WriteAndDeleteRequestItem(msg.Header.Id, notifyData)
	} else if err := proto.Unmarshal(decBody, aip); err == nil {
		if aiReq := aip.GetReq(); aiReq != nil {
			projects := config.GC.GetAIProjectsOfNode()
			aiBody := &protocol.AIProjectBody{
				Data: &protocol.AIProjectBody_Res{
					Res: types.AIProject2ProtocolMessage(projects, 0),
				},
			}
			resBody, err := proto.Marshal(aiBody)
			if err != nil {
				log.Logger.Warnf("Marshal AI Project Response Body %v", err)
				return
			}
			resBody, err = p2p.Encrypt(ctx, msg.Header.NodeId, resBody)
			res := protocol.Message{
				Header: &protocol.MessageHeader{
					ClientVersion: p2p.Hio.UserAgent,
					Timestamp:     time.Now().Unix(),
					Id:            msg.Header.Id,
					NodeId:        config.GC.Identity.PeerID,
					Receiver:      msg.Header.NodeId,
				},
				Type:          protocol.MessageType_AI_PROJECT,
				Body:          resBody,
				ResultCode:    0,
				ResultMessage: "ok",
			}
			if err == nil {
				res.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
			}
			resBytes, err := proto.Marshal(&res)
			if err != nil {
				log.Logger.Errorf("Marshal AI Project Response %v", err)
				return
			}
			publishChan <- resBytes
			log.Logger.Info("Sending AI Project Response")
		} else if aiRes := aip.GetRes(); aiRes != nil {
			res := types.AIProjectListResponse{
				Code:    int(msg.ResultCode),
				Message: msg.ResultMessage,
				Data:    types.ProtocolMessage2AIProject(aiRes),
			}
			notifyData, err := json.Marshal(res)
			if err != nil {
				log.Logger.Errorf("Marshal AI Project Protocol %v", err)
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

func handleChatCompletionRequest(ctx context.Context, req *protocol.ChatCompletionRequest, reqHeader *protocol.MessageHeader) (int, string, *protocol.ChatCompletionResponse) {
	chatReq := types.ChatModelRequest{
		Model:  req.Model,
		Stream: req.Stream,
		WalletVerification: types.WalletVerification{
			Wallet:    req.Wallet.Wallet,
			Signature: req.Wallet.Signature,
			Hash:      req.Wallet.Hash,
		},
	}
	for _, ccm := range req.Messages {
		chatReq.Messages = append(chatReq.Messages, types.ChatCompletionMessage{
			Role:    ccm.Role,
			Content: ccm.Content,
		})
	}
	chatRes := model.ChatModel(config.GC.GetModelAPI(req.Project, req.Model), chatReq)

	log.Logger.Infof("Execute model %s in %s result {code:%d, message:%s}", req.Project, req.Model, chatRes.Code, chatRes.Message)
	modelHistory := &types.ModelHistory{
		TimeStamp:    chatRes.Data.Created,
		ReqId:        reqHeader.Id,
		ReqNodeId:    reqHeader.NodeId,
		ResNodeId:    reqHeader.Receiver,
		Code:         chatRes.Code,
		Message:      chatRes.Message,
		Project:      req.GetProject(),
		Model:        req.GetModel(),
		ChatMessages: chatReq.Messages,
		ChatChoices:  chatRes.Data.Choices,
		ChatUsage:    chatRes.Data.Usage,
		ImagePrompt:  "",
		ImageChoices: []types.ImageResponseChoice{},
	}
	_ = db.WriteModelHistory(modelHistory)

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
	response.Usage = &protocol.ChatCompletionResponse_ChatResponseUsage{
		CompletionTokens: int32(chatRes.Data.Usage.CompletionTokens),
		PromptTokens:     int32(chatRes.Data.Usage.PromptTokens),
		TotalTokens:      int32(chatRes.Data.Usage.TotalTokens),
	}
	return chatRes.Code, chatRes.Message, response
}

func handleImageGenerationRequest(ctx context.Context, req *protocol.ImageGenerationRequest, reqHeader *protocol.MessageHeader) (int, string, *protocol.ImageGenerationResponse) {
	igReq := types.ImageGenModelRequest{
		Model:          req.GetModel(),
		Prompt:         req.GetPrompt(),
		Number:         int(req.GetNumber()),
		Size:           req.GetSize(),
		Width:          int(req.GetWidth()),
		Height:         int(req.GetHeight()),
		ResponseFormat: req.GetResponseFormat(),
		WalletVerification: types.WalletVerification{
			Wallet:    req.Wallet.Wallet,
			Signature: req.Wallet.Signature,
			Hash:      req.Wallet.Hash,
		},
	}
	igRes := model.ImageGenerationModel(config.GC.GetModelAPI(req.GetProject(), req.GetModel()), igReq)

	if igRes.Code == 0 {
		log.Logger.Infof("Execute model %s with (%q, %d, %s) result %v",
			req.GetModel(), req.GetPrompt(), req.GetNumber(), req.GetSize(), igRes.Data.Choices)
	} else {
		log.Logger.Errorf("Execute model %s with (%q, %d, %s) error {code:%d, message:%s}",
			req.GetModel(), req.GetPrompt(), req.GetNumber(), req.GetSize(), igRes.Code, igRes.Message)
	}
	modelHistory := &types.ModelHistory{
		TimeStamp:    igRes.Data.Created,
		ReqId:        reqHeader.Id,
		ReqNodeId:    reqHeader.NodeId,
		ResNodeId:    reqHeader.Receiver,
		Code:         igRes.Code,
		Message:      igRes.Message,
		Project:      req.GetProject(),
		Model:        req.GetModel(),
		ChatMessages: []types.ChatCompletionMessage{},
		ChatChoices:  []types.ChatResponseChoice{},
		ChatUsage:    types.ChatResponseUsage{},
		ImagePrompt:  req.GetPrompt(),
		ImageChoices: igRes.Data.Choices,
	}
	_ = db.WriteModelHistory(modelHistory)

	response := &protocol.ImageGenerationResponse{}
	if igRes.Code != 0 {
		return igRes.Code, igRes.Message, response
	}
	response.Created = igRes.Data.Created
	for _, choice := range igRes.Data.Choices {
		response.Choices = append(response.Choices, &protocol.ImageGenerationResponse_ImageResponseChoice{
			Url:           choice.Url,
			B64Json:       choice.B64Json,
			RevisedPrompt: choice.RevisedPrompt,
		})
	}
	return igRes.Code, igRes.Message, response
}
