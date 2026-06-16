package httpauthproviders

import (
	"encoding/base64"
	"net/http"
)

type BasicAuthProvider struct {
	username string
	password string
}

func NewBasicAuthProvider(username, password string) *BasicAuthProvider {
	return &BasicAuthProvider{
		username: username,
		password: password,
	}
}
func (a *BasicAuthProvider) AddAuthHeaders(req *http.Request) {

	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(a.username+":"+a.password)))
}
