package generic

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/persistence"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type repository struct{}

func NewRepo() *repository {
	return &repository{}
}

func (*repository) GetPaginationInfo(ctx context.Context, sourceName string, filter string, params []interface{}) (*entities.PaginationInfo, error) {
	log.C(ctx).Infof("getting pagination info from source %s", sourceName)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persitance from context in generic repo, error %s", err.Error())
		return nil, err
	}

	sqlQuery := fmt.Sprintf(
		`WITH filtered AS(
						SELECT id FROM %s %s 
					)
					 SELECT
    				(SELECT id FROM filtered
					ORDER BY id ASC LIMIT 1) AS first_id,
					(SELECT id FROM filtered
    				ORDER BY id DESC LIMIT 1) AS last_id,
    				COUNT(*) AS total_count FROM filtered`, sourceName, filter)

	var paginationInfo entities.PaginationInfo
	if err = persist.GetContext(ctx, &paginationInfo, sqlQuery, params...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Warn("empty table ...")
			return nil, nil
		}
		log.C(ctx).Errorf("failed to get pagination info ids in generic repo, error %v when trying to execute sql query", err)
		return nil, err
	}

	return &paginationInfo, nil
}
