package arcjson

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spotlightpa/almanack/internal/errutil"
)

const (
	StatusWorking    = 1
	StatusAssigning  = 2
	StatusSecondEdit = 3
	StatusRim        = 4
	StatusSlot       = 5
	StatusDone       = 6
)

type API struct {
	Version  string     `json:"apiVersion"`
	Contents []Contents `json:"contents"`
}

type Contents struct {
	AdditionalProperties ContentProperties  `json:"additional_properties"`
	CanonicalURL         string             `json:"canonical_url"`
	CanonicalWebsite     string             `json:"canonical_website"`
	Comments             Comments           `json:"comments"`
	ContentElements      []*json.RawMessage `json:"content_elements"`
	CreatedDate          time.Time          `json:"created_date"`
	Credits              Credits            `json:"credits"`
	Description          Description        `json:"description"`
	DisplayDate          time.Time          `json:"display_date,omitempty"`
	Distributor          Distributor        `json:"distributor"`
	FirstPublishDate     time.Time          `json:"first_publish_date,omitempty"`
	Headlines            Headlines          `json:"headlines"`
	ID                   string             `json:"_id"`
	Label                Label              `json:"label"`
	Language             string             `json:"language"`
	LastUpdatedDate      time.Time          `json:"last_updated_date"`
	Owner                Owner              `json:"owner"`
	Planning             Planning           `json:"planning"`
	PromoItems           PromoItems         `json:"promo_items"`
	PublishDate          time.Time          `json:"publish_date,omitempty"`
	Publishing           Publishing         `json:"publishing"`
	Slug                 string             `json:"slug"`
	Source               Source             `json:"source"`
	Subheadlines         Subheadlines       `json:"subheadlines"`
	Subtype              string             `json:"subtype"`
	Syndication          Syndication        `json:"syndication"`
	Type                 string             `json:"type"`
	Version              string             `json:"version"`
	Website              string             `json:"website"`
	WebsiteURL           string             `json:"website_url,omitempty"`
	Workflow             Workflow           `json:"workflow,omitempty"`
}

type ContentProperties struct {
	HasPublishedCopy bool            `json:"has_published_copy"`
	IsPublished      bool            `json:"is_published"`
	PublishDate      json.RawMessage `json:"publish_date"`
}

type _contentElement struct {
	ID        string            `json:"_id"`
	Type      string            `json:"type"`
	Content   string            `json:"content"`
	Caption   string            `json:"caption"`
	Items     []_contentElement `json:"items,omitempty"`
	Level     int               `json:"level"`
	ListType  string            `json:"list_type"`
	Owner     Owner             `json:"owner"`
	RawOembed RawOembed         `json:"raw_oembed"`
	URL       string            `json:"url"`
	Width     int               `json:"width"`
}

type ContentElementType struct {
	Type *string `json:"type"`
}

type ContentElementText struct {
	Content *string `json:"content"`
}

type ContentElementHeading struct {
	Content string `json:"content"`
	Level   int    `json:"level"`
}

type ContentElementImage struct {
	Credits Credits `json:"credits"`
	Caption string  `json:"caption"`
	URL     string  `json:"url"`
	Width   int     `json:"width"`
}

type ContentElementList struct {
	Items    []struct{ Type, Content string } `json:"items"`
	ListType string                           `json:"list_type"`
}

type ContentElementOembed struct {
	RawOembed RawOembed `json:"raw_oembed"`
}

type RawOembed struct {
	ID           string `json:"_id"`
	AuthorName   string `json:"author_name"`
	AuthorURL    string `json:"author_url"`
	CacheAge     string `json:"cache_age"`
	HTML         string `json:"html"`
	ProviderName string `json:"provider_name"`
	ProviderURL  string `json:"provider_url"`
	Type         string `json:"type"`
	URL          string `json:"url"`
	Version      string `json:"version"`
	Width        int    `json:"width"`
}

type Headlines struct {
	Basic     string `json:"basic"`
	Mobile    string `json:"mobile"`
	Native    string `json:"native"`
	Print     string `json:"print"`
	Tablet    string `json:"tablet"`
	Web       string `json:"web"`
	MetaTitle string `json:"meta_title"`
}

type Owner struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Sponsored bool   `json:"sponsored"`
}

type Comments struct {
	AllowComments   bool `json:"allow_comments"`
	DisplayComments bool `json:"display_comments"`
}

type Workflow struct {
	StatusCode int    `json:"status_code"`
	Note       string `json:"note"`
}

type Syndication struct {
	ExternalDistribution bool `json:"external_distribution"`
	Search               bool `json:"search"`
}

type Subheadlines struct {
	Basic string `json:"basic"`
}

type Description struct {
	Basic string `json:"basic"`
}

type Source struct {
	System     string `json:"system"`
	Name       string `json:"name"`
	SourceType string `json:"source_type"`
}

type Eyebrows struct {
	Text    string `json:"text"`
	URL     string `json:"url"`
	Display bool   `json:"display"`
}

type Label struct {
	Eyebrows Eyebrows `json:"eyebrows"`
}

type Distributor struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
}
type Scheduling struct {
	PlannedPublishDate time.Time `json:"planned_publish_date"`
}
type StoryLength struct {
	WordCountActual  int `json:"word_count_actual"`
	LineCountActual  int `json:"line_count_actual"`
	InchCountActual  int `json:"inch_count_actual"`
	WordCountPlanned int `json:"word_count_planned"`
}
type Planning struct {
	Scheduling   Scheduling  `json:"scheduling"`
	InternalNote string      `json:"internal_note"`
	StoryLength  StoryLength `json:"story_length"`
	BudgetLine   string      `json:"budget_line"`
}
type Image struct {
	URL     string `json:"url"`
	Version string `json:"version"`
}

type SocialLinks struct {
	Site string `json:"site"`
	URL  string `json:"url"`
}

type By struct {
	ID          string        `json:"_id"`
	Type        string        `json:"type"`
	Version     string        `json:"version"`
	Name        string        `json:"name"`
	Image       Image         `json:"image"`
	Description string        `json:"description"`
	URL         string        `json:"url"`
	Slug        string        `json:"slug"`
	SocialLinks []SocialLinks `json:"social_links"`
}

type Credits struct {
	By []By `json:"by"`
}

type ScheduledOperations struct {
	PublishEdition   []interface{} `json:"publish_edition"`
	UnpublishEdition []interface{} `json:"unpublish_edition"`
}

type Publishing struct {
	ScheduledOperations ScheduledOperations `json:"scheduled_operations"`
}

type Revision struct {
	RevisionID string        `json:"revision_id"`
	ParentID   string        `json:"parent_id"`
	Editions   []interface{} `json:"editions"`
	Branch     string        `json:"branch"`
	UserID     string        `json:"user_id"`
}

type Taxonomy struct {
	SeoKeywords []string `json:"seo_keywords"`
}

type PromoItems struct {
	Basic struct {
		ID                   string `json:"_id"`
		AdditionalProperties struct {
			FullSizeResizeURL string        `json:"fullSizeResizeUrl"`
			Galleries         []interface{} `json:"galleries"`
			MimeType          string        `json:"mime_type"`
			OriginalName      string        `json:"originalName"`
			OriginalURL       string        `json:"originalUrl"`
			ProxyURL          string        `json:"proxyUrl"`
			Published         bool          `json:"published"`
			ResizeURL         string        `json:"resizeUrl"`
			Restricted        bool          `json:"restricted"`
			TakenOn           time.Time     `json:"takenOn"`
			Version           int           `json:"version"`
		} `json:"additional_properties"`
		Caption     string      `json:"caption"`
		CreatedDate time.Time   `json:"created_date"`
		Credits     PromoCredit `json:"credits"`
		Distributor struct {
			AdditionalProperties struct {
			} `json:"additional_properties"`
			Category string `json:"category"`
			Name     string `json:"name"`
		} `json:"distributor"`
		Height          int       `json:"height"`
		LastUpdatedDate time.Time `json:"last_updated_date"`
		Licensable      bool      `json:"licensable"`
		Owner           struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"owner"`
		Source struct {
			AdditionalProperties struct {
				ClickabilityID string `json:"clickability_id"`
			} `json:"additional_properties"`
			Name       string `json:"name"`
			SourceID   string `json:"source_id"`
			SourceType string `json:"source_type"`
			System     string `json:"system"`
		} `json:"source"`
		Type    string `json:"type"`
		URL     string `json:"url"`
		Version string `json:"version"`
		Width   int    `json:"width"`
	} `json:"basic"`
}

type PromoCredit struct {
	Affiliation []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"affiliation"`
	By []struct {
		Byline string `json:"byline"`
		Name   string `json:"name"`
		Type   string `json:"type"`
	} `json:"by"`
}

func (feed API) Get(id string) (*Contents, error) {
	found := -1
	for i, article := range feed.Contents {
		if article.ID == id {
			if found != -1 {
				return nil, fmt.Errorf("multiple matching IDs found")
			}
			found = i
		}
	}
	if found == -1 {
		return nil, errutil.NotFound
	}
	return &feed.Contents[found], nil
}
