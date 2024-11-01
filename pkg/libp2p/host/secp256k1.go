package host

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	ErrNotSecp256k1PubKey  = fmt.Errorf("not secp256k1 public key")
	ErrNotSecp256k1PrivKey = fmt.Errorf("not secp256k1 private key")
)

func Encrypt(ctx context.Context, peerid string, plaintext []byte) ([]byte, error) {
	peer, err := peer.Decode(peerid)
	if err != nil {
		return plaintext, err
	}
	pubKey, err := Hio.GetPublicKey(ctx, peer)
	if err != nil {
		return plaintext, err
	}

	pubK, ok := pubKey.(*crypto.Secp256k1PublicKey)
	if !ok {
		return plaintext, ErrNotSecp256k1PubKey
	}
	secp256k1PubKey := (*secp256k1.PublicKey)(pubK)

	privK, ok := Hio.PrivKey.(*crypto.Secp256k1PrivateKey)
	if !ok {
		return plaintext, ErrNotSecp256k1PrivKey
	}
	secp256k1PrivKey := (*secp256k1.PrivateKey)(privK)

	sharedKey := secp256k1.GenerateSharedSecret(secp256k1PrivKey, secp256k1PubKey)
	return gcmEncrypt(sharedKey, plaintext)
}

func Decrypt(pubKey, plaintext []byte) ([]byte, error) {
	if pubKey == nil || plaintext == nil {
		return plaintext, nil
	}
	cpub, err := crypto.UnmarshalPublicKey(pubKey)
	if err != nil {
		return plaintext, err
	}

	pubK, ok := cpub.(*crypto.Secp256k1PublicKey)
	if !ok {
		return plaintext, ErrNotSecp256k1PubKey
	}
	secp256k1PubKey := (*secp256k1.PublicKey)(pubK)

	privK, ok := Hio.PrivKey.(*crypto.Secp256k1PrivateKey)
	if !ok {
		return plaintext, ErrNotSecp256k1PrivKey
	}
	secp256k1PrivKey := (*secp256k1.PrivateKey)(privK)

	sharedKey := secp256k1.GenerateSharedSecret(secp256k1PrivKey, secp256k1PubKey)
	return gcmDecrypt(sharedKey, plaintext)
}

func gcmEncrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return plaintext, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext, err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return plaintext, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	return append(nonce, ciphertext...), nil
}

func gcmDecrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return ciphertext, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return ciphertext, err
	}
	if len(ciphertext) < aesgcm.NonceSize() {
		return ciphertext, fmt.Errorf("ciphertext too short")
	}
	nonce := ciphertext[:aesgcm.NonceSize()]
	ciphertext = ciphertext[aesgcm.NonceSize():]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return ciphertext, err
	}
	return plaintext, nil
}
