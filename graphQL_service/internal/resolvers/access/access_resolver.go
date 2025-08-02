package access

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/graph/utils"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

//go:generate mockery --name=accessConverter --exported --output=./mocks --outpkg=mocks --filename=access_converter.go --with-expecter=true
type accessConverter interface {
	ToGQL(*models.CallbackResponse) *gql.Access
	ToHandlerModelRefresh(*gql.RefreshTokenInput) *handler_models.Refresh
}

//go:generate mockery --name=jsonMarshaller --exported --output=./mocks --outpkg=mocks --filename=json_marshaller.go --with-expecter=true
type jsonMarshaller interface {
	Marshal(interface{}) ([]byte, error)
}

//go:generate mockery --name=httpResponseGetter --exported --output=./mocks --outpkg=mocks --filename=http_response_getter.go --with-expecter=true
type httpService interface {
	GetHttpResponse(context.Context, string, string, io.Reader) (*http.Response, error)
}
type resolver struct {
	restUrl        string
	converter      accessConverter
	jsonMarshaller jsonMarshaller
	httpService    httpService
}

func NewResolver(converter accessConverter, jsonMarshaller jsonMarshaller, httpService httpService, restUrl string) *resolver {
	return &resolver{
		restUrl:        restUrl,
		converter:      converter,
		jsonMarshaller: jsonMarshaller,
		httpService:    httpService,
	}
}

func (r *resolver) ExchangeRefreshToken(ctx context.Context, input gql.RefreshTokenInput) (*gql.Access, error) {
	log.C(ctx).Info("exchanging refresh token for new jwt and refresh token in access resolver")

	url := r.restUrl + gql_constants.TOKENS_PATH + gql_constants.REFRESH_PATH

	jsonRefreshPayload, err := r.jsonMarshaller.Marshal(r.converter.ToHandlerModelRefresh(&input))
	if err != nil {
		log.C(ctx).Errorf("failed to JSON marshal refresh payload, error %s", err.Error())
		return &gql.Access{}, errors.New("error when trying to marshal JSON refresh payload")
	}

	resp, err := r.httpService.GetHttpResponse(ctx, http.MethodPost, url, bytes.NewReader(jsonRefreshPayload))
	if err != nil {
		log.C(ctx).Errorf("failed to exchnage refresh token, error %s when trying to get http response", err.Error())
		return &gql.Access{}, errors.New("error when trying to get http response")
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get collaborators in a list in list resolver, error %s due to bad response status code", err.Error())
		return &gql.Access{}, err
	}

	var refreshResponse models.CallbackResponse
	if err = json.NewDecoder(resp.Body).Decode(&refreshResponse); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return &gql.Access{}, err
	}

	return r.converter.ToGQL(&refreshResponse), nil
}
