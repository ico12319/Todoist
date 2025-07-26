package oauth

import (
	"Todo-List/internProject/todo_app_service/internal/oauth/mocks"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_HandleLogin(t *testing.T) {
	globalRequest := setUpHttpRequestForLogin()
	httpRecorder := getHttpRecorder()

	tests := []struct {
		testName         string
		oauthServiceMock func() *mocks.OauthService
		httpServiceMock  func() *mocks.HttpService
		expectedHttpCode int
		expectHttpError  bool
	}{
		{
			testName: "Successfully handling login",

			oauthServiceMock: func() *mocks.OauthService {
				mck := &mocks.OauthService{}

				mck.EXPECT().
					LoginUrl(context.TODO()).
					Return(expectedOauthUrl, expectedOauthState, nil).
					Once()

				return mck
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mck.EXPECT().SetCookie(httpRecorder, &http.Cookie{
					Name:     cookieName,
					Path:     cookiePath,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
					Value:    expectedOauthState,
					Secure:   true,
				}).
					Once()

				mck.EXPECT().
					Redirect(httpRecorder, globalRequest, expectedOauthUrl, http.StatusTemporaryRedirect).
					Once()

				return mck
			},

			expectedHttpCode: http.StatusOK,
		},

		{
			testName: "Failed to handle callback error when calling oauth service",

			oauthServiceMock: func() *mocks.OauthService {
				mck := &mocks.OauthService{}

				mck.EXPECT().
					LoginUrl(context.TODO()).
					Return("", "", errWhenGeneratingState).
					Once()

				return mck
			},

			expectHttpError: true,

			expectedHttpCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			oauthServiceMock := &mocks.OauthService{}
			if test.oauthServiceMock != nil {
				oauthServiceMock = test.oauthServiceMock()
			}

			httpServiceMock := &mocks.HttpService{}
			if test.httpServiceMock != nil {
				httpServiceMock = test.httpServiceMock()
			}

			oauthHandler := NewHandler(oauthServiceMock, nil, httpServiceMock)
			oauthHandler.HandleLogin(httpRecorder, globalRequest)

			if test.expectHttpError {
				extractErrorFromResponseRecorder(t, httpRecorder, errWhenGeneratingState)
			}

			require.Equal(t, test.expectedHttpCode, httpRecorder.Code)
			mock.AssertExpectationsForObjects(t, oauthServiceMock, httpServiceMock)
		})
	}
}

func TestHandler_HandleCallback(t *testing.T) {

	tests := []struct {
		testName         string
		oauthServiceMock func() *mocks.OauthService
		jwtIssuerMock    func() *mocks.JwtIssuer
		requestMock      func() *http.Request
		err              error
		expectedHttpCode int
		expectedTokens   *models.CallbackResponse
	}{
		{
			testName: "Successfully handling callback",

			oauthServiceMock: func() *mocks.OauthService {
				return setUpCorrectOauthServiceMock()
			},

			jwtIssuerMock: func() *mocks.JwtIssuer {
				mck := &mocks.JwtIssuer{}

				mck.EXPECT().GetTokens(mock.Anything, accessToken).Return(&models.CallbackResponse{
					JwtToken:     jwtToken,
					RefreshToken: refreshToken,
				}, nil).Once()

				return mck
			},

			requestMock: func() *http.Request {
				return setUpCorrectRequestMock()
			},

			expectedHttpCode: http.StatusOK,

			expectedTokens: &models.CallbackResponse{
				JwtToken:     jwtToken,
				RefreshToken: refreshToken,
			},
		},

		{
			testName: "Failed to handle callback, empty auth code in url",

			requestMock: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				return req
			},

			expectedHttpCode: http.StatusBadRequest,

			err: errMissingAuthCodeInUrl,
		},

		{
			testName: "Failed to handle callback, error when calling oauth service",

			oauthServiceMock: func() *mocks.OauthService {
				mck := &mocks.OauthService{}

				mck.EXPECT().
					ExchangeCodeForToken(mock.Anything, authCode).
					Return("", errWhenCallingOauthServiceWithExchangeCodeForToken).
					Once()

				return mck
			},

			requestMock: func() *http.Request {
				return setUpCorrectRequestMock()
			},

			err: errWhenCallingOauthServiceWithExchangeCodeForToken,

			expectedHttpCode: http.StatusInternalServerError,
		},

		{
			testName: "Failed to handle callback, error when calling jwt issuer",

			oauthServiceMock: func() *mocks.OauthService {
				return setUpCorrectOauthServiceMock()
			},

			jwtIssuerMock: func() *mocks.JwtIssuer {
				mck := &mocks.JwtIssuer{}

				mck.EXPECT().
					GetTokens(mock.Anything, accessToken).
					Return(nil, errWhenCallingJwtIssuerGetTokens).
					Once()

				return mck
			},

			requestMock: func() *http.Request {
				return setUpCorrectRequestMock()
			},

			expectedHttpCode: http.StatusInternalServerError,

			err: errWhenCallingJwtIssuerGetTokens,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			httpRecorder := getHttpRecorder()

			oauthServiceMock := &mocks.OauthService{}
			if test.oauthServiceMock != nil {
				oauthServiceMock = test.oauthServiceMock()
			}

			jwtIssuerMock := &mocks.JwtIssuer{}
			if test.jwtIssuerMock != nil {
				jwtIssuerMock = test.jwtIssuerMock()
			}

			oauthHandler := NewHandler(oauthServiceMock, jwtIssuerMock, nil)
			oauthHandler.HandleCallback(httpRecorder, test.requestMock())

			if test.err != nil {
				extractErrorFromResponseRecorder(t, httpRecorder, test.err)
			} else {
				receivedTokens := extractCallbackResponseFromResponseRecorder(t, httpRecorder)
				require.Equal(t, test.expectedTokens, receivedTokens)
			}

			require.Equal(t, test.expectedHttpCode, httpRecorder.Code)
			mock.AssertExpectationsForObjects(t, oauthServiceMock, jwtIssuerMock)
		})
	}

}
