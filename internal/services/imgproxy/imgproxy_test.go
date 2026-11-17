package imgproxy_test

import (
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/services/imgproxy"
)

func TestURL(t *testing.T) {
	be := assert.FailsNow(t)
	for _, tc := range []struct {
		have imgproxy.URL
		want string
	}{
		{},
		{imgproxy.URL{
			FilePath:  "abc/123",
			Width:     100,
			Height:    100,
			Extension: "jpeg",
		}, "https://images.data.spotlightpa.org/insecure/rt:fill/w:100/h:100/g:sm/el:1/YWJjLzEyMw==.jpeg"},
	} {
		be.Equal(tc.have.String(), tc.want)
	}
}
