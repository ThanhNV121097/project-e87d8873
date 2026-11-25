-- The owner's sessions (admin story). One row per sign-in that is still
-- valid: the owner is one account made from two secrets, so there is no
-- accounts table, and a session holds the email it was made for.
--
-- token_hash is an HMAC of the cookie's text keyed by a fingerprint of the two
-- secrets, never the token itself: a copy of this table is no use to anyone,
-- and changing either secret ends every session. expires_at is set by the
-- application's own clock (seven days after the sign-in); rows that have
-- expired are purged on the next sign-in.
CREATE TABLE "sessions" (
  "id"         uuid        NOT NULL DEFAULT gen_random_uuid(),
  "token_hash" bytea       NOT NULL,
  "email"      text        NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "expires_at" timestamptz NOT NULL,
  CONSTRAINT sessions_pkey PRIMARY KEY ("id"),
  CONSTRAINT uq_sessions_token_hash UNIQUE ("token_hash"),
  CONSTRAINT ck_sessions_token_hash_length CHECK (octet_length("token_hash") = 32),
  CONSTRAINT ck_sessions_expiry_after_creation CHECK ("expires_at" > "created_at")
);
