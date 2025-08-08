package todos

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"errors"
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
	GetTodos(ctx context.Context, f *filters.TodoFilters) ([]entities.Todo, error)
	GetTodosByListId(ctx context.Context, listId string, f *filters.TodoFilters) ([]entities.Todo, error)
	GetTodoAssigneeTo(ctx context.Context, todoId string) (*entities.User, error)
	GetTodoByListId(ctx context.Context, listId string, todoId string) (*entities.Todo, error)
	UnassignUserFromTodos(ctx context.Context, userId string, listId string) error
	GetTodosPaginationInfo(ctx context.Context) (*entities.PaginationInfo, error)
	GetTodosFromListPaginationInfo(ctx context.Context, listID string) (*entities.PaginationInfo, error)
}

type listRepo interface {
	GetList(ctx context.Context, listId string) (*entities.List, error)
	GetListOwner(ctx context.Context, listId string) (*entities.User, error)
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
	ToModel(todo *entities.Todo) *models.Todo
	ToEntity(todo *models.Todo) *entities.Todo
	ManyToPage(todos []entities.Todo, pageInfo *entities.PaginationInfo) *models.TodoPage
	ConvertFromCreateHandlerModelToModel(todo *handler_models.CreateTodo) *models.Todo
	ConvertFromUpdateHandlerModelToModel(todo *handler_models.UpdateTodo) *models.Todo
}

type userConverter interface {
	ToModel(user *entities.User) *models.User
}

type service struct {
	tRepo      todoRepo
	lRepo      listRepo
	uuidGen    uuidGenerator
	timeGen    timeGenerator
	tConverter todoConverter
	uConverter userConverter
}

func NewService(tRepo todoRepo, lRepo listRepo, uuidGen uuidGenerator, timeGen timeGenerator,
	todoConverter todoConverter, userConverter userConverter) *service {
	return &service{
		tRepo:      tRepo,
		lRepo:      lRepo,
		uuidGen:    uuidGen,
		timeGen:    timeGen,
		tConverter: todoConverter,
		uConverter: userConverter,
	}
}

func (s *service) CreateTodoRecord(ctx context.Context, todo *handler_models.CreateTodo, creator *models.User) (*models.Todo, error) {
	log.C(ctx).Info("creating todo in todo service")

	if creator.Role != constants.Admin {
		if err := s.checkWhetherUserHasAccessToTodo(ctx, creator.Id, todo.ListId, errors.New("only users that are part of the list can create todo")); err != nil {
			log.C(ctx).Errorf("failed to create todo, error %s user trying to create todo does not have access to it", err.Error())
			return nil, err
		}
	}

	if todo.AssignedTo != nil {
		if err := s.checkWhetherUserHasAccessToTodo(ctx, *todo.AssignedTo, todo.ListId, errors.New("only the list owner and the list collaborators can be assigned to todo")); err != nil {
			log.C(ctx).Errorf("failed to create todo, error %s", err.Error())
			return nil, err
		}
	}

	modelTodo := s.tConverter.ConvertFromCreateHandlerModelToModel(todo)

	modelTodo.Id = s.uuidGen.Generate()
	modelTodo.LastUpdated = s.timeGen.Now()
	modelTodo.CreatedAt = s.timeGen.Now()

	entityTodo := s.tConverter.ToEntity(modelTodo)

	returnedEntity, err := s.tRepo.CreateTodo(ctx, entityTodo)
	if err != nil {
		log.C(ctx).Errorf("failed to create todo in todo service dut to an error in todo repository")
		return nil, err
	}

	return s.tConverter.ToModel(returnedEntity), nil
}

func (s *service) DeleteTodoRecord(ctx context.Context, todoId string) error {
	log.C(ctx).Infof("deleting todo with id %s in todo service", todoId)

	if err := s.tRepo.DeleteTodo(ctx, todoId); err != nil {
		log.C(ctx).Errorf("failed to delete todo with id %s", err.Error())
		return err
	}

	return nil
}

func (s *service) DeleteTodosRecordsByListId(ctx context.Context, listId string) error {
	log.C(ctx).Infof("deleting todos from a list with id %s", listId)

	if _, err := s.lRepo.GetList(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to delete todos from a list with id %s, error %s when calling list repo", listId, err.Error())
		return err
	}

	if err := s.tRepo.DeleteTodosByListId(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to delete todos by list id %s, error %s", listId, err.Error())
		return err
	}

	return nil
}

func (s *service) DeleteTodosRecords(ctx context.Context) error {
	log.C(ctx).Info("deleting all todos")

	if err := s.tRepo.DeleteTodos(ctx); err != nil {
		log.C(ctx).Errorf("failed to delete todos in todo sercice, error %s", err.Error())
		return err
	}
	return nil
}

func (s *service) UpdateTodoRecord(ctx context.Context, todoId string, todo *handler_models.UpdateTodo) (*models.Todo, error) {
	log.C(ctx).Infof("updating todo with id %s in todo service", todoId)

	sqlExecParams := map[string]interface{}{"id": todoId}
	sqlFields := make([]string, 0)

	todoEntity, err := s.tRepo.GetTodo(ctx, todoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo with id %s when trying to update todo, error %s", err.Error())
		return nil, err
	}

	modelTodo := s.tConverter.ConvertFromUpdateHandlerModelToModel(todo)
	if todo.AssignedTo != nil {
		assignError := fmt.Errorf("only the list owner and the list collaborators can be assigned to todo")

		if err = s.checkWhetherUserHasAccessToTodo(ctx, *todo.AssignedTo, todoEntity.ListId.String(), assignError); err != nil {
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

	return s.tConverter.ToModel(entity), nil
}

func (s *service) GetTodoRecord(ctx context.Context, todoId string) (*models.Todo, error) {
	log.C(ctx).Infof("getting todo with id %s", todoId)

	entity, err := s.tRepo.GetTodo(ctx, todoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo with id %s due to an error in todo repository", todoId)
		return nil, err
	}

	return s.tConverter.ToModel(entity), nil
}

func (s *service) GetTodoAssigneeToRecord(ctx context.Context, todoId string) (*models.User, error) {
	log.C(ctx).Infof("getting user assigned to todo with id %s", todoId)

	assignee, err := s.tRepo.GetTodoAssigneeTo(ctx, todoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get user assigned to todo with id %s, error %s when calling todo repo", todoId, err.Error())
		return nil, err
	}

	return s.uConverter.ToModel(assignee), nil
}

func (s *service) GetTodoRecords(ctx context.Context, filters *filters.TodoFilters) (*models.TodoPage, error) {
	log.C(ctx).Info("getting todos in todo service")

	eTodos, err := s.tRepo.GetTodos(ctx, filters)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos due to an error in todo service %s", err.Error())
		return nil, err
	}

	paginationInfo, err := s.tRepo.GetTodosPaginationInfo(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get first and last ids of lists, error %s", err.Error())
		return nil, err
	}

	return s.tConverter.ManyToPage(eTodos, paginationInfo), nil
}

func (s *service) GetTodosByListId(ctx context.Context, filters *filters.TodoFilters, listId string) (*models.TodoPage, error) {
	log.C(ctx).Infof("getting todos of list with id %s in list service", listId)

	if _, err := s.lRepo.GetList(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to get todos of list with id %s, error when calling list repo", listId)
		return nil, err
	}

	eTodos, err := s.tRepo.GetTodosByListId(ctx, listId, filters)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos of list with id %s, error when calling todo repo", listId)
		return nil, err
	}

	paginationInfo, err := s.tRepo.GetTodosFromListPaginationInfo(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get first and last ids of lists, error %s", err.Error())
		return nil, err
	}

	return s.tConverter.ManyToPage(eTodos, paginationInfo), nil
}

func (s *service) GetTodoByListId(ctx context.Context, listId string, todoId string) (*models.Todo, error) {
	log.C(ctx).Infof("getting todo with id %s, from list with id %s in todo service", todoId, listId)

	if _, err := s.lRepo.GetList(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to get todo with id %s, invalid list id", todoId)
		return nil, err
	}

	entityTodo, err := s.tRepo.GetTodoByListId(ctx, listId, todoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo with id %s, invalid todo id", todoId)
		return nil, err
	}

	return s.tConverter.ToModel(entityTodo), nil
}

func (s *service) checkWhetherUserHasAccessToTodo(ctx context.Context, userId string, listId string, desiredErr error) error {
	log.C(ctx).Info("checking whether user with id %s has access to todos from list with id %s", userId, listId)

	todoListOwner, err := s.lRepo.GetListOwner(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to check whether user has access to todo, error %s when trying to get list owner", err.Error())
		return err
	}

	isCollaborator, err := s.lRepo.CheckWhetherUserIsCollaborator(ctx, listId, userId)
	if err != nil {
		log.C(ctx).Errorf("failed to create todo record, error %s when trying to check whether user with id %s is collaborator", err.Error(), userId)
		return err
	}

	if userId != todoListOwner.Id.String() && !isCollaborator {
		log.C(ctx).Debug("user is not owner and is not part of collaborators...")
		return desiredErr
	}

	return nil
}

func (s *service) UnassignUserFromTodos(ctx context.Context, userId string, listId string) error {
	log.C(ctx).Infof("unassigning user with id %s from todo from id %s", userId, listId)

	if err := s.tRepo.UnassignUserFromTodos(ctx, userId, listId); err != nil {
		log.C(ctx).Errorf("failed to unassign user with id %s from todos of list with id %s, error %s", userId, listId, err.Error())
		return err
	}

	return nil
}
