# ERD extension — Place an order and hold the bike

Story: `docs/orders/stories/place-an-order-and-hold-the-bike.md` · Requirements: ORD-001 to ORD-008 (`docs/orders/SRS.md`) · Plan item: *Place an order and hold the bike* (feature, P1) · Migration slot: `20261118000300_orders` · Companion: `docs/architecture/services/place-an-order-and-hold-the-bike.md`

## 1. Result

**This story adds no entity, no column, no index and no constraint to the model.** `orders`, `order_lines` and `order_counters` — their columns, nullability, `CHECK`s, the one foreign key, the indexes and the number allocation — are already defined by the project-wide design (`docs/architecture/erd.md` §3.9 to §3.11, §3.13, §4, §5 patterns 9 to 11, §7, §8 row 3). Nothing of that is restated here. This story **builds** the three tables in migration slot `000300` and writes to them, and to `products.stock`, from one transaction.

There is no mock module to read for this story: PR #17 deleted `src/lib/mock/orders.ts`, and `src/lib/orders.ts` is the API client. The shape the frontend renders is its `Order` (the eleven keys of the answer). It was judged against the model: every key has a column to come from, and no column holds something the answer needs that the model lacks.

| Answer key | Where it comes from |
|---|---|
| `number` | `orders.number`, allocated from `order_counters` (statement S4) |
| `placed` | `orders.placed_at` — the clock instant of the order, written RFC 3339 with `+07:00` |
| `name`, `phone`, `delivery`, `address`, `payment` | `orders.*`, as stored (trimmed) |
| `lines[].id` / `.name` / `.price` / `.qty` | `order_lines.product_id` / `.name` / `.unit_price` / `.quantity` |
| `lines[].variant` | `order_lines.size`; the key is left out when the stored size is `''` |
| `lines[].image` | `products.image`, **read in statement S1 inside the transaction**; not stored on the line (ORD-001.2). The key is left out when the bike has no image |
| `subtotal`, `delivery_fee`, `total` | `orders.*` |
| *(not in the answer)* `note` | `orders.note`, stored and never returned |

What this file adds, because the project-wide design left it to the builder and this is the only gate where the transaction is read against the real tables:

| § | What |
|---|---|
| 2 | The seven statements of the transaction: what each reads or writes, its lock, and the index that serves it |
| 3 | Clarifications of the model. None changes a table; each pins how a statement is written, and says why |
| 4 | Build rules for migration 3 |
| 5 | The model checked against the real bikes and the shop record |
| 6 | The SQL the verification step runs (counts, the foreign-key check, the failure injection) |
| 7 | The migration plan: forward, backward, safety on a populated table, order of delivery |

## 2. What the story reads and writes

One transaction per accepted order, `READ COMMITTED` (the default; the row locks below do the serialising, so `SERIALIZABLE` would add retries and no safety). Every value is a bound parameter. Column names are quoted, as in migration 1.

```sql
-- S1  lock and read every distinct bike of the order, in one fixed order
SELECT "id", "name", "brand", "status", "stock", "price", "sizes", "sizes_out", "image"
FROM "products" WHERE "id" = ANY($1::text[]) ORDER BY "id" FOR UPDATE;

-- S2  the courier charge, only when delivery = 'courier' (a plain read: nothing writes "shop" at run time)
SELECT "delivery_fee" FROM "shop" WHERE "id" = 'shop';

-- S3  hold the bikes: ONE statement, the quantities of each distinct bike summed
UPDATE "products" AS p SET "stock" = p."stock" - h.qty
FROM unnest($1::text[], $2::bigint[]) AS h(id, qty)
WHERE p."id" = h.id AND p."status" = 'live' AND p."stock" >= h.qty;

-- S4  the order number: the upsert of erd.md §3.11, $1 = the Vietnam date of the clock instant (as ::date)

-- S5  the order
INSERT INTO "orders" ("number","placed_at","name","phone","delivery","address","payment","note","subtotal","delivery_fee","total")
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id";

-- S6  all lines in one statement; "position" is the 1-based place in the request
INSERT INTO "order_lines" ("order_id","position","product_id","name","size","quantity","unit_price")
SELECT $1, l.pos::smallint, l.product_id, l.name, l.size, l.qty::integer, l.price
FROM unnest($2::int[], $3::text[], $4::text[], $5::text[], $6::int[], $7::bigint[]) AS l(pos, product_id, name, size, qty, price);
```

The order of the statements is the contract's (`services.md` §3.10): S1, S2, then the decision in Go, then S3 (stock), S4 (number), S5 (order), S6 (lines), `COMMIT`. A refusal happens after S1/S2 and before S3, so a refused order writes nothing and spends no number; any error after S3 rolls back all of it, the counter included.

| Statement | Table, columns | Lock taken | Index that serves it |
|---|---|---|---|
| S1 | `products`: the nine columns | row lock (`FOR UPDATE`) on each distinct bike, taken in `id` order — ids are immutable (`erd.md` R1), so the order is the same for every transaction and two orders over overlapping bikes cannot deadlock | primary key, `= ANY` over at most 20 ids |
| S2 | `shop."delivery_fee"` | none | primary key |
| S3 | `products."stock"` (and `version`, `updated_at` through `trg_products_touch`) | already held from S1; no new lock | primary key. `RowsAffected` must equal the number of distinct bikes; anything else is `INTERNAL` (unreachable under the lock, and `ck_products_stock_nonneg` is the second guard) |
| S4 | `order_counters` | row lock on today's row, held to `COMMIT` — **last among the locks**, so the only lock an order can wait on after the bike rows is one every other order also takes after its own | primary key (`day`) |
| S5, S6 | `orders`, `order_lines` | none beyond the new rows | `uq_orders_number` (checked at insert), `uq_order_lines_order_id_position` |

Statements per order: S1, S3, S4, S5, S6, plus S2 for a courier order — **five or six, whatever the number of lines**, plus `BEGIN` and `COMMIT`. With a row lock on each bike this is what keeps an order of 20 lines inside ORD §6's 1 s at the 95th percentile (SC-36).

Concurrency, read against the tables: twenty orders for a bike with stock 1 queue on S1's lock for that bike. The first holds it through `COMMIT`; each of the nineteen then re-reads the row (`READ COMMITTED` returns the committed version after the wait) and sees `stock = 0`, so it is refused before S3 (SC-30). The pool has 10 connections (`internal/db`): a transaction waiting for a lock holds a connection, but the holder of the lock already has its own, so a full pool slows the queue and can never starve the holder. Do not enlarge it for this story.

An owner's save of the same bike (ADMIN-008) is one `UPDATE … WHERE id = $1 AND version = $2` on a single row and never waits for a second resource while holding its lock, so it cannot close a cycle with an order. An owner's delete of a bike that an order has locked waits for the order; after the order commits, the delete removes the row and the order's lines stay (`order_lines.product_id` is not a foreign key, `erd.md` §3.10).

## 3. Clarifications of the model

None of these changes a table, a column, a constraint or an index. Each pins a choice the model left open.

| # | Where the model is open | Decision | Rejected alternative, and why |
|---|---|---|---|
| C1 | `erd.md` §3.9 says `placed_at` is `now()`, "the instant the number's day is read from"; §3.11 computes the day as `(now() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date`. The story's SC-32 and SC-33 need a clock the test sets (story §6) | The application reads its injected clock **once**, after `BEGIN`, and passes that one instant as `placed_at` (S5) and, converted to Vietnam time, as the day of S4. The column default `now()` stays (it serves a row inserted by hand) and is never relied on. `order_counters.updated_at` keeps the `now()` of the upsert: it is bookkeeping, not a customer-visible time | Keeping `now()` and `AT TIME ZONE` in SQL: the instant cannot be set from a unit test, so "18:00 UTC on 16 Nov is `DEO-261117-`" and the midnight roll-over could only be shown at the real midnight |
| C2 | The zone of the day is named `Asia/Ho_Chi_Minh` | In Go the day is `clock.In(time.FixedZone("ICT", 7*3600))`. Vietnam has had no daylight saving since 1975, so a fixed `+07:00` gives the same date as the named zone for every date this shop can have, and the rendered `placed` carries `+07:00` as the contract requires. It needs no tzdata in the unit-test image (the runtime image has it; the CI test image is not promised to) | `time.LoadLocation("Asia/Ho_Chi_Minh")`: identical result, one more way for a test to fail in an image without tzdata |
| C3 | `erd.md` §5 pattern 9 writes the hold as one `UPDATE` "per bike" | One statement for all bikes (S3, `unnest`), and one for all lines (S6). The behaviour is the same; the round trips do not grow with the line count | A statement per line: correct, but 20 lines are about 60 round trips and SC-36's second is then spent on the network, not the database |
| C4 | `erd.md` §3.10 says `image` is "read at response time, not stored" | It is read in S1, inside the transaction, and the answer is built from the rows S1 read and the values S3 to S6 wrote. **Nothing is read after `COMMIT`.** The `image` in the answer is therefore the one the bike had when it was ordered, even if the owner edits it a millisecond later | A read after commit: a second connection, a second failure mode after the order is stored, and an answer that could disagree with the order |
| C5 | Story SC-31 describes the simulated failure as "after the order row is written and before the stock is lowered"; the contract's order lowers stock first (S3) and writes the order after (S5, S6) | The contract's order stands. What SC-31 proves is that no step's failure leaves a half-order, and that does not depend on which step fails: in the running stack the failure is injected with a trigger on `order_lines` (§6), which fails **after** S3, S4 and S5 — the worst place, with the stock lowered and the counter and the order row written; in unit tests each step of the transaction sits behind a small interface so a fake fails any one of them (services extension §6). No code in the product exists only for a test | Moving S3 after S6 to match the sentence literally: it changes a contract paragraph and an arranged test case (`set_db_state` trigger on `order_lines`) for no gain, because the observable property is all-or-nothing and both orders give it |
| C6 | `erd.md` §3.9: `address` is `NOT NULL DEFAULT ''` and "required unless `pickup`"; `order_lines.size` is `''` "for a bike with no sizes" | A `pickup` order that sends an address stores it, trimmed, within the 300-character limit (an address is *not needed* for a pickup, not *forbidden*; `ck_orders_address` already allows it). A line for a bike with no sizes stores `size = ''` whatever variant was sent, and the answer leaves `variant` out (the contract's "ignored") | Refusing an address on `pickup`, or blanking it: a rule no requirement states, and SC-4 wants the stored address to equal the answered one |
| C7 | `erd.md` §3.9 `ck_orders_name` uses `btrim`, which trims only the space character | The application trims with Unicode white space (`strings.TrimSpace`) before it checks or stores; the stored value is already trimmed, so the `CHECK` never disagrees with the validator. A text with a NUL character (U+0000) is refused as a value error before any statement, because Postgres cannot store it and would answer a 500 | Relying on the `CHECK`: a name of a tab or a no-break space would pass `btrim` and be stored as a blank name |

## 4. Build rules for migration 3 (`20261118000300_orders`)

1. **Quote every column name** in the DDL and in every statement, as migration 1 does (`"position"`, `"number"`, `"size"`, `"day"`, `"note"` are legal bare in some places and not in others; one rule is cheaper than remembering).
2. **Constraint and index names are exactly those of `erd.md` §3.9 to §3.11** (`ck_orders_cod_local`, `uq_orders_number`, `fk_order_lines_order_id`, `uq_order_lines_order_id_position`, `ck_order_counters_last_seq_range`, …). The Go code never names one and no answer carries one: a failed statement is a generic 500.
3. **`number` is `UNIQUE` as a named constraint** (`CONSTRAINT uq_orders_number UNIQUE ("number")`) so the index has the name the model gives it.
4. **The number's `CHECK` is a plain single-quoted literal**, `'^DEO-\d{6}-\d{4}$'`: `standard_conforming_strings` is on, so `\d` is a backslash and a `d`. No `E'…'` and no dollar quoting anywhere in the file; the runner's statement splitter reads a single-quoted string as one unit.
5. **`order_lines."product_id"` has no foreign key.** Neither does anything else in the file point at `products`. The file therefore depends on no other table, and SC-5's "deleting a bike leaves its lines unchanged" is true by construction (§6 checks it).
6. **`fk_order_lines_order_id`** is `REFERENCES "orders"("id") ON DELETE CASCADE ON UPDATE RESTRICT`, as `erd.md` §3.10.
7. **Defaults:** `"id" uuid DEFAULT gen_random_uuid()` (built in since Postgres 13, no extension), `"placed_at" timestamptz NOT NULL DEFAULT now()`, `"address"` and `"note"` `NOT NULL DEFAULT ''`, `"size"` `NOT NULL DEFAULT ''`. `"quantity"` is `integer`, `"position"` is `smallint`, money is `bigint`.
8. **No `CONCURRENTLY` anywhere in the file, in any case** (the runner would move the whole file out of its transaction on seeing the word, `docs/architecture/erd/site-data-from-the-database.md` §5 item 5), and no data statement: this file only creates tables.
9. **The down file** is exactly three statements, in this order: `DROP TABLE "order_lines";`, `DROP TABLE "orders";`, `DROP TABLE "order_counters";`. It mentions no other table and creates or drops no function or trigger (this story creates none).
10. **`migrations_test.go`** gains a test that reads the pair: the up file creates exactly `orders`, `order_lines` and `order_counters`; the down file drops exactly those; neither has a `REFERENCES` to `products` or a statement on it, and neither contains `concurrently`. The test checks text, not Postgres: CI has no database.

## 5. The model checked against the real data

The four bikes the story's cases use and the `shop` record were read in full from `code/frontend/src/data/` on `main`; the other ten bikes were not read for this story (the seed's own pre-flight, `docs/architecture/erd/site-data-from-the-database.md` §3, covers every record against the same constraints).

| Check | What the data holds | Result |
|---|---|---|
| `talon-2` (SC-1 to SC-4, SC-7 to SC-9, SC-11, SC-23 to SC-25, SC-28, SC-30) | Giant, "Talon 2", `live`, price 18990000, stock 9, sizes XS S M L XL, `sizes_out` `[]`, image `/images/ref/talon-2.jpg` | The answer's line is `Giant Talon 2`, 18990000, `/images/ref/talon-2.jpg` — exactly SC-1. Five sizes, so a 20-line order of one bike (SC-13) is four lines per size and needs its stock raised first (the test cases do) |
| `stance-0` (SC-10, SC-26, SC-28) | Giant, "Stance 0", `live`, price 66490000, stock 7, sizes XS to XL, none out | Orderable in any size |
| `anthem-advanced-sl-0` (SC-29) | `live`, stock 1, sizes S M L XL, **`sizes_out` S and XL** | Only **M and L** can be ordered. An order in S or XL is a `409` "not available in size", so SC-29's "last bike" must be ordered in M or L. After it, stock is 0, `sizes_out` is unchanged (ORD-005.4) and the page reads "Out of stock" because the page reads `stock` |
| `pique-advanced-sl-se` (SC-23) | Liv, "Pique Advanced SL SE", `live`, stock 0, **every size** in `sizes_out` | The stock check comes before the size check (`services.md` §3.10), so the sentence is "Liv Pique Advanced SL SE is out of stock.", not a size sentence |
| `shop` (SC-8, SC-9) | `delivery_fee` 350000; `stock_rule` "Each bike has a number in stock, kept by the shop. An order holds the bike until the confirming call; a size at zero shows as out of stock." | A courier order's `delivery_fee` is 350000; the rule is what the hold implements |
| `ck_order_lines_unit_price` (`> 0`) | Every live bike has a price above 0 (`ck_products_live_has_price`; the seed pre-flight read every bike) | A line can never carry a 0 |
| Money cannot overflow | The largest price the owner can save is 9,007,199,254,740,991 (`services.md` §3.11); × quantity 10 × 20 lines = 1.8 × 10^18, below the `bigint` limit of 9.2 × 10^18 | `subtotal` and `total` are safe in `bigint` and in Go's `int64` |
| Daily counter | 9,999 orders a day (`ck_order_counters_last_seq_range`) | Far above a bike shop's volume; the 10,000th would roll back and answer a 500 (`erd.md` §3.11) |
| **The seeded stock figures are the team's own** | `_invented` on `talon-2`, `anthem-advanced-sl-0`, `pique-advanced-sl-se` and `stance-0` lists `stock` | From the first real order the hold lowers a number that was never counted on a shelf. The owner must set the true stock in the management area before the shop is announced (§8) |

## 6. SQL the verification step runs

CI has no database, so every criterion that needs Postgres is proved against the running stack with these reads (read-only unless stated).

**An accepted order is stored as answered (SC-4).**

```sql
SELECT o.number, o.name, o.phone, o.delivery, o.address, o.payment, o.note,
       o.subtotal, o.delivery_fee, o.total,
       l.position, l.product_id, l.name AS line_name, l.size, l.quantity, l.unit_price
FROM orders o JOIN order_lines l ON l.order_id = o.id
WHERE o.number = $1 ORDER BY l.position;
```

**No foreign key from a line to a bike (SC-5).** The catalogue has one row, to `orders`; the delete is rolled back, so nothing is lost.

```sql
SELECT conname, confrelid::regclass FROM pg_constraint
WHERE conrelid = 'order_lines'::regclass AND contype = 'f';      -- one row: fk_order_lines_order_id, orders
BEGIN; DELETE FROM products WHERE id = 'talon-2';
SELECT count(*) FROM order_lines WHERE product_id = 'talon-2';   -- unchanged by the delete
ROLLBACK;
```

**A refused order changed nothing (SC-21, SC-26, SC-31).** Read before and after, compare:

```sql
SELECT count(*) FROM orders;  SELECT count(*) FROM order_lines;
SELECT id, stock FROM products ORDER BY id;  SELECT day, last_seq FROM order_counters ORDER BY day;
```

**A failure injected after the stock was lowered (SC-31).** No code in the product is involved: a trigger on `order_lines` raises on the insert of S6, which runs after S3 (stock), S4 (counter) and S5 (order row). The request must answer `500` with the generic sentence and the reads above must be unchanged.

```sql
CREATE FUNCTION fail_line() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected'; END $$;
CREATE TRIGGER trg_fail_line BEFORE INSERT ON order_lines FOR EACH ROW EXECUTE FUNCTION fail_line();
-- afterwards, always:
DROP TRIGGER trg_fail_line ON order_lines;  DROP FUNCTION fail_line();
```

**Invariants that must hold after any run, concurrent or not.** Each returns no rows or a zero:

```sql
SELECT count(*) FROM products WHERE stock < 0;                                          -- 0 (SC-30)
SELECT number FROM orders o WHERE subtotal <> (SELECT sum(unit_price * quantity) FROM order_lines WHERE order_id = o.id);
SELECT number FROM orders o WHERE NOT EXISTS (SELECT 1 FROM order_lines l WHERE l.order_id = o.id);   -- no order without a line
SELECT day FROM order_counters c                                                         -- no gap, no duplicate: the counter equals the orders of its day
WHERE last_seq <> (SELECT count(*) FROM orders WHERE number LIKE 'DEO-' || to_char(c.day, 'YYMMDD') || '-%');
```

The last query holds only while no order has been purged by hand (`erd.md` §9, question 1); the verification step runs it on a database it started from the seed.

## 7. Migration plan

| Aspect | Decision |
|---|---|
| Forward | One file, `20261118000300_orders.up.sql`: the three tables of `erd.md` §3.9 to §3.11 with their constraints, the one foreign key and the two indexes. Applied by the server on boot in filename order — after `000100` and `000200`, before `000400` — in one transaction with its `schema_migrations` row, under the advisory lock |
| Backward | `20261118000300_orders.down.sql` (§4 item 9). **For development only**: it destroys every order, and the runner never applies a down file or deletes the `schema_migrations` row, so a manual rollback deletes that row by hand. Rolling back a shop that has taken orders is a restore from a backup, not a migration. Rolling the *binary* back while the tables stay is harmless: an older binary has no route that writes them |
| Safe on a populated table | Yes. Every object is new; `products` and `shop` are not touched, so no lock is taken on any table that holds data, and the file has no foreign key to one. A shop that already holds bikes, owner edits and sessions loses and changes nothing |
| Failure | A statement that fails rolls back the whole file and the server **does not start**, with the migration's own message. Nothing is half-applied; the next start retries the same file |
| Time and size | Three `CREATE TABLE`s and two indexes on empty tables: milliseconds |
| Order of delivery | This story's backend reads `products` and `shop`, which the *Site data from the database* backend creates and fills (`000100`, `000200`). It merges **after** that backend; if it were deployed before it, `POST /v1/orders` would answer a `500` (the tables are missing) and nothing else would break. Within the story the frontend PR (#17) merges first and the backend PR second, as the delivery flow says; the checkout shows its generic failure sentence in between, and nothing is deployed between the two merges |
| Not in this story's migrations | `sessions` (`000400`), which belongs to *Shop admin: real sign-in and saved edits* |

## 8. Open questions

| # | Question | Owner | Blocking |
|---|---|---|---|
| 1 | **The seeded stock figures were invented** (§5, last row). The first real order lowers a number nobody counted. Proposed default: the owner sets the true stock of every bike in the management area before announcing the shop; until then an order can be refused as "out of stock" for a bike the shop has, or accepted for one it lacks | Stakeholder | No |
| 2 | How long orders are kept stays open: `erd.md` §9, question 1. This story stores the personal data that question is about and changes nothing in it | PM / stakeholder | No |
