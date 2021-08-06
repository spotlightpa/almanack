package db

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/carlmjohnson/errutil"
	"github.com/spotlightpa/almanack/internal/timeutil"
)

func (page *Page) ToTOML() (string, error) {
	var buf strings.Builder
	buf.WriteString("+++\n")
	enc := toml.NewEncoder(&buf)
	// Remove blank values
	frontmatter := Map{}
	for key, val := range page.Frontmatter {
		if val == nil {
			continue
		}
		if s, ok := val.(string); ok && s == "" {
			continue
		}
		if t, ok := val.(time.Time); ok && t.IsZero() {
			continue
		}
		if n, ok := val.(float64); ok && n == 0.0 {
			continue
		}
		frontmatter[key] = val
	}
	if err := enc.Encode(frontmatter); err != nil {
		return "", err
	}
	buf.WriteString("+++\n\n")
	buf.WriteString(page.Body)
	buf.WriteString("\n")
	return buf.String(), nil
}

func (page *Page) FromTOML(content string) (err error) {
	defer errutil.Prefix(&err, "problem reading TOML")

	const delimiter = "+++\n"
	var frontmatter, body string

	if !strings.HasPrefix(content, delimiter) {
		return fmt.Errorf("could not parse frontmatter: no prefix delimiter")
	}
	frontmatter = content[len(delimiter):]
	if end := strings.Index(frontmatter, delimiter); end != -1 {
		body = frontmatter[end+len(delimiter):]
		frontmatter = frontmatter[:end]
	} else {
		return fmt.Errorf("could not parse frontmatter: no end delimiter")
	}

	m := map[string]interface{}{}
	if _, err := toml.Decode(frontmatter, &m); err != nil {
		return err
	}
	page.Frontmatter = m
	body = strings.TrimPrefix(body, "\n")
	body = strings.TrimSuffix(body, "\n")
	page.Body = body
	return nil
}

func (page *Page) SetURLPath() {
	if page.URLPath.Valid && page.URLPath.String != "" {
		return
	}
	if u, _ := page.Frontmatter["url"].(string); u != "" {
		page.URLPath.String = u
		page.URLPath.Valid = true
		return
	}
	upath := page.FilePath
	upath = strings.TrimPrefix(upath, "content")
	upath = strings.TrimSuffix(upath, ".md")
	dir, fname := path.Split(upath)
	if dir == "/news/" {
		if pub, ok := timeutil.GetTime(page.Frontmatter, "published"); ok {
			pub = timeutil.ToEST(pub)
			dir = pub.Format("/news/2006/01/")
		}
	}
	if slug, _ := page.Frontmatter["slug"].(string); slug != "" {
		fname = slug
	}

	upath = path.Join(dir, fname)
	if upath != "" && !strings.HasSuffix(upath, "/") {
		upath += "/"
	}
	page.URLPath.String = upath
	page.URLPath.Valid = upath != ""
}
