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

func NewHttpServe(addr string, publishChan chan<- []byte) {
	http.HandleFunc("/api/v0/id", func(w http.ResponseWriter, r *http.Request) {
		id := p2p.Hio.GetIdentifyProtocol()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(id)
	})
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
	http.HandleFunc("/api/v0/peers", func(w http.ResponseWriter, r *http.Request) {
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
	})
	http.HandleFunc("/api/v0/peer", func(w http.ResponseWriter, r *http.Request) {
		var msg PeerRequest
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		rsp := PeerResponse{
			Code:    0,
			Message: "ok",
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

		publishChan <- reqBytes

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
	})
	// 启动 HTTP 服务器
	log.Logger.Info("HTTP server is running on http://localhost", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Logger.Fatalf("Start HTTP Server: %v", err)
	}
}
