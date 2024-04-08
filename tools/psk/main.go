package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

func main() {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalf("Generate Pre-Shared key: %v", err)
	}
	log.Println("Generate Pre-Shared key:", hex.EncodeToString(key))
}
