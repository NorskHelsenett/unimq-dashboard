package dex

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"golang.org/x/oauth2"
)

type DexClient struct {
	Config   *oauth2.Config
	Verifier *oidc.IDTokenVerifier
}

func NewDexClient(ctx context.Context, config *config.OIDCConfig) (*DexClient, error) {

	provider, err := oidc.NewProvider(ctx, config.OIDCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     config.OIDCClientID,
		ClientSecret: config.OIDCClientSecret,
		RedirectURL:  config.OIDCRedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	}

	return &DexClient{
		Config: &oauth2Config,
		Verifier: provider.Verifier(&oidc.Config{
			ClientID: config.OIDCClientID,
		}),
	}, nil
}

func (d *DexClient) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	// generate a CSRF state
	state := fmt.Sprintf("st_%d", time.Now().UnixNano())
	// Simple example using a cookie:
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(10 * time.Minute),
	})
	http.Redirect(w, r, d.Config.AuthCodeURL(state), http.StatusFound)
}

func (d *DexClient) HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")

	slog.Info("Received OAuth2 callback", "state", state)

	oauth2Token, err := d.Config.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		httpsuite.WriteJSONError(w, "failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		httpsuite.WriteJSONError(w, "no id_token field in oauth2 token", http.StatusInternalServerError)
		return
	}

	// Parse and verify ID Token payload.
	idToken, err := d.Verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		httpsuite.WriteJSONError(w, "failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract custom claims.
	var claims struct {
		Email    string   `json:"email"`
		Verified bool     `json:"email_verified"`
		Groups   []string `json:"groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		httpsuite.WriteJSONError(w, "failed to parse claims: "+err.Error(), http.StatusInternalServerError)
	}
}
