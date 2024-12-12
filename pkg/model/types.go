package model

import (
	"sync"

	"AIComputingNode/pkg/types"
)

type ProjectMap struct {
	mutex    sync.RWMutex
	elements map[string]map[string]types.ModelInfo
}

var projects = ProjectMap{
	mutex:    sync.RWMutex{},
	elements: make(map[string]map[string]types.ModelInfo),
}

func InitModels(ms []types.AIProjectConfig) error {
	for _, pc := range ms {
		models := make(map[string]types.ModelInfo)
		for _, model := range pc.Models {
			models[model.Model] = types.ModelInfo{
				API:  model.API,
				Type: model.Type,
				Idle: 0,
			}
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

func GetAIProjects() map[string]map[string]types.ModelInfo {
	projects.mutex.RLock()
	defer projects.mutex.RUnlock()
	// Do not modify the returned value
	// return projects.elements
	res := make(map[string]map[string]types.ModelInfo)
	for pn, models := range projects.elements {
		ms := make(map[string]types.ModelInfo)
		for mn, model := range models {
			ms[mn] = model
		}
		res[pn] = ms
	}
	return res
}

func RegisterAIProject(pjt types.AIProjectConfig) {
	projects.mutex.Lock()
	defer projects.mutex.Unlock()

	models := make(map[string]types.ModelInfo)
	for _, model := range pjt.Models {
		models[model.Model] = types.ModelInfo{
			API:  model.API,
			Type: model.Type,
			Idle: 0,
		}
	}
	if old, ok := projects.elements[pjt.Project]; ok {
		for key, value := range old {
			if mi, ok := models[key]; ok {
				mi.Idle = value.Idle
				models[key] = mi
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
func IncRef(project, model string) {
	projects.mutex.Lock()
	defer projects.mutex.Unlock()
	if models, ok := projects.elements[project]; ok {
		if mi, ok := models[model]; ok {
			// mi.Idle = mi.Idle + 1
			models[model] = types.ModelInfo{
				API:  mi.API,
				Type: mi.Type,
				Idle: mi.Idle + 1,
			}
		}
	}
}

// Decrease reference
func DecRef(project, model string) {
	projects.mutex.Lock()
	defer projects.mutex.Unlock()
	if models, ok := projects.elements[project]; ok {
		if mi, ok := models[model]; ok && mi.Idle > 0 {
			// mi.Idle = mi.Idle - 1
			models[model] = types.ModelInfo{
				API:  mi.API,
				Type: mi.Type,
				Idle: mi.Idle - 1,
			}
		}
	}
}

/*
var projects = sync.Map{}

func InitModels(ms []types.AIProjectConfig) error {
	for _, pc := range ms {
		models := make(map[string]types.ModelInfo)
		for _, model := range pc.Models {
			models[model.Model] = types.ModelInfo{
				API:  model.API,
				Type: model.Type,
				Idle: 0,
			}
		}
		projects.Store(pc.Project, models)
	}
	return nil
}

func GetAIProjects() map[string]map[string]types.ModelInfo {
	projs := make(map[string]map[string]types.ModelInfo)
	projects.Range(func(key, value any) bool {
		project := key.(string)
		models := value.(map[string]types.ModelInfo)
		projs[project] = models
		return true
	})
	return projs
}

func RegisterAIProject(pjt types.AIProjectConfig) {
	models := make(map[string]types.ModelInfo)
	for _, model := range pjt.Models {
		models[model.Model] = types.ModelInfo{
			API:  model.API,
			Type: model.Type,
			Idle: 0,
		}
	}
	if old, ok := projects.Load(pjt.Project); ok {
		if oldModels, ok := old.(map[string]types.ModelInfo); ok {
			for key, value := range oldModels {
				if mi, ok := models[key]; ok {
					mi.Idle = value.Idle
					models[key] = mi
				}
			}
		}
	}
	projects.Store(pjt.Project, models)
}

func UnregisterAIProject(project string) {
	projects.Delete(project)
}

func UpdateModel(project string, updateFunc func(interface{}) interface{}) {
	if old, ok := projects.Load(project); ok {
		if models := updateFunc(old); models != nil {
			projects.Store(project, models)
		}
	}
}

// The external map uses the Load method to obtain a reference to the internal map, and the
// concurrency safety of the internal map cannot be guaranteed unless both layers of maps use sync.Map.
// Concurrent modification by multiple threads may cause a crash of "fatal error: concurrent map writes".

// Increase Reference
func IncRef(project, model string) {
	UpdateModel(project, func(old interface{}) interface{} {
		if models, ok := old.(map[string]types.ModelInfo); ok {
			new := make(map[string]types.ModelInfo)
			for key, value := range models {
				if key == model {
					new[key] = types.ModelInfo{
						API:  value.API,
						Type: value.Type,
						Idle: value.Idle + 1,
					}
				} else {
					new[key] = value
				}
			}
			return new
			// if mi, ok := models[model]; ok {
			// 	mi.Idle = mi.Idle + 1
			// 	models[model] = mi
			// 	return models
			// }
		}
		return nil
	})
}

// Decrease reference
func DecRef(project, model string) {
	UpdateModel(project, func(old interface{}) interface{} {
		if models, ok := old.(map[string]types.ModelInfo); ok {
			new := make(map[string]types.ModelInfo)
			for key, value := range models {
				if key == model {
					idle := value.Idle - 1
					if idle < 0 {
						idle = 0
					}
					new[key] = types.ModelInfo{
						API:  value.API,
						Type: value.Type,
						Idle: idle,
					}
				} else {
					new[key] = value
				}
			}
			return new
			// if mi, ok := models[model]; ok {
			// 	if mi.Idle > 0 {
			// 		mi.Idle = mi.Idle - 1
			// 		models[model] = mi
			// 		return models
			// 	}
			// }
		}
		return nil
	})
}
*/
