package oauth

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	config "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type oauthServ interface {
	LoginUrl(context.Context) (string, string, error)
	ExchangeCodeForToken(context.Context, string) (string, error)
	GetTokens(context.Context, string) (*models.CallbackResponse, error)
	GetRenewedTokens(context.Context, *handler_models.Refresh) (*models.CallbackResponse, error)
}

type handler struct {
	service oauthServ
}

func NewHandler(service oauthServ) *handler {
	return &handler{service: service}
}

func (h *handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	url, state, err := h.service.LoginUrl(ctx)
	if err != nil {
		config.C(ctx).Errorf("failed to handle login, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	authCode := r.URL.Query().Get("code")

	accessToken, err := h.service.ExchangeCodeForToken(ctx, authCode)
	if err != nil {
		config.C(ctx).Errorf("failed to handle callback, error %s when trying to get access token", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokens, err := h.service.GetTokens(ctx, accessToken)
	if err != nil {
		config.C(ctx).Errorf("failed to handle callback, error %s when trying to get jwt", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(tokens); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var refresh handler_models.Refresh
	if err := json.NewDecoder(r.Body).Decode(&refresh); err != nil {
		config.C(ctx).Errorf("failed to decode refresh handler model %s", err.Error())
		utils.EncodeError(w, constants.INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	renewedTokens, err := h.service.GetRenewedTokens(ctx, &refresh)
	if err != nil {
		config.C(ctx).Errorf("failed to refresh tokens, error %s when calling service method", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(renewedTokens); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
