package crsf

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/QuangTung97/weblib/null"
	"github.com/QuangTung97/weblib/router"
)

func NewMiddleware(
	core *Core,
	getSessionID func(ctx router.Context) null.Null[string],
	getCsrfToken func(ctx router.Context) null.Null[string],
) router.Middleware {
	m := middlewareLogic{
		core:             core,
		getSessionIDFunc: getSessionID,
		getCsrfTokenFunc: getCsrfToken,
	}
	return m.runMiddleware
}

const (
	preSessionCookieName = "pre_session_id"
	csrfCookieName       = "csrf_token"
)

type middlewareLogic struct {
	core *Core

	getSessionIDFunc func(ctx router.Context) null.Null[string]
	getCsrfTokenFunc func(ctx router.Context) null.Null[string]
}

func (m *middlewareLogic) getSessionIDOrPreSession(ctx router.Context) null.Null[string] {
	sessionID := m.getSessionIDFunc(ctx)
	if sessionID.Valid {
		return sessionID
	}

	preSessCookie, err := ctx.Request.Cookie(preSessionCookieName)
	if err == nil {
		return null.New(preSessCookie.Value)
	}

	return null.Null[string]{}
}

func (m *middlewareLogic) getSessionIDOrGenNew(ctx router.Context) (string, func()) {
	sessionID := m.getSessionIDOrPreSession(ctx)
	if sessionID.Valid {
		return sessionID.Data, func() {}
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

	sessionID, updateFn := m.getSessionIDOrGenNew(ctx)
	updateFn()

	_, err = ctx.Request.Cookie(csrfCookieName)
	if err != nil {
		csrfToken := m.core.Generate(sessionID)
		http.SetCookie(ctx.GetWriter(), &http.Cookie{
			Name:  csrfCookieName,
			Value: csrfToken,
		})
	}

	return resp, nil
}

func (m *middlewareLogic) handleNonGet(
	ctx router.Context, handler router.GenericHandler, req any,
) (any, error) {
	sessionID := m.getSessionIDOrPreSession(ctx)
	if !sessionID.Valid {
		return nil, fmt.Errorf("not found session id or pre-session id")
	}

	csrfToken := m.getCsrfTokenFunc(ctx)
	if !csrfToken.Valid {
		return nil, fmt.Errorf("not found csrf token")
	}

	if err := m.core.Validate(sessionID.Data, csrfToken.Data); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func (m *middlewareLogic) runMiddleware(
	handler router.GenericHandler,
) router.GenericHandler {
	return func(ctx router.Context, req any) (any, error) {
		if ctx.Request.Method == http.MethodGet {
			return m.handleGet(ctx, handler, req)
		}

		return m.handleNonGet(ctx, handler, req)
	}
}
