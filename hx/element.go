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
	return NewNormalTag("div", children...)
}

func Button(children ...Elem) Elem {
	return NewNormalTag("button", children...)
}

func A(children ...Elem) Elem {
	return NewNormalTag("a", children...)
}

func Script(children ...Elem) Elem {
	return NewNormalTag("script", children...)
}

func Ul(children ...Elem) Elem {
	return NewNormalTag("ul", children...)
}

func Li(children ...Elem) Elem {
	return NewNormalTag("li", children...)
}

func Select(children ...Elem) Elem {
	return NewNormalTag("select", children...)
}

func Option(children ...Elem) Elem {
	return NewNormalTag("option", children...)
}

// ----------------------------------------------------
// Attributes
// ----------------------------------------------------

func Class(className string) Elem {
	return NewNormalAttr("class", className)
}

func ID(id ElemID) Elem {
	return NewNormalAttr("id", string(id))
}

func Name(val string) Elem {
	return NewNormalAttr("name", val)
}

func Href(urlPath string) Elem {
	return NewUnsafeAttr("href", urlPath)
}

func Src(urlPath string) Elem {
	return NewUnsafeAttr("src", urlPath)
}

func Rel(value string) Elem {
	return NewNormalAttr("rel", value)
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
	return NewSimpleTag("br")
}

func Input(children ...Elem) Elem {
	return NewSimpleTag("input", children...)
}

func Link(children ...Elem) Elem {
	return NewSimpleTag("link", children...)
}

// ----------------------------------------------------
// Empty Attribute
// ----------------------------------------------------

func Required() Elem {
	return NewEmptyAttr("required")
}

func Disabled() Elem {
	return NewEmptyAttr("disabled")
}

// ----------------------------------------------------
// Helper Functions
// ----------------------------------------------------

func NewNormalTag(name string, children ...Elem) Elem {
	e := Elem{
		elemType: elemTypeNormalTag,
		name:     []byte(name),
	}

	if len(children) > 0 {
		e.children = slices.Values(children)
	}

	return e
}

func NewNormalAttr(name string, value string) Elem {
	return Elem{
		elemType: elemTypeAttribute,
		name:     []byte(name),
		value:    []byte(html.EscapeString(value)),
	}
}

func NewUnsafeAttr(name string, value string) Elem {
	return Elem{
		elemType: elemTypeAttribute,
		name:     []byte(name),
		value:    []byte(value),
	}
}

func NewSimpleTag(name string, children ...Elem) Elem {
	e := Elem{
		elemType: elemTypeSimpleTag,
		name:     []byte(name),
	}

	if len(children) > 0 {
		e.children = slices.Values(children)
	}

	return e
}

func NewEmptyAttr(name string) Elem {
	return Elem{
		elemType: elemTypeEmptyAttribute,
		name:     []byte(name),
	}
}
