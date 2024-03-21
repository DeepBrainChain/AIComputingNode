package main

import (
	"flag"
	"log"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func main() {
	peerKeyPath := flag.String("peerkey", "", "the file path of peer key")
	flag.Parse()

	if *peerKeyPath == "" {
		log.Fatal("Please provide a filepath to save peer key")
	}

	privKey, pubKey, err := LoadPeerKey(*peerKeyPath)
	if err != nil {
		privKey, pubKey, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
		if err != nil {
			log.Fatalf("Generate peer key: %v", err)
		}
		log.Println("Generate peer key success")
		err := SavePeerKey(*peerKeyPath, privKey)
		if err != nil {
			log.Fatalf("Save peer key: %v", err)
		}
	} else {
		log.Println("Load peer key success")
	}

	privkeyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		log.Fatalf("Marshal Private Key err: %v", err)
	}
	log.Println("Encode private key:", crypto.ConfigEncodeKey(privkeyBytes))

	pubkeyBytes, err := crypto.MarshalPublicKey(pubKey)
	if err != nil {
		log.Fatalf("Marshal Public Key err: %v", err)
	}
	log.Println("Encode public key:", crypto.ConfigEncodeKey(pubkeyBytes))

	id, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		log.Fatalf("Transform Peer ID err: %v", err)
	} else {
		log.Println("Transform Peer ID:", id)
	}
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
