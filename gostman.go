package gostman

import (
	_ "embed"
	"net/http"
	neturl "net/url"
	"strings"
	"testing"
)

//go:embed VERSION
var gostmanVersion string

// Gostman represents an API development set.
type Gostman struct {
	t *testing.T
}

// New returns new Gostman.
func New(t *testing.T) *Gostman {
	return &Gostman{
		t: t,
	}
}

// Request run the request.
func (gm *Gostman) Request(name, method, url string, fn func(*Request)) {
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

// GET run the request using GET method.
func (gm *Gostman) GET(name, url string, fn func(*Request)) {
	gm.Request(name, http.MethodGet, url, fn)
}

// POST run the request using POST method.
func (gm *Gostman) POST(name, url string, fn func(*Request)) {
	gm.Request(name, http.MethodPost, url, fn)
}

// PUT run the request using PUT method.
func (gm *Gostman) PUT(name, url string, fn func(*Request)) {
	gm.Request(name, http.MethodPut, url, fn)
}

// PATCH run the request using PATCH method.
func (gm *Gostman) PATCH(name, url string, fn func(*Request)) {
	gm.Request(name, http.MethodPatch, url, fn)
}

// DELETE run the request using DELETE method.
func (gm *Gostman) DELETE(name, url string, fn func(*Request)) {
	gm.Request(name, http.MethodDelete, url, fn)
}

// HEAD run the request using HEAD method.
func (gm *Gostman) HEAD(name, url string, fn func(*Request)) {
	gm.Request(name, http.MethodHead, url, fn)
}

// OPTIONS run the request using OPTIONS method.
func (gm *Gostman) OPTIONS(name, url string, fn func(*Request)) {
	gm.Request(name, http.MethodOptions, url, fn)
}

// SetV sets a variable.
func (gm *Gostman) SetV(name, val string) {
	runtime.setEnvVar(name, val)
}

// V returns a variable.
func (gm *Gostman) V(name string) string {
	return runtime.envVar(name)
}
