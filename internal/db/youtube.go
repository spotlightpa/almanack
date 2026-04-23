package db

import (
	"fmt"
	"strings"

	"github.com/spotlightpa/almanack/internal/stringx"
)

func slugifyVideoID(s string) string {
	s, _ = stringx.StripAccents(s)
	hadDash := false
	f := func(r rune) rune {
		switch {
		case
			r >= 'A' && r <= 'Z',
			r >= 'a' && r <= 'z',
			r >= '0' && r <= '9',
			r == '.',
			r == '_':
			hadDash = false
			return r
		case hadDash:
			return -1
		}
		hadDash = true
		return '-'
	}
	return strings.Map(f, s)
}

func (video Youtube) FilePath() string {
	return fmt.Sprintf("content/videos/%s.md", slugifyVideoID(video.ExternalID))
}

func (video Youtube) YouTubeID() string {
	return strings.TrimPrefix(video.ExternalID, "yt:video:")
}

func (video Youtube) IsShort() bool {
	return strings.Contains(video.URL, "/shorts/")
}
