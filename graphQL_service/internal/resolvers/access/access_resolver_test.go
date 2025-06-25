package access

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/resolvers/access/mocks"
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"io"
	"net/http"
	"testing"
)

func TestResolver_ExchangeRefreshToken(t *testing.T) {
	refreshInput := initRefreshInput()
	handlerInput := initHandlerModel()
	model := initModel()
	gqlModel := initGqlModel()
	refreshInputBytes := getBytesOfEntity(refreshInput)
	modelBytes := getBytesOfEntity(model)
	mockRequest := initMockRequest(refreshInput)

	tests := []struct {
		testName           string
		mockClient         func() *mocks.HttpClient
		mockAuthSetter     func() *mocks.RequestAuthSetter
		mockConverter      func() *mocks.AccessConverter
		mockJsonMarshaller func() *mocks.JsonMarshaller
		mockHttpRequester  func() *mocks.HttpRequester
		err                error
		expectedOutput     *gql.Access
	}{
		{
			testName: "Successfully receiving expected Access output",

			mockClient: func() *mocks.HttpClient {
				mClient := &mocks.HttpClient{}

				mockResponse := &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(bytes.NewReader(modelBytes)),
				}

				mClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPost && req.URL.Path == "/tokens/refresh"
				})).Return(mockResponse, nil).Once()

				return mClient
			},

			mockAuthSetter: func() *mocks.RequestAuthSetter {
				mAuthSetter := &mocks.RequestAuthSetter{}

				decoratedRequest := mockRequest
				decoratedRequest.Header.Set("Authorization", AUTH)

				mAuthSetter.EXPECT().
					DecorateRequest(context.TODO(), mockRequest).
					Return(decoratedRequest, nil).Once()

				return mAuthSetter
			},

			mockConverter: func() *mocks.AccessConverter {
				mConverter := &mocks.AccessConverter{}

				mConverter.EXPECT().
					ToHandlerModelRefresh(refreshInput).
					Return(handlerInput).Once()

				mConverter.EXPECT().
					ToGQL(model).
					Return(gqlModel).Once()
				return mConverter
			},

			mockJsonMarshaller: func() *mocks.JsonMarshaller {
				mJsonMarshaller := &mocks.JsonMarshaller{}

				mJsonMarshaller.EXPECT().
					Marshal(handlerInput).
					Return(refreshInputBytes, nil).Once()

				return mJsonMarshaller
			},

			mockHttpRequester: func() *mocks.HttpRequester {
				mRequest := &mocks.HttpRequester{}

				mRequest.EXPECT().
					NewRequestWithContext(context.TODO(), http.MethodPost, URL, bytes.NewReader(refreshInputBytes)).
					Return(mockRequest, nil).Once()

				return mRequest
			},

			expectedOutput: gqlModel,
		},
		{
			testName: "Failed to exchange refresh token, error when trying to JSON marshal",

			mockConverter: func() *mocks.AccessConverter {
				mConverter := &mocks.AccessConverter{}

				mConverter.EXPECT().
					ToHandlerModelRefresh(refreshInput).
					Return(handlerInput).Once()
				return mConverter
			},

			mockJsonMarshaller: func() *mocks.JsonMarshaller {
				mJsonMarshaller := &mocks.JsonMarshaller{}

				mJsonMarshaller.EXPECT().Marshal(handlerInput).
					Return(nil, assert.AnError).Once()
				return mJsonMarshaller
			},

			err: jsonMarshalError,

			expectedOutput: &gql.Access{},
		},
		{
			testName: "Failed to exchange refresh token, error when trying to make http request",

			mockConverter: func() *mocks.AccessConverter {
				mConverter := &mocks.AccessConverter{}

				mConverter.EXPECT().
					ToHandlerModelRefresh(refreshInput).
					Return(handlerInput).Once()
				return mConverter
			},

			mockJsonMarshaller: func() *mocks.JsonMarshaller {
				mJsonMarshaller := &mocks.JsonMarshaller{}

				mJsonMarshaller.EXPECT().Marshal(handlerInput).
					Return(refreshInputBytes, nil).Once()
				return mJsonMarshaller
			},

			mockHttpRequester: func() *mocks.HttpRequester {
				mHttpRequester := &mocks.HttpRequester{}

				mHttpRequester.EXPECT().
					NewRequestWithContext(context.TODO(), http.MethodPost, URL, bytes.NewReader(refreshInputBytes)).
					Return(nil, assert.AnError).Once()

				return mHttpRequester
			},

			err: requestError,

			expectedOutput: &gql.Access{},
		},
		{
			testName: "Failed to exchange refresh token, error when trying to decorate http request",

			mockConverter: func() *mocks.AccessConverter {
				mConverter := &mocks.AccessConverter{}

				mConverter.EXPECT().
					ToHandlerModelRefresh(refreshInput).
					Return(handlerInput).Once()
				return mConverter
			},

			mockJsonMarshaller: func() *mocks.JsonMarshaller {
				mJsonMarshaller := &mocks.JsonMarshaller{}

				mJsonMarshaller.EXPECT().Marshal(handlerInput).
					Return(refreshInputBytes, nil).Once()
				return mJsonMarshaller
			},

			mockHttpRequester: func() *mocks.HttpRequester {
				mHttpRequester := &mocks.HttpRequester{}

				mHttpRequester.EXPECT().
					NewRequestWithContext(context.TODO(), http.MethodPost, URL, bytes.NewReader(refreshInputBytes)).
					Return(mockRequest, nil).Once()

				return mHttpRequester
			},

			mockAuthSetter: func() *mocks.RequestAuthSetter {
				mAuthSetter := &mocks.RequestAuthSetter{}

				mAuthSetter.EXPECT().
					DecorateRequest(context.TODO(), mockRequest).
					Return(nil, errorByRequestDecorator).Once()

				return mAuthSetter
			},
			err: errorByRequestDecorator,

			expectedOutput: &gql.Access{},
		},
		{
			testName: "Failed to exchange refresh token, error when trying to get http response",

			mockConverter: func() *mocks.AccessConverter {
				mConverter := &mocks.AccessConverter{}

				mConverter.EXPECT().
					ToHandlerModelRefresh(refreshInput).
					Return(handlerInput).Once()
				return mConverter
			},

			mockJsonMarshaller: func() *mocks.JsonMarshaller {
				mJsonMarshaller := &mocks.JsonMarshaller{}

				mJsonMarshaller.EXPECT().Marshal(handlerInput).
					Return(refreshInputBytes, nil).Once()
				return mJsonMarshaller
			},

			mockHttpRequester: func() *mocks.HttpRequester {
				mHttpRequester := &mocks.HttpRequester{}

				mHttpRequester.EXPECT().
					NewRequestWithContext(context.TODO(), http.MethodPost, URL, bytes.NewReader(refreshInputBytes)).
					Return(mockRequest, nil).Once()

				return mHttpRequester
			},

			mockAuthSetter: func() *mocks.RequestAuthSetter {
				mAuthSetter := &mocks.RequestAuthSetter{}

				decoratedRequest := mockRequest
				decoratedRequest.Header.Set("Authorization", AUTH)

				mAuthSetter.EXPECT().
					DecorateRequest(context.TODO(), mockRequest).
					Return(decoratedRequest, nil).Once()

				return mAuthSetter
			},

			mockClient: func() *mocks.HttpClient {
				mClient := &mocks.HttpClient{}

				mClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPost && req.Header.Get("Authorization") == AUTH && req.URL.Path == "/tokens/refresh"
				})).Return(nil, httpClientError).Once()

				return mClient
			},

			err: httpClientError,

			expectedOutput: &gql.Access{},
		},
		{
			testName: "Failed to exchange refresh token, error Internal Server error internal server error returner by rest_api",
			mockConverter: func() *mocks.AccessConverter {
				mConverter := &mocks.AccessConverter{}

				mConverter.EXPECT().
					ToHandlerModelRefresh(refreshInput).
					Return(handlerInput).Once()
				return mConverter
			},

			mockJsonMarshaller: func() *mocks.JsonMarshaller {
				mJsonMarshaller := &mocks.JsonMarshaller{}

				mJsonMarshaller.EXPECT().Marshal(handlerInput).
					Return(refreshInputBytes, nil).Once()
				return mJsonMarshaller
			},

			mockHttpRequester: func() *mocks.HttpRequester {
				mHttpRequester := &mocks.HttpRequester{}

				mHttpRequester.EXPECT().
					NewRequestWithContext(context.TODO(), http.MethodPost, URL, bytes.NewReader(refreshInputBytes)).
					Return(mockRequest, nil).Once()

				return mHttpRequester
			},

			mockAuthSetter: func() *mocks.RequestAuthSetter {
				mAuthSetter := &mocks.RequestAuthSetter{}

				decoratedRequest := mockRequest
				decoratedRequest.Header.Set("Authorization", AUTH)

				mAuthSetter.EXPECT().
					DecorateRequest(context.TODO(), mockRequest).
					Return(decoratedRequest, nil).Once()

				return mAuthSetter
			},

			mockClient: func() *mocks.HttpClient {
				mClient := &mocks.HttpClient{}

				mockResponse := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPost && req.Header.Get("Authorization") == AUTH && req.URL.Path == "/tokens/refresh"
				})).Return(mockResponse, nil).Once()

				return mClient
			},

			err: &gqlerror.Error{
				Message:    "Internal error, please try again later.",
				Extensions: map[string]interface{}{"code": "INTERNAL_SERVER_ERROR"},
			},
			expectedOutput: &gql.Access{},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mClient := &mocks.HttpClient{}
			if test.mockClient != nil {
				mClient = test.mockClient()
			}

			mAuthSetter := &mocks.RequestAuthSetter{}
			if test.mockAuthSetter != nil {
				mAuthSetter = test.mockAuthSetter()
			}

			mConverter := &mocks.AccessConverter{}
			if test.mockConverter != nil {
				mConverter = test.mockConverter()
			}

			mJsonMarshaller := &mocks.JsonMarshaller{}
			if test.mockJsonMarshaller != nil {
				mJsonMarshaller = test.mockJsonMarshaller()
			}

			mHttpRequester := &mocks.HttpRequester{}
			if test.mockHttpRequester != nil {
				mHttpRequester = test.mockHttpRequester()
			}

			aResolver := NewResolver(mClient, mAuthSetter, mConverter, mJsonMarshaller, mHttpRequester, restUrl)

			gotAccess, err := aResolver.ExchangeRefreshToken(context.TODO(), *refreshInput)
			if test.err != nil {
				require.EqualError(t, test.err, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedOutput, gotAccess)

			mock.AssertExpectationsForObjects(t, mClient, mAuthSetter, mConverter, mJsonMarshaller, mHttpRequester)
		})
	}
}
