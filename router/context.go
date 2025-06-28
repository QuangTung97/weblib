package router

import (
	"context"
	"net/http"
)

type Context struct {
	Request *http.Request
	writer  http.ResponseWriter
}

func (c Context) Context() context.Context {
	return c.Request.Context()
}

func NewContext(writer http.ResponseWriter, req *http.Request) Context {
	return Context{
		Request: req,
		writer:  writer,
	}
}
