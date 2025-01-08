package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"AIComputingNode/pkg/types"

	"github.com/mattn/go-isatty"
	"github.com/multiformats/go-multiaddr"
)

var GC *Config

type Config struct {
	Bootstrap  []string                `json:"Bootstrap"`
	Addresses  []string                `json:"Addresses"`
	API        APIConfig               `Json:"API"`
	Identity   IdentityConfig          `json:"Identity"`
	Swarm      SwarmConfig             `json:"Swarm"`
	Pubsub     PubsubConfig            `json:"Pubsub"`
	Routing    RoutingConfig           `json:"Routing"`
	App        AppConfig               `json:"App"`
	AIProjects []types.AIProjectConfig `json:"AIProjects"`
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
	DialTimeout          string                  `json:"DialTimeout"`
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
	LogLevel     string            `json:"LogLevel"`
	LogFile      string            `json:"LogFile"`
	LogOutput    string            `json:"LogOutput"`
	PreSharedKey string            `json:"PreSharedKey"`
	TopicName    string            `json:"TopicName"`
	Datastore    string            `json:"Datastore"`
	AutoUpgrade  AutoUpgradeConfig `json:"AutoUpgrade"`
	// peers collect config
	PeersCollect AppPeersCollectConfig `json:"PeersCollect"`
}

type AutoUpgradeConfig struct {
	Enabled      bool   `json:"Enabled"`
	TimeInterval string `json:"TimeInterval"`
}

type AppPeersCollectConfig struct {
	Enabled           bool   `json:"Enabled"`
	HeartbeatInterval string `json:"HeartbeatInterval"`
	ClientProject     string `json:"ClientProject"`
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
	peer, err := PeerIDFromPrivKeyString(config.PrivKey)
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
	if config.Datastore == "" {
		return fmt.Errorf("datastore can not be empty")
	}
	if s, err := os.Stat(config.Datastore); err != nil || !s.IsDir() {
		return fmt.Errorf("datastore must be a folder that already exists")
	}
	if err := config.AutoUpgrade.Validate(); err != nil {
		return err
	}
	if err := config.PeersCollect.Validate(); err != nil {
		return err
	}
	return nil
}

func (config AutoUpgradeConfig) Validate() error {
	if _, err := time.ParseDuration(config.TimeInterval); err != nil {
		return err
	}
	return nil
}

func (config AppPeersCollectConfig) Validate() error {
	if _, err := time.ParseDuration(config.HeartbeatInterval); err != nil {
		return err
	}
	return nil
}

// func (config Config) GetModelAPI(projectName, modelName, cid string) (*types.AIModelConfig, error) {
// 	mi := &types.AIModelConfig{}
// 	if projectName == "" || modelName == "" {
// 		return mi, fmt.Errorf("empty project or model")
// 	}
// 	for _, project := range config.AIProjects {
// 		if project.Project == projectName {
// 			for _, model := range project.Models {
// 				if model.Model == modelName && (cid == "" || model.CID == cid) {
// 					*mi = model
// 					break
// 				}
// 			}
// 		}
// 	}
// 	if mi.API == "" {
// 		return mi, fmt.Errorf("can not find in registered project and model")
// 	}
// 	return mi, nil
// }

func (config Config) SaveConfig(configPath string) error {
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		// fmt.Println("Marshal pretty json:", err)
		return err
	}
	if err := os.WriteFile(configPath, jsonData, 0600); err != nil {
		// fmt.Println("Failed to save json:", err)
		return err
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

	if GC.Swarm.ConnMgr.GracePeriod == "" {
		GC.Swarm.ConnMgr.GracePeriod = "20s"
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

	if GC.App.AutoUpgrade.TimeInterval == "" {
		GC.App.AutoUpgrade.TimeInterval = "1h"
	}

	if GC.App.PeersCollect.HeartbeatInterval == "" {
		GC.App.PeersCollect.HeartbeatInterval = "180s"
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
