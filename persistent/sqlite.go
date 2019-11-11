package persistent

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"nz.cloudwalker/wireguard-webadmin/wg"
)

var (
	tableMigrations = [][]string{
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
)

const (
	optionSchemaVersion = "schema_version"
)

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

	row := tx.QueryRowx("SELECT CAST(value AS INTEGER) FROM options WHERE name = %1", optionSchemaVersion)
	if err = row.Err(); err == nil {
		if err = row.Scan(&schemaVersion); err != nil {
			return nil, err
		}
	}

	for v := schemaVersion; v < targetSchemaVersion; v++ {
		for _, s := range tableMigrations[v] {
			if _, err = tx.Exec(s); err != nil {
				return nil, err
			}
		}
	}

	_, err = tx.Exec(fmt.Sprintf("INSERT OR REPLACE INTO options(name, value) VALUES ('%v', '%v')",
		optionSchemaVersion, targetSchemaVersion))
	if err != nil {
		return nil, err
	}

	return db, nil
}

type sqlRepository struct {
	*sqlx.DB
}

func (s sqlRepository) GetNames(t MetaType, ids []string) (ret []string, err error) {
	var rows *sqlx.Rows

	switch t {
	case META_DEVICE: {
		if rows, err = s.DB.Queryx("SELECT ")
		break
	}
	}
}

func (s sqlRepository) SetNames(t MetaType, names map[string]string) error {
	panic("implement me")
}

func (s sqlRepository) RemoveNames(t MetaType, ids []string) error {
	panic("implement me")
}

func (s sqlRepository) SaveDevices(devices []wg.DeviceConfig) error {
	panic("implement me")
}

func (s sqlRepository) RemoveDevices(ids []string) error {
	panic("implement me")
}

func (s sqlRepository) ListDevices() ([]wg.DeviceConfig, error) {
	panic("implement me")
}

func NewSqliteRepository(dsn string) (DeviceRepository, MetaRepository, error) {
	db, err := createDb(dsn, len(tableMigrations))
	if err != nil {
		return nil, nil, err
	}

	repo := &sqlRepository{DB: db}
	return repo, repo, nil
}
