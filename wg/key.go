package wg

import (
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/curve25519"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const keySize = 32

type Key [keySize]byte

func (k *Key) Scan(src interface{}) error {
	if text, ok := src.(string); ok {
		var err error
		if *k, err = NewKeyFromString(text); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return fmt.Errorf("key: expecting a string but %v is given", src)
	}
}

func (k Key) Value() (driver.Value, error) {
	return k.String(), nil
}

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

func NewRandom() (Key, error) {
	if k, err := wgtypes.GenerateKey(); err != nil {
		return Key{}, err
	} else {
		return Key(k), nil
	}
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
