package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"AIComputingNode/pkg/types"
)

type ModelRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ModelResponse struct {
	Code     int    `json:"code"`
	Status   string `json:"status"`
	ImageUrl string `json:"imageUrl"`
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
