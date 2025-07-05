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
		newNormalTag("title", Text(title)),
		newSimpleTag("meta",
			newNormalAttr("charset", "UTF-8"),
		),
		head,
	}

	htmlContent := newNormalTag(
		"html",
		newNormalAttr("lang", conf.lang),

		newNormalTag("head", headList...),
		newNormalTag("body", body),
	)

	return Group(
		newSimpleTag("!DOCTYPE html"),
		htmlContent,
	)
}
