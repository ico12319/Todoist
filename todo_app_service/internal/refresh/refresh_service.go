package refresh

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
)

//go:generate mockery --name=refreshRepository --exported --output=./mocks --outpkg=mocks --filename=refresh_repository.go --with-expecter=true
type refreshRepository interface {
	CreateRefreshToken(context.Context, *entities.Refresh) (*entities.Refresh, error)
	UpdateRefreshToken(context.Context, string, string) (*entities.Refresh, error)
	GetTokenOwner(context.Context, string) (*entities.User, error)
}

//go:generate mockery --name=converter --exported --output=./mocks --outpkg=mocks --filename=converter.go --with-expecter=true
type converter interface {
	ToEntity(*models.Refresh) *entities.Refresh
	ToModel(*entities.Refresh) *models.Refresh
}

//go:generate mockery --name=userService --exported --output=./mocks --outpkg=mocks --filename=user_service.go --with-expecter=true
type userService interface {
	GetUserRecordByEmail(context.Context, string) (*models.User, error)
}

//go:generate mockery --name=userConverter --exported --output=./mocks --outpkg=mocks --filename=user_converter.go --with-expecter=true
type userConverter interface {
	ConvertFromDBEntityToModel(*entities.User) *models.User
}

type service struct {
	uService   userService
	repo       refreshRepository
	conv       converter
	uConverter userConverter
}

func NewService(repo refreshRepository, uService userService, conv converter, uConverter userConverter) *service {
	return &service{repo: repo, uService: uService, conv: conv, uConverter: uConverter}
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

func (s *service) UpsertRefreshToken(ctx context.Context, refresh *models.Refresh, userEmail string) error {
	if _, err := s.repo.UpdateRefreshToken(ctx, refresh.RefreshToken, refresh.UserId); err != nil {
		if utils.CheckForNotFoundError(err) {

			if _, err = s.CreateRefreshToken(ctx, userEmail, refresh.RefreshToken); err != nil {
				log.C(ctx).Errorf("faile to create refresh token in jwt issuer, error %s", err.Error())
				return err
			}

		} else {

			log.C(ctx).Errorf("failed to update refresh token, error %s when generating...", err.Error())
			return err

		}
	}

	return nil
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
