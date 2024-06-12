package model

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"AIComputingNode/pkg/types"
)

type ChatCompletionRequest struct {
	Model    string                        `json:"model"`
	Messages []types.ChatCompletionMessage `json:"messages"`
}

type ImageGenerationRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Number int    `json:"n"`
	Size   string `json:"size"`
}

type OldImageGenerationResponse struct {
	Code     int    `json:"code"`
	Status   string `json:"status"`
	ImageUrl string `json:"imageUrl"`
}

type ImageModelChoice struct {
	Url string `json:"url"`
}

type ImageGenerationResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Images  []ImageModelChoice `json:"data"`
}

//	curl http://127.0.0.1:1042/v1/chat/completions -H "Content-Type: application/json" -d "{
//	   \"model\": \"Llama3-8B\",
//	   \"messages\": [
//	     {
//	       \"role\": \"system\",
//	       \"content\": \"You are a helpful assistant.\"
//	     },
//	     {
//	       \"role\": \"user\",
//	       \"content\": \"Hello\"
//	     }
//	   ]
//	 }"
func ChatModel(api string, chatReq ChatCompletionRequest) *types.ChatCompletionResponse {
	result := &types.ChatCompletionResponse{
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
func ImageGenerationModel(api string, req ImageGenerationRequest) (int, string, []string) {
	images := make([]string, 0)
	if api == "" {
		return int(types.ErrCodeModel), "Model API configuration is empty", images
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return int(types.ErrCodeModel), "Marshal model request error", images
	}
	resp, err := http.Post(
		api,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil || resp.StatusCode != 200 {
		return int(types.ErrCodeModel), "Post HTTP request error", images
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return int(types.ErrCodeModel), "Read model response error", images
	}
	response := ImageGenerationResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		oldresponse := OldImageGenerationResponse{}
		if err := json.Unmarshal(body, &oldresponse); err != nil {
			return int(types.ErrCodeModel), "Unmarshal model response error", images
		}
		return oldresponse.Code, oldresponse.Status, append(images, oldresponse.ImageUrl)
	}
	for _, image := range response.Images {
		images = append(images, image.Url)
	}
	return response.Code, response.Message, images
}
