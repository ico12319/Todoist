package middlewares

import (
	"context"
	"github.com/stretchr/testify/require"
	"internProject/todo_app_service/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestObjectCreationMiddleware_ServeHTTP(t *testing.T) {
	tests := []struct {
		testName         string
		encodesError     bool
		responseBody     string
		shouldCallNext   bool
		contextValue     *models.User
		expectedHttpCode int
	}{
		{
			"User from the context is an admin and the middleware successfully sends it through the context" +
				"and calls next",
			false,
			"",
			true,
			&models.User{
				Email: "email",
				Role:  "admin",
			},
			http.StatusOK,
		},
		{
			"User from the context is a writer and the middleware successfully sends it through the context" +
				"and calls next",
			false,
			"",
			true,
			&models.User{
				Email: "email",
				Role:  "writer",
			},
			http.StatusOK,
		},
		{
			"User from the context is not an admin or writer so the middleware enocodes httpStatusUnauthorized",
			true,
			`{"error":"unauthorized user"}` + "\n",
			false,
			&models.User{
				Email: "email",
				Role:  "role",
			},
			http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			isNextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				isNextCalled = true
				user := r.Context().Value(UserKey).(models.User)
				require.Equal(t, test.contextValue, user)
			})

			rr := httptest.NewRecorder()
			ctx := context.Background()
			ctx = context.WithValue(ctx, UserKey, test.contextValue)
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

			middleware := ObjectCreationMiddlewareFunc(next)
			middleware.ServeHTTP(rr, req)

			if test.encodesError {
				require.Equal(t, test.responseBody, rr.Body.String())
			}

			require.Equal(t, test.expectedHttpCode, rr.Code)
			require.Equal(t, test.shouldCallNext, isNextCalled)
		})
	}
}
