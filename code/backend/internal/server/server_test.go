package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakeRow struct{ err error }

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*(dest[0].(*int)) = 1
	return nil
}

type fakeDB struct{ err error }

func (d fakeDB) QueryRow(context.Context, string, ...any) pgx.Row { return fakeRow{err: d.err} }

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestHealthzIsOKOnlyWhenTheDatabaseAnswers(t *testing.T) {
	if rec := get(Router(fakeDB{}), "/healthz"); rec.Code != http.StatusOK {
		t.Errorf("database up: status = %d, want 200", rec.Code)
	}
	if rec := get(Router(fakeDB{err: errors.New("connection refused")}), "/healthz"); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("database down: status = %d, want 503", rec.Code)
	}
}

func TestUnknownAddressAnswersTheEnvelope(t *testing.T) {
	rec := get(Router(fakeDB{}), "/v1/nothing/here")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] == "" {
		t.Errorf("body = %q, want the error envelope", rec.Body.String())
	}
}

func TestEveryAnswerIsUncacheable(t *testing.T) {
	for _, path := range []string{"/healthz", "/missing"} {
		if got := get(Router(fakeDB{}), path).Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("%s: Cache-Control = %q, want no-store", path, got)
		}
	}
}
