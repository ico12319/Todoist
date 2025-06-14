package middlewares

import (
	"github.com/stretchr/testify/require"
	"internProject/todo_app_service/pkg/constants"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContentTypeMiddleware_ServeHTTP(t *testing.T) {
	isNextHandlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isNextHandlerCalled = true
	})

	middleware := ContentTypeMiddlewareFunc(next)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	middleware.ServeHTTP(rr, req)

	require.Equal(t, constants.CONTENT_TYPE, rr.Header().Get("Content-Type"))
	require.True(t, isNextHandlerCalled)
}
