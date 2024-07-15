package p2p

import (
	"bufio"
	"net/http"
	"net/url"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/log"

	"github.com/libp2p/go-libp2p/core/network"
)

const ChatProxyProtocol = "/chat-proxy/0.0.1"

// streamHandler is our function to handle any libp2p-net streams that belong
// to our protocol. The streams should contain an HTTP request which we need
// to parse, make on behalf of the original node, and then write the response
// on the stream, before closing it.
func ChatProxyStreamHandler(stream network.Stream) {
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
	modelUrl := config.GC.GetModelAPI(projectName, modelName)
	if modelUrl == "" {
		stream.Reset()
		log.Logger.Errorf("Get model api interface failed: %v", err)
		return
	}

	req.URL, err = url.Parse(modelUrl)
	if err != nil {
		stream.Reset()
		log.Logger.Errorf("Parse model api interface failed: %v", err)
		return
	}
	queryValues.Del("project")
	queryValues.Del("model")
	req.URL.RawQuery = queryValues.Encode()

	outreq := new(http.Request)
	*outreq = *req

	// We now make the request
	log.Logger.Infof("Making request to %s\n", req.URL)
	resp, err := http.DefaultTransport.RoundTrip(outreq)
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
