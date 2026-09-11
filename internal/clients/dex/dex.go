package dex

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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
		ClientID: config.OIDCClientID,
		Endpoint: provider.Endpoint(),
		Scopes:   []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	}

	return &DexClient{
		Config: &oauth2Config,
		Verifier: provider.Verifier(&oidc.Config{
			ClientID: config.OIDCClientID,
		}),
	}, nil
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
					httpsuite.WithInternalErrorMessage("Authorization header does not start with 'Bearer '"),
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
					httpsuite.WithInternalErrorMessage("Failed to parse claims"),
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
