package rest

import "net/http"

func (rc *RestClient) Delete(url string, out any, params ...Params) (int, error) {
	return rc.request(http.MethodDelete, url, nil, &out, params)
}
func (rc *RestClient) Get(url string, out any, params ...Params) (int, error) {
	return rc.request(http.MethodGet, url, nil, out, params)
}

func (rc *RestClient) Patch(url string, body any, out any, params ...Params) (int, error) {
	return rc.request(http.MethodPatch, url, body, &out, params)
}

func (rc *RestClient) Post(url string, body any, out any, params ...Params) (int, error) {
	return rc.request(http.MethodPost, url, body, out, params)
}

func (rc *RestClient) Put(url string, body any, out any, params ...Params) (int, error) {
	return rc.request(http.MethodPut, url, body, out, params)
}
