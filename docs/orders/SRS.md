# SRS — Orders (place an order and hold the bike)

Module: `orders`
Design: [View the approved design](http://localhost:9400)
Design system: `design/design-system.md`

> One file per module. This one covers what happens when a customer places an order: the server
> keeps it, prices it, checks the bike and holds it. The cart, the checkout page and the confirmation
> page are already built and are described in `docs/site/SRS.md`; this module changes what stands
> behind them.

## 1. Purpose

The brief asks for "a cart and online ordering" and for stock availability to be true. Today an order is
only an object in the customer's browser: its prices and total are whatever the browser says, the owner
learns of it only from one message, and stock never changes, so the last bike of a size can be ordered ten
times. This module makes the order a real record on the server, priced by the server, with the bike held
for it. Without it the shop can be talked into selling a bike it does not have, at a price it did not set.

## 2. Actors

| Actor | Who they are | What they may do in this module |
|---|---|---|
| Customer | A visitor with bikes in the bag, no sign-in | Place an order; see the order the server stored, on the confirmation page |
| Shop owner | Runs Đèo Bikes | Reads each placed order as a message in Agent Team; corrects stock in the management area when an order falls through. Cannot read orders back through the site |
| Bot | A program filling every field of a form | Is dropped without effect when it fills the hidden field |

## 3. Scope

**In scope** — the functions specified below, by their plan titles:

- Place an order and hold the bike

**Out of scope**

- The cart, the checkout page and its fields, the confirmation page's layout — already built; `site` module (SITE-007, SITE-008, SITE-009).
- Card payment on the site — deliberately not built; payment is bank transfer (a VietQR code is sent when the shop calls), cash on delivery in town, or card or cash at the shop.
- A management screen that lists orders — deliberately not built; the owner reads orders as messages in Agent Team.
- Emailing the customer — deliberately not built.
- Releasing a held bike when an order falls through — deliberately not built; the owner corrects stock in the management area (`admin` module).
- Reading an order back, listing orders, changing or cancelling an order — not offered by any endpoint.
- Marking a frame size out of stock when an order is placed — stock is one figure per bike; the owner marks sizes out (`admin` module).

## 4. Functional requirements

### 4.1 Place an order and hold the bike

**Requirement ORD-001 — Placing an order stores it and answers with what was stored**

*As a* customer, *I want* my order to be kept by the shop and confirmed back to me with its number, *so that* I know the shop has it.

Behaviour:

1. Placing an order sends the customer's name, phone, delivery method (`local`, `courier` or `pickup`), address, payment method (`transfer`, `cod` or `shop`), an optional note and the bikes — each as the bike's id, the chosen size and a quantity. The request carries no price and no total that the server would use.
2. When it is accepted, the server stores the order and its lines and answers 201 with the stored order: `number`, `placed`, `name`, `phone`, `delivery`, `address`, `payment`, `lines`, `subtotal`, `delivery_fee` and `total` — the same shape as `Order` in `src/lib/mock/orders.ts`. A line is `{ id, name, price, image, variant, qty }`: the bike's id, its name as the bag shows it (`<brand> <model>`), the unit price, the bike's main photograph, the chosen size and the quantity. The note is stored and is not part of the answer.
3. `placed` is the moment the server stored the order, as an ISO 8601 time with its offset.
4. The order and its lines persist: for each order the database keeps the number, the time placed, name, phone, delivery, address, payment, note, subtotal, delivery fee and total, and for each line the bike's id, name, size, quantity and unit price.
5. Reading an order back is not offered: no endpoint lists or returns orders.

**Acceptance criteria** — each is proved by at least one test case in `docs/orders/test-cases/place-an-order-and-hold-the-bike.md`.

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | Bike `talon-2` live with stock 9, name "Talon 2", brand "Giant", price 18990000, size M available; delivery `local`, payment `transfer`, name "Minh", phone "0901234567", address "12 Hùng Vương" | An order for 2 of `talon-2` in size M is placed | 201 with `lines` equal to one line `{ id: "talon-2", name: "Giant Talon 2", price: 18990000, image: "/images/ref/talon-2.jpg", variant: "M", qty: 2 }`, `subtotal` 37980000, `delivery_fee` 0, `total` 37980000 |
| AC-2 | The order of AC-1 | The answer is read | It has exactly the keys `number`, `placed`, `name`, `phone`, `delivery`, `address`, `payment`, `lines`, `subtotal`, `delivery_fee`, `total` and `placed` parses as a time with an offset |
| AC-3 | The order of AC-1 | The order is placed with the note "Please call after 5pm" | The note is stored with the order and is not in the answer |
| AC-4 | The order of AC-1 was accepted | The database is read | One order exists whose name, phone, delivery, address, payment, note, subtotal, delivery fee and total equal the answer's, with one line holding `talon-2`, "Giant Talon 2", size M, quantity 2 and unit price 18990000 |
| AC-5 | The order of AC-1 | The answer's `lines` are compared with the bike's record | `image` equals the bike's `image` field |
| AC-6 | An accepted order with a number | An order is asked for by that number through any endpoint | None is served (404 or 405) and no list of orders is served |
| AC-7 | Two accepted orders | They are compared | Their `number`s differ |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input, refused bike, refused stock | See ORD-003 and ORD-004 | 400 or 409; nothing is stored |
| Permission | — | Not applicable: placing an order needs no sign-in; the customer has no account |
| Upstream failure | The database is unreachable | A 5xx with one generic sentence in the error envelope; nothing is stored and no stock changes |

**Data touched**

| Field | Type | Required | Rule |
|---|---|---|---|
| number | text | yes | `DEO-YYMMDD-NNNN`, unique (ORD-006) |
| placed | time | yes | Set by the server |
| name | text | yes | At most 100 characters |
| phone | text | yes | At most 30 characters |
| delivery | text | yes | `local`, `courier` or `pickup` |
| address | text | when not `pickup` | At most 300 characters |
| payment | text | yes | `transfer`, `cod` or `shop` |
| note | text | no | At most 1000 characters; stored, not returned |
| subtotal, delivery_fee, total | whole đồng | yes | Computed by the server (ORD-002) |
| line: bike id | text | yes | A bike that exists |
| line: name | text | yes | The bike's `<brand> <model>` when ordered |
| line: size | text | when the bike has sizes | One of the bike's sizes |
| line: quantity | whole number | yes | 1 to 10 |
| line: unit price | whole đồng | yes | The bike's price when ordered |

**Requirement ORD-002 — The server prices the order; the browser's money is never trusted**

*As a* shop owner, *I want* every order priced from my records, *so that* no customer can pay a price I did not set.

Behaviour:

1. Each line's name and unit price are read from the bike's record when the order is stored. Any name or price the browser sends is ignored.
2. `subtotal` is the sum over the lines of unit price times quantity. `delivery_fee` is the shop's own courier charge (`delivery_fee` of the shop's record) when the delivery is `courier` and 0 otherwise. `total` is `subtotal` plus `delivery_fee`.
3. A price the owner changes applies to the next order placed; a line already stored keeps the price it was ordered at.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | `talon-2` priced 18990000 | An order for 1, with each line also carrying `"price": 1` and the order `"total": 1` | `lines[0].price` is 18990000, `subtotal` and `total` are 18990000; the extra keys have no effect |
| AC-2 | `talon-2` priced 18990000, courier charge 350000 | An order for 2 with delivery `courier` and an address | `subtotal` 37980000, `delivery_fee` 350000, `total` 38330000 |
| AC-3 | The same bikes | An order for 2 with delivery `local` | `delivery_fee` is 0 |
| AC-4 | The same bikes | An order for 2 with delivery `pickup` and no address | `delivery_fee` is 0 |
| AC-5 | The shop's courier charge is changed to 400000 | A courier order is placed | `delivery_fee` is 400000 |
| AC-6 | `talon-2` and `stance-0` | One order of 1 of each is placed | `subtotal` is the sum of the two unit prices |
| AC-7 | The owner changes `talon-2`'s price to 17990000 | An order for 1 is placed, then the earlier order is read from the database | The new order's line is 17990000 and the earlier order's line is unchanged |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Price changed after the bag was filled | The customer's bag holds an old price | The order is priced at the shop's current price and accepted; the confirmation shows the server's total (SITE-007 failure table). There is no warning before the order, as the design draws none |
| Other cases | — | Not applicable: no roles; invalid input is ORD-003 |

**Data touched** — bike price and name; the shop's courier charge; the order's subtotal, fee and total.

**Requirement ORD-003 — What the customer must give, and what is refused as invalid**

*As a* shop owner, *I want* an order refused when it is incomplete or inconsistent, *so that* I can call every customer and deliver every bike.

Behaviour:

1. Name and phone are required: each must hold at least one character other than spaces. At least one line and at most 20 are required; each line's quantity is a whole number from 1 to 10. A missing or wrong kind of value is refused.
2. `delivery` must be `local`, `courier` or `pickup`; `payment` must be `transfer`, `cod` or `shop`.
3. `cod` is accepted only with delivery `local`.
4. An address is required unless the delivery is `pickup`. For `pickup` none is needed.
5. Text is limited to: name 100, phone 30, address 300, note 1000, a size 20 characters; a longer text is refused. Spaces around a text are trimmed.
6. A refusal is a 400 with one sentence naming what is wrong, in the error envelope of the service contract; nothing is stored and no stock changes.
7. A filled hidden field `website` marks a bot: the request is dropped with a 202 and an empty answer; nothing is stored, no stock is held and the owner is told nothing.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A valid order | The name is missing, and separately is "   " | Each is 400 naming the name; no order is stored |
| AC-2 | A valid order | The phone is missing | 400 naming the phone |
| AC-3 | A valid order | `lines` is empty or missing | 400 |
| AC-4 | A valid order with 20 lines of distinct bikes | It is placed | 201 (20 lines is accepted) |
| AC-5 | A valid order with 21 lines | It is placed | 400 |
| AC-6 | A valid order | A quantity of 1 and, separately, 10 | Each is accepted (when stock allows) |
| AC-7 | A valid order | A quantity of 0, 11, 1.5, -1 and "2" | Each is 400 |
| AC-8 | A valid order | `delivery` is "drone" | 400 |
| AC-9 | A valid order | `payment` is "card" | 400 |
| AC-10 | An order with payment `cod` | Delivery `courier` and, separately, `pickup` | Each is 400 naming cash on delivery |
| AC-11 | An order with payment `cod` and delivery `local` and an address | It is placed | Accepted |
| AC-12 | An order with delivery `local` or `courier` | The address is empty | 400 naming the address |
| AC-13 | An order with delivery `pickup` | No address is given | Accepted |
| AC-14 | An order with a name of 100 characters, and one of 101 | Each is placed | 100 is accepted; 101 is 400 |
| AC-15 | An order | The hidden field `website` has any text | 202 with an empty object, no order stored, no stock changed |
| AC-16 | A body that is not an object, or not JSON | It is placed | 400 |
| AC-17 | Any refused order | The database is read | No order and no line were stored and no stock changed |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | Every row of Behaviour 1 to 5 | 400 with one sentence (AC-1 to AC-16) |
| Boundary | Lines 20 and 21; quantity 1, 10, 0 and 11; text lengths at and past their limits | At the limit accepted; one past, refused (AC-4 to AC-7, AC-14) |
| Bot | The hidden field is filled | Dropped with 202 (AC-15) |

**Data touched** — the order's fields and lines as listed in ORD-001.

**Requirement ORD-004 — A bike that cannot be had is refused, and nothing is stored**

*As a* shop owner, *I want* an order refused when a bike is not for sale, is short, or has no such size, *so that* I never sell what I cannot deliver.

*As a* customer, *I want* to be told which bike is the problem, *so that* I can change my bag.

Behaviour:

1. An order is refused with a 409 and one sentence naming the bike when, for any line, at the moment the order would be stored:
   - the bike does not exist: "A bike in your bag is no longer in the shop."
   - the bike is not live: "<Brand> <Model> is no longer in the shop."
   - the bikes' stock, counting every line of that bike together, is below the quantity asked: "<Brand> <Model> is out of stock." when its stock is 0, and "<Brand> <Model>: only <N> in stock." when it is above 0
   - the bike has frame sizes and the line's size is not one of them, is one the owner marked out, or is missing: "<Brand> <Model> is not available in size <size>." (for a missing size: "<Brand> <Model> needs a frame size.")
2. When several lines fail, the sentence is about the first failing line in the order the lines were sent.
3. A refused order is refused whole: no order or line is stored and no stock of any bike changes, however many of its other lines were fine.
4. A bike that has no frame sizes needs no size.
5. A quantity equal to the stock is accepted.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | Bike `talon-2` in draft | An order for it is placed | 409 "Giant Talon 2 is no longer in the shop."; no order stored |
| AC-2 | Bike `pique-advanced-sl-se` with stock 0 | An order for 1 | 409 "<Brand> <Model> is out of stock." naming it |
| AC-3 | Bike `talon-2` with stock 3 | An order for 4 | 409 "Giant Talon 2: only 3 in stock." |
| AC-4 | Bike `talon-2` with stock 3 | An order for 3 | 201 and the stock is 0 |
| AC-5 | Bike `talon-2` with sizes XS to XL | An order in size "XXL" | 409 "Giant Talon 2 is not available in size XXL." |
| AC-6 | Bike `talon-2` with size XL in `sizes_out` | An order in size XL | 409 "Giant Talon 2 is not available in size XL." |
| AC-7 | Bike `talon-2` with sizes | An order with no size | 409 "Giant Talon 2 needs a frame size." |
| AC-8 | Bike `talon-2` with stock 3 | An order with two lines, size M quantity 2 and size S quantity 2 | 409 "Giant Talon 2: only 3 in stock." |
| AC-9 | An id `no-such-bike` | An order for it | 409 "A bike in your bag is no longer in the shop." |
| AC-10 | An order whose first line is `stance-0` (fine) and second is a bike with stock 0 | It is placed | 409 about the second bike; `stance-0`'s stock is unchanged and no order exists |
| AC-11 | An order whose first and second lines both fail | It is placed | The sentence names the first line's bike |
| AC-12 | A bike with no frame sizes | An order with no size | Accepted |
| AC-13 | Any 409 | The database is read | No order, no line and no changed stock |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Conflict | The bike changes between the bag and the order (sold out, set to draft, size marked out) | The order is refused with the sentence for what is true now (AC-1 to AC-6) |
| Boundary | Quantity equal to stock; quantity one above | Accepted; refused (AC-3, AC-4) |
| Other cases | — | Not applicable: no roles |

**Data touched** — a bike's status, stock, sizes and sizes out, brand and name.

**Requirement ORD-005 — An accepted order holds the bike**

*As a* shop owner, *I want* each accepted order to take its bikes out of stock at once, *so that* two customers cannot buy the last one.

Behaviour:

1. Following the shop's own rule — "an order holds the bike until the confirming call" — an accepted order lowers each ordered bike's stock by the quantity ordered. The lowering and the storing of the order happen together: either both, or neither.
2. A bike's stock never goes below 0. When two orders compete for the last bike, exactly one is accepted and the other is refused as in ORD-004.
3. The new stock shows at once on the shop's pages: the bike's status reads "In stock · N left", or "Out of stock" at 0.
4. An order changes a bike's stock figure only. It never marks a size out; the owner does (`admin` module).
5. An order that fails after part of the work was done leaves no order and no changed stock.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | `talon-2` with stock 9 | An order for 2 is accepted | `talon-2`'s record then has stock 7 |
| AC-2 | `anthem-advanced-sl-0` with stock 1 | An order for 1 is accepted | Its stock is 0 and its page reads "Out of stock" with no "Add to bag" |
| AC-3 | `talon-2` with stock 9 | An order of 1 `talon-2` and 1 `stance-0` is accepted | Each bike's stock is one lower |
| AC-4 | A bike with stock 1 | Two orders for 1 are placed at the same moment | Exactly one is 201 and the other 409; the stock is 0, not -1; one order exists |
| AC-5 | A bike with stock 1 | Twenty orders for 1 are placed at the same moment | One 201 and nineteen 409; stock 0; one order stored |
| AC-6 | A bike with stock 9 | Five orders of 2 are placed at the same moment | Four are 201 and one is 409; the stock is 1 |
| AC-7 | The database fails after the order row is written and before the stock is lowered (simulated) | The order is placed | A 5xx; no order row and the stock unchanged |
| AC-8 | `talon-2` with size XL not in `sizes_out` and stock 1 | An order for 1 in XL is accepted | `sizes_out` is unchanged |
| AC-9 | An order accepted for `talon-2` | The shop's page for `talon-2` is reloaded | It shows the lowered stock |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Concurrent orders | Two or more orders for the same last bikes | Exactly as many succeed as stock allows; the rest are 409 (AC-4 to AC-6) |
| Concurrent owner edit | The owner saves the same bike while an order is placed | Neither overwrites the other silently — specified in ADMIN-008 |
| Upstream failure | The database fails mid-way | Neither the order nor the change is kept (AC-7) |

**Data touched** — a bike's stock; the order and its lines.

**Requirement ORD-006 — The server issues the order number**

*As a* shop owner, *I want* a number I can read out on the phone, *so that* the customer and I name the same order.

Behaviour:

1. The number is `DEO-YYMMDD-NNNN`: `YYMMDD` is the date the order was stored in Vietnam time (Asia/Ho_Chi_Minh), and `NNNN` is a four-digit counter that starts at `0001` for each day and rises by one for each order stored that day.
2. A number is issued only by the server, is unique among all orders, and is never taken from the browser.
3. Orders stored at the same moment get different numbers.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | The first order stored on 16 Nov 2026 in Vietnam | Its number is read | `DEO-261116-0001` |
| AC-2 | A second order the same day | Its number is read | `DEO-261116-0002` |
| AC-3 | An order stored at 18:00 UTC on 16 Nov 2026, which is 01:00 on 17 Nov in Vietnam | Its number is read | It begins `DEO-261117-` |
| AC-4 | An order stored on the next day | Its number is read | It ends `-0001` |
| AC-5 | Twenty orders stored at the same moment | The numbers are read | Twenty different numbers |
| AC-6 | A request that carries a `number` | The order is placed | The request's number is ignored; the server's is used |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Not applicable | — | Not applicable: no roles or failure state of its own; a refused order issues no number |

**Data touched** — the order's number.

**Requirement ORD-007 — The checkout sends the order to the server and shows what it says**

*As a* customer, *I want* to place my order on the checkout page and, when it is refused, to read why and keep my bag, *so that* I can fix it and try again.

Behaviour:

1. Pressing "Place order" on a form that passes the browser's required-field check sends the order of ORD-001 to the server: name, phone, delivery, address, payment, note and each bag line as the bike's id, size and quantity, with the hidden field `website` as it stands. The browser invents no order number.
2. While it is being sent, the button reads "Sending…" and is disabled (SITE-008): one press sends one order.
3. On a 201 the browser saves the server's order in this browser, empties the bag and goes to the confirmation page, which shows the server's number, lines and total (ORD-008, SITE-009).
4. On a 409 or a 400 the server's sentence is shown in the form's alert line (`role="alert"`, the danger colour); the bag is kept and what was typed and chosen stays.
5. When the server cannot be reached or answers 5xx, the line "The order could not be sent. Please try again." is shown in the same alert line; the bag is kept.
6. The module `src/lib/orders.ts` replaces `src/lib/mock/orders.ts`: it keeps `Order`, `saveOrder`, `lastOrder`, `EXAMPLE_ORDER` and `EXAMPLE_LINES` and what they do, the browser no longer makes a number (the number function is gone), and the old file is deleted. The pages that imported from it change that one import line.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A bag of 2 × `talon-2` size M and a valid form | "Place order" is pressed | One request goes to the order endpoint carrying name, phone, delivery, address, payment, note and lines of id, size and quantity only, and no price or total |
| AC-2 | The same | The request is in flight | The button reads "Sending…" and is disabled; a second press sends nothing |
| AC-3 | The server answers 201 with number `DEO-261116-0001` | The page is read | The bag is empty and the browser is at /checkout/done showing `DEO-261116-0001`, the server's lines and the server's total |
| AC-4 | The server answers 409 "Giant Talon 2: only 3 in stock." | The page is read | That sentence is in a `role="alert"` line in the danger colour, the bag still holds its lines, the typed name, phone and address are still in the fields, and the page is still /checkout |
| AC-5 | The server answers 400 with a sentence | The page is read | The same as AC-4 with that sentence |
| AC-6 | The network is down | "Place order" is pressed | "The order could not be sent. Please try again." is in the alert line and the bag is kept |
| AC-7 | The server answers 500 | "Place order" is pressed | The same as AC-6 |
| AC-8 | A refused order was shown | The customer changes the bag and presses "Place order" again | A new request is sent; the old sentence is replaced by the new outcome |
| AC-9 | The finished story | The change's file list is read | `src/components/OrderForm.tsx`, `src/lib/orders.ts` (added), `src/lib/mock/orders.ts` (deleted), and one import line in each of `Bag.tsx`, `Checkout.tsx` and `CheckoutDone.tsx` |
| AC-10 | `src/lib/orders.ts` | Its exports are read | `Order`, `saveOrder`, `lastOrder`, `EXAMPLE_ORDER`, `EXAMPLE_LINES`, and no function that makes an order number |
| AC-11 | A saved order and a reload of /checkout/done | The page is read | The order is shown from this browser's saved copy |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Refused | 409 or 400 | The server's sentence in the alert line; bag kept (AC-4, AC-5) |
| Upstream failure | Network down or 5xx | The generic sentence; bag kept (AC-6, AC-7) |
| Double press | A second press while sending | No second order (AC-2) |
| Lost answer | The order was stored but the answer never arrived | The customer may press again and place a second order; see §7, open question 2 |
| Not connected | The site has no line to the owner | The line "This form is not connected yet — please try again later." stays as designed (SITE-008); the order is still stored and the owner message is skipped |

**Data touched** — the bag; the order the server returned, kept in this browser.

**Requirement ORD-008 — The order first, then the owner's message; a failed message never undoes the order**

*As a* shop owner, *I want* every placed order to reach me in Agent Team, *so that* I can call the customer and confirm it.

*As a* customer, *I want* my confirmed order to stand even if the shop's message could not be sent, *so that* I am never told my order failed when it did not.

Behaviour:

1. The order is saved on the server first. Only after a 201 is the owner's message sent, through the same line to Agent Team that the contact form uses. If the order was not saved, no message is sent.
2. The message shows: the customer's name and phone (as the reply-to); the server's order number; one line per bike, "<quantity> × <name> (size <size>) — <line total>"; the delivery method and the address; the payment method; the total; and the customer's note when there is one.
3. If the message cannot be sent — the line to Agent Team is down, refuses it, or the site has none — nothing is shown to the customer: the saved order is kept, the bag is emptied and the customer goes to the confirmation page as in ORD-007.
4. The confirmation page shows what the server stored: its number, its lines and its total, and the time it was placed.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A valid order | "Place order" is pressed | The request that stores the order is made before the request to Agent Team |
| AC-2 | The server refuses the order with 409 | "Place order" is pressed | No request to Agent Team is made |
| AC-3 | An accepted order for "Minh", phone "0901234567", 2 × Giant Talon 2 size M, delivery `courier`, address "12 Hùng Vương", payment `transfer`, note "Call after 5pm" | The message is read in Agent Team | It shows Minh as the sender with that phone, the server's order number, "2 × Giant Talon 2 (size M) — 37.980.000 ₫", delivery and address, payment, the total 38.330.000 ₫ and the note |
| AC-4 | An accepted order and Agent Team refusing the message | The page is read | The browser is at /checkout/done with the server's order; no error line was shown and the bag is empty |
| AC-5 | An accepted order and no line to the owner | The page is read | The same as AC-4 |
| AC-6 | The confirmation page after an accepted order | It is read | It shows the server's number, the server's lines and total, and the placed date and time |
| AC-7 | An accepted order with no note | The message is read | There is no empty note line |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Notification fails | Agent Team is down or refuses | The order stands; the customer is not told (AC-4); see §7, open question 1 |
| Order refused | 400 or 409 | No message (AC-2) |

**Data touched** — the stored order's number, lines, delivery, address, payment and total; the customer's name, phone and note.

## 5. Screens

This module draws no screen of its own. It changes what stands behind two screens the design already
draws; the states each draws are those of `docs/site/SRS.md`.

| Screen | Section in the design | Functions it serves | States that must exist |
|---|---|---|---|
| Checkout `/checkout` | `pages/Checkout.tsx`, `OrderForm` | ORD-001 to ORD-005, ORD-007, ORD-008 | default, sending, failed, not connected, empty bag |
| Order received `/checkout/done` | `pages/CheckoutDone.tsx` | ORD-001, ORD-006, ORD-008 | order, no order |

## 6. Non-functional requirements

| Area | Requirement |
|---|---|
| Integrity | A bike's stock never goes below 0. An order and its stock change are stored together or not at all. Twenty orders placed at the same moment for a bike with stock 1 give one 201, nineteen 409 and a stock of 0 |
| Performance | Placing an order of 20 lines answers within 1 s at the 95th percentile over 50 sequential requests against the seeded database on the same host |
| Privacy | An order stores the customer's name, phone, address and note, needed to call and to deliver. No endpoint returns them; they appear only in the database, in the owner's message in Agent Team and in the customer's own browser |
| Localisation | Money is whole đồng. The owner's message writes line totals and the total in the site's price format (`37.980.000 ₫`) |

## 7. Dependencies and assumptions

- **Depends on:** `data` (the bikes table, its stock and the shop's record for the courier charge); the service contract in `docs/architecture/services.md` for the error envelope; Agent Team's inbox, through the key in `src/site.json`.
- **Depended on by:** `admin` (an owner's edit of a bike's stock must not silently overwrite an order's hold, ADMIN-008).
- **Assumption:** the shop's stock rule is the one in the shop's record: "Each bike has a number in stock, kept by the shop. An order holds the bike until the confirming call; a size at zero shows as out of stock." The owner corrects stock when an order falls through.

| Open question | Proposed default | Who decides |
|---|---|---|
| 1. When the owner's message cannot be sent the order is stored but nobody is told, and there is no orders screen | Accept; the owner would see the stock fall in the management area. Add an orders screen later if wanted | Stakeholder |
| 2. A customer who presses "Place order" again after the answer was lost makes a second order that holds a second bike; there is no duplicate detection | Accept; the confirming phone call catches it and the owner corrects stock | Stakeholder |

## 8. Traceability

| Plan item | Requirement ids | Test cases |
|---|---|---|
| Place an order and hold the bike | ORD-001, ORD-002, ORD-003, ORD-004, ORD-005, ORD-006, ORD-007, ORD-008 | `test-cases/place-an-order-and-hold-the-bike.md` |
