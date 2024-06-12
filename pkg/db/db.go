package db

import (
	"encoding/binary"
	"encoding/json"
	"path/filepath"
	"time"

	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/types"

	"github.com/syndtr/goleveldb/leveldb"
)

var peersDB *leveldb.DB
var modelsDB *leveldb.DB

type InitOptions struct {
	Folder       string
	PeersDBName  string
	ModelsDBName string
}

type peerInfo struct {
	Address             string `json:"Address"`             // connection address
	LastConnectTime     int64  `json:"LastConnectTime"`     // Time of last successful connection
	ConsecutiveFailures int32  `json:"ConsecutiveFailures"` // Number of consecutive failures
}

func InitDb(opts InitOptions) error {
	if opts.PeersDBName == "" {
		opts.PeersDBName = "peers.db"
	}
	if opts.ModelsDBName == "" {
		opts.ModelsDBName = "models.db"
	}
	var err error
	peersDB, err = leveldb.OpenFile(filepath.Join(opts.Folder, opts.PeersDBName), nil)
	if err != nil {
		return err
	}
	modelsDB, err = leveldb.OpenFile(filepath.Join(opts.Folder, opts.ModelsDBName), nil)
	return err
}

func LoadPeers() map[string]string {
	peers := make(map[string]string)
	iter := peersDB.NewIterator(nil, nil)
	for iter.Next() {
		var pi peerInfo
		if err := json.Unmarshal(iter.Value(), &pi); err != nil {
			log.Logger.Warn("Parse failed when load peer of ", iter.Key(), err)
			continue
		}
		peers[string(iter.Key())] = pi.Address
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		log.Logger.Warnf("Iterator failed when load peers %v", err)
	}
	return peers
}

func PeerConnected(id string, addr string) {
	pi := peerInfo{
		Address:             addr,
		LastConnectTime:     time.Now().Unix(),
		ConsecutiveFailures: 0,
	}
	updatePeer(id, pi)
}

func PeerConnectFailed(id string) {
	value, err := peersDB.Get([]byte(id), nil)
	if err != nil {
		log.Logger.Warnf("Get connectiong db item failed %v", err)
	}
	var pi peerInfo
	err = json.Unmarshal(value, &pi)
	if err != nil {
		log.Logger.Warnf("Parse connectiong db item failed %v", err)
	}

	pi.ConsecutiveFailures = pi.ConsecutiveFailures + 1
	if pi.ConsecutiveFailures >= 10 {
		lastTime := time.Unix(pi.LastConnectTime, 0)
		currentTime := time.Now()
		duration := currentTime.Sub(lastTime)
		if duration.Abs().Hours()/24 > 30.0 {
			log.Logger.Infof("Connection with %s is too old and failed too many times, %s",
				id, "will be deleted soon")
			if err := peersDB.Delete([]byte(id), nil); err != nil {
				log.Logger.Warnf("Delete item %s failed %v", id, err)
			}
		}
	} else {
		updatePeer(id, pi)
	}
}

func updatePeer(id string, pi peerInfo) error {
	value, err := json.Marshal(pi)
	if err != nil {
		log.Logger.Warnf("Marshal failed when update connectiong db %v", err)
	}
	err = peersDB.Put([]byte(id), value, nil)
	if err != nil {
		log.Logger.Warnf("Update connectiong db failed %v", err)
	}
	return nil
}

func WriteModelHistory(mh *types.ModelHistory) error {
	value, err := json.Marshal(mh)
	if err != nil {
		log.Logger.Warnf("Marshal failed when write model history %v", err)
		return err
	}
	keyBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(keyBytes, uint64(mh.TimeStamp))

	if err := modelsDB.Put(keyBytes, value, nil); err != nil {
		log.Logger.Warnf("Put model history failed %v", err)
		return err
	}
	log.Logger.Infof("Put model history success")
	return nil
}
