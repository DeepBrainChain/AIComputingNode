# InferenceComputingNetwork

Distributed Inference Computing Network for AI

## Table of Contents

- [`Compiling`](#compiling)
- [`Protobuf`](#protobuf)
- [`Command Line`](#command-line)
- [`Node Deployment`](#node-deployment)
- [`Tools`](#tools)
  - [`Ipfs`](#ipfs)
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

## Node Deployment

The following steps describe how to deploy this distributed network communication node.

### Step 1, Download executable program

This project will automatically compile and upload the latest executable program, and supports Linux/Windows/macOS systems.

Check the [latest version](https://github.com/DeepBrainChain/AIComputingNode/releases/latest) here and download the corresponding executable program according to the user's system.

After downloading to the system, make sure that the executable program is used for execution permissions.

### Step 2, Generate JSON Configuration File

If your machine has a public IP address, please run the `host -init server` command, which will generate a configuration file named `server.json` in the same path as the program.

If your machine does not have a public IP address, please run the `host -init xxx` command, which will generate a configuration file named `xxx.json` in the same path as the program.

After generating the JSON configuration file, you need to add some bootstrap nodes to the configuration file, and ensure that the open ports do not conflict and the API meets your needs.

Please refer to [JSON Configuration File](./docs/configuration.md) for detailed configuration items.

### Step 3, Run the program using the JSON configuration file

Assuming that the JSON configuration file generated above is named `config.json`, run it with `host -config ./config.json`.

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

[AI Model Interface Standard Documentation](./docs/model_api.md)

[AIComputingNode HTTP API Interface Documentation](./docs/api.md)

[JSON Configuration File](./docs/configuration.md)
