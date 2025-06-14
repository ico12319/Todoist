ALTER TABLE todos
ADD CONSTRAINT chk_due_after_start
CHECK (due_date > created_at);