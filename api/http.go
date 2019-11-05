package api

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"nz.cloudwalker/wireguard-webadmin/repo"
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

func writeHttpResult(data interface{}, err error, writer http.ResponseWriter) {
	var r result
	if err != nil {
		var ok bool

		if r.Error, ok = err.(*displayableError); !ok {
			r.Error = wrapError(err)
		}
	} else {
		r.Data = data
	}

	writer.Header().Set("Content-Type", "application/json")
	if r.Error != nil {
		writer.WriteHeader(r.Error.StatusCode)
	}
	_ = json.NewEncoder(writer).Encode(r)
}

func NewHttpApi(repository repo.Repository) (http.Handler, error) {
	api := httpApi{Repo: repository}
	r := httprouter.New()
	r.PanicHandler = func(writer http.ResponseWriter, request *http.Request, i interface{}) {
		if err, ok := i.(error); ok {
			writeHttpResult(nil, err, writer)
		} else {
			writeHttpResult(nil, newError(unknownError), writer)
		}
	}

	r.GET("/peers", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var offset, limit uint32
		mustScanQueryParameter(&offset, request, "offset", "%d", true)
		mustScanQueryParameter(&limit, request, "limit", "%d", true)
		if r, err := api.ListPeers(offset, limit); err != nil {
			panic(err)
		} else {
			writeHttpResult(r, nil, writer)
		}
	})
	return r, nil
}
