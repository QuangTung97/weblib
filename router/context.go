package router

import (
	"context"
	"net/http"
)

type Context struct {
	Request *http.Request
	writer  http.ResponseWriter

	state *contextState
}

type contextState struct {
	responded bool
}

func NewContext(writer http.ResponseWriter, req *http.Request) Context {
	return Context{
		Request: req,
		writer:  writer,

		state: &contextState{},
	}
}

func (c Context) Context() context.Context {
	return c.Request.Context()
}

func (c Context) GetWriter() http.ResponseWriter {
	return c.writer
}

func (c Context) HttpRedirect(redirectURL string) {
	c.state.responded = true
	http.Redirect(c.writer, c.Request, redirectURL, http.StatusTemporaryRedirect)
}

func (c Context) IsHxRequest() bool {
	return c.Request.Header.Get("Hx-Request") == "true"
}
