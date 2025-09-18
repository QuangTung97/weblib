package hx

type htmlConfig struct {
	lang string
}

type HtmlOption func(conf *htmlConfig)

func WithHtmlLang(lang string) HtmlOption {
	return func(conf *htmlConfig) {
		conf.lang = lang
	}
}

func Html(
	title string,
	head Elem,
	body Elem,
	options ...HtmlOption,
) Elem {
	conf := htmlConfig{
		lang: "en",
	}
	for _, fn := range options {
		fn(&conf)
	}

	headList := []Elem{
		NewNormalTag("title", Text(title)),
		NewSimpleTag("meta",
			NewNormalAttr("charset", "UTF-8"),
		),
		NewSimpleTag("meta",
			NewNormalAttr("name", "viewport"),
			NewNormalAttr("content", "width=device-width, initial-scale=1.0"),
		),
		head,
	}

	htmlContent := NewNormalTag(
		"html",
		NewNormalAttr("lang", conf.lang),

		NewNormalTag("head", headList...),
		NewNormalTag("body", body),
	)

	return Group(
		NewSimpleTag("!DOCTYPE html"),
		htmlContent,
	)
}
