package jwthook_test

import (
	"bufio"
	"github.com/earthboundkid/assert"
	jwt "github.com/spotlightpa/almanack/internal/services/jwthook"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestVerifyRequest(t *testing.T) {
	be := assert.FailNow(t)
	for _, tc := range []struct {
		name string
		iat  time.Time
	}{
		{"testdata/validate-umd.txt", time.Unix(1660249215, 0)},
		{"testdata/signup-umd.txt", time.Unix(1660249215, 0)},
		{"testdata/login-spotlight.txt", time.Unix(1660249192, 0)},
	} {
		var event any
		be.NotZero(jwt.VerifyRequest(
			getreq(t, tc.name),
			"abc", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "gotrue",
			tc.iat.AddDate(0, 0, 1),
			&event))
		be.NotZero(jwt.VerifyRequest(
			getreq(t, tc.name),
			"abc", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "gotrue",
			tc.iat.AddDate(0, 0, -1),
			&event))
		be.NotZero(jwt.VerifyRequest(
			getreq(t, tc.name),
			"123", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "gotrue",
			tc.iat,
			&event))
		be.NotZero(jwt.VerifyRequest(
			getreq(t, tc.name),
			"abc", "xyz", "gotrue",
			tc.iat,
			&event))
		be.NotZero(jwt.VerifyRequest(
			getreq(t, tc.name),
			"abc", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "xxx",
			tc.iat,
			&event))
		be.Zero(jwt.VerifyRequest(
			getreq(t, tc.name),
			"abc", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "gotrue",
			tc.iat,
			&event))
	}
	var event any
	be.NotZero(jwt.VerifyRequest(
		getreq(t, "testdata/login-spotlight-tampered.txt"),
		"abc", "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "gotrue",
		time.Unix(1660249192, 0),
		&event))
}

func getreq(t *testing.T, name string) *http.Request {
	t.Helper()
	be := assert.FailNow(t)
	f := be.OK(os.Open(name))
	defer f.Close()
	buf := bufio.NewReader(f)
	return be.OK(http.ReadRequest(buf))
}
