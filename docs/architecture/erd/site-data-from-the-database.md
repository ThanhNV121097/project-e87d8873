# ERD extension — Site data from the database

Story: `docs/data/stories/site-data-from-the-database.md` · Requirements: DATA-001 to DATA-008 (`docs/data/SRS.md`) · Plan item: *Site data from the database* (feature, P1) · Migration slots: `20261118000100_collections`, `20261118000200_seed_collections`

## 1. Result

**This story adds no entity, no column, no index and no constraint to the model.** The eight collection tables, the three helper functions, the version trigger, the closed sets and the access patterns are already defined by the project-wide design (`docs/architecture/erd.md` §3.0 to §3.8, §3.14, §4, §5) and the two migration slots by §8 rows 1 and 2. This story **builds** them. Nothing of that is restated here.

What this file adds, because the project-wide design left it to the builder and this is the only gate where anybody reads the real data against it:

| § | What |
|---|---|
| 2 | Which request each table serves and which plan it gets |
| 3 | The model checked against the 14 + 4 + 12 + 7 + 4 + 6 + 8 + 1 real records (the seed's pre-flight) |
| 4, 5 | Build rules for the two migration files that the model leaves open (quoting, constraint names, how the seed is written and checked, one trap in the runner) |
| 6 | Four clarifications of the model. None changes a table; each changes how a statement is written, and says why |
| 7 | The migration plan: forward, backward, safety on a populated table, order of delivery |

## 2. What the story reads, and what serves it

The story only reads: no endpoint of it inserts, updates or deletes (the owner's writes are the admin story's, the stock hold is the orders story's). The `version` column is read once, for the `ETag`.

| Request (`services.md` §3.2, §3.3) | Table and columns | Plan | Index (`erd.md` §3) |
|---|---|---|---|
| `GET /v1/{collection}` with no condition | every column of the collection's field list, plus `seq` | scan of ≤ 14 rows, ordered by `seq`, paged | `uq_<table>_seq` (the tie-break of every ordering) |
| … with `<field>=v` on `category`, `bike`, `type` or `id` | `products.category`, `reviews.bike`, `gallery.type`, `<t>.id` | `col = ANY (<folded values>)` (clarification C1) | `idx_products_category`, `idx_reviews_bike`, `idx_gallery_type`, the primary keys |
| … with any other condition, `q`, `search`, `sort` | the named columns | sequential scan of the whole table (§3.2: no index on `status`, `brand`, `wheel`, `price`, `stock`; at ≤ 500 bikes it is the fastest plan, `erd.md` §6) | none |
| `GET /v1/{collection}/{id}` | one row | primary-key lookup (pattern 5); `version` read for `products` and `categories` only | primary key |

A bike's Reviews tab (`reviews?bike=<id>`, DATA-008 AC-11) and a riding type's bikes (`products?category=<id>`) are the two filters an index exists for.

## 3. The model checked against the real data

The records were read from `code/frontend/src/data/*.json` on `main`. Every reviews, gallery, team, services, faqs, shop and categories record was read in full; of the products, the first twelve in full and the last two (`tempt-e-plus-1`, `embolden-e-plus-0`) by their ids only. The seed migration is the check for those two: a record that breaks a `CHECK` or a reference stops the migration and rolls it back (§5).

| Check | What the data holds | Result |
|---|---|---|
| Counts (SC-1) | products 14, categories 4, reviews 12, gallery 7, team 4, services 6, faqs 8, shop 1 | as `erd.md` §8 row 2 |
| `ck_*_id_format` `^[a-z0-9][a-z0-9-]{0,63}$` | every id read: `talon-e-plus-1`, `first-bike`, `r-thao`, `pass-climb`, `shop`, `e-mtb`, … | passes |
| `ck_products_brand`, `_wheel`, `_status` | brand `Giant` or `Liv`; wheel `29` or `mixed`; status `live` on every bike | passes |
| `ck_products_live_has_price` | every live bike has a price above 0 | passes |
| `ck_reviews_date_iso` | all 12 dates are `YYYY-MM-DD` | passes |
| Foreign keys | every `reviews.bike` is a bike (`embolden-e-plus-0` is the last one); every `products.category` and `gallery.type` is one of `hardtail`, `cross-country`, `trail`, `e-mtb` | passes, provided `categories` is inserted first and `products` before `reviews` (§5) |
| R3, `NULL` = no key | no record in any file holds `null`; an absent field is simply a missing key | lossless |
| `was` (SC-4) | exactly two bikes (`talon-29-1`, `anthem-x-advanced-sl-3`) | every other bike has `was` NULL |
| `month` | only `trance-x-1` | |
| `stock` | `0` on two bikes (`pique-advanced-sl-se` and one more), 1 to 9 on the rest, so `stock_from=1` is 12 (SC-22) | `0` is a stored value, not NULL |
| `deal` | `true` on `talon-2`, `stance-0`, `talon-e-plus-1`; an explicit `false` on every other bike read | `deal=false` matches the stored `false`; Home's six picks rely on it (SC-76) |
| `numeric` exactness | `rating` is `4.8` on some bikes and the whole number `5` on others (`anthem-advanced-sl-0`, `pique-advanced-sl-se`, and reviews); `year` 2025 to 2027 | `numeric` keeps `5` as `5` and `4.8` as `4.8`; the seed writes each number as it is written in the file |
| Lists (SC-3) | `sizes_out` is `[]` on `talon-2`; `_invented` is a list of field names on most records | `'{}'` and `[]` stay different from NULL |
| Keys outside the model | none seen; every key is a `schema.json` key, `id`, `_source` or `_invented` | the seed generator must still fail on a key with no column (§5), because a typo would otherwise drop data silently |
| Text that breaks a careless SQL writer | apostrophes (`maker's`, `women's`, `Liv's`, `Giant's`, `frame's`), a dollar sign (`A$1,099 RRP`), a percent sign (`90% built`), `—`, `×`, `·`, `–`, Vietnamese letters (`Thảo`, `Bảo`, `Vân`) | handled by §5's literal rule |
| Words in the data that the runner reacts to | none of the eight files contains the word `concurrently` | see the trap in §5 |

## 4. Build rules for migration 1 (`20261118000100_collections`)

1. **Quote every column name** in the DDL, and in every statement the Go code builds. `date`, `time`, `text`, `type`, `year`, `month`, `role`, `line`, `name`, `a` and `q` are legal bare in most places and keywords or type names in others; one rule for all of them is cheaper than remembering which.
2. **`seq`** is `bigint GENERATED ALWAYS AS IDENTITY` with `CONSTRAINT uq_<table>_seq UNIQUE (seq)` so the index has the name `erd.md` gives it. `ALWAYS` means no statement can supply a `seq`; the seed relies on that (§5).
3. **Constraint and index names are exactly those of `erd.md` §3.1 to §3.8** (`ck_products_live_has_price`, `fk_reviews_bike`, `idx_gallery_type`, …). The Go code never names one, and no answer ever carries one (a failed statement is a generic 500), so a name is for the person reading the migration's own error message (SC-9).
4. **`version` and the trigger** exist on `products` and `categories` only. `version bigint NOT NULL DEFAULT 1`; `touch_record()` is `BEFORE UPDATE … FOR EACH ROW`. The seed inserts, so every seeded row has `version = 1` and its `ETag` is `"1"`.
5. **The three functions are created in the order `fold_text`, `natural_key`, `touch_record`** (`natural_key` calls `fold_text`; a SQL-language function body is checked when it is created). Bodies are those of `erd.md` §3.14, with two facts the builder must keep: `normalize()` needs a UTF-8 database (the runtime's and the compose profile's both are; a failure says so at the first call, not at creation), and `lpad(…, 40, '0')` **truncates** a digit run longer than 40, so two names that agree in their first 40 digits of a longer run compare equal there (accepted: no name has such a run, and `seq` then breaks the tie).
6. **No `CREATE INDEX CONCURRENTLY`** in this file: the tables are new, plain `CREATE INDEX` is fine, and a concurrent index would move the whole file outside a transaction (the runner's rule), which breaks "a migration that cannot complete leaves the database as it was" (SC-9).
7. **The down file** drops, in this order, `reviews`, `gallery`, `products`, `categories`, `team`, `services`, `faqs`, `shop`, then `natural_key(text)`, `fold_text(text)`, `touch_record()`. Triggers go with their tables.

## 5. Build rules for migration 2 (`20261118000200_seed_collections`)

The seed is generated from the JSON by a throwaway script that is not committed (it is not product code and the JSON is deleted in the same story). What the generator must do:

1. **Source and freshness.** Read the eight files from the tip of `main` **at the time the backend PR is opened, and again before it merges**. The owner edits these files in the Studio; a record changed between the two reads is a record lost (DATA-001: "none may be lost"). The second read is a diff of `git log -- code/frontend/src/data/` against the commit the seed was generated from; any change means regenerate.
2. **One `INSERT` per record**, in the order `categories`, `products`, `reviews`, `gallery`, `team`, `services`, `faqs`, `shop`, and inside each table in the file's order. An explicit column list; **never `seq`**. `ON CONFLICT (id) DO NOTHING` on every statement (a manual second run changes nothing and, importantly, **never overwrites an edit the owner has since made** to a seeded row, which `DO UPDATE` would).
3. **Literals.** Text in single quotes with every `'` doubled; no `E'…'` escapes and no dollar quoting (the runner's statement splitter reads a single-quoted string as one unit, so `$1` in `A$1,099` and `--` inside text are safe; a hand-written dollar tag is an avoidable risk). Numbers exactly as the file writes them (`4.8`, `5`, `18990000`), never through a float. Booleans `true`/`false`. A list is `ARRAY['XS','S']::text[]`; an empty list is `ARRAY[]::text[]`, never omitted and never NULL. An absent key is a column left out of the list (so it is NULL). The element order of every list is the file's.
4. **Fail on surprises.** The generator stops on a key that has no column, on a value of the wrong JSON type for its column, on a `null`, and on a duplicate id within a file. Nothing is skipped or coerced.
5. **The trap in the runner.** `internal/migrate` decides whether a file may run inside a transaction by looking for the word `CONCURRENTLY` anywhere in a statement's text, strings included (`needsNoTransaction`). A seed text that contained it would run the whole file outside a transaction, statement by statement, and a failure half-way would leave half the shop stored. None of the eight files has the word today. `migrations_test.go` gains one test: no `.up.sql` that holds a table or data statement contains `concurrently`, in any case. A later seed change that adds the word then fails in CI instead of in production.
6. **Atomicity (SC-9).** The runner applies the file in one transaction with its `schema_migrations` row. A reference that does not exist stops it with Postgres's own message (`violates foreign key constraint "fk_products_category"`, `Key (category)=(…) is not present in table "categories"`), and nothing is stored.
7. **The down file** deletes the seeded ids, as an `IN (…)` list per table, in the order `reviews`, `gallery`, `products`, `categories`, `team`, `services`, `faqs`, `shop`. It fails by `RESTRICT` if an owner-added bike still uses a seeded riding type; that is intended (`erd.md` §8 row 2).
8. **Proof that nothing was lost** is the running-stack case, not a unit test: each record returned by `GET /v1/<collection>/<id>` equals the record of the deleted file (taken from git history) key for key (SC-2). The unit tests cannot see it, because CI has no database and the files are gone.

## 6. Clarifications of the model

None of these changes a table, a column, a constraint or an index. Each pins a choice the model left open, so the builder and the reviewer read the same thing.

| # | Where the model is open | Decision | Rejected alternative, and why |
|---|---|---|---|
| C1 | `erd.md` §3.2 says `idx_products_category` serves the shop's `category=` filter, while `services.md` §3.2 says text equality ignores capitals and accents (`fold_text` on the column) | Equality on **`id` and on every `ref:*` column** folds the **parameter side only**: `col = ANY (ARRAY(SELECT fold_text(v) FROM unnest($n::text[]) AS v))`. This is the same answer as folding both sides, because a stored `id` or reference is already folded by construction (`ck_*_id_format` allows only `a-z`, `0-9`, `-`, and a reference must equal a referenced id). On every other text column both sides are folded, as the contract says. It makes the index claim true and costs nothing | Folding the column everywhere: correct, but `fold_text(category)` cannot use `idx_products_category`, so the index the model draws for this filter would serve only the delete check |
| C2 | `erd.md` §5 pattern 3 writes the search as `fold_text(col) LIKE '%' \|\| fold_text($q) \|\| '%'` with `%`, `_`, `\` escaped | The search is `strpos(fold_text(col), fold_text($n::text)) > 0`, one placeholder reused for every searched column. `strpos` has no wildcards, so `%`, `_` and `\` mean themselves with **nothing to escape**; the observable contract (DATA-004 AC-7, SC-38) is identical | `LIKE … ESCAPE '\'`: needs an escaping function whose one mistake turns a visitor's `100%` into "every bike" (SC-38 exists because that mistake is easy), and gives no speed at this size |
| C3 | `erd.md` §2.7 of the services contract says a record is "one row serialised by key" without naming how | The answer's keys come from an **explicit column list per collection** (the same list the query validation uses), projected in a subquery, then `to_jsonb(row)` and `jsonb_strip_nulls`. `seq` is carried for ordering and removed with `- 'seq'`; `version` is read as its own column for the `ETag` and never selected into the record | `to_jsonb(table_row) - 'seq' - 'version' - 'created_at' - 'updated_at'`: a deny-list, so the next column anyone adds to a table (an internal flag, a timestamp) leaks into every public answer until somebody remembers to extend the list. An allow-list cannot leak |
| C4 | Number encoding in the answer | `to_jsonb(numeric)` and `to_jsonb(bigint)` write JSON numbers with their stored digits; the handler passes the `jsonb` text through as raw bytes and **never decodes it into Go floats**. A rating of `4.8`, a stock of `0` and a price of `322990000` arrive as written | Decoding into `map[string]any` and re-encoding: every number becomes a `float64`; harmless at these values, wrong for a price near 2^53 and for `5` becoming `5` but `5.0` becoming `5` (DATA-001 AC-2 compares key for key) |

## 7. Migration plan

| Aspect | Decision |
|---|---|
| Forward | Two files, applied by the server on boot in filename order, before it listens: `20261118000100_collections.up.sql` (schema), then `20261118000200_seed_collections.up.sql` (data). Each runs in one transaction with its `schema_migrations` row, under the advisory lock, so two starting containers apply each file once. The slots are this story's alone; the orders (`…000300`) and sessions (`…000400`) stories are stamped later and never collide |
| Backward | The two `.down.sql` files of §4 item 7 and §5 item 7, **for development**: the runner never applies a down file and does not delete the `schema_migrations` row, so a manual rollback deletes that row by hand. Rolling back a shop that holds owner edits or orders is a restore from a backup, not a migration |
| Safe on a populated table | Migration 1: yes. Every object is new and the only earlier file is the empty baseline; plain `CREATE TABLE` and `CREATE INDEX` take no lock on anything that exists. Migration 2: yes. It only inserts, with `ON CONFLICT (id) DO NOTHING`, so applying it by hand to a shop whose owner has edited, added or deleted records changes none of them (a deleted seeded record would be put back; that is why the file is never re-run by the runner and a manual re-run is a decision). A skipped duplicate may leave a gap in `seq`; a gap changes no ordering |
| Failure | A statement that fails rolls back its whole file and the server **does not start**: it exits with the migration's own message (SC-9). Nothing is half-applied, and the next start retries the same file |
| Time and size | Migration 2 is 14 + 4 + 12 + 7 + 4 + 6 + 8 + 1 = 56 single-row inserts: milliseconds. Nothing is batched because nothing is large |
| Order of delivery | The frontend PR deletes the eight JSON files and so cannot serve a real site without this migration and the read API; the backend PR cannot show anything until the frontend module calls it. They merge back to back (steps 11 and 12 of the delivery flow) and the deployment waits for both. The orders story depends on `products.stock` from this migration; the admin story on the tables, `version` and the `touch_record` trigger |
| Not in this story's migrations | `orders`, `order_lines`, `order_counters` (slot `000300`) and `sessions` (`000400`): other stories |

## 8. Open questions

None. The three open questions of `docs/data/SRS.md` §7 keep their proposed defaults (the story's §8.6), and none touches the model.
