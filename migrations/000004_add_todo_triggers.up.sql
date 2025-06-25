BEGIN;

CREATE FUNCTION set_todo_last_updated()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.last_updated := now();
RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER todo_before_update
    BEFORE UPDATE ON todos
    FOR EACH ROW
    EXECUTE FUNCTION set_todo_last_updated();

COMMIT;