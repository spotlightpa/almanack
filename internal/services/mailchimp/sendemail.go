package mailchimp

import (
	"context"
	"net/http"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/earthboundkid/errorx/v2"
	"github.com/earthboundkid/resperr/v2"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (v3 V3) SendEmail(ctx context.Context, subject, body string) (err error) {
	l := almlog.FromContext(ctx)
	var mcErr ErrorResponse
	defer func() {
		if err != nil {
			err = resperr.New(http.StatusBadGateway,
				"MailChimp [%3d] %s: %s: %q: %w",
				mcErr.Status, mcErr.Title, mcErr.Detail, mcErr.Errors, err,
			)
		}
	}()
	defer errorx.Trace(&err)

	var res PostCampaignResponse
	if err = requests.
		New(v3.config).
		Path("campaigns").
		BodyJSON(PostCampaignRequest{
			Type:       "plaintext",
			Recipients: Recipients{ListID: v3.listID},
			Settings: Settings{
				FromName:    "Spotlight PA",
				ReplyTo:     "press@spotlightpa.org",
				SubjectLine: subject,
				Title:       subject,
			},
		}).
		ErrorJSON(&mcErr).
		ToJSON(&res).
		Fetch(ctx); err != nil {
		return err
	}

	l.InfoContext(ctx, "mailchimp.SendEmail: created campaign", "campaign_id", res.ID)

	var putRes PutCampaignResponse
	if err = requests.
		New(v3.config).
		Put().
		Pathf("campaigns/%s/content", res.ID).
		BodyJSON(PutCampaignRequest{
			PlainText: body,
		}).
		ErrorJSON(&mcErr).
		ToJSON(&putRes).
		Fetch(ctx); err != nil {
		return err
	}

	l.InfoContext(ctx, "mailchimp.SendEmail: configured campaign", "campaign_id", res.ID)

	if err = requests.
		New(v3.config).
		Pathf("campaigns/%s/actions/send", res.ID).
		BodyBytes(nil).
		ErrorJSON(&mcErr).
		Fetch(ctx); err != nil {
		return err
	}

	l.InfoContext(ctx, "mailchimp.SendEmail: sent campaign", "campaign_id", res.ID)

	return nil
}

type ErrorResponse struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
	Errors   []any  `json:"errors"`
}

type PostCampaignRequest struct {
	Type            string           `json:"type"`
	Recipients      Recipients       `json:"recipients"`
	Settings        Settings         `json:"settings"`
	VariateSettings *VariateSettings `json:"variate_settings,omitempty"`
	Tracking        *Tracking        `json:"tracking,omitempty"`
	RssOpts         *RssOpts         `json:"rss_opts,omitempty"`
	SocialCard      *SocialCard      `json:"social_card,omitempty"`
	ContentType     string           `json:"content_type,omitempty"`
}

type PostCampaignResponse struct {
	ID                string          `json:"id"`
	WebID             int             `json:"web_id"`
	ParentCampaignID  string          `json:"parent_campaign_id"`
	Type              string          `json:"type"`
	CreateTime        time.Time       `json:"create_time"`
	ArchiveURL        string          `json:"archive_url"`
	LongArchiveURL    string          `json:"long_archive_url"`
	Status            string          `json:"status"`
	EmailsSent        int             `json:"emails_sent"`
	SendTime          string          `json:"send_time"`
	ContentType       string          `json:"content_type"`
	NeedsBlockRefresh bool            `json:"needs_block_refresh"`
	Resendable        bool            `json:"resendable"`
	Recipients        Recipients      `json:"recipients"`
	Settings          Settings        `json:"settings"`
	VariateSettings   VariateSettings `json:"variate_settings"`
	Tracking          Tracking        `json:"tracking"`
	RssOpts           RssOpts         `json:"rss_opts"`
	AbSplitOpts       AbSplitOpts     `json:"ab_split_opts"`
	SocialCard        SocialCard      `json:"social_card"`
	ReportSummary     ReportSummary   `json:"report_summary"`
	DeliveryStatus    DeliveryStatus  `json:"delivery_status"`
	Links             []Links         `json:"_links"`
}

type SegmentOpts struct {
	SavedSegmentID    int    `json:"saved_segment_id,omitempty"`
	PrebuiltSegmentID string `json:"prebuilt_segment_id,omitempty"`
	Match             string `json:"match,omitempty"`
	Conditions        []any  `json:"conditions,omitempty"`
}

type Recipients struct {
	ListID         string      `json:"list_id,omitempty"`
	ListIsActive   bool        `json:"list_is_active,omitempty"`
	ListName       string      `json:"list_name,omitempty"`
	SegmentText    string      `json:"segment_text,omitempty"`
	RecipientCount int         `json:"recipient_count,omitempty"`
	SegmentOpts    SegmentOpts `json:"segment_opts"`
}

type Settings struct {
	SubjectLine     string   `json:"subject_line,omitempty"`
	PreviewText     string   `json:"preview_text,omitempty"`
	Title           string   `json:"title,omitempty"`
	FromName        string   `json:"from_name,omitempty"`
	ReplyTo         string   `json:"reply_to,omitempty"`
	UseConversation bool     `json:"use_conversation,omitempty"`
	ToName          string   `json:"to_name,omitempty"`
	FolderID        string   `json:"folder_id,omitempty"`
	Authenticate    bool     `json:"authenticate,omitempty"`
	AutoFooter      bool     `json:"auto_footer,omitempty"`
	InlineCSS       bool     `json:"inline_css,omitempty"`
	AutoTweet       bool     `json:"auto_tweet,omitempty"`
	AutoFbPost      []string `json:"auto_fb_post,omitempty"`
	FbComments      bool     `json:"fb_comments,omitempty"`
	Timewarp        bool     `json:"timewarp,omitempty"`
	TemplateID      int      `json:"template_id,omitempty"`
	DragAndDrop     bool     `json:"drag_and_drop,omitempty"`
}

type Combinations struct {
	ID                 string `json:"id,omitempty"`
	SubjectLine        int    `json:"subject_line,omitempty"`
	SendTime           int    `json:"send_time,omitempty"`
	FromName           int    `json:"from_name,omitempty"`
	ReplyTo            int    `json:"reply_to,omitempty"`
	ContentDescription int    `json:"content_description,omitempty"`
	Recipients         int    `json:"recipients,omitempty"`
}

type VariateSettings struct {
	WinningCombinationID string         `json:"winning_combination_id,omitempty"`
	WinningCampaignID    string         `json:"winning_campaign_id,omitempty"`
	WinnerCriteria       string         `json:"winner_criteria,omitempty"`
	WaitTime             int            `json:"wait_time,omitempty"`
	TestSize             int            `json:"test_size,omitempty"`
	SubjectLines         []string       `json:"subject_lines,omitempty"`
	SendTimes            []time.Time    `json:"send_times,omitempty"`
	FromNames            []string       `json:"from_names,omitempty"`
	ReplyToAddresses     []string       `json:"reply_to_addresses,omitempty"`
	Contents             []string       `json:"contents,omitempty"`
	Combinations         []Combinations `json:"combinations,omitempty"`
}

type Salesforce struct {
	Campaign bool `json:"campaign,omitempty"`
	Notes    bool `json:"notes,omitempty"`
}

type Capsule struct {
	Notes bool `json:"notes,omitempty"`
}

type Tracking struct {
	Opens           bool       `json:"opens,omitempty"`
	HTMLClicks      bool       `json:"html_clicks,omitempty"`
	TextClicks      bool       `json:"text_clicks,omitempty"`
	GoalTracking    bool       `json:"goal_tracking,omitempty"`
	Ecomm360        bool       `json:"ecomm360,omitempty"`
	GoogleAnalytics string     `json:"google_analytics,omitempty"`
	Clicktale       string     `json:"clicktale,omitempty"`
	Salesforce      Salesforce `json:"salesforce"`
	Capsule         Capsule    `json:"capsule"`
}

type DailySend struct {
	Sunday    bool `json:"sunday,omitempty"`
	Monday    bool `json:"monday,omitempty"`
	Tuesday   bool `json:"tuesday,omitempty"`
	Wednesday bool `json:"wednesday,omitempty"`
	Thursday  bool `json:"thursday,omitempty"`
	Friday    bool `json:"friday,omitempty"`
	Saturday  bool `json:"saturday,omitempty"`
}

type Schedule struct {
	Hour            int       `json:"hour,omitempty"`
	DailySend       DailySend `json:"daily_send"`
	WeeklySendDay   string    `json:"weekly_send_day,omitempty"`
	MonthlySendDate int       `json:"monthly_send_date,omitempty"`
}

type RssOpts struct {
	FeedURL         string    `json:"feed_url,omitempty"`
	Frequency       string    `json:"frequency,omitempty"`
	Schedule        Schedule  `json:"schedule"`
	LastSent        time.Time `json:"last_sent"`
	ConstrainRssImg bool      `json:"constrain_rss_img,omitempty"`
}

type AbSplitOpts struct {
	SplitTest      string    `json:"split_test,omitempty"`
	PickWinner     string    `json:"pick_winner,omitempty"`
	WaitUnits      string    `json:"wait_units,omitempty"`
	WaitTime       int       `json:"wait_time,omitempty"`
	SplitSize      int       `json:"split_size,omitempty"`
	FromNameA      string    `json:"from_name_a,omitempty"`
	FromNameB      string    `json:"from_name_b,omitempty"`
	ReplyEmailA    string    `json:"reply_email_a,omitempty"`
	ReplyEmailB    string    `json:"reply_email_b,omitempty"`
	SubjectA       string    `json:"subject_a,omitempty"`
	SubjectB       string    `json:"subject_b,omitempty"`
	SendTimeA      time.Time `json:"send_time_a"`
	SendTimeB      time.Time `json:"send_time_b"`
	SendTimeWinner string    `json:"send_time_winner,omitempty"`
}

type SocialCard struct {
	ImageURL    string `json:"image_url,omitempty"`
	Description string `json:"description,omitempty"`
	Title       string `json:"title,omitempty"`
}

type Ecommerce struct {
	TotalOrders  int `json:"total_orders,omitempty"`
	TotalSpent   int `json:"total_spent,omitempty"`
	TotalRevenue int `json:"total_revenue,omitempty"`
}

type ReportSummary struct {
	Opens            int       `json:"opens,omitempty"`
	UniqueOpens      int       `json:"unique_opens,omitempty"`
	OpenRate         int       `json:"open_rate,omitempty"`
	Clicks           int       `json:"clicks,omitempty"`
	SubscriberClicks int       `json:"subscriber_clicks,omitempty"`
	ClickRate        int       `json:"click_rate,omitempty"`
	Ecommerce        Ecommerce `json:"ecommerce"`
}

type DeliveryStatus struct {
	Enabled        bool   `json:"enabled,omitempty"`
	CanCancel      bool   `json:"can_cancel,omitempty"`
	Status         string `json:"status,omitempty"`
	EmailsSent     int    `json:"emails_sent,omitempty"`
	EmailsCanceled int    `json:"emails_canceled,omitempty"`
}

type Links struct {
	Rel          string `json:"rel,omitempty"`
	Href         string `json:"href,omitempty"`
	Method       string `json:"method,omitempty"`
	TargetSchema string `json:"targetSchema,omitempty"`
	Schema       string `json:"schema,omitempty"`
}

type PutCampaignRequest struct {
	PlainText       string    `json:"plain_text,omitempty"`
	HTML            string    `json:"html,omitempty"`
	URL             string    `json:"url,omitempty"`
	Template        *Template `json:"template,omitempty"`
	Archive         *Archive  `json:"archive,omitempty"`
	VariateContents []any     `json:"variate_contents,omitempty"`
}

type Template struct {
	ID       int               `json:"id,omitempty"`
	Sections map[string]string `json:"sections,omitempty"`
}

type Archive struct {
	ArchiveContent string `json:"archive_content,omitempty"`
	ArchiveType    string `json:"archive_type,omitempty"`
}

type PutCampaignResponse struct {
	VariateContents []VariateContents `json:"variate_contents"`
	PlainText       string            `json:"plain_text"`
	HTML            string            `json:"html"`
	ArchiveHTML     string            `json:"archive_html"`
	Links           []Links           `json:"_links"`
}

type VariateContents struct {
	ContentLabel string `json:"content_label"`
	PlainText    string `json:"plain_text"`
	HTML         string `json:"html"`
}

type TestEmailRequest struct {
	TestEmails []string `json:"test_emails"`
	SendType   string   `json:"send_type"`
}
