package oauth

import (
	"encoding/base64"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/QuangTung97/weblib/sliceutil"
)

type serviceTest struct {
	now        time.Time
	randData   string
	randInputs []int

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

	s.svc = NewService(
		&oauth2.Config{
			RedirectURL:  "http://localhost:8000/auth/google/callback",
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
			Endpoint:     google.Endpoint,
		},
		func() time.Time {
			return s.now
		},
		func(n int) []byte {
			s.randInputs = append(s.randInputs, n)
			return []byte(s.randData)
		},
	).(*serviceImpl)

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
				"oauth_login_sess=cmFuZDAx; Expires=Sat, 28 Jun 2025 10:40:00 GMT",
			},
		}, s.writer.Header())
	})

	t.Run("handle login", func(t *testing.T) {
		s := newServiceTest()

		queryParams := url.Values{
			"redirect": {"/user/123"},
		}
		req := httptest.NewRequest(
			http.MethodGet, "/oauth/login?"+queryParams.Encode(), nil,
		)
		// do handle login
		s.svc.HandleLogin(s.writer, req)

		// check header
		checkHeader := maps.Clone(s.writer.Header())
		redirectURL := checkHeader.Get("Location")
		delete(checkHeader, "Location")
		assert.Equal(t, http.Header{
			"Content-Type": {
				"text/html; charset=utf-8",
			},
			"Set-Cookie": {
				"oauth_login_sess=cmFuZDAx; Expires=Sat, 28 Jun 2025 10:40:00 GMT",
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
