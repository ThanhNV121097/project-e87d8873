# SRS — Data (the shop's records)

Module: `data`
Design: [View the approved design](http://localhost:9400)
Design system: `design/design-system.md`

> One file per module. This one covers where the shop's records live and how the site reads them:
> the database, the read API and the one frontend module that talks to it. The records themselves
> are described in `code/frontend/src/data/schema.json`.

## 1. Purpose

Every page of the shop shows records — bikes, riding types, reviews, the team, gallery photographs,
workshop jobs, questions and the shop's own numbers. Today they are JSON files bundled into the site, so
nothing the owner changes can survive and nothing can be searched on a server. This module moves them into
the project's database and serves them through a paged, filterable, searchable read API, so the storefront,
the search page and the management area read the same true records. The product loses its stock, its prices
and its search without it.

## 2. Actors

| Actor | Who they are | What they may do in this module |
|---|---|---|
| Visitor | Anyone browsing; their browser calls the API | Read any record of any collection, listed or one by one. No sign-in |
| Shop owner | Signed in to the management area | Everything a visitor may; writes to bikes and riding types are in the `admin` module |

The eight collections hold no personal data a visitor could not already see on the site, so reads are public.

## 3. Scope

**In scope** — the functions specified below, by their plan titles:

- Site data from the database

**Out of scope**

- Creating, changing or deleting a record through the API — belongs to the `admin` module (bikes and riding types only).
- Placing orders and holding stock — belongs to the `orders` module.
- The pages themselves and their layout — the `site` module. Their behaviour given the records is unchanged by this module.
- Reviews, team, gallery, workshop jobs, questions and the shop's own numbers being edited anywhere but the Studio — deliberately not offered by the management area.

## 4. Functional requirements

### 4.1 Site data from the database

**Requirement DATA-001 — The eight collections are stored, and every record is kept exactly**

*As a* shop owner, *I want* every record I edited in the Studio to be in the database exactly as it was, *so that* none is lost, renamed or re-priced by the move.

Behaviour:

1. The database holds one collection each for riding types (`categories`), `faqs`, `gallery`, `products`, `reviews`, `services`, `shop` and `team`. Every field the schema lists is stored, with the schema's meaning: text, long text, image and url as text; number as a number; price as an amount of the currency the schema names (đồng, whole); bool as true or false; list as an ordered list of text; `ref:<collection>` as a reference to a record of that collection. Each record has a text id that is unique within its collection, and the record's origin (`_source`) is stored.
2. Every record of every `src/data/<name>.json` is inserted exactly as it is: the same id, every field's value, every list's order and the order of the records in the file.
3. A record is returned with the same field names its JSON used — including `_source` and, where it had one, `_invented` — and the same kinds of value: numbers as numbers, booleans as booleans, lists as lists.
4. A field a record never had is left out of the answer; it is never returned as `null`. A list that is empty is returned as an empty list.
5. A reference to a record that does not exist cannot be stored: a bike's riding type, a review's bike and a gallery photograph's riding type each name a record that exists.
6. Applying the migrations again changes nothing and duplicates nothing.

**Acceptance criteria** — each is proved by at least one test case in `docs/data/test-cases/site-data-from-the-database.md`.

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | The migrations have been applied to an empty database | Each collection is listed | `products` has 14 records, `categories` 4, `faqs` 8, `gallery` 7, `reviews` 12, `services` 6, `shop` 1 and `team` 4, equal to the number of records in each JSON file |
| AC-2 | The seeded database | Every record of every JSON file is compared with the record the API returns for the same id | They are equal, key for key, including `_source` and `_invented` |
| AC-3 | Bike `talon-2` | Its record is read | `price` is the number 18990000, `rating` is the number 4.8, `deal` is the boolean true, `sizes` is a list of five sizes in the file's order and `sizes_out` is an empty list |
| AC-4 | Bike `talon-2`, which has no earlier price | Its record is read | The answer has no `was` key, and no field of any record in any answer is `null` |
| AC-5 | The seeded database | The products are listed with no ordering asked | `talon-2` comes first and `embolden-e-plus-0` last, as in `products.json` |
| AC-6 | An attempt to store a bike whose riding type does not exist | It is made | It is refused and nothing is stored |
| AC-7 | The migrations were applied | They are applied again | Every collection still has the counts of AC-1 |
| AC-8 | The seeded database | The review of `r-thao` is read | Its `bike` is `talon-e-plus-1` and `date` is the text `2026-09-14` |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Migration fails | One migration cannot complete | The database is left as it was before that migration; nothing is half-applied |
| Reference missing | A seed record points at a record that does not exist | The migration stops with a message naming the record (AC-6) |
| Permission | — | Not applicable: this requirement stores records; reads are public and writes are in the `admin` module |

**Data touched** — all eight collections, every field of `schema.json`.

**Requirement DATA-002 — Lists come a page at a time, in a stable order**

*As a* visitor's browser, *I want* a page of a collection and the number of every match, *so that* a list can show "14 bikes" and load more without fetching everything.

Behaviour:

1. A request to list a collection with a page number and a page size answers the page asked for, the total number of records that match, the page number and the page size: `{ "items": […], "total": n, "page": p, "limit": l }`.
2. The page defaults to 1 and the size to 12. A page size above 100 is answered as 100 and the answer says `limit: 100`. A page number or size that is not a positive whole number is refused with a 400 and one sentence, in the error envelope of the service contract.
3. With no ordering asked the records come in the order they were seeded; records added later come after them in the order they were added. The same request always returns the same order, and consecutive pages neither repeat nor skip a record.
4. A page past the last answers an empty `items` and the true `total`; it is not an error.
5. A collection that does not exist answers 404.
6. Reading needs no sign-in.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | 14 bikes | The products are listed with page 1 and size 12 | 12 items, `total` 14, `page` 1, `limit` 12 |
| AC-2 | 14 bikes | Page 2 with size 12 is asked | 2 items, `total` 14 |
| AC-3 | 14 bikes | Pages 1, 2 and 3 with size 5 are asked | The three pages hold 5, 5 and 4 items, together every bike once, in the seeded order |
| AC-4 | 14 bikes | Page 9 with size 12 is asked | `items` is empty, `total` is 14 and the status is 200 |
| AC-5 | A list with no page and no size | It is asked | `page` is 1 and `limit` is 12 |
| AC-6 | A list asked with size 500 | It is asked | At most 100 items come back and `limit` is 100 |
| AC-7 | A list asked with page 0, size 0, size -1 or size "abc" | It is asked | 400 with one sentence; no items |
| AC-8 | A collection named "nothing" | It is listed | 404 |
| AC-9 | A request with no session | The bikes are listed | 200 with the page |
| AC-10 | The same list asked twice | The answers are compared | Identical, in the same order |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | Page or size is not a positive whole number | 400 (AC-7) |
| Boundary | Size 100 and size 101 | 100 is answered as 100; 101 is answered as 100 (AC-6) |
| Not found | Unknown collection | 404 (AC-8) |
| Upstream failure | The database is unreachable | A 5xx with one generic sentence in the error envelope; no record is invented |

**Data touched** — any collection's records; the page number and size asked.

**Requirement DATA-003 — Narrow a list by a field's value, a set of values or a range**

*As a* visitor's browser, *I want* to ask for only the records whose fields match, *so that* the shop can show one riding type, one brand, bikes in stock or bikes in a price range.

Behaviour:

1. Any field of the collection's schema, named as a parameter with a value, keeps only records whose field equals that value. Text, reference, image and url fields are compared ignoring capitals and accents. Number and price fields are compared as numbers and a value that is not a number is refused with a 400. A bool field takes `true` or `false` and refuses anything else with a 400. A list field matches a record whose list contains the value.
2. A comma-separated value means "any of": `category=trail,e-mtb` keeps bikes of either riding type.
3. `<field>_from` and `<field>_to` keep records whose field is at or above and at or below the given figure, both ends inclusive. For numbers and prices the comparison is numeric; for text it is by text, which is how ISO dates (`2026-09-01`) compare. A range on a bool or a list field is refused with a 400.
4. Conditions on several fields all apply together and combine with the page, the page size, the search words and the ordering; `total` counts every record that matches.
5. An empty value — `category=` — filters nothing. A record that has no value for a field does not match a condition on that field.
6. A field the collection does not have is refused with a 400 naming the field. The parameters `page`, `limit`, `q`, `search` and `sort` are the API's own; `q` is always the search words, even for the questions collection, which has a field called `q` — that field's exact value cannot be filtered on.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | 14 live bikes | Bikes with `status=live` are listed | `total` is 14 |
| AC-2 | 3 bikes marked `deal` | `deal=true` is asked | `total` is 3 and the items are `talon-2`, `stance-0` and `talon-e-plus-1` |
| AC-3 | 2 bikes with stock 0 | `stock=0` is asked | `total` is 2 |
| AC-4 | 12 bikes with stock above 0 | `status=live&stock_from=1` is asked | `total` is 12 |
| AC-5 | Riding types `trail` and `e-mtb` | `category=trail,e-mtb` is asked | Every item is a bike of either type and none of another |
| AC-6 | The bike `talon-2` priced 18990000 | `price_to=18990000` is asked | `talon-2` is among the items (the upper end is inclusive); `price_from=18990001` leaves it out |
| AC-7 | Reviews dated 2026-10-28, 2026-10-12, 2026-09-30, 2026-09-14 and earlier | `date_from=2026-09-01&date_to=2026-10-31` is asked on the reviews | `total` is 4 |
| AC-8 | Any collection | `category=` is asked | The result is the same as with no condition |
| AC-9 | Bikes | `category=Trail` and `category=trail` are asked | The same bikes |
| AC-10 | Bikes | `brand=Giant&wheel=29&category=trail` is asked | Only bikes of all three conditions |
| AC-11 | Bikes | `stock=abc`, `deal=maybe` and `deal_from=1` are each asked | Each is a 400 with one sentence |
| AC-12 | Bikes | `colourful=red` is asked | 400 naming the field `colourful` |
| AC-13 | A conditioned list with size 5 | Page 2 is asked | `total` is the number of all matches, not of the page |
| AC-14 | The questions collection | `q=free` is asked | It is treated as the search words, not as a filter on the field `q` |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | A non-numeric number, a non-boolean bool, an unknown field, a range on a bool or list | 400 with one sentence (AC-11, AC-12) |
| Boundary | A figure exactly at a range end | Included (AC-6) |
| No value | A record lacks the field | It does not match a condition on it |

**Data touched** — any field of any collection; the conditions asked.

**Requirement DATA-004 — Search the words of a collection, ignoring capitals and accents**

*As a* visitor's browser, *I want* to find records by words they contain, *so that* "cua lo" finds "Cửa Lò" and the shop's search page can look across the site.

Behaviour:

1. `q=<words>` keeps records in which the words — the whole text as one phrase, trimmed — appear anywhere in a searched field, ignoring capitals and accents. The Vietnamese letter "đ" is treated as "d", so "deo" finds "Đèo".
2. With `search=a,b` only the fields named are searched. Without it, every text-like field is searched: text, long text, image, url, reference, and the record's id — never numbers, prices, booleans or lists, and never a field whose name begins with an underscore.
3. `q` combines with conditions, ordering, the page and the page size; `total` counts every match.
4. `%`, `_` and `\` in the words mean themselves. Empty or blank words filter nothing. Words longer than 100 characters are refused with a 400. A `search` naming a field the collection does not have is refused with a 400 naming it.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | The reviews, including the one by "Thảo" | `q=thao` and `q=THAO` are asked | Both return the review `r-thao` |
| AC-2 | A record whose text holds "Cửa Lò" | `q=cua lo` is asked | That record is found |
| AC-3 | A record whose text holds "Đèo" | `q=deo` is asked | That record is found |
| AC-4 | The questions | `q=free` is asked | `total` is 2: the questions `delivery` and `service` |
| AC-5 | The questions | `q=free&search=q` is asked | `total` is 1: `service` |
| AC-6 | The bikes | `q=stance&status=live` is asked | Only live bikes whose searched fields hold "stance" |
| AC-7 | Any collection | `q=100%` is asked | Only records that hold the characters "100%" — not every record |
| AC-8 | Any collection | `q=` or `q=   ` is asked | The result is the same as with no words |
| AC-9 | Any collection | 101 characters are sent as `q` | 400 with one sentence |
| AC-10 | The questions | `search=nothing` is asked | 400 naming `nothing` |
| AC-11 | The bikes | `q=Stance` is asked with `limit=1` | One item and `total` equal to every match |
| AC-12 | The bikes | `q=20990000` is asked | No bike is found by its price: numbers are not searched |
| AC-13 | The bikes | `q='; drop table products; --` is asked | A normal empty list; no error and no table lost |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | Words over 100 characters; `search` names an unknown field | 400 with one sentence (AC-9, AC-10) |
| Special characters | `%`, `_`, `\`, quotes | Treated as the characters themselves (AC-7, AC-13) |

**Data touched** — the text-like fields of any collection.

**Requirement DATA-005 — Order a list by a field**

*As a* visitor's browser, *I want* a list ordered by price or by name, *so that* the shop can sort bikes from cheapest or dearest.

Behaviour:

1. `sort=<field>` orders ascending and `sort=-<field>` descending. Numbers and prices are ordered as numbers; text in natural order, ignoring capitals and accents, so digits inside the text compare as numbers ("Talon 2" comes before "Talon 29 3").
2. Records with equal values, and with no ordering asked, keep the order they were seeded or added. A record with no value for the field comes after those that have one, in either direction.
3. A field the collection does not have is refused with a 400 naming it. `total` is the same with or without an ordering.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | The 14 bikes | `sort=price` is asked | Each price is at or above the one before and the first is the lowest of all 14 |
| AC-2 | The 14 bikes | `sort=-price` is asked | Each price is at or below the one before and the first is the highest |
| AC-3 | Bikes priced 9.990.000 ₫ and 18.990.000 ₫ | `sort=price` is asked | 9.990.000 comes first (numeric, not by text) |
| AC-4 | The bikes `Talon 2`, `Talon 29 1`, `Talon 29 3` | `sort=name` is asked | They come in that order |
| AC-5 | Bikes with equal stock | `sort=stock` is asked | Those with equal stock keep the seeded order |
| AC-6 | The bikes | `sort=colourful` is asked | 400 naming the field |
| AC-7 | The bikes | `sort=price&category=trail&page=2&limit=2` is asked | The second page of the trail bikes by price; `total` is the number of trail bikes |
| AC-8 | The bikes | `sort=stock` is asked | A bike with no stock value comes after those with one |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | Unknown field | 400 naming it (AC-6) |
| Other cases | — | Not applicable: no roles or writes |

**Data touched** — the field ordered by.

**Requirement DATA-006 — One record by its id**

*As a* visitor's browser, *I want* one record by its id, *so that* a bike's own page can show it.

Behaviour:

1. Asking for a collection and an id answers the record, with the same field names and values as in a list (DATA-001).
2. An id that no record has, and a collection that does not exist, answer 404 with one sentence.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | Bike `stance-0` | Its record is asked | 200 with the same record the list holds for it |
| AC-2 | No bike `no-such-bike` | It is asked | 404 with one sentence |
| AC-3 | No collection "nothing" | A record of it is asked | 404 |
| AC-4 | A request with no session | A record is asked | 200 |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Not found | Unknown id or collection | 404 (AC-2, AC-3) |
| Upstream failure | The database is unreachable | A 5xx with one generic sentence in the error envelope |

**Data touched** — one record of one collection.

**Requirement DATA-007 — The site reads its records from the API through one module**

*As a* visitor, *I want* every page to show the shop's real records with real waiting, *so that* what the owner keeps true is what I see.

Behaviour:

1. The site's records are read only through the existing data module (`useCollection`, `useRecord`), which asks `<base>/v1/<collection>…` where `<base>` is the environment's `VITE_API_URL` when it is set and `/api` otherwise. The module keeps every name it exports today — `useCollection`, `useRecord`, `schema` and the types `Row`, `Source`, `Field`, `Collection`, `Filter`, `Query` — and what each returns: from `useCollection` `items`, `total`, `pages`, `page`, `hasMore`, `loading`, `error` and `loadMore`; from `useRecord` `item`, `loading` and `error`.
2. Without a `page`, the first page is requested when the component appears and `loadMore()` requests the next and adds it to the records already held; `hasMore` is true while fewer records are held than `total`. With a `page`, `items` is that page alone and `pages` is the number of pages `total` makes; a page past the last shows the last page, as today.
3. The query is sent in exactly the API's terms: a condition on a field as `field=value` (a list as `field=a,b`, a range as `field_from` and `field_to`), `q`, `search=a,b`, `sort`, `page`, and `limit` for the page size. A condition that is empty — `""`, an empty list, a range with neither end, `undefined` — is not sent.
4. A page size above 100 is never requested: the module asks for 100.
5. `loading` is true from the moment a request goes out until it is answered. `error` is an `Error` when a request fails — the network, or an answer that is not a success — and `null` again after a later request succeeds. When the query changes, `items` start again from the first page of the new query and never mix records of two queries; the answer of an earlier query that arrives late is discarded.
6. `useRecord` asks for one record: `item` is empty until it is answered and stays empty when the answer is 404; `loading` is true while waiting; `error` is `null` for a 404 and an `Error` for any other failure.
7. When another module saves or deletes a record of a collection, every list and record on screen for that collection asks again and is redrawn, without a reload. (The management area uses this, ADMIN-007.)


**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | `VITE_API_URL` unset | `useCollection("products", { pageSize: 3 })` mounts | The first request is `/api/v1/products?page=1&limit=3` |
| AC-2 | `VITE_API_URL` is `https://x.test` | The same hook mounts | The request goes to `https://x.test/v1/products?page=1&limit=3` |
| AC-3 | 14 bikes and page size 5 | `loadMore()` is called once | The second request has `page=2`; `items` holds 10; `hasMore` is true; after two more calls `items` holds 14 and `hasMore` is false |
| AC-4 | 14 bikes, `page: 2`, `pageSize: 5` | The hook mounts | `items` is the five of page 2 alone; `pages` is 3 |
| AC-5 | 14 bikes, `page: 7`, `pageSize: 5` | The hook settles | `page` is 3 and `items` is the last page |
| AC-6 | `where: { status: "live", category: ["trail","e-mtb"], stock: { from: 1 }, price: { from: "", to: 70000000 }, deal: false }` | The hook mounts | The request carries `status=live`, `category=trail,e-mtb`, `stock_from=1`, `price_to=70000000` and `deal=false`, and no `price_from` |
| AC-7 | `q: "cua lo"`, `search: ["name","fact"]`, `sort: "-price"` | The hook mounts | The request carries `q=cua%20lo`, `search=name,fact` and `sort=-price` |
| AC-8 | `where: { category: [], brand: "", month: undefined }` | The hook mounts | None of the three is sent |
| AC-9 | `pageSize: 500` | The hook mounts | The request has `limit=100` |
| AC-10 | A request is in flight | The hook is read | `loading` is true; after the answer it is false |
| AC-11 | The API answers 500 | The hook settles | `error` is an `Error`, `loading` is false, and `items` are unchanged; a later successful request sets `error` back to `null` |
| AC-12 | A slow request for query A is in flight | The query changes to B and B answers first, then A answers | `items` are B's records only |
| AC-13 | `useRecord("products", "stance-0")` | The API answers the record | `item` is it; before the answer `item` is empty and `loading` is true |
| AC-14 | `useRecord("products", "no-such-bike")` | The API answers 404 | `item` is empty, `loading` false, `error` null |
| AC-15 | The module | Every name of Behaviour 1 is imported | Each exists, and the project type-checks with no page or component changed |
| AC-16 | A list of bikes on screen | Another module saves a bike | The list asks again and shows the new record without a reload |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Upstream failure | The API cannot be reached, or answers 5xx | `error` is set and `items` stay as they were (AC-11). The design draws no error screen: the pages show the empty states they already have. See §7, open question 1 |
| Not found | A record is 404 | `item` empty, `error` null (AC-14) |
| Boundary | A page size above 100 | Asked as 100 (AC-9) |
| Stale answer | A late answer of an old query | Discarded (AC-12) |

**Data touched** — every record the pages read; the query terms of Behaviour 3.

**Requirement DATA-008 — The JSON files are gone and no page changed**

*As a* shop owner, *I want* the shop to behave as before the move, *so that* nothing I approved is lost to the migration.

Behaviour:

1. `code/frontend/src/data/` holds only `schema.json` and the data module; the eight JSON files are deleted and the site builds without them.
2. No other file of the frontend changes: no page, component, hook or library.
3. For every page the same records appear in the same order as before the move. In particular: on the shop the riding-type chips, the brand and wheel boxes and the price sort return the same bikes as before; the frame-size and price-band narrowing stays on the loaded list, and the shop asks for all live bikes in one request of 100; and the search page finds bikes, riding types, workshop jobs, questions and gallery photographs from the API, ignoring accents.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | The finished story | `src/data/` is listed | `schema.json` and `index.ts` only |
| AC-2 | The finished story | The change's file list is read | It contains `src/data/index.ts` and the deleted JSON files and nothing else |
| AC-3 | /shop with no filters | It opens | The count reads "14 bikes", 9 cards are drawn and the button reads "Show more bikes (5)" |
| AC-4 | /shop | Each riding-type chip is pressed in turn | The bikes listed, and their order, equal those the same chip listed before the move |
| AC-5 | /shop | The brand boxes and the wheel boxes are ticked in turn | The bikes equal those listed before the move |
| AC-6 | /shop | "Price: low to high" and "Price: high to low" are chosen | The bikes are in the same order as before the move |
| AC-7 | /shop with size "M" and a price band ticked | The list is read | The bikes equal those listed before the move (this narrowing is done on the loaded 14) |
| AC-8 | /search?q=size and /search?q=SIZE | Each opens | The result counts per kind are equal to each other and to those before the move; the question "How do I know which frame size to choose?" is found |
| AC-9 | /search?q=bleed | It opens | The "Workshop" block lists "Brake bleed" |
| AC-10 | Home | It opens | "12 bikes in stock", the three tinted cards `talon-2`, `stance-0`, `talon-e-plus-1`, the bike of the month `trance-x-1`, and six picks that are the first six live bikes not marked `deal`, as before |
| AC-11 | /shop/stance-0 | The "Reviews" tab is read | It holds the reviews whose `bike` is `stance-0` — two |
| AC-12 | /about | It opens | Four team cards, six workshop rows, six reviews with "Show all reviews (12)" and eight questions |
| AC-13 | /gallery | It opens | "7 photographs" |
| AC-14 | /admin/products and /admin/kinds, signed in | Each opens | 14 bike rows (one page of 20) and 4 riding-type rows |
| AC-15 | /admin | It opens | The overview asks for its bikes with page size 500 and receives 14 (asked as 100): Bikes 14, Riding types 4, Running low 14, Out of stock 2 |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Boundary | The overview asks for 500 | Asked as 100; with more than 100 bikes its counts would stop at 100 (§7, open question 3) |
| Other cases | — | Not applicable: a parity requirement with no roles or writes |

**Data touched** — all eight collections.

## 5. Screens

This module has no screen of its own. It serves the records every screen reads; the states each screen
draws are those of `docs/site/SRS.md` and `docs/admin/SRS.md` and do not change.

| Screen | Section in the design | Functions it serves | States that must exist |
|---|---|---|---|
| Shop `/shop` | `pages/Shop.tsx` | DATA-002, DATA-003, DATA-005, DATA-007, DATA-008 | default, loading, no match |
| Search `/search` | `pages/Search.tsx` | DATA-003, DATA-004, DATA-007, DATA-008 | nothing typed, results, nothing found |
| Home, Bike, About, Gallery, Bag, Checkout | `pages/*.tsx` | DATA-001, DATA-006, DATA-007, DATA-008 | as in the site SRS |
| Management lists and record page | `pages/Admin*.tsx` | DATA-002, DATA-003, DATA-005, DATA-007, DATA-008 | as in the admin SRS |

## 6. Non-functional requirements

| Area | Requirement |
|---|---|
| Performance | Listing the bikes with size 100, with or without `q`, answers within 500 ms at the 95th percentile over 100 sequential requests against the seeded database on the same host |
| Limits | A page size is at most 100; search words are at most 100 characters |
| Security | Every parameter is passed to the database as a value, never as part of a query's text: no request in AC-7 and AC-13 of DATA-004 returns a 5xx or changes data. The eight collections hold no private data, so reads are public |
| Localisation | Search treats accents and capitals as equal and "đ" as "d". Prices are whole đồng |

## 7. Dependencies and assumptions

- **Depends on:** the project's PostgreSQL database; the service contract in `docs/architecture/services.md` for the error envelope and the route prefix (the edge strips `/api`, so the browser calls `/api/v1/…` and this server serves `/v1/…`).
- **Depended on by:** `orders` (the bikes table and its stock), `admin` (the tables, the read API and the module's refresh), and `site` (the bike page's wait for its record).
- **Assumption:** the way another module tells the data module "this collection changed" is the module's existing `setCollection` name kept working; its exact form is TL's.
- **Assumption:** the Studio's preview frame and the design walk do not call this API; whether they keep working once the JSON files are gone is TL's concern.

| Open question | Proposed default | Who decides |
|---|---|---|
| 1. When the API cannot be reached the design draws no error screen. The shop would show its empty-list panel ("No bike matches all of that.") and a bike's page would stay empty | Accept at launch; the hook exposes `error` and no new screen is drawn | Stakeholder |
| 2. The old in-page search did not treat "đ" as "d"; this API does, so "deo" finds "Đèo". It can return more results than before for words with d | Keep, as a Vietnamese shop's customers type without marks | Stakeholder |
| 3. The management overview asks for up to 500 bikes and is limited to 100 by the API's cap; its counts would stop at 100 if the shop ever lists more than 100 bikes | Accept; the shop has 14 | Stakeholder |

## 8. Traceability

| Plan item | Requirement ids | Test cases |
|---|---|---|
| Site data from the database | DATA-001, DATA-002, DATA-003, DATA-004, DATA-005, DATA-006, DATA-007, DATA-008 | `test-cases/site-data-from-the-database.md` |
