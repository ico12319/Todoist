package main

import (
	"Todo-List/internProject/graphQL_service/graph"
	"Todo-List/internProject/graphQL_service/internal/directives"
	"Todo-List/internProject/graphQL_service/internal/gql_auth_header_setters"
	"Todo-List/internProject/graphQL_service/internal/gql_converters"
	"Todo-List/internProject/graphQL_service/internal/gql_middlewares"
	"Todo-List/internProject/graphQL_service/internal/health"
	"Todo-List/internProject/graphQL_service/internal/resolvers/access"
	"Todo-List/internProject/graphQL_service/internal/resolvers/activity"
	"Todo-List/internProject/graphQL_service/internal/resolvers/list"
	"Todo-List/internProject/graphQL_service/internal/resolvers/todo"
	"Todo-List/internProject/graphQL_service/internal/resolvers/user"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	_ "Todo-List/internProject/graphQL_service/internal/url_decorators/url_decorators_creators"
	"Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/http_helpers"
	"Todo-List/internProject/todo_app_service/pkg/jwt"
	"context"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/mux"
	"github.com/vektah/gqlparser/v2/ast"
	"net/http"
)

func main() {
	configManager := configuration.GetInstance()

	port := configManager.GraphConfig.Port
	restUrl := configManager.RestConfig.TodoApiUrl

	httpClient := &http.Client{}

	sConverter := gql_converters.NewStatusConverter()
	pConverter := gql_converters.NewPriorityConverter()
	roleConverter := gql_converters.NewRoleConverter()
	listConv := gql_converters.NewListConverter()
	userConv := gql_converters.NewUserConverter(roleConverter)
	todoConv := gql_converters.NewTodoConverter(pConverter, sConverter)
	accessConv := gql_converters.NewAccessConverter()
	activityConverter := gql_converters.NewActivityConverter()

	urlDecoratorFactory := url_decorators.GetUrlDecoratorFactoryInstance()
	requestDecorator := gql_auth_header_setters.NewRequestAuthHeader()

	httpRequester := http_helpers.NewHttpRequester()
	httpService := http_helpers.NewService(httpClient, requestDecorator, httpRequester)
	jsonMarshaller := http_helpers.NewJsonMarshaller()

	listResolver := list.NewResolver(listConv, userConv, todoConv, restUrl, urlDecoratorFactory, httpService, jsonMarshaller)
	todoResolver := todo.NewResolver(urlDecoratorFactory, todoConv, userConv, listConv, restUrl, jsonMarshaller, httpService)
	userResolver := user.NewResolver(userConv, listConv, todoConv, restUrl, urlDecoratorFactory, httpService)
	accessResolver := access.NewResolver(accessConv, jsonMarshaller, httpService, restUrl)
	activityResolver := activity.NewResolver(restUrl, httpService, activityConverter)
	jwtParserHelper := jwt.NewJwtManager()
	jwtParser := jwt.NewJwtParseService(jwtParserHelper)

	roleDirective := directives.NewRoleDirectiveImplementation(jwtParser)

	root := graph.NewResolver(listResolver, todoResolver, userResolver, accessResolver, activityResolver)
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

	restHealthService := health.NewService(httpService)
	restHealthHandler := health.NewHandler(restHealthService, restUrl)

	r := mux.NewRouter()
	r.HandleFunc("/api/healthz", restHealthHandler.HandleCheckingRESTHealthz).Methods(http.MethodGet)
	r.HandleFunc("/api/readyz", restHealthHandler.HandleCheckingRestReadyz).Methods(http.MethodGet)

	mainGql := r.PathPrefix("/query").Subrouter()
	mainGql.Handle("", srv).Methods(http.MethodPost, http.MethodOptions, http.MethodGet, http.MethodOptions)
	mainGql.Use(gql_middlewares.ContentTypeMiddlewareFunc(),
		gql_middlewares.NewJwtPopulateMiddleware(),
		gql_middlewares.CorsMiddlewareFunc(configManager.CorsConfig.FrontendUrl))

	configuration.C(context.Background()).Fatal(http.ListenAndServe(":"+port, r))
}
