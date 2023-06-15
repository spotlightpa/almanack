package jwthook_test

import (
	"bufio"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	jwt "github.com/spotlightpa/almanack/internal/jwthook"
)

func TestVerifyRequest(t *testing.T) {
	for _, tc := range []struct {
		name string
		iat  time.Time
	}{
		{"testdata/validate-umd.txt", time.Unix(1660249215, 0)},
		{"testdata/signup-umd.txt", time.Unix(1660249215, 0)},
		{"testdata/login-spotlight.txt", time.Unix(1660249192, 0)},
	} {
		var event any
		be.Nonzero(t, jwt.VerifyRequest(
			getreq(t, tc.name),
			"abc", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "gotrue",
			tc.iat.AddDate(0, 0, 1),
			&event))
		be.Nonzero(t, jwt.VerifyRequest(
			getreq(t, tc.name),
			"abc", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "gotrue",
			tc.iat.AddDate(0, 0, -1),
			&event))
		be.Nonzero(t, jwt.VerifyRequest(
			getreq(t, tc.name),
			"123", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "gotrue",
			tc.iat,
			&event))
		be.Nonzero(t, jwt.VerifyRequest(
			getreq(t, tc.name),
			"abc", "xyz", "gotrue",
			tc.iat,
			&event))
		be.Nonzero(t, jwt.VerifyRequest(
			getreq(t, tc.name),
			"abc", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "xxx",
			tc.iat,
			&event))
		be.NilErr(t, jwt.VerifyRequest(
			getreq(t, tc.name),
			"abc", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "gotrue",
			tc.iat,
			&event))
	}
	var event any
	be.Nonzero(t, jwt.VerifyRequest(
		getreq(t, "testdata/login-spotlight-tampered.txt"),
		"abc", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "gotrue",
		time.Unix(1660249192, 0),
		&event))
}

func getreq(t *testing.T, name string) *http.Request {
	t.Helper()
	f, err := os.Open(name)
	be.NilErr(t, err)
	defer f.Close()
	buf := bufio.NewReader(f)
	req, err := http.ReadRequest(buf)
	be.NilErr(t, err)
	return req
}
