package api

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"nz.cloudwalker/wireguard-webadmin/repo"
	"strconv"
)

type httpPeer struct {
	PublicKey string `json:"public_key"`
	Name      string `json:"name"`
}

type paginatedResult struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
	Total uint32      `json:"total"`
}

func (p *httpPeer) FromPeerInfo(info repo.PeerInfo) {
	p.PublicKey = info.PublicKey.String()
	p.Name = info.Name
}

type httpApi struct {
	repo.Repository
}

func (api httpApi) ListPeers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var result paginatedResult
	var err error

	defer func() {
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(500)
		}

		_ = json.NewEncoder(w).Encode(result)
	}()

	var offset uint64
	var limit uint64

	offsetParam := r.URL.Query()["offset"]
	if len(offsetParam) > 0 {
		if offset, err = strconv.ParseUint(offsetParam[0], 10, 32); err != nil {
			result.Error = "Offset is an invalid value"
			return
		}
	}

	limitParam := r.URL.Query()["limit"]
	if len(limitParam) > 0 {
		if limit, err = strconv.ParseUint(limitParam[0], 10, 32); err != nil {
			result.Error = "Limit is an invalid value"
			return
		}
	}

	var peerInfo []repo.PeerInfo

	if result.Total, err = api.Repository.ListAllPeers(&peerInfo, uint32(offset), uint32(limit)); err != nil {
		result.Error = err.Error()
		return
	}

	var peers []httpPeer
	for _, info := range peerInfo {
		var p httpPeer
		p.FromPeerInfo(info)

		peers = append(peers, p)
	}

	result.Data = peers
}

func NewHttpApi(repository repo.Repository) (http.Handler, error) {
	api := httpApi{repository}

	r := httprouter.New()
	r.GET("/peers", api.ListPeers)

	return r, nil
}
