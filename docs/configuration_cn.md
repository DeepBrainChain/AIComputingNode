# JSON 配置文件

本项目的 bootstrap 节点、开放端口、日志等均可在 JSON 文件中进行配置，本文详细介绍了所有配置项，并提供了两个示例。

可以使用命令来自动生成 JSON 配置文件，然后再手动向配置文件中添加引导节点即可。
如果要部署的机器拥有公网 IP 地址，即部署 Input 节点，请使用 `host -init input` 命令，否则请使用 `host -init worker` 命令部署 Worker 节点。

节点支持的 AI 项目列表虽然也保存在此配置文件中，但是推荐使用 HTTP API 接口来管理，不建议直接编辑配置文件来修改此项。

## 配置项

```json
{
  // 引导节点列表，用于节点发现和路由功能，需要部署在有公网 IP 地址的服务器上，最好带域名。
  "Bootstrap": [
    "/ip4/122.99.183.54/tcp/6001/p2p/16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
    "/ip4/8.219.75.114/tcp/6001/p2p/16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF"
  ],
  // 用于节点连接和通信的监听地址
  "Addresses": [
    "/ip4/0.0.0.0/tcp/6001",
    "/ip6/::/tcp/6001"
  ],
  // 提供 HTTP 服务的 API 接口
  "API": {
    // 以 "host:port" 形式提供，例如 "localhost:8080", "127.0.0.1:8080" 或者 "0.0.0.0:8080"，其中 "0.0.0.0:8080" 可以简写成 ":8080"。
    "Addr": "0.0.0.0:6000"
  },
  // 节点的身份信息，可以使用 `host -peerkey ./peer.key` 命令读取和生成，保存在 `./peer.key` 文件中，做好备份请勿删除。
  // 以下仅为示例，实际部署时请自行生成并填写。
  // PeerID 是公钥经过算法生成的，程序会验证 PeerID 与 PrivKey 是否匹配。
  "Identity": {
    // 节点的 NodeID，是分布式通信网络中标识不同节点的唯一 ID。
    "PeerID": "16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y",
    // 节点的私钥(private key)。
    "PrivKey": "CAISINAckM6QODvCrez5I0Q3RZyo9PeV4jDeB1L71AHnSU/H"
  },
  // 节点连接参数配置
  "Swarm": {
    // 连接管理器
    "ConnMgr": {
      // 连接管理器的类型
      // "basic" - 具有 "HighWater" 和 "LowWater" 等参数的基础管理器
      // "none" - 不使用连接管理器
      "Type": "basic",
      // GracePeriod 是新连接不会被连接管理器关闭的持续时间
      "GracePeriod": "60s",
      // 当使用基础连接管理器的节点达到 HighWater 个空闲连接时，它将关闭最没有用的连接，直到达到 LowWater 个空闲连接。
      "HighWater": 400,
      "LowWater": 100
    },
    // 禁用自动 NAT 端口转发
    // 当未禁用(默认)时，本地节点会要求 NAT 设备(例如路由器)打开外部端口并将其转发到正在运行的节点端口。
    // 当此操作有效时(即当您的路由器支持 NAT 端口转发时)，它使您的本地节点可以从公共互联网访问。
    "DisableNatPortMap": false,
    // 中继客户端配置，中继服务的的使用者
    "RelayClient": {
      // 是否为节点启用 "自动中继用户" 模式
      // 如果节点部署在有公网 IP 的服务器上，请禁用它，否则请启用它。
      "Enabled": true,
      "StaticRelays": []
    },
    // 中继服务配置，为网络中其他对等点提供中继服务
    "RelayService": {
      // 是否启用它为网络上的其他对等点提供 "/p2p-circuit" v2 中继服务
      // 如果节点部署在有公网 IP 的服务器上，请启用它，否则请禁用它。
      "Enabled": false
    },
    // 实验性配置
    // 当您无法进行 NAT 端口转发时，可以启用打洞以进行 NAT 穿透(默认禁用)。
    // 启用后，本地节点将使用中继连接与对方协商，尽可能升级到透过 NAT/防火墙的直接连接。
    // 此功能需要将 Swarm.RelayClient.Enabled 设置为 true。
    "EnableHolePunching": false,
    // 如果您想帮助其他对等点确定它们是否位于 NAT 后面，您也可以启动 AutoNAT 的服务器端(RelayClient 已经运行客户端)。
    // 这会将节点配置为向对等点提供服务以确定其可达性状态。
    // 如果节点部署在有公网 IP 的服务器上，请启用它，否则请禁用它。
    // 此服务具有严格的速率限制，不会导致任何性能问题。
    "EnableAutoNATService": true,
    // 拨打连接别的节点的超时时间，包括拨打原始网络连接、协议选择以及握手（如果适用）之间的时间。
    // 可以设置为 "5s" "10s" "15s" 等值，默认为空(使用 libp2p 库的默认值)。
    "DialTimeout": ""
  },
  // 发布和订阅配置
  "Pubsub": {
    // 默认启用
    "Enabled": true,
    // 路由类型，以下选项之一，默认 "gossipsub"
    // "floodsub" - 是一种基本路由，它只是简单地将消息泛洪到所有连接的对等点的。这种路由效率极低，但非常可靠，适用于节点较少的小型网络。
    // "gossipsub" - 是一种更高级的路由，具有 mesh 网络形式和八卦传播功能。适用于节点较多的大型网络。
    "Router": "gossipsub",
    // 在 "gossipsub" 中使用泛洪发布将消息发送给所有得分大于 "publishThreshold" 的对等点。
    "FloodPublish": true
  },
  // DHT 节点路由配置
  "Routing": {
    // 路由类型，以下选项之一
    // "dhtclient" - 客户端模式，适用于 NAT 后面的节点；
    // "dhtserver" - 服务器模式，适用于部署在有公网 IP 地址的服务器上的节点；
    // "auto" - 根据网络情况自动确定客户端模式或服务器模式；
    // "none" - 不使用 DHT 路由发现其他节点；
    "Type": "dhtclient",
    // 路由协议的前缀。不同协议的节点无法建立连接。
    "ProtocolPrefix": "/dbc"
  },
  // 应用程序配置
  "App": {
    // 日志级别, "debug"、"info"、"warn"、"error"、"panic" 和 "fatal" 之一
    "LogLevel": "info",
    // 日志文件路径，用于保存程序产生的日志记录
    "LogFile": "./test.log",
    // 日志输出模式，通过 "stderr", "file" 或者 "+" 自由组合
    // "stderr" - 输出到控制台
    // "file" - 输出到 "LogFile" 指定的文件
    // "stderr+file" - 同时输出到控制台和指定文件
    "LogOutput": "stderr+file",
    // 预生成共享密钥, 密钥不同的节点无法建立连接
    "PreSharedKey": "f504f536a912a8cf7d00adacee8ed20270c5040d961d7f3da4fccbcbec0ec48a",
    // 订阅的主题名称，用于集中式节点查询服务，例如查询网络中所有节点的 Node ID 列表。
    // 不同的主题名称可能会阻止其他人查询你部署的节点。
    "TopicName": "DeepBrainChain",
    // 数据存储文件夹持久化了成功连接的对等信息等信息，方便节点启动时快速重新连接网络。
    // 请提前创建此文件夹并确保本节点程序有读写权限。
    "Datastore": "./datastore",
    // 收集节点广播的心跳信息，包括支持的AI项目和模型。
    "PeersCollect": {
      // 仅当节点拥有公网 IP 地址，且开启收集功能时，节点信息才会保存到 leveldb 数据库中。
      "Enabled": false,
      // 心跳间隔
      "HeartbeatInterval": "180s",
      // 客户端节点所属的项目名称，例如 "DecentralGPT" and "SuperImage".
      // 该配置项用于为指定 AI 项目部署的专用客户端节点。
      // 默认值为空，表示公共的客户端节点，如果不为空，则本节点拒绝不属于本项目的模型节点的连接。
      "ClientProject": ""
    }
  },
  // 节点支持的 AI 项目列表，可使用 registration/unregistration 接口管理，但不推荐手动修改。
  "AIProjects": [
    {
      // 项目名称
      "Project": "SuperImage",
      // 模型列表，一个项目可以包含多个模型。
      "Models": [
        {
          // 模型名称
          "Model": "SuperImageAI",
          // 模型的访问接口，用于接收 HTTP 请求，根据请求中的提示词和其他参数运行模型，并返回生成的文本或图片。
          // 在执行 AI 模型的节点上这是必要的配置。
          "API": "http://127.0.0.1:1088/models",
          // 模型类型，以下选项之一
          // 0 - 文生文模型
          // 1 - 文生图模型
          // 2 - 图生图模型
          "Type": 1
        }
      ]
    }
  ]
}
```

## Worker 节点配置示例

```json
{
  "Bootstrap": [
    "/ip4/122.99.183.54/tcp/6001/p2p/16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
    "/ip4/8.219.75.114/tcp/6001/p2p/16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF"
  ],
  "Addresses": [
    "/ip4/0.0.0.0/tcp/6001",
    "/ip6/::/tcp/6001"
  ],
  "API": {
    "Addr": "127.0.0.1:6000"
  },
  "Identity": {
    "PeerID": "16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y",
    "PrivKey": "CAISINAckM6QODvCrez5I0Q3RZyo9PeV4jDeB1L71AHnSU/H"
  },
  "Swarm": {
    "ConnMgr": {
      "Type": "none",
      "GracePeriod": "60s",
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
    "EnableAutoNATService": true,
    "DialTimeout": ""
  },
  "Pubsub": {
    "Enabled": true,
    "Router": "gossipsub",
    "FloodPublish": true
  },
  "Routing": {
    "Type": "dhtclient",
    "ProtocolPrefix": "/dbc"
  },
  "App": {
    "LogLevel": "info",
    "LogFile": "./test.log",
    "LogOutput": "stderr+file",
    "PreSharedKey": "f504f536a912a8cf7d00adacee8ed20270c5040d961d7f3da4fccbcbec0ec48a",
    "TopicName": "DeepBrainChain",
    "Datastore": "./datastore",
    "PeersCollect": {
      "Enabled": false,
      "HeartbeatInterval": "180s",
      "ClientProject": ""
    }
  },
  "AIProjects": [
    {
      "Project": "SuperImage",
      "Models": [
        {
          "Model": "SuperImageAI",
          "API": "http://127.0.0.1:1088/models",
          "Type": 1
        }
      ]
    }
  ]
}
```

## Input 节点配置示例

```json
{
  "Bootstrap": [
    "/ip4/122.99.183.54/tcp/6001/p2p/16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
    "/ip4/8.219.75.114/tcp/6001/p2p/16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF"
  ],
  "Addresses": [
    "/ip4/0.0.0.0/tcp/6001",
    "/ip6/::/tcp/6001"
  ],
  "API": {
    "Addr": "0.0.0.0:6000"
  },
  "Identity": {
    "PeerID": "16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y",
    "PrivKey": "CAISINAckM6QODvCrez5I0Q3RZyo9PeV4jDeB1L71AHnSU/H"
  },
  "Swarm": {
    "ConnMgr": {
      "Type": "basic",
      "GracePeriod": "60s",
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
    "EnableAutoNATService": true,
    "DialTimeout": ""
  },
  "Pubsub": {
    "Enabled": true,
    "Router": "floodsub",
    "FloodPublish": true
  },
  "Routing": {
    "Type": "dhtserver",
    "ProtocolPrefix": "/dbc"
  },
  "App": {
    "LogLevel": "info",
    "LogFile": "./test.log",
    "LogOutput": "stderr+file",
    "PreSharedKey": "f504f536a912a8cf7d00adacee8ed20270c5040d961d7f3da4fccbcbec0ec48a",
    "TopicName": "DeepBrainChain",
    "Datastore": "./datastore",
    "PeersCollect": {
      "Enabled": true,
      "HeartbeatInterval": "180s",
      "ClientProject": ""
    }
  },
  "AIProjects": []
}
```
