package orders

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Bike is what the order reads of a bike, under the lock the transaction took.
// A bike with no brand, stock, price or sizes in its record arrives here with
// the zero value.
type Bike struct {
	ID       string
	Name     string
	Brand    string
	Status   string
	Stock    int64
	Price    int64
	Sizes    []string
	SizesOut []string
	Image    string
}

// PricedLine is a line as it will be stored: the bike's name and price when it
// was ordered, never what the browser said.
type PricedLine struct {
	ID    string
	Name  string
	Size  string
	Image string
	Price int64
	Qty   int64
}

// Priced is an accepted order with its money, computed from the shop's records.
type Priced struct {
	Lines       []PricedLine
	Subtotal    int64
	DeliveryFee int64
	Total       int64
}

// Refusal is a bike that cannot be had. Its sentence names the bike and is
// shown to the customer as it stands.
type Refusal struct{ Sentence string }

func (r *Refusal) Error() string { return r.Sentence }

func refuse(format string, a ...any) *Refusal {
	return &Refusal{Sentence: fmt.Sprintf(format, a...)}
}

// errBrokenRecord is a bike the shop calls live that has no price: a fault in
// the shop's data, not a mistake of the customer, so it is never a refusal.
var errBrokenRecord = errors.New("a live bike has no price")

// Decide walks the lines in the order they were sent and refuses at the first
// that fails, checking each in this order: the bike exists, is live, has the
// stock (the quantities of one bike are counted together), and has the size.
// It never touches the database. courierFee is the shop's courier charge; it
// applies only to a courier delivery.
func Decide(req Request, bikes map[string]Bike, courierFee int64) (Priced, error) {
	held := map[string]int64{}
	lines := make([]PricedLine, 0, len(req.Lines))
	for _, l := range req.Lines {
		bike, ok := bikes[l.ID]
		if !ok {
			return Priced{}, refuse("A bike in your bag is no longer in the shop.")
		}
		label := strings.TrimSpace(bike.Brand + " " + bike.Name)
		if bike.Status != "live" {
			return Priced{}, refuse("%s is no longer in the shop.", label)
		}
		held[l.ID] += l.Qty
		if held[l.ID] > bike.Stock {
			if bike.Stock <= 0 {
				return Priced{}, refuse("%s is out of stock.", label)
			}
			return Priced{}, refuse("%s: only %d in stock.", label, bike.Stock)
		}
		size := ""
		if len(bike.Sizes) > 0 {
			if l.Variant == "" {
				return Priced{}, refuse("%s needs a frame size.", label)
			}
			if !slices.Contains(bike.Sizes, l.Variant) || slices.Contains(bike.SizesOut, l.Variant) {
				return Priced{}, refuse("%s is not available in size %s.", label, l.Variant)
			}
			size = l.Variant
		}
		lines = append(lines, PricedLine{ID: bike.ID, Name: label, Size: size, Image: bike.Image, Price: bike.Price, Qty: l.Qty})
	}

	p := Priced{Lines: lines}
	for _, l := range lines {
		if l.Price <= 0 {
			return Priced{}, fmt.Errorf("%w: %s", errBrokenRecord, l.ID)
		}
		p.Subtotal += l.Price * l.Qty
	}
	if req.Delivery == deliveryCourier {
		p.DeliveryFee = courierFee
	}
	p.Total = p.Subtotal + p.DeliveryFee
	return p, nil
}
