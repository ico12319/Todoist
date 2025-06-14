package middlewares

/*
import (
	"github.com/stretchr/testify/require"
	"internProject/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthorisationMiddleware_ServeHTTP(t *testing.T) {
	authorizedUsers := initAuthorizedUsers()

	tests := []struct {
		testName          string
		shouldCallNext    bool
		expectedHttpCode  int
		shouldEncodeError bool
		responseBody      string
		userEmail         string
		hasAuthHeader     bool
	}{
		{
			"The email of the user extracted from the context belongs to an admin and the middleware" +
				"sends it to the next handler and calls next",
			true,
			http.StatusOK,
			false,
			"",
			"hristo_partenov@abv.bg",
			true,
		},
		{
			"The email of the user extracted from the context belongs to a writer and the middleware" +
				"sends it to the next handler and calls next",
			true,
			http.StatusOK,
			false,
			"",
			"test@yahoo.com",
			true,
		},
		{
			"The email of the user extracted from the context belongs to a reader" +
				"and the middleware sends it to the next handler and calls next",
			true,
			http.StatusOK,
			false,
			"",
			"fake@gmail.com",
			true,
		},
		{
			"The email of the user extracted from the contest belongs to an unauthorized user and the middleware" +
				"encodes error and httpStatusUnauthorized",
			false,
			http.StatusUnauthorized,
			true,
			`{"error":"unauthorized"}` + "\n",
			"unauthorizedUser@abv.bg",
			true,
		},
		{
			"Authorization Header is missing and the middleware encodes error and httpStatusUnauthorized",
			false,
			http.StatusUnauthorized,
			true,
			`{"error":"unauthorized"}` + "\n",
			"",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			isNextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				isNextCalled = true
				userFromContext := r.Context().Value(UserKey).(models.User)
				require.Equal(t, authorizedUsers[test.userEmail], userFromContext)
			})
			middleware := newAuthorisationMiddleware(next, authorizedUsers)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			if test.hasAuthHeader {
				req.Header.Set("Authorization", test.userEmail)
			}
			middleware.ServeHTTP(rr, req)

			if test.shouldEncodeError {
				require.Equal(t, test.responseBody, rr.Body.String())
			}
			require.Equal(t, test.expectedHttpCode, rr.Code)
			require.Equal(t, test.shouldCallNext, isNextCalled)
		})
	}
}

func initAuthorizedUsers() map[string]models.User {
	return map[string]models.User{
		"hristo_partenov@abv.bg": {
			Email: "hristo_partenov@abv.bg",
			Role:  "admin",
		},
		"test@yahoo.com": {
			Email: "test@yahoo.com",
			Role:  "writer",
		},
		"fake@gmail.com": {
			Email: "fake@gmail.com",
			Role:  "reader",
		},
	}
}
*/
