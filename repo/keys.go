package repo

import (
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/curve25519"
)

const (
	noisePublicKeySize  = 32
	noisePrivateKeySize = 32
)

func GetPublicKey(privateKey string) (string, error) {
	var rawPrivateKey [noisePrivateKeySize]byte
	if n, err := hex.Decode(rawPrivateKey[:], []byte(privateKey)); err != nil {
		return "", err
	} else if n != noisePrivateKeySize {
		return "", errors.New("invalid private key")
	}

	var out [noisePublicKeySize]byte
	curve25519.ScalarBaseMult(&out, &rawPrivateKey)
	return hex.EncodeToString(out[:]), nil
}
