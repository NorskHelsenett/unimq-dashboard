package dex

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"golang.org/x/oauth2"
)

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

	authURL := d.Config.AuthCodeURL(state)

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

	// Verify before storing so a malformed token is rejected here rather than
	// on the next API call, and so the cookie can inherit the token's expiry.
	verified, err := d.ValidateToken(r.Context(), idToken)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("internal server error"),
			httpsuite.WithInternalErrorMessage("id_token failed verification"),
		)
		return
	}

	setSessionCookie(w, r, idToken, verified.Expiry)

	// Without a refresh token the session simply ends when the ID token
	// expires, so this is not fatal — but it is worth knowing about.
	if token.RefreshToken != "" {
		setRefreshCookie(w, r, token.RefreshToken)
	} else {
		slog.WarnContext(r.Context(), "no refresh token issued; sessions will end at ID token expiry",
			"scopes", d.Config.Scopes)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// @Summary		Refresh the session
// @Description	Exchanges the refresh token for a new ID token and reissues the session cookie.
// @Tags			Authentication
// @Produce		json
// @Success		204	{string}	string	"no content"
// @Failure		401	{object}	httpsuite.ErrorResponse
// @Failure		500	{object}	httpsuite.ErrorResponse
// @Router			/v1/login/refresh [post]
func (d *DexClient) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil || cookie.Value == "" {
		httpsuite.WriteJSONError(w,
			http.StatusUnauthorized,
			httpsuite.WithExternalErrorMessage("unauthorized"),
			httpsuite.WithInternalErrorMessage("no refresh token on request"),
		)
		return
	}

	// An already-expired token with no access token forces TokenSource to go
	// straight to the refresh grant.
	source := d.Config.TokenSource(r.Context(), &oauth2.Token{
		RefreshToken: cookie.Value,
		Expiry:       time.Now().Add(-time.Minute),
	})

	token, err := source.Token()
	if err != nil {
		// Spent, rotated away or revoked. Clear both cookies so the browser
		// stops retrying and falls back to a full login.
		clearSessionCookie(w, r)
		clearRefreshCookie(w, r)
		httpsuite.WriteJSONError(w,
			http.StatusUnauthorized,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("unauthorized"),
			httpsuite.WithInternalErrorMessage("refresh token exchange failed"),
		)
		return
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithExternalErrorMessage("internal server error"),
			httpsuite.WithInternalErrorMessage("id_token not found in refresh response"),
		)
		return
	}

	verified, err := d.ValidateToken(r.Context(), idToken)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("internal server error"),
			httpsuite.WithInternalErrorMessage("refreshed id_token failed verification"),
		)
		return
	}

	setSessionCookie(w, r, idToken, verified.Expiry)

	// Providers that rotate refresh tokens hand back a new one; oauth2 carries
	// the previous value forward when they do not, so this is always current.
	if token.RefreshToken != "" {
		setRefreshCookie(w, r, token.RefreshToken)
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary		Log out
// @Description	Clears the session cookie. The OIDC provider session is untouched because Dex exposes no end_session_endpoint.
// @Tags			Authentication
// @Produce		json
// @Success		204	{string}	string	"no content"
// @Router			/v1/logout [post]
func (d *DexClient) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w, r)
	clearRefreshCookie(w, r)
	w.WriteHeader(http.StatusNoContent)
}
