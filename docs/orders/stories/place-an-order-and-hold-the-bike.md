# Story — Place an order and hold the bike

Module: `orders` · Plan item 3 · Priority P1
Requirements: ORD-001, ORD-002, ORD-003, ORD-004, ORD-005, ORD-006, ORD-007, ORD-008 (`docs/orders/SRS.md`)
Contract: `docs/architecture/services.md` §3.10 (`POST /v1/orders`), §2.3 (error envelope), §5 (Agent Team line); `docs/architecture/erd.md` §3.9–3.11 (`orders`, `order_lines`, `order_counters`)
Design: `design/design-system.md` (no new screen, no new component, no new token)
Test cases: `docs/orders/test-cases/place-an-order-and-hold-the-bike.md`

## 1. User story

*As a* customer with bikes in my bag (no account, no sign-in), *I want* pressing "Place order" to make the shop keep my order, price it from its own records and hold my bikes — and to read, in one sentence, which bike is the problem when one cannot be had — *so that* the bike I chose is mine until the shop's confirming call, and I am never told an order failed when it stood.

The shop owner is the second actor: each placed order reaches them as one message in Agent Team, and each bike's stock figure falls by what was ordered. The owner cannot read orders back through the site.

This is one slice, not two: the checkout that sends the order and the server that stores, prices, numbers and holds it are useless apart (today the order is only an object in the browser, so a price can be talked down and the last bike of a size can be ordered ten times).

## 2. Scope

### In scope

Backend (`code/backend`)
- Migration `20261118000300_orders` (`.up.sql` and `.down.sql`): tables `orders`, `order_lines`, `order_counters` exactly as `erd.md` §3.9–3.11 draws them (number unique, `DEO-YYMMDD-NNNN`; one row per day in `order_counters`; `order_lines.product_id` is a copy of the bike's id, not a foreign key).
- `POST /v1/orders` (the one route this story adds, one line in `internal/server` Router): validation, bot drop, pricing from the `products` and `shop` records, the four availability refusals, the stock hold and the insert in **one transaction**, the number from `order_counters`, and the `201` answer with the eleven keys of `Order`.
- The refusals: `400` (`VALIDATION_FAILED` / `MALFORMED_REQUEST`), `409` (`BIKE_UNAVAILABLE`), `202 {}` for a filled `website`, `413`, `500`/`503` with the generic sentence — all in the error envelope of `services.md` §2.3.

Frontend (`code/frontend`)
- `src/lib/orders.ts` (added) replaces `src/lib/mock/orders.ts` (deleted): keeps `Order`, `saveOrder`, `lastOrder`, `EXAMPLE_ORDER`, `EXAMPLE_LINES` and what they do; adds the one function that sends the order to `POST /v1/orders`; the browser no longer makes an order number.
- `src/components/OrderForm.tsx`: the submit sends the order to the server, shows the server's sentence on a refusal, then (only after a `201`) saves the server's order, sends the owner's message through `sendEnquiry`, empties the bag and goes to `/checkout/done`.
- One import line in each of `src/pages/Bag.tsx`, `src/pages/Checkout.tsx`, `src/pages/CheckoutDone.tsx` (`"../lib/mock/orders"` becomes `"../lib/orders"`) — nothing else in those pages changes.

### Out of scope (and why)

- The bag, the checkout page's fields and layout, the confirmation page's layout — already built (`site` module, SITE-007, SITE-008, SITE-009). Only the import line above touches those pages.
- Card payment on the site — deliberately not built (SRS §3); payment is transfer (a VietQR code is sent when the shop calls), cash on delivery in town, or card or cash at the shop.
- A management screen listing orders, any endpoint that reads, lists, changes or cancels an order, and emailing the customer — deliberately not built; the owner reads orders as messages in Agent Team.
- Releasing a held bike when an order falls through, and marking a frame size out of stock when an order is placed — the owner corrects stock and sizes in the management area (`admin` module, ADMIN-008 covers an owner's save racing an order's hold).
- A warning before the order when the bag's price is out of date — the design draws none; the order is priced at the shop's current price and the confirmation shows the server's total (ORD-002 failure table).
- Duplicate-order detection and an `Idempotency-Key` — decided against (SRS §7 question 2, `services.md` §2.6); a customer who presses again after a lost answer makes a second order and the confirming call catches it.
- A rate limit on `POST /v1/orders` — none at launch (`services.md` §9 question 2).
- Changes to `src/content.json`, `src/theme.css`, `src/index.css`, `src/App.tsx`, `src/site.ts` and `src/data/*` — the file list of §7 is closed. The failure sentence "The order could not be sent. Please try again." is a constant in `src/lib/orders.ts`, as `site.ts` already keeps its own fallback sentence in code.
- The Agent Team inbox itself and its key: the message goes through the existing `sendEnquiry` and `src/site.json`; no new account, key or setup step is needed.

## 3. UI scope

No new screen and no new state is drawn. The story changes what stands behind two screens the approved design already draws; every state below already exists in `design/design-system.md` §2 and `docs/site/SRS.md`.

| Screen | Files | States this story must make true |
|---|---|---|
| Checkout `/checkout` | `pages/Checkout.tsx` (one import line), `components/OrderForm.tsx` | default; **sending** (button "Sending…", disabled, `disabled:opacity-60`); **failed** (the server's sentence, or the generic sentence, in the butter panel's `role="alert"` line, `text-danger`, button enabled again); **not connected** (the `text-sm text-ink-soft` line "This form is not connected yet — please try again later." above the button, unchanged); empty bag (button disabled, unchanged) |
| Order received `/checkout/done` | `pages/CheckoutDone.tsx` (one import line) | **order** (now drawn from the server's number, lines, total and placed time); no order (unchanged) |
| Bag `/bag` | `pages/Bag.tsx` (one import line) | unchanged — it still reads `EXAMPLE_LINES` |

Rules the UI review holds the PR to:
- Tailwind classes that name the design's tokens only (`text-danger`, `text-ink-soft`, `bg-accent`…): no raw `#hex`, `rgb()`, arbitrary-value colour or palette colour from outside the theme. The alert line keeps its existing classes.
- No new words in the design: the only new sentence is the generic failure sentence the SRS fixes (ORD-007.5); the refusal sentences are the server's.
- After a refusal the typed name, phone, address and note, and the chosen delivery and payment, are still there (the form's fields are not reset; delivery and payment live in `Checkout`'s state).
- In demo mode (`?demo=1`, set by the Studio and the walk) an empty bag shows `EXAMPLE_LINES`, and the form sends the lines it is given; that is the design preview and not a customer path. A test that wants a real order fills the bag first.

## 4. Decisions this story makes

The SRS and the contract leave these open; they are decided here so Dev, Test and TL build to one reading. None needs the stakeholder.

1. **A `202` is not an order.** A `202 {}` (the server dropped the request as a bot) carries no order. The browser treats it as "could not be sent": the generic sentence in the alert line, no saved order, no emptied bag, no navigation, no message to the owner. (A person's autofill could fill the hidden field; they see the line and the bag is intact. ORD-007 specifies only 201, 409, 400, 5xx and network failure.)
2. **Which sentence shows.** Status `400` and `409` show the server's `error` sentence (it is fixed text or names the field); every other outcome — network failure, `5xx`, `413`, any other status, an unreadable body, a `400`/`409` with no `error` sentence — shows "The order could not be sent. Please try again."
3. **The sender lives in `src/lib/orders.ts`** (`services.md` §3.10: "Sent by `src/lib/orders.ts`"). ORD-007 AC-10 lists the five names the module keeps; the one function that sends the order is the only export added, and no function makes an order number.
4. **The owner's message waits at most 10 s.** After the `201`, `sendEnquiry` is raced against a 10 s timer (`services.md` §5: `sendEnquiry` takes no abort signal); whichever finishes first, the customer goes on to `/checkout/done`. There is no retry of the message.
5. **No line to the owner** (`canSend` false): the order is still sent and stored; the message is not attempted (today `sendEnquiry` would throw and block the order — that behaviour goes). The "not connected" line stays visible above the button as designed.
6. **The message is built from the server's answer, not the bag**: name and reply-to from the answer's `name` and `phone`, the number, lines and total from the answer, the note from the form (the answer has none). It keeps the shape `sendEnquiry` has today: `message` (one line per bike, then a blank line and the note when there is one), `details` `{ order, delivery, address, payment, total }`. The bot field is not sent with it (an accepted order's `website` was empty).
7. **`address` for a pickup** is sent as an empty string (the field is not drawn); the server stores `""`.

## 5. Acceptance criteria

Numbered `SC-n`; each names the requirement and the SRS criterion it narrows. Test writes at least one case per criterion; TL's test-case review, the UI review and the backend review count this list. "The database is read" means a direct SQL read in the verification step. Every `409` and `400` is in the error envelope of `services.md` §2.3 (`error`, `code`, optional `field`); tests match the status and the quoted key words, and the **fixed** sentences of ORD-004 verbatim.

Seed used below: bike `talon-2` — brand "Giant", name "Talon 2", price 18990000, stock 9, sizes XS to XL, image `/images/ref/talon-2.jpg`; the shop's courier charge 350000; a valid customer is name "Minh", phone "0901234567", delivery `local`, address "12 Hùng Vương", payment `transfer`.

### 5.1 The order is stored and answered (ORD-001)

- **SC-1 [ORD-001 AC-1, AC-5]:** Placing an order for 2 × `talon-2` in size M answers `201` with `lines` equal to exactly one line `{ "id": "talon-2", "name": "Giant Talon 2", "price": 18990000, "image": "/images/ref/talon-2.jpg", "variant": "M", "qty": 2 }` (`image` equal to the bike record's `image` field), `subtotal` 37980000, `delivery_fee` 0 and `total` 37980000.
- **SC-2 [ORD-001 AC-2]:** The `201` body has exactly the keys `number`, `placed`, `name`, `phone`, `delivery`, `address`, `payment`, `lines`, `subtotal`, `delivery_fee`, `total`; `placed` parses as an ISO 8601 / RFC 3339 time with an offset (`+07:00`).
- **SC-3 [ORD-001 AC-3]:** Placed with the note "Please call after 5pm", the order is stored with that note and the `201` body has no `note` key.
- **SC-4 [ORD-001 AC-4]:** After an accepted order, the database holds one order whose name, phone, delivery, address, payment, note, subtotal, delivery fee and total equal the answer's, and one line holding `talon-2`, "Giant Talon 2", size M, quantity 2 and unit price 18990000.
- **SC-5 [ORD-001 AC-4]:** On a fresh database the service starts and migration `20261118000300_orders` applies; it creates `orders`, `order_lines` and `order_counters` and its `.down.sql` removes them without touching `products`; `order_lines` has no foreign key to `products`, so deleting a bike row leaves its order lines unchanged.
- **SC-6 [ORD-001 AC-6]:** With an accepted order and its number, asking for it by that number through any path (`GET /v1/orders/{number}`, `GET /v1/orders/{id}`, and the same with `PUT`/`DELETE`) is `404` or `405`; `GET /v1/orders` serves no list. Nothing in any answer carries a customer's name, phone, address or note.

### 5.2 The server prices the order (ORD-002)

- **SC-7 [ORD-002 AC-1]:** An order for 1 × `talon-2` whose body also carries `"total": 1`, `"subtotal": 1`, `"delivery_fee": 1`, `"number": "X"`, `"placed": "x"` and, on the line, `"price": 1`, `"name": "Cheap"`, `"image": "x"` is accepted; `lines[0].price` is 18990000, the name is "Giant Talon 2", `subtotal` and `total` are 18990000, and the number is the server's. (This is also the "price changed after the bag was filled" case: the order is priced at the shop's current price and accepted.)
- **SC-8 [ORD-002 AC-2, AC-3, AC-4]:** An order for 2 × `talon-2` with delivery `courier` and an address has `subtotal` 37980000, `delivery_fee` 350000, `total` 38330000; with delivery `local` `delivery_fee` is 0; with delivery `pickup` and no address `delivery_fee` is 0.
- **SC-9 [ORD-002 AC-5]:** After the shop's courier charge is changed to 400000, a courier order answers `delivery_fee` 400000.
- **SC-10 [ORD-002 AC-6]:** One order of 1 × `talon-2` and 1 × `stance-0` has `subtotal` equal to the sum of the two bikes' unit prices.
- **SC-11 [ORD-002 AC-7]:** After the owner changes `talon-2`'s price to 17990000, a new order's line is 17990000 and the database still holds the earlier order's line at 18990000.

### 5.3 What is refused as invalid (ORD-003)

- **SC-12 [ORD-003 AC-1, AC-2]:** A valid order with the name missing, and separately with the name "   ", is `400` naming the name; with the phone missing it is `400` naming the phone; no order is stored.
- **SC-13 [ORD-003 AC-3, AC-4, AC-5]:** `lines` empty or missing is `400`; an order of 20 lines (a bike repeated in different sizes, stock for all) is `201`; 21 lines is `400`.
- **SC-14 [ORD-003 AC-6, AC-7]:** A quantity of 1 and, separately, 10 is accepted (stock allowing); a quantity of 0, 11, 1.5, -1 and "2" is each `400` (`"2"` as `MALFORMED_REQUEST`, the rest as `VALIDATION_FAILED`, the sentence naming the quantity).
- **SC-15 [ORD-003 AC-8, AC-9]:** `delivery` "drone" is `400`; `payment` "card" is `400`.
- **SC-16 [ORD-003 AC-10, AC-11]:** Payment `cod` with delivery `courier`, and separately with `pickup`, is each `400` naming cash on delivery; payment `cod` with delivery `local` and an address is accepted.
- **SC-17 [ORD-003 AC-12, AC-13]:** Delivery `local` or `courier` with an empty or missing address is `400` naming the address; delivery `pickup` with no address is accepted.
- **SC-18 [ORD-003 AC-14]:** Text limits hold at the limit and one past it: name 100 accepted / 101 `400`; and likewise phone 30, address 300, note 1000 and a size of 20 characters. The limit counts the text after its surrounding spaces are trimmed, and the stored and answered text is the trimmed one (" Minh " is stored and answered as "Minh").
- **SC-19 [ORD-003 AC-15]:** An otherwise valid order whose hidden field `website` holds any text answers `202` with `{}`; no order is stored, no stock changes and no number is spent.
- **SC-20 [ORD-003 AC-16]:** A body that is not a JSON object (an array, a string, `null`) and a body that is not JSON are each `400`.
- **SC-21 [ORD-003 AC-17, ORD-004 AC-13]:** After any refused order (every `400` and `409` above and below) the database holds no new order and no new line, no bike's stock has changed, and the next accepted order that day gets the next number with no gap.

### 5.4 A bike that cannot be had is refused (ORD-004)

- **SC-22 [ORD-004 AC-1, AC-9]:** An order for `talon-2` while it is in draft is `409` "Giant Talon 2 is no longer in the shop."; an order for the id `no-such-bike` is `409` "A bike in your bag is no longer in the shop."
- **SC-23 [ORD-004 AC-2, AC-3, AC-4]:** An order for 1 of `pique-advanced-sl-se` (stock 0) is `409` "<Brand> <Model> is out of stock." naming that bike; an order for 4 of `talon-2` with stock 3 is `409` "Giant Talon 2: only 3 in stock."; an order for 3 with stock 3 is `201` and the stock is then 0.
- **SC-24 [ORD-004 AC-5, AC-6, AC-7]:** An order for `talon-2` in size "XXL" is `409` "Giant Talon 2 is not available in size XXL."; in size XL while XL is in `sizes_out` is `409` "Giant Talon 2 is not available in size XL."; with no size is `409` "Giant Talon 2 needs a frame size."
- **SC-25 [ORD-004 AC-8]:** With `talon-2` stock 3, an order of two lines — size M quantity 2 and size S quantity 2 — is `409` "Giant Talon 2: only 3 in stock." (the quantities of one bike are counted together).
- **SC-26 [ORD-004 AC-10, AC-11]:** An order whose first line is `stance-0` (fine) and second is a bike with stock 0 is `409` about the second bike, `stance-0`'s stock is unchanged and no order exists; when the first and second lines both fail, the sentence names the first line's bike.
- **SC-27 [ORD-004 AC-12]:** A bike that has no frame sizes is accepted with no size.

### 5.5 An accepted order holds the bike (ORD-005)

- **SC-28 [ORD-005 AC-1, AC-3, AC-8]:** An accepted order of 2 × `talon-2` (stock 9) leaves `talon-2`'s record with stock 7; an accepted order of 1 `talon-2` and 1 `stance-0` leaves each bike's stock one lower; an accepted order for `talon-2` size XL (not in `sizes_out`, stock 1) leaves `sizes_out` unchanged.
- **SC-29 [ORD-005 AC-2, AC-9]:** After an accepted order for the last `anthem-advanced-sl-0` (stock 1) its stock is 0 and, on reload, its page reads "Out of stock" with no "Add to bag"; after an accepted order for `talon-2`, reloading `talon-2`'s page shows the lowered stock ("In stock · N left").
- **SC-30 [ORD-005 AC-4, AC-5, AC-6]:** Placed at the same moment: two orders for 1 of a bike with stock 1 give one `201`, one `409`, stock 0 (not -1) and one stored order; twenty such orders give one `201`, nineteen `409`, stock 0, one stored order; five orders of 2 for a bike with stock 9 give four `201` and one `409`, stock 1.
- **SC-31 [ORD-005 AC-7]:** With a failure simulated after the order row is written and before the stock is lowered, the order answers a `5xx` with the generic sentence, no order row exists and the stock is unchanged.

### 5.6 The server issues the number (ORD-006)

- **SC-32 [ORD-006 AC-1, AC-2, AC-4]:** The first order stored on 16 Nov 2026 in Vietnam is `DEO-261116-0001`; a second the same day is `DEO-261116-0002`; the first order on the next day ends `-0001`.
- **SC-33 [ORD-006 AC-3]:** An order stored at 18:00 UTC on 16 Nov 2026 (01:00 on 17 Nov in Vietnam) has a number beginning `DEO-261117-`.
- **SC-34 [ORD-006 AC-5, ORD-001 AC-7]:** Twenty orders stored at the same moment get twenty different numbers; any two accepted orders have different numbers.
- **SC-35 [ORD-006 AC-6]:** A request that carries a `number` is placed with the server's number, not the request's.

### 5.7 Speed (ORD §6)

- **SC-36 [ORD §6 Performance]:** An order of 20 lines answers within 1 s at the 95th percentile over 50 sequential requests against the seeded database on the same host.

### 5.8 The checkout sends the order and shows what the server says (ORD-007)

- **SC-37 [ORD-007 AC-1]:** With a bag of 2 × `talon-2` size M and a valid form, pressing "Place order" makes exactly one `POST <base>/v1/orders` (`<base>` is `VITE_API_URL` when set, else `/api`) with a JSON body carrying `name`, `phone`, `delivery`, `address`, `payment`, `note`, `website` (as it stands, empty) and `lines` of `id`, `variant` (when the line has a size) and `qty` only — no price, total, subtotal, delivery fee, number, placed time, line name or image. For a pickup, `address` is `""`.
- **SC-38 [ORD-007 AC-2]:** While the request is in flight the button reads "Sending…" and is disabled; a second press sends nothing, so one press is one order.
- **SC-39 [ORD-007 AC-3]:** When the server answers `201` with number `DEO-261116-0001`, the bag is empty, the browser is at `/checkout/done`, and the order saved in this browser is the server's answer unchanged (no field taken from the bag or the form).
- **SC-40 [ORD-008 AC-6]:** `/checkout/done` after an accepted order shows the server's number, each of the server's lines (photograph, name, "Size M · 2 × 18.990.000 ₫", line total), the server's total, and the date and time the server stored the order (`placed`). When the server's price differs from the price the bag held, the page shows the server's.
- **SC-41 [ORD-007 AC-11]:** After a saved order, reloading `/checkout/done` shows the same order from this browser's saved copy; with nothing saved the "No order / yet" state shows as before.
- **SC-42 [ORD-007 AC-4, AC-5]:** When the server answers `409` "Giant Talon 2: only 3 in stock." (and, separately, `400` with a sentence), that sentence is in a `role="alert"` line in the danger colour in the form's butter panel, the bag still holds its lines, the typed name, phone, address and note are still in their fields, the chosen delivery and payment are still chosen, the page is still `/checkout` and the button is enabled again.
- **SC-43 [ORD-007 AC-6, AC-7]:** When the network is down, and separately when the server answers `500` (or `503`, `413`, or an answer whose body cannot be read), pressing "Place order" shows "The order could not be sent. Please try again." in the same alert line; the bag, the typed values and the choices are kept.
- **SC-44 [ORD-007 AC-8]:** After a refusal was shown, the customer changes the bag (or a field) and presses "Place order" again: a new request is sent, the old sentence is gone while it is sending and is replaced by the new outcome — the next refusal's sentence, or the confirmation page on a `201`.
- **SC-45 [ORD-003 AC-15, ORD-007 AC-6]:** When the server answers `202` with `{}` the browser behaves as in SC-43: the generic sentence, the bag kept, no order saved, no navigation and no message to the owner (decision 1).
- **SC-46 [ORD-007 AC-9, AC-10]:** The change's file list is `src/components/OrderForm.tsx` (changed), `src/lib/orders.ts` (added), `src/lib/mock/orders.ts` (deleted) and one import line in each of `src/pages/Bag.tsx`, `src/pages/Checkout.tsx`, `src/pages/CheckoutDone.tsx`, plus the backend files of §7; `src/lib/orders.ts` exports `Order`, `saveOrder`, `lastOrder`, `EXAMPLE_ORDER`, `EXAMPLE_LINES` with the behaviour they had, and the one sending function, and no function that makes an order number; nothing imports `lib/mock/orders` and the build, lint and tests pass.

### 5.9 The order first, then the owner's message (ORD-008)

- **SC-47 [ORD-008 AC-1]:** For a valid order, the request that stores the order (`POST …/v1/orders`) is made before the request to Agent Team, and the Agent Team request is made only after the `201` has arrived.
- **SC-48 [ORD-008 AC-2]:** When the server refuses the order with `409` (and likewise `400`, `202`, `5xx` or a network failure), no request to Agent Team is made.
- **SC-49 [ORD-008 AC-3, AC-7]:** For an accepted order for "Minh", phone "0901234567", 2 × Giant Talon 2 size M, delivery `courier`, address "12 Hùng Vương", payment `transfer`, note "Call after 5pm", the Agent Team request carries `name` "Minh", `reply_to` "0901234567", a `message` containing "2 × Giant Talon 2 (size M) — 37.980.000 ₫" and the note, and `details` with the server's order number, the delivery, the address "12 Hùng Vương", the payment and the total "38.330.000 ₫". It is built from the server's answer (the server's number, name, price and total), not from the bag. An accepted order with no note has no empty note line in the message.
- **SC-50 [ORD-008 AC-4]:** When the order was accepted and Agent Team refuses the message (non-2xx, or unreachable), the browser is at `/checkout/done` showing the server's order, no error line was shown and the bag is empty.
- **SC-51 [ORD-008 AC-5, ORD-007 not connected]:** When the site has no line to the owner (`canSend` false), pressing "Place order" still sends and stores the order, makes no Agent Team request, and ends as SC-50 (confirmation page, no error line, bag empty); on `/checkout` the line "This form is not connected yet — please try again later." is shown above the button as designed.
- **SC-52 [ORD-008 AC-4, decision 4]:** When the order was accepted and Agent Team does not answer, the browser goes to `/checkout/done` 10 s after the `201` at the latest (the button reads "Sending…" until then), with the order saved and the bag empty, and sends no second message.

## 6. Dependencies

- **Lands first — plan item 2, "Site data from the database"** (`docs/data/SRS.md`): the migrations `20261118000100` (collections, including the `products` table with `status`, `stock`, `sizes`, `sizes_out`, `brand`, `name`, `price`, `image` and the `touch_record` version trigger) and `000200` (the seed with `talon-2`, `stance-0`, `pique-advanced-sl-se`, `anthem-advanced-sl-0`, and the `shop` record with `delivery_fee` and the stock rule); the read API and the rewritten data module, without which SC-29 (the page shows the lowered stock) cannot be seen. This story's migration is slot `20261118000300`.
- **Runs after this story — plan item 4, "Shop admin"**: ADMIN-008 (an owner's save must not silently overwrite an order's hold) relies on this story's hold taking the bike's row lock and bumping `version` through the same trigger (`services.md` §1, §3.10). Nothing here waits on admin.
- **Unrelated, no ordering — plan item 5, "Bike page waits for its record"**: a bug in a page this story does not edit.
- **Already in the repository:** the bag (`src/lib/cart.ts`), `OrderForm`, `Checkout`, `CheckoutDone`, `src/site.ts` (`sendEnquiry`, `canSend`) and `src/site.json` (the Agent Team key). The edge proxy strips `/api`; locally `docker compose --profile local up --build --wait` gives the database.
- **External accounts and secrets:** none. No payment gateway, e-mail or SMS service; no `request_secret` or `request_setup` is raised by this story.
- **Testability the build must provide (for Dev and TL):** an injectable clock for the order number's day (SC-32, SC-33); a seam that can fail after the order row and before the stock is lowered (SC-31); the 10 s wait of SC-52 driven by a fake timer. CI has no database, so the criteria that need Postgres — SC-4, SC-5, SC-11, SC-21, SC-28 to SC-36 — are proved in the verification step against the running story, and everything else (validation, sentences, the frontend) runs without one.

## 7. Files this story changes

| Area | File | Change |
|---|---|---|
| Frontend | `code/frontend/src/components/OrderForm.tsx` | changed: sends the order, the order of events, the outcomes |
| Frontend | `code/frontend/src/lib/orders.ts` | added: `Order`, `saveOrder`, `lastOrder`, `EXAMPLE_ORDER`, `EXAMPLE_LINES`, the sending function, the generic failure sentence |
| Frontend | `code/frontend/src/lib/mock/orders.ts` | deleted (with `newOrderNumber`) |
| Frontend | `code/frontend/src/pages/Bag.tsx`, `Checkout.tsx`, `CheckoutDone.tsx` | one import line each: `"../lib/mock/orders"` becomes `"../lib/orders"` |
| Backend | `code/backend/migrations/20261118000300_orders.up.sql`, `.down.sql` | added |
| Backend | `code/backend/internal/orders/` | added: handler, validation, the one transaction, the number |
| Backend | `code/backend/internal/server` (Router) | one added line for `POST /v1/orders` |

Test files are Test's. Anything else in `code/frontend` (including `content.json`, `theme.css`, `index.css`, `App.tsx`, `site.ts`, `data/*`) belongs to the scaffold or to another story and is not edited.

## 8. Accepted risks (decided, not blocking)

Built to the defaults the SRS and the contract propose; revisit only if the stakeholder asks.

- When the owner's message cannot be sent, the order stands and nobody is told; the owner would see the stock fall (SRS §7 question 1). An orders screen can be added later.
- A second press after a lost answer makes a second order that holds a second bike (SRS §7 question 2).
- Anyone can place orders for every bike and make the shop look sold out until the owner corrects stock, because an order holds stock and nothing is paid (`services.md` §9 question 2). A per-address limit is an additive change if it happens.
