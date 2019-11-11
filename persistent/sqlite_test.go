package persistent

import (
	"github.com/jmoiron/sqlx"
	"nz.cloudwalker/wireguard-webadmin/wg"
	"reflect"
	"testing"
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
			_, _, err := NewSqliteRepository(tt.args.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSqliteRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_createDb(t *testing.T) {
	type args struct {
		dsn string
	}
	tests := []struct {
		name    string
		args    args
		want    *sqlx.DB
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createDb(tt.args.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("createDb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createDb() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sqlRepository_GetNames(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		t   MetaType
		ids []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			got, err := s.GetNames(tt.args.t, tt.args.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNames() got = %v, want %v", got, tt.want)
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
		want    []wg.DeviceConfig
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sqlRepository{
				DB: tt.fields.DB,
			}
			got, err := s.ListDevices()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListDevices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListDevices() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sqlRepository_RemoveDevices(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		ids []string
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

func Test_sqlRepository_RemoveNames(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		t   MetaType
		ids []string
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
			if err := s.RemoveNames(tt.args.t, tt.args.ids); (err != nil) != tt.wantErr {
				t.Errorf("RemoveNames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sqlRepository_SaveDevices(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		devices []wg.DeviceConfig
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

func Test_sqlRepository_SetNames(t *testing.T) {
	type fields struct {
		DB *sqlx.DB
	}
	type args struct {
		t     MetaType
		names map[string]string
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
			if err := s.SetNames(tt.args.t, tt.args.names); (err != nil) != tt.wantErr {
				t.Errorf("SetNames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
