package wg

import (
	"net"
	"time"
)

type PeerConfig struct {
	PublicKey           Key
	PreSharedKey        Key
	Endpoint            *net.UDPAddr
	AllowedIPs          []net.IPNet
	PersistentKeepAlive time.Duration
}

type Peer struct {
	PublicKey           Key
	PreSharedKey        Key
	Endpoint            *net.UDPAddr
	AllowedIPs          []net.IPNet
	PersistentKeepAlive time.Duration
	LastHandshake       *time.Time
}

type DeviceConfig struct {
	Name       string
	PrivateKey Key
	Peers      []PeerConfig
	ListenPort uint16
	Address    *net.IPNet
}

type Device struct {
	Name       string
	PrivateKey Key
	Peers      []Peer
	ListenPort uint16
	Address    *net.IPNet
}

type Client interface {
	Up(deviceId string, config DeviceConfig) (Device, error)
	Down(deviceId string) error
	Configure(deviceId string, configurator func(config *DeviceConfig) error) error
	Devices() ([]Device, error)
	Device(id string) (Device, error)

	Close() error
}
