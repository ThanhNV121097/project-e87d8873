package collections

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"syscall"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
)

// noOwner is a guard that refuses everyone: nobody is signed in.
func noOwner(http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteCoded(w, http.StatusUnauthorized, httpx.CodeUnauthenticated, "You are not signed in.", "")
	})
}

// anyOwner is a guard that lets everyone through: the owner is signed in.
func anyOwner(next http.Handler) http.Handler { return next }

// fakeWriter answers the store's calls from functions a test replaces, and
// remembers what it was asked.
type fakeWriter struct {
	add            func(id string) ([]byte, int64, bool, error)
	lookup         func() (int64, bool, bool, error)
	categoryExists func() (bool, error)
	replace        func(version int64) ([]byte, int64, bool, error)
	versionOf      func() (int64, bool, error)
	remove         func() (bool, error)
	inUse          func() (string, int64, bool, error)

	calls   []string
	addedID []string
	values  []any
	version int64
}

func newFakeWriter() *fakeWriter {
	return &fakeWriter{
		add:            func(id string) ([]byte, int64, bool, error) { return []byte(`{"id":"` + id + `","name":"x"}`), 1, true, nil },
		lookup:         func() (int64, bool, bool, error) { return 4, true, true, nil },
		categoryExists: func() (bool, error) { return true, nil },
		replace: func(version int64) ([]byte, int64, bool, error) {
			return []byte(`{"id":"talon-2","name":"x"}`), version + 1, true, nil
		},
		versionOf: func() (int64, bool, error) { return 5, true, nil },
		remove:    func() (bool, error) { return true, nil },
		inUse:     func() (string, int64, bool, error) { return "Trail", 4, true, nil },
	}
}

func (f *fakeWriter) Add(_ context.Context, _ Spec, id string, values []any) ([]byte, int64, bool, error) {
	f.calls = append(f.calls, "add")
	f.addedID = append(f.addedID, id)
	f.values = values
	return f.add(id)
}

func (f *fakeWriter) Lookup(context.Context, Spec, string, *string) (int64, bool, bool, error) {
	f.calls = append(f.calls, "lookup")
	return f.lookup()
}

func (f *fakeWriter) CategoryExists(context.Context, string) (bool, error) {
	f.calls = append(f.calls, "category")
	return f.categoryExists()
}

func (f *fakeWriter) Replace(_ context.Context, _ Spec, _ string, version int64, values []any) ([]byte, int64, bool, error) {
	f.calls = append(f.calls, "replace")
	f.version, f.values = version, values
	return f.replace(version)
}

func (f *fakeWriter) Version(context.Context, Spec, string) (int64, bool, error) {
	f.calls = append(f.calls, "version")
	return f.versionOf()
}

func (f *fakeWriter) Remove(context.Context, Spec, string) (bool, error) {
	f.calls = append(f.calls, "remove")
	return f.remove()
}

func (f *fakeWriter) InUse(context.Context, string) (string, int64, bool, error) {
	f.calls = append(f.calls, "inuse")
	return f.inUse()
}

func writeRouter(wr Writer, guard Guard) http.Handler {
	db := &fakeDB{scan: listRows(0, `[]`)}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/{collection}", collection(db, wr, guard))
	mux.HandleFunc("/v1/{collection}/{id}", record(db, wr, guard))
	return httpx.RequestID(httpx.NoStore(mux))
}

// call sends a JSON request; headers are key, value pairs.
func call(h http.Handler, method, target, body string, headers ...string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(headers); i += 2 {
		r.Header.Add(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

func TestTheDispatcherAnswersInTheContractsOrder(t *testing.T) {
	guarded := 0
	counting := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			guarded++
			noOwner(next).ServeHTTP(w, r)
		})
	}
	h := writeRouter(newFakeWriter(), counting)

	for _, s := range All {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
			target := "/v1/" + s.Name
			if method != http.MethodPost {
				target += "/x"
			}
			guarded = 0
			rec := call(h, method, target, `{"not":"read"}`)
			switch {
			case s.Writable():
				if rec.Code != http.StatusUnauthorized || guarded != 1 {
					t.Errorf("%s %s: status %d, guard %d times, want 401 from the guard", method, target, rec.Code, guarded)
				}
			default:
				if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "GET" || guarded != 0 {
					t.Errorf("%s %s: status %d, Allow %q, guard %d times, want 405 and GET with no session asked", method, target, rec.Code, rec.Header().Get("Allow"), guarded)
				}
			}
		}
	}

	// Methods the route shape does not offer are 405 with this collection's Allow, whoever asks.
	for method, wantAllow := range map[string]map[string]string{
		http.MethodPut:     {"/v1/products": "GET, POST"},
		http.MethodPatch:   {"/v1/products": "GET, POST", "/v1/products/x": "GET, PUT, DELETE"},
		http.MethodHead:    {"/v1/products": "GET, POST", "/v1/products/x": "GET, PUT, DELETE"},
		http.MethodOptions: {"/v1/categories": "GET, POST", "/v1/categories/x": "GET, PUT, DELETE"},
		http.MethodPost:    {"/v1/categories/x": "GET, PUT, DELETE"},
	} {
		for target, allow := range wantAllow {
			if rec := call(h, method, target, ""); rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != allow {
				t.Errorf("%s %s: status %d, Allow %q, want 405 and %q", method, target, rec.Code, rec.Header().Get("Allow"), allow)
			}
		}
	}

	// An unknown collection is 404 for every method, before the method and before the session.
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		guarded = 0
		rec := call(h, method, "/v1/unicorns", "")
		if rec.Code != http.StatusNotFound || guarded != 0 {
			t.Errorf("%s /v1/unicorns: status %d, guard %d times, want 404 and no session asked", method, rec.Code, guarded)
		}
	}
}

func TestTheSessionIsCheckedBeforeTheBodyTheHeaderAndTheRecord(t *testing.T) {
	wr := newFakeWriter()
	h := writeRouter(wr, noOwner)
	for _, tc := range []struct{ method, target, body string }{
		{http.MethodPost, "/v1/products", `not json`},
		{http.MethodPut, "/v1/products/no-such-bike", `{"colourful":1}`},
		{http.MethodDelete, "/v1/categories/trail", ``},
	} {
		rec := call(h, tc.method, tc.target, tc.body, "Content-Type", "text/plain")
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status %d, want 401", tc.method, tc.target, rec.Code)
		}
		if got := envelope(t, rec); got["error"] != "You are not signed in." || got["code"] != "UNAUTHENTICATED" {
			t.Errorf("%s %s: body %v", tc.method, tc.target, got)
		}
	}
	if len(wr.calls) != 0 {
		t.Errorf("the store was asked %v for a visitor", wr.calls)
	}
}

func TestAReadStillReachesTheReadHandlerForAVisitor(t *testing.T) {
	h := writeRouter(newFakeWriter(), noOwner)
	if rec := call(h, http.MethodGet, "/v1/products", ""); rec.Code != http.StatusOK {
		t.Errorf("a visitor's list: status %d, want 200", rec.Code)
	}
}

func TestAddingABikeWithoutAnIDMakesOne(t *testing.T) {
	wr := newFakeWriter()
	rec := call(writeRouter(wr, anyOwner), http.MethodPost, "/v1/products", `{"name":"Test Bike","brand":"Giant","category":"trail","price":25000000,"status":"draft","stock":3}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
	}
	if !regexp.MustCompile(`^products-[a-z0-9]{6}$`).MatchString(wr.addedID[0]) {
		t.Errorf("made id %q, want products-<6 of a-z0-9>", wr.addedID[0])
	}
	if rec.Header().Get("ETag") != `"1"` || rec.Header().Get("Location") != "" {
		t.Errorf("ETag %q, Location %q, want \"1\" and none", rec.Header().Get("ETag"), rec.Header().Get("Location"))
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	if !strings.Contains(rec.Body.String(), `"id":"`+wr.addedID[0]+`"`) {
		t.Errorf("body %s does not carry the id", rec.Body.String())
	}
	if want := []string{"category", "add"}; fmt.Sprint(wr.calls) != fmt.Sprint(want) {
		t.Errorf("store calls %v, want %v: the riding type is checked before the insert", wr.calls, want)
	}
}

func TestAddingWithAClientIDTriesItOnce(t *testing.T) {
	wr := newFakeWriter()
	h := writeRouter(wr, anyOwner)
	if rec := call(h, http.MethodPost, "/v1/products", `{"id":"my-bike-1","name":"x"}`); rec.Code != http.StatusCreated || wr.addedID[0] != "my-bike-1" {
		t.Errorf("a free id: status %d, id %v, want 201 and my-bike-1", rec.Code, wr.addedID)
	}
	wr.add = func(string) ([]byte, int64, bool, error) { return nil, 0, false, nil }
	wr.addedID = nil
	rec := call(h, http.MethodPost, "/v1/products", `{"id":"my-bike-1","name":"x"}`)
	if rec.Code != http.StatusConflict || len(wr.addedID) != 1 {
		t.Fatalf("a taken id: status %d after %d tries, want 409 after one", rec.Code, len(wr.addedID))
	}
	if got := envelope(t, rec); got["error"] != "That id is already used." || got["code"] != "ID_TAKEN" {
		t.Errorf("body %v", got)
	}
}

func TestAServerMadeIDIsRetriedFiveTimesThenFails(t *testing.T) {
	wr := newFakeWriter()
	wr.add = func(string) ([]byte, int64, bool, error) { return nil, 0, false, nil }
	rec := call(writeRouter(wr, anyOwner), http.MethodPost, "/v1/categories", `{"name":"Gravel"}`)
	if rec.Code != http.StatusInternalServerError || len(wr.addedID) != 5 {
		t.Fatalf("status %d after %d tries, want 500 after 5", rec.Code, len(wr.addedID))
	}
	seen := map[string]bool{}
	for _, id := range wr.addedID {
		if !strings.HasPrefix(id, "categories-") || seen[id] {
			t.Errorf("made ids %v: want a new categories-… id each time", wr.addedID)
		}
		seen[id] = true
	}
}

func TestAMissingRidingTypeIsRefusedBeforeTheIDIsTried(t *testing.T) {
	wr := newFakeWriter()
	wr.categoryExists = func() (bool, error) { return false, nil }
	rec := call(writeRouter(wr, anyOwner), http.MethodPost, "/v1/products", `{"id":"my-bike-1","name":"x","category":"unicycle"}`)
	got := envelope(t, rec)
	if rec.Code != http.StatusBadRequest || got["field"] != "category" || got["code"] != "VALIDATION_FAILED" {
		t.Errorf("status %d, body %v, want a 400 on category", rec.Code, got)
	}
	for _, c := range wr.calls {
		if c == "add" {
			t.Error("the insert was tried for a missing riding type")
		}
	}

	// The riding type deleted in the instant after it was checked: the same 400.
	wr = newFakeWriter()
	wr.add = func(string) ([]byte, int64, bool, error) { return nil, 0, false, &pgconn.PgError{Code: "23503"} }
	rec = call(writeRouter(wr, anyOwner), http.MethodPost, "/v1/products", `{"name":"x","category":"trail"}`)
	if got := envelope(t, rec); rec.Code != http.StatusBadRequest || got["field"] != "category" {
		t.Errorf("a foreign-key failure: status %d, body %v, want the category 400", rec.Code, got)
	}
}

func TestARefusedBodySavesNothing(t *testing.T) {
	wr := newFakeWriter()
	h := writeRouter(wr, anyOwner)
	for _, body := range []string{`{"colourful":"red","name":"x"}`, `{"name":"x","price":-1}`, `{"name":"x","status":"live"}`, `[]`} {
		if rec := call(h, http.MethodPost, "/v1/products", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", body, rec.Code)
		}
	}
	if rec := call(h, http.MethodPut, "/v1/products/talon-2", `{"name":"x","stock":-1}`, "If-Match", `"4"`); rec.Code != http.StatusBadRequest {
		t.Errorf("a refused change: status %d, want 400", rec.Code)
	}
	if len(wr.calls) != 0 {
		t.Errorf("the store was asked %v for refused bodies", wr.calls)
	}
}

func TestAnOversizeBodyIs413BeforeAnythingIsParsed(t *testing.T) {
	wr := newFakeWriter()
	body := `{"name":"x","description":"` + strings.Repeat("a", 1<<20) + `"}`
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		target := "/v1/products"
		if method == http.MethodPut {
			target += "/talon-2"
		}
		rec := call(writeRouter(wr, anyOwner), method, target, body, "Content-Type", "text/plain")
		if got := envelope(t, rec); rec.Code != http.StatusRequestEntityTooLarge || got["code"] != "PAYLOAD_TOO_LARGE" {
			t.Errorf("%s: status %d, body %v, want 413 (before the content type is looked at)", method, rec.Code, got)
		}
	}
	if len(wr.calls) != 0 {
		t.Errorf("the store was asked %v", wr.calls)
	}
}

func TestAChangeChecksInTheContractsOrder(t *testing.T) {
	const body = `{"id":"talon-2","name":"Talon 2","category":"trail","stock":5}`
	badTokens := []string{`7`, `W/"7"`, `*`, `"abc"`, `"7", "8"`, `"0"`, `"07"`, `""`, ` `, `"99999999999999999999"`}

	// Unknown id and no If-Match: the 404 comes first.
	wr := newFakeWriter()
	wr.lookup = func() (int64, bool, bool, error) { return 0, false, true, nil }
	if rec := call(writeRouter(wr, anyOwner), http.MethodPut, "/v1/products/talon-2", body); rec.Code != http.StatusNotFound {
		t.Errorf("unknown id, no If-Match: status %d, want 404", rec.Code)
	}

	// Known id and no If-Match: 428 with its own code.
	wr = newFakeWriter()
	rec := call(writeRouter(wr, anyOwner), http.MethodPut, "/v1/products/talon-2", body)
	if got := envelope(t, rec); rec.Code != http.StatusPreconditionRequired || got["code"] != "PRECONDITION_REQUIRED" || !strings.Contains(got["error"], "Reload") {
		t.Errorf("no If-Match: status %d, body %v, want 428", rec.Code, got)
	}

	for _, token := range badTokens {
		rec := call(writeRouter(newFakeWriter(), anyOwner), http.MethodPut, "/v1/products/talon-2", body, "If-Match", token)
		if got := envelope(t, rec); rec.Code != http.StatusBadRequest || got["code"] != "MALFORMED_REQUEST" || got["field"] != "" {
			t.Errorf("If-Match %q: status %d, body %v, want a 400 with no field", token, rec.Code, got)
		}
	}
	// Two header lines are as malformed as one line with two tokens.
	rec = call(writeRouter(newFakeWriter(), anyOwner), http.MethodPut, "/v1/products/talon-2", body, "If-Match", `"4"`, "If-Match", `"5"`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("two If-Match lines: status %d, want 400", rec.Code)
	}

	// A missing riding type is a 400 before staleness is looked at.
	wr = newFakeWriter()
	wr.lookup = func() (int64, bool, bool, error) { return 4, true, false, nil }
	rec = call(writeRouter(wr, anyOwner), http.MethodPut, "/v1/products/talon-2", body, "If-Match", `"1"`)
	if got := envelope(t, rec); rec.Code != http.StatusBadRequest || got["field"] != "category" {
		t.Errorf("a missing riding type: status %d, body %v", rec.Code, got)
	}
	for _, c := range wr.calls {
		if c == "replace" {
			t.Error("the replace ran for a missing riding type")
		}
	}

	// A body id that is not the address's id.
	rec = call(writeRouter(newFakeWriter(), anyOwner), http.MethodPut, "/v1/products/talon-2", `{"id":"other","name":"x"}`, "If-Match", `"4"`)
	if got := envelope(t, rec); rec.Code != http.StatusBadRequest || got["field"] != "id" {
		t.Errorf("a different body id: status %d, body %v", rec.Code, got)
	}
}

func TestAChangeReplacesAtTheVersionTheCallerRead(t *testing.T) {
	wr := newFakeWriter()
	rec := call(writeRouter(wr, anyOwner), http.MethodPut, "/v1/products/talon-2", `{"name":"Talon 2","stock":5,"price":""}`, "If-Match", `"7"`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
	}
	if wr.version != 7 {
		t.Errorf("replace ran at version %d, want the header's 7", wr.version)
	}
	if rec.Header().Get("ETag") != `"8"` {
		t.Errorf("ETag %q, want the new version \"8\"", rec.Header().Get("ETag"))
	}
	var sawStock bool
	for _, v := range wr.values {
		sawStock = sawStock || v == int64(5)
	}
	if !sawStock {
		t.Errorf("the replace was not given the stock: %#v", wr.values)
	}
	if want := []string{"lookup", "replace"}; fmt.Sprint(wr.calls) != fmt.Sprint(want) {
		t.Errorf("store calls %v, want %v", wr.calls, want)
	}
}

func TestAStaleChangeIs409AndAGoneRecordIs404(t *testing.T) {
	wr := newFakeWriter()
	wr.replace = func(int64) ([]byte, int64, bool, error) { return nil, 0, false, nil }
	rec := call(writeRouter(wr, anyOwner), http.MethodPut, "/v1/products/talon-2", `{"name":"x"}`, "If-Match", `"4"`)
	got := envelope(t, rec)
	const sentence = "This record changed while you were editing it — for example an order was placed. Reload and make your change again."
	if rec.Code != http.StatusConflict || got["code"] != "STALE_RECORD" || got["error"] != sentence {
		t.Errorf("status %d, body %v, want the 409 with the fixed sentence", rec.Code, got)
	}

	wr.versionOf = func() (int64, bool, error) { return 0, false, nil }
	if rec := call(writeRouter(wr, anyOwner), http.MethodPut, "/v1/products/talon-2", `{"name":"x"}`, "If-Match", `"4"`); rec.Code != http.StatusNotFound {
		t.Errorf("deleted in the instant between: status %d, want 404", rec.Code)
	}

	wr = newFakeWriter()
	wr.replace = func(int64) ([]byte, int64, bool, error) { return nil, 0, false, &pgconn.PgError{Code: "23503"} }
	rec = call(writeRouter(wr, anyOwner), http.MethodPut, "/v1/products/talon-2", `{"name":"x","category":"trail"}`, "If-Match", `"4"`)
	if got := envelope(t, rec); rec.Code != http.StatusBadRequest || got["field"] != "category" {
		t.Errorf("a foreign-key failure: status %d, body %v, want the category 400", rec.Code, got)
	}
}

func TestDeletingARecordNeedsNoVersionAndAnswers204WithNoBody(t *testing.T) {
	wr := newFakeWriter()
	h := writeRouter(wr, anyOwner)
	rec := call(h, http.MethodDelete, "/v1/products/talon-2", "", "If-Match", `"garbage"`)
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 || rec.Header().Get("Content-Type") != "" {
		t.Errorf("status %d, body %q, Content-Type %q, want 204 with neither", rec.Code, rec.Body.String(), rec.Header().Get("Content-Type"))
	}
	wr.remove = func() (bool, error) { return false, nil }
	if rec := call(h, http.MethodDelete, "/v1/products/talon-2", ""); rec.Code != http.StatusNotFound {
		t.Errorf("a record that is not there: status %d, want 404", rec.Code)
	}
}

func TestARidingTypeInUseIsRefusedWithHowManyBikesUseIt(t *testing.T) {
	for _, tc := range []struct {
		bikes int64
		want  string
	}{
		{4, "Trail is still the riding type of 4 bikes. Give those bikes another riding type first."},
		{1, "Trail is still the riding type of 1 bike. Give those bikes another riding type first."},
	} {
		wr := newFakeWriter()
		wr.remove = func() (bool, error) { return false, &pgconn.PgError{Code: "23503"} }
		wr.inUse = func() (string, int64, bool, error) { return "Trail", tc.bikes, true, nil }
		rec := call(writeRouter(wr, anyOwner), http.MethodDelete, "/v1/categories/trail", "")
		got := envelope(t, rec)
		if rec.Code != http.StatusConflict || got["code"] != "IN_USE" || got["error"] != tc.want {
			t.Errorf("%d bikes: status %d, body %v, want 409 %q", tc.bikes, rec.Code, got, tc.want)
		}
	}
}

func TestARidingTypeDeleteUnderChangingBikes(t *testing.T) {
	// Deleted in the instant between the failed delete and the count: 404.
	wr := newFakeWriter()
	wr.remove = func() (bool, error) { return false, &pgconn.PgError{Code: "23503"} }
	wr.inUse = func() (string, int64, bool, error) { return "", 0, false, nil }
	if rec := call(writeRouter(wr, anyOwner), http.MethodDelete, "/v1/categories/trail", ""); rec.Code != http.StatusNotFound {
		t.Errorf("gone in the instant between: status %d, want 404", rec.Code)
	}

	// The last bike moved away in the instant between: the delete is tried again, at most 3 times in all.
	wr = newFakeWriter()
	removes := 0
	wr.remove = func() (bool, error) {
		removes++
		if removes < 3 {
			return false, &pgconn.PgError{Code: "23503"}
		}
		return true, nil
	}
	wr.inUse = func() (string, int64, bool, error) { return "Trail", 0, true, nil }
	if rec := call(writeRouter(wr, anyOwner), http.MethodDelete, "/v1/categories/trail", ""); rec.Code != http.StatusNoContent || removes != 3 {
		t.Errorf("a type freed on the third try: status %d after %d tries, want 204 after 3", rec.Code, removes)
	}
	removes = 0
	wr.remove = func() (bool, error) { removes++; return false, &pgconn.PgError{Code: "23503"} }
	if rec := call(writeRouter(wr, anyOwner), http.MethodDelete, "/v1/categories/trail", ""); rec.Code != http.StatusInternalServerError || removes != 3 {
		t.Errorf("never freed: status %d after %d tries, want 500 after 3", rec.Code, removes)
	}
}

func TestAFailedStoreIs500UnlessTheDatabaseCannotBeReached(t *testing.T) {
	bug := &pgconn.PgError{Code: "42P01", Message: `relation "products" does not exist`}
	refused := fmt.Errorf("dial tcp: %w", syscall.ECONNREFUSED)
	steps := map[string]func(wr *fakeWriter, err error){
		"category": func(wr *fakeWriter, err error) { wr.categoryExists = func() (bool, error) { return false, err } },
		"add":      func(wr *fakeWriter, err error) { wr.add = func(string) ([]byte, int64, bool, error) { return nil, 0, false, err } },
		"lookup":   func(wr *fakeWriter, err error) { wr.lookup = func() (int64, bool, bool, error) { return 0, false, false, err } },
		"replace":  func(wr *fakeWriter, err error) { wr.replace = func(int64) ([]byte, int64, bool, error) { return nil, 0, false, err } },
		"version": func(wr *fakeWriter, err error) {
			wr.replace = func(int64) ([]byte, int64, bool, error) { return nil, 0, false, nil }
			wr.versionOf = func() (int64, bool, error) { return 0, false, err }
		},
		"remove": func(wr *fakeWriter, err error) { wr.remove = func() (bool, error) { return false, err } },
	}
	requests := map[string]func(h http.Handler) *httptest.ResponseRecorder{
		"category": func(h http.Handler) *httptest.ResponseRecorder {
			return call(h, http.MethodPost, "/v1/products", `{"name":"x","category":"trail"}`)
		},
		"add": func(h http.Handler) *httptest.ResponseRecorder {
			return call(h, http.MethodPost, "/v1/products", `{"name":"x"}`)
		},
		"lookup": func(h http.Handler) *httptest.ResponseRecorder {
			return call(h, http.MethodPut, "/v1/products/talon-2", `{"name":"x"}`, "If-Match", `"1"`)
		},
		"replace": func(h http.Handler) *httptest.ResponseRecorder {
			return call(h, http.MethodPut, "/v1/products/talon-2", `{"name":"x"}`, "If-Match", `"1"`)
		},
		"version": func(h http.Handler) *httptest.ResponseRecorder {
			return call(h, http.MethodPut, "/v1/products/talon-2", `{"name":"x"}`, "If-Match", `"1"`)
		},
		"remove": func(h http.Handler) *httptest.ResponseRecorder {
			return call(h, http.MethodDelete, "/v1/products/talon-2", "")
		},
	}
	for step, fail := range steps {
		for name, tc := range map[string]struct {
			err  error
			want int
		}{"a bug": {bug, 500}, "a refused connection": {refused, 503}, "a deadline": {context.DeadlineExceeded, 500}} {
			wr := newFakeWriter()
			fail(wr, tc.err)
			rec := requests[step](writeRouter(wr, anyOwner))
			if rec.Code != tc.want {
				t.Errorf("%s, %s: status %d, want %d", step, name, rec.Code, tc.want)
			}
			if strings.Contains(rec.Body.String(), "products") || strings.Contains(rec.Body.String(), "relation") {
				t.Errorf("%s, %s: the answer names a table: %s", step, name, rec.Body.String())
			}
			if got := envelope(t, rec); got["error"] == "" || got["code"] == "" {
				t.Errorf("%s, %s: body %v is not the envelope", step, name, got)
			}
		}
	}
}

func TestAConstraintFailureIsInternalNotARetry(t *testing.T) {
	for _, code := range []string{"23502", "23505", "23514"} {
		wr := newFakeWriter()
		wr.add = func(string) ([]byte, int64, bool, error) { return nil, 0, false, &pgconn.PgError{Code: code} }
		if rec := call(writeRouter(wr, anyOwner), http.MethodPost, "/v1/products", `{"id":"a","name":"x"}`); rec.Code != http.StatusInternalServerError {
			t.Errorf("SQLSTATE %s: status %d, want 500", code, rec.Code)
		}
	}
}

func TestMadeIDsAreAlwaysAValidID(t *testing.T) {
	shape := regexp.MustCompile(`^[a-z]+-[a-z0-9]{6}$`)
	seen := map[string]bool{}
	for i := 0; i < 2000; i++ {
		id := newID("products")
		if !shape.MatchString(id) || !idShape.MatchString(id) {
			t.Fatalf("made id %q", id)
		}
		seen[id] = true
	}
	if len(seen) < 1990 {
		t.Errorf("%d different ids in 2000", len(seen))
	}
}

func TestTheStatementsAreFixedTextBuiltFromTheFieldTable(t *testing.T) {
	for _, s := range All {
		stmts, ok := writeStatements[s.Name]
		if !s.Writable() {
			if ok {
				t.Errorf("%s has write statements", s.Name)
			}
			continue
		}
		if !ok {
			t.Fatalf("%s has none", s.Name)
		}
		n := len(s.Fields)
		if got := strings.Count(stmts.add, "$"); got != n+1 {
			t.Errorf("%s add: %d placeholders, want %d", s.Name, got, n+1)
		}
		if got := strings.Count(stmts.replace, "$"); got != n+2 {
			t.Errorf("%s replace: %d placeholders, want %d", s.Name, got, n+2)
		}
		for i, f := range s.Fields {
			if !strings.Contains(stmts.add, fmt.Sprintf("$%d%s", i+2, castOf(f))) || !strings.Contains(stmts.replace, fmt.Sprintf("%s = $%d%s", quote(f.Key), i+3, castOf(f))) {
				t.Errorf("%s.%s: the placeholder or its cast is missing", s.Name, f.Key)
			}
		}
		for _, sql := range []string{stmts.add, stmts.replace} {
			head := sql[:strings.Index(sql, "RETURNING")]
			for _, never := range []string{`"seq"`, `SET "version"`, `, "version"`, `"created_at"`, `"updated_at"`, `"_source"`} {
				if strings.Contains(head, never) {
					t.Errorf("%s: a write names %s:\n%s", s.Name, never, head)
				}
			}
			if !strings.Contains(sql, returningList(s)) {
				t.Errorf("%s: the RETURNING list is not the answer columns plus the version", s.Name)
			}
		}
		if !strings.Contains(stmts.add, `ON CONFLICT ("id") DO NOTHING`) || !strings.Contains(stmts.replace, `"version" = $2::bigint`) {
			t.Errorf("%s: the add must not fail on a taken id and the replace must compare the version", s.Name)
		}
		if !strings.HasSuffix(returningList(s), `"_source", "version"`) {
			t.Errorf("%s: RETURNING list %s", s.Name, returningList(s))
		}
	}
	// Only the bikes' statement reads the riding type.
	if strings.Contains(writeStatements["categories"].lookup, "$2") || !strings.Contains(writeStatements["products"].lookup, "$2::text IS NULL") {
		t.Error("the lookup of riding types has a second placeholder, or the one of bikes lacks it")
	}
	if !strings.Contains(sqlInUse, `p."category" = c."id"`) || !strings.Contains(sqlCategoryExists, `"categories"`) {
		t.Error("the riding-type statements do not read the riding types")
	}
}

func TestEveryRequestedValueIsOnlyEverAPlaceholder(t *testing.T) {
	for _, stmts := range writeStatements {
		for _, sql := range []string{stmts.add, stmts.replace, stmts.lookup, stmts.version, stmts.remove} {
			// The one literal is the key the answer drops: to_jsonb(w) - 'version'.
			if strings.ContainsAny(strings.ReplaceAll(sql, "- 'version'", ""), "'") {
				t.Errorf("a statement holds a quoted literal: %s", sql)
			}
		}
	}
}

func TestACollectionsWritesAreOnlyToBikesAndRidingTypes(t *testing.T) {
	var got []string
	for name := range writeStatements {
		got = append(got, name)
	}
	if len(got) != 2 || writeStatements["products"].add == "" || writeStatements["categories"].add == "" {
		t.Errorf("write statements for %v, want products and categories only", got)
	}
}
