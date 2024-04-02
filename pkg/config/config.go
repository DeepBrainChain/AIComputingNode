package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"AIComputingNode/pkg/p2p"

	"github.com/mattn/go-isatty"
	"github.com/multiformats/go-multiaddr"
)

var GC *Config

type Config struct {
	Bootstrap []string       `json:"Bootstrap"`
	Addresses []string       `json:"Addresses"`
	API       APIConfig      `Json:"API"`
	Identity  IdentityConfig `json:"Identity"`
	Swarm     SwarmConfig    `json:"Swarm"`
	Pubsub    PubsubConfig   `json:Pubsub`
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

type PubsubConfig struct {
	Enabled      bool   `json:"Enabled"`
	Router       string `json:"Router"`
	FloodPublish bool   `json:"FloodPublish"`
}

type RoutingConfig struct {
	Type           string `json:"Type"`
	ProtocolPrefix string `json:"ProtocolPrefix"`
}

type AppConfig struct {
	LogLevel     string `json:"LogLevel"`
	LogFile      string `json:"LogFile"`
	LogOutput    string `json:"LogOutput"`
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

	err = config.Pubsub.Validate()
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

func (config PubsubConfig) Validate() error {
	if config.Router != "gossipsub" && config.Router != "floodsub" {
		return fmt.Errorf("unknowned pubsub router")
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
	outputOptions := strings.Split(config.LogOutput, "+")
	for _, opt := range outputOptions {
		switch opt {
		case "stdout":
		case "stderr":
			continue
		case "file":
			if config.LogFile == "" {
				return fmt.Errorf("need specify a LogFile when LogOutput contained 'file'")
			}
			continue
		default:
			return fmt.Errorf("unknowned log output")
		}
	}
	if pathIsTerm(config.LogFile) {
		return fmt.Errorf("illegal log file")
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

	GC.Pubsub.Enabled = true
	if GC.Pubsub.Router == "" {
		GC.Pubsub.Router = "gossipsub"
	}

	if GC.Routing.Type == "" {
		GC.Routing.Type = "auto"
	}

	if GC.App.LogLevel == "" {
		GC.App.LogLevel = "info"
	}

	if GC.App.LogOutput == "" {
		GC.App.LogOutput = "stderr"
	}

	return GC, nil
}

func isTerm(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

func pathIsTerm(p string) bool {
	// !!!no!!! O_CREAT, if we fail - we fail
	f, err := os.OpenFile(p, os.O_WRONLY, 0)
	if f != nil {
		defer f.Close() // nolint:errcheck
	}
	return err == nil && isTerm(f)
}
