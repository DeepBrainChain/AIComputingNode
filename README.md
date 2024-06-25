# InferenceComputingNetwork

Distributed Inference Computing Network for AI

## Table of Contents

- [`Compiling`](#compiling)
- [`Protobuf`](#protobuf)
- [`Command Line`](#command-line)
- [`Configuration`](#configuration)
  - [`Client Configuration Example`](#client-configuration-example)
  - [`Server Configuration Example`](#server-configuration-example)
- [`Tools`](#tools)
  - [`Ipfs`](#ipfs)
- [`Document`](#document)

## Compiling

```shell
$ go mod tidy
$ version=$(git describe --tags)
$ go build -ldflags "-X main.version=$version" -o host host\main.go
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

## Command Line

`host [-h] [-config ./config.json] [-version] [-init mode] [-peerkey ./peer.key] [-psk]`

- h: Show command line help
- config: Run program using the specified configuration file
- version: Show version number and exit
- init: Initialize configuration in client/server mode
- peerkey: Parse or generate a key file based on the specified file path
- psk: Generate a random Pre-Shared Key

```shell
$ host.exe -init client
Generate peer key success at D:\Code\AIComputingNode\host\peer.key
Important notice: Please save this key file. You can use tools to retrieve the ID and private key in the future.
Encode private key: CAISINAckM6QODvCrez5I0Q3RZyo9PeV4jDeB1L71AHnSU/H
Encode public key: CAISIQNbA9ZWCwFM7X/eTUUBvwSRzTurMLkb9jg38wn5IRL4BQ==
Transform Peer ID: 16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y
Create datastore directory at D:\Code\AIComputingNode\host\datastore
Generate configuration success at D:\Code\AIComputingNode\host\client.json
Run "host -config D:\Code\AIComputingNode\host\client.json" command to start the program
$ 
$ host.exe -peerkey D:\Code\AIComputingNode\host\peer.key
Load peer key success
Encode private key: CAISINAckM6QODvCrez5I0Q3RZyo9PeV4jDeB1L71AHnSU/H
Encode public key: CAISIQNbA9ZWCwFM7X/eTUUBvwSRzTurMLkb9jg38wn5IRL4BQ==
Transform Peer ID: 16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y
$ 
$ host.exe -config D:\Code\AIComputingNode\host\client.json
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

## Configuration

```json
{
  // Boot node list, used for node routing and discovery functions, needs to be deployed on
  // a public network server, preferably with a domain name
  "Bootstrap": [
    "/ip4/122.99.183.54/tcp/6001/p2p/16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
    "/ip4/82.157.50.32/tcp/6001/p2p/16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF"
  ],
  // Listening address for node connection communication
  "Addresses": [
    "/ip4/0.0.0.0/tcp/6001",
    "/ip6/::/tcp/6001"
  ],
  // API that provides HTTP services
  "API": {
    // Provided in the form "host:port", such as "localhost:8080", "127.0.0.1:8080"
    // or "0.0.0.0:8080", where "0.0.0.0:8080" can be abbreviated as ":8080".
    "Addr": ":6000"
  },
  // The identity information of the node is generated by the peer-key tool. The following is
  // just an example. Please generate and fill it out by yourself during actual deployment.
  // PeerID is generated by algorithm using public key, and the program will verify whether
  // PeerID and PrivKey match.
  "Identity": {
    // Node ID of the node
    "PeerID": "16Uiu2HAmJnGqxBqtWGkymSsy5WDKJY5A5NctcUduENADDptQFF4Y",
    // Node’s private key
    "PrivKey": "CAISINAckM6QODvCrez5I0Q3RZyo9PeV4jDeB1L71AHnSU/H"
  },
  // Configure node connection parameters
  "Swarm": {
    // Connection manager
    "ConnMgr": {
      // Type of connection manager
      // "basic" - Basic manager with parameters such as "HighWater" "LowWater"
      // "none" - Not using a connection manager
      "Type": "basic",
      // GracePeriod is a time duration that new connections are immune from being closed by
      // the connection manager.
      "GracePeriod": "20s",
      // When a node using the basic connection manager reaches HighWater idle connections,
      // it will close the least useful ones until it reaches LowWater idle connections.
      "HighWater": 400,
      "LowWater": 100
    },
    // Disable automatic NAT port forwarding.
    // When not disabled (default), the local node asks NAT devices (e.g., routers), to open up
    // an external port and forward it to the port node is running on. When this works (i.e., when
    // your router supports NAT port forwarding), it makes the local node accessible from the
    // public internet.
    "DisableNatPortMap": false,
    // Configuration options for the relay client to use relay services.
    "RelayClient": {
      // Enables "automatic relay user" mode for this node.
      // If the node is deployed on a public server, please disable it, otherwise enable it.
      "Enabled": true,
      "StaticRelays": []
    },
    // Configuration options for the relay service that can be provided to other peers on the network
    "RelayService": {
      // Enables providing `/p2p-circuit` v2 relay service to other peers on the network.
      // If the node is deployed on a public server, please enable it, otherwise disable it.
      "Enabled": false
    },
    // Experimental
    // Enable hole punching for NAT traversal when port forwarding is not possible. (default: disabled)
    // When enabled, the local node will coordinate with the counterparty using a relayed connection,
    // to upgrade to a direct connection through a NAT/firewall whenever possible. This feature
    // requires Swarm.RelayClient.Enabled to be set to true.
    "EnableHolePunching": false,
    // If you want to help other peers to figure out if they are behind NATs, you can launch the
    // server-side of AutoNAT too (RelayClient already runs the client).
    // This configures the node to provide a service to peers for determining their reachability status.
    // If the node is deployed on a public server, please enable it, otherwise disable it.
    // This service is highly rate-limited and should not cause any performance issues.
    "EnableAutoNATService": true
  },
  // Publish and subscribe configuration
  "Pubsub": {
    // Enabled by default
    "Enabled": true,
    // Route type, default "gossipsub"
    // "floodsub" - floodsub is a basic router that simply floods messages to all connected peers.
    // This router is extremely inefficient but very reliable. Suitable for small networks with few nodes.
    // "gossipsub" - gossipsub is a more advanced router with mesh formation and gossip propagation.
    // Suitable for large networks with many nodes.
    "Router": "gossipsub",
    // Use flood publishing in "gossipsub" to send messages to all peers with a score greater than publishThreshold.
    "FloodPublish": true
  },
  // Node routing configuration
  "Routing": {
    // Route type, one of the following options
    // "dhtclient" - Client mode, suitable for nodes behind NAT
    // "dhtserver" - Server mode, suitable for nodes deployed on public network servers
    // "auto" - Automatically determine client mode or server mode based on network conditions
    // "none" - Not using DHT routing to discover other nodes
    "Type": "dhtclient",
    // The prefix of the routing protocol. Nodes with different protocols cannot establish connections.
    "ProtocolPrefix": "/dbc"
  },
  // Application configuration
  "App": {
    // Log level, one of "debug"、"info"、"warn"、"error"、"panic" and "fatal"
    "LogLevel": "info",
    // Log file path, used to save log records generated by the program
    "LogFile": "./test.log",
    // Log output mode, freely combined by "stderr", "file" or "+"
    // "stderr" - output to console
    // "file" - Output to the file specified by "LogFile"
    // "stderr+file" - Output to the console and the specified file at the same time
    "LogOutput": "stderr+file",
    // Pre-generated shared key, nodes with different keys cannot establish connections
    "PreSharedKey": "f504f536a912a8cf7d00adacee8ed20270c5040d961d7f3da4fccbcbec0ec48a",
    // The subscribed topic name, used for centralized node query services, such as querying
    // the Node ID list of all nodes in the network.
    // Different topic names may prevent others from querying the node you deployed.
    "TopicName": "DeepBrainChain",
    // The data storage folder persists the successfully connected peer information and other
    // information to facilitate the node to quickly reconnect to the network when it starts.
    // Please create this folder in advance and ensure that this node program has read and
    // write permissions.
    "Datastore": "./datastore",
    // The HTTP RPC interface of the IPFS storage node is used to store image resources generated by
    // the AI model. The node executing the AI model needs to configure this, but the storage node
    // specified in the user request will be used first.
    "IpfsStorageAPI": "http://122.99.183.54:4002",
    // Collect the heartbeat information broadcast by the node, which includes the supported
    // AI projects and models.
    "PeersCollect": {
      // The node information will be saved to the leveldb database only if the node is located on
      // a public network server and collection is enabled.
      "Enabled": false,
      // Heartbeat interval
      "HeartbeatInterval": "180s"
    }
  },
  // List of supported AI projects for Node, which can be managed using interface registration/unregistration
  "AIProjects": [
    {
      // Project name
      "Project": "SuperImage",
      // Model list, a project can contain multiple models
      "Models": [
        {
          // Model name
          "Model": "SuperImageAI",
          // The access interface of the model is used to receive HTTP requests, run the model according to
          // the prompt words and other parameters in the request, and return the generated text or images.
          // This must be configured on the node that executes the AI model.
          "API": "http://127.0.0.1:1088/models",
          // Model type, one of the following options
          // 0 - Text chat dialogue model
          // 1 - Text generation picture model
          // 2 - Image generation image model
          "Type": 1
        }
      ]
    }
  ]
}
```

### Client Configuration Example

```json
{
  "Bootstrap": [
    "/ip4/122.99.183.54/tcp/6001/p2p/16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
    "/ip4/82.157.50.32/tcp/6001/p2p/16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF"
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
    "IpfsStorageAPI": "http://122.99.183.54:4002",
    "PeersCollect": {
      "Enabled": false,
      "HeartbeatInterval": "180s"
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

### Server Configuration Example

```json
{
  "Bootstrap": [
    "/ip4/122.99.183.54/tcp/6001/p2p/16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
    "/ip4/82.157.50.32/tcp/6001/p2p/16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF"
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
    "IpfsStorageAPI": "",
    "PeersCollect": {
      "Enabled": true,
      "HeartbeatInterval": "180s"
    }
  },
  "AIProjects": []
}
```

## Tools

### Ipfs

Upload files to or download files from IPFS nodes.

```shell
$ ./ipfs -node /ip4/192.168.1.159/tcp/4002 -upload ./test.png
2024/04/09 17:01:35 Upload file /ipfs/QmStNRFDoBzuEn4g7wWXV7UEFXtXGBf4SgJjzmjjDvD1Hb
2024/04/09 17:01:35 File uploaded successfully.
$ ./ipfs -node /ip4/192.168.1.159/tcp/4002 -download QmStNRFDoBzuEn4g7wWXV7UEFXtXGBf4SgJjzmjjDvD1Hb -save ./test.png
2024/04/09 17:03:22 File downloaded successfully.
```

When we upload a file to an IPFS node (such as `/ip4/192.168.1.159/tcp/4002` in the above example), a CID identification (such as `QmStNRFDoBzuEn4g7wWXV7UEFXtXGBf4SgJjzmjjDvD1Hb` in the above example) will be returned, and then we can view the file in the browser through `http://192.168.1.159:4040/ipfs/cid`, or even view it on the public Internet through `https://ipfs.io/ipfs/cid/`.

## Document

[AI Model Interface Standard Documentation](./model_api.md)

[AIComputingNode HTTP API Interface Documentation](./api.md)
