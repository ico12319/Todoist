package users

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/persistence"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type genericRepository interface {
	GetPaginationInfo(ctx context.Context, sourceName string, filter string, params []interface{}) (*entities.PaginationInfo, error)
}

type sqlDecoratorFactory interface {
	CreateSqlDecorator(ctx context.Context, filters sql_query_decorators.Filters, initialQuery string) (sql_query_decorators.SqlQueryRetriever, error)
}

type repository struct {
	genericRepo genericRepository
	factory     sqlDecoratorFactory
}

func NewRepo(genericRepo genericRepository, factory sqlDecoratorFactory) *repository {
	return &repository{
		genericRepo: genericRepo,
		factory:     factory,
	}
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

func (r *repository) GetUsers(ctx context.Context, f *filters.UserFilters) ([]entities.User, error) {
	log.C(ctx).Info("getting users in user repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return nil, err
	}

	baseQuery := `SELECT id, email, role FROM users`

	decorator, err := r.factory.CreateSqlDecorator(ctx, f, baseQuery)
	if err != nil {
		log.C(ctx).Errorf("failed to get users, error when calling factory function")
		return nil, err
	}

	sqlQueryString := decorator.DetermineCorrectSqlQuery(ctx)
	completeQuery := fmt.Sprintf(`SELECT id, email, role FROM (%s) ORDER BY id`, sqlQueryString)

	var userEntities []entities.User
	if err = persist.SelectContext(ctx, &userEntities, completeQuery); err != nil {
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

func (r *repository) GetTodosAssignedToUser(ctx context.Context, userID string, f *filters.UserFilters) ([]entities.Todo, error) {
	log.C(ctx).Info("getting todos assigned to user in user repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return nil, err
	}

	baseQuery := `SELECT id, name, description, list_id, status, created_at,
				last_updated, assigned_to, due_date, priority FROM user_todos WHERE user_id = $1`

	decorator, err := r.factory.CreateSqlDecorator(ctx, f, baseQuery)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user in user service, error when calling factory function")
		return nil, err
	}

	sqlQuery := decorator.DetermineCorrectSqlQuery(ctx)
	completeQuery := fmt.Sprintf(`SELECT id, name, description, list_id, status, created_at,
				last_updated, assigned_to, due_date, priority FROM (%s) ORDER BY id`, sqlQuery)

	var todos []entities.Todo
	if err = persist.SelectContext(ctx, &todos, completeQuery, userID); err != nil {
		log.C(ctx).Errorf("failed to get todos assigned to user, error %s when executing sql query", err.Error())
		return nil, errors.New("unexpected database error")
	}

	return todos, nil
}

func (r *repository) GetUserLists(ctx context.Context, userID string, f *filters.UserFilters) ([]entities.List, error) {
	log.C(ctx).Info("getting user lists in user repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in user repo, error %s", err.Error())
		return nil, err
	}

	baseQuery := `SELECT DISTINCT id, name, created_at, last_updated, owner, description FROM lists_and_users WHERE owner = $1`

	decorator, err := r.factory.CreateSqlDecorator(ctx, f, baseQuery)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates in, error when calling factory function")
		return nil, err
	}

	sqlQueryString := decorator.DetermineCorrectSqlQuery(ctx)
	completeQuery := fmt.Sprintf(`SELECT id, name, created_at, last_updated, owner, description FROM (%s) ORDER BY id`, sqlQueryString)

	log.C(ctx).Errorf(completeQuery)
	var listEntities []entities.List
	if err = persist.SelectContext(ctx, &listEntities, completeQuery, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("sql err no rows when trying to get lists of user id %s, error %s", userID, err.Error())
			return nil, application_errors.NewNotFoundError(constants.USER_TARGET, userID)
		}
		log.C(ctx).Errorf("failed to get lists where user participates, error %s when executing sql query", err.Error())
		return nil, errors.New("unexpected database error")
	}

	return listEntities, nil
}

func (r *repository) GetUsersPaginationInfo(ctx context.Context) (*entities.PaginationInfo, error) {
	log.C(ctx).Info("getting users pagination info in user repository")

	return r.genericRepo.GetPaginationInfo(ctx, "users", "TRUE", nil)
}

func (r *repository) GetUserListsPaginationInfo(ctx context.Context, userID string) (*entities.PaginationInfo, error) {
	log.C(ctx).Infof("getting user lists pagination info of user with id %s in user repository", userID)

	return r.genericRepo.GetPaginationInfo(ctx, "lists_and_users", `owner = $1`, []interface{}{userID})
}

func (r *repository) GetTodosAssignedToUserPaginationInfo(ctx context.Context, userID string) (*entities.PaginationInfo, error) {
	log.C(ctx).Infof("getting todos assigned to user with id %s pagination info in user repository", userID)

	return r.genericRepo.GetPaginationInfo(ctx, "user_todos", `user_id = $1`, []interface{}{userID})
}
