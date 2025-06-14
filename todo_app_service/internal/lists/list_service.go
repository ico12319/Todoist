package lists

import (
	"context"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/entities"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/handler_models"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/models"
	"time"
)

//go:generate mockery --name=ListRepo --output=./mocks --outpkg=mocks --filename=list_repo.go --with-expecter=true
type listRepo interface {
	GetLists(ctx context.Context, queryBuilder sqlQueryRetriever) ([]entities.List, error)
	GetList(ctx context.Context, listId string) (*entities.List, error)
	GetListCollaborators(ctx context.Context, sqlQueryBuilder sqlQueryRetriever) ([]entities.User, error)
	GetListOwner(ctx context.Context, listId string) (*entities.User, error)
	DeleteList(ctx context.Context, listID string) error
	DeleteLists(ctx context.Context) error
	CreateList(ctx context.Context, entity *entities.List) (*entities.List, error)
	UpdateList(ctx context.Context, sqlExecParams map[string]interface{}, sqlFields []string) (*entities.List, error)
	UpdateListSharedWith(ctx context.Context, listId string, userId string) error
	DeleteCollaborator(ctx context.Context, listId string, userId string) error
	CheckWhetherUserIsCollaborator(ctx context.Context, listId string, userId string) (bool, error)
}

//go:generate mockery --name=IUuidGenerator --output=./mocks --outpkg=mocks --filename=iuuid_generator.go --with-expecter=true
type uuidGenerator interface {
	Generate() string
}

//go:generate mockery --name=ITimeGenerator --output=./mocks --outpkg=mocks --filename=itime_generator.go --with-expecter=true
type timeGenerator interface {
	Now() time.Time
}

//go:generate mockery --name=IListConverter --output=./mocks --outpkg=mocks --filename=ilist_converter.go --with-expecter=true
type listConverter interface {
	ConvertFromDBEntityToModel(list *entities.List) *models.List
	ConvertFromModelToDBEntity(list *models.List) *entities.List
	ManyToModel(lists []entities.List) []*models.List
	FromUpdateHandlerModelToModel(list *handler_models.UpdateList) *models.List
	FromCreateHandlerModelToModel(list *handler_models.CreateList) *models.List
}

//go:generate mockery --name=IUserConverter --output=./mocks --outpkg=mocks --filename=iuser_converter.go --with-expecter=true
type userConverter interface {
	ConvertFromDBEntityToModel(user *entities.User) *models.User
	ConvertFromModelToEntity(user *models.User) *entities.User
	ManyToModel(users []entities.User) []*models.User
}

type userService interface {
	GetUserRecord(ctx context.Context, userId string) (*models.User, error)
}

type concreteSqlQueryFactory interface {
	CreateListQueryDecorator(ctx context.Context, initialQuery string, listFilters *filters.ListFilters) (sql_query_decorators.SqlQueryRetriever, error)
}

type service struct {
	lRepo      listRepo
	uService   userService
	uuidGen    uuidGenerator
	timeGen    timeGenerator
	lConverter listConverter
	uConverter userConverter
	factory    concreteSqlQueryFactory
}

func NewService(repo listRepo, uuidGen uuidGenerator, timeGen timeGenerator,
	listConverter listConverter, uService userService, userConverter userConverter, factory concreteSqlQueryFactory) *service {
	return &service{lRepo: repo, uuidGen: uuidGen, timeGen: timeGen,
		lConverter: listConverter, uService: uService, uConverter: userConverter, factory: factory}
}

func (s *service) GetListsRecords(ctx context.Context, lFilters *filters.ListFilters) ([]*models.List, error) {
	log.C(ctx).Info("getting lists from list service")

	retriever, err := s.factory.CreateListQueryDecorator(ctx, baseListGetQuery, lFilters)
	if err != nil {
		log.C(ctx).Error("failed to determine sql query, error when calling decorator factory")
		return nil, err
	}

	listEntities, err := s.lRepo.GetLists(ctx, retriever)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists records, error %s when calling list repo", err.Error())
		return nil, err
	}

	return s.lConverter.ManyToModel(listEntities), nil
}

func (s *service) DeleteListRecord(ctx context.Context, listId string) error {
	log.C(ctx).Infof("deleting with id %s list record from list service layer", listId)

	if err := s.lRepo.DeleteList(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to delete list with id %s, error %s when calling list repo", listId, err.Error())
		return err
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

	convertedEntity := s.lConverter.ConvertFromModelToDBEntity(newlyCreatedList)

	resEntity, err := s.lRepo.CreateList(ctx, convertedEntity)
	if err != nil {
		log.C(ctx).Errorf("failed to create list, error %s when calling list repo", err.Error())
		return nil, err
	}

	return s.lConverter.ConvertFromDBEntityToModel(resEntity), nil
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

	return s.lConverter.ConvertFromDBEntityToModel(entity), nil
}

func (s *service) AddCollaborator(ctx context.Context, listId string, userId string) (*models.User, error) {
	log.C(ctx).Debugf("adding collaborator with id %s in list with id %s", userId, listId)

	user, err := s.uService.GetUserRecord(ctx, userId)
	if err != nil {
		log.C(ctx).Errorf("failed to get user with id %s, error when calling user repo function", userId)
		return nil, err
	}

	if err = s.lRepo.UpdateListSharedWith(ctx, listId, userId); err != nil {
		log.C(ctx).Errorf("failed to add collaborator with id %s in list with id %s, error when calling repo function", userId, listId)
		return nil, err
	}

	return user, nil
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

	return nil
}

func (s *service) DeleteLists(ctx context.Context) error {
	log.C(ctx).Info("deleting lists in list service")

	return s.lRepo.DeleteLists(ctx)
}

func (s *service) GetCollaborators(ctx context.Context, lFilters *filters.ListFilters) ([]*models.User, error) {
	log.C(ctx).Info("getting list collaborators from list service")

	if _, err := s.lRepo.GetList(ctx, lFilters.ListId); err != nil {
		log.C(ctx).Errorf("failed to get list collaborators, error %s when calling list repo", err.Error())
		return nil, err
	}

	sqlQueryBuilder, err := s.factory.CreateListQueryDecorator(ctx, baseCollaboratorsGetQuery, lFilters)
	if err != nil {
		log.C(ctx).Errorf("failed to get collaborators, error when calling decorator factory")
		return nil, err
	}

	entitiesCollaborators, err := s.lRepo.GetListCollaborators(ctx, sqlQueryBuilder)
	if err != nil {
		log.C(ctx).Errorf("failed to get list collaborators, error %s when calling list repo", err.Error())
		return nil, err
	}

	return s.uConverter.ManyToModel(entitiesCollaborators), nil
}

func (s *service) GetListRecord(ctx context.Context, listId string) (*models.List, error) {
	log.C(ctx).Infof("getting list with id %s from list service", listId)

	listEntity, err := s.lRepo.GetList(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list record with id %s, error %s when calling list repo", listId, err.Error())
		return nil, err
	}

	return s.lConverter.ConvertFromDBEntityToModel(listEntity), nil
}

func (s *service) GetListOwnerRecord(ctx context.Context, listId string) (*models.User, error) {
	log.C(ctx).Infof("getting list owner of list with id %s from list service", listId)

	entityOwner, err := s.lRepo.GetListOwner(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list owner, error %s when calling list repo", err.Error())
		return nil, err
	}

	return s.uConverter.ConvertFromDBEntityToModel(entityOwner), nil
}

func (s *service) CheckWhetherUserIsCollaborator(ctx context.Context, listId string, userId string) (bool, error) {
	log.C(ctx).Infof("checking whether a user with id %s is collaborator in list with id %s", userId, listId)

	if _, err := s.uService.GetUserRecord(ctx, userId); err != nil {
		log.C(ctx).Errorf("failed to check whether user with id %s is collaborator, error %s when trying to get it", userId, err.Error())
		return false, err
	}

	return s.lRepo.CheckWhetherUserIsCollaborator(ctx, listId, userId)
}
