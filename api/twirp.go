package api

import (
	"net/http"
	"nz.cloudwalker/wireguard-webadmin/api/twirp"
	"nz.cloudwalker/wireguard-webadmin/repo"
)

type twirpService struct {
}

func NewTwirpService(repository repo.Repository) (http.Handler, error) {
	return twirp.NewWireGuardServiceServer(nil, nil), nil
}
