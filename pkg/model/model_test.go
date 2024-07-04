package model

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"AIComputingNode/pkg/test"
	"AIComputingNode/pkg/types"
)

var steamRequest = `阅读下面的材料，根据要求写作。
随着互联网的普及、人工智能的应用，越来越多的问题能很快得到答案。那么，我们的问题是否会越来越少？
以上材料引发了你怎样的联想和思考？请写一篇文章。
要求：选准角度，确定立意，明确文体，自拟标题；不要套作，不得抄袭；不得泄露个人信息；不少于800字。
`

// go test -v -timeout 30s -count=1 -run TestChatModel AIComputingNode/pkg/model
func TestChatModel(t *testing.T) {
	config, err := test.LoadConfig("D:/Code/AIComputingNode/test.json")
	if err != nil {
		t.Fatalf("Error loading test config file: %v", err)
	}

	req := types.ChatModelRequest{
		Model:    config.Models.Llama3.Name,
		Messages: []types.ChatCompletionMessage{},
		Stream:   false,
	}
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "system",
		Content: "You are a helpful assistant.",
	})
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "user",
		Content: "Hello",
	})
	res := ChatModel(config.Models.Llama3.API, req)
	if res.Code != 0 {
		t.Fatalf("Execute model %s error %s", config.Models.Llama3.Name, res.Message)
	}
	t.Logf("Execute model %s result %v", config.Models.Llama3.Name, res.Data)
}

// https://blog.csdn.net/QSTARTmachine/article/details/131993746
func StreamChatModel(api string, chatReq types.ChatModelRequest) (code int, message string) {
	code = int(types.ErrCodeModel)
	if api == "" {
		message = "Model API configuration is empty"
		return code, message
	}
	jsonData, err := json.Marshal(chatReq)
	if err != nil {
		message = "Marshal model request error"
		return code, message
	}
	resp, err := http.Post(
		api,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil || resp.StatusCode != 200 {
		message = "Post HTTP request error"
		return code, message
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	for {
		pack := types.StreamChatModelResponseData{}
		if err := dec.Decode(&pack); err != nil {
			fmt.Println(err, err.Error())
			if err != io.EOF {
				code = 0
				message = "EOF"
				return code, message
			}
			message = "json decode nil"
			return code, message
		}
		tempjson, err := json.Marshal(pack)
		if err != nil {
			fmt.Println("marshal json failed")
		} else {
			fmt.Println(string(tempjson))
		}
		if len(pack.Choices) > 0 && pack.Choices[0].FinishReason == "stop" {
			message = "stop"
			break
		}
	}
	code = 0
	return code, message
}

// https://blog.csdn.net/QSTARTmachine/article/details/131993746
func StreamChatModel2(api string, chatReq types.ChatModelRequest) (code int, message string) {
	code = int(types.ErrCodeModel)
	if api == "" {
		message = "Model API configuration is empty"
		return code, message
	}
	jsonData, err := json.Marshal(chatReq)
	if err != nil {
		message = "Marshal model request error"
		return code, message
	}
	resp, err := http.Post(
		api,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil || resp.StatusCode != 200 {
		message = "Post HTTP request error"
		return code, message
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	for {
		pack := types.StreamChatModelResponseData{}
		if !sc.Scan() {
			fmt.Println(sc.Err().Error())
		}
		line := sc.Bytes()
		if len(line) < 8 {
			continue
		}
		if string(line) == "data: [DONE]" {
			message = "stop"
			break
		}
		length := len(line)
		if err := json.Unmarshal(line[6:length], &pack); err != nil {
			fmt.Println(err, err.Error())
			if err != io.EOF {
				code = 0
				message = "EOF"
				return code, message
			}
			message = "json decode nil"
			return code, message
		}
		// fmt.Println(string(line))
		if len(pack.Choices) > 0 {
			fmt.Print(pack.Choices[0].Delta.Content)
			if pack.Choices[0].FinishReason == "stop" {
				fmt.Printf("\n{completion_tokens: %d, prompt_tokens: %d, total_tokens: %d}\n",
					pack.Usage.CompletionTokens, pack.Usage.PromptTokens, pack.Usage.TotalTokens)
				// message = "stop"
				// break
			}
		}
	}
	code = 0
	return code, message
}

// go test -v -timeout 300s -count=1 -run TestStreamChatModel AIComputingNode/pkg/model
func TestStreamChatModel(t *testing.T) {
	config, err := test.LoadConfig("D:/Code/AIComputingNode/test.json")
	if err != nil {
		t.Fatalf("Error loading test config file: %v", err)
	}

	req := types.ChatModelRequest{
		Model:    config.Models.Qwen2.Name,
		Messages: []types.ChatCompletionMessage{},
		Stream:   true,
	}
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "system",
		Content: "你是一名参加高考的高三学生",
	})
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "user",
		Content: steamRequest,
	})
	code, message := StreamChatModel2(config.Models.Qwen2.API, req)
	t.Logf("Execute stream chat model %v %s", code, message)
}

// go test -v -timeout 300s -count=1 -run TestImageModel AIComputingNode/pkg/model
func TestImageModel(t *testing.T) {
	config, err := test.LoadConfig("D:/Code/AIComputingNode/test.json")
	if err != nil {
		t.Fatalf("Error loading test config file: %v", err)
	}

	var (
		prompt = "bird"
	)

	req := types.ImageGenModelRequest{
		Model:  config.Models.SuperImage.Name,
		Prompt: prompt,
	}
	code, message, image := ImageGenerationModel(config.Models.SuperImage.API, req)
	if code != 0 {
		t.Fatalf("Execute model %s with %q error %s", config.Models.SuperImage.Name, prompt, message)
	}
	t.Logf("Execute model %s with %q result %v", config.Models.SuperImage.Name, prompt, image)
}
