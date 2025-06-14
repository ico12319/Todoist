package users

import (
	"context"
	"database/sql"
	"errors"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/application_errors"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/entities"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/constants"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type sqlQueryRetriever interface {
	DetermineCorrectSqlQuery(ctx context.Context) string
}

type sqlUserDB struct {
	db *sqlx.DB
}

func NewSQLUserDB(db *sqlx.DB) *sqlUserDB {
	return &sqlUserDB{db: db}
}

func (s *sqlUserDB) CreateUser(ctx context.Context, user *entities.User) (*entities.User, error) {
	log.C(ctx).Info("creating user in user repository")

	sqlQueryString := `INSERT INTO users (id, email, role) VALUES (:id,:email,:role)`

	_, err := s.db.NamedExecContext(ctx, sqlQueryString, user)
	if err != nil {
		log.C(ctx).Errorf("failed to create user, error %s when executing sql query", err.Error())
		return nil, err
	}

	return s.GetUser(ctx, user.Id.String())
}

func (s *sqlUserDB) DeleteUser(ctx context.Context, id string) error {
	log.C(ctx).Infof("deleting user with id %s in user repository", id)

	sqlQueryString := `DELETE FROM users WHERE id = $1`

	_, err := s.db.ExecContext(ctx, sqlQueryString, id)
	if err != nil {
		log.C(ctx).Errorf("failed to delete user, error when executing sql query")
		return err
	}

	return nil
}

func (s *sqlUserDB) DeleteUsers(ctx context.Context) error {
	log.C(ctx).Info("failed to delete users in user repository")

	sqlQueryString := `DELETE FROM users`
	if _, err := s.db.ExecContext(ctx, sqlQueryString); err != nil {
		log.C(ctx).Errorf("failed to delete lists, error %s when trying to execute sql query", err.Error())
		return err
	}

	return nil
}
func (s *sqlUserDB) UpdateUserPartially(ctx context.Context, sqlExecParams map[string]interface{}, sqlFields []string) (*entities.User, error) {
	log.C(ctx).Info("updating user partially in user repository")

	sqlQueryString := parseUserQuery(sqlFields)
	userId := sqlExecParams["id"].(string)

	res, err := s.db.NamedExecContext(ctx, sqlQueryString, sqlExecParams)
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

	return s.GetUser(ctx, userId)
}

func (s *sqlUserDB) UpdateUser(ctx context.Context, id string, user *entities.User) (*entities.User, error) {
	log.C(ctx).Info("updating user in user repository")

	sqlQueryString := `UPDATE users SET (id, email, role) = ($1, $2, $3) WHERE id = $4`

	res, err := s.db.ExecContext(ctx, sqlQueryString, user.Id, user.Email, user.Role, id)
	if err != nil {
		log.C(ctx).Errorf("failed to update user with id %s, error when trying to exec sql query", id)
		return nil, err
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

	return s.GetUser(ctx, user.Id.String())
}

func (s *sqlUserDB) GetUsers(ctx context.Context, retriever sqlQueryRetriever) ([]entities.User, error) {
	log.C(ctx).Info("getting users in user repository")

	sqlQueryString := retriever.DetermineCorrectSqlQuery(ctx)

	var entities []entities.User
	if err := s.db.Select(&entities, sqlQueryString); err != nil {
		log.C(ctx).Errorf("failed to get users from user repository due to a database error %s", err.Error())
		return nil, err
	}
	return entities, nil
}

func (s *sqlUserDB) GetUser(ctx context.Context, userId string) (*entities.User, error) {
	log.C(ctx).Infof("getting user with user_id %s from user repository", userId)

	sqlQueryString := `SELECT id,email,role FROM users WHERE id = $1`

	entity := &entities.User{}
	if err := s.db.GetContext(ctx, entity, sqlQueryString, userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("failed to get user with id %s due to a sqlErrNoRows", err.Error())
			return nil, application_errors.NewNotFoundError(constants.USER_TARGET, userId)
		}

		log.C(ctx).Errorf("failed to get user with id %s due to a database error %s", userId, err.Error())
		return nil, err
	}
	return entity, nil
}

func (s *sqlUserDB) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	log.C(ctx).Infof("getting user by email %s in user repository", email)

	sqlQueryString := `SELECT id, email, role FROM users WHERE email = $1`

	entity := &entities.User{}
	if err := s.db.GetContext(ctx, entity, sqlQueryString, email); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("failed to get user by email %s due to a sqlErrNoRows", err.Error())
			return nil, application_errors.NewNotFoundError(constants.USER_TARGET, email)
		}

		log.C(ctx).Errorf("failed to get user by email %s due to a database error %s", email, err.Error())
		return nil, err
	}
	return entity, nil
}

func (s *sqlUserDB) GetUserIdByEmail(ctx context.Context, email string) (string, error) {
	log.C(ctx).Infof("getting user id by email %s in user repository", email)

	sqlQueryString := `SELECT id FROM users WHERE email = $1`
	var userId uuid.UUID
	if err := s.db.GetContext(ctx, &userId, sqlQueryString, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("failed to get user id by email %s due to a sqlErrNoRows", email)
			return "", application_errors.NewAlreadyExistError(constants.USER_TARGET, email)
		}

		log.C(ctx).Errorf("failed to get user id by email %s due to a database error %s", email, err.Error())
		return "", err
	}
	return userId.String(), nil
}

func (s *sqlUserDB) GetTodosAssignedToUser(ctx context.Context, retriever sqlQueryRetriever) ([]entities.Todo, error) {
	log.C(ctx).Info("getting todos assigned to user in user repository")

	sqlQuery := retriever.DetermineCorrectSqlQuery(ctx)

	var todos []entities.Todo
	if err := s.db.SelectContext(ctx, &todos, sqlQuery); err != nil {
		log.C(ctx).Errorf("failed to get todos assigned to user, error %s when executing sql query", err.Error())
		return nil, err
	}
	return todos, nil
}

func (s *sqlUserDB) GetUserLists(ctx context.Context, retriever sqlQueryRetriever) ([]entities.List, error) {
	log.C(ctx).Info("getting user lists in user repository")

	sqlQueryString := retriever.DetermineCorrectSqlQuery(ctx)

	var listEntities []entities.List
	if err := s.db.SelectContext(ctx, &listEntities, sqlQueryString); err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates, error %s when executing sql query", err.Error())
		return nil, err
	}

	return listEntities, nil
}
