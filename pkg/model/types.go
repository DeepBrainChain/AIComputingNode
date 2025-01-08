package model

import (
	"fmt"
	"sync"

	"AIComputingNode/pkg/types"
)

type ProjectMap struct {
	mutex    sync.RWMutex
	elements map[string][]types.ModelIdle
}

var projects = ProjectMap{
	mutex:    sync.RWMutex{},
	elements: make(map[string][]types.ModelIdle),
}

func InitModels(ms []types.AIProjectConfig) error {
	for _, pc := range ms {
		models := make([]types.ModelIdle, 0)
		for _, model := range pc.Models {
			models = append(models, types.ModelIdle{
				AIModelConfig: model,
				Idle:          0,
			})
		}
		projects.elements[pc.Project] = models
	}
	return nil
}

func IdleCount() int {
	projects.mutex.RLock()
	defer projects.mutex.RUnlock()
	idleCount := 0
	for _, models := range projects.elements {
		for _, model := range models {
			idleCount += model.Idle
		}
	}
	return idleCount
}

func GetAIProjects() map[string][]types.ModelIdle {
	projects.mutex.RLock()
	defer projects.mutex.RUnlock()
	// Do not modify the returned value
	// return projects.elements
	res := make(map[string][]types.ModelIdle)
	for pn, models := range projects.elements {
		ms := make([]types.ModelIdle, len(models))
		copy(ms, models)
		res[pn] = ms
	}
	return res
}

func GetModelInfo(projectName, modelName, cid string) (*types.ModelIdle, error) {
	mi := &types.ModelIdle{}
	if projectName == "" || modelName == "" {
		return mi, fmt.Errorf("empty project or model")
	}
	projects.mutex.RLock()
	defer projects.mutex.RUnlock()
	for pn, models := range projects.elements {
		if pn == projectName {
			for _, model := range models {
				if model.Model == modelName && (cid == "" || model.CID == cid) {
					if mi.API == "" {
						*mi = model
					} else {
						if model.Idle < mi.Idle {
							*mi = model
						}
					}
				}
			}
		}
	}
	if mi.API == "" {
		return mi, fmt.Errorf("can not find in registered project and model")
	}
	return mi, nil
}

func RegisterAIProject(pjt types.AIProjectConfig) {
	projects.mutex.Lock()
	defer projects.mutex.Unlock()

	models := make([]types.ModelIdle, 0)
	for _, model := range pjt.Models {
		models = append(models, types.ModelIdle{
			AIModelConfig: model,
			Idle:          0,
		})
	}
	if old, ok := projects.elements[pjt.Project]; ok {
		for _, value := range old {
			for i, mi := range models {
				if value.Model == mi.Model && value.CID == mi.CID {
					mi.Idle = value.Idle
					models[i] = mi
					break
				}
			}
		}
	}
	projects.elements[pjt.Project] = models
}

func UnregisterAIProject(project string) {
	projects.mutex.Lock()
	defer projects.mutex.Unlock()
	delete(projects.elements, project)
}

// Increase Reference
func IncRef(project, model, cid string) {
	projects.mutex.Lock()
	defer projects.mutex.Unlock()
	if models, ok := projects.elements[project]; ok {
		for i, mi := range models {
			if mi.Model == model && mi.CID == cid {
				mi.Idle = mi.Idle + 1
				models[i] = mi
				break
			}
		}
	}
}

// Decrease reference
func DecRef(project, model, cid string) {
	projects.mutex.Lock()
	defer projects.mutex.Unlock()
	if models, ok := projects.elements[project]; ok {
		for i, mi := range models {
			if mi.Model == model && mi.CID == cid {
				if mi.Idle > 0 {
					mi.Idle = mi.Idle - 1
					models[i] = mi
				}
				break
			}
		}
	}
}
