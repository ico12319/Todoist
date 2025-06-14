package tokens

import (
	"context"
	"errors"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/application_errors"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/handler_models"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type userService interface {
	GetUserRecordByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUserRecord(ctx context.Context, user *handler_models.CreateUser) (*models.User, error)
}

type timeGenerator interface {
	Now() time.Time
}
type jwtCreationService struct {
	uService userService
	timeGen  timeGenerator
}

func NewJwtService(uService userService, timeGen timeGenerator) *jwtCreationService {
	return &jwtCreationService{uService: uService, timeGen: timeGen}
}

func (j *jwtCreationService) GenerateJWT(ctx context.Context, email string, role string) (string, error) {
	configManager := log.GetInstance()
	log.C(ctx).Info("generating jwt token in jwt service")
	expirationTime := j.timeGen.Now().Add(30 * time.Minute)

	user, err := j.uService.GetUserRecordByEmail(ctx, email)
	if err != nil {
		var applicationError *application_errors.NotFoundError
		if errors.As(err, &applicationError) {
			user, err = j.uService.CreateUserRecord(ctx, &handler_models.CreateUser{
				Email: email,
				Role:  role,
			})

			if err != nil {
				log.C(ctx).Errorf("failed to generate JWT, error %s when trying to create unregistred user", err.Error())
				return "", err
			}
		}
		return "", err
	}

	claims := &Claims{
		UserId: user.Id,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(j.timeGen.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	stringToken, err := jwtToken.SignedString(configManager.JwtConfig.Secret)
	if err != nil {
		log.C(ctx).Errorf("failed to sign jwt, error %s", err.Error())
		return "", err
	}

	return stringToken, nil
}
