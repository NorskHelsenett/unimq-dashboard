package v1_test

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	routes "github.com/sisneve/rabbitmq-dashboard/internal/api/routes/v1"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/dex"
)

// Helper function to walk the routes and return a map of registered routes.
func walkRoutes(t *testing.T, r chi.Router) map[string]bool {
	t.Helper()

	got := make(map[string]bool)
	err := chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		got[method+" "+route] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walking routes: %v", err)
	}

	return got
}

// authRouter mirrors how SetupRoutes nests the authentication routes, so the
// paths below are the real ones a browser hits.
func authRouter(t *testing.T) chi.Router {
	t.Helper()

	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// The handlers are never invoked here, only registered, so a
			// zero-value client is enough and no Dex provider is needed.
			routes.SetupAuthenticationRoutes(r, &dex.DexClient{})
		})
	})

	return r
}

func TestAuthenticationRoutesAreRegistered(t *testing.T) {
	got := walkRoutes(t, authRouter(t))

	for _, want := range []string{
		"GET /api/v1/login/redirect",
		"GET /api/v1/login/callback",
		"POST /api/v1/login/refresh",
		"POST /api/v1/logout",
	} {
		if !got[want] {
			t.Errorf("route %q is not registered; registered routes: %v", want, got)
		}
	}
}

// TestRefreshCookiePathMatchesRoute pins the coupling between the refresh
// cookie's Path attribute and the route the handler is mounted on. If the route
// moves without the constant following, browsers stop sending the refresh
// token: sessions would silently stop renewing and every user would be thrown
// back to Dex at ID token expiry, with nothing failing at build or test time.
func TestRefreshCookiePathMatchesRoute(t *testing.T) {
	got := walkRoutes(t, authRouter(t))

	want := "POST " + dex.RefreshCookiePath
	if !got[want] {
		t.Errorf(
			"refresh cookie Path is %q but no route is mounted there; registered routes: %v",
			dex.RefreshCookiePath, got,
		)
	}
}
