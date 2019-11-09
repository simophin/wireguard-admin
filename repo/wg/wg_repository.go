package wg

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"nz.cloudwalker/wireguard-webadmin/repo"
	"os/exec"
)

type wgRepository struct {
	repo.DefaultChangeNotificationHandler
	Client *wgctrl.Client
}

func runIpCommand(command string, args ...interface{}) error {
	argStrings := make([]string, 0, len(args)+1)
	argStrings = append(argStrings, command)

	for _, arg := range args {
		if v, ok := arg.(string); ok {
			argStrings = append(argStrings, v)
		} else {
			argStrings = append(argStrings, fmt.Sprint(arg))
		}
	}
	cmd := exec.Command("ip", argStrings...)
	return cmd.Run()
}

func (w wgRepository) listDeviceMap() (map[string]repo.DeviceInfo, error) {
	devices, err := w.Client.Devices()
	if err != nil {
		return nil, err
	}

	ret := make(map[string]repo.DeviceInfo, len(devices))
	for _, d := range devices {
		ret[d.PublicKey.String()] = repo.DeviceInfo{
			PrivateKey: d.PrivateKey.String(),
			ListenPort: uint16(d.ListenPort),
			Name:       d.Name,
		}
	}

	return ret, nil
}

func (w wgRepository) ListDevices() ([]repo.DeviceInfo, error) {
	devices, err := w.Client.Devices()
	if err != nil {
		return nil, err
	}

	var ret []repo.DeviceInfo
	for _, d := range devices {
		ret = append(ret, repo.DeviceInfo{
			PrivateKey: d.PrivateKey.String(),
			ListenPort: uint16(d.ListenPort),
			Name:       d.Name,
		})
	}

	return ret, nil
}

func (w wgRepository) UpdateDevices(devices []repo.DeviceInfo) error {

}

func (w wgRepository) RemoveDevices(pubKeys []string) error {
	panic("implement me")
}

func (w wgRepository) ReplaceAllDevices(devices []repo.DeviceInfo) error {
	panic("implement me")
}

func (w wgRepository) ListPeersByDevices(pubKeys []string, order repo.PeerOrder, offset uint, limit uint) (data []repo.PeerInfo, total uint, err error) {
	panic("implement me")
}

func (w wgRepository) ListPeersByKeys(devicePubKey string, pubKeys []string, order repo.PeerOrder, offset uint, limit uint) (data []repo.PeerInfo, total uint, err error) {
	panic("implement me")
}

func (w wgRepository) ListPeers(order repo.PeerOrder, offset uint, limit uint) (data []repo.PeerInfo, total uint, err error) {
	panic("implement me")
}

func (w wgRepository) RemovePeers(devicePubKey string, publicKeys []string) error {
	panic("implement me")
}

func (w wgRepository) UpdatePeers(devicePubKey string, peers []repo.PeerInfo) error {
	panic("implement me")
}

func (w wgRepository) ReplaceAllPeers(devicePubKey string, peers []repo.PeerInfo) error {
	panic("implement me")
}

func NewWgRepository() (repo.Repository, error) {
	if client, err := wgctrl.New(); err != nil {
		return nil, err
	} else {
		return &wgRepository{
			Client: client,
		}, nil
	}
}
