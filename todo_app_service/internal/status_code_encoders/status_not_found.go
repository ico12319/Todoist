package status_code_encoders

import (
	"context"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/utils"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"net/http"
)

type statusNotFoundErrorEncoder struct {
	errorToEncode error
}

func newStatusNotFoundErrorEncoder(errorToEncode error) *statusNotFoundErrorEncoder {
	return &statusNotFoundErrorEncoder{errorToEncode: errorToEncode}
}

func (s *statusNotFoundErrorEncoder) EncodeErrorWithCorrectStatusCode(ctx context.Context, w http.ResponseWriter) {
	log.C(ctx).Info("encoding http status not found in status not found error encoder")

	utils.EncodeError(w, s.errorToEncode.Error(), http.StatusNotFound)
}
