package crsf

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/QuangTung97/weblib/null"
	"github.com/QuangTung97/weblib/router"
)

type middlewareTest struct {
	sessionID null.Null[string]
	core      *Core
	ctx       router.Context
	writer    *httptest.ResponseRecorder

	middleware router.Middleware

	actions []string
}

func newMiddlewareTest(
	method string,
) *middlewareTest {
	m := &middlewareTest{}

	m.core = NewCore(
		"secret-key-01",
		func(n int) []byte {
			return bytes.Repeat([]byte("*"), n)
		},
	)

	req := httptest.NewRequest(method, "/", nil)
	m.writer = httptest.NewRecorder()
	m.ctx = router.NewContext(m.writer, req)

	m.middleware = NewMiddleware(
		m.core,
		func(ctx router.Context) null.Null[string] {
			return m.sessionID
		},
	)

	return m
}

func (m *middlewareTest) addAction(a string) {
	m.actions = append(m.actions, a)
}

func TestMiddleware(t *testing.T) {
	t.Run("get without session id and pre-session", func(t *testing.T) {
		m := newMiddlewareTest(http.MethodGet)

		handler := m.middleware(func(ctx router.Context, req any) (any, error) {
			m.addAction("handler")
			return "success", nil
		})

		// do handle
		resp, err := handler(m.ctx, "input")
		assert.Equal(t, nil, err)
		assert.Equal(t, "success", resp)

		// check cookie
		assert.Equal(t, http.Header{
			"Set-Cookie": {
				"pre_session_id=KioqKioqKioqKioqKioqKioqKio=; Max-Age=604800; HttpOnly",
				fmt.Sprintf(
					"csrf_token=%s",
					m.core.Generate("KioqKioqKioqKioqKioqKioqKio="),
				),
			},
		}, m.writer.Header())
	})

	t.Run("with pre session", func(t *testing.T) {
		m := newMiddlewareTest(http.MethodGet)

		m.ctx.Request.AddCookie(&http.Cookie{
			Name:     preSessionCookieName,
			Value:    "random-pre-session",
			HttpOnly: true,
		})

		handler := m.middleware(func(ctx router.Context, req any) (any, error) {
			m.addAction("handler")
			return "success", nil
		})

		// do handle
		resp, err := handler(m.ctx, "input")
		assert.Equal(t, nil, err)
		assert.Equal(t, "success", resp)

		// check cookie
		assert.Equal(t, http.Header{
			"Set-Cookie": {
				fmt.Sprintf(
					"csrf_token=%s",
					m.core.Generate("random-pre-session"),
				),
			},
		}, m.writer.Header())

		assert.Equal(t, []string{"handler"}, m.actions)
	})
}
