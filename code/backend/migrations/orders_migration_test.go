package migrations

import (
	"regexp"
	"strings"
	"testing"
)

const (
	ordersUp   = "20261118000300_orders.up.sql"
	ordersDown = "20261118000300_orders.down.sql"
)

var createdTable = regexp.MustCompile(`(?i)CREATE TABLE "([a-z_]+)"`)
var droppedTable = regexp.MustCompile(`(?i)DROP TABLE "([a-z_]+)"`)

// The orders migration creates the three tables of the design and nothing
// else, depends on no other table, and its down file undoes exactly it.
func TestTheOrdersMigrationCreatesThreeTablesAndDependsOnNoBike(t *testing.T) {
	up, down := read(t, ordersUp), read(t, ordersDown)

	var created []string
	for _, m := range createdTable.FindAllStringSubmatch(up, -1) {
		created = append(created, m[1])
	}
	if strings.Join(created, ",") != "orders,order_lines,order_counters" {
		t.Errorf("the up file creates %v, want orders, order_lines, order_counters", created)
	}
	if n := strings.Count(strings.ToLower(up), "create table"); n != 3 {
		t.Errorf("the up file holds %d CREATE TABLE statements, want 3", n)
	}

	var dropped []string
	for _, m := range droppedTable.FindAllStringSubmatch(down, -1) {
		dropped = append(dropped, m[1])
	}
	if strings.Join(dropped, ",") != "order_lines,orders,order_counters" {
		t.Errorf("the down file drops %v, want order_lines, orders, order_counters", dropped)
	}

	for name, body := range map[string]string{ordersUp: up, ordersDown: down} {
		lower := strings.ToLower(body)
		if strings.Contains(lower, "products") {
			t.Errorf("%s mentions products: an order line copies the bike's id and has no foreign key to it", name)
		}
		if strings.Contains(lower, "concurrent"+"ly") {
			t.Errorf("%s mentions the word that moves a file outside its transaction", name)
		}
	}
	if n := strings.Count(strings.ToLower(up), "references"); n != 1 || !strings.Contains(up, `REFERENCES "orders" ("id")`) {
		t.Errorf("the up file has %d foreign keys, want exactly one, to orders", n)
	}
}

// The names the design gives are the names the database carries, and the
// number's pattern is the plain literal the runner's splitter reads as one string.
func TestTheOrdersMigrationUsesTheDesignsConstraintNames(t *testing.T) {
	up := read(t, ordersUp)
	for _, want := range []string{
		"uq_orders_number", "ck_orders_number_format", "ck_orders_name", "ck_orders_phone", "ck_orders_delivery",
		"ck_orders_payment", "ck_orders_cod_local", "ck_orders_address", "ck_orders_note_length", "ck_orders_money",
		"ck_orders_fee_courier_only", "fk_order_lines_order_id", "uq_order_lines_order_id_position",
		"ck_order_lines_position", "ck_order_lines_quantity", "ck_order_lines_unit_price", "ck_order_lines_name_not_blank",
		"ck_order_lines_size_length", "ck_order_counters_last_seq_range", `'^DEO-\d{6}-\d{4}$'`,
		"ON DELETE CASCADE ON UPDATE RESTRICT",
	} {
		if !strings.Contains(up, want) {
			t.Errorf("the orders migration does not contain %q", want)
		}
	}
	if strings.Contains(up, "E'") || strings.Contains(up, "$$") {
		t.Error("the orders migration uses an escape string or dollar quoting; the design asks for plain literals")
	}
}
