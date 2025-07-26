package health

import (
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/gql_errors"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"errors"
	"fmt"
	"net/http"
)

type healthService interface {
	CheckRESTProbes(ctx context.Context, url string) error
}

type handler struct {
	hService healthService
	restUrl  string
}

func NewHandler(hService healthService, restUrl string) *handler {
	return &handler{
		hService: hService,
		restUrl:  restUrl,
	}
}

func (h *handler) HandleCheckingRESTHealthz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("handle checking REST healthz probe in health handler in GQL")

	formattedSuffix := fmt.Sprintf("%s%s", gql_constants.API_ENDPOINT, gql_constants.HEALTH_ENDPOINT)
	url := h.restUrl + formattedSuffix

	h.handleCheckingRESTProbes(w, r, url)
}

func (h *handler) HandleCheckingRestReadyz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("handle checking REST readyz probe in health handler in GQL")

	formattedSuffix := fmt.Sprintf("%s%s", gql_constants.API_ENDPOINT, gql_constants.READY_ENDPOINT)
	url := h.restUrl + formattedSuffix

	h.handleCheckingRESTProbes(w, r, url)
}

func (h *handler) handleCheckingRESTProbes(w http.ResponseWriter, r *http.Request, url string) {
	ctx := r.Context()
	log.C(ctx).Info("handle checking REST probes in health handler in GQL")

	if err := h.hService.CheckRESTProbes(ctx, url); err != nil {
		var restHealthError *gql_errors.RestHealthError
		if errors.As(err, &restHealthError) {
			w.WriteHeader(restHealthError.BadStatusCode)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
