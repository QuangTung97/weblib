package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	state           *routerState
	middlewares     []Middleware
	urlPrefix       string
	paramValidators []func(params any)
}

func NewRouter() *Router {
	return NewRouterWithChi(
		chi.NewRouter(),
	)
}

func NewRouterWithChi(chiRouter chi.Router) *Router {
	r := &Router{
		state: &routerState{
			chi:        chiRouter,
			registered: map[endpointKey]struct{}{},
		},
	}

	r.state.handleHtmlError = r.DefaultHtmlErrorHandler

	return r
}

func (r *Router) WithGroup(groupPrefix string) *Router {
	newRouter := *r
	newRouter.urlPrefix += groupPrefix
	return &newRouter
}

func (r *Router) GetChi() chi.Router {
	return r.state.chi
}

func (r *Router) WithParamValidator(
	validators ...func(params any),
) *Router {
	newRouter := *r
	newRouter.paramValidators = append(newRouter.paramValidators, validators...)
	return &newRouter
}

// -------------------------------------------------------------------------
// Internal Implementation
// -------------------------------------------------------------------------

type endpointKey struct {
	method  string
	pattern string
}
type routerState struct {
	chi        chi.Router
	registered map[endpointKey]struct{}

	finalHooks []Middleware

	handleHtmlError func(err error, writer http.ResponseWriter)
}
