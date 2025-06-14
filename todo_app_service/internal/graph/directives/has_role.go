package directives

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/gql_middlewares"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/oauth/tokens"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/constants"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type jwtParser interface {
	ParseJWT(ctx context.Context, tokenString string) (*tokens.Claims, error)
}
type roleDirective struct {
	parser jwtParser
}

func NewRoleDirectiveImplementation(parser jwtParser) *roleDirective {
	return &roleDirective{parser: parser}
}

func (i *roleDirective) HasRole(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	log.C(ctx).Info("checking whether user has right to see the role")

	jwtToken, ok := ctx.Value(gql_middlewares.AuthToken).(string)
	if !ok {
		log.C(ctx).Debugf("nil token in context...")
		return nil, &gqlerror.Error{
			Message: "missing jwt token",
		}
	}

	claims, err := i.parser.ParseJWT(ctx, jwtToken)
	if err != nil {
		log.C(ctx).Errorf("failed to check whether user has rights to see the role, error %s when trying to parse jwt", err.Error())
		return nil, &gqlerror.Error{
			Message: "unable to parse JWT",
		}
	}

	if claims.Role != string(constants.Admin) {
		log.C(ctx).Debugf("only admins can see user's roles, actual role %s", claims.Role)
		return nil, nil
	}

	return next(ctx)
}
