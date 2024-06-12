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

var connsDB *leveldb.DB
var modelsDB *leveldb.DB

type InitOptions struct {
	Folder       string
	ConnsDBName  string
	ModelsDBName string
}

type connInfo struct {
	Address             string `json:"Address"`             // connection address
	LastConnectTime     int64  `json:"LastConnectTime"`     // Time of last successful connection
	ConsecutiveFailures int32  `json:"ConsecutiveFailures"` // Number of consecutive failures
}

func InitDb(opts InitOptions) error {
	if opts.ConnsDBName == "" {
		opts.ConnsDBName = "conns.db"
	}
	if opts.ModelsDBName == "" {
		opts.ModelsDBName = "models.db"
	}
	var err error
	connsDB, err = leveldb.OpenFile(filepath.Join(opts.Folder, opts.ConnsDBName), nil)
	if err != nil {
		return err
	}
	modelsDB, err = leveldb.OpenFile(filepath.Join(opts.Folder, opts.ModelsDBName), nil)
	return err
}

func LoadPeerConnHistory() map[string]string {
	conns := make(map[string]string)
	iter := connsDB.NewIterator(nil, nil)
	for iter.Next() {
		var pi connInfo
		if err := json.Unmarshal(iter.Value(), &pi); err != nil {
			log.Logger.Warn("Parse failed when load peer conn history of ", iter.Key(), err)
			continue
		}
		conns[string(iter.Key())] = pi.Address
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		log.Logger.Warnf("Iterator failed when load peer conns history %v", err)
	}
	return conns
}

func PeerConnected(id string, addr string) {
	pi := connInfo{
		Address:             addr,
		LastConnectTime:     time.Now().Unix(),
		ConsecutiveFailures: 0,
	}
	updatePeerConnHistory(id, pi)
}

func PeerConnectFailed(id string) {
	value, err := connsDB.Get([]byte(id), nil)
	if err != nil {
		log.Logger.Warnf("Get connectiong db item failed %v", err)
	}
	var pi connInfo
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
			if err := connsDB.Delete([]byte(id), nil); err != nil {
				log.Logger.Warnf("Delete item %s failed %v", id, err)
			}
		}
	} else {
		updatePeerConnHistory(id, pi)
	}
}

func updatePeerConnHistory(id string, pi connInfo) error {
	value, err := json.Marshal(pi)
	if err != nil {
		log.Logger.Warnf("Marshal failed when update connectiong db %v", err)
	}
	err = connsDB.Put([]byte(id), value, nil)
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
