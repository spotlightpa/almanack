package layouts

import (
	"embed"
	"html/template"
)

//go:embed *.html
var FS embed.FS

func makeTemplate(names ...string) *template.Template {
	baseName := names[0]
	return template.Must(
		template.
			New(baseName).
			Funcs(nil).
			ParseFS(FS, names...))
}

var Error = makeTemplate("error.html")
