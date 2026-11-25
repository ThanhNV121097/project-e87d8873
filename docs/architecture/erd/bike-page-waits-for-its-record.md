# ERD extension — Bike page waits for its record

Story: `docs/site/stories/bike-page-waits-for-its-record.md` · Requirement: SITE-016 (`docs/site/SRS.md` §4.2) · Plan item: *Bike page waits for its record* (bug, P1)

## 1. Result

**This story adds no entity, no column, no index, no constraint and no migration.** The ERD is not extended.

The story changes one frontend file, `code/frontend/src/pages/Product.tsx`, so that the bike's place stays empty while the record is awaited or after the request failed, and the not-found view shows only after the shop has answered that there is no such bike (story §3, §4). The story states it itself: "Backend, schema and API: none" (story §4, last bullet).

## 2. What the page reads, and where it already lives

The page reads one bike by its id. That read is served by tables and an access pattern that merged designs already define; this story neither changes nor restates them:

| What the page needs | Where it is defined |
|---|---|
| One bike by id, 404 when no row has that id | `products` (primary key `id`), access pattern 5 of `docs/architecture/erd.md` §5 ("One record by id", index: primary key) |
| The bike `stance-0` exists and `no-such-bike` does not (story §7, "Data that must exist") | Migration slot `20261118000200_seed_collections` of *Site data from the database*; `stance-0` is one of the 14 seeded bikes and no seeded or owner-added row has the id `no-such-bike` |

The lookup is a primary-key read, so it needs no index of its own and has no new query to name.

One caution about the data this story's checks stand on, not a schema change: `no-such-bike` is absent because nothing seeds it, not because the table forbids it. The owner's "add a bike" write accepts any id matching `^[a-z0-9][a-z0-9-]{0,63}$` (`docs/architecture/services/_base.md` §3.4), which includes `no-such-bike`. A run that adds bikes must therefore not use that id, or must delete what it added before the 404 cases (SC-5) are read; the seed (`code/backend/migrations/20261118000200_seed_collections.up.sql`) inserts 14 bikes, among them `stance-0`, and no bike with the id `no-such-bike`.

## 3. Entities, columns, foreign keys, indexes

None.

## 4. Migration plan

| Aspect | Decision |
|---|---|
| Forward | No file. This story owns no migration slot; the slots `20261118000100`…`000400` of `docs/architecture/erd.md` §8 belong to *Site data from the database*, *Place an order and hold the bike* and *Shop admin: real sign-in and saved edits* |
| Backward | Nothing to reverse. Reverting the story is reverting the one-line change to `Product.tsx` |
| Safe on a populated table | Not applicable: no table is touched, no lock is taken, and no row is read or written by anything this story adds |

## 5. Ordering dependency (not a schema dependency)

The change is invisible until `useRecord` reports a real `loading`, which arrives with *Site data from the database* (DATA-007). Until the table `products` and the route of `docs/architecture/services/bike-page-waits-for-its-record.md` §2 exist, the page runs against the file-backed mock whose `loading` is the constant `false`. That ordering is a delivery dependency and is recorded in story §7; it adds nothing to the data model.
