package imgproxy

import (
	"cmp"
	"encoding/base64"
	"fmt"
)

type URL struct {
	FilePath       string
	Width, Height  int
	ResizingType   string
	Quality        int
	Gravity        string
	Extension      string
	DisableEnlarge bool
}

func (u URL) String() string {
	if u.FilePath == "" {
		return ""
	}

	const (
		baseURL   = "https://images.data.spotlightpa.org"
		signature = "insecure"
	)

	gravity := cmp.Or(u.Gravity, "sm")
	resizingType := cmp.Or(u.ResizingType, "fill")

	qualityStr := ""
	if u.Quality != 0 {
		qualityStr = fmt.Sprintf("/q:%d", u.Quality)
	}
	enlarge := "1"
	if u.DisableEnlarge {
		enlarge = "0"
	}

	encodedSourceURL := base64.StdEncoding.EncodeToString([]byte(u.FilePath))

	return fmt.Sprintf(
		"%s/%s/rt:%s/w:%d/h:%d/g:%s/el:%s%s/%s.%s",
		baseURL, signature, resizingType, u.Width, u.Height, gravity, enlarge,
		qualityStr, encodedSourceURL, u.Extension,
	)
}
