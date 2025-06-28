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
