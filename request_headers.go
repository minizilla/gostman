package gostman

import "net/http"

// Headers sets request headers.
func (r *Request) Headers(f func(http.Header)) {
	f(r.headers)
}

func recomendedHeader() http.Header {
	headers := make(http.Header)
	headers.Add("Host", "<calculated when request is sent>")
	headers.Add("User-Agent", "Gostman/"+gostmanVersion)
	headers.Add("Accept", "*/*")
	// headers.Add("Accept-Encoding", "gzip, deflate, br") // TODO
	headers.Add("Connection", "keep-alive")
	return headers
}
