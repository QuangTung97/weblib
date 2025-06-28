package hx

import (
	"io"
	"iter"
)

type Elem struct {
	elemType elemType
	name     []byte
	value    []byte
	children iter.Seq[Elem]
	extra    *elemExtraInfo
}

type ElemID string

type elemExtraInfo struct {
	childValidator    func(child Elem, w *writerHelper)
	afterTravelRender func(w *writerHelper)
}

type elemType int

const (
	elemTypeNone elemType = iota
	elemTypeNormalTag
	elemTypeSimpleTag
	elemTypeContent
	elemTypeAttribute
	elemTypeGroup
	elemTypeIter
)

func (e Elem) Render(writer io.Writer) error {
	w := writerHelper{
		writer:       writer,
		validateFunc: func(elem Elem, w *writerHelper) {},
	}
	e.renderWithHelper(&w)
	return w.err
}

var openTagBegin = []byte("<")
var openTagEnd = []byte(">")
var closeTagBegin = []byte("</")
var singleSpace = []byte(" ")
var equalSign = []byte("=")
var doubleQuote = []byte(`"`)

func (e Elem) renderWithHelper(w *writerHelper) {
	// setup empty children
	if e.children == nil {
		e.children = func(yield func(Elem) bool) {}
	}

	if e.extra != nil {
		w.validateFunc = e.extra.childValidator
	}

	switch e.elemType {
	case elemTypeNormalTag:
		w.writeBytes(openTagBegin)
		w.writeBytes(e.name)
		for child := range e.children {
			child.renderAttribute(w)
		}
		w.writeBytes(openTagEnd)

		for child := range e.children {
			w.validateFunc(child, w)
			child.renderWithHelper(w)
			if w.err != nil {
				return
			}
		}

		if w.validateFailed {
			if e.extra != nil {
				e.extra.afterTravelRender(w)
			}
		}

		w.writeBytes(closeTagBegin)
		w.writeBytes(e.name)
		w.writeBytes(openTagEnd)

	case elemTypeSimpleTag:
		w.writeBytes(openTagBegin)
		w.writeBytes(e.name)
		for child := range e.children {
			child.renderAttribute(w)
		}
		w.writeBytes(openTagEnd)

	case elemTypeContent:
		w.writeBytes(e.value)

	case elemTypeGroup:
		for child := range e.children {
			child.renderWithHelper(w)
		}

	case elemTypeIter:
		for child := range e.children {
			child.renderWithHelper(w)
		}

	default:
	}
}

func (e Elem) renderAttribute(w *writerHelper) {
	switch e.elemType {
	case elemTypeAttribute:
		w.writeBytes(singleSpace)
		w.writeBytes(e.name)
		w.writeBytes(equalSign)
		w.writeBytes(doubleQuote)
		w.writeBytes(e.value)
		w.writeBytes(doubleQuote)

	case elemTypeGroup:
		for child := range e.children {
			child.renderAttribute(w)
		}

	default:
	}
}

type writerHelper struct {
	writer io.Writer
	err    error

	validateFunc   func(elem Elem, w *writerHelper)
	validateFailed bool
}

func (w *writerHelper) writeBytes(data []byte) {
	if w.err != nil {
		return
	}

	_, err := w.writer.Write(data)
	if err != nil {
		w.err = err
	}
}
