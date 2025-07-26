package refresh

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
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
	jwtToken                            = "valid jwt token"
	refreshToken                        = "valid refresh token"
)

var (
	errInvalidRequestBody   = errors.New(constants.INVALID_REQUEST_BODY)
	errWhenCallingJwtIssuer = errors.New("error when calling jwt issuer")
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

func getInjectedHandlerModelRefreshInRequestBody() *handler_models.Refresh {
	return &handler_models.Refresh{
		RefreshToken: refreshToken,
	}
}

func getExpectedCallbackResponse() *models.CallbackResponse {
	return &models.CallbackResponse{
		JwtToken:     jwtToken,
		RefreshToken: refreshToken,
	}
}

func extractErrorFromResponseRecorder(tb testing.TB, rr *httptest.ResponseRecorder, err error) {
	tb.Helper()
	var got map[string]string
	require.NoError(tb, json.Unmarshal(rr.Body.Bytes(), &got))
	expect := map[string]string{
		"error": err.Error(),
	}
	require.Equal(tb, expect, got)
}

func extractCallbackResponseFromHttpRecorder(t *testing.T, rr *httptest.ResponseRecorder) *models.CallbackResponse {
	t.Helper()
	var callbackResponse models.CallbackResponse
	err := json.NewDecoder(rr.Body).Decode(&callbackResponse)
	require.NoError(t, err)

	return &callbackResponse
}

func getValidHttpRequestWithInjectedHandlerModel(t *testing.T) *http.Request {
	injectedHandlerModelRefreshBytes, err := json.Marshal(getInjectedHandlerModelRefreshInRequestBody())
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/tokens/refresh", bytes.NewReader(injectedHandlerModelRefreshBytes))
	return req
}

func getEmptyCallbackResponse() *models.CallbackResponse {
	return &models.CallbackResponse{
		JwtToken:     "",
		RefreshToken: "",
	}
}
