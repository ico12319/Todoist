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
	"errors"
	"io"
	"net/http"
)

//go:generate mockery --name=httpClient --exported --output=./mocks --outpkg=mocks --filename=http_client.go --with-expecter=true
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

//go:generate mockery --name=requestAuthSetter --exported --output=./mocks --outpkg=mocks --filename=request_auth_setter.go --with-expecter=true
type requestAuthSetter interface {
	DecorateRequest(context.Context, *http.Request) (*http.Request, error)
}

//go:generate mockery --name=accessConverter --exported --output=./mocks --outpkg=mocks --filename=access_converter.go --with-expecter=true
type accessConverter interface {
	ToGQL(*models.CallbackResponse) *gql.Access
	ToHandlerModelRefresh(*gql.RefreshTokenInput) *handler_models.Refresh
}

//go:generate mockery --name=jsonMarshaller --exported --output=./mocks --outpkg=mocks --filename=json_marshaller.go --with-expecter=true
type jsonMarshaller interface {
	Marshal(interface{}) ([]byte, error)
}

//go:generate mockery --name=httpRequester --exported --output=./mocks --outpkg=mocks --filename=http_requester2.go --with-expecter=true
type httpRequester interface {
	NewRequestWithContext(context.Context, string, string, io.Reader) (*http.Request, error)
}

type resolver struct {
	client         httpClient
	restUrl        string
	authSetter     requestAuthSetter
	converter      accessConverter
	jsonMarshaller jsonMarshaller
	httpRequester  httpRequester
}

func NewResolver(client httpClient, authSetter requestAuthSetter, converter accessConverter, jsonMarshaller jsonMarshaller, httpRequester httpRequester, restUrl string) *resolver {
	return &resolver{
		client:         client,
		authSetter:     authSetter,
		converter:      converter,
		jsonMarshaller: jsonMarshaller,
		httpRequester:  httpRequester,
		restUrl:        restUrl,
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

	req, err := r.httpRequester.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonRefreshPayload))
	if err != nil {
		log.C(ctx).Errorf("failed to make http request, error %s", err.Error())
		return &gql.Access{}, errors.New("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request, error %s", err.Error())
		return &gql.Access{}, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to receive http response, error %s", err.Error())
		return &gql.Access{}, errors.New("error when trying to get http response")
	}
	defer resp.Body.Close()

	callbackResponse, err := utils.HandleHttpCode[*models.CallbackResponse](resp)
	if err != nil {
		log.C(ctx).Errorf("failed to handler http code prperly, error %s", err.Error())
		return &gql.Access{}, err
	}

	return r.converter.ToGQL(callbackResponse), nil
}
