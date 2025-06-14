package main

import (
	"context"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/directives"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/gql_converters"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/gql_middlewares"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/resolvers/auth_header_setters"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/resolvers/list"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/resolvers/todo"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/resolvers/user"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/url_decorators"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/middlewares"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/oauth/tokens"

	//"github.com/I763039/Todo-List/internProject/todo_app_service/internal/middlewares"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/gorilla/mux"
	"github.com/vektah/gqlparser/v2/ast"
	"net/http"
)

func main() {
	port := log.GetInstance().GraphConfig.Port
	restUrl := log.GetInstance().RestConfig.TodoApiUrl

	httpClient := &http.Client{}

	sConverter := gql_converters.NewStatusConverter()
	pConverter := gql_converters.NewPriorityConverter()
	roleConverter := gql_converters.NewRoleConverter()
	listConv := gql_converters.NewListConverter()
	userConv := gql_converters.NewUserConverter(roleConverter)
	todoConv := gql_converters.NewTodoConverter(pConverter, sConverter)

	commonFactory := url_decorators.NewCommonDecoratorFactory()
	rFactory := url_decorators.NewQueryParamsRetrieverFactory(sConverter, pConverter, commonFactory)
	requestDecorator := auth_header_setters.NewRequestAuthHeader()
	listResolver := list.NewResolver(httpClient, listConv, userConv, todoConv, restUrl, commonFactory, rFactory, requestDecorator)
	todoResolver := todo.NewResolver(httpClient, rFactory, todoConv, userConv, listConv, restUrl, requestDecorator)
	userResolver := user.NewResolver(httpClient, userConv, listConv, todoConv, restUrl, commonFactory, requestDecorator)

	jwtParser := tokens.NewJwtParseService()
	roleDirective := directives.NewRoleDirectiveImplementation(jwtParser)

	root := graph.NewResolver(listResolver, todoResolver, userResolver)
	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: root,
		Directives: graph.DirectiveRoot{
			HasRole: roleDirective.HasRole,
		},
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	r := mux.NewRouter()
	r.Handle("/query", srv).Methods(http.MethodGet, http.MethodPost)
	r.Use(gql_middlewares.NewJwtPopulateMiddleware(), middlewares.ContentTypeMiddlewareFunc)
	log.C(context.Background()).Fatal(http.ListenAndServe(":"+port, r))
}
