package router

import (
	"github.com/QuangTung97/weblib/hx"
	"github.com/QuangTung97/weblib/urls"
)

func HttpGet[T any](
	router *Router,
	urlPath urls.Path[T],
	handler func(ctx Context, params T) (hx.Elem, error),
) {
}
