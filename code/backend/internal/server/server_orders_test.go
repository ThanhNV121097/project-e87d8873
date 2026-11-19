package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func send(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestTheOrdersRouteIsMountedWithoutTheApiPrefix(t *testing.T) {
	// A filled hidden field is dropped before any database is asked.
	if rec := send(Router(fakeDB{}), http.MethodPost, "/v1/orders", `{"website":"x"}`); rec.Code != http.StatusAccepted {
		t.Errorf("POST /v1/orders: status = %d, want 202", rec.Code)
	}
	for _, path := range []string{"/api/v1/orders", "/v1/orders/"} {
		if rec := send(Router(fakeDB{}), http.MethodPost, path, `{"website":"x"}`); rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: status = %d, want 404 (the edge strips /api)", path, rec.Code)
		}
	}
}

// No order is ever read back: the one address takes POST only, and nothing
// under it is an order.
func TestOnlyPostIsOfferedOnOrdersAndNothingIsServedUnderIt(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions} {
		rec := send(Router(fakeDB{}), method, "/v1/orders", "")
		if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "POST" {
			t.Errorf("%s /v1/orders: status %d Allow %q, want 405 and POST", method, rec.Code, rec.Header().Get("Allow"))
		}
	}
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete} {
		if rec := send(Router(fakeDB{}), method, "/v1/orders/DEO-261116-0001", `{"website":"x"}`); rec.Code != http.StatusNotFound {
			t.Errorf("%s /v1/orders/DEO-261116-0001: status = %d, want 404", method, rec.Code)
		}
	}
}
