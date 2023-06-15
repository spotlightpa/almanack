package layouts

import (
	"html/template"
)

func makeTemplate(names ...string) *template.Template {
	return nil
}

var MailChimp = makeTemplate("mailchimp.html")

var Error = makeTemplate("error.html")
