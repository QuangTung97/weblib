package hx

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func formatXML(data []byte) ([]byte, error) {
	b := &bytes.Buffer{}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	encoder := xml.NewEncoder(b)
	encoder.Indent("", "\t")
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			if err := encoder.Flush(); err != nil {
				return nil, err
			}
			return b.Bytes(), nil
		}
		if err != nil {
			return nil, err
		}
		err = encoder.EncodeToken(token)
		if err != nil {
			return nil, err
		}
	}
}

func assertXmlContent(t *testing.T, expected string, elem Elem) {
	t.Helper()

	var buf bytes.Buffer
	if err := elem.Render(&buf); err != nil {
		panic(err)
	}

	formattedData, err := formatXML(buf.Bytes())
	if err != nil {
		panic(err)
	}

	assert.Equal(t, strings.TrimSpace(expected), string(formattedData))
}

func assertSimpleContent(t *testing.T, expected string, elem Elem) {
	t.Helper()
	var buf bytes.Buffer
	if err := elem.Render(&buf); err != nil {
		panic(err)
	}
	assert.Equal(t, strings.TrimSpace(expected), buf.String())
}

func TestElem_Render(t *testing.T) {
	t.Run("simple empty div", func(t *testing.T) {
		assertXmlContent(t, "<div></div>", Div())
	})

	t.Run("div inside div", func(t *testing.T) {
		expected := `
<div>
	<div></div>
	<div></div>
</div>
`
		assertXmlContent(t, expected, Div(
			Div(),
			Div(),
		))
	})

	t.Run("simple div with text", func(t *testing.T) {
		assertXmlContent(t, "<div>Hello World &lt;&gt;</div>", Div(
			Text("Hello World <>"),
		))
	})

	t.Run("with attribute class", func(t *testing.T) {
		expected := `<div class="px-2 shadow" id="test-id">Hello</div>`
		assertXmlContent(t, expected, Div(
			Class("px-2 shadow"),
			ID("test-id"),
			Text("Hello"),
		))
	})

	t.Run("line break", func(t *testing.T) {
		expected := `<br>`
		assertSimpleContent(t, expected, Br())
	})

	t.Run("input tag", func(t *testing.T) {
		expected := `<input name="name01">`
		assertSimpleContent(t, expected, Input(
			Name("name01"),
		))
	})

	t.Run("none", func(t *testing.T) {
		expected := ``
		assertSimpleContent(t, expected, None())
	})

	t.Run("div inside div, with grouping", func(t *testing.T) {
		expected := `
<div>
	<div id="test-id"></div>
	<div></div>
	<div></div>
</div>
`
		assertXmlContent(t, expected, Div(
			Group(
				Div(
					ID("test-id"),
				),
				Div(),
			),
			Div(),
		))
	})

	t.Run("group attribute", func(t *testing.T) {
		expected := `
<div id="test-id" class="class01"></div>
`
		assertXmlContent(t, expected, Div(
			Group(
				ID("test-id"),
				Class("class01"),
			),
		))
	})

	t.Run("iterator", func(t *testing.T) {
		iter := slices.Values([]Elem{
			Ul(),
			Div(),
		})

		expected := `
<div>
	<ul></ul>
	<div></div>
</div>
`
		assertXmlContent(t, expected, Div(
			Collect(iter),
		))
	})
}

func TestElem_Basic(t *testing.T) {
	t.Run("a tag", func(t *testing.T) {
		assertXmlContent(t, `<a href="/user/login"></a>`, A(
			Href("/user/login"),
		))
	})
}

type errorWriter struct {
	data string
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	w.data += string(p)
	return 0, errors.New("write error")
}

func TestElem_Render__With_Writer_Error(t *testing.T) {
	elem := Div(
		Ul(),
		Ul(),
	)

	var w errorWriter
	err := elem.Render(&w)
	assert.Equal(t, errors.New("write error"), err)
	assert.Equal(t, "<", w.data)
}
