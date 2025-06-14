package status_code_encoders

import (
	"context"
	"errors"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/application_errors"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"net/http"
)

type StatusCodeEncoder interface {
	EncodeErrorWithCorrectStatusCode(ctx context.Context, w http.ResponseWriter)
}
type statusCodeEncoderFactory struct{}

func NewStatusCodeEncoderFactory() *statusCodeEncoderFactory {
	return &statusCodeEncoderFactory{}
}

func (*statusCodeEncoderFactory) CreateStatusCodeEncoder(ctx context.Context, w http.ResponseWriter, err error) StatusCodeEncoder {
	log.C(ctx).Info("creating status code encoder in status code encoder factory")

	var nfErr *application_errors.NotFoundError
	if errors.As(err, &nfErr) {
		return newStatusNotFoundErrorEncoder(err)
	}

	return newStatusInternalServerError(err)
}
