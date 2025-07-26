package users

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/persistence"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
	"database/sql"
	"errors"
)

//go:generate mockery --name=sqlQueryRetriever --exported --output=./mocks --outpkg=mocks --filename=sqlQuery_retriever.go --with-expecter=true
type sqlQueryRetriever interface {
	DetermineCorrectSqlQuery(ctx context.Context) string
}

type repository struct{}

func NewRepo() *repository {
	return &repository{}
}

func (*repository) CreateUser(ctx context.Context, user *entities.User) (*entities.User, error) {
	log.C(ctx).Info("creating user in user repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `INSERT INTO users (id, email, role) VALUES (:id,:email,:role)`

	_, err = persist.NamedExecContext(ctx, sqlQueryString, user)
	if err != nil {
		log.C(ctx).Errorf("failed to create user, error %s when executing sql query", err.Error())
		return nil, persistence.MapPostgresUserError(err, user)
	}

	return user, nil
}

func (*repository) DeleteUser(ctx context.Context, id string) error {
	log.C(ctx).Infof("deleting user with id %s in user repository", id)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return err
	}

	sqlQueryString := `DELETE FROM users WHERE id = $1`

	_, err = persist.ExecContext(ctx, sqlQueryString, id)
	if err != nil {
		log.C(ctx).Errorf("failed to delete user, error when executing sql query")
		return errors.New("unexpected database error")
	}

	return nil
}

func (*repository) DeleteUsers(ctx context.Context) error {
	log.C(ctx).Info("failed to delete users in user repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return err
	}

	sqlQueryString := `DELETE FROM users`
	if _, err = persist.ExecContext(ctx, sqlQueryString); err != nil {
		log.C(ctx).Errorf("failed to delete lists, error %s when trying to execute sql query", err.Error())
		return errors.New("unexpected database error")
	}

	return nil
}
func (r *repository) UpdateUserPartially(ctx context.Context, sqlExecParams map[string]interface{}, sqlFields []string) (*entities.User, error) {
	log.C(ctx).Info("updating user partially in user repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := parseUserQuery(sqlFields)
	userId := sqlExecParams["id"].(string)

	res, err := persist.NamedExecContext(ctx, sqlQueryString, sqlExecParams)
	if err != nil {
		log.C(ctx).Error("failed to update user partially, error when trying to exec sql query")
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.C(ctx).Error("failed to update user partially, error when trying to get the number of rows affected")
		return nil, err
	}

	if rowsAffected == 0 {
		log.C(ctx).Errorf("failed to update user partially, error because the number of rows affected is 0")
		return nil, application_errors.NewNotFoundError(constants.USER_TARGET, userId)
	}

	return r.GetUser(ctx, userId)
}

func (r *repository) UpdateUser(ctx context.Context, id string, user *entities.User) (*entities.User, error) {
	log.C(ctx).Info("updating user in user repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `UPDATE users SET (id, email, role) = ($1, $2, $3) WHERE id = $4`

	res, err := persist.ExecContext(ctx, sqlQueryString, user.Id, user.Email, user.Role, id)
	if err != nil {
		log.C(ctx).Errorf("failed to update user with id %s, error when trying to exec sql query", id)
		return nil, errors.New("unexpected database error")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.C(ctx).Errorf("failed to update user with id %s, error when trying to get the number of rows affected", id)
		return nil, err
	}

	if rowsAffected == 0 {
		log.C(ctx).Errorf("failed to update user with id %s, error because the number of rows affected is 0", id)
		return nil, application_errors.NewNotFoundError(constants.USER_TARGET, id)
	}

	return r.GetUser(ctx, user.Id.String())
}

func (*repository) GetUsers(ctx context.Context, retriever sqlQueryRetriever) ([]entities.User, error) {
	log.C(ctx).Info("getting users in user repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := retriever.DetermineCorrectSqlQuery(ctx)

	var userEntities []entities.User
	if err = persist.SelectContext(ctx, &userEntities, sqlQueryString); err != nil {
		log.C(ctx).Errorf("failed to get users from user repository due to a database error %s", err.Error())
		return nil, errors.New("unexpected database error")
	}

	return userEntities, nil
}

func (*repository) GetUser(ctx context.Context, userId string) (*entities.User, error) {
	log.C(ctx).Infof("getting user with user_id %s from user repository", userId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `SELECT id,email,role FROM users WHERE id = $1`

	entity := &entities.User{}
	if err = persist.GetContext(ctx, entity, sqlQueryString, userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("failed to get user with id %s due to a sqlErrNoRows", err.Error())
			return nil, application_errors.NewNotFoundError(constants.USER_TARGET, userId)
		}

		log.C(ctx).Errorf("failed to get user with id %s due to a database error %s", userId, err.Error())
		return nil, errors.New("unexpected database error")
	}
	return entity, nil
}

func (*repository) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	log.C(ctx).Infof("getting user by email %s in user repository", email)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `SELECT id, email, role FROM users WHERE email = $1`

	entity := &entities.User{}
	if err = persist.GetContext(ctx, entity, sqlQueryString, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("failed to get user by email %s due to a sqlErrNoRows", err.Error())
			return nil, application_errors.NewNotFoundError(constants.USER_TARGET, email)
		}

		log.C(ctx).Errorf("failed to get user by email %s due to a database error %s", email, err.Error())
		return nil, errors.New("unexpected database error")
	}
	return entity, nil
}

func (*repository) GetTodosAssignedToUser(ctx context.Context, userId string, retriever sqlQueryRetriever) ([]entities.Todo, error) {
	log.C(ctx).Info("getting todos assigned to user in user repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return nil, err
	}

	sqlQuery := retriever.DetermineCorrectSqlQuery(ctx)

	var todos []entities.Todo
	if err := persist.SelectContext(ctx, &todos, sqlQuery, userId); err != nil {
		log.C(ctx).Errorf("failed to get todos assigned to user, error %s when executing sql query", err.Error())
		return nil, errors.New("unexpected database error")
	}

	return todos, nil
}

func (*repository) GetUserLists(ctx context.Context, retriever sqlQueryRetriever) ([]entities.List, error) {
	log.C(ctx).Info("getting user lists in user repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := retriever.DetermineCorrectSqlQuery(ctx)

	var listEntities []entities.List
	if err := persist.SelectContext(ctx, &listEntities, sqlQueryString); err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates, error %s when executing sql query", err.Error())
		return nil, errors.New("unexpected database error")
	}

	return listEntities, nil
}
