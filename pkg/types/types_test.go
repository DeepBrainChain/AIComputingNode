package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type testJson struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	EValue string `json:"evalue,omitempty"`
}

type testPerson struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (p testPerson) Say() {
	fmt.Println("I am Person struct")
}

type testStudent struct {
	testPerson
	Grade int `json:"grade"`
}

func (p testStudent) Say() {
	fmt.Println("I am Student struct")
}

func TestJsonStruct(t *testing.T) {
	test := testJson{
		Key:   "Test",
		Value: "value",
	}
	jsonData, err := json.Marshal(test)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s", string(jsonData))

	var value string = "{\"key\":\"Test\",\"value\":\"abc\"}"
	js := testJson{}
	if err := json.Unmarshal([]byte(value), &js); err != nil {
		t.Fatalf("Unmarshal json %v", err)
	}
	t.Logf("Unmarshal json sucess %v", js)

	value = "{\"KEY\":\"Test\",\"Value\":\"abc\"}"
	js2 := testJson{}
	if err := json.Unmarshal([]byte(value), &js2); err != nil {
		t.Fatalf("Unmarshal json %v", err)
	}
	t.Logf("Unmarshal json sucess %v", js2)

	test2 := testJson{
		Key: "Test",
	}
	jsonData, err = json.Marshal(test2)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s from struct %v", string(jsonData), test2)

	value = "{\"Model\":\"degpt\",\"API\":\"https://api.openai.com/v1/chat/completions/\",\"type\":1,\"cid\":\"246\"}"
	js3 := AIModelConfig{}
	if err := json.Unmarshal([]byte(value), &js3); err != nil {
		t.Fatalf("Unmarshal json %v", err)
	}
	t.Logf("Unmarshal json sucess %v", js3)
}

func TestDynamicJsonStruct(t *testing.T) {
	type DynamicMessage struct {
		Role    string      `json:"role"`
		Content interface{} `json:"content"`
	}

	type TextContentPart struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type ImageContentPart struct {
		Type     string `json:"type"`
		ImageUrl struct {
			Url    string `json:"url"`
			Detail string `json:"detail"`
		} `json:"image_url"`
	}

	type AudioContentPart struct {
		Type       string `json:"type"`
		InputAudio struct {
			Data   string `json:"data"`
			Format string `json:"format"`
		} `json:"input_audio"`
	}

	type DynamicContentPart struct {
		Type     string `json:"type"`
		Text     string `json:"text,omitempty"`
		ImageUrl struct {
			Url    string `json:"url"`
			Detail string `json:"detail,omitempty"`
		} `json:"image_url,omitempty"`
		InputAudio struct {
			Data   string `json:"data"`
			Format string `json:"format"`
		} `json:"input_audio,omitempty"`
	}

	type StringMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type ArrayMessage struct {
		Role    string               `json:"role"`
		Content []DynamicContentPart `json:"content"`
	}

	var value1 string = `
	{
	  "role": "user",
	  "content": "Hello!"
	}
	`
	sdmc := DynamicMessage{}
	if err := json.Unmarshal([]byte(value1), &sdmc); err != nil {
		t.Fatalf("Unmarshal string message content failed: %v", err)
	}
	switch v := sdmc.Content.(type) {
	case string:
		t.Logf("Content: %v - %v", sdmc.Content, v)
	case []interface{}:
		t.Logf("Content: %v - %v", sdmc.Content, v)
	}

	var value2 string = `
	{
	  "role": "user",
	  "content": [
      {
		    "type": "text",
				"text": "What's in this image?"
		  },
      {
		    "type": "image_url",
				"image_url": {
				  "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
				}
		  }
		]
	}
	`
	idmc := DynamicMessage{}
	if err := json.Unmarshal([]byte(value2), &idmc); err != nil {
		t.Fatalf("Unmarshal string message content failed: %v", err)
	}
	switch v := idmc.Content.(type) {
	case string:
		t.Logf("Content: %v - %v", idmc.Content, v)
	case []interface{}:
		t.Logf("Content: %v - %v", idmc.Content, v)
		t.Log(reflect.TypeOf(v))
		for k, dc := range v {
			t.Log(k, reflect.TypeOf(dc), dc)
			switch dv := dc.(type) {
			case TextContentPart:
				t.Logf("TextContentPart: %v", dv)
			case ImageContentPart:
				t.Logf("ImageContentPart: %v", dv)
			case DynamicContentPart:
				t.Logf("DynamicContentPart: %v", dv)
			case map[string]interface{}:
				for key, value := range dv {
					t.Logf("key-value: %v - %v", key, value)
				}
			default:
				t.Logf("unknown %v type: %v", reflect.TypeOf(dv), dv)
			}
		}
	}

	t.Log("The above analysis is very problematic")

	switch v := idmc.Content.(type) {
	case string:
		t.Logf("Content: %v - %v", idmc.Content, v)
	case []DynamicContentPart:
		t.Logf("Content: %v - %v", idmc.Content, v)
	default:
		t.Log(reflect.TypeOf(v))
	}

	t.Log("The above analysis is very problematic")

	iamc := ArrayMessage{}
	if err := json.Unmarshal([]byte(value2), &iamc); err != nil {
		t.Fatalf("Unmarshal string message content failed: %v", err)
	}
	for _, dc := range iamc.Content {
		t.Logf("%v, %v", reflect.TypeOf(dc), dc)
		t.Logf(dc.ImageUrl.Url)
	}

	t.Log("See if the string parsing with array content is successful")
	smc := StringMessage{}
	if err := json.Unmarshal([]byte(value2), &smc); err != nil {
		t.Fatalf("Unmarshal string message content failed: %v", err)
	}
	t.Log("the string parsing with array content is successful")
}

// https://www.jb51.net/article/284192.htm
func TestJsonRawMessage(t *testing.T) {
	type Message struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}

	type Request struct {
		Model     string    `json:"model"`
		Messages  []Message `json:"messages"`
		MaxTokens int       `json:"max_tokens"`
	}

	type ImageUrlObject struct {
		Url    string `json:"url"`
		Detail string `json:"detail,omitempty"`
	}

	type InputAudioObject struct {
		Data   string `json:"data"`
		Format string `json:"format"`
	}

	type DynamicContentPart struct {
		Type       string            `json:"type"`
		Text       string            `json:"text,omitempty"`
		ImageUrl   *ImageUrlObject   `json:"image_url,omitempty"`
		InputAudio *InputAudioObject `json:"input_audio,omitempty"`
	}

	var value1 string = `
	{
    "model": "NVLM-D-72B",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": [
          {
            "type": "text",
            "text": "What's in this image?"
          },
          {
            "type": "image_url",
            "image_url": {
              "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
            }
          }
        ]
      }
    ],
    "max_tokens": 300
  }
	`

	parseJson := func(value []byte) {
		req := Request{}
		if err := json.Unmarshal([]byte(value), &req); err != nil {
			t.Fatalf("Unmarshal request failed: %v", err)
		}
		t.Logf("model: %v, max_tokens: %v", req.Model, req.MaxTokens)
		// t.Logf("messages: %v", req.Messages)

		for _, message := range req.Messages {
			t.Logf("role: %v, content type: %v, content: %v", message.Role, reflect.TypeOf(message.Content), message.Content)
			dcp := []DynamicContentPart{}
			if err := json.Unmarshal(message.Content, &dcp); err != nil {
				t.Logf("Text content: %v", string(message.Content))
			} else {
				t.Logf("Array of content parts")
				for _, content := range dcp {
					switch content.Type {
					case "text":
						t.Logf("type: %v, text: %v", content.Type, content.Text)
					case "image_url":
						t.Logf("type: %v, image_url: %v", content.Type, content.ImageUrl.Url)
					case "input_audio":
						t.Logf("type: %v, input_audio: %v", content.Type, content.InputAudio.Data)
					default:
						t.Logf("unknowned type: %v", content.Type)
					}
				}
			}
		}
	}

	parseJson([]byte(value1))

	userContentParts := []DynamicContentPart{
		{
			Type: "text",
			Text: "What's in this image?",
		},
		{
			Type: "image_url",
			ImageUrl: &ImageUrlObject{
				Url: "https://example.com",
			},
		},
		{
			Type: "input_audio",
			InputAudio: &InputAudioObject{
				Data:   "data:audio/xxxx",
				Format: "mp3",
			},
		},
	}

	dcps, err := json.Marshal(userContentParts)
	if err != nil {
		t.Fatalf("Marshal array of user content parts failed: %v", err)
	}
	t.Logf("Marshal array of user content parts: %v", string(dcps))

	plainText := "You are a helpful assistant."
	jsonText, err := json.Marshal(plainText)
	if err != nil {
		t.Fatalf("Marshal plain text failed: %v", err)
	}
	t.Logf("plain text %v : %v", len(plainText), []byte(plainText))
	t.Logf("marshal plain text %v : %v", len(jsonText), []byte(jsonText))

	req := Request{
		Model: "NVLM-D-72B",
		Messages: []Message{
			// If json.RawMessage is a string, it must be wrapped in ""
			// json.RawMessage
			// {
			// 	Role:    "system",
			// 	Content: []byte("You are a helpful assistant."),
			// },
			{
				Role:    "system",
				Content: jsonText,
			},
			{
				Role:    "system",
				Content: []byte(`"You are a helpful assistant."`),
			},
			{
				Role:    "system",
				Content: []byte("\"You are a helpful assistant.\""),
			},
			{
				Role:    "user",
				Content: dcps,
			},
		},
		MaxTokens: 300,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal request with json.RawMessage failed: %v", err)
	}
	t.Logf("Marshal request with json.RawMessage: %v", string(reqBytes))

	parseJson(reqBytes)
}

func TestComposition(t *testing.T) {
	person := testPerson{
		Name: "Test",
		Age:  18,
	}
	person.Say()
	jsonData, err := json.Marshal(person)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s", string(jsonData))

	student := testStudent{
		testPerson: testPerson{
			Name: "Test1",
			Age:  15,
		},
		Grade: 12,
	}
	student.Say()
	student.testPerson.Say()
	jsonData, err = json.Marshal(student)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s", string(jsonData))
}

func TestNodeType(t *testing.T) {
	var nt NodeType = 0x00
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt |= PublicIpFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt |= PeersCollectFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt |= ModelFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt &= ^ModelFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt &= ^PeersCollectFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())

	nt &= ^PublicIpFlag
	t.Logf("%v ~ {%v, %v, %v}", nt, nt.IsPublicNode(), nt.IsClientNode(), nt.IsModelNode())
}

func TestJsonMap(t *testing.T) {
	ss := make(map[string]testPerson)
	ss["s1"] = testPerson{
		Name: "s1-name",
		Age:  18,
	}
	ss["s2"] = testPerson{
		Name: "s2-name",
		Age:  19,
	}

	// error will occur when Person is map[int]testPerson
	type Class struct {
		Name    string
		Persons map[string]testPerson
	}

	cls := Class{
		Name:    "c1",
		Persons: ss,
	}
	jsonData, err := json.Marshal(cls)
	if err != nil {
		t.Fatalf("Marshal json %v", err)
	}
	t.Logf("orignal json %s", string(jsonData))

	js := Class{}
	if err := json.Unmarshal(jsonData, &js); err != nil {
		t.Fatalf("Unmarshal json %v", err)
	}
	t.Logf("Unmarshal json sucess %v", js)
}
