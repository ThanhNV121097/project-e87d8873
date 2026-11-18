# Story — Shop admin: real sign-in and saved edits

Module: `admin` · Plan item: "Shop admin: real sign-in and saved edits" (feature, P1)
Requirements: `docs/admin/SRS.md` ADMIN-001 to ADMIN-013
Contract the build follows: `docs/architecture/services.md` (§2.2 authentication, §2.3 errors, §3.4 to §3.9 endpoints, §3.11 body rules) and `docs/architecture/erd.md` (`sessions`, the `version` trigger)
Design: `design/design-system.md` §1.1 "Management area" and §2.27 "Management area kit". **No new screen, no new field and no restyle.**

## 1. User story

As the **shop owner**, I want to sign in with my own email and password and have every change I make to bikes and riding types saved — with stock, price and status showing on the public pages at once — so that customers see what is true after launch, and nobody else can change my shop.

Actors (SRS §2): the **shop owner** (the one signed-in account) and the **visitor** (anyone else, who must be kept out of every write and every management page).

## 2. What exists and what changes

Today the five management screens (`/admin/sign-in`, `/admin`, `/admin/products`, `/admin/products/:id`, `/admin/kinds`) are the approved design running on mock data: any email signs in with no password check (`src/lib/session.ts`, `components/admin/SignInPage.tsx`) and every edit is held in memory and gone on reload (`components/admin/store.ts`). After this story:

- one owner account, whose email and password are two secrets, signs in for real and gets a server-side session cookie;
- every write to a bike or a riding type needs that session, is validated against the schema, is saved in the database and is visible to every visitor on the next read;
- the screens wait for the server, keep what the owner typed when a save or delete fails, and say why in the sentence the server gave.

## 3. In scope

**Secrets (requested from the owner in the console by this plan — no key is ever typed in chat):** `ADMIN_EMAIL`, `ADMIN_PASSWORD`. Until both are supplied nobody can sign in (ADMIN-001 AC-4); the public shop works regardless. Both names are listed, without values, in `code/backend/.env.example`.

**Backend** (Go, as in `docs/architecture/services.md`):

- Sign-in, the `me` check and sign-out: `POST /v1/sessions`, `GET /v1/me`, `DELETE /v1/sessions/current` — argon2id hash of `ADMIN_PASSWORD` made at start and held in memory, `deo_session` cookie (`HttpOnly`, `SameSite=Lax`, `Secure` over HTTPS, 7 days), the `sessions` table (migration slot `20261118000400`), the in-memory throttle (5 failures in 15 minutes per client address → 429), the `requireOwner` check.
- Writes for bikes and riding types only: `POST /v1/{products|categories}`, `PUT /v1/{products|categories}/{id}` (with `If-Match`), `DELETE /v1/{products|categories}/{id}`; every other collection answers 405 to a write. Body checks per `services.md` §3.11 (required `name`, types, closed values, `category` exists, live bike needs a price, ids, 5000 characters, empty means no value), server-made ids, the riding-type-in-use refusal, compare-and-swap on the record's version so an owner's edit and an order's stock hold never overwrite each other silently.
- `GET /v1/{collection}/{id}` sends `ETag` for bikes and riding types (`services.md` §3.3). If the data story left it out, this story adds it in the same handler.
- Go unit tests that run without a database (CI has none): the validators, the throttle (injected clock), the hash check, the cookie attributes, the 405/401 ordering. Everything that needs the database is proved in the local run and the verification pass.

**Frontend** (files in §9):

- `src/lib/session.ts` keeps `useSession`, `signIn`, `signOut`, `useRole` and asks the server; `components/admin/SignInPage.tsx` checks the password for real.
- `components/admin/store.ts` keeps `saveRow`, `deleteRow`, `newId`, calls the API, and tells the data module the collection changed so every list and record on screen asks again.
- The screens wait for the answer: `ListPage.tsx`, `RecordPage.tsx`, the `Confirm` component in `components/admin/index.tsx`, and `pages/AdminOverview.tsx` (loading state).
- `src/test/admin.test.ts` is rewritten for the store that now calls the API (the field-derivation tests in it stay as they are).

## 4. Out of scope (deliberately, with the reason)

- More than one account, roles, password reset by email, changing the password in the site — SRS §3: the owner changes the account by changing its two secrets in the console.
- An orders screen — orders reach the owner as messages in Agent Team (`orders` module).
- Editing reviews, team, gallery, workshop jobs (`services`), questions (`faqs`) and the shop's numbers from the management area — they stay edited in the Studio; the API answers 405 to a write on them.
- Releasing a bike held by an order that fell through — the owner corrects the stock here by hand.
- Uploading picture files — the picture fields keep doing whatever they do in the approved design.
- New screens, new fields, a new component, or any restyle of the management kit. The failure notice is the kit's existing `<Notice tone="alert">`.
- Sliding sessions, "remember me", ending the owner's other sessions on sign-out, a password-strength rule (`services.md` §9, question 1: none for now).
- Any change to `App.tsx` (the `SignedIn` gate), `DashboardLayout.tsx`, `pages/AdminProducts.tsx`, `AdminProduct.tsx`, `AdminKinds.tsx`, `AdminSignIn.tsx`, `src/data/index.ts` (the data story's file), `theme.css`, `index.css`.
- The storefront's own pages waiting for a bike (plan item "Bike page waits for its record") and the checkout's refusal logic (orders story) — this story only proves their result after a real edit.
- Keeping the Studio's preview frame and the design walk signed in by writing the demo session key (SRS §7: TL's concern).

## 5. UI scope

No new screen. These existing screens and states are touched; every state below already exists in the design or is drawn with a part the kit already has.

| Screen | Touched by | States that must exist after this story |
|---|---|---|
| `/admin/sign-in` (`pages/AdminSignIn.tsx`, `SignInForm`) | real password check | default; empty-email message; the server's refusal sentence (401, 429) as a visible error with the typed email kept; "Could not sign in. Please try again."; the signed-in "Signed in as <email>." + "Continue" line; arrival after a session ended showing "Your session has ended. Sign in again." |
| Management frame (`DashboardLayout`) | the owner shown in the sidebar footer now comes from the server session | unchanged markup: initial, name made from the email, email, "Sign out" |
| `/admin` overview | live records; loading | loading (no number anywhere), default, nothing low, nothing matches |
| `/admin/products` and `/admin/kinds` (`ListPage`) | save, delete, edit read, riding-type names | loading ("Loading…"), default, nothing yet, nothing matches, add dialog, edit dialog, dialog saving (primary button disabled), dialog refused (alert notice inside, values kept), delete question, delete question refused (alert notice inside, "Delete" enabled again) |
| `/admin/products/:id` (`RecordPage`) | save, delete, loading | loading (neither form nor "not here"), default, "Saved." notice, alert notice above the form with values kept, delete question and its refused state, "This record is not here" |

## 6. Decisions made in this plan (so nothing is left to guess)

1. **Routes follow the architecture contract, not the plan item's sketch.** The plan item wrote `/v1/auth/sign-in`, `/v1/auth/sign-out`, `/v1/auth/me`; `docs/architecture/services.md` (the contract, written after it) fixes `POST /v1/sessions`, `DELETE /v1/sessions/current`, `GET /v1/me`. The contract wins. The browser reaches them as `/api/v1/...` (or `<VITE_API_URL>/v1/...`, the same base the data module uses). Every write is sent as `application/json` with the session cookie.
2. **The unchanged gate and a server session.** `SignedIn` in `App.tsx` reads `useSession()` synchronously and is not edited. So `session.ts` makes the `me` check once per page load, when the first page that reads the session mounts (no storefront page waits on it), and while that first answer is pending `useSession()` reports the owner this browser last signed in as — a non-secret hint (the email only, kept in the browser, never trusted for anything else). A `200` makes the server's email the session; a `401`, any other answer or no answer clears the hint and means nobody is signed in. Consequence the criteria are written around: a browser that signed in through the form reloads or opens a deep link without ever showing the sign-in page; a browser that holds a valid cookie but no hint (storage cleared, or the cookie obtained by calling the API directly) is sent to the sign-in page once and sees "Signed in as <email>. Continue" there. If TL judges that edge unacceptable, the fix is an edit to `SignedIn` in `App.tsx`, which this story does not own: raise it in the UI review, do not widen silently.
3. **The version token travels with the values it was read with.** A list carries no version (`services.md` §3.3), and a save must be refused if the record changed since the screen loaded it (ADMIN-008). So the token is never fetched at save time: **Edit** in the list and the record page read the record by id first and keep its `ETag` together with the values they show; `saveRow` sends it as `If-Match` on a change, keeps the token the answer returns for the next save, and sends none on an add or a delete. This adds one function to `store.ts` (additive; the three existing names are unchanged; its name is Dev's) and a read-by-id step in the two screens — a third widening of the two screens' changes, inside the same two files. In the list, if that read fails, the dialog does not open and an alert notice above the table shows the server's sentence (404: "That record does not exist."), or "The record could not be opened. Please try again." when no usable answer came — the only new sentence in this story.
4. **What the server is sent.** A change sends the whole record the screen holds (the server replaces the record), without `_source`, which the server refuses as an undefined key and never touches; hidden fields the form does not show (`price_listed`, `_invented`) are sent back as read, otherwise the replace would erase them. An add sends no `id`: the server makes it (`products-k3f9a2`). `newId` stays exported and still returns a text no other call returns, but no screen needs it. Numbers are sent as numbers, lists as arrays, `wheel` as the text `"29"` or `"mixed"`.
5. **Which sentence is shown.** If the answer is the service's error envelope, its `error` sentence is shown as it is. If there is no usable answer (network down, a proxy page instead of the envelope), the generic sentence is shown: "The change could not be saved. Please try again." (save), "The record could not be deleted. Please try again." (delete), "Could not sign in. Please try again." (sign-in).
6. **A 401 on a write** is shown as "Your session has ended. Sign in again." and `session.ts` drops the session, so the existing gate sends the owner to `/admin/sign-in`. The gated page unmounts at once, so the sentence is shown where it can be read: in the sign-in form's error place on arrival, until the next sign-in attempt. Values typed in an open dialog are lost in that one case.
7. **Sign-out that fails.** The session is dropped in the browser only after the server has answered `204`. If the request fails, the owner stays signed in and nothing else changes on screen (the button can be pressed again) — the server never reports "signed out" while the session still works (`services.md` §3.9).
8. **Riding types by name in the bikes list.** `ListPage` itself loads the riding types (`useCollection("categories")`) so the "Riding type" column and the form's riding-type choice read names, not ids such as `trail` (`fields.tsx` reads names from what some screen has loaded).
9. **Showing a change at once.** After an accepted write, `store.ts` signals the data module that the collection changed (DATA-007 Behaviour 7; the form of the signal is TL's — data SRS §7) so every `useCollection` and `useRecord` on screen for it asks again. A refused or failed write sends no signal. Public visitors need no signal: every API answer is `no-store`, so their next read is the new data.
10. **While loading, the overview shows no number at all** — not the figure and not the count inside a note ("0 still in draft" is also a 0). Dev's choice of what stands there (blank or a dash) is free as long as it is not a number.

## 7. Acceptance criteria

Numbered `SC-n`; each names the requirement and the SRS criterion it narrows (`SRS` ids are in `docs/admin/SRS.md`). Test writes at least one case per `SC`; the reviews count this list. Statuses, sentences and cookie attributes are those of `docs/architecture/services.md`.

### A. Backend — the owner account, signing in, the session

- SC-1 [ADMIN-001 AC-1, AC-3]: With `ADMIN_EMAIL=owner@deo.test` and a known `ADMIN_PASSWORD`, `POST /v1/sessions` with those answers 200 `{"email":"owner@deo.test"}`; the same with the email typed `  OWNER@deo.test ` is also 200.
- SC-2 [ADMIN-001 AC-2, AC-5]: The password's value is found nowhere — not in the repository (including `.env.example`), not in the sign-in, `me` or refused-sign-in answers, not in the server's log, not in any table. What the server holds is an argon2id hash made at start; `sessions` holds only hashes of tokens.
- SC-3 [ADMIN-001 AC-4]: With `ADMIN_PASSWORD` (or `ADMIN_EMAIL`) not set, the service still starts and public reads answer 200; every `POST /v1/sessions` answers 401 "That email and password do not match." and the log says, once per attempt, that the admin account is not configured.
- SC-4 [ADMIN-002 AC-1]: A successful sign-in sets the cookie `deo_session` with `HttpOnly`, `SameSite=Lax`, `Path=/` and an expiry 7 days after the sign-in (`Max-Age=604800` and a matching `Expires`); the body holds the email and no password.
- SC-5 [ADMIN-002 AC-7]: When the request arrived over HTTPS (`X-Forwarded-Proto: https`) the cookie also carries `Secure`; over plain HTTP it does not.
- SC-6 [ADMIN-002 AC-2]: A wrong password and, separately, a wrong email each answer 401 with the identical body: the sentence "That email and password do not match." and code `UNAUTHENTICATED`.
- SC-7 [ADMIN-002 AC-3]: A body with no password, and a body that is not an object, each answer 400.
- SC-8 [ADMIN-002 AC-4, AC-5]: After five wrong sign-ins from one address within 15 minutes, a sixth with the right password answers 429 "Too many attempts. Try again in a few minutes." (with `Retry-After`); once 15 minutes have passed since the last failed attempt, the right password is accepted.
- SC-9 [ADMIN-002 AC-6]: Four wrong sign-ins, then a right one, then four more wrong ones, then a right one: the last is accepted (the right one reset the count).
- SC-10 [ADMIN-003 AC-1, AC-2]: `GET /v1/me` with a valid session cookie answers 200 `{"email":"owner@deo.test"}`; with no cookie it answers 401 with one sentence ("You are not signed in.").
- SC-11 [ADMIN-003 AC-3; ADMIN §3]: A session older than 7 days, and a cookie with one character changed, each answer 401 on `GET /v1/me`; so does a cookie issued before the owner changed `ADMIN_EMAIL` or `ADMIN_PASSWORD` and the service restarted.
- SC-12 [ADMIN-003 AC-7 (cookie), AC-8, AC-9]: `DELETE /v1/sessions/current` answers 204 and sends a cookie that clears `deo_session` (`Max-Age=0`); replaying the old cookie value to `GET /v1/me` is then 401; the same call with no session also answers 204 and changes nothing.

### B. Backend — who may write, what is saved, what is deleted

- SC-13 [ADMIN-004 AC-1, AC-2, AC-3; ADMIN-006 AC-10]: With no session, and with an expired one, `POST`, `PUT` and `DELETE` on bikes and on riding types each answer 401 "You are not signed in." and the records are unchanged.
- SC-14 [ADMIN-004 AC-5]: For a signed-in owner, a change to a review, a change to the shop record, a delete of a gallery photograph and an add of a team member each answer 405 (with `Allow: GET`) and change nothing; without a session they answer 405 too.
- SC-15 [ADMIN-004 AC-4, AC-6]: With no session, listing the bikes and reading one each answer 200; a signed-in owner's add of a bike is accepted (SC-16).
- SC-16 [ADMIN-005 AC-1, AC-2, AC-19]: A signed-in owner adds a bike with a model name, brand `Giant`, riding type `trail`, price 25000000, status `draft`, stock 3 and no id: 201, the record carries a server-made id like `products-k3f9a2` and the same values, and reading it returns it. The same body with `"id": "my-bike-1"` is 201 with that id, and adding it again is 409 "That id is already used." A riding type with a name only is 201 with a server-made id and the name.
- SC-17 [ADMIN-005 AC-3, AC-4, AC-5]: A change of `talon-2` with stock 5 is 200 and reading it shows stock 5 and every other field as sent; a field the old record had and the body omits is gone from the saved record; a change to `no-such-bike` is 404.
- SC-18 [ADMIN-005 AC-6, AC-7, AC-20; Behaviour 8]: A body with `"colourful": "red"` is 400 naming `colourful`; a bike body with no model name, and one with a blank model name, are 400 naming the name; a riding type with no name is 400 naming the name. Every refusal is one sentence in the error envelope (`error`, `code`, and `field` where one field is at fault).
- SC-19 [ADMIN-005 AC-8, AC-9, AC-13]: `price` "cheap", -1 and 1.5 are each 400 naming `price`; `stock` -1 and 2.5 are each 400 naming `stock`; `sizes` "M" (not a list) and `deal` "yes" are each 400 naming the field.
- SC-20 [ADMIN-005 AC-10, AC-11, AC-12]: `status` "sold", `wheel` "26", `brand` "Trek" and `category` "unicycle" are each 400 naming the field.
- SC-21 [ADMIN-005 AC-14, AC-15]: A `live` bike with no price, and one with price 0, are each 400 "A live bike needs a price."; a `draft` bike with no price is accepted.
- SC-22 [ADMIN-005 AC-16]: A body with `was` empty saves a record with no `was` key, and no key of any saved record is `null`.
- SC-23 [ADMIN-005 AC-17, AC-18]: On a change, a body `id` that differs from the address is 400; an id with capital letters or a space is 400.
- SC-24 [ADMIN-005 AC-21]: A text of 5000 characters is accepted; one of 5001 is 400.
- SC-25 [ADMIN-005 AC-22]: After any refused body, reading the records shows nothing was saved.
- SC-26 [ADMIN-006 AC-1, AC-9]: Deleting a bike that has no orders is 204; reading it is then 404 and the shop, Home and search no longer show it; deleting an id that does not exist is 404.
- SC-27 [ADMIN-006 AC-2, AC-3, AC-4]: Deleting the riding type `trail`, used by 4 bikes, is 409 "Trail is still the riding type of 4 bikes. Give those bikes another riding type first." and the type still exists; one used by exactly 1 bike is 409 with "…of 1 bike."; one no bike uses is 204.
- SC-28 [ADMIN-006 AC-5, AC-6]: After a riding type a gallery photograph carried is deleted, the photograph is still listed and has no riding type; after a bike with a review is deleted, the review still exists and has no `bike`.
- SC-29 [ADMIN-006 AC-7, AC-8]: Deleting a bike that has an order is 204 and the order's lines are unchanged; a bag still holding that bike is then refused at checkout with 409 "A bike in your bag is no longer in the shop." (the refusal is the orders story's; it is proved here after a real delete).
- SC-30 [ADMIN-007 AC-1, AC-2]: After the owner saves `talon-2` (stock 9) with stock 0, a visitor's read of it shows stock 0; the visitor's shop, Home and the bike's page show "Out of stock", the button reads "Ask when it is back", and Home's "bikes in stock" count is one lower.
- SC-31 [ADMIN-007 AC-3, AC-4]: After the owner saves `talon-2` as `draft`, a visitor's `/shop`, Home, `/search?q=talon` and the "More" row of a bike of the same type no longer list it; after the owner saves its price as 17990000 (from 18990000), the visitor's card and bike page show 17.990.000 ₫ and the shop's price order uses it.
- SC-32 [ADMIN-007 AC-7, AC-8]: After a save refused with 400, the public pages show what they showed before; after a bike is deleted, a visitor's read of it is 404 and the shop no longer lists it.
- SC-33 [ADMIN-008 AC-1, AC-2, AC-3]: The owner reads `talon-2` at stock 9; an order for 2 is then accepted (stock 7); a `PUT` with the token from the first read is 409 with the sentence "This record changed while you were editing it — for example an order was placed. Reload and make your change again.", and reading `talon-2` shows stock 7 and the old price. A `PUT` made after reading again is 200 and the record has stock 7 and the new price. Two clients that read the same bike: the one that saves second is 409 and the first one's change stands. A `PUT` with no `If-Match` is 428.
- SC-34 [ADMIN-008 AC-4, AC-5]: After the owner saves stock 5, an order for 1 leaves stock 4. For a bike at stock 9, a save of stock 5 and an order for 1 sent at the same moment end either as save 200 + order 201 with stock 4, or as save 409 with stock 8 — never stock 5 with the order accepted.
- SC-35 [ADMIN-008 AC-6]: Deleting a bike that has changed since a screen loaded it is 204 (a delete sends no token and is not refused for staleness).

### C. Frontend — the sign-in screen and the session

"Signed in through the form" means the owner pressed "Sign in" in this browser (decision 2).

- SC-36 [ADMIN-002 AC-8, AC-16]: `/admin/sign-in` still reads as designed with the real check in place: the photograph, the navy panel with "Mountain bikes for passes, trails and rough ground — chosen, built and set up by people who ride them.", the logo, the `h1` "Welcome / back", "Sign in to run Đèo Bikes.", "Email" (hint "you@example.com"), "Password" (characters hidden), the button "Sign in", "Ask the shop's owner for an account." and the link "Back to the shop" to the home page; in a 390 px wide window the photograph and navy panel are above the form.
- SC-37 [ADMIN-002 AC-9, AC-14]: With the right email and password typed, pressing Enter in the password field sends `{email, password}` to the sign-in endpoint, and on success the browser goes to `/admin`. A wrong password never reaches `/admin` (the password is really checked).
- SC-38 [ADMIN-002 AC-10]: With the email field empty, pressing "Sign in" shows "An email address is needed" under the email field and sends no request.
- SC-39 [ADMIN-002 AC-11, AC-12]: A wrong password shows "That email and password do not match." as a visible error in the form with the typed email still in the field; an answer of 429 shows "Too many attempts. Try again in a few minutes." in the same place.
- SC-40 [ADMIN-002 AC-13]: When the server cannot be reached, pressing "Sign in" shows "Could not sign in. Please try again." in the form.
- SC-41 [ADMIN-002 AC-15]: For a signed-in owner, `/admin/sign-in` shows under the button "Signed in as owner@deo.test." and a "Continue" button that goes to `/admin`.
- SC-42 [ADMIN-003 AC-4, AC-5]: An owner signed in through the form who reloads `/admin` sees the overview, and one who types `/admin/products/stance-0` into the address bar sees that bike's record page; in both, `/admin/sign-in` is never shown at any moment (it never enters the browser's history during the load).
- SC-43 [ADMIN-003 AC-6, AC-12]: With no session, opening `/admin`, `/admin/products`, `/admin/products/stance-0` and `/admin/kinds` each lands on `/admin/sign-in`; when the `me` check cannot reach the server, opening `/admin` also lands there.
- SC-44 [ADMIN-003 AC-7, AC-13]: Pressing "Sign out" in the sidebar leaves the browser at `/admin/sign-in` with the cookie cleared by the server's answer; pressing the browser's Back button then shows the sign-in page, not a management page.
- SC-45 [ADMIN-003 AC-10, AC-11; Behaviour 5]: For the owner `lan.tran@deo.test` the sidebar footer shows "L", "Lan Tran", "lan.tran@deo.test" and a "Sign out" button; `useRole()` returns the empty text; `useSession`, `signIn`, `signOut` and `useRole` are all still exported by `src/lib/session.ts`, and `App.tsx` and `DashboardLayout.tsx` are unchanged.

### D. Frontend — saving, waiting and failing

- SC-46 [ADMIN-013 AC-1]: `saveRow`, `deleteRow` and `newId` are importable from `components/admin/store.ts` under those names; two `newId` calls return different texts; `saveRow` and `deleteRow` return promises that resolve on success and reject carrying the sentence to show on failure (a unit test with the network faked).
- SC-47 [ADMIN-013 AC-12; ADMIN-007 AC-5]: A successful save in the edit dialog closes it and the list shows the change without a reload; stock saved from the record page is shown by the bikes list, opened again, without a reload.
- SC-48 [ADMIN-013 AC-2]: While a save is in flight the dialog's "Add" or "Save changes" button is disabled, and so is the record page's "Save changes"; while a delete is in flight the question's "Delete" is disabled; one press makes one request.
- SC-49 [ADMIN-013 AC-3; ADMIN-008 AC-7]: In the edit dialog, when the server answers 409 with the sentence of ADMIN-008, "Save changes" leaves the dialog open, holding the typed values, with the sentence in an alert notice inside it.
- SC-50 [ADMIN-013 AC-4]: In the add dialog, when the server answers 400 "A live bike needs a price.", "Add" leaves the dialog open with its values and shows the sentence.
- SC-51 [ADMIN-013 AC-5]: On the record page, when the server answers 400, "Save changes" keeps the form's values, shows the sentence in an alert notice above the form, and does not show "Saved.".
- SC-52 [ADMIN-013 AC-6]: In the delete question for the riding type `trail`, when the server answers 409, "Delete" leaves the question open with the sentence in an alert notice inside it and "Delete" enabled again — on the riding types list and, for a bike the server refuses, on the record page.
- SC-53 [ADMIN-013 AC-7]: When the server cannot be reached, "Save changes" and, separately, "Delete" show "The change could not be saved. Please try again." and "The record could not be deleted. Please try again." in the same places as the server's sentences.
- SC-54 [ADMIN-013 AC-8]: When the server answers 401 to a save, "Your session has ended. Sign in again." shows and the owner is sent to `/admin/sign-in`, where the same sentence is visible in the form's error place.
- SC-55 [ADMIN-013 AC-9]: After a failed save, the management lists and the public pages show what they showed before.
- SC-56 [ADMIN-013 AC-13]: A delete confirmed from the record page opens the bikes list; a delete confirmed from a list closes its question with the row gone from the table.
- SC-57 [ADMIN-008 AC-1, AC-3 (the screens' half)]: With the edit dialog (and, separately, the record page) open on `talon-2` at stock 9, an order for 2 is accepted; changing the price and pressing "Save changes" is refused with the ADMIN-008 sentence, the typed values stay, and the server still holds stock 7. After closing and opening Edit again the dialog shows stock 7, and a save is accepted (the token comes from the read that gave the values, never from a read at save time).
- SC-58 [ADMIN-005 AC-3, AC-4; ADMIN-011 AC-4]: Opening a seeded bike (it has `_source`, `price_listed` and `_invented`), changing only its stock and saving — from the dialog and from the record page — is accepted, and the saved record equals the seeded one except for the stock: the hidden fields are kept, and the server never refuses `_source`.
- SC-59 [ADMIN-013 AC-7; decision 3]: In a list, when the by-id read that "Edit" makes fails, the dialog does not open and an alert notice above the table shows the server's sentence ("That record does not exist." for a bike deleted in another window) or "The record could not be opened. Please try again." when no usable answer came.

### E. Frontend — what each management screen does on live data

- SC-60 [ADMIN-009 AC-1, AC-2, AC-4]: Every management page still shows the sidebar (shop name, "Overview", group "Shop" with "Bikes" and "Riding types", the owner at the foot), the title bar and the marked current entry; below 768 px the sidebar is a drawer. On the overview, "Bikes", "Running low" and "Out of stock" open `/admin/products`; "Riding types" opens `/admin/kinds`.
- SC-61 [ADMIN-009 AC-3, AC-5, AC-6]: On the seeded database `/admin` shows "Bikes 14 — 0 still in draft", "Riding types 4 — shown on the site", "Running low 14 — under ten in stock" and "Out of stock 2 — marked out of stock". A bike with stock 9 is listed and counted as running low and one with stock 10 is not; a draft bike with stock 2 is counted and listed and "Bikes" then shows "1 still in draft".
- SC-62 [ADMIN-009 AC-7 to AC-12]: The "Running low" panel, on live records, has the columns "Bike", "Riding type", "In stock" ordered fewest first with a danger badge at 0; "Stance" typed in "Search bikes" narrows by model name ignoring capitals, shows "Clear filters" and is in the address; "Any riding type" narrows to one type; "Bike" pressed twice sorts by model name then reverses; the foot reads "1–14 of 14", has a rows-per-page select of 20, 50 and 100, "Export these 14" downloading a CSV of the 14 rows, and "Previous" (disabled), "1", "Next".
- SC-63 [ADMIN-009 AC-13, AC-14]: When every bike has 10 or more in stock the panel reads "Nothing is running low." and shows no table; a search that matches nothing reads "No bike running low matches. Clear the search or the riding type to see them all."
- SC-64 [ADMIN-009 AC-15; ADMIN-007 AC-6]: With the overview open, a bike's stock saved as 12 from another screen of the same window makes "Running low" one lower and the bike leaves the panel, without a reload.
- SC-65 [ADMIN-009 AC-16; ADMIN-013 AC-11]: Opening `/admin` by link, at no moment before the records arrive does any card show a number (the figure or the count in its note) or the panel read "Nothing is running low."
- SC-66 [ADMIN-010 AC-1, AC-2, AC-17, AC-18]: `/admin/products` opens with the title "Bikes", "Add", "Search by model", "Status" and a table of 14 rows with the photograph, "Model", "Brand", "Riding type", "Model year" and "Price" (as `18.990.000 ₫`, right-aligned), where a bike's riding type reads "Trail", not "trail"; at 390 px only the photograph, "Model" and "Brand" show; while the first answer is awaited the table reads "Loading…"; a "Model" link opens `/admin/products/<id>`.
- SC-67 [ADMIN-010 AC-3, AC-4, AC-5, AC-6, AC-19]: "Model" pressed three times sorts ascending, descending, then the seeded order; "Price" sorts by number; "stance" typed in "Search by model" lists only bikes with "stance" in their text, shows "Clear filters" and is in the address; "Draft" in "Status" lists only draft bikes (with none, "Nothing matches those filters"); the rows-per-page select set to 50 puts 50 in the address and the foot still reads "1–14 of 14".
- SC-68 [ADMIN-010 AC-7, AC-8, AC-11, AC-13, AC-20]: A row's "Actions" shows "Edit" and, below a rule, "Delete"; "Add" opens "Add to bikes" with the visible fields in the schema's order, their labels and hints, "Cancel" and "Add", and the "Bikes in stock" field carries its hint "0 shows 'Out of stock' and turns the button into 'Ask when it is back'."; "Edit" opens "Edit <model>" with "Changes show on the site as soon as they are saved.", "Cancel", "Save changes" and the bike's values; "Delete" opens "Delete <model>?" with "This cannot be undone.", "Cancel" and "Delete".
- SC-69 [ADMIN-010 AC-9]: "Add" with "Model" empty shows "Model is needed" under the field and sends nothing.
- SC-70 [ADMIN-010 AC-10]: Adding a bike with a model "Test Bike", brand Giant, riding type Trail closes the dialog, the bike appears in the list, and a visitor reading the bikes finds it.
- SC-71 [ADMIN-010 AC-12]: Saving the edit dialog with stock 0 closes it, and a visitor's page for that bike reads "Out of stock".
- SC-72 [ADMIN-010 AC-14]: Confirming "Delete" removes the bike from the list, and a visitor's read of it is 404.
- SC-73 [ADMIN-010 AC-15, AC-16]: With no bikes and no filter the list reads "Nothing here yet" with an "Add" button; a filter that matches nothing reads "Nothing matches those filters", "Clear a filter to see more." and "Clear filters".
- SC-74 [ADMIN-011 AC-1, AC-2, AC-3, AC-4, AC-11]: `/admin/products/stance-0` shows the breadcrumb "Bikes › <model>" (its "Bikes" opens the list), the model as heading, "Delete" and "Save changes"; the four tabs "Basics", "Stock", "Specification" and "Words" carry the fields of ADMIN-011 Behaviour 2, each hint under its control (including "Shown struck through beside the price. Leave empty when nothing was reduced." under "Price before"); "Price on the maker's page" and the invented-fields list are not shown.
- SC-75 [ADMIN-011 AC-5, AC-6]: On the "Stock" tab, stock changed to 4 and "Save changes" shows "Saved." with a dismiss button, and a visitor's read of the bike shows stock 4; editing any field then hides "Saved.".
- SC-76 [ADMIN-011 AC-7]: With the model cleared on "Basics", "Save changes" shows "Model is needed" under the field and sends nothing.
- SC-77 [ADMIN-011 AC-8]: "Delete" pressed and confirmed deletes the bike and opens the bikes list.
- SC-78 [ADMIN-011 AC-9]: `/admin/products/no-such-bike`, after the server has answered, shows "This record is not here", "It may have been deleted since the link was made." and a "Back to the list" button to the list.
- SC-79 [ADMIN-011 AC-10]: A bike's riding type changed on "Basics" and saved shows the new riding type in the bikes list and in the shop.
- SC-80 [ADMIN-011 AC-12; ADMIN-013 AC-10]: Opening `/admin/products/stance-0` by link, at no moment before the form appears does "This record is not here" show, and the form is not drawn before the record has arrived.
- SC-81 [ADMIN-012 AC-1, AC-2, AC-3, AC-4]: `/admin/kinds` opens with the title "Riding types", "Add", a search box and 4 rows with "Name", "Who it is for" and "Where it rides best"; a row is not a link and its "Actions" menu has "Edit" and "Delete"; "Add" opens "Add to riding types" with the three fields; with the name empty "Add" shows "Name is needed" and sends nothing.
- SC-82 [ADMIN-012 AC-5, AC-6]: A riding type "Gravel" added shows a chip "Gravel", a footer link and a "where to start" row for it on the shop; a riding type renamed "Cross country" shows the new name on the shop's chips, the breadcrumbs, the cards and the bikes list's "Riding type" column.
- SC-83 [ADMIN-012 AC-7, AC-8]: "Delete" confirmed on `trail` (used by 4 bikes) leaves the question open showing "Trail is still the riding type of 4 bikes. Give those bikes another riding type first." and the type still listed; "Delete" confirmed on a type no bike uses removes it from the list and from the shop's chips.

## 8. Dependencies

| Needs | From | Why |
|---|---|---|
| Tables for bikes and riding types with the `version` column and the `touch_record` trigger; the read API incl. `GET /v1/{collection}/{id}` | Story "Site data from the database" | Every write, the token of ADMIN-008 and the 405/404 rules sit on them |
| The data module with `useCollection` / `useRecord` returning `loading`, and its "this collection changed" signal (DATA-007 Behaviour 7), `collections()` | Story "Site data from the database" | The screens redraw after a save and read riding-type names |
| `POST /v1/orders` and the conditional stock hold | Story "Place an order and hold the bike" | SC-29, SC-33, SC-34 need a real order lowering a bike's stock; until it lands SC-33 can be shown with a direct update of the bike row (the trigger bumps `version`), SC-34 cannot |
| `ADMIN_EMAIL`, `ADMIN_PASSWORD` set in the run environment | The owner, in the console (requested by this plan) | SC-1, SC-4 to SC-12 and all of section C need a known account. If the owner has not supplied them yet, the local run and the verification use throwaway values that are not committed, and say so; SC-3 is proved by their absence |
| Edge forwards `X-Forwarded-For` and `X-Forwarded-Proto`; the site and `/api` share one origin | The deployment (existing) | The throttle keys on the forwarded address and `Secure` follows the forwarded protocol; with `VITE_API_URL` pointing at another origin the owner's calls do not work (`services.md` §9, question 3) |

No provider account, webhook or redirect address is involved, so there is no `request_setup`.

## 9. Files this story owns

**Frontend** (exactly these; nothing else in `code/frontend/` changes):

- `src/lib/session.ts` (and a new `src/test/session.test.ts` for it)
- `src/components/admin/SignInPage.tsx`
- `src/components/admin/store.ts`
- `src/components/admin/ListPage.tsx`, `src/components/admin/RecordPage.tsx`
- `src/components/admin/index.tsx` — the `Confirm` component only: it does not close until the delete has succeeded, shows the server's sentence in a `<Notice tone="alert">` inside it when the delete fails, and disables "Delete" while the request is in flight; no other component in that file changes
- `src/pages/AdminOverview.tsx`
- `src/test/admin.test.ts`

**Backend** (new files beside the data story's; the read handlers are not rewritten): the auth package (sign-in, `me`, sign-out, throttle, `requireOwner`), the write handlers and validators for `products` and `categories`, one added line per route in `internal/server` `Router`, migration `20261118000400_sessions.up.sql` / `.down.sql`, `code/backend/.env.example` (the two names, no values), and Go tests that run without a database.

## 10. Traceability

| Requirement | Criteria |
|---|---|
| ADMIN-001 One owner account, from secrets | SC-1, SC-2, SC-3 |
| ADMIN-002 Signing in | SC-4 to SC-9 (server); SC-36 to SC-41 (screen) |
| ADMIN-003 The session | SC-10 to SC-12 (server); SC-42 to SC-45 (screens) |
| ADMIN-004 Only the owner may write, only bikes and riding types | SC-13, SC-14, SC-15 |
| ADMIN-005 Saving a bike or a riding type | SC-16 to SC-25; SC-58 |
| ADMIN-006 Deleting | SC-13, SC-26 to SC-29 |
| ADMIN-007 An edit shows at once | SC-30 to SC-32; SC-47; SC-64 |
| ADMIN-008 An edit and an order do not overwrite each other | SC-33 to SC-35; SC-49; SC-57 |
| ADMIN-009 Frame and overview | SC-60 to SC-65 |
| ADMIN-010 Bikes list | SC-66 to SC-73 |
| ADMIN-011 Bike record page | SC-58; SC-74 to SC-80 |
| ADMIN-012 Riding types list | SC-81 to SC-83 |
| ADMIN-013 Screens wait and show a failure | SC-46 to SC-56; SC-59; SC-65; SC-80 |

## 11. Notes for the reviews

- **Decision 3 is a third widening of the same two screens** (read the record by id before editing, plus the one new sentence). The two earlier widenings (await the promise, `Confirm`, `AdminOverview`) are in the plan item; this one is needed because the contract's token cannot reach `saveRow` otherwise. It adds no file.
- **Decision 2's edge** (cookie without a hint) is a known, stated limit, not a defect to fix inside these files.
- **No blocking question for the owner.** The SRS §7 open items already carry defaults that the criteria use (7 days, 5 failures in 15 minutes, a live bike needs a price, deletes leave reviews and photographs); the owner can change them later. The only thing the owner must do is enter the two secrets in the console.
