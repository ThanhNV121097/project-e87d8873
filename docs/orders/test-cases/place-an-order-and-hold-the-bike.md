# Test cases — Place an order and hold the bike

Module: `orders` · Plan item 3 · Story: `docs/orders/stories/place-an-order-and-hold-the-bike.md`
Requirements: ORD-001 to ORD-008 (`docs/orders/SRS.md`) · Contract: `docs/architecture/services.md` §2.3, §2.5, §3.10, §5, §6 · Schema: `docs/architecture/erd.md` §3.9–3.11

**Risk level: HIGH.** The endpoint writes customer data and money, takes stock out of a shared figure under concurrency, and is public (no sign-in, no rate limit). A wrong price, a negative stock, a half-stored order or a lost refusal sentence each cost the shop a sale or a bike it cannot deliver. So every refusal path, every boundary the SRS or contract names, the concurrent cases and the failure paths (transaction fails, database out, service out) get their own case. Presentation-only criteria get one case each.

## 0. How every case is run

**Endpoint.** `POST <origin>/api/v1/orders` (the edge strips `/api`; the backend mounts `/v1/orders`), `Content-Type: application/json`, no cookie and no `Authorization` header — the customer has no account (ORD §2).

**Valid body `V`** (the story's seed customer):
`{"name":"Minh","phone":"0901234567","delivery":"local","address":"12 Hùng Vương","payment":"transfer","lines":[{"id":"talon-2","variant":"M","qty":1}],"website":""}`. A case says what it changes from `V`.

**Seed state `R`** (re-arranged with `set_db_state` before a case that says "Given R"; the arrangement is put back at the end of the case): `orders`, `order_lines` and `order_counters` are empty; `talon-2` is live, Giant, "Talon 2", price 18990000, stock 9, sizes XS S M L XL, `sizes_out` `[]`, image `/images/ref/talon-2.jpg`; `stance-0` live with its seeded price and stock 7; `pique-advanced-sl-se` live with stock 0; `anthem-advanced-sl-0` live with stock 1; the `shop` record's `delivery_fee` is 350000. Bike facts a case quotes (brand, name, price, image) are read first from `GET /v1/products/{id}` so the expected sentence is built from the record, not from memory.

**Reading the database.** No endpoint reads an order back (ORD-001.5), so an order's rows can only be read by SQL. Where a `Then` says "the database holds", it is read by SQL in the verification step (story §5). Wherever possible the same fact is also proved over HTTP, and that HTTP proof decides when no SQL read is available: *a refused order stored nothing* = the bike's `stock` from `GET /v1/products/{id}` is unchanged **and** the next accepted order's number is exactly the previous number plus one (a refused order spends no number, ORD-006, erd §3.11).

**Concurrent cases.** "At the same moment" means requests issued in parallel, not one after another. If `fetch_url` serialises them, the case reports what it saw and the strictly concurrent part is UNVERIFIED (the Go test of the transaction covers it); it never silently becomes a sequential pass.

**UI cases** drive the real site with `interact_page`. Demo mode must be off (no `agentteam:demo` key in `localStorage`, no `?demo` in the address) — in demo mode an empty bag shows `EXAMPLE_LINES`, which is the design preview and not a customer path. The bag is filled the way a customer does it:
- **Bag step `B(n, size)`** — open `/shop/talon-2`, press `[data-testid="size-<size>"]`, press `[aria-label="One more"]` `n − 1` times, press `[data-testid="product-add"]` (the product page's stepper stops at the bike's stock).
- **Form step `F`** — on `/checkout` type `#order-name` "Minh", `#order-phone` "0901234567", `#order-address` "12 Hùng Vương", optionally `#order-note`; choose a delivery with `[data-testid="deliver-local|courier|pickup"]` and a payment with `[data-testid="pay-transfer|cod|shop"]`; the submit button is `[data-testid="place-order"]`.
- **Recorder** — before pressing "Place order" a wrapper is installed on `window.fetch` and `navigator.sendBeacon` that records each call's URL, method, body and the response status; it is read back after the press. The Agent Team line is `<api>/<key>/messages` with `api` and `key` from `src/site.json`; in the verification environment that address may refuse the connection or answer a non-2xx — every case that depends on the Agent Team call holds either way, and a case says what it records.

**Arranging a slow or failing order** (UI and concurrency cases), through `set_db_state`, removed at the end of the case: *slow* = a `BEFORE INSERT` trigger on `orders` that runs `pg_sleep(3)`; *fails mid-way* = a `BEFORE INSERT` trigger on `order_lines` that raises an exception (the order row and the stock update have already run in the transaction by then).

**Instruments.** `fetch_url` for every HTTP answer (status, body, headers); `render_url` for what a page shows; `measure_styles` for a computed colour or opacity; `interact_page` for anything a customer does. `manual` is used only where no instrument can observe the fact; the reason is written in the case.

**Statuses and codes** are the contract's (§2.3): every non-2xx answer has exactly `error` (one sentence), `code` and, only for a single field at fault, `field`. Cases match the status, `code`, `field` and the quoted key words; the **fixed** ORD-004 sentences are matched word for word.

## 1. The order is stored and answered (ORD-001)

**Scenario**: A valid order is accepted and answered with the stored order, priced from the bike's record
**Given**: Seed state R. `GET /v1/products/talon-2` read first: brand "Giant", name "Talon 2", price 18990000, stock 9, `image` I (`/images/ref/talon-2.jpg`).
**When**: `POST /v1/orders` with `V` changed to 2 × `talon-2` in size M (`"lines":[{"id":"talon-2","variant":"M","qty":2}]`).
**Then**: Status 201. `lines` equals, key for key, exactly one line `{"id":"talon-2","name":"Giant Talon 2","price":18990000,"image":"/images/ref/talon-2.jpg","variant":"M","qty":2}`; that line's `image` equals the `image` field of the bike's record (I). `subtotal` 37980000, `delivery_fee` 0, `total` 37980000, each a JSON integer (no quotes, no decimal point). `name` "Minh", `phone` "0901234567", `delivery` "local", `payment` "transfer", `address` exactly "12 Hùng Vương" with its diacritics intact.
Traces: SC-1 (ORD-001 AC-1, AC-5; ORD-002.2)
Check: fetch_url

**Scenario**: The 201 body has exactly eleven keys and a placed time with an offset
**Given**: Seed state R.
**When**: The order of the previous case is placed and the raw response is read (body and headers).
**Then**: The body's keys are exactly `address`, `delivery`, `delivery_fee`, `lines`, `name`, `number`, `payment`, `phone`, `placed`, `subtotal`, `total` — no `note`, no `id`, no `website`. `placed` matches `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\+07:00$`, parses as an RFC 3339 time, and is within 5 minutes of the moment the request was made. `number` matches `^DEO-\d{6}-\d{4}$`. Each line has exactly the keys `id`, `name`, `price`, `image`, `variant`, `qty`. Headers: `Content-Type: application/json; charset=utf-8`, `Cache-Control: no-store`, an `X-Request-Id` of 32 hex characters, and no `Location` header.
Traces: SC-2 (ORD-001 AC-2; ORD-001.3; contract POST /v1/orders success, §2.1)
Check: fetch_url

**Scenario**: The note is stored with the order and is not in the answer
**Given**: Seed state R.
**When**: `V` is posted with `"note":"Please call after 5pm"`.
**Then**: Status 201 and the body has no `note` key (its keys are the eleven of the previous case). The database holds one order whose `note` is exactly "Please call after 5pm". A second order posted with no `note` key is 201 and its stored `note` is `''`.
Traces: SC-3 (ORD-001 AC-3)
Check: fetch_url

**Scenario**: The database holds the order and its line exactly as answered
**Given**: Seed state R (so exactly one order will exist).
**When**: 2 × `talon-2` size M is placed with `V`'s customer, `delivery` "courier", `payment` "transfer" and `"note":"Call after 5pm"`; the 201 body is kept; then `orders` and `order_lines` are read by SQL.
**Then**: `orders` holds one row whose `number`, `name`, `phone`, `delivery`, `address`, `payment`, `subtotal`, `delivery_fee` and `total` equal the answer's (38330000 total, 350000 fee), whose `note` is "Call after 5pm", and whose `placed_at` is the answer's `placed` instant. `order_lines` holds one row for that order with `position` 1, `product_id` "talon-2", `name` "Giant Talon 2", `size` "M", `quantity` 2 and `unit_price` 18990000.
Traces: SC-4 (ORD-001 AC-4; ORD-001.4)
Check: fetch_url (the 201) with the database read by SQL in the verification step

**Scenario**: Migration 20261118000300 has created the three tables and the service is healthy
**Given**: The running story, migrated at boot.
**When**: `GET /healthz` is read and the schema is read by SQL.
**Then**: `/healthz` is 200 `{"status":"ok"}`. `orders`, `order_lines` and `order_counters` exist; migration `20261118000300` is recorded as applied exactly once; the `number` column of `orders` has a unique index; `order_lines` has no foreign key to `products` (its only foreign key is `order_id` → `orders`).
Traces: SC-5 (ORD-001 AC-4; erd §3.9–3.11)
Check: fetch_url (`/healthz`) with the schema read by SQL

**Scenario**: Deleting a bike leaves its order lines unchanged
**Given**: Seed state R. A throwaway live bike `test-ephemeral` (Giant, name "Test Ephemeral", price 5000000, stock 5, sizes `["M"]`) inserted with `set_db_state`. One order of 1 × `test-ephemeral` size M is accepted (201).
**When**: The `test-ephemeral` row is deleted from `products` and `order_lines` is read by SQL.
**Then**: The delete succeeds (it is not refused by a foreign key) and the order's line still exists with `product_id` "test-ephemeral", `name` "Giant Test Ephemeral", `size` "M", `quantity` 1 and `unit_price` 5000000; the order row is unchanged. `GET /v1/products/test-ephemeral` is 404.
Traces: SC-5 (ORD-001 AC-4; erd §3.10)
Check: fetch_url with the database read by SQL in the verification step

**Scenario**: The down migration removes the three order tables and nothing of the bikes
**Given**: The PR diff for `code/backend/migrations/20261118000300_orders.up.sql` and `.down.sql`.
**When**: The `.down.sql` is read.
**Then**: It drops `order_lines`, `order_counters` and `orders` (lines before orders) and contains no statement on `products`, `categories`, `shop` or any other table; the `.up.sql` creates exactly those three tables. Reason `manual`: running the down migration would destroy the running story's tables, so no instrument is allowed to run it; the migration pair is also checked by `code/backend/migrations/migrations_test.go`.
Traces: SC-5 (ORD-001 AC-4)
Check: manual

**Scenario**: No endpoint reads, lists, changes or cancels an order, and no answer carries customer data
**Given**: Seed state R. An order for "Minh", phone "0901234567", address "12 Hùng Vương", note "Please call after 5pm" is accepted with number N; the order's internal id is read by SQL as U.
**When**: These are requested: `GET /v1/orders`, `GET /v1/orders/N`, `GET /v1/orders/U`, `PUT /v1/orders/N` (body `{}`), `DELETE /v1/orders/N`, `POST /v1/orders/N` (body `V`), and `POST /v1/orders/` (trailing slash, body `V`).
**Then**: Each answers 404 `NOT_FOUND` or 405 `METHOD_NOT_ALLOWED` (never 200, 201, 204) in the error envelope (`error`, `code`, no order data); `GET /v1/orders` serves no list. None of the seven bodies, nor `GET /v1/products`, `GET /v1/shop` and `GET /v1/products/talon-2`, contains the text "Minh", "0901234567", "12 Hùng Vương" or "Please call after 5pm". The database still holds the one order unchanged.
Traces: SC-6 (ORD-001 AC-6; ORD-001.5; contract §3 intro, §2.1 paths are exact; ORD §6 Privacy)
Check: fetch_url

## 2. The server prices the order (ORD-002)

**Scenario**: The browser's own money, number and names are ignored
**Given**: Seed state R; `talon-2` priced 18990000.
**When**: `V` with 1 × `talon-2` size M is posted with the extra keys `"total":1`, `"subtotal":1`, `"delivery_fee":1`, `"number":"X"`, `"placed":"x"` and, on the line, `"price":1`, `"name":"Cheap"`, `"image":"x"`. A second request carries `"total":"not a number"` and `"price":null` on the line.
**Then**: Both are 201 (the advisory keys are accepted and never read, not type-checked). `lines[0].price` is 18990000, `lines[0].name` is "Giant Talon 2", `lines[0].image` is "/images/ref/talon-2.jpg", `subtotal` and `total` are 18990000, `delivery_fee` is 0, `number` matches `^DEO-\d{6}-\d{4}$` and is not "X", `placed` is a real time with `+07:00`. The database holds `unit_price` 18990000 for both lines.
Traces: SC-7 (ORD-002 AC-1; contract §3.10 advisory keys; also the "price changed after the bag was filled" row of ORD-002)
Check: fetch_url

**Scenario**: A courier order carries the shop's courier charge
**Given**: Seed state R; the `shop` record's `delivery_fee` is 350000.
**When**: 2 × `talon-2` size M is posted with `delivery` "courier" and `address` "12 Hùng Vương".
**Then**: 201 with `subtotal` 37980000, `delivery_fee` 350000, `total` 38330000, `delivery` "courier".
Traces: SC-8 (ORD-002 AC-2)
Check: fetch_url

**Scenario**: Local delivery and pickup carry no delivery fee
**Given**: Seed state R; courier charge 350000.
**When**: 2 × `talon-2` size M is posted (a) with `delivery` "local" and an address, and (b) with `delivery` "pickup" and no `address` key.
**Then**: Both are 201. (a) `delivery_fee` 0, `total` 37980000. (b) `delivery_fee` 0, `total` 37980000, and `address` is present in the answer as `""`.
Traces: SC-8 (ORD-002 AC-3, AC-4; contract answer `address` always present)
Check: fetch_url

**Scenario**: A changed courier charge applies to the next courier order
**Given**: Seed state R; the `shop` record's `delivery_fee` changed to 400000 with `set_db_state` (put back to 350000 at the end).
**When**: A courier order for 1 × `talon-2` size M with an address is posted.
**Then**: 201 with `delivery_fee` 400000 and `total` 19390000.
Traces: SC-9 (ORD-002 AC-5)
Check: fetch_url

**Scenario**: Two bikes in one order are summed from their own prices
**Given**: Seed state R. `GET /v1/products/talon-2` and `GET /v1/products/stance-0` read first: prices P1 and P2.
**When**: One order of 1 × `talon-2` size M and 1 × `stance-0` size M is posted.
**Then**: 201 with two lines in the order sent (`talon-2` first), `lines[0].price` P1, `lines[1].price` P2, `subtotal` P1 + P2 and `total` P1 + P2.
Traces: SC-10 (ORD-002 AC-6)
Check: fetch_url

**Scenario**: A price the owner changes applies to the next order and an earlier order keeps its price
**Given**: Seed state R. Order A for 1 × `talon-2` size M is accepted at 18990000. Then `talon-2`'s price is changed to 17990000 with `set_db_state` (put back at the end).
**When**: Order B for 1 × `talon-2` size M is posted, then `order_lines` is read by SQL.
**Then**: B is 201 with `lines[0].price` 17990000, `subtotal` 17990000, `total` 17990000. The database holds A's line with `unit_price` 18990000 (unchanged) and B's line with 17990000.
Traces: SC-11 (ORD-002 AC-7; ORD-002.3)
Check: fetch_url with the database read by SQL in the verification step

## 3. What is refused as invalid (ORD-003)

Every 400 below is in the error envelope (`error`, `code`, `field` where one field is at fault). A refusal stores nothing; the bookkeeping check of "Refused orders leave no trace" (section 3, last case but one) proves it for the whole family.

**Scenario**: A missing name and a name of only spaces are refused, naming the name
**Given**: Seed state R.
**When**: `V` is posted (a) with no `name` key, (b) with `"name":"   "`, (c) with `"name":""`.
**Then**: Each is 400, `code` `VALIDATION_FAILED`, `field` "name", and `error` is one sentence containing the word "name" (default: "Your name is needed."). `talon-2`'s stock is still 9.
Traces: SC-12 (ORD-003 AC-1; contract §3.10 step 4)
Check: fetch_url

**Scenario**: A missing phone is refused, naming the phone
**Given**: Seed state R.
**When**: `V` is posted (a) with no `phone` key, (b) with `"phone":"  "`.
**Then**: Each is 400, `code` `VALIDATION_FAILED`, `field` "phone", `error` contains "phone" (default: "Your phone number is needed.").
Traces: SC-12 (ORD-003 AC-2)
Check: fetch_url

**Scenario**: Empty or missing lines are refused
**Given**: Seed state R.
**When**: `V` is posted (a) with `"lines":[]`, (b) with no `lines` key.
**Then**: Each is 400 `VALIDATION_FAILED`, `field` "lines"; (a)'s sentence is "Your bag is empty." No order is stored.
Traces: SC-13 (ORD-003 AC-3)
Check: fetch_url

**Scenario**: Twenty lines are accepted and twenty-one are refused
**Given**: Seed state R with `talon-2` stock raised to 100 (put back at the end). A 20-line list: `talon-2` sizes XS, S, M, L, XL repeated four times, each `qty` 1; a 21-line list is the same plus one more `talon-2` size M line.
**When**: `V` is posted with the 20-line list, then with the 21-line list.
**Then**: The 20-line order is 201; its `lines` has 20 entries in the order sent, `subtotal` 379800000 (20 × 18990000); `talon-2`'s stock is then 80 (the quantities of one bike are counted together). The 21-line order is 400 `VALIDATION_FAILED`, `field` "lines", the sentence says at most 20 lines ("An order holds at most 20 lines."), and the stock is still 80.
Traces: SC-13 (ORD-003 AC-4, AC-5)
Check: fetch_url

**Scenario**: A quantity of 1 and of 10 are accepted
**Given**: Seed state R with `talon-2` stock raised to 30 (put back at the end).
**When**: `V` is posted with `qty` 1, then (separately) with `qty` 10.
**Then**: Both are 201; the lines carry `qty` 1 and 10; the stock is 29 and then 19.
Traces: SC-14 (ORD-003 AC-6)
Check: fetch_url

**Scenario**: A quantity of 0, 11, 1.5, -1, 2.0 or "2" is refused
**Given**: Seed state R with `talon-2` stock raised to 30 (put back at the end).
**When**: `V` is posted six times, with the line's `qty` as the JSON values `0`, `11`, `1.5`, `-1`, `2.0` and `"2"`; and once with the `qty` key absent.
**Then**: Each is 400. For `0`, `11`, `1.5`, `-1` and the absent key: `code` `VALIDATION_FAILED`, `field` "lines[0].qty", and the sentence names the quantity ("The quantity of line 1 must be a whole number from 1 to 10."). For `"2"`: `code` `MALFORMED_REQUEST` (a string where a number is expected). For `2.0`: 400 (an integer field takes an integer literal; either code). The stock is still 30 afterwards.
Traces: SC-14 (ORD-003 AC-7; contract §2.5.3 reject, do not coerce)
Check: fetch_url

**Scenario**: A delivery method or payment method outside the closed sets is refused
**Given**: Seed state R.
**When**: `V` is posted (a) with `"delivery":"drone"`, (b) with no `delivery` key, (c) with `"payment":"card"`, (d) with no `payment` key.
**Then**: Each is 400 `VALIDATION_FAILED`; (a) and (b) `field` "delivery" with a sentence that says "delivery" and lists local, courier, pickup; (c) and (d) `field` "payment" with a sentence that says "payment" and lists transfer, cod, shop.
Traces: SC-15 (ORD-003 AC-8, AC-9)
Check: fetch_url

**Scenario**: Cash on delivery is refused for courier and pickup and accepted for local delivery
**Given**: Seed state R.
**When**: Posted: (a) `payment` "cod" with `delivery` "courier" and an address; (b) `payment` "cod" with `delivery` "pickup" and no address; (c) `payment` "cod" with `delivery` "local" and address "12 Hùng Vương".
**Then**: (a) and (b) are each 400 `VALIDATION_FAILED`, `field` "payment", and the sentence names cash on delivery ("Cash on delivery is only available for delivery in town."). (c) is 201 with `payment` "cod" and `delivery_fee` 0.
Traces: SC-16 (ORD-003 AC-10, AC-11)
Check: fetch_url

**Scenario**: An address is required unless the customer collects from the shop
**Given**: Seed state R.
**When**: Posted: (a) `delivery` "local" with `"address":""`; (b) `delivery` "local" with no `address` key; (c) `delivery` "courier" with `"address":"   "`; (d) `delivery` "pickup" with no `address` key; (e) `delivery` "pickup" with `"address":""`.
**Then**: (a), (b) and (c) are each 400 `VALIDATION_FAILED`, `field` "address", the sentence contains "address" ("A delivery address is needed."). (d) and (e) are 201 and `address` in the answer is `""`.
Traces: SC-17 (ORD-003 AC-12, AC-13)
Check: fetch_url

**Scenario**: Name, phone, address and note are accepted at their limits
**Given**: Seed state R.
**When**: `V` is posted with `name` of 100 characters, `phone` of 30 characters, `address` of 300 characters and `note` of 1000 characters (the letter "a" repeated, `phone` all digits).
**Then**: 201. The answer's `name`, `phone` and `address` have exactly 100, 30 and 300 characters; the stored `note` has 1000 characters (database read by SQL).
Traces: SC-18 (ORD-003 AC-14)
Check: fetch_url with the database read by SQL in the verification step

**Scenario**: One character past each text limit is refused, naming the field
**Given**: Seed state R.
**When**: `V` is posted four times, with `name` of 101 characters, `phone` of 31, `address` of 301 and `note` of 1001; and once with the line's `variant` of 21 characters.
**Then**: Each is 400 `VALIDATION_FAILED`; `field` is respectively "name", "phone", "address", "note" and "lines[0].variant"; the sentence says a text is too long. `talon-2`'s stock is unchanged.
Traces: SC-18 (ORD-003 AC-14; ORD-003.5)
Check: fetch_url

**Scenario**: A frame size of 20 characters passes validation (and is then refused as an unknown size)
**Given**: Seed state R.
**When**: `V` is posted with the line's `variant` of exactly 20 characters ("XXXXXXXXXXXXXXXXXXXX").
**Then**: The answer is not a 400: it is 409 `BIKE_UNAVAILABLE` "Giant Talon 2 is not available in size XXXXXXXXXXXXXXXXXXXX." — the length limit is 20, and this size is only unknown to the bike. (For a bike with no sizes the same value is accepted and ignored: see the case "A bike with no frame sizes is accepted without a size".)
Traces: SC-18 (ORD-003 AC-14; size limit 20)
Check: fetch_url

**Scenario**: Text is trimmed before it is counted, stored and answered
**Given**: Seed state R.
**When**: `V` is posted with `"name":" Minh "`, `"phone":" 0901234567 "`, `"address":" 12 Hùng Vương "`, `"note":"  hi  "`, and the line `"variant":" M "`; then with `name` of 100 characters surrounded by two spaces each side; then with `name` of 101 characters surrounded by two spaces each side.
**Then**: The first is 201 with `name` "Minh", `phone` "0901234567", `address` "12 Hùng Vương", the line's `variant` "M"; the database holds `note` "hi". The second is 201 (the limit counts the trimmed text; the answer's name has 100 characters). The third is 400 `VALIDATION_FAILED`, `field` "name".
Traces: SC-18 (ORD-003 AC-14; ORD-003.5 trimming)
Check: fetch_url with the database read by SQL in the verification step

**Scenario**: Limits count characters, not bytes
**Given**: Seed state R.
**When**: `V` is posted with `name` of 100 times "ư" (two bytes each in UTF-8), then with 101 times "ư".
**Then**: The first is 201 and the answer's `name` is the 100 characters; the second is 400 `VALIDATION_FAILED`, `field` "name".
Traces: SC-18 (ORD-003 AC-14; contract §2.5.4 "characters means code points")
Check: fetch_url

**Scenario**: A filled hidden field is dropped with 202 and nothing happens
**Given**: Seed state R. `GET /v1/products/talon-2` read: stock 9 and its ETag E.
**When**: `V` is posted with `"website":"http://spam.example"`; then a valid order without that text is posted.
**Then**: The first answer is 202 with a body that is exactly `{}`. The database holds no order and no line; `talon-2`'s stock is 9 and its ETag is E (no hold, no version change). The second order is 201 and its number ends `-0001` (the dropped request spent no number).
Traces: SC-19 (ORD-003 AC-15; ORD-003.7; contract §3.10 step 2)
Check: fetch_url

**Scenario**: The bot check comes before every other check; a non-text hidden field is a malformed request
**Given**: Seed state R.
**When**: Posted: (a) `V` with `"website":"x"` and `"name":""`; (b) `V` with `"website":"x"` and an undefined key `"coupon":"A"`; (c) `V` with `"website":5`.
**Then**: (a) and (b) are 202 `{}` (nothing is validated once the bot field is filled), nothing stored. (c) is 400 `MALFORMED_REQUEST`.
Traces: SC-19 (ORD-003 AC-15; contract §3.10 steps 2–3)
Check: fetch_url

**Scenario**: A body that is not a JSON object, or not JSON, is refused
**Given**: Seed state R.
**When**: These raw bodies are posted with `Content-Type: application/json`: `[]`, `"just text"`, `null`, `5`, `not json`, an empty body, and `{"name":"Minh"}{"x":1}` (trailing data).
**Then**: Each is 400 `MALFORMED_REQUEST` in the error envelope; nothing is stored.
Traces: SC-20 (ORD-003 AC-16; contract §2.5.2)
Check: fetch_url

**Scenario**: Refused orders leave no trace: no order, no line, no stock change, no number spent
**Given**: Seed state R. One accepted order O1 (`V`) so the counter stands at 0001; `talon-2` and `stance-0` stock read (S1, S2) and their ETags.
**When**: These are posted: a 400 (name missing), a 400 (`qty` 11), a 400 (cash on delivery with courier), a 409 (`stance-0` fine, then `pique-advanced-sl-se` with stock 0), a 409 (`talon-2` quantity 10 over stock 9), a 409 (size "XXL"), and a bot 202. Then an accepted order O2 (`V`).
**Then**: After the refusals `GET /v1/products/talon-2` and `/stance-0` show stock S1 and S2 and the same ETags; `orders` holds one row and `order_lines` one row (SQL). O2's number is O1's number plus exactly one (`…-0002`, no gap), and the database then holds two orders.
Traces: SC-21 (ORD-003 AC-17; ORD-004 AC-13; ORD-006 refused order issues no number)
Check: fetch_url with the database read by SQL in the verification step

## 4. A bike that cannot be had is refused, and nothing is stored (ORD-004)

Every 409 is `code` `BIKE_UNAVAILABLE` in the error envelope, with no `field`. The sentences are fixed (ORD-004) and matched word for word. `<Brand> <Model>` is the record's `brand`, a space, the record's `name`, read first from `GET /v1/products/{id}`.

**Scenario**: A bike in draft is refused as no longer in the shop
**Given**: Seed state R with `talon-2` set to `status` "draft" with `set_db_state` (put back at the end).
**When**: 1 × `talon-2` size M is posted.
**Then**: 409 `BIKE_UNAVAILABLE` with `error` exactly "Giant Talon 2 is no longer in the shop."; `talon-2`'s stock is still 9; no order is stored.
Traces: SC-22 (ORD-004 AC-1)
Check: fetch_url

**Scenario**: A bike id that does not exist is refused with the generic bag sentence
**Given**: Seed state R; no bike `no-such-bike`.
**When**: 1 × `no-such-bike` (no `variant`) is posted.
**Then**: 409 `BIKE_UNAVAILABLE` with `error` exactly "A bike in your bag is no longer in the shop."
Traces: SC-22 (ORD-004 AC-9)
Check: fetch_url

**Scenario**: A bike with stock 0 is refused as out of stock, naming it
**Given**: Seed state R; `pique-advanced-sl-se` live with stock 0; its brand B and name N read from its record.
**When**: 1 × `pique-advanced-sl-se` in a size the bike has (`M`) is posted.
**Then**: 409 `BIKE_UNAVAILABLE` with `error` exactly "B N is out of stock." (the sentence names the bike; there is no ": only 0 in stock" form).
Traces: SC-23 (ORD-004 AC-2)
Check: fetch_url

**Scenario**: A quantity above the stock is refused with the stock in the sentence
**Given**: Seed state R with `talon-2` stock set to 3 (put back at the end).
**When**: 4 × `talon-2` size M is posted.
**Then**: 409 `BIKE_UNAVAILABLE` with `error` exactly "Giant Talon 2: only 3 in stock."; the stock is still 3; no order is stored.
Traces: SC-23 (ORD-004 AC-3)
Check: fetch_url

**Scenario**: A quantity equal to the stock is accepted and takes the stock to 0
**Given**: Seed state R with `talon-2` stock set to 3 (put back at the end).
**When**: 3 × `talon-2` size M is posted; then 1 × `talon-2` size M is posted.
**Then**: The first is 201 and `GET /v1/products/talon-2` then reads `stock` 0. The second is 409 `BIKE_UNAVAILABLE` "Giant Talon 2 is out of stock."
Traces: SC-23 (ORD-004 AC-4; ORD-004.5; ORD-005 boundary)
Check: fetch_url

**Scenario**: A size the bike does not have, a size the owner marked out, and a missing size are refused
**Given**: Seed state R; `talon-2` sizes XS S M L XL. Second arrangement: `talon-2` `sizes_out` set to `["XL"]` (put back at the end).
**When**: With `sizes_out` empty, posted: (a) 1 × `talon-2` size "XXL"; (b) 1 × `talon-2` with no `variant` key; (c) 1 × `talon-2` size "m" (lower case). With `sizes_out` `["XL"]`, posted: (d) 1 × `talon-2` size "XL".
**Then**: All four are 409 `BIKE_UNAVAILABLE`. (a) "Giant Talon 2 is not available in size XXL." (b) "Giant Talon 2 needs a frame size." (c) "Giant Talon 2 is not available in size m." (sizes are compared exactly). (d) "Giant Talon 2 is not available in size XL." Stock is unchanged after each.
Traces: SC-24 (ORD-004 AC-5, AC-6, AC-7)
Check: fetch_url

**Scenario**: The quantities of one bike on several lines are counted together
**Given**: Seed state R with `talon-2` stock set to 3 (put back at the end).
**When**: Posted: (a) two lines, `talon-2` size M qty 2 and `talon-2` size S qty 2; (b) two lines, `talon-2` size M qty 2 and `talon-2` size S qty 1.
**Then**: (a) is 409 `BIKE_UNAVAILABLE` "Giant Talon 2: only 3 in stock." and the stock is still 3. (b) is 201 with two lines and the stock is then 0.
Traces: SC-25 (ORD-004 AC-8)
Check: fetch_url

**Scenario**: One failing line refuses the whole order and no other bike's stock changes
**Given**: Seed state R; `stance-0` stock S2 read; `pique-advanced-sl-se` stock 0, brand B and name N read.
**When**: An order is posted whose first line is 1 × `stance-0` size M (fine) and whose second line is 1 × `pique-advanced-sl-se` size M.
**Then**: 409 `BIKE_UNAVAILABLE` "B N is out of stock." (about the second bike); `GET /v1/products/stance-0` still shows stock S2; no order exists (the next accepted order's number is `…-0001`).
Traces: SC-26 (ORD-004 AC-10; ORD-004.3)
Check: fetch_url

**Scenario**: When two lines fail, the sentence names the first line's bike
**Given**: Seed state R; `pique-advanced-sl-se` stock 0; a second arrangement with `talon-2` set to `status` "draft" (put back at the end).
**When**: Posted: (a) first line 1 × `pique-advanced-sl-se` size M, second line 1 × `talon-2` size M; (b) first line 1 × `talon-2` size M, second line 1 × `pique-advanced-sl-se` size M.
**Then**: (a) is 409 "B N is out of stock." naming the Pique bike; (b) is 409 "Giant Talon 2 is no longer in the shop." naming `talon-2`. Each names the first failing line's bike, in the order the lines were sent.
Traces: SC-26 (ORD-004 AC-11; ORD-004.2)
Check: fetch_url

**Scenario**: A bike with no frame sizes is accepted without a size
**Given**: Seed state R. A live bike `test-no-sizes` (Giant, name "Test No Sizes", price 7000000, stock 5, no `sizes` value) inserted with `set_db_state` (removed at the end).
**When**: Posted: (a) 1 × `test-no-sizes` with no `variant` key; (b) 1 × `test-no-sizes` with `"variant":"M"`.
**Then**: Both are 201. In (a) the answer's line has no `variant` key (`id`, `name` "Giant Test No Sizes", `price` 7000000, `qty` 1 and, when the bike has a photograph, `image`). In (b) the size is ignored: the answer's line has no `variant` key and the database holds `size` `''` for that line. Stock is 3 afterwards.
Traces: SC-27 (ORD-004 AC-12; ORD-004.4; contract `lines[].variant`)
Check: fetch_url with the database read by SQL in the verification step

**Scenario**: A bike with no brand is named by its model alone
**Given**: Seed state R. A live bike `test-no-brand` (name "Test No Brand", no `brand` value, price 7000000, stock 5, sizes `["M"]`) inserted with `set_db_state` (removed at the end).
**When**: 1 × `test-no-brand` size M is posted; then 6 × `test-no-brand` size M.
**Then**: The first is 201 with the line's `name` exactly "Test No Brand". The second is 409 "Test No Brand: only 4 in stock." (no leading space, no "undefined").
Traces: contract (POST /v1/orders, `<Brand> <Model>` rule; ORD-004.1)
Check: fetch_url

## 5. An accepted order holds the bike (ORD-005)

**Scenario**: An accepted order lowers the bike's stock by the quantity and moves its version
**Given**: Seed state R. `GET /v1/products/talon-2` read: `stock` 9 and ETag E1.
**When**: 2 × `talon-2` size M is posted (201); then `GET /v1/products/talon-2` is read again.
**Then**: The record's `stock` is 7; its ETag E2 differs from E1 (the hold bumps the version, so an owner's stale save is a 409 — ADMIN-008); no other key of the record changed (`status`, `price`, `sizes`, `sizes_out` as before).
Traces: SC-28 (ORD-005 AC-1; ORD-005.1; contract §1 writers)
Check: fetch_url

**Scenario**: An order for two bikes lowers each bike's stock by one
**Given**: Seed state R; `talon-2` stock 9 and `stance-0` stock S2 (7).
**When**: One order of 1 × `talon-2` size M and 1 × `stance-0` size M is posted.
**Then**: 201; `GET /v1/products/talon-2` reads `stock` 8 and `GET /v1/products/stance-0` reads `stock` S2 − 1.
Traces: SC-28 (ORD-005 AC-3)
Check: fetch_url

**Scenario**: An order never marks a size out
**Given**: Seed state R with `talon-2` stock set to 1 and `sizes_out` `[]` (put back at the end).
**When**: 1 × `talon-2` size XL is posted.
**Then**: 201; the record's `stock` is 0 and its `sizes_out` is still `[]` (and `sizes` still XS S M L XL).
Traces: SC-28 (ORD-005 AC-8; ORD-005.4)
Check: fetch_url

**Scenario**: The last bike sold shows "Out of stock" on its page, with no "Add to bag"
**Given**: Seed state R; `anthem-advanced-sl-0` live with stock 1; its page `/shop/anthem-advanced-sl-0` read first shows "In stock" and a button "Add to bag".
**When**: 1 × `anthem-advanced-sl-0` size M is posted (201) and the page `/shop/anthem-advanced-sl-0` is loaded again.
**Then**: `GET /v1/products/anthem-advanced-sl-0` reads `stock` 0; the page's stock line reads "Out of stock" and the page has no button reading "Add to bag" (the button reads "Ask when it is back").
Traces: SC-29 (ORD-005 AC-2; ORD-005.3)
Check: render_url

**Scenario**: After an order the bike's page shows the lowered stock
**Given**: Seed state R with `talon-2` stock set to 5 (put back at the end); `/shop/talon-2` read first shows "In stock" (a stock above 3 shows no count: the design's StockLine draws "· N left" only for 1 to 3 left).
**When**: 3 × `talon-2` size M is posted (201) and `/shop/talon-2` is loaded again.
**Then**: `GET /v1/products/talon-2` reads `stock` 2; the page's stock line reads "In stock" followed by "· 2 left". A second check with the seed stock: from 9, one order of 2 leaves stock 7 and the page reads "In stock" with no "left" text.
Traces: SC-29 (ORD-005 AC-9; ORD-005.3)
Check: render_url

**Scenario**: Two orders for the last bike at the same moment: one accepted, one refused, stock 0
**Given**: Seed state R with `talon-2` stock set to 1 (put back at the end).
**When**: Two orders for 1 × `talon-2` size M are posted at the same moment.
**Then**: Exactly one answer is 201 and the other is 409 `BIKE_UNAVAILABLE` ("Giant Talon 2 is out of stock."); `GET /v1/products/talon-2` reads `stock` 0 (not -1); the database holds exactly one order and one line.
Traces: SC-30 (ORD-005 AC-4; ORD-005.2)
Check: fetch_url (parallel requests)

**Scenario**: Twenty orders for the last bike at the same moment: one accepted, nineteen refused
**Given**: Seed state R with `talon-2` stock set to 1 (put back at the end).
**When**: Twenty orders for 1 × `talon-2` size M are posted at the same moment.
**Then**: One answer is 201 and nineteen are 409 `BIKE_UNAVAILABLE`; the stock is 0; the database holds one order and one line; no answer is a 5xx.
Traces: SC-30 (ORD-005 AC-5; ORD §6 Integrity)
Check: fetch_url (parallel requests)

**Scenario**: Five orders of two for a stock of nine: four accepted, one refused, stock 1
**Given**: Seed state R; `talon-2` stock 9.
**When**: Five orders for 2 × `talon-2` size M are posted at the same moment.
**Then**: Four answers are 201 and one is 409 `BIKE_UNAVAILABLE` ("Giant Talon 2: only 1 in stock." or "… only 1 in stock" for the stock the loser saw); `GET /v1/products/talon-2` reads `stock` 1; the database holds four orders.
Traces: SC-30 (ORD-005 AC-6)
Check: fetch_url (parallel requests)

**Scenario**: A failure after part of the work leaves no order and no changed stock
**Given**: Seed state R. One accepted order O1 first (counter at 0001); `talon-2` stock 8 read. A trigger on `order_lines` that raises on insert is arranged with `set_db_state` (removed at the end), so the failure comes after the stock was lowered and the order row written, inside the transaction.
**When**: 2 × `talon-2` size M is posted; then the trigger is removed and a valid order is posted.
**Then**: The failing request answers 500 `INTERNAL` with `error` exactly "Something went wrong. Please try again." (no SQL, no table name, no constraint name in the body). `talon-2`'s stock is still 8; the database holds only O1 and its line (no order row left by the failed request). The next accepted order's number is O1's plus one (`…-0002`: the counter rolled back too).
Traces: SC-31 (ORD-005 AC-7; ORD-005.5; contract §3.10 `INTERNAL`)
Check: fetch_url with the database read by SQL in the verification step

**Scenario**: When the database refuses connections the order is refused whole
**Given**: Seed state R; `talon-2` stock 9; the database made unreachable with `set_db_available` while the backend keeps running.
**When**: `V` is posted; then the database is made available again.
**Then**: The answer is 503 `UNAVAILABLE` with `error` "The service is not ready." (one generic sentence, in the error envelope); after the database is back, `talon-2`'s stock is still 9, the database holds no order, and a valid order is then accepted with number `…-0001`.
Traces: ORD-001 failure table "Upstream failure"; contract (POST /v1/orders `UNAVAILABLE`)
Check: fetch_url

## 6. The server issues the number (ORD-006)

**Scenario**: The first order of the day is 0001 and the second is 0002, dated in Vietnam time
**Given**: Seed state R (`order_counters` empty). The Vietnam date D (YYMMDD) is the date part of the answer's `placed` (which carries `+07:00`).
**When**: Two valid orders are posted one after the other.
**Then**: The first number is `DEO-D-0001` and the second `DEO-D-0002`, where D is the six digits of the `placed` date written YYMMDD (for 16 Nov 2026: `DEO-261116-0001`, `DEO-261116-0002`). `order_counters` holds one row for that date with `last_seq` 2.
Traces: SC-32 (ORD-006 AC-1, AC-2; ORD-006.1)
Check: fetch_url with the database read by SQL in the verification step

**Scenario**: A new day starts again at 0001
**Given**: Seed state R. `order_counters` holds a row for yesterday's Vietnam date with `last_seq` 7 and there is no row for today (arranged with `set_db_state`), so the day has just turned.
**When**: One valid order is posted.
**Then**: Its number begins with today's Vietnam date (`DEO-<YYMMDD of placed>-`) and ends `-0001`; yesterday's row is unchanged at 7 and today's row has `last_seq` 1. (The exact roll-over of the clock at midnight is the backend's injected-clock test, run with the backend tests.)
Traces: SC-32 (ORD-006 AC-4)
Check: fetch_url with the database read by SQL in the verification step

**Scenario**: The number's date is the Vietnam date, not the UTC date
**Given**: Seed state R.
**When**: One valid order is posted and the answer's `number` and `placed` are compared with the UTC clock read at the same moment.
**Then**: Always: the six digits of the number equal the date part of `placed` (offset `+07:00`) written YYMMDD. When the check runs between 17:00 and 23:59 UTC the Vietnam date is the UTC date plus one day and the number carries the later date — for an order stored at 18:00 UTC on 16 Nov 2026 the number begins `DEO-261117-`. The exact 18:00 UTC instant needs a clock the running service does not let verification set; it is proved by the backend's injected-clock unit test (story §6), and is reported as such.
Traces: SC-33 (ORD-006 AC-3)
Check: fetch_url

**Scenario**: Twenty orders at the same moment get twenty different numbers
**Given**: Seed state R with `talon-2` stock raised to 100 (put back at the end).
**When**: Twenty valid orders for 1 × `talon-2` size M are posted at the same moment.
**Then**: All twenty are 201 and their `number`s are twenty distinct values, forming the run `…-0001` to `…-0020` of one date with no duplicate and no gap; the database holds twenty orders.
Traces: SC-34 (ORD-006 AC-5; ORD-001 AC-7)
Check: fetch_url (parallel requests)

**Scenario**: Two accepted orders never share a number
**Given**: Seed state R.
**When**: Two valid orders are posted one after the other.
**Then**: Both are 201 and their `number`s differ.
Traces: SC-34 (ORD-001 AC-7)
Check: fetch_url

**Scenario**: A number carried by the request is ignored
**Given**: Seed state R. Order A is accepted and its number NA read.
**When**: `V` is posted with `"number":"DEO-000000-9999"`, and again with `"number":NA` (the number an existing order already has).
**Then**: Both are 201 (no unique violation, no 500). Neither answer's `number` equals the request's value or NA; each matches `^DEO-\d{6}-\d{4}$` and is the next number of the day (`…-0002`, `…-0003`).
Traces: SC-35 (ORD-006 AC-6; ORD-006.2)
Check: fetch_url

## 7. Speed (ORD §6)

**Scenario**: An order of twenty lines answers within one second at the 95th percentile
**Given**: Seed state R with `talon-2` stock raised to 5000 (put back at the end). The 20-line list of section 3 (sizes XS to XL, four times each, qty 1).
**When**: 50 orders with the 20-line list are posted one after the other; each response time is recorded.
**Then**: All 50 are 201 and the 95th percentile of the response times is at most 1 s; `talon-2`'s stock is 5000 − 1000 = 4000.
Traces: SC-36 (ORD §6 Performance; contract §6)
Check: fetch_url

## 8. The checkout sends the order and shows what the server says (ORD-007)

All cases in sections 8 and 9 run on the real site against the real backend, with demo mode off, and use the bag steps `B(n, size)`, the form step `F` and the recorder of section 0. The refusals are produced by real server states (stock changed after the bag was filled, a blank phone, a note over its limit, a failing or absent database), not by stubs. Words checked: the button reads "Place order" (and "Sending…" while sending); the not-connected line is "This form is not connected yet — please try again later."; the generic failure line is "The order could not be sent. Please try again."

**Scenario**: Pressing "Place order" sends one request with the order and no money
**Given**: Seed state R. Demo mode off. Bag `B(2, "M")` (2 × `talon-2` size M). On `/checkout`, step `F` with `#order-name` "Minh", `#order-phone` "0901234567", `#order-address` "12 Hùng Vương", `#order-note` "Call after 5pm", delivery `courier`, payment `transfer`. Recorder installed.
**When**: `[data-testid="place-order"]` is pressed.
**Then**: The recorder holds exactly one call to `<base>/v1/orders` (`<base>` is `VITE_API_URL` when set, else `/api`), method POST, content type JSON. Its body's keys are exactly `name`, `phone`, `delivery`, `address`, `payment`, `note`, `website`, `lines`, with the values "Minh", "0901234567", "courier", "12 Hùng Vương", "transfer", "Call after 5pm", `""` (the hidden field as it stands), and `lines` equal to `[{"id":"talon-2","variant":"M","qty":2}]` — each line has only the keys `id`, `variant`, `qty`. The body has no `price`, `total`, `subtotal`, `delivery_fee`, `number`, `placed`, line `name` or `image` anywhere. The call's response status is 201.
Traces: SC-37 (ORD-007 AC-1; ORD-007.1)
Check: interact_page

**Scenario**: A pickup order sends an empty address, and a line with no size sends no variant
**Given**: Seed state R plus the live bike `test-no-sizes` of section 4 (no sizes, stock 5). Bag holds 1 × `test-no-sizes` (added from `/shop/test-no-sizes`, which has no size pills). On `/checkout` the name and phone are typed and `[data-testid="deliver-pickup"]` is chosen (no address field is drawn). Recorder installed.
**When**: `[data-testid="place-order"]` is pressed.
**Then**: The one recorded request body has `"address":""`, `"delivery":"pickup"` and `lines` equal to `[{"id":"test-no-sizes","qty":1}]` (no `variant` key); the response is 201 and its `address` is `""`.
Traces: SC-37 (ORD-007 AC-1; story decision 7)
Check: interact_page

**Scenario**: While the order is being sent the button reads "Sending…", is disabled, and a second press sends nothing
**Given**: Seed state R with the *slow* trigger on `orders` (3 s). Bag `B(1, "M")`; `F` filled; recorder installed.
**When**: `[data-testid="place-order"]` is pressed, then — at once and again 1 s later — pressed again.
**Then**: Right after the first press the button's text is "Sending…" (followed by the price) and its `disabled` property is true, and its computed opacity is 0.6 where the instrument reports it. After the response arrives the recorder holds exactly one POST to `/v1/orders`, and the database holds exactly one order (the next accepted order's number is `…-0002`).
Traces: SC-38 (ORD-007 AC-2; ORD-007.2)
Check: interact_page

**Scenario**: A 201 empties the bag, saves the server's order unchanged and shows the confirmation page
**Given**: Seed state R (so the server's first number is `…-0001`, which the browser's old random four-digit number could not give). Bag `B(2, "M")`; `F` filled; recorder installed.
**When**: `[data-testid="place-order"]` is pressed and the page settles.
**Then**: The address is `/checkout/done`; `[data-testid="order-confirmed"]` is present; the bag holds no line (`localStorage` key `agentteam:cart` absent, and `/bag` shows its empty state "…" from `pages.bag.empty`). `sessionStorage` key `deo:last-order` parses to an object deep-equal to the 201 body the recorder captured: same `number` (matching `^DEO-\d{6}-0001$`), same `placed`, same `lines` (name "Giant Talon 2", not the bag's own name), same `subtotal`, `delivery_fee`, `total` — no field taken from the bag or the form.
Traces: SC-39 (ORD-007 AC-3; ORD-007.3)
Check: interact_page

**Scenario**: The confirmation page shows the server's number, lines, total and placed time, not the bag's
**Given**: Seed state R. Bag `B(2, "M")` filled while `talon-2` costs 18990000; then `talon-2`'s price is changed to 17990000 with `set_db_state` (put back at the end), after `/checkout` has shown the old total 37.980.000 ₫ (`[data-testid="checkout-total"]`); `F` filled.
**When**: `[data-testid="place-order"]` is pressed and `/checkout/done` is read.
**Then**: The page shows the server's number as the 2xl text under "Order number" (`DEO-<YYMMDD>-0001`); one line with the photograph `/images/ref/talon-2.jpg`, the name "Giant Talon 2", the text "Size M · 2 × 17.990.000 ₫" and the line total "35.980.000 ₫"; the total "35.980.000 ₫" (the server's price, not the bag's 37.980.000 ₫); and a date and time written `dd/mm/yyyy, hh:mm` that equals the saved order's `placed` instant formatted in the site's locale (not the moment the page was drawn).
Traces: SC-40 (ORD-008 AC-6; ORD-008.4; ORD-002 "price changed after the bag was filled")
Check: interact_page

**Scenario**: A reload of the confirmation page shows the same saved order, and nothing saved shows the empty state
**Given**: The end state of the previous case (an order saved in this browser's session).
**When**: `/checkout/done` is reloaded; then `sessionStorage` key `deo:last-order` is removed and `/checkout/done` is reloaded again.
**Then**: After the first reload the page shows the same `number` and total as before the reload. After the key is removed the page shows the heading "No order yet" and the line "Orders you place on this browser show here." and no order number.
Traces: SC-41 (ORD-007 AC-11)
Check: interact_page

**Scenario**: A 409 shows the server's sentence in the alert line and keeps the bag, the typed values and the choices
**Given**: Seed state R with `talon-2` stock set to 4 (put back at the end). Bag `B(4, "M")` (the product page's stepper allows 4). Then `talon-2`'s stock is changed to 3 with `set_db_state` (the bike changed between the bag and the order). On `/checkout`, `F` typed: `#order-name` "Minh", `#order-phone` "0901234567", `#order-address` "12 Hùng Vương", `#order-note` "Call after 5pm", delivery `courier`, payment `shop`. Recorder installed.
**When**: `[data-testid="place-order"]` is pressed and the page settles.
**Then**: The server answered 409. An element `[role="alert"]` reads exactly "Giant Talon 2: only 3 in stock." and its computed colour equals the colour of the theme's `--danger` token. The address is still `/checkout`. The bag still holds 4 × `talon-2` size M (`localStorage` `agentteam:cart`). `#order-name` is "Minh", `#order-phone` "0901234567", `#order-address` "12 Hùng Vương", `#order-note` "Call after 5pm"; `[data-testid="deliver-courier"]` and `[data-testid="pay-shop"]` still have `aria-checked` "true". The button reads "Place order" with its price and is not disabled. `deo:last-order` is not saved.
Traces: SC-42 (ORD-007 AC-4; ORD-007.4)
Check: interact_page

**Scenario**: A 400 shows the server's sentence and keeps everything
**Given**: Seed state R. Bag `B(1, "M")`. On `/checkout` the name "Minh", the address "12 Hùng Vương" and `#order-phone` set to three spaces (it passes the browser's required check, and the server trims it to nothing). Second arrangement: a valid phone and `#order-note` of 1001 characters.
**When**: `[data-testid="place-order"]` is pressed (once for each arrangement).
**Then**: Each answer is a 400. With the blank phone the `[role="alert"]` line reads the server's sentence containing "phone" ("Your phone number is needed."); with the long note it reads the server's sentence that says a text is too long. In both the address is still `/checkout`, the bag still holds its line, the typed name, address and note are still in their fields, and the button is enabled.
Traces: SC-42 (ORD-007 AC-5)
Check: interact_page

**Scenario**: A failed transaction (500) shows the generic sentence and keeps the bag
**Given**: Seed state R with the *fails mid-way* trigger on `order_lines`. Bag `B(1, "M")`; `F` filled with delivery `local`, payment `transfer`.
**When**: `[data-testid="place-order"]` is pressed.
**Then**: The server answered 500. `[role="alert"]` reads exactly "The order could not be sent. Please try again." (not the server's text); the address is still `/checkout`; the bag still holds its line; the typed values and the two chosen options are kept; the button is enabled; nothing is saved in `deo:last-order`.
Traces: SC-43 (ORD-007 AC-7; ORD-007.5)
Check: interact_page

**Scenario**: A database that refuses connections (503) and a service that is stopped (gateway 502) show the generic sentence
**Given**: Bag `B(1, "M")`; `F` filled. First arrangement: the database made unreachable with `set_db_available` (backend running → 503 `UNAVAILABLE`). Second arrangement: the backend stopped (the gateway answers 502). Everything put back at the end.
**When**: `[data-testid="place-order"]` is pressed under each arrangement.
**Then**: In both, `[role="alert"]` reads exactly "The order could not be sent. Please try again."; the address is `/checkout`; the bag still holds its line; typed values and choices are kept; the button is enabled again.
Traces: SC-43 (ORD-007 AC-6, AC-7; story decision 2)
Check: interact_page

**Scenario**: An answer of 413 shows the generic sentence and keeps the bag
**Given**: Seed state R. Bag `B(1, "M")`; `F` filled; `#order-note` set to 33,000 characters (the request body is over 32 KiB).
**When**: `[data-testid="place-order"]` is pressed.
**Then**: The server answered 413 `PAYLOAD_TOO_LARGE`; `[role="alert"]` reads exactly "The order could not be sent. Please try again."; the bag, the typed values and the address `/checkout` are kept.
Traces: SC-43 (ORD-007 AC-7; story decision 2: 413 shows the generic sentence)
Check: interact_page

**Scenario**: A request that gets no response at all, or an unreadable body, shows the generic sentence
**Given**: A browser whose `fetch` rejects (no network) or whose answer body is not JSON.
**When**: "Place order" is pressed.
**Then**: `[role="alert"]` reads "The order could not be sent. Please try again."; the bag and the typed values are kept. Reason `manual`: stopping the backend yields the gateway's 502 (covered above), so no instrument can make a loaded page's `fetch` reject or hand it an unreadable 201/409 body; the frontend unit test of `src/lib/orders.ts` with a mocked `fetch` (run with the frontend tests) covers it.
Traces: SC-43 (ORD-007 AC-6; story decision 2)
Check: manual

**Scenario**: After a refusal the old sentence is replaced by the next outcome
**Given**: The end state of the 409 case (alert "Giant Talon 2: only 3 in stock.", bag of 4, values typed). Recorder installed. The *slow* trigger on `orders` is arranged (3 s).
**When**: The bag is changed — the line's quantity is lowered to 3 with the header bag drawer's `[aria-label="One fewer"]` — and `[data-testid="place-order"]` is pressed again; the button state is read at once.
**Then**: While the request is in flight the button reads "Sending…" and there is no `[role="alert"]` element (the old sentence is gone); the recorder holds a second POST to `/v1/orders`. When it ends, the browser is at `/checkout/done` with the server's order (3 × `talon-2`, stock now 0) and the old sentence appears nowhere.
Traces: SC-44 (ORD-007 AC-8)
Check: interact_page

**Scenario**: After a refusal a different refusal replaces the first sentence
**Given**: Seed state R. The 409 state of the previous case again (alert "Giant Talon 2: only 3 in stock."), then `talon-2` set to `status` "draft" with `set_db_state` (put back at the end).
**When**: `[data-testid="place-order"]` is pressed again without changing anything else.
**Then**: A new POST is recorded; the single `[role="alert"]` now reads exactly "Giant Talon 2 is no longer in the shop." and the first sentence is gone (one alert line, not two); the bag and typed values are still kept.
Traces: SC-44 (ORD-007 AC-8)
Check: interact_page

**Scenario**: A 202 (bot drop) is not an order: generic sentence, nothing saved, no navigation
**Given**: Seed state R. Bag `B(1, "M")`; `F` filled; the hidden field `input[name="website"]` given the text "x" (a person's autofill could do this). Recorder installed.
**When**: `[data-testid="place-order"]` is pressed.
**Then**: The server answered 202 with `{}`. `[role="alert"]` reads exactly "The order could not be sent. Please try again."; the address is still `/checkout`; the bag still holds its line; `deo:last-order` is not saved; the recorder holds no call to the Agent Team `/messages` address; the database holds no order (the next accepted order is `…-0001`).
Traces: SC-45 (ORD-003 AC-15; ORD-007 AC-6; story decision 1)
Check: interact_page

**Scenario**: The change replaces the mock module and touches only the allowed files
**Given**: The PR diff for the story.
**When**: The file list and `src/lib/orders.ts` are read.
**Then**: The frontend files changed are `src/components/OrderForm.tsx` (changed), `src/lib/orders.ts` (added), `src/lib/mock/orders.ts` (deleted) and exactly one changed import line in each of `src/pages/Bag.tsx`, `src/pages/Checkout.tsx`, `src/pages/CheckoutDone.tsx` (`"../lib/mock/orders"` becomes `"../lib/orders"`), plus the backend files of story §7; `content.json`, `theme.css`, `index.css`, `App.tsx`, `site.ts` and `src/data/*` are untouched. `src/lib/orders.ts` exports `Order`, `saveOrder`, `lastOrder`, `EXAMPLE_ORDER`, `EXAMPLE_LINES` and the one sending function, and no function that makes an order number (`newOrderNumber` is gone); nothing imports `lib/mock/orders`; the frontend build, lint and tests pass (`run_tests` frontend and CI). Reason `manual`: a diff's file list and a module's export list are not observable on any running page.
Traces: SC-46 (ORD-007 AC-9, AC-10; ORD-007.6)
Check: manual

## 9. The order first, then the owner's message (ORD-008)

**Scenario**: The order is stored before the Agent Team message is requested
**Given**: Seed state R. Bag `B(2, "M")`; `F` filled. Recorder installed (it records start order and response status of each call).
**When**: `[data-testid="place-order"]` is pressed and the page settles.
**Then**: The recorder holds the call to `<base>/v1/orders` first, with response status 201, and only after that response a call to `<api>/<key>/messages` (the address from `src/site.json`); the second call starts after the first call's response arrived. The browser ends at `/checkout/done`.
Traces: SC-47 (ORD-008 AC-1; ORD-008.1)
Check: interact_page

**Scenario**: No Agent Team message is requested after any refusal
**Given**: Bag `B(1, "M")`; `F` filled; recorder installed. Four arrangements, one press each: `talon-2` stock set to 0 (→ 409 "Giant Talon 2 is out of stock."); the phone typed as three spaces (→ 400); the hidden field filled (→ 202); the backend stopped (→ gateway 502). Everything put back after each.
**When**: `[data-testid="place-order"]` is pressed under each arrangement.
**Then**: In every one the alert line shows the outcome of its section-8 case, and the recorder holds no call whose address ends `/messages`.
Traces: SC-48 (ORD-008 AC-2; ORD-008.1)
Check: interact_page

**Scenario**: The owner's message is built from the server's answer
**Given**: Seed state R. The customer is "Minh", phone "0901234567", address "12 Hùng Vương". Bag `B(2, "M")` filled at 18990000; then `talon-2`'s price changed to 17990000 (put back at the end). On `/checkout`: delivery `courier`, payment `transfer`, `#order-note` "Call after 5pm". Courier charge 350000. Recorder installed.
**When**: `[data-testid="place-order"]` is pressed.
**Then**: The recorded call to `<api>/<key>/messages` is a POST whose JSON body has `name` "Minh" and `reply_to` "0901234567"; `message` contains the line "2 × Giant Talon 2 (size M) — 35.980.000 ₫" (the server's price, not the bag's 37.980.000 ₫) and, after a blank line, "Call after 5pm"; `details` has the server's order number (equal to the 201 body's `number`, not one made by the browser), `delivery` "courier", `address` "12 Hùng Vương", `payment` "transfer" and `total` "36.330.000 ₫" (35.980.000 + 350.000); the body has no `website` key; `page` is "/checkout".
Traces: SC-49 (ORD-008 AC-3; story decision 6)
Check: interact_page

**Scenario**: An accepted order with no note has no empty note line in the message
**Given**: Seed state R. Bag `B(1, "M")`; `F` filled with `#order-note` left empty. Recorder installed.
**When**: `[data-testid="place-order"]` is pressed.
**Then**: The recorded message body's `message` is exactly "1 × Giant Talon 2 (size M) — 18.990.000 ₫" — one line, no trailing newline, no blank line, no empty note paragraph.
Traces: SC-49 (ORD-008 AC-7)
Check: interact_page

**Scenario**: When Agent Team refuses or cannot be reached the saved order still stands and nothing is shown
**Given**: Seed state R. The Agent Team address in `src/site.json` is not accepting the message in this environment (connection refused, or a non-2xx answer; the recorder shows which). Bag `B(2, "M")`; `F` filled.
**When**: `[data-testid="place-order"]` is pressed and the page settles.
**Then**: The recorder shows a call to `/messages` that failed (or was answered non-2xx) after the 201. The browser is at `/checkout/done` showing the server's number and lines; no `[role="alert"]` element was ever shown on the way (read at once after the press and after it settles); the bag is empty; `deo:last-order` holds the server's order; the database holds the order and `talon-2`'s stock is 7.
Traces: SC-50 (ORD-008 AC-4; ORD-008.3)
Check: interact_page

**Scenario**: With no line to the owner the order is still stored and the form says it is not connected
**Given**: A build whose `src/site.json` has no key (`canSend` false).
**When**: `/checkout` is read, then "Place order" is pressed with a valid form and bag.
**Then**: On `/checkout` the line "This form is not connected yet — please try again later." is shown above the button; after the press the order is stored (201), no call to `/messages` is made, the browser ends at `/checkout/done` with no error line and an empty bag. Reason `manual`: `canSend` is fixed when the bundle is built from `src/site.json` and the served site has a key, so no instrument can serve this build; the frontend unit test with a keyless `site.json` (run with the frontend tests) covers it.
Traces: SC-51 (ORD-008 AC-5; ORD-007 "not connected")
Check: manual

**Scenario**: When Agent Team never answers the customer still reaches the confirmation within ten seconds
**Given**: An accepted order whose Agent Team call receives no answer at all (an address that accepts the connection and never replies).
**When**: "Place order" is pressed and the clock is read.
**Then**: The button reads "Sending…" until the 201 and the wait, and the browser is at `/checkout/done` no later than 10 s after the 201, with the order saved, the bag empty, and exactly one call to `/messages` (no retry). Reason `manual`: the verification environment's Agent Team address cannot be made to hang, so the ten-second race is proved by the frontend unit test with a fake timer (story §6); if the address refuses at once, the arrival is immediate and the 10 s bound is trivially met, which the previous case already shows.
Traces: SC-52 (ORD-008 AC-4; story decision 4)
Check: manual

## 10. The service contract for `POST /v1/orders` (`services.md` §2.3, §2.5, §3.10)

These cases come from the contract rather than from a numbered criterion: the undefined-key rule, the content type, the body cap, wrong JSON types, the error envelope and the order in which checks answer.

**Scenario**: A body key the contract does not define is refused with 400
**Given**: Seed state R. `talon-2` stock 9.
**When**: `V` is posted (a) with the extra top-level key `"coupon":"SAVE10"`, (b) with the extra key `"color":"red"` on the line, (c) with `"id":"abc"` at order level, (d) with `"note":"hi"` plus the undefined key `"delivery_fee_extra":0`.
**Then**: Each is 400 `MALFORMED_REQUEST`, `field` naming the undefined key ("coupon", "color", "id", "delivery_fee_extra") and the sentence names it; nothing is stored (`talon-2`'s stock is still 9). Only the advisory keys (`number`, `placed`, `subtotal`, `delivery_fee`, `total` at order level; `name`, `price`, `image` on a line) are accepted and never read — they are covered by the case "The browser's own money, number and names are ignored".
Traces: contract (POST /v1/orders, §2.5.1 undefined keys, §3.10 advisory keys)
Check: fetch_url

**Scenario**: The content type must be JSON
**Given**: Seed state R.
**When**: `V` is posted as the raw body with (a) `Content-Type: text/plain`, (b) no `Content-Type` header, (c) `Content-Type: application/json; charset=utf-8`.
**Then**: (a) and (b) are 400 `MALFORMED_REQUEST` and nothing is stored; (c) is 201.
Traces: contract (POST /v1/orders step 1, §2.5.2)
Check: fetch_url

**Scenario**: A body over 32 KiB is refused with 413 before anything is validated
**Given**: Seed state R.
**When**: A valid JSON object is posted whose `note` is 40,000 characters (a body of about 40 KB), and then one whose `lines` array holds 5,000 valid-looking lines.
**Then**: Each is 413 `PAYLOAD_TOO_LARGE` in the error envelope (not 400 for the note's length, not 5xx); nothing is stored.
Traces: contract (POST /v1/orders, §6 payload cap)
Check: fetch_url

**Scenario**: Values of the wrong JSON type are malformed requests
**Given**: Seed state R.
**When**: `V` is posted with (a) `"lines":{"id":"talon-2"}`, (b) `"lines":"talon-2"`, (c) `"name":5`, (d) `"phone":["0901234567"]`, (e) `"lines":["talon-2"]`, (f) a line `"id":7`, (g) a line `"variant":3`.
**Then**: Each is 400 `MALFORMED_REQUEST` in the error envelope; nothing is stored.
Traces: contract (POST /v1/orders step 3; §2.5.3)
Check: fetch_url

**Scenario**: A line's bike id must be non-empty and at most 200 characters
**Given**: Seed state R.
**When**: `V` is posted with the line `"id":""`, then with a line `id` of 201 characters.
**Then**: Each is 400 `VALIDATION_FAILED` with `field` "lines[0].id"; an id of 200 characters is not refused as invalid — it is 409 `BIKE_UNAVAILABLE` "A bike in your bag is no longer in the shop." (no such bike).
Traces: contract (POST /v1/orders `lines[].id`)
Check: fetch_url

**Scenario**: Every non-2xx answer has the error envelope and the headers of the contract
**Given**: Seed state R. One refusal of each kind is produced: 400 `VALIDATION_FAILED`, 400 `MALFORMED_REQUEST`, 409 `BIKE_UNAVAILABLE`, 413 `PAYLOAD_TOO_LARGE`, 404 (`GET /v1/orders`), and 405 or 404 (`PUT /v1/orders/DEO-261116-0001`). The requests are sent once with the header `X-Request-Id: order-check-1` and once without it.
**When**: The responses are read (body and headers).
**Then**: Each body is a JSON object whose keys are only `error` (a non-empty sentence with no SQL, stack trace, file path or host), `code` (one of the contract's catalog, `UPPER_SNAKE`) and, only for a single field at fault, `field`. `Content-Type` is `application/json; charset=utf-8`; `Cache-Control` is `no-store`; `X-Request-Id` is `order-check-1` when it was sent and 32 hex characters when it was not. A 405 carries an `Allow` header.
Traces: contract (§2.3 error contract, §2.1 caching and trace header)
Check: fetch_url

**Scenario**: A request that breaks several rules gets the first failing check's answer
**Given**: Seed state R; `pique-advanced-sl-se` stock 0.
**When**: Posted: (a) no `name`, `"delivery":"drone"` and empty `lines`; (b) `"payment":"cod"` with `"delivery":"courier"` and no `address`; (c) `"delivery":"local"` with no `address` and 21 lines; (d) `lines[0].qty` 0 with a bike that is out of stock (`pique-advanced-sl-se`); (e) `lines[0].qty` 11 and `lines[1].id` empty; (f) `"payment":"cod"` with `"delivery":"local"`, an address and an unknown bike `no-such-bike` (valid otherwise).
**Then**: (a) is 400 `field` "name"; (b) is 400 `field` "payment" naming cash on delivery (before the address); (c) is 400 `field` "address" (before `lines`); (d) is 400 `VALIDATION_FAILED` `field` "lines[0].qty" — validation answers before availability, so not 409; (e) is 400 `field` "lines[0].qty" (the first failing line in the order sent); (f) is 409 "A bike in your bag is no longer in the shop." (the value rules pass, then availability).
Traces: contract (POST /v1/orders "Checks, in this order"; §2.3 which check answers first)
Check: fetch_url

## 11. Coverage of the story's acceptance criteria

Every `SC-n` of the story has at least one case above, found by its `Traces:` line.

| SC | Cases (by scenario name) |
|---|---|
| SC-1 | A valid order is accepted and answered with the stored order, priced from the bike's record |
| SC-2 | The 201 body has exactly eleven keys and a placed time with an offset |
| SC-3 | The note is stored with the order and is not in the answer |
| SC-4 | The database holds the order and its line exactly as answered |
| SC-5 | Migration 20261118000300 has created the three tables…; Deleting a bike leaves its order lines unchanged; The down migration removes the three order tables… |
| SC-6 | No endpoint reads, lists, changes or cancels an order, and no answer carries customer data |
| SC-7 | The browser's own money, number and names are ignored |
| SC-8 | A courier order carries the shop's courier charge; Local delivery and pickup carry no delivery fee |
| SC-9 | A changed courier charge applies to the next courier order |
| SC-10 | Two bikes in one order are summed from their own prices |
| SC-11 | A price the owner changes applies to the next order and an earlier order keeps its price |
| SC-12 | A missing name and a name of only spaces…; A missing phone is refused… |
| SC-13 | Empty or missing lines are refused; Twenty lines are accepted and twenty-one are refused |
| SC-14 | A quantity of 1 and of 10 are accepted; A quantity of 0, 11, 1.5, -1, 2.0 or "2" is refused |
| SC-15 | A delivery method or payment method outside the closed sets is refused |
| SC-16 | Cash on delivery is refused for courier and pickup and accepted for local delivery |
| SC-17 | An address is required unless the customer collects from the shop |
| SC-18 | Name, phone, address and note are accepted at their limits; One character past each text limit…; A frame size of 20 characters…; Text is trimmed…; Limits count characters, not bytes |
| SC-19 | A filled hidden field is dropped with 202…; The bot check comes before every other check… |
| SC-20 | A body that is not a JSON object, or not JSON, is refused |
| SC-21 | Refused orders leave no trace… |
| SC-22 | A bike in draft is refused…; A bike id that does not exist is refused… |
| SC-23 | A bike with stock 0…; A quantity above the stock…; A quantity equal to the stock is accepted… |
| SC-24 | A size the bike does not have, a size the owner marked out, and a missing size are refused |
| SC-25 | The quantities of one bike on several lines are counted together |
| SC-26 | One failing line refuses the whole order…; When two lines fail, the sentence names the first line's bike |
| SC-27 | A bike with no frame sizes is accepted without a size |
| SC-28 | An accepted order lowers the bike's stock…; An order for two bikes…; An order never marks a size out |
| SC-29 | The last bike sold shows "Out of stock"…; After an order the bike's page shows the lowered stock |
| SC-30 | Two orders for the last bike…; Twenty orders for the last bike…; Five orders of two for a stock of nine… |
| SC-31 | A failure after part of the work leaves no order and no changed stock |
| SC-32 | The first order of the day is 0001…; A new day starts again at 0001 |
| SC-33 | The number's date is the Vietnam date, not the UTC date |
| SC-34 | Twenty orders at the same moment get twenty different numbers; Two accepted orders never share a number |
| SC-35 | A number carried by the request is ignored |
| SC-36 | An order of twenty lines answers within one second at the 95th percentile |
| SC-37 | Pressing "Place order" sends one request with the order and no money; A pickup order sends an empty address… |
| SC-38 | While the order is being sent the button reads "Sending…"… |
| SC-39 | A 201 empties the bag, saves the server's order unchanged and shows the confirmation page |
| SC-40 | The confirmation page shows the server's number, lines, total and placed time, not the bag's |
| SC-41 | A reload of the confirmation page shows the same saved order… |
| SC-42 | A 409 shows the server's sentence…; A 400 shows the server's sentence and keeps everything |
| SC-43 | A failed transaction (500)…; A database that refuses connections (503) and a service that is stopped (gateway 502)…; An answer of 413…; A request that gets no response at all… |
| SC-44 | After a refusal the old sentence is replaced by the next outcome; After a refusal a different refusal replaces the first sentence |
| SC-45 | A 202 (bot drop) is not an order… |
| SC-46 | The change replaces the mock module and touches only the allowed files |
| SC-47 | The order is stored before the Agent Team message is requested |
| SC-48 | No Agent Team message is requested after any refusal |
| SC-49 | The owner's message carries the order's lines…; The owner's message is built from the server's answer…; An accepted order with no note has no empty note line… |
| SC-50 | When Agent Team refuses or cannot be reached the saved order still stands and nothing is shown |
| SC-51 | With no line to the owner the order is still stored and the form says it is not connected |
| SC-52 | When Agent Team never answers the customer still reaches the confirmation within ten seconds |
| contract | A body key the contract does not define…; The content type must be JSON; A body over 32 KiB…; Values of the wrong JSON type…; A line's bike id…; Every non-2xx answer has the error envelope…; A request that breaks several rules…; A bike with no brand…; When the database refuses connections… |

## 12. Roles, risk and what is not covered

- **Roles.** The SRS names a customer, the shop owner and a bot. The customer needs no sign-in: every case above is sent with no cookie and no `Authorization` header, so "no account needed" is exercised by every accepted order. The shop owner has no way to read an order through the site: that denial is the case "No endpoint reads, lists, changes or cancels an order…". The bot is the cases "A filled hidden field is dropped…", "The bot check comes before every other check…" and the UI case "A 202 (bot drop) is not an order…".
- **Manual cases and why.** Four cases are `manual` because no instrument can observe them on the running story: the down migration (running it destroys the live tables), a `fetch` that rejects or returns an unreadable body (a loaded page's network cannot be broken, only the gateway's 502 can), a build with no Agent Team key (`canSend` is fixed at build time), and a hanging Agent Team (the address cannot be made to hang). The PR file list and exports case is `manual` too (a diff is not observable on a page). Each has a unit-test counterpart named in the case.
- **Not covered, on purpose.** A rate limit (none at launch, `services.md` §9 question 2); duplicate-order detection and `Idempotency-Key` (decided against, SRS §7 question 2); a warning before an order when the bag's price is old (the design draws none); releasing a held bike, an orders screen, emailing the customer and card payment (all out of scope in the SRS).

## 13. Findings against the requirements (reported, not resolved silently)

1. **"In stock · N left" is not drawn above three.** ORD-005.3 and SC-29 say the bike's status reads "In stock · N left" after an order, but the approved StockLine (`design-system.md` §2.13, `src/components/StockLine.tsx`) draws "· N left" only when 1 to 3 remain, and plain "In stock" above that. The cases follow the design: they arrange a stock of 5 so that 3 sold leaves "· 2 left", and check that 9 − 2 = 7 reads "In stock" with the exact figure in `GET /v1/products/talon-2`. If the owner wants the figure shown above three, that is a design change for the `site` module, not for this story.
2. **Backend test seams.** The number's date at 18:00 UTC (SC-33) and the 10 s wait of SC-52 need a clock the running service and the served site do not let verification set; they depend on the injectable clock and fake timer the story §6 asks for.
