package refresh

import (
	"Todo-List/internProject/todo_app_service/internal/persistence"
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
	GetRenewedTokens(ctx context.Context, refresh *handler_models.Refresh) (*models.CallbackResponse, error)
}

type httpService interface {
	SetCookie(w http.ResponseWriter, cookie *http.Cookie)
}

type Handler struct {
	issuer   jwtIssuer
	service  httpService
	transact persistence.Transactioner
}

func NewHandler(issuer jwtIssuer, service httpService, transact persistence.Transactioner) *Handler {
	return &Handler{
		issuer:   issuer,
		service:  service,
		transact: transact,
	}
}

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in refresh handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var refresh handler_models.Refresh
	if err = json.NewDecoder(r.Body).Decode(&refresh); err != nil {
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

	h.service.SetCookie(w, &http.Cookie{
		Name:   "access-token",
		Value:  renewedTokens.JwtToken,
		Path:   "/",
		MaxAge: constants.HTTP_COOKIES_MAX_AGE,
	})

	h.service.SetCookie(w, &http.Cookie{
		Name:   "refresh-token",
		Value:  renewedTokens.RefreshToken,
		Path:   "/",
		MaxAge: constants.HTTP_COOKIES_MAX_AGE,
	})

	if err = json.NewEncoder(w).Encode(renewedTokens); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction in refresh handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
