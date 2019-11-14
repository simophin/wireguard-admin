package persistent

import (
	"github.com/jmoiron/sqlx"
	"nz.cloudwalker/wireguard-webadmin/wg"
	"reflect"
	"testing"
	"time"
)

func TestNewSqliteRepository(t *testing.T) {
	type args struct {
		dsn string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "correct dsn",
			args: args{
				dsn: "file:test.db?cache=shared&mode=memory",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewSqliteRepository(tt.args.dsn)
			defer repo.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSqliteRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_device_ToDevice(t *testing.T) {
	type fields struct {
		Id         string
		Name       string
		PrivateKey wg.Key
		ListenPort uint16
		Address    string
	}
	type args struct {
		peersMap map[string][]peer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    wg.Device
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := device{
				Id:         tt.fields.Id,
				Name:       tt.fields.Name,
				PrivateKey: tt.fields.PrivateKey,
				ListenPort: tt.fields.ListenPort,
				Address:    tt.fields.Address,
			}
			got, err := d.ToDevice(tt.args.peersMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToDevice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToDevice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_device_UpdateFrom(t *testing.T) {
	type fields struct {
		Id         string
		Name       string
		PrivateKey wg.Key
		ListenPort uint16
		Address    string
	}
	type args struct {
		dev wg.Device
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &device{
				Id:         tt.fields.Id,
				Name:       tt.fields.Name,
				PrivateKey: tt.fields.PrivateKey,
				ListenPort: tt.fields.ListenPort,
				Address:    tt.fields.Address,
			}
		})
	}
}

func Test_peer_ToPeer(t *testing.T) {
	type fields struct {
		DeviceId            string
		PublicKey           wg.Key
		PreSharedKey        wg.Key
		Endpoint            string
		AllowedIPs          string
		PersistentKeepAlive time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		want    wg.Peer
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := peer{
				DeviceId:            tt.fields.DeviceId,
				PublicKey:           tt.fields.PublicKey,
				PreSharedKey:        tt.fields.PreSharedKey,
				Endpoint:            tt.fields.Endpoint,
				AllowedIPs:          tt.fields.AllowedIPs,
				PersistentKeepAlive: tt.fields.PersistentKeepAlive,
			}
			got, err := p.ToPeer()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToPeer() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_peer_UpdateFrom(t *testing.T) {
	type fields struct {
		DeviceId            string
		PublicKey           wg.Key
		PreSharedKey        wg.Key
		Endpoint            string
		AllowedIPs          string
		PersistentKeepAlive time.Duration
	}
	type args struct {
		d wg.Device
		o wg.Peer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &peer{
				DeviceId:            tt.fields.DeviceId,
				PublicKey:           tt.fields.PublicKey,
				PreSharedKey:        tt.fields.PreSharedKey,
				Endpoint:            tt.fields.Endpoint,
				AllowedIPs:          tt.fields.AllowedIPs,
				PersistentKeepAlive: tt.fields.PersistentKeepAlive,
			}
		})
	}
}

func Test_sqlRepository_GetDeviceMeta(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		key MetaKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[DeviceId]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			got, err := s.GetDeviceMeta(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeviceMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDeviceMeta() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sqlRepository_GetPeerMeta(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		key MetaKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[PeerId]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			got, err := s.GetPeerMeta(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPeerMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPeerMeta() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sqlRepository_ListDevices(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	tests := []struct {
		name    string
		fields  fields
		wantRet []wg.Device
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			gotRet, err := s.ListDevices()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListDevices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("ListDevices() gotRet = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func Test_sqlRepository_RemoveDeviceMeta(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		deviceId DeviceId
		key      MetaKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			if err := s.RemoveDeviceMeta(tt.args.deviceId, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("RemoveDeviceMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqlRepository_RemoveDevices(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		ids []DeviceId
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			if err := s.RemoveDevices(tt.args.ids); (err != nil) != tt.wantErr {
				t.Errorf("RemoveDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqlRepository_RemovePeerMeta(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		id  PeerId
		key MetaKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			if err := s.RemovePeerMeta(tt.args.id, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("RemovePeerMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqlRepository_SaveDevices(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		devices []wg.Device
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			if err := s.SaveDevices(tt.args.devices); (err != nil) != tt.wantErr {
				t.Errorf("SaveDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqlRepository_SetDeviceMeta(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		deviceId DeviceId
		key      MetaKey
		value    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			if err := s.SetDeviceMeta(tt.args.deviceId, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("SetDeviceMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqlRepository_SetPeerMeta(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		peerId PeerId
		key    MetaKey
		value  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			if err := s.SetPeerMeta(tt.args.peerId, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("SetPeerMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqlRepository_queryPeersMap(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string][]peer
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			got, err := s.queryPeersMap()
			if (err != nil) != tt.wantErr {
				t.Errorf("queryPeersMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queryPeersMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}
