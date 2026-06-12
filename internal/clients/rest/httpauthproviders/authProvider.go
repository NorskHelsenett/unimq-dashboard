package httpauthproviders

import "net/http"

type BearerTokenProvider struct {
	bearerToken string
}

func NewBearerTokenProvider(bearerToken string) *BearerTokenProvider {
	return &BearerTokenProvider{bearerToken: bearerToken}
}
func (a *BearerTokenProvider) AddAuthHeaders(req *http.Request) {

	req.Header.Add("Authorization", "bearer "+a.bearerToken)
}
