package mailchimp_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/spotlightpa/almanack/internal/mailchimp"
)

func assert(t *testing.T, v bool, format string, args ...interface{}) {
	t.Helper()
	if !v {
		t.Fatalf(format, args...)
	}
}

func assertErrNil(t *testing.T, err error) {
	t.Helper()
	assert(t, err == nil, "err != nil: %v", err)
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	assert(t, strings.Contains(s, substr), "%q missing from %q", substr, s)
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	assert(t, !strings.Contains(s, substr), "%q present in %q", substr, s)
}

func TestSimpleParse(t *testing.T) {
	doc, err := mailchimp.PageContent(strings.NewReader(`
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
	assertErrNil(t, err)
	assertContains(t, doc, "p2")
	assertNotContains(t, doc, "p3")
	assertNotContains(t, doc, "script")
}

func TestParse(t *testing.T) {
	doc, err := parseResponse()
	assertErrNil(t, err)
	os.WriteFile("body.html", []byte(doc), os.ModePerm)
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
	return mailchimp.PageContent(rsp.Body)
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
	return mailchimp.PageContent(bytes.NewReader(b))
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
	return mailchimp.PageContent(bufio.NewReader(rsp.Body))
}
