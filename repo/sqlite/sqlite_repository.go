package sqlite

import (
	"database/sql"
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
}

const (
	createDeviceTableSql = `CREATE TABLE devices (
		private_key TEXT NOT NULL PRIMARY,
		public_key TEXT NOT NULL,
		name TEXT NOT NULL
	)`

	createDeviceIndexSql = "CREATE UNIQUE INDEX devices_public_key ON devices(public_key)",
)

type peer struct {
	PublicKey                   string        `db:"public_key"`
	PreSharedKey                string        `db:"pre_shared_key"`
	Endpoint                    string        `db:"endpoint"`
	PersistentKeepaliveInterval time.Duration `db:"persistent_keepalive_interval"`
	AllowedIPs                  string        `db:"allowed_ips"`
	DevicePublicKey             string        `db:"device_public_key"`
	LastHandshake               sql.NullTime  `db:"last_handshake"`
	Name                        string        `db:"name"`
}

const (
	createPeerTableSql = `CREATE TABLE IF NOT EXISTS peers (
			public_key TEXT NOT NULL PRIMARY KEY,
			pre_shared_key TEXT NOT NULL,
			endpoint TEXT NOT NULL,
			persistent_keepalive_interval INTEGER NOT NULL DEFAULT 0,
			allowed_ips TEXT NOT NULL,
			device_public_key TEXT NOT NULL REFERENCES devices.public_key ON CASCADE DELETE,
			last_handshake INTEGER,
			name TEXT
		)`

	createPeerIndexSql1 = `CREATE INDEX peers_device_public_key ON peers(device_public_key)`
)

func fromPeerInfo(info repo.PeerInfo) peer {
	p := peer{
		PublicKey:                   info.PublicKey,
		PreSharedKey:                info.PresharedKey,
		PersistentKeepaliveInterval: info.PersistentKeepaliveInterval,
		DevicePublicKey:             info.DevicePublicKey,
		Name:                        info.Name,
	}

	if info.Endpoint != nil {
		p.Endpoint = info.Endpoint.String()
	}

	if info.LastHandshake != nil {
		p.LastHandshake.Time = *info.LastHandshake
	}

	var ips []string
	for _, ip := range info.AllowedIPs {
		ips = append(ips, ip.String())
	}
	p.AllowedIPs = strings.Join(ips, ",")

	return p
}

func (p peer) ToPeerInfo() (info repo.PeerInfo, err error) {
	info = repo.PeerInfo{
		PublicKey:                   p.PublicKey,
		PresharedKey:                p.PreSharedKey,
		PersistentKeepaliveInterval: p.PersistentKeepaliveInterval,
		DevicePublicKey:             p.DevicePublicKey,
		Name:                        p.Name,
	}

	if p.LastHandshake.Valid {
		t := p.LastHandshake.Time
		info.LastHandshake = &t
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

type sqliteRepository struct {
	repo.DefaultChangeNotificationHandler
	db        *sqlx.DB
	listeners map[chan<- interface{}]interface{}
}

func (s *sqliteRepository) Close() error {
	err := s.db.Close()
	s.listeners = nil
	s.db = nil
	return err
}

func (s sqliteRepository) UpdatePeers(peers []repo.PeerInfo) error {
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

	st, err := tx.PrepareNamed(updatePeerSql)
	if err != nil {
		return err
	}

	var p peer
	for _, peerInfo := range peers {
		p.FromPeerInfo(peerInfo)
		if _, err = st.Exec(p); err != nil {
			return err
		}
	}

	return nil
}

const (
	updatePeerSql = `
		INSERT OR REPLACE INTO peers(
			public_key, pre_shared_key, endpoint, persistent_keepalive_interval, allowed_ips, network_device_name, time_created, time_last_seen, name
		)
		VALUES (:public_key, :pre_shared_key, :endpoint, :persistent_keepalive_interval, :allowed_ips, :network_device_name, :time_created, :time_last_seen, :name)
	`
)

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
