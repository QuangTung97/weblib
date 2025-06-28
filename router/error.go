package router

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
