package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

type ModelInfo struct {
	API  string `json:"API"`
	Type int    `json:"Type"`
	Idle int    `json:"Idle"`
}

type PeerCollectInfo struct {
	AIProjects map[string]map[string]ModelInfo `json:"AIProjects"`
	NodeType   uint32                          `json:"NodeType"`
	Timestamp  int64                           `json:"timestamp"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("exec db_dir")
		os.Exit(1)
	}

	db, err := leveldb.OpenFile(os.Args[1], nil)
	if err != nil {
		panic(err)
	}

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		var info PeerCollectInfo
		if err := json.Unmarshal(iter.Value(), &info); err != nil {
			fmt.Printf("Parse failed when load peer collect info of %s : %v", string(iter.Key()), err)
			continue
		}
		fmt.Printf("%s - %v\n", string(iter.Key()), info)
	}

	iter.Release()
	if err := iter.Error(); err != nil {
		fmt.Printf("Iterator failed when load peer collect info %v", err)
	}
}
