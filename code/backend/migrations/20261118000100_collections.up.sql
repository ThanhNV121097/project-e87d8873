-- The eight collections the shop shows, one typed table each (erd.md section 3).
--
-- Column names are the schema's field keys, verbatim, and every column name is
-- quoted: date, time, text, type, year, month, line, name, a and q are keywords
-- or type names in some places. A NULL column means the record has no such
-- field (the API leaves the key out). seq is the order a collection lists in.
--
-- The database must be UTF-8: normalize() in fold_text needs it.

-- Folds text for search, equality and sort: lower case, no accents, and the
-- Vietnamese d-with-stroke as d. translate() runs first; chr(768)..chr(879) is
-- the combining-mark block U+0300..U+036F that NFD splits an accent into.
CREATE FUNCTION fold_text(t text) RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT AS $$
  SELECT lower(regexp_replace(normalize(translate(t, 'Đđ', 'Dd'), NFD), '[' || chr(768) || '-' || chr(879) || ']', '', 'g'))
$$;

-- Natural-order key: digit runs are zero-padded to 40 digits, so "Talon 2"
-- sorts before "Talon 29 3". Compare it with COLLATE "C".
CREATE FUNCTION natural_key(t text) RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT AS $$
  SELECT coalesce(string_agg(CASE WHEN m[1] IS NOT NULL THEN lpad(m[1], 40, '0') ELSE m[2] END, '' ORDER BY n), '')
  FROM regexp_matches(fold_text(t), '(\d+)|(\D+)', 'g') WITH ORDINALITY AS r(m, n)
$$;

-- Bumps the compare-and-swap token and the edit time on every update of a
-- writable collection (products, categories).
CREATE FUNCTION touch_record() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW."version" := OLD."version" + 1;
  NEW."updated_at" := now();
  RETURN NEW;
END $$;

CREATE TABLE categories (
  "id"         text        NOT NULL,
  "seq"        bigint      GENERATED ALWAYS AS IDENTITY,
  "version"    bigint      NOT NULL DEFAULT 1,
  "name"       text        NOT NULL,
  "line"       text,
  "terrain"    text,
  "_source"    text,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT categories_pkey PRIMARY KEY ("id"),
  CONSTRAINT uq_categories_seq UNIQUE ("seq"),
  CONSTRAINT ck_categories_id_format CHECK ("id" ~ '^[a-z0-9][a-z0-9-]{0,63}$'),
  CONSTRAINT ck_categories_name_not_blank CHECK (btrim("name") <> '')
);

CREATE TRIGGER trg_categories_touch BEFORE UPDATE ON categories
  FOR EACH ROW EXECUTE FUNCTION touch_record();

CREATE TABLE products (
  "id"           text        NOT NULL,
  "seq"          bigint      GENERATED ALWAYS AS IDENTITY,
  "version"      bigint      NOT NULL DEFAULT 1,
  "name"         text        NOT NULL,
  "brand"        text,
  "category"     text,
  "year"         numeric,
  "price"        bigint,
  "was"          bigint,
  "price_listed" text,
  "fact"         text,
  "status"       text,
  "stock"        bigint,
  "sizes"        text[],
  "sizes_out"    text[],
  "size_fit"     text[],
  "deal"         boolean,
  "month"        boolean,
  "wheel"        text,
  "travel"       text,
  "frame"        text,
  "fork"         text,
  "shock"        text,
  "motor"        text,
  "drivetrain"   text,
  "brakes"       text,
  "wheels"       text,
  "tyres"        text,
  "weight"       text,
  "colour"       text,
  "description"  text,
  "ride"         text,
  "rating"       numeric,
  "review_count" numeric,
  "image"        text,
  "image_back"   text,
  "image_detail" text,
  "_source"      text,
  "_invented"    text[],
  "created_at"   timestamptz NOT NULL DEFAULT now(),
  "updated_at"   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT products_pkey PRIMARY KEY ("id"),
  CONSTRAINT uq_products_seq UNIQUE ("seq"),
  CONSTRAINT fk_products_category FOREIGN KEY ("category") REFERENCES categories ("id")
    ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT ck_products_id_format CHECK ("id" ~ '^[a-z0-9][a-z0-9-]{0,63}$'),
  CONSTRAINT ck_products_name_not_blank CHECK (btrim("name") <> ''),
  CONSTRAINT ck_products_brand CHECK ("brand" IN ('Giant', 'Liv')),
  CONSTRAINT ck_products_status CHECK ("status" IN ('draft', 'live')),
  CONSTRAINT ck_products_wheel CHECK ("wheel" IN ('29', 'mixed')),
  CONSTRAINT ck_products_price_nonneg CHECK ("price" >= 0),
  CONSTRAINT ck_products_was_nonneg CHECK ("was" >= 0),
  CONSTRAINT ck_products_stock_nonneg CHECK ("stock" >= 0),
  CONSTRAINT ck_products_live_has_price CHECK ("status" IS DISTINCT FROM 'live' OR COALESCE("price", 0) > 0)
);

CREATE INDEX idx_products_category ON products ("category");

CREATE TRIGGER trg_products_touch BEFORE UPDATE ON products
  FOR EACH ROW EXECUTE FUNCTION touch_record();

CREATE TABLE reviews (
  "id"         text        NOT NULL,
  "seq"        bigint      GENERATED ALWAYS AS IDENTITY,
  "name"       text        NOT NULL,
  "rating"     numeric,
  "date"       text,
  "bike"       text,
  "text"       text,
  "_source"    text,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT reviews_pkey PRIMARY KEY ("id"),
  CONSTRAINT uq_reviews_seq UNIQUE ("seq"),
  CONSTRAINT fk_reviews_bike FOREIGN KEY ("bike") REFERENCES products ("id")
    ON DELETE SET NULL ON UPDATE RESTRICT,
  CONSTRAINT ck_reviews_id_format CHECK ("id" ~ '^[a-z0-9][a-z0-9-]{0,63}$'),
  CONSTRAINT ck_reviews_name_not_blank CHECK (btrim("name") <> ''),
  CONSTRAINT ck_reviews_date_iso CHECK ("date" ~ '^\d{4}-\d{2}-\d{2}$')
);

CREATE INDEX idx_reviews_bike ON reviews ("bike");

CREATE TABLE gallery (
  "id"         text        NOT NULL,
  "seq"        bigint      GENERATED ALWAYS AS IDENTITY,
  "image"      text        NOT NULL,
  "title"      text,
  "caption"    text,
  "type"       text,
  "_source"    text,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT gallery_pkey PRIMARY KEY ("id"),
  CONSTRAINT uq_gallery_seq UNIQUE ("seq"),
  CONSTRAINT fk_gallery_type FOREIGN KEY ("type") REFERENCES categories ("id")
    ON DELETE SET NULL ON UPDATE RESTRICT,
  CONSTRAINT ck_gallery_id_format CHECK ("id" ~ '^[a-z0-9][a-z0-9-]{0,63}$'),
  CONSTRAINT ck_gallery_image_not_blank CHECK (btrim("image") <> '')
);

CREATE INDEX idx_gallery_type ON gallery ("type");

CREATE TABLE team (
  "id"         text        NOT NULL,
  "seq"        bigint      GENERATED ALWAYS AS IDENTITY,
  "name"       text        NOT NULL,
  "role"       text,
  "line"       text,
  "rides"      text,
  "_source"    text,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT team_pkey PRIMARY KEY ("id"),
  CONSTRAINT uq_team_seq UNIQUE ("seq"),
  CONSTRAINT ck_team_id_format CHECK ("id" ~ '^[a-z0-9][a-z0-9-]{0,63}$'),
  CONSTRAINT ck_team_name_not_blank CHECK (btrim("name") <> '')
);

CREATE TABLE services (
  "id"         text        NOT NULL,
  "seq"        bigint      GENERATED ALWAYS AS IDENTITY,
  "name"       text        NOT NULL,
  "line"       text,
  "time"       text,
  "price"      bigint,
  "_source"    text,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT services_pkey PRIMARY KEY ("id"),
  CONSTRAINT uq_services_seq UNIQUE ("seq"),
  CONSTRAINT ck_services_id_format CHECK ("id" ~ '^[a-z0-9][a-z0-9-]{0,63}$'),
  CONSTRAINT ck_services_name_not_blank CHECK (btrim("name") <> ''),
  CONSTRAINT ck_services_price_nonneg CHECK ("price" >= 0)
);

CREATE TABLE faqs (
  "id"         text        NOT NULL,
  "seq"        bigint      GENERATED ALWAYS AS IDENTITY,
  "q"          text        NOT NULL,
  "a"          text,
  "_source"    text,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT faqs_pkey PRIMARY KEY ("id"),
  CONSTRAINT uq_faqs_seq UNIQUE ("seq"),
  CONSTRAINT ck_faqs_id_format CHECK ("id" ~ '^[a-z0-9][a-z0-9-]{0,63}$'),
  CONSTRAINT ck_faqs_q_not_blank CHECK (btrim("q") <> '')
);

CREATE TABLE shop (
  "id"                  text        NOT NULL,
  "seq"                 bigint      GENERATED ALWAYS AS IDENTITY,
  "founded"             numeric,
  "bikes_built"         numeric,
  "rating"              numeric,
  "review_count"        numeric,
  "reply_hours"         numeric,
  "warranty_years"      numeric,
  "free_service_months" numeric,
  "delivery_local"      text,
  "delivery_nationwide" text,
  "delivery_fee"        bigint,
  "pickup"              text,
  "payment"             text[],
  "confirm"             text,
  "stock_rule"          text,
  "_source"             text,
  "_invented"           text[],
  "created_at"          timestamptz NOT NULL DEFAULT now(),
  "updated_at"          timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT shop_pkey PRIMARY KEY ("id"),
  CONSTRAINT uq_shop_seq UNIQUE ("seq"),
  CONSTRAINT ck_shop_id_format CHECK ("id" ~ '^[a-z0-9][a-z0-9-]{0,63}$'),
  CONSTRAINT ck_shop_delivery_fee_nonneg CHECK ("delivery_fee" >= 0)
);
