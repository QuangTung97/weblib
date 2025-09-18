package router

import (
	"encoding/json"
	"net/http"
)

type HtmlErrorReason int

const (
	ReasonBadPathParam HtmlErrorReason = iota + 1
	ReasonBadFormParam
	ReasonBadResponseType
)

type HtmlError struct {
	Reason  HtmlErrorReason
	Message string
}

func (e *HtmlError) Error() string {
	return e.Message
}

func (r *Router) SetCustomHtmlErrorHandler(handler func(ctx Context, err error)) {
	r.state.handleHtmlError = handler
}

func (r *Router) DefaultHtmlErrorHandler(ctx Context, err error) {
	type errorMessage struct {
		Error string `json:"error"`
	}

	writer := ctx.GetWriter()
	writer.WriteHeader(http.StatusBadRequest)
	writer.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(writer)
	_ = enc.Encode(errorMessage{
		Error: err.Error(),
	})
}
