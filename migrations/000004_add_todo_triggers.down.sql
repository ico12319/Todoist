BEGIN;

DROP FUNCTION IF EXISTS set_todo_last_updated();

DROP TRIGGER IF EXISTS todo_before_update ON todos;

END;