BEGIN ;

CREATE OR REPLACE VIEW user_todos
AS

SELECT todos.id AS id, todos.name, todos.description, todos.list_id,
todos.status,todos.created_at, todos.last_updated, todos.assigned_to,
todos.due_date, todos.priority, users.id AS user_id FROM todos
JOIN users ON todos.assigned_to = users.id;


COMMIT;