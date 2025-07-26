package oauth

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

//go:generate mockery --name=oauthService --exported --output=./mocks --outpkg=mocks --filename=oauth_service.go --with-expecter=true
type oauthService interface {
	LoginUrl(context.Context) (string, string, error)
	ExchangeCodeForToken(context.Context, string) (string, error)
}

//go:generate mockery --name=jwtIssuer --exported --output=./mocks --outpkg=mocks --filename=jwt_issuer.go --with-expecter=true
type jwtIssuer interface {
	GetTokens(context.Context, string) (*models.CallbackResponse, error)
}

//go:generate mockery --name=httpService --exported --output=./mocks --outpkg=mocks --filename=http_service.go --with-expecter=true
type httpService interface {
	SetCookie(w http.ResponseWriter, cookie *http.Cookie)
	Redirect(w http.ResponseWriter, r *http.Request, url string, httpStatusCode int)
}

type handler struct {
	service     oauthService
	issuer      jwtIssuer
	httpService httpService
}

func NewHandler(service oauthService, issuer jwtIssuer, httpService httpService) *handler {
	return &handler{
		service:     service,
		issuer:      issuer,
		httpService: httpService,
	}
}

func (h *handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	url, state, err := h.service.LoginUrl(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to handle login, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.httpService.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	h.httpService.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	authCode := r.URL.Query().Get("code")
	if len(authCode) == 0 {
		log.C(ctx).Error("failed to handle callback, empty auth code in callback url...")
		utils.EncodeError(w, "missing auth code in callback url", http.StatusBadRequest)
		return
	}

	accessToken, err := h.service.ExchangeCodeForToken(ctx, authCode)
	if err != nil {
		log.C(ctx).Errorf("failed to handle callback, error %s when trying to get access token", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokens, err := h.issuer.GetTokens(ctx, accessToken)
	if err != nil {
		log.C(ctx).Errorf("failed to handle callback, error %s when trying to get jwt", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(tokens); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
