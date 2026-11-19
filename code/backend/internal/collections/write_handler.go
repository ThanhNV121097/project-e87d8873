package collections

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
)

// Guard wraps a handler that must only run for the signed-in owner. The auth
// package provides it; collections does not import auth.
type Guard func(http.Handler) http.Handler

const (
	// writeTimeout bounds the statements of one write: a deadline is a 500.
	writeTimeout = 10 * time.Second
	// maxBody is the largest write body read.
	maxBody = 1 << 20

	sentenceTooLarge      = "The request is too large."
	sentenceIDTaken       = "That id is already used."
	sentenceStale         = "This record changed while you were editing it — for example an order was placed. Reload and make your change again."
	sentenceNeedIfMatch   = "The save did not say which version of the record it was based on. Reload and make your change again."
	sentenceBadIfMatch    = "The If-Match header must be the version this record was read at, for example \"7\"."
	sentenceUnreadable    = "The request must be a JSON object."
	idCollisionAttempts   = 5
	deleteAttempts        = 3
	idAlphabet            = "abcdefghijklmnopqrstuvwxyz0123456789"
	madeIDChars           = 6
	foreignKeyViolationSQ = "23503"
)

var ifMatchShape = regexp.MustCompile(`^"[1-9][0-9]*"$`)

// refuse answers a refused write: one sentence in the error envelope, and one
// Info line (never the body, a field of the record or the sentence).
func refuse(w http.ResponseWriter, r *http.Request, s Spec, ref *Refusal) {
	slog.Info("write refused", "request_id", httpx.RequestIDFrom(r.Context()), "collection", s.Name, "code", ref.Code)
	httpx.WriteCoded(w, ref.Status, ref.Code, ref.Sentence, ref.Field)
}

func notFound(w http.ResponseWriter, r *http.Request, s Spec) {
	refuse(w, r, s, &Refusal{http.StatusNotFound, httpx.CodeNotFound, sentenceNoRecord, ""})
}

// writeFailed answers a failed statement: 503 only when the database cannot be
// reached, 500 for everything else — a CHECK or unique violation, a renamed
// table and a deadline included. The real error is logged with the request id
// and the collection and never sent.
func writeFailed(w http.ResponseWriter, r *http.Request, s Spec, op string, err error) {
	attrs := []any{"request_id", httpx.RequestIDFrom(r.Context()), "collection", s.Name, "op", op, "error", err}
	if errors.Is(err, context.Canceled) {
		slog.Info("write abandoned by the caller", attrs...)
	} else {
		slog.Error("write failed", attrs...)
	}
	if httpx.Unavailable(err) {
		httpx.WriteCoded(w, http.StatusServiceUnavailable, httpx.CodeUnavailable, sentenceNotReady, "")
		return
	}
	httpx.WriteCoded(w, http.StatusInternalServerError, httpx.CodeInternal, sentenceInternal, "")
}

// isForeignKey reports SQLSTATE 23503. On a bike's write it means the riding
// type was deleted in the instant after it was checked; on a riding type's
// delete it means a bike uses it.
func isForeignKey(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == foreignKeyViolationSQ
}

// readBody reads the whole body through the size limit before anything is
// parsed: an oversize body is a 413, which comes before every 400.
func readBody(w http.ResponseWriter, r *http.Request) ([]byte, *Refusal) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	data, err := io.ReadAll(r.Body)
	if err == nil {
		return data, nil
	}
	var tooBig *http.MaxBytesError
	if errors.As(err, &tooBig) {
		return nil, &Refusal{http.StatusRequestEntityTooLarge, httpx.CodePayloadTooLarge, sentenceTooLarge, ""}
	}
	return nil, &Refusal{http.StatusBadRequest, httpx.CodeMalformedRequest, sentenceUnreadable, ""}
}

// written answers an accepted add or replace: the version as the ETag and the
// row the statement returned as the body, written as raw JSON so no number is
// ever decoded into a float.
func written(w http.ResponseWriter, r *http.Request, s Spec, status int, rec []byte, version int64, id, op string) {
	slog.Info("record saved", "request_id", httpx.RequestIDFrom(r.Context()), "collection", s.Name, "id", id, "op", op)
	w.Header().Set("ETag", `"`+strconv.FormatInt(version, 10)+`"`)
	httpx.WriteJSON(w, status, json.RawMessage(rec))
}

// newID makes a record id: the collection's name, a hyphen and six characters
// of a-z0-9 from crypto/rand, by rejection sampling so no character is
// likelier than another.
func newID(collection string) string {
	const limit = 256 - 256%len(idAlphabet)
	out := make([]byte, 0, madeIDChars)
	buf := make([]byte, 16)
	for len(out) < madeIDChars {
		// crypto/rand.Read cannot fail since Go 1.24.
		_, _ = rand.Read(buf)
		for _, b := range buf {
			if int(b) < limit && len(out) < madeIDChars {
				out = append(out, idAlphabet[int(b)%len(idAlphabet)])
			}
		}
	}
	return collection + "-" + string(out)
}

// addHandler answers POST /v1/{collection}: validate, check the riding type,
// insert. A body id is tried once; a server-made id is retried on a collision.
func addHandler(s Spec, wr Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ref := readBody(w, r)
		if ref != nil {
			refuse(w, r, s, ref)
			return
		}
		rec, ref := Decode(s, r.Header.Get("Content-Type"), body, "", false)
		if ref != nil {
			refuse(w, r, s, ref)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), writeTimeout)
		defer cancel()

		if rec.Category != nil {
			ok, err := wr.CategoryExists(ctx, *rec.Category)
			if err != nil {
				writeFailed(w, r, s, "add", err)
				return
			}
			if !ok {
				refuse(w, r, s, noCategory())
				return
			}
		}
		attempts := 1
		if rec.ID == "" {
			attempts = idCollisionAttempts
		}
		for i := 0; i < attempts; i++ {
			id := rec.ID
			if id == "" {
				id = newID(s.Name)
			}
			out, version, added, err := wr.Add(ctx, s, id, rec.Values)
			switch {
			case isForeignKey(err):
				refuse(w, r, s, noCategory())
				return
			case err != nil:
				writeFailed(w, r, s, "add", err)
				return
			case added:
				written(w, r, s, http.StatusCreated, out, version, id, "add")
				return
			}
			if rec.ID != "" {
				refuse(w, r, s, &Refusal{http.StatusConflict, httpx.CodeIDTaken, sentenceIDTaken, ""})
				return
			}
		}
		writeFailed(w, r, s, "add", fmt.Errorf("%d made ids were all taken", attempts))
	})
}

// ifMatchVersion reads the version the caller's save was based on from the
// If-Match header: exactly one line, exactly a quoted positive integer.
func ifMatchVersion(r *http.Request) (int64, *Refusal) {
	values := r.Header.Values("If-Match")
	if len(values) == 0 {
		return 0, &Refusal{http.StatusPreconditionRequired, httpx.CodePreconditionRequired, sentenceNeedIfMatch, ""}
	}
	bad := &Refusal{http.StatusBadRequest, httpx.CodeMalformedRequest, sentenceBadIfMatch, ""}
	token := strings.TrimSpace(values[0])
	if len(values) > 1 || !ifMatchShape.MatchString(token) {
		return 0, bad
	}
	version, err := strconv.ParseInt(strings.Trim(token, `"`), 10, 64)
	if err != nil {
		return 0, bad
	}
	return version, nil
}

// changeHandler answers PUT /v1/{collection}/{id}: the body replaces the
// record, and only if nothing has changed it since the caller read it.
func changeHandler(s Spec, wr Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		body, ref := readBody(w, r)
		if ref != nil {
			refuse(w, r, s, ref)
			return
		}
		rec, ref := Decode(s, r.Header.Get("Content-Type"), body, id, true)
		if ref != nil {
			refuse(w, r, s, ref)
			return
		}
		if !cleanText(id) {
			// No record can have an id the database cannot store.
			notFound(w, r, s)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), writeTimeout)
		defer cancel()

		_, found, categoryOK, err := wr.Lookup(ctx, s, id, rec.Category)
		if err != nil {
			writeFailed(w, r, s, "replace", err)
			return
		}
		if !found {
			notFound(w, r, s)
			return
		}
		version, ref := ifMatchVersion(r)
		if ref != nil {
			refuse(w, r, s, ref)
			return
		}
		if !categoryOK {
			refuse(w, r, s, noCategory())
			return
		}
		out, newVersion, replaced, err := wr.Replace(ctx, s, id, version, rec.Values)
		switch {
		case isForeignKey(err):
			refuse(w, r, s, noCategory())
			return
		case err != nil:
			writeFailed(w, r, s, "replace", err)
			return
		case replaced:
			written(w, r, s, http.StatusOK, out, newVersion, id, "replace")
			return
		}
		// Nothing matched: the record is gone, or somebody changed it since the
		// caller read it. An order's stock hold counts as somebody.
		_, found, err = wr.Version(ctx, s, id)
		switch {
		case err != nil:
			writeFailed(w, r, s, "replace", err)
		case !found:
			notFound(w, r, s)
		default:
			refuse(w, r, s, &Refusal{http.StatusConflict, httpx.CodeStaleRecord, sentenceStale, ""})
		}
	})
}

// removeHandler answers DELETE /v1/{collection}/{id}. No body and no If-Match
// are read: a delete is not refused for being stale. A riding type that bikes
// use is refused with a sentence that names how many.
func removeHandler(s Spec, wr Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !cleanText(id) {
			notFound(w, r, s)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), writeTimeout)
		defer cancel()

		for i := 0; i < deleteAttempts; i++ {
			found, err := wr.Remove(ctx, s, id)
			if err == nil {
				if !found {
					notFound(w, r, s)
					return
				}
				slog.Info("record deleted", "request_id", httpx.RequestIDFrom(r.Context()), "collection", s.Name, "id", id, "op", "delete")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if !isForeignKey(err) {
				writeFailed(w, r, s, "delete", err)
				return
			}
			name, bikes, exists, err := wr.InUse(ctx, id)
			if err != nil {
				writeFailed(w, r, s, "delete", err)
				return
			}
			if !exists {
				notFound(w, r, s) // deleted in the instant between
				return
			}
			if bikes >= 1 {
				noun := "bikes"
				if bikes == 1 {
					noun = "bike"
				}
				refuse(w, r, s, &Refusal{http.StatusConflict, httpx.CodeInUse,
					fmt.Sprintf("%s is still the riding type of %d %s. Give those bikes another riding type first.", name, bikes, noun), ""})
				return
			}
			// The last bike moved away in the instant between: try the delete again.
		}
		writeFailed(w, r, s, "delete", fmt.Errorf("riding type %s kept changing under its delete", id))
	})
}
