# Test cases — Shop admin: real sign-in and saved edits

Module: `admin` · Plan item: "Shop admin: real sign-in and saved edits" (feature, P1)
Requirements: `docs/admin/SRS.md` ADMIN-001 to ADMIN-013 · Story: `docs/admin/stories/shop-admin-real-sign-in-and-saved-edits.md` (SC-1 to SC-83)
Contract: `docs/architecture/services.md` (§2.2 authentication, §2.3 errors, §3.3 to §3.9 endpoints, §3.11 body rules) and `docs/architecture/erd.md` (`sessions`, `touch_record`)

**Risk level: HIGH for sections A and B** (the owner's credentials, a session cookie, every write to the shop's prices and stock, and a race between an owner's save and a customer's order). Every refusal path is written as its own case, boundaries sit on both sides of each limit, and every denial (no session, expired session, signed-out cookie, read-only collection, locked address) has a case. **MEDIUM for sections C to E** (screens that already exist and are made real): each screen state of the story's §5 table is a case, and every failure named in ADMIN-013 happens in a case.

## 0. Conventions used by every case

**Paths.** API paths are written the way the backend mounts them (`/v1/...`); the instrument reaches them as `/api/v1/...` on the story's URL (the edge strips `/api`). Every API write is sent with `Content-Type: application/json` and, unless the case says otherwise, the session cookie.

**Credentials.** The account is `ADMIN_EMAIL=owner@deo.test` and a known `ADMIN_PASSWORD`, written here as `<ADMIN_PASSWORD>`. The value is supplied by the run environment (the owner's console secret, or a throwaway pair that is never committed, as the story's §8 says). **The value is never written in this file or anywhere in the repository** (SC-2). `SESSION` means the `deo_session` cookie returned by a right sign-in; "no session" means no cookie at all.

**Seeded data** (from the data story): 14 bikes (all under 10 in stock, 2 at stock 0, none in draft, 12 "in stock" on Home), 4 riding types (`hardtail`, `trail`, `e-mtb` and one more; `trail` is the riding type of 4 bikes), `talon-2` (live, stock 9, price 18990000, riding type `hardtail`, carries `_source`), `stance-0` (price 66490000), gallery photograph `rocky-descent` (riding type `trail`), reviews with a `bike` (for example `r-minh` → `stance-0`).

**Fixture bike `T`.** Made by `POST /v1/products` with `{"name":"Test Bike","brand":"Giant","category":"hardtail","price":25000000,"status":"live","stock":9,"sizes":["M"]}`; its server-made id is called `T`. It uses `hardtail`, never `trail`, so `trail` stays at 4 bikes. The case that makes `T` deletes it before it ends (`DELETE /v1/products/T` → 204).

**`ORDER(id, qty)`** means `POST /v1/orders` with `{"name":"Minh","phone":"0901234567","delivery":"pickup","payment":"shop","lines":[{"id":"<id>","variant":"M","qty":<qty>}]}`, answering 201 (the orders story owns its rules; here it is only the way stock is lowered by a customer). Until that story is live, "an order for 2" is arranged as a database update of the bike's stock (the `touch_record` trigger bumps `version`, exactly as the order's hold does), and the case says so in its report.

**Restore rule.** A case that changes a seeded record puts it back before it ends: read it first, and at the end `PUT` the original values with the `ETag` of the last read (or reset the row in the database). A row arranged in the database (a session, a gallery photograph, a review) is put back as it was. A stopped backend is started again; a database made unavailable is made available again.

**Sign-in failure budget.** The throttle keys on the client address and counts every `401` from `POST /v1/sessions` (wrong email, wrong password, not configured); `400` and `413` are not failures; a success clears the count. Each case states its failures as `F=n`. Cases whose cumulative `F` would pass 4 for one address are separated by a **backend restart** (it clears the throttle and keeps sessions), or use their own `X-Forwarded-For` address when the instrument can set one. SC-8 and SC-9 run last.

**Error envelope.** Every non-2xx answer is `{"error":"<one sentence>","code":"<UPPER_SNAKE>"[,"field":"<name>"]}` with `Cache-Control: no-store` and an `X-Request-Id` header. A case that names a status also names the `code`; a case that names a field names the `field`.

**Public pages** fetch in the browser, so "a visitor sees" is always `render_url` (never `fetch_url`). **Interactions** (typing, clicking, Enter, Back, reload) are always `interact_page`.

**"Signed in through the form"** means the owner pressed "Sign in" on `/admin/sign-in` in that browser (story decision 2). **"Fresh browser"** means no cookies and empty browser storage.

**Manual cases** are only those no instrument can answer; each states why, and names the evidence Test collects instead (`run_tests`, the PR diff).

---

## A. Backend — the owner account, signing in, the session (SC-1 to SC-12)

**Scenario**: SC-1a — The right email and password sign in
**Given**: The run has `ADMIN_EMAIL=owner@deo.test` and a known `ADMIN_PASSWORD`; no session; throttle clean.
**When**: `POST /v1/sessions` with `{"email":"owner@deo.test","password":"<ADMIN_PASSWORD>"}`.
**Then**: 200; the body is exactly `{"email":"owner@deo.test"}` (one key, no `password` key, no other key); a `Set-Cookie` header sets `deo_session`; `Cache-Control: no-store`. F=0.
Traces: SC-1 (ADMIN-001 AC-1; ADMIN-002 AC-1)
Check: fetch_url

**Scenario**: SC-1b — The email is compared ignoring capitals and spaces around it
**Given**: The same account; throttle clean.
**When**: `POST /v1/sessions` with `{"email":"  OWNER@deo.test ","password":"<ADMIN_PASSWORD>"}`.
**Then**: 200 and the body is exactly `{"email":"owner@deo.test"}` (trimmed and lower-cased, not echoed as typed); the cookie is set. F=0.
Traces: SC-1 (ADMIN-001 AC-3)
Check: fetch_url

**Scenario**: SC-2a — The password's value is in no answer
**Given**: The account of SC-1a; the password value `<ADMIN_PASSWORD>`.
**When**: Make a right sign-in, `GET /v1/me` with `SESSION`, and a refused sign-in (right email, wrong password); search every response body **and every response header** (including `Set-Cookie`) for the password's value.
**Then**: The value is found in none of the three answers' bodies or headers. F=1.
Traces: SC-2 (ADMIN-001 AC-2)
Check: fetch_url

**Scenario**: SC-2b — The password is nowhere else, and what the server holds is a hash
**Given**: The PR's branch; the running service.
**When**: Search the whole repository (code, docs, tests, `code/backend/.env.example`) for the password's value; read how the password is held in the auth package; read the service's log after a right and a wrong sign-in; read the `sessions` table.
**Then**: The value appears nowhere in the repository, and `.env.example` lists `ADMIN_EMAIL` and `ADMIN_PASSWORD` as names with no value; the code holds only an `argon2id` hash made at start (the plain string is dropped); the log holds neither the password nor the typed email; `sessions` holds only 32-byte `token_hash` values and no token and no password.
Traces: SC-2 (ADMIN-001 AC-2, AC-5)
Check: manual — no instrument reads a source tree, a server log or a table's contents over HTTP; Test greps the repository and reads the PR diff, and uses the database tool for `sessions` if it returns rows.

**Scenario**: SC-3a — Without the secrets nobody can sign in, and the shop still works
**Given**: The backend is started with `ADMIN_PASSWORD` unset (run it a second time with `ADMIN_EMAIL` unset instead), the database seeded.
**When**: `GET /healthz`; `GET /v1/products?limit=1`; `POST /v1/sessions` with `{"email":"owner@deo.test","password":"anything"}`; `POST /v1/sessions` with `{"email":"","password":""}`.
**Then**: `/healthz` 200 `{"status":"ok"}`; the list 200 with `total` 14; both sign-ins 401 with the body `{"error":"That email and password do not match.","code":"UNAUTHENTICATED"}`, byte for byte the body a wrong password gets on a configured server (SC-6a); no `Set-Cookie`. F=2 per run. If the run environment cannot start the backend without one of the secrets, report UNVERIFIED naming the missing instrument: a restart of the backend with an environment variable removed.
Traces: SC-3 (ADMIN-001 AC-4)
Check: fetch_url

**Scenario**: SC-3b — The log says once per attempt that the account is not configured
**Given**: The backend of SC-3a (a secret unset); its log is readable.
**When**: Three sign-ins are made.
**Then**: The log holds exactly three lines containing `admin account is not configured` (one per attempt), each with the request id and with no email and no password.
Traces: SC-3 (ADMIN-001 AC-4)
Check: manual — no instrument reads the service's log; Test reads the container log if it can and otherwise reads the unit test of the not-configured path in the PR.

**Scenario**: SC-4 — The cookie is HttpOnly, SameSite=Lax, Path=/ and lasts 7 days
**Given**: Throttle clean; plain HTTP (no `X-Forwarded-Proto`).
**When**: A right `POST /v1/sessions`; read the `Set-Cookie` header and the response `Date`.
**Then**: The cookie is `deo_session=<43 characters of [A-Za-z0-9_-]>` with `Path=/`, `Max-Age=604800`, an `Expires` within 60 seconds of `Date` + 7 days, `HttpOnly` and `SameSite=Lax`; the body holds `email` and no `password`.
Traces: SC-4 (ADMIN-002 AC-1)
Check: fetch_url

**Scenario**: SC-5 — Secure is added over HTTPS and only then
**Given**: Throttle clean.
**When**: A right `POST /v1/sessions` with the header `X-Forwarded-Proto: https`; then another right `POST /v1/sessions` without that header.
**Then**: The first `Set-Cookie` carries `Secure` as well as `HttpOnly; SameSite=Lax`; the second does not carry `Secure`.
Traces: SC-5 (ADMIN-002 AC-7)
Check: fetch_url

**Scenario**: SC-6a — A wrong password and a wrong email are refused with the identical answer
**Given**: The account; throttle clean.
**When**: `POST /v1/sessions` with the right email and a wrong password; then with a wrong email (`someone@else.test`) and the right password.
**Then**: Each is 401 with the body `{"error":"That email and password do not match.","code":"UNAUTHENTICATED"}` (no `field`), the two bodies byte for byte identical, and neither sets a cookie. F=2.
Traces: SC-6 (ADMIN-002 AC-2)
Check: fetch_url

**Scenario**: SC-6b — The password is compared exactly; empty strings are a failure, not a malformed request
**Given**: The account; throttle clean.
**When**: `POST /v1/sessions` with the right email and the right password followed by one space; then with `{"email":"","password":""}`.
**Then**: Both are 401 with the SC-6a body (the password is not trimmed; empty strings are accepted as values that do not match, so not 400). F=2.
Traces: contract (POST /v1/sessions §3.7)
Check: fetch_url

**Scenario**: SC-7a — A body with no password, and a body that is not an object, are 400
**Given**: Throttle clean.
**When**: `POST /v1/sessions` with `{"email":"owner@deo.test"}`; then with `["owner@deo.test","x"]`; then with the text `not json`.
**Then**: Each is 400 with `code` `MALFORMED_REQUEST` and a one-sentence `error`; the first names `field` `password`. None sets a cookie. F=0 (a 400 is not a failure: five more such requests do not lock the address — a following right sign-in is 200).
Traces: SC-7 (ADMIN-002 AC-3)
Check: fetch_url

**Scenario**: SC-7b — Sign-in rejects an undefined key, a non-string, a wrong content type and an oversize body
**Given**: Throttle clean.
**When**: `POST /v1/sessions` with `{"email":"owner@deo.test","password":"x","remember":true}`; with `{"email":"owner@deo.test","password":123}`; with a valid JSON body but `Content-Type: text/plain`; with a body of 9 KiB (a `password` string of 9000 characters).
**Then**: The first is 400 `MALFORMED_REQUEST` naming `remember`; the second 400 `MALFORMED_REQUEST` naming `password`; the third 400 `MALFORMED_REQUEST`; the fourth 413 `PAYLOAD_TOO_LARGE`. None is a failure for the throttle (F=0).
Traces: contract (POST /v1/sessions §3.7, §2.5 rule 1)
Check: fetch_url

**Scenario**: SC-8a — Five failures lock the address, even for the right password
**Given**: A backend restarted so the throttle is clean (or an address of its own).
**When**: Five `POST /v1/sessions` with the right email and a wrong password (each 401); then a sixth with the right password; then a seventh with a wrong password; then an eighth with the body `{}`; then `GET /v1/products?limit=1`.
**Then**: The sixth is 429 with the body `{"error":"Too many attempts. Try again in a few minutes.","code":"RATE_LIMITED"}`, a `Retry-After` header holding a whole number from 1 to 900, and no cookie; the seventh and the eighth are also 429 (the lock is checked before the body); the list is 200 (only sign-in is locked). F=5.
Traces: SC-8 (ADMIN-002 AC-4)
Check: fetch_url

**Scenario**: SC-8b — After 15 minutes without a failure the right password is accepted
**Given**: The locked address of SC-8a, and 15 minutes since its last failed attempt (the throttle's clock advanced).
**When**: `POST /v1/sessions` with the right credentials.
**Then**: 200 with the owner's email and a cookie.
Traces: SC-8 (ADMIN-002 AC-5)
Check: manual — no instrument advances a server's clock by 15 minutes or waits that long; the evidence is the Go unit test of the throttle with the injected clock (Test runs the backend tests with `run_tests` and reads that test's name and result).

**Scenario**: SC-9 — A right sign-in resets the failure count
**Given**: A backend restarted so the throttle is clean.
**When**: Four wrong sign-ins (each 401), one right sign-in, four more wrong sign-ins (each 401), then one right sign-in.
**Then**: The first right sign-in is 200, none of the eight wrong ones is 429, and the last right sign-in is 200 with a cookie (the right one reset the count; without the reset the fifth failure would have locked the address). F=8 across two resets.
Traces: SC-9 (ADMIN-002 AC-6)
Check: fetch_url

**Scenario**: SC-10a — `me` with a valid session answers the owner
**Given**: A signed-in `SESSION`.
**When**: `GET /v1/me` with the cookie.
**Then**: 200, body exactly `{"email":"owner@deo.test"}`, `Cache-Control: no-store`.
Traces: SC-10 (ADMIN-003 AC-1)
Check: fetch_url

**Scenario**: SC-10b — `me` with no session is 401 with one sentence
**Given**: No cookie; then a cookie `deo_session=` with an empty value.
**When**: `GET /v1/me` each time.
**Then**: Each is 401 with `{"error":"You are not signed in.","code":"UNAUTHENTICATED"}` (one sentence, no `field`).
Traces: SC-10 (ADMIN-003 AC-2)
Check: fetch_url

**Scenario**: SC-11a — A session past its 7 days is 401; one with an hour left is still 200
**Given**: A signed-in `SESSION` whose `sessions` row has `expires_at` set to one second in the past (database arranged); a second session whose row has `expires_at` one hour in the future and `created_at` 6 days 23 hours ago.
**When**: `GET /v1/me` with each.
**Then**: The expired one is 401 `UNAUTHENTICATED` "You are not signed in."; the other is 200 `{"email":"owner@deo.test"}`. Rows are removed afterwards.
Traces: SC-11 (ADMIN-003 AC-3)
Check: fetch_url

**Scenario**: SC-11b — A cookie with one character changed, or garbage, is 401 (never 500)
**Given**: A valid `SESSION` value.
**When**: `GET /v1/me` with the value's last character changed to a different one; with 43 `A` characters; with `deo_session=x`; with a 500-character value.
**Then**: Each is 401 `UNAUTHENTICATED`; none is 500; the untouched `SESSION` is still 200.
Traces: SC-11 (ADMIN-003 AC-3)
Check: fetch_url

**Scenario**: SC-11c — A session made for another email stops working
**Given**: A valid `SESSION` whose `sessions` row has its `email` changed to `other@deo.test` (database arranged; this stands for the owner having changed `ADMIN_EMAIL`).
**When**: `GET /v1/me` with the cookie; then put the row's email back and ask again.
**Then**: 401 `UNAUTHENTICATED` while the email differs; 200 once it is restored.
Traces: SC-11 (ADMIN §3; ADMIN-003 AC-3)
Check: fetch_url

**Scenario**: SC-11d — A cookie issued before the secrets changed is refused after a restart
**Given**: A valid `SESSION`; then the backend restarted with a different `ADMIN_PASSWORD` (and, separately, a different `ADMIN_EMAIL`).
**When**: `GET /v1/me` with the old cookie.
**Then**: 401 `UNAUTHENTICATED` in both runs (the stored token hash was keyed by the old secrets).
Traces: SC-11 (ADMIN §3; ADMIN-003 AC-3)
Check: manual — no instrument changes the container's secrets; the evidence is the Go unit test of the credential fingerprint and the session lookup in the PR (read with `run_tests`).

**Scenario**: SC-12a — Sign-out answers 204, clears the cookie and kills the old value
**Given**: A signed-in `SESSION`.
**When**: `DELETE /v1/sessions/current` with the cookie; then `GET /v1/me` with the same old cookie value.
**Then**: The first is 204 with no body and a `Set-Cookie` that clears `deo_session` (empty value, `Path=/`, `Max-Age=0`, `HttpOnly`, `SameSite=Lax`); the second is 401 `UNAUTHENTICATED`.
Traces: SC-12 (ADMIN-003 AC-7 (cookie), AC-8)
Check: fetch_url

**Scenario**: SC-12b — Signing out with no session also succeeds and changes nothing
**Given**: One valid `SESSION` (B), and a request that carries no cookie.
**When**: `DELETE /v1/sessions/current` with no cookie; then `GET /v1/me` with B.
**Then**: The sign-out is 204 with the clearing cookie; B's `me` is still 200 (nothing else was touched).
Traces: SC-12 (ADMIN-003 AC-9)
Check: fetch_url

**Scenario**: SC-A1 — Session routes keep to their shapes and methods
**Given**: Any state.
**When**: `GET /v1/sessions`; `PUT /v1/sessions`; `POST /v1/me`; `GET /v1/sessions/current`; `GET /v1/me/`; `GET /v1/ME`.
**Then**: `GET /v1/sessions` and `PUT /v1/sessions` are 405 `METHOD_NOT_ALLOWED` with `Allow: POST`; `POST /v1/me` is 405 with `Allow: GET`; `GET /v1/sessions/current` is 404 `NOT_FOUND` (a session is never readable); the trailing-slash and capitalised paths are 404 `NOT_FOUND`; every body is in the error envelope.
Traces: contract (§3 route shapes; §3.7 to §3.9)
Check: fetch_url

**Scenario**: SC-A2 — Every error from the account endpoints carries the envelope and its headers
**Given**: The refusals of SC-6a, SC-7a, SC-10b and SC-A1.
**When**: Read their headers and bodies; send one request with `X-Request-Id: trace-123` and one with `X-Request-Id: bad id!`.
**Then**: Every body has exactly `error` (string) and `code` (`UPPER_SNAKE`), plus `field` only on a 400 that names one; every response has `Cache-Control: no-store`; the valid request id is echoed, the invalid one is replaced by 32 hex characters; the request id is never in a body.
Traces: contract (§2.1, §2.3)
Check: fetch_url

**Scenario**: SC-A3 — With the database refusing connections the account endpoints say 503, and sign-out does not clear the cookie
**Given**: A valid `SESSION`; the database made unavailable while the service keeps running.
**When**: `GET /v1/me` with the cookie; `POST /v1/sessions` with right credentials; `DELETE /v1/sessions/current` with the cookie; `GET /healthz`.
**Then**: Each is 503 `UNAVAILABLE` with the sentence "The service is not ready." (`/healthz` 503 too); the sign-out carries no `Set-Cookie` that clears `deo_session` (the owner is never told they are signed out while the session still works). With the database available again, `GET /v1/me` with the same cookie is 200.
Traces: contract (§3.7, §3.8, §3.9 errors)
Check: fetch_url

**Scenario**: SC-A4 — A failed store query is 500 INTERNAL, never 503 and never a leak
**Given**: The `sessions` table renamed in the database (restored at the end); the service running.
**When**: `POST /v1/sessions` with right credentials; `GET /v1/me` with any cookie value.
**Then**: Each is 500 with `code` `INTERNAL` and the sentence "Something went wrong. Please try again."; no cookie is set; the body names no table, SQL or driver text.
Traces: contract (§2.3 INTERNAL; §3.7, §3.8)
Check: fetch_url

---

## B. Backend — who may write, what is saved, what is deleted (SC-13 to SC-35)

**Scenario**: SC-13a — With no session a bike cannot be added, changed or deleted
**Given**: No cookie; the seeded bikes (`total` 14); `talon-2` read once (its record and `ETag` kept).
**When**: `POST /v1/products` with `{"name":"Test Bike","brand":"Giant","category":"trail","price":25000000,"status":"draft","stock":3}`; `PUT /v1/products/talon-2` with the full valid record (stock 5) and the kept `If-Match`; `DELETE /v1/products/talon-2`.
**Then**: Each is 401 `{"error":"You are not signed in.","code":"UNAUTHENTICATED"}`; afterwards `GET /v1/products?limit=100` has `total` 14 and no "Test Bike"; `GET /v1/products/talon-2` equals the kept record key for key with the same `ETag`.
Traces: SC-13 (ADMIN-004 AC-1)
Check: fetch_url

**Scenario**: SC-13b — With no session a riding type cannot be added, changed or deleted
**Given**: No cookie; 4 riding types; `trail` read once (record and `ETag` kept).
**When**: `POST /v1/categories` with `{"name":"Gravel"}`; `PUT /v1/categories/trail` with the full valid record and the kept `If-Match`; `DELETE /v1/categories/trail`.
**Then**: Each is 401 `UNAUTHENTICATED` "You are not signed in."; `GET /v1/categories` has `total` 4 and no "Gravel"; `trail` is unchanged with the same `ETag`.
Traces: SC-13 (ADMIN-004 AC-2)
Check: fetch_url

**Scenario**: SC-13c — An expired session and a signed-out cookie are refused like no session
**Given**: A `SESSION` whose row is expired (`expires_at` one second in the past, database arranged); a second `SESSION` that was signed out with `DELETE /v1/sessions/current`.
**When**: With each cookie, `POST /v1/products` (valid body), `PUT /v1/products/talon-2` (valid body, a valid `If-Match`) and `DELETE /v1/products/talon-2`; the same three on `categories` with `trail`.
**Then**: All twelve are 401 `UNAUTHENTICATED` "You are not signed in."; `total` of bikes is still 14 and of riding types 4; `talon-2` and `trail` are unchanged.
Traces: SC-13 (ADMIN-004 AC-3; ADMIN-006 AC-10)
Check: fetch_url

**Scenario**: SC-13d — The session is checked before the body, the `If-Match` and the record
**Given**: No cookie.
**When**: `PUT /v1/products/talon-2` with no `If-Match` and the body `{"colourful":"red"}`; `POST /v1/products` with `Content-Type: text/plain`; `DELETE /v1/products/no-such-bike`; `PUT /v1/categories/no-such-type` with an invalid body.
**Then**: Each is 401 `UNAUTHENTICATED` (not 400, not 428, not 404).
Traces: SC-13 (contract §3.4 to §3.6 order of checks)
Check: fetch_url

**Scenario**: SC-14a — A signed-in owner still cannot write the read-only collections
**Given**: A valid `SESSION`; ids read from `GET /v1/reviews`, `GET /v1/gallery` (for example `rocky-descent`), `GET /v1/team`; the six collections' records read once.
**When**: `PUT /v1/reviews/r-minh` with its record; `PUT /v1/shop/shop` with its record; `DELETE /v1/gallery/rocky-descent`; `POST /v1/team` with `{"name":"Someone"}`; `POST /v1/services` with `{"name":"x"}`; `DELETE /v1/faqs/<an id>`.
**Then**: Each is 405 `METHOD_NOT_ALLOWED` with the header `Allow: GET`; every record read afterwards equals the one read before, key for key; `gallery` still holds `rocky-descent`.
Traces: SC-14 (ADMIN-004 AC-5)
Check: fetch_url

**Scenario**: SC-14b — Without a session the same writes are 405, not 401; an unknown collection is 404
**Given**: No cookie.
**When**: The six requests of SC-14a; `POST /v1/unicycles` and `PUT /v1/unicycles/x` (once with and once without a session); `PATCH /v1/products/talon-2`; `DELETE /v1/products`.
**Then**: The six are 405 `METHOD_NOT_ALLOWED` with `Allow: GET`; the two unknown-collection writes are 404 `NOT_FOUND` "That collection does not exist." with and without a session; `PATCH /v1/products/talon-2` is 405 with `Allow: GET, PUT, DELETE`; `DELETE /v1/products` is 405 with `Allow: GET, POST`.
Traces: SC-14 (ADMIN-004 AC-5; contract §3 route shapes, §3.4 check order)
Check: fetch_url

**Scenario**: SC-15a — Reading stays public
**Given**: No cookie.
**When**: `GET /v1/products` (with `limit=100`) and `GET /v1/products/talon-2`; `GET /v1/categories`.
**Then**: Each is 200; the bikes list has `total` 14 and `items`, `page`, `limit`; `talon-2` is returned with `status` `live`.
Traces: SC-15 (ADMIN-004 AC-6)
Check: fetch_url

**Scenario**: SC-15b — A bike or riding type read by id carries the `ETag`; the other collections do not
**Given**: No cookie.
**When**: `GET /v1/products/talon-2`; `GET /v1/categories/trail`; `GET /v1/reviews/r-minh`; `GET /v1/shop/shop`.
**Then**: The first two carry `ETag` holding one quoted whole number (for example `"7"`); the other two carry no `ETag`; no response body of any of them has a `version` key.
Traces: contract (§3.3); SC-33 (the token the save needs)
Check: fetch_url

**Scenario**: SC-16a — A signed-in owner adds a bike and the server makes its id
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `{"name":"Test Bike","brand":"Giant","category":"trail","price":25000000,"status":"draft","stock":3}`; read the answer; `GET /v1/products/<id>`; then repeat the `POST` once more.
**Then**: 201; the body has an `id` matching `^products-[a-z0-9]{6}$` and exactly the sent values (`name`, `brand`, `category`, `price`, `status`, `stock`), no `_source`, no `version`, no `null`; the header `ETag: "1"`; the `GET` returns the record equal to the answer key for key; the second `POST` gets a different `id`. Both bikes are deleted afterwards (204).
Traces: SC-16 (ADMIN-005 AC-1)
Check: fetch_url

**Scenario**: SC-16b — A given id is used, and a taken id is 409
**Given**: A valid `SESSION`; no bike `my-bike-1`.
**When**: `POST /v1/products` with `{"id":"my-bike-1","name":"Test Bike","status":"draft"}`; then the same body again.
**Then**: The first is 201 with `id` `my-bike-1`; the second is 409 with `code` `ID_TAKEN` and the sentence "That id is already used."; there is exactly one `my-bike-1`. It is deleted afterwards.
Traces: SC-16 (ADMIN-005 AC-2)
Check: fetch_url

**Scenario**: SC-16c — A riding type with a name only is added with a server-made id
**Given**: A valid `SESSION`; 4 riding types.
**When**: `POST /v1/categories` with `{"name":"Gravel"}`; `GET /v1/categories/<id>`.
**Then**: 201; the body is exactly `{"id":"categories-xxxxxx","name":"Gravel"}` with the id matching `^categories-[a-z0-9]{6}$` (no other key); `ETag: "1"`; the `GET` returns the same; `GET /v1/categories` has `total` 5. It is deleted afterwards (204).
Traces: SC-16 (ADMIN-005 AC-19)
Check: fetch_url

**Scenario**: SC-17a — A change replaces the record and reading shows it
**Given**: A valid `SESSION`; `talon-2` read (record `R`, `ETag` `E`; stock 9).
**When**: `PUT /v1/products/talon-2` with `R` minus `_source`, `stock` set to 5, and `If-Match: E`; then `GET /v1/products/talon-2`.
**Then**: 200 with the saved record: `stock` 5 and every other key as sent, `_source` still present with its seeded value; `ETag` is `E` + 1; the `GET` equals the answer key for key. Put back: `PUT` the original with the new `ETag`.
Traces: SC-17 (ADMIN-005 AC-3; ADMIN-007 AC-1)
Check: fetch_url

**Scenario**: SC-17b — A field the body leaves out is gone from the saved record
**Given**: A valid `SESSION`; `T` made with a `colour` of "Matte black" and a `fact` of "29-inch, 120 mm" (the record and `ETag` read).
**When**: `PUT /v1/products/T` with the same body without `colour`.
**Then**: 200; the answer and a following `GET` have no `colour` key (not `null`, not `""`) and still have `fact`. `T` is deleted afterwards.
Traces: SC-17 (ADMIN-005 AC-4)
Check: fetch_url

**Scenario**: SC-17c — A change to a record that does not exist is 404
**Given**: A valid `SESSION`.
**When**: `PUT /v1/products/no-such-bike` with a valid body and `If-Match: "1"`; `PUT /v1/categories/no-such-type` with `{"name":"X"}` and `If-Match: "1"`.
**Then**: Each is 404 with `code` `NOT_FOUND` and the sentence "That record does not exist."; `GET /v1/products?limit=100` still has `total` 14.
Traces: SC-17 (ADMIN-005 AC-5)
Check: fetch_url

**Scenario**: SC-18a — A key the contract does not define is 400 naming it
**Given**: A valid `SESSION`; `T` and its `ETag`.
**When**: `POST /v1/products` with `{"name":"X","colourful":"red"}`; `PUT /v1/products/T` with a valid body plus `"colourful":"red"`; the same with `"_source":"web:x"`, with `"version":3` and with `"seq":1`; `POST /v1/categories` with `{"name":"X","colourful":"red"}`.
**Then**: Each is 400 with `code` `MALFORMED_REQUEST` and `field` equal to the undefined key; the sentence contains the key (and says "…of a bike." or "…of a riding type."); nothing is saved (never a 200 that ignores the key).
Traces: SC-18 (ADMIN-005 AC-6; contract §2.5 rule 1)
Check: fetch_url

**Scenario**: SC-18b — A bike with no model name, or a blank one, is 400 naming `name`
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `{"brand":"Giant"}`; with `{"name":""}`; with `{"name":"   "}`; with `{"name":null}`.
**Then**: Each is 400 with `code` `VALIDATION_FAILED`, `field` `name` and the sentence "The field “name” is needed."; `total` of bikes is still 14.
Traces: SC-18 (ADMIN-005 AC-7)
Check: fetch_url

**Scenario**: SC-18c — A riding type with no name is 400 naming `name`
**Given**: A valid `SESSION`.
**When**: `POST /v1/categories` with `{"line":"For hills"}`; with `{"name":""}`.
**Then**: Each is 400 `VALIDATION_FAILED` with `field` `name`; `total` of riding types is still 4.
Traces: SC-18 (ADMIN-005 AC-20)
Check: fetch_url

**Scenario**: SC-19a — A bad `price` is 400 naming `price`
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `{"name":"X","price":"cheap"}`; with `{"name":"X","price":-1}`; with `{"name":"X","price":1.5}`; with `{"name":"X","was":-1}`; with `{"name":"X","price":9007199254740992}`; then with `{"name":"X","status":"draft","price":9007199254740991}`.
**Then**: The first is 400 `MALFORMED_REQUEST` (a string where a number is expected), `field` `price`; the second, third and fifth are 400 `VALIDATION_FAILED`, `field` `price`; the fourth is 400 `VALIDATION_FAILED`, `field` `was`; the sixth is 201 (the top of the range is accepted) and is deleted afterwards.
Traces: SC-19 (ADMIN-005 AC-8)
Check: fetch_url

**Scenario**: SC-19b — A bad `stock` is 400 naming `stock`
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `{"name":"X","stock":-1}`; with `{"name":"X","stock":2.5}`; with `{"name":"X","stock":"3"}`; then with `{"name":"X","status":"draft","stock":0}`.
**Then**: The first two are 400 `VALIDATION_FAILED`, `field` `stock`; the third is 400 `MALFORMED_REQUEST`, `field` `stock`; the fourth is 201 with `stock` 0 (kept as 0, not dropped) and is deleted afterwards.
Traces: SC-19 (ADMIN-005 AC-9)
Check: fetch_url

**Scenario**: SC-19c — Lists, booleans and numbers must have the right type and size
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `{"name":"X","sizes":"M"}`; with `{"name":"X","sizes":["M",5]}`; with `{"name":"X","deal":"yes"}`; with `{"name":"X","year":"2027"}`; with `sizes` holding 51 entries; then with `sizes` holding 50 entries and, separately, `"sizes":[]`.
**Then**: The first four are 400 `MALFORMED_REQUEST` naming `sizes`, `sizes`, `deal`, `year`; the 51-entry list is 400 `VALIDATION_FAILED`, `field` `sizes`, sentence "The field “sizes” must have at most 50 entries."; the 50-entry list and the empty list are 201 (the empty list is kept as `[]`). Created bikes are deleted afterwards.
Traces: SC-19 (ADMIN-005 AC-13)
Check: fetch_url

**Scenario**: SC-20a — `status`, `wheel` and `brand` accept only their closed values
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `{"name":"X","status":"sold"}`; with `{"name":"X","wheel":"26"}`; with `{"name":"X","wheel":29}`; with `{"name":"X","brand":"Trek"}`; then with `{"name":"X","status":"draft","wheel":"29","brand":"Liv"}` and with `{"name":"X","status":"draft","wheel":"mixed"}`.
**Then**: The `status`, `wheel` "26" and `brand` bodies are 400 `VALIDATION_FAILED` with `field` `status`, `wheel`, `brand` (the status sentence reads "The field “status” must be one of: draft, live."); the number `29` is 400 `MALFORMED_REQUEST`, `field` `wheel`; the last two are 201. Created bikes are deleted afterwards.
Traces: SC-20 (ADMIN-005 AC-10, AC-11)
Check: fetch_url

**Scenario**: SC-20b — `category` must name a riding type that exists
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `{"name":"X","status":"draft","category":"unicycle"}`; then with `"category":"trail"`.
**Then**: The first is 400 `VALIDATION_FAILED`, `field` `category`, sentence "The field “category” must be the id of a riding type that exists."; the second is 201. The created bike is deleted afterwards.
Traces: SC-20 (ADMIN-005 AC-12)
Check: fetch_url

**Scenario**: SC-21a — A live bike with no price, or a price of 0, is refused with the fixed sentence
**Given**: A valid `SESSION`; `T` and its `ETag`.
**When**: `POST /v1/products` with `{"name":"X","status":"live"}`; with `{"name":"X","status":"live","price":0}`; with `{"name":"X","status":"live","price":null}`; `PUT /v1/products/T` with the valid record but `status` `live` and `price` removed.
**Then**: Each is 400 `VALIDATION_FAILED`, `field` `price`, the sentence exactly "A live bike needs a price."; nothing is created or changed.
Traces: SC-21 (ADMIN-005 AC-14)
Check: fetch_url

**Scenario**: SC-21b — A draft or an unstated status needs no price; a live bike at price 1 is accepted
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `{"name":"X","status":"draft"}`; with `{"name":"X","status":"draft","price":0}`; with `{"name":"X"}`; with `{"name":"X","status":"live","price":1}`.
**Then**: All four are 201 (no 400 for a missing price on a draft or on a bike with no status; the boundary 1 is accepted for a live bike). The four bikes are deleted afterwards.
Traces: SC-21 (ADMIN-005 AC-15)
Check: fetch_url

**Scenario**: SC-22a — An empty optional value leaves the key out
**Given**: A valid `SESSION`; `T` made with `was` 29990000, `fact` "A line" and `sizes_out` ["S"] (record and `ETag` read).
**When**: `PUT /v1/products/T` with the valid record but `"was":""`, `"fact":""` and `"sizes_out":null`; `GET /v1/products/T`.
**Then**: 200; the answer and the `GET` have no `was`, no `fact` and no `sizes_out` key; the keys `sizes` (`["M"]`) and `price` are still there. `T` is deleted afterwards.
Traces: SC-22 (ADMIN-005 AC-16)
Check: fetch_url

**Scenario**: SC-22b — No saved record has a `null` anywhere
**Given**: The writes of SC-22a and SC-16a done.
**When**: Read the raw text of `GET /v1/products?limit=100` and `GET /v1/categories?limit=100`.
**Then**: Neither text contains `null` as a value (no `:null`, no `: null`) in any record.
Traces: SC-22 (ADMIN-005 AC-16; DATA-001 AC-4)
Check: fetch_url

**Scenario**: SC-23a — On a change the body's id must equal the address
**Given**: A valid `SESSION`; `T` and its `ETag`.
**When**: `PUT /v1/products/T` with a valid record whose `id` is `other-bike`; then with `id` `T`.
**Then**: The first is 400 `VALIDATION_FAILED`, `field` `id`, the sentence "The field “id” must equal the address's id."; the second is 200.
Traces: SC-23 (ADMIN-005 AC-17)
Check: fetch_url

**Scenario**: SC-23b — An id must be lowercase letters, digits and hyphens, 1 to 64, not starting with a hyphen
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `"id":"My Bike"`; with `"id":"Bike1"`; with `"id":"-bike"`; with a 65-character id of `a`; then with a 64-character id of `a` and with `"id":"a"` (each body also has `"name":"X","status":"draft"`).
**Then**: The first four are 400 `VALIDATION_FAILED` with `field` `id` and the sentence "The field “id” must be 1 to 64 lowercase letters, digits or hyphens, and not start with a hyphen."; the last two are 201. The created bikes are deleted afterwards.
Traces: SC-23 (ADMIN-005 AC-18)
Check: fetch_url

**Scenario**: SC-24a — A text of 5000 characters is accepted, one of 5001 is 400
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `description` of 5000 `a` characters; with 5001 `a`; with 5000 `Đ` characters (10,000 bytes); with `name` of 5001 characters; `POST /v1/categories` with `line` of 5001 characters.
**Then**: The 5000-character bodies (both kinds) are 201; the 5001-character `description` is 400 `VALIDATION_FAILED`, `field` `description`, "The field “description” must be at most 5000 characters."; the 5001-character `name` is 400 with `field` `name`; the 5001-character `line` is 400 with `field` `line` (characters are code points, not bytes). Created records are deleted afterwards.
Traces: SC-24 (ADMIN-005 AC-21)
Check: fetch_url

**Scenario**: SC-25 — After any refused body nothing is saved
**Given**: A valid `SESSION`; counts and `talon-2` read first.
**When**: Send, one by one: `POST /v1/products` with `colourful`, with no name, with `price` -1, with 5001 characters, with `category` `unicycle`; `PUT /v1/products/talon-2` with `price` "cheap", with `status` "sold", with a changed body `id`; then read the collections.
**Then**: Every request is 400; `GET /v1/products?limit=100` has `total` 14 and no record named "X"; `talon-2` equals the record read first, key for key, with the same `ETag` (a refused body does not bump the version).
Traces: SC-25 (ADMIN-005 AC-22)
Check: fetch_url

**Scenario**: SC-26a — A bike with no orders is deleted: 204, then 404
**Given**: A valid `SESSION`; `T` live (so it shows on the shop).
**When**: `DELETE /v1/products/T`; `GET /v1/products/T`; `GET /v1/products?limit=100`; `DELETE /v1/products/T` again.
**Then**: The first is 204 with no body; the `GET` is 404 `NOT_FOUND` "That record does not exist."; the list has `total` 14 and no `T`; the second delete is 404 `NOT_FOUND`.
Traces: SC-26 (ADMIN-006 AC-1, AC-9)
Check: fetch_url

**Scenario**: SC-26b — A deleted bike leaves the shop, Home and search
**Given**: A valid `SESSION`; `T` named "Zebra Test Bike", live, in stock; visitor pages read once and showing it on `/shop` and at `/search?q=zebra`.
**When**: `DELETE /v1/products/T`; then a visitor opens `/shop`, `/` and `/search?q=zebra`.
**Then**: "Zebra Test Bike" appears on none of the three pages; `/search?q=zebra` shows its empty-result state.
Traces: SC-26 (ADMIN-006 AC-1)
Check: render_url

**Scenario**: SC-26c — Deleting an id that does not exist is 404
**Given**: A valid `SESSION`.
**When**: `DELETE /v1/products/no-such-bike`; `DELETE /v1/categories/no-such-type`.
**Then**: Each is 404 `NOT_FOUND` "That record does not exist."
Traces: SC-26 (ADMIN-006 AC-9)
Check: fetch_url

**Scenario**: SC-27a — A riding type used by exactly one bike is 409 "…of 1 bike."
**Given**: A valid `SESSION`; a riding type "Solo" made by `POST /v1/categories` (id `S`), and one bike `X1` with `category` `S` (status `draft`).
**When**: `DELETE /v1/categories/S`; then `GET /v1/categories/S`; then delete `X1` and `DELETE /v1/categories/S` again.
**Then**: The first is 409 `IN_USE` with the sentence exactly "Solo is still the riding type of 1 bike. Give those bikes another riding type first."; the `GET` is 200 (the type still exists); after `X1` is deleted the second delete is 204.
Traces: SC-27 (ADMIN-006 AC-3)
Check: fetch_url

**Scenario**: SC-27b — A riding type used by many bikes is 409 with the count
**Given**: A valid `SESSION`; the seeded database, where `GET /v1/products?category=trail&limit=100` has `total` 4 (live and draft bikes both count).
**When**: `DELETE /v1/categories/trail`; then `GET /v1/categories/trail` and `GET /v1/products?category=trail&limit=100`.
**Then**: 409 `IN_USE` with the sentence exactly "Trail is still the riding type of 4 bikes. Give those bikes another riding type first."; the type is still 200; the bikes list still has `total` 4.
Traces: SC-27 (ADMIN-006 AC-2)
Check: fetch_url

**Scenario**: SC-27c — A riding type no bike uses is deleted
**Given**: A valid `SESSION`; a riding type "Gravel" just made (no bike uses it).
**When**: `DELETE /v1/categories/<its id>`; `GET /v1/categories/<its id>`; `GET /v1/categories`.
**Then**: 204; the `GET` is 404; `total` is back to 4.
Traces: SC-27 (ADMIN-006 AC-4)
Check: fetch_url

**Scenario**: SC-28a — A deleted riding type leaves its gallery photograph listed, without a riding type
**Given**: A valid `SESSION`; a new riding type `X` made by the API; gallery photograph `rocky-descent` has its `type` set to `X` (database arranged, original `trail` recorded); `GET /v1/gallery` read for its `total`.
**When**: `DELETE /v1/categories/X`; `GET /v1/gallery/rocky-descent`; `GET /v1/gallery`.
**Then**: 204; `rocky-descent` is 200 with no `type` key; the gallery `total` is unchanged and still lists it. Put back: `type` `trail`.
Traces: SC-28 (ADMIN-006 AC-5)
Check: fetch_url

**Scenario**: SC-28b — A deleted bike leaves its review, without a bike
**Given**: A valid `SESSION`; `T` made; review `r-minh` has its `bike` set to `T` (database arranged, original `stance-0` recorded); `GET /v1/reviews` read for its `total`.
**When**: `DELETE /v1/products/T`; `GET /v1/reviews/r-minh`; `GET /v1/reviews`.
**Then**: 204; the review is 200 with no `bike` key; the reviews `total` is unchanged. Put back: `bike` `stance-0`.
Traces: SC-28 (ADMIN-006 AC-6)
Check: fetch_url

**Scenario**: SC-29a — A bike that has an order is still deleted, and a bag holding it is then refused
**Given**: A valid `SESSION`; `T` live with stock 9, size M; `ORDER(T, 1)` answered 201.
**When**: `DELETE /v1/products/T`; then `ORDER(T, 1)` again.
**Then**: The delete is 204; the second order is 409 with `code` `BIKE_UNAVAILABLE` and the sentence exactly "A bike in your bag is no longer in the shop."
Traces: SC-29 (ADMIN-006 AC-7, AC-8)
Check: fetch_url

**Scenario**: SC-29b — The placed order's lines are unchanged after the bike is deleted
**Given**: The order of SC-29a, whose lines (`product_id`, name, price, qty, variant) are read from `order_lines` before the delete.
**When**: The bike is deleted; `order_lines` is read again.
**Then**: The rows are identical to the ones read before (the line holds a copy of the bike's id, name and price, not a key).
Traces: SC-29 (ADMIN-006 AC-7)
Check: manual — the API never serves an order (ORD-001.5), so no HTTP instrument can read order lines; the evidence is a read of `order_lines` through the database tool if it returns rows, and the `order_lines` definition in the migration (no foreign key to `products`).

**Scenario**: SC-30a — A saved stock of 0 shows on the shop, the bike's page and Home
**Given**: A valid `SESSION`; `talon-2` at stock 9; Home read first and showing "12" before "bikes in stock"; the shop card and the bike page showing no "Out of stock".
**When**: Save `talon-2` with stock 0 (`PUT` with its `ETag`); read `GET /v1/products/talon-2`; then a visitor opens `/shop`, `/shop/talon-2` and `/`.
**Then**: The read shows `stock` 0 (fetch of the record); on `/shop` the `talon-2` card reads "Out of stock"; on `/shop/talon-2` "Out of stock" shows and the button reads "Ask when it is back"; on `/` the count before "bikes in stock" is 11 (one lower than 12). Put `talon-2` back to stock 9.
Traces: SC-30 (ADMIN-007 AC-1, AC-2)
Check: render_url

**Scenario**: SC-30b — The visitor's read of the record shows the new stock at once
**Given**: A valid `SESSION`; `talon-2` at stock 9.
**When**: Save `talon-2` with stock 0; immediately `GET /v1/products/talon-2` with no cookie.
**Then**: 200 with `stock` 0 and `Cache-Control: no-store`; no older copy is served.
Traces: SC-30 (ADMIN-007 AC-1)
Check: fetch_url

**Scenario**: SC-31a — A bike saved as draft leaves every public list
**Given**: A valid `SESSION`; `talon-2` live (appears on `/shop`, on `/`, at `/search?q=talon`, and in the "More" row of another hardtail's page).
**When**: Save `talon-2` with `status` `draft`; a visitor opens `/shop`, `/`, `/search?q=talon` and the page of another `hardtail` bike; `GET /v1/products?status=live&limit=100`; `GET /v1/products/talon-2`.
**Then**: "Talon 2" is on none of the four pages (including the "More" row); the live list has no `talon-2`; `GET /v1/products/talon-2` is still 200 with `status` `draft`. Put `talon-2` back to `live`.
Traces: SC-31 (ADMIN-007 AC-3)
Check: render_url

**Scenario**: SC-31b — A saved price shows on the card and the page and drives the price order
**Given**: A valid `SESSION`; `talon-2` priced 18990000 (shown `18.990.000 ₫`).
**When**: Save `talon-2` with `price` 17990000; a visitor opens `/shop` and `/shop/talon-2`; `GET /v1/products?status=live&sort=price&limit=100`.
**Then**: The card and the page show `17.990.000 ₫` and not `18.990.000 ₫` as the price; in the API's price order `talon-2` sits where 17990000 belongs (between its neighbours by price). Put the price back to 18990000.
Traces: SC-31 (ADMIN-007 AC-4)
Check: render_url

**Scenario**: SC-32a — A refused save leaves the public pages as they were
**Given**: A valid `SESSION`; `/shop/talon-2` and `/shop` read first; `talon-2` and its `ETag` read.
**When**: `PUT /v1/products/talon-2` with `price` 1.5 (400), then with `colourful` (400), then with a stale `If-Match` and `price` 1000 (409); a visitor reopens `/shop/talon-2` and `/shop`.
**Then**: The three answers are 400, 400, 409; the pages show the same price, stock and status as before (`18.990.000 ₫`, in stock); `GET /v1/products/talon-2` is key for key the record read first.
Traces: SC-32 (ADMIN-007 AC-7)
Check: render_url

**Scenario**: SC-32b — A deleted bike is 404 to a visitor and gone from the shop
**Given**: A valid `SESSION`; `T` named "Zebra Test Bike", live.
**When**: `DELETE /v1/products/T`; then, with no cookie, `GET /v1/products/T`; a visitor opens `/shop`.
**Then**: The read is 404 `NOT_FOUND`; the shop shows no "Zebra Test Bike" and the grid has one card fewer than before.
Traces: SC-32 (ADMIN-007 AC-8)
Check: render_url

**Scenario**: SC-33a — A save made from before an order is refused and overwrites nothing
**Given**: A valid `SESSION`; `T` live (stock 9, price 25000000, size M) read as record `R1` with `ETag` `E1`; then `ORDER(T, 2)` accepted (201) so stock is 7 (until the orders story is live, the stock is set to 7 in the database).
**When**: `PUT /v1/products/T` with `R1` but `price` 23000000 and `If-Match: E1`; then `GET /v1/products/T`.
**Then**: 409 with `code` `STALE_RECORD` and the sentence exactly "This record changed while you were editing it — for example an order was placed. Reload and make your change again."; the `GET` shows `stock` 7 and `price` 25000000.
Traces: SC-33 (ADMIN-008 AC-1)
Check: fetch_url

**Scenario**: SC-33b — After reading again the save is accepted and the token can be reused
**Given**: The state of SC-33a; `T` read again (`stock` 7, `ETag` `E2`).
**When**: `PUT /v1/products/T` with that record but `price` 23000000 and `If-Match: E2`; then at once a second `PUT` with `price` 22000000 and the `ETag` returned by the first answer (no re-read).
**Then**: The first is 200 with `stock` 7 and `price` 23000000, and an `ETag` of `E2` + 1; the second is 200 with `price` 22000000. `T` is deleted afterwards.
Traces: SC-33 (ADMIN-008 AC-2; contract §3.5)
Check: fetch_url

**Scenario**: SC-33c — Two clients that read the same bike: the second to save is refused and the first change stands
**Given**: A valid `SESSION`; `T` read by client A and by client B with the same `ETag` `E`.
**When**: A `PUT`s `price` 24000000 with `E` (200); then B `PUT`s `price` 21000000 with `E`; then `GET /v1/products/T`.
**Then**: B's answer is 409 `STALE_RECORD` with the sentence of SC-33a; the `GET` shows `price` 24000000. `T` is deleted afterwards.
Traces: SC-33 (ADMIN-008 AC-3)
Check: fetch_url

**Scenario**: SC-33d — A `PUT` with no `If-Match` is 428; a malformed `If-Match` is 400
**Given**: A valid `SESSION`; `T` and its `ETag` `E`.
**When**: `PUT /v1/products/T` with a valid body and no `If-Match`; then with `If-Match` set to `7` (unquoted), `W/"7"`, `*`, `"abc"` and `"7", "8"`.
**Then**: The first is 428 `PRECONDITION_REQUIRED`; each of the five malformed values is 400 `MALFORMED_REQUEST`; `T` is unchanged. `T` is deleted afterwards.
Traces: SC-33 (ADMIN-008 AC-1 (the token); contract §3.5 steps 7)
Check: fetch_url

**Scenario**: SC-34a — Saving a stock and then an order leaves the lower stock
**Given**: A valid `SESSION`; `T` live, stock 9, size M, read with its `ETag`.
**When**: `PUT` `T` with stock 5 (200); then `ORDER(T, 1)`; then `GET /v1/products/T`.
**Then**: The save is 200, the order is 201, and the read shows `stock` 4. `T` is deleted afterwards.
Traces: SC-34 (ADMIN-008 AC-4)
Check: fetch_url

**Scenario**: SC-34b — A save and an order at the same moment end in one of the two allowed ways
**Given**: A valid `SESSION`; `T` live, stock 9, size M, read with its `ETag`; the save and the order sent so they overlap, repeated a number of times with fresh bikes.
**When**: `PUT` stock 5 and `ORDER(T, 1)` are sent at the same instant; then `T` is read.
**Then**: Every run ends as save 200 + order 201 with `stock` 4, or save 409 + order 201 with `stock` 8; never `stock` 5 with the order accepted.
Traces: SC-34 (ADMIN-008 AC-5)
Check: manual — no instrument sends two requests in the same instant; the sequential halves are SC-34a and SC-33a, and the invariant rests on both writers taking the bike's row lock (the conditional `UPDATE … WHERE version = $2` and the order's `stock >= qty`), which the backend review reads in the PR.

**Scenario**: SC-35 — Deleting is not refused for a change since the screen loaded the bike
**Given**: A valid `SESSION`; `T` read with `ETag` `E1`; then `T` saved by another client (`ETag` now `E2`).
**When**: `DELETE /v1/products/T` with no `If-Match`; then, for a second bike `T2` handled the same way, `DELETE /v1/products/T2` with the stale `If-Match: E1`.
**Then**: Both deletes are 204 (a delete reads no token); both bikes are then 404.
Traces: SC-35 (ADMIN-008 AC-6)
Check: fetch_url

**Scenario**: SC-B1 — A write must be a JSON object with the JSON content type
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with a valid body but `Content-Type: text/plain`; with `Content-Type: application/json; charset=utf-8` and a valid body; with the bodies `[]`, `"text"`, `null`, `{"name":"X"} {"name":"Y"}` (trailing data) and `{bad`.
**Then**: The `text/plain` request is 400 `MALFORMED_REQUEST`; the `charset` request is 201 (deleted afterwards); each of the five other bodies is 400 `MALFORMED_REQUEST` (for a non-object the sentence is "The request must be a JSON object."); nothing is saved.
Traces: contract (§2.5 rule 2, §3.4)
Check: fetch_url

**Scenario**: SC-B2 — A write body over 1 MiB is 413
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `{"name":"X","description":"<1,100,000 characters>"}`; `PUT /v1/products/talon-2` with the same oversize body.
**Then**: Each is 413 with `code` `PAYLOAD_TOO_LARGE`; nothing is saved.
Traces: contract (§3.4, §3.5, §6)
Check: fetch_url

**Scenario**: SC-B3 — Body rules answer in schema order, with the transaction rules last
**Given**: A valid `SESSION`.
**When**: `POST /v1/products` with `{"id":"Bad Id","name":""}`; with `{"name":"X","brand":"Trek","price":-1}`; with `{"name":"X","status":"live","stock":-1}`; with `{"name":"X","category":"unicycle","price":-1}`.
**Then**: Each answers with the first failing check only: 400 naming `id`; 400 naming `brand`; 400 naming `stock` (the live-needs-a-price rule is last); 400 naming `price` (the riding-type check is made after every body rule).
Traces: contract (§2.3 which check answers first; §3.4 order; §3.11)
Check: fetch_url

**Scenario**: SC-B4 — `PUT` and `DELETE` check in the contract's order
**Given**: A valid `SESSION`; `talon-2` and `T` with their `ETag`s; `E_old` a stale token for `T`.
**When**: `PUT /v1/products/no-such-bike` with an invalid body and no `If-Match`; `PUT /v1/products/no-such-bike` with a valid body and no `If-Match`; `PUT /v1/products/T` with `If-Match: E_old` and `category` `unicycle`; `PUT /v1/products/T` with `If-Match: E_old` and a valid body.
**Then**: 400 (body before record); 404 (record before `If-Match`); 400 `VALIDATION_FAILED` naming `category` (riding type before staleness); 409 `STALE_RECORD`.
Traces: contract (§3.5 order of checks)
Check: fetch_url

**Scenario**: SC-B5 — A failed write query is 500 and a refused connection is 503
**Given**: A valid `SESSION`; first the `products` table renamed in the database (restored afterwards); then, separately, the database made unavailable while the service runs.
**When**: With the table renamed, `POST /v1/products` with a valid body and `PUT /v1/products/talon-2`; with the database unavailable, `POST /v1/products` with a valid body.
**Then**: The renamed-table writes are 500 `INTERNAL` "Something went wrong. Please try again." with no SQL or table name in the body; the write with the database unavailable is 503 `UNAVAILABLE` "The service is not ready." (the session lookup is refused); with both put back, `GET /v1/products` is 200 and `total` is 14.
Traces: contract (§2.3 INTERNAL, UNAVAILABLE; §3.4)
Check: fetch_url

---

## C. Frontend — the sign-in screen and the session (SC-36 to SC-45)

**Scenario**: SC-36a — The sign-in screen still reads as designed with the real check in place
**Given**: A fresh browser (no cookie, empty storage) at 1440 × 790.
**When**: Open `/admin/sign-in`.
**Then**: The page shows a loaded riding photograph (`img` with a natural width above 0); a panel with the line "Mountain bikes for passes, trails and rough ground — chosen, built and set up by people who ride them."; the logo as a link to `/`; the `h1` reading "Welcome back" whose first word "Welcome" has a lighter computed `font-weight` than "back"; the line "Sign in to run Đèo Bikes."; the label "Email" with the placeholder "you@example.com"; the label "Password" over an `input[type=password]`; a full-width button "Sign in" (its width equals the form's width); the note "Ask the shop's owner for an account."; and a link "Back to the shop" whose `href` is `/`. Neither "Signed in as" nor "Continue" is on the page.
Traces: SC-36 (ADMIN-002 AC-8)
Check: render_url

**Scenario**: SC-36b — On a phone the photograph and navy panel come first, and the password is hidden as typed
**Given**: A fresh browser at 390 × 844.
**When**: Open `/admin/sign-in`; type `abc123` in "Password".
**Then**: The bottom edge of the box holding the photograph and the navy line is above the top edge of `form`; the panel's computed `background-color` is dark navy (each channel below 80 and blue not lower than red or green); the document is no wider than 390 px (`scrollWidth` ≤ 390); the password input's `type` is `password` and its value is `abc123` (shown as dots, not as text).
Traces: SC-36 (ADMIN-002 AC-8, AC-16)
Check: measure_styles

**Scenario**: SC-37a — Enter in the password field signs in with the right credentials and goes to /admin
**Given**: A fresh browser; the throttle clean.
**When**: Open `/admin/sign-in`; type `owner@deo.test` in "Email" and `<ADMIN_PASSWORD>` in "Password"; press Enter with focus in "Password".
**Then**: The browser's location becomes `/admin` and the overview (title "Overview") is shown; the sidebar's foot shows `owner@deo.test`. F=0.
Traces: SC-37 (ADMIN-002 AC-9, AC-14)
Check: interact_page

**Scenario**: SC-37b — A wrong password never reaches /admin
**Given**: A fresh browser; the throttle with at most 3 earlier failures.
**When**: Open `/admin/sign-in`; type `owner@deo.test` and a wrong password; press Enter; then open `/admin` directly.
**Then**: After Enter the location stays `/admin/sign-in` and no management page is shown; opening `/admin` lands on `/admin/sign-in` again (the demo's "any password" is gone). F=1.
Traces: SC-37 (ADMIN-002 AC-14)
Check: interact_page

**Scenario**: SC-38 — An empty email is refused on the screen and nothing is sent
**Given**: A fresh browser at `/admin/sign-in`; "Email" empty.
**When**: Press "Sign in"; then repeat with the email field holding only spaces.
**Then**: "An email address is needed" is shown directly under the "Email" field (its top edge is at or below the input's bottom edge and above the "Password" label); the location stays `/admin/sign-in`; the page made no request to the sign-in endpoint (the instrument's request list shows none, and the sign-in failure count is unchanged).
Traces: SC-38 (ADMIN-002 AC-10)
Check: interact_page

**Scenario**: SC-39a — A wrong password shows the server's sentence and keeps the typed email
**Given**: A fresh browser at `/admin/sign-in`; the throttle with at most 3 earlier failures.
**When**: Type `owner@deo.test` and a wrong password; press "Sign in".
**Then**: "That email and password do not match." is visible inside the form (its box has a height above 0 and is not hidden); the "Email" field still holds `owner@deo.test`; the location is `/admin/sign-in`. F=1.
Traces: SC-39 (ADMIN-002 AC-11)
Check: interact_page

**Scenario**: SC-39b — A locked address shows the 429 sentence in the same place
**Given**: The sign-in address made locked by five wrong sign-ins sent with `fetch_url` (F=5); a fresh browser at `/admin/sign-in`; the backend restarted afterwards.
**When**: Type `owner@deo.test` and `<ADMIN_PASSWORD>` (the right password); press "Sign in".
**Then**: "Too many attempts. Try again in a few minutes." is visible in the form in the same place as the sentence of SC-39a; the "Email" field still holds the typed email; the location stays `/admin/sign-in`.
Traces: SC-39 (ADMIN-002 AC-12)
Check: interact_page

**Scenario**: SC-40 — With the server out of reach the form says it could not sign in
**Given**: A fresh browser at `/admin/sign-in`, loaded while the backend was up; then the backend stopped (the gateway answers 502 with its own page).
**When**: Type `owner@deo.test` and `<ADMIN_PASSWORD>`; press "Sign in"; then start the backend again.
**Then**: "Could not sign in. Please try again." is visible in the form; the typed email is still in the field; the location is `/admin/sign-in`.
Traces: SC-40 (ADMIN-002 AC-13)
Check: interact_page

**Scenario**: SC-41a — A signed-in owner sees "Signed in as" and "Continue" on the sign-in page
**Given**: A fresh browser at `/admin/sign-in` showing neither "Signed in as" nor "Continue"; then the owner signs in through the form and opens `/admin/sign-in` again.
**When**: Read the page; press "Continue".
**Then**: Under the "Sign in" button (its top edge at or below the button's bottom edge) the page reads "Signed in as owner@deo.test." and has a "Continue" button; pressing it goes to `/admin`.
Traces: SC-41 (ADMIN-002 AC-15)
Check: interact_page

**Scenario**: SC-41b — A browser that holds a valid cookie but no hint is sent to sign-in once and offered "Continue"
**Given**: The owner signed in through the form; then the browser's local storage cleared (the cookie stays).
**When**: Open `/admin`.
**Then**: The browser lands on `/admin/sign-in`, where "Signed in as owner@deo.test." and "Continue" show once the `me` answer has arrived; pressing "Continue" opens `/admin` and the overview is shown. If the instrument cannot clear storage while keeping cookies, report UNVERIFIED naming that missing instrument.
Traces: SC-41 (story decision 2, the stated edge; ADMIN-002 AC-15)
Check: interact_page

**Scenario**: SC-42a — A reload keeps the owner on /admin and never shows the sign-in page
**Given**: The owner signed in through the form and on `/admin`.
**When**: Reload the page; record the page's location and visible text from the first frame the instrument can read until the overview's figures appear; then press the browser's Back button.
**Then**: Every recorded state has the path `/admin` and none contains the text "Welcome back" or the path `/admin/sign-in`; the overview (title "Overview") is shown at the end; Back does not land on `/admin/sign-in` (it never entered the history).
Traces: SC-42 (ADMIN-003 AC-4)
Check: interact_page

**Scenario**: SC-42b — A deep link opens the record page and never the sign-in page
**Given**: The owner signed in through the form.
**When**: Navigate straight to `/admin/products/stance-0` (as if typed in the address bar); record the location and text until the form is drawn; press Back.
**Then**: The record page shows the breadcrumb "Bikes › Stance 0" and the heading "Stance 0"; no recorded state has the path `/admin/sign-in` or the text "Welcome back"; Back does not land on `/admin/sign-in`.
Traces: SC-42 (ADMIN-003 AC-5)
Check: interact_page

**Scenario**: SC-43a — With no session every management address lands on sign-in
**Given**: A fresh browser (no cookie, empty storage).
**When**: Open, one after another, `/admin`, `/admin/products`, `/admin/products/stance-0` and `/admin/kinds`.
**Then**: Each ends at `/admin/sign-in` showing the "Sign in" button; none shows "Overview", "Bikes", the model "Stance 0" or "Riding types" as a page title.
Traces: SC-43 (ADMIN-003 AC-6)
Check: interact_page

**Scenario**: SC-43b — When the `me` check cannot reach the server, /admin lands on sign-in
**Given**: The owner signed in through the form (the browser holds its hint and its cookie); then the backend stopped (the gateway answers 502).
**When**: Open `/admin`; then start the backend again.
**Then**: The browser ends at `/admin/sign-in` showing "Sign in"; no management page content remains on screen.
Traces: SC-43 (ADMIN-003 AC-12, AC-7 of Behaviour: unreachable = nobody signed in)
Check: interact_page

**Scenario**: SC-43c — A refused `me` check (database down) also lands on sign-in
**Given**: The owner signed in through the form; then the database made unavailable (`GET /v1/me` answers 503).
**When**: Open `/admin`; then make the database available again.
**Then**: The browser ends at `/admin/sign-in` showing "Sign in".
Traces: SC-43 (ADMIN-003 AC-12; contract §3.8 errors)
Check: interact_page

**Scenario**: SC-43d — A hint with no valid session behind it is cleared and the owner is sent to sign-in
**Given**: The owner signed in through the form; then that session's row deleted from `sessions` (database arranged).
**When**: Open `/admin`; then open `/admin/products`.
**Then**: Both end at `/admin/sign-in` showing "Sign in"; on the second visit no management page is drawn even for a moment (the hint was cleared by the 401).
Traces: SC-43 (ADMIN-003 AC-6; story decision 2)
Check: interact_page

**Scenario**: SC-44a — "Sign out" ends the session, and Back never shows a management page
**Given**: The owner signed in through the form, on `/admin`; the sidebar's "Sign out" button visible.
**When**: Press "Sign out"; then press the browser's Back button; then open `/admin`.
**Then**: After "Sign out" the location is `/admin/sign-in`; after Back the page shown is the sign-in page (never "Overview" or another management page); opening `/admin` lands on `/admin/sign-in` (the cookie was cleared by the server's answer). The earlier cookie value is 401 on `GET /v1/me` (SC-12a).
Traces: SC-44 (ADMIN-003 AC-7, AC-13)
Check: interact_page

**Scenario**: SC-44b — A sign-out the server does not answer leaves the owner signed in
**Given**: The owner signed in through the form, on `/admin` (loaded while the backend was up); then the backend stopped.
**When**: Press "Sign out"; then start the backend and press "Sign out" again.
**Then**: After the first press the location is still `/admin`, the sidebar still shows `owner@deo.test` and the "Sign out" button (nothing else on screen changes); after the second press the location is `/admin/sign-in`.
Traces: SC-44 (story decision 6/7; contract §3.9)
Check: interact_page

**Scenario**: SC-45a — The sidebar foot shows the owner's initial, name, email and "Sign out"
**Given**: The owner signed in through the form as `owner@deo.test`.
**When**: Read the sidebar foot of `/admin`.
**Then**: It shows "O", "Owner" (the part before "@" with its first letter capitalised) and "owner@deo.test", and a "Sign out" button (an element with the accessible name "Sign out").
Traces: SC-45 (ADMIN-003 AC-10 (derivation))
Check: interact_page

**Scenario**: SC-45b — For `lan.tran@deo.test` the foot shows "L", "Lan Tran" and the email
**Given**: A run whose `ADMIN_EMAIL` is `lan.tran@deo.test`, or the frontend unit test of the name derivation with a faked `me` answer.
**When**: The owner signs in; the foot is read.
**Then**: "L", "Lan Tran", "lan.tran@deo.test" and a "Sign out" button.
Traces: SC-45 (ADMIN-003 AC-10)
Check: manual — the run's account is `owner@deo.test` and no instrument changes it; the evidence is `src/test/session.test.ts` (Test runs `run_tests` for the frontend and reads that test), and SC-45a proves the same derivation on the live account.

**Scenario**: SC-45c — The session module keeps its four names and the gate and layout are unchanged
**Given**: The PR's diff.
**When**: Read `src/lib/session.ts`, `src/App.tsx` and `src/components/admin/DashboardLayout.tsx` in the diff; run the frontend tests.
**Then**: `useSession`, `signIn`, `signOut` and `useRole` are all still exported by `session.ts`; `useRole()` returns the empty text for a signed-in owner (a test asserts it); `App.tsx` and `DashboardLayout.tsx` are absent from the diff.
Traces: SC-45 (ADMIN-003 AC-11; ADMIN-003 Behaviour 5)
Check: manual — no instrument reads source exports or a diff; the evidence is the PR's file list and `run_tests` for the frontend.

---

## D. Frontend — saving, waiting and failing (SC-46 to SC-59)

In this section the owner is signed in through the form, the bikes list is `/admin/products`, and fixture `T` (see §0) exists when a case says so.

**Scenario**: SC-46 — `saveRow`, `deleteRow` and `newId` keep their names and their promises
**Given**: `src/components/admin/store.ts` and its test `src/test/admin.test.ts` with the network faked.
**When**: Import the three names; call `newId()` twice; call `saveRow` and `deleteRow` against a faked success and against a faked refusal carrying a sentence and against a faked network failure.
**Then**: All three names import from `components/admin/store.ts`; the two `newId` texts differ; on success each promise resolves; on a refusal it rejects carrying the server's sentence; on a network failure it rejects carrying "The change could not be saved. Please try again." (save) and "The record could not be deleted. Please try again." (delete); the field-derivation tests in the file still pass.
Traces: SC-46 (ADMIN-013 AC-1)
Check: manual — a module's exports and a promise's rejection are not visible over HTTP or in the DOM; the evidence is the unit test run by `run_tests` for the frontend (the screens' behaviour is proved by SC-47 to SC-59).

**Scenario**: SC-47a — A successful edit closes the dialog and the list shows the change without a reload
**Given**: `T` live at price 25000000; `/admin/products` open (the row of "Test Bike" reads `25.000.000 ₫`).
**When**: Open the row's "Actions" → "Edit"; set "Price" to 24000000; press "Save changes".
**Then**: The dialog is gone (no element titled "Edit Test Bike"); the row of "Test Bike" reads `24.000.000 ₫` with no page reload between the press and the reading; `GET /v1/products/T` (fetch) shows `price` 24000000. `T` is deleted afterwards.
Traces: SC-47 (ADMIN-013 AC-12; ADMIN-007 AC-2)
Check: interact_page

**Scenario**: SC-47b — Stock saved from the record page is shown by the list's edit dialog without a reload
**Given**: `T` at stock 9; `/admin/products` open; then its record page opened by the "Model" link.
**When**: On the "Stock" tab set "Bikes in stock" to 4 and press "Save changes"; press the breadcrumb "Bikes"; open the row's "Actions" → "Edit".
**Then**: "Saved." shows after the save; the edit dialog of "Test Bike" shows "Bikes in stock" 4; no full page load happened between the steps. `T` is deleted afterwards.
Traces: SC-47 (ADMIN-007 AC-5; ADMIN-013 AC-12)
Check: interact_page

**Scenario**: SC-48a — One press of "Add" makes one bike, even when pressed twice quickly
**Given**: `/admin/products` open with 14 rows.
**When**: Press "Add"; type "Double Click Bike" in "Model"; double-click "Add" (two clicks in quick succession).
**Then**: The dialog closes; the list shows exactly one row "Double Click Bike" and 15 rows; `GET /v1/products?q=Double%20Click%20Bike` has `total` 1 (one request was made); if the instrument can read the button right after the first press, it is `disabled` until the answer arrives. The bike is deleted afterwards.
Traces: SC-48 (ADMIN-013 AC-2)
Check: interact_page

**Scenario**: SC-48b — A double press of "Save changes" makes one save and shows no refusal
**Given**: `T` with its record page open; the edit dialog of `T` opened on the list (two sub-runs).
**When**: On the record page change "Price" and double-click "Save changes"; then in the dialog change "Price" and double-click "Save changes".
**Then**: On the record page "Saved." shows and no alert notice shows (a second request with the old token would have been 409); the dialog closes with no alert; after each run `GET /v1/products/T` has an `ETag` one higher than before that run (one save, not two). `T` is deleted afterwards.
Traces: SC-48 (ADMIN-013 AC-2)
Check: interact_page

**Scenario**: SC-48c — A double press of "Delete" makes one delete
**Given**: `T`; `/admin/products` open; the row's "Actions" → "Delete" opened.
**When**: Double-click "Delete" in the question "Delete Test Bike?".
**Then**: The question closes with the row gone and no alert notice shown (a second delete would have been 404); `GET /v1/products/T` is 404.
Traces: SC-48 (ADMIN-013 AC-2)
Check: interact_page

**Scenario**: SC-49 — A stale save in the edit dialog keeps the dialog open with the typed values and the sentence
**Given**: `T` at price 25000000; `/admin/products` open; "Edit" opened on `T` (the dialog read `T` at its token `E1`); then another client saves `T` with price 24000000 (`PUT` with `E1`, so the token is now stale).
**When**: In the open dialog set "Price" to 22000000 and press "Save changes".
**Then**: The dialog stays open; its "Price" field still holds 22000000; an alert notice inside the dialog reads exactly "This record changed while you were editing it — for example an order was placed. Reload and make your change again."; "Save changes" is enabled again; `GET /v1/products/T` shows `price` 24000000 (the other client's change stands). `T` is deleted afterwards.
Traces: SC-49 (ADMIN-013 AC-3; ADMIN-008 AC-7)
Check: interact_page

**Scenario**: SC-50 — The add dialog refused with 400 keeps its values and shows the sentence
**Given**: `/admin/products` open; the add dialog open.
**When**: Type "No Price Live" in "Model", choose "Live" in "Status", leave "Price" empty; press "Add"; then type 1000000 in "Price" and press "Add" again.
**Then**: After the first press the dialog stays open, "Model" still holds "No Price Live", "Status" is still "Live", an alert notice inside it reads exactly "A live bike needs a price.", and `GET /v1/products?q=No%20Price%20Live` has `total` 0; after the second press the dialog closes and the list has the row "No Price Live". The bike is deleted afterwards.
Traces: SC-50 (ADMIN-013 AC-4)
Check: interact_page

**Scenario**: SC-51 — A refused save on the record page keeps the form and shows the sentence above it
**Given**: `T` live with price 25000000; its record page open.
**When**: On "Basics" set "One line under the name" to "Typed line"; clear "Price"; press "Save changes".
**Then**: An alert notice above the form reads exactly "A live bike needs a price."; "Price" is still empty and "One line under the name" still holds "Typed line"; "Saved." is not shown; `GET /v1/products/T` still has `price` 25000000 and no `fact` "Typed line". `T` is deleted afterwards.
Traces: SC-51 (ADMIN-013 AC-5)
Check: interact_page

**Scenario**: SC-52a — A refused delete of a riding type leaves the question open with "Delete" enabled
**Given**: `/admin/kinds` open; `trail` (4 bikes) listed.
**When**: Open `trail`'s "Actions" → "Delete"; press "Delete" in the question "Delete Trail?".
**Then**: The question stays open; an alert notice inside it reads exactly "Trail is still the riding type of 4 bikes. Give those bikes another riding type first."; the "Delete" button is enabled (not `disabled`); "Cancel" closes the question; the row "Trail" is still listed.
Traces: SC-52 (ADMIN-013 AC-6; ADMIN-012 AC-7)
Check: interact_page

**Scenario**: SC-52b — A refused delete on the record page leaves the question open too
**Given**: `T`; its record page open; `T` then deleted by another client (`DELETE /v1/products/T` with the owner's cookie).
**When**: Press "Delete", then "Delete" in the question "Delete Test Bike?".
**Then**: The question stays open; an alert notice inside it reads exactly "That record does not exist."; "Delete" is enabled again; the location is still `/admin/products/<T>`.
Traces: SC-52 (ADMIN-013 AC-6)
Check: interact_page

**Scenario**: SC-53a — With the server out of reach the edit dialog says the change could not be saved
**Given**: `T`; `/admin/products` open; the edit dialog of `T` open with "Price" set to 22000000; then the backend stopped.
**When**: Press "Save changes"; then start the backend.
**Then**: The dialog stays open with "Price" still 22000000 and an alert notice inside it reading exactly "The change could not be saved. Please try again."; "Save changes" is enabled again.
Traces: SC-53 (ADMIN-013 AC-7)
Check: interact_page

**Scenario**: SC-53b — The same failure on the record page shows the sentence above the form
**Given**: `T`; its record page open with "Price" changed to 22000000; then the backend stopped.
**When**: Press "Save changes"; then start the backend.
**Then**: An alert notice above the form reads exactly "The change could not be saved. Please try again."; the form keeps 22000000; "Saved." is not shown.
Traces: SC-53 (ADMIN-013 AC-7)
Check: interact_page

**Scenario**: SC-53c — With the server out of reach the delete question says the record could not be deleted
**Given**: `T`; `/admin/products` open with the question "Delete Test Bike?" open; then the backend stopped.
**When**: Press "Delete"; then start the backend.
**Then**: The question stays open with an alert notice reading exactly "The record could not be deleted. Please try again."; "Delete" is enabled again; `GET /v1/products/T` is 200 (nothing was deleted). `T` is deleted afterwards.
Traces: SC-53 (ADMIN-013 AC-7)
Check: interact_page

**Scenario**: SC-53d — The server's own sentence is shown as it is when the database refuses
**Given**: `T`; its record page open with "Price" changed; then the database made unavailable (the service answers 503 "The service is not ready.").
**When**: Press "Save changes"; then make the database available again.
**Then**: The alert notice above the form reads exactly "The service is not ready." (the envelope's sentence, not the generic one).
Traces: SC-53 (story decision 5; ADMIN-013 AC-4)
Check: interact_page

**Scenario**: SC-54a — A 401 on a save sends the owner to sign-in with the sentence in the form
**Given**: `T`; its record page open with "Price" changed; then the owner's session row deleted from `sessions` (database arranged).
**When**: Press "Save changes".
**Then**: The browser ends at `/admin/sign-in`; the sentence "Your session has ended. Sign in again." is visible in the sign-in form's error place; after signing in again with the right credentials the browser goes to `/admin` and the sentence is gone; opening `/admin/products/<T>` afterwards shows the record (the values typed before are not restored). `T` is deleted afterwards.
Traces: SC-54 (ADMIN-013 AC-8; story decision 6)
Check: interact_page

**Scenario**: SC-54b — A 401 on a delete does the same
**Given**: `T`; `/admin/products` open with the question "Delete Test Bike?" open; then the owner's session row deleted (database arranged).
**When**: Press "Delete".
**Then**: The browser ends at `/admin/sign-in` with "Your session has ended. Sign in again." visible in the form; `GET /v1/products/T` is still 200 (nothing deleted). `T` is deleted afterwards.
Traces: SC-54 (ADMIN-013 AC-8; story decision 6)
Check: interact_page

**Scenario**: SC-55 — After failed saves the lists and the public page show what they showed before
**Given**: `T` live at price 25000000; `/admin/products` showing `25.000.000 ₫` in its row; `/shop/<T>` read first.
**When**: Run SC-51 (a refused save on the record page, typed price removed) and SC-49 (a stale save in the dialog typed with 22000000); then open the list and a visitor opens `/shop/<T>`.
**Then**: The list row does not show 22.000.000 ₫ or an empty price: it shows `25.000.000 ₫`, or `24.000.000 ₫` after the other client's change of SC-49 (never a typed value); the visitor's page shows the server's price and stock with no typed value; the record on the server equals what `GET /v1/products/T` returned before the failed saves, apart from the other client's change. `T` is deleted afterwards.
Traces: SC-55 (ADMIN-013 AC-9; ADMIN-007 AC-7)
Check: interact_page

**Scenario**: SC-56a — A delete confirmed on the record page opens the bikes list
**Given**: `T`; its record page open.
**When**: Press "Delete" then "Delete" in the question.
**Then**: The location becomes `/admin/products`; the list has no row "Test Bike"; `GET /v1/products/T` is 404.
Traces: SC-56 (ADMIN-013 AC-13; ADMIN-011 AC-8)
Check: interact_page

**Scenario**: SC-56b — A delete confirmed from a list closes the question with the row gone
**Given**: `T`; `/admin/products` open with 15 rows.
**When**: Open the row's "Actions" → "Delete"; press "Delete".
**Then**: The question is gone; the table has 14 rows and none reads "Test Bike".
Traces: SC-56 (ADMIN-013 AC-13)
Check: interact_page

**Scenario**: SC-57a — An order placed while the edit dialog is open makes the save refused; reopening reads the new stock
**Given**: `T` live, stock 9, price 25000000; `/admin/products` open; the edit dialog of `T` open (it read stock 9); then `ORDER(T, 2)` accepted so the server holds stock 7 (or the stock set to 7 in the database until the orders story is live).
**When**: Set "Price" to 22000000; press "Save changes"; then press "Cancel"; open "Edit" on `T` again; set "Price" to 22000000; press "Save changes".
**Then**: The first save is refused: the dialog stays open, "Price" still holds 22000000, an alert notice inside it reads exactly "This record changed while you were editing it — for example an order was placed. Reload and make your change again.", and `GET /v1/products/T` still shows `stock` 7 and `price` 25000000; the reopened dialog shows "Bikes in stock" 7; the second save closes the dialog and `GET /v1/products/T` shows `price` 22000000 and `stock` 7. `T` is deleted afterwards.
Traces: SC-57 (ADMIN-008 AC-1, AC-3)
Check: interact_page

**Scenario**: SC-57b — The same on the record page
**Given**: `T` live, stock 9; its record page open (stock 9 read); then `ORDER(T, 2)` accepted so stock is 7.
**When**: Change "Price" and press "Save changes"; then reload the page; change "Price" again and press "Save changes".
**Then**: The first save shows the alert notice above the form with the sentence of SC-57a, the typed price stays, and the server still holds stock 7; after the reload the "Stock" tab shows 7; the second save shows "Saved." and the server holds the new price with stock 7. `T` is deleted afterwards.
Traces: SC-57 (ADMIN-008 AC-1, AC-3)
Check: interact_page

**Scenario**: SC-58a — Editing only the stock from the dialog keeps every hidden field of a seeded bike
**Given**: A seeded bike carrying `_source`, `price_listed` and `_invented` (found in `GET /v1/products?limit=100`; if none carries all three, the one carrying most of them), read as record `R` with its `ETag`; `/admin/products` open.
**When**: Open the bike's "Edit"; change only "Bikes in stock" (to a different whole number); press "Save changes".
**Then**: The dialog closes with no alert notice (the server did not refuse `_source`); `GET /v1/products/<id>` equals `R` key for key except `stock` (the keys `_source`, `price_listed` and `_invented` are present with their old values). Stock is put back afterwards.
Traces: SC-58 (ADMIN-005 AC-3, AC-4; ADMIN-011 AC-4)
Check: interact_page

**Scenario**: SC-58b — The same from the record page
**Given**: The bike of SC-58a read as record `R`; its record page open.
**When**: On "Stock" change "Bikes in stock"; press "Save changes".
**Then**: "Saved." shows and no alert notice shows; `GET /v1/products/<id>` equals `R` key for key except `stock`. Stock is put back afterwards.
Traces: SC-58 (ADMIN-005 AC-3, AC-4; ADMIN-011 AC-4)
Check: interact_page

**Scenario**: SC-59a — When "Edit" cannot read the record because it is gone, the dialog does not open and the server's sentence shows above the table
**Given**: `T`; `/admin/products` open showing its row; `T` then deleted by another client.
**When**: Open the row's "Actions" → "Edit".
**Then**: No dialog opens; an alert notice above the table reads exactly "That record does not exist."; the list stays usable (other rows' "Actions" still open).
Traces: SC-59 (ADMIN-013 AC-7; story decision 3)
Check: interact_page

**Scenario**: SC-59b — When "Edit" gets no usable answer, the dialog does not open and the generic sentence shows
**Given**: `/admin/products` open (loaded while the backend was up); then the backend stopped.
**When**: Open any row's "Actions" → "Edit"; then start the backend.
**Then**: No dialog opens; an alert notice above the table reads exactly "The record could not be opened. Please try again."
Traces: SC-59 (story decision 3)
Check: interact_page

---

## E. Frontend — what each management screen does on live data (SC-60 to SC-83)

The owner is signed in through the form for every case here. The seeded database is in place unless a case arranges otherwise.

**Scenario**: SC-60a — Every management page sits in the frame with its title and its marked entry
**Given**: The owner signed in; a 1440 × 790 window.
**When**: Open `/admin`, `/admin/products`, `/admin/products/stance-0`, `/admin/kinds`.
**Then**: Each page shows a sidebar with the shop's name "Đèo Bikes", the entry "Overview", the group "Shop" with the entries "Bikes" and "Riding types", and at its foot `owner@deo.test` with "Sign out"; the title bar shows the page's title ("Overview", "Bikes", "Stance 0", "Riding types"); the sidebar entry of the current section ("Overview", "Bikes", "Bikes", "Riding types") is marked — its computed `font-weight` or `background-color` differs from the other entries'.
Traces: SC-60 (ADMIN-009 AC-1)
Check: interact_page

**Scenario**: SC-60b — Below 768 px the sidebar is a drawer
**Given**: The owner signed in; a 390 × 844 window.
**When**: Open `/admin`; read the sidebar's state; press its toggle in the title bar.
**Then**: Before the press the entries "Overview", "Bikes" and "Riding types" are not visible (the sidebar is closed); after the press they are visible in a drawer over the page; the page has no horizontal scroll (`scrollWidth` ≤ 390).
Traces: SC-60 (ADMIN-009 AC-2)
Check: interact_page

**Scenario**: SC-60c — The overview's cards open the lists they count
**Given**: The owner on `/admin`.
**When**: Press the card "Bikes", go back and press "Running low", go back and press "Out of stock", go back and press "Riding types".
**Then**: The first three each land on `/admin/products`; the fourth lands on `/admin/kinds`.
Traces: SC-60 (ADMIN-009 AC-4)
Check: interact_page

**Scenario**: SC-61a — On the seeded data the four figures read as specified
**Given**: The seeded database (14 bikes, none in draft, 4 riding types, 2 bikes at stock 0, all bikes under 10).
**When**: Open `/admin`; wait until the records have arrived.
**Then**: The cards read "Bikes" 14 with "0 still in draft"; "Riding types" 4 with "shown on the site"; "Running low" 14 with "under ten in stock"; "Out of stock" 2 with "marked out of stock".
Traces: SC-61 (ADMIN-009 AC-3)
Check: render_url

**Scenario**: SC-61b — The boundary at 10, and a draft bike with low stock
**Given**: Three bikes made by the API with category `hardtail`: `L9` (live, price 1000000, stock 9), `L10` (live, price 1000000, stock 10) and `D2` (draft, stock 2).
**When**: Open `/admin`; wait for the records.
**Then**: "Bikes" shows 17 with "1 still in draft"; "Running low" shows 16 (the 14 seeded, `L9` and `D2`, not `L10`); the "Running low" panel lists `L9` and `D2` and does not list `L10`. The three bikes are deleted afterwards.
Traces: SC-61 (ADMIN-009 AC-5, AC-6)
Check: render_url

**Scenario**: SC-62a — The "Running low" panel's columns, order and danger badge
**Given**: `/admin` open at 1440 px on the seeded data.
**When**: Read the panel "Running low".
**Then**: The columns are "Bike", "Riding type" and "In stock"; the "In stock" figures read top to bottom never increase in the other direction (they are in order, fewest first: the two bikes at 0 come first, each figure in a badge whose computed `background-color` is a red-dominant danger colour, and different from the colour of the badge of a bike with stock above 0).
Traces: SC-62 (ADMIN-009 AC-7)
Check: measure_styles

**Scenario**: SC-62b — Searching by model name ignores capitals, shows "Clear filters" and is in the address
**Given**: `/admin` open on the seeded data.
**When**: Type "Stance" in "Search bikes"; then retype it as "STANCE"; then press "Clear filters".
**Then**: Only the rows whose "Bike" text contains "stance" (ignoring capitals) are listed, the same rows for both spellings; the button "Clear filters" is visible; the address query holds the typed words; after "Clear filters" the panel lists 14 rows again, the button is gone and the address holds no search.
Traces: SC-62 (ADMIN-009 AC-8)
Check: interact_page

**Scenario**: SC-62c — The riding-type select narrows, and "Bike" sorts then reverses
**Given**: `/admin` open on the seeded data.
**When**: Choose "Trail" in "Any riding type"; clear it; press "Bike" once, then a second time.
**Then**: After "Trail" only rows whose "Riding type" reads "Trail" are listed (4 on the seeded data); after one press of "Bike" the rows are in ascending model-name order (ignoring capitals, "Talon 2" before "Talon 10"); after the second press the same rows in the reverse order.
Traces: SC-62 (ADMIN-009 AC-9, AC-10)
Check: interact_page

**Scenario**: SC-62d — The panel's foot: range, rows per page, export and page buttons
**Given**: `/admin` open on the seeded data.
**When**: Read the foot of the "Running low" panel; press "Export these 14".
**Then**: The range reads "1–14 of 14"; a "rows per page" select offers 20, 50 and 100; the button reads "Export these 14"; "Previous" is disabled, a page button "1" and "Next" are present; pressing "Export these 14" downloads a `.csv` file with a header row of "Bike", "Riding type", "In stock" and 14 data rows. If the instrument cannot capture a download, the file half is reported UNVERIFIED naming that missing instrument.
Traces: SC-62 (ADMIN-009 AC-11, AC-12)
Check: interact_page

**Scenario**: SC-63a — With no bike running low the panel says so and draws no table
**Given**: Every bike's stock set to 10 in the database (the original stocks recorded first and restored afterwards).
**When**: Open `/admin` and wait for the records.
**Then**: The panel reads "Nothing is running low." and no `table` is drawn inside it; "Running low" shows 0 with "under ten in stock" (a real 0 after the records arrived).
Traces: SC-63 (ADMIN-009 AC-13)
Check: render_url

**Scenario**: SC-63b — A search that matches no low bike gets the second empty state
**Given**: `/admin` open on the seeded data.
**When**: Type "zzzzqq" in "Search bikes".
**Then**: The panel reads "No bike running low matches. Clear the search or the riding type to see them all." with no table and a "Clear filters" button; pressing it brings the 14 rows back.
Traces: SC-63 (ADMIN-009 AC-14)
Check: interact_page

**Scenario**: SC-64 — A stock saved as 12 from another screen takes the bike out of "Running low" without a reload
**Given**: The owner on `/admin` showing "Running low" 14 (seeded data); the bike `talon-2` (stock 9) is in the panel.
**When**: Without reloading, press the sidebar entry "Bikes", open `talon-2`'s "Actions" → "Edit", set "Bikes in stock" to 12, press "Save changes"; press the sidebar entry "Overview".
**Then**: The overview shows "Running low" 13 and the panel no longer lists "Talon 2"; no page reload happened between the steps. Stock is put back to 9 afterwards.
Traces: SC-64 (ADMIN-009 AC-15; ADMIN-007 AC-6)
Check: interact_page

**Scenario**: SC-65 — While the records are on their way the overview draws no figure and no "nothing" sentence
**Given**: The owner signed in; the overview opened by pressing the sidebar link or by navigating to `/admin` (a cold load).
**When**: Record the overview's text from the first frame the instrument can read until all four cards show numbers.
**Then**: No recorded state before the records arrive shows a number in any card (the figure or the count in its note, such as "0 still in draft") and none shows the sentence "Nothing is running low."; after arrival the cards read as in SC-61a. If no frame of the loading state can be captured (the answers come too fast), report UNVERIFIED naming the missing instrument: a way to hold the API's answer back.
Traces: SC-65 (ADMIN-009 AC-16; ADMIN-013 AC-11)
Check: interact_page

**Scenario**: SC-66a — The bikes list opens with its title, controls and 14 rows
**Given**: The owner on `/admin/products` at 1440 × 790 after the records arrived.
**When**: Read the page.
**Then**: The title is "Bikes"; a primary button "Add" is at the right of the title bar; "Search by model" and a "Status" select with "Live" and "Draft" are above the table; the table has 14 rows and the columns "Model", "Brand", "Riding type", "Model year", "Price" after a photograph column with no heading; the `talon-2` row's price reads `18.990.000 ₫` and its cell has computed `text-align` `right`; the "Riding type" cell of every row reads one of the four names with a capital letter (such as "Trail"), never an id such as "trail".
Traces: SC-66 (ADMIN-010 AC-1)
Check: render_url

**Scenario**: SC-66b — On a phone only the photograph, "Model" and "Brand" show
**Given**: The owner on `/admin/products` at 390 × 844.
**When**: Read the table.
**Then**: The photograph, "Model" and "Brand" cells are visible; the "Riding type", "Model year" and "Price" headings and cells are not (computed `display` `none` or zero width); no horizontal page scroll.
Traces: SC-66 (ADMIN-010 AC-2)
Check: measure_styles

**Scenario**: SC-66c — While the first answer is awaited the table reads "Loading…"
**Given**: The owner opening `/admin/products` by link.
**When**: Record the table's text from the first frame the instrument can read until the rows appear.
**Then**: Before the rows arrive the table area reads "Loading…" (and does not read "Nothing here yet"); then the 14 rows replace it. If no frame of the loading state can be captured, report UNVERIFIED naming the missing instrument: a way to hold the API's answer back.
Traces: SC-66 (ADMIN-010 AC-17; ADMIN-013 AC-6 of Behaviour)
Check: interact_page

**Scenario**: SC-66d — A "Model" link opens the bike's record page
**Given**: The owner on `/admin/products`.
**When**: Press the "Model" link of "Stance 0".
**Then**: The location becomes `/admin/products/stance-0` and the page's heading reads "Stance 0"; the link's `href` ends `/admin/products/stance-0`.
Traces: SC-66 (ADMIN-010 AC-18)
Check: interact_page

**Scenario**: SC-67a — "Model" sorts ascending, descending, then back to the seeded order
**Given**: `/admin/products` open; the rows' model order recorded as the seeded order.
**When**: Press "Model" three times, reading the rows after each press.
**Then**: After the first press the models are in ascending order (ignoring capitals, natural numbers); after the second in the exact reverse; after the third in the recorded seeded order.
Traces: SC-67 (ADMIN-010 AC-3)
Check: interact_page

**Scenario**: SC-67b — "Price" sorts by number
**Given**: `/admin/products` open.
**When**: Press "Price"; read the price column; press it again.
**Then**: After the first press the prices, read as numbers (`9.990.000 ₫` is smaller than `18.990.000 ₫`), rise from top to bottom; after the second they fall.
Traces: SC-67 (ADMIN-010 AC-4)
Check: interact_page

**Scenario**: SC-67c — Search by model and the status filter narrow the list and live in the address
**Given**: `/admin/products` open on the seeded data (no draft bike).
**When**: Type "stance" in "Search by model"; clear it; choose "Draft" in "Status"; then, with a draft bike `D1` made by the API, reopen the list with "Draft" chosen.
**Then**: After "stance" only rows with "stance" in their text are listed, "Clear filters" appears and the address holds "stance"; with "Draft" and no draft bike the panel reads "Nothing matches those filters"; with `D1` the list shows exactly one row, `D1`. `D1` is deleted afterwards.
Traces: SC-67 (ADMIN-010 AC-5, AC-6)
Check: interact_page

**Scenario**: SC-67d — The rows-per-page select puts its value in the address
**Given**: `/admin/products` open with 14 rows.
**When**: Choose 50 in the rows-per-page select.
**Then**: The address holds 50 and the foot still reads "1–14 of 14".
Traces: SC-67 (ADMIN-010 AC-19)
Check: interact_page

**Scenario**: SC-68a — A row's "Actions" menu has "Edit" and, below a rule, "Delete"
**Given**: `/admin/products` open.
**When**: Press a row's "Actions" button (its accessible name is "Actions").
**Then**: A menu shows "Edit" and then "Delete", with a separating rule between them; "Delete" is below "Edit" and its computed text colour is a red-dominant danger colour.
Traces: SC-68 (ADMIN-010 AC-7)
Check: interact_page

**Scenario**: SC-68b — The add dialog shows the visible fields, their labels and hints, and the two buttons
**Given**: `/admin/products` open.
**When**: Press "Add"; read the dialog.
**Then**: The title is "Add to bikes"; the fields appear in the schema's order with their labels — "Photograph", "Model", "Brand", "Riding type", "Model year", "Price", "Price before", "One line under the name", "Status", then "Bikes in stock", the size fields and the specification and wording fields as ADMIN-011 Behaviour 2 lists them — and the hints the schema carries (including "Shown struck through beside the price. Leave empty when nothing was reduced." under "Price before"); "Bikes in stock" has the hint "0 shows 'Out of stock' and turns the button into 'Ask when it is back'." under it; "Price on the maker's page" is not shown; the buttons read "Cancel" and "Add".
Traces: SC-68 (ADMIN-010 AC-8, AC-20)
Check: interact_page

**Scenario**: SC-68c — The edit dialog shows its title, its line, its buttons and the bike's values
**Given**: `/admin/products` open; `GET /v1/products/talon-2` read.
**When**: Open `talon-2`'s "Actions" → "Edit".
**Then**: The title is "Edit Talon 2"; the line "Changes show on the site as soon as they are saved." is shown; the buttons read "Cancel" and "Save changes"; the fields hold the bike's values — "Model" "Talon 2", "Brand" "Giant", "Riding type" the name of its riding type, "Price" 18990000 (or its formatted form), "Bikes in stock" 9 — equal to the record read.
Traces: SC-68 (ADMIN-010 AC-11)
Check: interact_page

**Scenario**: SC-68d — The delete question asks, and "Cancel" deletes nothing
**Given**: `/admin/products` open with 14 rows.
**When**: Open `talon-2`'s "Actions" → "Delete"; read the question; press "Cancel".
**Then**: The question reads "Delete Talon 2?" with the line "This cannot be undone." and the buttons "Cancel" and "Delete"; after "Cancel" it is gone, the list still has 14 rows and `GET /v1/products/talon-2` is 200.
Traces: SC-68 (ADMIN-010 AC-13)
Check: interact_page

**Scenario**: SC-69 — Adding with no model is refused on the screen and nothing is sent
**Given**: `/admin/products` open (14 bikes); the add dialog open with every field empty.
**When**: Press "Add".
**Then**: "Model is needed" shows directly under the "Model" field; the dialog stays open; `GET /v1/products?limit=1` still has `total` 14 (nothing was created).
Traces: SC-69 (ADMIN-010 AC-9)
Check: interact_page

**Scenario**: SC-70 — A bike added in the dialog appears in the list and is found by a visitor
**Given**: `/admin/products` open.
**When**: Press "Add"; type "Test Bike" in "Model"; choose "Giant" in "Brand" and "Trail" in "Riding type"; press "Add"; then `GET /v1/products?q=Test%20Bike` with no cookie.
**Then**: The dialog closes; the list shows a row "Test Bike" with "Giant" and "Trail" (15 rows); the read returns `total` 1 with a record whose `id` matches `^products-[a-z0-9]{6}$` (the screen sent no id), `brand` `Giant` and `category` `trail`. The bike is deleted afterwards.
Traces: SC-70 (ADMIN-010 AC-10)
Check: interact_page

**Scenario**: SC-71 — Saving stock 0 in the dialog makes the visitor's page read "Out of stock"
**Given**: `T` live with stock 9; `/admin/products` open.
**When**: Open `T`'s "Actions" → "Edit"; set "Bikes in stock" to 0; press "Save changes"; then a visitor opens `/shop/<T>`.
**Then**: The dialog closes; the visitor's page shows "Out of stock" and a button reading "Ask when it is back". `T` is deleted afterwards.
Traces: SC-71 (ADMIN-010 AC-12)
Check: render_url

**Scenario**: SC-72 — Confirming "Delete" removes the bike from the list and from the visitor's reach
**Given**: `T`; `/admin/products` open with 15 rows.
**When**: Open `T`'s "Actions" → "Delete"; press "Delete"; then `GET /v1/products/T` with no cookie.
**Then**: The question closes and the list has 14 rows with no "Test Bike"; the read is 404 `NOT_FOUND`.
Traces: SC-72 (ADMIN-010 AC-14)
Check: interact_page

**Scenario**: SC-73a — With no bikes at all the list says "Nothing here yet"
**Given**: The `products` table emptied in the database (its rows copied to a scratch table first, with `seq` and `version`, and put back afterwards).
**When**: Open `/admin/products`.
**Then**: The list reads "Nothing here yet" with a button "Add"; no table rows; no "Nothing matches those filters".
Traces: SC-73 (ADMIN-010 AC-15)
Check: render_url

**Scenario**: SC-73b — A filter that matches nothing says so and offers "Clear filters"
**Given**: `/admin/products` open on the seeded data.
**When**: Type "zzzzqq" in "Search by model"; press "Clear filters".
**Then**: The list reads "Nothing matches those filters", "Clear a filter to see more." and a button "Clear filters"; after pressing it the 14 rows are back and the search box is empty.
Traces: SC-73 (ADMIN-010 AC-16)
Check: interact_page

**Scenario**: SC-74a — The record page shows its breadcrumb, heading and two buttons
**Given**: The owner opens `/admin/products/stance-0`.
**When**: Read the page after the form is drawn.
**Then**: The breadcrumb reads "Bikes › Stance 0" with "Bikes" a link to `/admin/products`; the heading reads "Stance 0"; at the right a "Delete" button and a primary "Save changes" button; pressing the breadcrumb's "Bikes" opens `/admin/products`.
Traces: SC-74 (ADMIN-011 AC-1, AC-11)
Check: interact_page

**Scenario**: SC-74b — The four tabs carry the fields of ADMIN-011 Behaviour 2, each hint under its control
**Given**: `/admin/products/stance-0` open.
**When**: Open each tab in turn and read its fields.
**Then**: "Basics" shows photograph, model, brand, riding type, model year, price, price before, one line under the name, status and the two close-up photographs; "Stock" shows bikes in stock, frame sizes, sizes out of stock, rider height for each size, lead with it, bike of the month; "Specification" shows wheel size, suspension travel, frame, fork, rear shock, motor and battery, drivetrain, brakes, wheels, tyres, weight, colour; "Words" shows what it is for, what it is like to own, rating, number of ratings; every hint sits directly under its control, including "Shown struck through beside the price. Leave empty when nothing was reduced." under "Price before".
Traces: SC-74 (ADMIN-011 AC-2, AC-3)
Check: interact_page

**Scenario**: SC-74c — The hidden fields are not shown
**Given**: `/admin/products/stance-0` open.
**When**: Search all four tabs for the label "Price on the maker's page" and for an invented-fields list.
**Then**: Neither exists on any tab (no label "Price on the maker's page", no field or list named for invented fields).
Traces: SC-74 (ADMIN-011 AC-4)
Check: interact_page

**Scenario**: SC-75 — "Saved." shows after a save and goes away on the next edit
**Given**: `T` at stock 9; its record page open.
**When**: On "Stock" set "Bikes in stock" to 4 and press "Save changes"; then edit any field.
**Then**: "Saved." shows above the form with a dismiss button (accessible name "Dismiss") and `GET /v1/products/T` (no cookie) shows `stock` 4; after the next edit "Saved." is gone. `T` is deleted afterwards.
Traces: SC-75 (ADMIN-011 AC-5, AC-6)
Check: interact_page

**Scenario**: SC-76 — A cleared model is refused on the screen and nothing is sent
**Given**: `T`'s record page open.
**When**: On "Basics" clear "Model"; press "Save changes".
**Then**: "Model is needed" shows under the field; "Saved." is not shown; `GET /v1/products/T` still has `name` "Test Bike" and the `ETag` it had before. `T` is deleted afterwards.
Traces: SC-76 (ADMIN-011 AC-7)
Check: interact_page

**Scenario**: SC-77 — "Delete" confirmed on the record page deletes the bike and opens the list
**Given**: `T`'s record page open.
**When**: Press "Delete"; read the question "Delete Test Bike?" with "This cannot be undone."; press "Delete".
**Then**: The location is `/admin/products`; the list has no "Test Bike"; `GET /v1/products/T` is 404.
Traces: SC-77 (ADMIN-011 AC-8)
Check: interact_page

**Scenario**: SC-78 — A record page for a bike that does not exist says so, after the server has answered
**Given**: The owner opens `/admin/products/no-such-bike`.
**When**: Wait until the server has answered; read the page.
**Then**: The page reads "This record is not here" and "It may have been deleted since the link was made." with a button "Back to the list" and shows no form; pressing the button opens `/admin/products`.
Traces: SC-78 (ADMIN-011 AC-9)
Check: interact_page

**Scenario**: SC-79 — A changed riding type shows in the list and in the shop
**Given**: `T` with riding type "Hardtail"; its record page open.
**When**: On "Basics" choose "Trail" in "Riding type"; press "Save changes"; open `/admin/products`; a visitor opens `/shop/<T>`.
**Then**: The list row of "Test Bike" reads "Trail" in "Riding type"; the visitor's page shows "Trail" as the bike's riding type (breadcrumb or label). `T` is deleted afterwards.
Traces: SC-79 (ADMIN-011 AC-10)
Check: interact_page

**Scenario**: SC-80 — Opening a record by link never flashes "This record is not here"
**Given**: The owner signed in; `/admin/products/stance-0` opened by link (a cold load).
**When**: Record the page's text from the first frame the instrument can read until the form is drawn.
**Then**: No recorded state contains "This record is not here" and no state shows the form before the record has arrived (the model field holds "Stance 0" from the first frame in which the form exists). If no loading frame can be captured, report UNVERIFIED naming the missing instrument: a way to hold the API's answer back.
Traces: SC-80 (ADMIN-011 AC-12; ADMIN-013 AC-10)
Check: interact_page

**Scenario**: SC-81a — The riding types list opens with its title, controls and 4 rows
**Given**: The owner on `/admin/kinds`.
**When**: Read the page; press a row's "Actions".
**Then**: The title is "Riding types"; a button "Add" and a search box are shown; no "Status" select; the table has 4 rows with the columns "Name", "Who it is for" and "Where it rides best"; no cell holds a link to a record page (the row is not a link); the "Actions" menu has "Edit" and "Delete".
Traces: SC-81 (ADMIN-012 AC-1, AC-2)
Check: interact_page

**Scenario**: SC-81b — The add dialog for a riding type has three fields, and an empty name is refused
**Given**: `/admin/kinds` open.
**When**: Press "Add"; read the dialog; press "Add" with "Name" empty.
**Then**: The title is "Add to riding types" with the three fields "Name", "Who it is for" and "Where it rides best" and the buttons "Cancel" and "Add"; after the press "Name is needed" shows under "Name"; `GET /v1/categories` still has `total` 4.
Traces: SC-81 (ADMIN-012 AC-3, AC-4)
Check: interact_page

**Scenario**: SC-82a — A riding type added in the dialog is listed and saved
**Given**: `/admin/kinds` open.
**When**: Press "Add"; type "Gravel" in "Name"; press "Add".
**Then**: The dialog closes; the list shows a row "Gravel" (5 rows); `GET /v1/categories?q=Gravel` has `total` 1.
Traces: SC-82 (ADMIN-012 AC-5)
Check: interact_page

**Scenario**: SC-82b — The new riding type shows on the shop as a chip, a footer link and a "where to start" row
**Given**: The riding type "Gravel" of SC-82a exists.
**When**: A visitor opens `/shop` and `/`.
**Then**: `/shop` shows a chip "Gravel" and the footer has a link "Gravel"; the "Not sure where to start?" band on `/` has a row for "Gravel" (five rows). "Gravel" is deleted afterwards.
Traces: SC-82 (ADMIN-012 AC-5)
Check: render_url

**Scenario**: SC-82c — A renamed riding type shows its new name in the bikes list
**Given**: `trail` read with its `ETag`; `/admin/kinds` open.
**When**: Open `trail`'s "Actions" → "Edit"; read the title "Edit Trail" and the line "Changes show on the site as soon as they are saved."; set "Name" to "Cross country"; press "Save changes"; open `/admin/products`.
**Then**: The dialog closes with the row now reading "Cross country"; the bikes list's "Riding type" column reads "Cross country" for the four bikes that were "Trail". The name is put back to "Trail" afterwards.
Traces: SC-82 (ADMIN-012 AC-6)
Check: interact_page

**Scenario**: SC-82d — The renamed riding type shows on the shop's chips, breadcrumbs and cards
**Given**: `trail` renamed to "Cross country" (as in SC-82c).
**When**: A visitor opens `/shop` and the page of a bike whose riding type is `trail`.
**Then**: The chip reads "Cross country" (no chip "Trail"); the bike's breadcrumb or riding-type label reads "Cross country"; the shop cards for those bikes show "Cross country". The name is put back afterwards.
Traces: SC-82 (ADMIN-012 AC-6)
Check: render_url

**Scenario**: SC-83a — Deleting a riding type that bikes use leaves the question open with the sentence
**Given**: `/admin/kinds` open; `trail` is the riding type of 4 bikes.
**When**: Open `trail`'s "Actions" → "Delete"; press "Delete".
**Then**: The question stays open and shows exactly "Trail is still the riding type of 4 bikes. Give those bikes another riding type first."; after "Cancel" the row "Trail" is still listed.
Traces: SC-83 (ADMIN-012 AC-7)
Check: interact_page

**Scenario**: SC-83b — Deleting an unused riding type removes it from the list and from the shop
**Given**: A riding type "Gravel" made by the API that no bike uses; `/admin/kinds` open showing it.
**When**: Open its "Actions" → "Delete"; press "Delete"; then a visitor opens `/shop`.
**Then**: The question closes and the row "Gravel" is gone (4 rows); `/shop` has no chip "Gravel".
Traces: SC-83 (ADMIN-012 AC-8)
Check: interact_page

---

## F. Checks on the PR itself

**Scenario**: SC-F1 — The backend's own unit tests run without a database and pass
**Given**: The PR's branch.
**When**: Run the backend tests (no database, as in CI).
**Then**: They pass, and the suite contains tests named for the validators (each rule of §3.11 at its limit), the throttle with an injected clock (5 failures, 15 minutes, reset on success), the password hash check, the cookie attributes (`HttpOnly`, `SameSite=Lax`, `Max-Age=604800`, `Secure` only over HTTPS), and the 405-before-401 ordering.
Traces: story §3 (Backend: Go unit tests); SC-4, SC-5, SC-8, SC-9, SC-14, SC-18 to SC-24
Check: manual — no HTTP or DOM instrument runs a test suite; the evidence is `run_tests` for the backend, read for pass/fail and for the names above.

**Scenario**: SC-F2 — The frontend's tests, lint and build pass, and nothing outside the story's files changed
**Given**: The PR's branch.
**When**: Run the frontend tests; read the PR's file list against the story's §9.
**Then**: `npm test`, lint and the type-check build pass; `src/test/admin.test.ts` keeps its field-derivation tests and tests the store against a faked network; `src/test/session.test.ts` exists; the changed files are only those of §9 (`session.ts`, `SignInPage.tsx`, `store.ts`, `ListPage.tsx`, `RecordPage.tsx`, the `Confirm` component in `components/admin/index.tsx`, `AdminOverview.tsx`, the two tests); `App.tsx`, `DashboardLayout.tsx`, `AdminProducts.tsx`, `AdminProduct.tsx`, `AdminKinds.tsx`, `AdminSignIn.tsx`, `src/data/index.ts`, `theme.css` and `index.css` are not in the diff.
Traces: story §3 (Frontend), §4, §9
Check: manual — a file list and a test run are not observable over HTTP; the evidence is `run_tests` for the frontend and the PR's diff.

---

## G. Coverage — every acceptance criterion of the story has a case

| Criterion | Cases | Criterion | Cases |
|---|---|---|---|
| SC-1 | SC-1a, SC-1b | SC-43 | SC-43a, SC-43b, SC-43c, SC-43d |
| SC-2 | SC-2a, SC-2b | SC-44 | SC-44a, SC-44b |
| SC-3 | SC-3a, SC-3b | SC-45 | SC-45a, SC-45b, SC-45c |
| SC-4 | SC-4 | SC-46 | SC-46 |
| SC-5 | SC-5 | SC-47 | SC-47a, SC-47b |
| SC-6 | SC-6a, SC-6b | SC-48 | SC-48a, SC-48b, SC-48c |
| SC-7 | SC-7a, SC-7b | SC-49 | SC-49 |
| SC-8 | SC-8a, SC-8b | SC-50 | SC-50 |
| SC-9 | SC-9 | SC-51 | SC-51 |
| SC-10 | SC-10a, SC-10b | SC-52 | SC-52a, SC-52b |
| SC-11 | SC-11a, SC-11b, SC-11c, SC-11d | SC-53 | SC-53a, SC-53b, SC-53c, SC-53d |
| SC-12 | SC-12a, SC-12b | SC-54 | SC-54a, SC-54b |
| SC-13 | SC-13a, SC-13b, SC-13c, SC-13d | SC-55 | SC-55 |
| SC-14 | SC-14a, SC-14b | SC-56 | SC-56a, SC-56b |
| SC-15 | SC-15a, SC-15b | SC-57 | SC-57a, SC-57b |
| SC-16 | SC-16a, SC-16b, SC-16c | SC-58 | SC-58a, SC-58b |
| SC-17 | SC-17a, SC-17b, SC-17c | SC-59 | SC-59a, SC-59b |
| SC-18 | SC-18a, SC-18b, SC-18c | SC-60 | SC-60a, SC-60b, SC-60c |
| SC-19 | SC-19a, SC-19b, SC-19c | SC-61 | SC-61a, SC-61b |
| SC-20 | SC-20a, SC-20b | SC-62 | SC-62a, SC-62b, SC-62c, SC-62d |
| SC-21 | SC-21a, SC-21b | SC-63 | SC-63a, SC-63b |
| SC-22 | SC-22a, SC-22b | SC-64 | SC-64 |
| SC-23 | SC-23a, SC-23b | SC-65 | SC-65 |
| SC-24 | SC-24a | SC-66 | SC-66a, SC-66b, SC-66c, SC-66d |
| SC-25 | SC-25 | SC-67 | SC-67a, SC-67b, SC-67c, SC-67d |
| SC-26 | SC-26a, SC-26b, SC-26c | SC-68 | SC-68a, SC-68b, SC-68c, SC-68d |
| SC-27 | SC-27a, SC-27b, SC-27c | SC-69 | SC-69 |
| SC-28 | SC-28a, SC-28b | SC-70 | SC-70 |
| SC-29 | SC-29a, SC-29b | SC-71 | SC-71 |
| SC-30 | SC-30a, SC-30b | SC-72 | SC-72 |
| SC-31 | SC-31a, SC-31b | SC-73 | SC-73a, SC-73b |
| SC-32 | SC-32a, SC-32b | SC-74 | SC-74a, SC-74b, SC-74c |
| SC-33 | SC-33a, SC-33b, SC-33c, SC-33d | SC-75 | SC-75 |
| SC-34 | SC-34a, SC-34b | SC-76 | SC-76 |
| SC-35 | SC-35 | SC-77 | SC-77 |
| SC-36 | SC-36a, SC-36b | SC-78 | SC-78 |
| SC-37 | SC-37a, SC-37b | SC-79 | SC-79 |
| SC-38 | SC-38 | SC-80 | SC-80 |
| SC-39 | SC-39a, SC-39b | SC-81 | SC-81a, SC-81b |
| SC-40 | SC-40 | SC-82 | SC-82a, SC-82b, SC-82c, SC-82d |
| SC-41 | SC-41a, SC-41b | SC-83 | SC-83a, SC-83b |
| SC-42 | SC-42a, SC-42b | | |

Cases that trace to the contract, not to a numbered criterion: SC-6b, SC-7b, SC-15b, SC-A1 to SC-A4, SC-B1 to SC-B5 (error envelope, route shapes and methods, content type and size, the order in which checks answer, 500 for a failed query and 503 for a refused connection). Cases that trace to the story's scope: SC-F1, SC-F2.

**Failures the criteria name, each with its own case where the failure happens:** a wrong password and a wrong email (SC-6a), a malformed sign-in (SC-7a, SC-7b), the lock (SC-8a), an expired, tampered, mismatched and signed-out session (SC-11a to SC-11d, SC-12a, SC-13c), no session on every write (SC-13a to SC-13d), read-only collections (SC-14a, SC-14b), every refused body (SC-18 to SC-24, SC-B1 to SC-B3), a riding type in use (SC-27a, SC-27b, SC-52a, SC-83a), a stale save and a missing `If-Match` (SC-33a, SC-33c, SC-33d, SC-49, SC-57a, SC-57b), a refused connection and a failed query (SC-A3, SC-A4, SC-B5, SC-53d), a server out of reach on the screens (SC-40, SC-43b, SC-44b, SC-53a to SC-53c, SC-59b), a session that ended during a save or a delete (SC-54a, SC-54b), and a record deleted in another window (SC-52b, SC-59a).

## H. Cases that are `manual`, and why

| Case | Why no instrument answers it | Evidence Test collects instead |
|---|---|---|
| SC-2b | No instrument reads a source tree, a server log or a table's contents over HTTP | Repository grep for the value, `.env.example`, the PR diff; the database tool for `sessions` if it returns rows |
| SC-3b | No instrument reads the service's log | The container log if readable; otherwise the PR's unit test of the not-configured path |
| SC-8b | No instrument advances the server's clock 15 minutes | The Go unit test of the throttle with the injected clock (`run_tests`, backend) |
| SC-11d | No instrument changes the container's secrets | The Go unit test of the credential fingerprint and the session lookup (`run_tests`, backend) |
| SC-29b | The API never serves an order (ORD-001.5), so order lines cannot be read over HTTP | The database tool on `order_lines` if it returns rows; the migration (no foreign key to `products`) |
| SC-34b | No instrument sends two requests in the same instant | SC-34a and SC-33a (the two orders of arrival, one after the other); the backend review reads the row lock |
| SC-45b | The run's account is `owner@deo.test`; no instrument changes it | `src/test/session.test.ts` through `run_tests` (frontend); SC-45a proves the same rule live |
| SC-45c | No instrument reads source exports or a diff | The PR's file list and the frontend tests |
| SC-46 | A module's exports and a promise's rejection are not visible in the DOM or over HTTP | `src/test/admin.test.ts` through `run_tests` (frontend); SC-47 to SC-59 prove the screens |
| SC-F1, SC-F2 | A test run and a file list are not observable over HTTP | `run_tests` for both halves and the PR's diff |

That is 11 of 182 cases. Cases whose Given needs a state some runs cannot make say what to report as UNVERIFIED and which instrument was missing (SC-3a, SC-41b, SC-62d, SC-65, SC-66c, SC-80).

## I. Findings against the requirements and the plan (reported, not silently resolved)

1. **The "session ended" sentence versus the existing sign-in form (SC-54).** Story decision 6 says "Your session has ended. Sign in again." stays in the form's error place "until the next sign-in attempt". The form as it exists clears its error whenever the email field changes (`setError("")` in the email `onChange`). The cases assert the sentence on arrival and its disappearance after a new sign-in (SC-54a) and do not assert that typing keeps it. TL should say which is meant; if it must survive typing, the cases gain a step.
2. **Three criteria are about a state that lasts milliseconds (SC-65, SC-66c, SC-80).** "At no moment before the records arrive does a figure read 0 / does 'This record is not here' show" can only be observed if the instrument reads a frame during the wait. If the answers come back faster than the first readable frame, these report UNVERIFIED with the missing instrument named (a way to hold the API's answer back), and the PR diff (the `loading` branch of `AdminOverview`, `ListPage`, `RecordPage`) is the fallback evidence.
3. **Cases that need the backend stopped or the database made unavailable between two steps of one browser session** (SC-40, SC-43b, SC-43c, SC-44b, SC-53a to SC-53d, SC-59b) need the instrument to change the server between page loads without losing the page. If it cannot, they report UNVERIFIED naming that.
4. **The orders story may not be live when this one is verified.** SC-29a, SC-33, SC-34a, SC-57 and SC-61b's stock changes then use the database arrangement the story's §8 allows (a direct update of `stock`, which the `touch_record` trigger turns into a new `ETag`); SC-34b needs the real order and stays manual either way.
5. **The sign-in throttle is shared by every sign-in attempt from the verifier's address.** The cases carry their failure counts (`F=n`) and the restart points are in §0; a verification pass that ignores them will see a spurious 429.
6. **Routes.** The plan item's `/v1/auth/*` paths are replaced by the contract's `POST /v1/sessions`, `GET /v1/me` and `DELETE /v1/sessions/current` (story decision 1); these cases use the contract's.
7. **The password value is a placeholder in this file on purpose** (`<ADMIN_PASSWORD>`): writing the run's value here would break SC-2 itself.
