package middlewares

import (
	"Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

type uuidGenerator interface {
	Generate() string
}

type logEnrichMiddleware struct {
	next      http.Handler
	generator uuidGenerator
}

func newLogEnrichMiddleware(next http.Handler, generator uuidGenerator) *logEnrichMiddleware {
	return &logEnrichMiddleware{next: next, generator: generator}
}

func (l *logEnrichMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logEntry := log.C(ctx)

	logEntry = logEntry.WithFields(logrus.Fields{
		constants.REQUEST_ID: l.generator.Generate(),
	})

	ctx = log.ContextWithLogger(ctx, logEntry)
	l.next.ServeHTTP(w, r.WithContext(ctx))
}

func LogEnrichMiddlewareFunc(generator uuidGenerator) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newLogEnrichMiddleware(next, generator)
	}
}
