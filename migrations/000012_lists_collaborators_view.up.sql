BEGIN ;

CREATE OR REPLACE VIEW lists_collaborators
AS

SELECT users.id AS id,users.email,users.role,list_id FROM users
JOIN user_lists ON users.id = user_lists.user_id;

COMMIT ;