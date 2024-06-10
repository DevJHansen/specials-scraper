package internal

type Special struct {
	IsActive     bool
	DownloadLink string
	WebsiteLink  string
	Title        string
	Category     string
	BeenSent     bool
	ScrapingID   *string
	DateAdded    int64
}

type WebsiteCategory struct {
	Title     string
	TabNumber string
}
