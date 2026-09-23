package dex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"golang.org/x/oauth2"
)

type DexClient struct {
	url      string
	Config   *oauth2.Config
	Verifier *oidc.IDTokenVerifier
}

func NewDexClient(ctx context.Context, config *config.OIDCConfig) (*DexClient, error) {

	provider, err := oidc.NewProvider(ctx, config.OIDCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientSecret: config.OIDCClientSecret,
		ClientID:     config.OIDCClientID,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"},
		RedirectURL:  config.OIDCRedirectURL,
	}

	return &DexClient{
		url:    config.OIDCURL,
		Config: &oauth2Config,
		Verifier: provider.Verifier(&oidc.Config{
			ClientID: config.OIDCClientID,
		}),
	}, nil
}

func (d *DexClient) Ping(ctx context.Context) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.url+"/healthz", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping Dex server: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.ErrorContext(context.Background(), "error closing response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dex server returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

func (d *DexClient) ValidateToken(ctx context.Context, token string) (*oidc.IDToken, error) {
	idToken, err := d.Verifier.Verify(ctx, token)
	if err != nil {
		slog.ErrorContext(ctx, "failed to verify ID Token", "error", err)
		return nil, fmt.Errorf("failed to verify ID Token: %w", err)
	}

	return idToken, nil
}

const bearerPrefix = "Bearer "

// Authorization is a middleware that checks for a valid OIDC token in the Authorization header.
// If the token is valid, it adds the claims to the request context.
func (d *DexClient) Authorization() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := d.getBearerToken(r)
			if err != nil {
				switch {

				case errors.Is(err, ErrBearerTokenMissing):
					httpsuite.WriteJSONError(w,
						http.StatusUnauthorized,
						httpsuite.WithExternalErrorMessage("unauthorized"),
						httpsuite.WithInternalErrorMessage("Authorization header missing"),
						httpsuite.WithError(err),
					)
					return
				case errors.Is(err, ErrBearerTokenInvalid):
					httpsuite.WriteJSONError(w,
						http.StatusUnauthorized,
						httpsuite.WithExternalErrorMessage("unauthorized"),
						httpsuite.WithInternalErrorMessage("invalid token"),
						httpsuite.WithError(err),
					)
					return
				case errors.Is(err, ErrBearerTokenExpired):
					httpsuite.WriteJSONError(w,
						http.StatusUnauthorized,
						httpsuite.WithExternalErrorMessage("unauthorized"),
						httpsuite.WithInternalErrorMessage("token expired"),
						httpsuite.WithError(err),
					)
					return
				default:
					slog.ErrorContext(r.Context(), "failed to get bearer token", "error", err)
					httpsuite.WriteJSONError(w,
						http.StatusInternalServerError,
						httpsuite.WithExternalErrorMessage("internal server error"),
						httpsuite.WithInternalErrorMessage("failed to get bearer token"),
						httpsuite.WithError(err),
					)
					return

				}
			}

			var claims map[string]any

			if err := token.Claims(&claims); err != nil {
				httpsuite.WriteJSONError(w,
					http.StatusInternalServerError,
					httpsuite.WithError(err),
					httpsuite.WithExternalErrorMessage("unauthorized"),
					httpsuite.WithInternalErrorMessage("failed to parse claims"),
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				httpsuite.ClaimsContextKey,
				claims,
			)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// @Summary		Redirect to Dex for authentication
// @Description	Redirects the user to the Dex server for authentication
// @Tags			Authentication
// @Produce		json
// @Success		302	{string}	string	"redirect"
// @Router			/v1/login/redirect [get]
func (d *DexClient) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	authURL := d.Config.AuthCodeURL("state", oauth2.AccessTypeOffline)

	http.Redirect(w, r, authURL, http.StatusFound)
}

// @Summary		Handle Dex OAuth callback
// @Description	Handles the OAuth callback from Dex and exchanges the code for a token
// @Tags			Authentication
// @Produce		json
// @Param			code	query		string	true	"Authorization code"
// @Success		302		{string}	string	"redirect"
// @Failure		400		{object}	httpsuite.ErrorResponse
// @Failure		500		{object}	httpsuite.ErrorResponse
// @Router			/v1/login/callback [get]
func (d *DexClient) OauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get the code from the query parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithExternalErrorMessage("bad request"),
			httpsuite.WithInternalErrorMessage("code query parameter missing"),
		)
		return
	}

	// Exchange the code for a token
	token, err := d.Config.Exchange(r.Context(), code)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("internal server error"),
			httpsuite.WithInternalErrorMessage("failed to exchange code for token"),
		)
		return
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithExternalErrorMessage("internal server error"),
			httpsuite.WithInternalErrorMessage("id_token not found in token response"),
		)
		return
	}

	// Redirect to the frontend with the token as a query parameter
	redirectURL := fmt.Sprintf("/?token=%s", idToken)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

type LoginRequest struct {
	Username string `json:"username" example:"olanordmann@test.com"`
	Password string `json:"password" example:"password"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (lr *LoginRequest) Validate() error {
	if lr.Username == "" {
		return fmt.Errorf("username is required")
	}
	if lr.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

// @Summary		Login with Dex
// @Description	Login with Dex using username and password
// @Tags			Authentication
// @Accept			json
// @Produce		json
// @Param			loginRequest	body		LoginRequest	true	"Login Request"
// @Success		200				{object}	TokenResponse
// @Failure		400				{object}	httpsuite.ErrorResponse
// @Failure		401				{object}	httpsuite.ErrorResponse
// @Failure		500				{object}	httpsuite.ErrorResponse
// @Router			/v1/login [post]
func (d *DexClient) LoginHandler(w http.ResponseWriter, r *http.Request) {

	loginReq := &LoginRequest{}
	err := httpsuite.ReadResponse(w, r, loginReq)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("bad request"),
			httpsuite.WithInternalErrorMessage("failed to read login request"),
		)
		return
	}

	err = loginReq.Validate()
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("bad request"),
			httpsuite.WithInternalErrorMessage("invalid login request"),
		)
		return
	}

	form := url.Values{}
	form.Add("username", loginReq.Username)
	form.Add("password", loginReq.Password)
	form.Add("client_id", d.Config.ClientID)
	form.Add("client_secret", d.Config.ClientSecret)
	form.Add("grant_type", "password")
	form.Add("scope", strings.Join(d.Config.Scopes, " "))
	request, err := http.NewRequestWithContext(r.Context(), http.MethodPost, d.url+"/token", strings.NewReader(form.Encode()))
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("internal server error"),
			httpsuite.WithInternalErrorMessage("failed to create request to Dex"),
		)
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("internal server error"),
			httpsuite.WithInternalErrorMessage("failed to send request to Dex"),
		)
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.ErrorContext(r.Context(), "error closing response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		httpsuite.WriteJSONError(w,
			http.StatusUnauthorized,
			httpsuite.WithExternalErrorMessage("unauthorized"),
			httpsuite.WithInternalErrorMessage(fmt.Sprintf("Dex returned status code %d", resp.StatusCode)),
		)
		return
	}

	var tokenResponse TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("internal server error"),
			httpsuite.WithInternalErrorMessage("failed to decode Dex response"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "login successful", http.StatusOK, &tokenResponse)
}

var (
	ErrBearerTokenMissing = fmt.Errorf("bearer token missing")
	ErrBearerTokenInvalid = fmt.Errorf("bearer token invalid")
	ErrBearerTokenExpired = fmt.Errorf("bearer token expired")
)

func (d *DexClient) getBearerToken(r *http.Request) (*oidc.IDToken, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return nil, ErrBearerTokenMissing
	}

	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return nil, fmt.Errorf("%w, token does not have the correct 'Bearer ' prefix", ErrBearerTokenInvalid)
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return nil, fmt.Errorf("%w, token empty after trimming prefix", ErrBearerTokenInvalid)
	}

	idToken, err := d.ValidateToken(r.Context(), token)
	if err != nil {
		return nil, fmt.Errorf("%w, token did not validate. %w", ErrBearerTokenInvalid, err)
	}

	if idToken.Expiry.Before(time.Now()) {
		return nil, ErrBearerTokenExpired
	}

	return idToken, nil

}
