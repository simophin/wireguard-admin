package repo

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"time"
)

type PeerInfo struct {
	PublicKey                   wgtypes.Key
	PresharedKey                wgtypes.Key
	Endpoint                    *net.UDPAddr
	PersistentKeepaliveInterval time.Duration
	AllowedIPs                  []*net.IPNet
	NetworkDeviceName           string

	Name string
}

type Repository interface {
	ListAllPeers() (<-chan []PeerInfo, error)
	RemovePeer(publicKey wgtypes.Key) error
	UpdatePeers(peers []PeerInfo) error
}
