package gostman

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// BodyJSON creates request body by marshaling v using JSON.
// Returns io.Reader, Content-Type of application/json, and its length.
func BodyJSON(v interface{}) (io.Reader, string, int) {
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(v); err != nil {
		logrus.Fatal(err)
	}
	return &buff, "application/json", buff.Len()
}

// BodyFormURLEncoded creates request body by encoding url values.
// Returns io.Reader, Content-Type of application/x-www-form-urlencoded, and its length.
func BodyFormURLEncoded(f func(url.Values)) (io.Reader, string, int) {
	v := make(url.Values)
	f(v)
	form := v.Encode()
	return strings.NewReader(form), "application/x-www-form-urlencoded", len(form)
}

// Body Text creates request body using raw text.
// Returns io.Reader, Content-Type of text/plain, and its length.
func BodyText(s string) (io.Reader, string, int) {
	return strings.NewReader(s), "text/plain", len(s)
}
