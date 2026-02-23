package youtube

import (
	"context"
	"encoding/xml"
	"flag"
	"net/http"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqxml"
	"github.com/earthboundkid/errorx/v2"
)

type Feed struct {
	ChannelID string
}

func AddFlags(fl *flag.FlagSet) (feed *Feed) {
	feed = new(Feed)
	fl.StringVar(&feed.ChannelID, "youtube-channel-id", "", "`URL` for YouTube feed")
	return feed
}

func (svc *Feed) FetchFeed(ctx context.Context, cl *http.Client) (entries []Entry, err error) {
	defer errorx.Trace(&err)

	var feed XML
	if err = requests.
		URL("https://www.youtube.com/feeds/videos.xml").
		Client(cl).
		Param("channel_id", svc.ChannelID).
		Handle(reqxml.To(&feed)).
		Fetch(ctx); err != nil {
		return nil, err
	}
	return feed.Entries, nil
}

type XML struct {
	XMLName xml.Name `xml:"feed"`
	Title   string   `xml:"title"`
	Author  Author   `xml:"author"`
	Entries []Entry  `xml:"entry"`
}

type Author struct {
	Name string `xml:"name"`
	URI  string `xml:"uri"`
}

type Entry struct {
	ID         string     `xml:"id"`
	VideoID    string     `xml:"videoId"`
	ChannelID  string     `xml:"channelId"`
	Title      string     `xml:"title"`
	Link       Link       `xml:"link"`
	Author     Author     `xml:"author"`
	Published  time.Time  `xml:"published"`
	Updated    time.Time  `xml:"updated"`
	MediaGroup MediaGroup `xml:"group"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type MediaGroup struct {
	Title       string         `xml:"title"`
	Description string         `xml:"description"`
	Thumbnail   MediaThumbnail `xml:"thumbnail"`
	Community   MediaCommunity `xml:"community"`
}

type MediaThumbnail struct {
	URL    string `xml:"url,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
}

type MediaCommunity struct {
	StarRating MediaStarRating `xml:"starRating"`
	Statistics MediaStatistics `xml:"statistics"`
}

type MediaStarRating struct {
	Count   int     `xml:"count,attr"`
	Average float64 `xml:"average,attr"`
}

type MediaStatistics struct {
	Views int `xml:"views,attr"`
}
