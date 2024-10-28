package serve

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/model"
	"AIComputingNode/pkg/p2p"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/types"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (hs *httpService) handleChatCompletionRequest(ctx context.Context, req *types.ChatCompletionRequest, rsp *types.ChatCompletionResponse) {
	if req.NodeID == config.GC.Identity.PeerID {
		modelAPI := config.GC.GetModelAPI(req.Project, req.Model)
		*rsp = *model.ChatModel(modelAPI, req.ChatModelRequest)
		log.Logger.Infof("Execute model %s result {code:%d, message:%s}", req.Model, rsp.Code, rsp.Message)
		return
	}

	requestID, err := uuid.NewRandom()
	if err != nil {
		rsp.Code = int(types.ErrCodeUUID)
		rsp.Message = err.Error()
		return
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
			},
		},
	}
	body, err := proto.Marshal(pi)
	if err != nil {
		rsp.SetCode(int(types.ErrCodeProtobuf))
		rsp.SetMessage(err.Error())
		return
	}
	body, err = p2p.Encrypt(ctx, req.NodeID, body)
	if err != nil {
		rsp.Code = int(types.ErrCodeEncrypt)
		rsp.Message = types.ErrCodeEncrypt.String()
		return
	}

	msg := &protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: p2p.Hio.UserAgent,
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
	msg.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
	hs.handleRequest(msg, rsp)
}

func (hs *httpService) handleChatCompletionStreamRequest(ctx context.Context, w http.ResponseWriter, req *types.ChatCompletionRequest, rsp *types.ChatCompletionResponse) {
	if req.NodeID == config.GC.Identity.PeerID {
		modelAPI := config.GC.GetModelAPI(req.Project, req.Model)
		log.Logger.Info("Received chat completion stream request from the node itself")
		if modelAPI == "" {
			rsp.Code = int(types.ErrCodeModel)
			rsp.Message = "Model API configuration is empty"
			// w.Header().Set("Content-Type", "application/json")
			// json.NewEncoder(w).Encode(rsp)
			return
		}

		jsonData, err := json.Marshal(req.ChatModelRequest)
		if err != nil {
			rsp.Code = int(types.ErrCodeModel)
			rsp.Message = "Marshal model request body failed"
			log.Logger.Errorf("Marshal model request body failed: %v", err)
			return
		}
		hreq, err := http.NewRequestWithContext(ctx, "POST", modelAPI, bytes.NewBuffer(jsonData))
		if err != nil {
			rsp.Code = int(types.ErrCodeModel)
			rsp.Message = "Copy http request body failed"
			// w.Header().Set("Content-Type", "application/json")
			// json.NewEncoder(w).Encode(rsp)
			return
		}
		hreq.Header.Set("Content-Type", "application/json")

		log.Logger.Infof("Making request to %s\n", hreq.URL)
		resp, err := http.DefaultTransport.RoundTrip(hreq)
		if err != nil {
			rsp.Code = int(types.ErrCodeModel)
			rsp.Message = fmt.Sprintf("RoundTrip chat request failed: %v", err)
			log.Logger.Errorf("RoundTrip chat request failed: %v", err)
			// w.Header().Set("Content-Type", "application/json")
			// json.NewEncoder(w).Encode(rsp)
			return
		}
		defer resp.Body.Close()

		if resp.Header.Get("Content-Type") == "application/json" {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				rsp.Code = int(types.ErrCodeStream)
				rsp.Message = "Read model response json error"
				log.Logger.Errorf("Read model response json error: %v", err)
				return
			}
			if err := json.Unmarshal(body, rsp); err != nil {
				rsp.Code = int(types.ErrCodeStream)
				rsp.Message = "Unmarshal model response json error"
				log.Logger.Errorf("Unmarshal model response json error: %v", err)
				return
			}
			return
		}

		for k, v := range resp.Header {
			for _, s := range v {
				w.Header().Set(k, s)
			}
		}

		w.WriteHeader(resp.StatusCode)

		log.Logger.Info("Copy roundtrip response")
		io.Copy(w, resp.Body)
		// resp.Body.Close()
		log.Logger.Info("Handle chat completion stream request over from the node itself")
		rsp.Code = 0
		rsp.Message = ""
		return
	}

	log.Logger.Info("Received chat completion stream request")
	stream, err := p2p.Hio.NewStream(ctx, req.NodeID)
	if err != nil {
		rsp.Code = int(types.ErrCodeStream)
		rsp.Message = "Open stream with peer node failed"
		log.Logger.Errorf("Open stream with peer node failed: %v", err)
		// w.Header().Set("Content-Type", "application/json")
		// json.NewEncoder(w).Encode(rsp)
		return
	}
	stream.SetDeadline(time.Now().Add(p2p.ChatProxyStreamTimeout))
	defer stream.Close()
	log.Logger.Infof("Create libp2p stream with %s success", req.NodeID)

	jsonData, err := json.Marshal(req.ChatModelRequest)
	if err != nil {
		rsp.Code = int(types.ErrCodeStream)
		rsp.Message = "Marshal model request body failed"
		log.Logger.Errorf("Marshal model request body failed: %v", err)
		return
	}
	hreq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:8080/api/v0/chat/completion", bytes.NewBuffer(jsonData))
	if err != nil {
		rsp.Code = int(types.ErrCodeStream)
		rsp.Message = "Create http request for stream failed"
		log.Logger.Errorf("Create http request for stream failed: %v", err)
		return
	}

	hreq.Header.Set("Content-Type", "application/json")
	queryValues := hreq.URL.Query()
	queryValues.Add("project", req.Project)
	queryValues.Add("model", req.Model)
	hreq.URL.RawQuery = queryValues.Encode()

	err = hreq.Write(stream)
	if err != nil {
		stream.Reset()
		rsp.Code = int(types.ErrCodeStream)
		rsp.Message = "Write chat stream failed"
		log.Logger.Errorf("Write chat stream failed: %v", err)
		// w.Header().Set("Content-Type", "application/json")
		// json.NewEncoder(w).Encode(rsp)
		return
	}

	log.Logger.Info("Read the response that was send from dest peer")
	buf := bufio.NewReader(stream)
	resp, err := http.ReadResponse(buf, hreq)
	if err != nil {
		stream.Reset()
		rsp.Code = int(types.ErrCodeStream)
		rsp.Message = "Read chat stream failed"
		log.Logger.Errorf("Read chat stream failed: %v", err)
		// w.Header().Set("Content-Type", "application/json")
		// json.NewEncoder(w).Encode(rsp)
		return
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") == "application/json" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			stream.Reset()
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Read model response json error"
			log.Logger.Errorf("Read model response json error: %v", err)
			return
		}
		if err := json.Unmarshal(body, rsp); err != nil {
			stream.Reset()
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Unmarshal model response json error"
			log.Logger.Errorf("Unmarshal model response json error: %v", err)
			return
		}
		stream.Reset()
		return
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
	rsp.Code = 0
	rsp.Message = ""
}

func (hs *httpService) chatCompletionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	rsp := types.ChatCompletionResponse{
		Code:    0,
		Message: "ok",
	}

	var msg types.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if msg.Stream {
		hs.handleChatCompletionStreamRequest(r.Context(), w, &msg, &rsp)
		if rsp.Code != 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rsp)
		}
		return
	}

	hs.handleChatCompletionRequest(r.Context(), &msg, &rsp)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rsp)
}

func (hs *httpService) chatCompletionProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	rsp := types.ChatCompletionResponse{
		Code:    0,
		Message: "ok",
	}

	var msg types.ChatCompletionProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rsp)
		return
	}

	ids, code := db.GetPeersOfAIProjects(msg.Project, msg.Model, 20)
	if code != 0 {
		rsp.Code = int(types.ErrCodeProxy)
		rsp.Message = types.ErrorCode(code).String()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rsp)
		return
	}

	peers := []types.AIProjectPeerInfo{}
	for _, id := range ids {
		conn := p2p.Hio.Connectedness(id)
		if conn != 1 {
			continue
		}
		latency := p2p.Hio.Latency(id).Nanoseconds()
		if latency == 0 {
			continue
		}
		peers = append(peers, types.AIProjectPeerInfo{
			NodeID:       id,
			Connectivity: 1,
			Latency:      latency,
		})
	}
	if len(peers) == 0 {
		rsp.Code = int(types.ErrCodeProxy)
		rsp.Message = "Not enough available and directly connected nodes"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rsp)
		return
	}

	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Latency < peers[j].Latency
	})

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
			hs.handleChatCompletionStreamRequest(r.Context(), w, &chatReq, &rsp)
			if rsp.Code != 0 {
				log.Logger.Warnf("Roundtrip chat completion proxy %v %v to %s in %d time", rsp.Code, rsp.Message, peer.NodeID, failed_count)
				failed_count += 1
				continue
			} else {
				return
			}
		}
		hs.handleChatCompletionRequest(r.Context(), &chatReq, &rsp)
		if rsp.Code != 0 {
			log.Logger.Warnf("Roundtrip chat completion proxy %v %v to %s in %d time", rsp.Code, rsp.Message, peer.NodeID, failed_count)
			failed_count += 1
			continue
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rsp)
			return
		}
	}

	rsp.Code = int(types.ErrCodeProxy)
	rsp.Message = fmt.Sprintf("Failed %d times", failed_count)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rsp)
}

func (hs *httpService) handleImageGenRequest(ctx context.Context, req *types.ImageGenerationRequest, rsp *types.ImageGenerationResponse) {
	if req.NodeID == config.GC.Identity.PeerID {
		modelAPI := config.GC.GetModelAPI(req.Project, req.Model)
		if modelAPI == "" {
			rsp.Code = int(types.ErrCodeModel)
			rsp.Message = "Model API configuration is empty"
			return
		}
		*rsp = *model.ImageGenerationModel(modelAPI, req.ImageGenModelRequest)
		log.Logger.Infof("Execute model %s result {code:%d, message:%s}", req.Model, rsp.Code, rsp.Message)
		return
	}

	if req.ResponseFormat == "b64_json" {
		log.Logger.Info("Received image gen b64_json request")
		stream, err := p2p.Hio.NewStream(ctx, req.NodeID)
		if err != nil {
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Open stream with peer node failed"
			log.Logger.Errorf("Open stream with peer node failed: %v", err)
			return
		}
		stream.SetDeadline(time.Now().Add(p2p.ChatProxyStreamTimeout))
		defer stream.Close()
		log.Logger.Infof("Create libp2p stream with %s success", req.NodeID)

		jsonData, err := json.Marshal(req.ImageGenModelRequest)
		if err != nil {
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Marshal model request body failed"
			log.Logger.Errorf("Marshal model request body failed: %v", err)
			return
		}
		hreq, err := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:8080/api/v0/image/gen", bytes.NewBuffer(jsonData))
		if err != nil {
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Create http request for stream failed"
			log.Logger.Errorf("Create http request for stream failed: %v", err)
			return
		}

		hreq.Header.Set("Content-Type", "application/json")
		queryValues := hreq.URL.Query()
		queryValues.Add("project", req.Project)
		queryValues.Add("model", req.Model)
		hreq.URL.RawQuery = queryValues.Encode()

		err = hreq.Write(stream)
		if err != nil {
			stream.Reset()
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Write image gen request into libp2p stream failed"
			log.Logger.Errorf("Write image gen request into libp2p stream failed: %v", err)
			return
		}
		log.Logger.Info("Write image gen request into libp2p stream success")

		reader := bufio.NewReader(stream)
		responseCh := make(chan *types.ImageGenModelResponse, 1)

		go func() {
			response := &types.ImageGenModelResponse{}
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
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = fmt.Sprintf("Context canceled or timed out: %v", ctx.Err())
			log.Logger.Errorf("Handle image gen stream request time out: %v", ctx.Err())
			return
		case resp := <-responseCh:
			rsp.Code = resp.Code
			rsp.Message = resp.Message
			rsp.Data.Created = resp.Created
			rsp.Data.Choices = resp.Data
			log.Logger.Info("Handle image gen stream request over")
			return
		}
	}

	requestID, err := uuid.NewRandom()
	if err != nil {
		rsp.Code = int(types.ErrCodeUUID)
		rsp.Message = err.Error()
		return
	}

	pi := &protocol.ImageGenerationBody{
		Data: &protocol.ImageGenerationBody_Req{
			Req: &protocol.ImageGenerationRequest{
				Project:        req.Project,
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
			},
		},
	}
	body, err := proto.Marshal(pi)
	if err != nil {
		rsp.SetCode(int(types.ErrCodeProtobuf))
		rsp.SetMessage(err.Error())
		return
	}
	body, err = p2p.Encrypt(ctx, req.NodeID, body)
	if err != nil {
		rsp.Code = int(types.ErrCodeEncrypt)
		rsp.Message = types.ErrCodeEncrypt.String()
		return
	}

	msg := &protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: p2p.Hio.UserAgent,
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
	msg.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
	hs.handleRequest(msg, rsp)
}

func (hs *httpService) imageGenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	rsp := types.ImageGenerationResponse{
		Code:    0,
		Message: "ok",
	}
	w.Header().Set("Content-Type", "application/json")

	var msg types.ImageGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestProcessTimeout)
	defer cancel()
	hs.handleImageGenRequest(ctx, &msg, &rsp)
	json.NewEncoder(w).Encode(rsp)
}

func (hs *httpService) imageGenProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	rsp := types.ImageGenerationResponse{
		Code:    0,
		Message: "ok",
	}
	w.Header().Set("Content-Type", "application/json")

	var msg types.ImageGenerationProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	ids, code := db.GetPeersOfAIProjects(msg.Project, msg.Model, 20)
	if code != 0 {
		rsp.Code = int(types.ErrCodeProxy)
		rsp.Message = types.ErrorCode(code).String()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	peers := []types.AIProjectPeerInfo{}
	for _, id := range ids {
		// conn := p2p.Hio.Connectedness(id)
		// if conn != 1 {
		// 	continue
		// }
		latency := p2p.Hio.Latency(id).Nanoseconds()
		if latency == 0 {
			continue
		}
		peers = append(peers, types.AIProjectPeerInfo{
			NodeID:       id,
			Connectivity: p2p.Hio.Connectedness(id),
			Latency:      latency,
		})
	}
	if len(peers) == 0 {
		rsp.Code = int(types.ErrCodeProxy)
		rsp.Message = "Not enough available and directly connected nodes"
		json.NewEncoder(w).Encode(rsp)
		return
	}

	sort.Slice(peers, func(i, j int) bool {
		if (peers[i].Connectivity == 1 && peers[j].Connectivity == 1) ||
			(peers[i].Connectivity != 1 && peers[j].Connectivity != 1) {
			return peers[i].Latency < peers[j].Latency
		}
		if peers[i].Connectivity == 1 {
			return true
		}
		return false
	})

	ctx, cancel := context.WithTimeout(r.Context(), requestProcessTimeout)
	defer cancel()
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
		hs.handleImageGenRequest(ctx, &igReq, &rsp)
		if rsp.Code != 0 {
			log.Logger.Warnf("Handle image gen proxy response %v %v to %v in %d time",
				rsp.Code, rsp.Message, peer.NodeID, failed_count)
			failed_count += 1
			continue
		}
		json.NewEncoder(w).Encode(rsp)
		return
	}

	rsp.Code = int(types.ErrCodeProxy)
	rsp.Message = fmt.Sprintf("Failed %d times", failed_count)
	json.NewEncoder(w).Encode(rsp)
}
