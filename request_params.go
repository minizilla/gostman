package gostman

import "net/url"

// Params sets request params.
func (r *Request) Params(f func(url.Values)) {
	f(r.params)
}
