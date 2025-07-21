package users

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"errors"
	"github.com/gofrs/uuid"
	uuid2 "github.com/google/uuid"
	"time"
)

var (
	idKey                          = "id"
	emailKey                       = "email"
	roleKey                        = "role"
	emailField                     = "email = :email"
	roleField                      = "role = :role"
	sqlQueryUpdateUserPartially    = "UPDATE users SET email = ?, role = ? WHERE id = ?"
	sqlQueryDeleteUsers            = "DELETE FROM users"
	sqlQueryGetUsers               = `SELECT id,email,role FROM (SELECT id, email, role FROM users ORDER BY id)`
	sqlQueryGetTodosAssignedToUser = `WITH sorted_todo_cte AS ( 
				   SELECT todos.id, todos.name, todos.description, todos.list_id,
				   todos.status,todos.created_at, todos.last_updated, todos.assigned_to,
		           todos.due_date, todos.priority, users.id AS user_id FROM todos
				   JOIN users ON todos.assigned_to = users.id ORDER BY todos.id
 			      )
				SELECT id, name, description, list_id, status, created_at,
				last_updated, assigned_to, due_date, priority FROM sorted_todo_cte`
	sqlQueryGetUserLists = `WITH sorted_lists_and_users AS(
				   SELECT * FROM lists LEFT JOIN user_lists ON
  				   lists.id = user_lists.list_id
  				  )
 				SELECT id, name, created_at, last_updated, owner, description FROM sorted_lists_and_users`
	sqlQueryDeleteUser         = "DELETE FROM users WHERE id = $1"
	sqlQueryGetUser            = "SELECT id,email,role FROM users WHERE id = $1"
	sqlQueryCreateUser         = "INSERT INTO users (id, email, role) VALUES (?,?,?)"
	sqlUpdateUser              = "UPDATE users SET (id, email, role) = ($1, $2, $3) WHERE id = $4"
	sqlQueryGetUserByEmail     = "SELECT id, email, role FROM users WHERE email = $1"
	existingUserId             = uuid2.New().String()
	existingUserId2            = uuid2.New().String()
	nonExistingUserId          = uuid2.New().String()
	todo1Id                    = uuid2.New().String()
	todo2Id                    = uuid2.New().String()
	list1Id                    = uuid2.New().String()
	list2Id                    = uuid2.New().String()
	existingUserEmail          = "existinguser@test.test"
	existingUserEmail2         = "existinguser2@test.test"
	nonExistingUserEmail       = "nonExistinguser@test.test"
	updatedUserEmail           = "updatedEmail@test.test"
	existingUserRole           = "admin"
	nonExistingUserRole        = "reader"
	errInvalidUserId           = application_errors.NewNotFoundError(constants.USER_TARGET, nonExistingUserId)
	errInvalidUserEmail        = application_errors.NewNotFoundError(constants.USER_TARGET, nonExistingUserEmail)
	errUnexpectedDatabaseError = errors.New("unexpected database error")
	errAlreadyExistingEmail    = application_errors.NewAlreadyExistError(constants.USER_TARGET, existingUserEmail)
	errAlreadyExistingUserId   = application_errors.NewAlreadyExistError(constants.USER_TARGET, existingUserId)
	testDate                   = time.Date(2025, time.January, 15, 10, 30, 0, 0, time.UTC)
)

func initEntityUser(id string, email string, role string) *entities.User {
	return &entities.User{
		Id:    uuid.FromStringOrNil(id),
		Email: email,
		Role:  role,
	}
}

func initMockTodos() []entities.Todo {
	return []entities.Todo{
		{
			Id:          uuid.FromStringOrNil(todo1Id),
			Name:        "name1",
			Description: "desc1",
			ListId:      uuid.FromStringOrNil(list1Id),
			Status:      "status1",
			CreatedAt:   testDate,
			LastUpdated: testDate,
			AssignedTo: uuid.NullUUID{
				UUID:  uuid.FromStringOrNil(existingUserId),
				Valid: true,
			},
			Priority: "priority1",
		},
		{
			Id:          uuid.FromStringOrNil(todo2Id),
			Name:        "name2",
			Description: "desc2",
			ListId:      uuid.FromStringOrNil(list2Id),
			Status:      "status2",
			CreatedAt:   testDate,
			LastUpdated: testDate,
			AssignedTo: uuid.NullUUID{
				UUID:  uuid.FromStringOrNil(existingUserId),
				Valid: true,
			},
			Priority: "priority2",
		},
	}
}

func initMockLists() []entities.List {
	return []entities.List{
		{
			Id:          uuid.FromStringOrNil(list1Id),
			Name:        "list1",
			CreatedAt:   testDate,
			LastUpdated: testDate,
			Owner:       uuid.FromStringOrNil(existingUserId),
			Description: "desc1",
		},
		{
			Id:          uuid.FromStringOrNil(list2Id),
			Name:        "list2",
			CreatedAt:   testDate,
			LastUpdated: testDate,
			Owner:       uuid.FromStringOrNil(existingUserId),
			Description: "desc2",
		},
	}
}

func initMockUsers() []entities.User {
	return []entities.User{
		{
			Id:    uuid.FromStringOrNil(existingUserId),
			Email: existingUserEmail,
			Role:  "reader",
		},
		{
			Id:    uuid.FromStringOrNil(existingUserId2),
			Email: existingUserEmail2,
			Role:  "admin",
		},
	}
}
