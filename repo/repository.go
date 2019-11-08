package repo

import (
	"errors"
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
	LastHandshake               int64

	Name string
}

type DeviceInfo struct {
	PrivateKey string
	PublicKey  string
	ListenPort uint16
	Name       string
}

type ChangeNotification interface {
	io.Closer
	AddChangeNotification(channel chan<- interface{})
	RemoveChangeNotification(channel chan<- interface{})
}

type PeerOrder int

const (
	OrderNameAsc           PeerOrder = 0
	OrderNameDesc          PeerOrder = 1
	OrderLastHandshakeAsc  PeerOrder = 2
	OrderLastHandshakeDesc PeerOrder = 3
)

var (
	InvalidPeerOrder = errors.New("invalid peer order")
)

func (o PeerOrder) LessFunc(peers []PeerInfo) func(lh, rh int) bool {
	switch o {
	case OrderNameAsc:
		return func(i, j int) bool {
			lh := peers[i]
			rh := peers[j]
			if lh.Name == rh.Name {
				return lh.PublicKey < rh.PublicKey
			}
			return lh.Name < rh.Name
		}
	case OrderLastHandshakeAsc:
		return func(i, j int) bool {
			lh := peers[i]
			rh := peers[j]
			if lh.LastHandshake == rh.LastHandshake {
				return lh.PublicKey < rh.PublicKey
			}
			return lh.LastHandshake < rh.LastHandshake
		}

	case OrderNameDesc:
		{
			f := OrderNameAsc.LessFunc(peers)
			return func(lh, rh int) bool {
				return !f(lh, rh)
			}
		}
	case OrderLastHandshakeDesc:
		{
			f := OrderLastHandshakeAsc.LessFunc(peers)
			return func(lh, rh int) bool {
				return !f(lh, rh)
			}
		}
	default:
		{
			panic(errors.New("repository: unable to create less func: unknown order type"))
		}
	}
}

type Repository interface {
	ChangeNotification

	ListDevices() ([]DeviceInfo, error)
	UpdateDevices(devices []DeviceInfo) error
	RemoveDevices(pubKeys []string) error
	ReplaceAllDevices(devices []DeviceInfo) error

	ListPeersByDevices(pubKeys []string, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error)
	ListPeersByKeys(pubKeys []string, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error)
	ListPeers(order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error)

	RemovePeers(publicKeys []string) error
	UpdatePeers(peers []PeerInfo) error
	ReplaceAllPeers(peers []PeerInfo) error
}

type DefaultChangeNotificationHandler struct {
	listeners      map[chan<- interface{}]interface{}
	listenersMutex sync.Mutex
}

func (d *DefaultChangeNotificationHandler) Close() error {
	d.listenersMutex.Lock()
	defer d.listenersMutex.Unlock()

	for c, _ := range d.listeners {
		close(c)
	}

	d.listeners = make(map[chan<- interface{}]interface{})
	return nil
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
