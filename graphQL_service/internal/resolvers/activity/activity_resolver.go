package activity

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/graph/utils"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type activityConverter interface {
	ToGQL(*models.RandomActivity) *gql.RandomActivity
}

type httpResponseGetter interface {
	GetHttpResponse(context.Context, string, string, io.Reader) (*http.Response, error)
}
type resolver struct {
	restUrl        string
	responseGetter httpResponseGetter
	converter      activityConverter
}

func NewResolver(restUrl string, responseGetter httpResponseGetter, converter activityConverter) *resolver {
	return &resolver{restUrl: restUrl, responseGetter: responseGetter, converter: converter}
}

func (r *resolver) RandomActivity(ctx context.Context) (*gql.RandomActivity, error) {
	log.C(ctx).Info("getting a random activity in activity resolver")

	url := r.restUrl + gql_constants.ACTIVITIES_PATH + gql_constants.RANDOM_PATH

	resp, err := r.responseGetter.GetHttpResponse(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in activity resolver, error %s", err.Error())
		return &gql.RandomActivity{}, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get collaborators in a list in list resolver, error %s due to bad response status code", err.Error())
		return &gql.RandomActivity{}, err
	}

	var activity models.RandomActivity
	if err = json.NewDecoder(resp.Body).Decode(&activity); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return &gql.RandomActivity{}, err
	}

	return r.converter.ToGQL(&activity), nil
}
