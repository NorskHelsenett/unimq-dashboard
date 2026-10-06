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
		Scopes: []string{
			oidc.ScopeOpenID,
			oidc.ScopeProfile,
			oidc.ScopeEmail,
			"groups",
			oidc.ScopeOfflineAccess,
		},
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
			slog.ErrorContext(ctx, "error closing response body", "error", err)
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

// oauthStateCookieName holds the per-request CSRF state for the authorization
// code flow.
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
		HttpOnly: true, // keep the cookie out of reach of XSS and page scripts.
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

// sessionCookieName carries the verified ID token for browser clients.
const sessionCookieName = "unimq_session"

// setSessionCookie sets the session cookie with the ID token and expiry.
func setSessionCookie(w http.ResponseWriter, r *http.Request, idToken string, expiry time.Time) {
	maxAge := int(time.Until(expiry).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    idToken,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSessionCookie clears the session cookie by setting it to an empty value and a past expiry date.
func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

// RefreshCookieName is the name of the OIDC refresh token cookie.
const refreshCookieName = "unimq_refresh"

// RefreshCookiePath must match the route the refresh handler is mounted on.
// Is is exported so tests can assert the two stay in agreement and fail if not.
const RefreshCookiePath = "/api/v1/login/refresh"

// refreshTokenTTL binds how long the browser can keep using the refresh token.
// this only stops a stale cookie lingering.
const refreshTokenTTL = 8 * time.Hour

func setRefreshCookie(w http.ResponseWriter, r *http.Request, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     RefreshCookiePath,
		MaxAge:   int(refreshTokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteStrictMode,
	})
}

func clearRefreshCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     RefreshCookiePath,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteStrictMode,
	})
}

// verifyState compares the state echoed back by Dex against the one this
// server issued. It verifies the single-use cookie is present and matches the query parameter.
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

var (
	ErrBearerTokenMissing = fmt.Errorf("bearer token missing")
	ErrBearerTokenInvalid = fmt.Errorf("bearer token invalid")
	ErrBearerTokenExpired = fmt.Errorf("bearer token expired")
)

// rawTokenFromRequest returns the encoded ID token for a request. Swagger and
// other API clients send it as a bearer token; the browser sends it in the
// session cookie. The header wins when both are present.
func (d *DexClient) rawTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader != "" {
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			return "", fmt.Errorf("%w, token does not have the correct 'Bearer ' prefix", ErrBearerTokenInvalid)
		}

		token := strings.TrimPrefix(authHeader, bearerPrefix)
		if token == "" {
			return "", fmt.Errorf("%w, token empty after trimming prefix", ErrBearerTokenInvalid)
		}

		return token, nil
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", ErrBearerTokenMissing
	}

	return cookie.Value, nil
}

// getBearerToken returns the verified ID token for a request. It checks the
// bearer token in the Authorization header first, then the session cookie. If
// neither is present, it returns [ErrBearerTokenMissing]. If the token is present
// but invalid, it returns [ErrBearerTokenInvalid]. If the token is expired, it
// returns [ErrBearerTokenExpired].
func (d *DexClient) getBearerToken(r *http.Request) (*oidc.IDToken, error) {
	token, err := d.rawTokenFromRequest(r)
	if err != nil {
		return nil, err
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
