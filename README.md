# InferenceComputingNetwork

Distributed Inference Computing Network for AI

## Table of Contents

- [`Compiling`](#compiling)
- [`Protobuf`](#protobuf)
- [`Command Line`](#command-line)
- [`Node Deployment`](#node-deployment)
- [`Tools`](#tools)
- [`Document`](#document)

## Compiling

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

## Command Line

`host [-h] [-config ./config.json] [-version] [-init mode] [-peerkey ./peer.key] [-psk]`

- h: Show command line help
- config: Run program using the specified configuration file
- version: Show version number and exit
- init: Initialize configuration in input/worker mode
- peerkey: Parse or generate a key file based on the specified file path
- psk: Generate a random Pre-Shared Key

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

## Node type description

- Input node: A node with a public IP address, which is used to receive project HTTP API requests and return results (it would be better if domain name resolution is done well), and is responsible for node discovery and routing in a distributed network.
- Worker node: A node without a public IP address, which is deployed on the machine where the model is located, receives model requests sent to it from the Input node or the distributed network, forwards them to the model interface and returns the results.
- Bootstrap node: A concept inherited from the libp2p and IPFS projects, which is equivalent to the Input node in this project.

After the model is deployed, you need to register the model's calling interface and the project to which it belongs to the Worker node. The Worker node will propagate this information to the entire distributed communication network, and then you can call the deployed model through the Input node with a domain name or public IP address.

```
                  +----------------+                +----------------+             +----------------+
 HTTP Request     |    DBC AI      |                |    DBC AI      |             |                |
+----------------->                | libp2p stream  |                |  HTTP       |                |
                  |  Input Node    <---------------->  Worker Node   <------------->    AI Model    |
<-----------------+                |                |                | Req & Resp  |                |
  HTTP Response   |  CPU Machine   |                |  GPU Machine   |             |   GPU Machine  |
                  +----------------+                +----------------+             +----------------+
```

1. If you have a GPU server, you need to deploy a Worker node.
2. If you have a CPU machine and a domain name/public IP address, you can deploy an Input node.
3. If the GPU machine itself has a domain name/public IP address, you can directly deploy an Input node and register the model interface to the Input node without a Worker node.

## Node Deployment

The following steps describe how to deploy this distributed network communication node.

### Step 1, Download executable program

This project will automatically compile and upload the latest executable program, and supports Linux/Windows/macOS systems.

Check the [latest version](https://github.com/DeepBrainChain/AIComputingNode/releases/latest) here and download the corresponding executable program according to the user's system.

After downloading to the system, make sure that the executable program is used for execution permissions.

### Step 2, Generate JSON Configuration File

If you need to deploy an Input node, please run the `host -init input` command, which will generate a configuration file named `input.json` in the same path as the program.

If you need to deploy a Worker node, please run the `host -init worker` command, which will generate a configuration file named `worker.json` in the same path as the program.

After generating the JSON configuration file, you need to add some bootstrap nodes to the configuration file, and ensure that the open ports do not conflict and the API meets your needs.

Please refer to [JSON Configuration File](./docs/configuration.md) for detailed configuration items.

### Step 3, Run the program using the JSON configuration file

Assuming that the JSON configuration file generated above is named `worker.json`, run it with `host -config ./worker.json`.

## Tools

[Ipfs](./tools/ipfs/README.md)

## Document

[AI Model Interface Standard Documentation](./docs/model_api.md)

[AIComputingNode HTTP API Interface Documentation](./docs/api.md)

[JSON Configuration File](./docs/configuration.md)
