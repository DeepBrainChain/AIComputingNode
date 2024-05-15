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

	"AIComputingNode/pkg/log"
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
	ImageName string `json:"image_name"`
}

func UploadImage(ctx context.Context, addr string, filePath string) (string, int, error) {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		log.Logger.Errorf("Failed to upload image: %v", err)
		return "", int(types.ErrCodeUpload), err
	}
	node, err := rpc.NewApi(maddr)
	if err != nil {
		log.Logger.Errorf("Failed to create ipfs endpoint: %v", err)
		return "", int(types.ErrCodeUpload), err
	}

	abspath, err := filepath.Abs(filePath)
	if err != nil {
		log.Logger.Errorf("Invalid file path: %v", err)
		return "", int(types.ErrCodeUpload), err
	}
	file, err := os.Open(abspath)
	if err != nil {
		log.Logger.Errorf("Failed to open file: %v", err)
		return "", int(types.ErrCodeUpload), err
	}
	defer file.Close()

	fn := files.NewReaderFile(file)
	defer fn.Close()

	pip, err := node.Unixfs().Add(ctx, fn, options.Unixfs.Pin(true))
	if err != nil {
		log.Logger.Error("Failed to upload file: %v", err)
		return "", int(types.ErrCodeUpload), err
	}
	return pip.RootCid().String(), 0, nil
}

func WriteMFSHistory(timestamp int64, ipfsServer, model, prompt, cid, image string) error {
	his := history{
		TimeStamp: timestamp,
		Model:     model,
		Prompt:    prompt,
		ImageName: image,
	}
	jsonData, err := json.MarshalIndent(his, "", "  ")
	if err != nil {
		log.Logger.Errorf("Marshal json failed when write ipfs mfs file %v", err)
		return err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fmt.Sprintf("%s.json", cid))
	if err != nil {
		log.Logger.Errorf("Create multipart failed when write ipfs mfs file %v", err)
		return err
	}
	part.Write(jsonData)
	if err := writer.Close(); err != nil {
		log.Logger.Errorf("Close multipart failed when write ipfs mfs file %v", err)
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf(
			"%s/api/v0/files/write?arg=/models/%s.json&create=true&parents=true",
			ipfsServer,
			cid,
		),
		writer.FormDataContentType(),
		body,
	)
	if err != nil {
		log.Logger.Errorf("Send ipfs mfs write request failed %v", err)
		return err
	}
	if resp.StatusCode != 200 {
		log.Logger.Errorf("Ipfs mfs write request result %s", &resp.Status)
		return fmt.Errorf("http response status code %d", resp.StatusCode)
	}
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Errorf("Read ipfs mfs write response failed %v", err)
		return err
	}
	log.Logger.Infof("Write ipfs mfs file success %s", string(resBody))
	return nil
}
