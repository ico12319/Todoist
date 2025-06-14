package middlewares

import (
	"context"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListDeletionMiddleware_ServeHTTP(t *testing.T) {
	tests := []struct {
		testName         string
		shouldCallNext   bool
		encodesError     bool
		respBody         string
		requestContext   userRoleKey
		expectedHttpCode int
	}{
		{
			"User from the context is an admin so the middlewares successfully calls next",
			true,
			false,
			"",
			userRoleKey{
				role:    "admin",
				isOwner: false,
			},
			http.StatusOK,
		},
		{
			"User from the context is an owner so the middleware successfully calls next",
			true,
			false,
			"",
			userRoleKey{
				role:    "writer",
				isOwner: true,
			},
			http.StatusOK,
		},
		{
			"User from the context is unauthorized so the middlewares encodes httpStatusUnauthorized",
			false,
			true,
			`{"error":"unauthorized"}` + "\n",
			userRoleKey{
				role:    "writer",
				isOwner: false,
			},
			http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			isNextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				isNextCalled = true
			})

			rr := httptest.NewRecorder()
			ctx := context.Background()
			ctx = context.WithValue(ctx, UserRoleKey, test.requestContext)
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

			middleware := ListDeletionMiddlewareFunc(next)
			middleware.ServeHTTP(rr, req)

			if test.encodesError {
				require.Equal(t, test.respBody, rr.Body.String())
			}
			require.Equal(t, test.shouldCallNext, isNextCalled)
			require.Equal(t, test.expectedHttpCode, rr.Code)
		})
	}
}
