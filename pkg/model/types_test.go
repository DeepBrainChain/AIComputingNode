package model

import (
	"AIComputingNode/pkg/types"
	"sync"
	"testing"
)

// go test -v -timeout 30s -count=1 -run TestSyncMap AIComputingNode/pkg/model
func TestSyncMap(t *testing.T) {
	pjt := make([]types.AIProjectConfig, 0)
	pjt = append(pjt, types.AIProjectConfig{
		Project: "P1",
		Models: []types.AIModelConfig{
			{
				Model: "P1-M1",
				API:   "P1-url1",
				Type:  0,
				CID:   "P1-M1",
			},
			{
				Model: "P1-M2",
				API:   "P1-url2",
				Type:  0,
				CID:   "P1-M2",
			},
		},
	})
	pjt = append(pjt, types.AIProjectConfig{
		Project: "P2",
		Models: []types.AIModelConfig{
			{
				Model: "P2-M1",
				API:   "P2-url1",
				Type:  0,
				CID:   "P2-M1",
			},
			{
				Model: "P2-M2",
				API:   "P2-url2",
				Type:  0,
				CID:   "P2-M2",
			},
		},
	})

	InitModels(pjt)

	t.Log("Init models")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	IncRef("P1", "P1-M1", "")
	IncRef("P1", "P1-M1", "P1-M1")
	IncRef("P1", "P1-M2", "P1-M2")

	IncRef("P2", "P2-M1", "P2-M1")
	IncRef("P2", "P2-M2", "P2-M2")
	IncRef("P2", "P2-M2", "P2-M2")

	t.Log("Increase reference")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	DecRef("P2", "P2-M1", "P2-M1")

	t.Log("Decrease reference")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	DecRef("P2", "P2-M1", "P2-M1")

	t.Log("Decrease reference")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	RegisterAIProject(types.AIProjectConfig{
		Project: "P2",
		Models: []types.AIModelConfig{
			{
				Model: "P2-M1",
				API:   "P2-url1",
				Type:  0,
				CID:   "P2-M1",
			},
			{
				Model: "P2-M2",
				API:   "P2-url2",
				Type:  0,
				CID:   "P2-M2",
			},
			{
				Model: "P2-M3",
				API:   "P2-url3",
				Type:  0,
				CID:   "P2-M3",
			},
		},
	})

	t.Log("Register AI Project")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	RegisterAIProject(types.AIProjectConfig{
		Project: "P2",
		Models: []types.AIModelConfig{
			{
				Model: "P2-M1",
				API:   "P2-url1",
				Type:  0,
				CID:   "P2-M1",
			},
		},
	})

	t.Log("Register AI Project again")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	UnregisterAIProject("P2")

	t.Log("Unregister AI Project")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	RegisterAIModel(types.AIModelRegister{
		AIModelConfig: types.AIModelConfig{
			Model: "P2-M1",
			API:   "P2-url1",
			Type:  0,
			CID:   "P2-M1",
		},
		Project: "P2",
	})

	RegisterAIModel(types.AIModelRegister{
		AIModelConfig: types.AIModelConfig{
			Model: "P2-M2",
			API:   "P2-url2",
			Type:  0,
			CID:   "P2-M2",
		},
		Project: "P2",
	})

	t.Log("Register AI Model")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	IncRef("P2", "P2-M1", "P2-M1")
	IncRef("P2", "P2-M2", "P2-M2")

	t.Log("Increase reference")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	DecRef("P2", "P2-M1", "P2-M1")

	t.Log("Decrease reference")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	RegisterAIModel(types.AIModelRegister{
		AIModelConfig: types.AIModelConfig{
			Model: "P2-M2",
			API:   "P2-url3",
			Type:  1,
			CID:   "P2-M2",
		},
		Project: "P2",
	})

	t.Log("Register AI Model again with modify")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	UnregisterAIModel("P2", "P2-M1", "")

	t.Log("Unregister AI Model")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	UnregisterAIModel("P2", "P2-M1", "P2-M1")

	t.Log("Unregister AI Model again")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	UnregisterAIModel("P2", "P2-M2", "P2-M2")

	t.Log("Unregister AI Model again again")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}
}

// go test -v -timeout 30s -count=1 -run TestConcurrentMap AIComputingNode/pkg/model
func TestConcurrentMap(t *testing.T) {
	pjt := make([]types.AIProjectConfig, 0)
	pjt = append(pjt, types.AIProjectConfig{
		Project: "P1",
		Models: []types.AIModelConfig{
			{
				Model: "P1-M1",
				API:   "P1-url1",
				Type:  0,
				CID:   "P1-M1",
			},
			{
				Model: "P1-M2",
				API:   "P1-url2",
				Type:  0,
				CID:   "P1-M2",
			},
		},
	})
	pjt = append(pjt, types.AIProjectConfig{
		Project: "P2",
		Models: []types.AIModelConfig{
			{
				Model: "P2-M1",
				API:   "P2-url1",
				Type:  0,
				CID:   "P2-M1",
			},
			{
				Model: "P2-M2",
				API:   "P2-url2",
				Type:  0,
				CID:   "P2-M2",
			},
		},
	})

	InitModels(pjt)

	t.Log("Init models")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}

	t.Log("Before concurrent goroutine")

	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer DecRef("P1", "P1-M1", "P1-M1")
			IncRef("P1", "P1-M1", "P1-M1")
		}()
	}

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	IncRef("P1", "P1-M1", "P1-M1")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	IncRef("P1", "P1-M1", "P1-M1")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	DecRef("P1", "P1-M1", "P1-M1")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	IncRef("P1", "P1-M2", "P1-M2")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	IncRef("P1", "P1-M3", "P1-M3")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	IncRef("P3", "P3-M1", "P3-M1")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	for pn, models := range GetAIProjects() {
	// 		t.Log(pn, models)
	// 	}
	// }()

	wg.Wait()

	t.Log("After concurrent goroutine")
	for pn, models := range GetAIProjects() {
		t.Log(pn, models)
	}
}
