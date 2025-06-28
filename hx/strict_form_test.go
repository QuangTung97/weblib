package hx

import (
	"testing"

	"github.com/QuangTung97/weblib/urls"
)

type testParams struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestStrictForm(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		elem := StrictForm(
			urls.New[testParams]("/home"),
			Div(
				Name("age"),
			),
			Div(
				Name("info"),
			),
		)

		expected := `
<form>
	<div name="age"></div>
	<div name="info"></div>
	<div class="text-2xl text-red-700">StrictForm violation on input name &#39;age&#39;</div>
</form>
`
		assertXmlContent(t, expected, elem)
	})

	t.Run("normal", func(t *testing.T) {
		elem := StrictForm(
			urls.New[testParams]("/home"),
			Div(
				Name("id"),
			),
			Div(
				Name("name"),
			),
		)

		expected := `
<form>
	<div name="id"></div>
	<div name="name"></div>
</form>
`
		assertXmlContent(t, expected, elem)
	})
}
