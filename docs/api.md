# AIComputingNode HTTP API Interface Documentation

This document describes the HTTP API interface of the AIComputingNode distributed communication node.

Detailed test cases can be viewed on the Apifox platform: <https://xr03hymjol.apifox.cn>.

Common interface conventions:
1. When an HTTP request is successfully processed and returns the expected result, a status code of 200 OK is given.
2. When processing HTTP requests, if there are internal server problems such as database errors, logic errors, calculation errors, etc., a 500 Internal Server Error status code is usually returned instead of 200, and the following JSON is included to tell the client what error occurred, helping developers locate the problem.
```json
{
  "code": 1010,
  "message": "Unsupported function"
}
```

## Common query interfaces

Interface for querying common node information, node list or machine information.

### Get the node's own PeerInfo

PeerInfo is a structure containing basic information such as node ID, protocol, version and listening address.

Node ID is a unique ID that identifies a distributed communication node.

This interface is only used to query the PeerInfo of the node itself, without the participation of the distributed network.

- request method: GET
- request URL: http://127.0.0.1:6000/api/v0/id
- request Body: None
- return example:
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

### Get the node list

This interface is used to query the node list in the distributed communication network.

- request method: GET
- request URL: http://127.0.0.1:6000/api/v0/peers
- request Body: None
- return example:
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

Generally, the request response will have several fields such as code message and data.

code indicates the error code, 0 indicates success, non-0 indicates failure, and common error codes can be viewed at the end of the article.

message indicates error information.

data contains the result information of the interface request (valid when code = 0).

### Get PeerInfo of any node

This interface can be used to query the PeerInfo of any node in the distributed communication network.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/peer
- request Body:
```json
{
  // Node to be queried
  "node_id": "16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM"
}
```
- return example:
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

### Get the machine information of any node

This interface is used to query the machine hardware and software information of any node in the distributed communication network.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/host/info
- request Body:
```json
{
  // Node to be queried
  "node_id": "16Uiu2HAmDBYxgdKxeCbmn8hYiqwK3xHR9533WDdEYmpEDQ259GTe"
}
```
- return example:
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

## Model call interface

Interface for calling AI models.

### Text generation text model

This interface is used to call text to generate text models

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/chat/completion
- request Body:
```json
{
  // Node running the model
  "node_id": "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL",
  // AI project name
  "project": "DecentralGPT",
  // Model name you want to request
  "model": "Llama3-70B",
  // Preset system assistant behavior mode and alternating question and answer records
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
  // If this is set to true, the returned data will be streamed in increments of one message at a time, and the data stream ends with data: [DONE].
  "stream": false,
  // User’s wallet public key
  "wallet": "",
  // Wallet signature
  "signature": "",
  // Original data hash
  "hash": ""
}
```
- return example:
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

### Text generation text model(Use project name)

**This interface has been deprecated since v0.1.2 and has been restored since v0.1.4.**

This interface uses the project name to call the text-to-text model. The Input node selects some Worker nodes running the specified project and model, sort them according to the strategy (RTT connection latency or GPU idle value, etc.), and send model requests to the Worker nodes in turn until a correct response is obtained. If there are too many failures, an error will be reported.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/chat/completion/proxy
- request Body:
```json
{
  // AI project name
  "project": "DecentralGPT",
  // Model name you want to request
  "model": "Llama3-70B",
  // Preset system assistant behavior mode and alternating question and answer records
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
  // If this is set to true, the returned data will be streamed in increments of one message at a time, and the data stream ends with data: [DONE].
  "stream": false,
  // User’s wallet public key
  "wallet": "",
  // Wallet signature
  "signature": "",
  // Original data hash
  "hash": ""
}
```
- return example:
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

### Text generation image model

This interface is used to call the text-to-image model.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/image/gen
- request Body:
```json
{
  // Node running the model
  "node_id": "16Uiu2HAm49H3Hcae8rxKBdw8PfFcFAnBXQS8ierXA1VoZwhdDadV",
  // AI project name
  "project": "SuperImage",
  // Model name you want to request
  "model": "superImage",
  // Text description prompt words for the required image
  "prompt": "a bird flying in the sky",
  // The number of images to be generated, at least one
  "n": 2,
  // The size of the image to be generated
  "size": "1024x1024",
  "width": 1024,
  "height": 1024,
  // The format in which the generated images are returned. Must be one of url or b64_json
  "response_format": "url",
  // User’s wallet public key
  "wallet": "",
  // Wallet signature
  "signature": "",
  // Original data hash
  "hash": ""
}
```
- return example:
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

### Text generation image model(Use project name)

**This interface has been deprecated since v0.1.2 and has been restored since v0.1.4.**

This interface uses the project name to call the text-to-image model. The Input node selects some Worker nodes running the specified project and model, sort them according to the strategy (RTT connection latency or GPU idle value, etc.), and send model requests to the Worker nodes in turn until a correct response is obtained. If there are too many failures, an error will be reported.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/image/gen/proxy
- request Body:
```json
{
  // AI project name
  "project": "SuperImage",
  // Model name you want to request
  "model": "superImage",
  // Text description prompt words for the required image
  "prompt": "a bird flying in the sky",
  // The number of images to be generated, at least one
  "n": 2,
  // The size of the image to be generated
  "size": "1024x1024",
  "width": 1024,
  "height": 1024,
  // The format in which the generated images are returned. Must be one of url or b64_json
  "response_format": "url",
  // User’s wallet public key
  "wallet": "",
  // Wallet signature
  "signature": "",
  // Original data hash
  "hash": ""
}
```
- return example:
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

### Image generation image model

This interface is used to call the image-to-image model.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/image/edit?node_id=xxx&cid=xxx&project=SuperImage&model=superImage
- request Query parameters:
  - node_id: Node running the model, Required
  - project: AI project name, Required
  - model: Model name you want to request, Required
  - cid: Container ID running the model, Optional
  - wallet: User’s wallet public key
  - signature: Wallet signature
  - hash: Original data hash
- request multipart/form-data：
  - image: Image to be edited
  - prompt: Text description prompt words for the required image
  - mask: An additional image whose fully transparent areas (e.g. where alpha is zero) indicate where image should be edited. Must be a valid PNG file, optional
  - model: Model name you want to request
  - n: The number of images to be generated, at least one
  - size: The size of the image to be generated, such as 256x256, 512x512 or 1024x1024
  - response_format: The format in which the generated images are returned. Must be one of url or b64_json
- return example:
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

### Image generation image model(Use project name)

This interface uses the project name to call the image-to-image model. The Input node selects some Worker nodes running the specified project and model, sort them according to the strategy (RTT connection latency or GPU idle value, etc.), and send model requests to the Worker nodes in turn until a correct response is obtained. If there are too many failures, an error will be reported.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/image/edit/proxy?project=SuperImage&model=superImage
- request Query parameters:
  - project: AI project name, Required
  - model: Model name you want to request, Required
  - wallet: User’s wallet public key
  - signature: Wallet signature
  - hash: Original data hash
- request multipart/form-data：
  - image: Image to be edited
  - prompt: Text description prompt words for the required image
  - mask: An additional image whose fully transparent areas (e.g. where alpha is zero) indicate where image should be edited. Must be a valid PNG file, optional
  - model: Model name you want to request
  - n: The number of images to be generated, at least one
  - size: The size of the image to be generated, such as 256x256, 512x512 or 1024x1024
  - response_format: The format in which the generated images are returned. Must be one of url or b64_json
- return example:
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

### Get AI project list

This interface is used to query the list of AI projects running in the distributed communication network.

- request method: GET
- request URL: http://127.0.0.1:6000/api/v0/ai/projects/list?number=20
- request Query parameters:
  - number: positive integer type optional parameter - Indicates the maximum number of projects you want to query, the default value is 100
- request Body: None
- return example:
```json
{
  "data": [
    "DecentralGPT",
    "SuperImage"
  ]
}
```

### Get the model list of AI projects

This interface is used to query the model list of the specified AI project running in the distributed communication network.

- request method: GET
- request URL: http://127.0.0.1:6000/api/v0/ai/projects/models?project=DecentralGPT&number=20
- request Query parameters:
  - project: AI project name
  - number: positive integer type optional parameter - Indicates the maximum number of models you want to query, the default value is 100
- request Body: None
- return example:
```json
{
  "data": [
    "Qwen2-72B",
    "Llama3-70B"
  ]
}
```

### Get the node list running the specified AI project and model

This interface is used to query the list of nodes running the specified AI project and model in the distributed communication network.

- request method: GET
- request URL: http://127.0.0.1:6000/api/v0/ai/projects/peers?project=DecentralGPT&model=Qwen2.5-72B&number=20
- request Query parameters:
  - project: AI project name
  - model: Model name
  - number: positive integer type optional parameter - Indicates the maximum number of nodes you want to query, the default value is 20
- request Body: None
- return example:
```json
{
  "data": [
    // For versions v0.1.2 and earlier, the "data" field is a string array consisting of node IDs.
    // "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL",
    // "16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF"
    // Starting from version v0.1.3, the "data" field was changed to a structure array, and
    // information such as node connection latency time was added.
    // "node_id" indicates the node ID
    // "connectivity" indicates the connectivity between nodes, and there are the following cases:
    // 0 - NotConnected, means no connection to peer
    // 1 - Connected, means has an open, live connection to peer
    // 2 - CanConnect, means recently connected to peer, terminated gracefully
    // 3 - CannotConnect, means recently attempted connecting but failed to connect
    // "latency" indicates the round-trip time (RTT) of the node connection, int64 type non-negative
    // integer, in microseconds as the time unit, there are the following cases:
    // 0 - unconnected node, latency time cannot be calculated, default is 0
    // Positive integer - normal node connection latency time
    // ps: 1 second = 1e3 milliseconds = 1e6 microseconds = 1e9 nanoseconds
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

## Model registration/deregistration interface

When the model is running, it needs to be registered with the distributed network communication node. Only the registered model can be known and called by each node in the distributed communication network. When the model stops running, don't forget to deregister.

A project can have multiple models. The following AI model registration/unregistration interface can only operate one model at a time, while the AI ​​project registration/unregistration interface can operate multiple models at a time.

If a machine has 4 GPUs, and 4 identical models are deployed for a project, each model uses 1 GPU. At this time, the project name and model name of the 4 models are the same, and cid (Docker container ID) is needed to distinguish the 4 models. Therefore, the distributed network uses `node ID`, `project name`, `model name` and `Docker container ID` to distinguish and call different models.

### Register AI model

This interface is used to accept the registration and update of AI models and share them among distributed network nodes.

> [!NOTE]
> The AI ​​model and the registered node must be on the same machine.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/ai/model/register
- request Body:
```json
{
  // AI project name
  "project": "DecentralGPT",
  // AI model name
  "model": "Llama3-70B",
  // HTTP Url for executing model
  "api": "http://127.0.0.1:1042/v1/chat/completions",
  // Model type, default 0
  // 0 - Text generation text model
  // 1 - Text generation image model
  // 2 - Image editing model
  "type": 0,
  // docker container ID
  "cid": "d15c4007271b"
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok"
}
```

### Unregister AI model

This interface is used to accept the deregistration of AI models and share them among distributed network nodes.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/ai/model/unregister
- request Body:
```json
{
  // AI project name
  "project": "DecentralGPT",
  // AI model name
  "model": "Llama3-70B",
  // docker container ID
  "cid": "d15c4007271b"
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok"
}
```

### Register AI project

This interface is used to accept registration and update of AI projects (which can contain multiple models) and share them among distributed network nodes.

> [!NOTE]
> The AI ​​model and the registered node must be on the same machine.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/ai/project/register
- request Body:
```json
{
  // AI project name
  "project": "DecentralGPT",
  // List of AI model and HTTP interface information
  "models": [
    {
      // Model name
      "model": "Llama3-70B",
      // HTTP Url for executing model
      "api": "http://127.0.0.1:1042/v1/chat/completions",
      // Model type, default 0
      // 0 - Text generation text model
      // 1 - Text generation image model
      // 2 - Image editing model
      "type": 0,
      // Docker container ID
      "cid": "d15c4007271b"
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
  "message": "ok"
}
```

### Unregister AI project

This interface is used to accept the deregistration of AI projects, which will deregister all models contained in this project and share them among distributed network nodes.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/ai/project/unregister
- request Body:
```json
{
  // AI project name
  "project": "DecentralGPT"
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok"
}
```

### Query the model information of AI projects registered on any node

This interface is used to query the model information of AI projects registered on any node in the distributed communication network.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/ai/project/peer
- request Body:
```json
{
  // Node to be queried
  "node_id": "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL"
}
```
- return example:
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

## Node control interface

Interface for controlling node connection and registration status.

### Lists peers with established connections

This interface is used to query information about other peers that have established connections with this node.

- request method: GET
- request URL: http://127.0.0.1:6000/api/v0/swarm/peers
- request Body: None
- return example:
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

### List known connection addresses

This interface is used to query the connection addresses of other nodes known to this node.

- request method: GET
- request URL: http://127.0.0.1:6000/api/v0/swarm/addrs
- request Body: None
- return example:
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

### Connect to a specified node

This interface is used to manually connect to a specified node.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/swarm/connect
- request Body:
```json
{
  // Node connection address
  "node_addr": "/ip4/10.0.20.43/tcp/7001/p2p/12D3KooWN7ZYLCxpr6T5FgFfXQDjoQYBT9jEucKq1ziF5dwT5Kzs"
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok"
}
```

### Disconnect from a specified node

This interface is used to manually disconnect from a specified node.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/swarm/disconnect
- request Body:
```json
{
  // Node connection address
  "node_addr": "/ip4/10.0.20.43/tcp/7001/p2p/12D3KooWN7ZYLCxpr6T5FgFfXQDjoQYBT9jEucKq1ziF5dwT5Kzs"
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok"
}
```

### List a list of subscribed nodes

This interface is used to query the list of nodes subscribed to the same topic, limited to other nodes known to this node, so it cannot be used to query the list of all nodes.s

- request method: GET
- request URL: http://127.0.0.1:6000/api/v0/pubsub/peers
- request Body: None
- return example:
```json
[
  "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL",
  "16Uiu2HAm49H3Hcae8rxKBdw8PfFcFAnBXQS8ierXA1VoZwhdDadV",
  "16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
  "16Uiu2HAmDBYxgdKxeCbmn8hYiqwK3xHR9533WDdEYmpEDQ259GTe"
]
```

## Error code

The following lists the common error codes and error messages defined by this program, but does not include error codes customized by AI projects and models.

| Error Code | Description |
| ---- | ---- |
| 1001 | Parameter error |
| 1002 | JSON and other parsing errors |
| 1003 | JSON serialization/deserialization errors |
| 1004 | Protobuf serialization/deserialization errors |
| 1005 | Timeout |
| 1006 | Rendezvous node discovery and routing errors |
| 1007 | AI model error |
| 1008 | IPFS upload/download errors |
| 1009 | Buffer errors |
| 1010 | Permission errors |
| 1011 | Not supported/not implemented yet |
| 1012 | Failed to obtain machine information |
| 1013 | Encryption failed |
| 1014 | Decryption failed |
| 1015 | UUID error |
| 1016 | Database error |
| 1017 | Unable to find supported nodes or all nodes report errors when executing AI requests using project names |
| 1018 | Stream error for text-to-text model |
| 1019 | Deprecated functions |
| .... | Reserved for future expansion |
| 5000 | Internal error |
