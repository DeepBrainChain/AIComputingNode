package db

import (
	"os"
	"testing"
)

func TestLevelDB(t *testing.T) {
	if err := InitDb(InitOptions{Folder: ".", ConnsDBName: "test.db"}); err != nil {
		t.Fatal("Init db failed", err)
	}
	PeerConnected("16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF",
		"/ip4/8.219.75.114/tcp/6001")
	PeerConnected("16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
		"/ip4/122.99.183.54/tcp/6001")
	peers := LoadPeerConnHistory()
	for key, value := range peers {
		t.Log("load db item", key, value)
	}
	connsDB.Close()
	os.RemoveAll("./test.db")
}
