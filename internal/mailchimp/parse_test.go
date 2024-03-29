package mailchimp_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"golang.org/x/net/html"
)

func readContent(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", err
	}
	return mailchimp.PageContent(doc)
}

func TestSimpleParse(t *testing.T) {
	doc, err := readContent(strings.NewReader(`
<html>
<script>s1</script>
<body>
<p>p1</p>
<script>s2</script>
<p>p2</p>
<div>
	<p id="awesomewrap">p3</p>
</div>
<script>s3</script>
</body>
</html>
	`))
	be.NilErr(t, err)
	be.In(t, "p2", doc)
	be.NotIn(t, "p3", doc)
	be.NotIn(t, "script", doc)
}

var globalDoc string

func BenchmarkParseResponse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		doc, err := parseResponse()
		if err != nil {
			b.Fatalf("err: %v", err)
		}
		globalDoc = doc
	}
}

func parseResponse() (string, error) {
	u := `http://eepurl.com/hve6Dv`
	rsp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", rsp.Status)
	}
	return readContent(rsp.Body)
}

func BenchmarkParseBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		doc, err := parseBytes()
		if err != nil {
			b.Fatalf("err: %v", err)
		}
		globalDoc = doc
	}
}

func parseBytes() (string, error) {
	u := `http://eepurl.com/hve6Dv`
	rsp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", rsp.Status)
	}
	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	return readContent(bytes.NewReader(b))
}

func BenchmarkParseBuf(b *testing.B) {
	u := `http://eepurl.com/hve6Dv`
	for i := 0; i < b.N; i++ {
		doc, err := parseBuf(u)
		if err != nil {
			b.Fatalf("err: %v", err)
		}
		globalDoc = doc
	}
}

func parseBuf(u string) (string, error) {
	rsp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", rsp.Status)
	}
	return readContent(bufio.NewReader(rsp.Body))
}
