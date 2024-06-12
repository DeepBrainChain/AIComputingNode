package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"AIComputingNode/pkg/types"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/ipfs/kubo/core/coreiface/options"
	"github.com/multiformats/go-multiaddr"
)

type history struct {
	TimeStamp int64  `json:"timestamp"`
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	ReqNodeId string `json:"req_node_id"`
	ResNodeId string `json:"res_node_id"`
	// Image Choices
	Choices []types.ImageResponseChoice `json:"choices"`
}

func UploadFile(ctx context.Context, addr string, filePath string) (string, int, error) {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return "", int(types.ErrCodeUpload), fmt.Errorf("invalid address of ipfs node")
	}
	node, err := rpc.NewApi(maddr)
	if err != nil {
		return "", int(types.ErrCodeUpload), fmt.Errorf("ipfs API construct error")
	}

	abspath, err := filepath.Abs(filePath)
	if err != nil {
		return "", int(types.ErrCodeUpload), fmt.Errorf("invalid file path")
	}
	file, err := os.Open(abspath)
	if err != nil {
		return "", int(types.ErrCodeUpload), fmt.Errorf("open file failed")
	}
	defer file.Close()

	fn := files.NewReaderFile(file)
	defer fn.Close()

	pip, err := node.Unixfs().Add(ctx, fn, options.Unixfs.Pin(true))
	if err != nil {
		return "", int(types.ErrCodeUpload), fmt.Errorf("upload failed %v", err.Error())
	}
	return pip.RootCid().String(), 0, nil
}

func WriteMFSHistory(timestamp int64, reqId, resId, fileName, ipfsServer, model, prompt string, images []types.ImageResponseChoice) error {
	his := history{
		TimeStamp: timestamp,
		Model:     model,
		Prompt:    prompt,
		ReqNodeId: reqId,
		ResNodeId: resId,
		Choices:   images,
	}
	jsonData, err := json.MarshalIndent(his, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json failed")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fmt.Sprintf("%s.json", fileName))
	if err != nil {
		return fmt.Errorf("create multipart failed")
	}
	part.Write(jsonData)
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart failed")
	}

	resp, err := http.Post(
		fmt.Sprintf(
			"%s/api/v0/files/write?arg=/models/%s.json&create=true&parents=true",
			ipfsServer,
			fileName,
		),
		writer.FormDataContentType(),
		body,
	)
	if err != nil {
		return fmt.Errorf("post ipfs mfs request failed")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("http response status code %d", resp.StatusCode)
	}
	// resBody, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return fmt.Errorf("read ipfs mfs response failed")
	// }
	// log.Logger.Infof("Write ipfs mfs file to %s success. %s", ipfsServer, string(resBody))
	return nil
}

func ReadMFSHistory(ipfsServer, fileName string) ([]byte, error) {
	resp, err := http.PostForm(
		fmt.Sprintf("%s/api/v0/files/read?arg=/models/%s.json", ipfsServer, fileName),
		nil,
	)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}
