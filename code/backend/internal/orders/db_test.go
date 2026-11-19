package orders

import (
	"context"
	"errors"
	"os"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/migrate"
	"github.com/ThanhNV121097/project-e87d8873/backend/migrations"
)

// These tests run the six statements, the row locks and the CHECKs against a
// real Postgres, which no fake can prove. They need TEST_DATABASE_URL (an empty
// database or one already migrated) and are skipped without it. Every bike
// they use has an id starting zz-order-test, and every order they place is
// deleted again.

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := migrate.Up(ctx, pool, migrations.FS); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

// addBike inserts a live bike with sizes M and L (L marked out) and removes it
// when the test ends.
func addBike(t *testing.T, pool *pgxpool.Pool, id string, stock int) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `INSERT INTO products (id, name, brand, status, price, stock, sizes, sizes_out, image)
		VALUES ($1, 'Test bike', 'Giant', 'live', 1000000, $2, ARRAY['M','L'], ARRAY['L'], '/images/test.jpg')`, id, stock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id) })
}

// forget deletes the orders a test placed, and their lines with them.
func forget(t *testing.T, pool *pgxpool.Pool, numbers ...string) {
	t.Helper()
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM orders WHERE number = ANY($1)`, numbers) })
}

func stockOf(t *testing.T, pool *pgxpool.Pool, id string) int64 {
	t.Helper()
	var stock int64
	if err := pool.QueryRow(context.Background(), `SELECT stock FROM products WHERE id = $1`, id).Scan(&stock); err != nil {
		t.Fatal(err)
	}
	return stock
}

func linesOf(t *testing.T, pool *pgxpool.Pool, id string) int64 {
	t.Helper()
	var n int64
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM order_lines WHERE product_id = $1`, id).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func one(id, size string, qty int64) Request {
	return Request{Name: "Minh", Phone: "0901234567", Delivery: "local", Address: "12 Hùng Vương", Payment: "transfer",
		Lines: []Line{{ID: id, Variant: size, Qty: qty}}}
}

func TestPlacingAnOrderStoresItAndHoldsTheBike(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	const id = "zz-order-test-a"
	addBike(t, pool, id, 5)

	req := one(id, "M", 2)
	req.Note = "Please call after 5pm"
	placed, err := NewStore(pool).Place(ctx, req, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	forget(t, pool, placed.Number)

	if !regexp.MustCompile(`^DEO-\d{6}-\d{4}$`).MatchString(placed.Number) {
		t.Errorf("number = %q", placed.Number)
	}
	if at, err := time.Parse(time.RFC3339, placed.Placed); err != nil || at.Before(time.Now().Add(-time.Minute)) {
		t.Errorf("placed = %q (%v)", placed.Placed, err)
	}
	wantLine := PlacedLine{ID: id, Name: "Giant Test bike", Price: 1000000, Image: "/images/test.jpg", Variant: "M", Qty: 2}
	if len(placed.Lines) != 1 || placed.Lines[0] != wantLine || placed.Subtotal != 2000000 || placed.DeliveryFee != 0 || placed.Total != 2000000 {
		t.Errorf("answer = %+v", placed)
	}
	if got := stockOf(t, pool, id); got != 3 {
		t.Errorf("stock = %d, want 3", got)
	}

	var (
		name, phone, delivery, address, payment, note string
		subtotal, fee, total                          int64
		product, lineName, size                       string
		qty                                           int
		unit                                          int64
	)
	err = pool.QueryRow(ctx, `SELECT o.name, o.phone, o.delivery, o.address, o.payment, o.note, o.subtotal, o.delivery_fee, o.total,
			l.product_id, l.name, l.size, l.quantity, l.unit_price
		FROM orders o JOIN order_lines l ON l.order_id = o.id WHERE o.number = $1`, placed.Number).
		Scan(&name, &phone, &delivery, &address, &payment, &note, &subtotal, &fee, &total, &product, &lineName, &size, &qty, &unit)
	if err != nil {
		t.Fatal(err)
	}
	if name != "Minh" || phone != "0901234567" || delivery != "local" || address != "12 Hùng Vương" || payment != "transfer" || note != "Please call after 5pm" ||
		subtotal != 2000000 || fee != 0 || total != 2000000 || product != id || lineName != "Giant Test bike" || size != "M" || qty != 2 || unit != 1000000 {
		t.Errorf("the stored order differs from the answer")
	}

	// A bike the owner deletes leaves the line it was sold on.
	if _, err := pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
	if n := linesOf(t, pool, id); n != 1 {
		t.Errorf("%d lines left after the bike was deleted, want 1", n)
	}
}

func TestACourierOrderIsChargedTheShopsOwnFee(t *testing.T) {
	pool := testPool(t)
	const id = "zz-order-test-fee"
	addBike(t, pool, id, 5)

	var fee int64
	if err := pool.QueryRow(context.Background(), `SELECT COALESCE(delivery_fee, 0) FROM shop WHERE id = 'shop'`).Scan(&fee); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal(err)
	}
	req := one(id, "M", 1)
	req.Delivery = "courier"
	placed, err := NewStore(pool).Place(context.Background(), req, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	forget(t, pool, placed.Number)
	if placed.DeliveryFee != fee || placed.Total != 1000000+fee {
		t.Errorf("fee %d total %d, want the shop's %d", placed.DeliveryFee, placed.Total, fee)
	}
}

func TestARefusedOrderChangesNothing(t *testing.T) {
	pool := testPool(t)
	const id = "zz-order-test-b"
	addBike(t, pool, id, 3)
	st := NewStore(pool)

	for name, req := range map[string]Request{
		"more than the stock": one(id, "M", 4),
		"a size marked out":   one(id, "L", 1),
		"a size that is none": one(id, "XXL", 1),
		"no size":             one(id, "", 1),
	} {
		_, err := st.Place(context.Background(), req, time.Now)
		var refusal *Refusal
		if !errors.As(err, &refusal) {
			t.Errorf("%s: err = %v, want a refusal", name, err)
		}
	}
	if got := stockOf(t, pool, id); got != 3 {
		t.Errorf("stock = %d after refusals, want 3", got)
	}
	if n := linesOf(t, pool, id); n != 0 {
		t.Errorf("%d lines stored for refused orders", n)
	}
}

func TestTwentyOrdersForTheLastBikeGiveOneOrderAndNoNegativeStock(t *testing.T) {
	pool := testPool(t)
	const id = "zz-order-test-c"
	addBike(t, pool, id, 1)
	st := NewStore(pool)

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		numbers  []string
		accepted int
		refused  int
		failed   []error
	)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			placed, err := st.Place(context.Background(), one(id, "M", 1), time.Now)
			mu.Lock()
			defer mu.Unlock()
			var refusal *Refusal
			switch {
			case err == nil:
				accepted++
				numbers = append(numbers, placed.Number)
			case errors.As(err, &refusal):
				refused++
			default:
				failed = append(failed, err)
			}
		}()
	}
	wg.Wait()
	forget(t, pool, numbers...)

	if accepted != 1 || refused != 19 || len(failed) != 0 {
		t.Errorf("accepted %d, refused %d, failed %v; want 1, 19 and none", accepted, refused, failed)
	}
	if got := stockOf(t, pool, id); got != 0 {
		t.Errorf("stock = %d, want 0", got)
	}
	if n := linesOf(t, pool, id); n != 1 {
		t.Errorf("%d lines stored, want 1", n)
	}
}

func TestOrdersPlacedAtTheSameMomentGetDifferentNumbers(t *testing.T) {
	pool := testPool(t)
	const id = "zz-order-test-d"
	addBike(t, pool, id, 20)
	st := NewStore(pool)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		numbers []string
		failed  []error
	)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			placed, err := st.Place(context.Background(), one(id, "M", 1), time.Now)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failed = append(failed, err)
				return
			}
			numbers = append(numbers, placed.Number)
		}()
	}
	wg.Wait()
	forget(t, pool, numbers...)

	seen := map[string]bool{}
	for _, n := range numbers {
		seen[n] = true
	}
	if len(failed) != 0 || len(numbers) != 20 || len(seen) != 20 {
		t.Errorf("%d orders, %d different numbers, failures %v; want 20, 20 and none", len(numbers), len(seen), failed)
	}
	if got := stockOf(t, pool, id); got != 0 {
		t.Errorf("stock = %d, want 0", got)
	}
}

// A failure after the stock was lowered and the order written leaves neither:
// the trigger raises on the insert of the lines, the last write of the
// transaction. No code of the product exists for this test.
func TestAFailureAfterTheStockWasLoweredKeepsNeitherTheOrderNorTheChange(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	const id = "zz-order-test-e"
	addBike(t, pool, id, 4)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DROP TRIGGER IF EXISTS zz_fail_line ON order_lines`)
		_, _ = pool.Exec(ctx, `DROP FUNCTION IF EXISTS zz_fail_line()`)
	})
	if _, err := pool.Exec(ctx, `CREATE FUNCTION zz_fail_line() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected'; END $$`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `CREATE TRIGGER zz_fail_line BEFORE INSERT ON order_lines FOR EACH ROW EXECUTE FUNCTION zz_fail_line()`); err != nil {
		t.Fatal(err)
	}

	var before int64
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM orders`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	_, err := NewStore(pool).Place(ctx, one(id, "M", 2), time.Now)
	var refusal *Refusal
	if err == nil || errors.As(err, &refusal) {
		t.Fatalf("err = %v, want a failure that is not a refusal", err)
	}
	var after int64
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM orders`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Errorf("%d orders before, %d after: the order row survived", before, after)
	}
	if got := stockOf(t, pool, id); got != 4 {
		t.Errorf("stock = %d, want the unchanged 4", got)
	}
}
