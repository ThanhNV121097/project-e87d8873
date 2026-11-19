-- Undoes 20261118000200_seed_collections.up.sql: deletes the seeded ids, the
-- children first. It stops (RESTRICT) if a bike the owner added still uses a
-- seeded riding type; that is intended. For development only: the runner never
-- applies a down file. Delete the schema_migrations row of this version by
-- hand after running it.
DELETE FROM reviews WHERE "id" IN ('r-minh', 'r-lan', 'r-huy', 'r-thao', 'r-quan', 'r-ha', 'r-tuan', 'r-phuc', 'r-trang', 'r-khanh', 'r-dung', 'r-vy');
DELETE FROM gallery WHERE "id" IN ('pass-climb', 'rocky-descent', 'forest-singletrack', 'group-ride', 'dirt-jump', 'sunset-rest', 'ridge-view');
DELETE FROM products WHERE "id" IN ('talon-2', 'talon-29-3', 'talon-29-1', 'tempt-29-0', 'anthem-x-advanced-sl-3', 'anthem-advanced-sl-0', 'pique-advanced-sl-se', 'stance-0', 'stance', 'trance-x-1', 'intrigue-x-2', 'talon-e-plus-1', 'tempt-e-plus-1', 'embolden-e-plus-0');
DELETE FROM categories WHERE "id" IN ('hardtail', 'cross-country', 'trail', 'e-mtb');
DELETE FROM team WHERE "id" IN ('quang', 'thu', 'bao', 'van');
DELETE FROM services WHERE "id" IN ('build', 'fit', 'service', 'fork', 'bleed', 'tubeless');
DELETE FROM faqs WHERE "id" IN ('size', 'first-bike', 'test', 'delivery', 'pay', 'confirm', 'warranty', 'service');
DELETE FROM shop WHERE "id" IN ('shop');
