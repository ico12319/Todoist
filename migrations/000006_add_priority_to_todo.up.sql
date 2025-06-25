BEGIN TRANSACTION;

CREATE TYPE todo_priority AS ENUM('very low','low','medium','high','very high');

ALTER TABLE todos
ADD COLUMN priority todo_priority NOT NULL
DEFAULT 'low';

END;