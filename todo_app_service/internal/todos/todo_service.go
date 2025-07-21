package todos

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"fmt"
	"time"
)

//go:generate mockery --name=TodoRepo --output=./mocks --outpkg=mocks --filename=todo_repo.go --with-expecter=true
type todoRepo interface {
	CreateTodo(ctx context.Context, entity *entities.Todo) (*entities.Todo, error)
	UpdateTodo(ctx context.Context, sqlExecParams map[string]interface{}, sqlFields []string) (*entities.Todo, error)
	DeleteTodo(ctx context.Context, todoId string) error
	DeleteTodosByListId(ctx context.Context, listId string) error
	DeleteTodos(ctx context.Context) error
	GetTodo(ctx context.Context, todoId string) (*entities.Todo, error)
	GetTodos(ctx context.Context, sqlRetriever sqlQueryRetriever) ([]entities.Todo, error)
	GetTodosByListId(ctx context.Context, decorator sqlQueryRetriever, listId string) ([]entities.Todo, error)
	GetTodoAssigneeTo(ctx context.Context, todoId string) (*entities.User, error)
	GetTodoByListId(ctx context.Context, listId string, todoId string) (*entities.Todo, error)
}

type listService interface {
	GetListRecord(ctx context.Context, listId string) (*models.List, error)
	GetCollaborators(ctx context.Context, lFilters *filters.ListFilters) ([]*models.User, error)
	GetListOwnerRecord(ctx context.Context, listId string) (*models.User, error)
	CheckWhetherUserIsCollaborator(ctx context.Context, listId string, userId string) (bool, error)
}

//go:generate mockery --name=IUuidGenerator --output=./mocks --outpkg=mocks --filename=IUuidGenerator.go --with-expecter=true
type uuidGenerator interface {
	Generate() string
}

//go:generate mockery --name=ITimeGenerator --output=./mocks --outpkg=mocks --filename=ITimeGenerator.go --with-expecter=true
type timeGenerator interface {
	Now() time.Time
}

type todoConverter interface {
	ConvertFromDBEntityToModel(todo *entities.Todo) *models.Todo
	ConvertFromModelToDBEntity(todo *models.Todo) *entities.Todo
	ManyToModel(todos []entities.Todo) []*models.Todo
	ConvertFromCreateHandlerModelToModel(todo *handler_models.CreateTodo) *models.Todo
	ConvertFromUpdateHandlerModelToModel(todo *handler_models.UpdateTodo) *models.Todo
}

type userConverter interface {
	ConvertFromDBEntityToModel(user *entities.User) *models.User
}

type sqlDecoratorFactory interface {
	CreateSqlDecorator(context.Context, sql_query_decorators.Filters, string) (sql_query_decorators.SqlQueryRetriever, error)
}

type service struct {
	tRepo      todoRepo
	lService   listService
	uuidGen    uuidGenerator
	timeGen    timeGenerator
	tConverter todoConverter
	uConverter userConverter
	factory    sqlDecoratorFactory
}

func NewService(tRepo todoRepo, lService listService, uuidGen uuidGenerator, timeGen timeGenerator,
	todoConverter todoConverter, userConverter userConverter, factory sqlDecoratorFactory) *service {
	return &service{tRepo: tRepo, lService: lService, uuidGen: uuidGen, timeGen: timeGen,
		tConverter: todoConverter, uConverter: userConverter, factory: factory}
}

func (s *service) CreateTodoRecord(ctx context.Context, todo *handler_models.CreateTodo, creator *models.User) (*models.Todo, error) {
	log.C(ctx).Info("creating todo in todo service")

	if creator.Role != constants.Admin {
		createErr := fmt.Errorf("only users that are part of the list can create todo")
		if err := s.checkWhetherUserHasAccessToTodo(ctx, creator.Id, todo.ListId, createErr); err != nil {
			log.C(ctx).Errorf("failed to create todo, error %s user trying to create todo does not have access to it", err.Error())
			return nil, err
		}

		if todo.AssignedTo != nil {
			assignError := fmt.Errorf("only the list owner and the list collaborators can be assigned to todo")
			if err := s.checkWhetherUserHasAccessToTodo(ctx, *todo.AssignedTo, todo.ListId, assignError); err != nil {
				log.C(ctx).Errorf("failed to create todo, error %s", err.Error())
				return nil, err
			}
		}
	}

	modelTodo := s.tConverter.ConvertFromCreateHandlerModelToModel(todo)

	modelTodo.Id = s.uuidGen.Generate()
	modelTodo.LastUpdated = s.timeGen.Now()
	modelTodo.CreatedAt = s.timeGen.Now()

	entityTodo := s.tConverter.ConvertFromModelToDBEntity(modelTodo)

	returnedEntity, err := s.tRepo.CreateTodo(ctx, entityTodo)
	if err != nil {
		log.C(ctx).Errorf("failed to create todo in todo service dut to an error in todo repository")
		return nil, err
	}

	return s.tConverter.ConvertFromDBEntityToModel(returnedEntity), nil
}

func (s *service) DeleteTodoRecord(ctx context.Context, todoId string) error {
	log.C(ctx).Infof("deleting todo with id %s in todo service", todoId)

	return s.tRepo.DeleteTodo(ctx, todoId)
}

func (s *service) DeleteTodosRecordsByListId(ctx context.Context, listId string) error {
	log.C(ctx).Infof("deleting todos from a list with id %s", listId)

	if _, err := s.lService.GetListRecord(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to delete todos from a list with id %s, error %s when calling list repo", listId, err.Error())
		return err
	}

	return s.tRepo.DeleteTodosByListId(ctx, listId)
}

func (s *service) DeleteTodosRecords(ctx context.Context) error {
	log.C(ctx).Info("deleting all todos")

	return s.tRepo.DeleteTodos(ctx)
}

func (s *service) UpdateTodoRecord(ctx context.Context, todoId string, todo *handler_models.UpdateTodo) (*models.Todo, error) {
	log.C(ctx).Infof("updating todo with id %s in todo service", todoId)

	sqlExecParams := map[string]interface{}{"id": todoId}
	sqlFields := make([]string, 0)

	modelTodo := s.tConverter.ConvertFromUpdateHandlerModelToModel(todo)
	if modelTodo.AssignedTo != nil {
		assignError := fmt.Errorf("only the list owner and the list collaborators can be assigned to todo")

		if err := s.checkWhetherUserHasAccessToTodo(ctx, *modelTodo.AssignedTo, modelTodo.ListId, assignError); err != nil {
			log.C(ctx).Errorf("failed to update todo, error %s", err.Error())
			return nil, err
		}
	}

	determineSqlFieldsAndParamsTodo(modelTodo, sqlExecParams, &sqlFields)

	entity, err := s.tRepo.UpdateTodo(ctx, sqlExecParams, sqlFields)
	if err != nil {
		log.C(ctx).Errorf("failed to update todo record %s, error in todo repository", todoId)
		return nil, err
	}

	return s.tConverter.ConvertFromDBEntityToModel(entity), nil
}

func (s *service) GetTodoRecord(ctx context.Context, todoId string) (*models.Todo, error) {
	log.C(ctx).Infof("getting todo with id %s", todoId)

	entity, err := s.tRepo.GetTodo(ctx, todoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo with id %s due to an error in todo repository", todoId)
		return nil, err
	}

	return s.tConverter.ConvertFromDBEntityToModel(entity), nil
}

func (s *service) GetTodoAssigneeToRecord(ctx context.Context, todoId string) (*models.User, error) {
	log.C(ctx).Infof("getting user assigned to todo with id %s", todoId)

	assignee, err := s.tRepo.GetTodoAssigneeTo(ctx, todoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get user assigned to todo with id %s, error %s when calling todo repo", todoId, err.Error())
		return nil, err
	}

	return s.uConverter.ConvertFromDBEntityToModel(assignee), nil
}

func (s *service) GetTodoRecords(ctx context.Context, filters *filters.TodoFilters) ([]*models.Todo, error) {
	log.C(ctx).Info("getting todos in todo service")

	// this is just the base part of the sql query,
	//the decorator decorates it and pass the already decorated query to the repo layer
	//where the actual execution happens

	baseQuery := `WITH sorted_todos AS(
		SELECT * FROM todos ORDER BY id
	)
	SELECT id,name,description,list_id,status,created_at,last_updated,assigned_to,due_date,priority FROM sorted_todos`

	decorator, err := s.factory.CreateSqlDecorator(ctx, filters, baseQuery)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos, error when calling factory function")
		return nil, err
	}

	eTodos, err := s.tRepo.GetTodos(ctx, decorator)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos due to an error in todo service %s", err.Error())
		return nil, err
	}

	return s.tConverter.ManyToModel(eTodos), nil
}

func (s *service) GetTodosByListId(ctx context.Context, filters *filters.TodoFilters, listId string) ([]*models.Todo, error) {
	log.C(ctx).Infof("getting todos of list with id %s in list service", listId)

	if _, err := s.lService.GetListRecord(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to get todos of list with id %s, error when calling list repo", listId)
		return nil, err
	}

	baseQuery := `WITH sorted_todos AS(
					SELECT * FROM todos ORDER BY id
                 )
				SELECT id,name,description,list_id,status,created_at,last_updated,assigned_to,due_date,priority FROM sorted_todos WHERE list_id = $1`

	decorator, err := s.factory.CreateSqlDecorator(ctx, filters, baseQuery)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos of list with id %s, error when creating query decorator", listId)
		return nil, err
	}

	eTodos, err := s.tRepo.GetTodosByListId(ctx, decorator, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos of list with id %s, error when calling todo repo", listId)
		return nil, err
	}

	return s.tConverter.ManyToModel(eTodos), nil
}

func (s *service) GetTodoByListId(ctx context.Context, listId string, todoId string) (*models.Todo, error) {
	log.C(ctx).Infof("getting todo with id %s, from list with id %s in todo service", todoId, listId)

	if _, err := s.lService.GetListRecord(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to get todo with id %s, invalid list id", todoId)
		return nil, err
	}

	entityTodo, err := s.tRepo.GetTodoByListId(ctx, listId, todoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo with id %s, invalid todo id", todoId)
		return nil, err
	}

	return s.tConverter.ConvertFromDBEntityToModel(entityTodo), nil
}

func (s *service) checkWhetherUserHasAccessToTodo(ctx context.Context, userId string, listId string, desiredErr error) error {
	todoListOwner, err := s.lService.GetListOwnerRecord(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to create todo, error %s when trying to get list owner", err.Error())
		return err
	}

	isCollaborator, err := s.lService.CheckWhetherUserIsCollaborator(ctx, listId, userId)
	if err != nil {
		log.C(ctx).Errorf("failed to create todo record, error %s when trying to chech whether user with id %s is collaborator", err.Error(), userId)
		return err
	}

	if userId != todoListOwner.Id && !isCollaborator {
		log.C(ctx).Debug("user is not owner and is not part of collaborators...")
		return desiredErr
	}

	return nil
}
