package hx

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/QuangTung97/weblib/urls"
)

func StrictForm[T any](
	urlPath urls.Path[T],
	children ...Elem,
) Elem {
	paramSet := map[string]struct{}{}
	for _, param := range urlPath.GetNonPathParams() {
		paramSet[param] = struct{}{}
	}

	e := newNormalTag("form", children)

	var violatedName string

	validateFunc := func(child Elem, w *writerHelper) {
		if child.elemType != elemTypeAttribute {
			return
		}
		if !bytes.Equal(child.name, []byte("name")) {
			return
		}

		nameKey := string(child.value)
		_, ok := paramSet[nameKey]
		if ok {
			return
		}

		if !w.validateFailed {
			w.validateFailed = true
			violatedName = nameKey
		}
	}

	e.extra = &elemExtraInfo{
		childValidator: validateFunc,
		afterTravelRender: func(w *writerHelper) {
			// log using slog
			StrictFormErrorLogFunc(violatedName)

			validateErrorMsg := fmt.Sprintf("StrictForm violation on input name '%s'", violatedName)
			elem := Div(
				Class("text-2xl text-red-700"),
				Text(validateErrorMsg),
			)
			elem.renderWithHelper(w)
		},
	}

	return e
}

var StrictFormErrorLogFunc = func(inputName string) {
	slog.Error("strict form violation", "name", inputName)
}
