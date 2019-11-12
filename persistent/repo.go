package persistent

import (
	"net"
	"nz.cloudwalker/wireguard-webadmin/wg"
)

type MetaKey string

type DeviceId string

type PeerId struct {
	DeviceId DeviceId
	Id       string
}

type Device struct {
	Id         DeviceId
	Name       string
	PrivateKey wg.Key
	Peers      []Peer
	ListenPort uint16
	Address    *net.IPNet
}

type Peer struct {
	wg.PeerConfig
	PeerId
}

const (
	MetaKeyName MetaKey = "name"
)

type Repository interface {
	SaveDevices(devices []Device) error
	RemoveDevices(ids []DeviceId) error
	ListDevices() ([]Device, error)

	SavePeers(peers []Peer) error
	RemovePeers(ids []PeerId) error
	ListPeersByDevice(deviceId DeviceId) ([]Peer, error)
	ListPeers(ids []PeerId) ([]Peer, error)

	GetDeviceMeta(ids []DeviceId, key MetaKey) (map[DeviceId]string, error)
	SaveDeviceMeta(id DeviceId, data map[MetaKey]string) error
	RemoveDeviceMeta(id DeviceId, keys []MetaKey) error

	GetPeerMeta(ids []PeerId, key MetaKey) (map[PeerId]string, error)
	SavePeerMeta(id PeerId, data map[MetaKey]string) error
	RemovePeerMeta(id PeerId, keys []MetaKey)
}
