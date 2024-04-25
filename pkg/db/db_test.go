package db

import (
	"os"
	"testing"
	"time"
)

func TestTimeDiff(t *testing.T) {
	t1 := time.Now()
	t.Log("current time", t1)
	t1i64 := t1.Unix()
	t.Log("current time in unix", t1i64)
	t2 := time.Unix(t1i64, 0)
	t.Log("time from int64", t2)
	td1 := t2.Sub(t1)
	t.Log("time sub", td1)

	st := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.Local)
	t.Log("start time", st)
	td2 := t2.Sub(st)
	t.Log("time sub", td2)
	t.Logf("time diff %.2f days ~ %.f hours ~ %.f minutes ~ %.f seconds",
		td2.Hours()/24, td2.Hours(), td2.Minutes(), td2.Seconds())
}

func TestLevelDB(t *testing.T) {
	if err := InitDb(InitOptions{Folder: ".", PeersDBName: "test.db"}); err != nil {
		t.Fatal("Init db failed", err)
	}
	PeerConnected("16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF",
		"/ip4/8.219.75.114/tcp/6001")
	PeerConnected("16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
		"/ip4/122.99.183.54/tcp/6001")
	peers := LoadPeers()
	for key, value := range peers {
		t.Log("load db item", key, value)
	}
	peersDB.Close()
	os.RemoveAll("./test.db")
}
