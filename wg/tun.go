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
	var link netlink.Link
	if name, err := tunIf.Name(); err != nil {
		return err
	} else {
		if link, err = netlink.LinkByName(name); err != nil {
			return err
		}

		addresses, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		if err != nil {
			return err
		}

		for _, addr := range addresses {
			_ = netlink.AddrDel(link, &addr)
		}

		if config.Address != nil {
			addr := &netlink.Addr{
				IPNet: config.Address,
			}

			if err = netlink.AddrAdd(link, addr); err != nil {
				return err
			}
		}
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return err
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

			for _, ip := range p.AllowedIPs {
				if ones, _ := ip.Mask.Size(); ones > 0 {
					err := netlink.RouteAdd(&netlink.Route{
						LinkIndex: link.Attrs().Index,
						Dst:       &ip,
					})

					if err != nil {
						fmt.Printf("wg-tun: unable to add route for %v: %v\n", ip, err)
					}
				} else {
					fmt.Printf("wg-tun: catch all ip %v is unsupported\n", ip)
				}
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
		Device: Device{
			Id: deviceId,
		},
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
