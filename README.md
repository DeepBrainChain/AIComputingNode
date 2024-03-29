# InferenceComputingNetwork

Distributed Inference Computing Network for Text-to-Image Generation

## Table of Contents

- [`Compiling`](#compiling)
- [`Protobuf`](#protobuf)
- [`Command Line`](#command-line)
- [`Configuration`](#configuration)
  - [`Client Configuration Example`](#client-configuration-example)
  - [`Server Configuration Example`](#server-configuration-example)
- [`Tools`](#tools)
  - [`PeerKey`](#peerkey)
- [`HTTP API`](#http-api)

## Compiling

```shell
$ go mod tidy
$ version=$(git describe --tags)
$ go build -ldflags "-X main.version=$version" -o host host\main.go
```

## Protobuf

```shell
# 安装到 ~/go/bin/
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ protoc -I=./pkg/protocol/ --go_out=./pkg/protocol/ ./pkg/protocol/protocol.proto 
protoc-gen-go: program not found or is not executable
--go_out: protoc-gen-go: Plugin failed with status code 1.
$ export PATH=$PATH:/home/dbtu/go/bin
$ protoc -I=./pkg/protocol/ --go_out=./pkg/protocol/ ./pkg/protocol/protocol.proto 
```

## Command Line

```shell
$ host -config ./config.json [-version]
```

- config: Specify the json configuration file path of the program
- version: Show version number and exit

## Configuration

```json
{
  // 引导节点列表，用于节点路由和发现功能，需要在公网服务器部署，最好有域名
  "Bootstrap": [
    "/ip4/122.99.183.54/tcp/6001/p2p/12D3KooWSpgWzEXE5GNjY6hgdAhuuBLe4d3ocqWDnVLdCa8U3cig",
    "/ip4/82.157.50.32/tcp/6001/p2p/12D3KooWFrTcDtocZWEvEAk2X4poyn13LzT3G7JMBRoPD73YPAoB"
  ],
  // 节点连接通信的监听地址
  "Addresses": [
    "/ip4/0.0.0.0/tcp/6001",
    "/ip6/::/tcp/6001"
  ],
  // 提供 HTTP 服务的 API
  "API": {
    // 以冒号:端口号的形式提供
    "Addr": ":6000"
  },
  // 节点的身份识别信息，由 peer-key 工具生成，下面提供的只是示例，实际部署时请自行生成和填写
  // PeerID 由公钥经算法生成，程序会验证 PeerID 与 PrivKey 是否匹配
  "Identity": {
    // 节点的 Node ID
    "PeerID": "12D3KooWLzae3a7n7eJSQbSgARHQizJzChS4Rtpo3gziugpV4vRz",
    // 节点的私钥
    "PrivKey": "CAESQBOuYLpTajlQ0Xzh84i6ey0Zot24Tc2qA2yDgw7dd2vGpg53BtN+RQgJ2erx2lRgDo4BiOXy0ZFRRxIZRQ5wJ3U="
  },
  // 配置节点连接参数
  "Swarm": {
    // 连接管理器
    "ConnMgr": {
      // 连接管理器类型
      // "basic" - 带有 "HighWater" "LowWater" 等参数的基础管理器
      // "none" - 不使用连接管理器
      "Type": "basic",
      // 如果节点不能在宽限期内重新连接，其数据将被删除。
      "GracePeriod": "20s",
      // 控制保持连接的节点数量，当连接数超过 HighWater 时，部分节点连接将被终止，直到保留 LowWater。
      "HighWater": 400,
      "LowWater": 100
    },
    // 对于 NAT 后的节点，是否禁用 uPNP。
    "DisableNatPortMap": false,
    // 中继客户端: 公网服务器上部署的节点请禁用，反之则开启
    "RelayClient": {
      // 是否启动中继客户端
      "Enabled": true,
      "StaticRelays": []
    },
    // 中继服务端: 公网服务器上部署的节点请开启，反之则禁用
    "RelayService": {
      // 是否启动中继服务
      "Enabled": false
    },
    // 启用 P2P 打洞，实验性功能
    "EnableHolePunching": false,
    // 帮助其他节点确定它们是否位于 NAT 之后，如果是公网服务器上部署的节点请务必开启
    // 该服务的速率受到严格限制，不会导致任何性能问题。
    "EnableAutoNATService": true
  },
  // 节点路由
  "Routing": {
    // 路由类型
    // "dhtclient" - 客户端模式，适用于 NAT 之后的节点
    // "dhtserver" - 服务器模式，适用于公网服务器上部署的节点
    // "auto" - 根据网络条件自动判断客户端模式或者服务器模式
    // "none" - 不使用 DHT 路由来发现其他节点
    "Type": "dhtclient",
    // 路由协议的前缀，协议不同的节点无法建立连接。
    "ProtocolPrefix": "/DeepBrainChain"
  },
  // 应用程序配置
  "app": {
    // 日志等级，选择 "debug"、"info"、"warn"、"error"、"panic"、"fatal" 中的一个
    "LogLevel": "info",
    // 日志文件路径，用来保存程序生成的日志记录
    "LogFile": "./test.log",
    // 日志输出模式，由 "stderr"、"file" 或者 "+" 自由组合
    // "stderr" - 输出到控制台
    // "file" - 输出到由 "LogFile" 指定的文件中
    // "stderr+file" - 同时输出到控制台和指定的文件中
    "LogOutput": "stderr+file",
    // 预生成的共享密钥，密钥不同的节点无法建立连接
    "PreSharedKey": "f504f536a912a8cf7d00adacee8ed20270c5040d961d7f3da4fccbcbec0ec48a",
    // 订阅的主题名称，用于集中式的节点查询服务，例如查询网络中所有的节点 Node ID 列表。
    // 主题名称不同可能会导致别人无法查询到您部署的节点。
    "TopicName": "SuperImageAI"
  }
}
```

### Client Configuration Example

```json
{
  "Bootstrap": [
    "/ip4/122.99.183.54/tcp/6001/p2p/12D3KooWSpgWzEXE5GNjY6hgdAhuuBLe4d3ocqWDnVLdCa8U3cig",
    "/ip4/82.157.50.32/tcp/6001/p2p/12D3KooWFrTcDtocZWEvEAk2X4poyn13LzT3G7JMBRoPD73YPAoB"
  ],
  "Addresses": [
    "/ip4/0.0.0.0/tcp/6001",
    "/ip6/::/tcp/6001"
  ],
  "API": {
    "Addr": ":6000"
  },
  "Identity": {
    "PeerID": "12D3KooWLzae3a7n7eJSQbSgARHQizJzChS4Rtpo3gziugpV4vRz",
    "PrivKey": "CAESQBOuYLpTajlQ0Xzh84i6ey0Zot24Tc2qA2yDgw7dd2vGpg53BtN+RQgJ2erx2lRgDo4BiOXy0ZFRRxIZRQ5wJ3U="
  },
  "Swarm": {
    "ConnMgr": {
      "Type": "none",
      "GracePeriod": "20s",
      "HighWater": 400,
      "LowWater": 100
    },
    "DisableNatPortMap": false,
    "RelayClient": {
      "Enabled": true,
      "StaticRelays": []
    },
    "RelayService": {
      "Enabled": false
    },
    "EnableHolePunching": true,
    "EnableAutoNATService": true
  },
  "Routing": {
    "Type": "dhtclient",
    "ProtocolPrefix": "/DeepBrainChain"
  },
  "app": {
    "LogLevel": "info",
    "LogFile": "./test.log",
    "LogOutput": "stderr+file",
    "PreSharedKey": "f504f536a912a8cf7d00adacee8ed20270c5040d961d7f3da4fccbcbec0ec48a",
    "TopicName": "SuperImageAI"
  }
}
```

### Server Configuration Example

```json
{
  "Bootstrap": [
    "/ip4/122.99.183.54/tcp/6001/p2p/12D3KooWSpgWzEXE5GNjY6hgdAhuuBLe4d3ocqWDnVLdCa8U3cig",
    "/ip4/82.157.50.32/tcp/6001/p2p/12D3KooWFrTcDtocZWEvEAk2X4poyn13LzT3G7JMBRoPD73YPAoB"
  ],
  "Addresses": [
    "/ip4/0.0.0.0/tcp/6001",
    "/ip6/::/tcp/6001"
  ],
  "API": {
    "Addr": ":6000"
  },
  "Identity": {
    "PeerID": "12D3KooWLzae3a7n7eJSQbSgARHQizJzChS4Rtpo3gziugpV4vRz",
    "PrivKey": "CAESQBOuYLpTajlQ0Xzh84i6ey0Zot24Tc2qA2yDgw7dd2vGpg53BtN+RQgJ2erx2lRgDo4BiOXy0ZFRRxIZRQ5wJ3U="
  },
  "Swarm": {
    "ConnMgr": {
      "Type": "basic",
      "GracePeriod": "20s",
      "HighWater": 400,
      "LowWater": 100
    },
    "DisableNatPortMap": true,
    "RelayClient": {
      "Enabled": false,
      "StaticRelays": []
    },
    "RelayService": {
      "Enabled": true
    },
    "EnableHolePunching": false,
    "EnableAutoNATService": true
  },
  "Routing": {
    "Type": "dhtserver",
    "ProtocolPrefix": "/DeepBrainChain"
  },
  "app": {
    "LogLevel": "info",
    "LogFile": "./test.log",
    "LogOutput": "stderr+file",
    "PreSharedKey": "f504f536a912a8cf7d00adacee8ed20270c5040d961d7f3da4fccbcbec0ec48a",
    "TopicName": "SuperImageAI"
  }
}
```

## Tools

### PeerKey

读取命令行参数 `-peerkey` 指定的密钥文件路径，解析并展示密钥中的私钥/公钥，并转换成对应的 PeerID。

如果命令行参数 `-peerkey` 指定的密钥文件不存在，则在此文件路径创建一个新的密钥。

```shell
$ ./peer-key.exe -peerkey ./peer.key
2024/03/28 18:58:12 Generate peer key success
2024/03/28 18:58:12 Encode private key: CAESQBOuYLpTajlQ0Xzh84i6ey0Zot24Tc2qA2yDgw7dd2vGpg53BtN+RQgJ2erx2lRgDo4BiOXy0ZFRRxIZRQ5wJ3U=
2024/03/28 18:58:12 Encode public key: CAESIKYOdwbTfkUICdnq8dpUYA6OAYjl8tGRUUcSGUUOcCd1
2024/03/28 18:58:12 Transform Peer ID: 12D3KooWLzae3a7n7eJSQbSgARHQizJzChS4Rtpo3gziugpV4vRz
$ ./peer-key.exe -peerkey ./peer.key
2024/03/29 09:59:55 Load peer key success
2024/03/29 09:59:55 Encode private key: CAESQBOuYLpTajlQ0Xzh84i6ey0Zot24Tc2qA2yDgw7dd2vGpg53BtN+RQgJ2erx2lRgDo4BiOXy0ZFRRxIZRQ5wJ3U=
2024/03/29 09:59:55 Encode public key: CAESIKYOdwbTfkUICdnq8dpUYA6OAYjl8tGRUUcSGUUOcCd1
2024/03/29 09:59:55 Transform Peer ID: 12D3KooWLzae3a7n7eJSQbSgARHQizJzChS4Rtpo3gziugpV4vRz
```

## HTTP API

[Rest HTTP API](./api.http)

开发者在 Apifox 中邀请你加入团队 DBC https://app.apifox.com/invite?token=XJYVYIvuR6VDnQCKJfIqV
