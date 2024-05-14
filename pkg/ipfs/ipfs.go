package ipfs

import (
	"context"
	"os"
	"path/filepath"

	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/types"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/ipfs/kubo/core/coreiface/options"
	"github.com/multiformats/go-multiaddr"
)

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
