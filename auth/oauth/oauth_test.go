package oauth

import (
	"context"
	"encoding/base64"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/QuangTung97/weblib/router"
	"github.com/QuangTung97/weblib/sliceutil"
)

type serviceTest struct {
	now        time.Time
	randData   string
	randInputs []int

	exchangeCodes []string
	accessTokens  []string

	svc *serviceImpl

	writer *httptest.ResponseRecorder
}

func newTime(s string) time.Time {
	const format = "2006-01-02 15:04"
	t, err := time.Parse(format, s)
	if err != nil {
		panic(err)
	}
	return t.UTC()
}

func newServiceTest() *serviceTest {
	s := &serviceTest{}
	s.now = newTime("2025-06-28 10:20")
	s.randData = "rand01"

	conf := &oauth2.Config{
		RedirectURL:  "http://localhost:8000/auth/google/callback",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	s.svc = NewService(
		conf,
		func(ctx router.Context, accessToken string) error {
			s.accessTokens = append(s.accessTokens, accessToken)
			return nil
		},
		func() time.Time {
			return s.now
		},
		func(n int) []byte {
			s.randInputs = append(s.randInputs, n)
			return []byte(s.randData)
		},
	).(*serviceImpl)

	s.svc.exchangeFunc = func(ctx context.Context, code string) (string, error) {
		s.exchangeCodes = append(s.exchangeCodes, code)
		return "new-access-token", nil
	}

	s.writer = httptest.NewRecorder()

	return s
}

func TestService_HandleLogin(t *testing.T) {
	t.Run("generate state", func(t *testing.T) {
		s := newServiceTest()

		state := s.svc.generateStateOauthCookie(s.writer, "/user/123")

		// check state
		data, err := base64.URLEncoding.DecodeString(state)
		assert.Equal(t, nil, err)
		assert.Equal(t, `{"login_session":"cmFuZDAx","redirect_url":"/user/123"}`, string(data))

		assert.Equal(t, []int{20}, s.randInputs)

		// check header
		assert.Equal(t, http.Header{
			"Set-Cookie": {
				"oauth_login_sess=cmFuZDAx; Path=/; Expires=Sat, 28 Jun 2025 10:40:00 GMT",
			},
		}, s.writer.Header())
	})

	t.Run("handle login", func(t *testing.T) {
		s := newServiceTest()

		req := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)

		ctx := router.NewContext(s.writer, req)

		// do handle login
		_, err := s.svc.HandleLogin(ctx, LoginParams{Redirect: "/user/123"})
		assert.Equal(t, nil, err)

		// check header
		checkHeader := maps.Clone(s.writer.Header())
		redirectURL := checkHeader.Get("Location")
		delete(checkHeader, "Location")
		assert.Equal(t, http.Header{
			"Content-Type": {
				"text/html; charset=utf-8",
			},
			"Set-Cookie": {
				"oauth_login_sess=cmFuZDAx; Path=/; Expires=Sat, 28 Jun 2025 10:40:00 GMT",
			},
		}, checkHeader)

		// check http status
		assert.Equal(t, http.StatusTemporaryRedirect, s.writer.Code)

		// check redirect url
		urlVal, _ := url.Parse(redirectURL)

		assert.Equal(t, "https", urlVal.Scheme)
		assert.Equal(t, "accounts.google.com", urlVal.Host)
		assert.Equal(t, "/o/oauth2/auth", urlVal.Path)

		// check redirect query params
		outputParams := urlVal.Query()
		assert.Equal(t, []string{
			"client_id",
			"redirect_uri",
			"response_type",
			"scope",
			"state",
		}, sliceutil.GetMapKeys(outputParams))

		// check state query param
		state := outputParams["state"][0]
		stateRaw, _ := base64.URLEncoding.DecodeString(state)
		assert.Equal(t, `{"login_session":"cmFuZDAx","redirect_url":"/user/123"}`, string(stateRaw))

		// check client_id
		assert.Equal(t, "test-client-id", outputParams["client_id"][0])
		assert.Equal(t, "http://localhost:8000/auth/google/callback", outputParams["redirect_uri"][0])
		assert.Equal(t, "code", outputParams["response_type"][0])
		assert.Equal(t, "https://www.googleapis.com/auth/userinfo.email", outputParams["scope"][0])
	})
}

func TestService_HandleCallback(t *testing.T) {
	t.Run("not found cookie", func(t *testing.T) {
		s := newServiceTest()

		req := httptest.NewRequest(http.MethodGet, "/oauth/callback", nil)
		ctx := router.NewContext(s.writer, req)

		_, err := s.svc.HandleCallback(ctx, CallbackParams{})
		assert.Error(t, err)
		assert.Equal(t, "invalid oauth login session: http: named cookie not present", err.Error())
	})

	t.Run("invalid state", func(t *testing.T) {
		s := newServiceTest()

		req := httptest.NewRequest(http.MethodGet, "/oauth/callback", nil)
		req.AddCookie(&http.Cookie{
			Name:  oauthLoginSessionCookie,
			Value: "login-cookie-value",
		})

		ctx := router.NewContext(s.writer, req)

		_, err := s.svc.HandleCallback(ctx, CallbackParams{
			State: "invalid-base64",
		})
		assert.Error(t, err)
		assert.Equal(t, "invalid base64 state: illegal base64 data at input byte 12", err.Error())
	})

	t.Run("success", func(t *testing.T) {
		s := newServiceTest()

		// construct input state
		state := s.svc.generateStateOauthCookie(httptest.NewRecorder(), "/user/123")
		req := httptest.NewRequest(http.MethodGet, "/oauth/callback", nil)
		req.AddCookie(&http.Cookie{
			Name:  oauthLoginSessionCookie,
			Value: "cmFuZDAx",
		})

		ctx := router.NewContext(s.writer, req)

		// do handle
		_, err := s.svc.HandleCallback(ctx, CallbackParams{
			State: state,
			Code:  "input-exchange-code",
		})
		assert.Equal(t, nil, err)

		// check exchange code & access token
		assert.Equal(t, []string{
			"input-exchange-code",
		}, s.exchangeCodes)
		assert.Equal(t, []string{
			"new-access-token",
		}, s.accessTokens)

		// check header
		assert.Equal(t, http.Header{
			"Content-Type": {"text/html; charset=utf-8"},
			"Location":     {"/user/123"},
		}, s.writer.Header())
	})

	t.Run("mismatch cookie", func(t *testing.T) {
		s := newServiceTest()

		// construct input state
		state := s.svc.generateStateOauthCookie(httptest.NewRecorder(), "/user/123")
		req := httptest.NewRequest(http.MethodGet, "/oauth/callback", nil)
		req.AddCookie(&http.Cookie{
			Name:  oauthLoginSessionCookie,
			Value: "cmFuZDAx-invalid",
		})

		ctx := router.NewContext(s.writer, req)

		// do handle
		_, err := s.svc.HandleCallback(ctx, CallbackParams{
			State: state,
		})
		assert.Equal(t, errors.New("mismatch oauth callback state and login session"), err)
	})
}
