package router

import (
	"bytes"
	"errors"
	"io"
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

func (h *htmlTest) doMethod(method string, postURL string, data url.Values) {
	var body io.Reader
	if data != nil {
		var buf bytes.Buffer
		buf.WriteString(data.Encode())
		body = &buf
	}

	h.req = httptest.NewRequest(method, postURL, body)
	if data != nil {
		h.req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

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

func TestHtmlGet__Handler_Error__WithCustomErrorHandler(t *testing.T) {
	h := newHtmlTest()

	urlPath := urls.New[htmlParams]("/users/{id}")

	var inputParams []htmlParams
	HtmlGet(h.router, urlPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
		h.addAction("handler")
		inputParams = append(inputParams, params)
		return hx.None(), errors.New("handler error")
	})

	h.router.SetCustomHtmlErrorHandler(func(ctx Context, err error) {
		writer := ctx.GetWriter()
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(err.Error()))
	})

	h.doGet("/users/123?age=81")

	// check input
	assert.Equal(t, []htmlParams{
		{ID: 123, Age: 81},
	}, inputParams)

	// check output
	assert.Equal(t, 500, h.writer.Code)
	assert.Equal(t,
		"handler error",
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

	h.doMethod(http.MethodPost, "/users/123", url.Values{
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

func TestHtmlPut__Normal(t *testing.T) {
	h := newHtmlTest()

	urlPath := urls.New[htmlParams]("/users/{id}")

	var inputParams []htmlParams
	HtmlPut(h.router, urlPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
		inputParams = append(inputParams, params)
		return hx.None(), nil
	})

	h.doMethod(http.MethodPut, "/users/123", url.Values{
		"age": {"81"},
	})

	// check input
	assert.Equal(t, []htmlParams{
		{ID: 123, Age: 81},
	}, inputParams)

	// check output
	assert.Equal(t, 200, h.writer.Code)
	assert.Equal(t, "", h.writer.Body.String())
}

func TestHtmlDelete_Normal(t *testing.T) {
	h := newHtmlTest()

	urlPath := urls.New[htmlParams]("/users/{id}")

	var inputParams []htmlParams
	HtmlDelete(h.router, urlPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
		inputParams = append(inputParams, params)
		return hx.None(), nil
	})

	h.doMethod(http.MethodDelete, "/users/123?age=82", nil)

	// check input
	assert.Equal(t, []htmlParams{
		{ID: 123, Age: 82},
	}, inputParams)

	// check output
	assert.Equal(t, 200, h.writer.Code)
	assert.Equal(t, "", h.writer.Body.String())
}

func TestHtmlGet__Multi__Duplicated_Pattern__Panic(t *testing.T) {
	h := newHtmlTest()

	urlPath := urls.New[htmlParams]("/users/{id}")

	HtmlGet(h.router, urlPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
		return hx.None(), nil
	})

	assert.PanicsWithValue(t, "GET /users/{id} is already defined", func() {
		HtmlGet(h.router, urlPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
			return hx.None(), nil
		})
	})
}

func TestHtmlGet__With_Group_Prefix__Not_Match__Panic(t *testing.T) {
	h := newHtmlTest()

	urlPath := urls.New[htmlParams]("/api/users/{id}")
	productPath := urls.New[htmlParams]("/api/products/{id}")

	h.router = h.router.WithGroup("/api")
	h.router = h.router.WithGroup("/users")

	HtmlGet(h.router, urlPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
		return hx.None(), nil
	})

	assert.PanicsWithValue(t, "GET /api/products/{id} not satisfy url prefix '/api/users'", func() {
		HtmlGet(h.router, productPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
			return hx.None(), nil
		})
	})
}

func TestHtmlGet__With_Redirect(t *testing.T) {
	h := newHtmlTest()

	urlPath := urls.New[htmlParams]("/users/{id}")
	HtmlGet(h.router, urlPath, func(ctx Context, params htmlParams) (hx.Elem, error) {
		ctx.HttpRedirect("https://example.com")
		return hx.Div(), nil
	})

	h.doGet("/users/123?search=test01")

	// check output
	assert.Equal(t, http.StatusTemporaryRedirect, h.writer.Code)
	assert.Equal(t, `<a href="https://example.com">Temporary Redirect</a>.`+"\n\n", h.writer.Body.String())

	// check headers
	assert.Equal(t, http.Header{
		"Content-Type": []string{"text/html; charset=utf-8"},
		"Location": {
			"https://example.com",
		},
	}, h.writer.Header())
}

type testParams01 struct {
	ID int64 `json:"id"`
}

type testParamInterface interface {
	GetID() int64
}

func (p testParams01) GetID() int64 {
	return p.ID
}

var _ testParamInterface = testParams01{}

type testParams02 struct {
	ID int64 `json:"id"`
}

func TestHtmlGet__With_Params_Validator(t *testing.T) {
	h := newHtmlTest()

	router := h.router.WithParamValidator(
		func(params any) {
			_, ok := params.(testParamInterface)
			if !ok {
				panic("Missing GetID()")
			}
		},
	)

	// success
	urlPath01 := urls.New[testParams01]("/users/{id}")
	HtmlGet(router, urlPath01, func(ctx Context, params testParams01) (hx.Elem, error) {
		return hx.Div(), nil
	})

	// panic
	urlPath02 := urls.New[testParams02]("/users-v2/{id}")
	assert.PanicsWithValue(t, "Missing GetID()", func() {
		HtmlGet(router, urlPath02, func(ctx Context, params testParams02) (hx.Elem, error) {
			return hx.Div(), nil
		})
	})
}

func TestHtmlGet__Normal__Middleware__Redirect(t *testing.T) {
	h := newHtmlTest()

	h.router = h.router.WithMiddlewares(
		func(handler GenericHandler) GenericHandler {
			return func(ctx Context, req any) (any, error) {
				h.addAction("middleware01")
				ctx.HttpRedirect("/login")
				return nil, nil
			}
		},
	)

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
	assert.Equal(t, []htmlParams(nil), inputParams)

	// check output
	assert.Equal(t, 307, h.writer.Code)
	assert.Equal(t, `<a href="/login">Temporary Redirect</a>.`+"\n\n", h.writer.Body.String())

	// check headers
	assert.Equal(t, http.Header{
		"Content-Type": {"text/html; charset=utf-8"},
		"Location":     {"/login"},
	}, h.writer.Header())

	// check actions
	assert.Equal(t, []string{
		"middleware01",
	}, h.actions)
}
