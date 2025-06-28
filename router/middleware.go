package router

type GenericHandler = func(ctx Context, req any) (any, error)

type Middleware func(handler GenericHandler) GenericHandler
