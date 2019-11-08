package sqlite

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"net"
	"nz.cloudwalker/wireguard-webadmin/repo"
	"strings"
	"time"
)

type device struct {
	PublicKey  string `db:"public_key"`
	PrivateKey string `db:"private_key"`
	Name       string `db:"name"`
	ListenPort uint16 `db:"listen_port"`
}

const (
	createDeviceTableSql = `CREATE TABLE devices (
		private_key TEXT NOT NULL PRIMARY KEY,
		public_key TEXT NOT NULL,
		name TEXT NOT NULL,
		listen_port INTEGER NOT NULL CHECK (listen_port >= 0 AND listen_port < 65536)
	)`

	createDeviceIndexSql = "CREATE UNIQUE INDEX devices_public_key ON devices(public_key)"

	updateDeviceSql = `INSERT OR REPLACE INTO devices (private_key, public_key, name, listen_port)
		VALUES (:private_key, :public_key, :name, :listen_port)`
)

type peer struct {
	PublicKey                   string        `db:"public_key"`
	PreSharedKey                string        `db:"pre_shared_key"`
	Endpoint                    string        `db:"endpoint"`
	PersistentKeepaliveInterval time.Duration `db:"persistent_keepalive_interval"`
	AllowedIPs                  string        `db:"allowed_ips"`
	DevicePublicKey             string        `db:"device_public_key"`
	LastHandshake               int64         `db:"last_handshake"`
	Name                        string        `db:"name"`
}

const (
	createPeerTableSql = `CREATE TABLE peers (
                       public_key TEXT NOT NULL PRIMARY KEY,
                       pre_shared_key TEXT NOT NULL,
                       endpoint TEXT NOT NULL,
                       persistent_keepalive_interval INTEGER NOT NULL DEFAULT 0,
                       allowed_ips TEXT NOT NULL,
                       device_public_key TEXT NOT NULL REFERENCES devices(public_key) ON DELETE CASCADE ,
                       last_handshake INTEGER NOT NULL DEFAULT 0,
                       name TEXT
	)`

	createPeerIndexSql1 = `CREATE INDEX peers_device_public_key ON peers(device_public_key)`

	updatePeerSql = `
		INSERT OR REPLACE INTO peers(
			public_key, pre_shared_key, endpoint, persistent_keepalive_interval, allowed_ips, device_public_key, last_handshake, name
		)
		VALUES (:public_key, :pre_shared_key, :endpoint, :persistent_keepalive_interval, :allowed_ips, :device_public_key, :last_handshake, :name)
	`
)

func (p *peer) FromPeerInfo(info repo.PeerInfo) {
	p.PublicKey = info.PublicKey
	p.PreSharedKey = info.PresharedKey
	p.PersistentKeepaliveInterval = info.PersistentKeepaliveInterval
	p.DevicePublicKey = info.DevicePublicKey
	p.Name = info.Name
	p.LastHandshake = info.LastHandshake

	if info.Endpoint != nil {
		p.Endpoint = info.Endpoint.String()
	} else {
		p.Endpoint = ""
	}

	var ips []string
	for _, ip := range info.AllowedIPs {
		ips = append(ips, ip.String())
	}
	p.AllowedIPs = strings.Join(ips, ",")
}

func (p peer) ToPeerInfo() (info repo.PeerInfo, err error) {
	info = repo.PeerInfo{
		PublicKey:                   p.PublicKey,
		PresharedKey:                p.PreSharedKey,
		PersistentKeepaliveInterval: p.PersistentKeepaliveInterval,
		DevicePublicKey:             p.DevicePublicKey,
		LastHandshake:               p.LastHandshake,
		Name:                        p.Name,
	}

	info.Endpoint, err = net.ResolveUDPAddr("udp", p.Endpoint)
	if err != nil {
		return
	}

	ips := strings.Split(p.AllowedIPs, ",")
	for _, ip := range ips {
		var ipnet *net.IPNet
		_, ipnet, err = net.ParseCIDR(ip)
		if err != nil {
			return
		}

		info.AllowedIPs = append(info.AllowedIPs, *ipnet)
	}

	return
}

func (d device) ToDeviceInfo() repo.DeviceInfo {
	return repo.DeviceInfo{
		PrivateKey: d.PrivateKey,
		PublicKey:  d.PublicKey,
		ListenPort: d.ListenPort,
		Name:       d.Name,
	}
}

func (d *device) fromDeviceInfo(info repo.DeviceInfo) {
	d.PublicKey = info.PublicKey
	d.PrivateKey = info.PrivateKey
	d.Name = info.Name
	d.ListenPort = info.ListenPort
}

type sqliteRepository struct {
	repo.DefaultChangeNotificationHandler
	db *sqlx.DB
}

func (s sqliteRepository) ListDevices() (info []repo.DeviceInfo, err error) {
	var rows *sqlx.Rows
	rows, err = s.db.Queryx("SELECT * FROM devices")
	if err != nil {
		return
	}

	defer rows.Close()

	var d device
	for rows.Next() {
		if err = rows.StructScan(&d); err != nil {
			return
		}

		info = append(info, d.ToDeviceInfo())
	}
	return
}

func (s sqliteRepository) upsertDevices(removeAll bool, devices []repo.DeviceInfo) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			if err = tx.Commit(); err == nil {
				s.NotifyChange()
			}
		}
	}()

	if removeAll {
		if _, err = tx.Exec("DELETE FROM devices WHERE 1"); err != nil {
			return err
		}
	}

	st, err := tx.PrepareNamed(updateDeviceSql)
	if err != nil {
		return err
	}

	var d device
	for _, info := range devices {
		d.fromDeviceInfo(info)
		if _, err = st.Exec(d); err != nil {
			return err
		}
	}

	return err
}

func (s sqliteRepository) UpdateDevices(devices []repo.DeviceInfo) error {
	return s.upsertDevices(false, devices)
}

func (s sqliteRepository) RemoveDevices(pubKeys []string) error {
	if _, err := s.db.Exec("DELETE FROM devices WHERE public_key IN (:1)", pubKeys); err != nil {
		return err
	} else {
		s.NotifyChange()
		return nil
	}
}

func (s sqliteRepository) ReplaceAllDevices(devices []repo.DeviceInfo) error {
	return s.upsertDevices(true, devices)
}

func (s sqliteRepository) listPeersCommon(offset uint, limit uint, order repo.PeerOrder, whereStatement string, args ...interface{}) (data []repo.PeerInfo, total uint, err error) {
	var tx *sqlx.Tx
	tx, err = s.db.Beginx()
	if err != nil {
		return
	}

	defer func() {
		_ = tx.Commit()
	}()

	var row *sqlx.Row
	if row = tx.QueryRowx(fmt.Sprint("SELECT COUNT(public_key) FROM peers WHERE", whereStatement), args...); row.Err() != nil {
		err = row.Err()
		return
	}

	if err = row.Scan(&total); err != nil {
		return
	}

	var orderByStatement string
	switch order {
	case repo.OrderLastHandshakeAsc:
		orderByStatement = "last_handshake ASC, public_key ASC"
		break
	case repo.OrderLastHandshakeDesc:
		orderByStatement = "last_handshake DESC, public_key ASC"
		break
	case repo.OrderNameAsc:
		orderByStatement = "name ASC"
		break
	case repo.OrderNameDesc:
		orderByStatement = "name DESC"
		break
	default:
		panic(repo.InvalidPeerOrder)
	}

	var st string
	if limit > 0 {
		st = fmt.Sprintf("SELECT * FROM peers WHERE %s ORDER BY %s LIMIT %v, %v", whereStatement, orderByStatement, offset, limit)
	} else {
		st = fmt.Sprintf("SELECT * FROM peers WHERE %s ORDER BY %s", whereStatement, orderByStatement)
	}

	var rows *sqlx.Rows
	if rows, err = tx.Queryx(st, args...); err != nil {
		return
	}
	defer rows.Close()

	var p peer
	for rows.Next() {
		if err = rows.StructScan(&p); err != nil {
			return
		}

		if info, e := p.ToPeerInfo(); e != nil {
			err = e
			return
		} else {
			data = append(data, info)
		}
	}
	return
}

func (s sqliteRepository) ListPeersByDevices(pubKeys []string, order repo.PeerOrder, offset uint, limit uint) (data []repo.PeerInfo, total uint, err error) {
	return s.listPeersCommon(offset, limit, order, "device_public_key IN (:1)", pubKeys)
}

func (s sqliteRepository) ListPeersByKeys(pubKeys []string, order repo.PeerOrder, offset uint, limit uint) (data []repo.PeerInfo, total uint, err error) {
	return s.listPeersCommon(offset, limit, order, "public_keys IN (:1)", pubKeys)
}

func (s sqliteRepository) ListPeers(order repo.PeerOrder, offset uint, limit uint) (data []repo.PeerInfo, total uint, err error) {
	return s.listPeersCommon(offset, limit, order, "1")
}

func (s sqliteRepository) RemovePeers(publicKeys []string) error {
	if _, err := s.db.Exec("DELETE FROM peers WHERE public_key IN (:1)", publicKeys); err != nil {
		return err
	} else {
		s.NotifyChange()
		return nil
	}
}

func (s sqliteRepository) ReplaceAllPeers(peers []repo.PeerInfo) error {
	return s.upsertPeers(true, peers)
}

func (s *sqliteRepository) Close() error {
	err := s.db.Close()
	s.db = nil
	err = s.DefaultChangeNotificationHandler.Close()
	return err
}

func (s sqliteRepository) upsertPeers(removeAll bool, peers []repo.PeerInfo) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else if err = tx.Commit(); err == nil {
			s.NotifyChange()
		}
	}()

	if removeAll {
		if _, err = tx.Exec("DELETE FROM peers WHERE 1"); err != nil {
			return err
		}
	}

	st, err := tx.PrepareNamed(updatePeerSql)
	if err != nil {
		return err
	}

	defer st.Close()

	var p peer
	for _, peerInfo := range peers {
		p.FromPeerInfo(peerInfo)
		if _, err = st.Exec(p); err != nil {
			return err
		}
	}

	return nil
}

func (s sqliteRepository) UpdatePeers(peers []repo.PeerInfo) error {
	return s.upsertPeers(false, peers)
}

func NewSqliteRepository(dsn string) (repo repo.Repository, err error) {
	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return
	}

	tx, err := db.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if e, ok := recover().(error); ok {
			err = e
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	db.MustExec(createDeviceTableSql)
	db.MustExec(createDeviceIndexSql)
	db.MustExec(createPeerTableSql)
	db.MustExec(createPeerIndexSql1)

	repo = &sqliteRepository{
		db: db,
	}

	return
}
