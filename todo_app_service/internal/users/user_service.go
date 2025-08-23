package users

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/resource_identifier"
	"Todo-List/internProject/todo_app_service/internal/source"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
)

type userRepo interface {
	CreateUser(ctx context.Context, user *entities.User) (*entities.User, error)
	GetUsers(ctx context.Context, f filters.SqlFilters) ([]entities.User, error)
	GetUser(ctx context.Context, userId string) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	UpdateUserPartially(ctx context.Context, params map[string]interface{}, fields []string) (*entities.User, error)
	UpdateUser(ctx context.Context, userId string, user *entities.User) (*entities.User, error)
	DeleteUser(ctx context.Context, userId string) error
	DeleteUsers(ctx context.Context) error
	GetTodosAssignedToUser(ctx context.Context, userID string, f filters.SqlFilters) ([]entities.Todo, error)
	GetUserLists(ctx context.Context, userID string, f filters.SqlFilters) ([]entities.List, error)
	GetPaginationInfo(ctx context.Context, f filters.SqlFilters, s source.Source) (*entities.PaginationInfo, error)
}

type userConverter interface {
	ToModel(user *entities.User) *models.User
	ToEntity(user *models.User) *entities.User
	ConvertFromUpdateModelToModel(user *handler_models.UpdateUser) *models.User
	ConvertFromCreateHandlerModelToModel(user *handler_models.CreateUser) *models.User
	ManyToPage(users []entities.User, paginationInfo *entities.PaginationInfo) *models.UserPage
}

type uuidGenerator interface {
	Generate() string
}

type listConverter interface {
	ManyToPage(lists []entities.List, paginationInfo *entities.PaginationInfo) *models.ListPage
}

type todoConverter interface {
	ManyToPage(todos []entities.Todo, paginationInfo *entities.PaginationInfo) *models.TodoPage
}

type resourceIdentifierAdapter interface {
	AdaptResourceIdentifier(rf resource_identifier.ResourceIdentifier) string
}
type service struct {
	repo       userRepo
	converter  userConverter
	lConverter listConverter
	tConverter todoConverter
	uuidGen    uuidGenerator
	rfAdapter  resourceIdentifierAdapter
}

func NewService(repo userRepo, converter userConverter, lConverter listConverter,
	tConverter todoConverter, uuidGen uuidGenerator, rfAdapter resourceIdentifierAdapter) *service {
	return &service{
		repo:       repo,
		converter:  converter,
		lConverter: lConverter,
		tConverter: tConverter,
		uuidGen:    uuidGen,
		rfAdapter:  rfAdapter,
	}
}

func (s *service) GetUsersRecords(ctx context.Context, uFilters filters.SqlFilters, rf resource_identifier.ResourceIdentifier) (*models.UserPage, error) {
	log.C(ctx).Info("getting users in user service")

	userEntities, err := s.repo.GetUsers(ctx, uFilters)
	if err != nil {
		log.C(ctx).Errorf("failed to get users in user service due to an error in user repository %s", err.Error())
		return nil, err
	}

	sqlSource := prepareSqlSource(s.rfAdapter, rf)
	paginationInfo, err := s.repo.GetPaginationInfo(ctx, uFilters, sqlSource)
	if err != nil {
		log.C(ctx).Errorf("failed to get first and last id in user service, error %s", err.Error())
		return nil, err
	}

	usersPage := s.converter.ManyToPage(userEntities, paginationInfo)

	return usersPage, nil
}

func (s *service) GetUserRecord(ctx context.Context, userId string) (*models.User, error) {
	log.C(ctx).Infof("getting user with id %s in user service", userId)

	entity, err := s.repo.GetUser(ctx, userId)
	if err != nil {
		log.C(ctx).Errorf("failed to get user with id %s in user service due to an error in user repository", userId)
		return nil, err
	}

	return s.converter.ToModel(entity), nil
}

func (s *service) GetUserRecordByEmail(ctx context.Context, email string) (*models.User, error) {
	log.C(ctx).Infof("getting user by email %s in user service", email)

	userEntity, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		log.C(ctx).Errorf("failed to get user by email %s in user service due to an error in user repository", email)
		return nil, err
	}

	return s.converter.ToModel(userEntity), nil
}

func (s *service) CreateUserRecord(ctx context.Context, user *handler_models.CreateUser) (*models.User, error) {
	log.C(ctx).Info("creating user in user service")

	createModel := s.converter.ConvertFromCreateHandlerModelToModel(user)
	createModel.Id = s.uuidGen.Generate()

	convertedEntity := s.converter.ToEntity(createModel)

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

	if err := s.repo.DeleteUsers(ctx); err != nil {
		log.C(ctx).Errorf("failed to delete users, error %s", err.Error())
		return err
	}

	return nil
}

func (s *service) UpdateUserRecord(ctx context.Context, id string, user *models.User) (*models.User, error) {
	log.C(ctx).Infof("updating user with id %s in user service", id)

	convertedEntity := s.converter.ToEntity(user)

	updatedEntity, err := s.repo.UpdateUser(ctx, id, convertedEntity)
	if err != nil {
		log.C(ctx).Errorf("failed to update user in user service, error %s when calling user repo", err.Error())
		return nil, err
	}

	return s.converter.ToModel(updatedEntity), nil
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

	return s.converter.ToModel(updatedEntity), nil
}

func (s *service) GetUserListsRecords(ctx context.Context, userId string, lFilter filters.SqlFilters, rf resource_identifier.ResourceIdentifier) (*models.ListPage, error) {
	log.C(ctx).Info("getting lists where user participates in user service")

	listEntities, err := s.repo.GetUserLists(ctx, userId, lFilter)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates in, error %s when calling user repo function", err.Error())
		return nil, err
	}

	sqlSource := prepareSqlSource(s.rfAdapter, rf)
	paginationInfo, err := s.repo.GetPaginationInfo(ctx, lFilter, sqlSource)
	if err != nil {
		log.C(ctx).Errorf("failed to get first and last ids in user service, error %s", err.Error())
		return nil, err
	}

	return s.lConverter.ManyToPage(listEntities, paginationInfo), nil
}

func (s *service) GetTodosAssignedToUser(ctx context.Context, userId string, tFilters filters.SqlFilters, rf resource_identifier.ResourceIdentifier) (*models.TodoPage, error) {
	log.C(ctx).Info("getting todos assigned to user in user service")

	if _, err := s.repo.GetUser(ctx, userId); err != nil {
		log.C(ctx).Errorf("failed to get todos assigned to user, error %s when calling user repo", err.Error())
		return nil, err
	}

	todoEntities, err := s.repo.GetTodosAssignedToUser(ctx, userId, tFilters)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user in user service, error when calling repo")
		return nil, err
	}

	sqlSource := prepareSqlSource(s.rfAdapter, rf)
	paginationInfo, err := s.repo.GetPaginationInfo(ctx, tFilters, sqlSource)
	if err != nil {
		log.C(ctx).Errorf("failed to get first and last ids in user service, error %s", err.Error())
		return nil, err
	}

	return s.tConverter.ManyToPage(todoEntities, paginationInfo), nil
}

func prepareSqlSource(adapter resourceIdentifierAdapter, rf resource_identifier.ResourceIdentifier) source.Source {
	adaptedRf := adapter.AdaptResourceIdentifier(rf)
	sqlSource := &source.SqlSource{}
	sqlSource.SetSource(adaptedRf)

	return sqlSource
}
