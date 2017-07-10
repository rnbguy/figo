package core

import (
	"crypto/sha1"
	"encoding/base64"
	"math/rand"
	"time"
)

func GetNick(size int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetHashes(names []string) []string {
	hashes := make([]string, len(names))
	for i, name := range names {
		hashes[i] = GetHash(name)
	}
	return hashes
}

func GetHash(name string) string {
	hasher := sha1.New()
	hasher.Write([]byte(name))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func ZeroconfHandler(zeroconfQuit chan bool) {
	return
}
