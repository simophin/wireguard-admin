package sqlite

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"nz.cloudwalker/wireguard-webadmin/repo"
	"strings"
	"time"
)

type sqlitePeer struct {
	PublicKey                   string         `db:"public_key"`
	PresharedKey                sql.NullString `db:"pre_shared_key"`
	Endpoint                    sql.NullString `db:"endpoint"`
	PersistentKeepaliveInterval sql.NullInt32  `db:"persistent_keepalive_interval"`
	AllowedIPs                  sql.NullString `db:"allowed_ips"`
	NetworkDeviceName           sql.NullString `db:"network_device_name"`
	Name                        sql.NullString `db:"name"`
	TimeCreated                 uint64         `db:"time_created"`
	TimeLastSeen                sql.NullInt64  `db:"time_last_seen"`
}

func (p *sqlitePeer) FromPeerInfo(info repo.PeerInfo) {
	p.PublicKey = info.PublicKey.String()
	p.PresharedKey.String = info.PresharedKey.String()

	var ips []string
	for _, ip := range info.AllowedIPs {
		ips = append(ips, ip.String())
	}
	p.AllowedIPs.String = strings.Join(ips, ",")

	if info.Endpoint != nil {
		p.Endpoint.String = info.Endpoint.String()
	}

	p.Name.String = info.Name
	p.NetworkDeviceName.String = info.NetworkDeviceName
	p.PersistentKeepaliveInterval.Int32 = int32(info.PersistentKeepaliveInterval)
}

func (p sqlitePeer) ToPeerInfo(info *repo.PeerInfo) error {
	var err error
	if info.PublicKey, err = wgtypes.ParseKey(p.PublicKey); err != nil {
		return err
	}

	if p.PresharedKey.Valid {
		if info.PresharedKey, err = wgtypes.ParseKey(p.PresharedKey.String); err != nil {
			return err
		}
	}

	if p.Endpoint.Valid {
		if info.Endpoint, err = net.ResolveUDPAddr("udp", p.Endpoint.String); err != nil {
			return err
		}
	}

	if p.PersistentKeepaliveInterval.Valid {
		info.PersistentKeepaliveInterval = time.Duration(p.PersistentKeepaliveInterval.Int32)
	}

	if p.AllowedIPs.Valid {
		ips := strings.Split(p.AllowedIPs.String, ",")
		for _, ip := range ips {
			if _, ip, e := net.ParseCIDR(ip); e != nil {
				return e
			} else {
				info.AllowedIPs = append(info.AllowedIPs, ip)
			}
		}
	}

	if p.NetworkDeviceName.Valid {
		info.NetworkDeviceName = p.NetworkDeviceName.String
	}

	if p.Name.Valid {
		info.Name = p.Name.String
	}

	return nil
}

type sqliteRepository struct {
	db        *sqlx.DB
	listeners map[chan<- interface{}]interface{}
}

func (s sqliteRepository) notifyChange() {
	go func() {
		defer func() {
			recover()
		}()

		for c, _ := range s.listeners {
			c <- nil
		}
	}()
}

func (s *sqliteRepository) AddChangeNotification(channel chan<- interface{}) {
	s.listeners[channel] = nil
}

func (s *sqliteRepository) RemoveChangeNotification(channel chan<- interface{}) {
	delete(s.listeners, channel)
}

func (s *sqliteRepository) Close() error {
	err := s.db.Close()
	s.listeners = nil
	s.db = nil
	return err
}

func (s sqliteRepository) ListAllPeers(out *[]repo.PeerInfo, offset uint32, limit uint32) (total uint32, err error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	if err = tx.Get(&total, "SELECT COUNT(public_key) FROM peers ORDER BY COALESCE(time_last_seen, time_created) DESC"); err != nil {
		return
	}

	var rows *sqlx.Rows

	statement := "SELECT * FROM peers ORDER BY COALESCE(time_last_seen, time_created) DESC"
	if limit > 0 {
		statement += fmt.Sprintf(" LIMIT %d, %d", offset, limit)
	}

	if rows, err = tx.Queryx(statement); err != nil {
		return
	}

	defer rows.Close()

	var p sqlitePeer
	var peerInfo repo.PeerInfo

	for rows.Next() {
		if err = rows.StructScan(&p); err != nil {
			return
		}

		if err = p.ToPeerInfo(&peerInfo); err != nil {
			return
		}

		*out = append(*out, peerInfo)
	}

	return
}

func (s sqliteRepository) RemovePeer(publicKey wgtypes.Key) error {
	if _, err := s.db.Exec("DELETE FROM peers WHERE public_key = :1", publicKey.String()); err != nil {
		return err
	} else {
		s.notifyChange()
		return nil
	}
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
				s.notifyChange()
			}
		}
	}()

	st, err := tx.PrepareNamed(updatePeerSql)
	if err != nil {
		return err
	}

	var p sqlitePeer
	for _, peerInfo := range peers {
		p.FromPeerInfo(peerInfo)
		if _, err = st.Exec(p); err != nil {
			return err
		}
	}

	return nil
}

const (
	createTableSql = `
		CREATE TABLE IF NOT EXISTS peers (
			public_key TEXT NOT NULL PRIMARY KEY,
			pre_shared_key TEXT,
			endpoint TEXT,
			persistent_keepalive_interval INTEGER,
			allowed_ips TEXT,
			network_device_name TEXT,
			time_created INTEGER NOT NULL,
			time_last_seen INTEGER,
			name TEXT
		)
	`
	updatePeerSql = `
		INSERT OR REPLACE INTO peers(
			public_key, pre_shared_key, endpoint, persistent_keepalive_interval, allowed_ips, network_device_name, time_created, time_last_seen, name
		)
		VALUES (:public_key, :pre_shared_key, :endpoint, :persistent_keepalive_interval, :allowed_ips, :network_device_name, :time_created, :time_last_seen, :name)
	`
)

func NewSqliteRepository(dsn string) (repo.Repository, error) {
	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec(createTableSql); err != nil {
		return nil, err
	}

	return &sqliteRepository{
		db:        db,
		listeners: make(map[chan<- interface{}]interface{}),
	}, nil
}
