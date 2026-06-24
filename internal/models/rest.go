package models

import "net/http"

// HTTPAuthProvider abstracts how we had authorization headers to the Request object
// Types:
//
// - BearerTokenProvider
// - BasicAuthProvider
// - NoAuthProvider
//
// Which allows us to modify the headers per our required implementation.
type HTTPAuthProvider interface {
	AddAuthHeaders(req *http.Request)
}
