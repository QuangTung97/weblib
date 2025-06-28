package router

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"

	"github.com/QuangTung97/weblib/hx"
	"github.com/QuangTung97/weblib/urls"
)

// HtmlGet set up Get handler
func HtmlGet[T any](
	router *Router,
	urlPath urls.Path[T],
	handler func(ctx Context, params T) (hx.Elem, error),
) {
	htmlMethod(router, http.MethodGet, urlPath, handler)
}

// HtmlPost set up Post handler
func HtmlPost[T any](
	router *Router,
	urlPath urls.Path[T],
	handler func(ctx Context, params T) (hx.Elem, error),
) {
	htmlMethod(router, http.MethodPost, urlPath, handler)
}

// HtmlPut set up Put handler
func HtmlPut[T any](
	router *Router,
	urlPath urls.Path[T],
	handler func(ctx Context, params T) (hx.Elem, error),
) {
	htmlMethod(router, http.MethodPut, urlPath, handler)
}

// HtmlDelete set up Delete handler
func HtmlDelete[T any](
	router *Router,
	urlPath urls.Path[T],
	handler func(ctx Context, params T) (hx.Elem, error),
) {
	htmlMethod(router, http.MethodDelete, urlPath, handler)
}

func htmlMethod[T any](
	router *Router,
	method string,
	urlPath urls.Path[T],
	handler func(ctx Context, params T) (hx.Elem, error),
) {
	// check duplicate endpoint
	key := endpointKey{
		method:  method,
		pattern: urlPath.GetPattern(),
	}
	_, existed := router.state.registered[key]
	if existed {
		panic(fmt.Sprintf("%s %s is already defined", method, urlPath.GetPattern()))
	}
	router.state.registered[key] = struct{}{}

	// setup middlewares
	genericHandler := func(ctx Context, req any) (any, error) {
		resp, err := handler(ctx, req.(T))
		return resp, err
	}
	genericHandler = router.applyMiddlewares(genericHandler)

	stdHandlerError := func(writer http.ResponseWriter, req *http.Request) error {
		var params T
		err := urls.SetStructWithValues(&params, urlPath.GetPathParams(), func(name string) string {
			return chi.URLParam(req, name)
		})
		if err != nil {
			return &HtmlError{
				Reason:  ReasonBadPathParam,
				Message: err.Error(),
			}
		}

		err = urls.SetStructWithValues(&params, urlPath.GetNonPathParams(), func(name string) string {
			return req.FormValue(name)
		})
		if err != nil {
			return &HtmlError{
				Reason:  ReasonBadFormParam,
				Message: err.Error(),
			}
		}

		// call handler
		ctx := NewContext(writer, req)
		resp, err := genericHandler(ctx, params)
		if err != nil {
			return err
		}

		outputElem, ok := resp.(hx.Elem)
		if !ok {
			err := fmt.Errorf("failed to convert response to hx.Elem")
			return &HtmlError{
				Reason:  ReasonBadResponseType,
				Message: err.Error(),
			}
		}

		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = outputElem.Render(writer)
		return nil
	}

	router.state.chi.MethodFunc(method, urlPath.GetPattern(), func(writer http.ResponseWriter, req *http.Request) {
		if err := stdHandlerError(writer, req); err != nil {
			router.state.handleHtmlError(err, writer)
		}
	})
}

func (r *Router) applyMiddlewares(handler GenericHandler) GenericHandler {
	// setup final hooks
	for _, hook := range slices.Backward(r.state.finalHooks) {
		handler = hook(handler)
	}

	// setup normal middlewares
	for _, mw := range slices.Backward(r.middlewares) {
		handler = mw(handler)
	}

	return handler
}

type errorMessage struct {
	Error string `json:"error"`
}
