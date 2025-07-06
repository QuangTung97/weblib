package csrf

import (
	"github.com/QuangTung97/weblib/null"
	"github.com/QuangTung97/weblib/router"
)

func InitMiddleware(hmacSecretKey string) router.Middleware {
	core := InitCore(hmacSecretKey)
	return NewMiddleware(
		core,
		func(ctx router.Context) null.Null[string] {
			cookie, err := ctx.Request.Cookie(sessionIDCookieName)
			if err != nil {
				return null.Null[string]{}
			}
			return null.New(cookie.Value)
		},
		func(ctx router.Context) null.Null[string] {
			value := ctx.Request.Header.Get(csrfHeaderKey)
			if len(value) == 0 {
				return null.Null[string]{}
			}
			return null.New(value)
		},
	)
}
