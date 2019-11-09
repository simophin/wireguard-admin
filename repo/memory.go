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

	Devices map[PublicKey]*memDevice
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
		pubKey := d.PrivateKey.ToPublicKey()
		if md, ok := m.Devices[pubKey]; ok {
			md.Device = d
		} else {
			m.Devices[pubKey] = &memDevice{Device: d, Peers: make(map[PublicKey]*PeerInfo)}
		}
	}

	if len(devices) > 0 {
		m.NotifyChange()
	}

	return nil
}

func (m *memRepository) RemoveDevices(pubKeys []PublicKey) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	for _, k := range pubKeys {
		delete(m.Devices, k)
	}

	if len(pubKeys) > 0 {
		m.NotifyChange()
	}

	return nil
}

func (m *memRepository) ReplaceAllDevices(devices []DeviceInfo) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	newDevices := make(map[PublicKey]*memDevice, len(devices))
	for _, d := range devices {
		key := d.PrivateKey.ToPublicKey()
		if old, ok := m.Devices[key]; ok {
			old.Device = d
			newDevices[key] = old
		} else {
			newDevices[key] = &memDevice{
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

func (m *memRepository) ListPeersByDevices(pubKeys []PublicKey, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error) {
	keyMap := make(map[PublicKey]interface{})
	for _, k := range pubKeys {
		keyMap[k] = nil
	}

	return m.listPeersCommon(func(peer *PeerInfo) bool {
		_, ok := keyMap[peer.DevicePublicKey]
		return ok
	}, order, offset, limit)
}

func (m *memRepository) ListPeersByKeys(devicePubKey PublicKey, pubKeys []PublicKey, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error) {
	keyMap := make(map[PublicKey]interface{})
	for _, k := range pubKeys {
		keyMap[k] = nil
	}

	return m.listPeersCommon(func(peer *PeerInfo) bool {
		if !peer.DevicePublicKey.EqualTo(devicePubKey) {
			return false
		}

		_, ok := keyMap[peer.PublicKey]
		return ok
	}, order, offset, limit)
}

func (m *memRepository) ListPeers(order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error) {
	return m.listPeersCommon(nil, order, offset, limit)
}

func (m *memRepository) RemovePeers(devicePubKey PublicKey, publicKeys []PublicKey) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if d, ok := m.Devices[devicePubKey]; ok {
		for _, k := range publicKeys {
			delete(d.Peers, k)
		}

		m.NotifyChange()
	}

	return nil
}

func (m *memRepository) UpdatePeers(devicePubKey PublicKey, peers []PeerInfo) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if d, ok := m.Devices[devicePubKey]; ok {
		for _, p := range peers {
			d.Peers[p.PublicKey] = &p
		}

		m.NotifyChange()
	}

	return nil
}

func (m *memRepository) ReplaceAllPeers(devicePubKey PublicKey, peers []PeerInfo) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if d, ok := m.Devices[devicePubKey]; ok {
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
		Devices: make(map[PublicKey]*memDevice),
	}
}
