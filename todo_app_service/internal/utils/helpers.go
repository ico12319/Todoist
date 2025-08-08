package utils

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/gitHub"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	config "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
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
	limit := GetContentFromUrl(r, constants.FIRST)
	if len(limit) == 0 {
		limit = GetContentFromUrl(r, constants.LAST)
	}

	if len(limit) == 0 {
		limit = constants.DEFAULT_LIMIT_VALUE
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

func getOrganizationNames(organizations []*gitHub.Organization) []string {
	orgNames := make([]string, 0, len(organizations))
	for _, org := range organizations {
		orgNames = append(orgNames, org.Login)
	}
	return orgNames
}

func containsSubstring(slice []string, substr string) bool {
	return slices.ContainsFunc(slice, func(s string) bool {
		return strings.Contains(s, substr)
	})
}

func DetermineRole(organizations []*gitHub.Organization) (string, error) {
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

func DetermineErrorWhenSigningJWT(err error) error {
	if errors.Is(err, jwt.ErrInvalidKeyType) {
		return errors.New("invalid key type passed when trying to sign JWT")
	} else if errors.Is(err, jwt.ErrTokenInvalidClaims) {
		return errors.New("nil claims passed when trying to sign JWT")
	}

	return nil
}

func DetermineJWTErrorWhenParsingWithClaims(ctx context.Context, err error) error {
	if errors.Is(err, jwt.ErrTokenMalformed) {
		config.C(ctx).Errorf("failed to parse jwt, error %s empty or malformed token", err.Error())
		return errors.New("empty or malformed token")
	}

	if errors.Is(err, jwt.ErrSignatureInvalid) {
		config.C(ctx).Errorf("failed to parse jwt, error %s invalid signature", err.Error())
		return errors.New("invalid signature")
	}

	if errors.Is(err, jwt.ErrTokenExpired) {
		return errors.New("token expired")
	}

	return nil
}

func EncodeErrorWithCorrectStatusCode(w http.ResponseWriter, err error) {
	var nfErr *application_errors.NotFoundError
	if errors.As(err, &nfErr) {
		EncodeError(w, err.Error(), http.StatusNotFound)
	} else {
		EncodeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetServerStatus(statusValue string) map[string]string {
	return map[string]string{
		"status": statusValue,
	}
}

func OrchestrateGoRoutines(ctx context.Context, chan1 chan ChannelResult[string], chan2 chan ChannelResult[string]) (string, string, error) {
	var result1 string
	var result2 string

	for i := 0; i < constants.GOROUTINES_COUNT; i++ {
		select {
		case chRes := <-chan1:
			if chRes.Err != nil {
				config.C(ctx).Errorf("failed to get jwt, error %s when trying to determine user's github email", chRes.Err.Error())
				return "", "", chRes.Err
			}

			if len(chRes.Result) == 0 {
				config.C(ctx).Errorf("failed to get jwt, empty email...")
				return "", "", errors.New("email can't be empty")
			}

			result1 = chRes.Result

		case chRes := <-chan2:
			if chRes.Err != nil {
				config.C(ctx).Errorf("failed to get jwt, error %s when trying to get user's app role", chRes.Err.Error())
				return "", "", chRes.Err
			}
			if len(chRes.Result) == 0 {
				config.C(ctx).Errorf("failed to get jwt, empty role received...")
				return "", "", errors.New("role can't be determined")
			}

			result2 = chRes.Result

		case <-ctx.Done():
			return "", "", ctx.Err()
		}
	}

	return result1, result2, nil
}

func CheckForNotFoundError(err error) bool {
	var applicationError *application_errors.NotFoundError
	if errors.As(err, &applicationError) {
		return true
	}

	return false
}

func DetermineSortingOrder(f *filters.BaseFilters) string {
	if len(f.Last) != 0 {
		return constants.DESC_ORDER
	}
	return constants.ASC_ORDER
}
