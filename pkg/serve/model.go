package serve

import (
	"bufio"
	"bytes"
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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

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

	if msg.NodeID == config.GC.Identity.PeerID {
		modelAPI := config.GC.GetModelAPI(msg.Project, msg.Model)
		if msg.Stream {
			log.Logger.Info("Received chat completion stream request from the node itself")
			if modelAPI == "" {
				rsp.Code = int(types.ErrCodeModel)
				rsp.Message = "Model API configuration is empty"
				c.JSON(http.StatusInternalServerError, rsp)
				return
			}
			req := new(http.Request)
			*req = *c.Request
			var err error = nil
			req.URL, err = url.Parse(modelAPI)
			if err != nil {
				rsp.Code = int(types.ErrCodeModel)
				rsp.Message = "Parse model api interface failed"
				c.JSON(http.StatusInternalServerError, rsp)
				return
			}
			req.Host = req.URL.Host
			req.Body, req.ContentLength, err = msg.ChatModelRequest.RequestBody()
			if err != nil {
				rsp.Code = int(types.ErrCodeModel)
				rsp.Message = "Copy http request body failed"
				c.JSON(http.StatusInternalServerError, rsp)
				return
			}
			log.Logger.Infof("Making request to %s\n", req.URL)
			resp, err := http.DefaultTransport.RoundTrip(req)
			if err != nil {
				rsp.Code = int(types.ErrCodeModel)
				rsp.Message = "RoundTrip chat request failed"
				log.Logger.Errorf("RoundTrip chat request failed: %v", err)
				c.JSON(http.StatusInternalServerError, rsp)
				return
			}

			for k, v := range resp.Header {
				for _, s := range v {
					c.Writer.Header().Add(k, s)
				}
			}

			c.Writer.WriteHeader(resp.StatusCode)

			log.Logger.Info("Copy roundtrip response")
			io.Copy(c.Writer, resp.Body)
			resp.Body.Close()
			log.Logger.Info("Handle chat completion stream request over from the node itself")
			return
		} else {
			rsp = *model.ChatModel(modelAPI, msg.ChatModelRequest)
			c.JSON(http.StatusOK, rsp)
			return
		}
	}

	if msg.Stream {
		log.Logger.Info("Received chat completion stream request")
		stream, err := p2p.Hio.NewStream(c.Request.Context(), msg.NodeID)
		if err != nil {
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Open stream with peer node failed"
			log.Logger.Errorf("Open stream with peer node failed: %v", err)
			c.JSON(http.StatusInternalServerError, rsp)
			return
		}
		stream.SetDeadline(time.Now().Add(p2p.ChatProxyStreamTimeout))
		defer stream.Close()
		log.Logger.Infof("Create libp2p stream with %s success", msg.NodeID)

		url := new(url.URL)
		*url = *c.Request.URL
		queryValues := url.Query()
		queryValues.Add("project", msg.Project)
		queryValues.Add("model", msg.Model)
		url.RawQuery = queryValues.Encode()

		jsonData, err := json.Marshal(msg.ChatModelRequest)
		if err != nil {
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Copy http request body failed"
			log.Logger.Errorf("Copy http request body failed: %v", err)
			c.JSON(http.StatusInternalServerError, rsp)
			return
		}

		newReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", url.String(), bytes.NewBuffer(jsonData))
		if err != nil {
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "New request failed"
			log.Logger.Errorf("New request failed: %v", err)
			c.JSON(http.StatusInternalServerError, rsp)
			return
		}

		err = newReq.Write(stream)
		if err != nil {
			stream.Reset()
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Write chat stream failed"
			log.Logger.Errorf("Write chat stream failed: %v", err)
			c.JSON(http.StatusInternalServerError, rsp)
			return
		}

		log.Logger.Info("Read the response that was send from dest peer")
		buf := bufio.NewReader(stream)
		resp, err := http.ReadResponse(buf, newReq)
		if err != nil {
			stream.Reset()
			rsp.Code = int(types.ErrCodeStream)
			rsp.Message = "Read chat stream failed"
			log.Logger.Errorf("Read chat stream failed: %v", err)
			c.JSON(http.StatusInternalServerError, rsp)
			return
		}

		for k, v := range resp.Header {
			for _, s := range v {
				c.Writer.Header().Add(k, s)
			}
		}

		c.Writer.WriteHeader(resp.StatusCode)

		log.Logger.Info("Copy the body from libp2p stream")
		io.Copy(c.Writer, resp.Body)
		resp.Body.Close()
		log.Logger.Info("Handle chat completion stream request over")
		return
	}

	requestID, err := uuid.NewRandom()
	if err != nil {
		rsp.Code = int(types.ErrCodeUUID)
		rsp.Message = err.Error()
		c.JSON(http.StatusInternalServerError, rsp)
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
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}
	body, err = p2p.Encrypt(c.Request.Context(), msg.NodeID, body)
	if err != nil {
		rsp.Code = int(types.ErrCodeEncrypt)
		rsp.Message = types.ErrCodeEncrypt.String()
		c.JSON(http.StatusInternalServerError, rsp)
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
	handleRequest(c, publishChan, req, &rsp)
}

func ChatCompletionProxyHandler(c *gin.Context) {
	rsp := types.ChatCompletionResponse{}

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
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Latency < peers[j].Latency
	})

	urlScheme := "http"
	hp := strings.Split(c.Request.Host, ":")
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
		req := c.Request.Clone(c.Request.Context())
		req.URL.Host = c.Request.Host
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
						c.Writer.Header().Add(k, s)
					}
				}
				c.Writer.WriteHeader(resp.StatusCode)
				io.Copy(c.Writer, resp.Body)
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
			c.JSON(http.StatusOK, rsp)
			return
		}
	}

	rsp.Code = int(types.ErrCodeProxy)
	rsp.Message = fmt.Sprintf("Failed %d times", failed_count)
	c.JSON(http.StatusInternalServerError, rsp)
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

	if msg.NodeID == config.GC.Identity.PeerID {
		modelAPI := config.GC.GetModelAPI(msg.Project, msg.Model)
		if modelAPI == "" {
			rsp.Code = int(types.ErrCodeModel)
			rsp.Message = "Model API configuration is empty"
			c.JSON(http.StatusInternalServerError, rsp)
			return
		}
		rsp = *model.ImageGenerationModel(modelAPI, msg.ImageGenModelRequest)
		c.JSON(http.StatusOK, rsp)
		return
	}

	requestID, err := uuid.NewRandom()
	if err != nil {
		rsp.Code = int(types.ErrCodeUUID)
		rsp.Message = err.Error()
		c.JSON(http.StatusInternalServerError, rsp)
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
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}
	body, err = p2p.Encrypt(c.Request.Context(), msg.NodeID, body)
	if err != nil {
		rsp.Code = int(types.ErrCodeEncrypt)
		rsp.Message = types.ErrCodeEncrypt.String()
		c.JSON(http.StatusInternalServerError, rsp)
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
	handleRequest(c, publishChan, req, &rsp)
}

func ImageGenProxyHandler(c *gin.Context) {
	rsp := types.ImageGenerationResponse{}

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
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Latency < peers[j].Latency
	})

	urlScheme := "http"
	hp := strings.Split(c.Request.Host, ":")
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
		req := c.Request.Clone(c.Request.Context())
		req.URL.Host = c.Request.Host
		req.URL.Scheme = urlScheme
		req.URL.Path = "/api/v0/image/gen"
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
		c.JSON(http.StatusOK, rsp)
		return
		// }
	}

	rsp.Code = int(types.ErrCodeProxy)
	rsp.Message = fmt.Sprintf("Failed %d times", failed_count)
	c.JSON(http.StatusInternalServerError, rsp)
}
