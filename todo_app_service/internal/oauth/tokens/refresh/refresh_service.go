package refresh

import (
	"context"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/entities"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/models"
)

type refreshRepository interface {
	CreateRefreshToken(ctx context.Context, refresh *entities.Refresh) (*entities.Refresh, error)
	UpdateRefreshToken(ctx context.Context, refreshToken string, userId string) (*entities.Refresh, error)
	GetTokenOwner(ctx context.Context, refreshToken string) (*entities.User, error)
}

type converter interface {
	ToEntity(refresh *models.Refresh) *entities.Refresh
	ToModel(refresh *entities.Refresh) *models.Refresh
}

type userService interface {
	GetUserRecordByEmail(ctx context.Context, email string) (*models.User, error)
}

type userConverter interface {
	ConvertFromDBEntityToModel(user *entities.User) *models.User
}

type builder interface {
	GenerateRefreshToken(ctx context.Context) (string, error)
}

type service struct {
	tokenBuilder builder
	uService     userService
	repo         refreshRepository
	conv         converter
	uConverter   userConverter
}

func NewService(repo refreshRepository, uService userService, conv converter, uConverter userConverter, tokenBuilder builder) *service {
	return &service{repo: repo, uService: uService, conv: conv, uConverter: uConverter, tokenBuilder: tokenBuilder}
}

func (s *service) CreateRefreshToken(ctx context.Context, email string, refreshToken string) (*models.Refresh, error) {
	log.C(ctx).Info("creating refresh token in refresh service")

	user, err := s.uService.GetUserRecordByEmail(ctx, email)
	if err != nil {
		log.C(ctx).Errorf("failed to create refresh token, error %s when trying to get user by email %s", err.Error(), email)
		return nil, err
	}

	refreshModel := &models.Refresh{
		UserId:       user.Id,
		RefreshToken: refreshToken,
	}

	convertedEntity := s.conv.ToEntity(refreshModel)

	if _, err = s.repo.CreateRefreshToken(ctx, convertedEntity); err != nil {
		log.C(ctx).Errorf("failed to create refresh token, error %s", err.Error())
		return nil, err
	}

	return refreshModel, nil
}

func (s *service) UpdateRefreshToken(ctx context.Context, refreshToken string, userId string) (*models.Refresh, error) {
	log.C(ctx).Infof("updating refresh token %s of user with id %s", refreshToken, userId)

	refreshEntity, err := s.repo.UpdateRefreshToken(ctx, refreshToken, userId)
	if err != nil {
		log.C(ctx).Errorf("failed to update refresh token, error %s when callin refresh repo", err.Error())
		return nil, err
	}

	return s.conv.ToModel(refreshEntity), nil
}

func (s *service) GetTokenOwner(ctx context.Context, refreshToken string) (*models.User, error) {
	log.C(ctx).Info("getting refresh token owner in refresh service")

	ownerEntity, err := s.repo.GetTokenOwner(ctx, refreshToken)
	if err != nil {
		log.C(ctx).Errorf("failed to get token owner, error %s in refresh service", err.Error())
		return nil, err
	}

	return s.uConverter.ConvertFromDBEntityToModel(ownerEntity), nil
}
