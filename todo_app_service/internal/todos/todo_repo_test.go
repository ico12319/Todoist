package todos

import (
	application_errors2 "Todo-List/internProject/todo_app_service/internal/application_errors"
	entities2 "Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"database/sql"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
	"regexp"
	"testing"
)

func TestSqlTodoDB_GetTodo(t *testing.T) {
	tests := []struct {
		testName       string
		listId         string
		todoId         string
		mockSetup      func(mck sqlmock.Sqlmock)
		expectError    bool
		err            error
		expectedEntity *entities2.Todo
	}{
		{
			testName: "Successfully returns expected todo entity",
			listId:   nonExistingListId.String(),
			todoId:   nonExistingTodoId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "list_id",
					"status", "created_at", "last_updated", "assigned_to", "due_date", "priority"}).
					AddRow(nonExistingTodoId, todoName, todoDescription, nonExistingListId, todoStatus, testDate, testDate,
						assigneeNullId, testNullDate, priority)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodo)).WithArgs(nonExistingListId, nonExistingTodoId).
					WillReturnRows(rows)
			},
			expectError: false,
			err:         nil,
			expectedEntity: initTodoEntity(nonExistingTodoId, todoName, todoDescription, nonExistingListId, todoStatus, testDate,
				testDate, assigneeNullId, testNullDate, priority),
		},
		{
			testName: "Unable to get todo because of invalid todo_id",
			listId:   nonExistingListId.String(),
			todoId:   existingTodoId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodo)).WithArgs(nonExistingListId, existingTodoId).
					WillReturnError(sql.ErrNoRows)
			},
			expectError:    true,
			err:            application_errors2.InvalidListOrTodoIdError,
			expectedEntity: nil,
		},
		{
			testName: "Unable to get todo because of invalid list_id",
			listId:   existingListId.String(),
			todoId:   nonExistingTodoId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodo)).WithArgs(existingListId, nonExistingTodoId).
					WillReturnError(sql.ErrNoRows)
			},
			expectError:    true,
			err:            application_errors2.InvalidListOrTodoIdError,
			expectedEntity: nil,
		},
		{
			testName: "Unable to get todo because of a database error",
			listId:   existingListId.String(),
			todoId:   existingTodoId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodo)).WithArgs(existingListId, existingTodoId).
					WillReturnError(databaseError)
			},
			expectError:    true,
			err:            databaseError,
			expectedEntity: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.mockSetup(mock)

			todoRepo := NewSQLTodoDB(db)
			gotTodo, err := todoRepo.GetTodo(test.listId, test.todoId)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedEntity, gotTodo)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlTodoDB_GetTodosByListId(t *testing.T) {
	tests := []struct {
		testName         string
		listId           string
		mockSetup        func(mck sqlmock.Sqlmock)
		expectError      bool
		err              error
		expectedEntities []entities2.Todo
	}{
		{
			testName: "Successfully returns expected entities",
			listId:   nonExistingListId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "list_id",
					"status", "created_at", "last_updated", "assigned_to", "due_date", "priority"}).
					AddRow(nonExistingTodoId, todoName, todoDescription, nonExistingListId, todoDoneStatus, testDate, testDate,
						assigneeNullId, testNullDate, veryLowPriority).
					AddRow(nonExistingTodoId2, todoName, todoDescription, nonExistingListId, todoInProgressStatus, testDate, testDate,
						assigneeNullId, testNullDate, veryHighPriority)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodos)).WithArgs(nonExistingListId).WillReturnRows(rows)
			},
			expectError: false,
			err:         nil,
			expectedEntities: []entities2.Todo{
				*initTodoEntity(nonExistingTodoId, todoName, todoDescription, nonExistingListId,
					todoDoneStatus, testDate, testDate, assigneeNullId, testNullDate, veryLowPriority),
				*initTodoEntity(nonExistingTodoId2, todoName, todoDescription, nonExistingListId,
					todoInProgressStatus, testDate, testDate, assigneeNullId, testNullDate, veryHighPriority),
			},
		},
		{
			testName: "Unable to get entities because of database error",
			listId:   existingListId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodos)).WithArgs(existingListId).
					WillReturnError(databaseError)
			},
			expectError: true,
			err:         databaseError,
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.mockSetup(mock)
			todoRepo := NewSQLTodoDB(db)

			gotEntities, err := todoRepo.GetTodosByListId(test.listId)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedEntities, gotEntities)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlTodoDB_GetTodosByStatus(t *testing.T) {
	tests := []struct {
		testName         string
		listId           string
		status           string
		mockSetup        func(mck sqlmock.Sqlmock)
		expectError      bool
		err              error
		expectedEntities []entities2.Todo
	}{
		{
			testName: "Successfully returns expected entities",
			listId:   nonExistingListId.String(),
			status:   todoDoneStatus,
			mockSetup: func(mck sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "list_id",
					"status", "created_at", "last_updated", "assigned_to", "due_date", "priority"}).
					AddRow(nonExistingTodoId, todoName, todoDescription, nonExistingListId, todoDoneStatus, testDate, testDate,
						assigneeNullId, testNullDate, veryHighPriority).
					AddRow(nonExistingTodoId2, todoName, todoDescription, nonExistingListId, todoDoneStatus, testDate, testDate,
						assigneeNullId, testNullDate, veryLowPriority)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodosByStatus)).WithArgs(nonExistingListId, todoDoneStatus).WillReturnRows(rows)
			},
			expectError: false,
			err:         nil,
			expectedEntities: []entities2.Todo{
				*initTodoEntity(nonExistingTodoId, todoName, todoDescription, nonExistingListId,
					todoDoneStatus, testDate, testDate, assigneeNullId, testNullDate, veryHighPriority),
				*initTodoEntity(nonExistingTodoId2, todoName, todoDescription, nonExistingListId,
					todoDoneStatus, testDate, testDate, assigneeNullId, testNullDate, veryLowPriority),
			},
		},
		{
			testName: "Unable to get entities because of database error",
			listId:   existingListId.String(),
			status:   todoStatus,
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodosByStatus)).WithArgs(existingListId, todoStatus).
					WillReturnError(databaseError)
			},
			expectError:      true,
			err:              databaseError,
			expectedEntities: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.mockSetup(mock)
			todoRepo := NewSQLTodoDB(db)

			gotEntities, err := todoRepo.GetTodosByStatus(test.listId, test.status)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedEntities, gotEntities)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlTodoDB_GetTodosByPriority(t *testing.T) {
	tests := []struct {
		testName         string
		listId           string
		priority         string
		mockSetup        func(mck sqlmock.Sqlmock)
		expectError      bool
		err              error
		expectedEntities []entities2.Todo
	}{
		{
			testName: "Successfully returns expected entities",
			listId:   nonExistingListId.String(),
			priority: veryLowPriority,
			mockSetup: func(mck sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "list_id",
					"status", "created_at", "last_updated", "assigned_to", "due_date", "priority"}).
					AddRow(nonExistingTodoId, todoName, todoDescription, nonExistingListId, todoInProgressStatus, testDate, testDate,
						assigneeNullId, testNullDate, veryLowPriority).
					AddRow(nonExistingTodoId2, todoName, todoDescription, nonExistingListId, todoStatus, testDate, testDate,
						assigneeNullId, testNullDate, veryLowPriority)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodosByPriority)).WithArgs(nonExistingListId, veryLowPriority).WillReturnRows(rows)
			},
			expectError: false,
			err:         nil,
			expectedEntities: []entities2.Todo{
				*initTodoEntity(nonExistingTodoId, todoName, todoDescription, nonExistingListId,
					todoInProgressStatus, testDate, testDate, assigneeNullId, testNullDate, veryLowPriority),
				*initTodoEntity(nonExistingTodoId2, todoName, todoDescription, nonExistingListId,
					todoStatus, testDate, testDate, assigneeNullId, testNullDate, veryLowPriority),
			},
		},
		{
			testName: "Unable to get entities because of database error",
			listId:   existingListId.String(),
			priority: priority,
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodosByPriority)).WithArgs(existingListId, priority).
					WillReturnError(databaseError)
			},
			expectError: true,
			err:         databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.mockSetup(mock)
			todoRepo := NewSQLTodoDB(db)

			gotEntities, err := todoRepo.GetTodosByPriority(test.listId, test.priority)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedEntities, gotEntities)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlTodoDB_GetTodosByStatusAndPriority(t *testing.T) {
	tests := []struct {
		testName         string
		listId           string
		status           string
		priority         string
		mockSetup        func(mck sqlmock.Sqlmock)
		expectError      bool
		err              error
		expectedEntities []entities2.Todo
	}{
		{
			testName: "Successfully returns expected entities",
			listId:   nonExistingListId.String(),
			status:   todoInProgressStatus,
			priority: mediumPriority,
			mockSetup: func(mck sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "description", "list_id",
					"status", "created_at", "last_updated", "assigned_to", "due_date", "priority"}).
					AddRow(nonExistingTodoId, todoName, todoDescription, nonExistingListId, todoInProgressStatus, testDate, testDate,
						assigneeNullId, testNullDate, mediumPriority).
					AddRow(nonExistingTodoId2, todoName, todoDescription, nonExistingListId, todoInProgressStatus, testDate, testDate,
						assigneeNullId, testNullDate, mediumPriority)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodosByStatusAndPriority)).WithArgs(nonExistingListId, todoInProgressStatus, mediumPriority).
					WillReturnRows(rows)
			},
			expectError: false,
			err:         nil,
			expectedEntities: []entities2.Todo{
				*initTodoEntity(nonExistingTodoId, todoName, todoDescription, nonExistingListId,
					todoInProgressStatus, testDate, testDate, assigneeNullId, testNullDate, mediumPriority),
				*initTodoEntity(nonExistingTodoId2, todoName, todoDescription, nonExistingListId,
					todoInProgressStatus, testDate, testDate, assigneeNullId, testNullDate, mediumPriority),
			},
		},
		{
			testName: "Unable to get entities because of database error",
			listId:   existingListId.String(),
			status:   todoStatus,
			priority: priority,
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodosByStatusAndPriority)).WithArgs(existingListId, todoStatus, priority).
					WillReturnError(databaseError)
			},
			expectError: true,
			err:         databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.mockSetup(mock)
			todoRepo := NewSQLTodoDB(db)

			gotEntities, err := todoRepo.GetTodosByStatusAndPriority(test.listId, test.status, test.priority)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedEntities, gotEntities)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlTodoDB_GetTodoAssigneeTo(t *testing.T) {
	tests := []struct {
		testName         string
		todoId           string
		mockSetup        func(mck sqlmock.Sqlmock)
		expectError      bool
		err              error
		expectedAssignee *entities2.User
	}{
		{
			testName: "Successfully returning todo assignee",
			todoId:   existingTodoId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "role"}).
					AddRow(assigneeId, email, writerRole)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodoAssignee)).WithArgs(existingTodoId).
					WillReturnRows(rows)
			},
			expectError:      false,
			err:              nil,
			expectedAssignee: initUserEntity(assigneeId, email, writerRole),
		},
		{
			testName: "Unable to get todo assignee because of invalid todo_id",
			todoId:   nonExistingTodoId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodoAssignee)).WithArgs(nonExistingTodoId).
					WillReturnError(sql.ErrNoRows)
			},
			expectError:      true,
			err:              application_errors2.NewNotFoundError(constants.TODO_TARGET, nonExistingTodoId.String()),
			expectedAssignee: nil,
		},
		{
			testName: "Unable to get todo assignee because of invalid todo_id",
			todoId:   nonExistingTodoId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodoAssignee)).WithArgs(nonExistingTodoId).
					WillReturnError(databaseError)
			},
			expectError:      true,
			err:              databaseError,
			expectedAssignee: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.mockSetup(mock)
			todoRepo := NewSQLTodoDB(db)
			assignee, err := todoRepo.GetTodoAssigneeTo(test.todoId)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedAssignee, assignee)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlTodoDB_CreateTodo(t *testing.T) {
	tests := []struct {
		testName       string
		passedEntity   *entities2.Todo
		mockSetup      func(mck sqlmock.Sqlmock)
		expectError    bool
		err            error
		expectedEntity *entities2.Todo
	}{
		{
			testName: "Successfully creates todo",
			passedEntity: initTodoEntity(nonExistingTodoId, todoName, todoDescription,
				nonExistingListId, todoStatus, testDate, testDate, assigneeNullId, testNullDate, priority),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryCreateTodo)).WithArgs(nonExistingTodoId, todoName, todoDescription,
					nonExistingListId, todoStatus, testDate, testDate, assigneeNullId, testNullDate, priority).
					WillReturnResult(sqlmock.NewResult(0, 1))

				rows := sqlmock.NewRows([]string{"id", "name", "description", "list_id", "status",
					"created_at", "last_updated", "assigned_to", "due_date", "priority"}).
					AddRow(nonExistingTodoId, todoName, todoDescription,
						nonExistingListId, todoStatus, testDate, testDate, assigneeNullId, testNullDate, priority)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodo)).WithArgs(nonExistingListId, nonExistingTodoId).
					WillReturnRows(rows)
			},
			expectError: false,
			err:         nil,
			expectedEntity: initTodoEntity(nonExistingTodoId, todoName, todoDescription,
				nonExistingListId, todoStatus, testDate, testDate, assigneeNullId, testNullDate, priority),
		},
		{
			testName: "Unable to create todo because of invalid todo_id",
			passedEntity: initTodoEntity(existingTodoId, todoName, todoDescription,
				nonExistingListId, todoStatus, testDate, testDate, assigneeNullId, testNullDate, priority),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryCreateTodo)).WithArgs(existingTodoId, todoName, todoDescription,
					nonExistingListId, todoStatus, testDate, testDate, assigneeNullId, testNullDate, priority).
					WillReturnError(&pq.Error{
						Code: "23505",
					})
			},
			expectError:    true,
			err:            application_errors2.NewNotFoundError(constants.TODO_TARGET, existingTodoId.String()),
			expectedEntity: nil,
		},
		{
			testName: "Unable to create todo because of invalid list_id",
			passedEntity: initTodoEntity(nonExistingTodoId, todoName, todoDescription,
				existingListId, todoStatus, testDate, testDate, assigneeNullId, testNullDate, priority),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryCreateTodo)).WithArgs(nonExistingTodoId, todoName, todoDescription,
					existingListId, todoStatus, testDate, testDate, assigneeNullId, testNullDate, priority).
					WillReturnError(&pq.Error{
						Code:       "23503",
						Constraint: "todos_list_id_fkey",
					})
			},
			expectError:    true,
			err:            application_errors2.NewNotFoundError(constants.LIST_TARGET, existingListId.String()),
			expectedEntity: nil,
		},
		{
			testName: "Unable to create todo because of invalid assignee_id",
			passedEntity: initTodoEntity(nonExistingTodoId, todoName, todoDescription, nonExistingListId,
				todoStatus, testDate, testDate, nonExistingAssigneeNullId, testNullDate, priority),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryCreateTodo)).WithArgs(nonExistingTodoId, todoName, todoDescription,
					nonExistingListId, todoStatus, testDate, testDate, nonExistingAssigneeNullId, testNullDate, priority).
					WillReturnError(&pq.Error{
						Code:       "23503",
						Constraint: "todos_assigned_to_fkey",
					})
			},
			expectError:    true,
			err:            application_errors2.NewNotFoundError(constants.USER_TARGET, nonExistingAssigneeNullId.UUID.String()),
			expectedEntity: nil,
		},
		{
			testName: "Unable to create todo because of database error",
			passedEntity: initTodoEntity(nonExistingTodoId, todoName, todoDescription, nonExistingListId,
				todoStatus, testDate, testDate, nonExistingAssigneeNullId, testNullDate, priority),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryCreateTodo)).WithArgs(nonExistingTodoId, todoName, todoDescription,
					nonExistingListId, todoStatus, testDate, testDate, nonExistingAssigneeNullId, testNullDate, priority).
					WillReturnError(databaseError)
			},
			expectError:    true,
			err:            databaseError,
			expectedEntity: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.mockSetup(mock)
			todoRepo := NewSQLTodoDB(db)
			gotEntity, err := todoRepo.CreateTodo(test.passedEntity)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedEntity, gotEntity)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlTodoDB_DeleteTodo(t *testing.T) {
	tests := []struct {
		testName    string
		listId      string
		todoId      string
		mockSetup   func(mck sqlmock.Sqlmock)
		expectError bool
		err         error
	}{
		{
			testName: "Successfully deletes todo",
			listId:   existingListId.String(),
			todoId:   existingTodoId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryDeleteTodo)).WithArgs(existingListId, existingTodoId).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
			err:         nil,
		},
		{
			testName: "Unable to delete todo because of a database error",
			listId:   existingListId.String(),
			todoId:   existingTodoId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryDeleteTodo)).WithArgs(existingListId, existingTodoId).
					WillReturnError(databaseError)
			},
			expectError: true,
			err:         databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.mockSetup(mock)
			todoRepo := NewSQLTodoDB(db)
			err = todoRepo.DeleteTodo(test.listId, test.todoId)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlTodoDB_UpdateTodo(t *testing.T) {
	tests := []struct {
		testName     string
		passedEntity *entities2.Todo
		mockSetup    func(mck sqlmock.Sqlmock)
		expectError  bool
		err          error
	}{
		{
			testName: "Successfully updating todo",
			passedEntity: initTodoEntity(existingTodoId, todoName, todoDescription,
				existingListId, todoStatus, testDate, testDate, assigneeNullId, testNullDate, priority),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryUpdateTodo)).WithArgs(todoName, todoDescription, todoStatus,
					testDate, assigneeNullId, testNullDate, priority, existingTodoId, existingListId).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
			err:         nil,
		},
		{
			testName: "Unable to update todo because of invalid todo_id or list_id",
			passedEntity: initTodoEntity(nonExistingTodoId, todoName, todoDescription,
				nonExistingListId, todoStatus, testDate, testDate, assigneeNullId, testNullDate, priority),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryUpdateTodo)).WithArgs(todoName, todoDescription, todoStatus,
					testDate, assigneeNullId, testNullDate, priority, nonExistingTodoId, nonExistingListId).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectError: true,
			err:         application_errors2.InvalidListOrTodoIdError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.mockSetup(mock)
			todoRepo := NewSQLTodoDB(db)
			err = todoRepo.UpdateTodo(test.passedEntity)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
