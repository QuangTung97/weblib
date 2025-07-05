package crsf

import (
	"net/http"

	"github.com/QuangTung97/weblib/null"
	"github.com/QuangTung97/weblib/router"
)

func NewMiddleware(
	secretKey string,
	getSessionID func(ctx router.Context) null.Null[string],
) router.Middleware {
	core := InitCore(secretKey)
	return func(handler router.GenericHandler) router.GenericHandler {
		return handleCsrfToken(core, getSessionID, handler)
	}
}

const (
	preLoginSessionCookieName = "pre_login_session"
	csrfCookieName            = "csrf_token"
)

func handleCsrfToken(
	core *Core,
	getSessionID func(ctx router.Context) null.Null[string],
	handler router.GenericHandler,
) router.GenericHandler {
	return func(ctx router.Context, req any) (any, error) {
		if ctx.Request.Method == http.MethodGet {
			resp, err := handler(ctx, req)
			if err != nil {
				return nil, err
			}

			return resp, nil
		}

		return handler(ctx, req)
	}
}
