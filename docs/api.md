# AIComputingNode HTTP API Interface Documentation

This document describes the HTTP API interface of the AIComputingNode distributed communication node.

Detailed test cases can be viewed on the Apifox platform: <https://xr03hymjol.apifox.cn>.

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
  "code": 0,
  "message": "ok",
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
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok",
  "data": {
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
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok",
  "data": {
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
  "stream": false
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok",
  "data": {
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
}
```

### ~~Text generation text model(Use project name)~~

**This interface will be deprecated.**

This interface uses the project name to call the text-to-text model. The client will select some nodes running the specified project and model according to the strategy, send model requests to them respectively, and select a response with the correct result.

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
  ]
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok",
  "data": {
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
  "n": 1,
  // The size of the image to be generated
  "size": "1024x1024",
  // IPFS storage node, each project can deploy its own IPFS storage server
  "ipfs_node": ""
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok",
  "data": {
    "ipfs_node": "http://122.99.183.54:4002",
    "choices": [
      {
        "cid": "QmUsxJhQ13Gifj2iX9kTfzYDD6VsCL7UkdjqumPd9V2vKz",
        "image_name": "knhfkmpilha9l5f2.png"
      }
    ]
  }
}
```

### ~~Text generation image model(Use project name)~~

**This interface will be deprecated.**

This interface uses the project name to call the text-to-image model. The client will select some nodes running the specified project and model according to the strategy, send model requests to them respectively, and select a response with the correct result.

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
  "n": 1,
  // The size of the image to be generated
  "size": "1024x1024",
  // IPFS storage node, each project can deploy its own IPFS storage server
  "ipfs_node": ""
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok",
  "data": {
    "ipfs_node": "http://122.99.183.54:4002",
    "choices": [
      {
        "cid": "QmUsxJhQ13Gifj2iX9kTfzYDD6VsCL7UkdjqumPd9V2vKz",
        "image_name": "knhfkmpilha9l5f2.png"
      }
    ]
  }
}
```

### Get AI project list

This interface is used to query the list of AI projects running in the distributed communication network.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/ai/projects/list
- request Body: None
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok",
  "data": [
    "DecentralGPT",
    "SuperImage"
  ]
}
```

### Get the model list of AI projects

This interface is used to query the model list of the specified AI project running in the distributed communication network.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/ai/projects/models
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
  "message": "ok",
  "data": [
    "Qwen2-72B",
    "Llama3-70B"
  ]
}
```

### Get the node list running the specified AI project and model

This interface is used to query the list of nodes running the specified AI project and model in the distributed communication network.

- request method: POST
- request URL: http://127.0.0.1:6000/api/v0/ai/projects/peers
- request Body:
```json
{
  // AI project name
  "project": "DecentralGPT",
  // Model name
  "model": "Llama3-70B"
}
```
- return example:
```json
{
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok",
  "data": [
    "16Uiu2HAm5cygUrKCBxtNSMKKvgdr1saPM6XWcgnPyTvK4sdrARGL",
    "16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF"
  ]
}
```

## Node control interface

Interface for controlling node connection and registration status.

### Lists peers with established connections

This interface is used to query information about other peers that have established connections with this node.

- request method: POST
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

- request method: POST
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

- request method: POST
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

### Register AI project

This interface is used to accept registration and updates of AI projects and models, and share them among distributed network nodes.

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
      "type": 0
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

This interface is used to accept the unregistration of AI projects and models, and share them among distributed network nodes.

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

### Query the registration information of AI project model of any node

This interface is used to query the AI ​​project model registration information of any node in the distributed communication network.

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
  // Error code, 0 means success, non-0 means failure
  "code": 0,
  // Error message
  "message": "ok",
  "data": [
    {
      "project": "DecentralGPT",
      "models": [
        "Llama3-70B"
      ]
    }
  ]
}
```

## Error code

The following lists the common error codes and error messages defined by this program, but does not include error codes customized by AI projects and models.

| Error Code | Description |
| ---- | ---- |
| 1001 | Parameter error |
| 1002 | JSON and other parsing errors |
| 1003 | Protobuf serialization/deserialization errors |
| 1004 | Timeout |
| 1005 | Rendezvous node discovery and routing errors |
| 1006 | AI model error |
| 1007 | IPFS upload/download errors |
| 1008 | Buffer errors |
| 1009 | Permission errors |
| 1010 | Not supported/not implemented yet |
| 1011 | Failed to obtain machine information |
| 1012 | Encryption failed |
| 1013 | Decryption failed |
| 1014 | UUID error |
| 1015 | Database error |
| 1016 | Unable to find supported nodes or all nodes report errors when executing AI requests using project names |
| 1017 | Stream error for text-to-text model |
| .... | Reserved for future expansion |
| 5000 | Internal error |
