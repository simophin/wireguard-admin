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
	if r.Error != nil && r.Error.StatusCode >= 100 {
		writer.WriteHeader(r.Error.StatusCode)
	} else if r.Error != nil {
		writer.WriteHeader(500)
	}

	_ = json.NewEncoder(writer).Encode(r)
}

func getQueryParams(r *http.Request, n string, defaultValue string) string {
	values := r.URL.Query()[n]
	var v string
	if len(values) > 0 {
		v = values[0]
	}

	if len(v) == 0 {
		return defaultValue
	}

	return v
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
		offset, err := strconv.ParseUint(getQueryParams(request, "offset", "0"), 10, 32)
		if err != nil {
			panic(&displayableError{
				Name:        badRequest,
				Description: "Parameter offset is not valid",
				StatusCode:  400,
			})
		}
		limit, err := strconv.ParseUint(getQueryParams(request, "limit", "0"), 10, 32)
		if err != nil {
			panic(&displayableError{
				Name:        badRequest,
				Description: "Parameter limit is not valid",
				StatusCode:  400,
			})
		}

		if r, err := api.ListPeers(uint32(offset), uint32(limit)); err != nil {
			panic(err)
		} else {
			writeHttpResult(r, nil, writer)
		}
	})
	return r, nil
}
