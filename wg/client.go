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
	PeerConfig
	LastHandshake *time.Time
}

type DeviceConfig struct {
	Name       string
	PrivateKey Key
	Peers      []PeerConfig
	ListenPort uint16
	Address    *net.IPNet
}

type Device struct {
	Id         string
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

func (p *Peer) UpdateFromConfig(c PeerConfig) {
	p.PeerConfig = c
}

func (d Device) GetPeersMap() map[Key]*Peer {
	m := make(map[Key]*Peer, len(d.Peers))
	for _, p := range d.Peers {
		m[p.PublicKey] = &p
	}
	return m
}

func (d *Device) UpdateFromConfig(c DeviceConfig) {
	d.Name = c.Name
	d.PrivateKey = c.PrivateKey
	oldPeers := d.GetPeersMap()
	newPeers := make([]Peer, 0, len(c.Peers))

	for _, pc := range c.Peers {
		if p, ok := oldPeers[pc.PublicKey]; ok {
			p.PeerConfig = pc
			newPeers = append(newPeers, *p)
		} else {
			newPeers = append(newPeers, Peer{PeerConfig: pc})
		}
	}

	d.Peers = newPeers
	d.Address = c.Address
	d.ListenPort = c.ListenPort
}

func (d Device) ToConfig() DeviceConfig {
	config := DeviceConfig{
		Name:       d.Name,
		PrivateKey: d.PrivateKey,
		Peers:      make([]PeerConfig, 0, len(d.Peers)),
		ListenPort: d.ListenPort,
		Address:    d.Address,
	}

	for _, p := range d.Peers {
		config.Peers = append(config.Peers, p.PeerConfig)
	}

	return config
}
