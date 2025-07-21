package directives

import (
	"Todo-List/internProject/graphQL_service/internal/gql_middlewares"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/jwt"
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type jwtParser interface {
	ParseJWT(ctx context.Context, tokenString string) (*jwt.Claims, error)
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
			Message: err.Error(),
		}
	}

	if claims.Role != string(constants.Admin) {
		log.C(ctx).Debugf("only admins can see user's roles, actual role %s", claims.Role)
		return nil, nil
	}

	return next(ctx)
}
