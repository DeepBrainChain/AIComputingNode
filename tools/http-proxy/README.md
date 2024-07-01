# HTTP proxy service with libp2p

This example shows how to create a simple HTTP proxy service with libp2p:

```
                                                                                                    XXX
                                                                                                   XX  XXXXXX
                                                                                                  X         XX
                                                                                        XXXXXXX  XX          XX XXXXXXXXXX
                  +----------------+                +-----------------+              XXX      XXX            XXX        XXX
 HTTP Request     |                |                |                 |             XX                                    XX
+----------------->                | libp2p stream  |                 |  HTTP       X                                      X
                  |  Local peer    <---------------->  Remote peer    <------------->     HTTP SERVER - THE INTERNET      XX
<-----------------+                |                |                 | Req & Resp   XX                                   X
  HTTP Response   |  libp2p host   |                |  libp2p host    |               XXXX XXXX XXXXXXXXXXXXXXXXXXXX   XXXX
                  +----------------+                +-----------------+                                            XXXXX
```

In order to proxy an HTTP request, we create a local peer which listens on `localhost:8541`. HTTP requests performed to that address are tunneled via a libp2p stream to a remote peer, which then performs the HTTP requests and sends the response back to the local peer, which relays it to the user.

Note that this is a very simple approach to a proxy, and does not perform any header management, nor supports HTTPS. The `proxy.go` code is thoroughly commented, detailing what is happening in every step.

## Build

From the `tools` directory run the following:

```
> cd http-proxy/
> go build
```

## Usage

First run the "remote" peer as follows. It will print a local peer address. If you would like to run this on a separate machine, please replace the IP accordingly:

```sh
> ./http-proxy -l 8530 -peerkey ./peer1.key
2024/06/28 15:57:58 Load peer key success
2024/06/28 15:57:58 Listen addresses: [/ip4/10.0.20.21/tcp/8530 /ip4/127.0.0.1/tcp/8530]
2024/06/28 15:57:58 Node id: 16Uiu2HAm2u6o8MKSBhvPAFhEpnrFAGegxkWtPAy6aHfgcs8ZTM7u
2024/06/28 15:57:58 HTTP server is running on http://0.0.0.0: 8531
```

Then run the local peer, indicating that it will need to forward http requests to the remote peer as follows:

```sh
> ./http-proxy -l 8540 -peerkey ./peer2.key -d /ip4/10.0.20.21/tcp/8530/p2p/16Uiu2HAm2u6o8MKSBhvPAFhEpnrFAGegxkWtPAy6aHfgcs8ZTM7u
2024/06/28 17:56:45 Load peer key success
2024/06/28 17:56:45 Listen addresses: [/ip4/10.0.20.21/tcp/8540 /ip4/127.0.0.1/tcp/8540]
2024/06/28 17:56:45 Node id: 16Uiu2HAmKdTkftKbTB5QK5MkoPpyuoikjpSjCe78G1YjGTozEsMo
2024/06/28 17:56:45 HTTP server is running on http://0.0.0.0: 8541
```

As you can see, the proxy prints the listening address `127.0.0.1:8541`. You can now use this address as a proxy, for example with `curl`:

```sh
> curl -x "127.0.0.1:8541" "http://ipfs.io/p2p/QmfUX75pGRBRDnjeoMkQzuQczuCup2aYbeLxz5NzeSu9G6"
it works!

> curl -x "127.0.0.1:8541" "http://8.219.75.114:8081/api/v0/id"
{"peer_id":"16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF","protocol_version":"aicn/0.0.1","agent_version":"v0.1.0","addresses":["/ip4/127.0.0.1/tcp/6001","/ip4/172.19.236.172/tcp/6001","/ip6/::1/tcp/6001"],"protocols":["/ipfs/ping/1.0.0","/libp2p/circuit/relay/0.2.0/stop","/dbc/kad/1.0.0","/libp2p/autonat/1.0.0","/ipfs/id/1.0.0","/ipfs/id/push/1.0.0","/floodsub/1.0.0","/libp2p/circuit/relay/0.2.0/hop"]}

> curl -x "127.0.0.1:8541" "http://8.219.75.114:8081/api/v0/peer" -X POST --header "Content-Type: application/json" --data-raw "{\"node_id\": \"16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM\"}"
{"code":0,"message":"ok","data":{"peer_id":"16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM","protocol_version":"aicn/0.0.1","agent_version":"v0.1.0","addresses":["/ip4/122.99.183.54/tcp/6001","/ip4/127.0.0.1/tcp/6001","/ip6/::1/tcp/6001"],"protocols":["/ipfs/ping/1.0.0","/libp2p/circuit/relay/0.2.0/stop","/dbc/kad/1.0.0","/libp2p/autonat/1.0.0","/ipfs/id/1.0.0","/ipfs/id/push/1.0.0","/floodsub/1.0.0","/libp2p/circuit/relay/0.2.0/hop"]}}

> curl -x "127.0.0.1:8541" "http://122.99.183.52:1042/v1/chat/completions" -X POST --header "Content-Type: application/json" --data-raw "{\"model\": \"Llama3-8B\",\"messages\": [{\"role\": \"system\",\"content\": \"You are a helpful assistant.\"},{\"role\": \"user\",\"content\": \"Hello\"}]}"
{"object":"error","message":"The model `Llama3-8B` does not exist.","type":"NotFoundError","param":null,"code":404}

> curl -x "127.0.0.1:8541" "http://122.99.183.52:1042/v1/chat/completions" -X POST --header "Content-Type: application/json" --data-raw "{\"model\": \"Qwen2-72B\",\"messages\": [{\"role\": \"system\",\"content\": \"You are a helpful assistant.\"},{\"role\": \"user\",\"content\": \"Hello\"}]}"
{"id":"cmpl-949a4e63eb284d9fb141ed092d399c32","object":"chat.completion","created":1719569515,"model":"Qwen2-72B","choices":[{"index":0,"message":{"role":"assistant","content":"Hello! How can I assist you today?"},"logprobs":null,"finish_reason":"stop","stop_reason":null}],"usage":{"prompt_tokens":20,"total_tokens":30,"completion_tokens":10}}

> curl -x "127.0.0.1:8541" "http://122.99.183.52:1042/v1/chat/completions" -X POST --header "Content-Type: application/json" --data-raw "{\"model\": \"Qwen2-72B\",\"messages\": [{\"role\": \"system\",\"content\": \"You are a helpful assistant.\"},{\"role\": \"user\",\"content\": \"Hello\"}],\"stream\":true}"
data: {"id":"cmpl-aedb6bfb7eed4e0ea24d368705e90a9b","object":"chat.completion.chunk","created":1719569677,"model":"Qwen2-72B","choices":[{"index":0,"delta":{"role":"assistant"},"logprobs":null,"finish_reason":null}]}

data: {"id":"cmpl-aedb6bfb7eed4e0ea24d368705e90a9b","object":"chat.completion.chunk","created":1719569677,"model":"Qwen2-72B","choices":[{"index":0,"delta":{"content":"Hello"},"logprobs":null,"finish_reason":null}]}

data: {"id":"cmpl-aedb6bfb7eed4e0ea24d368705e90a9b","object":"chat.completion.chunk","created":1719569677,"model":"Qwen2-72B","choices":[{"index":0,"delta":{"content":"!"},"logprobs":null,"finish_reason":null}]}

data: {"id":"cmpl-aedb6bfb7eed4e0ea24d368705e90a9b","object":"chat.completion.chunk","created":1719569677,"model":"Qwen2-72B","choices":[{"index":0,"delta":{"content":" How"},"logprobs":null,"finish_reason":null}]}

data: {"id":"cmpl-aedb6bfb7eed4e0ea24d368705e90a9b","object":"chat.completion.chunk","created":1719569677,"model":"Qwen2-72B","choices":[{"index":0,"delta":{"content":" can"},"logprobs":null,"finish_reason":null}]}

data: {"id":"cmpl-aedb6bfb7eed4e0ea24d368705e90a9b","object":"chat.completion.chunk","created":1719569677,"model":"Qwen2-72B","choices":[{"index":0,"delta":{"content":" I"},"logprobs":null,"finish_reason":null}]}

data: {"id":"cmpl-aedb6bfb7eed4e0ea24d368705e90a9b","object":"chat.completion.chunk","created":1719569677,"model":"Qwen2-72B","choices":[{"index":0,"delta":{"content":" assist"},"logprobs":null,"finish_reason":null}]}

data: {"id":"cmpl-aedb6bfb7eed4e0ea24d368705e90a9b","object":"chat.completion.chunk","created":1719569677,"model":"Qwen2-72B","choices":[{"index":0,"delta":{"content":" you"},"logprobs":null,"finish_reason":null}]}

data: {"id":"cmpl-aedb6bfb7eed4e0ea24d368705e90a9b","object":"chat.completion.chunk","created":1719569677,"model":"Qwen2-72B","choices":[{"index":0,"delta":{"content":" today"},"logprobs":null,"finish_reason":null}]}

data: {"id":"cmpl-aedb6bfb7eed4e0ea24d368705e90a9b","object":"chat.completion.chunk","created":1719569677,"model":"Qwen2-72B","choices":[{"index":0,"delta":{"content":"?"},"logprobs":null,"finish_reason":null}]}

data: {"id":"cmpl-aedb6bfb7eed4e0ea24d368705e90a9b","object":"chat.completion.chunk","created":1719569677,"model":"Qwen2-72B","choices":[{"index":0,"delta":{"content":""},"finish_reason":"stop"}],"usage":{"prompt_tokens":20,"total_tokens":30,"completion_tokens":10}}

data: [DONE]


> curl -x "127.0.0.1:8541" "http://122.99.183.52:1042/v1/chat/completions" -X POST --header "Content-Type: application/json" --data-raw "{\"model\": \"Qwen2-72B\",\"messages\": [{\"role\": \"system\",\"content\": \"你是一名参加高考的高三学生\"},{\"role\": \"user\",\"content\": \"阅读下面的材料，根据要求写作。随着互联网的普及、人工智能的应用，越来越多的问题能很快得到答案。那么，我们的问题是否会越来越少？以上材料引发了你怎样的联想和思考？请写一篇文章。要求：选准角度，确定立意，明确文体，自拟标题；不要套作，不得抄袭；不得泄露个人信息；不少于800字。\"}],\"stream\":true}"
```
