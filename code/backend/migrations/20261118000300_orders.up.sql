-- Orders: what a customer placed, what was in it as it was when ordered, and the
-- daily counter the order number is issued from (erd.md sections 3.9 to 3.11).
-- Every object is new: no existing table is read or changed, and order_lines
-- keeps a copy of the bike's id, never a foreign key to the bike, so a bike the
-- owner deletes leaves its sold lines as they were.

CREATE TABLE "orders" (
  "id"           uuid        NOT NULL DEFAULT gen_random_uuid(),
  "number"       text        NOT NULL,
  "placed_at"    timestamptz NOT NULL DEFAULT now(),
  "name"         text        NOT NULL,
  "phone"        text        NOT NULL,
  "delivery"     text        NOT NULL,
  "address"      text        NOT NULL DEFAULT '',
  "payment"      text        NOT NULL,
  "note"         text        NOT NULL DEFAULT '',
  "subtotal"     bigint      NOT NULL,
  "delivery_fee" bigint      NOT NULL,
  "total"        bigint      NOT NULL,
  CONSTRAINT orders_pkey PRIMARY KEY ("id"),
  CONSTRAINT uq_orders_number UNIQUE ("number"),
  CONSTRAINT ck_orders_number_format CHECK ("number" ~ '^DEO-\d{6}-\d{4}$'),
  CONSTRAINT ck_orders_name CHECK (btrim("name") <> '' AND char_length("name") <= 100),
  CONSTRAINT ck_orders_phone CHECK (btrim("phone") <> '' AND char_length("phone") <= 30),
  CONSTRAINT ck_orders_delivery CHECK ("delivery" IN ('local', 'courier', 'pickup')),
  CONSTRAINT ck_orders_payment CHECK ("payment" IN ('transfer', 'cod', 'shop')),
  CONSTRAINT ck_orders_cod_local CHECK ("payment" <> 'cod' OR "delivery" = 'local'),
  CONSTRAINT ck_orders_address CHECK (char_length("address") <= 300 AND ("delivery" = 'pickup' OR btrim("address") <> '')),
  CONSTRAINT ck_orders_note_length CHECK (char_length("note") <= 1000),
  CONSTRAINT ck_orders_money CHECK ("subtotal" >= 0 AND "delivery_fee" >= 0 AND "total" = "subtotal" + "delivery_fee"),
  CONSTRAINT ck_orders_fee_courier_only CHECK ("delivery" = 'courier' OR "delivery_fee" = 0)
);

CREATE TABLE "order_lines" (
  "id"         uuid     NOT NULL DEFAULT gen_random_uuid(),
  "order_id"   uuid     NOT NULL,
  "position"   smallint NOT NULL,
  "product_id" text     NOT NULL,
  "name"       text     NOT NULL,
  "size"       text     NOT NULL DEFAULT '',
  "quantity"   integer  NOT NULL,
  "unit_price" bigint   NOT NULL,
  CONSTRAINT order_lines_pkey PRIMARY KEY ("id"),
  CONSTRAINT fk_order_lines_order_id FOREIGN KEY ("order_id") REFERENCES "orders" ("id")
    ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT uq_order_lines_order_id_position UNIQUE ("order_id", "position"),
  CONSTRAINT ck_order_lines_position CHECK ("position" BETWEEN 1 AND 20),
  CONSTRAINT ck_order_lines_quantity CHECK ("quantity" BETWEEN 1 AND 10),
  CONSTRAINT ck_order_lines_unit_price CHECK ("unit_price" > 0),
  CONSTRAINT ck_order_lines_name_not_blank CHECK (btrim("name") <> ''),
  CONSTRAINT ck_order_lines_size_length CHECK (char_length("size") <= 20)
);

CREATE TABLE "order_counters" (
  "day"        date        NOT NULL,
  "last_seq"   integer     NOT NULL,
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT order_counters_pkey PRIMARY KEY ("day"),
  CONSTRAINT ck_order_counters_last_seq_range CHECK ("last_seq" BETWEEN 1 AND 9999)
);
