package router

type GenericHandler = func(ctx Context, req any) (any, error)

type Middleware func(handler GenericHandler) GenericHandler

// WithMiddlewares creates a new Router object and append new middlewares to it.
// The old Router is unchanged.
func (r *Router) WithMiddlewares(middlewares ...Middleware) *Router {
	newRouter := *r
	newRouter.middlewares = append(newRouter.middlewares, middlewares...)
	return &newRouter
}

// AddFinalizeHook add a middleware that will be shared across *Router objects.
// And it wll be called after all Normal middlewares, but before the real handler
func (r *Router) AddFinalizeHook(middleware Middleware) {
	r.state.finalHooks = append(r.state.finalHooks, middleware)
}
