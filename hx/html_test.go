package hx

import (
	"bytes"
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/skeleton.html
var skeletonHTMLData string

func TestHtml(t *testing.T) {
	elem := Html(
		"Hello Title",
		Script(
			Src("/static/core.js"),
		),
		Div(
			Text("Body Test"),
		),
		WithHtmlLang("vi"),
	)

	var buf bytes.Buffer
	if err := elem.Render(&buf); err != nil {
		panic(err)
	}

	expected := skeletonHTMLData
	expected = strings.ReplaceAll(expected, "\n", "")
	expected = strings.ReplaceAll(expected, "    ", "")
	assert.Equal(t, expected, buf.String())
}
