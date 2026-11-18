# Service contract extension — Shop admin: real sign-in and saved edits

Story: `docs/admin/stories/shop-admin-real-sign-in-and-saved-edits.md` · Requirements: ADMIN-001 to ADMIN-013 (`docs/admin/SRS.md`) · Plan item: *Shop admin: real sign-in and saved edits* (feature, P1) · Migration slot: `20261118000400_sessions` · Companion: `docs/architecture/erd/shop-admin-real-sign-in-and-saved-edits.md`

## 1. Result

**This story adds no endpoint, no request or response field, no error code and no pagination or auth rule that the contract lacks.** It builds the six endpoints `docs/architecture/services.md` already specifies — §3.4 `POST /v1/{collection}`, §3.5 `PUT /v1/{collection}/{id}`, §3.6 `DELETE /v1/{collection}/{id}`, §3.7 `POST /v1/sessions`, §3.8 `GET /v1/me`, §3.9 `DELETE /v1/sessions/current` — with their bodies, answers, error tables and **fixed** sentences, under the account and session rules of §2.2 and the body rules of §3.11. Nothing of those sections, of §2.3 (the envelope), §2.5 (the validation boundary) or §2.6 (idempotency) is restated here. The one read it also relies on, `GET /v1/{collection}/{id}` with its `ETag` (§3.3), is the data story's.

There is no mock module: PR #20 deleted nothing and added none — `src/lib/session.ts` and `src/components/admin/store.ts` are the API clients. §7 reads them against the contract; **they depart from nothing and ask for no frontend change.**

What this file adds, because the contract leaves it to the builder:

| § | What |
|---|---|
| 2 | The routes, and what every other request to the same paths gets |
| 3 | The account and the session: start-up, the password check, the token and the cookie, the throttle, `requireOwner`, the order of the checks of each endpoint, every sentence the contract does not spell |
| 4 | The writes: how a body is read (order of checks, the rule of every field with its sentence and `field`), then the add, the replace and the delete step by step |
| 5 | Failures, headers, logging |
| 6 | Package layout and what is tested without a database |
| 7 | How the frontend speaks the contract (PR #20 read against it) |
| 8 | Conflicts with merged decisions, resolved |
| 9 | Rollout, dependencies, open questions |

## 2. Routes

In `internal/server.Router`, three new lines for the account, and the two data-story lines widened (the only edit to that story's code; the read paths are untouched):

```go
mux.HandleFunc("/v1/sessions", owner.SignIn)                  // method-less: the handler answers 405 to the rest
mux.HandleFunc("DELETE /v1/sessions/current", owner.SignOut)  // method-specific, see below
mux.HandleFunc("/v1/me", owner.Me)                            // method-less
mux.HandleFunc("/v1/{collection}", collections.Collection(db, owner.RequireOwner))
mux.HandleFunc("/v1/{collection}/{id}", collections.Record(db, owner.RequireOwner))
```

`Router(db DB)` becomes `Router(db DB, account *auth.Account)`: it builds the service from the account, a store over `db` and `time.Now`, and registers the lines above. `cmd/api/main.go` calls `server.Router(pool, auth.NewAccount(cfg.AdminEmail, cfg.AdminPassword))` and keeps no copy of either secret; every existing test call site passes `auth.NewAccount("", "")` (an account that is not configured). `server.DB` **does not grow**: every statement of the story is answered by one row, which `QueryRow` provides (`docs/architecture/erd/shop-admin-real-sign-in-and-saved-edits.md` §2).

| Request | Answer |
|---|---|
| `POST /v1/sessions` | §3.7 of the contract; this file §3.5 |
| `GET`, `HEAD`, `OPTIONS`, `PUT`, `PATCH`, `DELETE` on `/v1/sessions` | `405 METHOD_NOT_ALLOWED`, `Allow: POST`, "That method is not offered here." (SC-A1) |
| `DELETE /v1/sessions/current` | §3.9; this file §3.7 |
| **Any other method** on `/v1/sessions/current` (`GET`, `PUT`, `POST`, …) | **`404 NOT_FOUND`** "That collection does not exist." — not 405. The pattern carries the method, so the request falls through to `/v1/{collection}/{id}` with the unknown collection `sessions`, which the data story answers `404` for every method before it looks at the method. This is the contract's "a session is never readable" (§3 intro) and SC-A1; §3's general "wrong method on a path that exists is 405" does not apply to this one path |
| `GET /v1/me` | §3.8; this file §3.6 |
| `POST`, `HEAD`, `OPTIONS`, `PUT`, `PATCH`, `DELETE` on `/v1/me` | `405`, `Allow: GET` |
| `/v1/me/`, `/v1/sessions/`, `/v1/ME`, `/v1/Sessions` | `404` — the trailing slash by the catch-all ("That address does not exist."), the capitalised name as an unknown collection ("That collection does not exist.") |
| `POST /v1/{collection}` | §4; `Allow` of a wrong method: `GET, POST` on a writable collection, `GET` on the six read-only ones |
| `PUT`/`DELETE /v1/{collection}/{id}` | §4; `Allow` of a wrong method: `GET, PUT, DELETE` on a writable collection, `GET` on the six read-only ones |
| Any write method on an **unknown** collection | `404` "That collection does not exist.", before the method and before the session — the order of the contract's §3 intro |

The order for a collection route, which the dispatcher in `collections` pins (the contract's common order up to the session): **unknown collection → 404; a method the route shape does not offer at all (`HEAD`, `OPTIONS`, `PATCH`, or `PUT` on the collection path) → 405 with that collection's `Allow`; a write method on one of the six read-only collections → 405 `Allow: GET`, whether or not anyone is signed in; then `requireOwner` → 401; then the write.** A `GET` goes to the data story's read handler unchanged.

## 3. The account and the session

Everything here lives in `internal/auth`. Nothing in it imports `collections` or `server` (the package declares its own one-method `DB` interface with the shape of `server.DB`).

### 3.1 Start-up

`config.Config` gains `AdminEmail` and `AdminPassword`, read through the same `getenv` as the rest. `AdminEmail` is trimmed. `AdminPassword` is **kept exactly as read except that trailing carriage returns and line feeds are removed** (no form field can carry one, and a secret pasted into a console with a trailing newline would otherwise be a password nobody can type); leading and trailing spaces stay, because the password is "compared exactly" (§2.2). A secret that is empty after `strings.TrimSpace` is **not set**. Neither is required: `config.Load` still fails only on a missing `DATABASE_URL` or a bad port.

`auth.NewAccount(email, password)` runs **once, before the server listens**, and is the only place either secret is used:

| Step | Detail |
|---|---|
| Not configured | Either secret not set → an account with no email and no key. It still computes **one** argon2id over a random value with a random salt, as the `dummy`, so a refused attempt costs what a real one costs |
| Email | `strings.ToLower(strings.TrimSpace(email))`. What is compared and what is answered (`{"email": …}`) |
| Hash | `argon2id`, `m=19456 KiB, t=2, p=1`, a 16-byte salt from `crypto/rand`, a 32-byte output (`golang.org/x/crypto/argon2`, `IDKey(password, salt, 2, 19456, 1, 32)`). The hash and the salt are the account's; **the password is in no field of any struct, global or log** — `main` passes it to `NewAccount` and keeps nothing |
| Fingerprint | `SHA-256("deo-admin-v1" 0x00 email 0x00 password)` from the normalized email and the password as read; held as the HMAC key of §3.2 |

**The check** (`Account.Verify(email, password) bool`) always does the same work, so the time does not say which part was wrong: it computes `IDKey(typed password, salt, …)` with the account's salt (the dummy's when the account is not configured), compares the result with the stored hash by `subtle.ConstantTimeCompare`, compares **`SHA-256(normalized typed email)` with `SHA-256(normalized email)`** the same way (digests, so the length of the real email is not leaked by an early return), and answers true only if both match **and** the account is configured. No early return before the argon2id. **At most 4 computations run at once** (a buffered channel of 4; a request waits for a slot until its own deadline, then answers `500 INTERNAL`): each takes 19 MiB, and sign-in is the one endpoint anyone may call before the throttle has seen a failure.

### 3.2 The token and the cookie

- **Token** — 32 bytes from `crypto/rand`, base64url **without padding**, 43 characters (`^[A-Za-z0-9_-]{43}$`). It leaves the server only in the cookie.
- **Stored hash** — `HMAC-SHA256(key = fingerprint, message = the 43 characters as text)`, 32 bytes (erd extension C1). A cookie value that does not match `^[A-Za-z0-9_-]{43}$`, or any cookie when the account is not configured, is a `401` **without a query**.
- **Cookie** — built with `http.SetCookie` from `http.Cookie{Name: "deo_session", Value: token, Path: "/", MaxAge: 604800, Expires: now+7 days, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: https}`. The clearing cookie has the same name, `Path: "/"`, `HttpOnly`, `SameSite=Lax`, the same `Secure`, an empty value and `MaxAge: -1` (Go writes `Max-Age=0`). No `Domain`.
- **`https`** is true when the **last** entry of `X-Forwarded-Proto` (the one the edge appended), trimmed and compared ignoring capitals, is `https`, or when `r.TLS != nil`. Anything else is plain HTTP and the cookie has no `Secure` (SC-5).
- **Client address** (throttle key) — the last entry of `X-Forwarded-For`, trimmed; if the header is empty, the host part of `r.RemoteAddr` (`net.SplitHostPort`, the whole text if it has no port). The address is never logged.

### 3.3 The throttle

An in-memory table behind one mutex, keyed by client address; the clock is a `func() time.Time` injected at construction. An entry is `{failures []time.Time, last time.Time, lockedUntil time.Time}`.

| Operation | Rule |
|---|---|
| `Check(addr)` | Not locked while `lockedUntil` is zero or not after now. Locked when `lockedUntil` is after now: the answer carries `Retry-After` = the seconds left, **rounded up**, clamped to 1…900. When `lockedUntil` has passed, the entry is deleted (the lock is over: 15 minutes after the latest failure) |
| `Fail(addr)` | Drops failures 15 minutes old or older, appends now, sets `last = now`; **on the 5th failure within the window sets `lockedUntil = now + 15 min`**. A request that is answered `429` never reaches `Fail`, so a locked address cannot extend its own lock |
| `Reset(addr)` | Deletes the entry (a success clears the count, SC-9) |
| Housekeeping | On every operation entries whose `last` is 15 minutes old or older **and** that are not locked are deleted; above 10,000 entries the one with the oldest `last` is evicted |

What counts as a failure is a **`401` from `POST /v1/sessions`**: wrong email, wrong password, account not configured. A `400`, a `413`, a `429` and a database failure after correct credentials are **not** failures. A failure that has aged out one by one (four failures spread over more than 15 minutes) never locks; five within 15 minutes lock until 15 minutes after the fifth (SC-8). The table is lost on restart (`erd.md` §1).

### 3.4 `requireOwner`

A function `func(http.Handler) http.Handler` that the collection dispatcher applies to a write **after** the collection and writability checks (§2) and **before** it reads the body. It reads the first `deo_session` cookie, applies the shape rule of §3.2, computes the HMAC, runs statement A2 with the current normalized email and the clock's now, and:

| Outcome | Answer |
|---|---|
| a row | the request goes on; the owner's email is not put in the context (nothing downstream needs it) |
| no cookie, a malformed cookie, no row, the account not configured | `401 UNAUTHENTICATED`, "You are not signed in." — the same answer for expired, tampered, signed-out, issued before a secret changed |
| the statement failed | `503 UNAVAILABLE` or `500 INTERNAL` by the predicate of §5. **Never `401`**: a database that is down is not "nobody is signed in" |

It writes nothing and does not extend the session.

### 3.5 `POST /v1/sessions` — the checks, in this order

| # | Check | Answer |
|---|---|---|
| 1 | Method is `POST` | else `405`, `Allow: POST` |
| 2 | The client address is not locked (§3.3) | `429 RATE_LIMITED`, **fixed** "Too many attempts. Try again in a few minutes.", `Retry-After`. Before every other check: a locked address gets `429` whatever it sends |
| 3 | Size: the body is read completely through `http.MaxBytesReader(w, r.Body, 8192)` before anything is parsed | `413 PAYLOAD_TOO_LARGE`, "The request is too large." |
| 4 | `Content-Type` parses (`mime.ParseMediaType`) as `application/json`, parameters allowed | else `400 MALFORMED_REQUEST`, "The request must be sent as JSON." |
| 5 | The first byte that is not white space is `{`, and only white space follows the one JSON value | else `400 MALFORMED_REQUEST`, "The request must be a JSON object." (an array, a string, a number, `null`, an empty body, bad JSON and trailing data are all this one) |
| 6 | An undefined key, in alphabetical order | `400 MALFORMED_REQUEST`, `field` = the key, "The field “remember” is not part of a sign-in." |
| 7 | `email`, then `password`: absent or `null` → "The field “password” is needed."; not a JSON string → "The field “password” must be text." | `400 MALFORMED_REQUEST`, `field` = the key. An **empty string is a value** and simply does not match (a `401`) |
| 8 | `email` longer than 320 characters (code points), then `password` longer than 1,024 bytes | `400 VALIDATION_FAILED`, `field` = the key, "The email is too long." / "The password is too long." |
| 9 | `Account.Verify` | false → `Fail(addr)`, log, `401 UNAUTHENTICATED`, **fixed** "That email and password do not match." |
| 10 | A random token, its HMAC, statement A1 (the purge and the insert) | a failure → `503`/`500` by §5, **no cookie, no `Reset`, no failure counted** |
| 11 | `Reset(addr)`, then the cookie and `200 {"email": "<normalized>"}` | the cookie is set **only here**, after the row is stored |

The password and the typed email are never logged; the log line is `msg: "sign-in"` with `request_id` and `outcome` (`ok`, `bad_credentials`, `not_configured`, `locked`). When the account is not configured, **one** additional `Warn` line per attempt — `msg: "admin account is not configured"`, `request_id` — is what ADMIN-001 AC-4 reads (SC-3).

### 3.6 `GET /v1/me`

Method check (`405`, `Allow: GET`), then the work of `requireOwner` (§3.4) answering `200 {"email": "<normalized ADMIN_EMAIL>"}` on a row and the same `401 UNAUTHENTICATED` "You are not signed in." (or `503`/`500`) otherwise. One indexed read, no write.

### 3.7 `DELETE /v1/sessions/current`

Reads the first `deo_session` cookie. When it has the shape of §3.2 and the account is configured, runs statement A3 (`DELETE … WHERE token_hash = $1`): **a row or no row are both success**. On success — and when there is no cookie, or a cookie of the wrong shape — it answers `204` with the **clearing** cookie, always. When A3 itself failed it answers `503`/`500` and sends **no** `Set-Cookie`, so the owner is never told they are signed out while the session still works (SC-A3). It does not end the owner's other sessions.

## 4. The writes

Everything here lives in `internal/collections`, beside the data story's reads, and uses **the same field table** (`schema.go`): the order of the fields, their kinds, the closed sets (`options`) and the answer columns all come from it, so the validator, the statements and the reads cannot disagree. Two things the table must carry for a write and the reads did not need: the **closed sets** of `brand`, `status` and `wheel` (the schema's `options`), and a **"whole number" mark** on `price`, `was` and `stock` (`schema.json` types `stock` as a plain `number`, as it does `year`; erd extension §5). If the data story's table lacks either, this story adds it and extends that story's test that pins the table to `schema.json`.

### 4.1 Reading a body — one function, `Decode(collection, body, address)`, no database

It runs the contract's order and returns a typed record or the first refusal. The nouns of the sentences are "bike" (`products`) and "riding type" (`categories`).

| Step | Check | Answer |
|---|---|---|
| 1 | Size: the body is read completely through `http.MaxBytesReader(w, r.Body, 1<<20)` **before** anything is parsed | `413 PAYLOAD_TOO_LARGE`, "The request is too large." (a `413` comes before every `400`) |
| 2 | `Content-Type` is `application/json` (parameters allowed, capitals ignored) | `400 MALFORMED_REQUEST`, "The request must be sent as JSON." |
| 3 | The first non-white-space byte is `{` and only white space follows the one value | `400 MALFORMED_REQUEST`, "The request must be a JSON object." |
| 4 | **Undefined keys**, in alphabetical order. `seq`, `version`, `created_at`, `updated_at` and **`_source`** are undefined | `400 MALFORMED_REQUEST`, `field` = the key, "The field “colourful” is not part of a bike." / "…of a riding type." |
| 5 | **Each field in field-table order, `id` first: its JSON type, then its value** (the table below). The first failure answers, so a refused `brand` is reported before a malformed `price` because `brand` comes first | as the table |
| 6 | **Cross-field:** when the saved `status` is `live`, `price` must be present and above 0 | `400 VALIDATION_FAILED`, `field` = `price`, **fixed** "A live bike needs a price." (a `draft`, or no `status`, needs none; a price of `""` or absent or 0 on a live bike is this one) |

The body is decoded into `map[string]json.RawMessage` (a duplicate key keeps the last value), so a number is always the **exact text of its token**. `DisallowUnknownFields` is not used: it cannot name the key, and the type and value of each field are checked together in the table's order (§8, row 6). The empties of the contract (§3.11) apply first to every field: an **absent** key and `null` are "no value"; for a text, a number or a price, so is the string `""`; a list takes only `null`; `[]` is a real empty list; a boolean takes only `true`, `false`, `null`. "Characters" are code points. A field with no value is bound as `NULL` and the key leaves the record (erd extension C5); `name` has no empty (it is required).

| Field kind | Wrong JSON type → `MALFORMED_REQUEST` | Value refused → `VALIDATION_FAILED` (`field` = the key) |
|---|---|---|
| `id` | not a string: "The field “id” must be text." | `""` or not `^[a-z0-9][a-z0-9-]{0,63}$`: "The field “id” must be 1 to 64 lowercase letters, digits or hyphens, and not start with a hyphen."; on a **change**, one that differs from the address's id: "The field “id” must equal the address's id." |
| `name` (required) | not a string: "The field “name” must be text." | absent, `null`, `""` or only white space (`strings.TrimSpace`; stored **as sent**, not trimmed): "The field “name” is needed."; over 5,000 characters: "The field “name” must be at most 5000 characters." |
| other text (`image`, `price_listed`, `fact`, `travel`, `frame`, `fork`, `shock`, `motor`, `drivetrain`, `brakes`, `wheels`, `tyres`, `weight`, `colour`, `description`, `ride`, `image_back`, `image_detail`; `line`, `terrain`) | "The field “description” must be text." | over 5,000 characters (5,000 accepted, 5,001 refused): "The field “description” must be at most 5000 characters." |
| `brand`, `status`, `wheel` | "must be text" — **a JSON number `29` for `wheel` is this one**; the contract's value is the string `"29"` | not in the closed set: "The field “status” must be one of: draft, live." (`brand`: Giant, Liv; `wheel`: 29, mixed) |
| `category` | "must be text" | none at this step; its existence is checked after every body rule (§4.2, §4.3) |
| `year`, `rating`, `review_count` | not a number: "The field “year” must be a number." | a token over 40 characters, or one `strconv.ParseFloat` refuses (out of range), is "The field “year” must be a number." |
| `price`, `was` | "The field “price” must be a number." (`"cheap"` is this one) | the token is not `^-?(0|[1-9][0-9]*)$` (`1.5`, `5.0`, `1e3`), or is negative, or above 9,007,199,254,740,991 (`ParseInt`): "The field “price” must be a whole number of đồng, not below 0." (`-0` is `0`, accepted) |
| `stock` | "The field “stock” must be a number." | the same rule: "The field “stock” must be a whole number, not below 0." |
| `deal`, `month` | not a boolean: "The field “deal” must be true or false." (`"yes"` is this one) | — |
| lists: `sizes`, `sizes_out`, `size_fit`, `_invented` | not an array, or an entry that is not a string: "The field “sizes” must be a list of text." | over 50 entries: "The field “sizes” must have at most 50 entries."; an entry over 5,000 characters: "The field “sizes” must have entries of at most 5000 characters." |
| every text and every list entry | — | contains U+0000: "The text in “description” contains a character that cannot be stored." — Postgres cannot store it, and it would otherwise reach the statement and answer a 500 |

In a text field the checks run blank (`name` only), then the NUL check, then the length. `price_listed` and `_invented` are hidden from the owner's form but are schema fields, so they are accepted when sent (a replace must be able to send back what it read).

### 4.2 `POST /v1/{collection}` — after the checks of §2 (collection, writability, session)

1. `Decode` (§4.1), `id` optional: absent or `null` → the server makes one.
2. When the body names a `category`: **W5**. `false` → `400 VALIDATION_FAILED`, `field` = `category`, "The field “category” must be the id of a riding type that exists." (this comes **after every body rule** and before the id is tried: SC-B3).
3. **W1.** A body `id` is tried once; no row → `409 ID_TAKEN`, **fixed** "That id is already used.". A server-made id (`<collection>-` and 6 characters of `a–z0–9` from `crypto/rand`, by rejection sampling so no character is likelier than another) is tried up to 5 times with a new id each time, then `500 INTERNAL`. SQLSTATE `23503` from W1 (the riding type was deleted in the instant after W5) is the same 400 as step 2.
4. `201` with `ETag: "<version>"` (`"1"`) and the row W1 returned as the body, written as raw JSON (never decoded into floats).

### 4.3 `PUT /v1/{collection}/{id}` — after the checks of §2

1. `Decode` with the address's `id` (§4.1).
2. **W2.** `version` is `NULL` → `404 NOT_FOUND`, "That record does not exist." — **before** the `If-Match` header is read (SC-B4).
3. `If-Match`: no header line → `428 PRECONDITION_REQUIRED`, "The save did not say which version of the record it was based on. Reload and make your change again."; more than one header line, or one that is not exactly `^"[1-9][0-9]*"$` after trimming (unquoted `7`, `W/"7"`, `*`, `"abc"`, `"7", "8"`), or a number above `int64` → `400 MALFORMED_REQUEST`, "The If-Match header must be the version this record was read at, for example \"7\"." (no `field`: it is a header, not a body field).
4. `category_ok` of W2 is `false` → the category 400 of §4.2 (**before** staleness: SC-B4).
5. **W3** with the version of the header. A row → `200`, `ETag: "<version>"` (the new one), the row as the body. No row → **W4**: no row → `404`; a row → `409 STALE_RECORD`, **fixed** "This record changed while you were editing it — for example an order was placed. Reload and make your change again." SQLSTATE `23503` from W3 is the category 400.

A save is refused if **anything** — the owner's other window, an order's stock hold — has bumped the version since the caller's read; a replay of an accepted save is therefore `409` (§2.6 of the contract). The token the client keeps for its next save is the `ETag` of this answer.

### 4.4 `DELETE /v1/{collection}/{id}` — after the checks of §2

No body is read, no header is read (`If-Match` is ignored: a delete is not refused for staleness, SC-35).

- `products`: **W6**. No row → `404`, "That record does not exist."; a row → `204`.
- `categories`: **W6**. No row → `404`; a row → `204`; SQLSTATE `23503` → **W7**: no row → `404` (it was deleted in the instant between); `N` bikes with `N ≥ 1` → `409 IN_USE`, **fixed** "<Name> is still the riding type of <N> bike. Give those bikes another riding type first." (`bike`, not `bikes`, when `N` is 1; `<Name>` is the type's `name` as stored); `N = 0` (the last bike moved away in the instant between) → W6 again, at most 3 times in all, then `500 INTERNAL`.

A bike's reviews and a riding type's gallery photographs are changed by `ON DELETE SET NULL` inside W6; no code of this story touches them (SC-28). The `204` has no body and no `Content-Type`.

## 5. Failures, headers, logging

| The error is | Answer |
|---|---|
| A refused connection (`syscall.ECONNREFUSED` anywhere in the chain, or a dial `*net.OpError`), or a `*pgconn.PgError` of class `08` or `57P01`, `57P02`, `57P03` | `503 UNAVAILABLE`, "The service is not ready." |
| SQLSTATE `23503` where §4 says so | the 400 (a save, an add) or the 409 (a riding type's delete) of §4 |
| Anything else: a failed statement (a renamed table, SC-B5), a `CHECK` or unique violation (`23502`, `23505`, `23514`), a deadline, a scan error, a cancelled request | `500 INTERNAL`, "Something went wrong. Please try again." |

The two classes use the **one** predicate of the data story (`unavailable(err)`, `docs/architecture/services/site-data-from-the-database.md` §5.1). `internal/auth` needs it too and must not import `collections`, so it lives in `internal/httpx`: if it is still private to `internal/collections` when this story starts, move it with the same body (the orders story's design says the same); whichever of the two stories merges first moves it and the other calls it — never a second copy. This story adds the `httpx` code constants it needs beside the data story's: `CodeUnauthenticated`, `CodeRateLimited`, `CodeIDTaken`, `CodeStaleRecord`, `CodeInUse`, `CodePreconditionRequired`, and, if the orders story has not, `CodeMalformedRequest`, `CodeValidationFailed`, `CodePayloadTooLarge`. Every refusal is written by `httpx.WriteCoded`.

- **Deadlines** (contract §5, PostgreSQL row): `context.WithTimeout(r.Context(), 10*time.Second)` around the statements of one write and around sign-in (it covers the wait for an argon2id slot); 5 s around the `me` check and `requireOwner`. A deadline is a `500`. The statements of a request run one after the other on the request's context; nothing is retried by the server.
- **Headers.** `ETag: "<version>"` on `201` and `200` of a write only. `Retry-After` on `429`. `Allow` on every `405`. `Set-Cookie` on a successful sign-in (`200`) and on a sign-out that succeeded (`204`) and **nowhere else** — in particular not on a failed sign-in, a failed A1 or a failed A3. No `Location`, no CORS header (contract §2.8). `Cache-Control: no-store` and `X-Request-Id` come from the middleware chain on every answer.
- **Logging.** One `Info` line per accepted write from the handler: `msg: "record saved"` or `"record deleted"`, `request_id`, `collection`, `id`, `op` (`add`, `replace`, `delete`). One `Info` line per refusal: `msg: "write refused"`, `request_id`, `collection`, `code`. Sign-in, as §3.5. A failed statement is logged at `Error` with the real cause (SQLSTATE and message), `collection` or the operation, and `request_id`; a request the caller abandoned (`context.Canceled`) at `Info`. **Never logged:** the body, any field of a record, the cookie, the token, the HMAC, the typed email, the password, `ADMIN_EMAIL`, `ADMIN_PASSWORD`, the client address, the query string.
- **No new rate limit.** Only sign-in is throttled (§3.3); a write needs a session.

## 6. Layout and what is tested without a database

CI has no database, so everything that can be decided before SQL is a pure function with a unit test, the statements sit behind small interfaces a fake can fail, and the statements themselves are proved on the running stack.

| File | Holds |
|---|---|
| `internal/auth/account.go` | `Account`, `NewAccount`, `Verify` (§3.1) |
| `internal/auth/throttle.go` | the throttle with its injected clock (§3.3) |
| `internal/auth/cookie.go` | token, HMAC, set and clear cookie, `https`, client address (§3.2) |
| `internal/auth/store.go` | `Sessions` (an interface: `Create`, `Exists`, `Delete`) and its pgx implementation of A1 to A3 |
| `internal/auth/handler.go`, `require.go` | `SignIn`, `Me`, `SignOut`, `RequireOwner` (§3.4 to §3.7) |
| `internal/collections/validate.go` | `Decode` (§4.1): pure |
| `internal/collections/write.go` | the fixed statements W1 to W7 built once per collection from the field table, and `Writer` (an interface: `Add`, `Lookup`, `Replace`, `Remove`) with the pgx implementation |
| `internal/collections/write_handler.go`, `handler.go` | the three write handlers (§4.2 to §4.4) and the method dispatch of §2 (the one edit to the data story's handlers) |
| `internal/config/config.go`, `config_test.go` | the two fields and the trailing-newline rule of §3.1 |
| `internal/httpx/httpx.go` | the constants and the predicate of §5 |
| `internal/server/server.go`, `server_test.go` | the five route lines of §2; `Router` takes the account |
| `cmd/api/main.go` | builds the account before `ListenAndServe` |
| `migrations/20261118000400_sessions.{up,down}.sql`, `migrations_test.go` | erd extension §4 |
| `go.mod`, `go.sum` | `golang.org/x/crypto` (argon2), added with `run_go_mod_tidy` — the one new dependency |
| `.env.example` | **unchanged**: `ADMIN_EMAIL` and `ADMIN_PASSWORD` are already there, as names with no value |

The tests (new files; existing tests change only where the table says):

| Test | Proves (story criteria) |
|---|---|
| `Account` / `SignIn` against a fake `Sessions`: the right pair; the email typed `  OWNER@deo.test `; a wrong password and a wrong email give **byte-identical** 401 bodies; not configured (both unset, one unset, white space only) is that same 401, and the captured `slog` output holds the `admin account is not configured` line once per attempt and never the password or the typed email; the password's value is in no answer, no log line and no struct field | SC-1, SC-2, SC-3, SC-6 |
| Cookie attributes of a success: `HttpOnly`, `SameSite=Lax`, `Path=/`, `Max-Age=604800` and an `Expires` 7 days on from the injected clock, no `Domain`, the body holds the email only; `Secure` with `X-Forwarded-Proto: https` (also as the last of two values, in capitals) and with `r.TLS`, absent otherwise; the clearing cookie has `Max-Age=0` and the same other attributes | SC-4, SC-5, SC-12 |
| Sign-in shapes: no password key, `password: 123`, `remember: true`, `[]`, `null`, `{}`, empty body, trailing data, `text/plain`, a 9 KiB body (413), 321-character email and 1,025-byte password (400 `VALIDATION_FAILED`) — each answers its row of §3.5 with its `field`, **none counts as a failure** | SC-7 |
| Throttle with the injected clock: five failures then the right password is 429 with `Retry-After` between 1 and 900 and **no** `Set-Cookie`; 14 min 59 s later still 429; 15 min later the right password is accepted; a 429 does not extend the lock; four failures, a success, four failures, a success; four failures spread over 20 minutes never lock; two addresses are independent; the address is the last `X-Forwarded-For` entry; 10,001 addresses keep the table at 10,000 | SC-8, SC-9 |
| Session lookup with a fake `Sessions` that is a map keyed by hash: a cookie with each character changed (the last one too), a cookie signed in under another password or email (a different fingerprint, so a different hash), an expired row (the fake compares with the injected clock), a signed-out cookie, no cookie, a cookie of the wrong length (no store call at all), an account not configured (no store call at all) — each `GET /v1/me` is `401` "You are not signed in."; the right cookie is `200 {"email": …}`; sign-out is `204` with the clearing cookie, with and without a cookie, and the replayed cookie is then `401`; a store that fails answers `500`, one that is refused `503`, and sign-out then sends no `Set-Cookie` | SC-10, SC-11, SC-12, SC-A3, SC-A4 |
| Dispatcher, table-driven over the 8 collections × `POST`/`PUT`/`DELETE` × {no session, session} (the owner function is a stub): the six read-only collections are `405` `Allow: GET` with or without a session; `products`, `categories` are `401` without one; an unknown collection is `404` for every method; `PATCH`, `HEAD`, `OPTIONS` get the `Allow` of §2; a bad body, `text/plain` or an unknown id **without a session** is `401` (the session is checked before the body, the `If-Match` and the record); `GET` still reaches the read handler | SC-13, SC-14, SC-15 |
| `Decode`, table-driven: every row of §4.1 with its status, `code`, `field` and key words; **two rules broken at once** answer the one the order says (`{"id":"Bad Id","name":""}` → `id`; `brand: "Trek"` with `price: -1` → `brand`; `status: "live"` with `stock: -1` → `stock`; `category: "unicycle"` with `price: -1` → `price`); `price` `"cheap"` / `-1` / `1.5` / `5.0` / `1e3` / `9007199254740991` / `9007199254740992` / `""`; `stock` `-1` / `2.5` / `0`; `sizes` `"M"` / `[]` / 51 entries; `deal` `"yes"`; `wheel` `"29"` / `29` / `"26"`; text of 5,000 and 5,001 characters counted in code points (5,000 × `Đ` accepted); NUL in a text and in a list entry; `""` and `null` for each kind; `name` of spaces and of a no-break space; `id` `"my-bike-1"`, `"Bad Id"`, 64 and 65 characters, on a change a different id; `_source`, `seq`, `version`, `colourful` as undefined keys; a duplicate key; a number token of 40 and 41 characters; a live bike with price absent, `""`, 0 and a draft with none | SC-16 to SC-24, SC-B1 to SC-B3 |
| Statement builders: the placeholder count equals the number of writable columns plus 1 (W1) or 2 (W3); every placeholder has its cast; the `INSERT` and `SET` lists never hold `seq`, `version`, `created_at`, `updated_at` or `_source`; the `RETURNING` list equals the data story's answer-column list plus `version`; no request text is in any statement; a list `null` binds `nil` and `[]` binds a non-nil empty slice | SC-22, SC-17 (a field the body omits is gone), SC-58 |
| Write flow against a fake `Writer` that can fail any step: an add with no id (server-made, 5 collisions then `500`), with a free id, with a taken id (`409` "That id is already used."), with a missing riding type (`400` `category`, before the id); a change: invalid body and no `If-Match` → 400, unknown id and no `If-Match` → 404, no header → 428, the five malformed values of SC-33d → 400, stale and a missing riding type → 400, stale → 409 with the fixed sentence, a zero-row W3 then a gone record → 404, SQLSTATE `23503` → the category 400; a delete: gone → 404, a bike → 204, a riding type in use by 4 and by 1 → the two sentences, the type gone in the instant between → 404, 0 bikes three times → `500`; a `*pgconn.PgError` `42P01` → `500` with no table name in the body, an `ECONNREFUSED` → `503` | SC-25, SC-26, SC-27, SC-33, SC-35, SC-B4, SC-B5 |
| `config`: the two values are read; trailing `\r\n` is removed from the password and nothing else; white-space-only is "not set"; `Load` still fails only for `DATABASE_URL` and the port | SC-3 |
| `migrations_test.go`: the pair of erd extension §4 item 8 | — |

**Proved only on the running stack** (CI has no database): that a save and an order serialise on the row lock (SC-33, SC-34), that `ON CONFLICT`/`23503`/`SET NULL` behave as §4 says (SC-16, SC-27, SC-28, SC-29), that the public pages show a saved edit (SC-30 to SC-32), and the screens (SC-36 to SC-83). SC-34b (a save and an order in the same instant) is manual: the backend review reads the row lock instead.

## 7. How the frontend speaks the contract (PR #20 read against it)

PR #20 (`feature/shop-admin-real-sign-in-and-saved-edits-ui`, approved at UI review) has no mock module: `session.ts` and `store.ts` call `/api/v1/…` (or `<VITE_API_URL>/v1/…`) directly. Read call by call against the contract:

| Client call | Contract | Reading |
|---|---|---|
| The first page that reads the session asks `GET /v1/me` once per page load; `200` → the server's email, anything else → nobody | §3.8, §3.9 of this file | matches. A visitor with no cookie sees a `401` in the network log on a management page; that is the answer, not an error |
| `signIn(email, password)` → `POST /v1/sessions`, `Content-Type: application/json`, body `{"email", "password"}` | §3.7 | matches **only while the body has exactly those two keys**: the server refuses a third (`400`, `field` = the key). The client must never add `remember`, `role` or anything else |
| `signOut()` → `DELETE /v1/sessions/current`, the session forgotten only after `204` | §3.9 | matches: the server never answers `204` while the session still works |
| `readRow` → `GET /v1/{collection}/{id}`, keeps the `ETag` text as received (quotes included) | §3.3 | matches; a `404` carries "That record does not exist." and the client treats a finished `404` as "not here" |
| `saveRow` with no `id` → `POST /v1/{collection}`; body = the record without `_source` and without `id`; `201` + the record + `ETag: "1"` | §3.4, §4.2 | matches; the client keeps the `ETag` for the next save |
| `saveRow` with an `id` → `PUT /v1/{collection}/{id}`, `If-Match` = the kept `ETag`, body = the record without `_source`, **including** `id` equal to the address | §3.5, §4.3 | matches; every hidden field the form does not show (`price_listed`, `_invented`) goes back as read, which the replace needs. A save with no kept token is `428` and shows that sentence |
| `deleteRow` → `DELETE /v1/{collection}/{id}`, no body, no `If-Match` | §3.6, §4.4 | matches |
| Every failure shows `error` of the envelope, else a generic sentence; `code` is not read | §2.3 | matches (the contract tells consumers to branch on `code`; the client branches on the **status** — `401` only — and shows the sentence, which every sentence of this file is fit to show) |
| A `401` on a write is "Your session has ended. Sign in again." and drops the session | §3.4 of this file | matches: only `requireOwner` answers `401` to a write; the sign-in's own `401` is a different call and shows the server's sentence |
| Every call is same-origin through `/api` | §2.1, §9 question 3 of the contract | matches; the `ETag` is readable by the page only on the same origin |

**Departures: none. No frontend change is asked for.** Two facts the screens meet that the mock-less client could not have checked and that are not defects: a blank "Add" sends the form's real values (`sizes: []`, `deal: false`, an empty text as `""`), and the server stores the lists and booleans as given and the empties as no value (erd extension C5); a bike added with no `stock` has none stored, which `erd.md` §3.2 reads as "not tracked: cannot be ordered".

## 8. Conflicts with merged decisions, resolved

| # | Merged text | Conflict | Choice | Rejected, and why |
|---|---|---|---|---|
| 1 | Contract §3.4 to §3.6: the riding-type checks and the riding type's delete "run in one transaction" | Opening one needs `server.DB.Begin`, which the orders story adds; the contract's own backstop is the foreign key | No transaction: one statement per write, W2/W5 read first in the contract's order, the foreign key's `23503` as the authority (erd extension C3). The answers are the contract's | A `BEGIN` / `FOR UPDATE` / count / delete transaction: the same answers, a dependency on another story, and the foreign key still needed as the backstop |
| 2 | Contract §3 intro: "any other method on a path that exists is 405" and "`GET /v1/sessions/current` is 404" | They disagree for `/v1/sessions/current` | The path is registered with its method (`DELETE …`), so every other method is `404` as an unknown collection (§2). SC-A1 reads it this way; `/v1/sessions` and `/v1/me` stay method-less and answer `405` with `Allow` | Registering it method-less: `GET /v1/sessions/current` would then be `405`, which SC-A1 forbids |
| 3 | `erd.md` §3.12: `token_hash` is "SHA-256 of the random session token" | `services.md` §2.2 and `decision.session_binding` make it an HMAC keyed by the credential fingerprint | The HMAC (erd extension C1) | Plain SHA-256: an old session would survive a change of password |
| 4 | Contract §2.7 lists a bike's keys as `id`, `name`, `brand`, … ; §3.11 says "in the order of `schema.json`" | `schema.json` starts `image`, `name` | `schema.json`'s order, taken from the field table (erd extension §5) | Copying the §2.7 order: it is a list of keys, not the order of the checks, and a hand-kept second order would drift |
| 5 | Contract §3.4 and §3.5 step 4: "content type and size, else 400/413" | When both fail the contract does not say which answers | Size first, as `POST /v1/orders` and sign-in do (§3.5, §4.1): a `413` comes before every `400`, and the oversize body is never parsed | Content type first: the oversize body would have to be partly read to know its type |
| 6 | Contract §2.5 rule 1: "JSON decoding uses `DisallowUnknownFields`" | That cannot name the key in a sentence, nor check each field's type and value together in the table's order | A hand-written pass over `map[string]json.RawMessage` (§4.1), which refuses exactly the same bodies and names the key | `DisallowUnknownFields`: the status is the same, the `field` and the order of checks are lost |
| 7 | Contract §2.2: the password "is compared exactly"; `erd.md` §3.12 wants `expires_at > now()` | The first would make a secret with a pasted trailing newline untypeable; the second cannot be tested for 7 days | Strip only trailing `\r\n` from the **secret**, never from the typed value (§3.1); one injected clock for the session (erd extension C2) | Trimming all white space: a password with a leading space could never match |
| 8 | Contract §3.5 to §3.7: no sentence for a missing `If-Match`, a malformed one, an oversize body, or a sign-in body that is malformed or too long | Every refusal needs one sentence (§2.3) | The sentences of §3.5, §4.1 and §4.3. Tests match status, `code`, `field` and key words | Leaving them to the builder: two builders would write two sentences |

## 9. Rollout, dependencies, open questions

**Delivery.** One backend PR after the data story's backend has merged (it widens that story's two handlers and uses its field table, tables and trigger; erd extension §7). It is independent of the orders story except that SC-29, SC-33 and SC-34 need a real order (the verification uses the direct stock update of erd extension §6 until it lands). The frontend PR #20 merges first. Nothing in `docker-compose.yml`, the Dockerfiles or the deploy workflow changes; the backend learns the two secrets from its environment, as it already learns `DATABASE_URL`. `code/backend/.env.example` already lists both names, without values.

**What cannot be verified yet.** `ADMIN_EMAIL` and `ADMIN_PASSWORD` were requested from the owner in the console and are not supplied. Until they are, nobody can sign in on the deployed project (SC-3 holds), the public shop is unaffected, and the local run and the verification use throwaway values that are **not committed** and are reported as such.

| # | Question | Owner | Blocking |
|---|---|---|---|
| 1 | **The throttle trusts the edge.** The client address is the last `X-Forwarded-For` entry, which is the edge proxy's, so a caller cannot choose it **only while the backend is reachable through the edge alone**. If the backend's port were published to the internet, a caller could send its own header and evade the throttle. The local run must show the backend is not published beyond the compose network/localhost | TL (the local run) | No |
| 2 | **The Studio's preview frame and the design walk** sign in by writing a demo visitor into `agentteam:session`; with real sessions, `session.ts` clears that hint on the first non-200 `me`, so a walked or previewed `/admin` page shows the sign-in screen. UI review accepted this and left it to a separate scaffold change by TL (`docs/admin/SRS.md` §7). It does not change this design | TL | No |
| 3 | **A valid cookie with no hint in the browser** is sent to the sign-in page once and sees "Signed in as … Continue" there (story decision 2). Accepted at UI review; the fix would be an edit to `SignedIn` in `App.tsx`, which this story does not own | — | No |
| 4 | **No password-strength rule** (contract §9 question 1) and **no limit on writes** beyond the session | Stakeholder | No |
| 5 | **Cross-origin `VITE_API_URL`**: the owner's cookie and the `ETag` header do not cross origins (contract §9 question 3). The deployment uses the same origin through `/api` | PM | No |

There is nothing for the owner to decide before the build: the SRS §7 defaults (7 days, five failures in 15 minutes, a live bike needs a price, deletes leave reviews and photographs) are the ones this design builds. No `request_setup` is needed: no webhook, redirect address or third-party account is involved.
