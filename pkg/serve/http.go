package serve

import (
	"encoding/json"
	"net/http"
	"time"

	"AIComputingNode/pkg/config"
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

func (hs *httpService) peerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	requestID := generateUniqueID()
	req := &protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: p2p.Hio.UserAgent,
			Timestamp:     time.Now().Unix(),
			Id:            requestID,
			NodeId:        config.GC.Identity.PeerID,
			NodePubKey:    []byte(""),
			Sign:          []byte(""),
		},
		Type: *protocol.MesasgeType_PEER_IDENTITY_REQUEST.Enum(),
		Body: &protocol.Message_PiReq{
			PiReq: &protocol.PeerIdentityRequest{
				NodeId: msg.NodeID,
			},
		},
	}
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		rsp.Code = ErrCodeProtobuf
		rsp.Message = err.Error()
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
		json.Unmarshal(notifyData, &rsp.Data)
	case <-time.After(2 * time.Minute):
		log.Logger.Warn("request id ", requestID, " message type PEER_IDENTITY_REQUEST timeout")
		rsp.Code = ErrCodeTimeout
		rsp.Message = errMsg[rsp.Code]
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

	requestID := generateUniqueID()
	req := &protocol.Message{
		Header: &protocol.MessageHeader{
			ClientVersion: p2p.Hio.UserAgent,
			Timestamp:     time.Now().Unix(),
			Id:            requestID,
			NodeId:        config.GC.Identity.PeerID,
			NodePubKey:    []byte(""),
			Sign:          []byte(""),
		},
		Type: *protocol.MesasgeType_IMAGE_GENERATION_REQUEST.Enum(),
		Body: &protocol.Message_IgReq{
			IgReq: &protocol.ImageGenerationRequest{
				NodeId:     msg.NodeID,
				Model:      msg.Model,
				PromptWord: msg.PromptWord,
			},
		},
	}
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		rsp.Code = ErrCodeProtobuf
		rsp.Message = err.Error()
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
		log.Logger.Warn("request id ", requestID, " message type ", req.Type.String(), " timeout")
		rsp.Code = ErrCodeTimeout
		rsp.Message = errMsg[rsp.Code]
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
	log.Logger.Info("HTTP server is running on http://", config.GC.API.Addr)
	if err := http.ListenAndServe(config.GC.API.Addr, nil); err != nil {
		log.Logger.Fatalf("Start HTTP Server: %v", err)
	}
}
