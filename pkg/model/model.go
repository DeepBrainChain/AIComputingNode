package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"AIComputingNode/pkg/types"
)

type ChatCompletionResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	types.ChatModelResponseData
}

type ImageGenerationResponse struct {
	Code    int                         `json:"code"`
	Message string                      `json:"message"`
	Created int64                       `json:"created"`
	Data    []types.ImageResponseChoice `json:"data"`
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
func ChatModel(api string, chatReq types.ChatModelRequest) *types.ChatCompletionResponse {
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
	if err != nil {
		result.Message = fmt.Sprintf("Post HTTP request error, %v", err)
		return result
	}
	defer resp.Body.Close()
	if resp.Header.Get("Content-Type") == "application/json" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			result.Message = "Read model response error"
			return result
		}
		chatRes := ChatCompletionResponse{}
		if err := json.Unmarshal(body, &chatRes); err != nil {
			result.Message = "Unmarshal model response error"
			return result
		}
		result.Code = chatRes.Code
		result.Message = chatRes.Message
		result.Data = chatRes.ChatModelResponseData
	} else if resp.StatusCode != 200 {
		result.Message = fmt.Sprintf("Post HTTP request error, %s", resp.Status)
	} else {
		result.Message = "Model HTTP reponse is not JSON"
	}
	return result
}

// curl -X POST "http://127.0.0.1:8080/models/superimage" -H "Content-Type: application/json" -d "{\"prompt\":\"bird\"}"
// curl -X POST "http://127.0.0.1:1088/v1/images/generations" -H "Content-Type: application/json" -d "{\"model\":\"superimage\",\"prompt\":\"bird\",\"n\":1,\"size\":\"1024x1024\"}"
func ImageGenerationModel(api string, req types.ImageGenModelRequest) *types.ImageGenerationResponse {
	result := &types.ImageGenerationResponse{
		Code: int(types.ErrCodeModel),
	}
	if api == "" {
		result.Message = "Model API configuration is empty"
		return result
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		result.Message = "Marshal model request error"
		return result
	}
	resp, err := http.Post(
		api,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		result.Message = fmt.Sprintf("Post HTTP request error, %v", err)
		return result
	}
	defer resp.Body.Close()
	if resp.Header.Get("Content-Type") == "application/json" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			result.Message = "Read model response error"
			return result
		}
		response := ImageGenerationResponse{}
		if err := json.Unmarshal(body, &response); err != nil {
			result.Message = "Unmarshal model response error"
			return result
		}
		result.Code = response.Code
		result.Message = response.Message
		result.Data.Created = response.Created
		result.Data.Choices = response.Data
	} else if resp.StatusCode != 200 {
		result.Message = fmt.Sprintf("Post HTTP request error, %s", resp.Status)
	} else {
		result.Message = "Model HTTP reponse is not JSON"
	}
	return result
}
