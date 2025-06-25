package tokens_test

import (
	"Todo-List/internProject/todo_app_service/pkg/tokens"
	"Todo-List/internProject/todo_app_service/pkg/tokens/mocks"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJwtParserService_ParseJWT(t *testing.T) {
	tests := []struct {
		testName       string
		tokenString    string
		mockParser     func() *mocks.Parser
		expectedOutput *tokens.Claims
		err            error
	}{
		{
			testName: "Successfully parsing jwt token",

			tokenString: SIGNED_JWT_STRING,

			mockParser: func() *mocks.Parser {
				mParser := &mocks.Parser{}

				claims := initClaims()
				token := initJWT(claims, &jwt.SigningMethodHMAC{}, true)

				mParser.EXPECT().
					ParseWithClaims(SIGNED_JWT_STRING, &tokens.Claims{}).
					Return(token, claims, nil).Once()

				return mParser
			},

			expectedOutput: initClaims(),
		},
		{
			testName: "Failed to parse jwt, error malformed token when calling parser",

			mockParser: func() *mocks.Parser {
				mParser := &mocks.Parser{}

				mParser.EXPECT().
					ParseWithClaims("", &tokens.Claims{}).
					Return(nil, nil, jwt.ErrTokenMalformed).Once()

				return mParser
			},

			err: errMalformedToken,
		},
		{
			testName: "Failed to parse jwt, unexpected signing method",

			tokenString: SIGNED_JWT_STRING,

			mockParser: func() *mocks.Parser {
				mParser := &mocks.Parser{}

				claims := initClaims()
				token := initJWT(claims, &jwt.SigningMethodECDSA{}, true)

				mParser.EXPECT().
					ParseWithClaims(SIGNED_JWT_STRING, &tokens.Claims{}).
					Return(token, claims, nil).Once()

				return mParser
			},

			err: errInvalidSigningMethod,
		},
		{
			testName: "Failed to parse jwt, invalid token",

			tokenString: SIGNED_JWT_STRING,

			mockParser: func() *mocks.Parser {
				mParser := &mocks.Parser{}

				claims := initClaims()
				token := initJWT(claims, &jwt.SigningMethodHMAC{}, false)

				mParser.EXPECT().
					ParseWithClaims(SIGNED_JWT_STRING, &tokens.Claims{}).
					Return(token, claims, nil).Once()

				return mParser
			},

			err: errInvalidToken,
		},
		{
			testName: "Failed to parse jwt, error expired token when calling parser",

			mockParser: func() *mocks.Parser {
				mParser := &mocks.Parser{}

				mParser.EXPECT().
					ParseWithClaims("", &tokens.Claims{}).
					Return(nil, nil, jwt.ErrTokenExpired).Once()

				return mParser
			},

			err: errTokenExpired,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mParser := &mocks.Parser{}

			if test.mockParser != nil {
				mParser = test.mockParser()
			}

			parserService := tokens.NewJwtParseService(mParser)

			gotOutput, err := parserService.ParseJWT(context.TODO(), test.tokenString)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedOutput, gotOutput)
			mock.AssertExpectationsForObjects(t, mParser)
		})
	}
}
