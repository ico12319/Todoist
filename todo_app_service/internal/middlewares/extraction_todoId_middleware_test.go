package middlewares

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractionTodoIdMiddleware_ServeHTTP(t *testing.T) {
	tests := []struct {
		testName         string
		respBody         string
		shouldCallNext   bool
		expectedHttpCode int
		requestUrlVars   map[string]string
	}{
		{
			"Middleware successfully extracts listId from request and sends it via the context to the next handler",
			`id=123`,
			true,
			http.StatusOK,
			map[string]string{
				"todo_id": "123",
			},
		},
		{
			"Bad request so the middleware should encode httpStatusBadRequest",
			`{"error":"invalid request: missing todo_id"}` + "\n",
			false,
			http.StatusBadRequest,
			map[string]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			isNextCalled := false

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				isNextCalled = true
				todoId := r.Context().Value(TodoId).(string)
				_, err := w.Write([]byte("id=" + todoId))
				require.NoError(t, err)
			})

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = mux.SetURLVars(req, test.requestUrlVars)

			middleware := ExtractionTodoIdMiddlewareFunc(next)
			middleware.ServeHTTP(rr, req)

			require.Equal(t, test.shouldCallNext, isNextCalled)
			require.Equal(t, test.expectedHttpCode, rr.Code)
			require.Equal(t, test.respBody, rr.Body.String())
		})
	}
}
