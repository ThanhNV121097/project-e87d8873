-- Undoes 20261118000100_collections.up.sql. For development only: the runner
-- never applies a down file, and a shop that holds owner edits or orders is
-- restored from a backup, not rolled back. Delete the schema_migrations row of
-- this version by hand after running it.
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS gallery;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS team;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS faqs;
DROP TABLE IF EXISTS shop;
DROP FUNCTION IF EXISTS natural_key(text);
DROP FUNCTION IF EXISTS fold_text(text);
DROP FUNCTION IF EXISTS touch_record();
