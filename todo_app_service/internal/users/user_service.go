package users

import (
	"context"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/entities"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/handler_models"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/models"
)

type userRepo interface {
	CreateUser(ctx context.Context, user *entities.User) (*entities.User, error)
	GetUsers(ctx context.Context, retriever sqlQueryRetriever) ([]entities.User, error)
	GetUser(ctx context.Context, userId string) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	UpdateUserPartially(ctx context.Context, sqlExecParams map[string]interface{}, sqlFields []string) (*entities.User, error)
	UpdateUser(ctx context.Context, id string, user *entities.User) (*entities.User, error)
	DeleteUser(ctx context.Context, id string) error
	DeleteUsers(ctx context.Context) error
	GetTodosAssignedToUser(ctx context.Context, retriever sqlQueryRetriever) ([]entities.Todo, error)
	GetUserLists(ctx context.Context, retriever sqlQueryRetriever) ([]entities.List, error)
}

type userConverter interface {
	ConvertFromDBEntityToModel(user *entities.User) *models.User
	ConvertFromModelToEntity(user *models.User) *entities.User
	ConvertFromUpdateModelToModel(user *handler_models.UpdateUser) *models.User
	ConvertFromCreateHandlerModelToModel(user *handler_models.CreateUser) *models.User
	ManyToModel(users []entities.User) []*models.User
}

type uuidGenerator interface {
	Generate() string
}

type sqlConcreteDecoratorQueryFactory interface {
	CreateUserQueryDecorator(ctx context.Context, initialQuery string, userFilters *filters.UserFilters) (sql_query_decorators.SqlQueryRetriever, error)
}

type listConverter interface {
	ManyToModel(lists []entities.List) []*models.List
}

type todoConverter interface {
	ManyToModel(todos []entities.Todo) []*models.Todo
}
type service struct {
	repo       userRepo
	converter  userConverter
	lConverter listConverter
	tConverter todoConverter
	uuidGen    uuidGenerator
	factory    sqlConcreteDecoratorQueryFactory
}

func NewService(repo userRepo, converter userConverter, lConverter listConverter, tConverter todoConverter, uuidGen uuidGenerator, factory sqlConcreteDecoratorQueryFactory) *service {
	return &service{repo: repo, converter: converter, lConverter: lConverter, tConverter: tConverter, uuidGen: uuidGen, factory: factory}
}

func (s *service) GetUsersRecords(ctx context.Context, uFilters *filters.UserFilters) ([]*models.User, error) {
	log.C(ctx).Info("getting users in user service")

	retriever, err := s.factory.CreateUserQueryDecorator(ctx, baseUserGetQuery, uFilters)
	if err != nil {
		log.C(ctx).Errorf("failed to get users, error when calling factory function")
		return nil, err
	}

	entities, err := s.repo.GetUsers(ctx, retriever)
	if err != nil {
		log.C(ctx).Errorf("failed to get users in user service due to an error in user repository %s", err.Error())
		return nil, err
	}

	return s.converter.ManyToModel(entities), nil
}

func (s *service) GetUserRecord(ctx context.Context, userId string) (*models.User, error) {
	log.C(ctx).Infof("getting user with id %s in user service", userId)

	entity, err := s.repo.GetUser(ctx, userId)
	if err != nil {
		log.C(ctx).Errorf("failed to get user with id %s in user service due to an error in user repository")
		return nil, err
	}

	return s.converter.ConvertFromDBEntityToModel(entity), nil
}

func (s *service) GetUserRecordByEmail(ctx context.Context, email string) (*models.User, error) {
	log.C(ctx).Infof("getting user by email %s in user service", email)

	entity, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		log.C(ctx).Errorf("failed to get user by email %s in user service due to an error in user repository", email)
		return nil, err
	}

	return s.converter.ConvertFromDBEntityToModel(entity), nil
}

func (s *service) CreateUserRecord(ctx context.Context, user *handler_models.CreateUser) (*models.User, error) {
	log.C(ctx).Info("creating user in user service")

	createModel := s.converter.ConvertFromCreateHandlerModelToModel(user)
	createModel.Id = s.uuidGen.Generate()

	convertedEntity := s.converter.ConvertFromModelToEntity(createModel)

	_, err := s.repo.CreateUser(ctx, convertedEntity)
	if err != nil {
		log.C(ctx).Errorf("failed to create user in user service, error %s when calling user repo", err.Error())
		return nil, err
	}

	return createModel, nil
}

func (s *service) DeleteUserRecord(ctx context.Context, id string) error {
	log.C(ctx).Infof("deleting user with id %s in user service", id)

	if err := s.repo.DeleteUser(ctx, id); err != nil {
		log.C(ctx).Errorf("failed to update user in user service, error %s when calling user repo", err.Error())
		return err
	}

	return nil
}

func (s *service) DeleteUsers(ctx context.Context) error {
	log.C(ctx).Info("deleting users in user service")

	return s.repo.DeleteUsers(ctx)
}

func (s *service) UpdateUserRecord(ctx context.Context, id string, user *models.User) (*models.User, error) {
	log.C(ctx).Infof("updating user with id %s in user service", id)

	convertedEntity := s.converter.ConvertFromModelToEntity(user)

	updatedEntity, err := s.repo.UpdateUser(ctx, id, convertedEntity)
	if err != nil {
		log.C(ctx).Errorf("failed to update user in user service, error %s when calling user repo", err.Error())
		return nil, err
	}

	return s.converter.ConvertFromDBEntityToModel(updatedEntity), nil
}

func (s *service) UpdateUserRecordPartially(ctx context.Context, id string, user *handler_models.UpdateUser) (*models.User, error) {
	log.C(ctx).Infof("updating user with with id %s in user service", id)

	convertedModel := s.converter.ConvertFromUpdateModelToModel(user)

	sqlExecParams := map[string]interface{}{"id": id}
	sqlFields := make([]string, 0, 8)

	determineSqlFieldsAndParamsUser(convertedModel, sqlExecParams, &sqlFields)

	updatedEntity, err := s.repo.UpdateUserPartially(ctx, sqlExecParams, sqlFields)
	if err != nil {
		log.C(ctx).Errorf("failed to partially update user with id %s, errror when calling user repo %s", id, err.Error())
		return nil, err
	}

	return s.converter.ConvertFromDBEntityToModel(updatedEntity), nil
}

func (s *service) GetUserListsRecords(ctx context.Context, uFilter *filters.UserFilters) ([]*models.List, error) {
	log.C(ctx).Info("getting lists where user participates in user service")

	if _, err := s.repo.GetUser(ctx, uFilter.UserId); err != nil {
		log.C(ctx).Errorf("failed to get user lists, error %s when calling user repo", err.Error())
		return nil, err
	}

	retriever, err := s.factory.CreateUserQueryDecorator(ctx, baseUserGetLists, uFilter)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates in, error when calling factory function")
		return nil, err
	}

	listEntities, err := s.repo.GetUserLists(ctx, retriever)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates in, error %s when calling user repo function", err.Error())
		return nil, err
	}

	return s.lConverter.ManyToModel(listEntities), nil
}

func (s *service) GetTodosAssignedToUser(ctx context.Context, userFilters *filters.UserFilters) ([]*models.Todo, error) {
	log.C(ctx).Info("getting todos assigned to user in user service")

	if _, err := s.repo.GetUser(ctx, userFilters.UserId); err != nil {
		log.C(ctx).Errorf("failed to get todos assigned to user, error %s when calling user repo", err.Error())
		return nil, err
	}

	decorator, err := s.factory.CreateUserQueryDecorator(ctx, baseUserGetTodos, userFilters)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user in user service, error when calling factory function")
		return nil, err
	}

	todoEntities, err := s.repo.GetTodosAssignedToUser(ctx, decorator)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user in user service, error when calling repo")
		return nil, err
	}

	return s.tConverter.ManyToModel(todoEntities), nil
}
