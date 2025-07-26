package refresh

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"encoding/json"
	"net/http"
)

//go:generate mockery --name=jwtIssuer --exported --output=./mocks --outpkg=mocks --filename=jwt_issuer.go --with-expecter=true
type jwtIssuer interface {
	GetRenewedTokens(context.Context, *handler_models.Refresh) (*models.CallbackResponse, error)
}

type handler struct {
	issuer jwtIssuer
}

func NewHandler(issuer jwtIssuer) *handler {
	return &handler{issuer: issuer}
}

func (h *handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var refresh handler_models.Refresh
	if err := json.NewDecoder(r.Body).Decode(&refresh); err != nil {
		log.C(ctx).Errorf("failed to decode refresh handler model %s", err.Error())
		utils.EncodeError(w, constants.INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	renewedTokens, err := h.issuer.GetRenewedTokens(ctx, &refresh)
	if err != nil {
		log.C(ctx).Errorf("failed to refresh tokens, error %s when calling service method", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(renewedTokens); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
