package hx

import (
	"html"
	"iter"
	"slices"
	"strings"
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

func ClassGroup(children ...Elem) Elem {
	var buf strings.Builder
	var counter int

	for _, child := range children {
		if child.elemType != elemTypeAttribute {
			continue
		}
		if string(child.name) != "class" {
			continue
		}

		counter++
		if counter > 1 {
			buf.WriteString(" ")
		}
		buf.WriteString(string(child.value))
	}

	return Class(buf.String())
}

// ----------------------------------------------------
// Normal Tags
// ----------------------------------------------------

func Div(children ...Elem) Elem {
	return newNormalTag("div", children...)
}

func Button(children ...Elem) Elem {
	return newNormalTag("button", children...)
}

func A(children ...Elem) Elem {
	return newNormalTag("a", children...)
}

func Script(children ...Elem) Elem {
	return newNormalTag("script", children...)
}

func Ul(children ...Elem) Elem {
	return newNormalTag("ul", children...)
}

func Li(children ...Elem) Elem {
	return newNormalTag("li", children...)
}

func Select(children ...Elem) Elem {
	return newNormalTag("select", children...)
}

func Option(children ...Elem) Elem {
	return newNormalTag("option", children...)
}

// ----------------------------------------------------
// Attributes
// ----------------------------------------------------

func Class(className string) Elem {
	return newNormalAttr("class", className)
}

func ID(id ElemID) Elem {
	return newNormalAttr("id", string(id))
}

func Name(val string) Elem {
	return newNormalAttr("name", val)
}

func Href(urlPath string) Elem {
	return newUnsafeAttr("href", urlPath)
}

func Src(urlPath string) Elem {
	return newUnsafeAttr("src", urlPath)
}

func Rel(value string) Elem {
	return newNormalAttr("rel", value)
}

func Text(text string) Elem {
	return Elem{
		elemType: elemTypeContent,
		value:    []byte(html.EscapeString(text)),
	}
}

// ----------------------------------------------------
// Simple Tag
// ----------------------------------------------------

func Br() Elem {
	return newSimpleTag("br")
}

func Input(children ...Elem) Elem {
	return newSimpleTag("input", children...)
}

func Link(children ...Elem) Elem {
	return newSimpleTag("link", children...)
}

// ----------------------------------------------------
// Empty Attribute
// ----------------------------------------------------

func Required() Elem {
	return newEmptyAttr("required")
}

func Disabled() Elem {
	return newEmptyAttr("disabled")
}

// ----------------------------------------------------
// Helper Functions
// ----------------------------------------------------

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

func newEmptyAttr(name string) Elem {
	return Elem{
		elemType: elemTypeEmptyAttribute,
		name:     []byte(name),
	}
}
