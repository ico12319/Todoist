package oauth

import (
	"Todo-List/internProject/todo_app_service/internal/oauth/mocks"
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"testing"
)

func TestOauthService_LoginUrl(t *testing.T) {
	tests := []struct {
		testName              string
		stateGeneratorMock    func() *mocks.StateGenerator
		oauthConfiguratorMock func() *mocks.OauthConfigurator
		err                   error
		expectedState         string
		expectedUrl           string
	}{
		{
			testName: "Successfully getting state and url",

			stateGeneratorMock: func() *mocks.StateGenerator {
				mck := &mocks.StateGenerator{}
				mck.EXPECT().
					GenerateState().
					Return(expectedOauthState, nil).
					Once()
				return mck
			},

			oauthConfiguratorMock: func() *mocks.OauthConfigurator {
				mck := &mocks.OauthConfigurator{}
				mck.EXPECT().
					AuthCodeURL(expectedOauthState).
					Return(expectedOauthUrl).
					Once()
				return mck
			},
			expectedState: expectedOauthState,

			expectedUrl: expectedOauthUrl,
		},

		{
			testName: "Failed to get state and url because generated state is empty",

			stateGeneratorMock: func() *mocks.StateGenerator {
				mck := &mocks.StateGenerator{}

				mck.EXPECT().
					GenerateState().
					Return(emptyOauthState, nil).
					Once()

				return mck
			},

			err: errZeroLengthState,
		},

		{
			testName: "Failed to get state and url because state generator returned error",

			stateGeneratorMock: func() *mocks.StateGenerator {
				mck := &mocks.StateGenerator{}

				mck.EXPECT().
					GenerateState().
					Return("", errWhenGeneratingState).
					Once()

				return mck
			},

			err: errWhenGeneratingState,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			stateGeneratorMock := &mocks.StateGenerator{}
			if test.stateGeneratorMock != nil {
				stateGeneratorMock = test.stateGeneratorMock()
			}

			oauthConfiguratorMock := &mocks.OauthConfigurator{}
			if test.oauthConfiguratorMock != nil {
				oauthConfiguratorMock = test.oauthConfiguratorMock()
			}

			oaService := NewService(stateGeneratorMock, oauthConfiguratorMock)

			receivedUrl, receivedState, err := oaService.LoginUrl(context.TODO())
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedUrl, receivedUrl)
			require.Equal(t, test.expectedState, receivedState)
			mock.AssertExpectationsForObjects(t, stateGeneratorMock, oauthConfiguratorMock)
		})
	}

}

func TestService_ExchangeCodeForToken(t *testing.T) {
	tests := []struct {
		testName              string
		authCode              string
		oauthConfiguratorMock func() *mocks.OauthConfigurator
		err                   error
		expectedAccessToken   string
	}{
		{
			testName: "Successfully receiving expected access token",

			authCode: authCode,

			oauthConfiguratorMock: func() *mocks.OauthConfigurator {
				mck := &mocks.OauthConfigurator{}

				mck.EXPECT().Exchange(context.TODO(), authCode).Return(&oauth2.Token{
					AccessToken: accessToken,
				}, nil).
					Once()

				return mck
			},

			expectedAccessToken: accessToken,
		},

		{
			testName: "Failed to receive access token due to error when performing exchange",

			authCode: authCode,

			oauthConfiguratorMock: func() *mocks.OauthConfigurator {
				mck := &mocks.OauthConfigurator{}

				mck.EXPECT().
					Exchange(context.TODO(), authCode).
					Return(nil, errWhenExchangingToken).
					Once()

				return mck
			},

			err: errWhenExchangingToken,
		},

		{
			testName: "Failed to receive access token because passed auth code was empty",

			err: errEmptyAuthCode,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			oauthConfiguratorMock := &mocks.OauthConfigurator{}

			if test.oauthConfiguratorMock != nil {
				oauthConfiguratorMock = test.oauthConfiguratorMock()
			}

			oaService := NewService(nil, oauthConfiguratorMock)

			receivedAccessToken, err := oaService.ExchangeCodeForToken(context.TODO(), test.authCode)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedAccessToken, receivedAccessToken)
			mock.AssertExpectationsForObjects(t, oauthConfiguratorMock)
		})
	}

}
