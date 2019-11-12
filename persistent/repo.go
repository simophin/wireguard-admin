package persistent

import "nz.cloudwalker/wireguard-webadmin/wg"

type MetaKey string

const (
	MetaKeyName MetaKey = "name"
)

type DeviceId string
type PeerId struct {
	DeviceId  DeviceId
	PublicKey wg.Key
}

type Repository interface {
	SaveDevices(devices []wg.Device) error
	ListDevices() ([]wg.Device, error)
	RemoveDevices(ids []DeviceId) error

	SetDeviceMeta(deviceId DeviceId, key MetaKey, value string) error
	GetDeviceMeta(key MetaKey) (map[DeviceId]string, error)
	RemoveDeviceMeta(deviceId DeviceId, key MetaKey) error

	SetPeerMeta(peerId PeerId, key MetaKey, value string) error
	GetPeerMeta(key MetaKey) (map[PeerId]string, error)
	RemovePeerMeta(id PeerId, key MetaKey) error
}
