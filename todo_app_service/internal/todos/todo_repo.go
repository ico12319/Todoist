package todos

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
	CreateSqlDecorator(ctx context.Context, f sql_query_decorators.Filters, initialQuery string) (sql_query_decorators.SqlQueryRetriever, error)
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

func (r *repository) GetTodos(ctx context.Context, f *filters.TodoFilters) ([]entities.Todo, error) {
	log.C(ctx).Info("getting all todos in todo repository")

	baseQuery := `SELECT id,name,description,list_id,status,created_at,last_updated,assigned_to,due_date,priority, COUNT(*) OVER() AS total_count
FROM todos`

	decorator, err := r.factory.CreateSqlDecorator(ctx, f, baseQuery)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos, error when calling factory function")
		return nil, err
	}

	sqlQuery := decorator.DetermineCorrectSqlQuery(ctx)
	completeSqlQuery := fmt.Sprintf(`SELECT id,name,description,list_id,status,created_at,
       last_updated,assigned_to,due_date,priority FROM (%s) ORDER BY id`, sqlQuery)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in todo repo, error %s", err.Error())
		return nil, err
	}

	var todos []entities.Todo
	if err = persist.SelectContext(ctx, &todos, completeSqlQuery); err != nil {
		log.C(ctx).Errorf("failed to get todos due to a database error %s", err.Error())
		return nil, err
	}

	return todos, nil
}

func (*repository) DeleteTodosByListId(ctx context.Context, listId string) error {
	log.C(ctx).Infof("deleting todo from a list withd id %s", listId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in todo repo, error %s", err.Error())
		return err
	}

	sqlQueryString := `DELETE FROM todos WHERE list_id = $1`

	_, err = persist.ExecContext(ctx, sqlQueryString, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to delete todos by list_id %s due to a database error %s", listId, err.Error())
		return err
	}

	return nil
}

func (*repository) GetTodo(ctx context.Context, todoId string) (*entities.Todo, error) {
	log.C(ctx).Infof("getting todo with id %s in todo repository", todoId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in todo repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `SELECT id, name, description, list_id, status, 
       					created_at, last_updated, assigned_to, due_date, priority FROM todos 
       					WHERE id = $1`

	entity := &entities.Todo{}
	if err = persist.GetContext(ctx, entity, sqlQueryString, todoId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("failed to get todo with id %s due to sqlErrNoRows", todoId)
			return nil, application_errors.NewNotFoundError(constants.TODO_TARGET, todoId)
		}
		log.C(ctx).Errorf("failed to get todo with id %s because of a database error %s", todoId, err.Error())
		return nil, err
	}

	return entity, nil
}

func (*repository) DeleteTodo(ctx context.Context, todoId string) error {
	log.C(ctx).Infof("deleting todo with id %s in todo repository", todoId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in todo repo, error %s", err.Error())
		return err
	}

	sqlQueryString := `DELETE FROM todos WHERE id = $1`

	if _, err = persist.ExecContext(ctx, sqlQueryString, todoId); err != nil {
		log.C(ctx).Errorf("failed to delete todo due to a database error %s", err.Error())
		return err
	}

	return nil
}

func (*repository) DeleteTodos(ctx context.Context) error {
	log.C(ctx).Info("deleting todos in todo repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in todo repo, error %s", err.Error())
		return err
	}

	sqlQueryString := `DELETE FROM todos`

	if _, err = persist.ExecContext(ctx, sqlQueryString); err != nil {
		log.C(ctx).Errorf("failed to delete todos due to a database errror %s", err.Error())
		return err
	}

	return nil
}

func (r *repository) CreateTodo(ctx context.Context, entity *entities.Todo) (*entities.Todo, error) {
	log.C(ctx).Info("creating todo in todo repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in todo repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `INSERT INTO todos(id, name, description, 
                  list_id, created_at, last_updated, assigned_to, due_date, priority) VALUES(:id,:name,:description,
                  :list_id,:created_at,:last_updated,:assigned_to,:due_date,:priority)`

	_, err = persist.NamedExecContext(ctx, sqlQueryString, entity)
	if err != nil {
		log.C(ctx).Errorf("failed to create todo due to an error %s when executing sql query", err.Error())
		return nil, persistence.MapPostgresTodoError(err, entity)
	}
	return r.GetTodo(ctx, entity.Id.String())
}

func (r *repository) UpdateTodo(ctx context.Context, sqlExecParams map[string]interface{}, sqlFields []string) (*entities.Todo, error) {
	log.C(ctx).Info("updating todo in todo repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in todo repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := parseTodoQuery(sqlFields)
	todoId := sqlExecParams["id"].(string)

	res, err := persist.NamedExecContext(ctx, sqlQueryString, sqlExecParams)

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

	return r.GetTodo(ctx, todoId)
}

func (*repository) GetTodoAssigneeTo(ctx context.Context, todoId string) (*entities.User, error) {
	log.C(ctx).Infof("getting todo %s assignee in todo repository", todoId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in todo repo, error %s", err.Error())
		return nil, err
	}

	user := &entities.User{}
	sqlQueryString := `SELECT users.id,users.email,users.role FROM users JOIN todos on users.id = todos.assigned_to
WHERE todos.id = $1`

	if err = persist.GetContext(ctx, user, sqlQueryString, todoId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Error("failed to get todo assignee due to a sqlErrNoRows error")
			return nil, nil
		}
		log.C(ctx).Error("failed to get todo assignee due to a database error")
		return nil, err

	}
	return user, nil
}

func (r *repository) GetTodosByListId(ctx context.Context, listId string, f *filters.TodoFilters) ([]entities.Todo, error) {
	log.C(ctx).Infof("getting todos of list with id %s", listId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in todo repo, error %s", err.Error())
		return nil, err
	}

	baseQuery := `SELECT id,name,description,list_id,status,created_at,last_updated,assigned_to,due_date,priority FROM todos WHERE list_id = $1`

	decorator, err := r.factory.CreateSqlDecorator(ctx, f, baseQuery)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos of list with id %s, error when creating query decorator", listId)
		return nil, err
	}

	sqlQueryString := decorator.DetermineCorrectSqlQuery(ctx)

	completeSqlQuery := fmt.Sprintf(`SELECT id,name,description,list_id,status,created_at,last_updated,assigned_to,due_date,priority FROM (%s) ORDER BY id`, sqlQueryString)

	var todos []entities.Todo
	if err = persist.SelectContext(ctx, &todos, completeSqlQuery, listId); err != nil {
		log.C(ctx).Errorf("failed to get todos of list with id %s, error when trying to execute sql query", listId)
		return nil, err
	}
	return todos, nil
}

func (*repository) GetTodoByListId(ctx context.Context, listId string, todoId string) (*entities.Todo, error) {
	log.C(ctx).Infof("getting todo with id %s, from list with id %s", todoId, listId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in todo repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `SELECT id, name, description, list_id, status, created_at, last_updated, 
assigned_to, due_date, priority FROM todos WHERE list_id = $1 AND id = $2`

	todo := &entities.Todo{}
	if err = persist.GetContext(ctx, todo, sqlQueryString, listId, todoId); err != nil {
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

func (*repository) UnassignUserFromTodos(ctx context.Context, userId string, listId string) error {
	log.C(ctx).Infof("unassigning user with id %s from todo from a list with id %s", userId, listId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from context in todo repo, error %s", err.Error())
		return err
	}

	sqlQueryString := `UPDATE todos SET assigned_to = NULL 
WHERE assigned_to = $1 AND list_id = $2`

	if _, err = persist.ExecContext(ctx, sqlQueryString, userId, listId); err != nil {
		log.C(ctx).Errorf("failed to unnassign user from todo, error %s", err.Error())
		return err
	}

	return nil
}

func (r *repository) GetTodosPaginationInfo(ctx context.Context) (*entities.PaginationInfo, error) {
	log.C(ctx).Info("getting todos pagination info in todo repository")

	return r.genericRepo.GetPaginationInfo(ctx, "todos", "TRUE", nil)
}

func (r *repository) GetTodosFromListPaginationInfo(ctx context.Context, listID string) (*entities.PaginationInfo, error) {
	log.C(ctx).Infof("getting todo from list with id %s in todo repository", listID)

	return r.genericRepo.GetPaginationInfo(ctx, "todos", `list_id = $1`, []interface{}{listID})
}
