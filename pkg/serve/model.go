package serve

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sort"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/libp2p/host"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/model"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/timer"
	"AIComputingNode/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func handleChatCompletionRequest(ctx context.Context, publishChan chan<- []byte, req *types.ChatCompletionRequest, rsp *types.ChatCompletionResponse) (int, int, string) {
	if req.NodeID == config.GC.Identity.PeerID {
		mi, err := model.GetModelInfo(req.Project, req.Model, req.CID)
		if err != nil {
			return http.StatusInternalServerError, int(types.ErrCodeModel), err.Error()
		}
		model.IncRef(req.Project, req.Model, mi.CID)
		timer.SendAIProjects(publishChan)
		defer func() {
			model.DecRef(req.Project, req.Model, mi.CID)
			timer.SendAIProjects(publishChan)
		}()
		*rsp = *model.ChatModel(mi.API, req.ChatModelRequest)
		log.Logger.Infof("Execute model %s result {code:%d, message:%s}", req.Model, rsp.Code, rsp.Message)
		return http.StatusOK, rsp.Code, rsp.Message
	}

	requestID, err := uuid.NewRandom()
	if err != nil {
		return http.StatusInternalServerError, int(types.ErrCodeUUID), err.Error()
	}

	ccms := make([]*protocol.ChatCompletionMessage, 0)
	for _, ccm := range req.Messages {
		ccms = append(ccms, &protocol.ChatCompletionMessage{
			Role:    ccm.Role,
			Content: ccm.Content,
		})
	}

	pi := &protocol.ChatCompletionBody{
		Data: &protocol.ChatCompletionBody_Req{
			Req: &protocol.ChatCompletionRequest{
				Project:  req.Project,
				Model:    req.Model,
				Messages: ccms,
				Stream:   false,
				Wallet: &protocol.WalletVerification{
					Wallet:    req.Wallet,
					Signature: req.Signature,
					Hash:      req.Hash,
				},
				Cid:         req.CID,
				Temperature: req.Temperature,
				TopP:        req.TopP,
			},
		},
	}
	body, err := proto.Marshal(pi)
	if err != nil {
		return http.StatusInternalServerError, int(types.ErrCodeProtobuf), err.Error()
	}
	body, err = host.Encrypt(ctx, req.NodeID, body)
	if err != nil {
		return http.StatusInternalServerError, int(types.ErrCodeEncrypt), types.ErrCodeEncrypt.String()
	}

	msg := &protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: host.Hio.UserAgent,
			Timestamp:     time.Now().Unix(),
			Id:            requestID.String(),
			NodeId:        config.GC.Identity.PeerID,
			Receiver:      req.NodeID,
			NodePubKey:    nil,
			Sign:          nil,
		},
		Type:       *protocol.MessageType_CHAT_COMPLETION.Enum(),
		Body:       body,
		ResultCode: 0,
	}
	msg.Header.NodePubKey, _ = host.MarshalPubKeyFromPrivKey(host.Hio.PrivKey)
	return handleRequest(publishChan, msg, rsp, types.ChatCompletionRequestTimeout)
}

func handleChatCompletionStreamRequest(ctx context.Context, w http.ResponseWriter, req *types.ChatCompletionRequest, rsp *types.ChatCompletionResponse) (int, int, string) {
	if req.NodeID == config.GC.Identity.PeerID {
		mi, err := model.GetModelInfo(req.Project, req.Model, req.CID)
		log.Logger.Info("Received chat completion stream request from the node itself")
		if err != nil {
			return http.StatusInternalServerError, int(types.ErrCodeModel), err.Error()
		}

		jsonData, err := json.Marshal(req.ChatModelRequest)
		if err != nil {
			// rsp.Code = int(types.ErrCodeModel)
			// rsp.Message = "Marshal model request body failed"
			log.Logger.Errorf("Marshal model request body failed: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeModel), "Marshal model request body failed"
		}
		hreq, err := http.NewRequestWithContext(ctx, "POST", mi.API, bytes.NewBuffer(jsonData))
		if err != nil {
			// rsp.Code = int(types.ErrCodeModel)
			// rsp.Message = "Create http request for stream failed"
			log.Logger.Errorf("Create http request for stream failed: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeModel), "Create http request for stream failed"
		}
		hreq.Header.Set("Content-Type", "application/json")

		log.Logger.Infof("Making request to %s", hreq.URL)
		resp, err := http.DefaultTransport.RoundTrip(hreq)
		if err != nil {
			// rsp.Code = int(types.ErrCodeModel)
			// rsp.Message = fmt.Sprintf("RoundTrip chat request failed: %v", err)
			log.Logger.Errorf("RoundTrip chat request failed: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeModel), fmt.Sprintf("RoundTrip chat request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.Header.Get("Content-Type") == "application/json" {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				// rsp.Code = int(types.ErrCodeModel)
				// rsp.Message = "Read model response json error"
				log.Logger.Errorf("Read model response json error: %v", err)
				return http.StatusInternalServerError, int(types.ErrCodeModel), "Read model response json error"
			}
			if err := json.Unmarshal(body, rsp); err != nil {
				// rsp.Code = int(types.ErrCodeModel)
				// rsp.Message = "Unmarshal model response json error"
				log.Logger.Errorf("Unmarshal model response json error: %v", err)
				return http.StatusInternalServerError, int(types.ErrCodeModel), "Unmarshal model response json error"
			}
			return http.StatusOK, rsp.Code, rsp.Message
		}

		for k, v := range resp.Header {
			for _, s := range v {
				w.Header().Set(k, s)
			}
		}

		w.WriteHeader(resp.StatusCode)

		log.Logger.Info("Copy roundtrip chat completion response")
		io.Copy(w, resp.Body)
		// resp.Body.Close()
		log.Logger.Info("Handle chat completion stream request over from the node itself")
		// rsp.Code = 0
		// rsp.Message = ""
		return http.StatusOK, 0, ""
	}

	log.Logger.Info("Received chat completion stream request")

	jsonData, err := json.Marshal(req.ChatModelRequest)
	if err != nil {
		// rsp.Code = int(types.ErrCodeStream)
		// rsp.Message = "Marshal model request body failed"
		log.Logger.Errorf("Marshal model request body failed: %v", err)
		return http.StatusInternalServerError, int(types.ErrCodeStream), "Marshal model request body failed"
	}
	hreq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:8080/api/v0/chat/completion", bytes.NewBuffer(jsonData))
	if err != nil {
		// rsp.Code = int(types.ErrCodeStream)
		// rsp.Message = "Create http request for stream failed"
		log.Logger.Errorf("Create http request for stream failed: %v", err)
		return http.StatusInternalServerError, int(types.ErrCodeStream), "Create http request for stream failed"
	}

	hreq.Header.Set("Content-Type", "application/json")

	queryValues := hreq.URL.Query()
	queryValues.Add("project", req.Project)
	queryValues.Add("model", req.Model)
	queryValues.Add("cid", req.CID)
	hreq.URL.RawQuery = queryValues.Encode()

	stream, err := host.Hio.NewStream(ctx, req.NodeID)
	if err != nil {
		// rsp.Code = int(types.ErrCodeStream)
		// rsp.Message = "Open stream with peer node failed"
		log.Logger.Errorf("Open stream with peer node failed: %v", err)
		return http.StatusInternalServerError, int(types.ErrCodeStream), "Open stream with peer node failed"
	}
	stream.SetDeadline(time.Now().Add(types.ChatCompletionRequestTimeout))
	defer stream.Close()
	log.Logger.Infof("Create libp2p stream with %s success", req.NodeID)

	err = hreq.Write(stream)
	if err != nil {
		stream.Reset()
		// rsp.Code = int(types.ErrCodeStream)
		// rsp.Message = "Write chat stream failed"
		log.Logger.Errorf("Write chat stream failed: %v", err)
		return http.StatusInternalServerError, int(types.ErrCodeStream), "Write chat stream failed"
	}
	log.Logger.Info("Write chat request into libp2p stream success")

	buf := bufio.NewReader(stream)
	resp, err := http.ReadResponse(buf, hreq)
	if err != nil {
		stream.Reset()
		// rsp.Code = int(types.ErrCodeStream)
		// rsp.Message = "Read chat stream failed"
		log.Logger.Errorf("Read chat stream failed: %v", err)
		return http.StatusInternalServerError, int(types.ErrCodeStream), "Read chat stream failed"
	}
	defer resp.Body.Close()
	log.Logger.Info("Read chat response from libp2p stream success")

	if resp.Header.Get("Content-Type") == "application/json" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			stream.Reset()
			// rsp.Code = int(types.ErrCodeStream)
			// rsp.Message = "Read model response json error"
			log.Logger.Errorf("Read model response json error: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeStream), "Read model response json error"
		}
		if err := json.Unmarshal(body, rsp); err != nil {
			stream.Reset()
			// rsp.Code = int(types.ErrCodeStream)
			// rsp.Message = "Unmarshal model response json error"
			log.Logger.Errorf("Unmarshal model response json error: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeStream), "Unmarshal model response json error"
		}
		// stream.Reset()
		log.Logger.Infof("Read chat json response from libp2p stream {code: %v, message: %v}", rsp.Code, rsp.Message)
		return http.StatusOK, rsp.Code, rsp.Message
	}

	for k, v := range resp.Header {
		for _, s := range v {
			w.Header().Set(k, s)
		}
	}

	w.WriteHeader(resp.StatusCode)

	log.Logger.Info("Copy the body from libp2p stream")
	io.Copy(w, resp.Body)
	// resp.Body.Close()
	log.Logger.Info("Handle chat completion stream request over")
	// rsp.Code = 0
	// rsp.Message = ""
	return http.StatusOK, 0, ""
}

func ChatCompletionHandler(c *gin.Context, publishChan chan<- []byte) {
	rsp := types.ChatCompletionResponse{}

	var msg types.ChatCompletionRequest
	if err := c.ShouldBindJSON(&msg); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	if !msg.Stream {
		status, code, message := handleChatCompletionRequest(c.Request.Context(), publishChan, &msg, &rsp)
		if code != 0 {
			c.JSON(status, types.BaseHttpResponse{
				Code:    code,
				Message: message,
			})
		} else if rsp.Code != 0 {
			c.JSON(http.StatusInternalServerError, rsp)
		} else {
			c.JSON(http.StatusOK, rsp)
		}
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), types.ChatCompletionRequestTimeout)
	defer cancel()
	status, code, message := handleChatCompletionStreamRequest(ctx, c.Writer, &msg, &rsp)
	if code != 0 {
		c.JSON(status, types.BaseHttpResponse{
			Code:    code,
			Message: message,
		})
	} else if rsp.Code != 0 {
		c.JSON(http.StatusInternalServerError, rsp)
	}
}

func ChatCompletionProxyHandler(c *gin.Context, publishChan chan<- []byte) {
	rsp := types.ChatCompletionResponse{}

	if !config.GC.App.PeersCollect.Enabled {
		rsp.Code = int(types.ErrCodeUnsupported)
		rsp.Message = types.ErrCodeUnsupported.String()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	var msg types.ChatCompletionProxyRequest
	if err := c.ShouldBindJSON(&msg); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	ids, code := db.GetPeersOfAIProjects(msg.Project, msg.Model, 20)
	if code != 0 {
		rsp.Code = int(types.ErrCodeProxy)
		rsp.Message = types.ErrorCode(code).String()
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	peers := []types.AIProjectPeerInfo{}
	for id, mi := range ids {
		conn := host.Hio.Connectedness(id)
		if msg.Stream && conn != 1 {
			continue
		}
		latency := host.Hio.Latency(id).Nanoseconds()
		if msg.Stream && latency == 0 {
			continue
		}
		peers = append(peers, types.AIProjectPeerInfo{
			NodeID:       id,
			Connectivity: conn,
			Latency:      latency,
			Idle:         mi.Idle,
			CID:          mi.CID,
		})
	}
	if len(peers) == 0 {
		rsp.Code = int(types.ErrCodeProxy)
		rsp.Message = "Not enough available and directly connected nodes"
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	sort.Sort(types.AIProjectPeerOrder(peers))

	chatReq := types.ChatCompletionRequest{
		NodeID:           peers[0].NodeID,
		CID:              peers[0].CID,
		Project:          msg.Project,
		ChatModelRequest: msg.ChatModelRequest,
	}

	if !msg.Stream {
		status, code, message := handleChatCompletionRequest(c.Request.Context(), publishChan, &chatReq, &rsp)
		if code != 0 {
			c.JSON(status, types.BaseHttpResponse{
				Code:    code,
				Message: message,
			})
		} else if rsp.Code != 0 {
			c.JSON(http.StatusInternalServerError, rsp)
		} else {
			c.JSON(http.StatusOK, rsp)
		}
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), types.ChatCompletionRequestTimeout)
	defer cancel()
	status, code, message := handleChatCompletionStreamRequest(ctx, c.Writer, &chatReq, &rsp)
	if code != 0 {
		c.JSON(status, types.BaseHttpResponse{
			Code:    code,
			Message: message,
		})
	} else if rsp.Code != 0 {
		c.JSON(http.StatusInternalServerError, rsp)
	}
	/*
	   failed_count := 0

	   	for _, peer := range peers {
	   		if failed_count >= 3 {
	   			break
	   		}
	   		chatReq := types.ChatCompletionRequest{
	   			NodeID:           peer.NodeID,
	   			Project:          msg.Project,
	   			ChatModelRequest: msg.ChatModelRequest,
	   		}
	   		if msg.Stream {
	   			ctx, cancel := context.WithTimeout(c.Request.Context(), types.ChatCompletionRequestTimeout)
	   			defer cancel()
	   			status, code, message := handleChatCompletionStreamRequest(ctx, c.Writer, &chatReq, &rsp)
	   			if code != 0 {
	   				log.Logger.Warnf("Roundtrip chat completion proxy %v %v %v to %s in %d time", status, code, message, peer.NodeID, failed_count)
	   				failed_count += 1
	   				continue
	   			} else if rsp.Code != 0 {
	   				log.Logger.Warnf("Roundtrip chat completion proxy %v %v %v to %s in %d time", status, rsp.Code, rsp.Message, peer.NodeID, failed_count)
	   				failed_count += 1
	   				continue
	   			} else {
	   				log.Logger.Infof("Handle chat completion proxy stream request to %s success in %d time", peer.NodeID, failed_count)
	   				return
	   			}
	   		}
	   		status, code, message := handleChatCompletionRequest(c.Request.Context(), publishChan, &chatReq, &rsp)
	   		if code != 0 {
	   			log.Logger.Warnf("Roundtrip chat completion proxy %v %v %v to %s in %d time", status, code, message, peer.NodeID, failed_count)
	   			failed_count += 1
	   			continue
	   		} else if rsp.Code != 0 {
	   			log.Logger.Warnf("Roundtrip chat completion proxy %v %v %v to %s in %d time", status, rsp.Code, rsp.Message, peer.NodeID, failed_count)
	   			failed_count += 1
	   			continue
	   		} else {
	   			c.JSON(http.StatusOK, rsp)
	   			log.Logger.Infof("Handle chat completion proxy stream request to %s success in %d time", peer.NodeID, failed_count)
	   			return
	   		}
	   	}

	   rsp.Code = int(types.ErrCodeProxy)
	   rsp.Message = fmt.Sprintf("Failed %d times", failed_count)
	   c.JSON(http.StatusInternalServerError, rsp)
	*/
}

func handleImageGenRequest(ctx context.Context, publishChan chan<- []byte, req types.ImageGenerationRequest, rsp *types.ImageGenerationResponse) (int, int, string) {
	if req.NodeID == config.GC.Identity.PeerID {
		mi, err := model.GetModelInfo(req.Project, req.Model, req.CID)
		if err != nil {
			return http.StatusInternalServerError, int(types.ErrCodeModel), err.Error()
		}
		model.IncRef(req.Project, req.Model, mi.CID)
		timer.SendAIProjects(publishChan)
		defer func() {
			model.DecRef(req.Project, req.Model, mi.CID)
			timer.SendAIProjects(publishChan)
		}()
		*rsp = *model.ImageGenerationModel(mi.API, req.ImageGenModelRequest)
		log.Logger.Infof("Execute model %s result {code:%d, message:%s}", req.Model, rsp.Code, rsp.Message)
		return http.StatusOK, rsp.Code, rsp.Message
	}

	if req.ResponseFormat == "b64_json" {
		log.Logger.Info("Received image gen b64_json request")
		ctx, cancel := context.WithTimeout(ctx, types.ImageGenerationRequestTimeout)
		defer cancel()

		jsonData, err := json.Marshal(req.ImageGenModelRequest)
		if err != nil {
			// rsp.Code = int(types.ErrCodeStream)
			// rsp.Message = "Marshal model request body failed"
			log.Logger.Errorf("Marshal model request body failed: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeStream), "Marshal model request body failed"
		}
		hreq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:8080/api/v0/image/gen", bytes.NewBuffer(jsonData))
		if err != nil {
			// rsp.Code = int(types.ErrCodeStream)
			// rsp.Message = "Create http request for stream failed"
			log.Logger.Errorf("Create http request for stream failed: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeStream), "Create http request for stream failed"
		}

		hreq.Header.Set("Content-Type", "application/json")

		queryValues := hreq.URL.Query()
		queryValues.Add("project", req.Project)
		queryValues.Add("model", req.Model)
		queryValues.Add("cid", req.CID)
		hreq.URL.RawQuery = queryValues.Encode()

		stream, err := host.Hio.NewStream(ctx, req.NodeID)
		if err != nil {
			// rsp.Code = int(types.ErrCodeStream)
			// rsp.Message = "Open stream with peer node failed"
			log.Logger.Errorf("Open stream with peer node failed: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeStream), "Open stream with peer node failed"
		}
		stream.SetDeadline(time.Now().Add(types.ImageGenerationRequestTimeout))
		defer stream.Close()
		log.Logger.Infof("Create libp2p stream with %s success", req.NodeID)

		err = hreq.Write(stream)
		if err != nil {
			stream.Reset()
			// rsp.Code = int(types.ErrCodeStream)
			// rsp.Message = "Write image gen request into libp2p stream failed"
			log.Logger.Errorf("Write image gen request into libp2p stream failed: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeStream), "Write image gen request into libp2p stream failed"
		}
		log.Logger.Info("Write image gen request into libp2p stream success")

		reader := bufio.NewReader(stream)
		responseCh := make(chan *types.ImageGenerationResponse, 1)

		go func() {
			response := &types.ImageGenerationResponse{}
			select {
			case <-ctx.Done():
				response.Code = int(types.ErrCodeStream)
				response.Message = fmt.Sprintf("Context canceled or timed out: %v", ctx.Err())
				log.Logger.Errorf("Context canceled or timed out: %v", ctx.Err())
				responseCh <- response
				return
			default:
				resp, err := http.ReadResponse(reader, hreq)
				select {
				case <-ctx.Done():
					response.Code = int(types.ErrCodeStream)
					response.Message = fmt.Sprintf("Context canceled or timed out: %v", ctx.Err())
					log.Logger.Errorf("Context canceled or timed out: %v", ctx.Err())
					responseCh <- response
					return
				default:
					if err != nil {
						stream.Reset()
						response.Code = int(types.ErrCodeStream)
						response.Message = "Read image gen response from libp2p stream failed"
						log.Logger.Errorf("Read image gen response from libp2p stream failed: %v", err)
						responseCh <- response
						return
					}
					log.Logger.Info("Read image gen response from libp2p stream success")

					defer resp.Body.Close()
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						stream.Reset()
						response.Code = int(types.ErrCodeStream)
						response.Message = "Read image response body failed"
						log.Logger.Errorf("Read image response body failed: %v", err)
						responseCh <- response
						return
					}
					log.Logger.Info("Read image response body success")

					// response := types.ImageGenModelResponse{}
					if err := json.Unmarshal(body, &response); err != nil {
						stream.Reset()
						response.Code = int(types.ErrCodeStream)
						response.Message = "Unmarshal image response from stream error"
						log.Logger.Errorf("Unmarshal image response from stream error: %v", err)
						responseCh <- response
						return
					}
					responseCh <- response
				}
			}
		}()

		select {
		case <-ctx.Done():
			// rsp.Code = int(types.ErrCodeStream)
			// rsp.Message = fmt.Sprintf("Context canceled or timed out: %v", ctx.Err())
			log.Logger.Errorf("Handle image gen stream request time out: %v", ctx.Err())
			return http.StatusInternalServerError, int(types.ErrCodeStream), fmt.Sprintf("Context canceled or timed out: %v", ctx.Err())
		case resp := <-responseCh:
			rsp.Code = resp.Code
			rsp.Message = resp.Message
			rsp.Created = resp.Created
			rsp.Choices = resp.Choices
			log.Logger.Info("Handle image gen stream request over")
			return http.StatusOK, rsp.Code, rsp.Message
		}
	}

	requestID, err := uuid.NewRandom()
	if err != nil {
		return http.StatusInternalServerError, int(types.ErrCodeUUID), err.Error()
	}

	pi := &protocol.ImageGenerationBody{
		Data: &protocol.ImageGenerationBody_Req{
			Req: &protocol.ImageGenerationRequest{
				Project:        req.Project,
				File:           req.File,
				Model:          req.Model,
				Prompt:         req.Prompt,
				Number:         int32(req.Number),
				Size:           req.Size,
				Width:          int32(req.Width),
				Height:         int32(req.Height),
				ResponseFormat: req.ResponseFormat,
				Wallet: &protocol.WalletVerification{
					Wallet:    req.Wallet,
					Signature: req.Signature,
					Hash:      req.Hash,
				},
				Cid:  req.CID,
				Step: req.Step,
			},
		},
	}
	body, err := proto.Marshal(pi)
	if err != nil {
		return http.StatusInternalServerError, int(types.ErrCodeProtobuf), err.Error()
	}
	body, err = host.Encrypt(ctx, req.NodeID, body)
	if err != nil {
		return http.StatusInternalServerError, int(types.ErrCodeEncrypt), types.ErrCodeEncrypt.String()
	}

	msg := &protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: host.Hio.UserAgent,
			Timestamp:     time.Now().Unix(),
			Id:            requestID.String(),
			NodeId:        config.GC.Identity.PeerID,
			Receiver:      req.NodeID,
			NodePubKey:    nil,
			Sign:          nil,
		},
		Type:       *protocol.MessageType_IMAGE_GENERATION.Enum(),
		Body:       body,
		ResultCode: 0,
	}
	msg.Header.NodePubKey, _ = host.MarshalPubKeyFromPrivKey(host.Hio.PrivKey)
	return handleRequest(publishChan, msg, rsp, types.ImageGenerationRequestTimeout)
}

func ImageGenHandler(c *gin.Context, publishChan chan<- []byte) {
	rsp := types.ImageGenerationResponse{}

	var msg types.ImageGenerationRequest
	if err := c.ShouldBindJSON(&msg); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	status, code, message := handleImageGenRequest(c.Request.Context(), publishChan, msg, &rsp)
	if code != 0 {
		c.JSON(status, types.BaseHttpResponse{
			Code:    code,
			Message: message,
		})
	} else if rsp.Code != 0 {
		c.JSON(http.StatusInternalServerError, rsp)
	} else {
		c.JSON(http.StatusOK, rsp)
	}
}

func ImageGenProxyHandler(c *gin.Context, publishChan chan<- []byte) {
	rsp := types.ImageGenerationResponse{}

	if !config.GC.App.PeersCollect.Enabled {
		rsp.Code = int(types.ErrCodeUnsupported)
		rsp.Message = types.ErrCodeUnsupported.String()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	var msg types.ImageGenerationProxyRequest
	if err := c.ShouldBindJSON(&msg); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	ids, code := db.GetPeersOfAIProjects(msg.Project, msg.Model, 20)
	if code != 0 {
		rsp.Code = int(types.ErrCodeProxy)
		rsp.Message = types.ErrorCode(code).String()
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	peers := []types.AIProjectPeerInfo{}
	for id, mi := range ids {
		conn := host.Hio.Connectedness(id)
		if msg.ResponseFormat == "b64_json" && conn != 1 {
			continue
		}
		latency := host.Hio.Latency(id).Nanoseconds()
		if msg.ResponseFormat == "b64_json" && latency == 0 {
			continue
		}
		peers = append(peers, types.AIProjectPeerInfo{
			NodeID:       id,
			Connectivity: conn,
			Latency:      latency,
			Idle:         mi.Idle,
			CID:          mi.CID,
		})
	}
	if len(peers) == 0 {
		rsp.Code = int(types.ErrCodeProxy)
		rsp.Message = "Not enough available and directly connected nodes"
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	sort.Sort(types.AIProjectPeerOrder(peers))

	igReq := types.ImageGenerationRequest{
		NodeID:               peers[0].NodeID,
		CID:                  peers[0].CID,
		Project:              msg.Project,
		ImageGenModelRequest: msg.ImageGenModelRequest,
	}
	status, code, message := handleImageGenRequest(c.Request.Context(), publishChan, igReq, &rsp)
	if code != 0 {
		c.JSON(status, types.BaseHttpResponse{
			Code:    code,
			Message: message,
		})
	} else if rsp.Code != 0 {
		c.JSON(http.StatusInternalServerError, rsp)
	} else {
		c.JSON(http.StatusOK, rsp)
	}
	/*
	   failed_count := 0

	   	for _, peer := range peers {
	   		if failed_count >= 3 {
	   			break
	   		}
	   		igReq := types.ImageGenerationRequest{
	   			NodeID:               peer.NodeID,
	   			Project:              msg.Project,
	   			ImageGenModelRequest: msg.ImageGenModelRequest,
	   		}
	   		status, code, message := handleImageGenRequest(c.Request.Context(), publishChan, igReq, &rsp)
	   		if code != 0 {
	   			log.Logger.Warnf("Roundtrip image gen proxy %v %v %v to %s in %d time", status, code, message, peer.NodeID, failed_count)
	   			failed_count += 1
	   			continue
	   		} else if rsp.Code != 0 {
	   			log.Logger.Warnf("Roundtrip image gen proxy %v %v to %s in %d time", rsp.Code, rsp.Message, peer.NodeID, failed_count)
	   			failed_count += 1
	   			continue
	   		} else {
	   			c.JSON(http.StatusOK, rsp)
	   			log.Logger.Infof("Handle image gen proxy request to %s success in %d time", peer.NodeID, failed_count)
	   			return
	   		}
	   	}

	   rsp.Code = int(types.ErrCodeProxy)
	   rsp.Message = fmt.Sprintf("Failed %d times", failed_count)
	   c.JSON(http.StatusInternalServerError, rsp)
	*/
}

func handleImageEditRequest(ctx context.Context, publishChan chan<- []byte, w http.ResponseWriter, form *multipart.Form, req types.ImageGenerationRequest) (int, int, string) {
	if req.NodeID == config.GC.Identity.PeerID {
		mi, err := model.GetModelInfo(req.Project, req.Model, req.CID)
		if err != nil {
			return http.StatusInternalServerError, int(types.ErrCodeModel), err.Error()
		}
		model.IncRef(req.Project, req.Model, mi.CID)
		timer.SendAIProjects(publishChan)
		defer func() {
			model.DecRef(req.Project, req.Model, mi.CID)
			timer.SendAIProjects(publishChan)
		}()
		resp, err := model.ImageEditModel(mi.API, form)
		if err != nil {
			log.Logger.Errorf("RoundTrip image edit request failed: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeModel), fmt.Sprintf("RoundTrip image edit request failed: %v", err)
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			for _, s := range v {
				w.Header().Set(k, s)
			}
		}

		w.WriteHeader(resp.StatusCode)

		log.Logger.Info("Copy roundtrip image edit response")
		io.Copy(w, resp.Body)
		// resp.Body.Close()
		log.Logger.Info("Handle image edit stream request over from the node itself")
		// rsp.Code = 0
		// rsp.Message = ""
		return http.StatusOK, 0, ""
	}

	if host.Hio.Connectedness(req.NodeID) != 1 {
		return http.StatusInternalServerError, int(types.ErrCodeStream), "Not available and directly connected node"
	}

	log.Logger.Info("Received image edit b64_json request")
	ctx, cancel := context.WithTimeout(ctx, types.ImageGenerationRequestTimeout)
	defer cancel()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// copy file field in form
	for fieldName, files := range form.File {
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				msg := fmt.Sprintf("Failed to open %v %v", fieldName, err)
				log.Logger.Errorf(msg)
				return http.StatusInternalServerError, int(types.ErrCodeModel), msg
			}
			defer file.Close()

			part, err := writer.CreateFormFile(fieldName, fileHeader.Filename)
			if err != nil {
				msg := fmt.Sprintf("Failed to create form file for %v %v", fieldName, err)
				log.Logger.Errorf(msg)
				return http.StatusInternalServerError, int(types.ErrCodeModel), msg
			}

			_, err = io.Copy(part, file)
			if err != nil {
				msg := fmt.Sprintf("Failed to copy form file for %v %v", fieldName, err)
				log.Logger.Errorf(msg)
				return http.StatusInternalServerError, int(types.ErrCodeModel), msg
			}
		}
	}

	// copy other field in form
	for fieldName, values := range form.Value {
		for _, value := range values {
			if err := writer.WriteField(fieldName, value); err != nil {
				msg := fmt.Sprintf("Failed to copy %v field %v", fieldName, err)
				log.Logger.Errorf(msg)
				return http.StatusInternalServerError, int(types.ErrCodeModel), msg
			}
		}
	}

	if err := writer.Close(); err != nil {
		msg := fmt.Sprintf("Failed to close multipart writer %v", err)
		log.Logger.Errorf(msg)
		return http.StatusInternalServerError, int(types.ErrCodeModel), msg
	}
	hreq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:8080/api/v0/image/edit", body)
	if err != nil {
		// rsp.Code = int(types.ErrCodeModel)
		// rsp.Message = "Create http request for stream failed"
		log.Logger.Errorf("Create http request for stream failed: %v", err)
		return http.StatusInternalServerError, int(types.ErrCodeModel), "Create http request for stream failed"
	}

	hreq.Header.Set("Content-Type", writer.FormDataContentType())

	queryValues := hreq.URL.Query()
	queryValues.Add("project", req.Project)
	queryValues.Add("model", req.Model)
	queryValues.Add("cid", req.CID)
	hreq.URL.RawQuery = queryValues.Encode()

	stream, err := host.Hio.NewStream(ctx, req.NodeID)
	if err != nil {
		// rsp.Code = int(types.ErrCodeStream)
		// rsp.Message = "Open stream with peer node failed"
		log.Logger.Errorf("Open stream with peer node failed: %v", err)
		return http.StatusInternalServerError, int(types.ErrCodeStream), "Open stream with peer node failed"
	}
	stream.SetDeadline(time.Now().Add(types.ImageGenerationRequestTimeout))
	defer stream.Close()
	log.Logger.Infof("Create libp2p stream with %s success", req.NodeID)

	err = hreq.Write(stream)
	if err != nil {
		stream.Reset()
		// rsp.Code = int(types.ErrCodeStream)
		// rsp.Message = "Write image edit request into libp2p stream failed"
		log.Logger.Errorf("Write image edit request into libp2p stream failed: %v", err)
		return http.StatusInternalServerError, int(types.ErrCodeStream), "Write image edit request into libp2p stream failed"
	}
	log.Logger.Info("Write image edit request into libp2p stream success")

	reader := bufio.NewReader(stream)
	responseCh := make(chan *types.ImageGenerationResponse, 1)

	go func() {
		response := &types.ImageGenerationResponse{}
		select {
		case <-ctx.Done():
			response.Code = int(types.ErrCodeStream)
			response.Message = fmt.Sprintf("Context canceled or timed out: %v", ctx.Err())
			log.Logger.Errorf("Context canceled or timed out: %v", ctx.Err())
			responseCh <- response
			return
		default:
			resp, err := http.ReadResponse(reader, hreq)
			select {
			case <-ctx.Done():
				response.Code = int(types.ErrCodeStream)
				response.Message = fmt.Sprintf("Context canceled or timed out: %v", ctx.Err())
				log.Logger.Errorf("Context canceled or timed out: %v", ctx.Err())
				responseCh <- response
				return
			default:
				if err != nil {
					stream.Reset()
					response.Code = int(types.ErrCodeStream)
					response.Message = "Read image edit response from libp2p stream failed"
					log.Logger.Errorf("Read image edit response from libp2p stream failed: %v", err)
					responseCh <- response
					return
				}
				log.Logger.Info("Read image edit response from libp2p stream success")

				defer resp.Body.Close()

				for k, v := range resp.Header {
					for _, s := range v {
						w.Header().Set(k, s)
					}
				}

				w.WriteHeader(resp.StatusCode)

				log.Logger.Info("Copy the body from image edit stream")
				io.Copy(w, resp.Body)
				// resp.Body.Close()
				log.Logger.Info("Handle image edit stream request over")
				// rsp.Code = 0
				// rsp.Message = ""
				response.Code = 0
				response.Message = ""
				responseCh <- response
			}
		}
	}()

	select {
	case <-ctx.Done():
		// rsp.Code = int(types.ErrCodeStream)
		// rsp.Message = fmt.Sprintf("Context canceled or timed out: %v", ctx.Err())
		log.Logger.Errorf("Handle image edit stream request time out: %v", ctx.Err())
		return http.StatusInternalServerError, int(types.ErrCodeStream), fmt.Sprintf("Context canceled or timed out: %v", ctx.Err())
	case resp := <-responseCh:
		log.Logger.Info("Handle image edit stream request over")
		return http.StatusOK, resp.Code, resp.Message
	}
}

func ImageEditHandler(c *gin.Context, publishChan chan<- []byte) {
	rsp := types.ImageGenerationResponse{}

	var msg types.ImageGenerationRequest
	if err := c.ShouldBindQuery(&msg); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = types.ErrCodeParam.String()
		c.JSON(http.StatusUnprocessableEntity, rsp)
	}
	status, code, message := handleImageEditRequest(c.Request.Context(), publishChan, c.Writer, form, msg)
	if code != 0 {
		c.JSON(status, types.BaseHttpResponse{
			Code:    code,
			Message: message,
		})
	}
}

func ImageEditProxyHandler(c *gin.Context, publishChan chan<- []byte) {
	rsp := types.ImageGenerationResponse{}

	if !config.GC.App.PeersCollect.Enabled {
		rsp.Code = int(types.ErrCodeUnsupported)
		rsp.Message = types.ErrCodeUnsupported.String()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	var msg types.ImageGenerationProxyRequest
	if err := c.ShouldBindQuery(&msg); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = types.ErrCodeParam.String()
		c.JSON(http.StatusUnprocessableEntity, rsp)
	}

	ids, code := db.GetPeersOfAIProjects(msg.Project, msg.Model, 20)
	if code != 0 {
		rsp.Code = int(types.ErrCodeProxy)
		rsp.Message = types.ErrorCode(code).String()
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	peers := []types.AIProjectPeerInfo{}
	for id, mi := range ids {
		conn := host.Hio.Connectedness(id)
		if /*msg.ResponseFormat == "b64_json" &&*/ conn != 1 {
			continue
		}
		latency := host.Hio.Latency(id).Nanoseconds()
		if /*msg.ResponseFormat == "b64_json" &&*/ latency == 0 {
			continue
		}
		peers = append(peers, types.AIProjectPeerInfo{
			NodeID:       id,
			Connectivity: conn,
			Latency:      latency,
			Idle:         mi.Idle,
			CID:          mi.CID,
		})
	}
	if len(peers) == 0 {
		rsp.Code = int(types.ErrCodeProxy)
		rsp.Message = "Not enough available and directly connected nodes"
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	sort.Sort(types.AIProjectPeerOrder(peers))

	igReq := types.ImageGenerationRequest{
		NodeID:               peers[0].NodeID,
		CID:                  peers[0].CID,
		Project:              msg.Project,
		ImageGenModelRequest: msg.ImageGenModelRequest,
	}
	status, code, message := handleImageEditRequest(c.Request.Context(), publishChan, c.Writer, form, igReq)
	if code != 0 {
		c.JSON(status, types.BaseHttpResponse{
			Code:    code,
			Message: message,
		})
	}
}
