package auth

import (
	"math"
	"sync"
	"time"
)

const (
	// throttleWindow is how long a failure counts, and how long an address is
	// locked after its fifth.
	throttleWindow = 15 * time.Minute
	// throttleLimit is the number of failures inside the window that locks.
	throttleLimit = 5
	// throttleMaxEntries bounds the table: above it the entry that failed
	// longest ago is evicted.
	throttleMaxEntries = 10000
)

type attempts struct {
	failures    []time.Time
	last        time.Time
	lockedUntil time.Time
}

// Throttle counts failed sign-ins per client address, in memory: five
// failures inside 15 minutes lock the address until 15 minutes after the
// fifth. The clock is injected so a test can move it. The table is lost on
// restart, which is acceptable for an owner-only door.
type Throttle struct {
	mu      sync.Mutex
	now     func() time.Time
	entries map[string]*attempts
}

// NewThrottle returns an empty throttle reading time from now.
func NewThrottle(now func() time.Time) *Throttle {
	return &Throttle{now: now, entries: map[string]*attempts{}}
}

// Check reports whether addr is locked and, if so, how long is left.
func (t *Throttle) Check(addr string) (time.Duration, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	t.sweep(now)
	e, ok := t.entries[addr]
	if !ok {
		return 0, false
	}
	if e.lockedUntil.After(now) {
		return e.lockedUntil.Sub(now), true
	}
	if !e.lockedUntil.IsZero() {
		// The lock is over: 15 minutes after the latest failure.
		delete(t.entries, addr)
	}
	return 0, false
}

// Fail records one failed sign-in. A request that was answered 429 never
// reaches here, so a locked address cannot extend its own lock.
func (t *Throttle) Fail(addr string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	t.sweep(now)
	e, ok := t.entries[addr]
	if !ok {
		e = &attempts{}
		t.entries[addr] = e
	}
	kept := e.failures[:0]
	for _, f := range e.failures {
		if now.Sub(f) < throttleWindow {
			kept = append(kept, f)
		}
	}
	e.failures = append(kept, now)
	e.last = now
	if len(e.failures) >= throttleLimit {
		e.lockedUntil = now.Add(throttleWindow)
	}
	for len(t.entries) > throttleMaxEntries {
		t.evictOldest()
	}
}

// Reset forgets addr: a successful sign-in clears the count.
func (t *Throttle) Reset(addr string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, addr)
}

// Len is the number of addresses the table holds.
func (t *Throttle) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.entries)
}

// sweep drops every entry whose last failure is 15 minutes old or older and
// that is not locked. The caller holds the mutex.
func (t *Throttle) sweep(now time.Time) {
	for addr, e := range t.entries {
		if now.Sub(e.last) >= throttleWindow && !e.lockedUntil.After(now) {
			delete(t.entries, addr)
		}
	}
}

func (t *Throttle) evictOldest() {
	var (
		oldest string
		at     time.Time
		found  bool
	)
	for addr, e := range t.entries {
		if !found || e.last.Before(at) {
			oldest, at, found = addr, e.last, true
		}
	}
	if found {
		delete(t.entries, oldest)
	}
}

// retryAfterSeconds is the Retry-After value for a lock with d left: whole
// seconds, rounded up, from 1 to 900.
func retryAfterSeconds(d time.Duration) int {
	s := int(math.Ceil(d.Seconds()))
	if s < 1 {
		return 1
	}
	if max := int(throttleWindow.Seconds()); s > max {
		return max
	}
	return s
}
