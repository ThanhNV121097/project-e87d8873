package collections

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"syscall"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
)

type scanFunc func(dest ...any) error

type fakeRow struct{ scan scanFunc }

func (r fakeRow) Scan(dest ...any) error { return r.scan(dest...) }

// fakeDB stands in for the pool: it records what it was asked and answers
// with the scan function it was given.
type fakeDB struct {
	scan  scanFunc
	calls int
	sql   string
	args  []any
}

func (d *fakeDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	d.calls++
	d.sql, d.args = sql, args
	return fakeRow{d.scan}
}

func listRows(total int64, items string) scanFunc {
	return func(dest ...any) error {
		*(dest[0].(*int64)) = total
		*(dest[1].(*[]byte)) = []byte(items)
		return nil
	}
}

func recordRow(rec string, version int64) scanFunc {
	return func(dest ...any) error {
		*(dest[0].(*[]byte)) = []byte(rec)
		if len(dest) > 1 {
			*(dest[1].(*int64)) = version
		}
		return nil
	}
}

func failing(err error) scanFunc { return func(...any) error { return err } }

// router mounts the two reads the way server.Router does.
func router(db DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/{collection}", Collection(db))
	mux.HandleFunc("/v1/{collection}/{id}", Record(db))
	return httpx.RequestID(httpx.NoStore(mux))
}

func serve(db DB, method, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	router(db).ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func envelope(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q is not an error envelope: %v", rec.Body.String(), err)
	}
	return body
}

func TestAListAnswersItemsTotalPageAndLimitAndNothingElse(t *testing.T) {
	db := &fakeDB{scan: listRows(14, `[{"id":"talon-2"}]`)}
	rec := serve(db, http.MethodGet, "/v1/products?page=3&limit=500")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 4 {
		t.Errorf("keys = %v, want items, total, page and limit", body)
	}
	for key, want := range map[string]string{"total": "14", "page": "3", "limit": "100", "items": `[{"id":"talon-2"}]`} {
		if got := string(body[key]); got != want {
			t.Errorf("%s = %s, want %s", key, got, want)
		}
	}
	if rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-Request-Id") == "" {
		t.Errorf("headers = %v, want no-store and a request id", rec.Header())
	}
}

func TestNumbersReachTheBrowserAsTheDatabaseWroteThem(t *testing.T) {
	db := &fakeDB{scan: listRows(1, `[{"id":"a","rating":5.0,"price":322990000,"stock":0}]`)}
	body := serve(db, http.MethodGet, "/v1/products").Body.String()
	for _, want := range []string{`"rating":5.0`, `"price":322990000`, `"stock":0`} {
		if !strings.Contains(body, want) {
			t.Errorf("body %s does not contain %s", body, want)
		}
	}
}

func TestAPastTheEndPageIsAnEmptyListNotAnError(t *testing.T) {
	db := &fakeDB{scan: listRows(14, `[]`)}
	rec := serve(db, http.MethodGet, "/v1/products?page=9")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"items":[]`) || !strings.Contains(rec.Body.String(), `"total":14`) {
		t.Errorf("status %d, body %s", rec.Code, rec.Body.String())
	}
}

func TestAnUnknownCollectionIs404BeforeAnythingElse(t *testing.T) {
	db := &fakeDB{scan: listRows(0, `[]`)}
	for _, c := range []struct{ method, target string }{
		{http.MethodGet, "/v1/nothing"},
		{http.MethodGet, "/v1/nothing?limit=0"},
		{http.MethodPost, "/v1/nothing"},
		{http.MethodGet, "/v1/nothing/x"},
		{http.MethodGet, "/v1/PRODUCTS"},
		{http.MethodGet, "/v1/orders"},
	} {
		rec := serve(db, c.method, c.target)
		body := envelope(t, rec)
		if rec.Code != http.StatusNotFound || body["code"] != "NOT_FOUND" || body["error"] != "That collection does not exist." {
			t.Errorf("%s %s: status %d, body %v", c.method, c.target, rec.Code, body)
		}
	}
	if db.calls != 0 {
		t.Errorf("the database was asked %d times for a collection that does not exist", db.calls)
	}
}

func TestARefusedQueryIs400WithTheFieldAndNeverReachesTheDatabase(t *testing.T) {
	db := &fakeDB{scan: listRows(0, `[]`)}
	for target, field := range map[string]string{
		"/v1/products?limit=0":             "limit",
		"/v1/products?page=abc":            "page",
		"/v1/products?colourful=red":       "colourful",
		"/v1/products?stock=abc":           "stock",
		"/v1/products?deal=maybe":          "deal",
		"/v1/products?deal_from=1":         "deal_from",
		"/v1/faqs?search=nothing":          "nothing",
		"/v1/products?sort=colourful":      "colourful",
		"/v1/products?category=a&category=b": "category",
		"/v1/products?q=" + strings.Repeat("x", 101): "q",
	} {
		rec := serve(db, http.MethodGet, target)
		body := envelope(t, rec)
		if rec.Code != http.StatusBadRequest || body["code"] != "INVALID_QUERY" || body["field"] != field || body["error"] == "" {
			t.Errorf("%s: status %d, body %v, want 400 INVALID_QUERY naming %s", target, rec.Code, body, field)
		}
		if strings.Contains(rec.Body.String(), `"items"`) {
			t.Errorf("%s: the refusal carries items", target)
		}
	}
	// A query string that cannot be read at all has no one field to blame.
	for _, target := range []string{"/v1/products?q=%zz", "/v1/products?a=1;b=2"} {
		rec := serve(db, http.MethodGet, target)
		body := envelope(t, rec)
		if rec.Code != http.StatusBadRequest || body["code"] != "INVALID_QUERY" || body["field"] != "" {
			t.Errorf("%s: status %d, body %v", target, rec.Code, body)
		}
	}
	if db.calls != 0 {
		t.Errorf("the database was asked %d times for a refused query", db.calls)
	}
}

func TestOtherMethodsAre405WithAllowGet(t *testing.T) {
	db := &fakeDB{scan: listRows(0, `[]`)}
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions} {
		for _, target := range []string{"/v1/products", "/v1/products/talon-2"} {
			rec := serve(db, method, target)
			if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "GET" {
				t.Errorf("%s %s: status %d, Allow %q", method, target, rec.Code, rec.Header().Get("Allow"))
				continue
			}
			if method != http.MethodHead && envelope(t, rec)["code"] != "METHOD_NOT_ALLOWED" {
				t.Errorf("%s %s: wrong code", method, target)
			}
		}
	}
	if db.calls != 0 {
		t.Errorf("the database was asked %d times for a method that is not offered", db.calls)
	}
}

func TestARecordIsAnsweredAsTheDatabaseWroteIt(t *testing.T) {
	const rec = `{"id":"stance-0","name":"Stance 0","price":66490000,"sizes_out":[]}`
	db := &fakeDB{scan: recordRow(rec, 7)}
	res := serve(db, http.MethodGet, "/v1/products/Stance-0")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}
	if got := strings.TrimSpace(res.Body.String()); got != rec {
		t.Errorf("body = %s, want %s", got, rec)
	}
	if got := res.Header().Get("ETag"); got != `"7"` {
		t.Errorf("ETag = %q, want \"7\"", got)
	}
	if !reflect.DeepEqual(db.args, []any{"Stance-0"}) {
		t.Errorf("args = %#v: the id is matched exactly as sent", db.args)
	}
	if strings.Contains(res.Body.String(), "version") {
		t.Error("the version is in the body")
	}
}

func TestOnlyBikesAndRidingTypesCarryAnETag(t *testing.T) {
	for name, want := range map[string]bool{"products": true, "categories": true, "reviews": false, "gallery": false, "team": false, "services": false, "faqs": false, "shop": false} {
		db := &fakeDB{scan: recordRow(`{"id":"x"}`, 3)}
		res := serve(db, http.MethodGet, "/v1/"+name+"/x")
		if _, has := res.Header()["Etag"]; has != want {
			t.Errorf("%s: ETag present = %v, want %v", name, has, want)
		}
	}
}

func TestARecordThatIsNotThereIs404(t *testing.T) {
	db := &fakeDB{scan: failing(pgx.ErrNoRows)}
	rec := serve(db, http.MethodGet, "/v1/products/no-such-bike")
	body := envelope(t, rec)
	if rec.Code != http.StatusNotFound || body["code"] != "NOT_FOUND" || body["error"] != "That record does not exist." {
		t.Errorf("status %d, body %v", rec.Code, body)
	}
	if rec.Header().Get("ETag") != "" {
		t.Error("a 404 carries an ETag")
	}
}

func TestAnIdTheDatabaseCannotHoldIsAnswered404WithoutAsking(t *testing.T) {
	db := &fakeDB{scan: recordRow(`{}`, 1)}
	rec := serve(db, http.MethodGet, "/v1/products/a%00b")
	if rec.Code != http.StatusNotFound || db.calls != 0 {
		t.Errorf("status %d after %d calls, want 404 and none", rec.Code, db.calls)
	}
}

// retryable is an error pgconn says never left the client.
type retryable struct{}

func (retryable) Error() string {
	return "conn busy"
}

func (retryable) SafeToRetry() bool {
	return true
}

func TestOnlyAnUnreachableDatabaseOrAGoingAwayServerIs503(t *testing.T) {
	noHost := &net.DNSError{Err: "no such host", Name: "db", IsNotFound: true}
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"refused", syscall.ECONNREFUSED, 503},
		{"refused, wrapped", fmt.Errorf("connect: %w", &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED}), 503},
		{"a failed dial", &net.OpError{Op: "dial", Err: errors.New("no route to host")}, 503},
		{"a host that is gone", noHost, 503},
		{"a host that is gone, as pgx words it", fmt.Errorf("failed to connect to `user=app database=app`: hostname resolving error (%w)", noHost), 503},
		{"a connect error", errors.New("failed to connect to `user=app database=app`: dial error (something new)"), 503},
		{"a connection reset", &net.OpError{Op: "read", Net: "tcp", Err: syscall.ECONNRESET}, 503},
		{"a broken pipe", fmt.Errorf("failed to write: %w", &net.OpError{Op: "write", Net: "tcp", Err: syscall.EPIPE}), 503},
		{"closed under the query", fmt.Errorf("receive: %w", io.ErrUnexpectedEOF), 503},
		{"closed cleanly", io.EOF, 503},
		{"a closed connection", net.ErrClosed, 503},
		{"never sent", retryable{}, 503},
		{"connection exception", &pgconn.PgError{Code: "08006"}, 503},
		{"shutting down", &pgconn.PgError{Code: "57P01"}, 503},
		{"crash shutdown", &pgconn.PgError{Code: "57P02"}, 503},
		{"starting up", &pgconn.PgError{Code: "57P03"}, 503},
		{"starting up, at connect", fmt.Errorf("failed to connect to `user=app`: server error (%w)", &pgconn.PgError{Code: "57P03"}), 503},
		{"a refused password", fmt.Errorf("failed to connect to `user=app`: server error (%w)", &pgconn.PgError{Code: "28P01"}), 500},
		{"a missing table", &pgconn.PgError{Code: "42P01", Message: `relation "products" does not exist`}, 500},
		{"a constraint", &pgconn.PgError{Code: "23505", ConstraintName: "uq_products_seq"}, 500},
		{"a cancelled query", &pgconn.PgError{Code: "57014"}, 500},
		{"a deadline", context.DeadlineExceeded, 500},
		{"a deadline in the pool", fmt.Errorf("acquire: %w", context.DeadlineExceeded), 500},
		{"a network timeout", &net.OpError{Op: "read", Net: "tcp", Err: os.ErrDeadlineExceeded}, 500},
		{"an abandoned request", context.Canceled, 500},
		{"a scan error", errors.New(`can't scan into dest[0]`), 500},
		{"anything else", errors.New("boom"), 500},
	}
	for _, c := range cases {
		for _, target := range []string{"/v1/products?q=secret+words", "/v1/products/talon-2"} {
			rec := serve(&fakeDB{scan: failing(c.err)}, http.MethodGet, target)
			body := envelope(t, rec)
			if rec.Code != c.status {
				t.Errorf("%s %s: status = %d, want %d", c.name, target, rec.Code, c.status)
			}
			wantCode, wantSentence := "INTERNAL", "Something went wrong. Please try again."
			if c.status == 503 {
				wantCode, wantSentence = "UNAVAILABLE", "The service is not ready."
			}
			if len(body) != 2 || body["code"] != wantCode || body["error"] != wantSentence {
				t.Errorf("%s %s: body = %v, want only the generic %s sentence", c.name, target, body, wantCode)
			}
			for _, leak := range []string{"relation", "products", "uq_", "SELECT", "scan", "boom", "dial", "items", "host", "db"} {
				if strings.Contains(rec.Body.String(), leak) {
					t.Errorf("%s %s: the answer leaks %q: %s", c.name, target, leak, rec.Body.String())
				}
			}
		}
	}
}

func TestUnavailableLooksAtTheWholeErrorChain(t *testing.T) {
	if unavailable(nil) || unavailable(errors.New("x")) || unavailable(context.DeadlineExceeded) {
		t.Error("an error that is not a lost connection was called unavailable")
	}
	if !unavailable(fmt.Errorf("a: %w", fmt.Errorf("b: %w", syscall.ECONNREFUSED))) {
		t.Error("a wrapped refused connection was not called unavailable")
	}
	if !unavailable(fmt.Errorf("a: %w", fmt.Errorf("b: %w", &net.DNSError{Err: "no such host", Name: "db"}))) {
		t.Error("a wrapped name lookup failure was not called unavailable")
	}
}
