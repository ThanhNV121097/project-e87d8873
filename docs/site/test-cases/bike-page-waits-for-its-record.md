# Test cases — Bike page waits for its record

Module: `site` · Plan item: "Bike page waits for its record" (bug, P1)
Requirement: SITE-016 (`docs/site/SRS.md` §4.2), narrowing SITE-005 Behaviour 12 · Story: `docs/site/stories/bike-page-waits-for-its-record.md`
Screen: Bike `/shop/:id` (`code/frontend/src/pages/Product.tsx`) · Design: `design/design-system.md` §2.23

## 1. Risk, roles and scope of these cases

- **Risk level: medium.** One frontend file, no write, no money, no stored data. But it is a P1 because the defect is the first thing every visitor to a bike's link would see, and the failure it prevents is a *transient* one — a single rendered frame is enough to break it. So the depth goes to (a) every row of the story's state table (waiting, record in hand, 404, failure), (b) every way of arriving at a bike's page (typed address, shop grid, "More" row, Home), and (c) the log of **every** state the page passes through, not only the final one.
- **Roles.** The SRS names one actor, the visitor (SRS §2); the storefront has no sign-in, so there is no denied role to cover. The management area is out of scope (ADMIN-013).
- **Out of scope here (story §4):** a spinner or any new screen, a retry or error message, the data module's own behaviour (DATA-007), the management record page, draft bikes. No case asks for any of them; where a case sees one of them drawn, that is a failure of SC-3 or SC-6.
- **No request field is sent by any case in this file**; the only API calls are `GET`s with no body and no query. The contract cases of §4.1 are the data this story depends on (story §7), not new endpoints.

## 2. Shared arrangement — how the wait is observed

**Seed records the cases use** (the shop's records as seeded from the design; Test reads them first with `fetch_url` and uses what the API says if a value has since been edited):

| id | Shown as | Notes |
|---|---|---|
| `stance-0` | "Giant Stance 0", riding type "Trail", `66.490.000 ₫`, stock 7 | sizes XS S M L XL (none out); photographs `image`, `image_back`, `image_detail`; two reviews; no `motor` value, so nine specification rows |
| `talon-2` | "Giant Talon 2", `18.990.000 ₫`, stock 9 | sizes XS–XL, none out — the bike of SITE-005 AC-7 |
| `trance-x-1` | "Giant Trance X 1", riding type "Trail" | the bike of SITE-005 AC-1 |
| `stance`, `intrigue-x-2` | other live Trail bikes | with `stance-0` and `trance-x-1` they are the four Trail bikes, so `stance-0` has three bikes for "More" |
| `no-such-bike` | — | must not exist |

**Holding a request.** "Held" means the page driver delays the response to the matching request by **3000 ms** before letting it through unchanged (request interception on `GET /api/v1/products/<id>`; the list requests `GET /api/v1/products?…` and the other collections are **not** held unless a case says so). The hold is what turns a transient frame into a state a reading can catch.

**The state log.** Before the page's own scripts run, the driver registers a script that installs a `MutationObserver` on `document` (`subtree`, `childList`, `characterData`, `attributes`) and, on every callback and on every animation frame, appends one entry: `{ t, path: <the URL path>, main: <text of the main element>, mainChildren: <number of elements inside main>, h1: <text of every h1 inside main>, title: document.title }`. It is read back at the end with `interact_page`. **"At no moment" means in every entry of this log**, not only in the last one. A reading taken only after the page settles cannot fail this story and is not accepted as a pass.

- *Not-found marker* — an entry contains it if its `main` text includes `That bike is no longer in the shop.` or `main` contains a link or button named `See all bikes`. (`main` is the page's own area; the header, menu, search panel and footer sit outside it and may carry other words.)
- *Waiting entry* — an entry taken while the held request has not answered.

**Fallback, stated so it is never silently used.** If the driver cannot register a script before the page's scripts, or cannot hold a request, Test samples `main` every 50 ms from the first `DOMContentLoaded` instead and says in the report that the reading is weaker; a case whose arrangement cannot be produced at all is reported UNVERIFIED with the missing instrument named, never passed.

**Dependency (story §7).** The cases that hold a request need the story "Site data from the database" (DATA-007) on the build under test. On the file-backed mock `useRecord` answers `loading: false` at once, there is no request to hold, and TC-03, 06–08, 17, 20–22 cannot be observed — those are UNVERIFIED on such a build, not failed. **A flash of the not-found view for one frame that the log shows even though the page's own check is right is the data module's defect** (`useRecord` must answer `loading: true` on the very first render for an id, and on the first render after the id changes — DATA-007 AC-13): report it against DATA-007, not against `Product.tsx`.

**Viewports.** 1280×800 unless a case says otherwise; TC-05 repeats the core reading at 390×844.

## 3. Coverage map

| Story criterion | Cases |
|---|---|
| SC-1 — typed address, never the not-found view, bike appears | TC-03, TC-04, TC-05 (and the data they rely on: TC-01) |
| SC-2 — arriving by pressing a card (grid, "More", Home); previous bike not shown | TC-06, TC-07, TC-08 |
| SC-3 — while awaited the page draws nothing; header and footer stay | TC-05, TC-09, TC-10 |
| SC-4 — a bike that exists is drawn as SITE-005 (breadcrumb, price, sizes, add to bag, ask band, "More") | TC-11, TC-12, TC-13, TC-14, TC-15, TC-16 |
| SC-5 — `/shop/no-such-bike` after the 404: the line and "See all bikes" to `/shop`; neither before | TC-02, TC-17, TC-18, TC-19 |
| SC-6 — the request fails (unreachable, 500, 503): empty, never the not-found view, no error screen | TC-20, TC-21, TC-22 |
| SC-7 — a record in hand stays drawn while the shop is asked again | TC-23 (and TC-16 for the part a visitor can reach) |
| SC-8 — one file changed; hooks in the same order; lint, type-check, build pass | TC-24, TC-25 |

## 4. Cases

### 4.1 The data the story depends on (contract)

**Scenario**: TC-01 The API answers 200 with the record for `stance-0`
**Given**: The story is running; the seeded records are in place; no arrangement is active.
**When**: `GET /api/v1/products/stance-0` is requested with no body and no query.
**Then**: Status 200; the body is a JSON object whose `id` is `stance-0`, `name` is `Stance 0`, `brand` is `Giant`, `category` is `trail`; the body carries no `error` key.
Traces: contract (GET /v1/products/{id}) — data that must exist (story §7)
Check: fetch_url

**Scenario**: TC-02 The API answers 404 NOT_FOUND for `no-such-bike`
**Given**: The story is running; no bike has the id `no-such-bike`.
**When**: `GET /api/v1/products/no-such-bike` is requested.
**Then**: Status 404; the body is exactly the error envelope `{"error":"<one sentence>","code":"NOT_FOUND"}` — `code` is `NOT_FOUND`, `error` is a non-empty sentence, no record fields are present; the response carries `Cache-Control: no-store`.
Traces: SC-5 (SITE-016 AC-3), contract (GET /v1/products/{id})
Check: fetch_url

### 4.2 SC-1 — a typed address never shows the not-found view

**Scenario**: TC-03 Direct load of `/shop/stance-0` with the bike's answer held never shows the not-found view, then draws the bike
**Given**: The story is running at 1280×800. The state log (§2) is registered before the page's scripts. The response to `GET /api/v1/products/stance-0` is held for 3000 ms.
**When**: The address `/shop/stance-0` is opened by a full page load (typed, not clicked), the page is left alone for 3000 ms, then the held answer is let through and the page is read until `[data-testid='product-name']` appears (at most 5 s later).
**Then**: (1) No entry of the log — waiting entries and every entry after — contains the not-found marker before the bike is drawn: `main` never contains `That bike is no longer in the shop.` and never contains a link or button named `See all bikes`. (2) After the answer, `[data-testid='product-name']` reads exactly `Giant Stance 0` and is the only `h1` in `main`; `[data-testid='product-price']` reads `66.490.000 ₫`. (3) No later entry contains the not-found marker either.
Traces: SC-1 (SITE-016 AC-1)
Check: interact_page

**Scenario**: TC-04 Direct load of `/shop/stance-0` at the API's own speed never flashes the not-found view
**Given**: The story is running at 1280×800; no request is held; the state log is registered before the page's scripts.
**When**: `/shop/stance-0` is opened by a full page load and the log is read after `[data-testid='product-name']` has appeared. The same load is repeated five times, each from a fresh navigation.
**Then**: In none of the five logs does any entry contain the not-found marker; in each, the entries before the bike is drawn have `mainChildren` equal to 0 and an empty `main` text. (A one-frame flash would be a log entry with the marker; if one is seen while the page's code is right, it is DATA-007 AC-13's defect — see §2.)
Traces: SC-1 (SITE-016 AC-1)
Check: interact_page

**Scenario**: TC-05 The same wait holds on a phone: 390×844, nothing of the page and no not-found view while the answer is awaited
**Given**: The story is running; the viewport is 390×844 (a phone); the state log is registered; the response to `GET /api/v1/products/stance-0` is held for 3000 ms.
**When**: `/shop/stance-0` is opened by a full page load and read at 500 ms, 1500 ms and 2800 ms, then the answer is let through.
**Then**: At each reading and in every waiting entry of the log, `main` contains no text, no element (`mainChildren` is 0) and no `h1`, and the not-found marker is absent; the `header` and `footer` elements are present with text. After the answer `[data-testid='product-name']` reads `Giant Stance 0`, and the document is not wider than 390 px (`document.documentElement.scrollWidth` is at most 390).
Traces: SC-1 (SITE-016 AC-1), SC-3 (SITE-016 Behaviour 1)
Check: interact_page

### 4.3 SC-2 — arriving by pressing a bike card

**Scenario**: TC-06 Pressing a bike card on the shop grid never shows the not-found view before the bike appears
**Given**: The story is running at 1280×800; `/shop` is open and `[data-testid='card-stance-0']` is drawn; the state log is registered on this page; from now on the response to `GET /api/v1/products/stance-0` is held for 3000 ms.
**When**: `[data-testid='card-stance-0']` is clicked, the page is left alone for 3000 ms, then the answer is let through.
**Then**: After the click the URL path is `/shop/stance-0`. In every entry from the click until `[data-testid='product-name']` appears, `main` does not contain `That bike is no longer in the shop.` and has no link or button named `See all bikes`; while the answer is held `main` has no element and no `h1`. After the answer `[data-testid='product-name']` reads `Giant Stance 0`.
Traces: SC-2 (SITE-016 AC-2)
Check: interact_page

**Scenario**: TC-07 Pressing a card in the "More" row of another bike's page shows neither the not-found view nor the previous bike while the new bike is awaited
**Given**: The story is running at 1280×800; `/shop/stance-0` has loaded and shows `Giant Stance 0`; its "More Trail" row (`section[data-block='related']`) holds three cards; the state log is registered on this page; from now on every `GET /api/v1/products/<id>` is held for 3000 ms. Let `<new-id>` be the id in the `data-testid` (`card-<new-id>`) of the first card of that row, and `<new-name>` the card's bike name.
**When**: That card is clicked, the page is left alone for 3000 ms, then the answer is let through.
**Then**: After the click the URL path is `/shop/<new-id>`. In every entry from the click until the new bike appears: the not-found marker is absent, and **no entry whose `path` is the new address shows the previous bike** — `main` does not contain `Giant Stance 0` and `[data-testid='product-name']` is absent; while the answer is held `main` has no element. After the answer `[data-testid='product-name']` reads `Giant <new-name>` and `main` no longer contains the text `Stance 0` as a heading.
Traces: SC-2 (SITE-016 AC-2)
Check: interact_page

**Scenario**: TC-08 Pressing a bike card in a Home section never shows the not-found view before the bike appears
**Given**: The story is running at 1280×800; `/` is open and the lead card `[data-testid='tint-stance-0']` is drawn; the state log is registered on this page; from now on the response to `GET /api/v1/products/stance-0` is held for 3000 ms.
**When**: `[data-testid='tint-stance-0']` is clicked, the page is left alone for 3000 ms, then the answer is let through.
**Then**: After the click `location.pathname` is `/shop/stance-0`. In every entry from the click until the bike appears, the not-found marker is absent and, while the answer is held, `main` has no element (none of Home's sections remains under the new address). After the answer `[data-testid='product-name']` reads `Giant Stance 0`.
Traces: SC-2 (SITE-016 AC-2)
Check: interact_page

### 4.4 SC-3 — while the bike is awaited the page draws nothing

**Scenario**: TC-09 While the bike's answer is awaited `main` is empty and the header and footer are drawn
**Given**: The story is running at 1280×800; the response to `GET /api/v1/products/stance-0` is held for 3000 ms.
**When**: `/shop/stance-0` is opened by a full page load and, 1500 ms in (the answer still held), the page is inspected.
**Then**: Inside `main`: the text is empty; there are no child elements at all (so no breadcrumb `nav[aria-label='Breadcrumb']`, no `h1`, no `img`, no `svg`, no `button`, no `a`, no `progress`, no `[role='status']`, no `[aria-busy='true']`, no `[role='alert']`); the computed height of `main` is `0px` (no placeholder box, however pale). Outside `main`: the `header` element is present with its logo link and navigation, the `footer` element is present with text.
Traces: SC-3 (SITE-016 Behaviour 1)
Check: interact_page

**Scenario**: TC-10 The tab title is the bike page's title while the page waits
**Given**: The story is running; the response to `GET /api/v1/products/stance-0` is held for 3000 ms; the state log is registered.
**When**: `/shop/stance-0` is opened by a full page load and `document.title` is read at 1500 ms (answer held) and again after the bike is drawn.
**Then**: Both readings contain `A bike from Đèo Bikes` (the product page's title from `content.json`, `pages.product.title`), and they are equal to each other: the title does not change when the bike arrives.
Traces: SC-3 (SITE-016 Behaviour 1), story §5 "Tab title: unchanged"
Check: interact_page

### 4.5 SC-4 — a bike that exists is drawn exactly as SITE-005 describes

These are the story's spot checks plus the whole page once. The page must look as it did before the change, so every value below is the value the approved design draws.

**Scenario**: TC-11 The price of `talon-2` reads `18.990.000 ₫`
**Given**: The story is running at 1280×800; no request is held; `talon-2` has price 18990000 in the API (read with `GET /api/v1/products/talon-2` first).
**When**: `/shop/talon-2` is opened and rendered after the scripts have run.
**Then**: `[data-testid='product-price']` reads exactly `18.990.000 ₫`; `[data-testid='product-name']` reads exactly `Giant Talon 2`; the text `That bike is no longer in the shop.` is not on the page.
Traces: SC-4 (SITE-016 AC-4, SITE-005 AC-7)
Check: render_url

**Scenario**: TC-12 The price is set in large bold figures
**Given**: The story is running at 1280×800; `/shop/talon-2` has loaded and shows its price.
**When**: The computed style of `[data-testid='product-price']` is measured.
**Then**: `font-size` is `36px` and `font-weight` is `700` (the values the design draws — Tailwind `text-4xl` and `font-bold`); `font-variant-numeric` includes `tabular-nums`.
Traces: SC-4 (SITE-016 AC-4, SITE-005 AC-7)
Check: measure_styles

**Scenario**: TC-13 The breadcrumb of `trance-x-1` reads Bikes › Trail › the bike's name
**Given**: The story is running at 1280×800; `trance-x-1` is a Trail bike named `Trance X 1`.
**When**: `/shop/trance-x-1` is opened and rendered after the scripts have run.
**Then**: Inside `nav[aria-label='Breadcrumb']`, in order: a link `Bikes` with `href` `/shop`; a link `Trail` with `href` `/shop?category=trail`; then the text `Trance X 1`, which is not a link (the nav holds exactly two links).
Traces: SC-4 (SITE-016 AC-4, SITE-005 AC-1)
Check: render_url

**Scenario**: TC-14 The size pills of a bike show each size with M chosen; a size marked out is disabled
**Given**: The story is running at 1280×800; `talon-2` has sizes XS S M L XL with none out; `trance-x-1` has sizes S M L XL with XL out.
**When**: `/shop/talon-2` is opened and its pills are read; then `[data-testid='size-L']` is clicked; then `/shop/trance-x-1` is opened and its pills are read.
**Then**: On `talon-2`: five pills `size-XS`, `size-S`, `size-M`, `size-L`, `size-XL`, each showing its rider height under the size letter; `size-M` has `aria-checked="true"` and the other four `"false"`; none is disabled. After the click, `size-L` is `aria-checked="true"` and `size-M` is `"false"`. On `trance-x-1`: `size-XL` is disabled, `size-M` is `aria-checked="true"`.
Traces: SC-4 (SITE-016 AC-4, SITE-005 AC-10)
Check: interact_page

**Scenario**: TC-15 "Add to bag" puts the chosen quantity and size in the bag
**Given**: The story is running at 1280×800; the bag is empty (`localStorage` key `agentteam:cart` removed); `/shop/talon-2` has loaded with `size-M` chosen and quantity 1.
**When**: The button labelled `One more` is clicked once (quantity 2), then `[data-testid='product-add']` is clicked, then the link `[data-testid='product-to-bag']` is clicked.
**Then**: After "Add to bag": the button reads `In your bag` and the line `Added to your bag.` is drawn with a link `See your bag` whose `href` is `/bag`. After the link: `location.pathname` is `/bag`; `[data-testid='bag-line-talon-2']` contains `Giant Talon 2`, `Size M` and a quantity of `2`; `[data-testid='bag-total']` reads `37.980.000 ₫`.
Traces: SC-4 (SITE-016 AC-4, SITE-005 AC-13)
Check: interact_page

**Scenario**: TC-16 `stance-0` is drawn whole: every section, the ask band and the "More" row
**Given**: The story is running at 1280×800; no request is held; `stance-0` has nine specification values (no motor), five sizes, three photographs, two reviews, stock 7; the shop's record holds warranty 2 and reply 2 (read from the API first).
**When**: `/shop/stance-0` is opened and rendered after the scripts have run.
**Then**: The page holds, in this order down the page: the breadcrumb `Bikes › Trail › Stance 0`; the gallery with one main picture and two thumbnail buttons carrying `aria-pressed`; the head with the `h1` `Giant Stance 0`, the line `29 in · 140 / 130 mm · 1×12 · 2027`, five stars with `4.9` and `(23 ratings)`; the price `66.490.000 ₫` and the stock line `In stock`; `Frame size` with a link `Size guide` and the five pills; a button `Add to bag` and a link `Ask about it` whose `href` is `#ask`; the promise list with `2 years of frame warranty from the shop` and `A reply within 2 hours, seven days a week`; the heading `The specification` over nine rows in the order Frame, Fork, Rear shock, Drivetrain, Brakes, Wheels, Tyres, Weight, Colour — and no `Motor and battery` row; the heading `Which size?` over a table with columns `Size`, `Rider height`, `Stock`, five rows, and the note `Heights are the maker's guide. Long legs or a short back can move you a size — send us both and we will check.`; tabs `What it is for`, `Owning it`, `Delivery and payment`, `Reviews (2)`; `section#ask` with the heading `Ask about the Stance 0` and the line `Leave your number and your height. We call back with the right size and when you can ride it.`; and `section[data-block='related']` with the heading `More Trail` and exactly three bike cards, none of them `[data-testid='card-stance-0']`.
Traces: SC-4 (SITE-016 AC-4, SITE-005 AC-1, AC-2, AC-5, AC-15, AC-17, AC-18, AC-19, AC-20, AC-21, AC-24)
Check: render_url

**Scenario**: TC-17 The page waits for the bike only: the bike is drawn while the riding types, reviews, shop record and "More" row are still on their way
**Given**: The story is running at 1280×800; the response to `GET /api/v1/products/stance-0` is **not** held; the responses to every other API request the page makes — `GET /api/v1/categories…`, `GET /api/v1/reviews…`, `GET /api/v1/shop…` and `GET /api/v1/products?…` — are held for 3000 ms.
**When**: `/shop/stance-0` is opened by a full page load and read at 1500 ms, then again at 4500 ms (after the hold has ended).
**Then**: At 1500 ms: `[data-testid='product-name']` reads `Giant Stance 0`, `[data-testid='product-price']` reads `66.490.000 ₫`, the not-found marker is absent from every log entry, and these are **not yet** drawn: the riding-type pill `Trail` in the head, the middle breadcrumb link `Trail`, the promise list, a `Reviews` tab, and the `More Trail` row. At 4500 ms: the pill `Trail`, the middle breadcrumb link `Trail` (`href` `/shop?category=trail`), the promise list, the tab `Reviews (2)` and the `More Trail` row with three cards are all drawn, and the bike has not blanked or flickered out in between (no log entry after the first draw of the bike has `[data-testid='product-name']` absent).
Traces: SC-4 (SITE-016 AC-4), SC-7 (a bike in hand stays drawn), story §4 (the other records arrive when their own answers do)
Check: interact_page

### 4.6 SC-5 — a bike that does not exist

**Scenario**: TC-18 `/shop/no-such-bike` shows nothing while the request is awaited, then the line and button once the 404 has answered
**Given**: The story is running at 1280×800; no bike has the id `no-such-bike`; the state log (§2) is registered; the response to `GET /api/v1/products/no-such-bike` (a real `404` from the API) is held for 3000 ms.
**When**: `/shop/no-such-bike` is opened by a full page load, read at 1500 ms (answer held), then the answer is let through and the page is read 1000 ms later.
**Then**: In every waiting entry of the log, `main` has no element, no text, no `h1`, no `That bike is no longer in the shop.` and no link or button named `See all bikes`. After the answer: `main` contains exactly one `h1` whose text is exactly `That bike is no longer in the shop.` and one link whose text is exactly `See all bikes` with `href` `/shop`.
Traces: SC-5 (SITE-016 AC-3, SITE-005 AC-26)
Check: interact_page

**Scenario**: TC-19 The button "See all bikes" opens the shop
**Given**: The story is running at 1280×800; `/shop/no-such-bike` has loaded and shows the not-found view (nothing held).
**When**: The link `See all bikes` in `main` is clicked.
**Then**: `location.pathname` is `/shop`; the shop grid is drawn with `[data-testid='card-stance-0']` present; `main` no longer contains `That bike is no longer in the shop.`.
Traces: SC-5 (SITE-016 AC-3, SITE-005 AC-26)
Check: interact_page

**Scenario**: TC-20 "See all bikes" is the primary button of the design
**Given**: The story is running at 1280×800; `/shop/no-such-bike` has loaded and shows the not-found view.
**When**: The computed style of `main a[href='/shop']` is measured.
**Then**: `background-color` is the theme's accent, `rgb(25, 26, 46)`; `color` is the theme's accent ink, `rgb(255, 255, 255)`; `font-size` is `14px`; the computed `border-radius` is a full pill (at least half the element's rendered height). The `h1` in `main` has `color` `rgb(25, 26, 46)` (ink).
Traces: SC-5 (SITE-005 AC-26; story §5 "Not found: the existing Missing view, unchanged")
Check: measure_styles

### 4.7 SC-6 — the request for the bike fails

In each case the bike has **not** been said to be gone, so the not-found view is never drawn, and the design draws no error screen, so none appears (`docs/data/SRS.md` §7, open question 1: accepted at launch).

**Scenario**: TC-21 With the shop unreachable the bike's place stays empty and never says the bike is gone
**Given**: The story is running at 1280×800; the backend is stopped, so the gateway answers `502` to `GET /api/v1/products/stance-0` (confirmed with `fetch_url`); the frontend itself is still served; the state log (§2) is registered.
**When**: `/shop/stance-0` is opened by a full page load and read at 500 ms and at 5000 ms.
**Then**: In every log entry: `main` does not contain `That bike is no longer in the shop.` and has no link or button named `See all bikes`. At 5000 ms: `main` has no element, no text and no `h1` (no error message, no `[role='alert']`, no retry button, no spinner); the `header` and `footer` are still drawn. Restoring the arrangement (backend started) and reloading `/shop/stance-0` then draws `[data-testid='product-name']` reading `Giant Stance 0`.
Traces: SC-6 (SITE-016 failure table, "Upstream failure")
Check: interact_page

**Scenario**: TC-22 When the shop answers 500 the bike's place stays empty and never says the bike is gone
**Given**: The story is running at 1280×800; the store fails the bike query — the arrangement tool makes the products query fail (for example by renaming the `products` table) — so `GET /api/v1/products/stance-0` answers `500` with `code` `INTERNAL` (confirmed with `fetch_url`); the state log is registered.
**When**: `/shop/stance-0` is opened by a full page load and read at 500 ms and at 5000 ms.
**Then**: In every log entry the not-found marker is absent. At 5000 ms: `main` has no element, no text and no `h1`; no error message is drawn; the `header` and `footer` are still drawn. After the table is put back and the page reloaded, the bike is drawn (`Giant Stance 0`).
Traces: SC-6 (SITE-016 failure table, "Upstream failure"; story §3 table, row "answered 5xx")
Check: interact_page

**Scenario**: TC-23 When the database refuses connections (503) neither a real bike nor a missing one is called gone
**Given**: The story is running at 1280×800; the backend is up and the database is made unavailable (`set_db_available` off), so `GET /api/v1/products/stance-0` and `GET /api/v1/products/no-such-bike` both answer `503` with `code` `UNAVAILABLE` (confirmed with `fetch_url`); the state log is registered.
**When**: `/shop/stance-0` is opened and read at 5000 ms; then `/shop/no-such-bike` is opened and read at 5000 ms.
**Then**: For both addresses: in every log entry the not-found marker is absent; at 5000 ms `main` has no element, no text and no `h1`; the `header` and `footer` are drawn. (For `no-such-bike` the 404 was never heard, so "That bike is no longer in the shop." must not be said.) After the database is made available again, a reload of `/shop/stance-0` draws `Giant Stance 0` and a reload of `/shop/no-such-bike` draws the `h1` `That bike is no longer in the shop.`.
Traces: SC-6 (SITE-016 failure table, "Upstream failure"), SC-5 (a 404 is the only answer that means "no such bike")
Check: interact_page

### 4.8 SC-7 — a bike in hand stays drawn when the shop is asked again

**Scenario**: TC-24 A record in hand is drawn whatever `loading` and `error` say
**Given**: The PR head of this story; the page's early-return logic, read in `code/frontend/src/pages/Product.tsx`.
**When**: The logic is read for the state `item` present with `loading` true (the data module's refresh after another module saves or deletes a record, DATA-007 Behaviour 7 and AC-16), and for `item` present with `error` set.
**Then**: The bike's page is returned whenever `item` is present: the wait and the not-found decision are made only when `item` is empty (for instance `if (!item) { if (loading || error) return null; return <Missing />; }`). A test of `loading` or `error` that comes before the test of `item` — for example `if (loading) return null` placed first — fails this case, because it would blank a page the visitor is reading.
Traces: SC-7 (SITE-016 Behaviour 2)
Check: manual — *Reason:* the state (record in hand while the shop is asked again) is reached only when the data module re-asks after a save or delete made by another module, which a visitor on a mounted bike page cannot trigger, and the story adds no component test (SC-8 allows `Product.tsx` only), so no instrument (`fetch_url`, `render_url`, `measure_styles`, `interact_page`) can put the page in that state. The part a visitor can reach — the bike staying drawn while other requests are pending — is TC-17. If the PR nevertheless carries a component test that stubs `useRecord` with `item` present and `loading: true`, Test runs it with `run_tests` (frontend) and records its result here.

### 4.9 SC-8 — the change is one file and breaks nothing

**Scenario**: TC-25 The change is `Product.tsx` and nothing else, and the checks pass
**Given**: The PR for this story, at its head commit.
**When**: The PR's changed-file list and diff are read, and the CI result of that head commit is read.
**Then**: The changed files are exactly `code/frontend/src/pages/Product.tsx` — no other page, component, library, data-module file, `content.json`, `theme.css` or `index.css`. The diff reads `loading` and `error` from `useRecord` and decides in this order: a record in hand → the bike's page; none and (`loading` or `error`) → `null`; none, not loading, no error → `<Missing />`; the `Missing` function, the markup of the bike's page and every class on it are unchanged; `useCollection("categories", …)` and `usePageMeta(…)` are still called before the first early return, in the same order; no new key is added to `content.json` and no spinner, skeleton or "Loading…" text is added. The CI jobs `npm run lint`, the type-check and the build are green on the head commit.
Traces: SC-8 (SITE-016 AC-4)
Check: manual — *Reason:* a statement about the PR's contents and CI status, not about anything a response or a rendered page shows; read with `get_pr` and `get_ci_logs`. Every other case in this file only runs if the build compiled, which is independent evidence that type-check and build pass.

**Scenario**: TC-26 The order of hooks never changes: waiting → bike, and waiting → not found, draw without a React error
**Given**: The story is running at 1280×800; the response to `GET /api/v1/products/stance-0` is held for 3000 ms in the first run and the response to `GET /api/v1/products/no-such-bike` in the second; the driver records page errors and console errors.
**When**: First `/shop/stance-0`, then `/shop/no-such-bike`, is opened by a full page load, left alone for 3000 ms, and the held answer is let through. Then, from the loaded `/shop/stance-0`, the first card of the "More" row is clicked (the id changes while the page stays mounted) and the page is read after its answer.
**Then**: After each transition the page draws what it should (`[data-testid='product-name']` reads `Giant Stance 0`; the `h1` reads `That bike is no longer in the shop.`; the clicked bike's `[data-testid='product-name']` appears); the `header` is still present after each (the app did not unmount); none of the recorded page or console errors matches `Rendered more hooks`, `Rendered fewer hooks` or `change in the order of Hooks`.
Traces: SC-8 (story §3 last bullet: hooks stay above the early return, in the same order)
Check: interact_page

## 5. Read-back against the criteria list

| Criterion | Covered by | Instrument(s) |
|---|---|---|
| SC-1 | TC-03 (held), TC-04 (no delay), TC-05 (phone) | interact_page |
| SC-2 | TC-06 (shop grid), TC-07 ("More" row, previous bike not shown), TC-08 (Home) | interact_page |
| SC-3 | TC-05, TC-09 (empty `main`, header and footer stay), TC-10 (title) | interact_page |
| SC-4 | TC-11, TC-13, TC-16 (render_url); TC-12 (measure_styles); TC-14, TC-15, TC-17 (interact_page) | render_url, measure_styles, interact_page |
| SC-5 | TC-02 (404 envelope, fetch_url); TC-18, TC-19 (interact_page); TC-20 (measure_styles) | fetch_url, interact_page, measure_styles |
| SC-6 | TC-21 (unreachable), TC-22 (500), TC-23 (503, real and missing bike) | interact_page |
| SC-7 | TC-24 (manual — no instrument can force "record in hand while asking again"); TC-17 for the reachable part | manual, interact_page |
| SC-8 | TC-25 (manual — repository diff and CI), TC-26 (hook order at runtime) | manual, interact_page |

26 cases: 22 `interact_page`/`render_url`/`measure_styles`/`fetch_url` cases that observe the running story, 2 `manual` (TC-24, TC-25) for the reasons stated in their `Check:` lines — 8% of the file, well under the one-third that parks a story.
