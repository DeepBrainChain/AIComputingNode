package stream

import (
	"bufio"
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"

	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/model"
	"AIComputingNode/pkg/timer"
	"AIComputingNode/pkg/types"

	"github.com/libp2p/go-libp2p/core/network"
)

type Libp2pStream struct {
	pcn              chan<- []byte
	DefaultTransport http.RoundTripper
}

func NewLibp2pStream(publishChan chan<- []byte) *Libp2pStream {
	return &Libp2pStream{
		pcn: publishChan,
		DefaultTransport: &http.Transport{
			// Proxy: ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			DisableCompression: true,
			// ForceAttemptHTTP2:     true,
			MaxIdleConns: 100,
			// DisableKeepAlives:   true,
			// MaxIdleConnsPerHost: -1,
			DisableKeepAlives:     false,
			MaxIdleConnsPerHost:   20,
			MaxConnsPerHost:       128,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func NewHttpClientTrace() *httptrace.ClientTrace {
	trace := &httptrace.ClientTrace{
		DNSStart: func(di httptrace.DNSStartInfo) {
			log.Logger.Infof("HTTP client trace DNSStart{Host: %v}", di.Host)
		},
		DNSDone: func(di httptrace.DNSDoneInfo) {
			log.Logger.Infof("HTTP client trace DNSDone{Addrs: %v, Error: %v, Coalesced: %v}", di.Addrs, di.Err, di.Coalesced)
		},
		ConnectStart: func(network, addr string) {
			log.Logger.Infof("HTTP client trace ConnectStart{network: %s, addr: %s}", network, addr)
		},
		ConnectDone: func(network, addr string, err error) {
			log.Logger.Infof("HTTP client trace ConnectDone{network: %s, addr: %s, err: %v}", network, addr, err)
		},
		TLSHandshakeStart: func() {
			log.Logger.Info("HTTP client trace TLSHandshakeStart")
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			log.Logger.Infof("HTTP client trace TLSHandshakeDone{tls.ConnectionState: {Version: %v, ServerName: %v}, err %v}",
				cs.Version, cs.ServerName, err)
		},
		GetConn: func(hostPort string) {
			log.Logger.Infof("HTTP client trace GetConn{HostPort: %v}", hostPort)
		},
		GotConn: func(gci httptrace.GotConnInfo) {
			log.Logger.Infof("HTTP client trace GotConn{GotConnInfo: {Reused: %v, WasIdle: %v, IdleTime: %v}}",
				gci.Reused, gci.WasIdle, gci.IdleTime.String())
		},
		PutIdleConn: func(err error) {
			log.Logger.Infof("HTTP client trace PutIdleConn{err: %v}", err)
		},
		GotFirstResponseByte: func() {
			log.Logger.Info("HTTP client trace GotFirstResponseByte")
		},
	}
	return trace
}

// streamHandler is our function to handle any libp2p-net streams that belong
// to our protocol. The streams should contain an HTTP request which we need
// to parse, make on behalf of the original node, and then write the response
// on the stream, before closing it.
func (ls *Libp2pStream) ChatProxyStreamHandler(stream network.Stream) {
	// Remember to close the stream when we are done.
	defer stream.Close()

	log.Logger.Infof("Chat proxy stream with %s started", stream.ID())
	// Create a new buffered reader, as ReadRequest needs one.
	// The buffered reader reads from our stream, on which we
	// have sent the HTTP request (see ServeHTTP())
	buf := bufio.NewReader(stream)
	// Read the HTTP request from the buffer
	req, err := http.ReadRequest(buf)
	if err != nil {
		stream.Reset()
		log.Logger.Errorf("Read chat proxy request from stream failed: %v", err)
		return
	}
	defer req.Body.Close()

	// var msg types.ChatModelRequest
	// if err := json.NewDecoder(req.Body).Decode(&msg); err != nil {
	// 	stream.Reset()
	// 	log.Logger.Errorf("Parse chat proxy request failed: %v", err)
	// 	return
	// }

	// modelUrl := config.GC.GetModelAPI(msg.Project, msg.Model)
	// if modelUrl == "" {
	// 	stream.Reset()
	// 	log.Logger.Errorf("Get model api interface failed: %v", err)
	// 	return
	// }

	// req.URL, err = url.Parse(modelUrl)
	// if err != nil {
	// 	stream.Reset()
	// 	log.Logger.Errorf("Parse model api interface failed: %v", err)
	// 	return
	// }

	// We need to reset these fields in the request
	// URL as they are not maintained.
	// req.URL.Scheme = "http"
	// hp := strings.Split(req.Host, ":")
	// if len(hp) > 1 && hp[1] == "443" {
	// 	req.URL.Scheme = "https"
	// } else {
	// 	req.URL.Scheme = "http"
	// }
	// req.URL.Host = req.Host
	queryValues := req.URL.Query()
	projectName := queryValues.Get("project")
	modelName := queryValues.Get("model")
	cid := queryValues.Get("cid")
	mi, err := model.GetModelInfo(projectName, modelName, cid)
	if err != nil {
		stream.Reset()
		log.Logger.Errorf("Get model api interface failed: %v", err)
		return
	}

	req.URL, err = url.Parse(mi.API)
	if err != nil {
		stream.Reset()
		log.Logger.Errorf("Parse model api interface failed: %v", err)
		return
	}
	queryValues.Del("project")
	queryValues.Del("model")
	queryValues.Del("cid")
	req.URL.RawQuery = queryValues.Encode()
	req.Host = req.URL.Host

	// outreq := new(http.Request)
	// *outreq = *req
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := types.ChatCompletionRequestTimeout
	if mi.Type == 1 || mi.Type == 2 {
		timeout = types.ImageGenerationRequestTimeout
	}
	pctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	outreq := req.WithContext(httptrace.WithClientTrace(pctx, NewHttpClientTrace()))

	stream.SetDeadline(time.Now().Add(timeout))

	model.IncRef(projectName, modelName, mi.CID)
	timer.SendAIProjects(ls.pcn)
	defer func() {
		model.DecRef(projectName, modelName, mi.CID)
		timer.SendAIProjects(ls.pcn)
	}()

	// We now make the request
	log.Logger.Infof("Making request to %s", req.URL)
	resp, err := ls.DefaultTransport.RoundTrip(outreq)
	if err != nil {
		stream.Reset()
		log.Logger.Errorf("RoundTrip chat proxy request failed: %v", err)
		return
	}

	// resp.Write writes whatever response we obtained for our
	// request back to the stream.
	log.Logger.Info("Write roundtrip response into chat proxy stream")
	resp.Write(stream)
	log.Logger.Infof("Chat proxy stream with %s stopped", stream.ID())
}
