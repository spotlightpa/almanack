package jsonfeed

type Feed struct {
	Description string `json:"description"`
	FeedURL     string `json:"feed_url"`
	HomePageURL string `json:"home_page_url"`
	Items       []Item `json:"items"`
	Title       string `json:"title"`
	Version     string `json:"version"`
}

type Item struct {
	Author           string   `json:"author"`
	Authors          []string `json:"authors"`
	Category         string   `json:"category"`
	ContentHTML      string   `json:"content_html"`
	DateModified     string   `json:"date_modified"`
	DatePublished    string   `json:"date_published"`
	ID               string   `json:"id"`
	Image            string   `json:"image"`
	ImageCredit      string   `json:"image_credit"`
	ImageDescription string   `json:"image_description"`
	Language         string   `json:"language"`
	Title            string   `json:"title"`
	URL              string   `json:"url"`
}
