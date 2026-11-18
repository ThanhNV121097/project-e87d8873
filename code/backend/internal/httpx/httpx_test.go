package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteErrorUsesTheEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusConflict, "Giant Talon 2: only 3 in stock.")

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if len(body) != 1 || body["error"] != "Giant Talon 2: only 3 in stock." {
		t.Errorf("body = %v, want only the error sentence", body)
	}
}

func TestNoStoreSetsCacheControl(t *testing.T) {
	rec := httptest.NewRecorder()
	NoStore(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
}

func TestRecoverAnswers500WithTheEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	h := Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }))
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] == "" {
		t.Errorf("body = %q, want the error envelope", rec.Body.String())
	}
	if rec.Body.String() == "boom" {
		t.Error("the panic value leaked into the answer")
	}
}
