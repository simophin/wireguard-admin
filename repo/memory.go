package repo

import (
	"sort"
	"sync"
)

type memDevice struct {
	Device DeviceInfo
	Peers  map[PublicKey]*PeerInfo
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
		if md, ok := m.Devices[d.Name]; ok {
			md.Device = d
		} else {
			m.Devices[d.Name] = &memDevice{Device: d, Peers: make(map[PublicKey]*PeerInfo)}
		}
	}

	if len(devices) > 0 {
		m.NotifyChange()
	}

	return nil
}

func (m *memRepository) RemoveDevices(deviceNames []string) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for _, k := range deviceNames {
		delete(m.Devices, k)
	}

	if len(deviceNames) > 0 {
		m.NotifyChange()
	}

	return nil
}

func (m *memRepository) ReplaceAllDevices(devices []DeviceInfo) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	newDevices := make(map[string]*memDevice, len(devices))
	for _, d := range devices {
		if old, ok := m.Devices[d.Name]; ok {
			old.Device = d
			newDevices[d.Name] = old
		} else {
			newDevices[d.Name] = &memDevice{
				Device: d,
				Peers:  make(map[PublicKey]*PeerInfo),
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

func (m *memRepository) ListPeersByDevices(deviceNames []string, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error) {
	keyMap := make(map[string]interface{})
	for _, k := range deviceNames {
		keyMap[k] = nil
	}

	return m.listPeersCommon(func(peer *PeerInfo) bool {
		_, ok := keyMap[peer.DeviceName]
		return ok
	}, order, offset, limit)
}

func (m *memRepository) ListPeersByKeys(deviceName string, pubKeys []PublicKey, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error) {
	keyMap := make(map[PublicKey]interface{})
	for _, k := range pubKeys {
		keyMap[k] = nil
	}

	return m.listPeersCommon(func(peer *PeerInfo) bool {
		if peer.DeviceName != deviceName {
			return false
		}

		_, ok := keyMap[peer.PublicKey]
		return ok
	}, order, offset, limit)
}

func (m *memRepository) ListPeers(order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error) {
	return m.listPeersCommon(nil, order, offset, limit)
}

func (m *memRepository) RemovePeers(deviceName string, publicKeys []PublicKey) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if d, ok := m.Devices[deviceName]; ok {
		for _, k := range publicKeys {
			delete(d.Peers, k)
		}

		m.NotifyChange()
	}

	return nil
}

func (m *memRepository) UpdatePeers(deviceName string, peers []PeerInfo) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if d, ok := m.Devices[deviceName]; ok {
		for _, p := range peers {
			d.Peers[p.PublicKey] = &p
		}

		m.NotifyChange()
	}

	return nil
}

func (m *memRepository) ReplaceAllPeers(deviceName string, peers []PeerInfo) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if d, ok := m.Devices[deviceName]; ok {
		d.Peers = make(map[PublicKey]*PeerInfo)
		for _, p := range peers {
			d.Peers[p.PublicKey] = &p
		}
		m.NotifyChange()
	}

	return nil
}

func NewMemRepository() Repository {
	return &memRepository{
		Devices: make(map[string]*memDevice),
	}
}
