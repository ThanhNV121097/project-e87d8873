package orders

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
)

// rollbackWait bounds the ROLLBACK, which runs on a context of its own so that
// a request the visitor abandoned still releases its row locks at once.
const rollbackWait = 2 * time.Second

// DB is the part of the connection pool an order needs: a transaction.
// server.DB has the same method, so the pool satisfies both.
type DB interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// PlacedLine is a line of the answer: the bike's id, its name as the bag shows
// it, the unit price, its photograph when it has one, the size when the line
// has one, and the count.
type PlacedLine struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Price   int64  `json:"price"`
	Image   string `json:"image,omitempty"`
	Variant string `json:"variant,omitempty"`
	Qty     int64  `json:"qty"`
}

// Placed is the answer to an accepted order: exactly these eleven keys. The
// note is stored and never returned.
type Placed struct {
	Number      string       `json:"number"`
	Placed      string       `json:"placed"`
	Name        string       `json:"name"`
	Phone       string       `json:"phone"`
	Delivery    string       `json:"delivery"`
	Address     string       `json:"address"`
	Payment     string       `json:"payment"`
	Lines       []PlacedLine `json:"lines"`
	Subtotal    int64        `json:"subtotal"`
	DeliveryFee int64        `json:"delivery_fee"`
	Total       int64        `json:"total"`
}

// Store places an order: one transaction, all or nothing. The clock is read
// once, inside the transaction, and that one instant is the time the order was
// placed and the day of its number. A refusal comes back as a *Refusal; any
// other error is a failure of the store.
type Store interface {
	Place(ctx context.Context, req Request, clock func() time.Time) (Placed, error)
}

// OrderRow is the order as it is inserted.
type OrderRow struct {
	Number      string
	PlacedAt    time.Time
	Name        string
	Phone       string
	Delivery    string
	Address     string
	Payment     string
	Note        string
	Subtotal    int64
	DeliveryFee int64
	Total       int64
}

// txn is the steps of one transaction, in the order they run. A fake can fail
// any of them.
type txn interface {
	// LockBikes locks and reads the bikes, in id order.
	LockBikes(ctx context.Context, ids []string) (map[string]Bike, error)
	// CourierFee reads the shop's courier charge; none recorded is 0.
	CourierFee(ctx context.Context) (int64, error)
	// HoldStock lowers each bike's stock by its quantity and returns how many
	// bikes it changed.
	HoldStock(ctx context.Context, ids []string, qty []int64) (int64, error)
	// NextSeq issues the next counter of the day ("2006-01-02").
	NextSeq(ctx context.Context, day string) (int, error)
	InsertOrder(ctx context.Context, o OrderRow) (orderID string, err error)
	InsertLines(ctx context.Context, orderID string, lines []PricedLine) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type store struct {
	begin func(ctx context.Context) (txn, error)
}

// NewStore returns the Postgres store.
func NewStore(db DB) Store {
	return &store{begin: func(ctx context.Context) (txn, error) {
		tx, err := db.Begin(ctx)
		if err != nil {
			return nil, err
		}
		return pgTxn{tx: tx}, nil
	}}
}

// Place runs the contract's transaction: lock the bikes, read the courier
// charge, decide, hold the stock, number, store the order and its lines,
// commit. A refusal happens before anything is written; any error rolls
// everything back, the counter included.
func (s *store) Place(ctx context.Context, req Request, clock func() time.Time) (Placed, error) {
	tx, err := s.begin(ctx)
	if err != nil {
		return Placed{}, fmt.Errorf("begin: %w", err)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		rollback, cancel := context.WithTimeout(context.WithoutCancel(ctx), rollbackWait)
		defer cancel()
		if err := tx.Rollback(rollback); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Warn("order rollback failed", "error", err)
		}
	}()

	instant := clock()
	day := Day(instant)

	wanted := map[string]int64{}
	for _, l := range req.Lines {
		wanted[l.ID] += l.Qty
	}
	ids := slices.Sorted(maps.Keys(wanted))

	bikes, err := tx.LockBikes(ctx, ids)
	if err != nil {
		return Placed{}, fmt.Errorf("lock the bikes: %w", err)
	}
	var courierFee int64
	if req.Delivery == deliveryCourier {
		if courierFee, err = tx.CourierFee(ctx); err != nil {
			return Placed{}, fmt.Errorf("read the courier charge: %w", err)
		}
	}
	priced, err := Decide(req, bikes, courierFee)
	if err != nil {
		return Placed{}, err
	}

	qty := make([]int64, len(ids))
	for i, id := range ids {
		qty[i] = wanted[id]
	}
	changed, err := tx.HoldStock(ctx, ids, qty)
	if err != nil {
		return Placed{}, fmt.Errorf("hold the stock: %w", err)
	}
	if changed != int64(len(ids)) {
		// Unreachable under the row locks: the guard of Decide and the CHECK on
		// the stock agree. If it happens, nothing may be kept.
		return Placed{}, fmt.Errorf("hold the stock: %d bikes changed, want %d", changed, len(ids))
	}

	seq, err := tx.NextSeq(ctx, day.Format("2006-01-02"))
	if err != nil {
		return Placed{}, fmt.Errorf("issue the order number: %w", err)
	}
	number := Number(day, seq)

	orderID, err := tx.InsertOrder(ctx, OrderRow{
		Number: number, PlacedAt: instant, Name: req.Name, Phone: req.Phone, Delivery: req.Delivery,
		Address: req.Address, Payment: req.Payment, Note: req.Note,
		Subtotal: priced.Subtotal, DeliveryFee: priced.DeliveryFee, Total: priced.Total,
	})
	if err != nil {
		return Placed{}, fmt.Errorf("store order %s: %w", number, err)
	}
	if err := tx.InsertLines(ctx, orderID, priced.Lines); err != nil {
		return Placed{}, fmt.Errorf("store the lines of order %s: %w", number, err)
	}
	if err := tx.Commit(ctx); err != nil {
		// The outcome is unknown: the order may exist. The number is in the
		// log line so the operator can look for it.
		return Placed{}, fmt.Errorf("commit order %s: %w", number, err)
	}
	committed = true

	lines := make([]PlacedLine, len(priced.Lines))
	for i, l := range priced.Lines {
		lines[i] = PlacedLine{ID: l.ID, Name: l.Name, Price: l.Price, Image: l.Image, Variant: l.Size, Qty: l.Qty}
	}
	return Placed{
		Number: number, Placed: day.Format(placedLayout), Name: req.Name, Phone: req.Phone, Delivery: req.Delivery,
		Address: req.Address, Payment: req.Payment, Lines: lines,
		Subtotal: priced.Subtotal, DeliveryFee: priced.DeliveryFee, Total: priced.Total,
	}, nil
}

// The six statements of the design (erd/place-an-order-and-hold-the-bike.md
// section 2). Every value is a bound parameter.
const (
	// S1: lock and read every distinct bike, in one fixed order so that two
	// orders over overlapping bikes cannot deadlock. A field the record leaves
	// out reads as its zero value.
	sqlLockBikes = `SELECT "id", "name", COALESCE("brand", ''), COALESCE("status", ''), COALESCE("stock", 0), COALESCE("price", 0),
		COALESCE("sizes", ARRAY[]::text[]), COALESCE("sizes_out", ARRAY[]::text[]), COALESCE("image", '')
		FROM "products" WHERE "id" = ANY($1::text[]) ORDER BY "id" FOR UPDATE`
	// S2: the courier charge.
	sqlCourierFee = `SELECT COALESCE("delivery_fee", 0) FROM "shop" WHERE "id" = 'shop'`
	// S3: hold the bikes, one statement for all of them. The guard is a second
	// line of defence behind Decide and the CHECK on the stock.
	sqlHoldStock = `UPDATE "products" AS p SET "stock" = p."stock" - h.qty
		FROM unnest($1::text[], $2::bigint[]) AS h(id, qty)
		WHERE p."id" = h.id AND p."status" = 'live' AND p."stock" >= h.qty`
	// S4: the day's counter. The upsert takes the day row's lock, so orders are
	// numbered one after the other, and rolls back with a failed order.
	sqlNextSeq = `INSERT INTO "order_counters" ("day", "last_seq") VALUES ($1::text::date, 1)
		ON CONFLICT ("day") DO UPDATE SET "last_seq" = "order_counters"."last_seq" + 1, "updated_at" = now()
		RETURNING "last_seq"`
	// S5: the order.
	sqlInsertOrder = `INSERT INTO "orders" ("number", "placed_at", "name", "phone", "delivery", "address", "payment", "note", "subtotal", "delivery_fee", "total")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING "id"::text`
	// S6: every line in one statement; position is the 1-based place in the request.
	sqlInsertLines = `INSERT INTO "order_lines" ("order_id", "position", "product_id", "name", "size", "quantity", "unit_price")
		SELECT $1::text::uuid, l.pos::smallint, l.product_id, l.name, l.size, l.qty::integer, l.price
		FROM unnest($2::int[], $3::text[], $4::text[], $5::text[], $6::int[], $7::bigint[]) AS l(pos, product_id, name, size, qty, price)`
)

// pgTxn is the Postgres implementation of txn.
type pgTxn struct{ tx pgx.Tx }

func (t pgTxn) LockBikes(ctx context.Context, ids []string) (map[string]Bike, error) {
	rows, err := t.tx.Query(ctx, sqlLockBikes, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bikes := make(map[string]Bike, len(ids))
	for rows.Next() {
		var b Bike
		if err := rows.Scan(&b.ID, &b.Name, &b.Brand, &b.Status, &b.Stock, &b.Price, &b.Sizes, &b.SizesOut, &b.Image); err != nil {
			return nil, err
		}
		bikes[b.ID] = b
	}
	return bikes, rows.Err()
}

func (t pgTxn) CourierFee(ctx context.Context) (int64, error) {
	var fee int64
	err := t.tx.QueryRow(ctx, sqlCourierFee).Scan(&fee)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return fee, err
}

func (t pgTxn) HoldStock(ctx context.Context, ids []string, qty []int64) (int64, error) {
	tag, err := t.tx.Exec(ctx, sqlHoldStock, ids, qty)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (t pgTxn) NextSeq(ctx context.Context, day string) (int, error) {
	var seq int
	err := t.tx.QueryRow(ctx, sqlNextSeq, day).Scan(&seq)
	return seq, err
}

func (t pgTxn) InsertOrder(ctx context.Context, o OrderRow) (string, error) {
	var id string
	err := t.tx.QueryRow(ctx, sqlInsertOrder, o.Number, o.PlacedAt, o.Name, o.Phone, o.Delivery, o.Address, o.Payment, o.Note,
		o.Subtotal, o.DeliveryFee, o.Total).Scan(&id)
	return id, err
}

func (t pgTxn) InsertLines(ctx context.Context, orderID string, lines []PricedLine) error {
	n := len(lines)
	pos, qty := make([]int32, n), make([]int32, n)
	ids, names, sizes := make([]string, n), make([]string, n), make([]string, n)
	prices := make([]int64, n)
	for i, l := range lines {
		pos[i], qty[i] = int32(i+1), int32(l.Qty)
		ids[i], names[i], sizes[i], prices[i] = l.ID, l.Name, l.Size, l.Price
	}
	_, err := t.tx.Exec(ctx, sqlInsertLines, orderID, pos, ids, names, sizes, qty, prices)
	return err
}

func (t pgTxn) Commit(ctx context.Context) error   { return t.tx.Commit(ctx) }
func (t pgTxn) Rollback(ctx context.Context) error { return t.tx.Rollback(ctx) }
