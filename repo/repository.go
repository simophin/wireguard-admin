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
	TimeLastSeen                *time.Time

	Name        string
	TimeCreated time.Time
}

type Repository interface {
	AddChangeNotification(channel chan<- interface{})
	RemoveChangeNotification(channel chan<- interface{})

	ListAllPeers(out *[]PeerInfo, offset uint32, limit uint32) (total uint32, err error)
	RemovePeer(publicKey wgtypes.Key) error
	UpdatePeers(peers []PeerInfo) error
	Close() error
}
