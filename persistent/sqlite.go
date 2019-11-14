package persistent

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"net"
	"nz.cloudwalker/wireguard-webadmin/utils"
	"nz.cloudwalker/wireguard-webadmin/wg"
	"strings"
	"time"
)

type device struct {
	Id         string `sql:"id"`
	Name       string `sql:"name"`
	PrivateKey wg.Key `sql:"private_key"`
	ListenPort uint16 `sql:"listen_port"`
	Address    string `sql:"address"`
}

type peer struct {
	DeviceId            string        `sql:"device_id"`
	PublicKey           wg.Key        `sql:"public_key"`
	PreSharedKey        wg.Key        `sql:"pre_shared_key"`
	Endpoint            string        `sql:"endpoint"`
	AllowedIPs          string        `sql:"allowed_ips"`
	PersistentKeepAlive time.Duration `sql:"persistent_keep_alive"`
}

var tableMigrations = [][]string{
	{
		`CREATE TABLE options(
				name TEXT NOT NULL PRIMARY KEY,
				value TEXT
			)`,

		`CREATE TABLE devices(
				id TEXT NOT NULL PRIMARY KEY,
				name TEXT NOT NULL,
				private_key TEXT NOT NULL,
				listen_port INTEGER NOT NULL DEFAULT 0,
				address TEXT NOT NULL
			)`,

		`CREATE TABLE peers(
				device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
				public_key TEXT NOT NULL,
				pre_shared_key TEXT NOT NULL,
				endpoint TEXT NOT NULL,
				allowed_ips TEXT NOT NULL,
				persistent_keep_alive INTEGER NOT NULL,
				PRIMARY KEY (device_id, public_key) ON CONFLICT REPLACE
			)`,

		`CREATE INDEX peers_device_id ON peers(device_id)`,

		`CREATE TABLE device_meta(
				device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
				name TEXT NOT NULL,
				value TEXT NOT NULL,
				PRIMARY KEY (device_id, name)
			)`,

		`CREATE TABLE peer_meta(
				device_id TEXT NOT NULL,
				public_key TEXT NOT NULL,
				name TEXT NOT NULL,
				value TEXT NOT NULL,
				PRIMARY KEY (device_id, public_key, name),
				FOREIGN KEY (device_id, public_key)
					REFERENCES peers (device_id, public_key)
					ON DELETE CASCADE
			)`,
	},
}

const (
	insertDeviceSql = `INSERT OR REPLACE INTO devices(id, name, private_key, listen_port, address)
						VALUES (:id, :private_key, :name, :listen_port, :address)`

	insertPeerSql = `INSERT OR REPLACE INTO peers(device_id, public_key, pre_shared_key, endpoint, allowed_ips, persistent_keep_alive)
					  VALUES (:device_id, :public_key, :pre_shared_key, :endpoint, :allowed_ips, :persistent_keep_alive)`
)

const (
	optionSchemaVersion = "schema_version"
)

func (d *device) UpdateFrom(dev wg.Device) {
	d.Id = dev.Id
	if dev.Address != nil {
		d.Address = dev.Address.String()
	} else {
		d.Address = ""
	}
	d.PrivateKey = dev.PrivateKey
	d.ListenPort = dev.ListenPort
	d.Name = dev.Name
}

func (p *peer) UpdateFrom(d wg.Device, o wg.Peer) {
	p.PublicKey = o.PublicKey
	p.PersistentKeepAlive = o.PersistentKeepAlive
	p.DeviceId = d.Id

	ips := make([]string, 0, len(o.AllowedIPs))
	for _, ip := range o.AllowedIPs {
		ips = append(ips, ip.String())
	}
	p.AllowedIPs = strings.Join(ips, ",")
	if o.Endpoint != nil {
		p.Endpoint = o.Endpoint.String()
	} else {
		p.Endpoint = ""
	}
}

func (d device) ToDevice(peersMap map[string][]peer) (wg.Device, error) {
	peers, _ := peersMap[d.Id]

	ret := wg.Device{
		Id:         d.Id,
		Name:       d.Name,
		PrivateKey: d.PrivateKey,
		Peers:      make([]wg.Peer, 0, len(peers)),
		ListenPort: d.ListenPort,
		Address:    nil,
	}

	if len(d.Address) > 0 {
		var err error
		if ret.Address, err = utils.ParseCIDRAsIPNet(d.Address); err != nil {
			return ret, err
		}
	}

	for _, peer := range peers {
		if peer, err := peer.ToPeer(); err != nil {
			return ret, err
		} else {
			ret.Peers = append(ret.Peers, peer)
		}
	}

	return ret, nil
}

func (p peer) ToPeer() (wg.Peer, error) {
	allowedIPStrings := strings.Split(p.AllowedIPs, ",")
	ret := wg.Peer{
		PeerConfig: wg.PeerConfig{
			PublicKey:           p.PublicKey,
			PreSharedKey:        p.PreSharedKey,
			Endpoint:            nil,
			AllowedIPs:          make([]net.IPNet, 0, len(allowedIPStrings)),
			PersistentKeepAlive: p.PersistentKeepAlive,
		},
	}

	var err error
	if ret.Endpoint, err = net.ResolveUDPAddr("udp", p.Endpoint); err != nil {
		return ret, err
	}

	for _, ipString := range allowedIPStrings {
		if _, ipNet, err := net.ParseCIDR(ipString); err != nil {
			return ret, err
		} else {
			ret.AllowedIPs = append(ret.AllowedIPs, *ipNet)
		}
	}

	return ret, nil
}

func createDb(dsn string, targetSchemaVersion int) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	tx, err := db.Beginx()
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			_ = db.Close()
		} else {
			if err = tx.Commit(); err != nil {
				_ = db.Close()
				db = nil
			}
		}
	}()

	schemaVersion := 0

	row := tx.QueryRow("SELECT CAST(value AS INTEGER) FROM options WHERE name = $1", optionSchemaVersion)
	_ = row.Scan(&schemaVersion)

	for v := schemaVersion; v < targetSchemaVersion; v++ {
		for _, s := range tableMigrations[v] {
			if _, err = tx.Exec(s); err != nil {
				return nil, err
			}
		}
	}

	_, err = tx.Exec("INSERT OR REPLACE INTO options(name, value) VALUES ($1, $2)", optionSchemaVersion, targetSchemaVersion)
	if err != nil {
		return nil, err
	}

	return db, nil
}

type sqlRepository struct {
	*sqlx.DB
}

func (s sqlRepository) Close() error {
	return s.Close()
}

func (s sqlRepository) SaveDevices(devices []wg.Device) error {
	tx, err := s.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	devSt, err := s.PrepareNamed(insertDeviceSql)
	if err != nil {
		return err
	}

	peerSt, err := s.PrepareNamed(insertPeerSql)
	if err != nil {
		return err
	}

	var updatingDevice device
	var updatingPeer peer

	for _, d := range devices {
		updatingDevice.UpdateFrom(d)
		if _, err = devSt.Exec(updatingDevice); err != nil {
			return err
		}

		for _, p := range d.Peers {
			updatingPeer.UpdateFrom(d, p)
			if _, err = peerSt.Exec(p); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s sqlRepository) queryPeersMap() (map[string][]peer, error) {
	rows, err := s.Queryx("SELECT * FROM peers")
	if err != nil {
		return nil, err
	}

	ret := make(map[string][]peer)
	var p peer
	for rows.Next() {
		if err = rows.StructScan(&p); err != nil {
			return ret, err
		}

		devicePeers, _ := ret[p.DeviceId]
		devicePeers = append(devicePeers, p)
		ret[p.DeviceId] = devicePeers
	}

	return ret, nil
}

func (s sqlRepository) ListDevices() (ret []wg.Device, err error) {
	rows, err := s.Queryx("SELECT * FROM devices")
	if err != nil {
		return
	}

	peersMap, err := s.queryPeersMap()
	if err != nil {
		return
	}

	var dev device
	for rows.Next() {
		if err = rows.StructScan(&dev); err != nil {
			return
		}

		if dev, e := dev.ToDevice(peersMap); e != nil {
			err = e
			return
		} else {
			ret = append(ret, dev)
		}
	}

	return
}

func (s sqlRepository) SetDeviceMeta(deviceId DeviceId, key MetaKey, value string) error {
	_, err := s.Exec("INSERT OR REPLACE INTO device_meta (device_id, name, value) VALUES ($1, $2, $3)",
		deviceId, key, value)
	return err
}

func (s sqlRepository) GetDeviceMeta(key MetaKey) (map[DeviceId]string, error) {
	rows, err := s.Query("SELECT device_id, name, value FROM device_meta WHERE name = $1", key)
	if err != nil {
		return nil, err
	}

	var deviceId, name, value string
	ret := make(map[DeviceId]string)
	for rows.Next() {
		if err = rows.Scan(&deviceId, &name, &value); err != nil {
			return ret, err
		}

		ret[DeviceId(deviceId)] = value
	}

	return ret, nil
}

func (s sqlRepository) RemoveDeviceMeta(deviceId DeviceId, key MetaKey) error {
	_, err := s.Exec("DELETE FROM device_meta WHERE device_id = $1 AND name = $2", deviceId, key)
	return err
}

func (s sqlRepository) SetPeerMeta(peerId PeerId, key MetaKey, value string) error {
	_, err := s.Exec("INSERT OR REPLACE INTO peer_meta (device_id, public_key, name, value) VALUES ($1, $2, $3, $4)",
		peerId.DeviceId, peerId.PublicKey, key, value)

	return err
}

func (s sqlRepository) GetPeerMeta(key MetaKey) (map[PeerId]string, error) {
	rows, err := s.Query("SELECT device_id, public_key, value FROM peer_meta WHERE name = $1", key)
	if err != nil {
		return nil, err
	}

	var deviceId, value string
	var publicKey wg.Key
	ret := make(map[PeerId]string)
	for rows.Next() {
		if err = rows.Scan(&deviceId, &publicKey, &value); err != nil {
			return ret, err
		}

		ret[PeerId{
			DeviceId:  DeviceId(deviceId),
			PublicKey: publicKey,
		}] = value
	}

	return ret, nil
}

func (s sqlRepository) RemovePeerMeta(id PeerId, key MetaKey) error {
	_, err := s.Exec("DELETE FROM peer_meta WHERE device_id = $1 AND public_key = $2 AND name = $3",
		id.DeviceId, id.PublicKey, key)
	return err
}

func (s sqlRepository) RemoveDevices(ids []DeviceId) error {
	_, err := s.Exec("DELETE FROM devices WHERE id IN ($1)", ids)
	return err
}

func NewSqliteRepository(dsn string) (Repository, error) {
	db, err := createDb(dsn, len(tableMigrations))
	if err != nil {
		return nil, err
	}

	repo := &sqlRepository{DB: db}
	return repo, nil
}
