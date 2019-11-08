package repo

import (
	"math"
	"sort"
	"sync"
)

type memDevice struct {
	Device DeviceInfo
	Peers  []*PeerInfo
}

type memRepository struct {
	DefaultChangeNotificationHandler

	Mutex sync.Mutex

	Devices map[string]*memDevice
}

func (m *memRepository) ListDevices() ([]DeviceInfo, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	devices := make([]DeviceInfo, 0, len(m.Devices))
	for _, d := range m.Devices {
		devices = append(devices, d.Device)
	}

	return devices, nil
}

func (m *memRepository) UpdateDevices(devices []DeviceInfo) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for _, d := range devices {
		if md, ok := m.Devices[d.PublicKey]; ok {
			md.Device = d
		} else {
			m.Devices[d.PublicKey] = &memDevice{Device: d}
		}
	}

	if len(devices) > 0 {
		m.NotifyChange()
	}

	return nil
}

func (m *memRepository) RemoveDevices(pubKeys []string) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for _, k := range pubKeys {
		delete(m.Devices, k)
	}

	if len(pubKeys) > 0 {
		m.NotifyChange()
	}
}

func (m *memRepository) ReplaceAllDevices(devices []DeviceInfo) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	newDevices := make(map[string]*memDevice, len(devices))
	for _, d := range devices {
		if old, ok := m.Devices[d.PublicKey]; ok {
			old.Device = d
			newDevices[d.PublicKey] = old
		} else {
			newDevices[d.PublicKey] = &memDevice{
				Device: d,
			}
		}
	}

	m.Devices = newDevices
	m.NotifyChange()
	return nil
}

func (m *memRepository) listPeersCommon(filter func(peer *PeerInfo) bool, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for _, d := range m.Devices {
		for _, p := range d.Peers {
			if filter == nil || filter(p) {
				data = append(data, *p)
				total++
			}
		}
	}

	sort.Slice(data, order.LessFunc(data))

	if offset >= total {
		data = make([]PeerInfo, 0)
		return
	}

	end := offset + limit
	if end >= total {
		end = total
	}

	data = data[offset:end]
	return
}

func (m *memRepository) ListPeersByDevices(pubKeys []string, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error) {
	keyMap := make(map[string]interface{})
	for _, k := range pubKeys {
		keyMap[k] = nil
	}

	return m.listPeersCommon(func(peer *PeerInfo) bool {
		_, ok := keyMap[peer.DevicePublicKey]
		return ok
	}, order, offset, limit)
}

func (m memRepository) ListPeersByKeys(pubKeys []string, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error) {
	keyMap := make(map[string]interface{})
	for _, k := range pubKeys {
		keyMap[k] = nil
	}

	return m.listPeersCommon(func(peer *PeerInfo) bool {
		_, ok := keyMap[peer.PublicKey]
		return ok
	}, order, offset, limit)
}

func (m memRepository) ListPeers(order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error) {
	return m.listPeersCommon(nil, order, offset, limit)
}

func (m memRepository) RemovePeers(publicKeys []string) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

}

func (m memRepository) UpdatePeers(peers []PeerInfo) error {
	panic("implement me")
}

func (m memRepository) ReplaceAllPeers(peers []PeerInfo) error {
	panic("implement me")
}

func NewMemRepository() Repository {
	return &memRepository{
		Devices: make(map[string]memDevice),
	}
}
