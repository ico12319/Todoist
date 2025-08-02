package jwt

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type userService interface {
	GetUserRecordByEmail(context.Context, string) (*models.User, error)
	CreateUserRecord(context.Context, *handler_models.CreateUser) (*models.User, error)
	UpdateUserRecordPartially(context.Context, string, *handler_models.UpdateUser) (*models.User, error)
}

type timeGenerator interface {
	Now() time.Time
}

//go:generate mockery --name=jwtGetter --exported --output=./mocks --outpkg=mocks --filename=jwt_getter.go --with-expecter=true
type jwtGetter interface {
	GetJWTWithClaims(method jwt.SigningMethod, claims jwt.Claims) *jwt.Token
	GetSignedJWT(jwt *jwt.Token, key interface{}) (string, error)
}

type jwtCreationService struct {
	uService  userService
	timeGen   timeGenerator
	getter    jwtGetter
	jwtSecret []byte
}

func NewJwtService(uService userService, timeGen timeGenerator, getter jwtGetter, jwtSecret []byte) *jwtCreationService {
	return &jwtCreationService{
		uService:  uService,
		timeGen:   timeGen,
		getter:    getter,
		jwtSecret: jwtSecret,
	}
}

func (j *jwtCreationService) GenerateJWT(ctx context.Context, email string, role string) (string, error) {
	log.C(ctx).Info("generating jwt token in jwt service")

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
		} else {
			return "", err
		}
	}

	if string(user.Role) != role {
		if user, err = j.uService.UpdateUserRecordPartially(ctx, user.Id, &handler_models.UpdateUser{
			Role: &role,
		}); err != nil {
			log.C(ctx).Errorf("failed to update user role, error %s", err.Error())
			return "", err
		}
	}

	expirationTime := determineTokenExpirationTimeBasedOnUserRole(j.timeGen, user)

	claims := &Claims{
		UserId: user.Id,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(j.timeGen.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	jwtToken := j.getter.GetJWTWithClaims(jwt.SigningMethodHS256, claims)

	stringToken, err := j.getter.GetSignedJWT(jwtToken, j.jwtSecret)
	if err != nil {
		log.C(ctx).Errorf("failed to sign jwt, error %s", err.Error())
		return "", err
	}

	return stringToken, nil
}

func (j *jwtCreationService) GenerateRefreshToken(ctx context.Context) (string, error) {
	log.C(ctx).Info("generating refresh token")

	expirationTime := j.timeGen.Now().Add(150 * time.Hour)

	claims := &RefreshClaims{
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(j.timeGen.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	refreshJwtToken := j.getter.GetJWTWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := j.getter.GetSignedJWT(refreshJwtToken, j.jwtSecret)
	if err != nil {
		log.C(ctx).Errorf("failed to sign jwt, error %s", err.Error())
		return "", utils.DetermineErrorWhenSigningJWT(err)
	}

	return signedToken, nil
}

func determineTokenExpirationTimeBasedOnUserRole(generator timeGenerator, user *models.User) time.Time {
	var expirationTime time.Time

	if user.Role == constants.Admin {
		expirationTime = generator.Now().Add(24 * time.Hour)
	} else {
		expirationTime = generator.Now().Add(30 * time.Minute)
	}

	return expirationTime
}
