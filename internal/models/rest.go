package models

import "net/http"

// HTTPAuthProvider abstracts how we had authorization headers to the Request object
// Types:
//
// - BearerTokenProvider
//
// - NoAuthProvider.
type HTTPAuthProvider interface {
	AddAuthHeaders(req *http.Request)
}
