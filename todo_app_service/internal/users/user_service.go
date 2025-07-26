package users

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/persistence"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
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
	GetTodosAssignedToUser(context.Context, string, sqlQueryRetriever) ([]entities.Todo, error)
	GetUserLists(context.Context, sqlQueryRetriever) ([]entities.List, error)
}

type userConverter interface {
	ToModel(*entities.User) *models.User
	ToEntity(*models.User) *entities.User
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
	transact   persistence.Transactioner
}

func NewService(repo userRepo, converter userConverter, lConverter listConverter, tConverter todoConverter, uuidGen uuidGenerator, factory sqlDecoratorFactory, transact persistence.Transactioner) *service {
	return &service{
		repo:       repo,
		converter:  converter,
		lConverter: lConverter,
		tConverter: tConverter,
		uuidGen:    uuidGen,
		factory:    factory,
		transact:   transact,
	}
}

func (s *service) GetUsersRecords(ctx context.Context, uFilters *filters.UserFilters) ([]*models.User, error) {
	log.C(ctx).Info("getting users in user service")

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transactin in user service, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	baseQuery := `SELECT id,email,role FROM (SELECT id, email, role FROM users ORDER BY id)`

	decorator, err := s.factory.CreateSqlDecorator(ctx, uFilters, baseQuery)
	if err != nil {
		log.C(ctx).Errorf("failed to get users, error when calling factory function")
		return nil, err
	}

	userEntities, err := s.repo.GetUsers(ctx, decorator)
	if err != nil {
		log.C(ctx).Errorf("failed to get users in user service due to an error in user repository %s", err.Error())
		return nil, err
	}

	modelUsers := s.converter.ManyToModel(userEntities)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get user records, error %s", err.Error())
		return nil, err
	}

	return modelUsers, nil
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

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transactin in user service, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	userEntity, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		log.C(ctx).Errorf("failed to get user by email %s in user service due to an error in user repository", email)
		return nil, err
	}

	userModel := s.converter.ToModel(userEntity)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get user by email %s, error %s", userModel, err.Error())
		return nil, err
	}

	return userModel, nil
}

func (s *service) CreateUserRecord(ctx context.Context, user *handler_models.CreateUser) (*models.User, error) {
	log.C(ctx).Info("creating user in user service")

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transactin in user service, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	createModel := s.converter.ConvertFromCreateHandlerModelToModel(user)
	createModel.Id = s.uuidGen.Generate()

	convertedEntity := s.converter.ToEntity(createModel)

	_, err = s.repo.CreateUser(ctx, convertedEntity)
	if err != nil {
		log.C(ctx).Errorf("failed to create user in user service, error %s when calling user repo", err.Error())
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to create user, error %s", err.Error())
		return nil, err
	}

	return createModel, nil
}

func (s *service) DeleteUserRecord(ctx context.Context, id string) error {
	log.C(ctx).Infof("deleting user with id %s in user service", id)

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transactin in user service, error %s", err.Error())
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = s.repo.DeleteUser(ctx, id); err != nil {
		log.C(ctx).Errorf("failed to update user in user service, error %s when calling user repo", err.Error())
		return err
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to delete user with id %s, error %s", id, err.Error())
		return err
	}

	return nil
}

func (s *service) DeleteUsers(ctx context.Context) error {
	log.C(ctx).Info("deleting users in user service")

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transactin in user service, error %s", err.Error())
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = s.repo.DeleteUsers(ctx); err != nil {
		log.C(ctx).Errorf("failed to delete users, error %s", err.Error())
		return err
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to delete users, error %s", err.Error())
		return err
	}

	return nil
}

func (s *service) UpdateUserRecord(ctx context.Context, id string, user *models.User) (*models.User, error) {
	log.C(ctx).Infof("updating user with id %s in user service", id)

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transactin in user service, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedEntity := s.converter.ToEntity(user)

	updatedEntity, err := s.repo.UpdateUser(ctx, id, convertedEntity)
	if err != nil {
		log.C(ctx).Errorf("failed to update user in user service, error %s when calling user repo", err.Error())
		return nil, err
	}

	updatedModel := s.converter.ToModel(updatedEntity)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to update user with id %s, error %s", id, err.Error())
		return nil, err
	}

	return updatedModel, nil
}

func (s *service) UpdateUserRecordPartially(ctx context.Context, id string, user *handler_models.UpdateUser) (*models.User, error) {
	log.C(ctx).Infof("updating user with with id %s in user service", id)

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transactin in user service, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedModel := s.converter.ConvertFromUpdateModelToModel(user)

	sqlExecParams := map[string]interface{}{"id": id}
	sqlFields := make([]string, 0, 8)

	determineSqlFieldsAndParamsUser(convertedModel, sqlExecParams, &sqlFields)

	updatedEntity, err := s.repo.UpdateUserPartially(ctx, sqlExecParams, sqlFields)
	if err != nil {
		log.C(ctx).Errorf("failed to partially update user with id %s, errror when calling user repo %s", id, err.Error())
		return nil, err
	}

	updatedModel := s.converter.ToModel(updatedEntity)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to update user with id %s partially, error %s", id, err.Error())
		return nil, err
	}

	return updatedModel, nil
}

func (s *service) GetUserListsRecords(ctx context.Context, userId string, uFilter *filters.UserFilters) ([]*models.List, error) {
	log.C(ctx).Info("getting lists where user participates in user service")

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transactin in user service, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if _, err := s.repo.GetUser(ctx, userId); err != nil {
		log.C(ctx).Errorf("failed to get user lists, error %s when calling user repo", err.Error())
		return nil, err
	}

	baseQuery := `WITH sorted_lists_and_users AS(
				   SELECT * FROM lists LEFT JOIN user_lists ON
  				   lists.id = user_lists.list_id
  				  )
 				SELECT id, name, created_at, last_updated, owner, description FROM sorted_lists_and_users`

	decorator, err := s.factory.CreateSqlDecorator(ctx, uFilter, baseQuery)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates in, error when calling factory function")
		return nil, err
	}

	listEntities, err := s.repo.GetUserLists(ctx, decorator)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates in, error %s when calling user repo function", err.Error())
		return nil, err
	}

	listModels := s.lConverter.ManyToModel(listEntities)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get user lists, error %s", err.Error())
		return nil, err
	}

	return listModels, nil
}

func (s *service) GetTodosAssignedToUser(ctx context.Context, userId string, userFilters *filters.UserFilters) ([]*models.Todo, error) {
	log.C(ctx).Info("getting todos assigned to user in user service")

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transactin in user service, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if _, err = s.repo.GetUser(ctx, userId); err != nil {
		log.C(ctx).Errorf("failed to get todos assigned to user, error %s when calling user repo", err.Error())
		return nil, err
	}

	baseQuery := `WITH sorted_todo_cte AS ( 
				   SELECT todos.id, todos.name, todos.description, todos.list_id,
				   todos.status,todos.created_at, todos.last_updated, todos.assigned_to,
		           todos.due_date, todos.priority, users.id AS user_id FROM todos
				   JOIN users ON todos.assigned_to = users.id ORDER BY todos.id
 			      )
				SELECT id, name, description, list_id, status, created_at,
				last_updated, assigned_to, due_date, priority FROM sorted_todo_cte WHERE user_id = $1`

	decorator, err := s.factory.CreateSqlDecorator(ctx, userFilters, baseQuery)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user in user service, error when calling factory function")
		return nil, err
	}

	todoEntities, err := s.repo.GetTodosAssignedToUser(ctx, userId, decorator)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user in user service, error when calling repo")
		return nil, err
	}

	todoModels := s.tConverter.ManyToModel(todoEntities)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get todos assigned to user with id %s, error %s", userId, err.Error())
		return nil, err
	}

	return todoModels, nil
}
