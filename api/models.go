package api

import (
	"nz.cloudwalker/wireguard-webadmin/repo"
)

type peer struct {
	PublicKey string `json:"public_key"`
	Name      string `json:"name"`
}

type errorName string

type displayableError struct {
	Cause       error     `json:"-"`
	Name        errorName `json:"name"`
	Description string    `json:"description"`
	StatusCode  int       `json:"-"`
}

func (d displayableError) Error() string {
	if len(d.Description) > 0 {
		return d.Description
	}
	return string(d.Name)
}

const (
	unknownError errorName = "unknown"
	badRequest   errorName = "bad_request"
)

func newError(name errorName) *displayableError {
	return &displayableError{
		Name: name,
	}
}

func wrapError(cause error) *displayableError {
	if ret, ok := cause.(*displayableError); ok {
		return ret
	}

	return &displayableError{
		Cause: cause,
		Name:  unknownError,
	}
}

type result struct {
	Data  interface{}       `json:"data,omitempty"`
	Error *displayableError `json:"error,omitempty"`
}

type paginatedResult struct {
	Contents interface{} `json:"contents"`
	Total    uint32      `json:"total"`
}

func (p *peer) FromPeerInfo(info repo.PeerInfo) {
	p.PublicKey = info.PublicKey.String()
	p.Name = info.Name
}
