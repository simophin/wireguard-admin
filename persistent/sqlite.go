package persistent

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"nz.cloudwalker/wireguard-webadmin/wg"
	"strconv"
	"strings"
)

var tableMigrations = [][]string{
	{
		`CREATE TABLE options(
				name TEXT NOT NULL PRIMARY KEY,
				value TEXT
			)`,

		`CREATE TABLE devices(
				id TEXT NOT NULL PRIMARY KEY,
				private_key TEXT NOT NULL,
				listen_port INTEGER NOT NULL DEFAULT 0,
				address TEXT NOT NULL
			)`,

		`CREATE TABLE peers(
				device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
				id TEXT NOT NULL,
				public_key TEXT NOT NULL,
				pre_shared_key TEXT NOT NULL,
				endpoint TEXT NOT NULL,
				allowed_ips TEXT NOT NULL,
				persistent_keep_alive INTEGER NOT NULL,
				PRIMARY KEY (device_id, id) ON CONFLICT REPLACE
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
				peer_id TEXT NOT NULL,
				name TEXT NOT NULL,
				value TEXT NOT NULL,
				PRIMARY KEY (device_id, peer_id, name),
				FOREIGN KEY (device_id, peer_id)
					REFERENCES peers (device_id, id)
					ON DELETE CASCADE
			)`,
	},
}

const (
	insertDeviceSql = `INSERT OR REPLACE INTO devices(id, private_key, listen_port, address)
						VALUES (:id, :private_key, :listen_port, :address)`

	insertPeerSql = `INSERT OR REPLACE INTO peers(device_id, id, public_key, pre_shared_key, endpoint, allowed_ips, persistent_keep_alive)
					  VALUES (:device_id, :id, :public_key, :pre_shared_key, :endpoint, :allowed_ips, :persistent_keep_alive)`
)

const (
	optionSchemaVersion = "schema_version"
)

func namedParametersDevice(device Device) map[string]string {
	ret := map[string]string{
		"id":          string(device.Id),
		"private_key": device.PrivateKey.String(),
		"listen_port": strconv.Itoa(int(device.ListenPort)),
	}

	if device.Address != nil {
		ret["address"] = device.Address.String()
	}

	return ret
}

func namedParametersPeer(deviceId string, peer Peer) map[string]string {
	var endpoint string
	if peer.Endpoint != nil {
		endpoint = peer.Endpoint.String()
	}

	ips := make([]string, 0, len(peer.AllowedIPs))
	for _, ip := range peer.AllowedIPs {
		ips = append(ips, ip.String())
	}

	return map[string]string{
		"id":                    peer.Id,
		"device_id":             deviceId,
		"public_key":            peer.PublicKey.String(),
		"pre_shared_key":        peer.PreSharedKey.String(),
		"endpoint":              endpoint,
		"allowed_ips":           strings.Join(ips, ","),
		"persistent_keep_alive": peer.PersistentKeepAlive.String(),
	}
}

func createDb(dsn string, targetSchemaVersion int) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
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
	if err = row.Scan(&schemaVersion); err != nil {
		return nil, err
	}

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

func (s sqlRepository) RemoveDevices(ids []DeviceId) error {
	panic("implement me")
}

func (s sqlRepository) ListDevices() ([]Device, error) {
	panic("implement me")
}

func (s sqlRepository) SavePeers(peers []Peer) error {
	panic("implement me")
}

func (s sqlRepository) RemovePeers(ids []PeerId) error {
	panic("implement me")
}

func (s sqlRepository) ListPeersByDevice(deviceId DeviceId) ([]Peer, error) {
	panic("implement me")
}

func (s sqlRepository) ListPeers(ids []PeerId) ([]Peer, error) {
	panic("implement me")
}

func (s sqlRepository) GetDeviceMeta(ids []DeviceId, key MetaKey) (map[DeviceId]string, error) {
	panic("implement me")
}

func (s sqlRepository) SaveDeviceMeta(id DeviceId, data map[MetaKey]string) error {
	panic("implement me")
}

func (s sqlRepository) RemoveDeviceMeta(id DeviceId, keys []MetaKey) error {
	panic("implement me")
}

func (s sqlRepository) GetPeerMeta(ids []PeerId, key MetaKey) (map[PeerId]string, error) {
	panic("implement me")
}

func (s sqlRepository) SavePeerMeta(id PeerId, data map[MetaKey]string) error {
	panic("implement me")
}

func (s sqlRepository) RemovePeerMeta(id PeerId, keys []MetaKey) {
	panic("implement me")
}

func (s sqlRepository) SaveDevices(devices []Device) error {
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

	insertDeviceSt, err := tx.PrepareNamed(insertDeviceSql)
	if err != nil {
		return err
	}
	defer insertDeviceSt.Close()

	insertPeerSt, err := tx.PrepareNamed(insertPeerSql)
	if err != nil {
		return err
	}
	defer insertPeerSt.Close()

	for _, device := range devices {

		if _, err = insertDeviceSt.Exec(namedParametersDevice(device)); err != nil {
			return err
		}

		for _, peer := range device.Peers {
			if _, err = insertDeviceSt.Exec(namedParametersPeer(device.Id, peer)); err != nil {
				return err
			}
		}
	}

	return nil
}

func NewSqliteRepository(dsn string) (Repository, error) {
	db, err := createDb(dsn, len(tableMigrations))
	if err != nil {
		return nil, err
	}

	repo := &sqlRepository{DB: db}
	return repo, nil
}
