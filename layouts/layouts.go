package layouts

import (
	"embed"
	"html/template"
)

var FS embed.FS

func makeTemplate(names ...string) *template.Template {
	baseName := names[0]
	return template.Must(
		template.
			New(baseName).
			Funcs(nil).
			ParseFS(FS, names...))
}

var MailChimp = makeTemplate("mailchimp.html")

var Error = makeTemplate("error.html")
