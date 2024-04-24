package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	ProtocolPrefix string = "/DeepBrainChain"
	PreSharedKey   string = "f504f536a912a8cf7d00adacee8ed20270c5040d961d7f3da4fccbcbec0ec48a"
	TopicName      string = "SuperImageAI"
)

func Init(mode string) error {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get program directory:", err)
		return err
	}

	keyPath := GetUniqueFile(cwd, "peer", "key")
	privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Secp256k1, -1)
	if err != nil {
		fmt.Println("Generate peer key: ", err)
		return err
	}
	if err := SavePeerKey(keyPath, privKey); err != nil {
		fmt.Println("Save peer key:", err)
		return err
	}
	fmt.Println("Generate peer key success at", keyPath)
	fmt.Println("Important notice: Please save this key file.",
		"You can use tools to retrieve the ID and private key in the future.")

	privkeyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		fmt.Println("Marshal Private Key err:", err)
		return err
	}
	fmt.Println("Encode private key:", crypto.ConfigEncodeKey(privkeyBytes))

	pubkeyBytes, err := crypto.MarshalPublicKey(pubKey)
	if err != nil {
		fmt.Println("Marshal Public Key err:", err)
		return err
	}
	fmt.Println("Encode public key:", crypto.ConfigEncodeKey(pubkeyBytes))

	id, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		fmt.Println("Transform Peer ID err:", err)
		return err
	}
	fmt.Println("Transform Peer ID:", id)

	tcpPort := 7001
	for !CheckPortAvailability(tcpPort) {
		tcpPort = tcpPort + 1000
	}
	httpPort := tcpPort + 1
	for !CheckPortAvailability(httpPort) {
		httpPort++
	}

	dataPath := GetUniqueFile(cwd, "datastore", "")
	if err := os.Mkdir(dataPath, 0755); err != nil {
		fmt.Println("Create datastore directory failed:", err)
		return err
	}
	fmt.Println("Create datastore directory at", dataPath)

	config := Config{
		Bootstrap: []string{},
		Addresses: []string{
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", tcpPort),
			fmt.Sprintf("/ip6/::/tcp/%d", tcpPort),
		},
		API: APIConfig{
			Addr: fmt.Sprintf("127.0.0.1:%d", httpPort),
		},
		Identity: IdentityConfig{
			PeerID:  id.String(),
			PrivKey: crypto.ConfigEncodeKey(privkeyBytes),
		},
		Swarm: SwarmConfig{
			ConnMgr: SwarmConnMgrConfig{
				Type:        "basic",
				GracePeriod: "20s",
				HighWater:   400,
				LowWater:    100,
			},
			DisableNatPortMap: false,
			RelayClient: SwarmRelayClientConfig{
				Enabled: false,
			},
			RelayService: SwarmRelayServiceConfig{
				Enabled: false,
			},
			EnableHolePunching:   false,
			EnableAutoNATService: true,
		},
		Pubsub: PubsubConfig{
			Enabled:      true,
			Router:       "gossipsub",
			FloodPublish: true,
		},
		Routing: RoutingConfig{
			Type:           "dhtclient",
			ProtocolPrefix: ProtocolPrefix,
		},
		App: AppConfig{
			LogLevel:     "info",
			LogFile:      filepath.Join(cwd, "host.log"),
			LogOutput:    "file",
			PreSharedKey: PreSharedKey,
			TopicName:    TopicName,
			Datastore:    dataPath,
		},
	}
	if mode == "server" {
		config.Swarm.DisableNatPortMap = true
		config.Swarm.RelayService.Enabled = true
		config.Pubsub.Router = "floodsub"
		config.Routing.Type = "dhtserver"
		config.App.LogOutput = "stderr+file"
	}
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Marshal pretty json:", err)
		return err
	}
	configPath := GetUniqueFile(cwd, mode, "json")
	if err := os.WriteFile(configPath, jsonData, 0600); err != nil {
		fmt.Println("Failed to save json:", err)
		return err
	}
	fmt.Println("Generate configuration success at", configPath)
	fmt.Printf("Run \"host -config %s\" command to start the program", configPath)
	return nil
}

func GetUniqueFile(dir string, name string, ext string) string {
	filename := name
	if ext != "" {
		filename = fmt.Sprintf("%s.%s", name, ext)
	}
	filePath := filepath.Join(dir, filename)
	_, err := os.Stat(filePath)

	counter := 1
	for !os.IsNotExist(err) {
		if ext != "" {
			filename = fmt.Sprintf("%s-%d.%s", name, counter, ext)
		} else {
			filename = fmt.Sprintf("%s-%d", name, counter)
		}
		filePath = filepath.Join(dir, filename)
		_, err = os.Stat(filePath)
		counter++
	}
	return filePath
}

func PeerKeyParse(peerKeyPath string) error {
	privKey, pubKey, err := LoadPeerKey(peerKeyPath)
	if err != nil {
		privKey, pubKey, err = crypto.GenerateKeyPair(crypto.Secp256k1, -1)
		if err != nil {
			fmt.Println("Generate peer key:", err)
			return err
		}
		err := SavePeerKey(peerKeyPath, privKey)
		if err != nil {
			fmt.Println("Save peer key:", err)
			return err
		}
		fmt.Println("Generate peer key success at", peerKeyPath)
	} else {
		fmt.Println("Load peer key success")
	}

	privkeyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		fmt.Println("Marshal Private Key err:", err)
		return err
	}
	fmt.Println("Encode private key:", crypto.ConfigEncodeKey(privkeyBytes))

	pubkeyBytes, err := crypto.MarshalPublicKey(pubKey)
	if err != nil {
		fmt.Println("Marshal Public Key err:", err)
		return err
	}
	fmt.Println("Encode public key:", crypto.ConfigEncodeKey(pubkeyBytes))

	id, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		fmt.Println("Transform Peer ID err:", err)
		return err
	}
	fmt.Println("Transform Peer ID:", id)
	return nil
}

func LoadPeerKey(filePath string) (crypto.PrivKey, crypto.PubKey, error) {
	privBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	priv, err := crypto.UnmarshalPrivateKey(privBytes)
	return priv, priv.GetPublic(), err
}

func SavePeerKey(filePath string, priv crypto.PrivKey) error {
	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, privBytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

func CheckPortAvailability(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}
