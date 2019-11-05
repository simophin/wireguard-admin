package api

import (
	"context"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net/http"
	"nz.cloudwalker/wireguard-webadmin/api/rpc"
	"nz.cloudwalker/wireguard-webadmin/repo"
)

type twirpService struct {
	repo repo.Repository
}

func (t twirpService) UpdatePeers(ctx context.Context, r *rpc.UpdatePeerRequest) (*rpc.UpdatePeerResponse, error) {
	var peerKeys []string
	for _, k := range r.Peers {
		peerKeys = append(peerKeys, k.PublicKey)
	}

	peersMap := make(map[string]repo.PeerInfo)
	peers, err := t.repo.GetPeers(peerKeys)
	if err != nil {
		return nil, err
	}

	for _, p := range peers {
		peersMap[p.PublicKey] = p
	}

	for _, p := range r.Peers {
		peer := peersMap[p.PublicKey]
		peer.PublicKey = p.PublicKey
		peer.Name = p.Name
	}

	peers = peers[:0]
	for _, p := range peersMap {
		peers = append(peers, p)
	}
	if err := t.repo.UpdatePeers(peers); err != nil {
		return nil, err
	}

	return &rpc.UpdatePeerResponse{
		NumUpdated: uint32(len(peers)),
	}, nil

}

func (t twirpService) RemovePeers(ctx context.Context, r *rpc.RemovePeerRequest) (*rpc.RemovePeerResponse, error) {
	var publicKeys []wgtypes.Key
	for _, k := range r.PublicKeys {
		if k, err := wgtypes.ParseKey(k); err != nil {
			return nil, err
		} else {
			publicKeys = append(publicKeys, k)
		}
	}

	return &rpc.RemovePeerResponse{
		NumRemoved: uint32(len(publicKeys)),
	}, nil
}

func (t twirpService) ListPeers(ctx context.Context, r *rpc.ListPeerRequest) (*rpc.ListPeerResponse, error) {
	if peerInfos, total, err := t.repo.ListAllPeers(r.Offset, r.Limit); err != nil {
		return nil, err
	} else {
		var ret rpc.ListPeerResponse
		ret.Total = total
		for _, peerInfo := range peerInfos {
			var p rpc.Peer
			p.Name = peerInfo.Name
			p.PublicKey = peerInfo.PublicKey
			ret.Peers = append(ret.Peers, &p)
		}
		return &ret, nil
	}
}

func NewTwirpService(repository repo.Repository) (http.Handler, error) {
	return rpc.NewWireGuardServiceServer(&twirpService{repository}, nil), nil
}
