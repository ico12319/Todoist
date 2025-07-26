package activity

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/resolvers/activity/mocks"
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func TestResolver_RandomActivity(t *testing.T) {
	tests := []struct {
		testName               string
		httpServiceMock        func(t *testing.T) *mocks.HttpService
		converterMock          func() *mocks.ActivityConverter
		err                    error
		expectedRandomActivity *gql.RandomActivity
	}{
		{
			testName: "Successfully getting random activity",

			httpServiceMock: func(t *testing.T) *mocks.HttpService {
				mck := &mocks.HttpService{}

				expectedModelRandomActivity := initExpectedModelRandomActivity()
				expectedModelRandomActivityBytes, err := json.Marshal(expectedModelRandomActivity)
				require.NoError(t, err)

				mck.EXPECT().GetHttpResponse(mock.Anything, http.MethodGet, url, nil).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(expectedModelRandomActivityBytes)),
				}, nil).
					Once()

				return mck
			},

			converterMock: func() *mocks.ActivityConverter {
				mck := &mocks.ActivityConverter{}

				mck.EXPECT().
					ToGQL(initExpectedModelRandomActivity()).
					Return(initExpectedGqlRandomActivity()).
					Once()

				return mck
			},

			expectedRandomActivity: initExpectedGqlRandomActivity(),
		},

		{
			testName: "Failed to get random activity, error when trying to get http response",

			httpServiceMock: func(t *testing.T) *mocks.HttpService {
				mck := &mocks.HttpService{}

				mck.EXPECT().
					GetHttpResponse(mock.Anything, http.MethodGet, url, nil).
					Return(nil, errByHttpService).
					Once()

				return mck
			},

			err: errByHttpService,

			expectedRandomActivity: initEmptyGqlRandomActivity(),
		},

		{
			testName: "Failed to get random activity, error due to response containing bad status code",

			httpServiceMock: func(t *testing.T) *mocks.HttpService {
				mck := &mocks.HttpService{}

				mck.EXPECT().GetHttpResponse(mock.Anything, http.MethodGet, url, nil).Return(&http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil).
					Once()

				return mck
			},

			err: errHttpStatusInternalServerError,

			expectedRandomActivity: initEmptyGqlRandomActivity(),
		},

		{
			testName: "Failed to get random activity, error due to response ot containing random activity model",

			httpServiceMock: func(t *testing.T) *mocks.HttpService {
				mck := &mocks.HttpService{}

				mck.EXPECT().GetHttpResponse(mock.Anything, http.MethodGet, url, nil).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil).
					Once()

				return mck
			},

			err: errWhenDecodingJSON,

			expectedRandomActivity: initEmptyGqlRandomActivity(),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			httpServiceMock := &mocks.HttpService{}
			if test.httpServiceMock != nil {
				httpServiceMock = test.httpServiceMock(t)
			}

			converterMock := &mocks.ActivityConverter{}
			if test.converterMock != nil {
				converterMock = test.converterMock()
			}

			activityResolver := NewResolver(restUrl, httpServiceMock, converterMock)

			receivedRandomActivity, err := activityResolver.RandomActivity(context.TODO())
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedRandomActivity, receivedRandomActivity)
			mock.AssertExpectationsForObjects(t, httpServiceMock, converterMock)
		})
	}

}
