package refresh

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
	"testing"
)

func TestSqlRefreshDB_CreateRefreshToken(t *testing.T) {
	tests := []struct {
		testName       string
		passedEntity   *entities.Refresh
		dbMock         func(mck sqlmock.Sqlmock)
		err            error
		expectedEntity *entities.Refresh
	}{
		{
			testName: "Successfully creating refresh token record",

			passedEntity: initRefreshEntity(VALID_REFRESH_TOKEN, validUserId),

			dbMock: func(mck sqlmock.Sqlmock) {
				mck.ExpectQuery()
			},
		},
	}
}
