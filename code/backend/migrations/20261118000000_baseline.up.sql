-- Baseline. It changes nothing: it is the first row in schema_migrations, so a
-- booted backend proves the runner end to end, and it gives the embed
-- directive a file to hold. The tables arrive with the migrations of the
-- stories that need them, each in a file stamped later than this one.
SELECT 1;
