package todos

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/utils"
	"Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type sqlQueryRetriever interface {
	DetermineCorrectSqlQuery(ctx context.Context) string
}
type sqlTodoDB struct {
	db *sqlx.DB
}

func NewSQLTodoDB(db *sqlx.DB) *sqlTodoDB {
	return &sqlTodoDB{db: db}
}

func (s *sqlTodoDB) GetTodos(ctx context.Context, sqlRetriever sqlQueryRetriever) ([]entities.Todo, error) {
	log.C(ctx).Info("getting all todos in todo repository")

	sqlQuery := sqlRetriever.DetermineCorrectSqlQuery(ctx)

	var entities []entities.Todo
	if err := s.db.Select(&entities, sqlQuery); err != nil {
		log.C(ctx).Errorf("failed to get todos due to a database error %s", err.Error())
		return nil, err
	}
	return entities, nil
}

func (s *sqlTodoDB) DeleteTodosByListId(ctx context.Context, listId string) error {
	log.C(ctx).Infof("deleting todo from a list withd id %s", listId)

	sqlQueryString := `DELETE FROM todos WHERE list_id = $1`

	_, err := s.db.Exec(sqlQueryString, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to delete todos by list_id %s due to a database error %s", listId, err.Error())
		return err
	}

	return nil
}

func (s *sqlTodoDB) GetTodo(ctx context.Context, todoId string) (*entities.Todo, error) {
	log.C(ctx).Infof("getting todo with id %s in todo repository", todoId)

	sqlQueryString := `SELECT id, name, description, list_id, status, 
       					created_at, last_updated, assigned_to, due_date, priority FROM todos 
       					WHERE id = $1`

	entity := &entities.Todo{}
	if err := s.db.Get(entity, sqlQueryString, todoId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("failed to get todo with id %s due to sqlErrNoRows", todoId)
			return nil, application_errors.NewNotFoundError(constants.TODO_TARGET, todoId)
		}
		log.C(ctx).Errorf("failed to get todo with id %s because of a database error %s", todoId, err.Error())
		return nil, err
	}

	return entity, nil
}

func (s *sqlTodoDB) DeleteTodo(ctx context.Context, todoId string) error {
	log.C(ctx).Infof("deleting todo with id %s in todo repository", todoId)

	sqlQueryString := `DELETE FROM todos WHERE id = $1`

	if _, err := s.db.ExecContext(ctx, sqlQueryString, todoId); err != nil {
		log.C(ctx).Errorf("failed to delete todo due to a database error %s", err.Error())
		return err
	}

	return nil
}

func (s *sqlTodoDB) DeleteTodos(ctx context.Context) error {
	log.C(ctx).Info("deleting todos in todo repository")

	sqlQueryString := `DELETE FROM todos`

	if _, err := s.db.ExecContext(ctx, sqlQueryString); err != nil {
		log.C(ctx).Errorf("failed to delete todos due to a database errror %s", err.Error())
		return err
	}

	return nil
}

func (s *sqlTodoDB) CreateTodo(ctx context.Context, entity *entities.Todo) (*entities.Todo, error) {
	log.C(ctx).Info("creating todo in todo repository")

	sqlQueryString := `INSERT INTO todos(id, name, description, 
                  list_id, created_at, last_updated, assigned_to, due_date, priority) VALUES(:id,:name,:description,
                  :list_id,:created_at,:last_updated,:assigned_to,:due_date,:priority)`

	_, err := s.db.NamedExecContext(ctx, sqlQueryString, entity)
	if err != nil {
		log.C(ctx).Errorf("failed to create todo due to an error %s when executing sql query", err.Error())
		return nil, utils.MapPostgresTodoError(err, entity)
	}
	return s.GetTodo(ctx, entity.Id.String())
}

func (s *sqlTodoDB) UpdateTodo(ctx context.Context, sqlExecParams map[string]interface{}, sqlFields []string) (*entities.Todo, error) {
	log.C(ctx).Info("updating todo in todo repository")

	sqlQueryString := parseTodoQuery(sqlFields)
	todoId := sqlExecParams["id"].(string)

	res, err := s.db.NamedExecContext(ctx, sqlQueryString, sqlExecParams)

	if err != nil {
		log.C(ctx).Errorf("failed to update todo due to an error %s when executing sql query", err.Error())
		return nil, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.C(ctx).Errorf("error %s when trying to see the number of rows affected", err.Error())
		return nil, err
	}

	if rowsAffected == 0 {
		log.C(ctx).Error("failed to update todo due to 0 rows being affected")
		return nil, application_errors.NewNotFoundError(constants.TODO_TARGET, todoId)
	}

	return s.GetTodo(ctx, todoId)
}

func (s *sqlTodoDB) GetTodoAssigneeTo(ctx context.Context, todoId string) (*entities.User, error) {
	log.C(ctx).Infof("getting todo %s assignee in todo repository", todoId)

	user := &entities.User{}
	sqlQueryString := `SELECT users.id,users.email,users.role FROM users JOIN todos on users.id = todos.assigned_to
WHERE todos.id = $1`

	if err := s.db.GetContext(ctx, user, sqlQueryString, todoId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Error("failed to get todo assignee due to a sqlErrNoRows error")
			return nil, nil
		}
		log.C(ctx).Error("failed to get todo assignee due to a database error")
		return nil, err

	}
	return user, nil
}

func (s *sqlTodoDB) GetTodosByListId(ctx context.Context, decorator sqlQueryRetriever, listId string) ([]entities.Todo, error) {
	log.C(ctx).Infof("getting todos of list with id %s", listId)

	sqlQueryString := decorator.DetermineCorrectSqlQuery(ctx)

	var todos []entities.Todo
	if err := s.db.SelectContext(ctx, &todos, sqlQueryString, listId); err != nil {
		log.C(ctx).Errorf("failed to get todos of list with id %s, error when trying to execute sql query", listId)
		return nil, err
	}
	return todos, nil
}

func (s *sqlTodoDB) GetTodoByListId(ctx context.Context, listId string, todoId string) (*entities.Todo, error) {
	log.C(ctx).Infof("getting todo with id %s, from list with id %s", todoId, listId)

	sqlQueryString := `SELECT id, name, description, list_id, status, created_at, last_updated, 
assigned_to, due_date, priority FROM todos WHERE list_id = $1 AND id = $2`

	todo := &entities.Todo{}
	if err := s.db.Get(todo, sqlQueryString, listId, todoId); err != nil {
		log.C(ctx).Errorf("failed to get todo with id %s from list with id %s, error %s when trying to execute sql query", todoId, listId, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Info("error is SqlNoRows...")
			return nil, application_errors.NewNotFoundError(constants.TODO_TARGET, todoId)
		}
		log.C(ctx).Info("some other db error...")
		return nil, fmt.Errorf("internal errror")
	}
	return todo, nil
}
