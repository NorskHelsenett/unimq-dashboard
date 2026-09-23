package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

func requestWithVhostAndGroups(vhost string, groups []any) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("vhost", vhost)

	ctx := context.WithValue(request.Context(), chi.RouteCtxKey, routeContext)
	if groups != nil {
		ctx = context.WithValue(ctx, httpsuite.ClaimsContextKey, map[string]any{"groups": groups})
	}
	return request.WithContext(ctx)
}

func TestVhostACLMiddlewareAllowsAdmin(t *testing.T) {
	service := &APIService{AdminGroups: []string{"admins"}}
	request := requestWithVhostAndGroups("orders", []any{"admins"})
	recorder := httptest.NewRecorder()

	service.VhostACLMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestVhostACLMiddlewareRejectsMissingClaims(t *testing.T) {
	service := &APIService{AdminGroups: []string{"admins"}}
	request := requestWithVhostAndGroups("orders", nil)
	recorder := httptest.NewRecorder()

	service.VhostACLMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusForbidden)
	}
}
