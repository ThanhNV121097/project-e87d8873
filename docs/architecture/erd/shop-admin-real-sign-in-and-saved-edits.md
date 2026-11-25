# ERD extension — Shop admin: real sign-in and saved edits

Story: `docs/admin/stories/shop-admin-real-sign-in-and-saved-edits.md` · Requirements: ADMIN-001 to ADMIN-013 (`docs/admin/SRS.md`) · Plan item: *Shop admin: real sign-in and saved edits* (feature, P1) · Migration slot: `20261118000400_sessions` · Companion: `docs/architecture/services/shop-admin-real-sign-in-and-saved-edits.md`

## 1. Result

**This story adds no entity, no column, no index and no constraint to the model.** `sessions` — its columns, nullability, `CHECK`s, its one index and its lifecycle — is already defined by the project-wide design (`docs/architecture/erd.md` §3.12, §3.13, §5 pattern 12, §8 row 4). The tables it writes to, `products` and `categories`, with their `version` column, `touch_record` trigger, foreign keys and `CHECK`s, are the data story's (`erd.md` §3.1, §3.2, §3.14, §8 row 1). Nothing of that is restated here. This story **builds** one table (migration slot `000400`) and **writes** to two tables the data story builds.

There is no mock module to compare with the database: PR #20 has none (`store.ts` and `session.ts` are the API clients, written against `services.md`). The record the frontend renders is the record the read API already serves (`erd.md` R2, `services.md` §2.7); every key a write can carry is a column of `products` or `categories`, and no key the screens send lacks a column (§5).

What this file adds, because the project-wide design left it to the builder and this is the only gate where the writes are read against the real tables:

| § | What |
|---|---|
| 2 | The statements of the story, one table for the owner's writes and one for the session: what each reads or writes, its lock, the index that serves it |
| 3 | Eight clarifications of the model. None changes a table; each pins how a statement is written, and says why |
| 4 | Build rules for migration 4 |
| 5 | The writes checked against the real bikes (the seeded round trip of SC-58) |
| 6 | The SQL the verification step runs |
| 7 | The migration plan: forward, backward, safety on a populated table, order of delivery |
| 8 | Open questions |

## 2. What the story reads and writes

**No transaction is opened by this story.** Every owner write is **one statement answered by one row**, and every session operation is one statement, so `server.DB` keeps the one method the data story gave it (`QueryRow`) and this story does not depend on the `Begin` the orders story adds. The statements are fixed text built once per collection from the same field table the data story's reads use (`internal/collections/schema.go`, the only place a column name comes from); a value is only ever a placeholder with an explicit cast; column names are quoted, as in migration 1. Below, `<cols>` is the collection's **schema-order** list of writable columns (for `products`: `image`, `name`, `brand`, `category`, `year`, `price`, `was`, `price_listed`, `fact`, `status`, `stock`, `sizes`, `sizes_out`, `size_fit`, `deal`, `month`, `wheel`, `travel`, `frame`, `fork`, `shock`, `motor`, `drivetrain`, `brakes`, `wheels`, `tyres`, `weight`, `colour`, `description`, `ride`, `rating`, `review_count`, `image_back`, `image_detail`, `_invented` — 35 of them; for `categories`: `name`, `line`, `terrain`), and `<answer>` is the data story's answer-column list for the collection (the schema columns, `_source`), followed by `"version"`. **`seq`, `version`, `created_at`, `updated_at` and `_source` are never in an `INSERT` or `SET` list** (`seq` is `GENERATED ALWAYS`; `version` and `updated_at` belong to the trigger; `_source` is untouched by a replace and `NULL` on an add, `erd.md` §9 question 2).

### 2.1 The owner's writes (`products` and `categories`)

```sql
-- W1  add. One row, or none when the id is taken. Every column is written: an absent field is NULL.
WITH w AS (
  INSERT INTO "products" ("id", <cols>) VALUES ($1::text, $2::text, …, $36::text[])
  ON CONFLICT ("id") DO NOTHING
  RETURNING <answer>)
SELECT jsonb_strip_nulls(to_jsonb(w) - 'version') AS rec, w."version" FROM w;

-- W2  what the save needs to know before it writes: does the record exist (and at which version),
--     does the riding type exist. Always one row. (products; categories takes $1 only and has no second column)
SELECT (SELECT t."version" FROM "products" AS t WHERE t."id" = $1::text) AS version,
       ($2::text IS NULL OR EXISTS (SELECT 1 FROM "categories" AS c WHERE c."id" = $2::text)) AS category_ok;

-- W3  replace, compare-and-swap on the version the caller read. The trigger adds 1.
WITH w AS (
  UPDATE "products" AS t SET "image" = $3::text, "name" = $4::text, …, "_invented" = $37::text[]
  WHERE t."id" = $1::text AND t."version" = $2::bigint
  RETURNING <answer>)
SELECT jsonb_strip_nulls(to_jsonb(w) - 'version') AS rec, w."version" FROM w;

-- W4  after W3 changed no row: gone, or changed. One row at most.
SELECT "version" FROM "products" WHERE "id" = $1::text;

-- W5  add only: does the riding type exist (run only when the body carries a category)
SELECT EXISTS (SELECT 1 FROM "categories" WHERE "id" = $1::text);

-- W6  delete. No row = no such record. A riding type that a bike uses fails with SQLSTATE 23503 (RESTRICT).
DELETE FROM "products"   WHERE "id" = $1::text RETURNING "id";
DELETE FROM "categories" WHERE "id" = $1::text RETURNING "id";

-- W7  after the riding type's delete failed with 23503: its name and how many bikes use it, for the 409 sentence.
SELECT c."name", (SELECT count(*) FROM "products" AS p WHERE p."category" = c."id")
FROM "categories" AS c WHERE c."id" = $1::text;
```

`INSERT … VALUES` binds `$2…` in `<cols>` order; `W3` binds `$1` the id, `$2` the version, `$3…` in `<cols>` order. For `categories` the lists are `name`, `line`, `terrain` and the statements are otherwise the same with no `category` parameter. `to_jsonb(w) - 'version'` and `jsonb_strip_nulls` are the projection of the data story's one-record read, so **an add or a replace answers the same JSON a later `GET /v1/{collection}/{id}` returns** (SC-16 "reading it returns it"; SC-17).

| Statement | Table, columns | Lock taken | Index that serves it | Used by |
|---|---|---|---|---|
| W1 | `products` / `categories`: the row inserted; the `fk_products_category` check reads the `categories` row (key-share) | none beyond the new row | primary key (the `ON CONFLICT` arbiter), `idx_products_category` is not touched | `POST` |
| W2 | `products."version"`, `categories."id"` | none (plain reads) | primary keys | `PUT` |
| W3 | the whole `products` / `categories` row (every `<cols>` column, `version`, `updated_at` through the trigger) | the **row lock** of the update. An order's transaction holds the same lock from its `SELECT … FOR UPDATE` to `COMMIT`; the update waits for it, then re-reads the row's `version`, finds it changed and matches nothing | primary key | `PUT` |
| W4 | `products."version"` | none | primary key | `PUT` after a zero-row W3 |
| W5 | `categories."id"` | none | primary key | `POST` |
| W6 | the row; `reviews.bike`, `gallery.type` rows that point at it are changed by `ON DELETE SET NULL`; a `products.category` row that points at it makes the statement fail | row lock on the deleted row (waits for an order that holds a bike) | primary key; `idx_reviews_bike` and `idx_gallery_type` serve the `SET NULL` scans, `idx_products_category` serves the `RESTRICT` check | `DELETE` |
| W7 | `categories."name"`, `count(*)` of `products` | none | primary key, `idx_products_category` | `DELETE` of a riding type after a 23503 |

A save is **two statements** (W2, W3; W4 only when W3 matched nothing), an add is **one or two** (W5 only when the body names a riding type, then W1; a server-made id that collides is retried, §3 C4), a delete is **one** (W6; a riding type that is in use adds W7). All are single-row; the 1 s p95 of the admin SRS §6 has two orders of magnitude of room.

**Why a save cannot overwrite an order's hold (ADMIN-008).** The compare-and-swap is the `WHERE … AND "version" = $2` of W3, evaluated by Postgres under the row lock. `READ COMMITTED` re-evaluates a blocked `UPDATE`'s `WHERE` against the row the other transaction committed, so a save that waited for an order sees the new `version` and changes nothing (→ 409, SC-33); a save that lands first bumps `version` and the order's `SELECT … FOR UPDATE` then reads the saved row and lowers *its* stock (SC-34). The save never holds a second resource while it holds the row, so it cannot close a lock cycle with an order (`docs/architecture/erd/place-an-order-and-hold-the-bike.md` §2). Delete sends no version and is not refused for a stale one (SC-35); a delete that waits for an order's lock removes the row after the order commits, and the order's lines stay (`order_lines.product_id` is a copy, `erd.md` §3.10, SC-29).

### 2.2 The session

```sql
-- A1  sign-in: purge what has expired and store the new session, in ONE statement. $3 = the clock's now, $4 = now + 7 days
--     (both computed in Go from the one injected clock). The purge runs even though nothing reads its output.
WITH purged AS (DELETE FROM "sessions" WHERE "expires_at" <= $3::timestamptz RETURNING 1)
INSERT INTO "sessions" ("token_hash", "email", "expires_at") VALUES ($1::bytea, $2::text, $4::timestamptz)
RETURNING "id";

-- A2  the me check and the guard on every write. One row or none. $3 = the clock's now.
SELECT 1 FROM "sessions" WHERE "token_hash" = $1::bytea AND "email" = $2::text AND "expires_at" > $3::timestamptz;

-- A3  sign-out. No row is still success (ADMIN-003 AC-9).
DELETE FROM "sessions" WHERE "token_hash" = $1::bytea RETURNING 1;
```

| Statement | Table, columns | Lock | Index | Used by |
|---|---|---|---|---|
| A1 | `sessions`: the new row; every expired row deleted | row locks on the expired rows only; the table holds a handful of rows, so the purge is a scan and needs no index on `expires_at` (`erd.md` §3.12) | `uq_sessions_token_hash` (checked at insert) | `POST /v1/sessions` |
| A2 | `sessions` | none | `uq_sessions_token_hash` | `GET /v1/me` and `requireOwner` on every `POST`/`PUT`/`DELETE` of a collection |
| A3 | `sessions` | row lock on the deleted row | `uq_sessions_token_hash` | `DELETE /v1/sessions/current` |

A2 is the only statement a request makes **before** it does anything for the owner, so it is a plain indexed read: it writes nothing and does not extend the session (fixed 7 days, `erd.md` §3.12).

## 3. Clarifications of the model

None of these changes a table, a column, a constraint or an index. Each pins a choice the model left open, so the builder and the reviewer read the same thing.

| # | Where the model is open | Decision | Rejected alternative, and why |
|---|---|---|---|
| C1 | `erd.md` §3.12 describes `token_hash` as "SHA-256 of the random session token"; `services.md` §2.2 (and `decision.session_binding`) make it `HMAC-SHA256(key = credential_fingerprint, message = the cookie's text)` so that changing `ADMIN_EMAIL` or `ADMIN_PASSWORD` and restarting ends every old session | **`services.md` §2.2 governs.** The column is the same: `bytea`, 32 bytes, `CHECK (octet_length("token_hash") = 32)`, unique. The HMAC message is the **43-character cookie text as received**, never its decoded bytes, so a cookie with one character changed — including the last one, whose 2 low bits base64url ignores when decoding — hashes differently and finds nothing (SC-11) | Hashing the decoded bytes: two different cookie texts would reach the same session. Plain SHA-256: a copy of the database plus a guessed token would still work after the owner changed the password |
| C2 | `erd.md` §3.12 words the lookup as `expires_at > now()` and says `expires_at` is "set by the application" | **One clock, the application's, injected.** `expires_at` is `now + 7×24 h` computed in Go (not `interval '7 days'`, which follows the session's time zone and daylight saving), and the lookup and the purge compare with the same `$3`, not with `now()`. `created_at` keeps its `now()` default and is never read. The `CHECK (expires_at > created_at)` holds while the two clocks are within 7 days of each other, which a single host guarantees | `now()` in the lookup: the 7-day expiry and the "older than 7 days is 401" rule (SC-11) could then only be shown by waiting a week or editing a row; with the injected clock a unit test sets it |
| C3 | `services.md` §3.6 says a riding type's delete "runs in one transaction: lock the row, count its bikes, delete; if a bike took the type in the instant between … the count is read again"; §3.4/§3.5 say the riding-type check is made "inside the transaction" | **No transaction.** The delete is attempted first (W6); the foreign key `fk_products_category` (`RESTRICT`) is the authority, and its SQLSTATE `23503` — the only foreign key that can refuse a delete of a riding type, since `gallery.type` is `SET NULL` — is what turns it into the 409; W7 then reads the name and the count for the sentence. If W7 finds the type gone the answer is `404`; if it finds **0** bikes (a bike moved away in the instant between) W6 is tried again, at most 3 times, then `INTERNAL`. For the riding type of a bike or a save, W2/W5 read it first, in the order the contract fixes (400 before 409), and a `23503` from W1/W3 — the type deleted in the instant between — is the same 400 | A `BEGIN` / `SELECT … FOR UPDATE` / count / delete transaction: the same observable answers, but it needs `server.DB.Begin` (which another story adds) and still needs the foreign key as the backstop, so it is a second mechanism for the same guarantee |
| C4 | `services.md` §3.4 says an id the server makes is retried "on a collision up to 5 times" and the id given in a body "is taken" → 409 | The collision is detected by `ON CONFLICT ("id") DO NOTHING` returning no row (W1), not by catching SQLSTATE `23505`: the statement does not fail, so nothing is aborted and no savepoint is needed. A body id is tried once (no row → `ID_TAKEN`); a server-made id is tried up to 5 times with a new random id each time, then `INTERNAL`. The unique violation on `seq` cannot happen (identity). The identity value a skipped insert took leaves a gap in `seq`, which changes no ordering (`erd.md` §8 row 2) | Catch `23505` and look at the constraint name: the Go code would name a constraint, and the failed statement would abort an enclosing transaction |
| C5 | `erd.md` R3: an absent or empty field is `NULL`; `services.md` §3.5: a replace sets every absent schema field to `NULL` | The fixed `INSERT` and `UPDATE` carry **every** writable column; the handler binds Go `nil` for "no value" and the typed value otherwise. A **list**: `null` or absent binds `nil` (→ `NULL`, no key in the record); `[]` binds a **non-nil empty slice** (→ `'{}'`, the key `[]`). A Go `nil` slice and an empty slice are different values to pgx — the difference is the whole of R3 for lists and the reason `sizes_out: []` of `talon-2` survives an edit (§5) | Building the column list from the keys present: a different statement per body, so a replace would have to delete the absent columns separately, and the statement text would depend on request data |
| C6 | `erd.md` §3.0 types `year`, `rating`, `review_count` as `numeric`, "kept exactly" | A number goes in as the **text of its JSON token** and is cast in SQL (`$n::text::numeric`), never through a Go `float64`, as the data story reads them (its C4): `2027.0` is stored as `2027.0` and read back as `2027.0`, a rating of `5` as `5`. `price`, `was` and `stock` are whole numbers and bind as Go `int64` (`::bigint`). A numeric token longer than 40 characters is refused by the validator (`services` extension §4), because Postgres refuses a `numeric` with a scale over 16,383 digits with an error the handler could only call a 500 | Parsing every number to `float64`: a stored `5` would come back `5` but `2027.0` would come back `2027`, and a price near 2^53 would lose digits |
| C7 | `services.md` §2.5: a `CHECK` that fires means a validator is missing a rule, and the answer is a 500 | The handler classifies a failed statement by **SQLSTATE only**, never by constraint name: `23503` (foreign key) on a `products` write is the riding-type 400, on the `categories` delete the 409; **every other class `23` code** (`23502` not null, `23514` check, `23505` unique) is `INTERNAL` with the cause logged at `Error` and the request id — a text with U+0000 would be one if the validator missed it, which is why it does not (`services` extension §4) | Naming `ck_products_live_has_price` in Go to give a nicer sentence: the sentence already comes from the validator, and a name in Go is a second place to rename |
| C8 | `erd.md` §3.12: "No index on `expires_at`"; "expired rows are purged on each sign-in" | Unchanged. The purge is part of A1, so a failed purge fails the sign-in with the same 500 as a failed insert (no cookie is set), and a table that is never signed in to is never purged — it holds at most one row per sign-in of the last 7 days and nothing reads an expired row (A2 compares `expires_at`) | A scheduled purge: a worker for a table of a handful of rows (`services.md` §4: no background work) |

## 4. Build rules for migration 4 (`20261118000400_sessions`)

1. **Quote every column name** in the DDL and in the three statements (`"token_hash"`, `"email"`, `"expires_at"`), as migrations 1 and 3 do.
2. **Columns exactly as `erd.md` §3.12:** `"id" uuid NOT NULL DEFAULT gen_random_uuid()` as the primary key; `"token_hash" bytea NOT NULL`; `"email" text NOT NULL`; `"created_at" timestamptz NOT NULL DEFAULT now()`; `"expires_at" timestamptz NOT NULL`. No nullable column, no foreign key, no `updated_at` (the row is never changed).
3. **Constraint and index names** (the model gives the index its name and the two `CHECK`s none; the convention `ck_<table>_<rule>` names them): `CONSTRAINT uq_sessions_token_hash UNIQUE ("token_hash")`, `CONSTRAINT ck_sessions_token_hash_length CHECK (octet_length("token_hash") = 32)`, `CONSTRAINT ck_sessions_expiry_after_creation CHECK ("expires_at" > "created_at")`. The Go code names none of them (C7) and no answer carries one.
4. **No second index.** In particular none on `expires_at` (`erd.md` §3.12) and none on `email` (no query filters on it alone: A2's `email` is a predicate on a row already found by `token_hash`).
5. **No `CONCURRENTLY` anywhere in the file, in any case**, and no data statement: the runner moves a file out of its transaction on seeing the word, even inside a string (`docs/architecture/erd/site-data-from-the-database.md` §5 item 5). The file only creates a table.
6. **The file depends on no other table**, so it needs neither migration 1 nor migration 3 to have run; `gen_random_uuid()` is built in since Postgres 13 (no extension).
7. **The down file** is exactly one statement: `DROP TABLE "sessions";`. It mentions no other table and creates or drops no function or trigger (this story creates none).
8. **`migrations_test.go`** gains a test that reads the pair: the up file creates exactly `sessions`; the down file drops exactly it; neither has `REFERENCES`, a statement on another table, or the word `concurrently`. It checks text, not Postgres: CI has no database.

## 5. The writes checked against the real data

The schema (`code/frontend/src/data/schema.json`), the four riding types, and the stock, category and status of all 14 bikes were read in full from `main`; the first three bikes (`talon-2`, `talon-29-3`, `talon-29-1`) were read key by key. The seed's own pre-flight (`docs/architecture/erd/site-data-from-the-database.md` §3) covers every other record against the same constraints.

| Check | What the data holds | Result |
|---|---|---|
| **Field order of the validator and of the statements** | `schema.json` lists `products` as `image`, `name`, `brand`, `category`, `year`, `price`, `was`, `price_listed`, `fact`, `status`, `stock`, `sizes`, `sizes_out`, `size_fit`, `deal`, `month`, `wheel`, … `image_back`, `image_detail`, `_invented`: **`image` comes first, `name` second**, and `image_back` / `image_detail` sit at the end. The key table of `services.md` §2.7 lists `name` first | The contract's "fields are checked in the order of `schema.json`" governs (SC-B3). The write layer takes the order from the same field table as the reads, which the data story's test pins to `schema.json`; it does not copy the §2.7 table |
| **`stock` has no type of its own** | `schema.json` types `stock` as plain `number`, as `year`, `rating`, `review_count` are; `price` and `was` are `price` | The kind alone cannot tell `stock` (a whole number ≥ 0, `bigint`) from `year` (any number, `numeric`). The write layer therefore keys "whole" on the column: `price`, `was`, `stock` — the column mapping of `erd.md` §3.0 ("`number` named `stock` → `bigint`") |
| **Closed sets** | `brand` options `Giant`, `Liv`; `status` options `draft`, `live`; `wheel` options `29`, `mixed` (stored as text, `"29"`), each in `schema.json` as `options` | The validator's closed sets are the schema's `options` (the field table carries them; if the data story's table does not, this story adds them and extends that story's schema test). They equal the `CHECK`s `ck_products_brand`, `_status`, `_wheel` |
| **The seeded round trip (SC-58)** | `talon-2` is `id`, `_source` and 32 of the 35 schema keys (it has no `was`, `month` or `motor`; `_invented` is one of the 32); every key is a schema key, `id` or `_source`; `_invented` is `["weight","stock","rating","review_count","ride"]`; `price_listed` is `"A$1,099 RRP"`; `sizes_out` is `[]`; `deal` is `true` or an explicit `false`; `rating` is `4.8`, `year` `2027`; the longest text is a few hundred characters | A client that sends the record back **minus `_source`** passes every rule: no undefined key, every text far below 5,000 characters, every number token below 40 characters, `deal: false` is a real `false` (not "no value"), `sizes_out: []` is a real empty list (C5). The replace writes the same value into every column except the stock, and leaves `_source` alone, so the saved record equals the seeded one except for the stock, and the `ETag` has moved from `"1"` to `"2"` |
| **A seeded stock of 0 is a stored value** | `stock` is `0` on two bikes (the seventh bike and the thirteenth, both in `cross-country` / `e-mtb`) | `0` is a JSON integer, not an empty value: it binds as `0`, never as `NULL` (`""` and `null` are the only empties, `services.md` §3.11). SC-71 (stock 0 → "Out of stock") and SC-61's "Out of stock 2" depend on it |
| **Running low** | the 14 stocks are 9, 6, 4, 3, 2, 1, 0, 7, 5, 3, 4, 5, 0, 2: all under 10 | "Running low 14" (SC-61) is read by the screen from `GET /v1/products` with `limit` 100; this story's writes only change the figures |
| **Riding types in use** | `hardtail` 4 bikes, `cross-country` 3, `trail` 4, `e-mtb` 3 (14 in all, none unused) | `DELETE /v1/categories/trail` is the 409 of SC-27 and SC-83 with "Trail is still the riding type of 4 bikes. Give those bikes another riding type first." (`name` is `Trail`). **No seeded type can be deleted and none has exactly one bike**, so the 204 (SC-27, SC-83) and the "…of 1 bike." sentence need a riding type the verification adds first (and, for the singular, one bike moved onto it) |
| **`month`** | set on one bike only (`trance-x-1`, `trail`) | Absent elsewhere, so a replace that omits it keeps it absent; a replace of `trance-x-1` that omits it removes "Bike of the month" (replace semantics, ADMIN-005 AC-4) |
| **Draft bikes** | none: every seeded bike is `live` | "0 still in draft" (SC-61). A draft the owner adds with no price is stored with `price` `NULL` and `status` `'draft'`; `ck_products_live_has_price` passes because the status is not `live` |
| **Money cannot overflow** | the largest accepted `price`/`was`/`stock` is 9,007,199,254,740,991 (`services.md` §3.11) | Fits `bigint` and a JavaScript safe integer, so the screens read back exactly what the owner saved. `ck_products_price_nonneg`, `_was_nonneg`, `_stock_nonneg` are the backstop and never fire |
| **No key the screens send lacks a column** | the screens send the record they read (without `_source`), `wheel` as text, no `id` on an add | every key is a `<cols>` column or `id`; PR #20's `bodyOf` drops `_source` and an empty `id`, nothing else |

## 6. SQL the verification step runs

CI has no database, so everything that needs Postgres is proved against the running stack. Read-only unless a line says otherwise.

**A session holds a hash, never a token or a password (SC-2, SC-4, SC-11).**

```sql
SELECT id, octet_length(token_hash) AS hash_bytes, email, created_at, expires_at,
       expires_at - created_at AS lifetime            -- 7 days (to the second), and 32 bytes
FROM sessions ORDER BY created_at;
```

**A session older than 7 days answers 401 (SC-11a).** The verification ages one row, then reads `me` with that cookie. A write statement; the session is signed out afterwards:

```sql
UPDATE sessions SET expires_at = now() - interval '1 minute', created_at = now() - interval '7 days 1 minute'
WHERE id = $1;                                         -- $1 = the row of the cookie under test (`created_at` is moved too so the CHECK holds)
```

**A save and a delete wrote what the screen said (SC-17, SC-22, SC-58).** The record the owner saved equals what the table holds; no column holds an empty text or an `'{}'` where the body omitted the key:

```sql
SELECT id, version, stock, price, status, _source, _invented, was, month
FROM products WHERE id = 'talon-2';                    -- version rises by 1 per save and per order; _source never changes
SELECT count(*) FROM products WHERE name = '' OR image = '' OR fact = '' OR description = '';   -- 0: an empty text is NULL (R3)
```

**Delete leaves the right rows behind (SC-28, SC-29).**

```sql
SELECT id, bike FROM reviews WHERE bike IS NULL;        -- reviews of a deleted bike: still there, no bike
SELECT id, type FROM gallery WHERE type IS NULL;        -- photographs of a deleted riding type: still there, no type
SELECT count(*) FROM order_lines WHERE product_id = $1; -- unchanged by the delete of that bike (SC-29b)
```

**A refused write changed nothing (SC-25, SC-32).** Read before and after the refusal and compare:

```sql
SELECT id, version FROM products ORDER BY id;  SELECT id, version FROM categories ORDER BY id;
```

**The owner's save and an order do not overwrite each other (SC-33, SC-34).** Until the orders story is live, a direct update of the stock stands in for the order; the trigger bumps `version` exactly as an order's hold does:

```sql
UPDATE products SET stock = stock - 2 WHERE id = 'talon-2';     -- the stand-in for an order of 2; version +1
SELECT stock, version FROM products WHERE id = 'talon-2';       -- the PUT with the old ETag is 409 and these are unchanged by it
```

**Invariants that must hold after any run.** Each returns no rows or a zero:

```sql
SELECT count(*) FROM products WHERE stock < 0 OR price < 0 OR was < 0;                       -- 0
SELECT id FROM products WHERE status = 'live' AND COALESCE(price, 0) <= 0;                   -- no rows (SC-21)
SELECT id FROM products WHERE category IS NOT NULL AND category NOT IN (SELECT id FROM categories);   -- no rows
SELECT count(*) FROM sessions WHERE expires_at <= created_at;                                -- 0
```

## 7. Migration plan

| Aspect | Decision |
|---|---|
| Forward | One file, `20261118000400_sessions.up.sql`, applied by the server on boot in filename order, before it listens, in one transaction with its `schema_migrations` row, under the advisory lock. The slot is this story's alone; migrations 1 and 2 (data story) are stamped earlier and migration 3 (orders story) between them and this one, and never collide |
| Backward | `20261118000400_sessions.down.sql`: `DROP TABLE "sessions";`, **for development** (the runner never applies a down file and does not delete the `schema_migrations` row, so a manual rollback deletes that row by hand). Dropping the table signs the owner out everywhere; nothing else is lost, because a session holds no data of the shop. Rolling back a shop that holds owner edits is a restore from a backup, not this file |
| Safe on a populated table | Yes. The table is new and has no foreign key, so applying it touches no row of `products`, `categories`, `orders` or any other table. This story's other persistent change is not a migration: the owner's writes to `products` and `categories`, which need no schema change (the `version` column and the trigger are the data story's) |
| Failure | A statement that fails rolls back the file and the server does not start; the next start retries the same file |
| Order of delivery | The backend PR **cannot be built before the data story's backend merges**: its handlers widen `collections.Collection` and `collections.Record`, read the field table in `internal/collections/schema.go`, and every write needs the tables and the trigger of migration 1. Its migration (`000400`) is independent of all the others. The orders story is independent of it, except that SC-29, SC-33 and SC-34 need a real order: until it lands the verification uses the direct update of §6, and SC-34b stays manual. The frontend (PR #20) merges first (steps 11 and 12 of the delivery flow); until the backend merges the management screens show their generic "could not …" sentences |
| `ADMIN_EMAIL`, `ADMIN_PASSWORD` | Requested from the owner in the console (`request_secret`) and already named, without values, in `code/backend/.env.example`. Until both are supplied nobody can sign in (SC-3) and the public shop is unaffected; a local run and the verification that cannot wait use throwaway values that are **not committed** and say so |

## 8. Open questions

None blocking. The four open questions of `docs/admin/SRS.md` §7 keep their proposed defaults (the story's §11), and none touches the model. Question 2 of `erd.md` §9 (does a save keep `_source`?) is already decided: it does, because the replace never writes the column (§2.1).
