package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"AIComputingNode/pkg/types"
)

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

type ModelRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ModelResponse struct {
	Code     int    `json:"code"`
	Status   string `json:"status"`
	ImageUrl string `json:"imageUrl"`
}

func ChatModel(api string, chatReq ChatCompletionRequest) *ChatCompletionResponse {
	result := &ChatCompletionResponse{
		Code: int(types.ErrCodeModel),
	}
	if api == "" {
		result.Message = "Model API configuration is empty"
		return result
	}
	jsonData, err := json.Marshal(chatReq)
	if err != nil {
		result.Message = "Marshal model request error"
		return result
	}
	resp, err := http.Post(
		api,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil || resp.StatusCode != 200 {
		result.Message = "Post HTTP request error"
		return result
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Message = "Read model response error"
		return result
	}
	if err := json.Unmarshal(body, &result.Data); err != nil {
		result.Message = "Unmarshal model response error"
		return result
	}
	result.Code = 0
	return result
}

// curl -X POST "http://127.0.0.1:8080/models/superimage" -d "{\"prompt\":\"bird\"}"
// curl -X POST "http://127.0.0.1:8080/models" -d "{\"model\":\"superimage\",\"prompt\":\"bird\"}"
func ExecuteModel(api, model, prompt string) (int, string, string) {
	if api == "" {
		return int(types.ErrCodeModel), "Model API configuration is empty", ""
	}
	req := ModelRequest{
		Model:  model,
		Prompt: prompt,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return int(types.ErrCodeModel), "Marshal model request error", ""
	}
	resp, err := http.Post(
		fmt.Sprintf("%s/models", api),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil || resp.StatusCode != 200 {
		return int(types.ErrCodeModel), "Post HTTP request error", ""
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return int(types.ErrCodeModel), "Read model response error", ""
	}
	response := ModelResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return int(types.ErrCodeModel), "Unmarshal model response error", ""
	}
	return response.Code, response.Status, response.ImageUrl
}
