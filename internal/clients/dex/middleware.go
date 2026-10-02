package dex

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
)

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
