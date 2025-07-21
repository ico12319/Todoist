package refresh

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const (
	VALID_USER_EMAIL      = "valid email"
	INVALID_USER_EMAIL    = "invalid email"
	VALID_REFRESH_TOKEN   = "random valid refresh token"
	INVALID_REFRESH_TOKEN = "random invalid refresh token"
	CONST_USER_EMAIL      = "random valid user email"
	SIGNED_STRING         = "valid signed jwt string"
)

var (
	sqlQueryCreateRefreshToken = "INSERT INTO user_refresh_tokens(refresh_token, user_id) VALUES (?, ?) ON CONFLICT (user_id) DO UPDATE SET refresh_token = EXCLUDED.refresh_token"
)

var (
	validUserId                         = uuid.FromBytesOrNil([]byte("valid user id"))
	invalidUserId                       = uuid.FromBytesOrNil([]byte("invalid user id"))
	errByUserService                    = application_errors.NewNotFoundError(constants.USER_TARGET, INVALID_USER_EMAIL)
	errWhenCreatingRefreshToken         = errors.New("error when trying to create refresh token")
	errByRefreshRepo                    = application_errors.NewNotFoundError(constants.USER_TARGET, invalidUserId.String())
	errByRefreshRepoInvalidRefreshToken = fmt.Errorf("invalid refresh token %s", INVALID_REFRESH_TOKEN)
	errInvalidJwtKeyType                = errors.New("invalid key type passed when trying to sign JWT")
	mockIssuedTime                      = time.Now()
	mockExpiryTime                      = time.Now().Add(150 * time.Hour)
	jwtKey                              = []byte("valid jwt key")
	invalidJwt                          = []byte("invalid jwt")
)

func initUser(userId string, email string) *models.User {
	return &models.User{
		Id:    userId,
		Email: email,
		Role:  constants.Admin,
	}
}

func initRefresh(refreshToken string, userId string) *models.Refresh {
	return &models.Refresh{
		RefreshToken: refreshToken,
		UserId:       userId,
	}
}

func initRefreshEntity(refreshToken string, userId uuid.UUID) *entities.Refresh {
	return &entities.Refresh{
		RefreshToken: refreshToken,
		UserId:       userId,
	}
}

func initRefreshEntityFromModel(refresh *models.Refresh) *entities.Refresh {
	return &entities.Refresh{
		RefreshToken: refresh.RefreshToken,
		UserId:       uuid.FromStringOrNil(refresh.UserId),
	}
}

func initRefreshModelFromEntity(refresh *entities.Refresh) *models.Refresh {
	return &models.Refresh{
		RefreshToken: refresh.RefreshToken,
		UserId:       refresh.UserId.String(),
	}
}

func initUserEntity(userId uuid.UUID, userEmail string) *entities.User {
	return &entities.User{
		Id:    userId,
		Email: userEmail,
	}
}

func initClaims() *Claims {
	return &Claims{
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(mockIssuedTime),
			ExpiresAt: jwt.NewNumericDate(mockExpiryTime),
		},
	}
}

func initJwt(claims *Claims) *jwt.Token {
	return &jwt.Token{
		Method: jwt.SigningMethodHS256,
		Claims: claims,
	}
}
