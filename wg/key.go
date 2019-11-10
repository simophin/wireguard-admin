package wg

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/curve25519"
)

const keySize = 32

type Key [keySize]byte

func (k Key) String() string {
	return hex.EncodeToString(k[:])
}

func (k *Key) ToPublicKey() Key {
	var dst [32]byte
	curve25519.ScalarBaseMult(&dst, (*[32]byte)(k))
	return dst
}

func (k Key) IsZero() bool {
	for i := 0; i < keySize; i++ {
		if k[i] != 0 {
			return false
		}
	}

	return true
}

func NewKeyFromString(v string) (Key, error) {
	var k Key
	if n, err := hex.Decode(k[:], []byte(v)); err != nil {
		return k, err
	} else if n != keySize {
		return k, fmt.Errorf("key: decode key size = %v, expecting %v", n, keySize)
	}
	return k, nil
}
