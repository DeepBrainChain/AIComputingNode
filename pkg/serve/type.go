package serve

import (
	"AIComputingNode/pkg/p2p"
	"sync"
)

type RequestItem struct {
	ID     string
	Notify chan []byte
}

var RequestQueue = make([]RequestItem, 0)
var QueueLock = sync.Mutex{}

// error code
const (
	ErrCodeParam      = 1001
	ErrCodeParse      = 1002
	ErrCodeProtobuf   = 1003
	ErrCodeTimeout    = 1004
	ErrCodeRendezvous = 1005
	ErrCodeInternal   = 5000
)

// error message
var errMsg = map[int]string{
	ErrCodeParam:      "Parameter error",
	ErrCodeParse:      "Parsing error",
	ErrCodeProtobuf:   "Protobuf serialization error",
	ErrCodeTimeout:    "Processing timeout",
	ErrCodeRendezvous: "Rendezvous error",
	ErrCodeInternal:   "Internal server error",
}

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
	Data    []string `json:"data"`
}

type PeerRequest struct {
	NodeID string `json:"node_id"`
}

type PeerResponse struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    p2p.IdentifyProtocol `json:"data"`
}
