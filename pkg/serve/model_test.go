package serve

import (
	"sort"
	"testing"

	"AIComputingNode/pkg/types"
)

// type ModelInfo struct {
// 	Connectivity int   `json:"connectivity"`
// 	Latency      int64 `json:"latency"`
// 	Idle         int   `json:"Idle"`
// }

// type ByLatency []ModelInfo

// func (a ByLatency) Len() int           { return len(a) }
// func (a ByLatency) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
// func (a ByLatency) Less(i, j int) bool { return a[i].Latency < a[j].Latency }

func TestModelSort(t *testing.T) {
	models := []types.AIProjectPeerInfo{
		{Connectivity: 1, Latency: 100, Idle: 2},
		{Connectivity: 0, Latency: 50, Idle: 1},
		{Connectivity: 2, Latency: 90, Idle: 0},
		{Connectivity: 1, Latency: 200, Idle: 1},
		{Connectivity: 3, Latency: 30, Idle: 0},
		{Connectivity: 1, Latency: 150, Idle: 3},
	}

	sort.Sort(types.AIProjectPeerOrder(models))

	t.Log(models)
}
