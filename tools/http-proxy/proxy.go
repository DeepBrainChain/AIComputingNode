package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	golog "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
)

// Protocol defines the libp2p protocol that we will use for the libp2p proxy
// service that we are going to provide. This will tag the streams used for
// this service. Streams are multiplexed and their protocol tag helps
// libp2p handle them to the right handler functions.
const Protocol = "/proxy-example/0.0.1"

type HttpService struct {
	ctx  context.Context
	host host.Host
	dest peer.ID
	port int
}

func (hs *HttpService) idHandler(w http.ResponseWriter, r *http.Request) {
	if hs.dest == "" {
		http.Error(w, "No destination peer id", http.StatusInternalServerError)
		return
	}
	fmt.Printf("proxying request for %s to peer %s\n", r.URL, hs.dest)
	// We need to send the request to the remote libp2p peer, so
	// we open a stream to it
	stream, err := hs.host.NewStream(hs.ctx, hs.dest, Protocol)
	// If an error happens, we write an error for response.
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	// r.Write() writes the HTTP request to the stream.
	err = r.Write(stream)
	if err != nil {
		stream.Reset()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Now we read the response that was sent from the dest
	// peer
	buf := bufio.NewReader(stream)
	resp, err := http.ReadResponse(buf, r)
	if err != nil {
		stream.Reset()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Copy any headers
	for k, v := range resp.Header {
		for _, s := range v {
			w.Header().Add(k, s)
		}
	}

	// Write response status and headers
	w.WriteHeader(resp.StatusCode)

	// Finally copy the body
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func (hs *HttpService) chatCompletionHandler(w http.ResponseWriter, r *http.Request) {}

func (hs *HttpService) startService() {
	http.HandleFunc("/api/v0/id", hs.idHandler)
	if hs.dest != "" {
		http.HandleFunc("/api/v0/peer", hs.idHandler)
		http.HandleFunc("/v1/chat/completions", hs.idHandler)
	}
	http.HandleFunc("/api/v0/chat/completion", hs.chatCompletionHandler)
	log.Println("HTTP server is running on http://0.0.0.0:", hs.port)
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", hs.port), nil); err != nil {
		log.Fatalf("Start HTTP Server: %v", err)
	}
	log.Println("HTTP server is stopped")
}

func LoadPeerKey(filePath string) (crypto.PrivKey, crypto.PubKey, error) {
	privBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	priv, err := crypto.UnmarshalPrivateKey(privBytes)
	return priv, priv.GetPublic(), err
}

func SavePeerKey(filePath string, priv crypto.PrivKey) error {
	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, privBytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

func GeneratePeerKey(filePath string) (crypto.PrivKey, crypto.PubKey, error) {
	// priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	priv, pub, err := crypto.GenerateKeyPair(crypto.Secp256k1, -1)
	if err != nil {
		return nil, nil, err
	} else {
		err := SavePeerKey(filePath, priv)
		if err != nil {
			return nil, nil, err
		} else {
			return priv, pub, err
		}
	}
}

// streamHandler is our function to handle any libp2p-net streams that belong
// to our protocol. The streams should contain an HTTP request which we need
// to parse, make on behalf of the original node, and then write the response
// on the stream, before closing it.
func streamHandler(stream network.Stream) {
	// Remember to close the stream when we are done.
	defer stream.Close()

	// Create a new buffered reader, as ReadRequest needs one.
	// The buffered reader reads from our stream, on which we
	// have sent the HTTP request (see ServeHTTP())
	buf := bufio.NewReader(stream)
	// Read the HTTP request from the buffer
	req, err := http.ReadRequest(buf)
	if err != nil {
		stream.Reset()
		log.Println(err)
		return
	}
	defer req.Body.Close()

	// We need to reset these fields in the request
	// URL as they are not maintained.
	req.URL.Scheme = "http"
	hp := strings.Split(req.Host, ":")
	if len(hp) > 1 && hp[1] == "443" {
		req.URL.Scheme = "https"
	} else {
		req.URL.Scheme = "http"
	}
	req.URL.Host = req.Host

	outreq := new(http.Request)
	*outreq = *req

	// We now make the request
	fmt.Printf("Making request to %s\n", req.URL)
	resp, err := http.DefaultTransport.RoundTrip(outreq)
	if err != nil {
		stream.Reset()
		log.Println(err)
		return
	}

	// resp.Write writes whatever response we obtained for our
	// request back to the stream.
	resp.Write(stream)
}

// addAddrToPeerstore parses a peer multiaddress and adds
// it to the given host's peerstore, so it knows how to
// contact it. It returns the peer ID of the remote peer.
func addAddrToPeerstore(h host.Host, addr string) peer.ID {
	// The following code extracts target's the peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Fatalln(err)
	}
	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Fatalln(err)
	}

	peerid, err := peer.Decode(pid)
	if err != nil {
		log.Fatalln(err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peerid))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr, so we add
	// it to the peerstore so LibP2P knows how to contact it
	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid
}

func main() {
	// golog.SetAllLoggers(golog.LevelDebug)
	golog.SetAllLoggers(golog.LevelInfo)

	listenF := flag.Int("l", 6000, "listening port waiting for incoming connections")
	destPeer := flag.String("d", "", "destination peer address")
	peerKeyPath := flag.String("peerkey", "", "the file path of peer key")
	flag.Parse()

	if *peerKeyPath == "" {
		log.Fatal("Please provide a filepath to save peer key")
	}
	privKey, _, err := LoadPeerKey(*peerKeyPath)
	if err != nil {
		// log.Fatalf("Load peer key: %v", err)
		privKey, _, err = GeneratePeerKey(*peerKeyPath)
		if err != nil {
			log.Fatalf("Generate peer key: %v", err)
		}
	} else {
		log.Println("Load peer key success")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *listenF)),
		libp2p.Identity(privKey),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		log.Fatalf("Create libp2p host: %v", err)
	}

	log.Println("Listen addresses:", host.Addrs())
	log.Println("Node id:", host.ID())

	host.SetStreamHandler(Protocol, streamHandler)

	hs := HttpService{
		ctx:  ctx,
		host: host,
		port: *listenF + 1,
	}

	if *destPeer != "" {
		hs.dest = addAddrToPeerstore(host, *destPeer)
	}
	hs.startService()
}
