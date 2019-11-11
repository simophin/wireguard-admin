package persistent

import "nz.cloudwalker/wireguard-webadmin/wg"

type DeviceRepository interface {
	SaveDevices(devices []wg.DeviceConfig) error
	RemoveDevices(ids []string) error
	ListDevices() ([]wg.DeviceConfig, error)
}

type MetaId interface {
	primaryId() string
	secondaryId() string
}

type DeviceId string

func (d DeviceId) primaryId() string {
	return string(d)
}

func (d DeviceId) secondaryId() string {
	return ""
}

type PeerId struct {
	DeviceId string
	Id       string
}

func (p PeerId) primaryId() string {
	return p.DeviceId
}

func (p PeerId) secondaryId() string {
	return p.Id
}

type MetaRepository interface {
	GetNames(ids []MetaId) ([]string, error)
	SetNames(names map[MetaId]string) error
	RemoveNames(ids []MetaId) error
}
