# AI 模型接口标准文档

此文档描述 AI 模型提供的 API 接口标准，提供给分布式通信节点调用。接口设计参考 OpenAI API 协议 [API Reference - OpenAI API](https://platform.openai.com/docs/api-reference/chat/create)。

## 文生文模型

聊天对话，文字助理

- 请求方式：POST
- 请求 URL：http://127.0.0.1:1088/v1/chat/completions
- 请求 Body：
```json
{
  // 想要请求的模型名称
  "model": "Llama3-8B",
  // 预设的系统助理行为模式和交替问答记录
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
- 返回示例：
```json
{
  // 错误码，0 表示成功，非 0 表示失败
  "code": 0,
  // 错误信息
  "message": "success",
  "created": 1677652288,
  "model": "Llama3-8B",
  // AI 模型给出的回答，最少要给出一条
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

## 文生图模型

根据提示词生成图片

- 请求方式：POST
- 请求 URL：http://127.0.0.1:1088/v1/images/generations
- 请求 Body：
```json
{
  // 想要请求的模型名称
  "model": "SuperImage",
  // 所需图像的文本描述
  "prompt": "A cute baby sea otter",
  // 要生成的图像数量，最少一个
  "n": 2,
  // 要生成图像的大小
  "size": "1024x1024"
}
```
- 返回示例：
```json
{
  // 错误码，0 表示成功，非 0 表示失败
  "code": 0,
  // 错误信息
  "message": "success",
  "created": 1589478378,
  // AI 模型给出的回答，最少要给出一条
  "data": [
    {
      "url": "https://..."
    },
    {
      "url": "https://..."
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

## 修图模型

根据提示词修改图片

- 请求方式：POST
- 请求 URL：http://127.0.0.1:1088/v1/images/edits
- 请求 Body：
```json
{
  // 想要请求的模型名称
  "model": "SuperImage",
  // 要编辑的图片
  "image": "https://...",
  // 所需图像的文本描述
  "prompt": "A cute baby sea otter wearing a beret",
  // 要生成的图像数量
  "n": 2,
  // 要生成图像的大小
  "size": "1024x1024"
}
```
- 返回示例：
```json
{
  // 错误码，0 表示成功，非 0 表示失败
  "code": 0,
  // 错误信息
  "message": "success",
  "created": 1589478378,
  // AI 模型给出的回答，最少要给出一条
  "data": [
    {
      "url": "https://..."
    },
    {
      "url": "https://..."
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

## 模型列表

一个项目可以有多个模型，例如 DecentralGPT 提供了 Llama3 70B 和 Qwen1.5-110B 等多个模型，因此可以提供一个接口查询所有模型的信息。

这个接口可能不是必要，实际部署时需要调用分布式网络通信节点的注册接口来告知运行的模型和调用的 URL。

- 请求方式：GET
- 请求 URL：http://127.0.0.1:1088/v1/models
- 返回示例：
```json
{
  // 项目名称，例如 DecentralGPT，SuperImage
  "project": "xxx",
  // AI 模型给出的回答，最少要给出一条
  "data": [
    {
      "model": "Llama3-8B",
      "url": "https://..."
    },
    {
      "model": "Qwen1.5-110B",
      "url": "https://..."
    }
  ]
}
```

```shell
curl http://127.0.0.1:1088/v1/models
```
