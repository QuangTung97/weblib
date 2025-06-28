package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"

	"github.com/QuangTung97/weblib/hx"
	"github.com/QuangTung97/weblib/urls"
)

func HtmlGet[T any](
	router *Router,
	urlPath urls.Path[T],
	handler func(ctx Context, params T) (hx.Elem, error),
) {
	genericHandler := func(ctx Context, req any) (any, error) {
		resp, err := handler(ctx, req.(T))
		return resp, err
	}

	genericHandler = router.applyMiddlewares(genericHandler)

	router.state.chi.Get(urlPath.GetPattern(), func(writer http.ResponseWriter, req *http.Request) {
		var params T
		err := urls.SetStructWithValues(&params, urlPath.GetPathParams(), func(name string) string {
			return chi.URLParam(req, name)
		})
		if err != nil {
			router.handleHtmlError(err, writer)
			return
		}

		if err := req.ParseForm(); err != nil {
			router.handleHtmlError(err, writer)
			return
		}

		err = urls.SetStructWithValues(&params, urlPath.GetNonPathParams(), func(name string) string {
			return req.FormValue(name)
		})
		if err != nil {
			router.handleHtmlError(err, writer)
			return
		}

		// call handler
		ctx := NewContext(writer, req)
		resp, err := genericHandler(ctx, params)
		if err != nil {
			router.handleHtmlError(err, writer)
			return
		}

		outputElem, ok := resp.(hx.Elem)
		if !ok {
			err := fmt.Errorf("failed to convert response to hx.Elem")
			router.handleHtmlError(err, writer)
			return
		}

		_ = outputElem.Render(writer)
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

func (r *Router) handleHtmlError(err error, writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusBadRequest)
	enc := json.NewEncoder(writer)
	_ = enc.Encode(errorMessage{
		Error: err.Error(),
	})
}

type errorMessage struct {
	Error string `json:"error"`
}
