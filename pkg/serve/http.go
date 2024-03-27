package serve

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/SuperImageAI/AIComputingNode/pkg/config"
	"github.com/SuperImageAI/AIComputingNode/pkg/log"
	"github.com/SuperImageAI/AIComputingNode/pkg/p2p"
	"github.com/SuperImageAI/AIComputingNode/pkg/protocol"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type RequestItem struct {
	ID     string
	Notify chan []byte
}

var RequestQueue = make([]RequestItem, 0)
var QueueLock = sync.Mutex{}

type EchoMessage struct {
	Content string `json:"content"`
}

type EchoResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PeerListResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	List    []string `json:"list"`
}

type PeerRequest struct {
	NodeID string `json:"node_id"`
}

type PeerResponse struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    p2p.IdentifyProtocol `json:"data"`
}

func generateUniqueID() string {
	return uuid.New().String()
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
			rsp.Code = 1
			rsp.Message = err.Error()
		} else {
			for peer := range peerChan {
				// rsp.List = append(rsp.List, peer.String())
				rsp.List = append(rsp.List, peer.ID.String())
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

		rsp := PeerResponse{
			Code:    0,
			Message: "ok",
		}

		if msg.NodeID == config.GC.Identity.PeerID {
			rsp.Data = p2p.Hio.GetIdentifyProtocol()
			w.Header().Set("Content-Type", "application/json")
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
			rsp.Code = 1
			rsp.Message = err.Error()
			w.Header().Set("Content-Type", "application/json")
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
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rsp)
		case <-time.After(2 * time.Minute):
			http.Error(w, "Request timeout", http.StatusRequestTimeout)
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
	})
	// 启动 HTTP 服务器
	log.Logger.Info("HTTP server is running on http://localhost", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Logger.Fatalf("Start HTTP Server: %v", err)
	}
}
