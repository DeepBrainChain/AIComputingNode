package db

import (
	"AIComputingNode/pkg/types"
	"encoding/json"
	"os"
	"testing"
	"time"
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
	} else {
		t.Logf("Level db get item value %s", string(data))
	}

	none, err := connsDB.Get([]byte("1234567"), nil)
	if err != nil {
		t.Errorf("Level db get not existed item failed %v", err)
	} else {
		t.Logf("Level db get not existed item value %v", none)
	}

	connsDB.Close()
	os.RemoveAll("./conns.db")
	modelsDB.Close()
	os.RemoveAll("./models.db")
}

// go test -v -timeout 30s -count=1 -run TestGetPeersOfAIProject AIComputingNode/pkg/db
func TestGetPeersOfAIProject(t *testing.T) {
	if err := InitDb(InitOptions{
		Folder:             ".",
		ConnsDBName:        "conns.db",
		ModelsDBName:       "models.db",
		EnablePeersCollect: true,
	}); err != nil {
		t.Fatal("Init db failed", err)
	}

	UpdatePeerCollect(
		"16Uiu2HAkyKgcoYhNsTfrX3mq2wJiZAGnD4DrMdZZb8dbofZ3jrb8",
		PeerCollectInfo{
			AIProjects: map[string][]types.ModelIdle{
				"SuperImageAI": []types.ModelIdle{
					{
						AIModelConfig: types.AIModelConfig{
							Model: "superImage",
							API:   "http://127.0.0.1:1088/v1/images/generations",
							Type:  1,
							CID:   "",
						},
						Idle: 0,
					},
				},
			},
			NodeType:  4,
			Timestamp: time.Now().Unix(),
		},
	)
	UpdatePeerCollect(
		"16Uiu2HAm4szuwGhRmXTBs7F2anRo9y9cQrcfZH7RXwEsVa1GrXVd",
		PeerCollectInfo{
			AIProjects: map[string][]types.ModelIdle{
				"DecentralGPT": []types.ModelIdle{
					{
						AIModelConfig: types.AIModelConfig{
							Model: "Codestral-22B-v0.1",
							API:   "http://127.0.0.1:8100/v1/chat/completions",
							Type:  0,
							CID:   "",
						},
						Idle: 0,
					},
				},
			},
			NodeType:  4,
			Timestamp: time.Now().Unix(),
		},
	)
	UpdatePeerCollect(
		"16Uiu2HAmAZmg7WcW8jK6mkjFx6HBbc1HtPWFk88cjjDyvf6MYw8D",
		PeerCollectInfo{
			AIProjects: map[string][]types.ModelIdle{
				"DecentralGPT": []types.ModelIdle{
					{
						AIModelConfig: types.AIModelConfig{
							Model: "DeepSeek-Coder-V2",
							API:   "http://127.0.0.1:5042/v1/chat/completions",
							Type:  0,
							CID:   "",
						},
						Idle: 0,
					},
					{
						AIModelConfig: types.AIModelConfig{
							Model: "Llama-3.1-Nemotron-70B",
							API:   "http://127.0.0.1:1042/v1/chat/completions",
							Type:  0,
							CID:   "",
						},
						Idle: 0,
					},
					{
						AIModelConfig: types.AIModelConfig{
							Model: "NVLM-D-72B",
							API:   "http://127.0.0.1:3042/v1/chat/completions",
							Type:  0,
							CID:   "",
						},
						Idle: 0,
					},
					{
						AIModelConfig: types.AIModelConfig{
							Model: "Qwen2.5-72B",
							API:   "http://127.0.0.1:4042/v1/chat/completions",
							Type:  0,
							CID:   "",
						},
						Idle: 0,
					},
					{
						AIModelConfig: types.AIModelConfig{
							Model: "Qwen2.5-Coder-32B",
							API:   "http://127.0.0.1:6042/v1/chat/completions",
							Type:  0,
							CID:   "",
						},
						Idle: 0,
					},
				},
				"SuperImageAI": []types.ModelIdle{
					{
						AIModelConfig: types.AIModelConfig{
							Model: "FLUX.1-dev",
							API:   "http://127.0.0.1:2042/v1/images/generations",
							Type:  1,
							CID:   "",
						},
						Idle: 0,
					},
				},
			},
			NodeType:  5,
			Timestamp: time.Now().Unix(),
		},
	)
	UpdatePeerCollect(
		"16Uiu2HAmDBYxgdKxeCbmn8hYiqwK3xHR9533WDdEYmpEDQ259GTe",
		PeerCollectInfo{
			AIProjects: map[string][]types.ModelIdle{},
			NodeType:   0,
			Timestamp:  time.Now().Unix(),
		},
	)
	UpdatePeerCollect(
		"16Uiu2HAmDR6u7WCPFnhJwx4P9FsXhtu9hdtnyhTC2BF8BGXw1ZiG",
		PeerCollectInfo{
			AIProjects: map[string][]types.ModelIdle{},
			NodeType:   0,
			Timestamp:  time.Now().Unix(),
		},
	)
	UpdatePeerCollect(
		"16Uiu2HAmEav9vvFGZR4qLey63Khrc7Bqnu58ESDt2NHzWrZByntU",
		PeerCollectInfo{
			AIProjects: map[string][]types.ModelIdle{
				"DecentralGPT": []types.ModelIdle{
					{
						AIModelConfig: types.AIModelConfig{
							Model: "Llama-3.1-405B",
							API:   "http://127.0.0.1:1042/v1/chat/completions",
							Type:  0,
							CID:   "#1",
						},
						Idle: 1,
					},
					{
						AIModelConfig: types.AIModelConfig{
							Model: "Llama-3.1-405B",
							API:   "http://127.0.0.1:2042/v1/chat/completions",
							Type:  0,
							CID:   "#2",
						},
						Idle: 0,
					},
					{
						AIModelConfig: types.AIModelConfig{
							Model: "Llama-3.1-405B",
							API:   "http://127.0.0.1:3042/v1/chat/completions",
							Type:  0,
							CID:   "#3",
						},
						Idle: 2,
					},
				},
			},
			NodeType:  4,
			Timestamp: time.Now().Unix(),
		},
	)
	UpdatePeerCollect(
		"16Uiu2HAmKk7Fg4WysTpEGd5q1wH2NL4wmxyQ5Nj4HhkQHyB3bDhm",
		PeerCollectInfo{
			AIProjects: map[string][]types.ModelIdle{},
			NodeType:   2,
			Timestamp:  time.Now().Unix(),
		},
	)
	UpdatePeerCollect(
		"16Uiu2HAmLCRpZv6nUmeAmXoWpXeyrKjZ7pUvqx5m3e5gZMmUzScp",
		PeerCollectInfo{
			AIProjects: map[string][]types.ModelIdle{},
			NodeType:   3,
			Timestamp:  time.Now().Unix(),
		},
	)
	UpdatePeerCollect(
		"16Uiu2HAmRTpigc7jAbsLndB2xDEBMAXLb887SBEFhfdJeEJNtqRM",
		PeerCollectInfo{
			AIProjects: map[string][]types.ModelIdle{},
			NodeType:   3,
			Timestamp:  time.Now().Unix(),
		},
	)
	UpdatePeerCollect(
		"16Uiu2HAmS4CErxrmPryJbbEX2HFQbLK8r8xCA5rmzdSU59rHc9AF",
		PeerCollectInfo{
			AIProjects: map[string][]types.ModelIdle{},
			NodeType:   3,
			Timestamp:  time.Now().Unix(),
		},
	)

	iter := peersCollectDB.NewIterator(nil, nil)
	for iter.Next() {
		var info PeerCollectInfo
		if err := json.Unmarshal(iter.Value(), &info); err != nil {
			t.Logf("Parse failed when load peer collect info of %s : %v", string(iter.Key()), err)
			continue
		}
		t.Logf("%s - %v\n", string(iter.Key()), info)
	}

	iter.Release()
	if err := iter.Error(); err != nil {
		t.Logf("Iterator failed when load peer collect info %v", err)
	}

	ids, code := GetPeersOfAIProjects("DecentralGPT", "Codestral-22B-v0.1", 20)
	if code != 0 {
		t.Log("GetPeersOfAIProjects failed ", code)
	} else {
		t.Log("GetPeersOfAIProjects of Codestral-22B-v0.1 ", ids)
	}

	ids, code = GetPeersOfAIProjects("DecentralGPT", "Qwen2.5-Coder-32B", 20)
	if code != 0 {
		t.Log("GetPeersOfAIProjects failed ", code)
	} else {
		t.Log("GetPeersOfAIProjects of Qwen2.5-Coder-32B ", ids)
	}

	ids, code = GetPeersOfAIProjects("DecentralGPT", "Llama-3.1-405B", 20)
	if code != 0 {
		t.Log("GetPeersOfAIProjects failed ", code)
	} else {
		t.Log("GetPeersOfAIProjects of Llama-3.1-405B ", ids)
	}

	connsDB.Close()
	os.RemoveAll("./conns.db")
	modelsDB.Close()
	os.RemoveAll("./models.db")
	peersCollectDB.Close()
	os.RemoveAll("./peers_collect.db")
}
