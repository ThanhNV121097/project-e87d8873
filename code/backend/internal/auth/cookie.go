package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// CookieName is the session cookie.
const CookieName = "deo_session"

// sessionLife is how long a session lasts from its sign-in. It is fixed:
// nothing extends it.
const sessionLife = 7 * 24 * time.Hour

// tokenShape is what a cookie value must look like before the store is asked
// about it: 32 random bytes in base64url without padding.
var tokenShape = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

func newToken() string {
	b := make([]byte, 32)
	// crypto/rand.Read cannot fail since Go 1.24.
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// hashToken is what the sessions table stores for a cookie: an HMAC of the
// cookie's text keyed by the account's fingerprint. A copy of the table is no
// use without the secrets, and changing either secret makes every stored hash
// unreachable.
func (a *Account) hashToken(token string) []byte {
	m := hmac.New(sha256.New, a.fingerprint)
	m.Write([]byte(token))
	return m.Sum(nil)
}

// lastOf returns the last entry of a comma-separated header, however many
// lines carry it: the one the edge proxy appended.
func lastOf(r *http.Request, header string) string {
	values := strings.Join(r.Header.Values(header), ",")
	parts := strings.Split(values, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		if p := strings.TrimSpace(parts[i]); p != "" {
			return p
		}
	}
	return ""
}

// isHTTPS reports whether the request reached the edge over HTTPS.
func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(lastOf(r, "X-Forwarded-Proto"), "https")
}

// clientAddr is the throttle's key: the last X-Forwarded-For entry (the edge's),
// else the peer address without its port. It is never logged.
func clientAddr(r *http.Request) string {
	if addr := lastOf(r, "X-Forwarded-For"); addr != "" {
		return addr
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func sessionCookie(token string, now time.Time, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionLife.Seconds()),
		Expires:  now.Add(sessionLife),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	}
}

// clearingCookie tells the browser to drop the session (Max-Age=0).
func clearingCookie(secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	}
}
