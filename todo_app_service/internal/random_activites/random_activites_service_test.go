package random_activites

import (
	"Todo-List/internProject/todo_app_service/internal/random_activites/mocks"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func TestService_Suggest(t *testing.T) {
	tests := []struct {
		testName               string
		mockHttpService        func() *mocks.HttpService
		err                    error
		expectedRandomActivity *models.RandomActivity
	}{
		{
			testName: "Successfully fetching a random activity",

			mockHttpService: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				randomActivityBytes, err := json.Marshal(mockRandomActivity)
				require.NoError(t, err)

				mockResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(randomActivityBytes)),
				}

				mck.EXPECT().
					GetHttpResponse(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(mockResponse, nil).
					Once()

				return mck
			},

			expectedRandomActivity: mockRandomActivity,
		},

		{
			testName: "Failed to get activity, error when trying to get http response",

			mockHttpService: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mck.EXPECT().
					GetHttpResponse(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(nil, errWhenTryingToGetHttpResponse).
					Once()

				return mck
			},

			err: errWhenTryingToGetHttpResponse,
		},

		{
			testName: "Failed to get activity, bad http status code received",

			mockHttpService: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockResponse := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponse(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(mockResponse, nil).
					Once()

				return mck
			},

			err: errWhenBadHttpStatusCodeIsReceived,
		},

		{
			testName: "Failed to fetch activity, error when trying to decode response body",

			mockHttpService: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponse(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(mockResponse, nil).
					Once()

				return mck
			},

			err: errWhenTryingToDecodeHttpResponseBody,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockHttpService := &mocks.HttpService{}
			if test.mockHttpService != nil {
				mockHttpService = test.mockHttpService()
			}

			activityService := NewService(mockUrl, mockHttpService)

			receivedRandomActivity, err := activityService.Suggest(context.TODO())
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedRandomActivity, receivedRandomActivity)
			mock.AssertExpectationsForObjects(t, mockHttpService)
		})

	}
}
