package stream

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"AIComputingNode/pkg/test"
	"AIComputingNode/pkg/types"

	golog "github.com/ipfs/go-log/v2"
)

var tlog = golog.Logger("AIComputingNode")

// streamHandler is our function to handle any libp2p-net streams that belong
// to our protocol. The streams should contain an HTTP request which we need
// to parse, make on behalf of the original node, and then write the response
// on the stream, before closing it.
func ChatProxyStreamTestHandler(stream network.Stream) {
	stream.SetDeadline(time.Now().Add(3 * time.Minute))
	// Remember to close the stream when we are done.
	defer stream.Close()

	tlog.Infof("Chat proxy stream with %s started", stream.ID())
	// Create a new buffered reader, as ReadRequest needs one.
	// The buffered reader reads from our stream, on which we
	// have sent the HTTP request (see ServeHTTP())
	buf := bufio.NewReader(stream)
	// Read the HTTP request from the buffer
	req, err := http.ReadRequest(buf)
	if err != nil {
		stream.Reset()
		tlog.Errorf("Read chat proxy request from stream failed: %v", err)
		return
	}
	defer req.Body.Close()

	req.URL.Scheme = "http"
	req.URL.Host = req.Host

	tlog.Infof("stream test handle read http query %v", req.URL.RawQuery)

	outreq := new(http.Request)
	*outreq = *req

	// We now make the request
	tlog.Infof("Making request to %s", req.URL)
	resp, err := http.DefaultTransport.RoundTrip(outreq)
	if err != nil {
		stream.Reset()
		tlog.Errorf("RoundTrip chat proxy request failed: %v", err)
		return
	}

	// resp.Write writes whatever response we obtained for our
	// request back to the stream.
	tlog.Info("Write roundtrip response into chat proxy stream")
	resp.Write(stream)
	tlog.Infof("Chat proxy stream with %s stopped", stream.ID())
}

func TestHttpUrl(t *testing.T) {
	var testUrl = "http://127.0.0.1:8080/api/v0/test?p1=hello&p2=100"

	u1, err := url.Parse(testUrl)
	if err != nil {
		t.Fatalf("Parse url failed: %v", err)
	}
	t.Logf("Parse url success %s", u1.String())
	fmt.Println(*u1)
	fmt.Println(u1)
	t.Logf("{Scheme: %s, Opaque: %s, Host: %s, Path: %s, RawPath: %s, OmitHost: %v, ForceQuery: %v, RawQuery: %v, Fragment: %v, RawFragment: %v}",
		u1.Scheme, u1.Opaque, u1.Host, u1.Path, u1.RawPath, u1.OmitHost, u1.ForceQuery, u1.RawQuery, u1.Fragment, u1.RawFragment)

	u2, err := url.ParseRequestURI(testUrl)
	if err != nil {
		t.Fatalf("Parse url failed: %v", err)
	}
	t.Logf("Parse url success %s", u2.String())
	t.Logf("Query list %v", u2.Query())
	queryValues := u2.Query()
	queryValues.Add("project", "a b")
	queryValues.Add("model", "p2p")
	u2.RawQuery = queryValues.Encode()
	t.Logf("Modify url querys success %s", u2.String())
	newQuery := u2.Query()
	t.Logf("Parse query project %s model %s", newQuery.Get("project"), newQuery.Get("model"))

	_, err = url.Parse("")
	if err != nil {
		t.Fatalf("Parse url failed: %v", err)
	}
}

func TestHttpReuse(t *testing.T) {
	golog.SetAllLoggers(golog.LevelInfo)

	mux := http.NewServeMux()
	mux.HandleFunc("/v0/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "This is a test!")
	})
	mux.HandleFunc("/v0/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
	httpServer := &http.Server{
		Addr:         "localhost:9071",
		Handler:      mux,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		tlog.Info("HTTP server is running on http://localhost:9071")
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			tlog.Fatalf("Start HTTP Server: %v", err)
		}
		tlog.Info("HTTP server is stopped")
	}()

	trace := &httptrace.ClientTrace{
		DNSStart: func(di httptrace.DNSStartInfo) {
			tlog.Infof("HTTP client trace DNSStart{Host: %v}", di.Host)
		},
		DNSDone: func(di httptrace.DNSDoneInfo) {
			tlog.Infof("HTTP client trace DNSDone{Addrs: %v, Error: %v, Coalesced: %v}", di.Addrs, di.Err, di.Coalesced)
		},
		ConnectStart: func(network, addr string) {
			tlog.Infof("HTTP client trace ConnectStart{network: %s, addr: %s}", network, addr)
		},
		ConnectDone: func(network, addr string, err error) {
			tlog.Infof("HTTP client trace ConnectDone{network: %s, addr: %s, err: %v}", network, addr, err)
		},
		TLSHandshakeStart: func() {
			tlog.Info("HTTP client trace TLSHandshakeStart")
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			tlog.Infof("HTTP client trace TLSHandshakeDone{tls.ConnectionState: {Version: %v, ServerName: %v}, err %v}",
				cs.Version, cs.ServerName, err)
		},
		GetConn: func(hostPort string) {
			tlog.Infof("HTTP client trace GetConn{HostPort: %v}", hostPort)
		},
		GotConn: func(gci httptrace.GotConnInfo) {
			tlog.Infof("HTTP client trace GotConn{GotConnInfo: {Reused: %v, WasIdle: %v, IdleTime: %v}}",
				gci.Reused, gci.WasIdle, gci.IdleTime.String())
		},
		PutIdleConn: func(err error) {
			tlog.Infof("HTTP client trace PutIdleConn{err: %v}", err)
		},
		GotFirstResponseByte: func() {
			tlog.Info("HTTP client trace GotFirstResponseByte")
		},
	}
	transport := &http.Transport{
		// Proxy: ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		DisableCompression:    true,
		MaxIdleConns:          100,
		DisableKeepAlives:     false,
		MaxIdleConnsPerHost:   20,
		MaxConnsPerHost:       128,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	time.Sleep(3 * time.Second)

	for i := 0; i < 3; i++ {
		req1, err := http.NewRequestWithContext(
			httptrace.WithClientTrace(context.Background(), trace),
			"GET",
			"http://localhost:9071/v0/test",
			nil,
		)
		if err != nil {
			tlog.Errorf("New HTTP Request failed %v", err)
		}
		resp1, err := transport.RoundTrip(req1)
		if err != nil {
			tlog.Errorf("Round trip test request %v", err)
		}
		// defer resp1.Body.Close()
		body1, err := io.ReadAll(resp1.Body)
		if err != nil {
			tlog.Errorf("Read test response %v", err)
		}
		tlog.Infof("Read test response %s", string(body1))
		resp1.Body.Close()

		time.Sleep(1 * time.Second)

		req2, err := http.NewRequestWithContext(
			httptrace.WithClientTrace(context.Background(), trace),
			"GET",
			"http://localhost:9071/v0/hello",
			nil,
		)
		if err != nil {
			tlog.Errorf("New HTTP Request failed %v", err)
		}
		resp2, err := transport.RoundTrip(req2)
		if err != nil {
			tlog.Errorf("Round trip hello request %v", err)
		}
		defer resp2.Body.Close()
		body2, err := io.ReadAll(resp2.Body)
		if err != nil {
			tlog.Errorf("Read hello response %v", err)
		}
		tlog.Infof("Read hello response %s", string(body2))
	}

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownRelease()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		tlog.Fatalf("Shutdown HTTP Server: %v", err)
	} else {
		tlog.Info("HTTP server is shutdown gracefully")
	}
}

// go test -v -timeout 300s -count=1 -run TestStreamChatModel AIComputingNode/pkg/libp2p/stream > 1.log 2>&1
func TestStreamChatModel(t *testing.T) {
	golog.SetAllLoggers(golog.LevelInfo)
	config, err := test.LoadConfig("D:/Code/AIComputingNode/test.json")
	if err != nil {
		t.Fatalf("Error loading test config file: %v", err)
	}

	req := types.ChatModelRequest{
		Model:    config.Models.Qwen2.Name,
		Messages: []types.ChatCompletionMessage{},
		Stream:   true,
	}
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "system",
		Content: []byte(`"You are a helpful assistant."`),
	})
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "user",
		Content: []byte(`"Hello, What's the weather like today? Where is a good place to travel?"`),
	})
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal model request: %v", err)
	}

	request, err := http.NewRequest("POST", config.Models.Qwen2.API, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("New request failed %v", err)
	}

	privKey1, pubKey1, err := crypto.GenerateKeyPair(crypto.Secp256k1, -1)
	if err != nil {
		t.Fatalf("generate key pair %v", err)
	}
	peer1, err := peer.IDFromPublicKey(pubKey1)
	if err != nil {
		t.Fatalf("parse peer id %v", err)
	}
	privKey2, pubKey2, err := crypto.GenerateKeyPair(crypto.Secp256k1, -1)
	if err != nil {
		t.Fatalf("generate key pair %v", err)
	}
	peer2, err := peer.IDFromPublicKey(pubKey2)
	if err != nil {
		t.Fatalf("parse peer id %v", err)
	}

	t.Logf("peer id 1 %v", peer1)
	t.Logf("peer id 2 %v", peer2)

	host1, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000", // regular tcp connections
		),
		libp2p.Identity(privKey1),
	)
	if err != nil {
		t.Fatalf("create libp2p host %v", err)
	}
	host2, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9001", // regular tcp connections
		),
		libp2p.Identity(privKey2),
	)
	if err != nil {
		t.Fatalf("create libp2p host %v", err)
	}

	pi1 := peer.AddrInfo{
		ID:    host1.ID(),
		Addrs: host1.Addrs(),
	}

	host2.SetStreamHandler("/chat-test/0.0.1", ChatProxyStreamTestHandler)
	if err := host2.Connect(context.Background(), pi1); err != nil {
		t.Logf("host2 connect host1 failed %v", err)
	} else {
		t.Log("host2 connect host1 success")
	}

	stream, err := host1.NewStream(context.Background(), peer2, "/chat-test/0.0.1")
	if err != nil {
		t.Fatalf("New libp2p stream failed %v", err)
	}
	defer stream.Close()
	tlog.Info("Create libp2p stream success")

	err = request.Write(stream)
	if err != nil {
		t.Fatalf("Write libp2p stream failed %v", err)
	}

	buf := bufio.NewReader(stream)
	resp, err := http.ReadResponse(buf, request)
	if err != nil {
		t.Fatalf("Read libp2p stream response failed %v", err)
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	for {
		if !sc.Scan() {
			if err := sc.Err(); errors.Is(err, io.EOF) {
				tlog.Info("Scanner read response EOF")
			} else if err != nil {
				tlog.Errorf("Scanner read response failed: %v", err)
			}
			break
		}
		line := sc.Bytes()
		tlog.Infof("Scanner read response %s", string(line))
	}
	tlog.Info("Scanner stopped")
}
