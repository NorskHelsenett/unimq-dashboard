package dex

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func testClient() *DexClient {
	return &DexClient{
		Config: &oauth2.Config{
			ClientID:    "unimq-dashboard",
			RedirectURL: "http://localhost:8080/api/v1/login/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL:  "http://localhost:5556/dex/auth",
				TokenURL: "http://localhost:5556/dex/token",
			},
		},
	}
}

// issueState runs RedirectHandler and returns the state placed in the
// authorization URL along with the state cookie that was set.
func issueState(t *testing.T) (string, *http.Cookie) {
	t.Helper()

	rec := httptest.NewRecorder()
	testClient().RedirectHandler(rec, httptest.NewRequest(http.MethodGet, "/api/v1/login/redirect", nil))

	if rec.Code != http.StatusFound {
		t.Fatalf("redirect: got status %d, want %d", rec.Code, http.StatusFound)
	}

	loc, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("redirect: unparseable Location: %v", err)
	}
	state := loc.Query().Get("state")

	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == oauthStateCookieName {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("redirect: no state cookie was set")
	}
	return state, cookie
}

func TestRedirectHandlerIssuesRandomState(t *testing.T) {
	state1, cookie := issueState(t)
	state2, _ := issueState(t)

	if state1 == "" {
		t.Fatal("no state in authorization URL")
	}
	if state1 == "state" {
		t.Fatal("state is the hardcoded literal (U1 regression)")
	}
	if state1 == state2 {
		t.Fatalf("state is not per-request: %q reused", state1)
	}
	if cookie.Value != state1 {
		t.Fatalf("cookie %q does not match URL state %q", cookie.Value, state1)
	}
	if !cookie.HttpOnly {
		t.Error("state cookie is not HttpOnly")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("state cookie SameSite = %v, want Lax", cookie.SameSite)
	}
	if cookie.MaxAge <= 0 {
		t.Errorf("state cookie MaxAge = %d, want a positive TTL", cookie.MaxAge)
	}
}

func TestRedirectHandlerMarksCookieSecureBehindTLSProxy(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/login/redirect", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	rec := httptest.NewRecorder()
	testClient().RedirectHandler(rec, req)

	for _, c := range rec.Result().Cookies() {
		if c.Name == oauthStateCookieName && !c.Secure {
			t.Error("state cookie is not Secure on a forwarded https request")
		}
	}
}

func TestCallbackRejectsBadState(t *testing.T) {
	good, cookie := issueState(t)

	tests := []struct {
		name   string
		state  string
		cookie *http.Cookie
	}{
		{"no state and no cookie", "", nil},
		{"state but no cookie", good, nil},
		{"cookie but no state", "", cookie},
		{"forged state", "attacker-chosen", cookie},
		{"hardcoded literal", "state", cookie},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/login/callback?code=abc&state="+url.QueryEscape(tc.state), nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			rec := httptest.NewRecorder()
			testClient().OauthCallbackHandler(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want %d (state check did not reject)", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestCallbackAcceptsMatchingState(t *testing.T) {
	state, cookie := issueState(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/login/callback?code=abc&state="+state, nil)
	req.AddCookie(cookie)

	rec := httptest.NewRecorder()
	testClient().OauthCallbackHandler(rec, req)

	// State passed, so the handler proceeds to the token exchange, which has
	// no server to talk to here. Anything other than 400 proves the state
	// check let the request through.
	if rec.Code == http.StatusBadRequest {
		t.Fatalf("matching state was rejected: %s", rec.Body.String())
	}
}

func TestCallbackClearsStateCookie(t *testing.T) {
	state, cookie := issueState(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/login/callback?code=abc&state="+state, nil)
	req.AddCookie(cookie)

	rec := httptest.NewRecorder()
	testClient().OauthCallbackHandler(rec, req)

	var cleared bool
	for _, c := range rec.Result().Cookies() {
		if c.Name == oauthStateCookieName && c.MaxAge < 0 && c.Value == "" {
			cleared = true
		}
	}
	if !cleared {
		t.Error("state cookie was not expired after use; it remains replayable")
	}
}

func TestMissingCodeIsRejected(t *testing.T) {
	state, cookie := issueState(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/login/callback?state="+state, nil)
	req.AddCookie(cookie)

	rec := httptest.NewRecorder()
	testClient().OauthCallbackHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), "bad request") {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}
