package router

import (
	"github.com/go-chi/chi/v5"
)

type Router struct {
	state       *routerState
	middlewares []Middleware
}

func NewRouter() *Router {
	router := chi.NewRouter()
	return &Router{
		state: &routerState{
			chi:        router,
			registered: map[endpointKey]struct{}{},
		},
	}
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
}
