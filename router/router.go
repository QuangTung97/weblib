package router

import (
	"context"
	"net/http"
)

type Router struct {
}

type Context struct {
	Request *http.Request
	writer  http.ResponseWriter
}

func (c Context) Context() context.Context {
	return c.Request.Context()
}

type GenericHandler = func(ctx Context, req any) (any, error)

type Middleware func(handler GenericHandler) GenericHandler
