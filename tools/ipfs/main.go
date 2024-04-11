package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/ipfs/boxo/files"
	path "github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/ipfs/kubo/core/coreiface/options"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	ipfsNode := flag.String("node", "/ip4/192.168.1.159/tcp/4002", "the ipfs node that wanted to storage")
	upload := flag.String("upload", "", "the file path that wanted to upload")
	download := flag.String("download", "", "the file cid that wanted to download")
	saveFile := flag.String("save", filepath.Join(".", "test.txt"), "the file path to which the downloaded file is saved")
	flag.Parse()

	ctx := context.Background()
	maddr, err := multiaddr.NewMultiaddr(*ipfsNode)
	if err != nil {
		log.Fatal("ipfs node ", err)
	}

	if *upload != "" {
		Upload(ctx, maddr, *upload)
	}
	if *download != "" {
		cid, err := cid.Decode(*download)
		if err != nil {
			log.Fatal("Failed to parse file cid ", err)
		}
		Download(ctx, maddr, cid, *saveFile)
	}
}

func Upload(ctx context.Context, addr multiaddr.Multiaddr, filePath string) {
	node, err := rpc.NewApi(addr)
	if err != nil {
		log.Fatal("Failed to create ipfs rpc ", err)
	}

	abspath, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatal("Invalid file path ", err)
	}
	file, err := os.Open(abspath)
	if err != nil {
		log.Fatal("Failed to read file ", err)
	}
	defer file.Close()

	fn := files.NewReaderFile(file)
	defer fn.Close()

	pip, err := node.Unixfs().Add(ctx, fn, options.Unixfs.Pin(true))
	if err != nil {
		log.Fatal("Failed to upload file ", err)
	}
	log.Println("Upload file", pip)
	log.Println("File uploaded successfully.")
}

func Download(ctx context.Context, addr multiaddr.Multiaddr, cid cid.Cid, filePath string) {
	outfile, err := os.Create(filePath)
	if err != nil {
		log.Fatal("Failed to create file ", err)
	}
	defer outfile.Close()

	node, err := rpc.NewApi(addr)
	if err != nil {
		log.Fatal("Failed to create ipfs rpc ", err)
	}
	p := path.FromCid(cid)
	// p, err := path.NewPath("/ipfs/" + cidstring)
	// err = node.Pin().Add(ctx, p)
	// if err != nil {
	// 	log.Fatal("Failed to download file ", err)
	// }
	// list, err := node.Pin().Ls(ctx)
	// if err != nil {
	// 	log.Fatal("Failed to list file ", err)
	// }
	// for ip := range list {
	// 	fmt.Println("item ", ip)
	// }
	f, err := node.Unixfs().Get(ctx, p)
	if err != nil {
		log.Fatal("Get file use cid ", err)
	}
	defer f.Close()

	var file files.File
	switch f := f.(type) {
	case files.File:
		file = f
	default:
		log.Fatal("Failed to get reader of ipfs file")
	}

	_, err = io.Copy(outfile, file)
	if err != nil {
		log.Fatal("Failed to copy file ", err)
	}
	log.Println("File downloaded successfully.")
}
