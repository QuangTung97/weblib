package csrf

import (
	"bytes"
	"errors"
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
	csrfToken null.Null[string]

	core   *Core
	ctx    router.Context
	writer *httptest.ResponseRecorder

	middleware router.Middleware

	actions []string
}

func newMiddlewareTest(
	method string,
) *middlewareTest {
	m := &middlewareTest{}
	m.actions = []string{}

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
		func(ctx router.Context) null.Null[string] {
			return m.csrfToken
		},
	)

	return m
}

func (m *middlewareTest) addAction(a string) {
	m.actions = append(m.actions, a)
}

func TestMiddleware__HandleGet(t *testing.T) {
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

	t.Run("already had csrf_token, do nothing", func(t *testing.T) {
		m := newMiddlewareTest(http.MethodGet)

		m.ctx.Request.AddCookie(&http.Cookie{
			Name:     preSessionCookieName,
			Value:    "random-pre-session",
			HttpOnly: true,
		})
		m.ctx.Request.AddCookie(&http.Cookie{
			Name:  csrfCookieName,
			Value: m.core.Generate("random-pre-session"),
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
		assert.Equal(t, http.Header{}, m.writer.Header())

		assert.Equal(t, []string{"handler"}, m.actions)
	})

	t.Run("with handler error", func(t *testing.T) {
		m := newMiddlewareTest(http.MethodGet)

		handler := m.middleware(func(ctx router.Context, req any) (any, error) {
			m.addAction("handler")
			return nil, errors.New("handle err")
		})

		// do handle
		resp, err := handler(m.ctx, "input")
		assert.Equal(t, errors.New("handle err"), err)
		assert.Equal(t, nil, resp)

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
}

func TestMiddleware__HandlePost(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := newMiddlewareTest(http.MethodPost)

		m.ctx.Request.AddCookie(&http.Cookie{
			Name:     preSessionCookieName,
			Value:    "random-pre-sess-value",
			HttpOnly: true,
		})

		m.csrfToken = null.New(
			m.core.Generate("random-pre-sess-value"),
		)

		handler := m.middleware(func(ctx router.Context, req any) (any, error) {
			m.addAction("handler")
			return "success", nil
		})

		// do handle
		resp, err := handler(m.ctx, "input")
		assert.Equal(t, nil, err)
		assert.Equal(t, "success", resp)

		// check cookie, no new cookie set
		assert.Equal(t, http.Header{}, m.writer.Header())

		// check action
		assert.Equal(t, []string{"handler"}, m.actions)
	})

	t.Run("success, with normal session id", func(t *testing.T) {
		m := newMiddlewareTest(http.MethodPost)

		m.sessionID = null.New("real-session-id")

		m.csrfToken = null.New(
			m.core.Generate("real-session-id"),
		)

		handler := m.middleware(func(ctx router.Context, req any) (any, error) {
			m.addAction("handler")
			return "success", nil
		})

		// do handle
		resp, err := handler(m.ctx, "input")
		assert.Equal(t, nil, err)
		assert.Equal(t, "success", resp)

		// check cookie, no new cookie set
		assert.Equal(t, http.Header{}, m.writer.Header())

		// check action
		assert.Equal(t, []string{"handler"}, m.actions)
	})

	t.Run("no pre session or session id", func(t *testing.T) {
		m := newMiddlewareTest(http.MethodPost)

		m.csrfToken = null.New(
			m.core.Generate("random-pre-sess-value"),
		)

		handler := m.middleware(func(ctx router.Context, req any) (any, error) {
			m.addAction("handler")
			return "success", nil
		})

		// do handle
		resp, err := handler(m.ctx, "input")
		assert.Equal(t, errors.New("not found session id or pre-session id"), err)
		assert.Equal(t, nil, resp)

		// check cookie, no new cookie set
		assert.Equal(t, http.Header{}, m.writer.Header())

		// check action
		assert.Equal(t, []string{}, m.actions)
	})

	t.Run("no csrf token", func(t *testing.T) {
		m := newMiddlewareTest(http.MethodPost)

		m.ctx.Request.AddCookie(&http.Cookie{
			Name:     preSessionCookieName,
			Value:    "random-pre-sess-value",
			HttpOnly: true,
		})

		handler := m.middleware(func(ctx router.Context, req any) (any, error) {
			m.addAction("handler")
			return "success", nil
		})

		// do handle
		resp, err := handler(m.ctx, "input")
		assert.Equal(t, errors.New("not found csrf token"), err)
		assert.Equal(t, nil, resp)

		// check cookie, no new cookie set
		assert.Equal(t, http.Header{}, m.writer.Header())

		// check action
		assert.Equal(t, []string{}, m.actions)
	})

	t.Run("invalid csrf token", func(t *testing.T) {
		m := newMiddlewareTest(http.MethodPost)

		m.ctx.Request.AddCookie(&http.Cookie{
			Name:     preSessionCookieName,
			Value:    "random-pre-sess-value",
			HttpOnly: true,
		})

		m.csrfToken = null.New(
			m.core.Generate("random-pre-sess-invalid"),
		)

		handler := m.middleware(func(ctx router.Context, req any) (any, error) {
			m.addAction("handler")
			return "success", nil
		})

		// do handle
		resp, err := handler(m.ctx, "input")
		assert.Equal(t, &Error{Message: "invalid csrf token"}, err)
		assert.Equal(t, nil, resp)

		// check cookie, no new cookie set
		assert.Equal(t, http.Header{
			"Set-Cookie": {"csrf_token=; Max-Age=0"},
		}, m.writer.Header())

		// check action
		assert.Equal(t, []string{}, m.actions)
	})
}
