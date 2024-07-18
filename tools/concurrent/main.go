package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"sync"

	"AIComputingNode/pkg/types"

	golog "github.com/ipfs/go-log/v2"
)

var log = golog.Logger("test")

func main() {
	// golog.SetAllLoggers(golog.LevelDebug)
	golog.SetAllLoggers(golog.LevelInfo)

	url := flag.String("url", "", "ai model http url")
	model := flag.String("model", "", "ai model name")
	rt := flag.Int("routine", 3, "goroutine number")
	msg := flag.String("content", "Hello", "the content words that input to model")
	flag.Parse()

	if *url == "" {
		log.Fatal("Please run with -url param")
	}
	if *model == "" {
		log.Fatal("Please run with -url param")
	}

	req := types.ChatModelRequest{
		Model:    *model,
		Messages: []types.ChatCompletionMessage{},
		Stream:   true,
	}
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "system",
		Content: "You are a helpful assistant.",
	})
	req.Messages = append(req.Messages, types.ChatCompletionMessage{
		Role:    "user",
		Content: *msg,
	})
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("Marshal model request: %v", err)
	}

	concurrency := *rt
	var wg sync.WaitGroup

	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		log.Fatalf("Can not convert http.DefaultTransport")
	}
	log.Infof("http.DefaultTransport {DisableKeepAlives %v, MaxIdleConns %v, MaxIdleConnsPerHost %v, MaxConnsPerHost %v}",
		transport.DisableKeepAlives, transport.MaxIdleConns, transport.MaxIdleConnsPerHost, transport.MaxConnsPerHost)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i int, body []byte) {
			defer wg.Done()
			log.Infof("Goroutine %v started", i)
			req, err := http.NewRequest("POST", *url, bytes.NewBuffer(body))
			if err != nil {
				log.Errorf("Goroutine %v new request failed: %v", i, err)
				return
			}
			defer req.Body.Close()

			log.Infof("Making request to %s\n", req.URL)
			resp, err := http.DefaultTransport.RoundTrip(req)
			if err != nil {
				log.Errorf("Goroutine %v roundtrip request failed: %v", i, err)
				return
			}
			defer resp.Body.Close()

			log.Infof("Goroutine %v begin to read response stream", i)
			sc := bufio.NewScanner(resp.Body)
			for {
				if !sc.Scan() {
					if err := sc.Err(); errors.Is(err, io.EOF) {
						log.Infof("Goroutine %v read response EOF", i)
					} else if err != nil {
						log.Errorf("Goroutine %v read response failed: %v", i, err)
					}
					break
				}
				line := sc.Bytes()
				log.Infof("Goroutine %v read response %s", i, string(line))
			}
			log.Infof("Goroutine %v stopped", i)
		}(i, jsonData)
	}

	wg.Wait()
}
