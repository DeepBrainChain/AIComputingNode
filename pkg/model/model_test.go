package model

import (
	"AIComputingNode/pkg/types"
	"testing"
)

func TestChatModel(t *testing.T) {
	var (
		api   = "http://122.99.183.53:1042/v1/chat/completions"
		model = "Llama3-70B"
	)

	req := types.ChatModelRequest{
		Model:    model,
		Messages: []types.ChatCompletionMessage{},
	}
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "system",
		Content: "You are a helpful assistant.",
	})
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "user",
		Content: "Hello",
	})
	res := ChatModel(api, req)
	if res.Code != 0 {
		t.Fatalf("Execute model %s error %s", model, res.Message)
	}
	t.Logf("Execute model %s result %v", model, res.Data)
}

func TestImageModel(t *testing.T) {
	var (
		api    = "http://127.0.0.1:8080/models"
		model  = "superimage"
		prompt = "bird"
	)

	req := types.ImageGenModelRequest{
		Model:  model,
		Prompt: prompt,
	}
	code, message, image := ImageGenerationModel(api, req)
	if code != 0 {
		t.Fatalf("Execute model %s with %q error %s", model, prompt, message)
	}
	t.Logf("Execute model %s with %q result %v", model, prompt, image)
}
