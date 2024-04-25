package serve

import (
	"encoding/json"
	"net/http"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/host"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/p2p"
	"AIComputingNode/pkg/protocol"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type httpService struct {
	publishChan chan<- []byte
}

func generateUniqueID() string {
	return uuid.New().String()
}

func httpStatus(code int) int {
	switch code {
	case 0:
		return http.StatusOK
	case ErrCodeParam, ErrCodeParse:
		return http.StatusBadRequest
	case ErrCodeProtobuf, ErrCodeTimeout, ErrCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (hs *httpService) idHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	id := p2p.Hio.GetIdentifyProtocol()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(id)
}

func (hs *httpService) peersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	rsp := PeerListResponse{
		Code:    0,
		Message: "ok",
	}
	peerChan, err := p2p.Hio.FindPeers(config.GC.App.TopicName)
	if err != nil {
		log.Logger.Warnf("List peer message: %v", err)
		rsp.Code = ErrCodeRendezvous
		rsp.Message = err.Error()
	} else {
		for peer := range peerChan {
			// rsp.List = append(rsp.List, peer.String())
			rsp.Data = append(rsp.Data, peer.ID.String())
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rsp)
}

func (hs *httpService) handleRequest(w http.ResponseWriter, r *http.Request, req *protocol.Message, rsp Response) {
	requestID := req.Header.Id
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		rsp.SetCode(ErrCodeProtobuf)
		rsp.SetMessage(err.Error())
		json.NewEncoder(w).Encode(rsp)
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

	hs.publishChan <- reqBytes

	select {
	case notifyData := <-notifyChan:
		json.Unmarshal(notifyData, &rsp)
	case <-time.After(2 * time.Minute):
		log.Logger.Warnf("request id %s message type %s timeout", requestID, req.Type)
		rsp.SetCode(ErrCodeTimeout)
		rsp.SetMessage(errMsg[ErrCodeTimeout])
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
	json.NewEncoder(w).Encode(rsp)
}

func (hs *httpService) peerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	rsp := PeerResponse{
		Code:    0,
		Message: "ok",
	}
	w.Header().Set("Content-Type", "application/json")

	var msg PeerRequest
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		rsp.Code = ErrCodeParse
		rsp.Message = errMsg[rsp.Code]
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = ErrCodeParam
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if msg.NodeID == config.GC.Identity.PeerID {
		rsp.Data = p2p.Hio.GetIdentifyProtocol()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	pi := &protocol.PeerIdentityBody{
		Data: &protocol.PeerIdentityBody_Req{
			Req: &protocol.PeerIdentityRequest{
				NodeId: msg.NodeID,
			},
		},
	}
	body, err := proto.Marshal(pi)
	if err != nil {
		rsp.SetCode(ErrCodeProtobuf)
		rsp.SetMessage(err.Error())
		json.NewEncoder(w).Encode(rsp)
		return
	}
	body, err = p2p.Encrypt(msg.NodeID, body)

	requestID := generateUniqueID()
	req := &protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: p2p.Hio.UserAgent,
			Timestamp:     time.Now().Unix(),
			Id:            requestID,
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
	hs.handleRequest(w, r, req, &rsp)
}

func (hs *httpService) imageGenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	rsp := ImageGenerationResponse{
		Code:    0,
		Message: "ok",
	}
	w.Header().Set("Content-Type", "application/json")

	var msg ImageGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		rsp.Code = ErrCodeParse
		rsp.Message = errMsg[rsp.Code]
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = ErrCodeParam
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if msg.NodeID == config.GC.Identity.PeerID {
		rsp.Code = ErrCodeParam
		rsp.Message = "Cannot be sent to the node itself"
		json.NewEncoder(w).Encode(rsp)
		return
	}

	pi := &protocol.ImageGenerationBody{
		Data: &protocol.ImageGenerationBody_Req{
			Req: &protocol.ImageGenerationRequest{
				NodeId:     msg.NodeID,
				Model:      msg.Model,
				PromptWord: msg.PromptWord,
			},
		},
	}
	body, err := proto.Marshal(pi)
	if err != nil {
		rsp.SetCode(ErrCodeProtobuf)
		rsp.SetMessage(err.Error())
		json.NewEncoder(w).Encode(rsp)
		return
	}
	body, err = p2p.Encrypt(msg.NodeID, body)

	requestID := generateUniqueID()
	req := &protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: p2p.Hio.UserAgent,
			Timestamp:     time.Now().Unix(),
			Id:            requestID,
			NodeId:        config.GC.Identity.PeerID,
			Receiver:      msg.NodeID,
			NodePubKey:    nil,
			Sign:          nil,
		},
		Type:       *protocol.MessageType_IMAGE_GENERATION.Enum(),
		Body:       body,
		ResultCode: 0,
	}
	if err == nil {
		req.Header.NodePubKey, _ = p2p.MarshalPubKeyFromPrivKey(p2p.Hio.PrivKey)
	}
	hs.handleRequest(w, r, req, &rsp)
}

func (hs *httpService) hostInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	rsp := HostInfoResponse{
		Code:    0,
		Message: "ok",
	}
	w.Header().Set("Content-Type", "application/json")

	var msg HostInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		rsp.Code = ErrCodeParse
		rsp.Message = errMsg[rsp.Code]
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := msg.Validate(); err != nil {
		rsp.Code = ErrCodeParam
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if msg.NodeID == config.GC.Identity.PeerID {
		hd, err := host.GetHostInfo()
		if err != nil {
			rsp.Code = ErrCodeHostInfo
			rsp.Message = err.Error()
		} else {
			rsp.Data = *hd
		}
		json.NewEncoder(w).Encode(rsp)
		return
	}

	pi := &protocol.HostInfoBody{
		Data: &protocol.HostInfoBody_Req{
			Req: &protocol.HostInfoRequest{
				NodeId: msg.NodeID,
			},
		},
	}
	body, err := proto.Marshal(pi)
	if err != nil {
		rsp.SetCode(ErrCodeProtobuf)
		rsp.SetMessage(err.Error())
		json.NewEncoder(w).Encode(rsp)
		return
	}
	body, err = p2p.Encrypt(msg.NodeID, body)

	requestID := generateUniqueID()
	req := &protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: p2p.Hio.UserAgent,
			Timestamp:     time.Now().Unix(),
			Id:            requestID,
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
	hs.handleRequest(w, r, req, &rsp)
}

func (hs *httpService) swarmPeersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	pinfos := p2p.Hio.SwarmPeers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pinfos)
}

func (hs *httpService) swarmAddrsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	pinfos := p2p.Hio.SwarmAddrs()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pinfos)
}

func (hs *httpService) swarmConnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	rsp := SwarmConnectResponse{
		Code:    0,
		Message: "ok",
	}
	w.Header().Set("Content-Type", "application/json")

	var req SwarmConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rsp.Code = ErrCodeParse
		rsp.Message = errMsg[rsp.Code]
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := req.Validate(); err != nil {
		rsp.Code = ErrCodeParam
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := p2p.Hio.SwarmConnect(req.NodeAddr); err != nil {
		rsp.Code = ErrCodeInternal
		rsp.Message = err.Error()
	}
	json.NewEncoder(w).Encode(rsp)
}

func (hs *httpService) swarmDisconnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	rsp := SwarmConnectResponse{
		Code:    0,
		Message: "ok",
	}
	w.Header().Set("Content-Type", "application/json")

	var req SwarmConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rsp.Code = ErrCodeParse
		rsp.Message = errMsg[rsp.Code]
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := req.Validate(); err != nil {
		rsp.Code = ErrCodeParam
		rsp.Message = err.Error()
		json.NewEncoder(w).Encode(rsp)
		return
	}

	if err := p2p.Hio.SwarmDisconnect(req.NodeAddr); err != nil {
		rsp.Code = ErrCodeInternal
		rsp.Message = err.Error()
	}
	json.NewEncoder(w).Encode(rsp)
}

func (hs *httpService) pubsubPeersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}
	rsp := p2p.Hio.PubsubPeers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rsp)
}

func NewHttpServe(pcn chan<- []byte) {
	hs := &httpService{
		publishChan: pcn,
	}
	http.HandleFunc("/api/v0/id", hs.idHandler)
	// http.HandleFunc("/api/v0/echo", func(w http.ResponseWriter, r *http.Request) {
	// 	var msg EchoMessage
	// 	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 		return
	// 	}
	// 	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	// 	// defer cancel()
	// 	err := hs.Topic.Publish(hs.Ctx, []byte(msg.Content))
	// 	rsp := EchoResponse{
	// 		Code:    0,
	// 		Message: "ok",
	// 	}
	// 	if err != nil {
	// 		log.Logger.Warnf("Publish message: %v", err)
	// 		rsp.Code = 1
	// 		rsp.Message = err.Error()
	// 	}
	// 	w.Header().Set("Content-Type", "application/json")
	// 	json.NewEncoder(w).Encode(rsp)
	// })
	http.HandleFunc("/api/v0/peers", hs.peersHandler)
	http.HandleFunc("/api/v0/peer", hs.peerHandler)
	http.HandleFunc("/api/v0/image/gen", hs.imageGenHandler)
	http.HandleFunc("/api/v0/host/info", hs.hostInfoHandler)
	http.HandleFunc("/api/v0/swarm/peers", hs.swarmPeersHandler)
	http.HandleFunc("/api/v0/swarm/addrs", hs.swarmAddrsHandler)
	http.HandleFunc("/api/v0/swarm/connect", hs.swarmConnectHandler)
	http.HandleFunc("/api/v0/swarm/disconnect", hs.swarmDisconnectHandler)
	http.HandleFunc("/api/v0/pubsub/peers", hs.pubsubPeersHandler)
	log.Logger.Info("HTTP server is running on http://", config.GC.API.Addr)
	if err := http.ListenAndServe(config.GC.API.Addr, nil); err != nil {
		log.Logger.Fatalf("Start HTTP Server: %v", err)
	}
}
