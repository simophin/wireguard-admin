package wg

import (
	"os"
	"sync"
)

type memClient struct {
	*sync.RWMutex

	DeviceMap map[string]*Device
}

func (m memClient) Up(deviceId string, config DeviceConfig) (Device, error) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.DeviceMap[deviceId]; ok {
		return Device{}, os.ErrExist
	} else {
		d := Device{Id: deviceId}
		d.UpdateFromConfig(config)
		m.DeviceMap[deviceId] = &d
		return d, nil
	}
}

func (m memClient) Down(deviceId string) error {
	panic("implement me")
}

func (m memClient) Configure(deviceId string, configurator func(config *DeviceConfig) error) error {
	panic("implement me")
}

func (m memClient) Devices() ([]Device, error) {
	panic("implement me")
}

func (m memClient) Device(id string) (Device, error) {
	panic("implement me")
}

func (m memClient) Close() error {
	panic("implement me")
}

func NewMemClient() (Client, error) {
	return memClient{
		RWMutex:   &sync.RWMutex{},
		DeviceMap: make(map[string]*Device),
	}, nil
}
