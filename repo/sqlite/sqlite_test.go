package sqlite

import (
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
