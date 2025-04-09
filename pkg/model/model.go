package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"AIComputingNode/pkg/types"
)

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
		BaseHttpResponse: types.BaseHttpResponse{
			Code: int(types.ErrCodeModel),
		},
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
	client := &http.Client{
		Timeout: types.ChatCompletionRequestTimeout,
	}
	resp, err := client.Post(
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
		chatRes := types.ChatCompletionResponse{}
		if err := json.Unmarshal(body, &chatRes); err != nil {
			result.Message = "Unmarshal model response error"
			return result
		}
		result.BaseHttpResponse = chatRes.BaseHttpResponse
		result.ChatModelResponseData = chatRes.ChatModelResponseData
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
		BaseHttpResponse: types.BaseHttpResponse{
			Code: int(types.ErrCodeModel),
		},
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
	client := &http.Client{
		Timeout: types.ImageGenerationRequestTimeout,
	}
	resp, err := client.Post(
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
		response := types.ImageGenerationResponse{}
		if err := json.Unmarshal(body, &response); err != nil {
			result.Message = "Unmarshal model response error"
			return result
		}
		result.BaseHttpResponse = response.BaseHttpResponse
		result.ImageModelResponse = response.ImageModelResponse
	} else if resp.StatusCode != 200 {
		result.Message = fmt.Sprintf("Post HTTP request error, %s", resp.Status)
	} else {
		result.Message = "Model HTTP reponse is not JSON"
	}
	return result
}

func ImageEditModel(api string, form *multipart.Form) *types.ImageGenerationResponse {
	result := &types.ImageGenerationResponse{
		BaseHttpResponse: types.BaseHttpResponse{
			Code: int(types.ErrCodeModel),
		},
	}
	if api == "" {
		result.Message = "Model API configuration is empty"
		return result
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// copy file field in form
	for fieldName, files := range form.File {
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				result.Message = fmt.Sprintf("Failed to open %v %v", fieldName, err)
				return result
			}
			defer file.Close()

			part, err := writer.CreateFormFile(fieldName, fileHeader.Filename)
			if err != nil {
				result.Message = fmt.Sprintf("Failed to create form file for %v %v", fieldName, err)
				return result
			}

			_, err = io.Copy(part, file)
			if err != nil {
				result.Message = fmt.Sprintf("Failed to copy form file for %v %v", fieldName, err)
				return result
			}
		}
	}

	// copy other field in form
	for fieldName, values := range form.Value {
		for _, value := range values {
			if err := writer.WriteField(fieldName, value); err != nil {
				result.Message = fmt.Sprintf("Failed to copy %v field %v", fieldName, err)
				return result
			}
		}
	}

	if err := writer.Close(); err != nil {
		result.Message = fmt.Sprintf("Failed to close multipart writer %v", err)
		return result
	}

	client := &http.Client{
		Timeout: types.ImageGenerationRequestTimeout,
	}
	resp, err := client.Post(
		api,
		writer.FormDataContentType(),
		body,
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
		response := types.ImageGenerationResponse{}
		if err := json.Unmarshal(body, &response); err != nil {
			result.Message = "Unmarshal model response error"
			return result
		}
		result.BaseHttpResponse = response.BaseHttpResponse
		result.ImageModelResponse = response.ImageModelResponse
	} else if resp.StatusCode != 200 {
		result.Message = fmt.Sprintf("Post HTTP request error, %s", resp.Status)
	} else {
		result.Message = "Model HTTP reponse is not JSON"
	}
	return result
}
