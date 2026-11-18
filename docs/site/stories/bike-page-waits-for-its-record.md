# Story — Bike page waits for its record

Module: `site` · Plan item: "Bike page waits for its record" (bug, P1)
Requirement: SITE-016 (`docs/site/SRS.md` §4.2), which narrows SITE-005 Behaviour 12 (the not-found view)
Design: `design/design-system.md` §2.23 (Bike not found) · Screen: Bike `/shop/:id`, `pages/Product.tsx`

## 1. User story

As a **visitor** (the storefront's only actor, SRS §2), I want a bike's page to show the bike as soon as it
arrives and not say it is missing first, so that a link to a bike never greets me with "That bike is no longer
in the shop."

## 2. Why this exists

Once the records come from the API (plan item "Site data from the database", DATA-007), `useRecord` answers
after a request, not at once. `pages/Product.tsx` reads only `item` from `useRecord` and draws the not-found
view whenever `item` is empty. So every bike page opened from a link, from the shop grid or from a "More" card
would flash "That bike is no longer in the shop." before the bike arrives, and a bike that is genuinely gone
could not be told apart from one still on its way.

## 3. In scope

- `src/pages/Product.tsx` reads `loading` and `error` from `useRecord` (both already exported by the data
  module, DATA-007 Behaviour 1) and chooses what the bike's place on the page shows:

| What the shop has said | `item` | `loading` | `error` | The page draws |
|---|---|---|---|---|
| A record is in hand (first answer, or a re-ask in flight) | record | any | any | The bike's page, as SITE-005 |
| Nothing yet | empty | true | — | Nothing in the bike's place (header and footer stay) |
| "No such bike" (404) | empty | false | null | The existing not-found view: "That bike is no longer in the shop." and the primary button "See all bikes" to `/shop` |
| The shop could not be reached, or answered 5xx | empty | false | an `Error` | Nothing in the bike's place (SITE-016 failure table); the not-found view is not shown, because the bike has not been said to be gone |

- Rule in one line: a record in hand is always drawn; with none, wait while `loading` or after a failure
  (`error`); show the not-found view only once the shop has answered that there is no such bike.
- The hooks the page already calls (`useCollection` for the riding types, `usePageMeta`) stay above the new
  early return, in the same order, so the order of hooks never changes between renders.

## 4. Out of scope

- A spinner, skeleton, placeholder shape, "Loading…" line or any other new screen. The design draws none for
  a bike's page (§2.23 draws a loading line for the shop's grid only) and the plan item says none is invented.
  The wait is a short, transient blank, not an empty state: nothing is missing yet, so §2.23's "every empty
  state says what is missing" does not apply.
- The data module (`src/data/index.ts`) and what `useRecord` returns — that is DATA-007, in the story "Site
  data from the database". This story only reads what it already promises.
- Any other page or component: the management record page and overview wait for their records in ADMIN-013
  (plan item "Shop admin: real sign-in and saved edits"); Home, the shop grid and search already hide or label
  what is not there yet.
- A retry button, error message or offline screen for a failed request. The design draws none; accepted at
  launch (`docs/data/SRS.md` §7, open question 1).
- The riding type's pill, the middle link of the breadcrumb and the head's tint arrive when the riding types
  answer, as they do today (`Crumbs` and `Head` already cope with no type). SITE-016 waits for the bike only.
  The reviews tab, the promise list and the "More" row likewise appear when their own records arrive.
- A bike in draft opened by its address (SITE-005 open question 2) — left as designed.
- Any copy, token or content change: no new key in `content.json`, nothing in `theme.css` or `index.css`.
- Backend, schema and API: none. `GET /v1/products/{id}` and its `404 NOT_FOUND` are built by "Site data from
  the database" (`docs/architecture/services.md` §3.3; SITE-016 row of the requirement map). Steps 6–8 of the
  delivery flow have nothing to design or build here.

## 5. UI scope

Screen: Bike `/shop/:id` (`pages/Product.tsx`). States, per SRS §5: **default** and **not found**, both
unchanged, plus the moment before either: **waiting**, in which the bike's place is empty.

- Waiting: nothing is drawn by the page — no text, no icon, no spinner, no box. The frame (header, menu,
  search, footer, contact dock) belongs to the layout and stays, so only the page's own area is blank.
- Default: every section SITE-005 describes, byte-for-byte the same output as before this change — breadcrumb,
  gallery, tinted head, price and stock, sizes, buy buttons, promises, specification, size guide, tabs, ask
  band, "More" row. No class, token or copy changes.
- Not found: the existing `Missing` view, unchanged — the `h1` "That bike is no longer in the shop." and the
  primary pill "See all bikes" to `/shop`.
- Viewports: the behaviour is the same at 390 px and 1280 px; there is no layout work.
- Tab title: unchanged; `usePageMeta` still runs while the page waits.

How the wait is observed (for Test): hold the request for `GET /api/v1/products/<id>` for a few seconds, and
record the page's text from the start of navigation (for example a `MutationObserver` registered before the
page's scripts run). "At no moment" means in every recorded state, not only in the final one.

## 6. Acceptance criteria

SC-1 [SITE-016 AC-1]: With the bike's answer delayed, `/shop/stance-0` opened by typing the address (a full
page load) never contains the text "That bike is no longer in the shop." or the button "See all bikes" at any
moment before the bike is drawn; once the answer arrives the bike's page appears.

SC-2 [SITE-016 AC-2]: Arriving at a bike's page by pressing a bike card — on the shop grid, in the "More" row
of another bike's page, or in a Home section — the same holds: no not-found text at any moment between the
press and the bike appearing. When pressed from another bike's page the previous bike is not shown under the
new address while the new one is awaited.

SC-3 [SITE-016 Behaviour 1]: While the bike's answer is awaited the page draws nothing of its own: no text, no
spinner, no placeholder shape, no breadcrumb and no `h1`; the header and the footer are drawn as on any page.

SC-4 [SITE-016 AC-4]: When the shop answers with a bike that exists, the page is drawn exactly as SITE-005
describes. Spot checks that must still pass on the finished page: the breadcrumb Bikes › type › name
(SITE-005 AC-1); the price "18.990.000 ₫" in large bold figures (AC-7); size pills with "M" chosen (AC-10);
"Add to bag" putting the chosen quantity and size in the bag (AC-13); the ask band and the "More" row (AC-24).

SC-5 [SITE-016 AC-3, SITE-005 AC-26]: `/shop/no-such-bike`, once the shop has answered 404, shows the `h1`
"That bike is no longer in the shop." and a primary button "See all bikes" that opens `/shop`; neither is
drawn before that answer arrives (the area is empty while the request is held).

SC-6 [SITE-016 failure table, "Upstream failure"]: When the request for the bike fails — the shop cannot be
reached, or it answers 500 — the bike's place stays empty and "That bike is no longer in the shop." is never
shown; no error screen or message is drawn.

SC-7 [SITE-016 Behaviour 2]: A bike already in hand stays drawn when the shop is asked again (`item` present
with `loading` true, for example the data module's refresh after a record changes, DATA-007 AC-16): the page
does not go blank and does not show the not-found view. (Observable with the data hook stubbed in a component
test.)

SC-8 [SITE-016 AC-4]: The change's file list is `code/frontend/src/pages/Product.tsx` and nothing else: no
other page, component, library, data-module, `content.json`, `theme.css` or `index.css` change, and the
page's hooks are still called in the same order before the early returns. `npm run lint`, the type-check and
the build pass.

## 7. Dependencies

- **Must land first: "Site data from the database"** (plan item 2, module `data`, DATA-007). Until `useRecord`
  returns a real `loading`, it is the constant `false` of the file-backed mock and this change is invisible, so
  SC-1, SC-2, SC-3, SC-5 and SC-6 cannot be observed before it merges.
- **A precondition this story depends on and does not own:** `useRecord` must answer `loading: true` on the
  very first render for an id — and on the first render after the id changes — not one render later, from an
  effect. Otherwise the not-found view flashes for one frame whatever this page does (DATA-007 AC-13, AC-14).
  If Test sees that, the defect is the data module's hook (DATA-007), not this file: report it against that
  story and do not widen this change.
- **Data that must exist:** the bike `stance-0` in the records (it is one of the shop's live bikes), and no
  bike with id `no-such-bike`. The API route `GET /v1/products/{id}` (200, and 404 `NOT_FOUND` for an unknown id).
- No external account, key or secret. No backend, schema or API work.

## 8. Decisions made here (not blocking)

- **Failure shows nothing, not the not-found view.** SITE-016 already says an unreachable shop leaves the
  bike's place empty. Saying "no longer in the shop" about a bike nobody has said is gone would be untrue, so
  the page reads `error` as well as `loading`. Both are already returned by `useRecord`, and the change stays in
  one file.
- **A record in hand is always drawn**, so a background refresh never blanks a page the visitor is reading.
- **No spinner.** The design draws none and the owner approved the design; if the owner later wants one, it is
  a design change, not this story.
