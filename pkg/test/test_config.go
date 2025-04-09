package test

import (
	"encoding/json"
	"os"
)

type ModelInfo struct {
	Name string `json:"name"`
	API  string `json:"api"`
}

type Models struct {
	Llama3     ModelInfo `json:"Llama3"`
	Llama31    ModelInfo `json:"Llama3.1"`
	Qwen2      ModelInfo `json:"Qwen2"`
	SuperImage ModelInfo `json:"SuperImage"`
	StyleID    ModelInfo `json:"StyleID"`
}

type Config struct {
	Models Models
}

func LoadConfig(configFile string) (*Config, error) {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
