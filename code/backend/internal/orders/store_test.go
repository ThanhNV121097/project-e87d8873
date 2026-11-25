package orders

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

var errStep = errors.New("step failed")

// fakeTxn records the steps of a transaction and fails the one it is told to.
type fakeTxn struct {
	calls   []string
	bikes   map[string]Bike
	fee     int64
	short   bool
	failAt  string
	heldIDs []string
	heldQty []int64
	day     string
	row     OrderRow
	lines   []PricedLine
}

func (f *fakeTxn) step(name string) error {
	f.calls = append(f.calls, name)
	if f.failAt == name {
		return errStep
	}
	return nil
}

func (f *fakeTxn) LockBikes(context.Context, []string) (map[string]Bike, error) {
	if err := f.step("lock"); err != nil {
		return nil, err
	}
	return f.bikes, nil
}

func (f *fakeTxn) CourierFee(context.Context) (int64, error) {
	if err := f.step("fee"); err != nil {
		return 0, err
	}
	return f.fee, nil
}

func (f *fakeTxn) HoldStock(_ context.Context, ids []string, qty []int64) (int64, error) {
	f.heldIDs, f.heldQty = ids, qty
	if err := f.step("hold"); err != nil {
		return 0, err
	}
	if f.short {
		return int64(len(ids)) - 1, nil
	}
	return int64(len(ids)), nil
}

func (f *fakeTxn) NextSeq(_ context.Context, day string) (int, error) {
	f.day = day
	if err := f.step("seq"); err != nil {
		return 0, err
	}
	return 1, nil
}

func (f *fakeTxn) InsertOrder(_ context.Context, o OrderRow) (string, error) {
	f.row = o
	if err := f.step("order"); err != nil {
		return "", err
	}
	return "order-1", nil
}

func (f *fakeTxn) InsertLines(_ context.Context, _ string, lines []PricedLine) error {
	f.lines = lines
	return f.step("lines")
}

func (f *fakeTxn) Commit(context.Context) error   { return f.step("commit") }
func (f *fakeTxn) Rollback(context.Context) error { return f.step("rollback") }

func storeOver(f *fakeTxn) *store {
	return &store{begin: func(context.Context) (txn, error) { return f, nil }}
}

// instant is 01:00 on 17 November in Vietnam.
var instant = time.Date(2026, 11, 16, 18, 0, 0, 0, time.UTC)

type counted struct{ reads int }

func (c *counted) now() time.Time {
	c.reads++
	return instant
}

func newTxn() *fakeTxn {
	return &fakeTxn{bikes: shelf(talon(), stance()), fee: 350000}
}

func TestTheStepsRunInTheContractsOrder(t *testing.T) {
	line := Line{ID: "talon-2", Variant: "M", Qty: 2}
	f := newTxn()
	if _, err := storeOver(f).Place(context.Background(), order("local", line), (&counted{}).now); err != nil {
		t.Fatal(err)
	}
	if want := []string{"lock", "hold", "seq", "order", "lines", "commit"}; !reflect.DeepEqual(f.calls, want) {
		t.Errorf("local: %v, want %v (no fee read, no rollback)", f.calls, want)
	}

	f = newTxn()
	if _, err := storeOver(f).Place(context.Background(), order("courier", line), (&counted{}).now); err != nil {
		t.Fatal(err)
	}
	if want := []string{"lock", "fee", "hold", "seq", "order", "lines", "commit"}; !reflect.DeepEqual(f.calls, want) {
		t.Errorf("courier: %v, want %v", f.calls, want)
	}
}

func TestAnAcceptedOrderIsAnsweredFromWhatWasWritten(t *testing.T) {
	f := newTxn()
	clock := &counted{}
	req := order("courier", Line{ID: "talon-2", Variant: "M", Qty: 2})
	req.Note = "Please call after 5pm"
	got, err := storeOver(f).Place(context.Background(), req, clock.now)
	if err != nil {
		t.Fatal(err)
	}

	want := Placed{
		Number: "DEO-261117-0001", Placed: "2026-11-17T01:00:00+07:00", Name: "Minh", Phone: "0901234567", Delivery: "courier",
		Address: "12 Hùng Vương", Payment: "transfer",
		Lines:    []PlacedLine{{ID: "talon-2", Name: "Giant Talon 2", Price: 18990000, Image: "/images/ref/talon-2.jpg", Variant: "M", Qty: 2}},
		Subtotal: 37980000, DeliveryFee: 350000, Total: 38330000,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("answer = %+v, want %+v", got, want)
	}
	if clock.reads != 1 {
		t.Errorf("the clock was read %d times, want once", clock.reads)
	}
	if !f.row.PlacedAt.Equal(instant) || f.day != "2026-11-17" {
		t.Errorf("the order was stored at %v and numbered for %s, want one instant: %v and 2026-11-17", f.row.PlacedAt, f.day, instant)
	}
	if f.row.Note != "Please call after 5pm" || f.row.Number != want.Number || f.row.Total != 38330000 {
		t.Errorf("stored row = %+v", f.row)
	}
	if len(f.lines) != 1 || f.lines[0].Price != 18990000 {
		t.Errorf("stored lines = %+v", f.lines)
	}
}

func TestTheStockHeldForABikeIsTheSumOfItsLinesInIdOrder(t *testing.T) {
	f := newTxn()
	req := order("local", Line{ID: "talon-2", Variant: "M", Qty: 2}, Line{ID: "stance-0", Variant: "S", Qty: 1}, Line{ID: "talon-2", Variant: "S", Qty: 3})
	got, err := storeOver(f).Place(context.Background(), req, (&counted{}).now)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f.heldIDs, []string{"stance-0", "talon-2"}) || !reflect.DeepEqual(f.heldQty, []int64{1, 5}) {
		t.Errorf("held %v x %v, want stance-0 x 1 and talon-2 x 5", f.heldIDs, f.heldQty)
	}
	if len(got.Lines) != 3 || got.Lines[0].ID != "talon-2" || got.Lines[1].ID != "stance-0" {
		t.Errorf("the answer's lines are not in the order sent: %+v", got.Lines)
	}
}

func TestARefusalHoldsNothingAndStoresNothing(t *testing.T) {
	f := newTxn()
	f.bikes = shelf()
	_, err := storeOver(f).Place(context.Background(), order("local", Line{ID: "no-such-bike", Qty: 1}), (&counted{}).now)
	if got := sentenceOf(t, err); !strings.Contains(got, "no longer in the shop") {
		t.Errorf("sentence = %q", got)
	}
	if want := []string{"lock", "rollback"}; !reflect.DeepEqual(f.calls, want) {
		t.Errorf("calls = %v, want %v: a refusal must stop before the stock is touched", f.calls, want)
	}
}

func TestAnErrorInAnyStepRollsEverythingBackAndNeverCommits(t *testing.T) {
	line := Line{ID: "talon-2", Variant: "M", Qty: 2}
	for _, step := range []string{"lock", "fee", "hold", "seq", "order", "lines", "commit"} {
		f := newTxn()
		f.failAt = step
		_, err := storeOver(f).Place(context.Background(), order("courier", line), (&counted{}).now)
		if !errors.Is(err, errStep) {
			t.Errorf("%s: err = %v, want the step's error", step, err)
		}
		if last := f.calls[len(f.calls)-1]; last != "rollback" {
			t.Errorf("%s: calls = %v, want the last to be rollback", step, f.calls)
		}
		commits := 0
		for _, c := range f.calls {
			if c == "commit" {
				commits++
			}
		}
		if (step == "commit") != (commits == 1) || commits > 1 {
			t.Errorf("%s: committed %d times: %v", step, commits, f.calls)
		}
	}
}

func TestACommitThatFailedNamesTheOrderInItsError(t *testing.T) {
	f := newTxn()
	f.failAt = "commit"
	_, err := storeOver(f).Place(context.Background(), order("local", Line{ID: "talon-2", Variant: "M", Qty: 1}), (&counted{}).now)
	if err == nil || !strings.Contains(err.Error(), "DEO-261117-0001") {
		t.Errorf("err = %v, want the order number in it so the operator can find the order", err)
	}
}

func TestAHoldThatChangedTheWrongNumberOfBikesIsAFailure(t *testing.T) {
	f := newTxn()
	f.short = true
	_, err := storeOver(f).Place(context.Background(), order("local", Line{ID: "talon-2", Variant: "M", Qty: 1}), (&counted{}).now)
	var r *Refusal
	if err == nil || errors.As(err, &r) {
		t.Errorf("err = %v, want a failure that is not a refusal", err)
	}
	for _, c := range f.calls {
		if c == "commit" {
			t.Errorf("committed after a hold that changed the wrong number of bikes: %v", f.calls)
		}
	}
}

func TestATransactionThatCannotBeginStoresNothing(t *testing.T) {
	s := &store{begin: func(context.Context) (txn, error) { return nil, errStep }}
	if _, err := s.Place(context.Background(), order("local", Line{ID: "talon-2", Qty: 1}), (&counted{}).now); !errors.Is(err, errStep) {
		t.Errorf("err = %v, want the begin error kept in the chain", err)
	}
}
