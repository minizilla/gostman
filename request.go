package gostman

import (
	"io"
	"net/http"
	neturl "net/url"
	"testing"

	log "github.com/sirupsen/logrus"
)

// Request contains all necessary thing to create a Gostman request.
type Request struct {
	t       *testing.T
	client  http.Client
	method  string
	url     string
	params  neturl.Values
	headers http.Header
	body    io.Reader
}

// Send sends a request. The testing can be done inside f.
func (r *Request) Send(f func(t *testing.T, res *http.Response)) {
	url := r.url + "?" + r.params.Encode()
	req, err := http.NewRequest(r.method, url, r.body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header = r.headers

	res, err := r.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	f(r.t, res)
}
