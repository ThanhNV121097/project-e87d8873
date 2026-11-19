package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"sort"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
)

// The fixed sentences of the account's answers. Consumers branch on the code
// and the status, never on a sentence; these are what a person reads.
const (
	sentenceNotSignedIn   = "You are not signed in."
	sentenceNoMatch       = "That email and password do not match."
	sentenceTooMany       = "Too many attempts. Try again in a few minutes."
	sentenceMethod        = "That method is not offered here."
	sentenceNotReady      = "The service is not ready."
	sentenceInternal      = "Something went wrong. Please try again."
	sentenceTooLarge      = "The request is too large."
	sentenceNeedJSON      = "The request must be sent as JSON."
	sentenceNeedObject    = "The request must be a JSON object."
	maxSignInBody         = 8192
	maxEmailChars         = 320
	maxPasswordBytes      = 1024
	signInTimeout         = 10 * time.Second
	sessionCheckTimeout   = 5 * time.Second
	signInFieldEmail      = "email"
	signInFieldPassword   = "password"
	outcomeOK             = "ok"
	outcomeBadCredentials = "bad_credentials"
	outcomeNotConfigured  = "not_configured"
	outcomeLocked         = "locked"
)

// Service is the account's three endpoints and the guard of every owner
// write. Build it with New once per process.
type Service struct {
	account  *Account
	sessions Sessions
	throttle *Throttle
	now      func() time.Time
}

// New returns the service over db, reading time from now.
func New(account *Account, db DB, now func() time.Time) *Service {
	return newService(account, pgSessions{db}, now)
}

func newService(account *Account, sessions Sessions, now func() time.Time) *Service {
	return &Service{account: account, sessions: sessions, throttle: NewThrottle(now), now: now}
}

func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	httpx.WriteCoded(w, http.StatusMethodNotAllowed, httpx.CodeMethodNotAllowed, sentenceMethod, "")
}

func notSignedIn(w http.ResponseWriter) {
	httpx.WriteCoded(w, http.StatusUnauthorized, httpx.CodeUnauthenticated, sentenceNotSignedIn, "")
}

func internal(w http.ResponseWriter) {
	httpx.WriteCoded(w, http.StatusInternalServerError, httpx.CodeInternal, sentenceInternal, "")
}

// storeFailed answers a failed session statement: 503 only when the database
// cannot be reached, 500 for everything else. It is never a 401: a database
// that is down is not "nobody is signed in". The real cause is logged, never
// sent, and never carries a cookie, a token or an address.
func storeFailed(w http.ResponseWriter, r *http.Request, op string, err error) {
	attrs := []any{"request_id", httpx.RequestIDFrom(r.Context()), "op", op, "error", err}
	if errors.Is(err, context.Canceled) {
		slog.Info("session request abandoned by the caller", attrs...)
	} else {
		slog.Error("session statement failed", attrs...)
	}
	if httpx.Unavailable(err) {
		httpx.WriteCoded(w, http.StatusServiceUnavailable, httpx.CodeUnavailable, sentenceNotReady, "")
		return
	}
	internal(w)
}

// signInRefusal is one 400 or 413 of the sign-in's body checks.
type signInRefusal struct {
	status         int
	code, sentence string
	field          string
}

// readSignIn reads and checks a sign-in body in the contract's order: size,
// content type, a JSON object, no undefined key, email then password present
// and text, then their lengths.
func readSignIn(w http.ResponseWriter, r *http.Request) (email, password string, ref *signInRefusal) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSignInBody)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			return "", "", &signInRefusal{http.StatusRequestEntityTooLarge, httpx.CodePayloadTooLarge, sentenceTooLarge, ""}
		}
		return "", "", &signInRefusal{http.StatusBadRequest, httpx.CodeMalformedRequest, sentenceNeedObject, ""}
	}
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		return "", "", &signInRefusal{http.StatusBadRequest, httpx.CodeMalformedRequest, sentenceNeedJSON, ""}
	}
	var fields map[string]json.RawMessage
	if t := bytes.TrimSpace(data); len(t) == 0 || t[0] != '{' || json.Unmarshal(data, &fields) != nil {
		return "", "", &signInRefusal{http.StatusBadRequest, httpx.CodeMalformedRequest, sentenceNeedObject, ""}
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if k != signInFieldEmail && k != signInFieldPassword {
			return "", "", &signInRefusal{http.StatusBadRequest, httpx.CodeMalformedRequest, "The field “" + k + "” is not part of a sign-in.", k}
		}
	}
	texts := [2]string{}
	for i, k := range [2]string{signInFieldEmail, signInFieldPassword} {
		raw, present := fields[k]
		if !present || string(raw) == "null" {
			return "", "", &signInRefusal{http.StatusBadRequest, httpx.CodeMalformedRequest, "The field “" + k + "” is needed.", k}
		}
		if err := json.Unmarshal(raw, &texts[i]); err != nil {
			return "", "", &signInRefusal{http.StatusBadRequest, httpx.CodeMalformedRequest, "The field “" + k + "” must be text.", k}
		}
	}
	email, password = texts[0], texts[1]
	if utf8.RuneCountInString(email) > maxEmailChars {
		return "", "", &signInRefusal{http.StatusBadRequest, httpx.CodeValidationFailed, "The email is too long.", signInFieldEmail}
	}
	if len(password) > maxPasswordBytes {
		return "", "", &signInRefusal{http.StatusBadRequest, httpx.CodeValidationFailed, "The password is too long.", signInFieldPassword}
	}
	return email, password, nil
}

// SignIn answers POST /v1/sessions. The checks run in the contract's order;
// the cookie is set only after the session row is stored, and only a 401 from
// the password check counts against the address.
func (s *Service) SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	id := httpx.RequestIDFrom(r.Context())
	addr := clientAddr(r)
	if wait, locked := s.throttle.Check(addr); locked {
		tooManyAttempts(w, id, wait)
		return
	}
	email, password, ref := readSignIn(w, r)
	if ref != nil {
		httpx.WriteCoded(w, ref.status, ref.code, ref.sentence, ref.field)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), signInTimeout)
	defer cancel()
	ok, wait, locked, err := s.attempt(ctx, addr, email, password)
	if err != nil {
		storeFailed(w, r, "password check", err)
		return
	}
	if locked {
		tooManyAttempts(w, id, wait)
		return
	}
	if !ok {
		if s.account.configured {
			slog.Info("sign-in", "request_id", id, "outcome", outcomeBadCredentials)
		} else {
			slog.Warn("admin account is not configured", "request_id", id)
			slog.Info("sign-in", "request_id", id, "outcome", outcomeNotConfigured)
		}
		httpx.WriteCoded(w, http.StatusUnauthorized, httpx.CodeUnauthenticated, sentenceNoMatch, "")
		return
	}

	token := newToken()
	now := s.now()
	if err := s.sessions.Create(ctx, s.account.hashToken(token), s.account.email, now, now.Add(sessionLife)); err != nil {
		storeFailed(w, r, "create session", err)
		return
	}
	s.throttle.Reset(addr)
	slog.Info("sign-in", "request_id", id, "outcome", outcomeOK)
	http.SetCookie(w, sessionCookie(token, now, isHTTPS(r)))
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"email": s.account.email})
}

// tooManyAttempts answers the 429 of a locked address, with Retry-After and
// no cookie.
func tooManyAttempts(w http.ResponseWriter, requestID string, wait time.Duration) {
	slog.Info("sign-in", "request_id", requestID, "outcome", outcomeLocked)
	w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds(wait)))
	httpx.WriteCoded(w, http.StatusTooManyRequests, httpx.CodeRateLimited, sentenceTooMany, "")
}

// attempt is the password check behind the throttle. A caller that opens many
// connections at once gets every request past the first look at the lock,
// before any of them has failed, so the lock is looked at again here, with a
// hashing slot held and before any hashing: a locked address is answered 429
// without a hash, and the slot is not given back until a wrong pair has been
// counted. At most maxHashes-1 attempts that began before the lock engaged
// still answer 401. ok is the pair matching; locked is the lock being on, with
// wait left; err is only a slot that did not come before ctx ended.
func (s *Service) attempt(ctx context.Context, addr, email, password string) (ok bool, wait time.Duration, locked bool, err error) {
	release, err := s.account.acquire(ctx)
	if err != nil {
		return false, 0, false, err
	}
	defer release()
	if wait, locked := s.throttle.Check(addr); locked {
		return false, wait, true, nil
	}
	if s.account.matches(email, password) {
		return true, 0, false, nil
	}
	s.throttle.Fail(addr)
	return false, 0, false, nil
}

// authenticate reports whether the request carries a live session of the
// owner. A cookie of the wrong shape, or any cookie while the account is not
// configured, is "no" without a query. It writes nothing and does not extend
// the session.
func (s *Service) authenticate(r *http.Request) (bool, error) {
	c, err := r.Cookie(CookieName)
	if err != nil || !s.account.configured || !tokenShape.MatchString(c.Value) {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(r.Context(), sessionCheckTimeout)
	defer cancel()
	return s.sessions.Exists(ctx, s.account.hashToken(c.Value), s.account.email, s.now())
}

// Me answers GET /v1/me: the signed-in owner, or 401.
func (s *Service) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	ok, err := s.authenticate(r)
	if err != nil {
		storeFailed(w, r, "me", err)
		return
	}
	if !ok {
		notSignedIn(w)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"email": s.account.email})
}

// SignOut answers DELETE /v1/sessions/current: 204 and a cookie that clears
// the session, whether or not there was one. When the delete itself failed no
// cookie is sent, so the owner is never told they are signed out while the
// session still works.
func (s *Service) SignOut(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(CookieName); err == nil && s.account.configured && tokenShape.MatchString(c.Value) {
		ctx, cancel := context.WithTimeout(r.Context(), sessionCheckTimeout)
		defer cancel()
		if err := s.sessions.Delete(ctx, s.account.hashToken(c.Value)); err != nil {
			storeFailed(w, r, "delete session", err)
			return
		}
	}
	http.SetCookie(w, clearingCookie(isHTTPS(r)))
	w.WriteHeader(http.StatusNoContent)
}

// RequireOwner guards a write: the request goes on only with a live session
// of the owner, and every other case is the one 401. It runs before the body
// is read, so a visitor learns nothing about the records from the answer.
func (s *Service) RequireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, err := s.authenticate(r)
		if err != nil {
			storeFailed(w, r, "session check", err)
			return
		}
		if !ok {
			notSignedIn(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}
