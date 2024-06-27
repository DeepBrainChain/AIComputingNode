package ipfs

import (
	"context"
	gopath "path/filepath"
	"testing"
	"time"
)

func TestUploadFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var (
		ipfsAddr  string    = "/ip4/192.168.1.159/tcp/4002"
		httpAPI   string    = "http://192.168.1.159:4002"
		model     string    = "superImage"
		prompt    string    = "bird"
		filePath  string    = "C:\\Users\\Jerry\\Pictures\\tux.png"
		timestamp time.Time = time.Now()
	)
	cid, code, err := UploadFile(ctx, ipfsAddr, filePath)
	if code != 0 {
		t.Fatalf("Upload file failed %v", err)
	}
	t.Logf("Upload file success and cid %s", cid)
	t.Logf("Get file directory %s", gopath.Base(filePath))

	if err := WriteMFSHistory(timestamp.Unix(), "123", "456", cid, httpAPI, model, prompt, nil); err != nil {
		t.Fatalf("Write ipfs mfs history failed %v", err)
	}
	t.Log("Write ipfs mfs history success")

	body, err := ReadMFSHistory(httpAPI, cid)
	if err != nil {
		t.Fatalf("Read ipfs mfs history failed %v", err)
	}
	t.Logf("Read ipfs mfs history:\n%s", string(body))
}

func TestReadFile(t *testing.T) {
	body, err := ReadMFSHistory(
		"http://192.168.1.159:4002",
		"QmRE8s1WZTVAqyxiTXwziPjjXLFUxzt8rSEGoYN19tb513",
	)
	if err != nil {
		t.Fatalf("Read ipfs mfs history failed %v", err)
	}
	t.Logf("Read ipfs mfs history:\n%s", string(body))
}
