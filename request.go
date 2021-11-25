package gostman

import (
	"io"
	"net/http"
	neturl "net/url"
	"strconv"
	"testing"

	log "github.com/sirupsen/logrus"
)

type Request struct {
	t       *testing.T
	client  http.Client
	method  string
	url     string
	params  neturl.Values
	headers http.Header
	body    io.Reader
}

func recomendedHeader() http.Header {
	headers := make(http.Header)
	headers.Add("Host", "<calculated when request is sent>")
	headers.Add("User-Agent", "Gostman/"+gostmanVersion)
	headers.Add("Accept", "*/*")
	headers.Add("Accept-Encoding", "gzip, deflate, br")
	headers.Add("Connection", "keep-alive")
	return headers
}

// Params sets request params.
func (r *Request) Params(f func(neturl.Values)) {
	f(r.params)
}

// Headers sets request headers.
func (r *Request) Headers(f func(http.Header)) {
	f(r.headers)
}

// Body sets request body.
func (r *Request) Body(body io.Reader, contentType string, contentLength int) {
	r.body = body
	r.headers.Set("Content-Type", contentType)
	r.headers.Set("Content-Length", strconv.Itoa(contentLength))
}

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
