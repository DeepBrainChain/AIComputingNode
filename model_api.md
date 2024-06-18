# AI Model Interface Standard Documentation

This document describes the API interface standard provided by the AI ​​model, which is provided to the distributed communication node for calling. The interface design refers to the [OpenAI API protocol](https://platform.openai.com/docs/api-reference/chat/create)。

## Text generation text model

Chat dialogue, text assistant

- request method: POST
- request URL: http://127.0.0.1:1088/v1/chat/completions
- request Body:
```json
{
  // Model name you want to request
  "model": "Llama3-8B",
  // Preset system assistant behavior mode and alternating question and answer records
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "Hello!"
    }
  ]
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "success",
  "created": 1677652288,
  "model": "Llama3-8B",
  // The answer given by the AI ​​model must give at least one
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "\n\nHello there, how may I assist you today?",
    },
    "finish_reason": "stop"
  }]
}
```

```shell
curl http://127.0.0.1:1088/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Llama3-8B",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": "Hello!"
      }
    ]
  }'
```

## Text generation image model

Generate pictures based on prompt words

- request method: POST
- request URL: http://127.0.0.1:1088/v1/images/generations
- request Body:
```json
{
  // Model name you want to request
  "model": "SuperImage",
  // Text description prompt words for the required image
  "prompt": "A cute baby sea otter",
  // The number of images to be generated, at least one
  "n": 2,
  // The size of the image to be generated
  "size": "1024x1024"
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "success",
  "created": 1589478378,
  // The answer given by the AI ​​model must give at least one
  "data": [
    {
      "url": "/home/AI_project/ImageGenerationAI/photos/v4xxidnrc9ol7m80.png"
    },
    {
      "url": "/home/AI_project/ImageGenerationAI/photos/bwjwyeqmz0yn6wjv.png"
    }
  ]
}
```

```shell
curl http://127.0.0.1:1088/v1/images/generations \
  -H "Content-Type: application/json" \
  -d '{
    "model": "SuperImage",
    "prompt": "A cute baby sea otter",
    "n": 1,
    "size": "1024x1024"
  }'
```

## Image editing model

Modify images based on prompt words

- request method: POST
- request URL: http://127.0.0.1:1088/v1/images/edits
- request Body:
```json
{
  // Model name you want to request
  "model": "SuperImage",
  // Image to be edited
  "image": "https://...",
  // Text description prompt words for the required image
  "prompt": "A cute baby sea otter wearing a beret",
  // The number of images to be generated, at least one
  "n": 2,
  // The size of the image to be generated
  "size": "1024x1024"
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "success",
  "created": 1589478378,
  // The answer given by the AI ​​model must give at least one
  "data": [
    {
      "url": "/home/AI_project/ImageGenerationAI/photos/v4xxidnrc9ol7m80.png"
    },
    {
      "url": "/home/AI_project/ImageGenerationAI/photos/bwjwyeqmz0yn6wjv.png"
    }
  ]
}
```

```shell
curl http://127.0.0.1:1088/v1/images/edits \
  -H "Content-Type: application/json" \
  -d '{
    "model": "SuperImage",
    "image": "https://...",
    "prompt": "A cute baby sea otter wearing a beret",
    "n": 1,
    "size": "1024x1024"
  }'
```

## Model list

A project can have multiple models. For example, DecentralGPT provides multiple models such as Llama3 70B and Qwen1.5-110B, so an interface can be provided to query the information of all models.

This interface may not be necessary. In actual deployment, the registration interface of the distributed network communication node needs to be called to inform the running model and the calling URL.

- request method: GET
- request URL: http://127.0.0.1:1088/v1/models
- return example:
```json
{
  // Project name, such as DecentralGPT, SuperImage
  "project": "xxx",
  // List of information such as the name and URL of the AI ​​model
  "data": [
    {
      "model": "Llama3-8B",
      "url": "http://127.0.0.1:1088/v1/chat/completions"
    },
    {
      "model": "Qwen1.5-110B",
      "url": "http://127.0.0.1:1088/v1/chat/completions"
    },
    {
      "model": "SuperImage",
      "url": "http://127.0.0.1:1088/v1/images/generations"
    }
  ]
}
```

```shell
curl http://127.0.0.1:1088/v1/models
```
