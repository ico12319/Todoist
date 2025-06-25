package refresh

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/tokens/refresh/mocks"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRefreshTokenBuilder_GenerateRefreshToken(t *testing.T) {
	tests := []struct {
		testName       string
		mockTimeGen    func() *mocks.TimeGenerator
		mockJwtGetter  func() *mocks.JwtGetter
		expectedOutput string
		err            error
	}{
		{
			testName: "Successfully generating refresh token",

			mockTimeGen: func() *mocks.TimeGenerator {
				mTimeGen := &mocks.TimeGenerator{}

				mTimeGen.
					EXPECT().
					Now().
					Return(mockIssuedTime).Twice()

				return mTimeGen
			},

			mockJwtGetter: func() *mocks.JwtGetter {
				mJwtGetter := &mocks.JwtGetter{}

				claims := initClaims()

				jwToken := initJwt(claims)

				mJwtGetter.EXPECT().
					GetJWTWithClaims(jwt.SigningMethodHS256, claims).
					Return(jwToken).Once()

				configManager := log.GetInstance()
				configManager.JwtConfig.Secret = jwtKey

				mJwtGetter.EXPECT().
					GetSignedJWT(jwToken, jwtKey).
					Return(SIGNED_STRING, nil).Once()

				return mJwtGetter
			},

			expectedOutput: SIGNED_STRING,
		},
		{
			testName: "Failed to generate refresh token, nil invalid jwt key",

			mockTimeGen: func() *mocks.TimeGenerator {
				mTimeGen := &mocks.TimeGenerator{}

				mTimeGen.
					EXPECT().
					Now().
					Return(mockIssuedTime).Twice()

				return mTimeGen
			},

			mockJwtGetter: func() *mocks.JwtGetter {
				mJwtGetter := &mocks.JwtGetter{}

				claims := initClaims()
				jwToken := initJwt(claims)

				mJwtGetter.EXPECT().
					GetJWTWithClaims(jwt.SigningMethodHS256, claims).
					Return(jwToken).Once()

				configManager := log.GetInstance()
				configManager.JwtConfig.Secret = invalidJwt

				mJwtGetter.EXPECT().
					GetSignedJWT(jwToken, invalidJwt).
					Return("", jwt.ErrInvalidKeyType).Once()

				return mJwtGetter
			},
			err: errInvalidJwtKeyType,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mTimeGen := &mocks.TimeGenerator{}
			if test.mockTimeGen != nil {
				mTimeGen = test.mockTimeGen()
			}

			mJwtGetter := &mocks.JwtGetter{}
			if test.mockJwtGetter != nil {
				mJwtGetter = test.mockJwtGetter()
			}

			rTokenBuilder := NewRefreshTokenBuilder(mTimeGen, mJwtGetter)

			gotOutput, err := rTokenBuilder.GenerateRefreshToken(context.TODO())
			if test.err != nil {
				require.EqualError(t, test.err, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedOutput, gotOutput)
			mock.AssertExpectationsForObjects(t, mTimeGen, mJwtGetter)
		})
	}
}
