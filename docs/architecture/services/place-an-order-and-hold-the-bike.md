# Service contract extension — Place an order and hold the bike

Story: `docs/orders/stories/place-an-order-and-hold-the-bike.md` · Requirements: ORD-001 to ORD-008 (`docs/orders/SRS.md`) · Plan item: *Place an order and hold the bike* (feature, P1) · Companion: `docs/architecture/erd/place-an-order-and-hold-the-bike.md`

## 1. Result

**This story adds no endpoint, no request or response field, no error code and no pagination or auth rule that the contract lacks.** It builds the one endpoint `docs/architecture/services.md` already specifies — §3.10 `POST /v1/orders` — with its body, its answer, its error table and its sentences, and it is called by the frontend module `src/lib/orders.ts` that PR #17 already wrote. Nothing of §3.10, §2.3 (the envelope), §2.5 (the validation boundary) or §2.6 (no idempotency) is restated here.

There is no mock module to compare with the API: PR #17 deleted `src/lib/mock/orders.ts`, and its `placeOrder` is the API client. §7 reads that client against the contract; it departs from nothing and asks for no frontend change.

What this file adds, because the contract leaves it to the builder:

| § | What |
|---|---|
| 2 | The route, and what every other request to an `orders` path gets |
| 3 | Reading the request, step by step: the order of the checks, the grammar of `qty`, the sentence and `field` of every 400 the contract does not spell |
| 4 | The decision and the transaction: the line walk, the pricing, the answer, what cannot happen |
| 5 | Failures, deadlines, headers and logging |
| 6 | Package layout and what is tested without a database |
| 7 | How the frontend speaks the contract (PR #17 read against it) |
| 8 | Conflicts with merged decisions, resolved |
| 9 | Rollout, dependencies, open questions |

## 2. Route

In `internal/server.Router`, one line beside the others (no method in the pattern, so the handler, not the mux, decides what a method gets, as the data story's two lines do):

```go
mux.HandleFunc("/v1/orders", orders.Place(orders.NewStore(db), time.Now))
```

`server.DB` **grows by one method**, `Begin(ctx context.Context) (pgx.Tx, error)`, which `*pgxpool.Pool` already has, because an order is the first thing that needs a transaction and `QueryRow` cannot give one. The fake in `server_test.go` gains the same method (one line). `cmd/api/main.go` still calls `server.Router(pool)`.

| Request | Answer |
|---|---|
| `POST /v1/orders` | §3.10 |
| `GET`, `HEAD`, `OPTIONS`, `PUT`, `PATCH`, `DELETE` on exactly `/v1/orders` | `405 METHOD_NOT_ALLOWED`, `Allow: POST`, "That method is not offered here." No list of orders is served (SC-6) |
| Any method on `/v1/orders/<anything>` (`GET /v1/orders/DEO-261116-0001`, `PUT`, `DELETE`, `POST /v1/orders/N`) | Not this route: the literal `/v1/orders` does not match a longer path, so `/v1/{collection}/{id}` answers `404 NOT_FOUND` "That collection does not exist." for every method (an unknown collection is checked before the method; data story §2). Nothing about any order is ever returned |
| `POST /v1/orders/` (trailing slash) | `404 NOT_FOUND` "That address does not exist." by the catch-all |

SC-6 accepts `404` or `405` for all of these; this table is the one right answer for each, so a test may pin it.

## 3. Reading the request

One function, `Parse`, takes the method's body and `Content-Type` and returns a typed request, or the first refusal, or the bot flag. Everything after it trusts the typed value (`services.md` §2.5). It runs the contract's five steps in this order; the first failure answers.

**Step 1 — size.** A declared `Content-Length` above 32,768 is refused without reading. Otherwise the body is read completely through `http.MaxBytesReader(w, r.Body, 32768)` **before any parsing**, so a `413` comes before every `400`: `*http.MaxBytesError` → `413 PAYLOAD_TOO_LARGE`, "The order is too large."

**Step 2 — content type and object.** `mime.ParseMediaType` of `Content-Type` must give `application/json` (parameters such as `charset=utf-8` allowed, compared without regard to case); otherwise `400 MALFORMED_REQUEST`, "The request must be sent as JSON." The body's first byte that is not white space must be `{`, and after the one JSON value only white space may follow; otherwise `400 MALFORMED_REQUEST`, "The request must be a JSON object." — one sentence for an array, a string, a number, `null`, an empty body, malformed JSON and trailing data (SC-20). The byte test is needed because decoding `null` into a map succeeds silently. The object is decoded into `map[string]json.RawMessage`; a duplicate key keeps the last value.

**Step 3 — bot.** If the key `website` is present and is a JSON string of at least one character — **any** text, not trimmed, so a value of one space is still a bot — the answer is `202` with the body `{}`, and nothing else is looked at: no other check, no database, no number (SC-19). An absent `website` or `""` goes on. A `website` that is not a string is a shape error (step 4).

**Step 4 — shape, `400 MALFORMED_REQUEST`.** A body this contract does not define is refused, and the advisory keys of §3.10 are the one carve-out. `DisallowUnknownFields` cannot be used as it stands: the bot check must run before the shape check, and the advisory keys must pass without being read. So the pass is written by hand, in this order, the first failure answering:

1. *Top-level undefined keys*, in the alphabetical order of their names; `number`, `placed`, `subtotal`, `delivery_fee` and `total` are skipped and **never decoded**, so `"total": "not a number"` and `"number": null` are accepted (SC-7, SC-35).
2. *Top-level types*: `name`, `phone`, `delivery`, `address`, `payment`, `note` and `website` must each be a JSON string when present; `lines` must be an array when present. `null` is the wrong type everywhere: a key sent as `null` is not the same as an absent key.
3. *Each line, in the order sent*: it must be an object; undefined keys alphabetically (`name`, `price` and `image` are skipped, never decoded, so `"price": null` is accepted); then `id` and `variant` must be strings when present, `qty` must be a JSON number when present (a string, `true`, `null`, an array or an object is the wrong type: `"2"` is `MALFORMED_REQUEST`, SC-14).

An **absent** key is never a shape error: it is caught in step 5, so a body with no `name` answers "Your name is needed.".

| Shape error | Sentence | `field` |
|---|---|---|
| a top-level key the contract does not define | "The field “<key>” is not part of an order." | `<key>` |
| a line key the contract does not define | "The field “<key>” is not part of an order line." | `lines[i].<key>` |
| a text field that is not a string | "The field “<key>” must be text." | `<key>` |
| `lines` is not an array | "The field “lines” must be a list." | `lines` |
| a line is not an object | "Line N must be an object." | `lines[i]` |
| `id` or `variant` of a line is not a string | "The field “<key>” of line N must be text." | `lines[i].<key>` |
| `qty` is not a JSON number | "The quantity of line N must be a number." | `lines[i].qty` |

`N` is the one-based number of the line in the sentence and `i` the zero-based index in `field`, as in the contract. Tests match the status, the code and the quoted key words.

**Step 5 — values, `400 VALIDATION_FAILED`.** Every text is trimmed with `strings.TrimSpace` (Unicode white space) and counted in code points (`services.md` §2.5, rule 4: a name of 100 letters `Đ` is 100 characters and 200 bytes, and is accepted). The checks run in the contract's order — each field takes its turn, and within a field the blank check, then the NUL check, then the length check:

| # | Check | Sentence | `field` |
|---|---|---|---|
| 1 | `name` absent or blank | "Your name is needed." | `name` |
| 2 | `phone` absent or blank | "Your phone number is needed." | `phone` |
| 3 | `delivery` absent or not `local`, `courier`, `pickup` | "The delivery method must be local, courier or pickup." | `delivery` |
| 4 | `payment` absent or not `transfer`, `cod`, `shop` | "The payment method must be transfer, cod or shop." | `payment` |
| 5 | `payment` is `cod` and `delivery` is not `local` | "Cash on delivery is only available for delivery in town." | `payment` |
| 6 | `delivery` is not `pickup` and `address` is absent or blank | "A delivery address is needed." | `address` |
| 7 | `note` (optional): only its NUL and length checks apply | see below | `note` |
| 8 | `lines` absent or empty | "Your bag is empty." | `lines` |
| 9 | `lines` has more than 20 entries | "An order holds at most 20 lines." | `lines` |
| 10 | each line in the order sent: `id` absent, blank or over 200 characters | "Line N needs a bike." | `lines[i].id` |
| 11 | … then `variant` (optional): only its NUL and length checks apply | see below | `lines[i].variant` |
| 12 | … then `qty` absent, or not a whole number from 1 to 10 | "The quantity of line N must be a whole number from 1 to 10." | `lines[i].qty` |

Too long, wherever it comes up in a field's turn (`name` over 100, `phone` over 30, `address` over 300, `note` over 1,000, a `variant` over 20): "A text is too long: name 100, phone 30, address 300, note 1000, size 20 characters." with `field` naming that field. A text with a NUL character (U+0000) in any of `name`, `phone`, `address`, `note`, `variant`, `id`: "The text in “<field>” contains a character that cannot be stored." — Postgres cannot store it and would otherwise answer a 500. `address` and `note` are optional: absent means `""`; `variant` absent or blank means `""`; an `id` is **not** trimmed (it is compared exactly, so an id with a space is simply a bike that does not exist, a 409). The table is the contract's order (`name`, `phone`, `delivery`, `payment`, cash on delivery, `address`, `note`, `lines`, then each line's `id`, `variant`, `qty`) with every turn written out; the NUL and length checks of `name`, `phone` and `address` run inside their own turns (rows 1, 2 and 6).

**`qty` grammar.** The raw number token must match `^-?(0|[1-9][0-9]*)$` and its value must be 1 to 10; anything else that is a JSON number is `VALIDATION_FAILED` — `0`, `11`, `-1`, `1.5`, `2.0`, `1e0`, `00`, and an integer too large for 64 bits (never an overflow, never a 500; SC-14).

**Trim rule for the answer.** The stored and answered `name`, `phone`, `address` are the trimmed text (`" Minh "` is `"Minh"`, SC-18); `note` is stored trimmed and never answered.

## 4. The decision and the transaction

The contract (§3.10) lists the transaction's steps and the four refusals; the statements and their locks are in the ERD extension §2. What this section pins is how the decision is written, so that it is one pure function a unit test can run without a database.

**`Decide(request, bikes, courierFee)`** takes the typed request, the bikes S1 returned (a map by `id`) and the courier charge (0 when the order is not a courier order, or the `shop` row has no charge), and returns either the priced order or the first refusal. It never touches the database. The walk, in the order the lines were sent:

```
held := map[id]int64{}
for i, line := range request.Lines:
    bike, ok := bikes[line.ID]
    if !ok:                      refuse("A bike in your bag is no longer in the shop.")
    label := trim(bike.Brand + " " + bike.Name)             // the brand is left out when the bike has none
    if bike.Status != "live":    refuse(label + " is no longer in the shop.")
    held[line.ID] += line.Qty                                // the quantities of one bike are counted together
    stock := bike.Stock (a NULL stock is 0)
    if held[line.ID] > stock:
        stock == 0 ? refuse(label + " is out of stock.") : refuse(label + ": only <stock> in stock.")   // <stock> is the whole stock
    if len(bike.Sizes) > 0:
        line.Variant == "" ? refuse(label + " needs a frame size.")
        line.Variant not in bike.Sizes, or in bike.SizesOut ? refuse(label + " is not available in size <variant>.")
    else: the stored size is ""                              // a bike with no sizes needs none; a variant sent for it is ignored
```

These sentences are the contract's and are **fixed** (ORD-004). The checks for one line run in that order, so a line that fails on stock and on size answers the stock sentence. A quantity equal to the stock passes. `<variant>` is the trimmed text sent, compared exactly (capitals count); it is echoed into a JSON string by the encoder, and the frontend renders the sentence as text, so it cannot inject markup.

After the walk, each line's price is the bike's `price` read in S1: `subtotal` is the sum of `price × qty`, `delivery_fee` is the courier charge when the delivery is `courier` and 0 otherwise, `total` is their sum (ORD-002). A bike that is `live` always has a price above 0 (`ck_products_live_has_price`); a `NULL` or `0` price here is a broken row, not a customer's mistake, and is `INTERNAL` (logged with the bike's id), never a priced line.

`Decide` is followed by S3 to S6 (ERD extension §2). Three things cannot happen under the locks of S1, and each is `INTERNAL`, logged at `Error` with the constraint or the count and never shown: S3 changes a number of rows other than the number of distinct bikes; `uq_orders_number` is violated; any `CHECK` fires. A `CHECK` that fires means a validator is missing a rule (`services.md` §2.5).

**The number.** `Number(day, seq)` is `"DEO-" + day.Format("060102") + "-" + fmt.Sprintf("%04d", seq)`, with `day` the date of the clock instant in Vietnam time and `seq` the `last_seq` that S4 returns. The clock is `time.Now`, read **once**, after `BEGIN`, and the same instant is the `placed_at` of S5 and the day of S4 (ERD extension C1, C2). The 10,000th order of a day fails `ck_order_counters_last_seq_range`, rolls back and answers 500 (`erd.md` §3.11).

**The answer.** The `201` body is built from the values S1 read and S3 to S6 wrote, never from a second read (ERD extension C4). A struct with these JSON keys and no others: `number`, `placed`, `name`, `phone`, `delivery`, `address`, `payment`, `lines`, `subtotal`, `delivery_fee`, `total` — `address` always present (`""` for a pickup with none), `note` never. A line is `id`, `name`, `price`, `qty` always, `image` only when the bike has one, `variant` only when the stored size is not `""`. `placed` is the instant in Vietnam time written `2006-01-02T15:04:05-07:00`, which prints `+07:00`. Lines are in the order sent.

## 5. Failures, deadlines, headers, logging

| The error is | Answer |
|---|---|
| A refusal of `Decide` | `409 BIKE_UNAVAILABLE`, the sentence of §4. Nothing was written: S3 had not run |
| A refused connection (`syscall.ECONNREFUSED` anywhere in the chain, or a dial `*net.OpError`), or a `*pgconn.PgError` with SQLSTATE of class `08` or `57P01`, `57P02`, `57P03` | `503 UNAVAILABLE`, "The service is not ready." |
| Anything else: a failed statement, a `CHECK` or unique violation, a deadline, a cancelled request, a scan error, a commit that failed | `500 INTERNAL`, "Something went wrong. Please try again." |

The two classes use the **one** predicate of the data story (`unavailable(err)`, `docs/architecture/services/site-data-from-the-database.md` §5.1). If it is private to `internal/collections` when this story starts, move it to `internal/httpx` with the same body and have both packages call it; do not write a second copy. This story adds the four code constants it needs to `httpx` beside the data story's five: `CodeMalformedRequest`, `CodeValidationFailed`, `CodeBikeUnavailable`, `CodePayloadTooLarge`.

- **Deadline.** `context.WithTimeout(r.Context(), 10*time.Second)` covers the whole of §4: the wait for a pool connection, every statement and the commit (`services.md` §5, PostgreSQL row). A deadline is a `500`, as for the reads.
- **Rollback.** Any error or refusal after `BEGIN` ends in `ROLLBACK`, issued on a fresh context (`context.WithoutCancel` plus 2 s), so that a request the visitor abandoned releases its row locks at once instead of when the connection is reaped. A request cancelled before `COMMIT` stores nothing.
- **A commit whose outcome is unknown** (the error comes from `COMMIT` itself) answers `500`, although the order may exist. This is the lost-answer case the SRS already accepts (orders SRS §7, question 2); the log line carries the order number so the operator can find it.
- **A panic** is turned into the generic `500` by the existing `Recover`, and the transaction is rolled back by the deferred `ROLLBACK` before it propagates.
- **Headers.** Every answer, the `202` and every error included, carries `Cache-Control: no-store` and `X-Request-Id` from the middleware chain, and `Content-Type: application/json; charset=utf-8`. A `405` carries `Allow`. There is no `Location` and no `CORS` header (`services.md` §2.8).
- **Logging.** One `Info` line per accepted order from the handler: `msg: "order placed"`, `request_id`, `order_number`, `lines` (a count). One `Info` line per refusal: `msg: "order refused"`, `request_id`, `code`. Failures are logged at `Error` with the real cause and `request_id`. **Never logged**: the body, the name, phone, address and note, the query, and a refusal's sentence is not logged, because the code already says why. A request abandoned by the visitor (`context.Canceled`) is logged at `Info`.
- **No retry, no idempotency key, no rate limit** (`services.md` §2.6, §9 question 2).

## 6. Layout and what is tested without a database

CI has no database, so everything that can be decided before SQL is a pure function with a unit test, the transaction's steps sit behind a small interface a fake can fail, and the statements themselves are proved on the running stack.

| File | Holds |
|---|---|
| `internal/orders/request.go` | `Request`, `Line`, and `Parse` — steps 2 to 5 of §3 on the bytes and the content type; returns the request, or a `Problem{Status, Code, Sentence, Field}`, or the bot flag |
| `internal/orders/decide.go` | `Bike`, `Priced`, `Refusal`, `Decide` (§4): pure |
| `internal/orders/number.go` | the fixed Vietnam zone, `Day(instant)`, `Number(day, seq)` |
| `internal/orders/store.go` | `Store` (an interface: `Place(ctx, Request, now time.Time) (Placed, error)`), `NewStore(db)`, and `txn`, the interface of one transaction's steps (`LockBikes`, `CourierFee`, `HoldStock`, `NextSeq`, `InsertOrder`, `InsertLines`, `Commit`, `Rollback`) with the pgx implementation of the six statements of the ERD extension §2 |
| `internal/orders/handler.go` | `Place(store, now)`: method check, step 1 of §3, the bot answer, `Store.Place`, the mapping of §5, the log lines |
| `internal/httpx/httpx.go` | the four code constants; `unavailable` moved here if it was private (§5) |
| `internal/server/server.go`, `server_test.go` | the one route line; `DB` gains `Begin`; the fake gains it |
| `migrations/20261118000300_orders.{up,down}.sql`, `migrations_test.go` | ERD extension §4, §7 |

The tests (new files; existing tests change only where the table says):

| Test | Proves (story criteria) |
|---|---|
| `Parse`, table-driven: every row of §3 with its status, `code`, `field` and key words; **two rules broken at once** answers the one the order of §3 says (a shape error before a value error: `qty: "2"` with a blank name is `MALFORMED_REQUEST`); `qty` of 0, 1, 10, 11, -1, 1.5, 2.0, 1e0, 00, a 25-digit integer, `"2"`, `null`; lines 0, 1, 20, 21; text limits at and one past (name 100/101, phone 30/31, address 300/301, note 1000/1001, size 20/21), counted in code points (100 × `Đ` accepted); `" Minh "` returns `Minh`; NUL in each text; the advisory keys (`total`, `number`, line `price: null`) accepted and absent from the result; an undefined key at each level; `cod` with each delivery; an address on `pickup` kept; no address on `pickup` accepted; the bot flag for `"x"`, `" "` and not for `""` or absent; bodies `[]`, `"x"`, `null`, `` (empty), `{`, trailing `{}{}`; a wrong `Content-Type`; a body of 32,768 and of 32,769 bytes | SC-7, SC-12 to SC-20, SC-35 |
| `Decide`: each refusal with its exact sentence (not live, unknown id, out of stock, "only N in stock", size not in `sizes`, size in `sizes_out`, no size); stock equal to the quantity passes; two lines of one bike are counted together (stock 3, lines of 2 and 2 → "only 3 in stock"); the first failing line is the one named, and a refusal on line 2 names line 2's bike; a bike with no sizes accepted with a variant sent and with none (stored size `""`); a `NULL` stock is "out of stock"; a bike with no brand is labelled by its name; subtotal, courier fee (`courier` only; a missing charge is 0) and total for the seed's numbers; a `NULL` price is `INTERNAL`, never a line | SC-8, SC-10, SC-22 to SC-27 |
| `Day` and `Number`: `2026-11-16T18:00:00Z` gives `261117`; `2026-11-16T16:59:59Z` gives `261116`, `…T17:00:00Z` gives `261117`; the first and second numbers of a day; the rendered `placed` of a UTC instant ends `+07:00` | SC-32, SC-33 |
| Handler with a fake `Store`: `201` has **exactly** the eleven keys (no `note`) and `placed` parses with an offset; a line has `image` and `variant` only when set; `202` with `{}` for a bot and the store **not called**; `409` is the envelope with `BIKE_UNAVAILABLE` and the sentence; any other store error is `500` with the generic sentence and no SQL text, a refused connection is `503`; the method table of §2 (`GET`, `PUT`, `DELETE`, `PATCH`, `HEAD` on `/v1/orders` are `405` with `Allow: POST`; `GET`, `PUT`, `DELETE` on `/v1/orders/DEO-261116-0001` are `404`); every answer carries `Cache-Control: no-store` and `X-Request-Id`; no log line from the handler contains the name, phone, address or note of the request (the test captures the logger) | SC-1 (shape), SC-2, SC-3, SC-6, SC-19, SC-21 (HTTP part), SC-43 |
| `Store` with a fake `txn`: the steps run in the contract's order (lock, fee, hold, number, order, lines, commit); a refusal returns before `HoldStock`; **an error from each step in turn** (`HoldStock`, `NextSeq`, `InsertOrder`, `InsertLines`, `Commit`) calls `Rollback` and never `Commit`; the clock is read once and the same instant reaches `InsertOrder` and the day of `NextSeq`; the fee is read only for a courier order | SC-21, SC-31 (as far as CI can see), SC-32, SC-33 |
| `migrations_test.go` | ERD extension §4 item 10 |

**What no unit test can see, and who must:** the six SQL statements — their column lists, the `unnest` casts, the lock, the number's upsert, the `CHECK`s. A syntax error in a statement passes CI and fails every order on the first request. Dev runs the stack (`docker compose --profile local up --build --wait`) and places one order of each delivery kind, one refused order and one order for the last bike before opening the backend PR; TL's *Run local* step repeats it. The criteria that need Postgres — SC-4, SC-5, SC-11, SC-21 (database part), SC-28 to SC-36 — are proved in the verification step with the SQL of ERD extension §6.

## 7. How the frontend speaks the contract (`src/lib/orders.ts`, `OrderForm.tsx`, PR #17)

PR #17 was read against §3.10, the way a mock would be. Nothing departs from the contract and **no frontend change is asked for**.

| What the frontend does | The contract | Match |
|---|---|---|
| `POST <VITE_API_URL ?? "/api">/v1/orders`, `Content-Type: application/json`, base read at call time | `/v1/orders`; the edge strips `/api` (§2.1) | yes |
| Body keys `name`, `phone`, `delivery`, `address`, `payment`, `note`, `website`, `lines[{id, variant?, qty}]`; no money; `address` is `""` for a pickup; a line with no size has no `variant` key | every key defined; `address` may be empty for a pickup; `variant` optional | yes — no undefined key, so the shape pass of §3 never fires for a real form |
| `201` with a body that has `number`, `placed`, `lines`, `total` is an order; saved as received | the eleven keys; `lines[].image` and `.variant` optional, as `CartLine` types them (`image?`, `variant?`) | yes |
| `400` and `409`: the body's `error` sentence in the alert line | the envelope's `error` key; every sentence of §3 and §4 is one a customer can read | yes |
| `202`, `413`, `5xx`, a network failure, an unreadable body, a 400 or 409 without a sentence: the generic sentence | `202 {}` is a bot, not an order; `413`/`500`/`503` carry sentences the form must not show | yes |
| Pressing again after a refusal sends a new request; a pressed button is disabled | no idempotency key, no state kept by the server (`services.md` §2.6) | yes |

Observations, **none a defect and none asked of this story** (its file list is closed, SC-46):

1. **The bag's quantity stepper has no upper limit** (`Bag.tsx`, `CartDrawer.tsx` pass no `max`), and the bike page's stepper stops at the bike's stock, which the owner can raise above 10. A line of 11 or more reaches the server and comes back as `400` "The quantity of line N must be a whole number from 1 to 10." in the alert line; the customer lowers it in the bag and presses again. Capping both steppers at `min(stock, 10)` is an additive frontend change for a later story.
2. **`placed` is shown in the visitor's own time zone** (`formatDateTime` uses the browser's zone; the offset in `placed` makes the instant right). For a visitor in Vietnam that is the shop's time. No change.
3. **`?demo=1` with an empty bag sends the example lines as a real order** (story §3, "that is the design preview and not a customer path"). The server cannot tell; it holds those two bikes like any order.
4. **`placeOrder` has no timeout** (a nit of the UI review). Every request that reaches the server is answered within 10 s (§5, deadline), so a press that is never answered is the edge or the network hanging, outside this contract. No change.
5. **The frontend branches on the status** (400 and 409 show the sentence; anything else shows the generic one) where `services.md` §2.3 says consumers branch on `code`. The contract makes the sentences of these two statuses fixed and readable, which is the stated exception, so this is consistent; if a third status ever carries a customer sentence, the frontend must learn it.

## 8. Conflicts with merged decisions, resolved

| # | Merged decision | What this story needs | Options and cost | Choice |
|---|---|---|---|---|
| 1 | `services.md` §3 intro: `GET /v1/orders` and `GET /v1/orders/<number>` are `404` "unknown collection"; the route-shape paragraph: a method not offered on a path that exists is `405` with `Allow` | One answer for `GET /v1/orders`, now that this story registers the literal route (the data story's §2 says `/v1/orders` is `404` "until the stories that own those routes register their literal patterns") | `404` for `GET` and `405` for the rest: a path that answers two ways by method, and `Allow` that does not tell the truth about `GET`. `405 Allow: POST` for every other method: one rule, and it is what the header says | **`405` with `Allow: POST` for every method but `POST` on `/v1/orders`**; everything under `/v1/orders/…` stays the data story's `404`. SC-6 accepts either, so no test changes. The contract's `GET /v1/orders` → `404` sentence is narrowed to `/v1/orders/<anything>` |
| 2 | `services.md` §2.5 rule 1: JSON decoding uses `DisallowUnknownFields`; §3.10 step 2 (the bot) comes before step 3 (shape) | The bot answer must come before any shape error, and the advisory keys must pass unread | A struct with `DisallowUnknownFields`: it refuses an unknown key before the bot check can run, and decodes the advisory keys' wrong types | A hand-written pass over `map[string]json.RawMessage` (§3): the same rule, applied in the contract's order. The contract's carve-out for the advisory keys is what it already says |
| 3 | `services.md` §3.10 and `erd.md` §7: the stock is lowered before the order row is written | SC-31 is worded "after the order row is written and before the stock is lowered" | See ERD extension C5: move the hold, or read the sentence as "a failure between two writes of the one transaction" | The contract's order stands; the failure is injected after the last write (ERD extension §6) and a fake fails every step (§6) |
| 4 | `erd.md` §3.9 and §3.11 read the time from `now()` and `AT TIME ZONE 'Asia/Ho_Chi_Minh'` | A clock a test can set (story §6) | See ERD extension C1, C2 | The application passes the instant and the Vietnam date; the column default stays |
| 5 | The data story: "the `DB` interface does **not** grow" | A transaction, which `QueryRow` cannot give | Grow `DB` by `Begin`: the pool has it, the fake gains one line. Type-assert in the route line: a fake `DB` silently gets no route and a test of the route passes by accident. A second argument to `Router`: changes `main.go` and every call of `Router` | `DB` gains `Begin` (§2). The data story's sentence was about its own reads and stays true of them |

## 9. Rollout, dependencies, open questions

- **Depends on** the data story's backend for: the `products` and `shop` tables and their seed (`000100`, `000200`); `httpx.WriteCoded`, `RequestIDFrom` and the middleware chain `RequestID(Recover(Log(NoStore(mux))))`; the `/v1/{collection}/{id}` route that answers `404` under `/v1/orders/…`. This backend starts from a `main` that has them. It is depended on by *Shop admin* only for the stock hold's `version` bump, which the trigger of migration 1 already makes (ADMIN-008); nothing here waits on admin.
- **Order of delivery.** The frontend PR (#17) merges first, then this backend PR, as the delivery flow says. Between the two merges `main` has a checkout that shows its generic failure sentence on every press and no route; nothing is deployed between them.
- **Secrets and setup.** None. No payment gateway, e-mail or SMS service is called by `api`; the owner's message is sent by the browser through the existing key in `src/site.json` (§5 of `services.md`). No `request_secret` and no `request_setup`.
- **Open questions.** None new and none blocking. The contract's question 2 (no limit on `POST /v1/orders`: anyone can hold every bike) and the ERD's question 1 (retention) keep their proposed defaults; the ERD extension adds one non-blocking note about the seeded stock figures being invented.
