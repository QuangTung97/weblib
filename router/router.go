package router

type Router struct {
}

func NewRouter() *Router {
	return &Router{}
}

// -------------------------------------------------------------------------
// Internal Implementation
// -------------------------------------------------------------------------

type endpointKey struct {
	method  string
	pattern string
}
type routerState struct {
	registered map[endpointKey]struct{}
}
