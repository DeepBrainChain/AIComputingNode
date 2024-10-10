package serve

import (
	"bufio"
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
				w.Header().Add(k, s)
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
	hs.handleRequest(w, r, req, &rsp)
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
		latency := p2p.Hio.Latency(id).Milliseconds()
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

	if msg.NodeID == config.GC.Identity.PeerID {
		modelAPI := config.GC.GetModelAPI(msg.Project, msg.Model)
		if modelAPI == "" {
			rsp.Code = int(types.ErrCodeModel)
			rsp.Message = "Model API configuration is empty"
			json.NewEncoder(w).Encode(rsp)
			return
		}
		rsp = *model.ImageGenerationModel(modelAPI, msg.ImageGenModelRequest)
		json.NewEncoder(w).Encode(rsp)
		return
	}

	requestID, err := uuid.NewRandom()
	if err != nil {
		rsp.Code = int(types.ErrCodeUUID)
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	pi := &protocol.ImageGenerationBody{
		Data: &protocol.ImageGenerationBody_Req{
			Req: &protocol.ImageGenerationRequest{
				Project:        msg.Project,
				Model:          msg.Model,
				Prompt:         msg.Prompt,
				Number:         int32(msg.Number),
				Size:           msg.Size,
				Width:          int32(msg.Width),
				Height:         int32(msg.Height),
				ResponseFormat: msg.ResponseFormat,
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
		Type:       *protocol.MessageType_IMAGE_GENERATION.Enum(),
		Body:       body,
		ResultCode: 0,
	}
	req.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
	hs.handleRequest(w, r, req, &rsp)
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
		latency := p2p.Hio.Latency(id).Milliseconds()
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
		igReq := types.ImageGenerationRequest{
			NodeID:               peer.NodeID,
			Project:              msg.Project,
			ImageGenModelRequest: msg.ImageGenModelRequest,
		}
		req := r.Clone(r.Context())
		req.URL.Host = r.Host
		req.URL.Scheme = urlScheme
		req.URL.Path = "/api/v0/chat/completion"
		req.Body, req.ContentLength, err = igReq.RequestBody()
		if err != nil {
			log.Logger.Warnf("Make image gen proxy request body %v in %d time", err, failed_count)
			continue
		}
		resp, err := http.DefaultTransport.RoundTrip(req)
		if err != nil || resp.StatusCode != 200 {
			log.Logger.Warnf("Roundtrip image gen proxy %v %v to %s in %d time", err, resp.StatusCode, peer.NodeID, failed_count)
			failed_count += 1
			continue
		}
		defer resp.Body.Close()
		// if msg.Stream {
		// 	contentType := resp.Header.Get("Content-Type")
		// 	if contentType == "application/json" {
		// 		body, _ := io.ReadAll(resp.Body)
		// 		log.Logger.Warnf("Transform image gen proxy %v to %s in %d time", string(body), peer.NodeID, failed_count)
		// 		failed_count += 1
		// 		continue
		// 	} else {
		// 		for k, v := range resp.Header {
		// 			for _, s := range v {
		// 				w.Header().Add(k, s)
		// 			}
		// 		}
		// 		w.WriteHeader(resp.StatusCode)
		// 		io.Copy(w, resp.Body)
		// 		log.Logger.Infof("Handle image gen proxy stream request to %s success in %d time", peer.NodeID, failed_count)
		// 		return
		// 	}
		// } else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Logger.Warnf("Read image gen proxy response json failed in %d time", failed_count)
			failed_count += 1
			continue
		}
		if err := json.Unmarshal(body, &rsp); err != nil {
			log.Logger.Warnf("Parse image gen proxy response json failed in %d time", failed_count)
			failed_count += 1
			continue
		}
		if rsp.Code != 0 {
			log.Logger.Warnf("Handle image gen proxy response %v in %d time", rsp.Message, failed_count)
			failed_count += 1
			continue
		}
		json.NewEncoder(w).Encode(rsp)
		return
		// }
	}

	rsp.Code = int(types.ErrCodeProxy)
	rsp.Message = fmt.Sprintf("Failed %d times", failed_count)
	json.NewEncoder(w).Encode(rsp)
}
