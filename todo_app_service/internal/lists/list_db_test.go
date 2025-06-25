package lists

import (
	application_errors2 "Todo-List/internProject/todo_app_service/internal/application_errors"
	entities2 "Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"database/sql"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
	"regexp"
	"testing"
)

func TestSqlListDB_GetList(t *testing.T) {
	tests := []struct {
		testName       string
		listId         string
		mockSetup      func(mck sqlmock.Sqlmock)
		expectError    bool
		err            error
		expectedEntity *entities2.List
	}{
		{
			testName: "Existing list id passed so the function should return expectedEntity",
			listId:   existingListId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				rows := mck.NewRows([]string{"id", "name", "created_at", "last_updated", "owner"}).
					AddRow(existingListId, listName, testDate, testDate, ownerId)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetList)).WithArgs(existingListId.String()).
					WillReturnRows(rows)
			},
			expectError:    false,
			err:            nil,
			expectedEntity: initEntityList(existingListId, listName, testDate, testDate, ownerId),
		},
		{
			testName: "Non existing list id passed so the function returns error and nil list entity",
			listId:   nonExistingListId.String(),
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(sqlQueryGetList)).WithArgs(nonExistingListId.String()).
					WillReturnError(sql.ErrNoRows)
			},
			expectError:    true,
			err:            application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String()),
			expectedEntity: nil,
		},
		{
			testName: "Db error so the function returns error and nil list entity",
			listId:   nonExistingListId.String(),
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(sqlQueryGetList)).WithArgs(nonExistingListId.String()).
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
			listRepo := NewSQLListDB(db)

			test.mockSetup(mock)

			got, err := listRepo.GetList(test.listId)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedEntity, got)
			require.NoError(t, mock.ExpectationsWereMet())

		})
	}
}

func TestSqlListDB_DeleteList(t *testing.T) {
	tests := []struct {
		testName    string
		listId      string
		mockSetup   func(mck sqlmock.Sqlmock)
		expectError bool
	}{
		{
			testName: "Successfully deletes data",
			listId:   existingListId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryDeleteList)).WithArgs(existingListId.String()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			testName: "Fails because deleting from the list table results in an error",
			listId:   existingListId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryDeleteList)).WithArgs(existingListId).
					WillReturnError(assert.AnError)
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			listRepo := NewSQLListDB(db)

			test.mockSetup(mock)

			err = listRepo.DeleteList(test.listId)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlListDB_GetLists(t *testing.T) {
	tests := []struct {
		testName         string
		mockSetup        func(mck sqlmock.Sqlmock)
		expectError      bool
		expectedEntities []entities2.List
	}{
		{
			testName: "Successfully retrieving entities from the list table",
			mockSetup: func(mck sqlmock.Sqlmock) {
				rows := mck.NewRows([]string{"id", "name", "created_at", "last_updated", "owner"}).
					AddRow(existingListId, listName, testDate, testDate, ownerId).
					AddRow(dummyListId, listName, testDate2, testDate2, dummyListOwner)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetLists)).WillReturnRows(rows)
			},
			expectError: false,
			expectedEntities: []entities2.List{
				*initEntityList(existingListId, listName, testDate, testDate, ownerId),
				*initEntityList(dummyListId, listName, testDate2, testDate2, dummyListOwner),
			},
		},
		{
			testName: "Unable to retrieve entities from the list table because of database error",
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(sqlQueryGetLists).WillReturnError(assert.AnError)
			},
			expectError:      true,
			expectedEntities: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			listRepo := NewSQLListDB(db)

			test.mockSetup(mock)

			gotEntities, err := listRepo.GetLists()

			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedEntities, gotEntities)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlListDB_UpdateListName(t *testing.T) {
	tests := []struct {
		testName    string
		listId      string
		newName     string
		mockSetup   func(mck sqlmock.Sqlmock)
		expectError bool
		err         error
	}{
		{
			testName: "Successfully updating list name",
			listId:   existingListId.String(),
			newName:  newListName,
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryUpdateListName)).WithArgs(newListName, existingListId.String()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
			err:         nil,
		},
		{
			testName: "Unable to update list name because of a database error",
			listId:   existingListId.String(),
			newName:  newListName,
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryUpdateListName)).WithArgs(newListName, existingListId.String()).
					WillReturnError(databaseError)
			},
			expectError: true,
			err:         databaseError,
		},
		{
			testName: "Unable to update list name because of a invalid list_id provided",
			listId:   nonExistingListId.String(),
			newName:  newListName,
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryUpdateListName)).WithArgs(newListName, nonExistingListId.String()).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectError: true,
			err:         application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String()),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			listRepo := NewSQLListDB(db)
			test.mockSetup(mock)

			err = listRepo.UpdateListName(test.listId, test.newName)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())

		})
	}
}

func TestSqlListDB_UpdateListSharedWith(t *testing.T) {
	tests := []struct {
		testName    string
		listId      string
		userId      string
		mockSetup   func(mck sqlmock.Sqlmock)
		expectError bool
		err         error
	}{
		{
			testName: "Successfully updating list's shared with",
			listId:   existingListId.String(),
			userId:   userId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlInsertIntoUserListsTableQuery)).
					WithArgs(userId, existingListId).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
			err:         nil,
		},
		{
			testName: "Unable to update list's shared with because of invalid list_id",
			listId:   nonExistingListId.String(),
			userId:   userId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlInsertIntoUserListsTableQuery)).
					WithArgs(userId, nonExistingListId).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectError: true,
			err:         application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String()),
		},
		{
			testName: "Unable to update list's shared with because of error in the database",
			listId:   existingListId.String(),
			userId:   userId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlInsertIntoUserListsTableQuery)).
					WithArgs(userId, existingListId).WillReturnError(databaseError)
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

			listRepo := NewSQLListDB(db)
			test.mockSetup(mock)

			err = listRepo.UpdateListSharedWith(test.listId, test.userId)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlListDB_GetListCollaborators(t *testing.T) {
	tests := []struct {
		testName              string
		listId                string
		mockSetup             func(mck sqlmock.Sqlmock)
		expectError           bool
		err                   error
		expectedCollaborators []entities2.User
	}{
		{
			testName: "Successfully getting list's collaborators",
			listId:   existingListId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				rows := mck.NewRows([]string{"id", "email", "role"}).
					AddRow(userId, userEmail, adminRole).
					AddRow(userId2, userEmail2, writerRole).
					AddRow(userId3, userEmail3, readerRole)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetCollaborator)).WithArgs(existingListId.String()).
					WillReturnRows(rows)
			},
			expectError: false,
			err:         nil,
			expectedCollaborators: []entities2.User{
				*initEntityUser(userId, userEmail, adminRole),
				*initEntityUser(userId2, userEmail2, writerRole),
				*initEntityUser(userId3, userEmail3, readerRole),
			},
		},
		{
			testName: "Unable to get list's collaborators because of database error",
			listId:   existingListId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetCollaborator)).WithArgs(existingListId.String()).
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

			listRepo := NewSQLListDB(db)
			test.mockSetup(mock)

			gotCollaborators, err := listRepo.GetListCollaborators(test.listId)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedCollaborators, gotCollaborators)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSqlListDB_CreateList(t *testing.T) {
	tests := []struct {
		testName       string
		passedEntity   *entities2.List
		mockSetup      func(mck sqlmock.Sqlmock)
		expectError    bool
		err            error
		expectedEntity *entities2.List
	}{
		{
			testName:     "Successfully creates list and returns it",
			passedEntity: initEntityList(nonExistingListId, listName, testDate, testDate, ownerId),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryInsertList)).
					WithArgs(nonExistingListId, listName, testDate, testDate, ownerId).
					WillReturnResult(sqlmock.NewResult(0, 1))

				rows := sqlmock.NewRows([]string{"id", "name", "created_at", "last_updated", "owner"}).
					AddRow(nonExistingListId, listName, testDate, testDate, ownerId)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetList)).WithArgs(nonExistingListId).WillReturnRows(rows)
			},
			expectError:    false,
			err:            nil,
			expectedEntity: initEntityList(nonExistingListId, listName, testDate, testDate, ownerId),
		},
		{
			testName:     "Unable to create list because of already existing listId",
			passedEntity: initEntityList(existingListId, listName, testDate, testDate, ownerId),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryInsertList)).
					WithArgs(existingListId, listName, testDate, testDate, ownerId).WillReturnError(&pq.Error{
					Code: "23505",
				})
			},
			expectError:    true,
			err:            application_errors2.NewAlreadyExistError(constants.LIST_TARGET, listName),
			expectedEntity: nil,
		},
		{
			testName:     "Unable to create list because of invalid owner",
			passedEntity: initEntityList(existingListId, listName, testDate, testDate, dummyListOwner),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectExec(regexp.QuoteMeta(sqlQueryInsertList)).
					WithArgs(existingListId, listName, testDate, testDate, dummyListOwner).WillReturnError(&pq.Error{
					Code: "23503",
				})
			},
			expectError:    true,
			err:            application_errors2.NewNotFoundError(constants.USER_TARGET, dummyListOwner.String()),
			expectedEntity: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.mockSetup(mock)

			listRepo := NewSQLListDB(db)
			gotEntity, err := listRepo.CreateList(test.passedEntity)
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

func TestSqlListDB_GetListOwner(t *testing.T) {
	tests := []struct {
		testName    string
		listId      string
		mockSetup   func(mck sqlmock.Sqlmock)
		expectError bool
		err         error
		entityOwner *entities2.User
	}{
		{
			testName: "Successfully getting list owner",
			listId:   existingListId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "role"}).
					AddRow(ownerId, userEmail, adminRole)
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetListOwner)).
					WithArgs(existingListId).WillReturnRows(rows)
			},
			expectError: false,
			err:         nil,
			entityOwner: initEntityUser(ownerId, userEmail, adminRole),
		},
		{
			testName: "Unable to get list owner because of invalid list_id",
			listId:   nonExistingListId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetListOwner)).
					WithArgs(nonExistingListId).WillReturnError(sql.ErrNoRows)
			},
			expectError: true,
			err:         application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String()),
			entityOwner: nil,
		},
		{
			testName: "Unable to get list owner because of database error",
			listId:   existingListId.String(),
			mockSetup: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery(regexp.QuoteMeta(sqlQueryGetListOwner)).
					WithArgs(existingListId).WillReturnError(databaseError)
			},
			expectError: true,
			err:         databaseError,
			entityOwner: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.Newx()
			require.NoError(t, err)
			defer db.Close()

			test.mockSetup(mock)
			listRepo := NewSQLListDB(db)
			gotOwner, err := listRepo.GetListOwner(test.listId)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.entityOwner, gotOwner)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
