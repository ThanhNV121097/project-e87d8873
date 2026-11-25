package orders

import (
	"errors"
	"testing"
)

func talon() Bike {
	return Bike{ID: "talon-2", Name: "Talon 2", Brand: "Giant", Status: "live", Stock: 9, Price: 18990000,
		Sizes: []string{"XS", "S", "M", "L", "XL"}, Image: "/images/ref/talon-2.jpg"}
}

func stance() Bike {
	return Bike{ID: "stance-0", Name: "Stance 0", Brand: "Giant", Status: "live", Stock: 7, Price: 66490000,
		Sizes: []string{"XS", "S", "M", "L", "XL"}}
}

func shelf(bikes ...Bike) map[string]Bike {
	m := map[string]Bike{}
	for _, b := range bikes {
		m[b.ID] = b
	}
	return m
}

func order(delivery string, lines ...Line) Request {
	return Request{Name: "Minh", Phone: "0901234567", Delivery: delivery, Address: "12 Hùng Vương", Payment: "transfer", Lines: lines}
}

func sentenceOf(t *testing.T, err error) string {
	t.Helper()
	var r *Refusal
	if !errors.As(err, &r) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	return r.Sentence
}

func TestAnOrderIsPricedFromTheBikesRecordNotTheBrowser(t *testing.T) {
	p, err := Decide(order("local", Line{ID: "talon-2", Variant: "M", Qty: 2}), shelf(talon()), 350000)
	if err != nil {
		t.Fatal(err)
	}
	want := PricedLine{ID: "talon-2", Name: "Giant Talon 2", Size: "M", Image: "/images/ref/talon-2.jpg", Price: 18990000, Qty: 2}
	if len(p.Lines) != 1 || p.Lines[0] != want {
		t.Errorf("lines = %+v, want %+v", p.Lines, want)
	}
	if p.Subtotal != 37980000 || p.DeliveryFee != 0 || p.Total != 37980000 {
		t.Errorf("money = %d %d %d, want 37980000 0 37980000", p.Subtotal, p.DeliveryFee, p.Total)
	}
}

func TestTheCourierChargeAppliesToACourierDeliveryOnly(t *testing.T) {
	line := Line{ID: "talon-2", Variant: "M", Qty: 2}
	for _, c := range []struct {
		delivery string
		fee      int64
		total    int64
	}{{"courier", 350000, 38330000}, {"local", 0, 37980000}, {"pickup", 0, 37980000}} {
		p, err := Decide(order(c.delivery, line), shelf(talon()), 350000)
		if err != nil || p.DeliveryFee != c.fee || p.Total != c.total {
			t.Errorf("%s: fee %d total %d err %v, want %d and %d", c.delivery, p.DeliveryFee, p.Total, err, c.fee, c.total)
		}
	}
	if p, err := Decide(order("courier", line), shelf(talon()), 0); err != nil || p.DeliveryFee != 0 {
		t.Errorf("a shop with no charge recorded: fee %d, err %v, want 0", p.DeliveryFee, err)
	}
}

func TestTheSubtotalSumsEveryLine(t *testing.T) {
	p, err := Decide(order("local", Line{ID: "talon-2", Variant: "M", Qty: 1}, Line{ID: "stance-0", Variant: "S", Qty: 1}), shelf(talon(), stance()), 0)
	if err != nil || p.Subtotal != 18990000+66490000 {
		t.Errorf("subtotal = %d, err %v", p.Subtotal, err)
	}
}

func TestABikeThatCannotBeHadIsRefusedWithTheFixedSentence(t *testing.T) {
	draft := talon()
	draft.Status = "draft"
	soldOut := Bike{ID: "pique", Name: "Pique Advanced SL SE", Brand: "Liv", Status: "live", Stock: 0, Price: 1, Sizes: []string{"S"}, SizesOut: []string{"S"}}
	short := talon()
	short.Stock = 3
	out := talon()
	out.SizesOut = []string{"XL"}
	unknownStock := talon()
	unknownStock.Stock = 0
	noBrand := Bike{ID: "mystery", Name: "Mystery", Status: "live", Stock: 0, Price: 1}

	cases := []struct {
		name  string
		bikes map[string]Bike
		lines []Line
		want  string
	}{
		{"a draft", shelf(draft), []Line{{ID: "talon-2", Variant: "M", Qty: 1}}, "Giant Talon 2 is no longer in the shop."},
		{"no such bike", shelf(), []Line{{ID: "no-such-bike", Qty: 1}}, "A bike in your bag is no longer in the shop."},
		{"no stock", shelf(soldOut), []Line{{ID: "pique", Variant: "S", Qty: 1}}, "Liv Pique Advanced SL SE is out of stock."},
		{"a stock that is not recorded", shelf(unknownStock), []Line{{ID: "talon-2", Variant: "M", Qty: 1}}, "Giant Talon 2 is out of stock."},
		{"more than the stock", shelf(short), []Line{{ID: "talon-2", Variant: "M", Qty: 4}}, "Giant Talon 2: only 3 in stock."},
		{"two lines of one bike are counted together", shelf(short), []Line{{ID: "talon-2", Variant: "M", Qty: 2}, {ID: "talon-2", Variant: "S", Qty: 2}}, "Giant Talon 2: only 3 in stock."},
		{"a size the bike lacks", shelf(talon()), []Line{{ID: "talon-2", Variant: "XXL", Qty: 1}}, "Giant Talon 2 is not available in size XXL."},
		{"a size compared exactly", shelf(talon()), []Line{{ID: "talon-2", Variant: "m", Qty: 1}}, "Giant Talon 2 is not available in size m."},
		{"a size the owner marked out", shelf(out), []Line{{ID: "talon-2", Variant: "XL", Qty: 1}}, "Giant Talon 2 is not available in size XL."},
		{"no size", shelf(talon()), []Line{{ID: "talon-2", Qty: 1}}, "Giant Talon 2 needs a frame size."},
		{"stock is checked before size", shelf(short), []Line{{ID: "talon-2", Variant: "XXL", Qty: 4}}, "Giant Talon 2: only 3 in stock."},
		{"the first failing line is named", shelf(stance(), soldOut), []Line{{ID: "stance-0", Variant: "S", Qty: 1}, {ID: "pique", Variant: "S", Qty: 1}, {ID: "no-such-bike", Qty: 1}}, "Liv Pique Advanced SL SE is out of stock."},
		{"a bike with no brand is named by its name", shelf(noBrand), []Line{{ID: "mystery", Qty: 1}}, "Mystery is out of stock."},
	}
	for _, c := range cases {
		_, err := Decide(order("local", c.lines...), c.bikes, 0)
		if got := sentenceOf(t, err); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

func TestAQuantityEqualToTheStockIsAccepted(t *testing.T) {
	short := talon()
	short.Stock = 3
	if _, err := Decide(order("local", Line{ID: "talon-2", Variant: "M", Qty: 3}), shelf(short), 0); err != nil {
		t.Errorf("3 of 3 refused: %v", err)
	}
}

func TestABikeWithNoSizesNeedsNoneAndStoresNone(t *testing.T) {
	bare := Bike{ID: "frame", Name: "Frame", Brand: "Giant", Status: "live", Stock: 2, Price: 5}
	for _, variant := range []string{"", "M"} {
		p, err := Decide(order("local", Line{ID: "frame", Variant: variant, Qty: 1}), shelf(bare), 0)
		if err != nil || p.Lines[0].Size != "" {
			t.Errorf("variant %q: size %q, err %v, want accepted with no size", variant, p.Lines[0].Size, err)
		}
	}
}

// A live bike with no price is a fault in the shop's data, not a refusal of
// the customer's order, and is never turned into a priced line.
func TestALiveBikeWithNoPriceIsAFailureNotARefusal(t *testing.T) {
	broken := talon()
	broken.Price = 0
	_, err := Decide(order("local", Line{ID: "talon-2", Variant: "M", Qty: 1}), shelf(broken), 0)
	var r *Refusal
	if err == nil || errors.As(err, &r) {
		t.Errorf("err = %v, want an error that is not a refusal", err)
	}
}
