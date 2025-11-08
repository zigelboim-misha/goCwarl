package models

// CrawlRequest represents the API request
type CrawlRequest struct {
	Model string `json:"model" binding:"required" example:"anthropic/claude-3-opus"`
}

// Crawl4AIRequest represents the request to crawl4ai service
type Crawl4AIRequest struct {
	URLs          []string      `json:"urls"`
	CrawlerConfig CrawlerConfig `json:"crawler_config"`
}

type CrawlerConfig struct {
	Type   string       `json:"type"`
	Params ConfigParams `json:"params"`
}

type ConfigParams struct {
	ScrapingStrategy          ScrapingStrategy `json:"scraping_strategy"`
	TableExtraction           TableExtraction  `json:"table_extraction"`
	ExcludeSocialMediaDomains []string         `json:"exclude_social_media_domains"`
	ExcludeTags               []string         `json:"excluded_tags,omitempty"`
	DelayBeforeReturnHTML     float32          `json:"delay_before_return_html,omitempty"`
	ExcludeExternalLinks      bool             `json:"exclude_external_links,omitempty"`
	ExcludeSocialMediaLinks   bool             `json:"exclude_social_media_links,omitempty"`
}

type ScrapingStrategy struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

type TableExtraction struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}
