package sqlite

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"net"
	"nz.cloudwalker/wireguard-webadmin/repo"
	"reflect"
	"sort"
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
					LastHandshake:               123,
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
				LastHandshake:               123,
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
					LastHandshake:               0,
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
				LastHandshake:               0,
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
		LastHandshake               int64
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
				LastHandshake:               0,
				Name:                        "name",
			},
			wantInfo: repo.PeerInfo{
				PublicKey:                   "pubkey",
				PresharedKey:                "presharedkey",
				Endpoint:                    mustResolveUdp("1.2.3.4:5000"),
				PersistentKeepaliveInterval: 20,
				AllowedIPs:                  []net.IPNet{mustResolveIPNet("1.2.3.4/24"), mustResolveIPNet("4.5.6.7/32")},
				DevicePublicKey:             "device_pubkey",
				LastHandshake:               0,
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
				LastHandshake:               0,
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
				LastHandshake:               0,
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

var repoSeq = 0

func mustCreateRepository(t *testing.T) *sqliteRepository {
	repoSeq++
	r, err := NewSqliteRepository(fmt.Sprintf("file:test%d.db?cache=shared&mode=memory", repoSeq))
	if err != nil {
		panic(err)
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
	return ret
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
			wantInfo: nil,
			wantErr:  false,
		},
		{
			name:     "one element",
			arg:      genNewDevices(1),
			wantInfo: genNewDevices(1),
			wantErr:  false,
		},
		{
			name:     "multiple elements",
			arg:      genNewDevices(10),
			wantInfo: genNewDevices(10),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mustCreateRepository(t)
			if err := s.UpdateDevices(tt.arg); err != nil {
				t.Error("error updating devices:", err)
			}

			gotInfo, err := s.ListDevices()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListDevices() error = %v, wantErr %v", err, tt.wantErr)
			}

			sort.Slice(gotInfo, func(i, j int) bool {
				return gotInfo[i].Name < gotInfo[j].Name
			})

			sort.Slice(tt.wantInfo, func(i, j int) bool {
				return tt.wantInfo[i].Name < tt.wantInfo[j].Name
			})

			if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
				t.Errorf("ListDevices() gotInfo = %v, want %v", gotInfo, tt.wantInfo)
			}
		})
	}
}

func genPeers(devices []repo.DeviceInfo, numPeers int, order repo.PeerOrder, t *testing.T) []repo.PeerInfo {
	ret := make([]repo.PeerInfo, 0, numPeers)
	j := 0
	for i := 0; i < numPeers; i++ {
		now := time.Now().Add(time.Duration(i) * time.Minute)
		p := repo.PeerInfo{
			PublicKey:                   fmt.Sprint("pubkey", i),
			PresharedKey:                fmt.Sprint("sharekey", i),
			Endpoint:                    &net.UDPAddr{},
			PersistentKeepaliveInterval: time.Duration(i),
			AllowedIPs:                  []net.IPNet{mustResolveIPNet(fmt.Sprintf("1.2.3.%v/24", i%254))},
			DevicePublicKey:             devices[j%len(devices)].PublicKey,
		}
		if i%3 != 0 {
			p.LastHandshake = now.Unix()
			p.Name = fmt.Sprint("name", i)
		}

		ret = append(ret, p)
		j++
	}

	sort.Slice(ret, order.LessFunc(ret))
	return ret
}

func Test_sqliteRepository_ListPeers(t *testing.T) {
	type args struct {
		allPeers []repo.PeerInfo
		order    repo.PeerOrder
		offset   uint
		limit    uint
	}
	tests := []struct {
		name      string
		args      args
		wantData  []repo.PeerInfo
		wantTotal uint
		wantErr   bool
	}{
		{
			name: "offset & limit",
			args: args{
				allPeers: genPeers(genNewDevices(1), 10, repo.OrderNameAsc, t),
				order:    repo.OrderNameAsc,
				offset:   5,
				limit:    2,
			},
			wantData:  genPeers(genNewDevices(1), 10, repo.OrderNameAsc, t)[5:7],
			wantTotal: 10,
			wantErr:   false,
		},
		{
			name: "offset & no limit",
			args: args{
				allPeers: genPeers(genNewDevices(1), 20, repo.OrderNameDesc, t),
				order:    repo.OrderNameDesc,
				offset:   5,
			},
			wantData:  genPeers(genNewDevices(1), 20, repo.OrderNameDesc, t)[5:20],
			wantTotal: 20,
			wantErr:   false,
		},
		{
			name: "order by name asc",
			args: args{
				allPeers: genPeers(genNewDevices(1), 5, repo.OrderNameDesc, t),
				order:    repo.OrderNameAsc,
			},
			wantData:  genPeers(genNewDevices(1), 5, repo.OrderNameAsc, t),
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name: "order by name desc",
			args: args{
				allPeers: genPeers(genNewDevices(1), 5, repo.OrderNameAsc, t),
				order:    repo.OrderNameDesc,
			},
			wantData:  genPeers(genNewDevices(1), 5, repo.OrderNameDesc, t),
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name: "order by last handshake asc",
			args: args{
				allPeers: genPeers(genNewDevices(1), 5, repo.OrderLastHandshakeDesc, t),
				order:    repo.OrderLastHandshakeAsc,
			},
			wantData:  genPeers(genNewDevices(1), 5, repo.OrderLastHandshakeAsc, t),
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name: "order by last handshake desc",
			args: args{
				allPeers: genPeers(genNewDevices(1), 5, repo.OrderLastHandshakeAsc, t),
				order:    repo.OrderLastHandshakeDesc,
			},
			wantData:  genPeers(genNewDevices(1), 5, repo.OrderLastHandshakeDesc, t),
			wantTotal: 5,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mustCreateRepository(t)

			if err := s.UpdatePeers(tt.args.allPeers); err != nil {
				t.Error("ListPeers() updateError:", err)
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

			_ = s.Close()
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
			}
			if err := s.upsertPeers(tt.args.removeAll, tt.args.peers); (err != nil) != tt.wantErr {
				t.Errorf("upsertPeers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
