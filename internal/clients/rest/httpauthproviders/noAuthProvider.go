package httpauthproviders

import "net/http"

type NoAuthProvider struct {
}

func NewNoAuthProvider() *NoAuthProvider {
	return &NoAuthProvider{}
}

func (n NoAuthProvider) AddAuthHeaders(_ *http.Request) {

}
