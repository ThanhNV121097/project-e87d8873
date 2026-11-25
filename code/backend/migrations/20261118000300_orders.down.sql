-- Development only: this destroys every order. The runner never applies a down
-- file; rolling back a shop that has taken orders is a restore from a backup.

DROP TABLE "order_lines";
DROP TABLE "orders";
DROP TABLE "order_counters";
