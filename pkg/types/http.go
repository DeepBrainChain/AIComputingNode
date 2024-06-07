package types

import "errors"

type HttpResponse interface {
	SetCode(code int)
	SetMessage(message string)
}

type BaseHttpRequest struct {
	NodeID string `json:"node_id"`
}

type BaseHttpResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PeerListResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

type PeerRequest BaseHttpRequest

type PeerResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    IdentifyProtocol `json:"data"`
}

type ImageGenerationRequest struct {
	NodeID     string `json:"node_id"`
	Model      string `json:"model"`
	PromptWord string `json:"prompt_word"`
	IpfsNode   string `json:"ipfs_node"`
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

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponseChoice struct {
	Index        int                   `json:"index"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

type ChatCompletionRequest struct {
	NodeID   string                  `json:"node_id"`
	Model    string                  `json:"model"`
	Messages []ChatCompletionMessage `json:"messages"`
}

type ChatCompletionResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Created int64                `json:"created"`
		Choices []ChatResponseChoice `json:"choices"`
	} `json:"data"`
}

type HostInfoRequest BaseHttpRequest

type HostInfoResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    HostInfo `json:"data"`
}

type SwarmConnectRequest struct {
	NodeAddr string `json:"node_addr"`
}

type SwarmConnectResponse BaseHttpResponse

type AIProjectListRequest BaseHttpRequest

type AIProjectListResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    []AIProjectOfNode `json:"data"`
}

func (res *BaseHttpResponse) SetCode(code int) {
	res.Code = code
}

func (res *BaseHttpResponse) SetMessage(message string) {
	res.Message = message
}

func (res *PeerResponse) SetCode(code int) {
	res.Code = code
}

func (res *PeerResponse) SetMessage(message string) {
	res.Message = message
}

func (res *ImageGenerationResponse) SetCode(code int) {
	res.Code = code
}

func (res *ImageGenerationResponse) SetMessage(message string) {
	res.Message = message
}

func (res *ChatCompletionResponse) SetCode(code int) {
	res.Code = code
}

func (res *ChatCompletionResponse) SetMessage(message string) {
	res.Message = message
}

func (res *HostInfoResponse) SetCode(code int) {
	res.Code = code
}

func (res *HostInfoResponse) SetMessage(message string) {
	res.Message = message
}

func (res *AIProjectListResponse) SetCode(code int) {
	res.Code = code
}

func (res *AIProjectListResponse) SetMessage(message string) {
	res.Message = message
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

func (req ChatCompletionRequest) Validate() error {
	if req.NodeID == "" {
		return errors.New("empty node_id")
	}
	if req.Model == "" {
		return errors.New("empty model")
	}
	return nil
}

func (req HostInfoRequest) Validate() error {
	if req.NodeID == "" {
		return errors.New("empty node_id")
	}
	return nil
}

func (req AIProjectListRequest) Validate() error {
	if req.NodeID == "" {
		return errors.New("empty node_id")
	}
	return nil
}

func (req SwarmConnectRequest) Validate() error {
	if req.NodeAddr == "" {
		return errors.New("empty node_addr")
	}
	return nil
}
