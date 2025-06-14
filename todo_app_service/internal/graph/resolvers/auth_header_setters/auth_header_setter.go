package auth_header_setters

import (
	"context"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/gql_middlewares"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"net/http"
)

type requestAuthHeaderSetter struct{}

func NewRequestAuthHeader() *requestAuthHeaderSetter {
	return &requestAuthHeaderSetter{}
}

func (*requestAuthHeaderSetter) DecorateRequest(ctx context.Context, req *http.Request) (*http.Request, error) {
	log.C(ctx).Info("decorating http request's auth header")

	jwtAuthToken, ok := ctx.Value(gql_middlewares.AuthToken).(string)
	if !ok {
		log.C(ctx).Error("failed to decorate http request auth header, missing jwt token in context")
		return nil, fmt.Errorf("missing bearer token in http request")
	}

	req.Header.Set("Authorization", jwtAuthToken)
	return req, nil
}
