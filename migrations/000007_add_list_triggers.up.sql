BEGIN;

CREATE FUNCTION set_list_last_updated()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.last_updated := now();
RETURN NEW;
END;
$$
LANGUAGE plpgsql;


CREATE TRIGGER lists_before_update
    BEFORE UPDATE ON lists
    FOR EACH ROW
    EXECUTE FUNCTION set_list_last_updated();

COMMIT;






