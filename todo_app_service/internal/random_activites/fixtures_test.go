package random_activites

import (
	"Todo-List/internProject/todo_app_service/pkg/models"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	mockUrl = "test.com"
)

var (
	activity     = "activity"
	activityType = "type"
	participants = 2
	kidFriendly  = true
)

var (
	mockRandomActivity = initRandomActivity(activity, activityType, participants, kidFriendly)
	requestMock        = httptest.NewRequest(http.MethodGet, "/", nil)
)

var (
	errWhenTryingToGetHttpResponse        = errors.New("error when trying to make http response")
	errWhenBadHttpStatusCodeIsReceived    = errors.New("error when trying to suggest an activity")
	errWhenTryingToDecodeHttpResponseBody = errors.New("EOF")
	errWhenCallingActivityService         = errors.New("error when calling activity service")
	errWhenTryingToJSONEncode             = errors.New("error when trying to JSON encode")
)

func initRandomActivity(activity string, activityType string, participants int, kidFriendly bool) *models.RandomActivity {
	return &models.RandomActivity{
		Activity:     activity,
		Type:         activityType,
		Participants: participants,
		KidFriendly:  kidFriendly,
	}
}

func extractErrorFromResponseRecorder(t *testing.T, rr *httptest.ResponseRecorder) error {
	t.Helper()
	var got map[string]string
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))

	return errors.New(got["error"])
}

func extractRandomActivityFromResponseRecorder(t *testing.T, rr *httptest.ResponseRecorder) *models.RandomActivity {
	t.Helper()

	var randomActivity models.RandomActivity
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&randomActivity))

	return &randomActivity
}
