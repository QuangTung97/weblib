package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	state       *routerState
	middlewares []Middleware
	urlPrefix   string
}

func NewRouter() *Router {
	router := chi.NewRouter()
	r := &Router{
		state: &routerState{
			chi:        router,
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

func (r *Router) GetChi() *chi.Mux {
	return r.state.chi
}

// -------------------------------------------------------------------------
// Internal Implementation
// -------------------------------------------------------------------------

type endpointKey struct {
	method  string
	pattern string
}
type routerState struct {
	chi        *chi.Mux
	registered map[endpointKey]struct{}

	finalHooks []Middleware

	handleHtmlError func(err error, writer http.ResponseWriter)
}
