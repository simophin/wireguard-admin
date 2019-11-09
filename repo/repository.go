package repo

import (
	"errors"
	"io"
	"net"
	"nz.cloudwalker/wireguard-webadmin/utils"
	"sync"
	"time"
)

type PeerInfo struct {
	PublicKey                   PublicKey
	PreSharedKey                SymmetricKey
	Endpoint                    *net.UDPAddr
	PersistentKeepaliveInterval time.Duration
	AllowedIPs                  []net.IPNet
	DevicePublicKey             PublicKey
	LastHandshake               int64

	Name string
}

type DeviceInfo struct {
	PrivateKey PrivateKey
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
			lhs, rhs := peers[i], peers[j]
			if lhs.Name == rhs.Name {
				return lhs.PublicKey.LessThan(rhs.PublicKey)
			}
			return lhs.Name < rhs.Name
		}
	case OrderLastHandshakeAsc:
		return func(i, j int) bool {
			lhs, rhs := peers[i], peers[j]
			if lhs.LastHandshake == rhs.LastHandshake {
				return lhs.PublicKey.LessThan(rhs.PublicKey)
			}
			return lhs.LastHandshake < rhs.LastHandshake
		}

	case OrderNameDesc:
		return utils.ReverseLessFunc(OrderNameAsc.LessFunc(peers))
	case OrderLastHandshakeDesc:
		return utils.ReverseLessFunc(OrderLastHandshakeAsc.LessFunc(peers))
	default:
		panic(InvalidPeerOrder)
	}
}

type Repository interface {
	ChangeNotification

	ListDevices() ([]DeviceInfo, error)
	UpdateDevices(devices []DeviceInfo) error
	RemoveDevices(pubKeys []PublicKey) error
	ReplaceAllDevices(devices []DeviceInfo) error

	ListPeersByDevices(pubKeys []PublicKey, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error)
	ListPeersByKeys(devicePubKey PublicKey, pubKeys []PublicKey, order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error)
	ListPeers(order PeerOrder, offset uint, limit uint) (data []PeerInfo, total uint, err error)

	RemovePeers(devicePubKey PublicKey, publicKeys []PublicKey) error
	UpdatePeers(devicePubKey PublicKey, peers []PeerInfo) error
	ReplaceAllPeers(devicePubKey PublicKey, peers []PeerInfo) error
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
