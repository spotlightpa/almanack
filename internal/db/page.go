package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/carlmjohnson/errutil"
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
