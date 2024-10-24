package serve

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
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

	if msg.NodeID == config.GC.Identity.PeerID {
		modelAPI := config.GC.GetModelAPI(msg.Project, msg.Model)
		if msg.Stream {
			log.Logger.Info("Received chat completion stream request from the node itself")
			if modelAPI == "" {
				rsp.Code = int(types.ErrCodeModel)
				rsp.Message = "Model API configuration is empty"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(rsp)
				return
			}
			req := new(http.Request)
			*req = *r
			var err error = nil
			req.URL, err = url.Parse(modelAPI)
			if err != nil {
				rsp.Code = int(types.ErrCodeModel)
				rsp.Message = "Parse model api interface failed"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(rsp)
				return
			}
			req.Host = req.URL.Host
			req.Body, req.ContentLength, err = msg.ChatModelRequest.RequestBody()
			if err != nil {
				rsp.Code = int(types.ErrCodeModel)
				rsp.Message = "Copy http request body failed"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(rsp)
				return
			}
			log.Logger.Infof("Making request to %s\n", req.URL)
			resp, err := http.DefaultTransport.RoundTrip(req)
			if err != nil {
				rsp.Code = int(types.ErrCodeModel)
				rsp.Message = "RoundTrip chat request failed"
				log.Logger.Errorf("RoundTrip chat request failed: %v", err)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(rsp)
				return
			}

			for k, v := range resp.Header {
				for _, s := range v {
					w.Header().Add(k, s)
				}
			}

			w.WriteHeader(resp.StatusCode)

			log.Logger.Info("Copy roundtrip response")
			io.Copy(w, resp.Body)
			resp.Body.Close()
			log.Logger.Info("Handle chat completion stream request over from the node itself")
			return
		} else {
			rsp = *model.ChatModel(modelAPI, msg.ChatModelRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rsp)
			return
		}
	}

	if msg.Stream {
		log.Logger.Info("Received chat completion stream request")
		stream, err := p2p.Hio.NewStream(r.Context(), msg.NodeID)
		if err != nil {
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Open stream with peer node failed"
			log.Logger.Errorf("Open stream with peer node failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rsp)
			return
		}
		stream.SetDeadline(time.Now().Add(p2p.ChatProxyStreamTimeout))
		defer stream.Close()
		log.Logger.Infof("Create libp2p stream with %s success", msg.NodeID)

		r.Body, r.ContentLength, err = msg.ChatModelRequest.RequestBody()
		if err != nil {
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Copy http request body failed"
			log.Logger.Errorf("Copy http request body failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rsp)
			return
		}
		queryValues := r.URL.Query()
		queryValues.Add("project", msg.Project)
		queryValues.Add("model", msg.Model)
		r.URL.RawQuery = queryValues.Encode()

		err = r.Write(stream)
		if err != nil {
			stream.Reset()
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Write chat stream failed"
			log.Logger.Errorf("Write chat stream failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rsp)
			return
		}

		log.Logger.Info("Read the response that was send from dest peer")
		buf := bufio.NewReader(stream)
		resp, err := http.ReadResponse(buf, r)
		if err != nil {
			stream.Reset()
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Read chat stream failed"
			log.Logger.Errorf("Read chat stream failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rsp)
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
		resp.Body.Close()
		log.Logger.Info("Handle chat completion stream request over")
		return
	}
	w.Header().Set("Content-Type", "application/json")

	requestID, err := uuid.NewRandom()
	if err != nil {
		rsp.Code = int(types.ErrCodeUUID)
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	ccms := make([]*protocol.ChatCompletionMessage, 0)
	for _, ccm := range msg.Messages {
		ccms = append(ccms, &protocol.ChatCompletionMessage{
			Role:    ccm.Role,
			Content: ccm.Content,
		})
	}

	pi := &protocol.ChatCompletionBody{
		Data: &protocol.ChatCompletionBody_Req{
			Req: &protocol.ChatCompletionRequest{
				Project:  msg.Project,
				Model:    msg.Model,
				Messages: ccms,
				Stream:   false,
				Wallet: &protocol.WalletVerification{
					Wallet:    msg.Wallet,
					Signature: msg.Signature,
					Hash:      msg.Hash,
				},
			},
		},
	}
	body, err := proto.Marshal(pi)
	if err != nil {
		rsp.SetCode(int(types.ErrCodeProtobuf))
		rsp.SetMessage(err.Error())
		json.NewEncoder(w).Encode(rsp)
		return
	}
	body, err = p2p.Encrypt(r.Context(), msg.NodeID, body)
	if err != nil {
		rsp.Code = int(types.ErrCodeEncrypt)
		rsp.Message = types.ErrCodeEncrypt.String()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	req := &protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: p2p.Hio.UserAgent,
			Timestamp:     time.Now().Unix(),
			Id:            requestID.String(),
			NodeId:        config.GC.Identity.PeerID,
			Receiver:      msg.NodeID,
			NodePubKey:    nil,
			Sign:          nil,
		},
		Type:       *protocol.MessageType_CHAT_COMPLETION.Enum(),
		Body:       body,
		ResultCode: 0,
	}
	req.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
	hs.handleRequest(req, &rsp)
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

	urlScheme := "http"
	hp := strings.Split(r.Host, ":")
	if len(hp) > 1 && hp[1] == "443" {
		urlScheme = "https"
	}
	var err error = nil
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
		req := r.Clone(r.Context())
		req.URL.Host = r.Host
		req.URL.Scheme = urlScheme
		req.URL.Path = "/api/v0/chat/completion"
		req.Body, req.ContentLength, err = chatReq.RequestBody()
		if err != nil {
			log.Logger.Warnf("Make chat completion proxy request body %v in %d time", err, failed_count)
			continue
		}
		resp, err := http.DefaultTransport.RoundTrip(req)
		if err != nil || resp.StatusCode != 200 {
			log.Logger.Warnf("Roundtrip chat completion proxy %v %v to %s in %d time", err, resp.StatusCode, peer.NodeID, failed_count)
			failed_count += 1
			continue
		}
		defer resp.Body.Close()
		if msg.Stream {
			contentType := resp.Header.Get("Content-Type")
			if contentType == "application/json" {
				body, _ := io.ReadAll(resp.Body)
				log.Logger.Warnf("Transform chat completion proxy %v to %s in %d time", string(body), peer.NodeID, failed_count)
				failed_count += 1
				continue
			} else {
				for k, v := range resp.Header {
					for _, s := range v {
						w.Header().Add(k, s)
					}
				}
				w.WriteHeader(resp.StatusCode)
				io.Copy(w, resp.Body)
				log.Logger.Infof("Handle chat completion proxy stream request to %s success in %d time", peer.NodeID, failed_count)
				return
			}
		} else {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Logger.Warnf("Read chat completion proxy response json failed in %d time", failed_count)
				failed_count += 1
				continue
			}
			if err := json.Unmarshal(body, &rsp); err != nil {
				log.Logger.Warnf("Parse chat completion proxy response json failed in %d time", failed_count)
				failed_count += 1
				continue
			}
			if rsp.Code != 0 {
				log.Logger.Warnf("Handle chat completion proxy response %v in %d time", rsp.Message, failed_count)
				failed_count += 1
				continue
			}
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
			rsp.Message = "Copy http request body failed"
			log.Logger.Errorf("Copy http request body failed: %v", err)
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
		responseCh := make(chan *http.Response, 1)
		errorCh := make(chan error, 1)

		go func() {
			select {
			case <-ctx.Done():
				return
			default:
				resp, err := http.ReadResponse(reader, hreq)
				select {
				case <-ctx.Done():
					return
				case responseCh <- resp:
				}
				select {
				case <-ctx.Done():
					return
				case errorCh <- err:
				}
			}
		}()

		var resp *http.Response
		select {
		case <-ctx.Done():
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = fmt.Sprintf("Context canceled or timed out: %v", ctx.Err())
			log.Logger.Errorf("Context canceled or timed out: %v", ctx.Err())
			return
		case resp = <-responseCh:
		}
		select {
		case <-ctx.Done():
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = fmt.Sprintf("Context canceled or timed out: %v", ctx.Err())
			log.Logger.Errorf("Context canceled or timed out: %v", ctx.Err())
			return
		case err = <-errorCh:
		}

		// resp, err := http.ReadResponse(reader, hreq)
		if err != nil {
			stream.Reset()
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Read image gen response from libp2p stream failed"
			log.Logger.Errorf("Read image gen response from libp2p stream failed: %v", err)
			return
		}
		log.Logger.Info("Read image gen response from libp2p stream success")

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			stream.Reset()
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Read image response body failed"
			log.Logger.Errorf("Read image response body failed: %v", err)
			return
		}
		log.Logger.Info("Read image response body success")

		response := types.ImageGenModelResponse{}
		if err := json.Unmarshal(body, &response); err != nil {
			stream.Reset()
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Unmarshal image response from stream error"
			log.Logger.Errorf("Unmarshal image response from stream error: %v", err)
			return
		}
		rsp.Code = response.Code
		rsp.Message = response.Message
		rsp.Data.Created = response.Created
		rsp.Data.Choices = response.Data
		log.Logger.Info("Handle image gen stream request over")
		return
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
