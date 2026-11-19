// Package auth is the shop owner's one account and its sessions: the password
// check, the session cookie, the sign-in throttle and the guard every owner
// write passes through.
//
// The account comes from two secrets, ADMIN_EMAIL and ADMIN_PASSWORD, which
// NewAccount reads once at start. The password itself is never kept: only a
// salted argon2id hash of it, and a fingerprint that keys the session tokens.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"strings"

	"golang.org/x/crypto/argon2"
)

// argon2id parameters (OWASP's second recommended profile): 19 MiB, two
// passes, one lane, a 32-byte output.
const (
	argonTime    uint32 = 2
	argonMemory  uint32 = 19456
	argonThreads uint8  = 1
	argonKeyLen  uint32 = 32

	// maxHashes bounds the argon2id computations running at once: each takes
	// 19 MiB, and sign-in is the one endpoint anyone may call before the
	// throttle has seen a failure.
	maxHashes = 4
)

// Account is the one owner. It holds no password: the hash and its salt, the
// normalized email, and the fingerprint of the pair that session tokens are
// signed with, so that changing either secret ends every old session.
type Account struct {
	configured  bool
	email       string
	emailSum    [32]byte
	salt        []byte
	hash        []byte
	fingerprint []byte
	slots       chan struct{}
}

// Normalize is how an email is compared and answered: no surrounding space,
// no capitals.
func Normalize(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

// NewAccount builds the account from the two secrets, once, before the server
// listens. Either one empty (after trimming white space) leaves the account
// not configured: nobody can sign in, but the public shop still works. A
// not-configured account still computes one hash over a random value, so a
// refused attempt costs what a real one costs.
func NewAccount(email, password string) *Account {
	a := &Account{salt: randomBytes(16), slots: make(chan struct{}, maxHashes)}
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		a.hash = derive(randomBytes(32), a.salt)
		return a
	}
	a.configured = true
	a.email = Normalize(email)
	a.emailSum = sha256.Sum256([]byte(a.email))
	a.hash = derive([]byte(password), a.salt)

	fp := sha256.New()
	fp.Write([]byte("deo-admin-v1"))
	fp.Write([]byte{0})
	fp.Write([]byte(a.email))
	fp.Write([]byte{0})
	fp.Write([]byte(password))
	a.fingerprint = fp.Sum(nil)
	return a
}

// Configured reports whether both secrets were supplied.
func (a *Account) Configured() bool { return a.configured }

func derive(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	// crypto/rand.Read cannot fail since Go 1.24: it crashes the program
	// instead of returning an error.
	_, _ = rand.Read(b)
	return b
}

// Verify checks a typed pair. It always does the same work, so the time does
// not say which half was wrong: the password is hashed whatever the email is,
// and both comparisons run in constant time (the email by its digest, so the
// length of the real one is not leaked). At most maxHashes run at once; a
// request waits for a slot until its own deadline and then gets the error.
func (a *Account) Verify(ctx context.Context, email, password string) (bool, error) {
	select {
	case a.slots <- struct{}{}:
	case <-ctx.Done():
		return false, ctx.Err()
	}
	defer func() { <-a.slots }()

	got := derive([]byte(password), a.salt)
	passwordOK := subtle.ConstantTimeCompare(got, a.hash)
	typed := sha256.Sum256([]byte(Normalize(email)))
	emailOK := subtle.ConstantTimeCompare(typed[:], a.emailSum[:])
	return a.configured && passwordOK&emailOK == 1, nil
}
