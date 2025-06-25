package lists

import (
	entities2 "Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	models2 "Todo-List/internProject/todo_app_service/pkg/models"
	"fmt"
	"github.com/gofrs/uuid"
	"time"
)

const (
	listName         = "listName"
	newListName      = "changedListName"
	dbError          = "database error"
	userEmail        = "email@email.com"
	userEmail2       = "email2@email.com"
	userEmail3       = "email3@email.com"
	adminRole        = "admin"
	writerRole       = "writer"
	readerRole       = "reader"
	invalidUserEmail = "invalid@email.com"
	sqlQueryGetList  = `SELECT id, name, created_at, last_updated, owner 
						FROM lists where id = $1`
	sqlQueryDeleteList = `DELETE FROM lists WHERE id = $1`
	sqlQueryGetLists   = `SELECT id, name, created_at, last_updated, owner 
						FROM lists`
	sqlQueryUpdateListName           = `UPDATE lists SET name = $1 WHERE id = $2`
	sqlQueryFindUserIdQuery          = `SELECT id FROM users WHERE email = $1`
	sqlInsertIntoUserListsTableQuery = `INSERT INTO user_lists (user_id, list_id) VALUES ($1,$2)`
	sqlQueryGetCollaborator          = `SELECT users.id,users.email,users.role FROM users
						JOIN user_lists ON users.id = user_lists.user_id 
                        WHERE list_id = $1`
	sqlQueryInsertList = `INSERT INTO lists (id, name, created_at, last_updated, owner) 
						VALUES (?,?,?,?,?)`
	sqlQueryGetListOwner = `SELECT users.id,users.role,users.email FROM lists 
    JOIN users ON lists.owner = users.id WHERE lists.id = $1`
)

var (
	userId            = uuid.Must(uuid.NewV4())
	userId2           = uuid.Must(uuid.NewV4())
	userId3           = uuid.Must(uuid.NewV4())
	existingListId    = uuid.Must(uuid.NewV4())
	listId2           = uuid.Must(uuid.NewV4())
	listId3           = uuid.Must(uuid.NewV4())
	listId4           = uuid.Must(uuid.NewV4())
	dummyListId       = uuid.Must(uuid.NewV4())
	ownerId           = uuid.Must(uuid.NewV4())
	dummyListOwner    = uuid.Must(uuid.NewV4())
	nonExistingListId = uuid.Must(uuid.NewV4())
	testDate          = time.Date(2021, time.January, 15, 10, 30, 0, 0, time.UTC)
	testDate2         = time.Date(2025, time.January, 11, 6, 24, 2, 1, time.UTC)
	// errors
	invalidListIdError           = fmt.Errorf("invalid list_id %s", nonExistingListId.String())
	nonExistingListIdError       = fmt.Errorf("invalid list_id %s", nonExistingListId.String())
	databaseError                = fmt.Errorf(dbError)
	invalidOwnerIdError          = fmt.Errorf("owner %s not found", ownerId.String())
	alreadyExistingListNameError = fmt.Errorf("a list with name %s already exists", listName)
	invalidUserEmailError        = fmt.Errorf("invalid email provided %s", userEmail)
)

func initEntityList(listId uuid.UUID, name string, createdAt time.Time, lastUpdated time.Time, owner uuid.UUID) *entities2.List {
	return &entities2.List{
		Id:          listId,
		Name:        name,
		CreatedAt:   createdAt,
		LastUpdated: lastUpdated,
		Owner:       owner,
	}
}

func initEntityUser(userId uuid.UUID, email string, role string) *entities2.User {
	return &entities2.User{
		Id:    userId,
		Email: email,
		Role:  role,
	}
}

func initModelUser(userId string, email string, role string) *models2.User {
	return &models2.User{
		Id:    userId,
		Email: email,
		Role:  constants.UserRole(role),
	}
}

func initModelList(listId string, name string, createdAt time.Time, lastUpdated time.Time, owner string) *models2.List {
	return &models2.List{
		Id:          listId,
		Name:        name,
		CreatedAt:   createdAt,
		LastUpdated: lastUpdated,
		Owner:       owner,
	}
}
