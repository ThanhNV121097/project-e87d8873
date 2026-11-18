# Service contract extension — Site data from the database

Story: `docs/data/stories/site-data-from-the-database.md` · Requirements: DATA-001 to DATA-008 (`docs/data/SRS.md`) · Plan item: *Site data from the database* (feature, P1) · Companion: `docs/architecture/erd/site-data-from-the-database.md`

## 1. Result

**This story adds no endpoint, no request or response field, no error code and no pagination or auth rule.** It builds the two reads that `docs/architecture/services.md` already specifies — §3.2 `GET /v1/{collection}` and §3.3 `GET /v1/{collection}/{id}` — with every parameter, answer and error of §2.3, §2.4 and §2.7, and the frontend module that speaks them. Nothing of that is restated here.

There is no UI pull request and no mock module for this story: the "mock" is the eight JSON files and the in-memory filter of `src/data/index.ts`. Section 7.6 reads that module as unreviewed input, says where the API departs from it and what the frontend does about it.

What this file adds, because the contract leaves it to the builder:

| § | What |
|---|---|
| 2 | The routes, and what the other methods answer in this story |
| 3 | Reading the query, step by step: the order of the checks, how a name is classified, the sentence and `field` of every 400 |
| 4 | The two statements: how every filter, the search, the sort and the page become SQL with bound values only |
| 5 | Failures, headers, the request id and the `httpx` changes this story makes for every story after it |
| 6 | Package layout and what is tested without a database |
| 7 | How the frontend module speaks the contract: request terms, state, the refresh signal, shared in-flight requests, its differences from today's module, its tests |
| 8 | Conflicts with merged decisions, resolved |
| 9 | Rollout, dependencies, open questions |

## 2. Routes

In `internal/server.Router`, two lines beside the health route (no method in the pattern, so the handler, not the mux, decides what a method gets):

```go
mux.HandleFunc("/v1/{collection}", collections.Collection(db))
mux.HandleFunc("/v1/{collection}/{id}", collections.Record(db))
```

The `DB` interface does **not** grow: both reads are one statement answered by one row (§4), which `QueryRow` already provides.

| Request | Answer |
|---|---|
| `GET /v1/{collection}` / `GET /v1/{collection}/{id}` | §3.2 / §3.3 of the contract |
| Any other method (including `HEAD`, `OPTIONS`) on either shape, collection known | `405 METHOD_NOT_ALLOWED`, `Allow: GET`, "That method is not offered here." In the final contract `Allow` also lists `POST` (collection) and `PUT, DELETE` (record); the admin story widens it |
| Unknown collection, any method | `404 NOT_FOUND` "That collection does not exist." The collection is checked **before** the method. This differs from the contract's final order only for a method the contract never offers (a `PATCH` on an unknown collection is 404 here and 405 there); nothing depends on it |
| `/v1/products/` (trailing slash), `/v1/products/talon-2/extra`, `/v1/` | `404 NOT_FOUND` "That address does not exist." by the catch-all: Go's mux does not let `{id}` match an empty trailing segment, so no handler code is needed |
| `/v1/PRODUCTS` | `404` "That collection does not exist." (names are exact) |
| `/v1/orders`, `/v1/me`, `/v1/sessions/current` | unknown collection, `404`, until the stories that own those routes register their literal patterns (a literal wins over `{collection}`) |

## 3. Reading the query

One function, `ParseQuery(collection, url.Values) (Query, *QueryError)`, takes the whole query string and returns a typed value or the first refusal. Everything after it trusts the typed value. It runs these checks **in this order**; the first failure answers:

1. **A repeated parameter**: names in alphabetical order, the first with more than one value → 400, `field` = that name.
2. **`page`**, then **`limit`**. Absent: 1 and 12. Present: the whole value must be decimal digits (`^[0-9]+$`; empty, signs, spaces, `1.0`, `abc` all fail) with `page` between 1 and 1,000,000,000 and `limit` at least 1. A `limit` above 100 becomes 100 and is **echoed as 100**. Digit strings too long for an `int64` fail as out of range.
3. **`sort`**: trimmed; empty is "none". A leading `-` is descending. The rest must be `id` or a schema key of the collection, and not a list.
4. **`search`**: split on `,`, parts trimmed, empty parts dropped; none left is "default". Each must be `id` or a schema key, and text-like. `_source` is not a schema key.
5. **`q`**: trimmed (Unicode white space); empty means no search. Longer than 100 code points after trimming → 400. A `q` that folds to nothing at all (a lone combining mark) matches every record that has a searched value; accepted.
6. **Every other parameter, alphabetical.** Classify the *name*, whatever the value, **even an empty one** (`colourful=` is 400; an empty value only means "no condition" for a name that is a field):
   1. the name is `id` or a schema key → an equality condition;
   2. else it ends in `_from` or `_to` and the rest is `id` or a schema key → a range on that field. The field must be text-like or numeric; on a boolean or a list → 400;
   3. else → 400 "no such field". `_source`, `seq`, `version`, `created_at`, `updated_at` are not schema keys. `_invented` is (a list: equality only);
   4. an equality value is split on `,`, empty parts dropped (none left: no condition); a range value is **not** split and an empty one is no condition. Numeric values (equality parts and range ends) must match `^-?[0-9]+(\.[0-9]+)?$`; a boolean's parts must each be exactly `true` or `false`.

`q` on `faqs` is **always** the search words (SC-32); the question's text cannot be filtered on by equality, but `q_from`/`q_to` and `sort=q`, `search=q` work, because those names are not `q`.

**The 400 answers.** `code` is `INVALID_QUERY`; `field` and the sentence are fixed here so that two builders and a test agree. `<noun>` is: products → "bikes", categories → "riding types", reviews → "reviews", gallery → "photographs", team → "team members", services → "workshop jobs", faqs → "questions", shop → "shop records". Tests match the status, the `code`, `field` and the quoted name, never the rest of the wording.

| Trigger | `field` | `error` |
|---|---|---|
| a repeated name `x` | `x` | The parameter “x” was given more than once. Use a comma to give several values. |
| `page` | `page` | The page must be a whole number from 1 to 1000000000. |
| `limit` | `limit` | The limit must be a whole number of 1 or more. |
| `sort` names no field | the name (for `-` alone: `sort`) | There is no field “colourful” to sort bikes by. |
| `sort` names a list | the name | “sizes” is a list and cannot be sorted. |
| `search` names no field | the name | There is no field “nothing” to search questions by. |
| `search` names a number, price, boolean or list | the name | “price” cannot be searched. |
| `q` too long | `q` | The search words must be at most 100 characters. |
| an unknown name `x` (also `x_from`, `x_to` of no field) | the name as given | There is no field “colourful” to filter bikes by. |
| a non-number on a numeric field | the name as given (`stock`, `price_from`) | The value of “stock” must be a number. |
| anything but `true`/`false` on a boolean | the name | The value of “deal” must be true or false. |
| a range on a boolean or list (`deal_from`) | the name as given | “deal_from” cannot be used: a range needs a number, a price or text. |

An unknown collection is not a query error (§2); `GET /v1/nothing?limit=0` is `404`.

**What each kind means** is the table in `services.md` §3.2 and is not repeated. The one thing it leaves open: **text-like** is `id` and every field of type `text`, `longtext`, `image`, `url` and `ref:*`; **numeric** is `number` and `price`; the default `search` set is `id` followed by every text-like schema key in schema order — for `products` that includes the three image paths, `category` and `status`, exactly the fields the current module searches (it searches every string-valued key that does not begin with an underscore).

## 4. The two statements

Built by `BuildList(collection, Query) (sql string, args []any)` and `BuildRecord(collection) string`, from **fixed fragments and a fixed field table only**. A column name is written into the SQL from the table (`"price"`, quoted), never from the request; a value is written only as a placeholder. Every placeholder carries an explicit cast, so a prepared statement never has to guess a type.

### 4.1 The list: one statement, one row, one snapshot

```sql
WITH m AS MATERIALIZED (
  SELECT t."seq", t."id", t."name", t."brand", …          -- the collection's answer columns, schema order, then "_source"
  FROM "products" AS t
  WHERE <condition> AND <condition> AND …                  -- the whole clause is omitted when there is none
)
SELECT
  (SELECT count(*) FROM m) AS total,
  COALESCE((
    SELECT jsonb_agg(p.rec ORDER BY p.rn)
    FROM (
      SELECT jsonb_strip_nulls(to_jsonb(m) - 'seq') AS rec,
             row_number() OVER (ORDER BY <order>) AS rn
      FROM m
    ) AS p
    WHERE p.rn > $a::bigint AND p.rn <= $b::bigint
  ), '[]'::jsonb) AS items
```

It answers **exactly one row** even when the page is empty (`total` 14, `items` `[]` for page 9: SC-12), which is why the page is cut with `row_number()` and not `OFFSET`/`LIMIT`, and why there is no second statement to disagree with the first. `total` is `bigint`; `items` is `jsonb`. The handler scans `total` into an `int64` and `items` into a `[]byte`, and writes `{"items": <bytes as raw JSON>, "total": …, "page": …, "limit": …}` (`json.RawMessage`; the numbers are never decoded into floats). `$a` is `(page−1)·limit`, `$b` is `$a + limit`, both computed in Go from validated integers (at most 10^9 · 100).

**Conditions** (`t.` is the CTE's alias; `<c>` is `t."col"`; `$k` is the next placeholder; `v` is the unnest alias). Parameters are bound as `[]string` (equality), `string` (range), `[]bool` (boolean).

| Kind | Equality `field=a,b` | Range `field_from` / `field_to` |
|---|---|---|
| `id` and `ref:*` | `<c> = ANY (ARRAY(SELECT fold_text(v) FROM unnest($k::text[]) AS v))` (clarification C1 of the ERD extension: fold the values only, so the index serves it) | `<c> COLLATE "C" >= $k::text`, `<=` for `_to` |
| other text-like | `fold_text(<c>) = ANY (ARRAY(SELECT fold_text(v) FROM unnest($k::text[]) AS v))` | as above (the stored text, byte order: how ISO dates compare, SC-25) |
| numeric | `<c> = ANY ($k::text[]::numeric[])` | `<c> >= $k::text::numeric`, `<=` for `_to` |
| boolean | `<c> = ANY ($k::boolean[])` | refused in §3 |
| list | `EXISTS (SELECT 1 FROM unnest(<c>) AS e WHERE fold_text(e) = ANY (ARRAY(SELECT fold_text(v) FROM unnest($k::text[]) AS v)))` | refused in §3 |

A numeric value goes in as **text** and is cast in SQL, never parsed to a float in Go: `18990000` and `18990000.5` mean what they say, and `price_to=18990000` includes `talon-2` (SC-24). A `NULL` column makes every comparison above `NULL`, so a record with no value matches no condition on it (SC-33); `unnest(NULL)` is no rows, so the same holds for a list.

**Search** is one more condition, ANDed last: `(strpos(fold_text(t."id"), fold_text($k::text)) > 0 OR strpos(fold_text(t."name"), fold_text($k::text)) > 0 OR …)` over the searched fields, with **one** placeholder reused (clarification C2: `strpos` has no wildcards, so `%`, `_` and `\` mean themselves and nothing is escaped, SC-38). The value bound is the trimmed `q` as the visitor typed it; the folding is the database's. A `NULL` field is `NULL`, which `OR` ignores when another field matches.

**Order** (inside the window, so it uses the CTE's alias `m`), always ending in `m."seq"`:

| `sort` | `ORDER BY` |
|---|---|
| none | `m."seq"` |
| numeric or boolean field | `m."price" ASC NULLS LAST, m."seq"` (`DESC NULLS LAST` for `-price`) |
| text-like field (`id` included) | `natural_key(m."name") COLLATE "C" ASC NULLS LAST, m."seq"` (`DESC` for `-name`) |

`natural_key` is `NULL` for a `NULL` field, so a record with no value sorts last in both directions (SC-48), and equal values keep `seq` ascending **in both directions** (the current module reverses ties on a descending sort; see §7.6).

**Order of the conditions in the text** is deterministic so that unit tests can compare statements: the field conditions in the alphabetical order of their parameter names (§3 step 6), then the search; placeholders are numbered in that order, then `$a`, `$b`.

**Worked example.** `GET /v1/products?status=live&category=Trail,e-mtb&price_to=70000000&q=Stance&sort=-price&page=2&limit=5`:

| Placeholder | Value | Clause |
|---|---|---|
| `$1` | `[]string{"Trail","e-mtb"}` | `t."category" = ANY (ARRAY(SELECT fold_text(v) FROM unnest($1::text[]) AS v))` |
| `$2` | `"70000000"` | `t."price" <= $2::text::numeric` |
| `$3` | `[]string{"live"}` | `fold_text(t."status") = ANY (ARRAY(SELECT fold_text(v) FROM unnest($3::text[]) AS v))` |
| `$4` | `"Stance"` | the search clause over `id`, `name`, `brand`, `category`, `price_listed`, `fact`, `status`, `wheel`, `travel`, `frame`, `fork`, `shock`, `motor`, `drivetrain`, `brakes`, `wheels`, `tyres`, `weight`, `colour`, `description`, `ride`, `image`, `image_back`, `image_detail` |
| `$5`, `$6` | `5`, `10` | `p.rn > $5::bigint AND p.rn <= $6::bigint`; order `m."price" DESC NULLS LAST, m."seq"` |

`q='; drop table products; --` is `$4`'s value and is nowhere in the statement text (SC-44).

### 4.2 One record

```sql
SELECT jsonb_strip_nulls(to_jsonb(r) - 'version') AS rec, r."version"      -- products and categories
FROM (SELECT t."id", t."name", …, t."version" FROM "products" AS t WHERE t."id" = $1::text) AS r
```

For the six read-only collections the `version` column is neither selected nor scanned and the answer is `rec` alone. The match is **exact and case-sensitive** (`Stance-0` is not `stance-0`: 404, `services.md` §3.3), unlike the list's `id=` condition, which ignores capitals. `pgx.ErrNoRows` is the 404 "That record does not exist." A hit is `200` with the record and, for `products` and `categories` only, `ETag: "<version>"`, set before the body is written. The record is the same JSON the list holds for that id (SC-51), because both come from the same column list and the same `jsonb_strip_nulls` projection.

## 5. Failures, headers, and what this story changes for every story after it

### 5.1 Which failure is which

Both reads run under `context.WithTimeout(r.Context(), 5*time.Second)` (`services.md` §5, PostgreSQL row). A failed statement is answered by **one** predicate, `unavailable(err)`:

| The error is | Answer |
|---|---|
| a refused connection: `syscall.ECONNREFUSED` anywhere in the chain, or a dial `*net.OpError`; or a `*pgconn.PgError` whose SQLSTATE is in class `08` or is `57P01`, `57P02`, `57P03` (the server is going away or not yet accepting) | `503 UNAVAILABLE`, `code` `UNAVAILABLE`, "The service is not ready." |
| anything else: a failed statement, a deadline, a cancelled request, a scan error | `500 INTERNAL`, `code` `INTERNAL`, "Something went wrong. Please try again." |

The body is the envelope and nothing else: no `items`, no record, no SQL, host, path or constraint name (SC-19). The real error goes to the log at `Error` level with `collection`, the operation (`list` or `record`) and `request_id` — **never the query string and never `q`**, because a visitor's search words are private (`services.md` §3.2). A request the visitor abandoned (`context.Canceled`) is logged at `Info`, not `Error`.

A deadline and a request waiting for a free pool connection both arrive as `context.DeadlineExceeded` and cannot be told apart, so both are `500` (conflict 2 of §8).

### 5.2 `httpx` changes (`internal/httpx`; every later story uses them)

This is the first backend story to need the envelope of `services.md` §2.3 and the request id of §2.1, so it lands them. They are not optional extras: TC-97 reads `code` and `X-Request-Id` on every answer.

| Change | Detail |
|---|---|
| `WriteCoded(w, status, code, sentence, field string)` | Writes `{"error": sentence, "code": code}` and `"field": field` only when `field != ""`. `Content-Type: application/json; charset=utf-8` as `WriteJSON`. Constants for the five codes this story uses: `CodeNotFound`, `CodeMethodNotAllowed`, `CodeInvalidQuery`, `CodeInternal`, `CodeUnavailable`; a later story adds the constants it needs |
| `WriteError(w, status, sentence)` is **deleted** | It writes an envelope with no `code`, which the contract no longer allows. Its three callers move to `WriteCoded`: the catch-all in `Router` (404, `NOT_FOUND`, "That address does not exist."), `health` (503, `UNAVAILABLE`, "The service is not ready.") and `Recover` (500, `INTERNAL`, "Something went wrong. Please try again."). `TestWriteErrorUsesTheEnvelope` becomes `TestWriteCodedUsesTheEnvelope` (keys `error` and `code`; `field` only when given). `server_test.go`'s `map[string]string` decoding still holds. No other branch calls `WriteError` yet: the orders and admin backends start after this merges |
| `RequestID(next)` middleware and `RequestIDFrom(ctx) string` | Accepts the caller's `X-Request-Id` when it matches `^[A-Za-z0-9._-]{1,64}$`, otherwise makes 32 lower-case hex characters from 16 bytes of `crypto/rand`. Sets it on the response **before** the handler runs (so it is on every answer including a `Recover`ed 500), and keeps it in the request context. A header that does not match is replaced, never echoed |
| `Log` | Adds `request_id` to the one line per request. Method, path (never the query), status and duration as today |
| Chain in `Router` | `RequestID(Recover(Log(NoStore(mux))))`: the id exists before anything that can fail, and `Recover`'s panic line carries it |

`X-Request-Id` is a response header only; it is never in a body.

## 6. Layout and what is tested without a database

CI has no database, so everything that can be decided before SQL runs is a pure function with a unit test, and the statements are proved on the running stack.

| File | Holds |
|---|---|
| `internal/collections/schema.go` | The fixed table of the eight collections: name, table, noun, whether it carries `version`, and its fields in schema order — key, kind (`id`, `text`, `ref`, `number`, `boolean`, `list`), plus the answer-only column `_source`. This is the **only** place a column name comes from |
| `internal/collections/query.go` | `Query`, `QueryError{Field, Sentence}`, `ParseQuery` (§3) |
| `internal/collections/sql.go` | `BuildList`, `BuildRecord` (§4) |
| `internal/collections/handler.go` | `Collection(db)`, `Record(db)`: collection check, method check, parse, build, run, classify, answer (§2, §5) |
| `internal/httpx/httpx.go`, `httpx_test.go` | §5.2 |
| `internal/server/server.go`, `server_test.go` | The two route lines; the catch-all moved to `WriteCoded` |
| `migrations/20261118000100_collections.{up,down}.sql`, `20261118000200_seed_collections.{up,down}.sql`, `migrations_test.go` | The ERD extension §4, §5 |

The tests (new files; existing tests change only where §5.2 says):

| Test | Proves (story criteria) |
|---|---|
| The field table equals `code/frontend/src/data/schema.json`: same collections, same keys in the same order, kinds from the types; skipped with a message when the file is not there (the Docker build context is `code/backend` only) | A column in the table, the schema and the migration cannot drift apart without a red CI (guards SC-2's "no key the file did not have") |
| `ParseQuery`: defaults; `page`/`limit` bad values and the 100 clamp; empty values; `q` of 100 and 101 characters, blank `q`, `q` on `faqs`; every row of the 400 table with its `field`; the **order of the checks** when two parameters are bad; repeated names; `_source`, `seq`, `version` refused; comma splitting and empty parts; `search` kinds; ranges on a list and a boolean; numeric and boolean value grammar | SC-13 to SC-15, SC-26, SC-29, SC-30, SC-32, SC-33, SC-39 to SC-41, SC-49 |
| `BuildList`: one case per kind and per range end; `NULLS LAST` and the `seq` tie-break in both directions; the search clause over the default set and over `search=`; one placeholder for `q`; placeholder numbering and the `$a`/`$b` arithmetic (page 9, limit 12; page 2, limit 5); the WHERE clause absent when there is no condition; **hostile values** (`'; drop table products; --`, `%`, `_`, `\`, a double quote, a semicolon, a very long value, Vietnamese text) appear in `args` and **never** in the SQL text; the SQL text of two requests that differ only in values is identical | SC-38, SC-44, and "every value reaches SQL as a bound parameter" |
| Handlers with a fake `DB` whose `QueryRow` fills `(int64, []byte)` or `([]byte, int64)`: the list answer has exactly the keys `items`, `total`, `page`, `limit` and echoes `limit` 100 and the asked `page`; unknown collection 404 with its sentence; a record with `pgx.ErrNoRows` 404 with its sentence; `ETag` present and quoted for `products` and `categories` and absent for `reviews`; no `version` in the body; 405 with `Allow: GET` for `POST`, `PUT`, `DELETE`, `HEAD`; a refused connection is 503 and any other error 500, each with the generic sentence and no SQL; the answer carries `Cache-Control: no-store` and `X-Request-Id` | SC-12, SC-16, SC-19, SC-51 to SC-53, TC-97, TC-99 |
| `httpx`: `WriteCoded` with and without `field`; `RequestID` accepts a valid id, replaces an invalid one and a 65-character one, makes 32 hex when none is sent, and the id is on a `Recover`ed 500 | TC-97 |
| `migrations_test.go`: no `.up.sql` containing a data statement mentions `concurrently` | the trap of the ERD extension §5 item 5 |

**What no unit test can see, and who must:** the SQL itself — the column lists, the casts, the `fold_text` and `natural_key` bodies, `jsonb_strip_nulls`, the ordering. A syntax error in a generated statement passes CI and fails every list on the first request. Dev runs the stack (`docker compose --profile local up --build --wait`) and reads one list, one record and one search before opening the backend PR; TL's *Run local* step repeats it.

## 7. How the frontend module speaks the contract (`src/data/index.ts`, DATA-007)

The module is the only frontend file that changes (plus new tests). It keeps every export — `useCollection`, `useRecord`, `schema`, `collections`, `setCollection` and the types `Row`, `Source`, `Field`, `Collection`, `Filter`, `Query` — and what each hook returns. `Filter` and `Query` are unchanged, so no page changes.

### 7.1 The request

- **Base**, read **when a request is made**, not when the module loads (tests set `VITE_API_URL` per case): `(import.meta.env.VITE_API_URL ?? "/api")` with trailing slashes removed; the URL is `<base>/v1/<encodeURIComponent(name)>?<query>` (SC-54).
- **Query**, in this order, joined with `&`: `page`, `limit`; then each `where` entry in the object's key order; then `q`; `search`; `sort`. Each key and each value goes through `encodeURIComponent`, and a list is joined with a literal `,` — **not** `URLSearchParams`, which writes a space as `+` where SC-59 expects `cua%20lo`.
- **`where`**: `undefined`, `""`, `[]` and a range with neither end are **not sent**, not even as `category=`; `false` and `0` **are** sent (`deal=false`, `stock=0`: an answer, not an absence); a string, number or boolean is `key=value`; a list is `key=a,b`; a range is `key_from=` and/or `key_to=` for the ends that are set (SC-58, SC-60). `q` is sent trimmed and only when not blank; `search` only when it has entries; `sort` only when not blank (SC-59).
- **`limit`** is `min(100, max(1, floor(pageSize ?? 12)))`; no request ever carries more than 100 (SC-61). The admin overview's 500 and the reference pickers' 200 become 100.
- **`page`** is the 1-based page being fetched: in the accumulating mode the next one; with `query.page` that page (floored, at least 1).

### 7.2 `useCollection(name, query)`

- **Two modes, as today.** `query.page` given: `items` is that page alone (a table). `query.page` omitted: `items` grows with `loadMore()` (a catalogue).
- **State.** The hook keeps, per mounted instance, the answered pages of the current request identity (`name`, `where`, `q`, `search`, `sort`, `limit`, `query.page`), `total`, `loading`, `error`. `items` is the pages in page order. Nothing is kept across instances: **no cache**; every answer is `no-store` and a page that mounts asks again.
- **A request identity changes** (another filter, another sort, another page, another collection) → a new *generation*: the old requests' answers are **discarded when they arrive** (SC-64), page 1 is asked for, and the accumulated pages restart. Until that answer arrives the hook keeps showing the previous `items` and `total` **of the same collection** with `loading` true (no flash of the empty or no-match state while a filter chip is pressed); another collection shows none. `loading` is derived on the render in which the identity changed, not set by an effect afterwards, so no render ever shows `loading: false` beside records of the old identity.
- **Returns** `{ items, total, pages, page, hasMore, loading, error, loadMore }` as today. `total` is the API's `total`. `pages` is `max(1, ceil(total / limit))`. In the accumulating mode `page` is 1 and `hasMore` is `items.length < total`; with `query.page`, `hasMore` is false and `page` is `query.page` clamped to `[1, pages]`.
- **A page past the end** (SC-57): when an answer has `items: []`, `total > 0` and a `page` above `pages`, the hook asks for page `pages` once, keeps `loading` true meanwhile, and reports that page as `page`. The API does not clamp (`services.md` §3.2); the module does, for its own state. A changed `total` that again puts the page out of range simply asks again, and terminates because `total` is finite.
- **`loadMore()`** does nothing while a request is in flight or when `hasMore` is false: it never asks for a page twice or for a page that cannot exist (SC-55). It asks for the next page and appends it.
- **Failure** (SC-63): a non-2xx answer, a body that is not the list shape, or a rejected `fetch` sets `error` to an `Error` (the envelope's `error` sentence when there is one), `loading` to false, and **leaves `items` and `total` as they were**. A later successful answer sets `error` back to `null`. Nothing throws into a render; the design draws no error screen (SRS §7, question 1).

### 7.3 `useRecord(name, id)`

- `id` undefined: `{ item: undefined, loading: false, error: null }`, no request.
- Otherwise the hook asks `<base>/v1/<name>/<encodeURIComponent(id)>`. The result is **keyed by `name` and `id` and derived on render**: on the first render for an id, and on the first render after the id changes, it returns `{ item: undefined, loading: true, error: null }` — never the previous record, never `loading: false` (the bike page's one-frame flash, SITE-016, would otherwise be this module's defect).
- `200` → `item` is the record, `loading` false, `error` null. **`404` → `item` undefined, `loading` false, `error` null** (the only answer that means "no such record"). Any other status, a malformed body or a rejected `fetch` → `item` undefined, `loading` false, `error` an `Error` (SC-65). The `ETag` is ignored here; the admin story reads it for its own `PUT`.
- When the collection is reported changed (§7.4) the record is asked again; the record on screen stays until the new answer arrives.

### 7.4 `collections()`, `setCollection`, and the refresh signal

This is the mechanism the story left to TL. Two module-level pieces of state:

| State | Meaning |
|---|---|
| `loaded: Record<string, Row[]>` | What the module has seen, per collection, merged by `id` in first-seen order. `collections()` returns it |
| `epochs: Record<string, number>` | One counter per collection; the only thing hooks subscribe to |

- **Every successful answer** (a list page or a record) is merged into `loaded[name]` by `id` **without** touching `epochs` — if it did, every answer would trigger the refresh that fetched it. This is what makes `collections().categories` hold the four riding types, each with `id` and `name`, once any screen has loaded them (SC-67, DATA-007 Behaviour 8), and what the admin lists' "Riding type" lookup reads.
- **`setCollection(name, rows)`** does two things, synchronously: sets `loaded[name] = rows` (so `collections()[name]` is `rows` **at once**, which `src/test/admin.test.ts` depends on and which it may not be edited to avoid), and increments `epochs[name]`, then notifies subscribers. This is how the management area says "this collection changed" (SC-67).
- Hooks read `epochs[name]` through `useSyncExternalStore` (a stable number, so there is no tearing and no render loop). A change of the number is a new generation for **that collection's** mounted lists and records only: page 1 is asked for again (an accumulating list collapses back to its first page; a paged list asks for its current page), a record asks again. A list of another collection sends nothing (TC-77).
- **Consequence, accepted.** `saveRow` and `deleteRow` in `components/admin/store.ts` still call `setCollection` and still write nothing to the server until the admin story replaces them with real writes; until then the list asks the API again and shows the server's records, so a design-mode edit does not stay on screen. The Studio's frame (`editable.tsx`) also calls `setCollection` with its unsaved rows and posts `collections()` to the Studio: with the JSON files gone, that panel sees what the open page has loaded rather than the whole collection, and its live preview of unsaved edits is replaced by the server's records on the next ask. The story puts the Studio and the design walk out of scope (SRS §7) and they do not call this API; the owner edits the live shop in `/admin`.

### 7.5 Identical requests in flight are shared

The existing components ask for the same list from many places at once (`ProductCard` asks for `categories` once **per card**, `ReviewCard` for `products` once per review, every page for `shop`). The module therefore keeps `inflight: Map<collection, Map<url, Promise>>`: a request for a URL that is **already in flight** joins that promise. The entry is removed the moment it settles, so a later request always goes to the server — **this is not a cache**; no answer is ever reused. `setCollection(name, …)` also drops every entry of `name`, so an ask made because of a change cannot join a request sent before the change. Because requests are shared, they are never aborted; a stale answer is dropped by the generation check of §7.2, not by cancelling the request. The tests therefore may not assume that two hooks with the same query send two requests.

### 7.6 Where the API departs from today's module, and what the frontend does about it

Today's module filters and sorts the JSON in the browser. The API follows the contract (`services.md` §3.2), not the module; each difference is named, with the frontend consequence.

| Today's module | The API | Frontend change |
|---|---|---|
| Records are there on the first render; `loading` is false one tick later | `items` is `[]` and `loading` true until the first answer | None in a page: they read `items` and `loading` as before. A page that reads only `item` may draw its empty state during the wait; that is SITE-016 and ADMIN-013, not this story (story §8.5) |
| `collections()` is every record of every collection | The records the module has loaded so far | None in the module. Consumers that look a record up by id (`fields.tsx`) read what a screen has loaded; the admin lists load the riding types themselves (ADMIN-010) |
| A descending sort reverses the whole list, so equal values come out in reverse file order | Equal values keep `seq` order in both directions (DATA-005 AC-5) | None. Visible only with ties: stock 3, 4 and 5 each occur on two bikes, so `-stock` shows them in a different order; no criterion reads it |
| A record with no value sorts as the empty text: first ascending, last descending | After every record that has one, in both directions (DATA-005 AC-8) | None. The seeded bikes all have a price, stock and name |
| A range on text compares with `localeCompare`; a `where` key that is not a field matches nothing | Byte order on the stored text (ISO dates compare right); an unknown name is `400` | None. Every call site read for this design passes schema fields only: `Shop`, `Search`, `Home`, `Product`, `RelatedRow`, and `ListPage`'s `filters`. A `sort` on a list column (`sizes`) is a `400` and an empty table; ADMIN-010 owns which columns the admin tables offer |
| `where: { sizes: "M" }` matched nothing (arrays are never `===` a string) | A list field matches when it contains the value | None; no caller does it. The shop narrows by frame size on the loaded list |
| Text equality and search fold capitals and accents | The same, in the database (`fold_text`) | None; `"cua lo"` finds `Cửa Lò`. `đ` folds to `d`, so `q=deo` also finds `video` (TC-42's note); the old module did not fold `đ` |
| A fetch is not a thing | A request per mounted list, shared while in flight (§7.5), no cache | None |

### 7.7 Tests (new files under `src/test/` only)

`src/test/data.test.ts` with a stubbed global `fetch` that records each URL and answers pages of a fixture of the 14 bikes, using `renderHook` and `vi.stubEnv` (both already in the frozen `package.json`: `@testing-library/react` 16, `vitest` 2, `jsdom`). It covers SC-54 to SC-67, one case per criterion, as TC-63 to TC-77 describe them. The request-term builder is exercised through the hook (its URLs), not exported. `src/test/admin.test.ts` and `src/test/format.test.ts` pass unchanged.

## 8. Conflicts with merged decisions, resolved

| # | Merged decision | What this story needs | Options and cost | Choice |
|---|---|---|---|---|
| 1 | `erd.md` §5 pattern 3: search is `LIKE` with `%`, `_`, `\` escaped | The same observable behaviour (DATA-004 AC-7) | Keep `LIKE`: an escaping function, and a mistake there is a visitor's `100%` matching every bike. Use `strpos`: no wildcards, nothing to escape, same answers | `strpos` (ERD extension C2). The contract in `services.md` §3.2 says "escaped for `LIKE`"; the **behaviour** it specifies is unchanged, and only the mechanism differs |
| 2 | `services.md` §2.3 and §5: `UNAVAILABLE` is "the database refused the connection, **the pool cannot hand out a connection**, or the process is shutting down" | A way to tell them apart in code | A pool that cannot hand out a connection waits until the context's deadline and returns `context.DeadlineExceeded`, which is also what a slow statement returns; it cannot be told from one. A closed pool at shutdown cannot be hit by these reads, because the server drains in-flight requests before `main` closes the pool | `503` only for a refused connection and the SQLSTATEs of §5.1; a deadline is `500`, as `services.md` §5 already says for "any other failure, including a deadline". The §2.3 wording about the pool is narrowed to what code can see |
| 3 | `httpx.WriteError` (scaffold) and `convention.api` ("keeps `error` and gains sibling code/field") | Every non-2xx answer carries `code` | Keep `WriteError` beside a new helper: two ways to write an error, one of which breaks the contract. Replace it | `WriteCoded`, and `WriteError` deleted (§5.2). `convention.api` is updated |
| 4 | The story's scope lists routes and the `DB` interface, not `httpx` | `X-Request-Id` on every answer and `code` on every error (TC-97), both in `services.md` §2.1 and §2.3 | Leave them for "the first story that needs them" and have three stories each add their own | This story is the first backend story, so it carries them (§5.2). They are shared files; the orders and admin backends start from them |
| 5 | Story §8.1 / `erd.md` R2: columns keep `_source` and `_invented` verbatim | The plan item says "`_source` becomes a column `source`" | A renamed column needs a translation in every read and write | The ERD wins, as the story already says; nothing here changes it |

## 9. Rollout, dependencies, open questions

- **Order.** The frontend PR (module, tests, eight deleted files) and the backend PR (migrations, `collections`, `httpx`, routes) are reviewed separately and **merged back to back, frontend first**, as the delivery flow says; the deployment waits for both. Between the two merges `main` has no JSON files and no API, which is why nothing is deployed between them. The seed is regenerated if any JSON file changed since it was generated (ERD extension §5 item 1).
- **Depended on.** *Place an order and hold the bike* reads `products.stock` and the `shop` record through this API; *Shop admin* uses the tables, `ETag`, `collections()` and `setCollection`; *Bike page waits for its record* needs `useRecord`'s `loading` and `error` from §7.3 (its page already changed, PR #13, and is invisible until this lands).
- **Secrets and setup.** None. `DATABASE_URL` is injected by the runtime and already in `code/backend/.env.example`; this story adds no key and no `request_setup`.
- **Open questions.** None blocking. SRS §7's three defaults stand: no error screen for an unreachable API (the pages draw the empty states they have, TC-94); `q=deo` finds `Đèo` (and, harmlessly, `video`); the admin overview counts to 100 bikes.
