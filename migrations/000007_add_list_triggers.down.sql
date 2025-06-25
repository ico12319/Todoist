BEGIN;

DROP TRIGGER IF EXISTS lists_before_update
ON lists;

DROP FUNCTION IF EXISTS set_list_last_updated();

COMMIT;