package main

import (
	"Todo-List/internProject/todo_app_service/internal/converters"
	"Todo-List/internProject/todo_app_service/internal/generators"
	"Todo-List/internProject/todo_app_service/internal/lists"
	"Todo-List/internProject/todo_app_service/internal/oauth"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	_ "Todo-List/internProject/todo_app_service/internal/sql_query_decorators/sql_decorators_creators"
	"Todo-List/internProject/todo_app_service/internal/status_code_encoders"
	"Todo-List/internProject/todo_app_service/internal/todos"
	"Todo-List/internProject/todo_app_service/internal/users"
	"Todo-List/internProject/todo_app_service/internal/validators"
	"Todo-List/internProject/todo_app_service/pkg/application"
	config "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/tokens"
	"Todo-List/internProject/todo_app_service/pkg/tokens/refresh"
	_ "github.com/lib/pq"
	"net/http"
)

func main() {
	configManagerInstance := config.GetInstance()

	client := &http.Client{}
	db := config.OpenPostgres(configManagerInstance.DbConfig)

	tRepo := todos.NewSQLTodoDB(db)
	lRepo := lists.NewSQLListDB(db)
	uRepo := users.NewSQLUserDB(db)

	uuidGen := generators.NewUuidGenerator()
	timeGen := generators.NewTimeGenerator()

	todoConverter := converters.NewTodoConverter()
	uDBConverter := converters.NewUserConverter()
	listConverter := converters.NewListConverter()

	decoratorFactory := sql_query_decorators.GetDecoratorFactoryInstance()

	uService := users.NewService(uRepo, uDBConverter, listConverter, todoConverter, uuidGen, decoratorFactory)
	lService := lists.NewService(lRepo, uuidGen, timeGen, listConverter, uService, uDBConverter, decoratorFactory)
	tService := todos.NewService(tRepo, lService, uuidGen, timeGen, todoConverter, uDBConverter, decoratorFactory)

	fValidator := validators.GetInstance()
	statusCodeFactory := status_code_encoders.NewStatusCodeEncoderFactory()
	tHandler := todos.NewHandler(tService, fValidator, statusCodeFactory)
	lHandler := lists.NewHandler(lService, fValidator, statusCodeFactory)
	uHandler := users.NewHandler(uService, fValidator, statusCodeFactory)
	gitHubService := oauth.NewService(client)

	jwtGetter := tokens.NewJwtGetter()

	userInfoService := oauth.NewUserInfoService(gitHubService)
	stateGenerator := generators.NewStateGenerator()
	jwtCreator := tokens.NewJwtService(uService, timeGen, jwtGetter)

	refreshBuilder := refresh.NewRefreshTokenBuilder(timeGen, jwtGetter)
	refreshRepo := refresh.NewSqlRefreshDB(db)
	refreshConverter := converters.NewRefreshConverter()
	refreshService := refresh.NewService(refreshRepo, uService, refreshConverter, uDBConverter)
	oauthService := oauth.NewOauthService(userInfoService, refreshService, stateGenerator, refreshBuilder, jwtCreator, configManagerInstance)
	oauthHandler := oauth.NewHandler(oauthService)

	jwtParserHelper := tokens.NewJwtParser()
	jwtParser := tokens.NewJwtParseService(jwtParserHelper)

	restServer := application.NewServer(lHandler, tHandler, uHandler, oauthHandler, lService, uService, tService, uuidGen, jwtParser, statusCodeFactory)
	restServer.Start(configManagerInstance.RestConfig.Port)
}
