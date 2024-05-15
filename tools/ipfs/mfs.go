package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"
)

type history struct {
	Name      string `json:"name"`
	TimeStamp int64  `json:"timestamp"`
}

func main() {
	ipfsNode := flag.String("node", "http://192.168.1.159:4002", "the ipfs node that wanted to storage")
	mfsPath := flag.String("path", "", "the MFS file path that wanted to append")
	readFlag := flag.Bool("read", false, "read the mfs file")
	flag.Parse()

	if *mfsPath == "" {
		log.Fatal("Need MFS path paramater")
	}

	if *readFlag {
		ReadMFSFile(*ipfsNode, *mfsPath)
	} else {
		WriteMFSFile(*ipfsNode, *mfsPath)
	}
}

func ReadMFSFile(ipfsServer, filePath string) {
	resp, err := http.PostForm(
		fmt.Sprintf("%s/api/v0/files/read?arg=%s", ipfsServer, filePath),
		nil,
	)
	if err != nil {
		log.Fatalf("Send ipfs http rpc request: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Read http response: %v", err)
	}
	log.Println("Read mfs file: ", string(body))
}

func WriteMFSFile(ipfsServer, filePath string) {
	curTime := time.Now()
	his := history{
		Name:      curTime.String(),
		TimeStamp: curTime.Unix(),
	}
	jsonData, err := json.MarshalIndent(his, "", "  ")
	if err != nil {
		log.Fatalf("Marshal json failed: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "123.json")
	if err != nil {
		log.Fatalf("Create multipart failed: %v", err)
	}
	part.Write(jsonData)
	if err := writer.Close(); err != nil {
		log.Fatalf("Close multipart writer failed: %v", err)
	}

	resp, err := http.Post(
		fmt.Sprintf(
			"%s/api/v0/files/write?arg=%s&create=true&parents=true",
			ipfsServer,
			filePath,
		),
		writer.FormDataContentType(),
		body,
	)
	if err != nil {
		log.Fatalf("Send ipfs http rpc request: %v", err)
	}
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Write http response: %v", err)
	}
	log.Println("Write mfs file return: ", string(resBody))
}
