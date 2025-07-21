package users

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/mocks"
	"context"
	"database/sql"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
	"regexp"
	"testing"
)

func TestSqlUserDB_GetUser(t *testing.T) {
	tests := []struct {
		testName       string
		userId         string
		dbMock         func(mck sqlmock.Sqlmock)
		err            error
		expectedEntity *entities.User
	}{
		{
			testName: "Successfully getting user",
			userId:   existingUserId,
			dbMock: func(mck sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "role"}).
					AddRow(existingUserId, existingUserEmail, existingUserRole)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUser)).
					WithArgs(existingUserId).
					WillReturnRows(rows)
			},
			expectedEntity: initEntityUser(existingUserId, existingUserEmail, existingUserRole),
		},
		{
			testName: "Failed to get user with non-existing user id",
			userId:   nonExistingUserId,
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUser)).
					WithArgs(nonExistingUserId).
					WillReturnError(sql.ErrNoRows)
			},
			err: errInvalidUserId,
		},
		{
			testName: "Failed to get user with due to unexpected database error",
			userId:   existingUserId,
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUser)).
					WithArgs(existingUserId).
					WillReturnError(assert.AnError)
			},
			err: errUnexpectedDatabaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mck, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.dbMock(mck)

			repo := NewSQLUserDB(db)

			receivedUser, err := repo.GetUser(context.TODO(), test.userId)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedEntity, receivedUser)
			require.NoError(t, mck.ExpectationsWereMet())
		})
	}
}

func TestSqlUserDB_GetUserByEmail(t *testing.T) {
	tests := []struct {
		testName       string
		email          string
		dbMock         func(mck sqlmock.Sqlmock)
		err            error
		expectedEntity *entities.User
	}{
		{
			testName: "Successfully getting user by email",
			email:    existingUserEmail,
			dbMock: func(mck sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "role"}).
					AddRow(existingUserId, existingUserEmail, existingUserRole)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUserByEmail)).
					WithArgs(existingUserEmail).
					WillReturnRows(rows)
			},
			expectedEntity: initEntityUser(existingUserId, existingUserEmail, existingUserRole),
		},

		{
			testName: "Failed to get user by email due to invalid email",
			email:    nonExistingUserEmail,
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUserByEmail)).
					WithArgs(nonExistingUserEmail).
					WillReturnError(sql.ErrNoRows)
			},
			err: errInvalidUserEmail,
		},

		{
			testName: "Failed to get user by email due to database error",
			email:    existingUserEmail,
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUserByEmail)).
					WithArgs(existingUserEmail).
					WillReturnError(assert.AnError)
			},
			err: errUnexpectedDatabaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.dbMock(mock)

			repo := NewSQLUserDB(db)

			receivedEntity, err := repo.GetUserByEmail(context.TODO(), test.email)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedEntity, receivedEntity)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlUserDB_CreateUser(t *testing.T) {
	tests := []struct {
		testName       string
		passedEntity   *entities.User
		dbMock         func(mck sqlmock.Sqlmock)
		err            error
		expectedEntity *entities.User
	}{
		{
			testName:     "Successfully creating user",
			passedEntity: initEntityUser(nonExistingUserId, nonExistingUserEmail, nonExistingUserRole),
			dbMock: func(mck sqlmock.Sqlmock) {

				mck.ExpectExec(regexp.QuoteMeta(sqlQueryCreateUser)).
					WithArgs(nonExistingUserId, nonExistingUserEmail, nonExistingUserRole).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedEntity: initEntityUser(nonExistingUserId, nonExistingUserEmail, nonExistingUserRole),
		},

		{
			testName:     "Failed to create user due to already existing email",
			passedEntity: initEntityUser(nonExistingUserId, existingUserEmail, nonExistingUserRole),
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryCreateUser)).
					WithArgs(nonExistingUserId, existingUserEmail, nonExistingUserRole).WillReturnError(&pq.Error{
					Code:       "23505",
					Constraint: "users_email_key",
				})
			},
			err: errAlreadyExistingEmail,
		},

		{
			testName:     "Failed to create user due to already existing id",
			passedEntity: initEntityUser(existingUserId, nonExistingUserEmail, nonExistingUserRole),
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryCreateUser)).
					WithArgs(existingUserId, nonExistingUserEmail, nonExistingUserRole).WillReturnError(&pq.Error{
					Code:       "23505",
					Constraint: "users_pkey",
				})
			},

			err: errAlreadyExistingUserId,
		},

		{
			testName:     "Failed to create user due to unexpected database error",
			passedEntity: initEntityUser(existingUserId, existingUserEmail, existingUserRole),
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryCreateUser)).
					WithArgs(existingUserId, existingUserEmail, existingUserRole).
					WillReturnError(assert.AnError)
			},

			err: errUnexpectedDatabaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.dbMock(mock)

			repo := NewSQLUserDB(db)

			receivedEntity, err := repo.CreateUser(context.TODO(), test.passedEntity)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedEntity, receivedEntity)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}

}

func TestSqlUserDB_UpdateUser(t *testing.T) {
	tests := []struct {
		testName       string
		id             string
		user           *entities.User
		dbMock         func(mck sqlmock.Sqlmock)
		err            error
		expectedEntity *entities.User
	}{
		{
			testName: "Successfully updating user",
			id:       existingUserId,
			user:     initEntityUser(existingUserId, updatedUserEmail, existingUserRole),
			dbMock: func(mck sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "role"}).
					AddRow(existingUserId, updatedUserEmail, existingUserRole)

				mck.ExpectExec(regexp.QuoteMeta(sqlUpdateUser)).
					WithArgs(existingUserId, updatedUserEmail, existingUserRole, existingUserId).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUser)).
					WithArgs(existingUserId).
					WillReturnRows(rows)
			},
			expectedEntity: initEntityUser(existingUserId, updatedUserEmail, existingUserRole),
		},

		{
			testName: "Failed to update user due to number of affected rows being 0",
			id:       nonExistingUserId,
			user:     initEntityUser(nonExistingUserId, existingUserEmail, existingUserRole),
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlUpdateUser)).
					WithArgs(nonExistingUserId, existingUserEmail, existingUserRole, nonExistingUserId).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			err: errInvalidUserId,
		},

		{
			testName: "Failed to update user due to unexpected database error",
			id:       existingUserId,
			user:     initEntityUser(existingUserId, existingUserEmail, existingUserRole),
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlUpdateUser)).
					WithArgs(existingUserId, existingUserEmail, existingUserRole, existingUserId).
					WillReturnError(assert.AnError)
			},

			err: errUnexpectedDatabaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.dbMock(mock)

			repo := NewSQLUserDB(db)

			receivedEntity, err := repo.UpdateUser(context.TODO(), test.id, test.user)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedEntity, receivedEntity)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}

}

func TestSqlUserDB_DeleteUser(t *testing.T) {
	tests := []struct {
		testName string
		id       string
		dbMock   func(mck sqlmock.Sqlmock)
		err      error
	}{
		{
			testName: "Successfully deleting user",
			id:       existingUserId,
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryDeleteUser)).
					WithArgs(existingUserId).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},

		{
			testName: "Failed to delete user due to unexpected database error",
			id:       existingUserId,
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryDeleteUser)).
					WithArgs(existingUserId).
					WillReturnError(assert.AnError)
			},
			err: errUnexpectedDatabaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.dbMock(mock)

			repo := NewSQLUserDB(db)

			err = repo.DeleteUser(context.TODO(), test.id)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlUserDB_GetTodosAssignedToUser(t *testing.T) {
	tests := []struct {
		testName         string
		sqlRetrieverMock func() *mocks.SqlQueryRetriever
		dbMock           func(mck sqlmock.Sqlmock)
		err              error
		expectedEntities []entities.Todo
	}{
		{
			testName: "Successfully getting todos assigned to user",

			sqlRetrieverMock: func() *mocks.SqlQueryRetriever {
				mck := &mocks.SqlQueryRetriever{}

				mck.EXPECT().DetermineCorrectSqlQuery(context.TODO()).
					Return(sqlQueryGetTodosAssignedToUser).
					Once()
				return mck
			},

			dbMock: func(mck sqlmock.Sqlmock) {
				mockTodos := initMockTodos()

				rows := sqlmock.NewRows([]string{"id", "name", "description", "list_id",
					"status", "created_at", "last_updated", "assigned_to", "due_date", "priority"}).
					AddRow(mockTodos[0].Id.String(), mockTodos[0].Name, mockTodos[0].Description,
						mockTodos[0].ListId.String(), mockTodos[0].Status, mockTodos[0].CreatedAt,
						mockTodos[0].LastUpdated, mockTodos[0].AssignedTo.UUID.String(), mockTodos[0].DueDate, mockTodos[0].Priority).
					AddRow(mockTodos[1].Id.String(), mockTodos[1].Name, mockTodos[1].Description,
						mockTodos[1].ListId.String(), mockTodos[1].Status, mockTodos[1].CreatedAt,
						mockTodos[1].LastUpdated, mockTodos[1].AssignedTo.UUID.String(), mockTodos[1].DueDate, mockTodos[1].Priority)

				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodosAssignedToUser)).
					WillReturnRows(rows)
			},

			expectedEntities: initMockTodos(),
		},

		{
			testName: "Failed to get todos assigned to user due to unexpected database error",

			sqlRetrieverMock: func() *mocks.SqlQueryRetriever {
				mck := &mocks.SqlQueryRetriever{}

				mck.EXPECT().DetermineCorrectSqlQuery(context.TODO()).
					Return(sqlQueryGetTodosAssignedToUser).
					Once()
				return mck
			},

			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetTodosAssignedToUser)).
					WillReturnError(assert.AnError)
			},

			err: errUnexpectedDatabaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mck, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.dbMock(mck)

			sqlRetrieverMock := &mocks.SqlQueryRetriever{}
			if test.sqlRetrieverMock != nil {
				sqlRetrieverMock = test.sqlRetrieverMock()
			}

			repo := NewSQLUserDB(db)

			receivedEntities, err := repo.GetTodosAssignedToUser(context.TODO(), sqlRetrieverMock)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedEntities, receivedEntities)
			require.NoError(t, mck.ExpectationsWereMet())
			mock.AssertExpectationsForObjects(t, sqlRetrieverMock)
		})
	}
}

func TestHandler_HandleGetUserLists(t *testing.T) {
	tests := []struct {
		testName         string
		sqlRetrieverMock func() *mocks.SqlQueryRetriever
		dbMock           func(mck sqlmock.Sqlmock)
		err              error
		expectedEntities []entities.List
	}{
		{
			testName: "Successfully getting user lists",

			sqlRetrieverMock: func() *mocks.SqlQueryRetriever {
				mck := &mocks.SqlQueryRetriever{}

				mck.EXPECT().DetermineCorrectSqlQuery(context.TODO()).
					Return(sqlQueryGetUserLists).
					Once()
				return mck
			},

			dbMock: func(mck sqlmock.Sqlmock) {
				mockLists := initMockLists()

				rows := sqlmock.NewRows([]string{"id", "name", "created_at", "last_updated",
					"owner", "description"}).
					AddRow(mockLists[0].Id, mockLists[0].Name, mockLists[0].CreatedAt,
						mockLists[0].LastUpdated, mockLists[0].Owner, mockLists[0].Description).
					AddRow(mockLists[1].Id, mockLists[1].Name, mockLists[1].CreatedAt,
						mockLists[1].LastUpdated, mockLists[1].Owner, mockLists[1].Description)

				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUserLists)).
					WillReturnRows(rows)
			},

			expectedEntities: initMockLists(),
		},

		{
			testName: "Failed to get user lists due to unexpected database error",

			sqlRetrieverMock: func() *mocks.SqlQueryRetriever {
				mck := &mocks.SqlQueryRetriever{}

				mck.EXPECT().DetermineCorrectSqlQuery(context.TODO()).
					Return(sqlQueryGetUserLists).
					Once()
				return mck
			},

			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUserLists)).
					WillReturnError(assert.AnError)
			},

			err: errUnexpectedDatabaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mck, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.dbMock(mck)

			sqlRetrieverMock := &mocks.SqlQueryRetriever{}
			if test.sqlRetrieverMock != nil {
				sqlRetrieverMock = test.sqlRetrieverMock()
			}

			repo := NewSQLUserDB(db)

			receivedEntities, err := repo.GetUserLists(context.TODO(), sqlRetrieverMock)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedEntities, receivedEntities)
			require.NoError(t, mck.ExpectationsWereMet())
			mock.AssertExpectationsForObjects(t, sqlRetrieverMock)
		})
	}
}

func TestSqlUserDB_GetUsers(t *testing.T) {
	tests := []struct {
		testName         string
		sqlRetrieverMock func() *mocks.SqlQueryRetriever
		dbMock           func(mck sqlmock.Sqlmock)
		err              error
		expectedEntities []entities.User
	}{
		{
			testName: "Successfully getting users",

			sqlRetrieverMock: func() *mocks.SqlQueryRetriever {
				mck := &mocks.SqlQueryRetriever{}

				mck.EXPECT().DetermineCorrectSqlQuery(context.TODO()).
					Return(sqlQueryGetUsers).
					Once()
				return mck
			},

			dbMock: func(mck sqlmock.Sqlmock) {
				mockUsers := initMockUsers()

				rows := sqlmock.NewRows([]string{"id", "email", "role"}).
					AddRow(mockUsers[0].Id, mockUsers[0].Email, mockUsers[0].Role).
					AddRow(mockUsers[1].Id, mockUsers[1].Email, mockUsers[1].Role)

				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUsers)).
					WillReturnRows(rows)
			},

			expectedEntities: initMockUsers(),
		},

		{
			testName: "Failed to get users unexpected database error",

			sqlRetrieverMock: func() *mocks.SqlQueryRetriever {
				mck := &mocks.SqlQueryRetriever{}

				mck.EXPECT().DetermineCorrectSqlQuery(context.TODO()).
					Return(sqlQueryGetUsers).
					Once()
				return mck
			},

			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUsers)).
					WillReturnError(assert.AnError)
			},

			err: errUnexpectedDatabaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mck, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.dbMock(mck)

			sqlRetrieverMock := &mocks.SqlQueryRetriever{}
			if test.sqlRetrieverMock != nil {
				sqlRetrieverMock = test.sqlRetrieverMock()
			}

			repo := NewSQLUserDB(db)

			receivedEntities, err := repo.GetUsers(context.TODO(), sqlRetrieverMock)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedEntities, receivedEntities)
			require.NoError(t, mck.ExpectationsWereMet())
			mock.AssertExpectationsForObjects(t, sqlRetrieverMock)
		})
	}
}

func TestSqlUserDB_DeleteUsers(t *testing.T) {
	tests := []struct {
		testName string
		dbMock   func(mck sqlmock.Sqlmock)
		err      error
	}{
		{
			testName: "Successfully deleting users",
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryDeleteUsers)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},

		{
			testName: "Failed to delete users due to unexpected database error",
			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryDeleteUsers)).
					WillReturnError(assert.AnError)
			},

			err: errUnexpectedDatabaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mck, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.dbMock(mck)

			repo := NewSQLUserDB(db)

			err = repo.DeleteUsers(context.TODO())
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mck.ExpectationsWereMet())
		})
	}
}

func TestSqlUserDB_UpdateUserPartially(t *testing.T) {
	tests := []struct {
		testName       string
		sqlExecParams  map[string]interface{}
		sqlFields      []string
		dbMock         func(mck sqlmock.Sqlmock)
		err            error
		expectedEntity *entities.User
	}{
		{
			testName: "Successfully updating user partially",

			sqlExecParams: map[string]interface{}{
				idKey:    existingUserId,
				emailKey: existingUserEmail,
				roleKey:  existingUserRole,
			},

			sqlFields: []string{
				emailField,
				roleField,
			},

			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryUpdateUserPartially)).
					WithArgs(existingUserEmail, existingUserRole, existingUserId).
					WillReturnResult(sqlmock.NewResult(0, 1))

				rows := sqlmock.NewRows([]string{"id", "email", "role"}).
					AddRow(existingUserId, existingUserEmail, existingUserRole)

				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUser)).
					WithArgs(existingUserId).
					WillReturnRows(rows)
			},

			expectedEntity: initEntityUser(existingUserId, existingUserEmail, existingUserRole),
		},

		{
			testName: "Failed to update user partially due to number of rows being affected being 0",

			sqlExecParams: map[string]interface{}{
				idKey:    nonExistingUserId,
				emailKey: nonExistingUserEmail,
				roleKey:  nonExistingUserRole,
			},

			sqlFields: []string{
				emailField,
				roleField,
			},

			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryUpdateUserPartially)).
					WithArgs(nonExistingUserEmail, nonExistingUserRole, nonExistingUserId).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},

			err: errInvalidUserId,
		},

		{
			testName: "Failed to update user partially due to unexpected database error thrown by update function",

			sqlExecParams: map[string]interface{}{
				idKey:    existingUserId,
				emailKey: existingUserEmail,
				roleKey:  existingUserRole,
			},

			sqlFields: []string{
				emailField,
				roleField,
			},

			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryUpdateUserPartially)).
					WithArgs(nonExistingUserEmail, nonExistingUserRole, nonExistingUserId).
					WillReturnError(assert.AnError)
			},

			err: errUnexpectedDatabaseError,
		},

		{
			testName: "Failed to update user partially due to unexpected database error thrown by get function",

			sqlExecParams: map[string]interface{}{
				idKey:    existingUserId,
				emailKey: existingUserEmail,
				roleKey:  existingUserRole,
			},

			sqlFields: []string{
				emailField,
				roleField,
			},

			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryUpdateUserPartially)).
					WithArgs(existingUserEmail, existingUserRole, existingUserId).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetUser)).
					WithArgs(existingUserId).
					WillReturnError(assert.AnError)
			},

			err: errUnexpectedDatabaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mck, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.dbMock(mck)

			repo := NewSQLUserDB(db)

			receivedEntity, err := repo.UpdateUserPartially(context.TODO(), test.sqlExecParams, test.sqlFields)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedEntity, receivedEntity)
			require.NoError(t, mck.ExpectationsWereMet())
		})
	}

}
