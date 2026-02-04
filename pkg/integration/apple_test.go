package integration_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/anf"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/jsonfeed"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestPublishAppleNews(t *testing.T) {
	almlog.UseTestLogger(t)
	p := createTestDB(t)
	q := db.New(p)
	ctx := t.Context()
	cl := &http.Client{
		Transport: reqtest.Replay("testdata/anf"),
	}
	http.DefaultClient.Transport = requests.ErrorTransport(errors.New("used default client"))
	res := anf.Response{
		Data: anf.Data{
			ID: "abc123",
		},
	}
	svc := almanack.Services{
		Client:  cl,
		Queries: q,
		NewsFeed: &jsonfeed.NewsFeed{
			URL: "https://www.spotlightpa.org/feeds/full.json",
		},
		ANF: &anf.Service{
			ChannelID: "test-channel-id",
			Key:       "test-key",
			Secret:    "test-secret",
			Client: &http.Client{
				Transport: reqtest.ReplayJSON(200, &res),
			},
		},
	}

	// PublishAppleNewsFeed should migrate legacy config to channel 1
	// and then process items
	be.NilErr(t, svc.PublishAppleNewsFeed(ctx))

	// Verify channel 1 was created
	channel, err := q.GetAppleNewsChannel(ctx, 1)
	be.NilErr(t, err)
	be.Equal(t, "test-channel-id", channel.ChannelID)
	be.Equal(t, "https://www.spotlightpa.org/feeds/full.json", channel.FeedUrl)

	// Check items were uploaded (via anf_channel_item)
	channelItems, err := q.ListANFChannelItemsForChannel(ctx, 1)
	be.NilErr(t, err)
	be.Nonzero(t, channelItems)

	// Running again should not re-upload unchanged items
	be.NilErr(t, svc.PublishAppleNewsFeed(ctx))
}

func TestPublishAppleNewsMultiChannel(t *testing.T) {
	almlog.UseTestLogger(t)
	p := createTestDB(t)
	q := db.New(p)
	ctx := t.Context()
	cl := &http.Client{
		Transport: reqtest.Replay("testdata/anf"),
	}
	http.DefaultClient.Transport = requests.ErrorTransport(errors.New("used default client"))
	res := anf.Response{
		Data: anf.Data{
			ID: "abc123",
		},
	}

	// Create two channels in the database (same feed URL for testing)
	channel1, err := q.CreateAppleNewsChannel(ctx, db.CreateAppleNewsChannelParams{
		Name:      "Channel 1",
		ChannelID: "channel-1-id",
		Key:       "key-1",
		Secret:    "secret-1",
		FeedUrl:   "https://www.spotlightpa.org/feeds/full.json",
		Active:    true,
	})
	be.NilErr(t, err)

	channel2, err := q.CreateAppleNewsChannel(ctx, db.CreateAppleNewsChannelParams{
		Name:      "Channel 2",
		ChannelID: "channel-2-id",
		Key:       "key-2",
		Secret:    "secret-2",
		FeedUrl:   "https://www.spotlightpa.org/feeds/full.json",
		Active:    true,
	})
	be.NilErr(t, err)

	svc := almanack.Services{
		Client:  cl,
		Queries: q,
		// No legacy ANF/NewsFeed - using DB channels
		ANF: &anf.Service{
			Client: &http.Client{
				Transport: reqtest.ReplayJSON(200, &res),
			},
		},
	}

	// Publishing should process both channels
	be.NilErr(t, svc.PublishAppleNewsFeed(ctx))

	// Both channels should have uploaded items
	items1, err := q.ListANFChannelItemsForChannel(ctx, channel1.ID)
	be.NilErr(t, err)
	be.Nonzero(t, items1)

	items2, err := q.ListANFChannelItemsForChannel(ctx, channel2.ID)
	be.NilErr(t, err)
	be.Nonzero(t, items2)

	// Verify last_synced_at was updated for both channels
	channel1Updated, err := q.GetAppleNewsChannel(ctx, channel1.ID)
	be.NilErr(t, err)
	be.True(t, channel1Updated.LastSyncedAt.Valid)

	channel2Updated, err := q.GetAppleNewsChannel(ctx, channel2.ID)
	be.NilErr(t, err)
	be.True(t, channel2Updated.LastSyncedAt.Valid)
}

func TestPublishAppleNewsInactiveChannel(t *testing.T) {
	almlog.UseTestLogger(t)
	p := createTestDB(t)
	q := db.New(p)
	ctx := t.Context()
	cl := &http.Client{
		Transport: reqtest.Replay("testdata/anf"),
	}

	// Create one active and one inactive channel
	_, err := q.CreateAppleNewsChannel(ctx, db.CreateAppleNewsChannelParams{
		Name:      "Active Channel",
		ChannelID: "active-channel-id",
		Key:       "key",
		Secret:    "secret",
		FeedUrl:   "https://www.spotlightpa.org/feeds/full.json",
		Active:    true,
	})
	be.NilErr(t, err)

	inactiveChannel, err := q.CreateAppleNewsChannel(ctx, db.CreateAppleNewsChannelParams{
		Name:      "Inactive Channel",
		ChannelID: "inactive-channel-id",
		Key:       "key",
		Secret:    "secret",
		FeedUrl:   "https://www.spotlightpa.org/feeds/full.json",
		Active:    false,
	})
	be.NilErr(t, err)

	res := anf.Response{
		Data: anf.Data{
			ID: "abc123",
		},
	}
	svc := almanack.Services{
		Client:  cl,
		Queries: q,
		ANF: &anf.Service{
			Client: &http.Client{
				Transport: reqtest.ReplayJSON(200, &res),
			},
		},
	}

	be.NilErr(t, svc.PublishAppleNewsFeed(ctx))

	// Inactive channel should not have last_synced_at set
	inactiveChannelUpdated, err := q.GetAppleNewsChannel(ctx, inactiveChannel.ID)
	be.NilErr(t, err)
	be.False(t, inactiveChannelUpdated.LastSyncedAt.Valid)
}
