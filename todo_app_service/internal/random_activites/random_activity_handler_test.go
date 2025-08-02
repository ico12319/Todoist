package random_activites

import (
	"Todo-List/internProject/todo_app_service/internal/random_activites/mocks"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_HandleSuggestion(t *testing.T) {
	tests := []struct {
		testName                  string
		mockRandomActivityService func() *mocks.RandomActivityService
		expectedHttpStatusCode    int
		err                       error
		expectedRandomActivity    *models.RandomActivity
	}{
		{
			testName: "Successfully handling suggestion",

			mockRandomActivityService: func() *mocks.RandomActivityService {
				mck := &mocks.RandomActivityService{}

				mck.EXPECT().
					Suggest(mock.Anything).
					Return(mockRandomActivity, nil).
					Once()

				return mck
			},
			
			expectedHttpStatusCode: http.StatusOK,

			expectedRandomActivity: mockRandomActivity,
		},

		{
			testName: "Failed to handle suggestion, error when calling activity service",

			mockRandomActivityService: func() *mocks.RandomActivityService {
				mck := &mocks.RandomActivityService{}

				mck.EXPECT().
					Suggest(mock.Anything).
					Return(nil, errWhenCallingActivityService).
					Once()

				return mck
			},

			expectedHttpStatusCode: http.StatusInternalServerError,

			err: errWhenCallingActivityService,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			httpRecorder := httptest.NewRecorder()

			mockRandomActivityService := &mocks.RandomActivityService{}
			if test.mockRandomActivityService != nil {
				mockRandomActivityService = test.mockRandomActivityService()
			}

			randomActivityHandler := NewHandler(mockRandomActivityService)
			randomActivityHandler.HandleSuggestion(httpRecorder, requestMock)

			if test.err != nil {
				receivedError := extractErrorFromResponseRecorder(t, httpRecorder)
				require.EqualError(t, receivedError, test.err.Error())
			} else {
				receivedRandomActivity := extractRandomActivityFromResponseRecorder(t, httpRecorder)
				require.Equal(t, test.expectedRandomActivity, receivedRandomActivity)
			}

			require.Equal(t, test.expectedHttpStatusCode, httpRecorder.Code)

			mock.AssertExpectationsForObjects(t, mockRandomActivityService)
		})
	}

}
