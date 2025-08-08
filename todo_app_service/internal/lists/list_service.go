package lists

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"fmt"
	"time"
)

//go:generate mockery --name=listRepo --exported --output=./mocks --outpkg=mocks --filename=list_repo.go --with-expecter=true
type listRepo interface {
	GetLists(ctx context.Context, f *filters.BaseFilters) ([]entities.List, error)
	GetList(context.Context, string) (*entities.List, error)
	GetListCollaborators(ctx context.Context, listID string, f *filters.BaseFilters) ([]entities.User, error)
	GetListOwner(context.Context, string) (*entities.User, error)
	DeleteList(context.Context, string) error
	DeleteLists(context.Context) error
	CreateList(context.Context, *entities.List) (*entities.List, error)
	UpdateList(context.Context, map[string]interface{}, []string) (*entities.List, error)
	UpdateListSharedWith(context.Context, string, string) error
	DeleteCollaborator(context.Context, string, string) error
	CheckWhetherUserIsCollaborator(context.Context, string, string) (bool, error)
	GetListsPaginationInfo(ctx context.Context) (*entities.PaginationInfo, error)
	GetListCollaboratorsPaginationInfo(ctx context.Context, listID string) (*entities.PaginationInfo, error)
}

type todoRepo interface {
	UnassignUserFromTodos(ctx context.Context, userId string, listId string) error
}

//go:generate mockery --name=uuidGenerator --exported --output=./mocks --outpkg=mocks --filename=uuid_generator.go --with-expecter=true
type uuidGenerator interface {
	Generate() string
}

//go:generate mockery --name=timeGenerator --exported --output=./mocks --outpkg=mocks --filename=time_generator.go --with-expecter=true
type timeGenerator interface {
	Now() time.Time
}

//go:generate mockery --name=listConverter --exported --output=./mocks --outpkg=mocks --filename=list_converter.go --with-expecter=true
type listConverter interface {
	ToModel(list *entities.List) *models.List
	ToEntity(list *models.List) *entities.List
	ManyToPage(lists []entities.List, pageInfo *entities.PaginationInfo) *models.ListPage
	FromUpdateHandlerModelToModel(list *handler_models.UpdateList) *models.List
	FromCreateHandlerModelToModel(list *handler_models.CreateList) *models.List
}

//go:generate mockery --name=userConverter --exported --output=./mocks --outpkg=mocks --filename=user_converter.go --with-expecter=true
type userConverter interface {
	ToModel(*entities.User) *models.User
	ToEntity(*models.User) *entities.User
	ManyToPage(users []entities.User, pageInfo *entities.PaginationInfo) *models.UserPage
}

//go:generate mockery --name=userService --exported --output=./mocks --outpkg=mocks --filename=user_service.go --with-expecter=true
type userRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
}

type service struct {
	lRepo      listRepo
	uRepo      userRepo
	tRepo      todoRepo
	uuidGen    uuidGenerator
	timeGen    timeGenerator
	lConverter listConverter
	uConverter userConverter
}

func NewService(repo listRepo, uuidGen uuidGenerator, timeGen timeGenerator,
	listConverter listConverter, uRep userRepo, tRepo todoRepo, userConverter userConverter) *service {
	return &service{
		lRepo:      repo,
		uuidGen:    uuidGen,
		timeGen:    timeGen,
		lConverter: listConverter,
		uRepo:      uRep,
		uConverter: userConverter,
		tRepo:      tRepo,
	}
}

func (s *service) GetListsRecords(ctx context.Context, f *filters.BaseFilters) (*models.ListPage, error) {
	log.C(ctx).Info("getting lists from list service")

	listEntities, err := s.lRepo.GetLists(ctx, f)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists records, error %s when calling list repo", err.Error())
		return nil, err
	}

	paginationInfo, err := s.lRepo.GetListsPaginationInfo(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get first and last ids of lists, error %s", err.Error())
		return nil, err
	}

	return s.lConverter.ManyToPage(listEntities, paginationInfo), nil
}

func (s *service) DeleteListRecord(ctx context.Context, listId string) error {
	log.C(ctx).Infof("deleting with id %s list record from list service layer", listId)

	if err := s.lRepo.DeleteList(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to delete list with id %s, error %s when calling list repo", listId, err.Error())
		return fmt.Errorf("failed to delete list with id %s", listId)
	}

	return nil
}

func (s *service) CreateListRecord(ctx context.Context, list *handler_models.CreateList, owner string) (*models.List, error) {
	log.C(ctx).Info("creating list record in list service layer")

	modelList := s.lConverter.FromCreateHandlerModelToModel(list)

	newlyCreatedList := &models.List{
		Id:          s.uuidGen.Generate(),
		Name:        modelList.Name,
		Description: modelList.Description,
		CreatedAt:   s.timeGen.Now(),
		LastUpdated: s.timeGen.Now(),
		Owner:       owner,
	}

	convertedEntity := s.lConverter.ToEntity(newlyCreatedList)

	if _, err := s.lRepo.CreateList(ctx, convertedEntity); err != nil {
		log.C(ctx).Errorf("failed to create list, error %s when calling list repo", err.Error())
		return nil, err
	}

	return newlyCreatedList, nil
}

func (s *service) UpdateListPartiallyRecord(ctx context.Context, listId string, list *handler_models.UpdateList) (*models.List, error) {
	log.C(ctx).Infof("updating list with id %s in list service ", listId)

	modelList := s.lConverter.FromUpdateHandlerModelToModel(list)

	sqlExecParams := map[string]interface{}{"id": listId}
	sqlFields := make([]string, 0, 8)

	determineSqlFieldsAndParamsList(modelList, sqlExecParams, &sqlFields)

	entity, err := s.lRepo.UpdateList(ctx, sqlExecParams, sqlFields)
	if err != nil {
		log.C(ctx).Errorf("failed to update list name %s", err.Error())
		return nil, err
	}

	return s.lConverter.ToModel(entity), nil
}

func (s *service) AddCollaborator(ctx context.Context, listId string, userEmail string) (*models.User, error) {
	log.C(ctx).Debugf("adding collaborator with email %s in list with id %s", userEmail, listId)

	user, err := s.uRepo.GetUserByEmail(ctx, userEmail)
	if err != nil {
		log.C(ctx).Errorf("failed to get user with email %s, error when calling user repo function", userEmail)
		return nil, err
	}

	if err = s.lRepo.UpdateListSharedWith(ctx, listId, user.Id.String()); err != nil {
		log.C(ctx).Errorf("failed to add collaborator with email %s in list with id %s, error when calling repo function", userEmail, listId)
		return nil, err
	}

	return s.uConverter.ToModel(user), nil
}

func (s *service) DeleteCollaborator(ctx context.Context, listId string, userId string) error {
	log.C(ctx).Infof("deleting a collaborator with id %s from list with id %s in list service", userId, listId)

	if _, err := s.lRepo.GetList(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to delete collaborator with id %s, error %s when calling list repo", userId, err.Error())
		return err
	}

	if err := s.lRepo.DeleteCollaborator(ctx, listId, userId); err != nil {
		log.C(ctx).Errorf("failed to delete a collaborator with id %s from a list with id %s, error in list repo", userId, listId)
		return err
	}

	if err := s.tRepo.UnassignUserFromTodos(ctx, userId, listId); err != nil {
		log.C(ctx).Errorf("failed to unassign user with id %s from todos from list with id %s, error %s", userId, listId, err.Error())
		return err
	}

	return nil
}

func (s *service) DeleteLists(ctx context.Context) error {
	log.C(ctx).Info("deleting lists in list service")

	if err := s.lRepo.DeleteLists(ctx); err != nil {
		log.C(ctx).Errorf("failed to delete lists, error %s", err.Error())
		return err
	}

	return nil
}

func (s *service) GetCollaborators(ctx context.Context, listId string, f *filters.BaseFilters) (*models.UserPage, error) {
	log.C(ctx).Info("getting list collaborators from list service")

	if _, err := s.lRepo.GetList(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to get list collaborators, error %s when calling list repo", err.Error())
		return nil, err
	}

	entitiesCollaborators, err := s.lRepo.GetListCollaborators(ctx, listId, f)
	if err != nil {
		log.C(ctx).Errorf("failed to get list collaborators, error %s when calling list repo", err.Error())
		return nil, err
	}

	paginationInfo, err := s.lRepo.GetListCollaboratorsPaginationInfo(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get first and last ids of lists, error %s", err.Error())
		return nil, err
	}

	return s.uConverter.ManyToPage(entitiesCollaborators, paginationInfo), nil
}

func (s *service) GetListRecord(ctx context.Context, listId string) (*models.List, error) {
	log.C(ctx).Infof("getting list with id %s from list service", listId)

	listEntity, err := s.lRepo.GetList(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list record with id %s, error %s when calling list repo", listId, err.Error())
		return nil, err
	}

	return s.lConverter.ToModel(listEntity), nil
}

func (s *service) GetListOwnerRecord(ctx context.Context, listId string) (*models.User, error) {
	log.C(ctx).Infof("getting list owner of list with id %s from list service", listId)

	entityOwner, err := s.lRepo.GetListOwner(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list owner, error %s when calling list repo", err.Error())
		return nil, err
	}

	return s.uConverter.ToModel(entityOwner), nil
}
