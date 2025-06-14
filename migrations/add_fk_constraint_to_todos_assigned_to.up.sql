ALTER TABLE todos
ADD CONSTRAINT todos_assigned_to_fkey
FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL;

