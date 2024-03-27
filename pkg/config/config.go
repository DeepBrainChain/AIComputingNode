package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"AIComputingNode/pkg/p2p"

	"github.com/multiformats/go-multiaddr"
)

var GC *Config

type Config struct {
	Bootstrap []string       `json:"Bootstrap"`
	Addresses []string       `json:"Addresses"`
	API       APIConfig      `Json:"API"`
	Identity  IdentityConfig `json:"Identity"`
	Swarm     SwarmConfig    `json:"Swarm"`
	Routing   RoutingConfig  `json:"Routing"`
	App       AppConfig      `json:"App"`
}

type APIConfig struct {
	Addr string `json:"Addr"`
}

type IdentityConfig struct {
	PeerID  string `json:"PeerID"`
	PrivKey string `json:"PrivKey"`
}

type SwarmConfig struct {
	ConnMgr              SwarmConnMgrConfig      `json:"ConnMgr"`
	DisableNatPortMap    bool                    `json:"DisableNatPortMap"`
	RelayClient          SwarmRelayClientConfig  `json:"RelayClient"`
	RelayService         SwarmRelayServiceConfig `json:"RelayService"`
	EnableHolePunching   bool                    `json:"EnableHolePunching"`
	EnableAutoNATService bool                    `json:"EnableAutoNATService"`
}

type SwarmRelayClientConfig struct {
	Enabled      bool     `json:"Enabled"`
	StaticRelays []string `json:"StaticRelays"`
}

type SwarmRelayServiceConfig struct {
	Enabled bool `json:"Enabled"`
}

type SwarmConnMgrConfig struct {
	Type        string `json:"Type"`
	HighWater   int    `json:"HighWater"`
	LowWater    int    `json:"LowWater"`
	GracePeriod string `json:"GracePeriod"`
}

type RoutingConfig struct {
	Type           string `json:"Type"`
	ProtocolPrefix string `json:"ProtocolPrefix"`
}

type AppConfig struct {
	LogLevel     string `json:"LogLevel"`
	PreSharedKey string `json:"PreSharedKey"`
	TopicName    string `json:"TopicName"`
}

func (config Config) Validate() error {
	if len(config.Bootstrap) > 0 {
		for _, peer := range config.Bootstrap {
			_, err := multiaddr.NewMultiaddr(peer)
			if err != nil {
				return err
			}
		}
	}

	if len(config.Addresses) == 0 {
		return fmt.Errorf("addresses can not be empty")
	}
	for _, addr := range config.Addresses {
		_, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return err
		}
	}

	err := config.API.Validate()
	if err != nil {
		return err
	}

	err = config.Identity.Validate()
	if err != nil {
		return err
	}

	err = config.Swarm.Validate()
	if err != nil {
		return err
	}

	err = config.Routing.Validate()
	if err != nil {
		return err
	}

	err = config.App.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (config APIConfig) Validate() error {
	if config.Addr == "" {
		return fmt.Errorf("http api address can not be empty")
	}
	return nil
}

func (config IdentityConfig) Validate() error {
	peer, err := p2p.PeerIDFromPrivKeyString(config.PrivKey)
	if err != nil {
		return err
	}
	if peer.String() != config.PeerID {
		return fmt.Errorf("private key and peer id do not match")
	}
	return nil
}

func (config SwarmConfig) Validate() error {
	err := config.ConnMgr.Validate()
	if err != nil {
		return err
	}
	return nil
}

func (config SwarmConnMgrConfig) Validate() error {
	if config.Type != "basic" && config.Type != "none" {
		return fmt.Errorf("unknowned connect manager type")
	}
	_, err := time.ParseDuration(config.GracePeriod)
	if err != nil {
		return err
	}

	return nil
}

func (config RoutingConfig) Validate() error {
	if config.Type != "auto" && config.Type != "none" &&
		config.Type != "dhtclient" && config.Type != "dhtserver" {
		return fmt.Errorf("unknowned routing type")
	}
	return nil
}

func (config AppConfig) Validate() error {
	if config.LogLevel != "debug" && config.LogLevel != "info" &&
		config.LogLevel != "warn" && config.LogLevel != "error" &&
		config.LogLevel != "panic" && config.LogLevel != "fatal" {
		return fmt.Errorf("unknowned log level")
	}
	if config.TopicName == "" {
		return fmt.Errorf("topic name can not be empty")
	}
	return nil
}

func LoadConfig(configPath string) (*Config, error) {
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	GC = &Config{}
	err = json.Unmarshal(configFile, GC)
	if err != nil {
		return nil, err
	}

	if GC.Swarm.ConnMgr.Type == "" {
		GC.Swarm.ConnMgr.Type = "basic"
	}

	if GC.Swarm.ConnMgr.LowWater == 0 {
		GC.Swarm.ConnMgr.LowWater = 100
	}

	if GC.Swarm.ConnMgr.HighWater == 0 {
		GC.Swarm.ConnMgr.HighWater = 400
	}

	if GC.Routing.Type == "" {
		GC.Routing.Type = "auto"
	}

	if GC.App.LogLevel == "" {
		GC.App.LogLevel = "info"
	}

	return GC, nil
}
