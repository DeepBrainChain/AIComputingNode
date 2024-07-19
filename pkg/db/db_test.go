package db

import (
	"os"
	"testing"
)

func TestLevelDB(t *testing.T) {
	if err := InitDb(InitOptions{
		Folder:             ".",
		ConnsDBName:        "conns.db",
		ModelsDBName:       "models.db",
		EnablePeersCollect: false,
	}); err != nil {
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

	data, err := connsDB.Get([]byte("16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF"), nil)
	if err != nil {
		t.Errorf("Level db get item failed %v", err)
	}
	t.Logf("Level db get item value %s", string(data))

	none, err := connsDB.Get([]byte("1234567"), nil)
	if err != nil {
		t.Errorf("Level db get not existed item failed %v", err)
	}
	t.Logf("Level db get not existed item value %v", none)

	connsDB.Close()
	os.RemoveAll("./conns.db")
	modelsDB.Close()
	os.RemoveAll("./models.db")
}
