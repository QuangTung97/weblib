package crsf

import (
	"encoding/base64"
	"net/http"

	"github.com/QuangTung97/weblib/null"
	"github.com/QuangTung97/weblib/router"
)

func NewMiddleware(
	core *Core,
	getSessionID func(ctx router.Context) null.Null[string],
) router.Middleware {
	m := middlewareLogic{
		core:             core,
		getSessionIDFunc: getSessionID,
	}
	return m.runMiddleware
}

const (
	preSessionCookieName = "pre_session_id"
	csrfCookieName       = "csrf_token"
)

type middlewareLogic struct {
	core             *Core
	getSessionIDFunc func(ctx router.Context) null.Null[string]
}

func (m *middlewareLogic) getSessionID(ctx router.Context) (string, func()) {
	sessionID := m.getSessionIDFunc(ctx)
	if sessionID.Valid {
		return sessionID.Data, func() {}
	}

	preSessCookie, err := ctx.Request.Cookie(preSessionCookieName)
	if err == nil {
		return preSessCookie.Value, func() {}
	}

	preSessionID := base64.URLEncoding.EncodeToString(m.core.randFunc(20))
	updateFn := func() {
		http.SetCookie(ctx.GetWriter(), &http.Cookie{
			Name:     preSessionCookieName,
			Value:    preSessionID,
			HttpOnly: true,
			MaxAge:   7 * 24 * 3600, // 7 days
		})
	}

	return preSessionID, updateFn
}

func (m *middlewareLogic) handleGet(
	ctx router.Context, handler router.GenericHandler, req any,
) (any, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}

	sessionID, updateFn := m.getSessionID(ctx)
	updateFn()

	csrfToken := m.core.Generate(sessionID)
	http.SetCookie(ctx.GetWriter(), &http.Cookie{
		Name:  csrfCookieName,
		Value: csrfToken,
	})

	return resp, nil
}

func (m *middlewareLogic) runMiddleware(
	handler router.GenericHandler,
) router.GenericHandler {
	return func(ctx router.Context, req any) (any, error) {
		if ctx.Request.Method == http.MethodGet {
			return m.handleGet(ctx, handler, req)
		}

		return handler(ctx, req)
	}
}
