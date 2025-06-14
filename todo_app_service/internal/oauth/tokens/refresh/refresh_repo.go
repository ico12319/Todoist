package refresh

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/application_errors"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/entities"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/constants"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type sqlRefreshDB struct {
	db *sqlx.DB
}

func NewSqlRefreshDB(db *sqlx.DB) *sqlRefreshDB {
	return &sqlRefreshDB{db: db}
}

func (s *sqlRefreshDB) CreateRefreshToken(ctx context.Context, refresh *entities.Refresh) (*entities.Refresh, error) {
	log.C(ctx).Infof("creating refresh token for user with id %s", refresh.UserId)

	sqlQueryString := `INSERT INTO user_refresh_tokens(refresh_token, user_id) 
VALUES (:refresh_token, :user_id) ON CONFLICT (user_id) DO UPDATE SET refresh_token = EXCLUDED.refresh_token`

	if _, err := s.db.NamedExecContext(ctx, sqlQueryString, refresh); err != nil {
		log.C(ctx).Errorf("failed to create refresh token, error %s", err.Error())
		return nil, fmt.Errorf("error when trying to create refresh token")
	}

	return refresh, nil
}

func (s *sqlRefreshDB) UpdateRefreshToken(ctx context.Context, refreshToken string, userId string) (*entities.Refresh, error) {
	log.C(ctx).Infof("updating refresh token of user with id %s", userId)

	sqlQueryString := `UPDATE user_refresh_tokens SET refresh_token = $1 WHERE user_id = $2
RETURNING refresh_token, user_id`

	var returnedToken string
	var id uuid.UUID

	row := s.db.QueryRowContext(ctx, sqlQueryString, refreshToken, userId)
	if err := row.Scan(&returnedToken, &id); err != nil {
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

func (s *sqlRefreshDB) GetTokenOwner(ctx context.Context, refreshToken string) (*entities.User, error) {
	log.C(ctx).Infof("getting token %s owner", refreshToken)

	sqlQueryString := `SELECT id, email, role FROM users JOIN user_refresh_tokens ON users.id = user_refresh_tokens.user_id
WHERE refresh_token = $1`

	tokenOwner := &entities.User{}
	if err := s.db.Get(tokenOwner, sqlQueryString, refreshToken); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.C(ctx).Errorf("failed to get token owner, error %s when executing sql query", err.Error())
			return nil, fmt.Errorf("invalid refresh token %s", refreshToken)
		}
		log.C(ctx).Error("failed to get token owner, db error...")
		return nil, err
	}

	return tokenOwner, nil
}
