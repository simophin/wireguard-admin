package repo

import (
	"net"
	"time"
)

type PeerInfo struct {
	PublicKey                   string
	PresharedKey                string
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

	ListAllPeers(offset uint32, limit uint32) (peers []PeerInfo, total uint32, err error)
	GetPeers(publicKeys []string) ([]PeerInfo, error)
	RemovePeers(publicKeys []string) error
	UpdatePeers(peers []PeerInfo) error
	Close() error
}
