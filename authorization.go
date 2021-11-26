package gostman

import (
	"encoding/base64"
	"net/http"
)

// Authorization sets request authorization.
func (r *Request) Authorization(f func(http.Header)) {
	f(r.headers)
}

// AuthBasic sets auth using basic username and password.
func AuthBasic(username, password string) func(http.Header) {
	return func(h http.Header) {
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		h.Set("Authorization", "Basic "+auth)
	}
}

// AuthBearer sets auth using bearer token.
func AuthBearer(token string) func(http.Header) {
	return func(h http.Header) {
		h.Set("Authorization", "Bearer "+token)
	}
}

// AuthAPIKey sets auth using api key.
func AuthAPIKey(key, val string) func(http.Header) {
	return func(h http.Header) {
		h.Set(key, val)
	}
}
