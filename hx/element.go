package hx

import (
	"html"
	"iter"
	"slices"
)

func None() Elem {
	return Elem{elemType: elemTypeNone}
}

func Group(children ...Elem) Elem {
	return Elem{
		elemType: elemTypeGroup,
		children: slices.Values(children),
	}
}

func Collect(seq iter.Seq[Elem]) Elem {
	return Elem{
		elemType: elemTypeIter,
		children: seq,
	}
}

func newNormalTag(name string, children ...Elem) Elem {
	e := Elem{
		elemType: elemTypeNormalTag,
		name:     []byte(name),
	}

	if len(children) > 0 {
		e.children = slices.Values(children)
	}

	return e
}

func Div(children ...Elem) Elem {
	return newNormalTag("div", children...)
}

func A(children ...Elem) Elem {
	return newNormalTag("a", children...)
}

func Href(urlPath string) Elem {
	return newUnsafeAttr("href", urlPath)
}

func Script(children ...Elem) Elem {
	return newNormalTag("script", children...)
}

func Src(urlPath string) Elem {
	return newUnsafeAttr("src", urlPath)
}

func Rel(value string) Elem {
	return newUnsafeAttr("rel", value)
}

func Ul(children ...Elem) Elem {
	return newNormalTag("ul", children...)
}

func Text(text string) Elem {
	return Elem{
		elemType: elemTypeContent,
		value:    []byte(html.EscapeString(text)),
	}
}

func newNormalAttr(name string, value string) Elem {
	return Elem{
		elemType: elemTypeAttribute,
		name:     []byte(name),
		value:    []byte(html.EscapeString(value)),
	}
}

func newUnsafeAttr(name string, value string) Elem {
	return Elem{
		elemType: elemTypeAttribute,
		name:     []byte(name),
		value:    []byte(value),
	}
}

func Class(className string) Elem {
	return newNormalAttr("class", className)
}

func ID(id ElemID) Elem {
	return newNormalAttr("id", string(id))
}

func Name(val string) Elem {
	return newNormalAttr("name", val)
}

func newSimpleTag(name string, children ...Elem) Elem {
	e := Elem{
		elemType: elemTypeSimpleTag,
		name:     []byte(name),
	}

	if len(children) > 0 {
		e.children = slices.Values(children)
	}

	return e

}

func Br() Elem {
	return newSimpleTag("br")
}

func Input(children ...Elem) Elem {
	return newSimpleTag("input", children...)
}

func Link(children ...Elem) Elem {
	return newSimpleTag("link", children...)
}
