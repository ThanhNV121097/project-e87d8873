# Service contract extension — Bike page waits for its record

Story: `docs/site/stories/bike-page-waits-for-its-record.md` · Requirement: SITE-016 (`docs/site/SRS.md` §4.2) · Plan item: *Bike page waits for its record* (bug, P1)

## 1. Result

**This story adds no endpoint, no request or response field, no error code and no pagination or auth rule.** The service contract is not extended.

The only call the changed page makes is `GET /v1/products/{id}`, which is `GET /v1/{collection}/{id}` of `docs/architecture/services.md` §3.3. That section already lists SITE-016 in its *Traces to* line and in the requirement map, and its error table already gives the answers this story needs. Nothing below changes it; this file only pins how the page must read those answers, so the backend and the frontend agree on what each answer means.

## 2. The one call this story depends on

`GET /v1/products/{id}` — browsers reach it as `/api/v1/products/{id}` (the edge proxy strips `/api`; `services.md` §2.1). Written here without the `/api` prefix, as the backend mounts it.

| Aspect | Value (all from `services.md` §3.3, §2.3, §2.1) |
|---|---|
| Auth | None. Any visitor |
| Path parameter | `id`, any text, no format check. The page sends the id of the address (`stance-0`, `no-such-bike`) |
| Request body | None |
| Success | `200` with the bike record of §2.7 (the same record the list holds for that id) and the `ETag` header, which this page ignores |
| Caching | `Cache-Control: no-store` on every answer, so a held or repeated request is never answered from a cache |
| Pagination | Not applicable: one record |

## 3. Answers the page tells apart

The page's three outcomes depend on which answer arrives. The error codes below are the project's existing catalog (`services.md` §2.3); no new code is introduced.

| Answer | HTTP · `code` | Fixed sentence (`error`) | What the data module reports (DATA-007.6) | What the page draws (story §3) |
|---|---|---|---|---|
| The bike exists | `200` | — | `item` = record, `loading` false, `error` null | The bike's page (SITE-005) |
| No bike has that id | `404` `NOT_FOUND` | "That record does not exist." | `item` empty, `loading` false, **`error` null** | The not-found view: "That bike is no longer in the shop." and "See all bikes" → `/shop` |
| An id with capitals or any shape no bike can have (for example `Stance-0`) | `404` `NOT_FOUND` | "That record does not exist." | As the row above | The not-found view |
| The query failed | `500` `INTERNAL` | "Something went wrong. Please try again." | `item` empty, `loading` false, `error` an `Error` | Nothing in the bike's place; **not** the not-found view |
| The database cannot be reached (a connection that could not be opened, refused, lost or closed under the query) or the process is shutting down | `503` `UNAVAILABLE` | "The service is not ready." | As the row above | Nothing in the bike's place |
| `200` whose body is not a JSON object (never sent by the backend; a proxy's error page, say) | `200` | — | As the row above: `error` an `Error` | Nothing in the bike's place |
| The network failed, or the edge proxy answered `502`/`504` | no backend envelope | — | As the row above | Nothing in the bike's place |
| Before any answer | — | — | `item` empty, `loading` **true** from the very first render for the id | Nothing in the bike's place |

Two consequences, both already true of the merged contract, stated so neither side drifts:

1. **A `404` is the only answer that means "no such bike".** The backend answers `404 NOT_FOUND` for an unknown bike and never for a failure: a failed query is `500 INTERNAL`, a refused connection is `503 UNAVAILABLE` (`services.md` §2.3, §3.3). That is what lets the page treat `error` as "do not claim the bike is gone".
2. **The sentence of a failure is never shown on this page.** `services.md` §5 (the PostgreSQL row) says the storefront "shows its existing empty states" when a read fails. For the bike page the empty place of SITE-016 (failure table, "Upstream failure") is that state; the not-found view is not an empty state of a failed read, because the shop has not said the bike is gone. This clarifies §5 for this one page and contradicts nothing in it.

## 4. Where the `404` is decided in the backend

The route handler of §3.3 already answers `404 NOT_FOUND` after the collection check and a primary-key read that finds no row (check order of §2.3: route and method → collection → record existence). Backend work for this story: **none**. The handler and its tests belong to *Site data from the database* (SC-52 of `docs/data/stories/site-data-from-the-database.md`: `no-such-bike` answers 404 with `code` `NOT_FOUND`). The test cases of this story (`docs/site/test-cases/bike-page-waits-for-its-record.md`, TC-02 and TC-23) pin the same envelope from the outside.

## 5. Preconditions that sit outside this story

| Precondition | Owner | If broken |
|---|---|---|
| `useRecord` answers `loading: true` on the very first render for an id, and on the first render after the id changes (DATA-007 AC-13, AC-14), `error` null on a `404` and an `Error` on any other failure (DATA-007.6) | *Site data from the database*, `src/data/index.ts` | One frame of the not-found view shows whatever `Product.tsx` does. The defect is the data module's; report it against DATA-007, do not widen this change |
| The edge proxy forwards `/api/v1/products/{id}` and does not answer a `404` of its own for it | Deployment (`docker-compose` / edge config, committed and not edited by stories) | A proxy `404` would read as "no such bike", the same as a backend `404`. A mis-routed proxy fails every page on the site at once, not this page alone, so no per-story guard is added |

## 6. Frontend impact of the design

The design departs from nothing, so nothing changes in the frontend beyond the reviewed PR #13. That PR adds no mock module (the page reads the data module's existing `useRecord`), so there is no mock shape to compare with the API: the record is the one `useRecord` already returns and `Bike` already types.

## 7. Checked against what is on `main` (2026-11-19)

*Site data from the database* (PR #22) has merged, so the preconditions of §5 and the rows of §3 were read against the code, not only against the documents. Nothing in them changed; this section records what was found so the backend step does not look for work that is not there.

| What was read | Finding |
|---|---|
| `code/backend/internal/collections/handler.go`, `Record` | `404 NOT_FOUND` "That record does not exist." is written in exactly two places: `pgx.ErrNoRows` from the primary-key read, and an id the database cannot store (`cleanText` refuses it before any query, so no statement runs). A query that fails goes through `storeFailed`: `503` `UNAVAILABLE` when `unavailable(err)` (a connect-phase error whatever the SQLSTATE under it, SQLSTATE class `08`, `57P01`–`57P03`, a refused or lost connection), otherwise `500` `INTERNAL`; a deadline is `500`. No failure path is ever a `404`, which is what the page relies on |
| `code/frontend/src/data/index.ts`, `useRecord` | `loading` is derived from the key `[name, id]` on render, so the first render for an id, and the first render after the id changes, is `{ item: undefined, loading: true, error: null }` (DATA-007 AC-13, AC-14 are met). `404` settles `error: null`; a `200` that is not an object, any other status and a rejected request settle an `Error`. While the same key is asked again (the collection's epoch moved) the held answer stays, so a record in hand is never blanked by the data module either |
| `code/frontend/src/pages/Product.tsx` on PR #13 (`fix/bike-page-waits-for-its-record-ui`) | `if (!item) return loading \|\| error ? null : <Missing />;` reads exactly the three fields above; the page needs no other change and `Product.tsx` on `main` is untouched, so the PR merges over the data module cleanly |

One behaviour of the data module that is correct for this page and worth knowing when Test reads a trace: the module keeps a single held answer. Going from bike A to bike B and back to A *before B has answered* draws A's earlier answer at once while A is asked again — a record in hand is drawn. It is never a not-found view and never B's record under A's address, so it needs no handling here.

### 7.1 Second re-check (`main` at `760be06`, 2026-11-19)

Nothing in the code this story reads has moved since PR #22 merged (`76a0160`): every later commit on `main` is documentation (this design's PR #24, the index rebuilds). Re-read against it:

| What was read | Finding |
|---|---|
| `collections.Record` and its test `TestARecordIsAnsweredAsTheDatabaseWroteIt` | The id is bound exactly as sent (`Stance-0` is the argument `Stance-0`), and the statement is `WHERE t."id" = $1::text` (`sql_test.go`). An id with capitals therefore matches no row and is the `404` of the §3 row, not the record `stance-0` |
| `TestARecordThatIsNotThereIs404`, `TestAnIdTheDatabaseCannotHoldIsAnswered404WithoutAsking` | `no-such-bike` is `404 NOT_FOUND` "That record does not exist." with no `ETag`; an id with a NUL byte is the same `404` with no query run |
| The mock the task names (`code/frontend/src/lib/mock/bike-page-waits-for-its-record.ts`) | Does not exist. PR #13 changes `code/frontend/src/pages/Product.tsx` only, so there was no mock shape to judge or depart from; §6 stands |
| Open PR #25 (orders) and PR #26 (admin), both cut from `main` and not merged | Neither adds a route for `GET /v1/products/{id}`. PR #26 changes how the `collections` handlers are wired (they take the owner's guard, which is for writes) and moves the connection-failure predicate to `httpx.Unavailable`. Its description says the read answers are unchanged; that is a claim to confirm, not to assume, so after whichever of the two merges first, the cases that pin this page's call (TC-02 and TC-23 of `docs/site/test-cases/bike-page-waits-for-its-record.md`) are the regression check: `GET /v1/products/stance-0` is `200` and `GET /v1/products/no-such-bike` is `404` with **no session cookie** |

Result: the design above stands unchanged. No endpoint, field, code, schema or migration is added by this re-check either.

### 7.2 Third re-check (`main` at `074dc3b`, 2026-11-19)

The step-6 task was dispatched again. `main` has moved only by documentation since §7.1 (PR #27 and the index rebuild after it); no file under `code/` changed, so §7 and §7.1 still describe the code on `main`. This time the one claim §7.1 left open was read instead of repeated: whether the open admin backend (PR #26, head `d4b1e23`) leaves this page's call as it is.

| What was read | Finding |
|---|---|
| PR #26 diff of `code/backend/internal/collections/handler.go`, function `record` (formerly `Record`) | Only the non-`GET` branch changed: `PUT` and `DELETE` on the two writable collections go to the owner's guard, any other method is `405`. The `GET` path of a record is the same code, so `GET /v1/products/stance-0` is still `200` with no session and an id with no row is still `404` `NOT_FOUND` via `pgx.ErrNoRows`. The guard wraps the write handlers only; it is never applied to a read |
| PR #26 move of the connection-failure predicate | `unavailable(err)` became a one-line call to `httpx.Unavailable(err)` with the body moved unchanged, so the `503`/`500` split of §3 is the same. The read tests pin that name and still call it |
| PR #26 route table | A literal route (`/v1/sessions`, `/v1/me`, `DELETE /v1/sessions/current`) wins over `{collection}`; none of them shadows `/v1/products/{id}`. A `GET` on `/v1/sessions/current` is a `404` as an unknown collection, which is correct and unrelated to a bike |
| PR #25 (orders, head `c8a7a6a`) | Not re-read line by line here. Its description and the TL review recorded in the team's notes say it adds the one literal route `/v1/orders`, moves the same predicate into `httpx`, and changes no read answer. A literal `/v1/orders` cannot match `/v1/products/{id}` |

Nothing in this pass changes the contract, the ERD extension or the delivery table of §9. After whichever of PR #25 and PR #26 merges first, TC-02 and TC-23 remain the regression check for this page's call, as §7.1 says. Neither PR is a prerequisite of this story: PR #13 touches `Product.tsx` only and merges over `main` as it is.

### 7.3 Fourth re-check (`main` at `b6e6759`, 2026-11-19)

The step-6 task was dispatched once more. Since §7.2 (`main` at `074dc3b`) `main` has moved only by documentation (PR #28 and the index rebuild after it); no file under `code/` changed, and PR #25 (orders) and PR #26 (admin) are still open, so §7, §7.1 and §7.2 still describe the code this page runs against. The two reads that decide this story's correctness were repeated against the files themselves:

| What was read | Finding |
|---|---|
| `code/frontend/src/data/index.ts`, `useRecord` (lines 344–375) | With an id and no held answer for the key `[name, id]`, the hook returns `{ item: undefined, loading: true, error: null }` on the render itself, not from an effect, and does so again on the first render after the id changes (DATA-007 AC-13, AC-14). A `404` settles `item` undefined with `error` null (line 363); any other status, a `200` that is not an object, and a rejected request settle an `Error`. The same key asked again (the collection's epoch moved) keeps the held answer, so a record in hand is never blanked (story SC-7) |
| `code/frontend/src/pages/Product.tsx` on `main` | Still `const { item } = useRecord<Bike>("products", id)` and `if (!item) return <Missing />`. The defect of the plan item is therefore still open on `main`, and it is closed by PR #13 alone: one file, merges over the data module as it is |

Result: the design stands unchanged. No endpoint, field, error code, schema or migration is added by this pass, no mock module exists to judge (§6), and the delivery table of §9 is as written: **no backend branch and no backend pull request for this story; PR #13 is the only pull request to verify and merge.**

## 8. Migration and rollout

No schema step and no deploy step of its own: see `docs/architecture/erd/bike-page-waits-for-its-record.md` §4. Rolling the page change out before *Site data from the database* is harmless (it is invisible on the file-backed mock). Rolling it back is reverting one line of `Product.tsx`; the API is untouched either way.

## 9. Delivery: what this design asks of the steps after it

| Step of the story flow | What it means for this story |
|---|---|
| 7. Story BE and 8. BE review | **Nothing to build and nothing to review.** No backend, schema, migration or mock-swap change exists (§1, §4, `erd/bike-page-waits-for-its-record.md` §1), so no backend branch or pull request is to be opened: its diff would be empty. A re-run of step 7 that finds this is correct, not stuck |
| 9. Run local | Run `main` with the head of PR #13 applied; the API under it is the one `main` already has. The checks that need the database are the two of §7.1, plus the observation in story §5 (hold `GET /api/v1/products/<id>` for a few seconds and record the page's text from the start of navigation) |
| 10. Verify | Against that running story, SC-1 to SC-7 of the story; SC-8 (one file) is read from the PR's file list |
| 11–12. Merge | **PR #13 is the story's only pull request.** There is no backend PR to merge after it |

The one thing that could make a verifier report a defect against this story is the data module's first render (§5, first row). If Test sees the not-found view for one frame, the report goes to DATA-007; `useRecord` on `main` was read in §7 and answers `loading: true` on that first render.
