package application

import (
	"Todo-List/internProject/todo_app_service/internal/middlewares"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	"Todo-List/internProject/todo_app_service/internal/status_code_encoders"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"Todo-List/internProject/todo_app_service/pkg/tokens"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type uuidGenerator interface {
	Generate() string
}

type listHandler interface {
	HandleGetListRecord(w http.ResponseWriter, r *http.Request)
	HandleGetLists(w http.ResponseWriter, r *http.Request)
	HandleGetCollaborators(w http.ResponseWriter, r *http.Request)
	HandleGetListOwner(w http.ResponseWriter, r *http.Request)
	HandleDeleteList(w http.ResponseWriter, r *http.Request)
	HandleCreateList(w http.ResponseWriter, r *http.Request)
	HandleUpdateListPartially(w http.ResponseWriter, r *http.Request)
	HandleAddCollaborator(w http.ResponseWriter, r *http.Request)
	HandleDeleteCollaborator(w http.ResponseWriter, r *http.Request)
	HandleDeleteLists(w http.ResponseWriter, r *http.Request)
}

type todoHandler interface {
	HandleTodoCreation(w http.ResponseWriter, r *http.Request)
	HandleGetTodos(w http.ResponseWriter, r *http.Request)
	HandleGetTodo(w http.ResponseWriter, r *http.Request)
	HandleGetTodosByListId(w http.ResponseWriter, r *http.Request)
	HandleDeleteTodo(w http.ResponseWriter, r *http.Request)
	HandleDeleteTodos(w http.ResponseWriter, r *http.Request)
	HandleDeleteTodosByListId(w http.ResponseWriter, r *http.Request)
	HandleUpdateTodoRecord(w http.ResponseWriter, r *http.Request)
	HandleGetTodoByListId(w http.ResponseWriter, r *http.Request)
	HandleGetTodoAssignee(w http.ResponseWriter, r *http.Request)
	HandleDeleteTodoByListId(w http.ResponseWriter, r *http.Request)
}

type userHandler interface {
	HandleGetUser(w http.ResponseWriter, r *http.Request)
	HandleGetUsers(w http.ResponseWriter, r *http.Request)
	HandleDeleteUser(w http.ResponseWriter, r *http.Request)
	HandleGetUserLists(w http.ResponseWriter, r *http.Request)
	HandleGetTodosAssignedToUser(w http.ResponseWriter, r *http.Request)
	HandleDeleteUsers(w http.ResponseWriter, r *http.Request)
}

type listService interface {
	GetListRecord(ctx context.Context, listId string) (*models.List, error)
	GetCollaborators(ctx context.Context, lFilters *filters.ListFilters) ([]*models.User, error)
	GetListOwnerRecord(ctx context.Context, listId string) (*models.User, error)
}

type userService interface {
	GetUserRecordByEmail(ctx context.Context, email string) (*models.User, error)
}

type todoService interface {
	GetTodoAssigneeToRecord(ctx context.Context, todoId string) (*models.User, error)
	GetTodoRecord(ctx context.Context, todoId string) (*models.Todo, error)
}

type oauthHandler interface {
	HandleLogin(w http.ResponseWriter, r *http.Request)
	HandleCallback(w http.ResponseWriter, r *http.Request)
	HandleRefresh(w http.ResponseWriter, r *http.Request)
}

type jwtParser interface {
	ParseJWT(ctx context.Context, tokenString string) (*tokens.Claims, error)
}

type statusCodeEncoderFactory interface {
	CreateStatusCodeEncoder(ctx context.Context, w http.ResponseWriter, err error) status_code_encoders.StatusCodeEncoder
}

type server struct {
	listHandler  listHandler
	todoHandler  todoHandler
	userHandler  userHandler
	oauthHandler oauthHandler
	listServ     listService
	userServ     userService
	tServ        todoService
	generator    uuidGenerator
	tokenParser  jwtParser
	factory      statusCodeEncoderFactory
}

func NewServer(listHandler listHandler, todoHandler todoHandler, userHandler userHandler, oauthHandler oauthHandler, listServ listService,
	userServ userService, tServ todoService, generator uuidGenerator, tokenParser jwtParser, factory statusCodeEncoderFactory) *server {
	return &server{listHandler: listHandler, todoHandler: todoHandler, userHandler: userHandler, oauthHandler: oauthHandler, listServ: listServ, userServ: userServ, tServ: tServ, generator: generator, tokenParser: tokenParser, factory: factory}
}

// only admins, list owners and collaborators can modify lists
func (s *server) registerListIdAuthRoutes(router *mux.Router) {
	router.HandleFunc("/collaborators", s.listHandler.HandleAddCollaborator).Methods(http.MethodPost)
	router.HandleFunc("", s.listHandler.HandleUpdateListPartially).Methods(http.MethodPatch)
}

// only admins and list owners can modify delete a certain collaborator
func (s *server) registerListUserIdRoutes(router *mux.Router) {
	router.HandleFunc("", s.listHandler.HandleDeleteCollaborator).Methods(http.MethodDelete)
}

// only admins and list owners can modify delete a certain collaborator
func (s *server) registerListDeleteRoutes(router *mux.Router) {
	router.HandleFunc("", s.listHandler.HandleDeleteList).Methods(http.MethodDelete)
}

// only admins, list the list owner and the list collaborators of the list where todo is located can update and delete todo
func (s *server) registerTodoIdAuthRoutes(router *mux.Router) {
	router.HandleFunc("", s.todoHandler.HandleDeleteTodo).Methods(http.MethodDelete)
	router.HandleFunc("", s.todoHandler.HandleUpdateTodoRecord).Methods(http.MethodPatch)
}

// all authorized users can read user specific things
func (s *server) registerReadUserIdRoutes(router *mux.Router) {
	router.HandleFunc("", s.userHandler.HandleGetUser).Methods(http.MethodGet)
	router.HandleFunc("/lists", s.userHandler.HandleGetUserLists).Methods(http.MethodGet)
	router.HandleFunc("/todos", s.userHandler.HandleGetTodosAssignedToUser).Methods(http.MethodGet)
}

// only admins can modify users
func (s *server) registerAuthUserIdRoutes(router *mux.Router) {
	router.HandleFunc("", s.userHandler.HandleDeleteUser).Methods(http.MethodDelete)
}

// every one can access these endpoints, entry point of the API
func (s *server) registerOauthPaths(router *mux.Router) {
	// this will redirect you to github and will prompt you to type your credentials and you will be asked to grant my API
	// with the scopes my API need so it can access your github information, this will be used to determine your role in the API
	// and you will be issued with a JWT token that you can use in order to authorize in front of my API.
	router.HandleFunc("/github/login", s.oauthHandler.HandleLogin).Methods(http.MethodGet)
	// this will redirect you to a page where you will be granted with a JWT token and a Refresh token,
	// then you should put the JWT in the Auth header in Postman and you will be able to call the API,
	// keep you refresh token because you JWT token will expire after around 3 minutes.
	router.HandleFunc("/auth2/callback", s.oauthHandler.HandleCallback).Methods(http.MethodGet)
	// this is the url where you will be requesting your new JWT token when yours expires,
	// you should put you refresh token in the request body and if it is correct you will be issued with
	// new JWT and new Refresh token
	router.HandleFunc("/tokens/refresh", s.oauthHandler.HandleRefresh).Methods(http.MethodPost)
}

// only admins and writers can create todos and lists
func (s *server) registerPostPaths(router *mux.Router) {
	router.HandleFunc("/lists", s.listHandler.HandleCreateList).Methods(http.MethodPost)
	router.HandleFunc("/todos", s.todoHandler.HandleTodoCreation).Methods(http.MethodPost)
}

// all authorized users can read lists, todos and users
func (s *server) registerReadAllRolesPaths(router *mux.Router) {
	router.HandleFunc("/lists", s.listHandler.HandleGetLists).Methods(http.MethodGet)
	router.HandleFunc("/todos", s.todoHandler.HandleGetTodos).Methods(http.MethodGet)
	router.HandleFunc("/users", s.userHandler.HandleGetUsers).Methods(http.MethodGet)
}

// all authorized users can read todo specific things
func (s *server) registerReadTodoIdPaths(router *mux.Router) {
	router.HandleFunc("", s.todoHandler.HandleGetTodo).Methods(http.MethodGet)
	router.HandleFunc("/assignee", s.todoHandler.HandleGetTodoAssignee).Methods(http.MethodGet)
}

// all authorized users can read todos related to a certain list
func (s *server) registerReadTodoListPaths(router *mux.Router) {
	router.HandleFunc("", s.todoHandler.HandleGetTodosByListId).Methods(http.MethodGet)
}

// only the users who can modify the list where the todo belongs can delete all todos in it
func (s *server) registerAuthTodoListPaths(router *mux.Router) {
	router.HandleFunc("", s.todoHandler.HandleDeleteTodosByListId).Methods(http.MethodDelete)
}

// all users that are authorized with their jwt can read list specific things
func (s *server) registerReadListIdPaths(router *mux.Router) {
	router.HandleFunc("/collaborators", s.listHandler.HandleGetCollaborators).Methods(http.MethodGet)
	router.HandleFunc("", s.listHandler.HandleGetListRecord).Methods(http.MethodGet)
	router.HandleFunc("/owner", s.listHandler.HandleGetListOwner).Methods(http.MethodGet)
	router.HandleFunc("", s.listHandler.HandleGetCollaborators).Methods(http.MethodGet)
}

// only admins can delete all entities from a certain type!
func (s *server) registerAdminPaths(router *mux.Router) {
	router.HandleFunc("/lists", s.listHandler.HandleDeleteLists).Methods(http.MethodDelete)
	router.HandleFunc("/todos", s.todoHandler.HandleDeleteTodos).Methods(http.MethodDelete)
	router.HandleFunc("/users", s.userHandler.HandleDeleteUsers).Methods(http.MethodDelete)
}

func (s *server) Start(restServerPort string) {
	oauthRouter := mux.NewRouter()
	oauthRouter.Use(middlewares.ContentTypeMiddlewareFunc)
	s.registerOauthPaths(oauthRouter)

	router := oauthRouter.PathPrefix("").Subrouter()
	router.Use(middlewares.ContentTypeMiddlewareFunc, middlewares.LogEnrichMiddlewareFunc(s.generator),
		middlewares.AuthorisationMiddlewareFunc(s.userServ, s.tokenParser))

	postRouter := router.PathPrefix("").Subrouter()
	postRouter.Use(middlewares.ObjectCreationMiddlewareFunc)
	s.registerPostPaths(postRouter)

	globalReaderRouter := router.PathPrefix("").Subrouter()
	s.registerReadAllRolesPaths(globalReaderRouter)

	adminRouter := router.PathPrefix("").Subrouter()
	adminRouter.Use(middlewares.GlobalAccessMiddlewareFunc)
	s.registerAdminPaths(adminRouter)

	listRouter := router.PathPrefix("/lists").Subrouter()
	listIdAuthRouter := listRouter.PathPrefix(fmt.Sprintf("/{list_id:%s}", constants.UUID_REGEX)).Subrouter()
	listIdAuthRouter.Use(middlewares.ExtractionListIdMiddlewareFunc, middlewares.ListAccessPermissionMiddlewareFunc(s.listServ, s.factory))
	s.registerListIdAuthRoutes(listIdAuthRouter)

	listIdReadRouter := listRouter.PathPrefix(fmt.Sprintf("/{list_id:%s}", constants.UUID_REGEX)).Subrouter()
	listIdReadRouter.Use(middlewares.ExtractionListIdMiddlewareFunc)
	s.registerReadListIdPaths(listIdReadRouter)

	listDeletionRouter := listIdAuthRouter.Methods(http.MethodDelete).Subrouter()
	listDeletionRouter.Use(middlewares.ListDeletionMiddlewareFunc)
	s.registerListDeleteRoutes(listDeletionRouter)

	listUserIdRouter := listDeletionRouter.PathPrefix(fmt.Sprintf("/collaborators/{user_id:%s}", constants.UUID_REGEX)).Subrouter()
	listUserIdRouter.Use(middlewares.ExtractionUserIdMiddlewareFunc)
	s.registerListUserIdRoutes(listUserIdRouter)

	todoRouter := router.PathPrefix("/todos").Subrouter()
	todoIdReaderRouter := todoRouter.PathPrefix(fmt.Sprintf("/{todo_id:%s}", constants.UUID_REGEX)).Subrouter()
	todoRouter.Use(middlewares.ExtractionTodoIdMiddlewareFunc)
	s.registerReadTodoIdPaths(todoIdReaderRouter)

	todoIdAuthRouter := todoRouter.PathPrefix(fmt.Sprintf("/{todo_id:%s}", constants.UUID_REGEX)).Subrouter()
	todoIdAuthRouter.Use(middlewares.ExtractionTodoIdMiddlewareFunc, middlewares.NewTodoModifyMiddlewareFunc(s.tServ, s.listServ, s.factory))
	s.registerTodoIdAuthRoutes(todoIdAuthRouter)

	todoListReaderRouter := listIdReadRouter.PathPrefix("/todos").Subrouter()
	s.registerReadTodoListPaths(todoListReaderRouter)

	todoListAuthRouter := listIdAuthRouter.PathPrefix("/todos").Subrouter()
	s.registerAuthTodoListPaths(todoListAuthRouter)

	userRouter := router.PathPrefix("/users").Subrouter()
	userIdReadRouter := userRouter.PathPrefix(fmt.Sprintf("/{user_id:%s}", constants.UUID_REGEX)).Subrouter()
	userIdReadRouter.Use(middlewares.ExtractionUserIdMiddlewareFunc)
	s.registerReadUserIdRoutes(userIdReadRouter)

	userIdAuthRouter := userRouter.PathPrefix(fmt.Sprintf("/{user_id:%s}", constants.UUID_REGEX)).Subrouter()
	userIdAuthRouter.Use(middlewares.ExtractionUserIdMiddlewareFunc, middlewares.UserAccessMiddlewareFunc)
	s.registerAuthUserIdRoutes(userIdAuthRouter)

	port := fmt.Sprintf(":%s", restServerPort)
	log.Fatal(http.ListenAndServe(port, oauthRouter))
}
