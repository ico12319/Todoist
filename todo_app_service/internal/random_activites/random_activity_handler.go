package random_activites

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

//go:generate mockery --name=randomActivityService --exported --output=./mocks --outpkg=mocks --filename=randomActivity_service.go --with-expecter=true
type randomActivityService interface {
	Suggest(ctx context.Context) (*models.RandomActivity, error)
}

type handler struct {
	service randomActivityService
}

func NewHandler(service randomActivityService) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) HandleSuggestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	randomActivity, err := h.service.Suggest(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to suggest random activity, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(randomActivity); err != nil {
		log.C(ctx).Errorf("failed to encode random activity, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
