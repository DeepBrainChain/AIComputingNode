package model

import (
	"AIComputingNode/pkg/types"
	"sync"
	"testing"
)

func TestSyncMap(t *testing.T) {
	pjt := make([]types.AIProjectConfig, 0)
	pjt = append(pjt, types.AIProjectConfig{
		Project: "P1",
		Models: []types.AIModelConfig{
			{
				Model: "P1-M1",
				API:   "P1-url1",
				Type:  0,
			},
			{
				Model: "P1-M2",
				API:   "P1-url2",
				Type:  0,
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
			},
			{
				Model: "P2-M2",
				API:   "P2-url2",
				Type:  0,
			},
		},
	})

	InitModels(pjt)

	t.Log("Init models")
	projects.Range(func(key, value any) bool {
		project := key.(string)
		models := value.(map[string]types.ModelInfo)
		t.Log(project, models)
		return true
	})

	IncRef("P1", "P1-M1")
	IncRef("P1", "P1-M2")

	IncRef("P2", "P2-M1")
	IncRef("P2", "P2-M2")
	IncRef("P2", "P2-M2")

	t.Log("Increase reference")
	projects.Range(func(key, value any) bool {
		project := key.(string)
		models := value.(map[string]types.ModelInfo)
		t.Log(project, models)
		return true
	})

	DecRef("P2", "P2-M1")

	t.Log("Decrease reference")
	projects.Range(func(key, value any) bool {
		project := key.(string)
		models := value.(map[string]types.ModelInfo)
		t.Log(project, models)
		return true
	})

	DecRef("P2", "P2-M1")

	t.Log("Decrease reference")
	projects.Range(func(key, value any) bool {
		project := key.(string)
		models := value.(map[string]types.ModelInfo)
		t.Log(project, models)
		return true
	})

	RegisterAIProject(types.AIProjectConfig{
		Project: "P2",
		Models: []types.AIModelConfig{
			{
				Model: "P2-M1",
				API:   "P2-url1",
				Type:  0,
			},
			{
				Model: "P2-M2",
				API:   "P2-url2",
				Type:  0,
			},
			{
				Model: "P2-M3",
				API:   "P2-url3",
				Type:  0,
			},
		},
	})

	t.Log("Register AI Project")
	projects.Range(func(key, value any) bool {
		project := key.(string)
		models := value.(map[string]types.ModelInfo)
		t.Log(project, models)
		return true
	})

	RegisterAIProject(types.AIProjectConfig{
		Project: "P2",
		Models: []types.AIModelConfig{
			{
				Model: "P2-M1",
				API:   "P2-url1",
				Type:  0,
			},
		},
	})

	t.Log("Register AI Project again")
	projects.Range(func(key, value any) bool {
		project := key.(string)
		models := value.(map[string]types.ModelInfo)
		t.Log(project, models)
		return true
	})

	UnregisterAIProject("P2")

	t.Log("Unregister AI Project")
	projects.Range(func(key, value any) bool {
		project := key.(string)
		models := value.(map[string]types.ModelInfo)
		t.Log(project, models)
		return true
	})
}

func TestConcurrentMap(t *testing.T) {
	pjt := make([]types.AIProjectConfig, 0)
	pjt = append(pjt, types.AIProjectConfig{
		Project: "P1",
		Models: []types.AIModelConfig{
			{
				Model: "P1-M1",
				API:   "P1-url1",
				Type:  0,
			},
			{
				Model: "P1-M2",
				API:   "P1-url2",
				Type:  0,
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
			},
			{
				Model: "P2-M2",
				API:   "P2-url2",
				Type:  0,
			},
		},
	})

	InitModels(pjt)

	t.Log("Init models")
	projects.Range(func(key, value any) bool {
		project := key.(string)
		models := value.(map[string]types.ModelInfo)
		t.Log(project, models)
		return true
	})

	t.Log("Before concurrent goroutine")

	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer DecRef("P1", "P1-M1")
			IncRef("P1", "P1-M1")
		}()
	}

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	IncRef("P1", "P1-M1")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	IncRef("P1", "P1-M1")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	DecRef("P1", "P1-M1")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	IncRef("P1", "P1-M2")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	IncRef("P1", "P1-M3")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	IncRef("P3", "P3-M1")
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	pjts := GetAIProjects()
	// 	for pn, models := range pjts {
	// 		t.Log(pn, models)
	// 	}
	// }()

	wg.Wait()

	t.Log("After concurrent goroutine")
	projects.Range(func(key, value any) bool {
		project := key.(string)
		models := value.(map[string]types.ModelInfo)
		t.Log(project, models)
		return true
	})
}
