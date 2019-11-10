package wg

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"os"
	"sync"
)

type tunDevice struct {
	id         string
	device     *device.Device
	name       string
	privateKey Key
	listenPort uint16
	peers      map[Key]*tunPeer

	tunIf tun.Device
}

type tunPeer struct {
	PeerConfig
	device *tunDevice
}

type tunClient struct {
	mutex   sync.Mutex
	devices map[string]*tunDevice

	tunNameSeq int
}

func (p tunPeer) toPeer(dev *device.Device) Peer {
	return Peer{
		PublicKey:           p.PublicKey,
		Endpoint:            p.Endpoint,
		AllowedIPs:          p.AllowedIPs,
		PersistentKeepAlive: p.PersistentKeepAlive,
		LastHandshake:       nil,
	}
}

func (device tunDevice) toDevice() Device {
	ret := Device{
		Name:       device.name,
		PrivateKey: device.privateKey,
	}

	ret.Peers = make([]Peer, 0, len(device.peers))
	for _, p := range device.peers {
		ret.Peers = append(ret.Peers, p.toPeer(device.device))
	}

	return ret
}

func (k Key) ToNoisePrivateKey() device.NoisePrivateKey {
	return device.NoisePrivateKey(k)
}

func (k Key) ToNoisePublicKey() device.NoisePublicKey {
	return device.NoisePublicKey(k)
}

func (k Key) ToNoiseSymmetricKey() device.NoiseSymmetricKey {
	return device.NoiseSymmetricKey(k)
}

func configureDevice(tunIf tun.Device, dev *device.Device, config DeviceConfig) error {
	if name, err := tunIf.Name(); err != nil {
		return err
	} else {
		link, err := netlink.LinkByName(name)
		if err != nil {
			return err
		}

		var newAddr *netlink.Addr
		if config.Address != nil {
			if newAddr, err = netlink.ParseAddr(config.Address.String()); err != nil {
				return err
			}
		}

		addresses, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		if err != nil {
			return err
		}

		for _, addr := range addresses {
			_ = netlink.AddrDel(link, &addr)
		}

		if newAddr != nil {
			if err = netlink.AddrAdd(link, newAddr); err != nil {
				return err
			}
		}
	}

	if err := dev.SetPrivateKey(config.PrivateKey.ToNoisePrivateKey()); err != nil {
		return err
	}

	if err := dev.SetBindPort(config.ListenPort); err != nil {
		return err
	}

	dev.RemoveAllPeers()
	for _, p := range config.Peers {
		if peer, err := dev.NewPeer(p.PublicKey.ToNoisePublicKey()); err != nil {
			return err
		} else {
			err := peer.Configure(device.PeerConfig{
				PreSharedKey:        p.PreSharedKey.ToNoiseSymmetricKey(),
				Endpoint:            p.Endpoint,
				AllowedIPs:          p.AllowedIPs,
				PersistentKeepalive: p.PersistentKeepAlive,
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *tunClient) Up(deviceId string, config DeviceConfig) (Device, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	var ret Device

	if _, ok := t.devices[deviceId]; ok {
		return ret, os.ErrExist
	}

	tunIf, err := tun.CreateTUN(fmt.Sprint("utun", t.tunNameSeq), device.DefaultMTU)

	if err != nil {
		return ret, err
	}

	wgDevice := device.NewDevice(tunIf, device.NewLogger(device.LogLevelInfo, "wg-backend: "))

	if err := configureDevice(tunIf, wgDevice, config); err != nil {
		wgDevice.Close()
		_ = tunIf.Close()
		return ret, err
	}
	
	wgDevice.Up()

	t.tunNameSeq++

	dev := tunDevice{
		id:         deviceId,
		device:     wgDevice,
		name:       config.Name,
		privateKey: config.PrivateKey,
		peers:      make(map[Key]*tunPeer, len(config.Peers)),
		listenPort: config.ListenPort,
		tunIf:      tunIf,
	}

	for _, pc := range config.Peers {
		dev.peers[pc.PublicKey] = &tunPeer{
			PeerConfig: PeerConfig{
				PublicKey:           pc.PublicKey,
				PreSharedKey:        pc.PreSharedKey,
				Endpoint:            pc.Endpoint,
				AllowedIPs:          pc.AllowedIPs,
				PersistentKeepAlive: pc.PersistentKeepAlive,
			},
			device: &dev,
		}
	}

	t.devices[deviceId] = &dev
	return ret, nil
}

func (t tunClient) Down(deviceId string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if d, ok := t.devices[deviceId]; !ok {
		return os.ErrNotExist
	} else {
		d.device.Down()
		d.device.Close()
		err := d.tunIf.Close()
		delete(t.devices, deviceId)
		return err
	}
}

func (t *tunClient) Configure(deviceId string, configurator func(config *DeviceConfig) error) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	d, ok := t.devices[deviceId]
	if !ok {
		return os.ErrNotExist
	}

	config := DeviceConfig{
		Name:       d.name,
		PrivateKey: d.privateKey,
		Peers:      make([]PeerConfig, 0, len(d.peers)),
		ListenPort: d.peers,
		Address:    nil,
	}
	if err := configurator(&config); err != nil {
		return err
	}
}

func (t *tunClient) Devices() (devices []Device, err error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, d := range t.devices {
		devices = append(devices, d)
	}

	return
}

func (t tunClient) Device(id string) (Device, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if d, ok := t.devices[id]; ok {
		copied := *d
		return copied, nil
	}

	return nil, os.ErrNotExist
}

func (t tunClient) Close() error {
	panic("implement me")
}

func NewTunClient() (Client, error) {
	return &tunClient{devices: make(map[string]*tunDevice)}, nil
}
