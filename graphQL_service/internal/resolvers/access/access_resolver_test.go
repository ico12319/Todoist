package access

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/resolvers/access/mocks"
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

	tests := []struct {
		testName               string
		mockConverter          func() *mocks.AccessConverter
		mockJsonMarshaller     func() *mocks.JsonMarshaller
		mockHttpResponseGetter func() *mocks.HttpResponseGetter
		err                    error
		expectedOutput         *gql.Access
	}{
		{
			testName: "Successfully receiving expected Access output",

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

			mockHttpResponseGetter: func() *mocks.HttpResponseGetter {
				mResponseGetter := &mocks.HttpResponseGetter{}

				mockResponse := &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(bytes.NewReader(modelBytes)),
				}

				mResponseGetter.EXPECT().
					GetHttpResponse(context.TODO(), http.MethodPost, URL, bytes.NewReader(refreshInputBytes)).
					Return(mockResponse, nil).Once()

				return mResponseGetter
			},

			expectedOutput: gqlModel,
		},
		{
			testName: "Failed to exchange refresh token, error when trying to JSON marshal handler model",

			mockConverter: func() *mocks.AccessConverter {
				mConverter := &mocks.AccessConverter{}

				mConverter.EXPECT().
					ToHandlerModelRefresh(refreshInput).
					Return(handlerInput).Once()

				return mConverter
			},

			mockJsonMarshaller: func() *mocks.JsonMarshaller {
				mJsonMarshaller := &mocks.JsonMarshaller{}

				mJsonMarshaller.EXPECT().
					Marshal(handlerInput).
					Return(nil, assert.AnError).Once()

				return mJsonMarshaller

			},

			err: jsonMarshalError,

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

				mJsonMarshaller.EXPECT().
					Marshal(handlerInput).
					Return(refreshInputBytes, nil).Once()

				return mJsonMarshaller
			},

			mockHttpResponseGetter: func() *mocks.HttpResponseGetter {
				mHttpResponseGetter := &mocks.HttpResponseGetter{}

				mHttpResponseGetter.EXPECT().
					GetHttpResponse(context.TODO(), http.MethodPost, URL, bytes.NewReader(refreshInputBytes)).
					Return(nil, assert.AnError).Once()

				return mHttpResponseGetter
			},

			err: responseError,

			expectedOutput: &gql.Access{},
		},

		{
			testName: "Failed to exchange refresh token, http response with status code Internal server error returned",

			mockConverter: func() *mocks.AccessConverter {
				mConverter := &mocks.AccessConverter{}

				mConverter.EXPECT().
					ToHandlerModelRefresh(refreshInput).
					Return(handlerInput).Once()

				return mConverter
			},

			mockJsonMarshaller: func() *mocks.JsonMarshaller {
				mJsonMarshaller := &mocks.JsonMarshaller{}

				mJsonMarshaller.EXPECT().
					Marshal(handlerInput).
					Return(refreshInputBytes, nil).Once()

				return mJsonMarshaller
			},

			mockHttpResponseGetter: func() *mocks.HttpResponseGetter {
				mHttpResponseGetter := &mocks.HttpResponseGetter{}

				mockResponse := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mHttpResponseGetter.EXPECT().
					GetHttpResponse(context.TODO(), http.MethodPost, URL, bytes.NewReader(refreshInputBytes)).
					Return(mockResponse, nil).Once()

				return mHttpResponseGetter
			},

			err: errorWhenHandlingHttpStatusCodeInternalServerError,

			expectedOutput: &gql.Access{},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			mConverter := &mocks.AccessConverter{}
			if test.mockConverter != nil {
				mConverter = test.mockConverter()
			}

			mJsonMarshaller := &mocks.JsonMarshaller{}
			if test.mockJsonMarshaller != nil {
				mJsonMarshaller = test.mockJsonMarshaller()
			}

			mResponseGetter := &mocks.HttpResponseGetter{}
			if test.mockHttpResponseGetter != nil {
				mResponseGetter = test.mockHttpResponseGetter()
			}

			aResolver := NewResolver(mConverter, mJsonMarshaller, mResponseGetter, restUrl)

			gotAccess, err := aResolver.ExchangeRefreshToken(context.TODO(), *refreshInput)
			if test.err != nil {
				require.EqualError(t, test.err, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedOutput, gotAccess)

			mock.AssertExpectationsForObjects(t, mConverter, mJsonMarshaller, mResponseGetter)
		})
	}
}
