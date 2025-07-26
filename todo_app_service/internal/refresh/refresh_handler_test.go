package refresh

import (
	"Todo-List/internProject/todo_app_service/internal/refresh/mocks"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_HandleRefresh(t *testing.T) {
	tests := []struct {
		testName              string
		jwtIssuerMock         func() *mocks.JwtIssuer
		httpRequestMock       func(t *testing.T) *http.Request
		expectedHttpCode      int
		err                   error
		expectedRenewedTokens *models.CallbackResponse
	}{
		{
			testName: "Successfully handling refresh",

			jwtIssuerMock: func() *mocks.JwtIssuer {
				mck := &mocks.JwtIssuer{}

				mck.EXPECT().
					GetRenewedTokens(mock.Anything, getInjectedHandlerModelRefreshInRequestBody()).
					Return(getExpectedCallbackResponse(), nil).
					Once()

				return mck
			},

			httpRequestMock: func(t *testing.T) *http.Request {
				return getValidHttpRequestWithInjectedHandlerModel(t)
			},

			expectedHttpCode: http.StatusOK,

			expectedRenewedTokens: getExpectedCallbackResponse(),
		},

		{
			testName: "Failed to handle refresh missing handler model refresh in request",

			httpRequestMock: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/tokens/refresh", nil)
				return req
			},

			expectedHttpCode: http.StatusBadRequest,

			err: errInvalidRequestBody,

			expectedRenewedTokens: getEmptyCallbackResponse(),
		},

		{
			testName: "Failed to handle refresh error when calling jwt issuer",

			jwtIssuerMock: func() *mocks.JwtIssuer {
				mck := &mocks.JwtIssuer{}

				mck.EXPECT().
					GetRenewedTokens(mock.Anything, getInjectedHandlerModelRefreshInRequestBody()).
					Return(nil, errWhenCallingJwtIssuer).
					Once()

				return mck
			},

			httpRequestMock: func(t *testing.T) *http.Request {
				return getValidHttpRequestWithInjectedHandlerModel(t)
			},

			expectedHttpCode: http.StatusInternalServerError,

			err: errWhenCallingJwtIssuer,

			expectedRenewedTokens: getEmptyCallbackResponse(),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			httpRecorder := httptest.NewRecorder()

			jwtIssuerMock := &mocks.JwtIssuer{}
			if test.jwtIssuerMock != nil {
				jwtIssuerMock = test.jwtIssuerMock()
			}

			req := test.httpRequestMock(t)

			refreshHandler := NewHandler(jwtIssuerMock)
			refreshHandler.HandleRefresh(httpRecorder, req)

			if test.err != nil {
				extractErrorFromResponseRecorder(t, httpRecorder, test.err)
			}

			receivedRenewedTokens := extractCallbackResponseFromHttpRecorder(t, httpRecorder)
			require.Equal(t, test.expectedRenewedTokens, receivedRenewedTokens)
			require.Equal(t, test.expectedHttpCode, httpRecorder.Code)

			mock.AssertExpectationsForObjects(t, jwtIssuerMock)
		})
	}

}
