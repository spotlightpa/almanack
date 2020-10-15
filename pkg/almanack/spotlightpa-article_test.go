package almanack

import (
	"reflect"
	"testing"
)

func TestToFromTOML(t *testing.T) {
	cases := map[string]SpotlightPAArticle{
		"empty":   {},
		"body":    {Body: "\n ## subhead ! \n"},
		"fm":      {Hed: "Hello", Authors: []string{"john", "smith"}},
		"body+fm": {Hed: "Hello", Authors: []string{"john", "smith"}, Body: "## subhead !"},
		"extra-delimiters": {
			Hed: "Hello", Authors: []string{"john", "smith"},
			Body: "## subhead !\n+++\n\nmore\n+++\nstuff",
		},
	}
	for name, art := range cases {
		t.Run(name, func(t *testing.T) {
			toml, err := art.ToTOML()
			if err != nil {
				t.Fatalf("err ToTOML: %v", err)
			}
			var art2 SpotlightPAArticle
			if err = art2.FromTOML(toml); err != nil {
				t.Fatalf("err FromTOML: %v", err)
			}
			if !reflect.DeepEqual(art, art2) {
				t.Errorf("article did not round trip")
			}
		})
	}
}

func TestFromToTOML(t *testing.T) {
	cases := map[string]struct {
		ok      bool
		content string
	}{
		"empty": {false, ``},
		"blank": {true, `+++
arc-id = ""
internal-id = ""
internal-budget = ""
image = ""
image-description = ""
image-caption = ""
image-credit = ""
image-size = ""
published = 0000-01-01T00:00:00Z
slug = ""
byline = ""
title = ""
subtitle = ""
description = ""
blurb = ""
kicker = ""
linktitle = ""
suppress-featured = false
weight = 0
url = ""
modal-exclude = false
no-index = false
language-code = ""
layout = ""
extended-kicker = ""
+++


`},
		"fm": {true, `+++
arc-id = "123"
internal-id = "spl123"
internal-budget = "hello"
image = "xyz.jpeg"
image-description = "desc"
image-caption = "capt"
image-credit = "cred"
image-size = "inline"
published = 2006-01-01T00:00:00Z
slug = "slug"
authors = ["john", "doe"]
byline = "menen"
title = "hed"
subtitle = "subtitle"
description = "desc2"
blurb = "blurb"
kicker = "kick"
linktitle = "lt"
suppress-featured = true
weight = 2
url = "/url/"
modal-exclude = true
no-index = true
language-code = "es"
layout = "fancy"
extended-kicker = "More News"
+++


`},
		"fm+body": {true, `+++
arc-id = "123"
internal-id = "spl123"
internal-budget = "hello"
image = "xyz.jpeg"
image-description = "desc"
image-caption = "capt"
image-credit = "cred"
image-size = "inline"
published = 2006-01-01T00:00:00Z
slug = "slug"
authors = ["john", "doe"]
byline = "menen"
title = "hed"
subtitle = "subtitle"
description = "desc2"
blurb = "blurb"
kicker = "kick"
linktitle = "lt"
suppress-featured = true
weight = 2
url = "/url/"
modal-exclude = true
no-index = true
language-code = "es"
layout = "fancy"
extended-kicker = "More News"
+++

Hello, world!
+++
more~~
<h1></h1>
`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var art SpotlightPAArticle
			err := art.FromTOML(tc.content)
			if !tc.ok {
				if err == nil {
					t.Error("expected err FromTOML")
				}
				return
			}
			if err != nil {
				t.Fatalf("err FromTOML: %v", err)
			}

			toml, err := art.ToTOML()
			if err != nil {
				t.Fatalf("err ToTOML: %v", err)
			}
			if tc.content != toml {
				t.Errorf("article did not round trip: %q != %q",
					tc.content, toml,
				)
			}
		})
	}
}
