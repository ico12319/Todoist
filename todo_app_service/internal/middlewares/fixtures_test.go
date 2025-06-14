package middlewares

import (
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/constants"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/models"
	"time"
)

const (
	listId            = "list1"
	listName          = "list"
	adminRole         = "admin"
	readerRole        = "reader"
	writerRole        = "writer"
	testEmail         = "test@yahoo.com"
	adminEmail        = "admin@yahoo.com"
	ownerEmail        = "owner@yahoo.com"
	sharedWithMeEmail = "sharedWithMe@yahoo.com"
	notOwnerEmail     = "notowner@abv.bg"
)

var mockTime = time.Date(2025, time.January, 1,
	1, 1, 1, 1, time.UTC)

func initUser(email string, role constants.UserRole) models.User {
	return models.User{
		Email: email,
		Role:  role,
	}
}
