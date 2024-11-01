package model

import (
	"sync"

	"AIComputingNode/pkg/types"
)

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

// Increase Reference
func IncRef(project, model string) {
	UpdateModel(project, func(old interface{}) interface{} {
		if models, ok := old.(map[string]types.ModelInfo); ok {
			if mi, ok := models[model]; ok {
				mi.Idle = mi.Idle + 1
				models[model] = mi
				return models
			}
		}
		return nil
	})
}

// Decrease reference
func DecRef(project, model string) {
	UpdateModel(project, func(old interface{}) interface{} {
		if models, ok := old.(map[string]types.ModelInfo); ok {
			if mi, ok := models[model]; ok {
				if mi.Idle > 0 {
					mi.Idle = mi.Idle - 1
					models[model] = mi
					return models
				}
			}
		}
		return nil
	})
}
