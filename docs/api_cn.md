# AIComputingNode HTTP API 接口文档

此文档描述 AIComputingNode 分布式通信节点的 HTTP API 接口。

详细测试用例可以在 Apifox 平台上获取查看: <https://xr03hymjol.apifox.cn>。

公共的接口约定：
1. 当 HTTP 请求成功处理并返回预期的结果，给出状态码 200 OK。
2. 在处理 HTTP 请求时，遇到数据库错误、逻辑错误、计算错误等服务器内部问题时，通常返回 500 Internal Server Error 状态码，而不是 200，并且带有以下 JSON 告诉客户端发生了什么错误，帮助开发人员定位问题。
```json
{
  "code": 1010,
  "message": "Unsupported function"
}
```

## 常用查询接口

用于查询常用的节点信息、节点列表或者机器信息的接口。

### 获取节点自身的 PeerInfo

PeerInfo 为包含节点 ID、协议、版本和监听地址等基础信息的结构体。

节点 ID 是识别分布式通信节点的唯一 ID。

此接口只用来查询节点自身的 PeerInfo，不需要经过分布式网络的参与。

- 请求方式: GET
- 请求 URL: http://127.0.0.1:6000/api/v0/id
- 请求 Body: None
- 返回示例:
```json
{
  "peer_id": "16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF",
  "protocol_version": "aicn/0.0.1",
  "agent_version": "v0.0.9",
  "addresses": [
    "/ip4/127.0.0.1/tcp/6001",
    "/ip4/172.19.236.172/tcp/6001",
    "/ip6/::1/tcp/6001"
  ],
  "protocols": [
    "/ipfs/ping/1.0.0",
    "/libp2p/circuit/relay/0.2.0/stop",
    "/dbc/kad/1.0.0",
    "/libp2p/autonat/1.0.0",
    "/ipfs/id/1.0.0",
    "/ipfs/id/push/1.0.0",
    "/floodsub/1.0.0",
    "/libp2p/circuit/relay/0.2.0/hop"
  ]
}
```

### 获取节点列表

此接口用来查询分布式通信网络中的节点列表

- 请求方式: GET
- 请求 URL: http://127.0.0.1:6000/api/v0/peers
- 请求 Body: None
- 返回示例:
```json
{
  "data": [
    "16Uiu2HAm49H3Hcae8rxKBdw8PfFcFAnBXQS8ierXA1VoZwhdDadV",
    "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL",
    "16Uiu2HAmDBYxgdKxeCbmn8hYiqwK3xHR9533WDdEYmpEDQ259GTe",
    "16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
    "16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF"
  ]
}
```

一般请求的返回中都会有 code message 和 data 几个字段。

code 表示错误码，0 表示成功，非 0 表示失败，可在文章末尾处查看常用的错误码。

message 表示错误信息。

data 包含接口请求的结果信息(当 code = 0 时有效)。

### 获取任意节点的 PeerInfo

此接口可用来查询分布式通信网络中任意节点的 PeerInfo。

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/peer
- 请求 Body:
```json
{
  // 想要查询的节点
  "node_id": "16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM"
}
```
- 返回示例:
```json
{
  "peer_id": "16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
  "protocol_version": "aicn/0.0.1",
  "agent_version": "v0.0.9",
  "addresses": [
    "/ip4/122.99.183.54/tcp/6001",
    "/ip4/127.0.0.1/tcp/6001",
    "/ip6/::1/tcp/6001"
  ],
  "protocols": [
    "/ipfs/ping/1.0.0",
    "/libp2p/circuit/relay/0.2.0/stop",
    "/dbc/kad/1.0.0",
    "/libp2p/autonat/1.0.0",
    "/ipfs/id/1.0.0",
    "/ipfs/id/push/1.0.0",
    "/floodsub/1.0.0",
    "/libp2p/circuit/relay/0.2.0/hop"
  ]
}
```

### 获取任意节点的机器信息

此接口用来查询分布式通信网络中任意节点的机器软硬件信息。

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/host/info
- 请求 Body:
```json
{
  // 想要查询的节点
  "node_id": "16Uiu2HAmDBYxgdKxeCbmn8hYiqwK3xHR9533WDdEYmpEDQ259GTe"
}
```
- 返回示例:
```json
{
  "os": {
    "os": "windows",
    "platform": "Microsoft Windows 11 Pro",
    "platform_family": "Standalone Workstation",
    "platform_version": "10.0.22631.3737 Build 22631.3737",
    "kernel_version": "10.0.22631.3737 Build 22631.3737",
    "kernel_arch": "x86_64"
  },
  "cpu": [
    {
      "model_name": "Intel(R) Core(TM) i7-8700 CPU @ 3.20GHz",
      "total_cores": 6,
      "total_threads": 12
    }
  ],
  "memory": {
    "total_physical_bytes": 17179869184,
    "total_usable_bytes": 17105440768
  },
  "disk": [
    {
      "drive_type": "HDD",
      "size_bytes": 2000396321280,
      "model": "WDC WD20EJRX-89G3VY0",
      "serial_number": "WD-WCC4M2USUZ1V"
    },
    {
      "drive_type": "SSD",
      "size_bytes": 240054796800,
      "model": "TOSHIBA-TR200",
      "serial_number": "29KB71U8K46S"
    }
  ],
  "gpu": [
    {
      "vendor": "qdesk",
      "product": "Qdesk Virtual Display Adapter"
    },
    {
      "vendor": "NVIDIA",
      "product": "NVIDIA GeForce RTX 2080 Ti"
    }
  ]
}
```

## 模型调用接口

调用 AI 模型的接口。

### 文生文模型

此接口用来调用文生文模型

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/chat/completion
- 请求 Body:
```json
{
  // 运行模型的节点
  "node_id": "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL",
  // AI 项目名称
  "project": "DecentralGPT",
  // 模型名称
  "model": "Llama3-70B",
  // 预设的系统助理行为模式和交替问答记录
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "Hello"
    }
  ],
  // 如果此项设置为 true，返回数据会将消息增量一段一段以流式传输，数据流以 data: [DONE] 结束。
  "stream": false,
  // 用户的钱包公钥
  "wallet": "",
  // 钱包签名
  "signature": "",
  // 原始数据的 hash
  "hash": ""
}
```
- 返回示例:
```json
{
  "created": 1718691167,
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! It's nice to meet you. Is there something I can help you with, or would you like to chat for a bit? I'm here to assist you with any questions or tasks you might have."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "completion_tokens": 44,
    "prompt_tokens": 22,
    "total_tokens": 66
  }
}
```

### 文生文模型(使用项目名称)

**此接口从 v0.1.2 版本开始被弃用，从 v0.1.4 版本开始恢复使用。**

此接口使用项目名称来调用文生文模型。Input 节点会选择一些运行指定项目和模型的 Worker 节点，根据策略(RTT连接时延或者GPU空闲值等)排序，依次向 Worker 节点发送模型请求直到获得正确的应答，失败次数过多时会报错。

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/chat/completion/proxy
- 请求 Body:
```json
{
  // AI 项目名称
  "project": "DecentralGPT",
  // 模型名称
  "model": "Llama3-70B",
  // 预设的系统助理行为模式和交替问答记录
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "Hello"
    }
  ],
  // 如果此项设置为 true，返回数据会将消息增量一段一段以流式传输，数据流以 data: [DONE] 结束。
  "stream": false,
  // 用户的钱包公钥
  "wallet": "",
  // 钱包签名
  "signature": "",
  // 原始数据的 hash
  "hash": ""
}
```
- 返回示例:
```json
{
  "created": 1718691167,
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! It's nice to meet you. Is there something I can help you with, or would you like to chat for a bit? I'm here to assist you with any questions or tasks you might have."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "completion_tokens": 44,
    "prompt_tokens": 22,
    "total_tokens": 66
  }
}
```

### 文生图模型

此接口用来调用文生图模型。

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/image/gen
- 请求 Body:
```json
{
  // 运行模型的节点
  "node_id": "16Uiu2HAm49H3Hcae8rxKBdw8PfFcFAnBXQS8ierXA1VoZwhdDadV",
  // AI 项目名称
  "project": "SuperImage",
  // 模型名称
  "model": "superImage",
  // 所需图片的描述或者提示词
  "prompt": "a bird flying in the sky",
  // 要生成的图像数量，最少一个
  "n": 2,
  // 要生成图像的大小
  "size": "1024x1024",
  "width": 1024,
  "height": 1024,
  // 要生成图像的格式，必须是 url 或 b64_json 之一
  "response_format": "url",
  // 用户的钱包公钥
  "wallet": "",
  // 钱包签名
  "signature": "",
  // 原始数据的 hash
  "hash": ""
}
```
- 返回示例:
```json
{
  "created": 1589478378,
  "data": [
    {
      "b64_json": "",
      "url": "https://...",
      "revised_prompt": "..."
    },
    {
      "b64_json": "",
      "url": "https://...",
      "revised_prompt": "..."
    }
  ]
}
```

### 文生图模型(使用项目名称)

**此接口从 v0.1.2 版本开始被弃用，从 v0.1.4 版本开始恢复使用。**

此接口用来调用文生图模型。Input 节点会选择一些运行指定项目和模型的 Worker 节点，根据策略(RTT连接时延或者GPU空闲值等)排序，依次向 Worker 节点发送模型请求直到获得正确的应答，失败次数过多时会报错。

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/image/gen/proxy
- 请求 Body:
```json
{
  // AI 项目名称
  "project": "SuperImage",
  // 模型名称
  "model": "superImage",
  // 所需图片的描述或者提示词
  "prompt": "a bird flying in the sky",
  // 要生成的图像数量，最少一个
  "n": 2,
  // 要生成图像的大小
  "size": "1024x1024",
  "width": 1024,
  "height": 1024,
  // 要生成图像的格式，必须是 url 或 b64_json 之一
  "response_format": "url",
  // 用户的钱包公钥
  "wallet": "",
  // 钱包签名
  "signature": "",
  // 原始数据的 hash
  "hash": ""
}
```
- 返回示例:
```json
{
  "created": 1589478378,
  "data": [
    {
      "b64_json": "",
      "url": "https://...",
      "revised_prompt": "..."
    },
    {
      "b64_json": "",
      "url": "https://...",
      "revised_prompt": "..."
    }
  ]
}
```

### 图生图模型

此接口用来调用图生图模型。

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/image/edit?node_id=xxx&cid=xxx&project=SuperImage&model=superImage
- 请求 Query 参数:
  - node_id: 运行模型的节点，Required
  - project: AI 项目名称，Required
  - model: 模型名称，Required
  - cid: 运行模型的容器 ID，Optional
  - wallet: 用户的钱包公钥
  - signature: 钱包签名
  - hash: 原始数据的 hash
- 请求 multipart/form-data：
  - image: 要编辑的 PNG 图片
  - prompt: 所需图像的文本描述
  - mask: 透明遮罩的附加 PNG 图片，可选参数
  - model: 想要请求的模型名称
  - n: 要生成的图像数量
  - size: 要生成图像的大小，例如 256x256、512x512 或者 1024x1024
  - response_format: 生成图像的格式，必须是 url 或 b64_json 之一
- 返回示例:
```json
{
  "created": 1589478378,
  "data": [
    {
      "b64_json": "",
      "url": "https://...",
      "revised_prompt": "..."
    },
    {
      "b64_json": "",
      "url": "https://...",
      "revised_prompt": "..."
    }
  ]
}
```

### 图生图模型(使用项目名称)

此接口用来调用图生图模型。Input 节点会选择一些运行指定项目和模型的 Worker 节点，根据策略(RTT连接时延或者GPU空闲值等)排序，依次向 Worker 节点发送模型请求直到获得正确的应答，失败次数过多时会报错。

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/image/edit/proxy?project=SuperImage&model=superImage
- 请求 Query 参数:
  - project: AI 项目名称，Required
  - model: 模型名称，Required
  - wallet: 用户的钱包公钥
  - signature: 钱包签名
  - hash: 原始数据的 hash
- 请求 multipart/form-data：
  - image: 要编辑的 PNG 图片
  - prompt: 所需图像的文本描述
  - mask: 透明遮罩的附加 PNG 图片，可选参数
  - model: 想要请求的模型名称
  - n: 要生成的图像数量
  - size: 要生成图像的大小，例如 256x256、512x512 或者 1024x1024
  - response_format: 生成图像的格式，必须是 url 或 b64_json 之一
- 返回示例:
```json
{
  "created": 1589478378,
  "data": [
    {
      "b64_json": "",
      "url": "https://...",
      "revised_prompt": "..."
    },
    {
      "b64_json": "",
      "url": "https://...",
      "revised_prompt": "..."
    }
  ]
}
```

### 获取 AI 项目列表

此接口用来查询分布式通信网络中运行的 AI 项目列表。

- 请求方式: GET
- 请求 URL: http://127.0.0.1:6000/api/v0/ai/projects/list?number=20
- 请求 Query 参数:
  - number: 正整数类型可选参数 - 表示想要查询的最大项目数量，默认值为 100
- 请求 Body: None
- 返回示例:
```json
{
  "data": [
    "DecentralGPT",
    "SuperImage"
  ]
}
```

### 获取 AI 项目的模型列表

此接口用来查询分布式通信网络中运行指定 AI 项目的模型列表。

- 请求方式: GET
- 请求 URL: http://127.0.0.1:6000/api/v0/ai/projects/models?project=DecentralGPT&number=20
- 请求 Query 参数:
  - project: AI 项目名称
  - number: 正整数类型可选参数 - 表示想要查询的最大模型数量，默认值为 100
- 请求 Body: None
- 返回示例:
```json
{
  "data": [
    "Qwen2-72B",
    "Llama3-70B"
  ]
}
```

### 获取运行指定 AI 项目和模型的节点列表

此接口用来查询分布式通信网络中运行指定 AI 项目和模型的节点列表。

- 请求方式: GET
- 请求 URL: http://127.0.0.1:6000/api/v0/ai/projects/peers?project=DecentralGPT&model=Qwen2.5-72B&number=20
- 请求 Query 参数:
  - project: AI 项目名称
  - model: 模型名称
  - number: 正整数类型可选参数 - 表示想要查询的最大节点数量，默认值为 20
- 请求 Body: None
- 返回示例:
```json
{
  "data": [
    // v0.1.2 版本及其之前版本，"data" 字段为节点 ID 组成的字符串数组
    // "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL",
    // "16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF"
    // v0.1.3 版本开始将 "data" 字段修改为结构体数组，添加了节点连接延迟时间等信息
    // "node_id" 表示节点 ID
    // "connectivity" 表示节点之间的连通性，有下列几种情况:
    // 0 - 未连接
    // 1 - 已建立连接
    // 2 - 可连接，最近连接过，但已被关闭
    // 3 - 最近尝试建立连接失败了
    // "latency" 表示节点连接的往返时延 RTT(Round-Trip Time)，int64 类型非负整数，以微秒为时间单位，有下列几种情况:
    // 0 - 未连接的节点，无法计算延迟时间，默认为 0
    // 正整数 - 正常的节点连接延迟时间
    // ps: 1 秒 = 1e3 毫秒 = 1e6 微秒 = 1e9 纳秒
    {
      "node_id": "16Uiu2HAmPKuJU5VE2PCnydyUn1VcTN2Lt59UDJFFEiRbb7h1x4CV",
      "connectivity": 1,
      "latency": 89121
    },
    {
      "node_id": "16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF",
      "connectivity": 0,
      "latency": 0
    }
  ]
}
```

## 模型注册/反注册接口

模型运行起来时需要向分布式网络通信节点注册，只有注册后的模型才能被分布式通信网络中的各个节点知晓和调用，在模型停止运行时，不要忘记反注册。

一个项目可以有多个模型，下面的 AI 模型注册/反注册接口一次只能操作一个模型，而 AI 项目注册/反注册接口一次可以操作多个模型。

假如一台机器有4个 GPU，为一个项目部署了4个相同的模型，每个模型使用1个 GPU，此时4个模型的项目名称和模型名称都一样，就需要 cid (Docker 容器 ID)来区分这个4个模型，因此分布式网络使用 `节点 ID`、`项目名称`、`模型名称` 和 `Docker 容器 ID` 区分和调用不同的模型。

### 注册 AI 模型

此接口用于接受 AI 模型的注册与更新，并将其在分布式网络节点间共享。

> [!NOTE]
> AI 模型和注册的节点必须在同一台机器上。

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/ai/model/register
- 请求 Body:
```json
{
  // AI 项目名称
  "project": "DecentralGPT",
  // AI 模型名称
  "model": "Llama3-70B",
  // 执行模型的 HTTP Url
  "api": "http://127.0.0.1:1042/v1/chat/completions",
  // 模型类型，默认 0
  // 0 - 文生文模型
  // 1 - 文生图模型
  // 2 - 图生图模型
  "type": 0,
  // docker 容器的 ID
  "cid": "d15c4007271b"
}
```
- 返回示例:
```json
{
  // 错误码，0 表示成功，非 0 表示失败
  "code": 0,
  // 错误信息
  "message": "ok"
}
```

### 反注册 AI 模型

此接口用于接受 AI 模型的反注册，并将其在分布式网络节点间共享。

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/ai/model/unregister
- 请求 Body:
```json
{
  // AI 项目名称
  "project": "DecentralGPT",
  // AI 模型名称
  "model": "Llama3-70B",
  // docker 容器的 ID
  "cid": "d15c4007271b"
}
```
- 返回示例:
```json
{
  // 错误码，0 表示成功，非 0 表示失败
  "code": 0,
  // 错误信息
  "message": "ok"
}
```

### 注册 AI 项目

此接口用于接受 AI 项目(可包含多个模型)的注册与更新，并将其在分布式网络节点间共享。

> [!NOTE]
> AI 模型和注册的节点必须在同一台机器上。

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/ai/project/register
- 请求 Body:
```json
{
  // AI 项目名称
  "project": "DecentralGPT",
  // AI 模型和 HTTP 接口信息列表
  "models": [
    {
      // 模型名称
      "model": "Llama3-70B",
      // 执行模型的 HTTP Url
      "api": "http://127.0.0.1:1042/v1/chat/completions",
      // 模型类型，默认 0
      // 0 - 文生文模型
      // 1 - 文生图模型
      // 2 - 图生图模型
      "type": 0,
      // docker 容器的 ID
      "cid": "d15c4007271b"
    }
  ]
}
```
- 返回示例:
```json
{
  // 错误码，0 表示成功，非 0 表示失败
  "code": 0,
  // 错误信息
  "message": "ok"
}
```

### 反注册 AI 项目

此接口用于接受 AI 项目的反注册，会将此项目包含所有模型全都反注册，并将其在分布式网络节点间共享。

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/ai/project/unregister
- 请求 Body:
```json
{
  // AI 项目名称
  "project": "DecentralGPT"
}
```
- 返回示例:
```json
{
  // 错误码，0 表示成功，非 0 表示失败
  "code": 0,
  // 错误信息
  "message": "ok"
}
```

### 查询任意节点注册的 AI 项目模型信息

此接口用于查询分布式通信网络中任意节点上注册的 AI 项目模型信息

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/ai/project/peer
- 请求 Body:
```json
{
  // 要查询的节点
  "node_id": "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL"
}
```
- 返回示例:
```json
{
  "data": {
    "DecentralGPT": [
      {
        "model": "Llama3-70B",
        "api": "http://127.0.0.1:1042/v1/chat/completions",
        "Type": 0,
        "cid": "d15c4007271b",
        "idle": 0
      }
    ]
  }
}
```

## 节点控制接口

控制节点连接和注册状态的接口。

### 列出建立连接的对等点

此接口用于查询与本节点建立连接的其他对等点信息。

- 请求方式: GET
- 请求 URL: http://127.0.0.1:6000/api/v0/swarm/peers
- 请求 Body: None
- 返回示例:
```json
[
  {
    "id": "16Uiu2HAmR-7",
    "peer": "16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
    "addr": "/ip4/122.99.183.54/tcp/6001",
    "latency": "208.878448ms",
    "direction": "Inbound"
  },
  {
    "id": "16Uiu2HAmD-11",
    "peer": "16Uiu2HAmDBYxgdKxeCbmn8hYiqwK3xHR9533WDdEYmpEDQ259GTe",
    "addr": "/ip4/116.233.45.153/tcp/7001",
    "latency": "0s",
    "direction": "Inbound"
  },
  {
    "id": "16Uiu2HAm5-5",
    "peer": "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL",
    "addr": "/ip4/122.99.183.53/tcp/7001",
    "latency": "0s",
    "direction": "Inbound"
  },
  {
    "id": "16Uiu2HAm4-3",
    "peer": "16Uiu2HAm49H3Hcae8rxKBdw8PfFcFAnBXQS8ierXA1VoZwhdDadV",
    "addr": "/ip4/122.99.183.54/tcp/7001",
    "latency": "0s",
    "direction": "Inbound"
  }
]
```

### 列出已知的连接地址

此接口用于查询本节点已知的其他节点的连接地址。

- 请求方式: GET
- 请求 URL: http://127.0.0.1:6000/api/v0/swarm/addrs
- 请求 Body: None
- 返回示例:
```json
[
  {
    "peer": "16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF",
    "addrs": [
      "/ip4/127.0.0.1/tcp/6001",
      "/ip4/172.19.236.172/tcp/6001",
      "/ip6/::1/tcp/6001",
      "/ip4/8.219.75.114/tcp/6001"
    ]
  },
  {
    "peer": "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL",
    "addrs": [
      "/ip4/122.99.183.53/tcp/39190",
      "/ip4/127.0.0.1/tcp/7001",
      "/ip4/192.168.122.228/tcp/7001",
      "/ip6/::1/tcp/7001"
    ]
  },
  {
    "peer": "16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
    "addrs": [
      "/ip4/122.99.183.54/tcp/6001",
      "/ip4/127.0.0.1/tcp/6001",
      "/ip6/::1/tcp/6001"
    ]
  },
  {
    "peer": "12D3KooWRjzt77yw6dNzRDaEvW2Acthy9nPVchQTn1DUCcsFgyKZ",
    "addrs": [
      "/ip4/122.99.183.54/tcp/37642"
    ]
  },
  {
    "peer": "12D3KooWQAW58SsJEVynUrY6J1ZfcAyTFJ6NX23sbV3NkivA773F",
    "addrs": [
      "/ip4/122.99.183.54/tcp/60654"
    ]
  },
  {
    "peer": "16Uiu2HAmDBYxgdKxeCbmn8hYiqwK3xHR9533WDdEYmpEDQ259GTe",
    "addrs": [
      "/ip4/116.233.45.153/tcp/7001",
      "/ip4/10.0.20.43/tcp/7001",
      "/ip4/127.0.0.1/tcp/7001",
      "/ip6/::1/tcp/7001",
      "/ip4/8.219.75.114/tcp/6001/p2p/16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF/p2p-circuit",
      "/ip4/122.99.183.54/tcp/6001/p2p/16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM/p2p-circuit"
    ]
  },
  {
    "peer": "16Uiu2HAm49H3Hcae8rxKBdw8PfFcFAnBXQS8ierXA1VoZwhdDadV",
    "addrs": [
      "/ip4/122.99.183.54/tcp/49442",
      "/ip4/127.0.0.1/tcp/7001",
      "/ip4/192.168.122.66/tcp/7001",
      "/ip6/::1/tcp/7001"
    ]
  },
  {
    "peer": "12D3KooWPfQEyrS9kBeqfDSMJZaxQAypd7j7sVnseHQp9kEq4575",
    "addrs": [
      "/ip4/122.99.183.54/tcp/44916"
    ]
  },
  {
    "peer": "12D3KooWLxer3X6qhvoB9KWxD34JtUNUdZiEHdw3Smi8oeiwuTng",
    "addrs": [
      "/ip4/122.99.183.54/tcp/50184"
    ]
  },
  {
    "peer": "16Uiu2HAmKk7Fg4WysTpEGd5q1wH2NL4wmxyQ5Nj4HhkQHyB3bDhm",
    "addrs": [
      "/ip4/124.78.244.129/tcp/7001"
    ]
  }
]
```

### 连接指定的节点

此接口用于手动连接指定的节点

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/swarm/connect
- 请求 Body:
```json
{
  // 节点连接地址
  "node_addr": "/ip4/10.0.20.43/tcp/7001/p2p/12D3KooWN7ZYLCxpr6T5FgFfXQDjoQYBT9jEucKq1ziF5dwT5Kzs"
}
```
- 返回示例:
```json
{
  // 错误码，0 表示成功，非 0 表示失败
  "code": 0,
  // 错误信息
  "message": "ok"
}
```

### 断开指定节点的连接

此接口用于手动断开与指定节点的连接

- 请求方式: POST
- 请求 URL: http://127.0.0.1:6000/api/v0/swarm/disconnect
- 请求 Body:
```json
{
  // 节点连接地址
  "node_addr": "/ip4/10.0.20.43/tcp/7001/p2p/12D3KooWN7ZYLCxpr6T5FgFfXQDjoQYBT9jEucKq1ziF5dwT5Kzs"
}
```
- 返回示例:
```json
{
  // 错误码，0 表示成功，非 0 表示失败
  "code": 0,
  // 错误信息
  "message": "ok"
}
```

### 列出订阅的节点列表

此接口用于查询订阅了相同主题的节点列表，仅限本节点已知的其他节点，因此不能作为查询所有节点列表的用途。

- 请求方式: GET
- 请求 URL: http://127.0.0.1:6000/api/v0/pubsub/peers
- 请求 Body: None
- 返回示例:
```json
[
  "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL",
  "16Uiu2HAm49H3Hcae8rxKBdw8PfFcFAnBXQS8ierXA1VoZwhdDadV",
  "16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
  "16Uiu2HAmDBYxgdKxeCbmn8hYiqwK3xHR9533WDdEYmpEDQ259GTe"
]
```

## 错误码

下面列出本程序所定义的常用错误码和错误信息，但不包括 AI 项目和模型自定义的错误码。

| 错误码 | 说明 |
| ---- | ---- |
| 1001 | 参数错误 |
| 1002 | JSON 等解析错误 |
| 1003 | JSON 序列化/反序列化错误 |
| 1004 | Protobuf 序列化/反序列化错误 |
| 1005 | 超时 |
| 1006 | Rendezvous 节点发现与路由错误 |
| 1007 | AI 模型错误 |
| 1008 | IPFS 上传/下载错误 |
| 1009 | 缓冲区错误 |
| 1010 | 权限错误 |
| 1011 | 暂不支持/暂未实现 |
| 1012 | 获取机器信息失败 |
| 1013 | 加密失败 |
| 1014 | 解密失败 |
| 1015 | UUID 错误 |
| 1016 | 数据库错误 |
| 1017 | 使用项目名称执行 AI 请求时找不到支持的节点或者所有节点全部报错 |
| 1018 | 文生文模型流式传输错误 |
| 1019 | 已弃用的功能 |
| .... | 预留以备未来扩充 |
| 5000 | 内部错误 |
