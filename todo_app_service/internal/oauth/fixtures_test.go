package oauth

import (
	"Todo-List/internProject/todo_app_service/internal/oauth/mocks"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	emptyOauthState    = ""
	expectedOauthState = "expected ready state"
	expectedOauthUrl   = "expectedUrl.com"
	authCode           = "valid auth code"
	accessToken        = "valid access token"
)

var (
	errZeroLengthState                                 = errors.New("zero length state generated")
	errWhenGeneratingState                             = errors.New("error when generating state")
	errWhenExchangingToken                             = errors.New("error when trying to exchange token")
	errEmptyAuthCode                                   = errors.New("empty authorization code")
	errMissingAuthCodeInUrl                            = errors.New("missing auth code in callback url")
	errWhenCallingOauthServiceWithExchangeCodeForToken = errors.New("error when trying to exchange auth code for access token")
	errWhenCallingJwtIssuerGetTokens                   = errors.New("error when trying to get tokens from jwt issuer")
)

var (
	cookieName = "oauth_state"
	cookiePath = "/"
)

var (
	jwtToken     = "jwt token"
	refreshToken = "refresh token"
)

func MatchHttpCookie(cookie *http.Cookie) bool {
	return cookie.HttpOnly && cookie.Secure &&
		cookie.Name == cookieName && cookie.Path == cookiePath &&
		cookie.SameSite == http.SameSiteLaxMode
}

func setUpHttpRequestForLogin() *http.Request {
	return httptest.NewRequestWithContext(context.TODO(), http.MethodGet, "/", nil)
}

func extractErrorFromResponseRecorder(tb testing.TB, rr *httptest.ResponseRecorder, err error) {
	tb.Helper()
	var got map[string]string
	require.NoError(tb, json.Unmarshal(rr.Body.Bytes(), &got))
	expect := map[string]string{
		"error": err.Error(),
	}
	require.Equal(tb, expect, got)
}

func extractCallbackResponseFromResponseRecorder(t *testing.T, rr *httptest.ResponseRecorder) *models.CallbackResponse {
	var callback models.CallbackResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &callback))

	return &callback
}

func getHttpRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

func setUpCorrectRequestMock() *http.Request {
	target := "/?code=" + url.QueryEscape(authCode)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	return req
}

func setUpCorrectOauthServiceMock() *mocks.OauthService {
	mck := &mocks.OauthService{}

	mck.EXPECT().ExchangeCodeForToken(mock.Anything, authCode).
		Return(accessToken, nil).Once()

	return mck
}
