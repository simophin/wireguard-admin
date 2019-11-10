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
	Device

	Raw   *device.Device
	TunIf tun.Device
}

type tunClient struct {
	sync.RWMutex

	DeviceMap  map[string]*tunDevice
	TunNameSeq uint
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

func (t *tunDevice) Close() error {
	t.Raw.Down()
	t.Raw.Close()
	return t.TunIf.Close()
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

func (t *tunClient) Up(deviceId string, config DeviceConfig) (ret Device, err error) {
	t.Lock()
	defer t.Unlock()

	if _, ok := t.DeviceMap[deviceId]; ok {
		err = os.ErrExist
		return
	}

	tunIf, err := tun.CreateTUN(fmt.Sprint("utun", t.TunNameSeq), device.DefaultMTU)

	if err != nil {
		return
	}

	wgDevice := device.NewDevice(tunIf, device.NewLogger(device.LogLevelInfo, "wg-backend: "))

	if err = configureDevice(tunIf, wgDevice, config); err != nil {
		wgDevice.Close()
		_ = tunIf.Close()
		return ret, err
	}

	wgDevice.Up()

	t.TunNameSeq++

	td := tunDevice{
		Raw:   wgDevice,
		TunIf: tunIf,
	}

	td.Device.UpdateFromConfig(config)
	t.DeviceMap[deviceId] = &td
	ret = td.Device
	return
}

func (t tunClient) Down(deviceId string) error {
	t.Lock()
	defer t.Unlock()

	if d, ok := t.DeviceMap[deviceId]; !ok {
		return os.ErrNotExist
	} else {
		err := d.Close()
		delete(t.DeviceMap, deviceId)
		return err
	}
}

func (t *tunClient) Configure(deviceId string, configurator func(config *DeviceConfig) error) error {
	t.Lock()
	defer t.Unlock()

	d, ok := t.DeviceMap[deviceId]
	if !ok {
		return os.ErrNotExist
	}

	config := d.ToConfig()

	if err := configurator(&config); err != nil {
		return err
	}

	if err := configureDevice(d.TunIf, d.Raw, config); err != nil {
		return err
	}

	d.UpdateFromConfig(config)
	return nil
}

func (t *tunClient) Devices() (devices []Device, err error) {
	t.RLock()
	defer t.RUnlock()

	for _, d := range t.DeviceMap {
		devices = append(devices, d.Device)
	}

	return
}

func (t *tunClient) Device(id string) (Device, error) {
	t.RLock()
	defer t.RUnlock()

	if d, ok := t.DeviceMap[id]; ok {
		return d.Device, nil
	} else {
		return Device{}, nil
	}
}

func (t *tunClient) Close() error {
	t.Lock()
	defer t.Unlock()

	for _, d := range t.DeviceMap {
		_ = d.Close()
	}

	return nil
}

func NewTunClient() (Client, error) {
	return &tunClient{DeviceMap: make(map[string]*tunDevice)}, nil
}
