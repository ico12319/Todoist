package refresh

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/persistence"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
)

type repository struct{}

func NewRepo() *repository {
	return &repository{}
}

func (*repository) CreateRefreshToken(ctx context.Context, refresh *entities.Refresh) (*entities.Refresh, error) {
	log.C(ctx).Infof("creating refresh token for user with id %s", refresh.UserId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistance from context in refresh repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `INSERT INTO user_refresh_tokens(refresh_token, user_id) 
VALUES (:refresh_token, :user_id)`

	if _, err = persist.NamedExecContext(ctx, sqlQueryString, refresh); err != nil {
		log.C(ctx).Errorf("failed to create refresh token, error %s", err.Error())
		return nil, persistence.MapPostgresNonExistingUserInUserTable(err, refresh)
	}

	return refresh, nil
}

func (*repository) UpdateRefreshToken(ctx context.Context, refreshToken string, userId string) (*entities.Refresh, error) {
	log.C(ctx).Infof("updating refresh token of user with id %s", userId)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistance from context in refresh repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `UPDATE user_refresh_tokens SET refresh_token = $1 WHERE user_id = $2
RETURNING refresh_token, user_id`

	var returnedToken string
	var id uuid.UUID

	row := persist.QueryRowContext(ctx, sqlQueryString, refreshToken, userId)
	if err = row.Scan(&returnedToken, &id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("no rows updated for user_id=%s", userId)
			return nil, application_errors.NewNotFoundError(constants.USER_TARGET, userId)
		}
		log.C(ctx).Errorf("failed to scan returning, error %s", err.Error())
		return nil, err
	}

	return &entities.Refresh{
		RefreshToken: returnedToken,
		UserId:       id,
	}, nil
}

func (*repository) GetTokenOwner(ctx context.Context, refreshToken string) (*entities.User, error) {
	log.C(ctx).Infof("getting token %s owner", refreshToken)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to get persistance from context in refresh repo, error %s", err.Error())
		return nil, err
	}

	sqlQueryString := `SELECT id, email, role FROM users JOIN user_refresh_tokens ON users.id = user_refresh_tokens.user_id
WHERE refresh_token = $1`

	tokenOwner := &entities.User{}
	if err = persist.GetContext(ctx, tokenOwner, sqlQueryString, refreshToken); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("failed to get token owner, error %s when executing sql query", err.Error())
			return nil, fmt.Errorf("invalid refresh token %s", refreshToken)
		}
		log.C(ctx).Error("failed to get token owner, db error...")
		return nil, err
	}

	return tokenOwner, nil
}
