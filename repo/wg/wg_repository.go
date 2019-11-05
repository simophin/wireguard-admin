package wg

import (
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"nz.cloudwalker/wireguard-webadmin/repo"
)

type wgCtlRepository struct {
	repo.DefaultChangeNotificationHandler
	client wgctrl.Client
}

func toPeerInfo(device *wgtypes.Device, peer wgtypes.Peer, peerInfo *repo.PeerInfo) {
	peerInfo.PublicKey = peer.PublicKey.String()
	peerInfo.PersistentKeepaliveInterval = peer.PersistentKeepaliveInterval
	peerInfo.NetworkDeviceName = device.Name
	peerInfo.Endpoint = peer.Endpoint
	peerInfo.AllowedIPs = peer.AllowedIPs
	peerInfo.PresharedKey = peer.PresharedKey.String()
	peerInfo.TimeLastSeen = &peer.LastHandshakeTime
}

func (w wgCtlRepository) ListAllPeers(offset uint32, limit uint32) (peers []repo.PeerInfo, total uint32, err error) {
	var devices []*wgtypes.Device
	if devices, err = w.client.Devices(); err != nil {
		return
	}

	for _, dev := range devices {
		for _, p := range dev.Peers {
			var peerInfo repo.PeerInfo
			toPeerInfo(dev, p, &peerInfo)

		}
	}
}

func (w wgCtlRepository) GetPeers(publicKeys []string) ([]repo.PeerInfo, error) {
	panic("implement me")
}

func (w wgCtlRepository) RemovePeers(publicKeys []string) error {
	panic("implement me")
}

func (w wgCtlRepository) UpdatePeers(peers []repo.PeerInfo) error {
	panic("implement me")
}

func (w wgCtlRepository) Close() error {
	panic("implement me")
}

func NewWgCtlRepository(client wgctrl.Client) (repo.Repository, error) {
	return &wgCtlRepository{client: client}, nil
}
