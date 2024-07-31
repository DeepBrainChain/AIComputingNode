package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

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

type HostInfoRequest BaseHttpRequest

type HostInfoResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    HostInfo `json:"data"`
}

type WalletVerification struct {
	Wallet    string `json:"wallet,omitempty"`
	Signature string `json:"signature,omitempty"`
	Hash      string `json:"hash,omitempty"`
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

type StreamChatResponseChoice struct {
	Index        int                   `json:"index"`
	Delta        ChatCompletionMessage `json:"delta"`
	FinishReason string                `json:"finish_reason"`
}

type ChatModelRequest struct {
	Model    string                  `json:"model"`
	Messages []ChatCompletionMessage `json:"messages"`
	Stream   bool                    `json:"stream"`
	WalletVerification
}

type ChatCompletionRequest struct {
	NodeID  string `json:"node_id"`
	Project string `json:"project"`
	ChatModelRequest
}

type ChatCompletionProxyRequest struct {
	Project string `json:"project"`
	ChatModelRequest
}

type ChatResponseUsage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatModelResponseData struct {
	Created int64                `json:"created"`
	Choices []ChatResponseChoice `json:"choices"`
	Usage   ChatResponseUsage    `json:"usage"`
}

type StreamChatModelResponseData struct {
	Created int64                      `json:"created"`
	Choices []StreamChatResponseChoice `json:"choices"`
	Usage   ChatResponseUsage          `json:"usage"`
}

type ChatCompletionResponse struct {
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	Data    ChatModelResponseData `json:"data"`
}

type ImageResponseChoice struct {
	CID       string `json:"cid"`
	ImageName string `json:"image_name"`
}

type ImageGenModelRequest struct {
	Project string `json:"project"`
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Number  int    `json:"n"`
	Size    string `json:"size"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	WalletVerification
}

type ImageGenerationRequest struct {
	NodeID string `json:"node_id"`
	ImageGenModelRequest
	IpfsNode string `json:"ipfs_node"`
}

type ImageGenerationProxyRequest struct {
	ImageGenModelRequest
	IpfsNode string `json:"ipfs_node"`
}

type ImageGenerationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		IpfsNode string                `json:"ipfs_node"`
		Choices  []ImageResponseChoice `json:"choices"`
	} `json:"data"`
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

type GetModelsOfAIProjectRequest struct {
	Project string `json:"project"`
}

type GetPeersOfAIProjectRequest struct {
	Project string `json:"project"`
	Model   string `json:"model"`
}

type AIProjectPeerInfo struct {
	NodeID       string `json:"node_id"`
	Connectivity int    `json:"connectivity"`
	Latency      int64  `json:"latency"`
}

type GetPeersOfAIProjectResponse struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    []AIProjectPeerInfo `json:"data"`
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

func (res *HostInfoResponse) SetCode(code int) {
	res.Code = code
}

func (res *HostInfoResponse) SetMessage(message string) {
	res.Message = message
}

func (res *ChatCompletionResponse) SetCode(code int) {
	res.Code = code
}

func (res *ChatCompletionResponse) SetMessage(message string) {
	res.Message = message
}

func (res *ImageGenerationResponse) SetCode(code int) {
	res.Code = code
}

func (res *ImageGenerationResponse) SetMessage(message string) {
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

func (req HostInfoRequest) Validate() error {
	if req.NodeID == "" {
		return errors.New("empty node_id")
	}
	return nil
}

func (req ChatCompletionRequest) Validate() error {
	if req.NodeID == "" {
		return errors.New("empty node_id")
	}
	if req.Project == "" {
		return errors.New("empty project")
	}
	if req.Model == "" {
		return errors.New("empty model")
	}
	return nil
}

func (req ChatCompletionProxyRequest) Validate() error {
	if req.Project == "" {
		return errors.New("empty project")
	}
	if req.Model == "" {
		return errors.New("empty model")
	}
	return nil
}

func (req ImageGenerationRequest) Validate() error {
	if req.NodeID == "" {
		return errors.New("empty node_id")
	}
	if req.Project == "" {
		return errors.New("empty project")
	}
	if req.Model == "" {
		return errors.New("empty model")
	}
	return nil
}

func (req ImageGenerationProxyRequest) Validate() error {
	if req.Project == "" {
		return errors.New("empty project")
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

func (req AIProjectListRequest) Validate() error {
	if req.NodeID == "" {
		return errors.New("empty node_id")
	}
	return nil
}

func (req GetModelsOfAIProjectRequest) Validate() error {
	if req.Project == "" {
		return errors.New("empty project")
	}
	return nil
}

func (req GetPeersOfAIProjectRequest) Validate() error {
	if req.Project == "" {
		return errors.New("empty project")
	}
	return nil
}

func (req ChatModelRequest) RequestBody() (io.ReadCloser, int64, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, 0, err
	}
	return io.NopCloser(bytes.NewBuffer(jsonData)), int64(len(jsonData)), nil
}

func (req ChatCompletionRequest) RequestBody() (io.ReadCloser, int64, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, 0, err
	}
	return io.NopCloser(bytes.NewBuffer(jsonData)), int64(len(jsonData)), nil
}

func (req ImageGenerationRequest) RequestBody() (io.ReadCloser, int64, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, 0, err
	}
	return io.NopCloser(bytes.NewBuffer(jsonData)), int64(len(jsonData)), nil
}
