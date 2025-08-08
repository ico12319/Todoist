package oauth

import (
	"Todo-List/internProject/todo_app_service/internal/persistence"
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"fmt"
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
	transact    persistence.Transactioner
	frontendUrl string
}

func NewHandler(service oauthService, issuer jwtIssuer, httpService httpService, transact persistence.Transactioner, frontendUrl string) *handler {
	return &handler{
		service:     service,
		issuer:      issuer,
		httpService: httpService,
		frontendUrl: frontendUrl,
		transact:    transact,
	}
}

func (h *handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in oaut handler when handling callback, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in oaut handler when handling callback, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	h.httpService.SetCookie(w, &http.Cookie{
		Name:   "access-token",
		Value:  tokens.JwtToken,
		Path:   "/",
		MaxAge: constants.HTTP_COOKIES_MAX_AGE,
	})

	h.httpService.SetCookie(w, &http.Cookie{
		Name:   "refresh-token",
		Value:  tokens.RefreshToken,
		Path:   "/",
		MaxAge: constants.HTTP_COOKIES_MAX_AGE,
	})

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction in oauth handler, error %s when trying to handle callback", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("%s/index.html", h.frontendUrl)
	h.httpService.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("handling logout in oauth handler")

	h.httpService.SetCookie(w, &http.Cookie{
		Name:  "access-token",
		Value: "",
		Path:  "/",
	})

	h.httpService.SetCookie(w, &http.Cookie{
		Name:  "refresh-token",
		Value: "",
		Path:  "/",
	})

	url := fmt.Sprintf("%s/index.html#/login", h.frontendUrl)
	h.httpService.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
