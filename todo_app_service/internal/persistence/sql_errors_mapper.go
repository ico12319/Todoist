package persistence

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"errors"
	"github.com/lib/pq"
)

func MapPostgresListErrorToError(err error, entityList *entities.List) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return application_errors.NewAlreadyExistError(constants.LIST_TARGET, entityList.Name)
		case "23503":
			return application_errors.NewNotFoundError(constants.USER_TARGET, entityList.Owner.String())
		}
	}
	return err
}

func MapPostgresTodoError(err error, todo *entities.Todo) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return err
	}

	switch pqErr.Code {
	case "23505":
		return application_errors.NewNotFoundError(constants.TODO_TARGET, todo.Id.String())
	case "23503":
		switch pqErr.Constraint {
		case "todos_list_id_fkey":
			return application_errors.NewNotFoundError(constants.LIST_TARGET, todo.ListId.String())
		case "todos_assigned_to_fkey":
			if todo.AssignedTo.Valid {
				return application_errors.NewNotFoundError(constants.USER_TARGET, todo.AssignedTo.UUID.String())
			}
		}
	}
	return err
}

func MapPostgresUserError(err error, user *entities.User) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return errors.New("unexpected database error")
	}

	switch pqErr.Code {
	case "23505":
		switch pqErr.Constraint {
		case "users_pkey":
			return application_errors.NewAlreadyExistError(constants.USER_TARGET, user.Id.String())
		case "users_email_key":

			return application_errors.NewAlreadyExistError(constants.USER_TARGET, user.Email)
		}
	}

	return errors.New("unexpected database error")
}

func MapPostgresNonExistingUserInUserTable(err error, refresh *entities.Refresh) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return errors.New("unexpected database error")
	}

	switch pqErr.Code {
	case "23505":
		switch pqErr.Constraint {
		case "user_refresh_tokens_pkey":
			return application_errors.NewAlreadyExistError(constants.USER_TARGET, refresh.UserId.String())
		case "users_refresh_tokens_refresh_token_key":
			return application_errors.NewAlreadyExistError(constants.REFRESH_TARGET, refresh.RefreshToken)
		}
	case "23503":
		return application_errors.NewNotFoundError(constants.USER_TARGET, refresh.UserId.String())
	}

	return errors.New("unexpected database error")

}
