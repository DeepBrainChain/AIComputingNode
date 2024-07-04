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

```
