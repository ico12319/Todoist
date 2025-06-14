package todos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	entities2 "github.com/I763039/Todo-List/internProject/todo_app_service/internal/entities"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	email                = "email@email.com"
	writerRole           = "writer"
	readerRole           = "reader"
	adminRole            = "admin"
	dbError              = "database error"
	todoName             = "todoName"
	todoDescription      = "description"
	todoStatus           = "open"
	todoDoneStatus       = "done"
	todoInProgressStatus = "in progress"
	priority             = "low"
	highPriority         = "high"
	veryHighPriority     = "very high"
	veryLowPriority      = "very low"
	mediumPriority       = "medium"
	sqlQueryGetTodo      = `SELECT id, name, description, list_id, status,
created_at, last_updated, assigned_to, due_date, priority FROM todos
WHERE list_id = $1 AND id = $2`
	sqlQueryGetTodos = `SELECT id, name, description, list_id, status, 
       					created_at, last_updated, assigned_to, due_date, priority FROM todos
       					WHERE list_id = $1`
	sqlQueryGetTodosByStatus = `SELECT id, name, description, list_id, status, created_at, 
       last_updated, assigned_to, due_date, priority FROM todos WHERE list_id = $1 AND status = $2`
	sqlQueryGetTodosByPriority = `SELECT id, name, description, list_id, status, created_at, last_updated, 
       assigned_to, due_date, priority FROM todos WHERE list_id = $1 AND priority = $2`
	sqlQueryGetTodosByStatusAndPriority = `SELECT id, name, description, list_id, status, created_at, 
       last_updated, assigned_to, due_date, priority FROM todos WHERE list_id = $1 AND status = $2 
       AND priority = $3`
	sqlQueryGetTodoAssignee = `SELECT users.id,users.email,users.role FROM users JOIN todos on users.id = todos.assigned_to
						WHERE todos.id = $1`
	sqlQueryCreateTodo = `INSERT INTO todos(id, name, description, 
                  list_id, status, created_at, last_updated, assigned_to, due_date, priority) VALUES(?,?,?,
                  ?,?,?,?,?,?,?)`
	sqlQueryDeleteTodo = `DELETE FROM todos WHERE list_id = $1 AND id = $2`
	sqlQueryUpdateTodo = `UPDATE todos SET (name,description,status,assigned_to,due_date,priority) = (?,?,?,
                 	 	?,?,?) 
             			WHERE id = ? AND list_id = ?`
)

var (
	nonExistingListId         = uuid.Must(uuid.NewV4())
	nonExistingTodoId         = uuid.Must(uuid.NewV4())
	nonExistingTodoId2        = uuid.Must(uuid.NewV4())
	existingListId            = uuid.Must(uuid.NewV4())
	existingTodoId            = uuid.Must(uuid.NewV4())
	testDate                  = time.Date(2025, time.January, 15, 10, 30, 0, 0, time.UTC)
	assigneeId                = uuid.Must(uuid.NewV4())
	nonExistingAssigneeId     = uuid.Must(uuid.NewV4())
	nonExistingAssigneeNullId = uuid.NullUUID{UUID: nonExistingAssigneeId, Valid: true}
	assigneeNullId            = uuid.NullUUID{UUID: assigneeId, Valid: true}
	testNullDate              = sql.NullTime{Time: testDate, Valid: true}
	databaseError             = fmt.Errorf(dbError)
)

func convertToStringPointer(str string) *string {
	return &str
}

func errorMatchHelper(tb testing.TB, rr *httptest.ResponseRecorder, err error) {
	tb.Helper()
	var got map[string]string
	require.NoError(tb, json.Unmarshal(rr.Body.Bytes(), &got))
	expect := map[string]string{
		"error": err.Error(),
	}
	require.Equal(tb, expect, got)
}

func initTodoEntity(id uuid.UUID, name string, description string, listId uuid.UUID, status string,
	createdAt time.Time, lastUpdated time.Time, assignedTo uuid.NullUUID, dueDate sql.NullTime, priority string) *entities2.Todo {
	return &entities2.Todo{
		Id:          id,
		Name:        name,
		Description: description,
		ListId:      listId,
		Status:      status,
		CreatedAt:   createdAt,
		LastUpdated: lastUpdated,
		AssignedTo:  assignedTo,
		DueDate:     dueDate,
		Priority:    priority,
	}
}

func initUserEntity(id uuid.UUID, email string, role string) *entities2.User {
	return &entities2.User{
		Id:    id,
		Email: email,
		Role:  role,
	}
}
