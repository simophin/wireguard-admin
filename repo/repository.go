package repo

import (
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
	NetworkDeviceName           string
	TimeLastSeen                *time.Time

	Name        string
	TimeCreated time.Time
}

type ChangeNotification interface {
	AddChangeNotification(channel chan<- interface{})
	RemoveChangeNotification(channel chan<- interface{})
}

type Repository interface {
	ChangeNotification

	ListAllPeers(offset uint32, limit uint32) (peers []PeerInfo, total uint32, err error)
	GetPeers(publicKeys []string) ([]PeerInfo, error)
	RemovePeers(publicKeys []string) error
	UpdatePeers(peers []PeerInfo) error
	Close() error
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
