package serve

import (
	"AIComputingNode/pkg/p2p"
	"errors"
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
	ErrCodeModel      = 1006
	ErrCodeUpload     = 1007
	ErrCodeInternal   = 5000
)

// error message
var errMsg = map[int]string{
	ErrCodeParam:      "Parameter error",
	ErrCodeParse:      "Parsing error",
	ErrCodeProtobuf:   "Protobuf serialization error",
	ErrCodeTimeout:    "Processing timeout",
	ErrCodeRendezvous: "Rendezvous error",
	ErrCodeModel:      "Model error",
	ErrCodeUpload:     "Upload error",
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

type ImageGenerationRequest struct {
	NodeID     string   `json:"node_id"`
	Model      string   `json:"model"`
	PromptWord []string `json:"prompt_word"`
}

type ImageGenerationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		IpfsNode  string `json:"ipfs_node"`
		CID       string `json:"cid"`
		ImageName string `json:"image_name"`
	} `json:"data"`
}

type SwarmConnectRequest struct {
	NodeAddr string `json:"node_addr"`
}

type SwarmConnectResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (req PeerRequest) Validate() error {
	if req.NodeID == "" {
		return errors.New("empty node_id")
	}
	return nil
}

func (req ImageGenerationRequest) Validate() error {
	if req.NodeID == "" {
		return errors.New("empty node_id")
	}
	if req.Model == "" {
		return errors.New("empty model")
	}
	return nil
}

func (req SwarmConnectRequest) Validate() error {
	if req.NodeAddr == "" {
		return errors.New("empty node_addr")
	}
	return nil
}
