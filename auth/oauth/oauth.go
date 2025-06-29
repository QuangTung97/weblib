package oauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

type Service interface {
	HandleLogin(writer http.ResponseWriter, request *http.Request)
	HandleCallback(writer http.ResponseWriter, request *http.Request)
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

func (s *serviceImpl) HandleCallback(writer http.ResponseWriter, request *http.Request) {
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
