package types

import (
	"fmt"
	"net"
	"net/url"
	"time"
)

var (
	OrdinaryRequestTimeout        = 2 * time.Minute
	ChatCompletionRequestTimeout  = 3 * time.Minute
	ImageGenerationRequestTimeout = 5 * time.Minute
)

type AIProjectConfig struct {
	Project string          `json:"Project"`
	Models  []AIModelConfig `json:"Models"`
}

type AIModelConfig struct {
	Model string `json:"Model"`
	API   string `json:"API"`
	Type  int    `json:"Type"`
	CID   string `json:"CID"`
}

type AIModelRegister struct {
	Project string `json:"Project"`
	AIModelConfig
}

type AIModelUnregister struct {
	Project string `json:"Project"`
	Model   string `json:"Model"`
	CID     string `json:"CID"`
}

type ModelIdle struct {
	AIModelConfig
	Idle int `json:"Idle"`
}

func (config AIProjectConfig) Validate() error {
	if config.Project == "" {
		return fmt.Errorf("project name can not be empty")
	}
	return nil
}

func (config AIModelConfig) Validate() error {
	if config.Model == "" {
		return fmt.Errorf("model name can not be empty")
	}
	if config.API == "" {
		return fmt.Errorf("model api can not be empty")
	}
	if err := VerifyUrlAndLocalHost(config.API); err != nil {
		return err
	}
	if config.CID == "" {
		return fmt.Errorf("cid represents the docker container id, which can not be empty")
	}
	return nil
}

func (config AIModelRegister) Validate() error {
	if config.Project == "" {
		return fmt.Errorf("project name can not be empty")
	}
	if config.Model == "" {
		return fmt.Errorf("model name can not be empty")
	}
	if config.API == "" {
		return fmt.Errorf("model api can not be empty")
	}
	if err := VerifyUrlAndLocalHost(config.API); err != nil {
		return err
	}
	if config.CID == "" {
		return fmt.Errorf("cid represents the docker container id, which can not be empty")
	}
	return nil
}

func (config AIModelUnregister) Validate() error {
	if config.Project == "" {
		return fmt.Errorf("project name can not be empty")
	}
	if config.Model == "" {
		return fmt.Errorf("model name can not be empty")
	}
	// if config.CID == "" {
	// 	return fmt.Errorf("cid represents the docker container id, which can not be empty")
	// }
	return nil
}

func VerifyUrlAndLocalHost(hurl string) error {
	purl, err := url.Parse(hurl)
	if err != nil {
		return err
	}
	host := purl.Hostname()
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("parse ip from model api failed")
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return fmt.Errorf("list interface addressess failed")
	}
	for _, addr := range addrs {
		var ipNet *net.IPNet
		var ok bool
		if ipNet, ok = addr.(*net.IPNet); !ok {
			continue
		}

		if ipNet.IP.Equal(ip) {
			return nil
		}
	}
	return fmt.Errorf("the url and the node are not on the same machine")
}
