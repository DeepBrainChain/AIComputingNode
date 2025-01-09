package types

import (
	"testing"
)

// go test -v -timeout 30s -count=1 -run TestAIRegister AIComputingNode/pkg/types
func TestAIRegister(t *testing.T) {
	aipjts := make([]AIProjectConfig, 0)

	listHandler := func(pjts []AIProjectConfig) {
		for _, pjt := range pjts {
			t.Log(pjt)
		}
	}

	registerProjectHandler := func(req AIProjectConfig) {
		find := -1
		for i := range aipjts {
			if aipjts[i].Project == req.Project {
				find = i
				break
			}
		}
		if find == -1 {
			aipjts = append(aipjts, req)
		} else {
			aipjts[find].Models = req.Models
		}
	}

	unregisterProjectHandler := func(req AIProjectConfig) {
		find := -1
		for i := range aipjts {
			if aipjts[i].Project == req.Project {
				find = i
				break
			}
		}
		if find == -1 {
			t.Log("not existed")
			return
		} else {
			aipjts = append(aipjts[:find], aipjts[find+1:]...)
		}
	}

	registerModelHandler := func(req AIModelRegister) {
		pfind := -1
		mfind := -1
		for i, project := range aipjts {
			if project.Project == req.Project {
				pfind = i
				for j, model := range project.Models {
					if model.Model == req.Model && model.CID == req.CID {
						mfind = j
						break
					}
				}
			}
		}
		if pfind == -1 {
			models := make([]AIModelConfig, 0)
			models = append(models, AIModelConfig{
				Model: req.Model,
				API:   req.API,
				Type:  req.Type,
				CID:   req.CID,
			})
			aipjts = append(aipjts, AIProjectConfig{
				Project: req.Project,
				Models:  models,
			})
		} else if mfind == -1 {
			models := aipjts[pfind].Models
			models = append(models, AIModelConfig{
				Model: req.Model,
				API:   req.API,
				Type:  req.Type,
				CID:   req.CID,
			})
			aipjts[pfind].Models = models
		} else {
			models := aipjts[pfind].Models
			models[mfind] = AIModelConfig{
				Model: req.Model,
				API:   req.API,
				Type:  req.Type,
				CID:   req.CID,
			}
			aipjts[pfind].Models = models
		}
	}

	unregisterModelHandler := func(req AIModelUnregister) {
		pfind := -1
		mfind := -1
		for i, project := range aipjts {
			if project.Project == req.Project {
				pfind = i
				for j, model := range project.Models {
					if model.Model == req.Model && model.CID == req.CID {
						mfind = j
						break
					}
				}
			}
		}

		if mfind == -1 {
			t.Log("not existed")
			return
		}

		models := aipjts[pfind].Models
		models = append(models[:mfind], models[mfind+1:]...)
		aipjts[pfind].Models = models
		if len(models) == 0 {
			aipjts = append(aipjts[:pfind], aipjts[pfind+1:]...)
		}
	}

	listHandler(aipjts)

	registerProjectHandler(AIProjectConfig{
		Project: "P1",
		Models: []AIModelConfig{
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

	registerProjectHandler(AIProjectConfig{
		Project: "P2",
		Models: []AIModelConfig{
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

	t.Log("After register ai project")
	listHandler(aipjts)

	registerProjectHandler(AIProjectConfig{
		Project: "P2",
		Models: []AIModelConfig{
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

	t.Log("After register ai project")
	listHandler(aipjts)

	unregisterProjectHandler(AIProjectConfig{
		Project: "P1",
	})

	t.Log("After unregister ai project")
	listHandler(aipjts)

	unregisterProjectHandler(AIProjectConfig{
		Project: "P1",
	})

	t.Log("After unregister ai project which not existed")
	listHandler(aipjts)

	registerModelHandler(AIModelRegister{
		Project: "P1",
		AIModelConfig: AIModelConfig{
			Model: "P1-M1",
			API:   "P1-url1",
			Type:  0,
			CID:   "P1-M1",
		},
	})

	t.Log("After register ai model")
	listHandler(aipjts)

	registerModelHandler(AIModelRegister{
		Project: "P1",
		AIModelConfig: AIModelConfig{
			Model: "P1-M2",
			API:   "P1-url2",
			Type:  0,
			CID:   "",
		},
	})

	t.Log("After register ai model")
	listHandler(aipjts)

	registerModelHandler(AIModelRegister{
		Project: "P1",
		AIModelConfig: AIModelConfig{
			Model: "P1-M2",
			API:   "P1-url2",
			Type:  1,
			CID:   "",
		},
	})

	t.Log("After register ai model with modify")
	listHandler(aipjts)

	unregisterModelHandler(AIModelUnregister{
		Project: "P1",
		Model:   "P1-M1",
		CID:     "P1-M1",
	})

	t.Log("After unregister ai model")
	listHandler(aipjts)

	unregisterModelHandler(AIModelUnregister{
		Project: "P1",
		Model:   "P1-M2",
		CID:     "P1-M2",
	})

	t.Log("After unregister ai model which not existed")
	listHandler(aipjts)

	unregisterModelHandler(AIModelUnregister{
		Project: "P1",
		Model:   "P1-M2",
		CID:     "",
	})

	t.Log("After unregister ai model")
	listHandler(aipjts)
}
