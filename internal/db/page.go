package db

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/carlmjohnson/errutil"
)

func (page *Page) ToTOML() (string, error) {
	var buf strings.Builder
	buf.WriteString("+++\n")
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(page.Frontmatter); err != nil {
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