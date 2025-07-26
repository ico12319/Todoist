package refresh

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/persistence"
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
type userRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
}

//go:generate mockery --name=userConverter --exported --output=./mocks --outpkg=mocks --filename=user_converter.go --with-expecter=true
type userConverter interface {
	ToModel(*entities.User) *models.User
}

type service struct {
	uRepo      userRepo
	repo       refreshRepository
	conv       converter
	uConverter userConverter
	transact   persistence.Transactioner
}

func NewService(repo refreshRepository, uRepo userRepo, conv converter, uConverter userConverter, transact persistence.Transactioner) *service {
	return &service{
		repo:       repo,
		uRepo:      uRepo,
		conv:       conv,
		uConverter: uConverter,
		transact:   transact,
	}
}

func (s *service) CreateRefreshToken(ctx context.Context, email string, refreshToken string) (*models.Refresh, error) {
	log.C(ctx).Info("creating refresh token in refresh service")

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transact in refresh service, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	user, err := s.uRepo.GetUserByEmail(ctx, email)
	if err != nil {
		log.C(ctx).Errorf("failed to create refresh token, error %s when trying to get user by email %s", err.Error(), email)
		return nil, err
	}

	refreshModel := &models.Refresh{
		UserId:       user.Id.String(),
		RefreshToken: refreshToken,
	}

	convertedEntity := s.conv.ToEntity(refreshModel)

	if _, err = s.repo.CreateRefreshToken(ctx, convertedEntity); err != nil {
		log.C(ctx).Errorf("failed to create refresh token, error %s", err.Error())
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction in refresh service, error %s", err.Error())
		return nil, err
	}

	return refreshModel, nil
}

func (s *service) UpdateRefreshToken(ctx context.Context, refreshToken string, userId string) (*models.Refresh, error) {
	log.C(ctx).Infof("updating refresh token %s of user with id %s", refreshToken, userId)

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transact in refresh service, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	refreshEntity, err := s.repo.UpdateRefreshToken(ctx, refreshToken, userId)
	if err != nil {
		log.C(ctx).Errorf("failed to update refresh token, error %s when callin refresh repo", err.Error())
		return nil, err
	}

	refreshModel := s.conv.ToModel(refreshEntity)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction in refresh service, error %s", err.Error())
		return nil, err
	}

	return refreshModel, nil
}

func (s *service) UpsertRefreshToken(ctx context.Context, refresh *models.Refresh, userEmail string) error {
	log.C(ctx).Infof("upserting refresh token of user with email %s in refresh service", userEmail)

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transact in refresh service, error %s", err.Error())
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if _, err = s.repo.UpdateRefreshToken(ctx, refresh.RefreshToken, refresh.UserId); err != nil {
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

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction in refresh service, error %s", err.Error())
		return err
	}

	return nil
}

func (s *service) GetTokenOwner(ctx context.Context, refreshToken string) (*models.User, error) {
	log.C(ctx).Info("getting refresh token owner in refresh service")

	tx, err := s.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transact in refresh service, error %s", err.Error())
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	ownerEntity, err := s.repo.GetTokenOwner(ctx, refreshToken)
	if err != nil {
		log.C(ctx).Errorf("failed to get token owner, error %s in refresh service", err.Error())
		return nil, err
	}

	ownerModel := s.uConverter.ToModel(ownerEntity)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction in refresh service, error %s", err.Error())
		return nil, err
	}

	return ownerModel, nil
}
