package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestWriteCodedUsesTheEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteCoded(rec, http.StatusBadRequest, CodeInvalidQuery, "There is no field “colourful” to filter bikes by.", "colourful")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if len(body) != 3 || body["code"] != "INVALID_QUERY" || body["field"] != "colourful" || !strings.Contains(body["error"], "colourful") {
		t.Errorf("body = %v, want error, code and field", body)
	}
}

func TestWriteCodedLeavesTheFieldOutWhenNoneIsGiven(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteCoded(rec, http.StatusNotFound, CodeNotFound, "That record does not exist.", "")

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if _, has := body["field"]; has || len(body) != 2 || body["code"] != "NOT_FOUND" {
		t.Errorf("body = %v, want only error and code", body)
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
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] == "" || body["code"] != "INTERNAL" {
		t.Errorf("body = %q, want the error envelope with code INTERNAL", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Error("the panic value leaked into the answer")
	}
}

var hex32 = regexp.MustCompile(`^[0-9a-f]{32}$`)

func requestIDOf(t *testing.T, sent string, send bool) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if send {
		req.Header.Set("X-Request-Id", sent)
	}
	rec := httptest.NewRecorder()
	var inContext string
	RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		inContext = RequestIDFrom(r.Context())
	})).ServeHTTP(rec, req)
	got := rec.Header().Get("X-Request-Id")
	if got != inContext {
		t.Errorf("header %q and context %q differ", got, inContext)
	}
	return got
}

func TestRequestIDKeepsAValidCallerId(t *testing.T) {
	if got := requestIDOf(t, "trace-1.A_b", true); got != "trace-1.A_b" {
		t.Errorf("id = %q, want the caller's", got)
	}
	if got := requestIDOf(t, strings.Repeat("a", 64), true); got != strings.Repeat("a", 64) {
		t.Errorf("a 64-character id was replaced: %q", got)
	}
}

func TestRequestIDReplacesAnInvalidOrMissingId(t *testing.T) {
	for name, sent := range map[string]string{
		"65 characters": strings.Repeat("a", 65),
		"a space":       "has space",
		"a slash":       "a/b",
		"an accent":     "café",
	} {
		got := requestIDOf(t, sent, true)
		if got == sent || !hex32.MatchString(got) {
			t.Errorf("%s: id = %q, want 32 hex characters that are not the caller's", name, got)
		}
	}
	if got := requestIDOf(t, "", false); !hex32.MatchString(got) {
		t.Errorf("no id sent: id = %q, want 32 hex characters", got)
	}
	if a, b := requestIDOf(t, "", false), requestIDOf(t, "", false); a == b {
		t.Error("two requests got the same generated id")
	}
}

func TestRequestIDIsOnARecoveredPanic(t *testing.T) {
	h := RequestID(Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if got := rec.Header().Get("X-Request-Id"); !hex32.MatchString(got) {
		t.Errorf("X-Request-Id = %q on a recovered 500, want 32 hex characters", got)
	}
}
