package sqlite

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"net"
	"nz.cloudwalker/wireguard-webadmin/repo"
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
		{name: "Valid DSN", args: args{dsn: "file:test.db?cache=shared&mode=memory"}, wantErr: false},
		{name: "Invalid DSN", args: args{dsn: "file:test.db?cache=shared&mode=unknown"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSqliteRepository(tt.args.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSqliteRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_device_ToDeviceInfo(t *testing.T) {
	type fields struct {
		PublicKey  string
		PrivateKey string
		Name       string
		ListenPort uint16
	}
	tests := []struct {
		name   string
		fields fields
		want   repo.DeviceInfo
	}{
		{
			name: "Test",
			fields: fields{
				PublicKey:  "pk1",
				PrivateKey: "privkey",
				Name:       "name",
				ListenPort: 123,
			},
			want: repo.DeviceInfo{
				PrivateKey: "privkey",
				PublicKey:  "pk1",
				ListenPort: 123,
				Name:       "name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := device{
				PublicKey:  tt.fields.PublicKey,
				PrivateKey: tt.fields.PrivateKey,
				Name:       tt.fields.Name,
				ListenPort: tt.fields.ListenPort,
			}
			if got := d.ToDeviceInfo(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToDeviceInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_device_fromDeviceInfo(t *testing.T) {
	type args struct {
		info repo.DeviceInfo
	}
	tests := []struct {
		name string
		args args
		want device
	}{
		{
			name: "First time init",
			args: args{
				info: repo.DeviceInfo{
					PrivateKey: "privatekey",
					PublicKey:  "publickey",
					ListenPort: 456,
					Name:       "name",
				},
			},
			want: device{
				PublicKey:  "publickey",
				PrivateKey: "privatekey",
				Name:       "name",
				ListenPort: 456,
			},
		},
		{
			name: "Second time init",
			args: args{
				info: repo.DeviceInfo{
					PrivateKey: "privatekey1",
					PublicKey:  "publickey1",
					ListenPort: 457,
					Name:       "name1",
				},
			},
			want: device{
				PublicKey:  "publickey1",
				PrivateKey: "privatekey1",
				Name:       "name1",
				ListenPort: 457,
			},
		},
	}

	var got device
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got.fromDeviceInfo(tt.args.info)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromDeviceInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mustResolveUdp(address string) *net.UDPAddr {
	udp, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		panic(err)
	}

	return udp
}

func mustResolveIPNet(address string) net.IPNet {
	_, ret, err := net.ParseCIDR(address)
	if err != nil {
		panic(err)
	}

	return *ret
}

func createTimePointer(timestamp int64) *time.Time {
	t := createTime(timestamp)
	return &t
}

func createTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func Test_peer_FromPeerInfo(t *testing.T) {
	type args struct {
		info repo.PeerInfo
	}
	tests := []struct {
		name string
		args args
		want peer
	}{
		{
			name: "Empty variable test",
			args: args{
				info: repo.PeerInfo{
					PublicKey:                   "pubkey",
					PresharedKey:                "preshared_key",
					Endpoint:                    mustResolveUdp("1.2.3.4:1234"),
					PersistentKeepaliveInterval: 20,
					AllowedIPs:                  []net.IPNet{mustResolveIPNet("1.2.3.4/24"), mustResolveIPNet("4.5.6.7/32")},
					DevicePublicKey:             "pub_key",
					LastHandshake:               createTimePointer(123),
					Name:                        "name1",
				},
			},
			want: peer{
				PublicKey:                   "pubkey",
				PreSharedKey:                "preshared_key",
				Endpoint:                    "1.2.3.4:1234",
				PersistentKeepaliveInterval: 20,
				AllowedIPs:                  "1.2.3.0/24,4.5.6.7/32",
				DevicePublicKey:             "pub_key",
				LastHandshake:               sql.NullTime{Time: createTime(123), Valid: true},
				Name:                        "name1",
			},
		},
		{
			name: "Override variable test",
			args: args{
				info: repo.PeerInfo{
					PublicKey:                   "pubkey1",
					PresharedKey:                "preshared_key1",
					Endpoint:                    mustResolveUdp("1.2.3.5:1234"),
					PersistentKeepaliveInterval: 0,
					AllowedIPs:                  []net.IPNet{mustResolveIPNet("1.2.3.5/24"), mustResolveIPNet("4.5.6.8/32")},
					DevicePublicKey:             "pub_key",
					LastHandshake:               nil,
					Name:                        "name2",
				},
			},
			want: peer{
				PublicKey:                   "pubkey1",
				PreSharedKey:                "preshared_key1",
				Endpoint:                    "1.2.3.5:1234",
				PersistentKeepaliveInterval: 0,
				AllowedIPs:                  "1.2.3.0/24,4.5.6.8/32",
				DevicePublicKey:             "pub_key",
				LastHandshake:               sql.NullTime{},
				Name:                        "name2",
			},
		},
	}

	var got peer
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got.FromPeerInfo(tt.args.info)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromPeerInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_peer_ToPeerInfo(t *testing.T) {
	type fields struct {
		PublicKey                   string
		PreSharedKey                string
		Endpoint                    string
		PersistentKeepaliveInterval time.Duration
		AllowedIPs                  string
		DevicePublicKey             string
		LastHandshake               sql.NullTime
		Name                        string
	}
	tests := []struct {
		name     string
		fields   fields
		wantInfo repo.PeerInfo
		wantErr  bool
	}{
		{
			name: "correct info",
			fields: fields{
				PublicKey:                   "pubkey",
				PreSharedKey:                "presharedkey",
				Endpoint:                    "1.2.3.4:5000",
				PersistentKeepaliveInterval: 20,
				AllowedIPs:                  "1.2.3.0/24,4.5.6.7/32",
				DevicePublicKey:             "device_pubkey",
				LastHandshake:               sql.NullTime{},
				Name:                        "name",
			},
			wantInfo: repo.PeerInfo{
				PublicKey:                   "pubkey",
				PresharedKey:                "presharedkey",
				Endpoint:                    mustResolveUdp("1.2.3.4:5000"),
				PersistentKeepaliveInterval: 20,
				AllowedIPs:                  []net.IPNet{mustResolveIPNet("1.2.3.4/24"), mustResolveIPNet("4.5.6.7/32")},
				DevicePublicKey:             "device_pubkey",
				LastHandshake:               nil,
				Name:                        "name",
			},
			wantErr: false,
		},
		{
			name: "incorrect endpoint",
			fields: fields{
				PublicKey:                   "pubkey",
				PreSharedKey:                "presharedkey",
				Endpoint:                    "not an address",
				PersistentKeepaliveInterval: 20,
				AllowedIPs:                  "1.2.3.0/24,4.5.6.7/32",
				DevicePublicKey:             "device_pubkey",
				LastHandshake:               sql.NullTime{},
				Name:                        "name",
			},
			wantErr: true,
		},
		{
			name: "incorrect allowed ips",
			fields: fields{
				PublicKey:                   "pubkey",
				PreSharedKey:                "presharedkey",
				Endpoint:                    "1.2.3.4:5000",
				PersistentKeepaliveInterval: 20,
				AllowedIPs:                  "not an address",
				DevicePublicKey:             "device_pubkey",
				LastHandshake:               sql.NullTime{},
				Name:                        "name",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := peer{
				PublicKey:                   tt.fields.PublicKey,
				PreSharedKey:                tt.fields.PreSharedKey,
				Endpoint:                    tt.fields.Endpoint,
				PersistentKeepaliveInterval: tt.fields.PersistentKeepaliveInterval,
				AllowedIPs:                  tt.fields.AllowedIPs,
				DevicePublicKey:             tt.fields.DevicePublicKey,
				LastHandshake:               tt.fields.LastHandshake,
				Name:                        tt.fields.Name,
			}
			gotInfo, err := p.ToPeerInfo()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ToPeerInfo() error = %v, wantErr: %v", err, tt.wantErr)
				}
			} else if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
				t.Errorf("ToPeerInfo() gotInfo = %v, want %v", gotInfo, tt.wantInfo)
			}
		})
	}
}

func mustCreateRepository(t *testing.T) *sqliteRepository {
	r, err := NewSqliteRepository("file:test.db?cache=shared&mode=memory")
	if err != nil {
		return nil
	}

	return r.(*sqliteRepository)
}

func Test_sqliteRepository_Close(t *testing.T) {
	r := mustCreateRepository(t)

	if err := r.Close(); err != nil {
		t.Error("error closing repo:", err)
	}

	if r.db != nil {
		t.Error("database is not closed")
	}
}

func genNewDevices(num int) []repo.DeviceInfo {
	var ret []repo.DeviceInfo
	for i := 0; i < num; i++ {
		ret = append(ret, repo.DeviceInfo{
			PrivateKey: fmt.Sprint("privatekey", i),
			PublicKey:  fmt.Sprint("publickey", i),
			ListenPort: uint16(i),
			Name:       fmt.Sprint("device", i),
		})
	}
}

func Test_sqliteRepository_ListDevices(t *testing.T) {
	tests := []struct {
		name     string
		arg      []repo.DeviceInfo
		wantInfo []repo.DeviceInfo
		wantErr  bool
	}{
		{
			name:     "empty",
			arg:      nil,
			wantInfo: make([]repo.DeviceInfo, 0),
			wantErr:  false,
		},
		{
			name:     "one element",
			arg:      nil,
			wantInfo: make([]repo.DeviceInfo, 0),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mustCreateRepository(t)
			gotInfo, err := s.ListDevices()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListDevices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
				t.Errorf("ListDevices() gotInfo = %v, want %v", gotInfo, tt.wantInfo)
			}
		})
	}
}

func Test_sqliteRepository_ListPeers(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		order  repo.PeerOrder
		offset uint
		limit  uint
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantData  []repo.PeerInfo
		wantTotal uint
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			gotData, gotTotal, err := s.ListPeers(tt.args.order, tt.args.offset, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("ListPeers() gotData = %v, want %v", gotData, tt.wantData)
			}
			if gotTotal != tt.wantTotal {
				t.Errorf("ListPeers() gotTotal = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func Test_sqliteRepository_ListPeersByDevices(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		pubKeys []string
		order   repo.PeerOrder
		offset  uint
		limit   uint
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantData  []repo.PeerInfo
		wantTotal uint
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			gotData, gotTotal, err := s.ListPeersByDevices(tt.args.pubKeys, tt.args.order, tt.args.offset, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPeersByDevices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("ListPeersByDevices() gotData = %v, want %v", gotData, tt.wantData)
			}
			if gotTotal != tt.wantTotal {
				t.Errorf("ListPeersByDevices() gotTotal = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func Test_sqliteRepository_ListPeersByKeys(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		pubKeys []string
		order   repo.PeerOrder
		offset  uint
		limit   uint
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantData  []repo.PeerInfo
		wantTotal uint
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			gotData, gotTotal, err := s.ListPeersByKeys(tt.args.pubKeys, tt.args.order, tt.args.offset, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPeersByKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("ListPeersByKeys() gotData = %v, want %v", gotData, tt.wantData)
			}
			if gotTotal != tt.wantTotal {
				t.Errorf("ListPeersByKeys() gotTotal = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func Test_sqliteRepository_RemoveDevices(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		pubKeys []string
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
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			if err := s.RemoveDevices(tt.args.pubKeys); (err != nil) != tt.wantErr {
				t.Errorf("RemoveDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqliteRepository_RemovePeers(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		publicKeys []string
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
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			if err := s.RemovePeers(tt.args.publicKeys); (err != nil) != tt.wantErr {
				t.Errorf("RemovePeers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqliteRepository_ReplaceAllDevices(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		devices []repo.DeviceInfo
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
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			if err := s.ReplaceAllDevices(tt.args.devices); (err != nil) != tt.wantErr {
				t.Errorf("ReplaceAllDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqliteRepository_ReplaceAllPeers(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		peers []repo.PeerInfo
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
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			if err := s.ReplaceAllPeers(tt.args.peers); (err != nil) != tt.wantErr {
				t.Errorf("ReplaceAllPeers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqliteRepository_UpdateDevices(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		devices []repo.DeviceInfo
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
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			if err := s.UpdateDevices(tt.args.devices); (err != nil) != tt.wantErr {
				t.Errorf("UpdateDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqliteRepository_UpdatePeers(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		peers []repo.PeerInfo
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
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			if err := s.UpdatePeers(tt.args.peers); (err != nil) != tt.wantErr {
				t.Errorf("UpdatePeers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqliteRepository_listPeersCommon(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		offset         uint
		limit          uint
		order          repo.PeerOrder
		whereStatement string
		args           []interface{}
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantData  []repo.PeerInfo
		wantTotal uint
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			gotData, gotTotal, err := s.listPeersCommon(tt.args.offset, tt.args.limit, tt.args.order, tt.args.whereStatement, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("listPeersCommon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("listPeersCommon() gotData = %v, want %v", gotData, tt.wantData)
			}
			if gotTotal != tt.wantTotal {
				t.Errorf("listPeersCommon() gotTotal = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func Test_sqliteRepository_upsertDevices(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		removeAll bool
		devices   []repo.DeviceInfo
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
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			if err := s.upsertDevices(tt.args.removeAll, tt.args.devices); (err != nil) != tt.wantErr {
				t.Errorf("upsertDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqliteRepository_upsertPeers(t *testing.T) {
	type fields struct {
		DefaultChangeNotificationHandler repo.DefaultChangeNotificationHandler
		db                               *sqlx.DB
		listeners                        map[chan<- interface{}]interface{}
	}
	type args struct {
		removeAll bool
		peers     []repo.PeerInfo
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
			s := sqliteRepository{
				DefaultChangeNotificationHandler: tt.fields.DefaultChangeNotificationHandler,
				db:                               tt.fields.db,
				listeners:                        tt.fields.listeners,
			}
			if err := s.upsertPeers(tt.args.removeAll, tt.args.peers); (err != nil) != tt.wantErr {
				t.Errorf("upsertPeers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
