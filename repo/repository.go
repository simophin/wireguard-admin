package repo

import (
	"io"
	"net"
	"sync"
	"time"
)

type PeerInfo struct {
	PublicKey                   string
	PresharedKey                string
	Endpoint                    *net.UDPAddr
	PersistentKeepaliveInterval time.Duration
	AllowedIPs                  []net.IPNet
	DevicePublicKey             string
	LastHandshake               *time.Time

	Name string
}

type DeviceInfo struct {
	PrivateKey string
	PublicKey  string
	ListenPort uint16
	Name       string
}

type ChangeNotification interface {
	AddChangeNotification(channel chan<- interface{})
	RemoveChangeNotification(channel chan<- interface{})
}

type PeerOrder int

const (
	NameAsc           PeerOrder = 0
	NameDesc          PeerOrder = 1
	LastHandshakeAsc  PeerOrder = 2
	LastHandshakeDesc PeerOrder = 3
)

type Repository interface {
	ChangeNotification
	io.Closer

	ListDevices() ([]DeviceInfo, error)
	UpdateDevices(devices []DeviceInfo) error
	RemoveDevices(pubKeys []string) error
	ReplaceAllDevices(devices []DeviceInfo) error

	ListPeersByDevices(pubKeys []string, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error)
	ListPeersByKeys(pubKeys []string, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error)
	ListPeers(order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error)

	RemovePeers(publicKeys []string) error
	UpdatePeers(peers []PeerInfo) error
	ReplaceAllPeers(peers []PeerInfo)
}

type DefaultChangeNotificationHandler struct {
	listeners      map[chan<- interface{}]interface{}
	listenersMutex sync.Mutex
}

func (d DefaultChangeNotificationHandler) NotifyChange() {
	go func() {
		defer func() {
			recover()
		}()

		for c, _ := range d.listeners {
			c <- nil
		}
	}()
}

func (d *DefaultChangeNotificationHandler) AddChangeNotification(channel chan<- interface{}) {
	d.listenersMutex.Lock()
	defer d.listenersMutex.Unlock()
	if d.listeners != nil {
		delete(d.listeners, channel)
	}
}

func (d *DefaultChangeNotificationHandler) RemoveChangeNotification(channel chan<- interface{}) {
	d.listenersMutex.Lock()
	defer d.listenersMutex.Unlock()
	if d.listeners == nil {
		d.listeners = make(map[chan<- interface{}]interface{})
	}

	d.listeners[channel] = nil
}
