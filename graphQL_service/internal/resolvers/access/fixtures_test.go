package access

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"net/http"
	"net/http/httptest"
	"os"
)

const (
	AUTH              = "FJAAHFFAHFALAVAH"
	OLD_REFRESH_TOKEN = "random old refresh token"
	JWT               = "random jwt token"
	NEW_REFRESH_TOKEN = "random new refresh token"
)

var (
	restUrl                                            = os.Getenv("TODO_REST_API_URL")
	URL                                                = restUrl + gql_constants.TOKENS_PATH + gql_constants.REFRESH_PATH
	jsonMarshalError                                   = errors.New("error when trying to marshal JSON refresh payload")
	requestError                                       = errors.New("error when making http request")
	errorByRequestDecorator                            = errors.New("missing bearer token in http request")
	httpClientError                                    = errors.New("error when trying to get http response")
	errorWhenHandlingHttpStatusCodeInternalServerError = &gqlerror.Error{
		Message:    "Internal error, please try again later.",
		Extensions: map[string]interface{}{"code": "INTERNAL_SERVER_ERROR"},
	}
)

func initRefreshInput() *gql.RefreshTokenInput {
	return &gql.RefreshTokenInput{
		RefreshToken: OLD_REFRESH_TOKEN,
	}
}

func initHandlerModel() *handler_models.Refresh {
	return &handler_models.Refresh{
		RefreshToken: OLD_REFRESH_TOKEN,
	}
}

func initModel() *models.CallbackResponse {
	return &models.CallbackResponse{
		JwtToken:     JWT,
		RefreshToken: NEW_REFRESH_TOKEN,
	}
}

func initGqlModel() *gql.Access {
	return &gql.Access{
		JwtToken:     JWT,
		RefreshToken: NEW_REFRESH_TOKEN,
	}
}

func getBytesOfEntity[T any](refresh T) []byte {
	bodyBytes, _ := json.Marshal(refresh)
	return bodyBytes
}

func initMockRequest(refresh *gql.RefreshTokenInput) *http.Request {
	return httptest.NewRequest(http.MethodPost, URL, bytes.NewReader(getBytesOfEntity(refresh)))
}
