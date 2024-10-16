package serve

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/host"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/p2p"
	"AIComputingNode/pkg/protocol"
	"AIComputingNode/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type httpService struct {
	publishChan chan<- []byte
	configPath  string
}

var httpServer *http.Server

func httpStatus(code types.ErrorCode) int {
	switch code {
	case 0:
		return http.StatusOK
	case types.ErrCodeParam, types.ErrCodeParse:
		return http.StatusBadRequest
	case types.ErrCodeProtobuf, types.ErrCodeTimeout, types.ErrCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func IdHandler(c *gin.Context) {
	id := p2p.Hio.GetIdentifyProtocol()
	c.JSON(http.StatusOK, id)
}

func PeersHandler(c *gin.Context) {
	rsp := types.PeerListResponse{}
	rsp.Data, rsp.Code = db.FindPeers(100)
	if rsp.Code != 0 {
		rsp.Message = types.ErrorCode(rsp.Code).String()
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}
	c.JSON(http.StatusOK, rsp)
}

func handleRequest(c *gin.Context, publishChan chan<- []byte, req *protocol.Message, rsp any) {
	requestID := req.Header.Id
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.BaseHttpResponse{
			Code:    int(types.ErrCodeProtobuf),
			Message: err.Error(),
		})
		return
	}

	notifyChan := make(chan []byte, 1024)
	requestItem := RequestItem{
		ID:     requestID,
		Notify: notifyChan,
	}
	QueueLock.Lock()
	RequestQueue = append(RequestQueue, requestItem)
	QueueLock.Unlock()

	publishChan <- reqBytes

	select {
	case notifyData, ok := <-notifyChan:
		if ok {
			if err := json.Unmarshal(notifyData, &rsp); err != nil {
				c.JSON(http.StatusInternalServerError, types.BaseHttpResponse{
					Code:    int(types.ErrCodeParse),
					Message: "parse pubsub reponse error",
				})
			} else {
				c.JSON(http.StatusOK, rsp)
			}
		} else {
			c.JSON(http.StatusInternalServerError, types.BaseHttpResponse{
				Code:    int(types.ErrCodeInternal),
				Message: "pubsub channel error",
			})
		}
	case <-time.After(2 * time.Minute):
		log.Logger.Warnf("request id %s message type %s timeout", requestID, req.Type)
		c.JSON(http.StatusGatewayTimeout, types.BaseHttpResponse{
			Code:    int(types.ErrCodeTimeout),
			Message: types.ErrCodeTimeout.String(),
		})
		QueueLock.Lock()
		for i, item := range RequestQueue {
			if item.ID == requestID {
				RequestQueue = append(RequestQueue[:i], RequestQueue[i+1:]...)
				break
			}
		}
		QueueLock.Unlock()
		close(notifyChan)
	}
}

func PeerHandler(c *gin.Context, publishChan chan<- []byte) {
	// if r.Method != http.MethodPost {
	// 	http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
	// 	return
	// }
	rsp := types.PeerResponse{}

	var msg types.PeerRequest
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
		rsp.IdentifyProtocol = p2p.Hio.GetIdentifyProtocol()
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

	pi := &protocol.PeerIdentityBody{
		Data: &protocol.PeerIdentityBody_Req{
			Req: &protocol.PeerIdentityRequest{},
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
		Type:       *protocol.MessageType_PEER_IDENTITY.Enum(),
		Body:       body,
		ResultCode: 0,
	}
	if err == nil {
		req.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
	}
	handleRequest(c, publishChan, req, &rsp)
}

func HostInfoHandler(c *gin.Context, publishChan chan<- []byte) {
	rsp := types.HostInfoResponse{}

	var msg types.HostInfoRequest
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
		hd, err := host.GetHostInfo()
		if err != nil {
			rsp.Code = int(types.ErrCodeHostInfo)
			rsp.Message = err.Error()
			c.JSON(http.StatusInternalServerError, rsp)
			return
		}
		rsp.HostInfo = *hd
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

	pi := &protocol.HostInfoBody{
		Data: &protocol.HostInfoBody_Req{
			Req: &protocol.HostInfoRequest{},
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
		Type:       *protocol.MessageType_HOST_INFO.Enum(),
		Body:       body,
		ResultCode: 0,
	}
	if err == nil {
		req.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
	}
	handleRequest(c, publishChan, req, &rsp)
}

func RendezvousPeersHandler(c *gin.Context) {
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	rsp := types.PeerListResponse{}

	peerChan, err := p2p.Hio.FindPeers(ctx, config.GC.App.TopicName)
	if err != nil {
		log.Logger.Warnf("List peer message: %v", err)
		rsp.Code = int(types.ErrCodeRendezvous)
		rsp.Message = err.Error()
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}

	for peer := range peerChan {
		// rsp.List = append(rsp.List, peer.String())
		rsp.Data = append(rsp.Data, peer.ID.String())
	}
	c.JSON(http.StatusOK, rsp)
}

func SwarmPeersHandler(c *gin.Context) {
	pinfos := p2p.Hio.SwarmPeers()
	c.JSON(http.StatusOK, pinfos)
}

func SwarmAddrsHandler(c *gin.Context) {
	pinfos := p2p.Hio.SwarmAddrs()
	c.JSON(http.StatusOK, pinfos)
}

func SwarmConnectHandler(c *gin.Context) {
	rsp := types.SwarmConnectResponse{
		Code:    0,
		Message: "ok",
	}

	var req types.SwarmConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	if err := req.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	if err := p2p.Hio.SwarmConnect(ctx, req.NodeAddr); err != nil {
		rsp.Code = int(types.ErrCodeInternal)
		rsp.Message = err.Error()
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}
	c.JSON(http.StatusOK, rsp)
}

func SwarmDisconnectHandler(c *gin.Context) {
	rsp := types.SwarmConnectResponse{
		Code:    0,
		Message: "ok",
	}

	var req types.SwarmConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	if err := req.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	if err := p2p.Hio.SwarmDisconnect(req.NodeAddr); err != nil {
		rsp.Code = int(types.ErrCodeInternal)
		rsp.Message = err.Error()
		c.JSON(http.StatusInternalServerError, rsp)
		return
	}
	c.JSON(http.StatusOK, rsp)
}

func PubsubPeersHandler(c *gin.Context) {
	rsp := p2p.Hio.PubsubPeers()
	c.JSON(http.StatusOK, rsp)
}

/*
func (hs *httpService) registerAIProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	rsp := types.BaseHttpResponse{
		Code:    0,
		Message: "ok",
	}
	w.Header().Set("Content-Type", "application/json")

	var req types.AIProject
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := req.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	for _, model := range req.Models {
		if err := model.Validate(); err != nil {
			rsp.Code = int(types.ErrCodeParam)
			rsp.Message = err.Error()
			json.NewEncoder(w).Encode(rsp)
			return
		}
	}

	var find bool = false
	for i := range config.GC.AIProjects {
		if config.GC.AIProjects[i].Project == req.Project {
			config.GC.AIProjects[i].Models = req.Models
			find = true
		}
	}
	if !find {
		config.GC.AIProjects = append(config.GC.AIProjects, req)
	}

	if err := config.GC.SaveConfig(hs.configPath); err != nil {
		rsp.Code = int(types.ErrCodeInternal)
		rsp.Message = fmt.Sprintf("config save err %v", err)
	}
	json.NewEncoder(w).Encode(rsp)
	timer.AIT.SendAIProjects()
}

func (hs *httpService) unregisterAIProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	rsp := types.BaseHttpResponse{
		Code:    0,
		Message: "ok",
	}
	w.Header().Set("Content-Type", "application/json")

	var req types.AIProject
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := req.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	var find bool = false
	for i := range config.GC.AIProjects {
		if config.GC.AIProjects[i].Project == req.Project {
			config.GC.AIProjects = append(config.GC.AIProjects[:i], config.GC.AIProjects[i+1:]...)
			find = true
		}
	}
	if !find {
		rsp.Message = "not existed"
	}

	if err := config.GC.SaveConfig(hs.configPath); err != nil {
		rsp.Code = int(types.ErrCodeInternal)
		rsp.Message = fmt.Sprintf("config save err %v", err)
	}
	json.NewEncoder(w).Encode(rsp)
	timer.AIT.SendAIProjects()
}

func (hs *httpService) getAIProjectOfNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	rsp := types.AIProjectListResponse{
		Code:    0,
		Message: "ok",
	}
	w.Header().Set("Content-Type", "application/json")

	var msg types.AIProjectListRequest
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
		rsp.Data = config.GC.GetAIProjectsOfNode()
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

	pbody := &protocol.AIProjectBody{
		Data: &protocol.AIProjectBody_Req{
			Req: &protocol.AIProjectRequest{},
		},
	}
	body, err := proto.Marshal(pbody)
	if err != nil {
		rsp.SetCode(int(types.ErrCodeProtobuf))
		rsp.SetMessage(err.Error())
		json.NewEncoder(w).Encode(rsp)
		return
	}
	body, err = p2p.Encrypt(r.Context(), msg.NodeID, body)

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
		Type:       *protocol.MessageType_AI_PROJECT.Enum(),
		Body:       body,
		ResultCode: 0,
	}
	if err == nil {
		req.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
	}
	hs.handleRequest(w, r, req, &rsp)
}

func (hs *httpService) listAIProjectsHandler(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodGet {
	// 	http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
	// 	return
	// }
	rsp := types.PeerListResponse{
		Code:    0,
		Message: "ok",
	}
	rsp.Data, rsp.Code = db.ListAIProjects(100)
	if rsp.Code != 0 {
		rsp.Message = types.ErrorCode(rsp.Code).String()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rsp)
}

func (hs *httpService) getModelsOfAIProjectHandler(w http.ResponseWriter, r *http.Request) {
	rsp := types.PeerListResponse{
		Code:    0,
		Message: "ok",
	}
	w.Header().Set("Content-Type", "application/json")

	var req types.GetModelsOfAIProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := req.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	rsp.Data, rsp.Code = db.GetModelsOfAIProjects(req.Project, 100)
	if rsp.Code != 0 {
		rsp.Message = types.ErrorCode(rsp.Code).String()
	}
	json.NewEncoder(w).Encode(rsp)
}

func (hs *httpService) getPeersOfAIProjectHandler(w http.ResponseWriter, r *http.Request) {
	rsp := types.GetPeersOfAIProjectResponse{
		Code:    0,
		Message: "ok",
		Data:    make([]types.AIProjectPeerInfo, 0),
	}
	w.Header().Set("Content-Type", "application/json")

	var req types.GetPeersOfAIProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rsp.Code = int(types.ErrCodeParse)
		rsp.Message = types.ErrCodeParse.String()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := req.Validate(); err != nil {
		rsp.Code = int(types.ErrCodeParam)
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	number := 20
	if r.URL.Query().Has("number") {
		num := r.URL.Query().Get("number")
		if rnum, err := strconv.Atoi(num); err != nil || rnum <= 0 {
			rsp.Code = int(types.ErrCodeParse)
			rsp.Message = types.ErrCodeParse.String()
			json.NewEncoder(w).Encode(rsp)
			return
		} else {
			number = rnum
		}
		if number > 100 {
			number = 100
		}
	}

	ids, code := db.GetPeersOfAIProjects(req.Project, req.Model, number)
	if code != 0 {
		rsp.Code = code
		rsp.Message = types.ErrorCode(code).String()
		json.NewEncoder(w).Encode(rsp)
		return
	}
	for _, id := range ids {
		rsp.Data = append(rsp.Data, types.AIProjectPeerInfo{
			NodeID:       id,
			Connectivity: p2p.Hio.Connectedness(id),
			Latency:      p2p.Hio.Latency(id).Microseconds(),
		})
	}
	json.NewEncoder(w).Encode(rsp)
}
*/
// func NewHttpServe(router *gin.Engine, pcn chan<- []byte, configFilePath string) {
// 	hs := &httpService{
// 		publishChan: pcn,
// 		configPath:  configFilePath,
// 	}
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/api/v0/id", hs.idHandler)
// 	mux.HandleFunc("/api/v0/peers", hs.peersHandler)
// 	mux.HandleFunc("/api/v0/peer", hs.peerHandler)
// 	mux.HandleFunc("/api/v0/host/info", hs.hostInfoHandler)
// 	mux.HandleFunc("/api/v0/chat/completion", hs.chatCompletionHandler)
// 	mux.HandleFunc("/api/v0/chat/completion/proxy", hs.chatCompletionProxyHandler)
// 	mux.HandleFunc("/api/v0/image/gen", hs.imageGenHandler)
// 	mux.HandleFunc("/api/v0/image/gen/proxy", hs.imageGenProxyHandler)
// 	mux.HandleFunc("/api/v0/rendezvous/peers", hs.rendezvousPeersHandler)
// 	mux.HandleFunc("/api/v0/swarm/peers", hs.swarmPeersHandler)
// 	mux.HandleFunc("/api/v0/swarm/addrs", hs.swarmAddrsHandler)
// 	mux.HandleFunc("/api/v0/swarm/connect", hs.swarmConnectHandler)
// 	mux.HandleFunc("/api/v0/swarm/disconnect", hs.swarmDisconnectHandler)
// 	mux.HandleFunc("/api/v0/pubsub/peers", hs.pubsubPeersHandler)
// 	mux.HandleFunc("/api/v0/ai/project/register", hs.registerAIProjectHandler)
// 	mux.HandleFunc("/api/v0/ai/project/unregister", hs.unregisterAIProjectHandler)
// 	mux.HandleFunc("/api/v0/ai/project/peer", hs.getAIProjectOfNodeHandler)
// 	mux.HandleFunc("/api/v0/ai/projects/list", hs.listAIProjectsHandler)
// 	mux.HandleFunc("/api/v0/ai/projects/models", hs.getModelsOfAIProjectHandler)
// 	mux.HandleFunc("/api/v0/ai/projects/peers", hs.getPeersOfAIProjectHandler)

// 	// mux.Handle("/metrics", promhttp.Handler())
// 	mux.Handle("/debug/metrics/prometheus", promhttp.Handler())

// 	// Golang pprof
// 	// mux.HandleFunc("/debug/pprof/", pprof.Index)
// 	// mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
// 	// mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
// 	// mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
// 	// mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
// 	// runtime.SetBlockProfileRate(1)
// 	// runtime.SetMutexProfileFraction(1)

// 	httpServer = &http.Server{
// 		Addr:         config.GC.API.Addr,
// 		Handler:      mux,
// 		ReadTimeout:  20 * time.Second,
// 		WriteTimeout: 90 * time.Second,
// 		IdleTimeout:  90 * time.Second,
// 	}
// 	go func() {
// 		log.Logger.Info("HTTP server is running on http://", httpServer.Addr)
// 		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
// 			log.Logger.Fatalf("Start HTTP Server: %v", err)
// 		}
// 		log.Logger.Info("HTTP server is stopped")
// 	}()
// }

// func StopHttpService() {
// 	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer shutdownRelease()
// 	if err := httpServer.Shutdown(shutdownCtx); err != nil {
// 		log.Logger.Fatalf("Shutdown HTTP Server: %v", err)
// 	} else {
// 		log.Logger.Info("HTTP server is shutdown gracefully")
// 	}
// }
