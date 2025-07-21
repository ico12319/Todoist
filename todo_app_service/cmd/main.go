package main

import (
	_ "Todo-List/internProject/todo_app_service/internal/sql_query_decorators/sql_decorators_creators"
	"Todo-List/internProject/todo_app_service/pkg/application"
	_ "github.com/lib/pq"
)

func main() {
	restServer := application.NewServer()
	restServer.Start()
}
