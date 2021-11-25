package gostman

import (
	_ "embed"
	"net/http"
	neturl "net/url"
	"strings"
	"testing"
)

var (
	//go:embed VERSION
	gostmanVersion string
	gostman        gostmanRuntime
)

type Gostman struct {
	t *testing.T
}

func New(t *testing.T) *Gostman {
	gostman.initOnce()

	gm := &Gostman{
		t: t,
	}

	gostman.g[t.Name()] = gm
	return gm
}

func (gm *Gostman) AddRequest(name, method, url string, fn func(*Request)) {
	gm.t.Run(strings.ReplaceAll(name, " ", ""), func(t *testing.T) {
		fn(&Request{
			t:       t,
			method:  method,
			url:     url,
			params:  make(neturl.Values),
			headers: recomendedHeader(),
		})
	})
}

func (gm *Gostman) GET(name, url string, fn func(*Request)) {
	gm.AddRequest(name, http.MethodGet, url, fn)
}

func (gm *Gostman) POST(name, url string, fn func(*Request)) {
	gm.AddRequest(name, http.MethodPost, url, fn)
}

func (gm *Gostman) PUT(name, url string, fn func(*Request)) {
	gm.AddRequest(name, http.MethodPut, url, fn)
}

func (gm *Gostman) PATCH(name, url string, fn func(*Request)) {
	gm.AddRequest(name, http.MethodPatch, url, fn)
}

func (gm *Gostman) DELETE(name, url string, fn func(*Request)) {
	gm.AddRequest(name, http.MethodDelete, url, fn)
}

func (gm *Gostman) HEAD(name, url string, fn func(*Request)) {
	gm.AddRequest(name, http.MethodHead, url, fn)
}

func (gm *Gostman) OPTIONS(name, url string, fn func(*Request)) {
	gm.AddRequest(name, http.MethodOptions, url, fn)
}
