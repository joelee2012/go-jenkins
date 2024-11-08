package jenkins

import (
	"io"
	"net/http"
)

type Requester interface {
	Request(method, entry string, body io.Reader) (*http.Response, error)
}
