package types

import (
	"fmt"
	"net"
	"net/url"
)

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
	purl, err := url.Parse(config.API)
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
	return fmt.Errorf("the AI model and the node are not on the same machine")
}
