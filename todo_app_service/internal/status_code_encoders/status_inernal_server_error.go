package status_code_encoders

import (
	"context"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/utils"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"net/http"
)

type statusInternalServerError struct {
	errorToEncode error
}

func newStatusInternalServerError(errorToEncode error) *statusInternalServerError {
	return &statusInternalServerError{errorToEncode: errorToEncode}
}

func (s *statusInternalServerError) EncodeErrorWithCorrectStatusCode(ctx context.Context, w http.ResponseWriter) {
	log.C(ctx).Info("encoding http status not found in status internal server error encoder")

	utils.EncodeError(w, s.errorToEncode.Error(), http.StatusInternalServerError)
}
