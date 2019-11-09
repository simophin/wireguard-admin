package repo

import (
	"database/sql/driver"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type key string

func (k key) Value() (driver.Value, error) {
	return k.String(), nil
}

func (k key) String() string {
	return string(k)
}

func (k key) EqualTo(other fmt.Stringer) bool {
	return k.String() == other.String()
}

func (k key) LessThan(other fmt.Stringer) bool {
	return k.String() < other.String()
}

func (k *key) Scan(src interface{}) error {
	if str, ok := src.(string); ok {
		if _, err := wgtypes.ParseKey(str); err != nil {
			return err
		} else {
			*(*string)(k) = str
			return nil
		}
	} else {
		return fmt.Errorf("key: unable to decode")
	}
}

func (k key) ToKey() (wgtypes.Key, error) {
	return wgtypes.ParseKey(string(k))
}

type PublicKey struct {
	key
}

type PrivateKey struct {
	key
}

type SymmetricKey struct {
	key
}

func NewPublicKey(pk wgtypes.Key) PublicKey {
	return PublicKey{key(pk.String())}
}

func NewSymmetricKey(pk wgtypes.Key) SymmetricKey {
	return SymmetricKey{key(pk.String())}
}

func NewPrivateKey(pk wgtypes.Key) PrivateKey {
	return PrivateKey{key(pk.String())}
}

func (k PrivateKey) ToPublicKey() PublicKey {
	if k, err := k.ToKey(); err != nil {
		panic(err)
	} else {
		return NewPublicKey(k)
	}
}
