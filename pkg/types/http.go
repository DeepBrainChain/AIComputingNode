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
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type PeerListResponse struct {
	BaseHttpResponse
	Data []string `json:"data"`
}

type PeerRequest BaseHttpRequest

type PeerResponse struct {
	BaseHttpResponse
	IdentifyProtocol
}

type HostInfoRequest BaseHttpRequest

type HostInfoResponse struct {
	BaseHttpResponse
	HostInfo
}

type WalletVerification struct {
	Wallet    string `json:"wallet,omitempty"`
	Signature string `json:"signature,omitempty"`
	Hash      string `json:"hash,omitempty"`
}

type ChatImageContentPart struct {
	Url    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type ChatAudioContentPart struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

type ChatDynamicContentPart struct {
	Type       string                `json:"type"`
	Text       string                `json:"text,omitempty"`
	ImageUrl   *ChatImageContentPart `json:"image_url,omitempty"`
	InputAudio *ChatAudioContentPart `json:"input_audio,omitempty"`
}

type ChatCompletionMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type ChatModelRequest struct {
	Model       string                  `json:"model"`
	Messages    []ChatCompletionMessage `json:"messages"`
	Stream      bool                    `json:"stream"`
	Temperature float32                 `json:"temperature,omitempty"`
	TopP        float32                 `json:"top_p,omitempty"`
	WalletVerification
}

type ChatCompletionRequest struct {
	NodeID  string `json:"node_id"`
	Project string `json:"project"`
	CID     string `json:"cid"`
	ChatModelRequest
}

type ChatCompletionProxyRequest struct {
	Project string `json:"project"`
	ChatModelRequest
}

type ChatCompletionResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponseChoice struct {
	Index        int                           `json:"index"`
	Message      ChatCompletionResponseMessage `json:"message"`
	FinishReason string                        `json:"finish_reason"`
}

type StreamChatResponseChoice struct {
	Index        int                           `json:"index"`
	Delta        ChatCompletionResponseMessage `json:"delta"`
	FinishReason string                        `json:"finish_reason"`
}

type ChatResponseUsage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatModelResponseData struct {
	Id      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Choices []ChatResponseChoice `json:"choices"`
	Usage   ChatResponseUsage    `json:"usage"`
}

type StreamChatModelResponseData struct {
	Id      string                     `json:"id"`
	Object  string                     `json:"object"`
	Created int64                      `json:"created"`
	Choices []StreamChatResponseChoice `json:"choices"`
	Usage   ChatResponseUsage          `json:"usage"`
}

type ChatCompletionResponse struct {
	BaseHttpResponse
	ChatModelResponseData
}

type ImageGenModelRequest struct {
	File           string `json:"file,omitempty"`
	Model          string `json:"model" form:"model"`
	Prompt         string `json:"prompt"`
	Number         int    `json:"n"`
	Size           string `json:"size"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	Step           int32  `json:"step,omitempty"`
	ResponseFormat string `json:"response_format"`
	WalletVerification
}

type ImageGenerationRequest struct {
	NodeID  string `json:"node_id"  form:"node_id"`
	Project string `json:"project" form:"project"`
	CID     string `json:"cid" form:"cid"`
	ImageGenModelRequest
}

type ImageGenerationProxyRequest struct {
	Project string `json:"project" form:"project"`
	ImageGenModelRequest
}

type ImageResponseChoice struct {
	B64Json       string `json:"b64_json,omitempty"`
	Url           string `json:"url,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type ImageModelResponse struct {
	Created int64                 `json:"created,omitempty"`
	Choices []ImageResponseChoice `json:"data,omitempty"`
}

type ImageGenerationResponse struct {
	BaseHttpResponse
	ImageModelResponse
}

type SwarmConnectRequest struct {
	NodeAddr string `json:"node_addr"`
}

type SwarmConnectResponse BaseHttpResponse

type AIProjectListRequest BaseHttpRequest

type AIProjectListResponse struct {
	BaseHttpResponse
	Data map[string][]ModelIdle `json:"data"`
}

type GetAIProjectsRequest struct {
	Number int `json:"number" form:"number"`
}

type GetModelsOfAIProjectRequest struct {
	Project string `json:"project" form:"project"`
	Number  int    `json:"number" form:"number"`
}

type GetPeersOfAIProjectRequest struct {
	Project string `json:"project" form:"project"`
	Model   string `json:"model" form:"model"`
	Number  int    `json:"number" form:"number"`
}

type AIProjectPeerInfo struct {
	NodeID       string `json:"node_id"`
	Connectivity int    `json:"connectivity"`
	Latency      int64  `json:"latency"`
	Idle         int    `json:"Idle"`
	CID          string `json:"cid"`
}

type GetPeersOfAIProjectResponse struct {
	BaseHttpResponse
	Data []AIProjectPeerInfo `json:"data"`
}

type AIProjectPeerOrder []AIProjectPeerInfo

func (a AIProjectPeerOrder) Len() int      { return len(a) }
func (a AIProjectPeerOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AIProjectPeerOrder) Less(i, j int) bool {
	// First condition: Connectivity of 1 is ranked first
	// if a[i].Connectivity != a[j].Connectivity {
	// 	return a[i].Connectivity == 1
	// }
	if a[i].Connectivity == 1 && a[j].Connectivity != 1 {
		return true
	}
	if a[i].Connectivity != 1 && a[j].Connectivity == 1 {
		return false
	}
	// Second condition: The smaller the Idle, the higher the ranking
	if a[i].Idle != a[j].Idle {
		return a[i].Idle < a[j].Idle
	}
	// Third condition: The smaller the Latency, the higher the ranking
	return a[i].Latency < a[j].Latency
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

func (req WalletVerification) Validate() error {
	if req.Wallet == "" {
		return errors.New("empty wallet")
	}
	if req.Signature == "" {
		return errors.New("empty signature")
	}
	if req.Hash == "" {
		return errors.New("empty hash")
	}
	return nil
}

func (cdcp ChatDynamicContentPart) Validate() error {
	switch cdcp.Type {
	case "text":
	case "image_url":
		if cdcp.ImageUrl == nil {
			return errors.New("the url of image content object is nil")
		}
	case "input_audio":
		if cdcp.InputAudio == nil {
			return errors.New("the data of audio content object is nil")
		}
	default:
		return errors.New("unknowned part type of chat content")
	}
	return nil
}

func (ccm ChatCompletionMessage) Validate() error {
	switch ccm.Role {
	case "system":
	case "user":
	case "assistant":
	case "tool":
	case "function":
	default:
		return errors.New("unknowned role of chat message author")
	}
	if len(ccm.Content) < 2 {
		return errors.New("invalid content of chat message")
	}
	if ccm.Content[0] == '[' && ccm.Content[len(ccm.Content)-1] == ']' {
		parts := []ChatDynamicContentPart{}
		if err := json.Unmarshal(ccm.Content, &parts); err != nil {
			return err
		}
		for _, part := range parts {
			if err := part.Validate(); err != nil {
				return err
			}
		}
	} else if ccm.Content[0] == '"' && ccm.Content[len(ccm.Content)-1] == '"' {
		//
	} else {
		return errors.New("invalid content of chat message")
	}
	return nil
}

func (req ChatModelRequest) Validate() error {
	for _, ccm := range req.Messages {
		if err := ccm.Validate(); err != nil {
			return err
		}
	}
	// if err := req.WalletVerification.Validate(); err != nil {
	// 	return err
	// }
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
	if err := req.ChatModelRequest.Validate(); err != nil {
		return err
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
	if err := req.ChatModelRequest.Validate(); err != nil {
		return err
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
	// if err := req.ImageGenModelRequest.WalletVerification.Validate(); err != nil {
	// 	return err
	// }
	return nil
}

func (req ImageGenerationProxyRequest) Validate() error {
	if req.Project == "" {
		return errors.New("empty project")
	}
	if req.Model == "" {
		return errors.New("empty model")
	}
	// if err := req.ImageGenModelRequest.WalletVerification.Validate(); err != nil {
	// 	return err
	// }
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

func (req GetAIProjectsRequest) Validate() error {
	if req.Number < 0 {
		return errors.New("invalid number")
	}
	return nil
}

func (req GetModelsOfAIProjectRequest) Validate() error {
	if req.Project == "" {
		return errors.New("empty project")
	}
	if req.Number < 0 {
		return errors.New("invalid number")
	}
	return nil
}

func (req GetPeersOfAIProjectRequest) Validate() error {
	if req.Project == "" {
		return errors.New("empty project")
	}
	if req.Model == "" {
		return errors.New("empty model")
	}
	if req.Number < 0 {
		return errors.New("invalid number")
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

func (req ImageGenModelRequest) RequestBody() (io.ReadCloser, int64, error) {
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
