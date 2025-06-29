package oauth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/QuangTung97/weblib/hx"
	"github.com/QuangTung97/weblib/router"
)

type Service interface {
	HandleLogin(writer http.ResponseWriter, request *http.Request)
	HandleCallback(ctx router.Context, params CallbackParams) (hx.Elem, error)

	RedirectToLogin(ctx router.Context, backURL string)
}

type CallbackParams struct {
	State string `json:"state"`
	Code  string `json:"code"`
}

type serviceImpl struct {
	authConfig *oauth2.Config
	nowFunc    func() time.Time
	randFunc   func(n int) []byte
}

func NewService(
	authConfig *oauth2.Config,
	nowFunc func() time.Time,
	randFunc func(n int) []byte,
) Service {
	return &serviceImpl{
		authConfig: authConfig,
		nowFunc:    nowFunc,
		randFunc:   randFunc,
	}
}

func (s *serviceImpl) HandleLogin(writer http.ResponseWriter, request *http.Request) {
	state := s.generateStateOauthCookie(writer, request.URL.Query().Get("redirect"))

	redirectURL := s.authConfig.AuthCodeURL(state)
	http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
}

func (s *serviceImpl) HandleCallback(ctx router.Context, params CallbackParams) (hx.Elem, error) {
	sessCookie, err := ctx.Request.Cookie(oauthLoginSessionCookie)
	if err != nil {
		return hx.None(), fmt.Errorf("invalid oauth login session: %w", err)
	}

	stateData, err := base64.URLEncoding.DecodeString(params.State)
	if err != nil {
		return hx.None(), fmt.Errorf("invalid base64 state: %w", err)
	}

	var state oauthState
	if err := json.Unmarshal(stateData, &state); err != nil {
		return hx.None(), fmt.Errorf("invalid json state: %w", err)
	}

	if sessCookie.Value != state.LoginSession {
		return hx.None(), fmt.Errorf("mismatch oauth callback state and login session")
	}

	ctx.HttpRedirect(state.RedirectURL)
	return hx.None(), nil
}

const oauthLoginSessionCookie = "oauth_login_sess"

type oauthState struct {
	LoginSession string `json:"login_session"`
	RedirectURL  string `json:"redirect_url"`
}

func (s *serviceImpl) generateStateOauthCookie(w http.ResponseWriter, redirectURL string) string {
	var expiration = s.nowFunc().Add(20 * time.Minute)

	data := s.randFunc(20)
	sessStr := base64.URLEncoding.EncodeToString(data)
	cookie := http.Cookie{
		Name:    oauthLoginSessionCookie,
		Value:   sessStr,
		Expires: expiration,
	}
	http.SetCookie(w, &cookie)

	state := oauthState{
		LoginSession: sessStr,
		RedirectURL:  redirectURL,
	}

	data, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(data)
}

func (s *serviceImpl) RedirectToLogin(ctx router.Context, backURL string) {
}
