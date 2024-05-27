# AI 模型接口标准文档

遵循 OpenAI API 协议 [API Reference - OpenAI API](https://platform.openai.com/docs/api-reference/chat/create)

## 文生文模型

聊天对话，文字助理

- 请求方式：POST
- 请求 URL：https://api.openai.com/v1/chat/completions
- 请求 Body：
```json
{
  // 想要请求的模型名称
  "model": "gpt-4-turbo",
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
  "model": "gpt-4-turbo",
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
curl https://api.openai.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4-turbo",
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
- 请求 URL：https://api.openai.com/v1/images/generations
- 请求 Body：
```json
{
  // 想要请求的模型名称
  "model": "dall-e-3",
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
curl https://api.openai.com/v1/images/generations \
  -H "Content-Type: application/json" \
  -d '{
    "model": "dall-e-3",
    "prompt": "A cute baby sea otter",
    "n": 1,
    "size": "1024x1024"
  }'
```

## 修图模型

根据提示词修改图片

- 请求方式：POST
- 请求 URL：https://api.openai.com/v1/images/edits
- 请求 Body：
```json
{
  // 想要请求的模型名称
  "model": "dall-e-3",
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
curl https://api.openai.com/v1/images/edits \
  -H "Content-Type: application/json" \
  -d '{
    "model": "dall-e-3",
    "image": "https://...",
    "prompt": "A cute baby sea otter wearing a beret",
    "n": 1,
    "size": "1024x1024"
  }'
```

## 模型列表

一个项目可以有多个模型，例如 OpenAI 提供了 ChatPGT 3.5 和 4 等多个模型，因此可以提供一个接口查询所有模型的信息。

这个接口可能不是必要，实际部署时需要调用分布式网络通信节点的注册接口来告知运行的模型和调用的 URL。

- 请求方式：GET
- 请求 URL：https://api.openai.com/v1/models
- 返回示例：
```json
{
  // 项目名称，例如 DecentralGPT，SuperImage
  "project": "xxx",
  // AI 模型给出的回答，最少要给出一条
  "data": [
    {
      "model": "gpt-4-turbo",
      "url": "https://..."
    },
    {
      "model": "gpt-3.5-turbo",
      "url": "https://..."
    }
  ]
}
```

```shell
curl https://api.openai.com/v1/models
```
