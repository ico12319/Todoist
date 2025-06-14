package main

import (
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/converters"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/generators"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/lists"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/oauth"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/oauth/tokens"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/oauth/tokens/refresh"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/status_code_encoders"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/todos"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/users"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/validators"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/application"
	config "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
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

	commonFactory := sql_query_decorators.NewCommonFactory()
	concreteDecoratorFactory := sql_query_decorators.NewConcreteQueryDecoratorFactory(commonFactory)

	uService := users.NewService(uRepo, uDBConverter, listConverter, todoConverter, uuidGen, concreteDecoratorFactory)
	lService := lists.NewService(lRepo, uuidGen, timeGen, listConverter, uService, uDBConverter, concreteDecoratorFactory)
	tService := todos.NewService(tRepo, lService, uuidGen, timeGen, todoConverter, uDBConverter, concreteDecoratorFactory)

	fValidator := validators.GetInstance()
	statusCodeFactory := status_code_encoders.NewStatusCodeEncoderFactory()
	tHandler := todos.NewHandler(tService, fValidator, statusCodeFactory)
	lHandler := lists.NewHandler(lService, fValidator, statusCodeFactory)
	uHandler := users.NewHandler(uService, fValidator, statusCodeFactory)
	gitHubService := oauth.NewService(client)

	userInfoService := oauth.NewUserInfoService(gitHubService)
	stateGenerator := generators.NewStateGenerator()
	jwtCreator := tokens.NewJwtService(uService, timeGen)

	refreshBuilder := refresh.NewRefreshTokenBuilder(timeGen)
	refreshRepo := refresh.NewSqlRefreshDB(db)
	refreshConverter := converters.NewRefreshConverter()
	refreshService := refresh.NewService(refreshRepo, uService, refreshConverter, uDBConverter, refreshBuilder)
	oauthService := oauth.NewOauthService(userInfoService, refreshService, stateGenerator, refreshBuilder, jwtCreator, configManagerInstance)
	oauthHandler := oauth.NewHandler(oauthService)

	jwtParser := tokens.NewJwtParseService()

	restServer := application.NewServer(lHandler, tHandler, uHandler, oauthHandler, lService, uService, tService, uuidGen, jwtParser, statusCodeFactory)
	restServer.Start(configManagerInstance.RestConfig.Port)
}
