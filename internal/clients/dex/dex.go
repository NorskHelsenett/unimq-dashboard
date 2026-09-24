package dex

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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

// oauthStateCookieName holds the per-request CSRF state for the authorization
// code flow. It is short-lived and consumed exactly once, in the callback.
const oauthStateCookieName = "unimq_oauth_state"

// oauthStateTTL bounds how long a login attempt may stay in flight.
const oauthStateTTL = 10 * time.Minute

var ErrOAuthStateMismatch = errors.New("oauth state mismatch")

// newOAuthState returns 256 bits of cryptographically random, URL-safe state.
func newOAuthState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate oauth state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// isSecureRequest reports whether the original client request used TLS. The
// proxy header is needed because TLS normally terminates at the ingress, so
// r.TLS is nil even on an https:// request.
func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func setStateCookie(w http.ResponseWriter, r *http.Request, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   int(oauthStateTTL.Seconds()),
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		// Lax still travels on the top-level GET redirect back from Dex,
		// while keeping the cookie off cross-site subresource requests.
		SameSite: http.SameSiteLaxMode,
	})
}

func clearStateCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

// verifyState compares the state echoed back by Dex against the one this
// server issued. The cookie is single-use: callers must clear it either way.
func verifyState(r *http.Request) error {
	got := r.URL.Query().Get("state")
	if got == "" {
		return fmt.Errorf("%w: no state in callback", ErrOAuthStateMismatch)
	}

	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || cookie.Value == "" {
		return fmt.Errorf("%w: no state cookie on request", ErrOAuthStateMismatch)
	}

	if subtle.ConstantTimeCompare([]byte(got), []byte(cookie.Value)) != 1 {
		return fmt.Errorf("%w: callback state does not match issued state", ErrOAuthStateMismatch)
	}

	return nil
}

// @Summary		Redirect to Dex for authentication
// @Description	Redirects the user to the Dex server for authentication
// @Tags			Authentication
// @Produce		json
// @Success		302	{string}	string	"redirect"
// @Failure		500	{object}	httpsuite.ErrorResponse
// @Router			/v1/login/redirect [get]
func (d *DexClient) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	state, err := newOAuthState()
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("internal server error"),
			httpsuite.WithInternalErrorMessage("failed to generate oauth state"),
		)
		return
	}

	setStateCookie(w, r, state)

	authURL := d.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	http.Redirect(w, r, authURL, http.StatusFound)
}

// @Summary		Handle Dex OAuth callback
// @Description	Handles the OAuth callback from Dex and exchanges the code for a token
// @Tags			Authentication
// @Produce		json
// @Param			code	query		string	true	"Authorization code"
// @Param			state	query		string	true	"CSRF state issued by /v1/login/redirect"
// @Success		302		{string}	string	"redirect"
// @Failure		400		{object}	httpsuite.ErrorResponse
// @Failure		500		{object}	httpsuite.ErrorResponse
// @Router			/v1/login/callback [get]
func (d *DexClient) OauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	clearStateCookie(w, r)

	if err := verifyState(r); err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithExternalErrorMessage("bad request"),
			httpsuite.WithInternalErrorMessage(err.Error()),
		)
		return
	}

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
