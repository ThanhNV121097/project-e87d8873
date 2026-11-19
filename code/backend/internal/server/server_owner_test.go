package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/auth"
)

func post(h http.Handler, path, body string, headers ...string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		r.Header.Add(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

func code(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q is not JSON: %v", rec.Body.String(), err)
	}
	return body["code"]
}

func TestTheAccountRoutesAreMountedWithoutTheApiPrefix(t *testing.T) {
	h := router(fakeDB{})
	if rec := post(h, "/v1/sessions", `{"email":"a@b.c","password":"x"}`); rec.Code != http.StatusUnauthorized || code(t, rec) != "UNAUTHENTICATED" {
		t.Errorf("POST /v1/sessions: status %d, want 401 (the account is not configured)", rec.Code)
	}
	if rec := post(h, "/api/v1/sessions", `{}`); rec.Code != http.StatusNotFound {
		t.Errorf("POST /api/v1/sessions: status %d, want 404 (the edge strips /api)", rec.Code)
	}
	if rec := get(h, "/v1/me"); rec.Code != http.StatusUnauthorized {
		t.Errorf("GET /v1/me with no cookie: status %d, want 401", rec.Code)
	}
	if rec := get(h, "/api/v1/me"); rec.Code != http.StatusNotFound {
		t.Errorf("GET /api/v1/me: status %d, want 404", rec.Code)
	}
}

func TestMethodsOnTheAccountRoutes(t *testing.T) {
	h := router(fakeDB{})
	for _, m := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		if rec := do(h, m, "/v1/sessions"); rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "POST" {
			t.Errorf("%s /v1/sessions: status %d, Allow %q, want 405 and POST", m, rec.Code, rec.Header().Get("Allow"))
		}
	}
	if rec := do(h, http.MethodPost, "/v1/me"); rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "GET" {
		t.Errorf("POST /v1/me: status %d, Allow %q, want 405 and GET", rec.Code, rec.Header().Get("Allow"))
	}

	// Sign-out is the one method-specific route: another method on it is an unknown collection.
	rec := do(h, http.MethodDelete, "/v1/sessions/current")
	if rec.Code != http.StatusNoContent || !strings.Contains(rec.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Errorf("DELETE /v1/sessions/current: status %d, Set-Cookie %q, want 204 and a clearing cookie", rec.Code, rec.Header().Get("Set-Cookie"))
	}
	for _, m := range []string{http.MethodGet, http.MethodPut, http.MethodPost} {
		rec := do(h, m, "/v1/sessions/current")
		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "That collection does not exist.") {
			t.Errorf("%s /v1/sessions/current: status %d, body %s, want 404 as an unknown collection", m, rec.Code, rec.Body.String())
		}
	}
	for _, path := range []string{"/v1/me/", "/v1/sessions/", "/v1/ME", "/v1/Sessions"} {
		if rec := get(h, path); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: status %d, want 404", path, rec.Code)
		}
	}
}

func TestEveryWriteToBikesAndRidingTypesNeedsASession(t *testing.T) {
	h := router(fakeDB{})
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/v1/products"}, {http.MethodPut, "/v1/products/talon-2"}, {http.MethodDelete, "/v1/products/talon-2"},
		{http.MethodPost, "/v1/categories"}, {http.MethodPut, "/v1/categories/trail"}, {http.MethodDelete, "/v1/categories/trail"},
	} {
		for _, cookie := range []string{"", "deo_session=short", "deo_session=" + strings.Repeat("a", 43)} {
			var headers []string
			if cookie != "" {
				headers = []string{"Cookie", cookie}
			}
			r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{"name":"x"}`))
			for i := 0; i+1 < len(headers); i += 2 {
				r.Header.Add(headers[i], headers[i+1])
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, r)
			if rec.Code != http.StatusUnauthorized || code(t, rec) != "UNAUTHENTICATED" {
				t.Errorf("%s %s (cookie %q): status %d, want 401", tc.method, tc.path, cookie, rec.Code)
			}
		}
	}
}

func TestTheOtherSixCollectionsAre405ToAWriteWhoeverAsks(t *testing.T) {
	h := router(fakeDB{})
	for _, name := range []string{"reviews", "team", "gallery", "services", "faqs", "shop"} {
		for _, tc := range []struct{ method, path string }{
			{http.MethodPost, "/v1/" + name}, {http.MethodPut, "/v1/" + name + "/x"}, {http.MethodDelete, "/v1/" + name + "/x"},
		} {
			if rec := do(h, tc.method, tc.path); rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "GET" {
				t.Errorf("%s %s: status %d, Allow %q, want 405 and GET", tc.method, tc.path, rec.Code, rec.Header().Get("Allow"))
			}
		}
	}
}

// With a configured account the whole door works through the real router: a
// sign-in, then the me check with the cookie it gave. The stand-in database
// answers every statement with one row, so this proves the wiring, not SQL.
func TestASignInThroughTheRouterGivesACookieTheMeCheckAccepts(t *testing.T) {
	h := Router(fakeDB{}, auth.NewAccount("owner@deo.test", "a long test passphrase"))
	rec := post(h, "/v1/sessions", `{"email":"Owner@Deo.test","password":"a long test passphrase"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("sign-in: status %d, body %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-Request-Id") == "" {
		t.Errorf("headers %v, want no-store and a request id", rec.Header())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "deo_session" {
		t.Fatalf("cookies = %v, want deo_session", cookies)
	}
	r := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	r.AddCookie(cookies[0])
	me := httptest.NewRecorder()
	h.ServeHTTP(me, r)
	if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), "owner@deo.test") {
		t.Errorf("me: status %d, body %s", me.Code, me.Body.String())
	}
}
