package orders

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
)

// placeTimeout covers the whole transaction: the wait for a pool connection,
// every statement and the commit. A deadline is a 500.
const placeTimeout = 10 * time.Second

const (
	sentenceMethod   = "That method is not offered here."
	sentenceTooLarge = "The order is too large."
	sentenceUnread   = "The request could not be read."
	sentenceNotReady = "The service is not ready."
	sentenceInternal = "Something went wrong. Please try again."
)

// Place answers POST /v1/orders. Nothing about an order is ever read back, so
// every other method on this path is 405 with Allow: POST. now is the clock the
// order's time and number's day come from.
func Place(store Store, now func() time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			httpx.WriteCoded(w, http.StatusMethodNotAllowed, httpx.CodeMethodNotAllowed, sentenceMethod, "")
			return
		}
		id := httpx.RequestIDFrom(r.Context())

		// The size is met before the body is read, so a 413 comes before every 400.
		if r.ContentLength > MaxBody {
			refusedSize(w, id)
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBody))
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				refusedSize(w, id)
				return
			}
			slog.Info("order request could not be read", "request_id", id, "error", err)
			refuse(w, id, &Problem{Status: http.StatusBadRequest, Code: httpx.CodeMalformedRequest, Sentence: sentenceUnread})
			return
		}

		req, bot, problem := Parse(r.Header.Get("Content-Type"), body)
		if problem != nil {
			refuse(w, id, problem)
			return
		}
		if bot {
			// Dropped: nothing is validated, held, stored or numbered.
			httpx.WriteJSON(w, http.StatusAccepted, struct{}{})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), placeTimeout)
		defer cancel()
		placed, err := store.Place(ctx, req, now)
		if err != nil {
			var refusal *Refusal
			if errors.As(err, &refusal) {
				refuse(w, id, &Problem{Status: http.StatusConflict, Code: httpx.CodeBikeUnavailable, Sentence: refusal.Sentence})
				return
			}
			failed(w, id, err)
			return
		}
		// The log carries the number and the count, never the customer's words.
		slog.Info("order placed", "request_id", id, "order_number", placed.Number, "lines", len(placed.Lines))
		httpx.WriteJSON(w, http.StatusCreated, placed)
	}
}

func refusedSize(w http.ResponseWriter, id string) {
	refuse(w, id, &Problem{Status: http.StatusRequestEntityTooLarge, Code: httpx.CodePayloadTooLarge, Sentence: sentenceTooLarge})
}

// refuse answers a refusal in the error envelope. The log says why by its
// code; the sentence names the customer's own fields and is not logged.
func refuse(w http.ResponseWriter, id string, p *Problem) {
	slog.Info("order refused", "request_id", id, "code", p.Code)
	httpx.WriteCoded(w, p.Status, p.Code, p.Sentence, p.Field)
}

// failed answers a failure of the store: 503 only when the database cannot be
// reached, 500 for everything else. The cause is logged, never answered.
func failed(w http.ResponseWriter, id string, err error) {
	if errors.Is(err, context.Canceled) {
		slog.Info("order abandoned by the caller", "request_id", id, "error", err)
	} else {
		slog.Error("order failed", "request_id", id, "error", err)
	}
	if httpx.Unavailable(err) {
		httpx.WriteCoded(w, http.StatusServiceUnavailable, httpx.CodeUnavailable, sentenceNotReady, "")
		return
	}
	httpx.WriteCoded(w, http.StatusInternalServerError, httpx.CodeInternal, sentenceInternal, "")
}
