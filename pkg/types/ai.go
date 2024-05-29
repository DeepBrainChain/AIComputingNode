package types

import "fmt"

type AIProject struct {
	Project string    `json:"Project"`
	Models  []AIModel `json:"Models"`
}

type AIModel struct {
	Model string `json:"Model"`
	API   string `json:"API"`
	Type  int    `json:"Type"`
}

type AIProjectOfNode struct {
	Project string   `json:"project"`
	Models  []string `json:"models"`
}

func (config AIProject) Validate() error {
	if config.Project == "" {
		return fmt.Errorf("project name can not be empty")
	}
	return nil
}

func (config AIModel) Validate() error {
	if config.Model == "" {
		return fmt.Errorf("model name can not be empty")
	}
	if config.API == "" {
		return fmt.Errorf("model api can not be empty")
	}
	return nil
}
