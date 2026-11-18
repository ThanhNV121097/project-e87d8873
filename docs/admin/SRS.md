# SRS — Admin (the shop's management area)

Module: `admin`
Design: [View the approved design](http://localhost:9400)
Design system: `design/design-system.md`

> One file per module. This one covers the owner's side of the site: signing in for real, and keeping
> bikes and riding types true — stock, prices, status — with every edit saved. The screens are already
> drawn (`/admin/sign-in`, `/admin`, `/admin/products`, `/admin/products/:id`, `/admin/kinds`) and are a
> demo today; this module makes them real and adds no new screen.

## 1. Purpose

The brief asks that stock availability and the prices of what is offered be true. After launch only the
owner can keep them true: a bike sells, a price changes, a new bike arrives. Today any email signs in with no
password and every edit is held in memory and gone on reload. This module gives the owner one real account,
protects every write behind it, saves edits to the database and shows them on the public pages at once. Without
it the shop's stock and prices are frozen at what the team made up.

## 2. Actors

| Actor | Who they are | What they may do in this module |
|---|---|---|
| Visitor | Anyone not signed in | Open the sign-in page; read any public record. Cannot open a management page or write anything |
| Shop owner | The one signed-in account | Everything a visitor may, plus add, change and delete bikes and riding types, and sign out |

There is exactly one owner account and there are no roles: no other account, no second level of access.

## 3. Scope

**In scope** — the functions specified below, by their plan titles:

- Shop admin: real sign-in and saved edits

**Out of scope**

- More than one account, roles, password reset by email, changing the password in the site — deliberately not built; the owner changes the account by changing its two secrets in the console.
- A screen listing orders — deliberately not built; orders reach the owner as messages in Agent Team (`orders` module).
- Editing reviews, team, gallery, workshop jobs, questions and the shop's own numbers from the management area — deliberately not offered; those stay edited in the Studio.
- Releasing a bike held by an order that fell through — the owner corrects the bike's stock here by hand.
- New screens or new fields: the five screens and their fields are as the design draws them.

## 4. Functional requirements

### 4.1 Shop admin: real sign-in and saved edits

**Requirement ADMIN-001 — One owner account, from secrets**

*As the* shop owner, *I want* my email and password to be known only to me and the server, *so that* nobody else can run my shop.

Behaviour:

1. The owner's email and password are two secrets entered by the owner in the console, named `ADMIN_EMAIL` and `ADMIN_PASSWORD`. They are never written in the repository and no endpoint, page or log line ever returns or prints them.
2. The server keeps the password only as a salted, deliberately slow hash and checks a sign-in against that hash. The email is compared ignoring capitals and spaces around it.
3. While either secret is not set, nobody can sign in: every sign-in is refused exactly as a wrong password is, and the server records in its own log that the account is not configured.

**Acceptance criteria** — each is proved by at least one test case in `docs/admin/test-cases/shop-admin-real-sign-in-and-saved-edits.md`.

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | `ADMIN_EMAIL` is `owner@deo.test` and `ADMIN_PASSWORD` is a known value | Those are used to sign in | The sign-in is accepted |
| AC-2 | The same | The repository, the sign-in answer, the `me` answer and a refused sign-in answer are searched for the password's value | It is found nowhere |
| AC-3 | The same | The email is typed as `  OWNER@deo.test ` | The sign-in is accepted |
| AC-4 | `ADMIN_PASSWORD` is not set | Any sign-in is made | It is refused with the sentence of ADMIN-002 AC-2 and the server's log says the account is not configured |
| AC-5 | The server's stored secret | It is inspected | It is a hash, not the password |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Not configured | A secret is missing | Nobody can sign in (AC-4) |
| Other cases | — | Not applicable: no roles; the account cannot be created or changed through the site |

**Data touched** — the two secrets; the password's hash.

**Requirement ADMIN-002 — Signing in**

*As the* shop owner, *I want* to sign in with my email and password, *so that* I can reach the management area.

Behaviour:

1. Signing in with the right email and password answers 200 with the owner (`{ "email": … }`, never the password) and starts a session by setting a cookie that is `HttpOnly`, `SameSite=Lax`, `Secure` whenever the site is served over HTTPS, and expires 7 days after signing in.
2. A wrong email or a wrong password answers 401 with the same single sentence, whichever is wrong: "That email and password do not match."
3. After 5 failed sign-ins from one client address within 15 minutes, every further sign-in from that address, right or wrong, answers 429 with the sentence "Too many attempts. Try again in a few minutes." until 15 minutes after its last failed attempt. A successful sign-in before the fifth failure resets the count. A body that is not an object with an email and a password answers 400.
4. The sign-in screen (`/admin/sign-in`) is as designed: on the left a riding photograph and under it the shop's navy panel with the line "Mountain bikes for passes, trails and rough ground — chosen, built and set up by people who ride them."; on the right the logo (to the home page), the `h1` "Welcome / back" (light "Welcome", bold "back"), the line "Sign in to run Đèo Bikes.", the field "Email" (hint "you@example.com"), the field "Password" (its characters hidden), a full-width button "Sign in", the note "Ask the shop's owner for an account.", and a link "Back to the shop". On a phone the photograph and navy panel come first, then the form.
5. Pressing Enter in either field signs in. An empty email sends nothing and shows "An email address is needed" under the email field. A refusal shows the server's sentence as a visible error in the form and keeps what was typed. When the server cannot be reached the form shows "Could not sign in. Please try again." A success goes to the overview (`/admin`).
6. A visitor who already has a session sees, under the button, "Signed in as <email>." and a button "Continue" that goes to the overview.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | The right email and password | A sign-in is made | 200 with `{ "email": "owner@deo.test" }`; the cookie is `HttpOnly`, has `SameSite=Lax`, and expires 7 days after the sign-in; the answer holds no password |
| AC-2 | A wrong password, and separately a wrong email | Each is tried | Each answers 401 with the identical sentence "That email and password do not match." |
| AC-3 | A body with no password, and a body that is not an object | Each is sent | 400 |
| AC-4 | Five wrong sign-ins from one address within 15 minutes | The sixth is made with the right password | 429 "Too many attempts. Try again in a few minutes." |
| AC-5 | The same lock | 15 minutes have passed since the last failed attempt and the right password is used | The sign-in is accepted |
| AC-6 | Four wrong sign-ins then a right one, then four more wrong ones | The next right one is made | It is accepted (the right one reset the count) |
| AC-7 | A request over HTTPS | A sign-in is made | The cookie also carries `Secure` |
| AC-8 | /admin/sign-in opens | The page is read | The photograph, the navy panel with the shop's line, the logo, the `h1` "Welcome / back", "Sign in to run Đèo Bikes.", "Email", "Password" (hidden characters), the button "Sign in", the note "Ask the shop's owner for an account." and the link "Back to the shop" to the home page |
| AC-9 | The email and password typed | Enter is pressed in the password field | The sign-in is sent |
| AC-10 | The email field is empty | "Sign in" is pressed | "An email address is needed" shows under the email field and nothing is sent |
| AC-11 | A wrong password is sent | The answer arrives | "That email and password do not match." is visible in the form and the typed email is still there |
| AC-12 | The server answers 429 | The answer arrives | "Too many attempts. Try again in a few minutes." is visible in the form |
| AC-13 | The server cannot be reached | "Sign in" is pressed | "Could not sign in. Please try again." is visible in the form |
| AC-14 | The right credentials | "Sign in" is pressed | The browser goes to /admin |
| AC-15 | A signed-in owner | /admin/sign-in opens | "Signed in as owner@deo.test." and a "Continue" button to /admin are shown under the button |
| AC-16 | A 390 px wide window | /admin/sign-in opens | The photograph and the navy panel are above the form |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | Empty email; a malformed body | "An email address is needed" (AC-10); 400 (AC-3) |
| Wrong credentials | Email or password wrong | 401 and one generic sentence that does not say which (AC-2) |
| Throttled | Five failures | 429 (AC-4, AC-5) |
| Upstream failure | The server cannot be reached | "Could not sign in. Please try again." (AC-13) |

**Data touched** — the email and password typed; the session cookie; a count of failed attempts per client address.

**Requirement ADMIN-003 — The session: stay in, deep links, sign out, the gate**

*As the* shop owner, *I want* to stay signed in across reloads, land where I was going, and sign out, *so that* the management area is open to me and closed to everyone else.

Behaviour:

1. When the site loads, the browser asks the server who is signed in (the `me` check): 200 with the owner, or 401 for nobody. The site's existing gate then lets a signed-in owner into any management page and sends anyone else to `/admin/sign-in`.
2. A reload keeps the owner signed in until the session expires, 7 days after signing in; an expired, tampered or missing session means nobody is signed in.
3. Opening a management address directly — for example `/admin/products/stance-0` — while the session is valid shows that page after the check and never the sign-in page; with no valid session it shows the sign-in page.
4. Signing out ends the session on the server — the old cookie stops working — and clears the cookie. The "Sign out" button in the management sidebar signs out and the owner lands on the sign-in page. Signing out when there is no session also answers success.
5. The names the site's pages use — `useSession`, `signIn`, `signOut`, `useRole` — stay; `useRole` returns an empty text because there are no roles. The gate and the management pages are not changed by this.
6. The sidebar footer shows, when signed in, the owner's initial, a name made from the part of the email before the "@" (words separated by `.`, `_` or `-` start with capitals) and the email, and the "Sign out" button.
7. When the check cannot reach the server, nobody is treated as signed in and the sign-in page shows.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A valid session cookie | The `me` check is made | 200 with `{ "email": "owner@deo.test" }` |
| AC-2 | No cookie | The `me` check is made | 401 with one sentence |
| AC-3 | A cookie older than 7 days, and a cookie with changed characters | Each `me` check is made | 401 |
| AC-4 | The owner is on /admin | The page is reloaded | The overview is shown, and the sign-in page never shows at any moment |
| AC-5 | A valid session | /admin/products/stance-0 is typed into the address bar | That bike's record page is shown and the sign-in page never shows at any moment |
| AC-6 | No session | /admin, /admin/products, /admin/products/stance-0 and /admin/kinds are each opened | Each lands on /admin/sign-in |
| AC-7 | A signed-in owner | "Sign out" in the sidebar is pressed | The browser is at /admin/sign-in and the cookie is cleared |
| AC-8 | The cookie value from before the sign-out is replayed | The `me` check is made | 401 |
| AC-9 | No session | Sign-out is requested | Success; nothing changes |
| AC-10 | A signed-in owner `lan.tran@deo.test` | The sidebar footer is read | It shows "L", "Lan Tran", "lan.tran@deo.test" and a "Sign out" button |
| AC-11 | The code | `useRole()` is called for the owner | It returns the empty text |
| AC-12 | The server cannot be reached for the check | /admin opens | /admin/sign-in shows |
| AC-13 | A signed-in owner who signed out | The browser's Back button is pressed | A management page is not shown; the sign-in page is |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Not permitted | No valid session on a management address | The sign-in page, never the management page (AC-6) |
| Expired | The session is over 7 days old | Nobody is signed in (AC-3) |
| Upstream failure | The check cannot reach the server | Nobody is signed in (AC-12) |

**Data touched** — the session cookie; the owner's email.

**Requirement ADMIN-004 — Only the signed-in owner may write, and only bikes and riding types**

*As the* shop owner, *I want* nobody else to be able to change my shop, *so that* its prices and stock stay mine.

Behaviour:

1. Adding a bike or a riding type, changing one, and deleting one are each refused with 401 and one sentence when there is no valid session; nothing is changed.
2. Reading stays public: any record of any collection can be read without signing in.
3. Only bikes (`products`) and riding types (`categories`) can be written. For any other collection — reviews, team, gallery, workshop jobs, questions, the shop's numbers — every write is refused with 405 even for the signed-in owner.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | No session | A bike is added, changed and deleted, each as a separate request | Each is 401 and the bike records are unchanged |
| AC-2 | No session | A riding type is added, changed and deleted | Each is 401 and the riding types are unchanged |
| AC-3 | An expired session | A bike is changed | 401 |
| AC-4 | A signed-in owner | A bike is added | Accepted (ADMIN-005) |
| AC-5 | A signed-in owner | A review is changed, a shop record is changed, a gallery photograph is deleted and a team member is added | Each is 405 and nothing changes |
| AC-6 | No session | The bikes are listed and one is read | 200 for each |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Not permitted | No valid session on a write | 401 (AC-1 to AC-3) |
| Wrong collection | A write to a read-only collection | 405 even when signed in (AC-5) |

**Data touched** — none beyond the session.

**Requirement ADMIN-005 — Saving a bike or a riding type**

*As the* shop owner, *I want* each bike and riding type I save to be checked and kept, *so that* what shows on the site is complete and true.

Behaviour:

1. Adding answers 201 with the saved record; changing answers 200 with the saved record and replaces the whole record with the body; a change to an id that does not exist answers 404.
2. A body is checked against the collection's schema (`src/data/schema.json`). A field the schema does not have is refused with 400 naming it. A required field — a bike's model name, a riding type's name — that is missing or blank is refused with 400 naming it.
3. Types: number fields (`year`, `rating`, `review_count`) must be numbers; `price` and `was` must be whole numbers of đồng and not below 0; `stock` must be a whole number and not below 0; `deal` and `month` must be true or false; text fields must be text of at most 5000 characters; list fields (`sizes`, `sizes_out`, `size_fit`) must be lists of text of at most 50 entries.
4. Closed values: `status` is `draft` or `live`; `wheel` is `29` or `mixed`; `brand` is `Giant` or `Liv`. A bike's `category` must be the id of a riding type that exists. Any other value is refused with 400 naming the field.
5. A bike that is saved as `live` must have a price above 0; otherwise it is refused with 400 "A live bike needs a price." (a free bike must not be orderable by mistake).
6. An empty value for an optional field — an empty text, an empty number — means the field has no value: it is left out of the saved record, never stored as `null`.
7. A new record's id is made by the server when the body has none, in the form `<collection>-<random lowercase letters and digits>`; an id in the body is used when it is made of lowercase letters, digits and hyphens (1 to 64 characters, not starting with a hyphen) and is not taken, otherwise it is refused with 400 or, if taken, 409 "That id is already used." On a change, an id in the body must equal the address's id.
8. Refusals are 400 with one sentence naming the field, in the error envelope of the service contract; nothing is saved.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A signed-in owner and a body with a model name, brand `Giant`, riding type `trail`, price 25000000, status `draft`, stock 3 | A bike is added with no id | 201; the record has a server-made id such as `products-k3f9a2` and the same values; reading it returns it |
| AC-2 | The same body with `"id": "my-bike-1"` | A bike is added | 201 with id `my-bike-1`; adding it again is 409 "That id is already used." |
| AC-3 | The bike `talon-2` | Its body with stock 5 is sent as a change | 200; reading `talon-2` shows stock 5 and every other field as sent |
| AC-4 | A change that omits a field `talon-2` had | It is sent | The saved record no longer has that field (the body replaced it) |
| AC-5 | A change to `no-such-bike` | It is sent | 404 |
| AC-6 | A body with `"colourful": "red"` | It is sent | 400 naming `colourful` |
| AC-7 | A body with no model name, and one with a blank model name | Each is sent | 400 naming the name |
| AC-8 | A body with `price` "cheap", -1, 1.5 | Each is sent | 400 naming `price` |
| AC-9 | A body with `stock` -1 and one with `stock` 2.5 | Each is sent | 400 naming `stock` |
| AC-10 | A body with `status` "sold" | It is sent | 400 naming `status` |
| AC-11 | A body with `wheel` "26" and one with `brand` "Trek" | Each is sent | 400 naming the field |
| AC-12 | A body with `category` "unicycle" | It is sent | 400 naming `category` |
| AC-13 | A body with `sizes` "M" (not a list) and `deal` "yes" | Each is sent | 400 naming the field |
| AC-14 | A body with `status` `live` and no price, and one with price 0 | Each is sent | 400 "A live bike needs a price." |
| AC-15 | A body with status `draft` and no price | It is sent | 200 or 201 (a draft needs no price) |
| AC-16 | A body with `was` empty | It is sent | The saved record has no `was` field and no field is `null` |
| AC-17 | A body whose `id` differs from the address on a change | It is sent | 400 |
| AC-18 | A body with an id of capital letters or with a space | It is sent | 400 |
| AC-19 | A riding type body with a name only | It is added | 201 with a server-made id and the name |
| AC-20 | A riding type body with no name | It is sent | 400 naming the name |
| AC-21 | A body with a text of 5000 characters and one of 5001 | Each is sent | 5000 is accepted; 5001 is 400 |
| AC-22 | Any refused body | The records are read | Nothing was saved |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | Every row of Behaviour 2 to 7 | 400 naming the field; nothing saved (AC-6 to AC-18) |
| Boundary | Text at 5000 and 5001; price 0 on a draft and on a live bike | Accepted at the limit, refused past it (AC-21); live at 0 refused, draft at 0 accepted (AC-14, AC-15) |
| Not found | A change to an id that does not exist | 404 (AC-5) |
| Conflict | An id that is taken | 409 (AC-2); concurrent changes are ADMIN-008 |
| Not permitted | No session | 401 (ADMIN-004) |

**Data touched** — every field of a bike and of a riding type as `schema.json` lists them (the hidden `price_listed` and `_invented` are accepted when sent and never shown to the owner).

**Requirement ADMIN-006 — Deleting a bike or a riding type**

*As the* shop owner, *I want* to delete a bike or a riding type I no longer sell, *so that* the site shows only what is real — and be stopped when a riding type is still in use.

Behaviour:

1. Deleting answers 204 and removes the record; deleting an id that does not exist answers 404.
2. A riding type that bikes still use is refused with 409 and a sentence naming it and how many bikes use it: "<Name> is still the riding type of <N> bike" for one, "…of <N> bikes" for more, followed by ". Give those bikes another riding type first." Nothing is deleted.
3. A riding type no bike uses can be deleted. Gallery photographs that carried it stay, and lose their riding-type link.
4. A bike can be deleted. Its reviews stay and lose their bike link. Orders already placed keep their lines — name and price were copied into them — so the order history is unchanged. A bag that still holds the bike is refused at checkout (ORD-004). Its own page then shows "That bike is no longer in the shop."

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A signed-in owner and a bike with no orders | It is deleted | 204 and reading it is 404; the shop, Home and search no longer show it |
| AC-2 | The riding type `trail`, used by 4 bikes | It is deleted | 409 "Trail is still the riding type of 4 bikes. Give those bikes another riding type first." and the type still exists |
| AC-3 | A riding type used by exactly 1 bike | It is deleted | 409 with "…of 1 bike." |
| AC-4 | A riding type no bike uses | It is deleted | 204 |
| AC-5 | A gallery photograph carrying a riding type that is then deleted | The gallery is read | The photograph is still listed and has no riding type |
| AC-6 | A bike with a review | The bike is deleted | The review still exists and has no `bike` |
| AC-7 | A bike that has an order | The bike is deleted | 204 and the order's lines are unchanged |
| AC-8 | A bag holding the deleted bike | The order is placed | 409 "A bike in your bag is no longer in the shop." |
| AC-9 | An id that does not exist | It is deleted | 404 |
| AC-10 | No session | A delete is sent | 401 and nothing is deleted |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Conflict | A riding type that bikes still use | 409 with the sentence (AC-2, AC-3) |
| Not found | An id that does not exist | 404 (AC-9) |
| Not permitted | No session | 401 (AC-10) |

**Data touched** — a bike or riding type; reviews and gallery photographs that point at it; orders' lines (unchanged).

**Requirement ADMIN-007 — An edit shows at once, on the public pages and on every list on screen**

*As the* shop owner, *I want* a change of stock, price or status to show on the site the moment it is saved, *so that* customers see what is true.

Behaviour:

1. After a save or delete is accepted, every read of that record or collection — by any visitor, from any page — returns the new data; nothing older is served from a copy kept between the server and the browser.
2. In the browser where the owner saved, every list and record on screen for that collection asks again and redraws without a reload: the management lists, the record page and the overview's figures.
3. On the public pages: a bike set to `draft` no longer appears in the shop, on Home, in search or in a bike's "More" row; a bike's stock of 0 shows "Out of stock" and "Ask when it is back"; a changed price shows on its card and page and in the shop's price order.
4. A refused save changes nothing that is shown.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | The bike `talon-2` at stock 9 | The owner saves stock 0 and a visitor then reads `talon-2` | The visitor's read shows stock 0 |
| AC-2 | The same | A visitor reloads the shop, Home and the bike's page | "Out of stock" shows on its card and page, the button reads "Ask when it is back", and Home's "bikes in stock" count is one lower |
| AC-3 | `talon-2` live | The owner saves status `draft` | A visitor's /shop, Home, /search?q=talon and the "More" row of a bike of the same type no longer list it |
| AC-4 | `talon-2` priced 18990000 | The owner saves 17990000 | A visitor's bike page and card show 17.990.000 ₫ |
| AC-5 | The management bikes list is open | The owner saves a bike's stock from the record page, then returns to the list | The list shows the saved figures without a reload |
| AC-6 | The overview is open in the same window as an edit | A bike's stock is changed to 12 | The "Running low" figure falls by one without a reload |
| AC-7 | A save is refused with 400 | The public pages are read | They show what they showed before |
| AC-8 | A bike is deleted | A visitor reads it | 404; the shop no longer lists it |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Refused save | 400, 401 or 409 | Nothing shown changes (AC-7) |
| Other cases | — | Not applicable: no further failure state beyond ADMIN-013 |

**Data touched** — the saved record; the lists and records on screen.

**Requirement ADMIN-008 — An owner's edit and an order do not overwrite each other silently**

*As the* shop owner, *I want* a save that would overwrite an order's hold to be stopped, *so that* the stock I see is never one an order made untrue.

Behaviour:

1. A bike or riding type is saved only if it is unchanged since the screen loaded it. When it has changed since — an order lowered a bike's stock, or the owner saved it from another window — the save is refused with 409 and one sentence: "This record changed while you were editing it — for example an order was placed. Reload and make your change again."
2. Nothing is overwritten: the stock an order lowered, and an earlier save, both stand; the owner's typed values stay in the dialog or the form (ADMIN-013).
3. When a save and an order for the same bike happen at the same moment, the two are applied one after the other: if the save lands first, the order then lowers the stock the owner saved; if the order lands first, the save is refused with the sentence above. No order that was answered 201 is ever lost, and no saved stock is ever replaced by an older figure.
4. Deleting is not refused for a change since the screen loaded it; the owner confirmed deleting that record.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | The owner opened `talon-2` at stock 9; then an order for 2 was accepted (stock 7) | The owner saves a changed price | 409 with the sentence; reading `talon-2` shows stock 7 and the old price |
| AC-2 | The owner reloads and changes the price again | It is saved | 200; the record has stock 7 and the new price |
| AC-3 | Two windows opened the same bike; window A saved | Window B, which loaded earlier, saves | 409 with the sentence; A's change stands |
| AC-4 | The owner saves stock 5 | An order for 1 is accepted afterwards | Stock is 4 |
| AC-5 | A save of stock 5 and an order for 1 at the same moment for a bike at stock 9 | Both finish | If the save was 200 and the order 201, the stock is 4; if the save was 409, the stock is 8; never 5 with the order accepted |
| AC-6 | A bike changed since the screen loaded it | The owner deletes it | 204 |
| AC-7 | The sentence | It is shown in a dialog | The dialog still holds the owner's typed values |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Conflict | The record changed since the screen loaded it | 409 and the sentence; nothing overwritten (AC-1, AC-3) |
| Simultaneous order and save | Both at the same moment | Applied one after the other, with neither lost (AC-5) |

**Data touched** — a bike's stock and every other field; the order's hold (ORD-005).

**Requirement ADMIN-009 — The management frame and the overview**

*As the* shop owner, *I want* a first screen that tells me what needs me today and a menu to everything I can edit, *so that* I can keep stock true in a minute.

Behaviour:

1. Every management page sits in the management frame, in its own fixed neutral look: a sidebar with the shop's name, "Overview", then a group "Shop" with "Bikes" and "Riding types", and, at its foot when signed in, the owner (ADMIN-003). The page's title is in a bar across the top, with the sidebar's toggle at its left; the page's actions (such as "Add") are at its right. The current page's menu entry is marked. Below 768 px the sidebar is a drawer opened by the toggle.
2. The overview (`/admin`, title "Overview") opens with four figures, each a card that opens the list it counts: "Bikes" (the number of bikes, live and draft, with "<N> still in draft"), "Riding types" (the number of riding types, "shown on the site"), "Running low" (the number of bikes with fewer than 10 in stock, "under ten in stock") and "Out of stock" (the number with 0, "marked out of stock"). Bikes, Running low and Out of stock open the bikes list; Riding types opens the riding types list.
3. Under the figures a panel "Running low" lists every bike with fewer than 10 in stock, live or draft, fewest first. A search box ("Search bikes") narrows by model name, ignoring capitals; a select "Any riding type" narrows to one riding type; both live in the address, with a "Clear filters" button while either is set. The table's columns are "Bike" (a link to the bike's record page), "Riding type" (hidden below 1024 px) and "In stock" (the figure in a badge, marked as a danger at 0); "Bike" and "In stock" sort. Under the table: the range "1–14 of 14" (or "No records"), a "rows per page" select of 20, 50 or 100, a button "Export these <N>" that downloads the rows on screen as a CSV file, and the page buttons "Previous", the page numbers and "Next".
4. When no bike is low the panel reads "Nothing is running low." instead of the table; when the search or select leaves no row it reads "No bike running low matches. Clear the search or the riding type to see them all."
5. The figures and the list read the live records: they change when a bike's stock, status or riding type is saved (ADMIN-007).

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A signed-in owner | Any management page opens | A sidebar shows the shop's name, "Overview", the group "Shop" with "Bikes" and "Riding types", and the owner at its foot; the title bar shows the page's title; the current page's entry is marked |
| AC-2 | A 390 px wide window | A management page opens | The sidebar is closed; its toggle opens it as a drawer |
| AC-3 | The 14 seeded bikes (all under 10 in stock, 2 at 0, none in draft) and 4 riding types | /admin opens | "Bikes 14 — 0 still in draft", "Riding types 4 — shown on the site", "Running low 14 — under ten in stock" and "Out of stock 2 — marked out of stock" |
| AC-4 | The overview | Each figure is pressed | Bikes, Running low and Out of stock open /admin/products; Riding types opens /admin/kinds |
| AC-5 | A bike with stock 9 and one with stock 10 | The overview is read | The first is in the "Running low" list and counted; the second is not |
| AC-6 | A bike in draft with stock 2 | The overview is read | It is counted and listed, and "Bikes" shows "1 still in draft" |
| AC-7 | /admin | The "Running low" panel is read | The columns are "Bike", "Riding type", "In stock"; rows are in order of fewest in stock; a bike at 0 shows its figure in a danger badge |
| AC-8 | The panel | "Stance" is typed in "Search bikes" | Only bikes whose model name holds "stance" (ignoring capitals) are listed, "Clear filters" appears, and the address holds the words |
| AC-9 | The panel | A riding type is chosen in "Any riding type" | Only bikes of that type are listed |
| AC-10 | The panel | "Bike" is pressed twice | The rows are in model-name order, then reversed |
| AC-11 | 14 rows | The table's foot is read | "1–14 of 14", a select with 20, 50 and 100 per page, "Export these 14", and "Previous", "1", "Next" with "Previous" disabled |
| AC-12 | The panel | "Export these 14" is pressed | A CSV file of the 14 rows in the table's columns downloads |
| AC-13 | Every bike has 10 or more in stock | /admin opens | The panel reads "Nothing is running low." and shows no table |
| AC-14 | A search that matches nothing | The panel is read | It reads "No bike running low matches. Clear the search or the riding type to see them all." |
| AC-15 | The overview is open | A bike's stock is saved as 12 from another screen of the same window | "Running low" shows one fewer and the bike leaves the list, without a reload |
| AC-16 | The records are still loading | /admin opens | No figure reads 0 and the panel does not read "Nothing is running low." (ADMIN-013) |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Nothing low | No bike is under 10 | AC-13 |
| Nothing matches | The search or select leaves no row | AC-14 |
| Not permitted | Not signed in | The sign-in page (ADMIN-003) |
| Boundary | Stock 9, 10 and 0 | 9 low; 10 not low; 0 low and out (AC-5) |

**Data touched** — bikes (model, riding type, status, stock); riding types.

**Requirement ADMIN-010 — The bikes list: search, filter, add, edit, delete**

*As the* shop owner, *I want* a list of every bike that I can search, filter, add to, edit and delete from, *so that* I can keep the range true.

Behaviour:

1. `/admin/products` is titled "Bikes", with a primary button "Add" at the right of the title bar. Above the table: a search box "Search by model", a select "Status" with "Live" and "Draft", and a "Clear filters" button while any is set. Search words, the status, the sort and the page live in the address.
2. The table's columns are: the bike's photograph (small, no heading), "Model" (a link to the bike's record page), "Brand", "Riding type" (the riding type's name, not its id), "Model year" and "Price" (as `18.990.000 ₫`, right-aligned); the last three are hidden below 1024 px. Every heading except the photograph's sorts; pressing a heading again reverses the order and a third press clears it. The search finds a bike by any of its text.
3. Each row ends in an "Actions" menu with "Edit" and, below a rule in the danger colour, "Delete". The table's foot is as ADMIN-009 Behaviour 3 (range, rows per page, "Export these <N>", page buttons); 20 rows a page.
4. "Add" and "Edit" open a dialog. Adding is titled "Add to bikes" with the buttons "Cancel" and "Add"; editing is titled "Edit <model>", has the line "Changes show on the site as soon as they are saved." and the buttons "Cancel" and "Save changes". Its fields are the bike's visible fields in the schema's order, each with its label and, where it has one, its hint; fields the schema hides are not shown. A required field left empty shows "<Label> is needed" under it and nothing is saved ("Model is needed").
5. "Delete" opens a question "Delete <model>?" with the line "This cannot be undone." and the buttons "Cancel" and "Delete". Deleting removes the bike and closes the question.
6. With nothing in the list and no filter: "Nothing here yet" with an "Add" button. With a filter or search that matches nothing: "Nothing matches those filters", "Clear a filter to see more." and a "Clear filters" button. While the first answer is awaited the table reads "Loading…".

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A signed-in owner and 14 bikes | /admin/products opens | The title "Bikes", "Add", "Search by model", "Status", and a table of 14 rows with the six columns; a bike's riding type reads "Trail", not "trail" |
| AC-2 | A 390 px wide window | The table is read | Only the photograph, "Model" and "Brand" are shown |
| AC-3 | The list | "Model" is pressed three times | The rows are sorted by model ascending, then descending, then in the seeded order |
| AC-4 | The list | "Price" is pressed | The rows are in price order by number |
| AC-5 | The list | "stance" is typed in "Search by model" | Only bikes with "stance" in their text are listed, "Clear filters" appears and the address holds it |
| AC-6 | The list | "Draft" is chosen in "Status" | Only draft bikes are listed; with no draft bike the panel "Nothing matches those filters" shows |
| AC-7 | The list | A row's "Actions" is pressed | A menu with "Edit" and, below a rule, "Delete" shows |
| AC-8 | The list | "Add" is pressed | A dialog titled "Add to bikes" shows the visible fields in the schema's order with their labels and hints, "Cancel" and "Add" |
| AC-9 | The add dialog | "Add" is pressed with "Model" empty | "Model is needed" shows under the field and nothing is sent |
| AC-10 | The add dialog with a model "Test Bike", brand Giant, riding type Trail | "Add" is pressed | The dialog closes and the bike appears in the list, and a visitor reading the bikes finds it |
| AC-11 | A row | "Edit" is chosen | A dialog titled "Edit <model>" with the line "Changes show on the site as soon as they are saved." and "Cancel" and "Save changes" shows, holding the bike's values |
| AC-12 | The edit dialog with stock changed to 0 | "Save changes" is pressed | The dialog closes and a visitor's page for the bike reads "Out of stock" |
| AC-13 | A row | "Delete" is chosen | "Delete <model>?" with "This cannot be undone.", "Cancel" and "Delete" shows |
| AC-14 | The delete question | "Delete" is pressed | The bike leaves the list and a visitor's read of it is 404 |
| AC-15 | No bikes and no filter | The list opens | "Nothing here yet" and an "Add" button |
| AC-16 | A filter that matches nothing | The list is read | "Nothing matches those filters", "Clear a filter to see more." and "Clear filters" |
| AC-17 | The bikes are still loading | The list opens | The table reads "Loading…" |
| AC-18 | A bike | Its "Model" link is pressed | /admin/products/<id> opens |
| AC-19 | The list | The rows-per-page select is changed to 50 | The address holds 50 and the foot reads "1–14 of 14" |
| AC-20 | The add dialog | The "Bikes in stock" field is read | Its hint "0 shows 'Out of stock' and turns the button into 'Ask when it is back'." is under it |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | A required field is empty | "<Label> is needed", nothing sent (AC-9) |
| Refused or failed save or delete | The server refuses or cannot be reached | ADMIN-013 |
| Empty | No bike, or none matches | AC-15, AC-16 |
| Loading | Waiting for the first answer | AC-17 |
| Not permitted | Not signed in | The sign-in page (ADMIN-003) |

**Data touched** — bikes (the visible fields of `schema.json`); riding types, for the name and for the riding-type choice.

**Requirement ADMIN-011 — The bike's own page: every field on four tabs**

*As the* shop owner, *I want* one bike open with all its fields, *so that* I can change its stock, price, sizes and words in one place.

Behaviour:

1. `/admin/products/<id>` shows a breadcrumb "Bikes › <model>" (the first part a link to the list), the bike's model as the page heading, and at its right a quiet button "Delete" and a primary button "Save changes".
2. The bike's visible fields are on four tabs, in the order they first appear in the schema: "Basics" (photograph, model, brand, riding type, model year, price, price before, one line under the name, status, the two close-up photographs), "Stock" (bikes in stock, frame sizes, sizes out of stock, rider height for each size, lead with it, bike of the month), "Specification" (wheel size, suspension travel, frame, fork, rear shock, motor and battery, drivetrain, brakes, wheels, tyres, weight, colour) and "Words" (what it is for, what it is like to own, rating, number of ratings). Each field has its label and, where it has one, its hint. Fields the schema hides are not shown.
3. "Save changes" saves the bike. A required field left empty shows "<Label> is needed" under it and nothing is saved. On success the notice "Saved." shows above the form with a dismiss button; any further edit hides it.
4. "Delete" opens "Delete <model>?" with "This cannot be undone.", "Cancel" and "Delete"; deleting returns to the bikes list.
5. An address whose id has no bike shows, instead of the form, "This record is not here" with "It may have been deleted since the link was made." and a button "Back to the list".

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | /admin/products/stance-0 | The page opens | The breadcrumb "Bikes › <model>", the model as heading, "Delete" and "Save changes" are shown |
| AC-2 | The page | The tabs are read | "Basics", "Stock", "Specification" and "Words" are drawn with the fields listed in Behaviour 2 |
| AC-3 | The page | Every field with a hint is read | Each hint is under its control, including "Shown struck through beside the price. Leave empty when nothing was reduced." under "Price before" |
| AC-4 | The page | The hidden fields are looked for | "Price on the maker's page" and the invented-fields list are not shown |
| AC-5 | The "Stock" tab with stock changed to 4 | "Save changes" is pressed | "Saved." shows with a dismiss button and a visitor's read of the bike shows stock 4 |
| AC-6 | "Saved." is shown | Any field is edited | "Saved." is gone |
| AC-7 | The "Basics" tab with the model cleared | "Save changes" is pressed | "Model is needed" shows under the field and nothing is sent |
| AC-8 | The page | "Delete" is pressed and confirmed | The bike is deleted and the bikes list opens |
| AC-9 | /admin/products/no-such-bike, after the answer | The page opens | "This record is not here", "It may have been deleted since the link was made." and a "Back to the list" button to the list |
| AC-10 | A bike's riding type changed in the "Basics" tab | "Save changes" is pressed | The bikes list and the shop show the new riding type |
| AC-11 | The breadcrumb | "Bikes" is pressed | The bikes list opens |
| AC-12 | The record is still loading | The page opens | Neither the form nor "This record is not here" shows (ADMIN-013) |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Not found | The id has no bike | AC-9 |
| Invalid input | A required field empty | AC-7 |
| Refused or failed save or delete | The server refuses or cannot be reached | ADMIN-013 |
| Conflict | The bike changed since the page loaded | ADMIN-008 |
| Not permitted | Not signed in | The sign-in page (ADMIN-003) |

**Data touched** — one bike's visible fields.

**Requirement ADMIN-012 — The riding types list**

*As the* shop owner, *I want* to keep the riding types the bikes are sorted into, *so that* the shop's chips, filters and "where to start" stay true.

Behaviour:

1. `/admin/kinds` is titled "Riding types", with the same list as ADMIN-010 — "Add" button, search box, sortable table, "Actions" menu with "Edit" and "Delete", foot with range, rows per page, export and page buttons — and no status filter.
2. The columns are "Name", "Who it is for" and "Where it rides best". A row is not a link.
3. "Add" opens "Add to riding types" with "Cancel" and "Add"; "Edit" opens "Edit <name>" with the line "Changes show on the site as soon as they are saved." and "Save changes". A name left empty shows "Name is needed". "Delete" asks "Delete <name>?" with "This cannot be undone.", "Cancel", "Delete".
4. Deleting a riding type that bikes still use is refused (ADMIN-006): the question stays open with the server's sentence (ADMIN-013).
5. The empty states are those of ADMIN-010 Behaviour 6.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | A signed-in owner and 4 riding types | /admin/kinds opens | The title "Riding types", "Add", a search box, and a table of 4 rows with "Name", "Who it is for" and "Where it rides best" |
| AC-2 | The list | A row is read | It is not a link; its "Actions" menu has "Edit" and "Delete" |
| AC-3 | The list | "Add" is pressed | A dialog titled "Add to riding types" with the three fields shows |
| AC-4 | The add dialog with the name empty | "Add" is pressed | "Name is needed" shows under the field and nothing is sent |
| AC-5 | A riding type "Gravel" added | The shop is opened | A chip "Gravel", a footer link and a "where to start" row for it are shown |
| AC-6 | A riding type renamed to "Cross country" | The shop and the bikes list are read | The new name shows on the chips, the breadcrumbs, the cards and the "Riding type" column |
| AC-7 | The riding type `trail`, used by 4 bikes | "Delete" is pressed and confirmed | The question stays open showing "Trail is still the riding type of 4 bikes. Give those bikes another riding type first." and the type is still listed |
| AC-8 | A riding type no bike uses | "Delete" is pressed and confirmed | It leaves the list and the shop's chips |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Invalid input | The name is empty | AC-4 |
| Conflict | The type is still in use | AC-7 |
| Refused or failed save | The server refuses or cannot be reached | ADMIN-013 |
| Not permitted | Not signed in | The sign-in page (ADMIN-003) |

**Data touched** — riding types (name, who it is for, where it rides best).

**Requirement ADMIN-013 — The screens wait for the server and show a failure**

*As the* shop owner, *I want* a refused or failed save to leave my work on screen with the reason, *so that* I never lose what I typed or believe a failed change was made.

Behaviour:

1. `saveRow`, `deleteRow` and `newId` keep their names. `saveRow` and `deleteRow` call the server and report the outcome; the screens wait for it. `newId` still returns a text id that no other call returns.
2. While a save or a delete is in flight its button is disabled, so one press makes one request.
3. On success, the add or edit dialog closes with the list showing the change, the record page shows "Saved.", a confirmed delete closes its question, and on the record page returns to the list (ADMIN-007).
4. On failure, whatever the reason: the add or edit dialog stays open with the typed values and shows the server's sentence in an alert notice inside it; the record page keeps the form with the typed values and shows the alert notice above it, not "Saved."; the delete question stays open and shows the sentence in an alert notice inside it. When the server cannot be reached the sentences are "The change could not be saved. Please try again." for a save and "The record could not be deleted. Please try again." for a delete.
5. On a 401 — the session ended — the sentence is "Your session has ended. Sign in again." and the gate sends the owner to the sign-in page; values typed in an open dialog are lost in that one case.
6. While a record or list is being read the first time, the screens do not state an absence: the record page draws neither its form nor "This record is not here" until the server has answered; the overview draws no figure and no "Nothing is running low." until its records have arrived. A list shows "Loading…" in its table.

**Acceptance criteria**

| # | Given | When | Then |
|---|---|---|---|
| AC-1 | The code | `saveRow`, `deleteRow` and `newId` are imported | All three exist with those names; two `newId` calls return different texts |
| AC-2 | A save is in flight | The dialog is read | Its "Add" or "Save changes" button is disabled |
| AC-3 | An edit dialog and the server answering 409 with the sentence of ADMIN-008 | "Save changes" is pressed | The dialog stays open, holds the typed values, and shows the sentence in an alert notice inside it |
| AC-4 | The add dialog and the server answering 400 "A live bike needs a price." | "Add" is pressed | The dialog stays open with its values and shows the sentence |
| AC-5 | The record page and the server answering 400 | "Save changes" is pressed | The form keeps its values, an alert notice shows the sentence above it, and "Saved." is not shown |
| AC-6 | The delete question for the riding type `trail` | "Delete" is pressed and the server answers 409 | The question stays open with the sentence in an alert notice and "Delete" is enabled again |
| AC-7 | The server cannot be reached | "Save changes" and, separately, "Delete" are pressed | "The change could not be saved. Please try again." and "The record could not be deleted. Please try again." show in the same places |
| AC-8 | The server answers 401 to a save | "Save changes" is pressed | "Your session has ended. Sign in again." shows and the owner is sent to /admin/sign-in |
| AC-9 | A failed save | The lists and public pages are read | They show what they showed before |
| AC-10 | /admin/products/stance-0 opened by link | The page loads | At no moment before the form appears does "This record is not here" show |
| AC-11 | /admin opened by link | The page loads | At no moment before the records arrive does a figure read 0 or the panel read "Nothing is running low." |
| AC-12 | A successful save in the edit dialog | "Save changes" is pressed | The dialog closes and the list shows the change |
| AC-13 | A successful delete from the record page | "Delete" is confirmed | The bikes list opens |

**Failure, boundary and permission behaviour**

| Case | Condition | Expected behaviour |
|---|---|---|
| Refused | 400, 404, 409 | The server's sentence in an alert notice; values kept (AC-3 to AC-6) |
| Upstream failure | The server cannot be reached | The generic sentences (AC-7) |
| Session ended | 401 | AC-8 |
| Double press | A second press while sending | No second request (AC-2) |

**Data touched** — the values typed in the dialog or form; the server's sentences.

## 5. Screens

The design is `code/frontend/src` as approved at http://localhost:9400. The management area is drawn in a fixed neutral look of its own (`design/design-system.md` §1.1 and §2.27); this module does not restyle it.

| Screen | Section in the design | Functions it serves | States that must exist |
|---|---|---|---|
| Sign-in `/admin/sign-in` | `pages/AdminSignIn.tsx`, `SignInForm` | ADMIN-001, ADMIN-002 | default, error shown, signed-in "Continue" line |
| Management frame (sidebar, title bar) | `DashboardLayout` | ADMIN-003, ADMIN-009 | default |
| Overview `/admin` | `pages/AdminOverview.tsx`, `AdminLowStock` | ADMIN-009, ADMIN-013 | default, nothing low, nothing matches |
| Bikes list `/admin/products` | `pages/AdminProducts.tsx`, `ListPage` | ADMIN-005 to ADMIN-008, ADMIN-010, ADMIN-013 | default, loading, nothing yet, nothing matches, add dialog, edit dialog, delete question |
| Bike `/admin/products/:id` | `pages/AdminProduct.tsx`, `RecordPage` | ADMIN-005 to ADMIN-008, ADMIN-011, ADMIN-013 | default, saved notice, delete question, not here |
| Riding types `/admin/kinds` | `pages/AdminKinds.tsx`, `ListPage` | ADMIN-005, ADMIN-006, ADMIN-012, ADMIN-013 | default, loading, nothing yet, nothing matches, add dialog, edit dialog, delete question |

The failure notice of ADMIN-013 uses the management kit's existing alert notice (`Notice`, tone alert); it adds no new component.

## 6. Non-functional requirements

| Area | Requirement |
|---|---|
| Security | The password is held only as a salted, slow hash. The session cookie is `HttpOnly`, `SameSite=Lax` and `Secure` over HTTPS, and ends 7 days after sign-in. Five failed sign-ins within 15 minutes lock that client address out for 15 minutes after its last failure. A write needs the cookie and a JSON request body. No secret appears in the repository, in a response or in a log line |
| Performance | A bike or riding-type write answers within 1 s at the 95th percentile over 50 sequential requests against the seeded database on the same host |
| Responsive | Every management page works from 390 px wide up with no horizontal page scroll; below 768 px the sidebar is a drawer |
| Accessibility | Every management control is keyboard reachable with a visible focus ring; every field is inside a `<label>`; icon-only buttons carry an `aria-label` ("Actions", "Sign out", "Dismiss"); a failed save is announced in a notice |
| Localisation | Copy is in English; prices are written `18.990.000 ₫`; dates `16/10/2026` |
| Privacy | The only personal data kept is the owner's email, held as a secret and never returned except to the owner by the `me` check |

## 7. Dependencies and assumptions

- **Depends on:** `data` (the tables, the read API, the module's "this collection changed" signal, DATA-007); `orders` (the stock hold an owner's edit must not silently overwrite, ORD-005); the service contract in `docs/architecture/services.md` for the error envelope.
- **Secrets the story plan requests from the owner in the console:** `ADMIN_EMAIL`, `ADMIN_PASSWORD`. Until they are given nobody can sign in (ADMIN-001 AC-4).
- **Assumption:** the Studio's preview frame and the design walk sign in by writing the demo session key; whether they keep working once sessions are server-side is TL's concern.
- **Assumption:** the five management screens and their fields are exactly as drawn; this module adds none.

| Open question | Proposed default | Who decides |
|---|---|---|
| 1. When a bike is deleted, its reviews stay and lose their bike link, and gallery photographs of a deleted riding type stay and lose the type; nothing in the brief says | As written in ADMIN-006 | Stakeholder |
| 2. The session lasts 7 days and five failures lock the sign-in for 15 minutes; the brief says "an expiry" and "throttled" without figures | 7 days; 5 failures in 15 minutes | Stakeholder |
| 3. A live bike with no price is refused on save (ADMIN-005 Behaviour 5); the brief names no such rule, and without it a bike with no price could be ordered for nothing | As written | Stakeholder |
| 4. Changing the owner's email or password means the owner changing the two secrets in the console; there is no screen for it and no reset by email | As written (out of scope) | Stakeholder |

## 8. Traceability

| Plan item | Requirement ids | Test cases |
|---|---|---|
| Shop admin: real sign-in and saved edits | ADMIN-001, ADMIN-002, ADMIN-003, ADMIN-004, ADMIN-005, ADMIN-006, ADMIN-007, ADMIN-008, ADMIN-009, ADMIN-010, ADMIN-011, ADMIN-012, ADMIN-013 | `test-cases/shop-admin-real-sign-in-and-saved-edits.md` |
