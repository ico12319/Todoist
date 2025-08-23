BEGIN;

CREATE OR REPLACE VIEW lists_and_users AS

SELECT  lists.id AS id, name, created_at, last_updated, owner, description, user_lists.user_id AS user_id
FROM lists
LEFT JOIN user_lists ON lists.id = user_lists.list_id;

COMMIT;