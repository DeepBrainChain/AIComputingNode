package model

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"testing"

	"AIComputingNode/pkg/test"
	"AIComputingNode/pkg/types"
)

var steamRequest = `"阅读下面的材料，根据要求写作。
随着互联网的普及、人工智能的应用，越来越多的问题能很快得到答案。那么，我们的问题是否会越来越少？
以上材料引发了你怎样的联想和思考？请写一篇文章。
要求：选准角度，确定立意，明确文体，自拟标题；不要套作，不得抄袭；不得泄露个人信息；不少于800字。
"`

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
		Content: []byte(`"You are a helpful assistant."`),
	})
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "user",
		Content: []byte(`"Hello"`),
	})
	res := ChatModel(config.Models.Llama3.API, req)
	if res.Code != 0 {
		t.Fatalf("Execute model %s error {code: %v, message: %s}", config.Models.Llama3.Name, res.Code, res.Message)
	}
	t.Logf("Execute model %s result %v", config.Models.Llama3.Name, res.ChatModelResponseData)
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
	fmt.Println("Request: ", string(jsonData))
	resp, err := http.Post(
		api,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Println("Post HTTP request error", err)
		message = "Post HTTP request error"
		return code, message
	}
	defer resp.Body.Close()
	fmt.Println("HTTP response", resp.Status)
	fmt.Println("Response Content-Type:", resp.Header.Get("Content-Type"))
	if resp.Header.Get("Content-Type") == "application/json" {
		result := &types.ChatCompletionResponse{
			BaseHttpResponse: types.BaseHttpResponse{
				Code: int(types.ErrCodeModel),
			},
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			result.Message = "Read model response error"
			return result.Code, result.Message
		}
		if err := json.Unmarshal(body, &result); err != nil {
			result.Message = "Unmarshal model response error"
			return result.Code, result.Message
		}
		fmt.Printf("{code: %v, message: %v}\n", result.Code, result.Message)
		return result.Code, result.Message
	}
	if resp.StatusCode != 200 {
		message = "Post HTTP request error"
		return code, message
	}
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
		Content: []byte(`"你是一名参加高考的高三学生"`),
	})
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "user",
		Content: []byte(steamRequest),
	})
	code, message := StreamChatModel2(config.Models.Qwen2.API, req)
	t.Logf("Execute stream chat model %v %s", code, message)
}

func TestConcurrentStreamChatModel(t *testing.T) {
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
		Content: []byte(`"You are a helpful assistant."`),
	})
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "user",
		Content: []byte(`"Hello, What's the weather like today? Where is a good place to travel?"`),
	})
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal model request: %v", err)
	}

	concurrency := 5
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i int, body []byte) {
			defer wg.Done()
			t.Logf("Goroutine %v started", i)
			req, err := http.NewRequest("POST", config.Models.Qwen2.API, bytes.NewBuffer(body))
			if err != nil {
				t.Errorf("Goroutine %v new request failed: %v", i, err)
				return
			}
			defer req.Body.Close()

			resp, err := http.DefaultTransport.RoundTrip(req)
			if err != nil {
				t.Errorf("Goroutine %v roundtrip request failed: %v", i, err)
				return
			}
			defer resp.Body.Close()
			sc := bufio.NewScanner(resp.Body)
			for {
				if !sc.Scan() {
					if err := sc.Err(); errors.Is(err, io.EOF) {
						t.Logf("Goroutine %v read response EOF", i)
					} else if err != nil {
						t.Errorf("Goroutine %v read response failed: %v", i, err)
					}
					break
				}
				line := sc.Bytes()
				t.Logf("Goroutine %v read response %s", i, string(line))
			}
			t.Logf("Goroutine %v stopped", i)
		}(i, jsonData)
	}

	wg.Wait()
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
		Model:          config.Models.SuperImage.Name,
		Prompt:         prompt,
		Number:         1,
		Size:           "1024x1024",
		ResponseFormat: "url",
	}
	res := ImageGenerationModel(config.Models.SuperImage.API, req)
	if res.Code != 0 {
		t.Fatalf("Execute model %s with %q error {code: %v, message: %s}", config.Models.SuperImage.Name, prompt, res.Code, res.Message)
	}
	t.Logf("Execute model %s with %q result %v", config.Models.SuperImage.Name, prompt, res.ImageModelResponse)
}

// go test -v -timeout 300s -count=1 -run TestImageEdit AIComputingNode/pkg/model
func TestImageEdit(t *testing.T) {
	config, err := test.LoadConfig("/Volumes/data/code/AIComputingNode/test.json")
	if err != nil {
		t.Fatalf("Error loading test config file: %v", err)
	}

	var (
		prompt = "analog film photo of a man. faded film, desaturated, 35mm photo, grainy, vignette, vintage, Kodachrome, Lomography, stained, highly detailed, found footage, masterpiece, best quality"
		size   = "256x256"
	)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("face_image", "ai_func2.png")
	if err != nil {
		t.Fatalf("Create multipart form file failed: %v", err)
	}
	image, err := os.Open("../../ai_func2.png")
	if err != nil {
		t.Fatalf("Open image file failed: %v", err)
	}
	defer image.Close()
	_, err = io.Copy(part, image)
	if err != nil {
		t.Fatalf("Read image file failed: %v", err)
	}
	if err := writer.WriteField("prompt", prompt); err != nil {
		t.Fatalf("Write prompt failed: %v", err)
	}
	if err := writer.WriteField("size", size); err != nil {
		t.Fatalf("Write prompt failed: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer %v", err)
	}

	// res := ImageEditModel(config.Models.StyleID.API, writer.)
	// if res.Code != 0 {
	// 	t.Fatalf("Execute model %s with %q error {code: %v, message: %s}", config.Models.SuperImage.Name, prompt, res.Code, res.Message)
	// }
	// t.Logf("Execute model %s with %q result %v", config.Models.SuperImage.Name, prompt, res.ImageModelResponse)

	client := &http.Client{
		Timeout: types.ImageGenerationRequestTimeout,
	}
	resp, err := client.Post(
		config.Models.StyleID.API,
		writer.FormDataContentType(),
		body,
	)
	if err != nil {
		t.Fatalf("Post HTTP request error, %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("Reponse status code: %v", resp.Status)
	}
	t.Logf("Response Content-Type: %v", resp.Header.Get("Content-Type"))
	savefile, err := os.OpenFile("../../image-save.png", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0766)
	if err != nil {
		t.Fatalf("Failed to create file for saving: %v", err)
	}
	defer savefile.Close()
	written, err := io.Copy(savefile, resp.Body)
	if err != nil {
		t.Fatalf("Failed to save image file: %v", err)
	}
	t.Logf("Image %v bytes in reponse", written)
}
