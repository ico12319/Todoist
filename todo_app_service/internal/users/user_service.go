package users

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	"Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
)

type userRepo interface {
	CreateUser(context.Context, *entities.User) (*entities.User, error)
	GetUsers(context.Context, sqlQueryRetriever) ([]entities.User, error)
	GetUser(context.Context, string) (*entities.User, error)
	GetUserByEmail(context.Context, string) (*entities.User, error)
	UpdateUserPartially(context.Context, map[string]interface{}, []string) (*entities.User, error)
	UpdateUser(context.Context, string, *entities.User) (*entities.User, error)
	DeleteUser(context.Context, string) error
	DeleteUsers(context.Context) error
	GetTodosAssignedToUser(context.Context, sqlQueryRetriever) ([]entities.Todo, error)
	GetUserLists(context.Context, sqlQueryRetriever) ([]entities.List, error)
}

type userConverter interface {
	ConvertFromDBEntityToModel(*entities.User) *models.User
	ConvertFromModelToEntity(*models.User) *entities.User
	ConvertFromUpdateModelToModel(*handler_models.UpdateUser) *models.User
	ConvertFromCreateHandlerModelToModel(*handler_models.CreateUser) *models.User
	ManyToModel([]entities.User) []*models.User
}

type uuidGenerator interface {
	Generate() string
}

type sqlDecoratorFactory interface {
	CreateSqlDecorator(context.Context, sql_query_decorators.Filters, string) (sql_query_decorators.SqlQueryRetriever, error)
}

type listConverter interface {
	ManyToModel([]entities.List) []*models.List
}

type todoConverter interface {
	ManyToModel([]entities.Todo) []*models.Todo
}
type service struct {
	repo       userRepo
	converter  userConverter
	lConverter listConverter
	tConverter todoConverter
	uuidGen    uuidGenerator
	factory    sqlDecoratorFactory
}

func NewService(repo userRepo, converter userConverter, lConverter listConverter, tConverter todoConverter, uuidGen uuidGenerator, factory sqlDecoratorFactory) *service {
	return &service{repo: repo, converter: converter, lConverter: lConverter, tConverter: tConverter, uuidGen: uuidGen, factory: factory}
}

func (s *service) GetUsersRecords(ctx context.Context, uFilters *filters.UserFilters) ([]*models.User, error) {
	log.C(ctx).Info("getting users in user service")

	decorator, err := s.factory.CreateSqlDecorator(ctx, uFilters, baseUserGetQuery)
	if err != nil {
		log.C(ctx).Errorf("failed to get users, error when calling factory function")
		return nil, err
	}

	entities, err := s.repo.GetUsers(ctx, decorator)
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

	decorator, err := s.factory.CreateSqlDecorator(ctx, uFilter, baseUserGetLists)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates in, error when calling factory function")
		return nil, err
	}

	listEntities, err := s.repo.GetUserLists(ctx, decorator)
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

	decorator, err := s.factory.CreateSqlDecorator(ctx, userFilters, baseUserGetTodos)
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
