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

func RegisterAIModel(mr types.AIModelRegister) {
	projects.mutex.Lock()
	defer projects.mutex.Unlock()

	existed := -1
	for pn, models := range projects.elements {
		if pn == mr.Project {
			for i, model := range models {
				if model.Model == mr.Model && model.CID == mr.CID {
					existed = i
					break
				}
			}
		}
	}

	if existed == -1 {
		if models, ok := projects.elements[mr.Project]; ok {
			models = append(models, types.ModelIdle{
				AIModelConfig: mr.AIModelConfig,
				Idle:          0,
			})
			projects.elements[mr.Project] = models
		} else {
			models := make([]types.ModelIdle, 0)
			models = append(models, types.ModelIdle{
				AIModelConfig: mr.AIModelConfig,
				Idle:          0,
			})
			projects.elements[mr.Project] = models
		}
	} else {
		models := projects.elements[mr.Project]
		models[existed] = types.ModelIdle{
			AIModelConfig: mr.AIModelConfig,
			Idle:          models[existed].Idle,
		}
		projects.elements[mr.Project] = models
	}
}

func UnregisterAIModel(projectName, modelName, cid string) {
	projects.mutex.Lock()
	defer projects.mutex.Unlock()

	models, ok := projects.elements[projectName]
	if ok {
		for i, model := range models {
			if model.Model == modelName && model.CID == cid {
				models = append(models[:i], models[i+1:]...)
				break
			}
		}
	}
	if len(models) == 0 {
		delete(projects.elements, projectName)
	} else {
		projects.elements[projectName] = models
	}
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
