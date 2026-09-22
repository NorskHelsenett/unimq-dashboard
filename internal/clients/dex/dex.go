package dex

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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
		ClientID:    config.OIDCClientID,
		Endpoint:    provider.Endpoint(),
		Scopes:      []string{oidc.ScopeOpenID, "profile", "email", "groups"},
		RedirectURL: config.OIDCRedirectURL,
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
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				httpsuite.WriteJSONError(w,
					http.StatusUnauthorized,
					httpsuite.WithExternalErrorMessage("unauthorized"),
					httpsuite.WithInternalErrorMessage("Authorization header missing"),
				)
				return
			}

			if !strings.HasPrefix(authHeader, bearerPrefix) {
				httpsuite.WriteJSONError(w,
					http.StatusUnauthorized,
					httpsuite.WithExternalErrorMessage("unauthorized"),
					httpsuite.WithInternalErrorMessage("authorization header does not start with 'Bearer '"),
				)
				return
			}

			token := strings.TrimPrefix(authHeader, bearerPrefix)
			if token == "" {
				httpsuite.WriteJSONError(w,
					http.StatusUnauthorized,
					httpsuite.WithExternalErrorMessage("unauthorized"),
					httpsuite.WithInternalErrorMessage("token missing in Authorization header"),
				)
				return
			}

			idToken, err := d.ValidateToken(r.Context(), token)
			if err != nil {
				httpsuite.WriteJSONError(w,
					http.StatusUnauthorized,
					httpsuite.WithError(err),
					httpsuite.WithExternalErrorMessage("unauthorized"),
					httpsuite.WithInternalErrorMessage("failed to validate token"),
				)
				return
			}

			var claims map[string]any

			if err := idToken.Claims(&claims); err != nil {
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
// @Router			/api/login [get]
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
// @Router			/api/login/callback [get]
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
