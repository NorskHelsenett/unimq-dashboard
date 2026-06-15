// rest is a generic rest client that can be set up towards any REST API.
// The implementation is generic toward any API without a specific focus on a product.
package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	neturl "net/url"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest/httpauthproviders"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

const userAgent = "UniMQ"

type Params struct {
	Value any
	Key   HTTPRestClientOpts
}
type HTTPRestClientOpts string

const (
	HTTPTransportClientOptsNoAuth  HTTPRestClientOpts = "NOAUTH"
	HTTPTransportClientOptsHeaders HTTPRestClientOpts = "HEADERS"
	HTTPTransportClientOptsQuery   HTTPRestClientOpts = "QUERY"
	HTTPTransportClientTimeout     HTTPRestClientOpts = "TIMEOUT"
)

type RestClient struct {
	AuthProvider models.HTTPAuthProvider
	Context      context.Context
	HTTPClient   *http.Client
	BaseURL      string
}
type (
	Config struct {
		authProvider models.HTTPAuthProvider
		context      context.Context
		Username     string
		password     string
		Timeout      time.Duration
	}
	ConfigOption func(*Config)
)

func newConfig() *Config {
	return &Config{
		authProvider: httpauthproviders.NewNoAuthProvider(),
		context:      context.TODO(),
		Username:     "",
		password:     "",
		//nolint:mnd // sane default.
		Timeout: time.Second * 30,
	}
}

// WithContext adds a context.

func WithContext(ctx context.Context) ConfigOption {
	return func(c *Config) {
		c.context = ctx
	}
}

// WithAuthProvider adds a custom AuthProvider.
func WithAuthProvider(authProvider models.HTTPAuthProvider) ConfigOption {
	return func(c *Config) {
		c.authProvider = authProvider
	}
}
func WithUsername(username string) ConfigOption {
	return func(c *Config) {
		c.Username = username
	}
}
func WithPassword(password string) ConfigOption {
	return func(c *Config) {
		c.password = password
	}
}
func WithTimeout(timeout int) ConfigOption {
	return func(c *Config) {
		c.Timeout = time.Second * time.Duration(timeout)
	}
}
func NewRestClient(baseURL string, opts ...ConfigOption) (*RestClient, error) {
	config := newConfig()
	for _, opt := range opts {
		opt(config)
	}
	restClient := RestClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: config.Timeout,
		},
		AuthProvider: config.authProvider,
	}
	if restClient.BaseURL == "" {
		return nil, errors.New("baseUrl is empty")
	}
	return &restClient, nil
}
func (r *Config) AddAuthHeaders(req *http.Request) {
	r.authProvider.AddAuthHeaders(req)
}
func (r *RestClient) request(method string, url string, body *[]byte, out any, restclientParam []Params) (int, error) {

	u, err := neturl.ParseRequestURI(url)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("uable to parse relative url %v. %w", u, err)
	}
	endpoint, err := neturl.ParseRequestURI(r.BaseURL + u.String())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("uable to parse absolute url %v. %w", endpoint, err)
	}
	var req *http.Request
	if body == nil {
		req, err = http.NewRequestWithContext(r.Context, method, endpoint.String(), nil)
	} else {
		req, err = http.NewRequestWithContext(r.Context, method, endpoint.String(), bytes.NewReader(*body))
	}
	if err != nil {
		return http.StatusInternalServerError,
			fmt.Errorf("uable to create new request of method %v with endpoint %v. %w",
				http.MethodPost,
				endpoint,
				err,
			)
	}
	r.addCommonHeaders(req)
	r.processParams(restclientParam, req)
	slog.DebugContext(r.Context, "making request", "method", method, "url", endpoint.String())

	//nolint:gosec // if this causes an exploitation there are bigger issues.
	resp, err := r.HTTPClient.Do(req)
	if err != nil && resp != nil {
		berr := resp.Body.Close()
		return http.StatusInternalServerError, fmt.Errorf("uable to Do request. %w, %w", err, berr)
	}
	if err != nil && resp == nil {
		return http.StatusInternalServerError, fmt.Errorf("uable to Do request. %w", err)
	}
	defer func() {
		berr := resp.Body.Close()
		err = errors.Join(err, berr)
		if err != nil {
			slog.ErrorContext(r.Context, "error closing response body", "error", err)
		}
	}()
	slog.DebugContext(r.Context, "received response", "method", method, "status_code", resp.StatusCode)
	badStatusCodeCeiling := 399
	if resp.StatusCode > badStatusCodeCeiling {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, err
		}
		err = errors.New(string(bodyBytes))
		if body != nil {
			err = errors.Join(err, errors.New(string(*body)))
		}
		return resp.StatusCode, err
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("uable to read body of response. %w", err)

	}
	return resp.StatusCode, nil
}
func (r *RestClient) addCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Add("Accept", `application/json`)
	req.Header.Add("Content-Type", `application/json`)
}
func (r *RestClient) processParams(params []Params, req *http.Request) {
	auth := true
	for _, param := range params {
		switch param.Key {
		case HTTPTransportClientOptsNoAuth:
			auth = false
		case HTTPTransportClientOptsHeaders:
			params, ok := param.Value.(map[string]string)
			if ok {
				for key, value := range params {
					req.Header.Add(key, value)
				}
			}
		case HTTPTransportClientOptsQuery:
			q := req.URL.Query()
			params, ok := param.Value.(map[string]string)
			if ok {
				for key, value := range params {
					q.Add(key, value)
				}
			}
			req.URL.RawQuery = q.Encode()
		case HTTPTransportClientTimeout:
		}
	}
	if auth {
		r.AuthProvider.AddAuthHeaders(req)
	}
}
