package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
)

const (
	testEmail    = "owner@deo.test"
	testPassword = "a long test passphrase, 123"
)

// The accounts are built once: argon2id takes tens of milliseconds, and an
// Account is immutable.
var (
	testAccount        = NewAccount(testEmail, testPassword)
	otherPassword      = NewAccount(testEmail, "some other passphrase")
	unconfiguredAcount = NewAccount("", "")
)

type testClock struct{ t time.Time }

func (c *testClock) now() time.Time          { return c.t }
func (c *testClock) advance(d time.Duration) { c.t = c.t.Add(d) }

type sessionRow struct {
	email   string
	expires time.Time
}

// fakeSessions is a map keyed by hash that compares expiry with the clock it
// is given, and can be made to fail.
type fakeSessions struct {
	rows                            map[string]sessionRow
	createErr, existsErr, deleteErr error
	creates, exists, deletes        int
}

func newFakeSessions() *fakeSessions { return &fakeSessions{rows: map[string]sessionRow{}} }

func (f *fakeSessions) Create(_ context.Context, hash []byte, email string, _, expires time.Time) error {
	f.creates++
	if f.createErr != nil {
		return f.createErr
	}
	f.rows[string(hash)] = sessionRow{email, expires}
	return nil
}

func (f *fakeSessions) Exists(_ context.Context, hash []byte, email string, now time.Time) (bool, error) {
	f.exists++
	if f.existsErr != nil {
		return false, f.existsErr
	}
	row, ok := f.rows[string(hash)]
	return ok && row.email == email && row.expires.After(now), nil
}

func (f *fakeSessions) Delete(_ context.Context, hash []byte) error {
	f.deletes++
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.rows, string(hash))
	return nil
}

type harness struct {
	svc   *Service
	h     http.Handler
	store *fakeSessions
	clock *testClock
}

func newHarness(account *Account) *harness {
	return newHarnessOn(account, newFakeSessions(), &testClock{t: time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC)})
}

func newHarnessOn(account *Account, store *fakeSessions, clock *testClock) *harness {
	svc := newService(account, store, clock.now)
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/sessions", svc.SignIn)
	mux.HandleFunc("DELETE /v1/sessions/current", svc.SignOut)
	mux.HandleFunc("/v1/me", svc.Me)
	mux.Handle("/v1/guarded", svc.RequireOwner(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"guarded": "reached"})
	})))
	return &harness{svc: svc, h: httpx.RequestID(mux), store: store, clock: clock}
}

// send makes one request; headers are key, value pairs.
func (h *harness) send(method, path, body string, headers ...string) *httptest.ResponseRecorder {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
	}
	for i := 0; i+1 < len(headers); i += 2 {
		r.Header.Add(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	h.h.ServeHTTP(rec, r)
	return rec
}

func signInBody(email, password string) string {
	b, _ := json.Marshal(map[string]string{"email": email, "password": password})
	return string(b)
}

func (h *harness) signIn(email, password string, headers ...string) *httptest.ResponseRecorder {
	return h.send(http.MethodPost, "/v1/sessions", signInBody(email, password), append([]string{"Content-Type", "application/json"}, headers...)...)
}

// signedIn signs the owner in and returns the cookie's value.
func (h *harness) signedIn(t *testing.T) string {
	t.Helper()
	rec := h.signIn(testEmail, testPassword)
	if rec.Code != http.StatusOK {
		t.Fatalf("sign-in: status %d, body %s", rec.Code, rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == CookieName {
			return c.Value
		}
	}
	t.Fatal("sign-in set no session cookie")
	return ""
}

func (h *harness) me(cookie string) *httptest.ResponseRecorder {
	if cookie == "" {
		return h.send(http.MethodGet, "/v1/me", "")
	}
	return h.send(http.MethodGet, "/v1/me", "", "Cookie", CookieName+"="+cookie)
}

func envelope(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q is not JSON: %v", rec.Body.String(), err)
	}
	return body
}

func cookieNamed(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == CookieName {
			return c
		}
	}
	return nil
}

func TestTheRightPairSignsInAndAnswersTheEmailOnly(t *testing.T) {
	h := newHarness(testAccount)
	for _, typed := range []string{testEmail, "  OWNER@deo.test "} {
		rec := h.signIn(typed, testPassword)
		if rec.Code != http.StatusOK {
			t.Fatalf("%q: status %d, want 200", typed, rec.Code)
		}
		if got := envelope(t, rec); len(got) != 1 || got["email"] != testEmail {
			t.Errorf("%q: body %v, want only the email %s", typed, got, testEmail)
		}
		if strings.Contains(rec.Body.String(), testPassword) {
			t.Error("the answer carries the password")
		}
	}
}

func TestThePasswordIsInNoFieldOfTheAccount(t *testing.T) {
	v := reflect.ValueOf(*testAccount)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.String:
			if strings.Contains(f.String(), testPassword) {
				t.Errorf("field %s holds the password", v.Type().Field(i).Name)
			}
		case reflect.Slice:
			if f.Type().Elem().Kind() == reflect.Uint8 && bytes.Contains(f.Bytes(), []byte(testPassword)) {
				t.Errorf("field %s holds the password", v.Type().Field(i).Name)
			}
		}
	}
	if !testAccount.Configured() || unconfiguredAcount.Configured() {
		t.Error("Configured() does not follow the two secrets")
	}
}

func TestTheSessionCookieAttributes(t *testing.T) {
	cases := []struct {
		name    string
		headers []string
		secure  bool
	}{
		{"plain http", nil, false},
		{"https at the edge", []string{"X-Forwarded-Proto", "https"}, true},
		{"https as the last of two, in capitals", []string{"X-Forwarded-Proto", "http, HTTPS"}, true},
		{"https first, http last", []string{"X-Forwarded-Proto", "https, http"}, false},
	}
	for _, c := range cases {
		h := newHarness(testAccount)
		rec := h.signIn(testEmail, testPassword, c.headers...)
		ck := cookieNamed(rec)
		if ck == nil {
			t.Fatalf("%s: no cookie", c.name)
		}
		if !ck.HttpOnly || ck.SameSite != http.SameSiteLaxMode || ck.Path != "/" || ck.MaxAge != 604800 || ck.Domain != "" {
			t.Errorf("%s: cookie = %+v, want HttpOnly, SameSite=Lax, Path=/, Max-Age=604800 and no Domain", c.name, ck)
		}
		if want := h.clock.now().Add(7 * 24 * time.Hour); !ck.Expires.Equal(want) {
			t.Errorf("%s: Expires = %v, want %v", c.name, ck.Expires, want)
		}
		if ck.Secure != c.secure {
			t.Errorf("%s: Secure = %v, want %v", c.name, ck.Secure, c.secure)
		}
		if !regexpToken(ck.Value) {
			t.Errorf("%s: value %q is not 43 base64url characters", c.name, ck.Value)
		}
	}
	// TLS on the request itself counts as HTTPS.
	h := newHarness(testAccount)
	r := httptest.NewRequest(http.MethodPost, "/v1/sessions", strings.NewReader(signInBody(testEmail, testPassword)))
	r.Header.Set("Content-Type", "application/json")
	r.TLS = &tls.ConnectionState{}
	rec := httptest.NewRecorder()
	h.h.ServeHTTP(rec, r)
	if ck := cookieNamed(rec); ck == nil || !ck.Secure {
		t.Errorf("a request over TLS: cookie = %+v, want Secure", ck)
	}
}

func regexpToken(s string) bool { return tokenShape.MatchString(s) }

func TestAWrongPasswordAndAWrongEmailAreTheSameAnswer(t *testing.T) {
	h := newHarness(testAccount)
	a := h.signIn(testEmail, "not the password")
	b := h.signIn("someone@deo.test", testPassword)
	if a.Code != http.StatusUnauthorized || b.Code != http.StatusUnauthorized {
		t.Fatalf("statuses %d and %d, want 401", a.Code, b.Code)
	}
	if a.Body.String() != b.Body.String() {
		t.Errorf("bodies differ: %q and %q", a.Body.String(), b.Body.String())
	}
	got := envelope(t, a)
	if got["error"] != "That email and password do not match." || got["code"] != "UNAUTHENTICATED" {
		t.Errorf("body = %v", got)
	}
	if cookieNamed(a) != nil || cookieNamed(b) != nil {
		t.Error("a refused sign-in set a cookie")
	}
	if h.store.creates != 0 {
		t.Error("a refused sign-in stored a session")
	}
}

func TestSignInBodiesThatAreRefusedBeforeThePasswordIsChecked(t *testing.T) {
	long := strings.Repeat("é", 321)
	cases := []struct {
		name, body, contentType string
		status                  int
		code, field             string
	}{
		{"no password", `{"email":"a@b.c"}`, "application/json", 400, "MALFORMED_REQUEST", "password"},
		{"null email", `{"email":null,"password":"x"}`, "application/json", 400, "MALFORMED_REQUEST", "email"},
		{"password a number", `{"email":"a@b.c","password":123}`, "application/json", 400, "MALFORMED_REQUEST", "password"},
		{"a third key", `{"email":"a@b.c","password":"x","remember":true}`, "application/json", 400, "MALFORMED_REQUEST", "remember"},
		{"an array", `[]`, "application/json", 400, "MALFORMED_REQUEST", ""},
		{"null", `null`, "application/json", 400, "MALFORMED_REQUEST", ""},
		{"a string", `"x"`, "application/json", 400, "MALFORMED_REQUEST", ""},
		{"an empty body", ``, "application/json", 400, "MALFORMED_REQUEST", ""},
		{"trailing data", `{"email":"a@b.c","password":"x"} {}`, "application/json", 400, "MALFORMED_REQUEST", ""},
		{"not json", `{"email":`, "application/json", 400, "MALFORMED_REQUEST", ""},
		{"text/plain", `{"email":"a@b.c","password":"x"}`, "text/plain", 400, "MALFORMED_REQUEST", ""},
		{"json with a parameter is fine but the email is too long", signInBody(long, "x"), "application/json; charset=utf-8", 400, "VALIDATION_FAILED", "email"},
		{"password too long", signInBody("a@b.c", strings.Repeat("p", 1025)), "application/json", 400, "VALIDATION_FAILED", "password"},
		{"over 8 KiB", `{"email":"` + strings.Repeat("a", 9*1024) + `"}`, "application/json", 413, "PAYLOAD_TOO_LARGE", ""},
	}
	h := newHarness(testAccount)
	for _, c := range cases {
		rec := h.send(http.MethodPost, "/v1/sessions", c.body, "Content-Type", c.contentType)
		if rec.Code != c.status {
			t.Errorf("%s: status %d, want %d (%s)", c.name, rec.Code, c.status, rec.Body.String())
			continue
		}
		got := envelope(t, rec)
		if got["code"] != c.code || got["field"] != c.field || got["error"] == "" {
			t.Errorf("%s: body %v, want code %s and field %q", c.name, got, c.code, c.field)
		}
	}
	// None of them counted as a failed attempt: the right pair still gets in.
	if rec := h.signIn(testEmail, testPassword); rec.Code != http.StatusOK {
		t.Errorf("after %d refused bodies the right pair got %d", len(cases), rec.Code)
	}
}

func TestFiveFailuresLockTheAddressForFifteenMinutes(t *testing.T) {
	h := newHarness(testAccount)
	for i := 0; i < 5; i++ {
		if rec := h.signIn(testEmail, "wrong"); rec.Code != http.StatusUnauthorized {
			t.Fatalf("failure %d: status %d, want 401", i+1, rec.Code)
		}
	}
	rec := h.signIn(testEmail, testPassword)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("the sixth, with the right password: status %d, want 429", rec.Code)
	}
	got := envelope(t, rec)
	if got["error"] != "Too many attempts. Try again in a few minutes." || got["code"] != "RATE_LIMITED" {
		t.Errorf("body = %v", got)
	}
	var secs int
	if _, err := fmt.Sscan(rec.Header().Get("Retry-After"), &secs); err != nil || secs < 1 || secs > 900 {
		t.Errorf("Retry-After = %q, want 1 to 900 seconds", rec.Header().Get("Retry-After"))
	}
	if cookieNamed(rec) != nil {
		t.Error("a 429 set a cookie")
	}

	h.clock.advance(14*time.Minute + 59*time.Second)
	if rec := h.signIn(testEmail, testPassword); rec.Code != http.StatusTooManyRequests {
		t.Errorf("14 min 59 s later: status %d, want 429", rec.Code)
	}
	h.clock.advance(time.Second)
	if rec := h.signIn(testEmail, testPassword); rec.Code != http.StatusOK {
		t.Errorf("15 min later: status %d, want 200", rec.Code)
	}
}

func TestALockedAddressCannotExtendItsOwnLock(t *testing.T) {
	h := newHarness(testAccount)
	for i := 0; i < 5; i++ {
		h.signIn(testEmail, "wrong")
	}
	h.clock.advance(10 * time.Minute)
	for i := 0; i < 3; i++ {
		if rec := h.signIn(testEmail, "wrong"); rec.Code != http.StatusTooManyRequests {
			t.Fatalf("status %d, want 429", rec.Code)
		}
	}
	h.clock.advance(5 * time.Minute)
	if rec := h.signIn(testEmail, testPassword); rec.Code != http.StatusOK {
		t.Errorf("15 minutes after the fifth failure: status %d, want 200", rec.Code)
	}
}

func TestASuccessResetsTheCount(t *testing.T) {
	h := newHarness(testAccount)
	for round := 0; round < 2; round++ {
		for i := 0; i < 4; i++ {
			h.signIn(testEmail, "wrong")
		}
		if rec := h.signIn(testEmail, testPassword); rec.Code != http.StatusOK {
			t.Fatalf("round %d: the right pair after four failures got %d", round, rec.Code)
		}
	}
}

func TestFailuresSpreadOverMoreThanFifteenMinutesNeverLock(t *testing.T) {
	h := newHarness(testAccount)
	for i := 0; i < 4; i++ {
		h.signIn(testEmail, "wrong")
		h.clock.advance(6 * time.Minute)
	}
	h.signIn(testEmail, "wrong") // the fifth, but only the fourth is within 15 minutes of it
	if rec := h.signIn(testEmail, testPassword); rec.Code != http.StatusOK {
		t.Errorf("status %d, want 200", rec.Code)
	}
}

func TestTheThrottleKeysOnTheLastForwardedAddress(t *testing.T) {
	h := newHarness(testAccount)
	for i := 0; i < 5; i++ {
		h.signIn(testEmail, "wrong", "X-Forwarded-For", fmt.Sprintf("10.0.0.%d, 203.0.113.9", i))
	}
	if rec := h.signIn(testEmail, testPassword, "X-Forwarded-For", "10.0.0.99, 203.0.113.9"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("same edge address: status %d, want 429 (a caller's own header entries do not choose its key)", rec.Code)
	}
	if rec := h.signIn(testEmail, testPassword, "X-Forwarded-For", "203.0.113.10"); rec.Code != http.StatusOK {
		t.Errorf("another address: status %d, want 200", rec.Code)
	}
}

func TestTheThrottleTableIsBounded(t *testing.T) {
	clock := &testClock{t: time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC)}
	th := NewThrottle(clock.now)
	for i := 0; i < throttleMaxEntries+1; i++ {
		th.Fail(fmt.Sprintf("addr-%d", i))
		clock.advance(time.Millisecond)
	}
	if th.Len() != throttleMaxEntries {
		t.Errorf("the table holds %d addresses, want %d", th.Len(), throttleMaxEntries)
	}
	if _, locked := th.Check("addr-0"); locked {
		t.Error("an address that failed once is locked")
	}
}

func TestRetryAfterIsRoundedUpAndClamped(t *testing.T) {
	for d, want := range map[time.Duration]int{time.Millisecond: 1, 1500 * time.Millisecond: 2, 15 * time.Minute: 900, 20 * time.Minute: 900, 0: 1} {
		if got := retryAfterSeconds(d); got != want {
			t.Errorf("retryAfterSeconds(%v) = %d, want %d", d, got, want)
		}
	}
}

func TestAnAccountThatIsNotConfiguredRefusesEverySignInAndSaysSoInTheLog(t *testing.T) {
	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(prev)

	for name, account := range map[string]*Account{
		"both unset":           unconfiguredAcount,
		"no password":          NewAccount(testEmail, ""),
		"no email":             NewAccount("", testPassword),
		"white space only":     NewAccount("   ", "   "),
		"password of spaces":   NewAccount(testEmail, "   "),
		"email of white space": NewAccount("\t", testPassword),
	} {
		logs.Reset()
		h := newHarness(account)
		for i := 0; i < 2; i++ {
			rec := h.signIn(testEmail, testPassword)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s: status %d, want 401", name, rec.Code)
			}
			if got := envelope(t, rec); got["error"] != "That email and password do not match." || got["code"] != "UNAUTHENTICATED" {
				t.Errorf("%s: body %v", name, got)
			}
		}
		if n := strings.Count(logs.String(), "admin account is not configured"); n != 2 {
			t.Errorf("%s: the log says the account is not configured %d times for 2 attempts, want 2", name, n)
		}
		if strings.Contains(logs.String(), testPassword) || strings.Contains(logs.String(), testEmail) {
			t.Errorf("%s: the log holds the typed password or email:\n%s", name, logs.String())
		}
		if h.store.creates != 0 {
			t.Errorf("%s: a session was stored", name)
		}
	}
}

func TestNothingTheServiceLogsHoldsThePasswordOrTheEmail(t *testing.T) {
	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(prev)

	h := newHarness(testAccount)
	h.signIn(testEmail, "wrong password")
	cookie := h.signedIn(t)
	h.me(cookie)
	h.send(http.MethodDelete, "/v1/sessions/current", "", "Cookie", CookieName+"="+cookie)
	for _, secret := range []string{testPassword, "wrong password", testEmail, cookie} {
		if strings.Contains(logs.String(), secret) {
			t.Errorf("the log holds %q", secret)
		}
	}
}

func TestMeAnswersTheOwnerForALiveSessionAndOtherwise401(t *testing.T) {
	h := newHarness(testAccount)
	cookie := h.signedIn(t)

	rec := h.me(cookie)
	if rec.Code != http.StatusOK || envelope(t, rec)["email"] != testEmail {
		t.Fatalf("a live session: status %d, body %s", rec.Code, rec.Body.String())
	}
	rec = h.me("")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no cookie: status %d, want 401", rec.Code)
	}
	if got := envelope(t, rec); got["error"] != "You are not signed in." || got["code"] != "UNAUTHENTICATED" {
		t.Errorf("no cookie: body %v", got)
	}
}

func TestACookieWithOneCharacterChangedIsRefused(t *testing.T) {
	h := newHarness(testAccount)
	cookie := h.signedIn(t)
	for i := range cookie {
		swap := "A"
		if cookie[i] == 'A' {
			swap = "B"
		}
		changed := cookie[:i] + swap + cookie[i+1:]
		if rec := h.me(changed); rec.Code != http.StatusUnauthorized {
			t.Fatalf("character %d changed: status %d, want 401", i, rec.Code)
		}
	}
}

func TestASessionOlderThanSevenDaysIsRefused(t *testing.T) {
	h := newHarness(testAccount)
	cookie := h.signedIn(t)
	h.clock.advance(7*24*time.Hour - time.Second)
	if rec := h.me(cookie); rec.Code != http.StatusOK {
		t.Errorf("just under 7 days: status %d, want 200", rec.Code)
	}
	h.clock.advance(time.Second)
	if rec := h.me(cookie); rec.Code != http.StatusUnauthorized {
		t.Errorf("7 days: status %d, want 401", rec.Code)
	}
}

func TestACookieFromBeforeASecretChangedIsRefused(t *testing.T) {
	store, clock := newFakeSessions(), &testClock{t: time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC)}
	cookie := newHarnessOn(testAccount, store, clock).signedIn(t)
	for name, account := range map[string]*Account{
		"another password": otherPassword,
		"another email":    NewAccount("boss@deo.test", testPassword),
	} {
		if rec := newHarnessOn(account, store, clock).me(cookie); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: status %d, want 401", name, rec.Code)
		}
	}
}

func TestACookieOfTheWrongShapeOrAnUnconfiguredAccountAsksTheStoreNothing(t *testing.T) {
	h := newHarness(testAccount)
	for _, cookie := range []string{"short", strings.Repeat("a", 44), strings.Repeat("!", 43)} {
		if rec := h.me(cookie); rec.Code != http.StatusUnauthorized {
			t.Errorf("cookie %q: status %d, want 401", cookie, rec.Code)
		}
	}
	u := newHarness(unconfiguredAcount)
	if rec := u.me(strings.Repeat("a", 43)); rec.Code != http.StatusUnauthorized {
		t.Errorf("unconfigured: status %d, want 401", rec.Code)
	}
	if h.store.exists != 0 || u.store.exists != 0 {
		t.Errorf("the store was asked %d and %d times, want never", h.store.exists, u.store.exists)
	}
}

func TestSignOutEndsTheSessionAndClearsTheCookie(t *testing.T) {
	h := newHarness(testAccount)
	cookie := h.signedIn(t)

	rec := h.send(http.MethodDelete, "/v1/sessions/current", "", "Cookie", CookieName+"="+cookie)
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
		t.Fatalf("status %d, body %q, want 204 and none", rec.Code, rec.Body.String())
	}
	ck := cookieNamed(rec)
	if ck == nil || ck.Value != "" || ck.MaxAge >= 0 || !ck.HttpOnly || ck.SameSite != http.SameSiteLaxMode || ck.Path != "/" {
		t.Errorf("clearing cookie = %+v", ck)
	}
	if !strings.Contains(rec.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Errorf("Set-Cookie = %q, want Max-Age=0", rec.Header().Get("Set-Cookie"))
	}
	if rec := h.me(cookie); rec.Code != http.StatusUnauthorized {
		t.Errorf("the replayed cookie: status %d, want 401", rec.Code)
	}

	// With no session at all it is the same 204 and the same clearing cookie.
	rec = h.send(http.MethodDelete, "/v1/sessions/current", "")
	if rec.Code != http.StatusNoContent || cookieNamed(rec) == nil {
		t.Errorf("no session: status %d, cookie %v, want 204 and the clearing cookie", rec.Code, cookieNamed(rec))
	}
	rec = h.send(http.MethodDelete, "/v1/sessions/current", "", "X-Forwarded-Proto", "https")
	if ck := cookieNamed(rec); ck == nil || !ck.Secure {
		t.Errorf("over https the clearing cookie = %+v, want Secure", ck)
	}
}

func TestAStoreThatFailsIsNeverA401AndSignOutThenSendsNoCookie(t *testing.T) {
	refused := fmt.Errorf("dial: %w", syscall.ECONNREFUSED)
	for name, tc := range map[string]struct {
		err  error
		want int
	}{"refused": {refused, 503}, "a bug": {errors.New("relation \"sessions\" does not exist"), 500}} {
		h := newHarness(testAccount)
		cookie := h.signedIn(t)

		h.store.existsErr = tc.err
		if rec := h.me(cookie); rec.Code != tc.want {
			t.Errorf("%s, me: status %d, want %d", name, rec.Code, tc.want)
		}
		if rec := h.send(http.MethodGet, "/v1/guarded", "", "Cookie", CookieName+"="+cookie); rec.Code != tc.want {
			t.Errorf("%s, guard: status %d, want %d", name, rec.Code, tc.want)
		}
		h.store.deleteErr = tc.err
		rec := h.send(http.MethodDelete, "/v1/sessions/current", "", "Cookie", CookieName+"="+cookie)
		if rec.Code != tc.want || cookieNamed(rec) != nil || rec.Header().Get("Set-Cookie") != "" {
			t.Errorf("%s, sign-out: status %d, Set-Cookie %q, want %d and none", name, rec.Code, rec.Header().Get("Set-Cookie"), tc.want)
		}

		h.store.createErr = tc.err
		rec = h.signIn(testEmail, testPassword)
		if rec.Code != tc.want || cookieNamed(rec) != nil {
			t.Errorf("%s, sign-in: status %d, cookie %v, want %d and no cookie", name, rec.Code, cookieNamed(rec), tc.want)
		}
		if strings.Contains(rec.Body.String(), "sessions") {
			t.Errorf("%s: the answer names a table: %s", name, rec.Body.String())
		}
	}
}

func TestAFailedStoreAfterTheRightPasswordCountsNoFailure(t *testing.T) {
	h := newHarness(testAccount)
	h.store.createErr = errors.New("boom")
	for i := 0; i < 6; i++ {
		if rec := h.signIn(testEmail, testPassword); rec.Code != http.StatusInternalServerError {
			t.Fatalf("attempt %d: status %d, want 500, never 429", i+1, rec.Code)
		}
	}
}

func TestTheGuardLetsOnlyALiveSessionThrough(t *testing.T) {
	h := newHarness(testAccount)
	rec := h.send(http.MethodPost, "/v1/guarded", "")
	if rec.Code != http.StatusUnauthorized || strings.Contains(rec.Body.String(), "reached") {
		t.Errorf("no session: status %d, body %s", rec.Code, rec.Body.String())
	}
	cookie := h.signedIn(t)
	if rec := h.send(http.MethodPost, "/v1/guarded", "", "Cookie", CookieName+"="+cookie); rec.Code != http.StatusOK {
		t.Errorf("a live session: status %d, want 200", rec.Code)
	}
}

func TestOtherMethodsAre405WithAllow(t *testing.T) {
	h := newHarness(testAccount)
	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions} {
		if rec := h.send(m, "/v1/sessions", ""); rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "POST" {
			t.Errorf("%s /v1/sessions: status %d, Allow %q, want 405 and POST", m, rec.Code, rec.Header().Get("Allow"))
		}
	}
	for _, m := range []string{http.MethodPost, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions} {
		if rec := h.send(m, "/v1/me", ""); rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "GET" {
			t.Errorf("%s /v1/me: status %d, Allow %q, want 405 and GET", m, rec.Code, rec.Header().Get("Allow"))
		}
	}
}

func TestVerifyGivesUpWhenItsDeadlinePassesWhileWaitingForASlot(t *testing.T) {
	a := NewAccount(testEmail, testPassword)
	for i := 0; i < maxHashes; i++ {
		a.slots <- struct{}{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if ok, err := a.Verify(ctx, testEmail, testPassword); ok || !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Verify = %v, %v, want false and the deadline", ok, err)
	}
}
