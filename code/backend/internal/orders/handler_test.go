package orders

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
)

type fakeStore struct {
	placed Placed
	err    error
	calls  int
	got    Request
}

func (s *fakeStore) Place(_ context.Context, req Request, _ func() time.Time) (Placed, error) {
	s.calls++
	s.got = req
	return s.placed, s.err
}

func sample() Placed {
	return Placed{
		Number: "DEO-261116-0001", Placed: "2026-11-16T09:40:00+07:00", Name: "Zed Quill", Phone: "0999 888 777", Delivery: "local",
		Address: "9 Secret Street", Payment: "transfer",
		Lines:    []PlacedLine{{ID: "talon-2", Name: "Giant Talon 2", Price: 18990000, Image: "/images/ref/talon-2.jpg", Variant: "M", Qty: 2}},
		Subtotal: 37980000, Total: 37980000,
	}
}

// goodBody carries words that must never appear in a log line.
const goodBody = `{"name":"Zed Quill","phone":"0999 888 777","delivery":"local","address":"9 Secret Street","payment":"transfer","note":"private words","lines":[{"id":"talon-2","variant":"M","qty":2}],"website":""}`

func serveOrders(s Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/orders", Place(s, time.Now))
	return httpx.RequestID(httpx.NoStore(mux))
}

func send(h http.Handler, method, body, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/v1/orders", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func keysOf(t *testing.T, body []byte) []string {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("body %q is not an object: %v", body, err)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(old) })
	return &buf
}

func TestAnAcceptedOrderAnswers201WithTheElevenKeysAndNoNote(t *testing.T) {
	store := &fakeStore{placed: sample()}
	rec := send(serveOrders(store), http.MethodPost, goodBody, "application/json")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body)
	}
	want := "address,delivery,delivery_fee,lines,name,number,payment,phone,placed,subtotal,total"
	if got := strings.Join(keysOf(t, rec.Body.Bytes()), ","); got != want {
		t.Errorf("keys = %s, want %s", got, want)
	}
	var body struct {
		Placed string            `json:"placed"`
		Lines  []json.RawMessage `json:"lines"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if at, err := time.Parse(time.RFC3339, body.Placed); err != nil {
		t.Errorf("placed %q does not parse with its offset: %v", body.Placed, err)
	} else if _, offset := at.Zone(); offset != 7*3600 {
		t.Errorf("placed %q has offset %d, want +07:00", body.Placed, offset)
	}
	if got := strings.Join(keysOf(t, body.Lines[0]), ","); got != "id,image,name,price,qty,variant" {
		t.Errorf("line keys = %s", got)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	if rec.Header().Get("Location") != "" || rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-Request-Id") == "" {
		t.Errorf("headers = %v, want no Location, Cache-Control no-store and an X-Request-Id", rec.Header())
	}
	wantReq := Request{Name: "Zed Quill", Phone: "0999 888 777", Delivery: "local", Address: "9 Secret Street", Payment: "transfer", Note: "private words",
		Lines: []Line{{ID: "talon-2", Variant: "M", Qty: 2}}}
	if store.calls != 1 || fmt.Sprint(store.got) != fmt.Sprint(wantReq) {
		t.Errorf("the store was asked %d times for %+v, want once for %+v", store.calls, store.got, wantReq)
	}
}

func TestALineLeavesOutAnImageAndASizeItDoesNotHave(t *testing.T) {
	placed := sample()
	placed.Lines = []PlacedLine{{ID: "frame", Name: "Giant Frame", Price: 5, Qty: 1}}
	rec := send(serveOrders(&fakeStore{placed: placed}), http.MethodPost, goodBody, "application/json")
	var body struct{ Lines []json.RawMessage }
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(keysOf(t, body.Lines[0]), ","); got != "id,name,price,qty" {
		t.Errorf("line keys = %s, want id,name,price,qty", got)
	}
}

func TestARefusedBikeIs409InTheEnvelope(t *testing.T) {
	for _, err := range []error{
		&Refusal{Sentence: "Giant Talon 2: only 3 in stock."},
		fmt.Errorf("wrapped: %w", &Refusal{Sentence: "Giant Talon 2: only 3 in stock."}),
	} {
		rec := send(serveOrders(&fakeStore{err: err}), http.MethodPost, goodBody, "application/json")
		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409", rec.Code)
		}
		var body map[string]string
		if e := json.Unmarshal(rec.Body.Bytes(), &body); e != nil {
			t.Fatal(e)
		}
		if len(body) != 2 || body["code"] != "BIKE_UNAVAILABLE" || body["error"] != "Giant Talon 2: only 3 in stock." {
			t.Errorf("body = %v", body)
		}
	}
}

func TestABotIsDroppedWith202AndAnEmptyObjectAndTheStoreIsNotAsked(t *testing.T) {
	store := &fakeStore{placed: sample()}
	rec := send(serveOrders(store), http.MethodPost, `{"website":"x"}`, "application/json")
	if rec.Code != http.StatusAccepted || strings.TrimSpace(rec.Body.String()) != "{}" {
		t.Errorf("status %d body %q, want 202 and {}", rec.Code, rec.Body)
	}
	if store.calls != 0 {
		t.Errorf("the store was asked %d times for a bot", store.calls)
	}
	if rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-Request-Id") == "" {
		t.Errorf("headers = %v", rec.Header())
	}
}

func TestARefusedRequestIs400WithTheFieldAndNeverReachesTheStore(t *testing.T) {
	store := &fakeStore{placed: sample()}
	h := serveOrders(store)
	cases := []struct {
		body, ctype, code, field string
	}{
		{"not json", "application/json", "MALFORMED_REQUEST", ""},
		{"[]", "application/json", "MALFORMED_REQUEST", ""},
		{goodBody, "text/plain", "MALFORMED_REQUEST", ""},
		{`{"colour":"red"}`, "application/json", "MALFORMED_REQUEST", "colour"},
		{`{"phone":"1"}`, "application/json", "VALIDATION_FAILED", "name"},
	}
	for _, c := range cases {
		rec := send(h, http.MethodPost, c.body, c.ctype)
		var body map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		if rec.Code != http.StatusBadRequest || body["code"] != c.code || body["field"] != c.field || body["error"] == "" {
			t.Errorf("%q: status %d body %v, want 400 %s field %q", c.body, rec.Code, body, c.code, c.field)
		}
	}
	if store.calls != 0 {
		t.Errorf("the store was asked %d times for refused requests", store.calls)
	}
}

func TestABodyOverTheLimitIs413BeforeAnyOtherCheck(t *testing.T) {
	store := &fakeStore{placed: sample()}
	h := serveOrders(store)
	pad := func(n int) string { return goodBody + strings.Repeat(" ", n-len(goodBody)) }

	if rec := send(h, http.MethodPost, pad(MaxBody), "application/json"); rec.Code != http.StatusCreated {
		t.Errorf("a body of exactly %d bytes: status = %d, want 201", MaxBody, rec.Code)
	}
	rec := send(h, http.MethodPost, pad(MaxBody+1), "text/plain")
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusRequestEntityTooLarge || body["code"] != "PAYLOAD_TOO_LARGE" {
		t.Errorf("a declared size one past: status %d body %v, want 413 PAYLOAD_TOO_LARGE before the content type is looked at", rec.Code, body)
	}

	// A body of unknown length is cut off while it is read.
	req := httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(pad(MaxBody+1)))
	req.ContentLength = -1
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("an undeclared size one past: status = %d, want 413", rec.Code)
	}
	if store.calls != 1 {
		t.Errorf("the store was asked %d times, want only for the body that fit", store.calls)
	}
}

func TestAFailureOfTheStoreIsGenericAndOnlyAnUnreachableDatabaseIs503(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"a refused connection", fmt.Errorf("begin: %w", syscall.ECONNREFUSED), 503},
		{"a failed statement", errors.New(`ERROR: relation "products" does not exist (SQLSTATE 42P01)`), 500},
		{"a deadline", context.DeadlineExceeded, 500},
		{"an abandoned request", context.Canceled, 500},
		{"a broken record", errBrokenRecord, 500},
	}
	for _, c := range cases {
		rec := send(serveOrders(&fakeStore{err: c.err}), http.MethodPost, goodBody, "application/json")
		var body map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		wantCode, wantSentence := "INTERNAL", "Something went wrong. Please try again."
		if c.status == 503 {
			wantCode, wantSentence = "UNAVAILABLE", "The service is not ready."
		}
		if rec.Code != c.status || len(body) != 2 || body["code"] != wantCode || body["error"] != wantSentence {
			t.Errorf("%s: status %d body %v, want %d with only the generic %s sentence", c.name, rec.Code, body, c.status, wantCode)
		}
	}
}

func TestOnlyPostIsOffered(t *testing.T) {
	store := &fakeStore{placed: sample()}
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		rec := send(serveOrders(store), method, goodBody, "application/json")
		if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "POST" {
			t.Errorf("%s: status %d Allow %q, want 405 and POST", method, rec.Code, rec.Header().Get("Allow"))
		}
		if method != http.MethodHead {
			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["code"] != "METHOD_NOT_ALLOWED" {
				t.Errorf("%s: body %q, want the envelope with METHOD_NOT_ALLOWED", method, rec.Body)
			}
		}
	}
	if store.calls != 0 {
		t.Errorf("the store was asked %d times by methods that are not offered", store.calls)
	}
}

func TestTheLogNeverHoldsTheCustomersWords(t *testing.T) {
	logs := captureLog(t)
	for _, c := range []struct {
		store *fakeStore
		body  string
	}{
		{&fakeStore{placed: sample()}, goodBody},
		{&fakeStore{err: &Refusal{Sentence: "Giant Talon 2 is out of stock."}}, goodBody},
		{&fakeStore{err: errors.New("boom")}, goodBody},
		{&fakeStore{}, strings.Replace(goodBody, `"name":"Zed Quill",`, "", 1)},
	} {
		send(serveOrders(c.store), http.MethodPost, c.body, "application/json")
	}
	log := logs.String()
	for _, secret := range []string{"Zed Quill", "0999 888 777", "Secret Street", "private words"} {
		if strings.Contains(log, secret) {
			t.Errorf("the log holds %q:\n%s", secret, log)
		}
	}
	if !strings.Contains(log, "order placed") || !strings.Contains(log, "DEO-261116-0001") || !strings.Contains(log, "order refused") {
		t.Errorf("the log lacks the placed and refused lines:\n%s", log)
	}
}
