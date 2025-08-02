package lists

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/persistence"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
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
	GetLists(context.Context, sql_query_decorators.SqlQueryRetriever) ([]entities.List, error)
	GetList(context.Context, string) (*entities.List, error)
	GetListCollaborators(context.Context, string, sql_query_decorators.SqlQueryRetriever) ([]entities.User, error)
	GetListOwner(context.Context, string) (*entities.User, error)
	DeleteList(context.Context, string) error
	DeleteLists(context.Context) error
	CreateList(context.Context, *entities.List) (*entities.List, error)
	UpdateList(context.Context, map[string]interface{}, []string) (*entities.List, error)
	UpdateListSharedWith(context.Context, string, string) error
	DeleteCollaborator(context.Context, string, string) error
	CheckWhetherUserIsCollaborator(context.Context, string, string) (bool, error)
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
	ToModel(*entities.List) *models.List
	ToEntity(*models.List) *entities.List
	ManyToModel([]entities.List) *models.ListPage
	FromUpdateHandlerModelToModel(*handler_models.UpdateList) *models.List
	FromCreateHandlerModelToModel(*handler_models.CreateList) *models.List
}

//go:generate mockery --name=userConverter --exported --output=./mocks --outpkg=mocks --filename=user_converter.go --with-expecter=true
type userConverter interface {
	ToModel(*entities.User) *models.User
	ToEntity(*models.User) *entities.User
	ManyToModel(users []entities.User) *models.UserPage
}

//go:generate mockery --name=userService --exported --output=./mocks --outpkg=mocks --filename=user_service.go --with-expecter=true
type userRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
}

//go:generate mockery --name=sqlDecoratorFactory --exported --output=./mocks --outpkg=mocks --filename=sql_decorator_factory.go --with-expecter=true
type sqlDecoratorFactory interface {
	CreateSqlDecorator(context.Context, sql_query_decorators.Filters, string) (sql_query_decorators.SqlQueryRetriever, error)
}

type service struct {
	lRepo      listRepo
	uRepo      userRepo
	tRepo      todoRepo
	uuidGen    uuidGenerator
	timeGen    timeGenerator
	lConverter listConverter
	uConverter userConverter
	factory    sqlDecoratorFactory
	transact   persistence.Transactioner
}

func NewService(repo listRepo, uuidGen uuidGenerator, timeGen timeGenerator,
	listConverter listConverter, uRep userRepo, tRepo todoRepo, userConverter userConverter, factory sqlDecoratorFactory, transact persistence.Transactioner) *service {
	return &service{
		lRepo:      repo,
		uuidGen:    uuidGen,
		timeGen:    timeGen,
		lConverter: listConverter,
		uRepo:      uRep,
		uConverter: userConverter,
		factory:    factory,
		tRepo:      tRepo,
		transact:   transact,
	}
}

func (s *service) GetListsRecords(ctx context.Context, lFilters *filters.BaseFilters) (*models.ListPage, error) {
	log.C(ctx).Info("getting lists from list service")

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin ctx in list service when trying to delete collaorator, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	baseQuery := `WITH sorted_lists AS(
					SELECT * FROM lists ORDER BY id
 				)
				SELECT id, name, created_at, last_updated, owner, description, COUNT(*) OVER() AS total_count FROM sorted_lists`

	retriever, err := s.factory.CreateSqlDecorator(ctx, lFilters, baseQuery)
	if err != nil {
		log.C(ctx).Error("failed to determine sql query, error when calling decorator factory")
		return nil, err
	}

	listEntities, err := s.lRepo.GetLists(ctx, retriever)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists records, error %s when calling list repo", err.Error())
		return nil, err
	}

	modelLists := s.lConverter.ManyToModel(listEntities)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get lists, error %s", err.Error())
		return nil, nil
	}

	return modelLists, nil
}

func (s *service) DeleteListRecord(ctx context.Context, listId string) error {
	log.C(ctx).Infof("deleting with id %s list record from list service layer", listId)

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin ctx in list service when trying to delete collaorator, error %s", err.Error())
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = s.lRepo.DeleteList(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to delete list with id %s, error %s when calling list repo", listId, err.Error())
		return fmt.Errorf("failed to delete list with id %s", listId)
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to delete list with id %s, error %s", listId, err.Error())
		return err
	}

	return nil
}

func (s *service) CreateListRecord(ctx context.Context, list *handler_models.CreateList, owner string) (*models.List, error) {
	log.C(ctx).Info("creating list record in list service layer")

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin ctx in list service when trying to delete collaorator, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if _, err = s.lRepo.CreateList(ctx, convertedEntity); err != nil {
		log.C(ctx).Errorf("failed to create list, error %s when calling list repo", err.Error())
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to create list with id %s, error %s", newlyCreatedList.Id, err.Error())
		return nil, err
	}

	return newlyCreatedList, nil
}

func (s *service) UpdateListPartiallyRecord(ctx context.Context, listId string, list *handler_models.UpdateList) (*models.List, error) {
	log.C(ctx).Infof("updating list with id %s in list service ", listId)

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin ctx in list service when trying to delete collaorator, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	modelList := s.lConverter.FromUpdateHandlerModelToModel(list)

	sqlExecParams := map[string]interface{}{"id": listId}
	sqlFields := make([]string, 0, 8)

	determineSqlFieldsAndParamsList(modelList, sqlExecParams, &sqlFields)

	entity, err := s.lRepo.UpdateList(ctx, sqlExecParams, sqlFields)
	if err != nil {
		log.C(ctx).Errorf("failed to update list name %s", err.Error())
		return nil, err
	}

	readyModel := s.lConverter.ToModel(entity)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to update list with id %s partially, error %s", list, err.Error())
		return nil, err
	}

	return readyModel, nil
}

func (s *service) AddCollaborator(ctx context.Context, listId string, userEmail string) (*models.User, error) {
	log.C(ctx).Debugf("adding collaborator with email %s in list with id %s", userEmail, listId)

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin ctx in list service when trying to delete collaorator, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	user, err := s.uRepo.GetUserByEmail(ctx, userEmail)
	if err != nil {
		log.C(ctx).Errorf("failed to get user with email %s, error when calling user repo function", userEmail)
		return nil, err
	}

	if err = s.lRepo.UpdateListSharedWith(ctx, listId, user.Id.String()); err != nil {
		log.C(ctx).Errorf("failed to add collaborator with email %s in list with id %s, error when calling repo function", userEmail, listId)
		return nil, err
	}

	modelUser := s.uConverter.ToModel(user)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to add user with email %s as collaborator in list with id %s, error %s", userEmail, listId, err.Error())
		return nil, err
	}

	return modelUser, nil
}

func (s *service) DeleteCollaborator(ctx context.Context, listId string, userId string) error {
	log.C(ctx).Infof("deleting a collaborator with id %s from list with id %s in list service", userId, listId)

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin ctx in list service when trying to delete collaorator, error %s", err.Error())
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if _, err = s.lRepo.GetList(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to delete collaborator with id %s, error %s when calling list repo", userId, err.Error())
		return err
	}

	if err = s.lRepo.DeleteCollaborator(ctx, listId, userId); err != nil {
		log.C(ctx).Errorf("failed to delete a collaborator with id %s from a list with id %s, error in list repo", userId, listId)
		return err
	}

	if err = s.tRepo.UnassignUserFromTodos(ctx, userId, listId); err != nil {
		log.C(ctx).Errorf("failed to unassign user with id %s from todos from list with id %s, error %s", userId, listId, err.Error())
		return err
	}

	return tx.Commit()
}

func (s *service) DeleteLists(ctx context.Context) error {
	log.C(ctx).Info("deleting lists in list service")

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin ctx in list service when trying to delete collaorator, error %s", err.Error())
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = s.lRepo.DeleteLists(ctx); err != nil {
		log.C(ctx).Errorf("failed to delete lists, error %s", err.Error())
		return err
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transation when trying to delete lists, error %s", err.Error())
		return err
	}

	return nil
}

func (s *service) GetCollaborators(ctx context.Context, listId string, lFilters *filters.BaseFilters) (*models.UserPage, error) {
	log.C(ctx).Info("getting list collaborators from list service")

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin ctx in list service when trying to delete collaorator, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if _, err = s.lRepo.GetList(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to get list collaborators, error %s when calling list repo", err.Error())
		return nil, err
	}

	baseQuery := `WITH sorted_user_cte AS( 
					SELECT users.id,users.email,users.role,list_id FROM users 
					JOIN user_lists ON users.id = user_lists.user_id ORDER BY users.id
				   ) 
					SELECT id, email, role, COUNT(*) OVER() AS total_count FROM sorted_user_cte WHERE list_id = $1`

	sqlQueryBuilder, err := s.factory.CreateSqlDecorator(ctx, lFilters, baseQuery)
	if err != nil {
		log.C(ctx).Errorf("failed to get collaborators, error when calling decorator factory")
		return nil, err
	}

	entitiesCollaborators, err := s.lRepo.GetListCollaborators(ctx, listId, sqlQueryBuilder)
	if err != nil {
		log.C(ctx).Errorf("failed to get list collaborators, error %s when calling list repo", err.Error())
		return nil, err
	}

	modelCollaborators := s.uConverter.ManyToModel(entitiesCollaborators)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get collaborators of list with id %s, error %s", listId, err.Error())
		return nil, err
	}

	return modelCollaborators, nil
}

func (s *service) GetListRecord(ctx context.Context, listId string) (*models.List, error) {
	log.C(ctx).Infof("getting list with id %s from list service", listId)

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin ctx in list service when trying to delete collaorator, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	listEntity, err := s.lRepo.GetList(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list record with id %s, error %s when calling list repo", listId, err.Error())
		return nil, err
	}

	modelList := s.lConverter.ToModel(listEntity)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get list with id %s, error %s", listId, err.Error())
		return nil, err
	}

	return modelList, nil
}

func (s *service) GetListOwnerRecord(ctx context.Context, listId string) (*models.User, error) {
	log.C(ctx).Infof("getting list owner of list with id %s from list service", listId)

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin ctx in list service when trying to delete collaorator, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	entityOwner, err := s.lRepo.GetListOwner(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list owner, error %s when calling list repo", err.Error())
		return nil, err
	}

	modelOwner := s.uConverter.ToModel(entityOwner)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get the owner of list with id %s, error %s", listId, err.Error())
		return nil, err
	}

	return modelOwner, nil
}
