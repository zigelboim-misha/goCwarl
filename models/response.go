package models

// CrawlResponse represents the filtered API response
type CrawlResponse struct {
	URL      string `json:"url"`
	Success  bool   `json:"success"`
	Markdown string `json:"markdown"`
}

// Crawl4AIResponse represents the full response from crawl4ai
type Crawl4AIResponse struct {
	Results []CrawlResult `json:"results"`
}

type CrawlResult struct {
	URL      string   `json:"url"`
	Success  bool     `json:"success"`
	Markdown Markdown `json:"markdown"`
}

type Markdown struct {
	RawMarkdown string `json:"raw_markdown"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
