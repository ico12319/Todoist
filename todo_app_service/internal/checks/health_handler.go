package checks

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type pingSender interface {
	PingContext(ctx context.Context) error
}

type handler struct {
	sender pingSender
}

func NewHandler(sender pingSender) *handler {
	return &handler{sender: sender}
}

func (h *handler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("handling liveness in health handler")

	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	if err := h.sender.PingContext(ctx); err != nil {
		log.C(ctx).Errorf("failed to ping database, error %s", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)

		status := utils.GetServerStatus(constants.DATABASE_DOWN_STATUS)

		if err = json.NewEncoder(w).Encode(status); err != nil {
			utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	status := utils.GetServerStatus(constants.READY_STATUS)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("handling readiness in health handler")

	w.WriteHeader(http.StatusOK)

	status := utils.GetServerStatus(constants.OK_STATUS)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
