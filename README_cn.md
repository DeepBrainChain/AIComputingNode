# InferenceComputingNetwork

Distributed Inference Computing Network for AI

## 目录

- [`编译`](#编译)
- [`Protobuf`](#protobuf)
- [`命令行`](#命令行)
- [`节点部署`](#节点部署)
- [`工具`](#工具)
- [`文档`](#文档)

## 编译

```shell
$ go mod tidy
$ version=$(git describe --tags)
$ go build -ldflags "-X main.version=$version" -o host host/main.go
```

## Protobuf

```shell
# install to ~/go/bin/
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ protoc -I=./pkg/protocol/ --go_out=./pkg/protocol/ ./pkg/protocol/protocol.proto
protoc-gen-go: program not found or is not executable
--go_out: protoc-gen-go: Plugin failed with status code 1.
$ export PATH=$PATH:/home/dbtu/go/bin
$ protoc -I=./pkg/protocol/ --go_out=./pkg/protocol/ ./pkg/protocol/protocol.proto
```

## 命令行

`host [-h] [-config ./config.json] [-version] [-init mode] [-peerkey ./peer.key] [-psk]`

- h: 显示命令行帮助
- config: 使用指定的配置文件运行程序
- version: 显示版本号并退出
- init: 在 input/worker 模式下初始化和生成 JSON 配置文件
- peerkey: 根据指定的文件路径解析或生成密钥文件
- psk: 生成随机预共享密钥

```shell
$ host.exe -init worker
Generate peer key success at D:\Code\AIComputingNode\host\peer.key
Important notice: Please save this key file. You can use tools to retrieve the ID and private key in the future.
Encode private key: CAISINAckM6QODvCrez5I0Q3RZyo9PeV4jDeB1L71AHnSU/H
Encode public key: CAISIQNbA9ZWCwFM7X/eTUUBvwSRzTurMLkb9jg38wn5IRL4BQ==
Transform Peer ID: 16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y
Create datastore directory at D:\Code\AIComputingNode\host\datastore
Generate configuration success at D:\Code\AIComputingNode\host\worker.json
Run "host -config D:\Code\AIComputingNode\host\worker.json" command to start the program
$ 
$ host.exe -peerkey D:\Code\AIComputingNode\host\peer.key
Load peer key success
Encode private key: CAISINAckM6QODvCrez5I0Q3RZyo9PeV4jDeB1L71AHnSU/H
Encode public key: CAISIQNbA9ZWCwFM7X/eTUUBvwSRzTurMLkb9jg38wn5IRL4BQ==
Transform Peer ID: 16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y
$ 
$ host.exe -config D:\Code\AIComputingNode\host\worker.json
2024-04-24T10:44:07.065+0800	INFO	AIComputingNode	host/main.go:97	################################################################
2024-04-24T10:44:07.077+0800	INFO	AIComputingNode	host/main.go:98	#                          START                               #
2024-04-24T10:44:07.077+0800	INFO	AIComputingNode	host/main.go:99	################################################################
2024-04-24T10:44:07.123+0800	INFO	dht/RtRefreshManager	rtrefresh/rt_refresh_manager.go:322	starting refreshing cpl 0 with key CIQAAAUWCEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA (routing table size was 0)
2024-04-24T10:44:07.123+0800	WARN	dht/RtRefreshManager	rtrefresh/rt_refresh_manager.go:187	failed when refreshing routing table2 errors occurred:
	* failed to query for self, err=failed to find any peer in table
	* failed to refresh cpl=0, err=failed to find any peer in table


2024-04-24T10:44:07.128+0800	INFO	AIComputingNode	host/main.go:188	Listen addresses:[/ip4/10.0.20.21/tcp/7001 /ip4/127.0.0.1/tcp/7001 /ip6/::1/tcp/7001]
2024-04-24T10:44:07.128+0800	INFO	AIComputingNode	host/main.go:189	Node id:16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y
2024-04-24T10:44:07.168+0800	INFO	AIComputingNode	host/main.go:197	libp2p node address:[/ip4/10.0.20.21/tcp/7001/p2p/16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y /ip4/127.0.0.1/tcp/7001/p2p/16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y /ip6/::1/tcp/7001/p2p/16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y]
2024-04-24T10:44:07.169+0800	INFO	dht/RtRefreshManager	rtrefresh/rt_refresh_manager.go:322	starting refreshing cpl 0 with key CIQAAAIIMMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA (routing table size was 0)
2024-04-24T10:44:07.169+0800	WARN	dht/RtRefreshManager	rtrefresh/rt_refresh_manager.go:233	failed when refreshing routing table	{"error": "2 errors occurred:\n\t* failed to query for self, err=failed to find any peer in table\n\t* failed to refresh cpl=0, err=failed to find any peer in table\n\n"}
2024-04-24T10:44:07.169+0800	INFO	AIComputingNode	pubsub/tracer.go:48	Trace event: {Type: JOIN, PeerID: 16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y, Join{Topic: DeepBrainChain}}
2024-04-24T10:44:07.169+0800	INFO	AIComputingNode	host/main.go:343	listening for connections
2024-04-24T10:44:07.169+0800	INFO	AIComputingNode	serve/http.go:386	HTTP server is running on http://127.0.0.1:7002
```

## 节点类型说明

- Input 节点: 拥有公网 IP 地址的节点，用于接收项目 HTTP API 请求并返回结果的节点(做好域名解析会更好)，并且在分布式网络中承担节点发现和路由的功能。
- Worker 节点: 没有公网 IP 地址的节点，部署在模型所在的机器上，接收 Input 节点或者分布式网络中的发给自己的模型请求，转发给模型接口并将结果返回。
- Bootstrap 节点: 又名引导节点，是传承自 libp2p 和 IPFS 项目的概念，在本项目中等同于 Input 节点。

模型部署好后，需要将模型的调用接口和所属项目等信息注册到 Worker 节点，Worker 节点会将这些信息传播到整个分布式通信网络，然后就可以通过拥有域名或者公网 IP 地址的 Input 节点来调用部署的模型。

```
                  +----------------+                +----------------+             +----------------+
 HTTP Request     |    DBC AI      |                |    DBC AI      |             |                |
+----------------->                | libp2p stream  |                |  HTTP       |                |
                  |  Input Node    <---------------->  Worker Node   <------------->    AI Model    |
<-----------------+                |                |                | Req & Resp  |                |
  HTTP Response   |  CPU Machine   |                |  GPU Machine   |             |   GPU Machine  |
                  +----------------+                +----------------+             +----------------+
```

1. 拥有 GPU 服务器，需要部署 Worker 节点。
2. 拥有 CPU 机器且有域名/公网 IP 地址，可以部署 Input 节点。
3. 如果 GPU 机器本身有域名/公网 IP 地址，可以直接部署 Input 节点，向 Input 节点注册模型接口，不需要 Worker 节点。

## 节点部署

以下步骤描述了如何部署此分布式网络通信节点。

### 第一步，下载可执行程序

本项目会自动编译并上传最新的可执行程序，支持 Linux/Windows/macOS 系统。

在此处查看 [最新版本](https://github.com/DeepBrainChain/AIComputingNode/releases/latest)，并根据用户的系统下载相应的可执行程序。

下载到系统后，请确保可执行程序具有执行权限。

### 第二步，生成 JSON 配置文件

如果需要部署 Input 节点，请运行 `host -init input` 命令，该命令会在与程序同路径下生成名为 `input.json` 的配置文件。

如果需要部署 Worker 节点，请运行 `host -init worker` 命令，该命令会在与程序同路径下生成名为 `worker.json` 的配置文件。

生成 JSON 配置文件后，您需要在配置文件中添加一些 bootstrap 节点，确保开放的端口不冲突，API 符合您的需求。

详细配置项请参考 [JSON 配置文件](./docs/configuration_cn.md)。

### 第三步，使用 JSON 配置文件运行程序

假设上面生成的 JSON 配置文件名为 `worker.json`，则使用 `host -config ./worker.json` 运行程序即可。

## 工具

[Ipfs](./tools/ipfs/README.md)

## 文档

[AI 模型接口标准文档](./docs/model_api_cn.md)

[AIComputingNode HTTP API 接口文档](./docs/api_cn.md)

[JSON 配置文件](./docs/configuration_cn.md)
