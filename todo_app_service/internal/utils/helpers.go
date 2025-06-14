package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/application_errors"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/entities"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/oauth/githubModels"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/constants"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
)

func EncodeError(w http.ResponseWriter, errString string, httpStatusCode int) {
	w.WriteHeader(httpStatusCode)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": errString}); err != nil {
		http.Error(w, err.Error(), httpStatusCode)
		return
	}
}

func ExtractDueDateValueFromSQLNull(todo *entities.Todo) *time.Time {
	var dueDate *time.Time
	if todo.DueDate.Valid {
		d := todo.DueDate.Time
		dueDate = &d
	}
	return dueDate
}

func ConvertFromPointerToNullUUID(uuidCandidate *string) uuid.NullUUID {
	var assignedTo uuid.NullUUID
	if uuidCandidate != nil {
		u, _ := uuid.FromString(*uuidCandidate)
		assignedTo = uuid.NullUUID{
			UUID:  u,
			Valid: true,
		}
	}
	return assignedTo
}

func ConvertFromNullUuidToStringPtr(uuidCandidate uuid.NullUUID) *string {
	if !uuidCandidate.Valid {
		return nil
	}

	str := uuidCandidate.UUID.String()
	return &str
}

func ConvertFromPointerToSQLNullTime(t *time.Time) sql.NullTime {
	var dueDate sql.NullTime
	if t != nil {
		currTime := *t
		dueDate = sql.NullTime{
			Time:  currTime,
			Valid: true,
		}
	}
	return dueDate
}

func ConvertFromStringToUUID(uuidCandidate string) uuid.UUID {
	res := uuid.FromStringOrNil(uuidCandidate)
	return res
}

func MapPostgresListErrorToError(err error, entityList *entities.List) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return application_errors.NewAlreadyExistError(constants.LIST_TARGET, entityList.Name)
		case "23503":
			return application_errors.NewNotFoundError(constants.USER_TARGET, entityList.Owner.String())
		}
	}
	return err
}

func MapPostgresTodoError(err error, todo *entities.Todo) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return err
	}

	switch pqErr.Code {
	case "23505":
		return application_errors.NewNotFoundError(constants.TODO_TARGET, todo.Id.String())
	case "23503":
		switch pqErr.Constraint {
		case "todos_list_id_fkey":
			return application_errors.NewNotFoundError(constants.LIST_TARGET, todo.ListId.String())
		case "todos_assigned_to_fkey":
			if todo.AssignedTo.Valid {
				return application_errors.NewNotFoundError(constants.USER_TARGET, todo.AssignedTo.UUID.String())
			}
		}
	}
	return err
}

func AssignValueToStringPtr(ptr *string, uuid uuid.UUID) *string {
	str := uuid.String()
	ptr = &str
	return ptr
}

func GetValueFromContext[T any](ctx context.Context, valueKey interface{}) (T, error) {
	var receivedValue T
	receivedValue, ok := ctx.Value(valueKey).(T)
	if !ok {
		return receivedValue, fmt.Errorf("invalid value in context")
	}
	return receivedValue, nil
}

func GetContentFromUrl(r *http.Request, fieldName string) string {
	field := r.URL.Query().Get(fieldName)
	return field
}

func GetLimitFromUrl(r *http.Request) string {
	limit := GetContentFromUrl(r, constants.LIMIT)
	if len(limit) == 0 {
		return constants.DEFAULT_LIMIT_VALUE
	}
	return limit
}

type fieldValidator interface {
	Struct(st interface{}) error
}

func CheckForValidationError(val fieldValidator, s interface{}) (string, error) {
	if err := val.Struct(s); err != nil {
		for _, fe := range err.(validator.ValidationErrors) {
			return fe.Field(), err
		}
	}
	return "", nil
}

func CheckForNotFoundError(err error) bool {
	var nfErr *application_errors.NotFoundError
	if errors.As(err, &nfErr) {
		return true
	}
	return false
}

func getOrganizationNames(organizations []*githubModels.Organization) []string {
	orgNames := make([]string, 0, len(organizations))
	for _, org := range organizations {
		orgNames = append(orgNames, org.Login)
		log.Printf("name: %s", org)
	}
	return orgNames
}

func containsSubstring(slice []string, substr string) bool {
	return slices.ContainsFunc(slice, func(s string) bool {
		return strings.Contains(s, substr)
	})
}

func DetermineRole(organizations []*githubModels.Organization) (string, error) {
	orgNames := getOrganizationNames(organizations)

	if containsSubstring(orgNames, constants.ADMIN_ORG) {
		return string(constants.Admin), nil
	}

	if containsSubstring(orgNames, constants.WRITER_ORG) {
		return string(constants.Writer), nil
	}

	if containsSubstring(orgNames, constants.READER_ORG) {
		return string(constants.Reader), nil
	}

	return "", fmt.Errorf("invalid role")
}

func DetermineCorrectJwtErrorMessage(err error) string {
	if errors.Is(err, jwt.ErrSignatureInvalid) {
		return "invalid token signature"
	}

	if errors.Is(err, jwt.ErrTokenExpired) {
		return "token expired"
	}

	return "invalid token"

}
