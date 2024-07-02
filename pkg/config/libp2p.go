package config

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func PrivKeyFromString(pk string) (crypto.PrivKey, error) {
	privKeyBytes, err := crypto.ConfigDecodeKey(pk)
	if err != nil {
		return nil, err
	}
	privKey, err := crypto.UnmarshalPrivateKey(privKeyBytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func PeerIDFromPrivKeyString(pk string) (peer.ID, error) {
	privKey, err := PrivKeyFromString(pk)
	if err != nil {
		return "", err
	}
	peer, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return "", err
	}
	return peer, nil
}
