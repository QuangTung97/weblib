package router

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func (h *htmlTest) doPost(postURL string, data url.Values) {
	var body bytes.Buffer
	body.WriteString(data.Encode())

	h.req = httptest.NewRequest(http.MethodPost, postURL, &body)
	h.req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	h.writer = httptest.NewRecorder()
	h.router.GetChi().ServeHTTP(h.writer, h.req)
}

type htmlParams struct {
	ID     int    `json:"id"`
	Search string `json:"search"`
	Age    int64  `json:"age"`
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

	// check headers
	assert.Equal(t, http.Header{
		"Content-Type": []string{"text/html; charset=utf-8"},
	}, h.writer.Header())

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

func TestHtmlGet__Can_Not_Parse_Path_Param(t *testing.T) {
	h := newHtmlTest()

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

	h.doGet("/users/invalid")

	// check input
	assert.Equal(t, []htmlParams(nil), inputParams)

	// check output
	assert.Equal(t, 400, h.writer.Code)
	assert.Equal(t,
		`{"error":"can not set value 'invalid' to field 'id' with type 'int'"}`+"\n",
		h.writer.Body.String(),
	)

	// check headers
	assert.Equal(t, http.Header{
		"Content-Type": []string{"application/json"},
	}, h.writer.Header())

	// check actions
	assert.Equal(t, []string(nil), h.actions)
}

func TestHtmlGet__Can_Not_Parse_Query_Param(t *testing.T) {
	h := newHtmlTest()

	urlPath := urls.New[htmlParams]("/users/{id}")

	var inputParams []htmlParams
	HtmlGet(h.router, urlPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
		h.addAction("handler")
		inputParams = append(inputParams, params)
		return hx.Div(
			hx.Text("Hello World"),
		), nil
	})

	h.doGet("/users/123?age=invalid02")

	// check input
	assert.Equal(t, []htmlParams(nil), inputParams)

	// check output
	assert.Equal(t, 400, h.writer.Code)
	assert.Equal(t,
		`{"error":"can not set value 'invalid02' to field 'age' with type 'int64'"}`+"\n",
		h.writer.Body.String(),
	)
}

func TestHtmlGet__Handler_Error(t *testing.T) {
	h := newHtmlTest()

	urlPath := urls.New[htmlParams]("/users/{id}")

	var inputParams []htmlParams
	HtmlGet(h.router, urlPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
		h.addAction("handler")
		inputParams = append(inputParams, params)
		return hx.None(), errors.New("handler error")
	})

	h.doGet("/users/123?age=81")

	// check input
	assert.Equal(t, []htmlParams{
		{ID: 123, Age: 81},
	}, inputParams)

	// check output
	assert.Equal(t, 400, h.writer.Code)
	assert.Equal(t,
		`{"error":"handler error"}`+"\n",
		h.writer.Body.String(),
	)
}

func TestHtmlPost_Normal(t *testing.T) {
	h := newHtmlTest()

	urlPath := urls.New[htmlParams]("/users/{id}")

	var inputParams []htmlParams
	HtmlPost(h.router, urlPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
		h.addAction("handler")
		inputParams = append(inputParams, params)
		return hx.None(), nil
	})

	h.doPost("/users/123", url.Values{
		"age":     {"81"},
		"search":  {"hello02"},
		"another": {"invalid"},
	})

	// check input
	assert.Equal(t, []htmlParams{
		{ID: 123, Age: 81, Search: "hello02"},
	}, inputParams)

	// check output
	assert.Equal(t, 200, h.writer.Code)
	assert.Equal(t, "", h.writer.Body.String())
}
