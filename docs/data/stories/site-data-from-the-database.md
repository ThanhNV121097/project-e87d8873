# Story — Site data from the database

Module: `data` · Plan item: *Site data from the database* · Priority: P1 · Type: feature
Requirements: `docs/data/SRS.md` — DATA-001 to DATA-008
Contracts it builds to: `docs/architecture/services.md` (§2.3 errors, §2.4 paging, §2.7 records, §3.2 list, §3.3 one record), `docs/architecture/erd.md` (§3.0 to §3.8, §3.14, §8)
Design: the approved design — this story changes no page and no component

## 1. User story

*As a* **visitor** to Đèo Bikes, *I want* every page to show the shop's real records — bikes, riding types, reviews, the team, gallery photographs, workshop jobs, questions and the shop's own numbers — read from the shop's database through a paged, filterable, searchable API, *so that* the stock, prices and search results I see are the ones the owner keeps true.

The second actor is the **shop owner** (SRS §2): every record they edited in the Studio must arrive in the database exactly as it was, with nothing lost, renamed or re-priced by the move (DATA-001). Reading is public: no sign-in is involved anywhere in this story.

## 2. Scope

### In scope

**Backend (`code/backend`)**

- Migration `20261118000100_collections`: the three helper functions (`fold_text`, `natural_key`, `touch_record`), one typed table for each of the eight collections (`categories`, `products`, `reviews`, `gallery`, `team`, `services`, `faqs`, `shop`) with their checks, foreign keys, indexes and the `seq` column that fixes the listing order, and the version trigger on `products` and `categories` (`erd.md` §3.0 to §3.8, §3.14, §8 row 1).
- Migration `20261118000200_seed_collections`: every record of every `src/data/<name>.json` exactly as it is — same id, every field's value, every list's order, the records in the file's order — `categories` first, one insert per record, safe to apply twice (`erd.md` §8 row 2).
- `GET /v1/{collection}`: `page`, `limit`, `<field>=value` and `<field>=a,b`, `<field>_from` / `<field>_to`, `q`, `search`, `sort`, answering `{ "items", "total", "page", "limit" }` (`services.md` §3.2).
- `GET /v1/{collection}/{id}`: one record, with the `ETag` header for a bike or a riding type (`services.md` §3.3).
- Every refusal in the error envelope of `services.md` §2.3 (`INVALID_QUERY` with the `field` at fault, `NOT_FOUND`, `UNAVAILABLE`, `INTERNAL`).
- Route lines added in `internal/server.Router`; the `DB` interface the handlers take grows only by the methods the read handlers need.

**Frontend (`code/frontend`)**

- `src/data/index.ts` rewritten to call the API (base `import.meta.env.VITE_API_URL ?? "/api"`): `useCollection` fetches page by page with real `loading` and `error`, `useRecord` fetches one record, and lists and records on screen ask again when another module reports a change.
- Every export it has today stays, with what each returns: `useCollection`, `useRecord`, `schema`, the types `Row`, `Source`, `Field`, `Collection`, `Filter`, `Query` — and also `collections()` and `setCollection()`, which the management area, the Studio frame and the existing test suite still import.
- The eight JSON files in `src/data/` are deleted; `schema.json` stays.
- Tests for the above: the hook with a stubbed `fetch`, and the backend's request parsing, validation and statement building without a database (CI has none); new files only, under `src/test/` for the frontend.

### Out of scope

| Left out | Why, and where it goes |
|---|---|
| Creating, replacing or deleting a record (`POST`, `PUT`, `DELETE`), `If-Match`, `STALE_RECORD`, sign-in and sessions | The *Shop admin: real sign-in and saved edits* story (ADMIN-004 to ADMIN-008). This story only reads. The `version` column and its trigger are created here because they belong to the table definition (`erd.md` §3.14); the only read of `version` is the `ETag` header |
| Placing an order, holding stock, the `orders`, `order_lines` and `order_counters` tables | The *Place an order and hold the bike* story. It depends on the `products` table and its `stock` column from this one |
| Any change to a page or a component | DATA-007.9, DATA-008.2. The pages' own follow-ups are separate plan items: the bike page's wait for its record (SITE-016, *Bike page waits for its record*), and the management record page, overview, failure notices and the lists' "Riding type" lookups (ADMIN-010, ADMIN-013, *Shop admin…*). Until they land a page that reads only `item` can flash its "missing" or empty state before the answer arrives; the criteria below are read after the answer has settled |
| An error screen for an unreachable API | SRS §7, open question 1: the pages show the empty states they already have; the hook exposes `error`. Default accepted, not asked again |
| Editing reviews, team, gallery, workshop jobs, questions or the shop's numbers anywhere but the Studio | SRS §3: deliberately not offered by the management area |
| Frame-size and price-band narrowing on `/shop` | It stays on the loaded list, done by the page; the shop asks for all live bikes in one request of 100 (DATA-008.3) |
| Whether the Studio's data panel and the design walk keep working once the JSON files are gone | SRS §7: TL's concern. They do not call this API |
| Edit of `schema.json`, of `theme.css`, `index.css`, `package.json` or `eslint.config.js` | Frozen scaffold files; `schema.json` is read, never changed |
| Cursor paging, a search index (`pg_trgm`), response caching, CORS for the owner's cookie | `services.md` §2.1, §2.4 and `erd.md` §5: not needed at 14 bikes; every answer is `no-store` |
| The management overview counting past 100 bikes | SRS §7, open question 3: accepted, the shop has 14 |

## 3. UI scope

This story adds **no screen, no component and no changed page**. What a visitor sees is parity with the approved design: the same records in the same order, now arriving from the server, so the loading, empty and "no match" states the pages already draw now appear for a real wait.

| Screen | Design section | States that must exist (already drawn) | Criteria |
|---|---|---|---|
| Shop `/shop` | `pages/Shop.tsx` | default, loading, no match | SC-69 to SC-73 |
| Search `/search` | `pages/Search.tsx` | nothing typed, results, nothing found | SC-74, SC-75 |
| Home | `pages/Home.tsx` | as in the site SRS | SC-76 |
| Bike `/shop/:id` | `pages/Product.tsx` | as in the site SRS (read once the record has arrived) | SC-77 |
| About, Gallery | `pages/About.tsx`, `pages/Gallery.tsx` | as in the site SRS | SC-78, SC-79 |
| Management lists and overview | `pages/Admin*.tsx` | as in the admin SRS; signed in with the design's existing mock sign-in | SC-80, SC-81 |

How this fits the twelve steps: the **Story UI** (step 4) PR is the data module rewrite and the deleted JSON files — no page is touched — and is proved with a stubbed `fetch`; the **UI review** (step 5) checks the file list, the exports, the query terms and the loading and error behaviour. The page criteria (SC-68 to SC-81) can only be seen with the module and the API together, so they are observed at **Run local** and **Verify** (steps 9 and 10). The FE PR deletes the JSON files and so cannot serve a real site without the BE PR: TL merges the two back to back (steps 11 and 12) and the deployment waits for both.

## 4. How the criteria are proved

CI has no database, so the proof splits three ways. Test writes each case in one of these modes and says which.

| Mode | Proves | Criteria |
|---|---|---|
| **Unit, no database** | The hook against a stubbed `fetch`; the backend's parsing of every query parameter, every 400 sentence and code, the 404s, the 5xx for a failing database, and that every value reaches SQL as a bound parameter | SC-15, SC-16, SC-19, SC-29, SC-30, SC-40, SC-41, SC-49, SC-54 to SC-67 |
| **Running stack** (`docker compose --profile local up --build --wait`, seeded Postgres 16, HTTP calls) | What is stored, its order, counts, accent and case folding, natural order, `NULLS LAST`, constraints, idempotent migrations, performance | SC-1 to SC-14, SC-17, SC-18, SC-20 to SC-28, SC-31 to SC-39, SC-42 to SC-48, SC-50 to SC-53, SC-82 |
| **Browser on the running stack** | Page parity with the build before the move | SC-69 to SC-81 (SC-68 is read from the change's file list) |

**Records the seed does not have.** Four cases need a record the seed lacks. The case inserts it into a throwaway database or transaction and removes it afterwards; none becomes part of the seed: a record whose text holds "Cửa Lò" and one holding "Đèo" (SC-35); a bike with no `stock` (SC-48); a bike naming a riding type that does not exist (SC-6); a seed file with a broken reference (SC-9).

**"Before the move"** (SC-70 to SC-75): the baseline is the build on `main` before this story's FE PR, whose module still reads the JSON files. Test records the bikes, orders and counts it shows for each case in the test-case file, from that build, before the files are deleted.

## 5. Acceptance criteria

Each criterion is numbered `SC-n` and names the SRS requirement and criterion it narrows (`docs/data/SRS.md`; `AC-n` there is the SRS's own numbering). Error answers use the envelope of `services.md` §2.3: `error` (one sentence), `code`, and `field` where one parameter is at fault. A "list" below is `GET /v1/<collection>`; a "record" is `GET /v1/<collection>/<id>`; no request carries a session unless stated.

### A. The records are stored, exactly (DATA-001)

- **SC-1 [DATA-001 AC-1]:** On an empty database the migrations are applied. Listing each collection with `limit=100` reports `total` 14 for `products`, 4 for `categories`, 8 for `faqs`, 7 for `gallery`, 12 for `reviews`, 6 for `services`, 1 for `shop` and 4 for `team`, each equal to the number of records in its JSON file as it stood before the files were deleted.
- **SC-2 [DATA-001 AC-2]:** For every record of every JSON file, the record the API returns for the same id equals it key for key, `_source` and `_invented` included. The answer has no key the file did not have: `seq`, `version`, `created_at` and `updated_at` never appear. Each collection listed with `limit=100` holds its ids in the order of its file.
- **SC-3 [DATA-001 AC-3]:** The record of `talon-2` has `price` as the number 18990000, `rating` as the number 4.8, `deal` as the boolean true, `sizes` as a list of five sizes in the file's order and `sizes_out` as an empty list (`[]`, not omitted).
- **SC-4 [DATA-001 AC-4]:** The record of `talon-2` has no `was` key. Across all eight collections, no field of any record in any answer is `null`. The two bikes that do have an earlier price return `was` as a number.
- **SC-5 [DATA-001 AC-5]:** With no ordering asked, listing the products with `limit=100` puts `talon-2` first and `embolden-e-plus-0` last, as in `products.json`.
- **SC-6 [DATA-001 AC-6]:** Storing a bike whose riding type does not exist, a review whose bike does not exist, and a gallery photograph whose riding type does not exist is each refused by the database, and nothing is stored.
- **SC-7 [DATA-001 AC-7]:** Applying the migrations a second time changes nothing: every collection still has the counts of SC-1 and no record is duplicated.
- **SC-8 [DATA-001 AC-8]:** The record of `r-thao` has `bike` `talon-e-plus-1` and `date` the text `2026-09-14`.
- **SC-9 [DATA-001 failure table]:** A migration that cannot complete leaves the database as it was before that migration, with nothing half-applied. A seed record that points at a record that does not exist stops the migration, and the message names the missing reference (the foreign-key constraint and the missing id; this narrows the SRS's "naming the record" to what the database reports).

### B. A page at a time, in a stable order (DATA-002)

- **SC-10 [DATA-002 AC-1, AC-2]:** With 14 bikes, page 1 with size 12 answers 12 items, `total` 14, `page` 1, `limit` 12; page 2 with size 12 answers 2 items and `total` 14.
- **SC-11 [DATA-002 AC-3]:** Pages 1, 2 and 3 with size 5 hold 5, 5 and 4 items; together they hold every bike once, in the seeded order, with none repeated or skipped.
- **SC-12 [DATA-002 AC-4]:** Page 9 with size 12 answers status 200, an empty `items` and `total` 14.
- **SC-13 [DATA-002 AC-5]:** A list asked with no page and no size answers `page` 1 and `limit` 12.
- **SC-14 [DATA-002 AC-6, boundary]:** A list asked with size 500 returns at most 100 items and says `limit` 100; size 100 is answered as 100 and size 101 is answered as 100.
- **SC-15 [DATA-002 AC-7]:** A list asked with page 0, size 0, size -1, or size "abc" answers 400 with `code` `INVALID_QUERY`, one sentence, `field` naming `page` or `limit`, and no `items`.
- **SC-16 [DATA-002 AC-8]:** Listing a collection named "nothing" answers 404 with `code` `NOT_FOUND` ("That collection does not exist.").
- **SC-17 [DATA-002 AC-9]:** A request with no session lists the bikes and gets 200 with the page.
- **SC-18 [DATA-002 AC-10]:** The same list asked twice answers identically, in the same order.
- **SC-19 [DATA-002 upstream failure; DATA-006 upstream failure]:** When the database is unreachable a list and a record each answer a 5xx in the error envelope — 503 `UNAVAILABLE` ("The service is not ready.") when the connection is refused, otherwise 500 `INTERNAL` ("Something went wrong. Please try again.") — with one generic sentence that holds no SQL, host or path, and no record.

### C. Narrow a list by a value, a set of values or a range (DATA-003)

- **SC-20 [DATA-003 AC-1]:** Listing bikes with `status=live` answers `total` 14.
- **SC-21 [DATA-003 AC-2]:** `deal=true` answers `total` 3, and the items are `talon-2`, `stance-0` and `talon-e-plus-1`, in that order.
- **SC-22 [DATA-003 AC-3, AC-4]:** `stock=0` answers `total` 2; `status=live&stock_from=1` answers `total` 12.
- **SC-23 [DATA-003 AC-5]:** `category=trail,e-mtb` returns only bikes of either riding type and none of another.
- **SC-24 [DATA-003 AC-6]:** With `talon-2` priced 18990000, `price_to=18990000` includes `talon-2` (the upper end is inclusive) and `price_from=18990001` leaves it out.
- **SC-25 [DATA-003 AC-7]:** On the reviews, `date_from=2026-09-01&date_to=2026-10-31` answers `total` 4 (the reviews dated 2026-10-28, 2026-10-12, 2026-09-30 and 2026-09-14).
- **SC-26 [DATA-003 AC-8]:** On any collection, `category=` (an empty value) answers the same list as no condition.
- **SC-27 [DATA-003 AC-9]:** `category=Trail` and `category=trail` return the same bikes.
- **SC-28 [DATA-003 AC-10]:** `brand=Giant&wheel=29&category=trail` returns only bikes that meet all three conditions.
- **SC-29 [DATA-003 AC-11]:** `stock=abc`, `deal=maybe` and `deal_from=1` on the bikes each answer 400 with `code` `INVALID_QUERY` and one sentence.
- **SC-30 [DATA-003 AC-12]:** `colourful=red` on the bikes answers 400 with `field` `colourful`, and the sentence names `colourful`.
- **SC-31 [DATA-003 AC-13]:** A conditioned list with size 5, page 2, answers a `total` equal to the number of all matches, not of the page.
- **SC-32 [DATA-003 AC-14]:** On the questions, `q=free` is read as the search words and not as a filter on the field `q`; it answers the same as SC-36's first case.
- **SC-33 [DATA-003 Behaviour 1 and 5, failure table "No value"]:** On the bikes, a list field matches a record whose list contains the value (`sizes=XS` returns exactly the bikes whose `sizes` list holds XS); a bike with no value for a field does not match a condition on it (`was_from=0` returns only the two bikes that have a `was`); a range on a list field (`sizes_from=M`) answers 400.

### D. Search the words of a collection, ignoring capitals and accents (DATA-004)

- **SC-34 [DATA-004 AC-1]:** On the reviews, `q=thao` and `q=THAO` each return the review `r-thao` (reviewer "Thảo").
- **SC-35 [DATA-004 AC-2, AC-3]:** A record whose text holds "Cửa Lò" is found by `q=cua lo`; a record whose text holds "Đèo" is found by `q=deo`. The seed holds neither, so the case inserts a test-only record for each (§4).
- **SC-36 [DATA-004 AC-4, AC-5]:** On the questions, `q=free` answers `total` 2 — the questions `delivery` and `service`; `q=free&search=q` answers `total` 1 — `service`.
- **SC-37 [DATA-004 AC-6]:** On the bikes, `q=stance&status=live` returns only live bikes whose searched fields hold "stance".
- **SC-38 [DATA-004 AC-7]:** `%`, `_` and `\` in the words mean themselves: `q=100%` on the bikes returns only bikes that hold the characters "100%" and never every bike; `q=90%` on the questions returns `delivery` alone; `q=_` is not "any one character".
- **SC-39 [DATA-004 AC-8]:** On any collection, `q=` and `q=   ` answer the same list as no words.
- **SC-40 [DATA-004 AC-9]:** `q` of 101 characters answers 400 with `code` `INVALID_QUERY` and one sentence; `q` of exactly 100 characters is accepted.
- **SC-41 [DATA-004 AC-10]:** On the questions, `search=nothing` answers 400 with `field` `nothing`; a `search` that names a number, price, boolean or list field (for example `search=price` on the bikes) is refused with 400 as not searchable (`services.md` §3.2).
- **SC-42 [DATA-004 AC-11]:** On the bikes, `q=Stance&limit=1` answers one item and a `total` equal to every match.
- **SC-43 [DATA-004 AC-12; Behaviour 2]:** On the bikes, `q=20990000` finds no bike by its price (numbers are not searched), and a word that appears only in a bike's `_source` — for example `q=xedap`, from the source of `talon-29-3` — finds nothing (a field whose name begins with an underscore is never searched).
- **SC-44 [DATA-004 AC-13]:** On the bikes, `q='; drop table products; --` answers 200 with an empty list, no error, and every table is still there afterwards.

### E. Order a list by a field (DATA-005)

- **SC-45 [DATA-005 AC-1, AC-2]:** On the 14 bikes, `sort=price` gives each price at or above the one before, the first the lowest of all 14; `sort=-price` gives each at or below the one before, the first the highest.
- **SC-46 [DATA-005 AC-3]:** With bikes priced 9.990.000 ₫ and 18.990.000 ₫, `sort=price` puts 9.990.000 first (ordered as numbers, not as text).
- **SC-47 [DATA-005 AC-4]:** `sort=name` puts `Talon 2`, `Talon 29 1` and `Talon 29 3` in that order (digits inside the text compare as numbers; capitals and accents are ignored).
- **SC-48 [DATA-005 AC-5, AC-8]:** `sort=stock` keeps the seeded order among bikes with equal stock; a bike with no `stock` value (a test-only bike, §4) comes after every bike that has one, in `sort=stock` and in `sort=-stock`.
- **SC-49 [DATA-005 AC-6]:** `sort=colourful` on the bikes answers 400 naming `colourful`; `sort` on a list field (`sort=sizes`) also answers 400.
- **SC-50 [DATA-005 AC-7]:** `sort=price&category=trail&page=2&limit=2` returns the second page of the trail bikes by price, and `total` is the number of trail bikes.

### F. One record by its id (DATA-006)

- **SC-51 [DATA-006 AC-1, AC-4; services.md §3.3]:** With no session, the record of `stance-0` answers 200 with the same record the list holds for it. A bike and a riding type answer with an `ETag` header holding a quoted whole number and never hold `version` in the body; a review answers with no `ETag`.
- **SC-52 [DATA-006 AC-2]:** The record `no-such-bike` answers 404 with `code` `NOT_FOUND` and one sentence ("That record does not exist.").
- **SC-53 [DATA-006 AC-3]:** A record of the collection "nothing" answers 404 ("That collection does not exist.").

### G. The data module reads the API (DATA-007)

Each case is run on the hook with a stubbed `fetch`; "the request" is the URL the module asks.

- **SC-54 [DATA-007 AC-1, AC-2]:** With `VITE_API_URL` unset, `useCollection("products", { pageSize: 3 })` first asks `/api/v1/products?page=1&limit=3`; with `VITE_API_URL` set to `https://x.test`, the same hook asks `https://x.test/v1/products?page=1&limit=3`.
- **SC-55 [DATA-007 AC-3]:** With 14 bikes and page size 5, one `loadMore()` asks `page=2`, `items` holds 10 and `hasMore` is true; after two more calls `items` holds 14 and `hasMore` is false.
- **SC-56 [DATA-007 AC-4]:** With 14 bikes, `page: 2` and `pageSize: 5`, `items` is the five of page 2 alone and `pages` is 3.
- **SC-57 [DATA-007 AC-5]:** With 14 bikes, `page: 7` and `pageSize: 5`, once the hook settles `page` is 3 and `items` is the last page.
- **SC-58 [DATA-007 AC-6]:** `where: { status: "live", category: ["trail","e-mtb"], stock: { from: 1 }, price: { from: "", to: 70000000 }, deal: false }` sends `status=live`, `category=trail,e-mtb`, `stock_from=1`, `price_to=70000000` and `deal=false`, and no `price_from`.
- **SC-59 [DATA-007 AC-7]:** `q: "cua lo"`, `search: ["name","fact"]`, `sort: "-price"` sends `q=cua%20lo`, `search=name,fact` and `sort=-price`.
- **SC-60 [DATA-007 AC-8]:** `where: { category: [], brand: "", month: undefined }` sends none of the three.
- **SC-61 [DATA-007 AC-9]:** `pageSize: 500` asks `limit=100`; no request ever carries a `limit` above 100.
- **SC-62 [DATA-007 AC-10]:** `loading` is true while a request is in flight and false after it is answered.
- **SC-63 [DATA-007 AC-11]:** When the API answers 500 (or the network fails), `error` is an `Error`, `loading` is false and `items` are unchanged; a later successful request sets `error` back to `null`.
- **SC-64 [DATA-007 AC-12]:** With a slow request for query A in flight, the query changes to B, B answers first and then A answers: `items` are B's records only.
- **SC-65 [DATA-007 AC-13, AC-14]:** `useRecord("products", "stance-0")` has `item` empty and `loading` true until the API answers, then `item` is the record; `useRecord("products", "no-such-bike")` on a 404 leaves `item` empty, `loading` false and `error` `null`, and any other failure sets `error` to an `Error`.
- **SC-66 [DATA-007 AC-15]:** `useCollection`, `useRecord`, `schema` and the types `Row`, `Source`, `Field`, `Collection`, `Filter` and `Query` are all exported, with `collections` and `setCollection`, and each hook returns what it returns today (`items`, `total`, `pages`, `page`, `hasMore`, `loading`, `error`, `loadMore`; `item`, `loading`, `error`). `npm run build` (type-check), `npm run lint` and `npm test` pass with no page or component changed, and the existing `src/test/admin.test.ts` and `src/test/format.test.ts` pass unchanged — in particular `setCollection(name, rows)` makes `collections()[name]` return `rows` at once.
- **SC-67 [DATA-007 AC-16, AC-17]:** With a list of bikes on screen, `setCollection("products", rows)` — how the management area reports a save or delete — makes that list ask again and draw the new records without a reload; once a screen has loaded the riding types, `collections().categories` holds them, each with its `id` and `name`.

### H. The JSON files are gone and no page changed (DATA-008)

- **SC-68 [DATA-008 AC-1, AC-2]:** `code/frontend/src/data/` holds `schema.json` and `index.ts` only, and the change's file list holds `src/data/index.ts`, the eight deleted JSON files and, at most, new test files under `src/test/` — no page, component, hook or library file (DATA-008.2; the test-file allowance is written into the SRS with this story).
- **SC-69 [DATA-008 AC-3]:** `/shop` with no filters reads "14 bikes", draws 9 cards, and the button reads "Show more bikes (5)".
- **SC-70 [DATA-008 AC-4]:** On `/shop`, pressing each riding-type chip in turn lists the bikes, in the order, that the same chip listed before the move.
- **SC-71 [DATA-008 AC-5]:** On `/shop`, ticking the brand boxes and the wheel boxes in turn lists the bikes that were listed before the move.
- **SC-72 [DATA-008 AC-6]:** On `/shop`, "Price: low to high" and "Price: high to low" put the bikes in the order they had before the move.
- **SC-73 [DATA-008 AC-7]:** On `/shop` with size "M" and a price band ticked, the bikes listed equal those listed before the move (this narrowing is done on the loaded 14).
- **SC-74 [DATA-008 AC-8]:** `/search?q=size` and `/search?q=SIZE` show the same result counts per kind, equal to those before the move, and find the question "How do I know which frame size to choose?".
- **SC-75 [DATA-008 AC-9]:** `/search?q=bleed` lists "Brake bleed" in its "Workshop" block.
- **SC-76 [DATA-008 AC-10]:** Home shows "12 bikes in stock", the three tinted cards `talon-2`, `stance-0` and `talon-e-plus-1`, the bike of the month `trance-x-1`, and six picks that are the first six live bikes not marked `deal`, as before.
- **SC-77 [DATA-008 AC-11]:** On `/shop/stance-0`, once the bike has arrived, the "Reviews" tab holds the two reviews whose `bike` is `stance-0`.
- **SC-78 [DATA-008 AC-12]:** `/about` shows four team cards, six workshop rows, six reviews with "Show all reviews (12)" and eight questions.
- **SC-79 [DATA-008 AC-13]:** `/gallery` reads "7 photographs".
- **SC-80 [DATA-008 AC-14]:** Signed in with the design's mock sign-in, `/admin/products` shows 14 bike rows (one page of 20) and `/admin/kinds` shows 4 riding-type rows.
- **SC-81 [DATA-008 AC-15; Boundary]:** `/admin` asks for its bikes with page size 500, which the module sends as 100 (SC-61), and receives 14: Bikes 14, Riding types 4, Running low 14, Out of stock 2.

### I. Non-functional (SRS §6)

- **SC-82 [SRS §6 Performance]:** Listing the bikes with `limit=100`, with and without `q=stance`, over 100 sequential requests against the seeded database on the same host, answers within 500 ms at the 95th percentile. (The injection and special-character cases of §6 Security are SC-38 and SC-44.)

## 6. Traceability — every SRS criterion has a criterion here

| Requirement | SRS criterion → story criterion |
|---|---|
| DATA-001 | AC-1 → SC-1 · AC-2 → SC-2 · AC-3 → SC-3 · AC-4 → SC-4 · AC-5 → SC-5 · AC-6 → SC-6 · AC-7 → SC-7 · AC-8 → SC-8 · failure table → SC-9 |
| DATA-002 | AC-1, AC-2 → SC-10 · AC-3 → SC-11 · AC-4 → SC-12 · AC-5 → SC-13 · AC-6 → SC-14 · AC-7 → SC-15 · AC-8 → SC-16 · AC-9 → SC-17 · AC-10 → SC-18 · upstream failure → SC-19 |
| DATA-003 | AC-1 → SC-20 · AC-2 → SC-21 · AC-3, AC-4 → SC-22 · AC-5 → SC-23 · AC-6 → SC-24 · AC-7 → SC-25 · AC-8 → SC-26 · AC-9 → SC-27 · AC-10 → SC-28 · AC-11 → SC-29 · AC-12 → SC-30 · AC-13 → SC-31 · AC-14 → SC-32 · Behaviour 1 and 5 → SC-33 |
| DATA-004 | AC-1 → SC-34 · AC-2, AC-3 → SC-35 · AC-4, AC-5 → SC-36 · AC-6 → SC-37 · AC-7 → SC-38 · AC-8 → SC-39 · AC-9 → SC-40 · AC-10 → SC-41 · AC-11 → SC-42 · AC-12 → SC-43 · AC-13 → SC-44 |
| DATA-005 | AC-1, AC-2 → SC-45 · AC-3 → SC-46 · AC-4 → SC-47 · AC-5, AC-8 → SC-48 · AC-6 → SC-49 · AC-7 → SC-50 |
| DATA-006 | AC-1, AC-4 → SC-51 · AC-2 → SC-52 · AC-3 → SC-53 · upstream failure → SC-19 |
| DATA-007 | AC-1, AC-2 → SC-54 · AC-3 → SC-55 · AC-4 → SC-56 · AC-5 → SC-57 · AC-6 → SC-58 · AC-7 → SC-59 · AC-8 → SC-60 · AC-9 → SC-61 · AC-10 → SC-62 · AC-11 → SC-63 · AC-12 → SC-64 · AC-13, AC-14 → SC-65 · AC-15 → SC-66 · AC-16, AC-17 → SC-67 |
| DATA-008 | AC-1, AC-2 → SC-68 · AC-3 → SC-69 · AC-4 → SC-70 · AC-5 → SC-71 · AC-6 → SC-72 · AC-7 → SC-73 · AC-8 → SC-74 · AC-9 → SC-75 · AC-10 → SC-76 · AC-11 → SC-77 · AC-12 → SC-78 · AC-13 → SC-79 · AC-14 → SC-80 · AC-15 → SC-81 |
| SRS §6 | Performance → SC-82 · Security → SC-38, SC-44 |

## 7. Dependencies

**Must exist first**

- The merged scaffold: the migration runner (`internal/migrate`, applied on boot under an advisory lock, one transaction per file), `internal/db`, `internal/httpx` (error writer, `NoStore`, request log) and `internal/server.Router`.
- The contracts: `docs/architecture/erd.md` and `docs/architecture/services.md`, both on `main`. `docs/architecture/overview.md` is not in the repository (`services.md` §9, question 5); the constraints here are taken from those two files and the recorded project decisions.
- The design as built (plan item 1, done): the pages and the eight JSON files this story moves.
- PostgreSQL 16 with a UTF-8 database (`normalize()` needs it); locally the `local` compose profile starts the project's own.

**External accounts and secrets:** none. `DATABASE_URL` is injected by the runtime and already listed in `code/backend/.env.example`; this story adds no key, no `request_secret` and no `request_setup`.

**Migration files:** `code/backend/migrations/20261118000100_collections` and `20261118000200_seed_collections`, each with its `.up.sql` and `.down.sql` (the slots `erd.md` §8 assigns to this story, so stories built in parallel never collide).

**Depended on by**

- *Place an order and hold the bike*: the `products` table and its `stock` column.
- *Shop admin: real sign-in and saved edits*: the tables, the read API and its `ETag`, `setCollection` as the refresh signal, and `collections()`.
- *Bike page waits for its record*: `useRecord`'s `loading` and `error` (SC-65).

## 8. Decided here, so no later stage has to guess

1. **Storage names follow `erd.md` R2.** The plan item says `_source` becomes a column `source`; the ERD, which is later and more exact, keeps the schema's key names verbatim (`_source`, `_invented`). No behaviour differs: the API returns `_source` and `_invented` either way.
2. **A missing seed reference** is reported by the database's own foreign-key message (constraint and missing id), not by custom text (SC-9).
3. **`setCollection` and `collections` stay.** The SRS keeps them as assumptions (§7); the existing `src/test/admin.test.ts`, which this story may not edit, calls `setCollection(name, rows)` and then reads `collections()[name]`, so the call must update what `collections()` returns at once as well as tell mounted lists to ask again. The exact mechanism is TL's.
4. **New test files are allowed** under `src/test/` although DATA-008.2 says no other frontend file changes: that sentence is about product code, and the module's behaviour (SC-54 to SC-67) can only be proved by a test. SRS DATA-008.2 and AC-2 are amended to say so.
5. **Parity is read after the answer settles.** A page that reads only `item` can flash its "missing" or empty state during the wait; fixing that is SITE-016 and ADMIN-013, not this story.
6. **The three open questions of SRS §7 keep their proposed defaults** (no error screen; "deo" finds "Đèo"; the overview stops counting at 100 bikes). None blocks the build.
7. **Only `GET` is offered.** Other methods on these paths are the admin story's to add.
