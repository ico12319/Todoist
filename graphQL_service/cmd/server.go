package main

import (
	graph2 "Todo-List/internProject/graphQL_service/graph"
	"Todo-List/internProject/graphQL_service/internal/auth_header_setters"
	"Todo-List/internProject/graphQL_service/internal/directives"
	"Todo-List/internProject/graphQL_service/internal/gql_converters"
	gql_middlewares2 "Todo-List/internProject/graphQL_service/internal/gql_middlewares"
	"Todo-List/internProject/graphQL_service/internal/marshallers"
	"Todo-List/internProject/graphQL_service/internal/requesters"
	"Todo-List/internProject/graphQL_service/internal/resolvers/access"
	"Todo-List/internProject/graphQL_service/internal/resolvers/list"
	"Todo-List/internProject/graphQL_service/internal/resolvers/todo"
	"Todo-List/internProject/graphQL_service/internal/resolvers/user"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	_ "Todo-List/internProject/graphQL_service/internal/url_decorators/url_decorators_creators"
	"Todo-List/internProject/todo_app_service/pkg/tokens"
	"context"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"Todo-List/internProject/todo_app_service/pkg/configuration"
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
	accessConv := gql_converters.NewAccessConverter()

	urlDecoratorFactory := url_decorators.GetUrlDecoratorFactoryInstance()
	requestDecorator := auth_header_setters.NewRequestAuthHeader()

	httpRequester := requesters.NewHttpRequester()
	jsonMarshaller := marshallers.NewJsonMarshaller()

	listResolver := list.NewResolver(httpClient, listConv, userConv, todoConv, restUrl, urlDecoratorFactory, requestDecorator, jsonMarshaller, httpRequester)
	todoResolver := todo.NewResolver(httpClient, urlDecoratorFactory, todoConv, userConv, listConv, restUrl, requestDecorator, jsonMarshaller, httpRequester)
	userResolver := user.NewResolver(httpClient, userConv, listConv, todoConv, restUrl, urlDecoratorFactory, requestDecorator, httpRequester)
	accessResolver := access.NewResolver(httpClient, requestDecorator, accessConv, jsonMarshaller, httpRequester, restUrl)

	jwtParserHelper := tokens.NewJwtParser()
	jwtParser := tokens.NewJwtParseService(jwtParserHelper)
	
	roleDirective := directives.NewRoleDirectiveImplementation(jwtParser)

	root := graph2.NewResolver(listResolver, todoResolver, userResolver, accessResolver)
	srv := handler.New(graph2.NewExecutableSchema(graph2.Config{
		Resolvers: root,
		Directives: graph2.DirectiveRoot{
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
	r.Use(gql_middlewares2.NewJwtPopulateMiddleware(), gql_middlewares2.ContentTypeMiddlewareFunc())
	log.C(context.Background()).Fatal(http.ListenAndServe(":"+port, r))
}
