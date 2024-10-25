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
				c.Writer.Header().Set(k, s)
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
	status, code, message := handleRequest(publishChan, req, &rsp)
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

func handleImageGenRequest(ctx context.Context, publishChan chan<- []byte, req types.ImageGenerationRequest, rsp *types.ImageGenerationResponse) (int, int, string) {
	if req.NodeID == config.GC.Identity.PeerID {
		modelAPI := config.GC.GetModelAPI(req.Project, req.Model)
		if modelAPI == "" {
			return http.StatusInternalServerError, int(types.ErrCodeModel), "Model API configuration is empty"
		}
		*rsp = *model.ImageGenerationModel(modelAPI, req.ImageGenModelRequest)
		return http.StatusOK, rsp.Code, rsp.Message
	}

	if req.ResponseFormat == "b64_json" {
		log.Logger.Info("Received image gen b64_json request")
		stream, err := p2p.Hio.NewStream(ctx, req.NodeID)
		if err != nil {
			// rsp.Code = int(types.ErrCodeStream)
			// rsp.Message = "Open stream with peer node failed"
			log.Logger.Errorf("Open stream with peer node failed: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeStream), "Open stream with peer node failed"
		}
		stream.SetDeadline(time.Now().Add(p2p.ChatProxyStreamTimeout))
		defer stream.Close()
		log.Logger.Infof("Create libp2p stream with %s success", req.NodeID)

		jsonData, err := json.Marshal(req.ImageGenModelRequest)
		if err != nil {
			// rsp.Code = int(types.ErrCodeStream)
			// rsp.Message = "Copy http request body failed"
			log.Logger.Errorf("Copy http request body failed: %v", err)
			return http.StatusInternalServerError, int(types.ErrCodeStream), "Copy http request body failed"
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
		hreq.URL.RawQuery = queryValues.Encode()

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
		return http.StatusInternalServerError, int(types.ErrCodeProtobuf), err.Error()
	}
	body, err = p2p.Encrypt(ctx, req.NodeID, body)
	if err != nil {
		return http.StatusInternalServerError, int(types.ErrCodeEncrypt), types.ErrCodeEncrypt.String()
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
	return handleRequest(publishChan, msg, rsp)
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), requestProcessTimeout)
	defer cancel()
	status, code, message := handleImageGenRequest(ctx, publishChan, msg, &rsp)
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
		c.JSON(http.StatusInternalServerError, rsp)
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), requestProcessTimeout)
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
		status, code, message := handleImageGenRequest(ctx, publishChan, igReq, &rsp)
		if code != 0 {
			log.Logger.Warnf("Roundtrip image gen proxy %v %v %v to %s in %d time", status, code, message, peer.NodeID, failed_count)
			failed_count += 1
			continue
		} else if rsp.Code != 0 {
			log.Logger.Warnf("Roundtrip image gen proxy %v %v to %s in %d time", code, message, peer.NodeID, failed_count)
			failed_count += 1
			continue
		} else {
			c.JSON(http.StatusOK, rsp)
			return
		}
	}

	rsp.Code = int(types.ErrCodeProxy)
	rsp.Message = fmt.Sprintf("Failed %d times", failed_count)
	c.JSON(http.StatusInternalServerError, rsp)
}
