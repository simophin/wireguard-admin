package api

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"nz.cloudwalker/wireguard-webadmin/repo"
	"strconv"
)

type httpApi struct {
	Repo repo.Repository
}

func (api httpApi) ListPeers(offset uint32, limit uint32) (result paginatedResult, err error) {
	var peerInfo []repo.PeerInfo

	if peerInfo, result.Total, err = api.Repo.ListAllPeers(offset, limit); err != nil {
		return
	}

	var peers []peer
	var p peer

	for _, info := range peerInfo {
		p.FromPeerInfo(info)
		peers = append(peers, p)
	}

	result.Contents = peers
	return
}

func NewHttpApi(repository repo.Repository) (http.Handler, error) {
	api := httpApi{Repo: repository}
	r := httprouter.New()
	r.GET("/peers", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {

	})
	return r, nil
}
