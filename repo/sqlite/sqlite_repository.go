package sqlite

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"nz.cloudwalker/wireguard-webadmin/repo"
	"strings"
	"time"
)

type sqlitePeer struct {
	PublicKey                   string         `db:"public_key"`
	PresharedKey                sql.NullString `db:"preshared_key"`
	Endpoint                    sql.NullString `db:"endpoint"`
	PersistentKeepaliveInterval sql.NullInt32  `db:"persistent_keepalive_interval"`
	AllowedIPs                  sql.NullString `db:"allowed_ips"`
	NetworkDeviceName           sql.NullString `db:"network_device_name"`
	Name                        sql.NullString `db:"name"`
}

func (p sqlitePeer) ToPeerInfo() (info repo.PeerInfo, err error) {
	if info.PublicKey, err = wgtypes.ParseKey(p.PublicKey); err != nil {
		return
	}

	if p.PresharedKey.Valid {
		if info.PresharedKey, err = wgtypes.ParseKey(p.PresharedKey.String); err != nil {
			return
		}
	}

	if p.Endpoint.Valid {
		if info.Endpoint, err = net.ResolveUDPAddr("udp", p.Endpoint.String); err != nil {
			return
		}
	}

	if p.PersistentKeepaliveInterval.Valid {
		info.PersistentKeepaliveInterval = time.Duration(p.PersistentKeepaliveInterval.Int32)
	}

	if p.AllowedIPs.Valid {
		ips := strings.Split(p.AllowedIPs.String, ",")
		for _, ip := range ips {
			if _, ip, e := net.ParseCIDR(ip); e != nil {
				err = e
				return
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

	return
}

type sqliteRepository struct {
	db      *sqlx.DB
	changes chan interface{}
}

func (s sqliteRepository) ListAllPeers() (<-chan []repo.PeerInfo, error) {
	retChan := make(chan []repo.PeerInfo)
	var ret []repo.PeerInfo
	row := s.db.QueryRowx("SELECT * FROM peers")
	if row.Err() != nil {
		return retChan, row.Err()
	}


}

func (s sqliteRepository) RemovePeer(publicKey wgtypes.Key) error {
	panic("implement me")
}

func (s sqliteRepository) UpdatePeers(peers []repo.PeerInfo) error {
	panic("implement me")
}

func NewSqliteRepository(dsn string) (repo.Repository, error) {
	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	db.MustExec(`
		CREATE TABLE IF NOT EXISTS peers (
			public_key TEXT NOT NULL PRIMARY KEY,
			pre_shared_key TEXT,
			endpoint TEXT,
			persistent_keepalive_interval INTEGER,
			allowed_ips TEXT,
			network_device_name TEXT,
			name TEXT
		)
	`)

	return &sqliteRepository{
		db: db,
		changes: make(chan interface{}),
	}, nil
}
