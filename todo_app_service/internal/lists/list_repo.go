package lists

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/persistence"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
	"database/sql"
	"errors"
)

type repository struct{}

func NewRepo() *repository {
	return &repository{}
}

func (*repository) GetList(ctx context.Context, listId string) (*entities.List, error) {
	log.C(ctx).Infof("getting list with id %s from list repository", listId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from ctx, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `SELECT id, name, created_at, last_updated, owner, description 
FROM lists where id = $1`

	var entity entities.List
	if err = persist.GetContext(ctx, &entity, sqlQueryString, listId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("failed to get list with id %s due to sqlErrNoRows", listId)
			return nil, application_errors.NewNotFoundError(constants.LIST_TARGET, listId)
		}

		log.C(ctx).Errorf("failed to get list with id %s due to database error", listId)
		return nil, err
	}
	return &entity, nil
}

func (*repository) DeleteList(ctx context.Context, listID string) error {
	log.C(ctx).Infof("deleting list with id %s from list repository", listID)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from ctx, error %s", err.Error())
		return err
	}

	sqlQueryString := `DELETE FROM lists WHERE id = $1`

	_, err = persist.ExecContext(ctx, sqlQueryString, listID)
	if err != nil {
		log.C(ctx).Errorf("failed to delete list with id %s due to a failiure in the execution of the sql query", listID)
		return err
	}

	return nil
}

func (r *repository) CreateList(ctx context.Context, entity *entities.List) (*entities.List, error) {
	log.C(ctx).Info("creating list in list repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from ctx, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `INSERT INTO lists (id, name, created_at, last_updated, owner, description) 
						VALUES (:id,:name,:created_at,:last_updated,:owner,:description)`

	_, err = persist.NamedExecContext(ctx, sqlQueryString, entity)
	if err != nil {
		log.C(ctx).Errorf("failed to create list due to a failure in the execution of the sql query %s", err.Error())
		return nil, persistence.MapPostgresListErrorToError(err, entity)
	}

	return r.GetList(ctx, entity.Id.String())
}

func (*repository) DeleteLists(ctx context.Context) error {
	log.C(ctx).Info("deleting all lists in list repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from ctx, error %s", err.Error())
		return err
	}

	sqlQueryString := `DELETE FROM lists`
	if _, err = persist.ExecContext(ctx, sqlQueryString); err != nil {
		log.C(ctx).Errorf("failed to delete lists, error %s when trying to execute sql query", err.Error())
		return err
	}

	return nil
}

func (*repository) GetLists(ctx context.Context, queryBuilder sql_query_decorators.SqlQueryRetriever) ([]entities.List, error) {
	log.C(ctx).Info("getting all lists from list repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from ctx, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := queryBuilder.DetermineCorrectSqlQuery(ctx)

	var entities []entities.List
	if err = persist.SelectContext(ctx, &entities, sqlQueryString); err != nil {
		log.C(ctx).Errorf("failed to parse list entities due to a failure in the execution of the sql query %s", err.Error())
		return nil, err
	}

	return entities, nil
}

func (r *repository) UpdateList(ctx context.Context, sqlExecParams map[string]interface{}, sqlFields []string) (*entities.List, error) {
	listId := sqlExecParams["id"].(string)
	log.C(ctx).Infof("updating list with id %s in list repository", listId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from ctx, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := parseSqlUpdateListQuery(sqlFields)

	res, err := persist.NamedExecContext(ctx, sqlQueryString, sqlExecParams)
	if err != nil {
		log.C(ctx).Errorf("failed to update list, error when executing sql query %s", err.Error())
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.C(ctx).Error("failed to update list, error when trying to get the number of rows affected")
		return nil, err
	}

	if rowsAffected == 0 {
		log.C(ctx).Errorf("failed to update list, invalid list_id provided %s", listId)
		return nil, application_errors.NewNotFoundError(constants.LIST_TARGET, listId)
	}

	return r.GetList(ctx, listId)
}

func (*repository) UpdateListSharedWith(ctx context.Context, listId string, userId string) error {
	log.C(ctx).Infof("adding a user collaborator with id %s to a list with id %s in list repository", userId, listId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from ctx, error %s", err.Error())
		return err
	}

	sqlInsertIntoUserListsTableQuery := `INSERT INTO user_lists (user_id, list_id) VALUES ($1,$2)`

	res, err := persist.ExecContext(ctx, sqlInsertIntoUserListsTableQuery, userId, listId)
	if err != nil {
		log.C(ctx).Debug("failed to add a new user collaborator due to an error when executing sql query")
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.C(ctx).Errorf("failed to add collaborator to list with id %s, error when trying to get the number of rows affected", listId)
		return err
	}

	if rowsAffected == 0 {
		log.C(ctx).Debug("failed to add a new user collaborator due to an error caused by the number of rows affected being equal to 0")
		return application_errors.NewNotFoundError(constants.LIST_TARGET, listId)
	}

	return nil
}

func (*repository) DeleteCollaborator(ctx context.Context, listId string, userId string) error {
	log.C(ctx).Infof("deleting a collaborator with id %s form list with id %s", userId, listId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from ctx, error %s", err.Error())
		return err
	}

	sqlQueryString := `DELETE FROM user_lists WHERE list_id = $1 AND user_id = $2`

	_, err = persist.ExecContext(ctx, sqlQueryString, listId, userId)
	if err != nil {
		log.C(ctx).Errorf("failed to delete collaborator with id %s from list with id %s, error when executing sql query", userId, listId)
		return err
	}

	return nil
}

func (*repository) GetListCollaborators(ctx context.Context, listId string, sqlQueryBuilder sql_query_decorators.SqlQueryRetriever) ([]entities.User, error) {
	log.C(ctx).Info("getting list's collaborators in list repository")

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from ctx, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := sqlQueryBuilder.DetermineCorrectSqlQuery(ctx)

	var collaborators []entities.User
	if err = persist.SelectContext(ctx, &collaborators, sqlQueryString, listId); err != nil {
		log.C(ctx).Errorf("failed to get list's collaborators due to an error %s in the execution of the sql query", err.Error())
		return nil, err
	}
	return collaborators, nil
}

func (*repository) GetListOwner(ctx context.Context, listId string) (*entities.User, error) {
	log.C(ctx).Infof("getting the owner of a list with id %s", listId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from ctx, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `SELECT users.id,users.role,users.email FROM lists 
    JOIN users ON lists.owner = users.id WHERE lists.id = $1`

	entityOwner := &entities.User{}
	if err = persist.GetContext(ctx, entityOwner, sqlQueryString, listId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("failed to get list owner due to an sqlErrNoRows error %s", err.Error())
			return nil, application_errors.NewNotFoundError(constants.LIST_TARGET, listId)
		}
		log.C(ctx).Errorf("failed to get list owner due to an error %s caused by the execution of the sql query", err.Error())
		return nil, err
	}

	return entityOwner, nil
}

func (*repository) CheckWhetherUserIsCollaborator(ctx context.Context, listId string, userId string) (bool, error) {
	log.C(ctx).Infof("checking whether user with id %s is collaborator of list with id %s", userId, listId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistence from ctx, error %s", err.Error())
		return false, err
	}

	sqlQueryString := `SELECT 1 FROM lists JOIN user_lists ON lists.id = user_lists.list_id
WHERE list_id = $1 AND user_id = $2`

	res, err := persist.ExecContext(ctx, sqlQueryString, listId, userId)
	if err != nil {
		log.C(ctx).Errorf("failed to check whether user %s is collaborator in list %s, error %s when executing sql query", userId, listId, err.Error())
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.C(ctx).Errorf("failed to check whether user %s is collaborator in list %s, error %s when trying to check the numner of rows affected", userId, listId, err.Error())
		return false, err
	}

	if rowsAffected == 0 {
		log.C(ctx).Debugf("user is not a collaborator, the number of rows affected being equal to 0...")
		return false, nil
	}

	return true, nil
}
