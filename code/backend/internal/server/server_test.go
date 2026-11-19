package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/auth"
)

// Router is built here with an account that is not configured: nobody can sign
// in, and the public reads are unaffected.
func router(db DB) http.Handler { return Router(db, auth.NewAccount("", "")) }

type fakeRow struct{ err error }

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	switch d := dest[0].(type) {
	case *int:
		*d = 1
	case *int64:
		*d = 0
		*(dest[1].(*[]byte)) = []byte("[]")
	case *[]byte:
		*d = []byte(`{"id":"x"}`)
	}
	return nil
}

type fakeDB struct{ err error }

func (d fakeDB) QueryRow(context.Context, string, ...any) pgx.Row { return fakeRow{err: d.err} }

func do(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	return do(h, http.MethodGet, path)
}

func TestHealthzIsOKOnlyWhenTheDatabaseAnswers(t *testing.T) {
	if rec := get(router(fakeDB{}), "/healthz"); rec.Code != http.StatusOK {
		t.Errorf("database up: status = %d, want 200", rec.Code)
	}
	if rec := get(router(fakeDB{err: errors.New("connection refused")}), "/healthz"); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("database down: status = %d, want 503", rec.Code)
	}
}

func TestUnknownAddressAnswersTheEnvelope(t *testing.T) {
	for _, path := range []string{"/v1/nothing/here/at-all", "/v1/", "/v1/products/", "/missing"} {
		rec := get(router(fakeDB{}), path)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s: status = %d, want 404", path, rec.Code)
		}
		var body map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] == "" || body["code"] != "NOT_FOUND" {
			t.Errorf("%s: body = %q, want the error envelope with code NOT_FOUND", path, rec.Body.String())
		}
	}
}

func TestEveryAnswerIsUncacheableAndCarriesARequestID(t *testing.T) {
	for _, path := range []string{"/healthz", "/missing", "/v1/products", "/v1/products/talon-2", "/v1/nothing"} {
		rec := get(router(fakeDB{}), path)
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("%s: Cache-Control = %q, want no-store", path, got)
		}
		if rec.Header().Get("X-Request-Id") == "" {
			t.Errorf("%s: no X-Request-Id", path)
		}
	}
}

func TestTheReadRoutesAreMountedWithoutTheApiPrefix(t *testing.T) {
	if rec := get(router(fakeDB{}), "/v1/products"); rec.Code != http.StatusOK {
		t.Errorf("/v1/products: status = %d, want 200", rec.Code)
	}
	if rec := get(router(fakeDB{}), "/v1/products/talon-2"); rec.Code != http.StatusOK {
		t.Errorf("/v1/products/talon-2: status = %d, want 200", rec.Code)
	}
	if rec := get(router(fakeDB{}), "/api/v1/products"); rec.Code != http.StatusNotFound {
		t.Errorf("/api/v1/products: status = %d, want 404 (the edge strips /api)", rec.Code)
	}
}

func TestOtherMethodsOnAReadOnlyCollectionAre405WithAllow(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead} {
		for _, path := range []string{"/v1/reviews", "/v1/reviews/r1"} {
			rec := do(router(fakeDB{}), method, path)
			if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "GET" {
				t.Errorf("%s %s: status = %d, Allow = %q, want 405 and GET", method, path, rec.Code, rec.Header().Get("Allow"))
			}
		}
	}
}
