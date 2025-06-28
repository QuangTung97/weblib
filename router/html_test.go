package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/QuangTung97/weblib/hx"
	"github.com/QuangTung97/weblib/urls"
)

type htmlTest struct {
	router *Router

	actions []string

	req    *http.Request
	writer *httptest.ResponseRecorder
}

func newHtmlTest() *htmlTest {
	h := &htmlTest{}
	h.router = NewRouter()
	return h
}

func (h *htmlTest) addAction(action string) {
	h.actions = append(h.actions, action)
}

func (h *htmlTest) addHooks() {
	h.router.AddFinalizeHook(
		func(handler GenericHandler) GenericHandler {
			return func(ctx Context, req any) (any, error) {
				h.addAction("final-hook-01")
				return handler(ctx, req)
			}
		},
	)
	h.router.AddFinalizeHook(
		func(handler GenericHandler) GenericHandler {
			return func(ctx Context, req any) (any, error) {
				h.addAction("final-hook-02")
				return handler(ctx, req)
			}
		},
	)
}

func (h *htmlTest) addMiddlewares() {
	h.router = h.router.WithMiddlewares(
		func(handler GenericHandler) GenericHandler {
			return func(ctx Context, req any) (any, error) {
				h.addAction("middleware01")
				defer h.addAction("middleware01_end")
				return handler(ctx, req)
			}
		},
		func(handler GenericHandler) GenericHandler {
			return func(ctx Context, req any) (any, error) {
				h.addAction("middleware02")
				defer h.addAction("middleware02_end")
				return handler(ctx, req)
			}
		},
	)
}

func (h *htmlTest) doGet(getURL string) {
	h.req = httptest.NewRequest(http.MethodGet, getURL, nil)
	h.writer = httptest.NewRecorder()
	h.router.GetChi().ServeHTTP(h.writer, h.req)
}

type htmlParams struct {
	ID     int    `json:"id"`
	Search string `json:"search"`
}

func TestHtmlGet__Normal(t *testing.T) {
	h := newHtmlTest()

	h.addHooks()
	h.addMiddlewares()

	urlPath := urls.New[htmlParams]("/users/{id}")

	var inputParams []htmlParams
	HtmlGet(h.router, urlPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
		h.addAction("handler")
		inputParams = append(inputParams, params)
		return hx.Div(
			hx.Text("Hello World"),
		), nil
	})

	h.doGet("/users/123?search=test01")

	// check input
	assert.Equal(t, []htmlParams{
		{ID: 123, Search: "test01"},
	}, inputParams)

	// check output
	assert.Equal(t, 200, h.writer.Code)
	assert.Equal(t, `<div>Hello World</div>`, h.writer.Body.String())

	// check actions
	assert.Equal(t, []string{
		"middleware01",
		"middleware02",
		"final-hook-01",
		"final-hook-02",
		"handler",
		"middleware02_end",
		"middleware01_end",
	}, h.actions)
}
