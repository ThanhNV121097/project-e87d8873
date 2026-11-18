# Test cases — Site data from the database

Module: `data` · Plan item: *Site data from the database* · Story: `docs/data/stories/site-data-from-the-database.md` (SC-1 to SC-82)
Requirements: `docs/data/SRS.md` (DATA-001 to DATA-008) · Contract: `docs/architecture/services.md` (§2.1 base, §2.3 errors, §2.4 paging, §2.5 validation, §2.7 records, §3.2 list, §3.3 one record) · ERD: `docs/architecture/erd.md` (§3.0 to §3.8, §3.14)

**Risk level: medium-high.** The function writes nothing at run time, reads are public and no personal data is held, so the risk is not security. It is the *move*: a price, a size list, an accent or a record order that changes silently loses what the owner approved, and every page reads these records. The depth is therefore highest on exactness (section A), on the query rules every page leans on (sections C to E) and on page parity (section H); the hook (section G) is proved at module level because its request terms are not visible from a browser.

## 0. How to read this file

**Case format.** Every case is `**Scenario**`, `**Given**`, `**When**`, `**Then**`, `Traces:` and `Check:`. Cases are numbered `TC-nn` in the order of the story's criteria; `Traces:` names the story criterion (`SC-n`) and, in brackets, the SRS criterion it narrows, or `contract (…)` for a case the service contract demands.

**Paths.** Every API path below is the one a browser (and `fetch_url`) uses: `/api/v1/…`. The backend itself serves `/v1/…`; the edge strips `/api` (`services.md` §2.1). `fetch_url` sends no cookie, so every list and record case is also a no-session case.

**"One sentence".** The `error` value is a non-empty string that ends in a full stop, holds no line break and holds no SQL, host name, file path or stack text (`services.md` §2.3).

**"The envelope".** A non-2xx answer whose body is a JSON object with exactly the keys `error` (string), `code` (UPPER_SNAKE string) and, only when one parameter is at fault, `field` (string naming it, and the same name appears in the `error` sentence); `Content-Type: application/json; charset=utf-8` and `Cache-Control: no-store` (`services.md` §2.1, §2.3).

**"A page".** A 200 whose body is a JSON object with exactly the keys `items`, `total`, `page`, `limit` (no others) and `Cache-Control: no-store`.

**"The seeded order".** The order of Appendix A. `ids(x)` below means the list of `id` values of `x.items` in answer order.

**Instruments.** Each case names one of `fetch_url`, `render_url`, `measure_styles`, `interact_page`, `manual`. Where a case needs a state the application will not produce on its own, the *Given* says how it is arranged (`set_db_state` for a row or a renamed table, `set_db_available` for a database that refuses connections, `set_backend_state` for an API that is down); the *Then* is still read with the instrument named. **Every arrangement is undone before the next case** (rows deleted, table renamed back, database made available, backend started).

**`Check: manual` is used for exactly four kinds of case**, each with its reason in the case: (1) the hook's request terms (section G), which exist only in the story's own stubbed-`fetch` test under `src/test/` — the verifier proves them by running that suite (`run_tests`, frontend) and reading the diff; (2) the change's file list (SC-68), which is a property of the PR; (3) the 95th-percentile timing over 100 requests (SC-82), which no single-request instrument measures; (4) a migration that fails half-way (SC-9), which needs a deliberately broken migration file. Each is reported UNVERIFIED unless the verifier's other tools answer it.

## Appendix A — the seed, as the JSON files hold it (the baseline for every expected value)

**A1. The 14 bikes, in file order** (`products.json`). All 14 have `status` `live` and no other field is empty except where "—".

| # | id | name | brand | category | wheel | price | was | stock | deal | month | sizes (out) |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | `talon-2` | Talon 2 | Giant | hardtail | 29 | 18990000 | — | 9 | true | — | XS S M L XL () |
| 2 | `talon-29-3` | Talon 29 3 | Giant | hardtail | 29 | 13990000 | — | 6 | false | — | S M () |
| 3 | `talon-29-1` | Talon 29 1 | Giant | hardtail | 29 | 20490000 | 23990000 | 4 | false | — | S M L XL (XL) |
| 4 | `tempt-29-0` | Tempt 29 0 | Liv | hardtail | 29 | 28990000 | — | 3 | false | — | XS S M L () |
| 5 | `anthem-x-advanced-sl-3` | Anthem X Advanced SL 3 | Giant | cross-country | 29 | 100990000 | 125990000 | 2 | false | — | S M L XL (S) |
| 6 | `anthem-advanced-sl-0` | Anthem Advanced SL 0 | Giant | cross-country | 29 | 254990000 | — | 1 | false | — | S M L XL (S XL) |
| 7 | `pique-advanced-sl-se` | Pique Advanced SL SE | Liv | cross-country | 29 | 322990000 | — | 0 | false | — | XS S M L (all four) |
| 8 | `stance-0` | Stance 0 | Giant | trail | 29 | 66490000 | — | 7 | true | — | XS S M L XL () |
| 9 | `stance` | Stance | Giant | trail | 29 | 55990000 | — | 5 | false | — | S M L XL (L) |
| 10 | `trance-x-1` | Trance X 1 | Giant | trail | 29 | 93490000 | — | 3 | false | true | S M L XL (XL) |
| 11 | `intrigue-x-2` | Intrigue X 2 | Liv | trail | mixed | 76490000 | — | 4 | false | — | XS S M L () |
| 12 | `talon-e-plus-1` | Talon E+ 1 | Giant | e-mtb | 29 | 84990000 | — | 5 | true | — | S M L XL () |
| 13 | `tempt-e-plus-1` | Tempt E+ 1 | Liv | e-mtb | mixed | 82990000 | — | 0 | false | — | XS S M (all three) |
| 14 | `embolden-e-plus-0` | Embolden E+ 0 | Liv | e-mtb | 29 | 152990000 | — | 2 | false | — | XS S M L (XS) |

Derived sets used below: riding type `hardtail` = #1–4; `cross-country` = #5–7; `trail` = #8–11; `e-mtb` = #12–14. Giant = #1, 2, 3, 5, 6, 8, 9, 10, 12 (9 bikes); Liv = #4, 7, 11, 13, 14 (5 bikes). `wheel` 29 = every bike except #11 and #13 (12 bikes); `mixed` = #11, #13. `deal` true = #1, #8, #12; `deal` false = the other 11. `month` true = #10 only. `was` present = #3 and #5 only. `stock` 0 = #7 and #13. Sizes holding `XS` = #1, 4, 7, 8, 11, 13, 14 (7 bikes).

**A2. The other seven collections, in file order.**

| Collection | Count | Ids in order |
|---|---|---|
| `categories` | 4 | `hardtail`, `cross-country`, `trail`, `e-mtb` |
| `faqs` | 8 | `size`, `first-bike`, `test`, `delivery`, `pay`, `confirm`, `warranty`, `service` |
| `gallery` | 7 | `pass-climb`, `rocky-descent`, `forest-singletrack`, `group-ride`, `dirt-jump`, `sunset-rest`, `ridge-view` |
| `reviews` | 12 | `r-minh`, `r-lan`, `r-huy`, `r-thao`, `r-quan`, `r-ha`, `r-tuan`, `r-phuc`, `r-trang`, `r-khanh`, `r-dung`, `r-vy` |
| `services` | 6 | `build`, `fit`, `service`, `fork`, `bleed`, `tubeless` |
| `shop` | 1 | `shop` |
| `team` | 4 | `quang`, `thu`, `bao`, `van` |

Reviews by date (all text, ISO): `r-minh` 2026-10-28 (bike `stance-0`), `r-lan` 2026-10-12, `r-huy` 2026-09-30, `r-thao` 2026-09-14 (bike `talon-e-plus-1`), `r-quan` 2026-08-22, `r-ha` 2026-08-03, `r-tuan` 2026-07-19, `r-phuc` 2026-06-27, `r-trang` 2026-06-08, `r-khanh` 2026-05-21, `r-dung` 2026-04-30, `r-vy` 2026-04-11 (bike `stance-0`).

**A3. Values that are easy to lose in a move** (used by the exactness cases). Each is written as record → field → value:
- workshop job `build` → field `price` → the number `0` (a real price, not "empty");
- bike `pique-advanced-sl-se` → field `stock` → the number `0`, and field `rating` → the number `5` (a number, not the text `"5"`);
- bike `talon-29-3` → field `deal` → the boolean `false` (present, not omitted);
- bike `talon-2` → first entry of the list `size_fit` → `155–166 cm` (an en dash, U+2013); field `fact` → `29 in · 100 mm fork · 1×9` (middle dot U+00B7, multiplication sign U+00D7);
- review `r-thao` → field `name` → `Thảo`; team members `bao` and `van` → field `name` → `Bảo` and `Vân` (precomposed letters);
- the shop record → field `payment` → the list of three strings `Bank transfer (VietQR)`, `Cash on delivery (in town)`, `Card or cash at the shop` in that order; field `delivery_fee` → `350000`; field `rating` → `4.9`.

## A. The records are stored, exactly (DATA-001)

**Scenario**: TC-01 — Every collection holds as many records as its JSON file
**Given**: The story's stack has been started on a database that held none of these tables (the project's own empty Postgres, migrations applied on boot).
**When**: Each collection is listed with `GET /api/v1/<collection>?limit=100`, for `products`, `categories`, `faqs`, `gallery`, `reviews`, `services`, `shop` and `team`.
**Then**: Each answer is a page with status 200 whose `total` is, respectively, 14, 4, 8, 7, 12, 6, 1 and 4, whose `limit` is 100 and whose `items` length equals `total`; each count equals the number of records in the collection's JSON file on the default branch (the file is read from `main`, where it still exists).
Traces: SC-1 (DATA-001 AC-1)
Check: fetch_url

**Scenario**: TC-02 — Every record equals its JSON record, key for key
**Given**: The seeded stack; the eight JSON files as they stand on `main`.
**When**: Each collection is listed with `limit=100` and every returned record is compared with the record of the same `id` in its JSON file.
**Then**: For all 56 records the two objects have the same set of keys (including `_source` and, where the file has it, `_invented`) and equal values (numbers equal as numbers, booleans as booleans, lists equal element by element in the same order, text equal character for character). No record has a key its file lacks, and no file key is missing from the record. Key order is not compared.
Traces: SC-2 (DATA-001 AC-2)
Check: fetch_url

**Scenario**: TC-03 — Each collection lists its ids in the order of its file
**Given**: The seeded stack.
**When**: Each of the eight collections is listed with `limit=100` and no `sort`.
**Then**: `ids(products)` is the 14 ids of Appendix A1 in table order; the other seven lists equal the id lists of Appendix A2 in the order shown.
Traces: SC-2 (DATA-001 AC-2, Behaviour 2)
Check: fetch_url

**Scenario**: TC-04 — No internal column appears in any record
**Given**: The seeded stack.
**When**: All eight collections are listed with `limit=100`, and `GET /api/v1/products/talon-2` and `GET /api/v1/categories/trail` are read.
**Then**: No record in any of these answers has the key `seq`, `version`, `created_at` or `updated_at` (a text search of each body for those four quoted keys finds none).
Traces: SC-2 (DATA-001 Behaviour 3; services.md §2.7)
Check: fetch_url

**Scenario**: TC-05 — Values the move could lose arrive unchanged
**Given**: The seeded stack.
**When**: `GET /api/v1/services/build`, `/api/v1/products/pique-advanced-sl-se`, `/api/v1/products/talon-29-3`, `/api/v1/products/talon-2`, `/api/v1/reviews/r-thao`, `/api/v1/team/bao`, `/api/v1/team/van` and `/api/v1/shop/shop` are read.
**Then**: Each value of Appendix A3 is returned exactly: the job `build` has the key `price` present with the number `0`; the bike `pique-advanced-sl-se` has `stock` the number `0` and `rating` the number `5`; the bike `talon-29-3` has the key `deal` present with the boolean `false`; the bike `talon-2` has `size_fit[0]` equal to `155–166 cm` with U+2013 and `fact` equal to `29 in · 100 mm fork · 1×9`; the names are `Thảo`, `Bảo` and `Vân` as precomposed letters; the shop record's `payment` is the three strings of A3 in that order, its `delivery_fee` is `350000` and its `rating` is `4.9`.
Traces: SC-2 (DATA-001 AC-2, Behaviour 3)
Check: fetch_url

**Scenario**: TC-06 — talon-2 keeps its types: number, number, boolean, ordered list, empty list
**Given**: The seeded stack.
**When**: `GET /api/v1/products/talon-2` is read.
**Then**: `price` is the JSON number `18990000` (not a string, no fraction); `rating` is the JSON number `4.8`; `deal` is the JSON boolean `true`; `sizes` is the list `["XS","S","M","L","XL"]` in that order (five entries); `sizes_out` is `[]` — present and empty, not omitted and not `null`.
Traces: SC-3 (DATA-001 AC-3)
Check: fetch_url

**Scenario**: TC-07 — A field a record never had is left out, and nothing is ever null
**Given**: The seeded stack.
**When**: `GET /api/v1/products/talon-2` is read, and then all eight collections are listed with `limit=100` and their bodies searched for the JSON literal `null`.
**Then**: The `talon-2` record has no `was` key and no `month` key. The literal `null` appears nowhere in any of the eight bodies. `talon-29-1` has `was` equal to the number `23990000` and `anthem-x-advanced-sl-3` has `was` equal to the number `125990000`; no other bike has a `was` key.
Traces: SC-4 (DATA-001 AC-4)
Check: fetch_url

**Scenario**: TC-08 — With no ordering asked the bikes list talon-2 first and embolden-e-plus-0 last
**Given**: The seeded stack.
**When**: `GET /api/v1/products?limit=100` is read.
**Then**: `items[0].id` is `talon-2` and `items[13].id` is `embolden-e-plus-0`; `ids` is the Appendix A1 order.
Traces: SC-5 (DATA-001 AC-5)
Check: fetch_url

**Scenario**: TC-09 — A bike whose riding type does not exist is refused and nothing is stored
**Given**: The seeded stack; the database is reached with `set_db_state`.
**When**: The statement `INSERT INTO products (id, name, status, category) VALUES ('tc-orphan', 'TC orphan', 'draft', 'no-such-type')` is run.
**Then**: The database refuses it with a foreign-key violation (SQLSTATE 23503) naming the missing riding type, and afterwards `GET /api/v1/products/tc-orphan` answers 404 `NOT_FOUND` and `GET /api/v1/products?limit=100` still answers `total` 14.
Traces: SC-6 (DATA-001 AC-6)
Check: fetch_url

**Scenario**: TC-10 — A review whose bike does not exist is refused and nothing is stored
**Given**: The seeded stack; `set_db_state`.
**When**: `INSERT INTO reviews (id, name, bike) VALUES ('tc-orphan', 'TC', 'no-such-bike')` is run.
**Then**: The database refuses it with a foreign-key violation; afterwards `GET /api/v1/reviews/tc-orphan` answers 404 and `GET /api/v1/reviews?limit=100` still answers `total` 12.
Traces: SC-6 (DATA-001 AC-6)
Check: fetch_url

**Scenario**: TC-11 — A gallery photograph whose riding type does not exist is refused and nothing is stored
**Given**: The seeded stack; `set_db_state`.
**When**: `INSERT INTO gallery (id, image, type) VALUES ('tc-orphan', '/images/none.jpg', 'no-such-type')` is run.
**Then**: The database refuses it with a foreign-key violation; afterwards `GET /api/v1/gallery/tc-orphan` answers 404 and `GET /api/v1/gallery?limit=100` still answers `total` 7.
Traces: SC-6 (DATA-001 AC-6)
Check: fetch_url

**Scenario**: TC-12 — Applying the migrations again changes nothing and duplicates nothing
**Given**: The seeded stack with the counts of TC-01. Arranged in two steps: (a) `set_backend_state` restarts the backend, which runs its migration step on boot a second time; (b) `set_db_state` deletes the `schema_migrations` row of `20261118000200_seed_collections`, and the backend is restarted again so the seed file runs a second time.
**When**: After each restart, the eight collections are listed with `limit=100`.
**Then**: After (a) and after (b) the backend has started (a `GET /healthz` answers 200), the totals are still 14, 4, 8, 7, 12, 6, 1 and 4 and every collection's `ids` equal Appendix A in order; no id appears twice. If step (b) cannot be arranged, the case is reported on (a) alone and says so.
Traces: SC-7 (DATA-001 AC-7)
Check: fetch_url

**Scenario**: TC-13 — The review of r-thao names talon-e-plus-1 and a text date
**Given**: The seeded stack.
**When**: `GET /api/v1/reviews/r-thao` is read.
**Then**: `bike` is the string `talon-e-plus-1` and `date` is the string `2026-09-14` (a JSON string, not a number and not a timestamp such as `2026-09-14T00:00:00Z`); `name` is `Thảo`; `rating` is the number `4`.
Traces: SC-8 (DATA-001 AC-8)
Check: fetch_url

**Scenario**: TC-14 — A migration that cannot complete leaves nothing half-applied; a broken seed reference names the missing record
**Given**: A scratch database (not the running story's) in which the collections migration has been applied, and a copy of the seed migration whose last `reviews` insert names the bike `no-such-bike`.
**When**: The broken seed migration is applied through the runner (`internal/migrate`: one transaction per file together with its `schema_migrations` row).
**Then**: The run stops with the database's foreign-key message, which names the constraint on `reviews.bike` and the missing id `no-such-bike`; afterwards the eight tables hold no row from that file (none of the seed's records are present) and `schema_migrations` has no row for it.
Traces: SC-9 (DATA-001 failure table: Migration fails, Reference missing)
Check: manual — needs a deliberately broken migration file applied to a scratch database, which none of the five instruments does. The verifier should still try `set_db_state` with the same statements inside one transaction (a failing insert followed by a read showing no row survives) and read `internal/migrate/migrate.go` for the one-transaction-per-file rule; whatever it cannot run is UNVERIFIED.

## B. A page at a time, in a stable order (DATA-002)

**Scenario**: TC-15 — Page 1 of size 12 holds 12 bikes, page 2 holds the other two
**Given**: The seeded stack (14 bikes).
**When**: `GET /api/v1/products?page=1&limit=12` and then `GET /api/v1/products?page=2&limit=12` are read.
**Then**: The first is a page with 12 items (Appendix A1 rows 1–12 in order), `total` 14, `page` 1, `limit` 12. The second is a page with 2 items (`tempt-e-plus-1`, `embolden-e-plus-0`), `total` 14, `page` 2, `limit` 12.
Traces: SC-10 (DATA-002 AC-1, AC-2)
Check: fetch_url

**Scenario**: TC-16 — Three pages of five hold every bike once, in the seeded order
**Given**: The seeded stack.
**When**: `GET /api/v1/products?page=1&limit=5`, `?page=2&limit=5` and `?page=3&limit=5` are read.
**Then**: The pages hold 5, 5 and 4 items; their ids are rows 1–5, rows 6–10 and rows 11–14 of Appendix A1; concatenated they are the 14 ids of A1 in order with none repeated and none missing; each answers `total` 14 and `limit` 5.
Traces: SC-11 (DATA-002 AC-3)
Check: fetch_url

**Scenario**: TC-17 — A page past the last is an empty 200 with the true total
**Given**: The seeded stack.
**When**: `GET /api/v1/products?page=9&limit=12` is read.
**Then**: The status is 200 (not 4xx), `items` is `[]`, `total` is 14, `page` is 9 and `limit` is 12.
Traces: SC-12 (DATA-002 AC-4)
Check: fetch_url

**Scenario**: TC-18 — With no page and no size the answer is page 1 of size 12
**Given**: The seeded stack.
**When**: `GET /api/v1/products` is read with no query.
**Then**: It is a page with `page` 1, `limit` 12, `total` 14 and 12 items.
Traces: SC-13 (DATA-002 AC-5)
Check: fetch_url

**Scenario**: TC-19 — A page size above 100 is answered as 100
**Given**: The seeded stack.
**When**: `GET /api/v1/products?limit=500`, `?limit=100` and `?limit=101` are read.
**Then**: All three answer 200 with `limit` 100; each holds all 14 items (never more than 100); `total` is 14. (A size of 100 is answered as 100 and a size of 101 is answered as 100.)
Traces: SC-14 (DATA-002 AC-6, boundary)
Check: fetch_url

**Scenario**: TC-20 — A page or size that is not a positive whole number is refused
**Given**: The seeded stack.
**When**: Each of these is read: `/api/v1/products?page=0`, `?limit=0`, `?limit=-1`, `?limit=abc`, `?limit=` (empty), `?page=` (empty), `?page=1.5`, `?page=1000000001`.
**Then**: Each answers 400 in the envelope with `code` `INVALID_QUERY`, one sentence naming the parameter, and `field` equal to `page` for the `page` requests and `limit` for the `limit` requests; the body has no `items` key.
Traces: SC-15 (DATA-002 AC-7; services.md §2.4 "Bad values")
Check: fetch_url

**Scenario**: TC-21 — A collection that does not exist is 404
**Given**: The seeded stack.
**When**: `GET /api/v1/nothing` is read; then `GET /api/v1/orders` and `GET /api/v1/Products` (capital P).
**Then**: Each answers 404 in the envelope with `code` `NOT_FOUND`; the first with `error` "That collection does not exist." (the contract's fixed sentence); none holds an `items` key.
Traces: SC-16 (DATA-002 AC-8; services.md §3 intro)
Check: fetch_url

**Scenario**: TC-22 — Reading needs no sign-in
**Given**: No session of any kind (`fetch_url` sends no cookie).
**When**: `GET /api/v1/products` is read.
**Then**: The status is 200 with a page of 12 items and `total` 14; no `Set-Cookie` header is sent.
Traces: SC-17 (DATA-002 AC-9)
Check: fetch_url

**Scenario**: TC-23 — The same list asked twice answers identically
**Given**: The seeded stack.
**When**: Each of `/api/v1/products?limit=100`, `/api/v1/products?sort=stock&limit=100` and `/api/v1/products?page=2&limit=5` is read twice, one after the other.
**Then**: For each pair the two bodies are equal (same `total`, `page`, `limit` and the same ids in the same order).
Traces: SC-18 (DATA-002 AC-10)
Check: fetch_url

**Scenario**: TC-24 — A refused database connection is a 503 with one generic sentence and no record
**Given**: The stack is running and `set_db_available` takes the database away (the backend stays up).
**When**: `GET /api/v1/products` and `GET /api/v1/products/stance-0` are read. Then `set_db_available` gives the database back.
**Then**: Both answer 503 in the envelope with `code` `UNAVAILABLE` and `error` "The service is not ready."; neither body holds `items`, an `id`, or any SQL, host or path text. After the database is given back, `GET /api/v1/products` answers 200 with `total` 14.
Traces: SC-19 (DATA-002 upstream failure; DATA-006 upstream failure)
Check: fetch_url

**Scenario**: TC-25 — A failed query is a 500 with one generic sentence and no record
**Given**: `set_db_state` renames the table with `ALTER TABLE products RENAME TO products_tc_off`; the backend stays up and the database still accepts connections.
**When**: `GET /api/v1/products` and `GET /api/v1/products/stance-0` are read. Then the table is renamed back.
**Then**: Both answer 500 in the envelope with `code` `INTERNAL` and `error` "Something went wrong. Please try again."; the bodies hold no `items`, no `id`, no table name, no "relation", no SQL. After the rename back, `GET /api/v1/products` answers 200 with `total` 14.
Traces: SC-19 (DATA-002 upstream failure; DATA-006 upstream failure; services.md §2.3 `INTERNAL`)
Check: fetch_url

## C. Narrow a list by a value, a set of values or a range (DATA-003)

**Scenario**: TC-26 — status=live keeps all 14 bikes
**Given**: The seeded stack.
**When**: `GET /api/v1/products?status=live` is read.
**Then**: `total` is 14.
Traces: SC-20 (DATA-003 AC-1)
Check: fetch_url

**Scenario**: TC-27 — deal=true keeps exactly the three bikes the home page leads with
**Given**: The seeded stack.
**When**: `GET /api/v1/products?deal=true` and `GET /api/v1/products?deal=false` are read.
**Then**: The first has `total` 3 and ids `talon-2`, `stance-0`, `talon-e-plus-1` in that order; the second has `total` 11 (the other bikes; page size 12 holds all of them).
Traces: SC-21 (DATA-003 AC-2)
Check: fetch_url

**Scenario**: TC-28 — Stock equal to 0 and stock above 0
**Given**: The seeded stack.
**When**: `GET /api/v1/products?stock=0`, `GET /api/v1/products?status=live&stock_from=1` and `GET /api/v1/products?stock_to=0` are read.
**Then**: The first has `total` 2 with ids `pique-advanced-sl-se` and `tempt-e-plus-1`; the second has `total` 12 and neither of those two ids is among its items; the third equals the first.
Traces: SC-22 (DATA-003 AC-3, AC-4)
Check: fetch_url

**Scenario**: TC-29 — A comma means "any of"
**Given**: The seeded stack.
**When**: `GET /api/v1/products?category=trail,e-mtb&limit=100` and `GET /api/v1/products?category=trail,,e-mtb&limit=100` are read.
**Then**: Both have `total` 7 and ids `stance-0`, `stance`, `trance-x-1`, `intrigue-x-2`, `talon-e-plus-1`, `tempt-e-plus-1`, `embolden-e-plus-0`; every item's `category` is `trail` or `e-mtb` and no `hardtail` or `cross-country` bike is present (the empty part between two commas is ignored).
Traces: SC-23 (DATA-003 AC-5; services.md §3.2)
Check: fetch_url

**Scenario**: TC-30 — The upper end of a price range is inclusive
**Given**: The seeded stack; `talon-2` is priced 18990000 and `talon-29-3` 13990000.
**When**: `GET /api/v1/products?price_to=18990000`, `?price_from=18990001`, `?price_from=18990000&price_to=18990000` and `?price=18990000` are read (all with `limit=100`).
**Then**: The first has `total` 2 and its ids are `talon-2` and `talon-29-3`; `talon-2` is among the items (inclusive upper end). The second has `total` 12 and does not hold `talon-2` (one đồng above). The third has `total` 1: `talon-2` (both ends inclusive, a figure exactly at a range end is included). The fourth (numeric equality) has `total` 1: `talon-2`.
Traces: SC-24 (DATA-003 AC-6; Boundary)
Check: fetch_url

**Scenario**: TC-31 — A text range on ISO dates keeps the reviews of September and October
**Given**: The seeded stack.
**When**: `GET /api/v1/reviews?date_from=2026-09-01&date_to=2026-10-31`, `?date_from=2026-09-14&date_to=2026-09-14` and `?date_to=2026-09-13&date_from=2026-09-01` are read.
**Then**: The first has `total` 4 with ids `r-minh`, `r-lan`, `r-huy`, `r-thao`. The second has `total` 1: `r-thao` (both ends inclusive). The third has `total` 0 and `items` `[]` with status 200.
Traces: SC-25 (DATA-003 AC-7)
Check: fetch_url

**Scenario**: TC-32 — An empty value filters nothing, on any collection
**Given**: The seeded stack.
**When**: `GET /api/v1/products?category=` is compared with `GET /api/v1/products`; `GET /api/v1/team?name=` with `GET /api/v1/team`; `GET /api/v1/reviews?date_from=&date_to=` with `GET /api/v1/reviews`.
**Then**: In each pair the status is 200 and the two bodies are equal (same `total`, `page`, `limit`, items in the same order): 14, 4 and 12 matches respectively.
Traces: SC-26 (DATA-003 AC-8, Behaviour 5)
Check: fetch_url

**Scenario**: TC-33 — Text conditions ignore capitals and accents and compare the whole value
**Given**: The seeded stack.
**When**: `GET /api/v1/products?category=Trail`, `?category=trail`, `?category=TRAIL`, `?brand=GIANT`, `?name=stance` and `?name=Stance%200` are read; and `GET /api/v1/reviews?name=thao`, `?name=THAO` and `?name=Th%E1%BA%A3o` (with the accent).
**Then**: The three `category` requests return the same four bikes (`stance-0`, `stance`, `trance-x-1`, `intrigue-x-2`); `brand=GIANT` returns the nine Giant bikes; `name=stance` returns only `stance` (equality, not a substring: `stance-0` is not returned); `name=Stance 0` returns only `stance-0`; the three `reviews` requests each return only `r-thao`.
Traces: SC-27 (DATA-003 AC-9, Behaviour 1)
Check: fetch_url

**Scenario**: TC-34 — Conditions on several fields all apply together
**Given**: The seeded stack.
**When**: `GET /api/v1/products?brand=Giant&wheel=29&category=trail` is read; and `GET /api/v1/products?brand=Liv&wheel=mixed` is read.
**Then**: The first has `total` 3 with ids `stance-0`, `stance`, `trance-x-1` (every item is Giant, wheel 29 and trail; `intrigue-x-2`, which is Liv and mixed, is absent). The second has `total` 2: `intrigue-x-2` and `tempt-e-plus-1`.
Traces: SC-28 (DATA-003 AC-10)
Check: fetch_url

**Scenario**: TC-35 — A value of the wrong type for its field is refused
**Given**: The seeded stack.
**When**: Each is read: `/api/v1/products?stock=abc`, `?stock=1e3`, `?deal=maybe`, `?deal=1`, `?deal_from=1`, `?deal_to=true`, `?price=cheap`, `?price_from=x`.
**Then**: Each answers 400 in the envelope with `code` `INVALID_QUERY` and one sentence; `field` is the parameter named in the request (`stock`, `stock`, `deal`, `deal`, `deal_from`, `deal_to`, `price`, `price_from`) and the sentence names it; no `items` key.
Traces: SC-29 (DATA-003 AC-11)
Check: fetch_url

**Scenario**: TC-36 — A field the collection does not have is refused, naming it
**Given**: The seeded stack.
**When**: Each is read: `/api/v1/products?colourful=red`, `?_source=assumed`, `?seq=1`, `?version=1`, `?created_at_from=2026-01-01`; and `/api/v1/team?price=1` (a bike field on the team).
**Then**: Each answers 400 in the envelope with `code` `INVALID_QUERY`, `field` equal to the parameter's name (`colourful`, `_source`, `seq`, `version`, `created_at_from`, `price`) and a sentence that names the same word; never a 200 that ignores the parameter.
Traces: SC-30 (DATA-003 AC-12; services.md §2.5 rule 1)
Check: fetch_url

**Scenario**: TC-37 — A parameter given twice is refused, naming it
**Given**: The seeded stack.
**When**: `GET /api/v1/products?category=trail&category=hardtail` is read.
**Then**: 400 in the envelope, `code` `INVALID_QUERY`, `field` `category`, one sentence naming `category`.
Traces: contract (GET /v1/{collection} — a parameter given twice, services.md §3.2)
Check: fetch_url

**Scenario**: TC-38 — The total counts every match, not the page
**Given**: The seeded stack.
**When**: `GET /api/v1/products?category=trail&limit=2&page=2` and `GET /api/v1/products?status=live&limit=5&page=3` are read.
**Then**: The first has 2 items (`trance-x-1`, `intrigue-x-2`), `total` 4, `page` 2, `limit` 2. The second has 4 items (rows 11–14), `total` 14, `page` 3, `limit` 5.
Traces: SC-31 (DATA-003 AC-13)
Check: fetch_url

**Scenario**: TC-39 — On the questions, q is the search words and not a filter on the field q
**Given**: The seeded stack; the `faqs` records have a field named `q` (the question text).
**When**: `GET /api/v1/faqs?q=free` is read.
**Then**: 200 with `total` 2 and ids `delivery`, `service` — the records whose text holds the word "free" — and not a 400 and not an empty list (which a filter on the field `q` equal to "free" would give, as no question is exactly "free").
Traces: SC-32 (DATA-003 AC-14, Behaviour 6)
Check: fetch_url

**Scenario**: TC-40 — A list field matches a record whose list contains the value; a missing value never matches; a range on a list is refused
**Given**: The seeded stack.
**When**: `GET /api/v1/products?sizes=XS&limit=100`, `?sizes=xs&limit=100`, `?was_from=0`, `?id=talon-2,stance-0`, `?sizes_from=M` and `?sizes_to=M` are read.
**Then**: The first two have `total` 7 and ids `talon-2`, `tempt-29-0`, `pique-advanced-sl-se`, `stance-0`, `intrigue-x-2`, `tempt-e-plus-1`, `embolden-e-plus-0`. `was_from=0` has `total` 2: `talon-29-1` and `anthem-x-advanced-sl-3` (the 12 bikes without a `was` do not match). `id=talon-2,stance-0` has `total` 2 in the order `talon-2`, `stance-0`. The last two answer 400, `code` `INVALID_QUERY`, `field` `sizes_from` and `sizes_to`.
Traces: SC-33 (DATA-003 Behaviour 1, 3 and 5; failure table "No value")
Check: fetch_url

## D. Search the words of a collection, ignoring capitals and accents (DATA-004)

**Scenario**: TC-41 — Searching "thao" in any capitals finds the review by "Thảo"
**Given**: The seeded stack.
**When**: `GET /api/v1/reviews?q=thao`, `?q=THAO`, `?q=Th%E1%BA%A3o` (with the accent) and `?q=%20thao%20` (padded with spaces) are read.
**Then**: Each is 200 with `total` 1 and `items` holding only `r-thao`.
Traces: SC-34 (DATA-004 AC-1)
Check: fetch_url

**Scenario**: TC-42 — "cua lo" finds "Cửa Lò" and "deo" finds "Đèo"
**Given**: The seeded stack, which holds neither phrase. `set_db_state` inserts two test-only questions: `INSERT INTO faqs (id, q, a) VALUES ('tc-cualo', 'Do you ride to Cửa Lò?', 'Yes.'), ('tc-deo', 'Which bike for the Đèo pass?', 'A trail bike.')`.
**When**: `GET /api/v1/faqs?q=cua%20lo`, `?q=C%E1%BB%ACA%20L%C3%92` (the capitalised, accented words "CỬA LÒ"), `?q=deo`, `?q=%C4%90%C3%A8o` (the word "Đèo") and `?q=%C4%91%C3%A8o` ("đèo") are read. Afterwards both test rows are deleted with `DELETE FROM faqs WHERE id IN ('tc-cualo','tc-deo')`.
**Then**: Both `cua lo` requests (plain and accented) have `tc-cualo` among their items and `tc-deo` not; each `deo`/`Đèo`/`đèo` request has `tc-deo` among its items (it may also return `delivery`, whose text holds "video call", which contains "deo"). After the deletion `GET /api/v1/faqs` answers `total` 8 again.
Traces: SC-35 (DATA-004 AC-2, AC-3)
Check: fetch_url

**Scenario**: TC-43 — "free" finds two questions, and only the question field finds one
**Given**: The seeded stack.
**When**: `GET /api/v1/faqs?q=free`, `?q=free&search=q`, `?q=free&search=a` and `?q=free&search=q,a` are read.
**Then**: The first has `total` 2 with ids `delivery`, `service` in that order. The second has `total` 1: `service` (only its question, "Is the first service free?", holds the word). The third has `total` 2 (`delivery` and `service` answers hold "free"). The fourth has `total` 2.
Traces: SC-36 (DATA-004 AC-4, AC-5)
Check: fetch_url

**Scenario**: TC-44 — q combines with a condition: only live bikes that hold "stance"
**Given**: The seeded stack.
**When**: `GET /api/v1/products?q=stance&status=live&limit=100` and `GET /api/v1/products?q=stance&status=draft&limit=100` are read.
**Then**: The first is 200; `stance-0` and `stance` are among its items; every item has `status` `live` and holds the letters "stance" (ignoring capitals) in at least one text-like field (`id`, a text field, an image path); `total` equals the number of items. The second has `total` 0 and `items` `[]`.
Traces: SC-37 (DATA-004 AC-6)
Check: fetch_url

**Scenario**: TC-45 — Percent, underscore and backslash in the words mean themselves
**Given**: The seeded stack; no bike holds the text "100%"; no question holds an underscore; the question `delivery` and the record `shop` both hold "90% built". (Whether some bike's text holds a lone underscore or backslash is not assumed: those two requests are judged by what each returned item holds, below.)
**When**: `GET /api/v1/products?q=100%25`, `GET /api/v1/products?q=_&limit=100`, `GET /api/v1/products?q=%5C&limit=100`, `GET /api/v1/faqs?q=90%25`, `GET /api/v1/shop?q=90%25` and `GET /api/v1/faqs?q=_` are read.
**Then**: Each is 200 (no 4xx, no 5xx). `products?q=100%` has `total` 0 and `items` `[]` — it does not return every bike. `products?q=_` and `products?q=%5C` return only bikes that hold a literal underscore (respectively backslash) in a searched text field, and `total` is not 14 (so `_` is not "any one character", and `\` is not an escape). `faqs?q=90%` has `total` 1: `delivery`. `shop?q=90%` has `total` 1: `shop`. `faqs?q=_` has `total` 0.
Traces: SC-38 (DATA-004 AC-7; SRS §6 Security)
Check: fetch_url

**Scenario**: TC-46 — Empty or blank words filter nothing
**Given**: The seeded stack.
**When**: `GET /api/v1/products?q=`, `?q=%20%20%20` and no `q` at all are read, then `GET /api/v1/team?q=` and `GET /api/v1/team`.
**Then**: The three product answers are equal (200, `total` 14, 12 items, same order); the two team answers are equal (`total` 4).
Traces: SC-39 (DATA-004 AC-8)
Check: fetch_url

**Scenario**: TC-47 — Words of 100 characters are accepted, 101 are refused
**Given**: The seeded stack. Let `a100` be the letter "a" repeated 100 times and `a101` the same 101 times; `d100` be "đ" repeated 100 times (200 bytes) and `d101` 101 times.
**When**: `GET /api/v1/products?q=<a100>`, `?q=<a101>`, `?q=<d100>`, `?q=<d101>` and `?q=<a100>%20%20%20%20%20` (100 characters then five spaces) are read, with the text percent-encoded.
**Then**: `a100`, `d100` and the padded one answer 200 with `total` 0 (accepted; the length is counted in characters, after trimming, not in bytes). `a101` and `d101` answer 400 in the envelope with `code` `INVALID_QUERY`, `field` `q` and one sentence naming `q`; no `items` key.
Traces: SC-40 (DATA-004 AC-9; Boundary)
Check: fetch_url

**Scenario**: TC-48 — A search naming a field the collection lacks, or a field that is not text, is refused
**Given**: The seeded stack.
**When**: `GET /api/v1/faqs?q=free&search=nothing`, `GET /api/v1/products?q=a&search=price`, `?q=a&search=sizes`, `?q=a&search=deal`, `?q=a&search=_source` and `GET /api/v1/products?q=a&search=name,nothing` are read.
**Then**: Each answers 400 in the envelope with `code` `INVALID_QUERY`; `field` is the field at fault (`nothing`, `price`, `sizes`, `deal`, `_source`, `nothing`) and the sentence names it (for a number, list or boolean field, that it cannot be searched).
Traces: SC-41 (DATA-004 AC-10; services.md §3.2)
Check: fetch_url

**Scenario**: TC-49 — A page size does not change the count of matches
**Given**: The seeded stack.
**When**: `GET /api/v1/products?q=Stance&limit=100` and `GET /api/v1/products?q=Stance&limit=1` are read.
**Then**: The second has exactly 1 item and `limit` 1; its `total` equals the `total` of the first and the first's `items` length.
Traces: SC-42 (DATA-004 AC-11)
Check: fetch_url

**Scenario**: TC-50 — Numbers and underscore fields are never searched
**Given**: The seeded stack; `talon-2` is priced 18990000, `talon-29-3`'s `_source` holds "xedap.vn", and no text field of any bike holds either word.
**When**: `GET /api/v1/products?q=20990000`, `?q=18990000`, `?q=xedap` and `?q=assumed` (a `_source` word of the other collections) are read, and `GET /api/v1/categories?q=assumed`.
**Then**: Each has status 200, `total` 0 and `items` `[]`: no bike is found by its price (numbers are not searched), and no record is found by the contents of `_source`.
Traces: SC-43 (DATA-004 AC-12, Behaviour 2)
Check: fetch_url

**Scenario**: TC-51 — A SQL fragment in the words is plain text and harms nothing
**Given**: The seeded stack.
**When**: `GET /api/v1/products?q=%27%3B%20drop%20table%20products%3B%20--` (the text `'; drop table products; --`) is read; afterwards all eight collections are listed with `limit=100`.
**Then**: The first answer is 200 with `items` `[]` and `total` 0 (not an error); afterwards the totals are still 14, 4, 8, 7, 12, 6, 1 and 4.
Traces: SC-44 (DATA-004 AC-13; SRS §6 Security)
Check: fetch_url

**Scenario**: TC-52 — Words, ordering and paging combine; total counts every match
**Given**: The seeded stack.
**When**: `GET /api/v1/faqs?q=free&sort=-q&limit=1&page=2` and `GET /api/v1/faqs?q=free&sort=-q&limit=1&page=1` are read.
**Then**: The first has `items` `[delivery]` (one item), `total` 2, `page` 2, `limit` 1; the second has `items` `[service]`, `total` 2 (descending by the folded question: "is the first service free?" before "do you deliver…").
Traces: SC-34 to SC-44 (DATA-004 Behaviour 3)
Check: fetch_url

## E. Order a list by a field (DATA-005)

**Scenario**: TC-53 — sort=price and sort=-price order the bikes by number
**Given**: The seeded stack.
**When**: `GET /api/v1/products?sort=price&limit=100` and `GET /api/v1/products?sort=-price&limit=100` are read.
**Then**: Ascending, ids are `talon-29-3`, `talon-2`, `talon-29-1`, `tempt-29-0`, `stance`, `stance-0`, `intrigue-x-2`, `tempt-e-plus-1`, `talon-e-plus-1`, `trance-x-1`, `anthem-x-advanced-sl-3`, `embolden-e-plus-0`, `anthem-advanced-sl-0`, `pique-advanced-sl-se`; each price is at or above the one before and the first (13990000) is the lowest of the 14. Descending is exactly the reverse list; the first (`pique-advanced-sl-se`, 322990000) is the highest. `total` is 14 in both.
Traces: SC-45 (DATA-005 AC-1, AC-2)
Check: fetch_url

**Scenario**: TC-54 — Prices are ordered as numbers, not as text
**Given**: The seeded stack, in which a text sort would put `anthem-x-advanced-sl-3` (100990000) before `talon-29-3` (13990000) because "1000…" < "139…". And, as a second step, `set_db_state` inserts `INSERT INTO products (id, name, status, price) VALUES ('tc-cheap', 'TC cheap', 'draft', 9990000)`.
**When**: `GET /api/v1/products?sort=price&limit=100` is read before and after the insert; afterwards `DELETE FROM products WHERE id = 'tc-cheap'`.
**Then**: Before the insert `talon-29-3` (13990000) comes before `anthem-x-advanced-sl-3` (100990000). After it, `tc-cheap` (9990000) comes first and `talon-29-3` (13990000) second (a text sort would put "13990000" first); `total` is 15. After the delete `total` is 14 again.
Traces: SC-46 (DATA-005 AC-3)
Check: fetch_url

**Scenario**: TC-55 — Names sort in natural order, ignoring capitals; digits compare as numbers
**Given**: The seeded stack.
**When**: `GET /api/v1/products?sort=name&limit=100` and `?sort=-name&limit=100` are read.
**Then**: Ascending ids are `anthem-advanced-sl-0`, `anthem-x-advanced-sl-3`, `embolden-e-plus-0`, `intrigue-x-2`, `pique-advanced-sl-se`, `stance`, `stance-0`, `talon-2`, `talon-29-1`, `talon-29-3`, `talon-e-plus-1`, `tempt-29-0`, `tempt-e-plus-1`, `trance-x-1` — in particular `Talon 2`, `Talon 29 1`, `Talon 29 3` in that order, and `Stance` before `Stance 0`. Descending is the exact reverse.
Traces: SC-47 (DATA-005 AC-4)
Check: fetch_url

**Scenario**: TC-56 — Equal values keep the seeded order, and a bike with no value comes last in either direction
**Given**: The seeded stack. Then `set_db_state` inserts `INSERT INTO products (id, name, status) VALUES ('tc-nostock', 'TC no stock', 'draft')` (a bike with no `stock`), which is deleted at the end.
**When**: `GET /api/v1/products?sort=stock&limit=100` and `?sort=-stock&limit=100` are read, before the insert and after it.
**Then**: Before the insert, ascending ids are `pique-advanced-sl-se`, `tempt-e-plus-1` (stock 0), `anthem-advanced-sl-0` (1), `anthem-x-advanced-sl-3`, `embolden-e-plus-0` (2), `tempt-29-0`, `trance-x-1` (3), `talon-29-1`, `intrigue-x-2` (4), `stance`, `talon-e-plus-1` (5), `talon-29-3` (6), `stance-0` (7), `talon-2` (9); descending ids are `talon-2`, `stance-0`, `talon-29-3`, `stance`, `talon-e-plus-1`, `talon-29-1`, `intrigue-x-2`, `tempt-29-0`, `trance-x-1`, `anthem-x-advanced-sl-3`, `embolden-e-plus-0`, `anthem-advanced-sl-0`, `pique-advanced-sl-se`, `tempt-e-plus-1` (equal stocks in seeded order in both directions). After the insert, `tc-nostock` is the last item of both lists and the other 14 keep the order above; `total` is 15. After the delete, `total` is 14.
Traces: SC-48 (DATA-005 AC-5, AC-8)
Check: fetch_url

**Scenario**: TC-57 — Ordering by a field the collection lacks, or by a list, is refused
**Given**: The seeded stack.
**When**: `GET /api/v1/products?sort=colourful`, `?sort=-colourful`, `?sort=sizes`, `?sort=_source` and `GET /api/v1/faqs?sort=price` are read.
**Then**: Each answers 400 in the envelope with `code` `INVALID_QUERY`; `field` is the name sorted by (`colourful`, `colourful`, `sizes`, `_source`, `price`) and the sentence names it. (`sort=` with an empty value is not an error and leaves the seeded order.)
Traces: SC-49 (DATA-005 AC-6)
Check: fetch_url

**Scenario**: TC-58 — Ordering combines with a filter and a page; total is the same with or without it
**Given**: The seeded stack.
**When**: `GET /api/v1/products?sort=price&category=trail&page=2&limit=2`, `GET /api/v1/products?category=trail&limit=100&sort=price` and `GET /api/v1/products?category=trail&limit=100` are read; and `GET /api/v1/products?sort=price` is compared with `GET /api/v1/products` on `total`; and `GET /api/v1/products?sort=deal&limit=100` is read.
**Then**: The first has `items` `intrigue-x-2`, `trance-x-1` (the second page of the trail bikes by price: `stance` 55990000, `stance-0` 66490000, then these two), `total` 4, `page` 2, `limit` 2. The second and third both have `total` 4. `sort=price` and no sort both have `total` 14. `sort=deal` lists the 11 bikes with `deal` false first in the seeded order, then `talon-2`, `stance-0`, `talon-e-plus-1` (false before true, services.md §3.2).
Traces: SC-50 (DATA-005 AC-7, Behaviour 3)
Check: fetch_url

## F. One record by its id (DATA-006)

**Scenario**: TC-59 — A record read by id equals the record the list holds, with no session
**Given**: The seeded stack, no session.
**When**: `GET /api/v1/products/stance-0` is read and compared with the `stance-0` item of `GET /api/v1/products?limit=100`.
**Then**: 200; the two objects are equal key for key; the body holds `id` `stance-0`, `name` `Stance 0`, `price` the number `66490000` and no `version` key.
Traces: SC-51 (DATA-006 AC-1, AC-4)
Check: fetch_url

**Scenario**: TC-60 — A bike and a riding type carry an ETag; a review does not
**Given**: The seeded stack.
**When**: `GET /api/v1/products/stance-0`, `GET /api/v1/categories/trail` and `GET /api/v1/reviews/r-thao` are read, headers included.
**Then**: The first two answers carry an `ETag` header holding a quoted whole number (for example `"1"`: a pair of double quotes around digits only) and their bodies hold no `version` key. The review answer has no `ETag` header.
Traces: SC-51 (services.md §3.3)
Check: fetch_url

**Scenario**: TC-61 — An id no record has is a 404 with one sentence
**Given**: The seeded stack.
**When**: `GET /api/v1/products/no-such-bike` and `GET /api/v1/products/Stance-0` (capitals) are read.
**Then**: Both answer 404 in the envelope with `code` `NOT_FOUND` and `error` "That record does not exist." (the contract's sentence); neither has a `field` key; neither body holds a record.
Traces: SC-52 (DATA-006 AC-2)
Check: fetch_url

**Scenario**: TC-62 — A record of a collection that does not exist is a 404
**Given**: The seeded stack.
**When**: `GET /api/v1/nothing/stance-0` is read.
**Then**: 404 in the envelope with `code` `NOT_FOUND` and `error` "That collection does not exist.".
Traces: SC-53 (DATA-006 AC-3)
Check: fetch_url

## G. The data module reads the API (DATA-007)

These cases are about the module `src/data/index.ts` itself. Its request terms and its internal state (`loading`, `error`, which answer is kept) exist only inside the browser's JavaScript and are not visible in any page, so the story proves them with its own tests under `src/test/` that stub `fetch`: each case below is written as that test's scenario. The verifier runs that suite (`run_tests`, frontend) and reads that each scenario below is asserted by it; a scenario the suite does not assert is a gap in the suite and is reported as such. Where a page shows the same behaviour, a browser case in section H covers it as well (TC-83, TC-85, TC-89, TC-94). In every scenario "the stub" is a replacement `fetch` that records each URL and answers `{ "items": […], "total": n, "page": p, "limit": l }` built from a fixture of the 14 bikes of Appendix A1 (a page of the fixture for the `page` and `limit` it was asked), and "the fixture" is that list.

**Scenario**: TC-63 — With VITE_API_URL unset the first request is /api/v1/products?page=1&limit=3
**Given**: `VITE_API_URL` is not set; the stub is installed.
**When**: `useCollection("products", { pageSize: 3 })` mounts.
**Then**: The first request the stub sees is exactly `/api/v1/products?page=1&limit=3`.
Traces: SC-54 (DATA-007 AC-1)
Check: manual — module-level request term; proved by the story's stubbed-`fetch` test (see section G intro).

**Scenario**: TC-64 — With VITE_API_URL set the request goes to that base
**Given**: `VITE_API_URL` is `https://x.test`; the stub is installed.
**When**: `useCollection("products", { pageSize: 3 })` mounts.
**Then**: The first request is exactly `https://x.test/v1/products?page=1&limit=3` (no `/api` in it).
Traces: SC-54 (DATA-007 AC-2)
Check: manual — as TC-63.

**Scenario**: TC-65 — loadMore appends the next page; hasMore is true while fewer are held than total
**Given**: The fixture of 14 bikes; the stub; page size 5, no `page` in the query.
**When**: The hook mounts and settles, then `loadMore()` is called once, then twice more.
**Then**: After mount `items` holds 5 (rows 1–5) and `hasMore` is true. After the first `loadMore()` the second request carries `page=2`, `items` holds 10 (rows 1–10 in order, none repeated) and `hasMore` is true. After the second `loadMore()` the third request carries `page=3`, `items` holds 14 and `hasMore` is false; a further `loadMore()` leaves `items` at 14 with no row repeated and `hasMore` false. `total` is 14 throughout.
Traces: SC-55 (DATA-007 AC-3)
Check: manual — as TC-63.

**Scenario**: TC-66 — With a page, items is that page alone and pages is the number of pages
**Given**: The fixture; the stub; `page: 2`, `pageSize: 5`.
**When**: The hook mounts and settles.
**Then**: The request carries `page=2&limit=5`; `items` holds the five bikes of rows 6–10 and nothing from page 1; `pages` is 3; `page` is 2; `total` is 14.
Traces: SC-56 (DATA-007 AC-4)
Check: manual — as TC-63.

**Scenario**: TC-67 — A page past the last shows the last page
**Given**: The fixture; the stub (which answers page 7 of size 5 with `items` `[]`, `total` 14); `page: 7`, `pageSize: 5`.
**When**: The hook mounts and settles.
**Then**: `page` is 3 and `items` is the last page (rows 11–14, four bikes); `pages` is 3.
Traces: SC-57 (DATA-007 AC-5)
Check: manual — as TC-63.

**Scenario**: TC-68 — Conditions are sent in the API's own terms, and an empty end is left out
**Given**: The stub; `where: { status: "live", category: ["trail","e-mtb"], stock: { from: 1 }, price: { from: "", to: 70000000 }, deal: false }`.
**When**: The hook mounts.
**Then**: The request carries `status=live`, `category=trail,e-mtb` (the comma may be written `,` or `%2C`), `stock_from=1`, `price_to=70000000` and `deal=false` (the boolean `false` is sent, not dropped), plus `page=1&limit=12`; it carries no `price_from` and no `price`.
Traces: SC-58 (DATA-007 AC-6)
Check: manual — as TC-63. (Page-level view: TC-83.)

**Scenario**: TC-69 — Search words, searched fields and ordering are sent in the API's own terms
**Given**: The stub; `q: "cua lo"`, `search: ["name","fact"]`, `sort: "-price"`.
**When**: The hook mounts.
**Then**: The request carries `q` equal to `cua lo` once decoded (written `cua%20lo`; a `+` for the space is the same on the wire), `search` equal to `name,fact` and `sort=-price`.
Traces: SC-59 (DATA-007 AC-7)
Check: manual — as TC-63. (Page-level view: TC-85.)

**Scenario**: TC-70 — A condition that is empty is not sent
**Given**: The stub; `where: { category: [], brand: "", month: undefined }`; and a second hook with `where: { price: { from: "", to: undefined }, stock: { } }` and `q: ""`.
**When**: Each hook mounts.
**Then**: Neither request carries `category`, `brand`, `month`, `price`, `price_from`, `price_to`, `stock`, `stock_from`, `stock_to` or `q`, with or without a value (not even `category=`); each is just `page=1&limit=12`.
Traces: SC-60 (DATA-007 AC-8)
Check: manual — as TC-63.

**Scenario**: TC-71 — A page size above 100 is never requested
**Given**: The stub; hooks with `pageSize` 500, 101 and 100, and one with `page: 1` and `pageSize: 500`.
**When**: Each mounts and calls `loadMore()` once.
**Then**: Every request of the four carries `limit=100`; no request anywhere in the test carries a `limit` above 100.
Traces: SC-61 (DATA-007 AC-9, Boundary)
Check: manual — as TC-63. (The server also clamps, so a page cannot tell the two apart; only the module's own request shows it.)

**Scenario**: TC-72 — loading is true from the request going out until it is answered
**Given**: The stub holds each answer until the test releases it.
**When**: The hook mounts and the first answer is held; the answer is then released; `loadMore()` is called and its answer is held, then released.
**Then**: `loading` is true while the first request is unanswered and false once it is answered; true again while the `loadMore()` request is unanswered and false after.
Traces: SC-62 (DATA-007 AC-10)
Check: manual — as TC-63.

**Scenario**: TC-73 — A failed request sets error and keeps the records; a later success clears it
**Given**: The stub answers page 1 successfully (5 bikes, page size 5), then answers the next request with status 500; in a second run the next request rejects as a network failure.
**When**: The hook mounts and settles; `loadMore()` is called (500, or the network failure); then `loadMore()` is called again and the stub now answers with success.
**Then**: After the failure `error` is an `Error` (`error instanceof Error`, non-empty `message`), `loading` is false and `items` still holds the 5 bikes it had; after the later successful answer `error` is `null` and `items` holds the bikes of both pages. Both failure kinds (answer 500, network failure) behave the same.
Traces: SC-63 (DATA-007 AC-11, Upstream failure)
Check: manual — as TC-63. (Page-level view: TC-94.)

**Scenario**: TC-74 — A late answer of an earlier query is discarded
**Given**: The stub holds query A's answer (rows of one riding type) and answers query B's at once.
**When**: The hook mounts with query A; its request is in flight; the query changes to B; B's answer arrives; then A's answer is released.
**Then**: After B's answer `items` and `total` are B's; after A's late answer they are still B's only — no record of A is in `items`, `total` is B's, `loading` is false and `error` is `null`.
Traces: SC-64 (DATA-007 AC-12, Stale answer)
Check: manual — as TC-63.

**Scenario**: TC-75 — useRecord: empty until answered, the record when it arrives, nothing and no error on a 404, an Error on any other failure
**Given**: The stub answers `GET /api/v1/products/stance-0` with the record (status 200), `…/no-such-bike` with 404 and the envelope, and a third id with 500; a fourth run rejects as a network failure.
**When**: `useRecord("products", "stance-0")` mounts, then `useRecord("products", "no-such-bike")`, then the 500 and network cases.
**Then**: Before the answer `item` is empty (`undefined`) and `loading` is true; after it `item` is the record (`id` `stance-0`, `price` 66490000) and `loading` is false. For the 404: `item` empty, `loading` false, `error` `null`. For the 500 and the network failure: `item` empty, `loading` false and `error` an `Error`.
Traces: SC-65 (DATA-007 AC-13, AC-14, Behaviour 6)
Check: manual — as TC-63. (Page-level view: TC-89.)

**Scenario**: TC-76 — Every export stays, and the project type-checks, lints and passes its tests with no page or component changed
**Given**: The story's PR head.
**When**: `src/data/index.ts` is read and `npm run build`, `npm run lint` and `npm test` run (`run_tests`, frontend, and CI).
**Then**: `useCollection`, `useRecord`, `schema`, `collections`, `setCollection` and the types `Row`, `Source`, `Field`, `Collection`, `Filter`, `Query` are exported; `useCollection` returns `items`, `total`, `pages`, `page`, `hasMore`, `loading`, `error`, `loadMore` and `useRecord` returns `item`, `loading`, `error`; the build, the lint and the tests pass; the existing `src/test/admin.test.ts` and `src/test/format.test.ts` pass unchanged, and `setCollection(name, rows)` makes `collections()[name]` equal `rows` at once.
Traces: SC-66 (DATA-007 AC-15; DATA-007 Behaviour 1, 8)
Check: manual — a property of the build and the diff, not of a running page; `run_tests` and `get_pr` answer it.

**Scenario**: TC-77 — When another module saves a record, lists and records on screen for that collection ask again
**Given**: The stub; one mounted `useCollection("products", …)`, one mounted `useRecord("products", "stance-0")` and one mounted `useCollection("categories", …)`; the stub's bike fixture then gains a new bike.
**When**: `setCollection("products", rows)` is called (the way the management area reports a save or delete).
**Then**: The products list and the bike record each send a new request, and the list now holds the new bike without being remounted; the categories list sends no new request. Separately, after a mounted `useCollection("categories", …)` has settled with the four riding types, `collections().categories` holds four rows each with its `id` and `name`.
Traces: SC-67 (DATA-007 AC-16, AC-17, Behaviour 7, 8)
Check: manual — as TC-63.

## H. The pages show the same records as before (DATA-008)

"Before the move" is the build on `main` before this story's frontend change, whose module reads the JSON files; the expected lists below were derived from the same files and the same page code (Appendix A) and are the baseline. The pages are read after the answers have settled (story §8.5).

**Scenario**: TC-78 — /shop with no filters: 14 bikes, 9 cards, "Show more bikes (5)"
**Given**: The running story; `/shop` with no query.
**When**: The page has loaded.
**Then**: `[data-testid="shop-count"]` reads "14" followed by "bikes"; nine links `a[data-testid^="card-"]` (excluding those whose `data-testid` starts `card-was-`) are drawn, in the order `talon-2`, `talon-29-3`, `talon-29-1`, `tempt-29-0`, `anthem-x-advanced-sl-3`, `anthem-advanced-sl-0`, `pique-advanced-sl-se`, `stance-0`, `stance`; `[data-testid="shop-more"]` reads "Show more bikes (5)". The four riding-type chips (`chip-hardtail`, `chip-cross-country`, `chip-trail`, `chip-e-mtb`) and `chip-all` are drawn.
Traces: SC-69 (DATA-008 AC-3)
Check: render_url

**Scenario**: TC-79 — Each riding-type chip lists the bikes it listed before, in the same order
**Given**: `/shop` is open.
**When**: `chip-hardtail`, `chip-cross-country`, `chip-trail` and `chip-e-mtb` are pressed one after another (the same chip is not pressed twice in a row).
**Then**: After each press `shop-count` and the card order read: hardtail 4 — `talon-2`, `talon-29-3`, `talon-29-1`, `tempt-29-0`; cross-country 3 — `anthem-x-advanced-sl-3`, `anthem-advanced-sl-0`, `pique-advanced-sl-se`; trail 4 — `stance-0`, `stance`, `trance-x-1`, `intrigue-x-2`; e-mtb 3 — `talon-e-plus-1`, `tempt-e-plus-1`, `embolden-e-plus-0`. The pressed chip has `aria-pressed` true. Pressing `chip-all` returns "14".
Traces: SC-70 (DATA-008 AC-4)
Check: interact_page

**Scenario**: TC-80 — The brand boxes and the wheel boxes list the bikes they listed before
**Given**: `/shop` is open.
**When**: In the filter rail, group "Brand", the "Giant" box is ticked and read; then "Giant" is cleared and "Liv" ticked and read; then both ticked; then both cleared and, in the group "Wheel size", the box "29 inch" ticked, then the box "Mixed 29 / 27.5 inch" instead; then "Liv" and the mixed box together.
**Then**: Giant: count 9, cards (nine at first, in this order) `talon-2`, `talon-29-3`, `talon-29-1`, `anthem-x-advanced-sl-3`, `anthem-advanced-sl-0`, `stance-0`, `stance`, `trance-x-1`, `talon-e-plus-1`. Liv: 5 — `tempt-29-0`, `pique-advanced-sl-se`, `intrigue-x-2`, `tempt-e-plus-1`, `embolden-e-plus-0`. Both: 14. Wheel 29: 12 (all but `intrigue-x-2` and `tempt-e-plus-1`). Mixed: 2 — `intrigue-x-2`, `tempt-e-plus-1`. Liv with mixed: 2, the same two.
Traces: SC-71 (DATA-008 AC-5)
Check: interact_page

**Scenario**: TC-81 — "Price: low to high" and "Price: high to low" order the bikes as before
**Given**: `/shop` is open with "Recommended" chosen.
**When**: The sort control is opened and "Price: low to high" is chosen and the cards are read, then "Price: high to low".
**Then**: Low to high: count 14; the first nine cards are `talon-29-3`, `talon-2`, `talon-29-1`, `tempt-29-0`, `stance`, `stance-0`, `intrigue-x-2`, `tempt-e-plus-1`, `talon-e-plus-1`, and after "Show more bikes (5)" the last five are `trance-x-1`, `anthem-x-advanced-sl-3`, `embolden-e-plus-0`, `anthem-advanced-sl-0`, `pique-advanced-sl-se`. High to low is the exact reverse (first card `pique-advanced-sl-se`, last `talon-29-3`). Choosing "Recommended" again restores the seeded order of TC-78.
Traces: SC-72 (DATA-008 AC-6)
Check: interact_page

**Scenario**: TC-82 — A frame size with a price band narrows the loaded 14, as before
**Given**: `/shop` is open.
**When**: In the group "Frame size" the box "M" and in the group "Price" the box "25–70 million ₫" are ticked and read; then both cleared and "XL" with "Over 100 million ₫" ticked and read; then "XS" with "Under 25 million ₫".
**Then**: M with 25–70: count 3 — `tempt-29-0`, `stance-0`, `stance` (that order; a price of exactly 25000000 or 70000000 belongs to the band that starts there — none of the 14 sits on an edge). XL with over 100: count 2 — `anthem-x-advanced-sl-3`, `anthem-advanced-sl-0`. XS with under 25: count 1 — `talon-2`. (Each is the intersection of the loaded 14 with the page's own size and band rules.)
Traces: SC-73 (DATA-008 AC-7)
Check: interact_page

**Scenario**: TC-83 — Several riding types, a brand and a wheel narrow together through the API
**Given**: The running story.
**When**: `/shop?category=trail,e-mtb&brand=Giant&wheel=29` is opened.
**Then**: `shop-count` reads "4"; the cards are `stance-0`, `stance`, `trance-x-1`, `talon-e-plus-1` in that order (a list of riding types is sent as one "any of" condition, the brand and wheel as values, and only live bikes are asked for).
Traces: SC-58 (DATA-007 AC-6, seen from the page); SC-23 (DATA-003 AC-5)
Check: render_url

**Scenario**: TC-84 — /search?q=size and /search?q=SIZE find the same results, including the frame-size question
**Given**: The running story.
**When**: `/search?q=size` and `/search?q=SIZE` are opened.
**Then**: Both show `[data-testid="search-count"]` reading "3" followed by "results"; the chips show Questions 3 and no chip for Bikes, Riding types, Workshop or Gallery; `[data-testid="search-faqs"]` lists three questions — the question "How do I know which frame size to choose?" (`faqs:size`), "Can I try a bike before I buy it?" (`faqs:test`) and "When is my order final?" (`faqs:confirm`) — in that order. The two pages are identical in counts and in the questions listed. (No bike, riding type, job or photograph holds the word "size" in the fields the page searches.)
Traces: SC-74 (DATA-008 AC-8)
Check: render_url

**Scenario**: TC-85 — /search?q=bleed lists "Brake bleed" in the Workshop block, whatever the capitals, as a phrase
**Given**: The running story.
**When**: `/search?q=bleed`, `/search?q=BLEED` and `/search?q=brake%20bleed` are opened.
**Then**: Each shows `[data-testid="search-block-services"]` with the heading "Workshop" and the count 1, and its table holds one row `[data-record="services:bleed"]` reading "Brake bleed", "30 min" and "250.000 ₫"; `search-count` reads "1". (The third proves a phrase with a space is searched as one phrase.)
Traces: SC-75 (DATA-008 AC-9); SC-59 (DATA-007 AC-7, seen from the page)
Check: render_url

**Scenario**: TC-86 — The search page's two other states: nothing typed, and nothing found
**Given**: The running story.
**When**: `/search` and `/search?q=zzqxv` are opened.
**Then**: `/search` shows no `search-count` and shows the riding types the start panel lists (`[data-testid="search-type-hardtail"]`, `…-cross-country`, `…-trail`, `…-e-mtb`, four links read from the API). `/search?q=zzqxv` shows `[data-testid="search-none"]` containing the typed words “zzqxv” and shows no `search-block-*` element and no `search-count`.
Traces: SC-74, SC-75 (DATA-008 AC-8, AC-9; SRS §5 Search states)
Check: render_url

**Scenario**: TC-87 — The home page: 12 in stock, three tinted cards, the bike of the month and six picks
**Given**: The running story.
**When**: `/` is opened.
**Then**: The proof line reads "12" followed by "bikes in stock" (`[data-count]` holds 12); the tinted cards are `[data-testid="tint-talon-2"]`, `tint-stance-0`, `tint-talon-e-plus-1` in that order; the bike of the month is `trance-x-1` (`[data-testid="month-cta"]` links to `/shop/trance-x-1`); the six picks (`a[data-testid^="card-"]` in the picks section) are, in order, `talon-29-3`, `talon-29-1`, `tempt-29-0`, `anthem-x-advanced-sl-3`, `anthem-advanced-sl-0`, `pique-advanced-sl-se` (the first six live bikes not marked `deal`); the four riding-type marks `mark-hardtail`, `mark-cross-country`, `mark-trail`, `mark-e-mtb` are drawn; three review cards are drawn.
Traces: SC-76 (DATA-008 AC-10)
Check: render_url

**Scenario**: TC-88 — /shop/stance-0: the Reviews tab holds its two reviews
**Given**: `/shop/stance-0` is open after the bike has arrived.
**When**: The "Reviews" tab is pressed.
**Then**: Exactly two review cards are shown, `[data-record="reviews:r-minh"]` ("Minh") and `[data-record="reviews:r-vy"]` ("Yến"); no card of another bike is shown (for example no `reviews:r-lan`).
Traces: SC-77 (DATA-008 AC-11)
Check: interact_page

**Scenario**: TC-89 — A bike page shows its record; an unknown bike shows the "no longer in the shop" state and no error
**Given**: The running story.
**When**: `/shop/stance-0` and `/shop/no-such-bike` are opened, each after its answer has settled.
**Then**: The first shows `[data-testid="product-name"]` reading "Stance 0" and `[data-testid="product-price"]` reading "66.490.000" with "₫". The second shows the heading "That bike is no longer in the shop." (the 404 leaves `item` empty) and no price and no JavaScript error on the page.
Traces: SC-65 (DATA-007 AC-13, AC-14, seen from the page); SC-77 (DATA-008 AC-11)
Check: render_url

**Scenario**: TC-90 — /about: four team cards, six workshop rows, six reviews with "Show all reviews (12)", eight questions
**Given**: The running story.
**When**: `/about` is opened, then `[data-testid="reviews-more"]` is pressed.
**Then**: Before the press: four `[data-record^="team:"]` cards (`quang`, `thu`, `bao`, `van`), six `[data-record^="services:"]` rows (`build`, `fit`, `service`, `fork`, `bleed`, `tubeless`), six `[data-record^="reviews:"]` cards, eight `[data-record^="faqs:"]` questions, and `reviews-more` reads "Show all reviews (12)". After the press twelve review cards are shown and the button is gone.
Traces: SC-78 (DATA-008 AC-12)
Check: interact_page

**Scenario**: TC-91 — /gallery reads "7 photographs"
**Given**: The running story.
**When**: `/gallery` is opened.
**Then**: The page reads "7" followed by "photographs" and draws seven `[data-record^="gallery:"]` figures, the first `gallery:pass-climb`.
Traces: SC-79 (DATA-008 AC-13)
Check: render_url

**Scenario**: TC-92 — The management lists show 14 bike rows and 4 riding-type rows
**Given**: The running story. The visitor is signed in through the sign-in the build already has: `/admin/sign-in`, where today any email is accepted (admin SRS §1) — `owner@deo.test` is typed and the form submitted. If no sign-in can be completed the case is UNVERIFIED.
**When**: `/admin/products` and `/admin/kinds` are opened.
**Then**: `/admin/products` shows 14 bike rows on one page (page size 20) and the range "1–14 of 14"; `/admin/kinds` shows 4 riding-type rows (Hardtail, Cross-country, Trail, E-MTB).
Traces: SC-80 (DATA-008 AC-14)
Check: interact_page

**Scenario**: TC-93 — The overview counts 14 bikes, 4 riding types, 14 running low, 2 out of stock
**Given**: Signed in as in TC-92.
**When**: `/admin` is opened and its four figures are read.
**Then**: "Bikes" 14, "Riding types" 4, "Running low" 14 and "Out of stock" 2 (the page asks for its bikes with a page size of 500, which the module sends as 100, and the answer holds the 14).
Traces: SC-81 (DATA-008 AC-15, Boundary)
Check: interact_page

**Scenario**: TC-94 — With the API out of reach the shop shows its empty list and recovers when the API returns
**Given**: `set_backend_state` stops the backend (the gateway answers 502).
**When**: `/shop` is opened; then the backend is started again and `chip-trail` is pressed on the same page; then `chip-all`.
**Then**: While the API is down the page draws its existing empty state — the text "No bike matches all of that." and `shop-clear` — with `shop-count` "0", no card and no uncaught JavaScript error (SRS §7, open question 1: no new error screen). After the backend is back, pressing `chip-trail` lists `stance-0`, `stance`, `trance-x-1`, `intrigue-x-2` (count 4) and `chip-all` shows "14": a later successful request clears the failure.
Traces: SC-63 (DATA-007 AC-11, Upstream failure; failure table "Upstream failure")
Check: interact_page

**Scenario**: TC-95 — The change touches only the data module, the deleted JSON files and new test files
**Given**: The story's frontend PR.
**When**: `code/frontend/src/data/` is listed on the PR head and the PR's file list is read.
**Then**: `src/data/` holds `schema.json` and `index.ts` only; the file list holds `src/data/index.ts`, the eight deleted files (`categories.json`, `faqs.json`, `gallery.json`, `products.json`, `reviews.json`, `services.json`, `shop.json`, `team.json`) and, at most, new files under `src/test/` — no page, component, hook, library, `schema.json`, `theme.css`, `index.css`, `package.json` or `eslint.config.js`.
Traces: SC-68 (DATA-008 AC-1, AC-2)
Check: manual — a property of the PR's file list; `get_pr` and `list_files` answer it.

## I. Performance (SRS §6)

**Scenario**: TC-96 — Listing 100 bikes, with and without q, stays within 500 ms at the 95th percentile
**Given**: The seeded database on the same host as the API.
**When**: `GET /api/v1/products?limit=100` and `GET /api/v1/products?limit=100&q=stance` are each requested 100 times in sequence.
**Then**: For each of the two series the 95th-percentile response time is at most 500 ms (and a single record, `GET /api/v1/products/stance-0`, answers within 100 ms at the 95th percentile, services.md §6).
Traces: SC-82 (SRS §6 Performance)
Check: manual — a percentile needs 100 sequential timed requests and no instrument loops or reports timings. The verifier samples with `fetch_url` and reports any timing it is given; without timings the case is UNVERIFIED.

## J. The service contract's own rules (GET /v1/{collection} and GET /v1/{collection}/{id})

**Scenario**: TC-97 — A success and an error both carry the contract's headers; the list's keys are exactly four
**Given**: The seeded stack.
**When**: `GET /api/v1/products`, `GET /api/v1/products/stance-0`, `GET /api/v1/products?limit=0` (400) and `GET /api/v1/products/no-such-bike` (404) are read, headers included.
**Then**: All four carry `Content-Type: application/json; charset=utf-8`, `Cache-Control: no-store` and an `X-Request-Id` header (32 hexadecimal characters when the caller sent none). The list body has exactly the keys `items`, `total`, `page`, `limit`. The two error bodies have exactly `error`, `code` (and `field` on the 400 only); `X-Request-Id` is not in any body.
Traces: contract (services.md §2.1, §2.3, §3.2)
Check: fetch_url

**Scenario**: TC-98 — A path that matches no route, or an exact path in the wrong shape, is a 404 in the envelope
**Given**: The seeded stack.
**When**: `GET /api/v1/products/` (trailing slash), `GET /api/v1/products/talon-2/extra`, `GET /api/v1/` and `GET /api/v1/PRODUCTS` are read.
**Then**: Each answers 404 in the envelope with `code` `NOT_FOUND`; the first three with `error` "That address does not exist." (the fourth names an unknown collection: "That collection does not exist."); none holds a record.
Traces: contract (services.md §2.1 "Paths are exact", §3 intro)
Check: fetch_url

**Scenario**: TC-99 — A write is not offered on these paths and changes nothing
**Given**: The seeded stack, no session. This story offers only `GET` (story §8.7); the contract answers any other method on an existing path 405 with an `Allow` header.
**When**: `POST /api/v1/products` (body `{"name":"TC write"}`), `PUT /api/v1/products/talon-2` (body `{"name":"TC write"}`) and `DELETE /api/v1/products/talon-2` are sent, then `GET /api/v1/products/talon-2` and `GET /api/v1/products?limit=100`.
**Then**: Each write answers 405 in the envelope with `code` `METHOD_NOT_ALLOWED` and an `Allow` header that lists `GET`; afterwards `talon-2` still has `name` `Talon 2` and `price` 18990000 and the list still has `total` 14. (If `fetch_url` cannot send a method other than GET, the case is UNVERIFIED. When the admin story lands, these three answer 401 for products and categories, and this case is then superseded.)
Traces: contract (services.md §2.3 `METHOD_NOT_ALLOWED`, §3 intro)
Check: fetch_url

## Coverage — every story criterion has at least one case

| Criterion | Cases | Criterion | Cases | Criterion | Cases |
|---|---|---|---|---|---|
| SC-1 | TC-01 | SC-29 | TC-35 | SC-57 | TC-67 |
| SC-2 | TC-02, TC-03, TC-04, TC-05 | SC-30 | TC-36 | SC-58 | TC-68, TC-83 |
| SC-3 | TC-06 | SC-31 | TC-38 | SC-59 | TC-69, TC-85 |
| SC-4 | TC-07 | SC-32 | TC-39 | SC-60 | TC-70 |
| SC-5 | TC-08 | SC-33 | TC-40 | SC-61 | TC-71 |
| SC-6 | TC-09, TC-10, TC-11 | SC-34 | TC-41, TC-52 | SC-62 | TC-72 |
| SC-7 | TC-12 | SC-35 | TC-42 | SC-63 | TC-73, TC-94 |
| SC-8 | TC-13 | SC-36 | TC-43 | SC-64 | TC-74 |
| SC-9 | TC-14 | SC-37 | TC-44 | SC-65 | TC-75, TC-89 |
| SC-10 | TC-15 | SC-38 | TC-45 | SC-66 | TC-76 |
| SC-11 | TC-16 | SC-39 | TC-46 | SC-67 | TC-77 |
| SC-12 | TC-17 | SC-40 | TC-47 | SC-68 | TC-95 |
| SC-13 | TC-18 | SC-41 | TC-48 | SC-69 | TC-78 |
| SC-14 | TC-19 | SC-42 | TC-49 | SC-70 | TC-79 |
| SC-15 | TC-20 | SC-43 | TC-50 | SC-71 | TC-80 |
| SC-16 | TC-21 | SC-44 | TC-51 | SC-72 | TC-81 |
| SC-17 | TC-22 | SC-45 | TC-53 | SC-73 | TC-82 |
| SC-18 | TC-23 | SC-46 | TC-54 | SC-74 | TC-84, TC-86 |
| SC-19 | TC-24, TC-25 | SC-47 | TC-55 | SC-75 | TC-85, TC-86 |
| SC-20 | TC-26 | SC-48 | TC-56 | SC-76 | TC-87 |
| SC-21 | TC-27 | SC-49 | TC-57 | SC-77 | TC-88, TC-89 |
| SC-22 | TC-28 | SC-50 | TC-58 | SC-78 | TC-90 |
| SC-23 | TC-29 | SC-51 | TC-59, TC-60 | SC-79 | TC-91 |
| SC-24 | TC-30 | SC-52 | TC-61 | SC-80 | TC-92 |
| SC-25 | TC-31 | SC-53 | TC-62 | SC-81 | TC-93 |
| SC-26 | TC-32 | SC-54 | TC-63, TC-64 | SC-82 | TC-96 |
| SC-27 | TC-33 | SC-55 | TC-65 | contract | TC-37, TC-97, TC-98, TC-99 |
| SC-28 | TC-34 | SC-56 | TC-66 | | |

Instruments: `fetch_url` on the HTTP cases (TC-01 to TC-62 except TC-14, plus TC-97 to TC-99); `render_url` on TC-78, TC-83 to TC-87, TC-89, TC-91; `interact_page` on TC-79 to TC-82, TC-88, TC-90, TC-92 to TC-94; `manual` on TC-14, TC-63 to TC-77, TC-95 and TC-96. No case needs `measure_styles`: the story draws nothing new and changes no style.

## Notes for the reviewer — ambiguities found, and how the cases resolve them

1. **DATA-005 AC-3 names a 9.990.000 ₫ bike the seed does not have.** The cheapest seeded bike is 13.990.000 ₫. TC-54 proves the numeric order twice: with the seed (13.990.000 before 100.990.000, which a text sort reverses) and with a test-only bike priced 9.990.000 that it removes afterwards.
2. **DATA-004 AC-2 and AC-3 need records the seed lacks** ("Cửa Lò", "Đèo"). TC-42 inserts two test-only questions and deletes them. It also records that `q=deo` legitimately returns `delivery` too (its text holds "video").
3. **SC-9 narrows the SRS** ("naming the record") to the foreign-key message. TC-14 is written to the story. It is `manual` because it needs a deliberately broken migration file.
4. **DATA-002 AC-4 and the contract on `page` past the last.** The SRS says the answer carries "the page number"; the contract says it is "as asked, not clamped". TC-17 expects `page` 9, as the contract does.
5. **SC-63 / SRS §7 question 1.** With the API out of reach nothing new is drawn; TC-94 holds the pages to the empty state they already have, not to an error screen.
6. **TC-99 depends on this story offering only `GET`.** `services.md` describes `POST`, `PUT` and `DELETE` too; the admin story adds them and changes the answer for products and categories from 405 to 401. The case is for this story's head only.
7. **SC-82** asks for a percentile no single-request instrument measures; TC-96 is honest about it.
8. **A record edited after the seed is not specified.** SRS DATA-001 AC-7 says re-applying the migrations "changes nothing"; the SRS does not say whether re-running the seed must restore an edited row or leave it. TC-12 therefore checks counts, ids and order only.
9. **Search results for "size" (TC-84) are derived, not recorded.** They follow from the seed and the search page's own field list; if the baseline build on `main` shows different counts, the baseline wins and the case is revised.
