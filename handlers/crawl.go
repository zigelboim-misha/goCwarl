package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mishazigelboim/gocrawl/k8s"
	"github.com/mishazigelboim/gocrawl/models"
)

type CrawlHandler struct {
	podManager *k8s.PodManager
	namespace  string
}

func NewCrawlHandler(namespace string) (*CrawlHandler, error) {
	pm, err := k8s.NewPodManager(namespace)
	if err != nil {
		return nil, err
	}

	return &CrawlHandler{
		podManager: pm,
		namespace:  namespace,
	}, nil
}

// CrawlModel handles the crawl request
// @Summary Crawl an AI model page
// @Description Creates a temporary crawl4ai pod, crawls the OpenRouter model page, and returns filtered results
// @Tags crawl
// @Accept json
// @Produce json
// @Param request body models.CrawlRequest true "Crawl Request"
// @Success 200 {object} models.CrawlResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/crawl [post]
func (h *CrawlHandler) CrawlModel(c *gin.Context) {
	var req models.CrawlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Validate model format (vendor/model-name)
	if !isValidModelFormat(req.Model) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "model must be in format: vendor/model-name",
		})
		return
	}

	ctx := c.Request.Context()

	// Generate unique pod name
	podName := fmt.Sprintf("crawl4ai-%d", time.Now().Unix())
	fmt.Printf("[%s] Starting crawl request for model: %s\n", podName, req.Model)

	// Create the pod
	fmt.Printf("[%s] Creating pod...\n", podName)
	_, err := h.podManager.CreateCrawlPod(ctx, podName)
	if err != nil {
		fmt.Printf("[%s] ERROR: Failed to create pod: %v\n", podName, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: fmt.Sprintf("failed to create pod: %v", err),
		})
		return
	}
	fmt.Printf("[%s] Pod created successfully\n", podName)

	// Ensure pod is deleted after request completes
	defer func() {
		fmt.Printf("[%s] Cleaning up pod...\n", podName)
		deleteCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := h.podManager.DeletePod(deleteCtx, podName); err != nil {
			// Log error but don't fail the request
			fmt.Printf("[%s] ERROR: Failed to delete pod: %v\n", podName, err)
		} else {
			fmt.Printf("[%s] Pod deleted successfully\n", podName)
		}
	}()

	// Wait for pod to be ready (max 2 minutes)
	fmt.Printf("[%s] Waiting for pod to be ready (max 2 minutes)...\n", podName)
	if err := h.podManager.WaitForPodReady(ctx, podName, 2*time.Minute); err != nil {
		fmt.Printf("[%s] ERROR: Pod failed to become ready: %v\n", podName, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: fmt.Sprintf("pod failed to become ready: %v", err),
		})
		return
	}
	fmt.Printf("[%s] Pod is ready\n", podName)

	// Get pod IP
	podIP, err := h.podManager.GetPodIP(ctx, podName)
	if err != nil {
		fmt.Printf("[%s] ERROR: Failed to get pod IP: %v\n", podName, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: fmt.Sprintf("failed to get pod IP: %v", err),
		})
		return
	}
	fmt.Printf("[%s] Pod IP: %s\n", podName, podIP)

	// Add a delay to allow the service inside the pod to start
	fmt.Printf("[%s] Waiting 10 seconds for crawl4ai service to start...\n", podName)
	time.Sleep(10 * time.Second)

	// Make request to crawl4ai
	fmt.Printf("[%s] Making crawl request to http://%s:11235/crawl\n", podName, podIP)
	crawlResp, err := h.makeCrawlRequest(ctx, podIP, req.Model)
	if err != nil {
		fmt.Printf("[%s] ERROR: Crawl request failed: %v\n", podName, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: fmt.Sprintf("crawl request failed: %v", err),
		})
		return
	}
	fmt.Printf("[%s] Crawl request successful\n", podName)

	// Filter and return response
	if len(crawlResp.Results) == 0 {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "no results returned from crawler",
		})
		return
	}

	result := crawlResp.Results[0]
	c.JSON(http.StatusOK, models.CrawlResponse{
		URL:      result.URL,
		Success:  result.Success,
		Markdown: result.Markdown.RawMarkdown,
	})
}

func (h *CrawlHandler) makeCrawlRequest(ctx context.Context, podIP, model string) (*models.Crawl4AIResponse, error) {
	// Build the crawl4ai request
	crawlReq := models.Crawl4AIRequest{
		URLs: []string{fmt.Sprintf("https://openrouter.ai/%s", model)},
		CrawlerConfig: models.CrawlerConfig{
			Type: "CrawlerRunConfig",
			Params: models.ConfigParams{
				ScrapingStrategy: models.ScrapingStrategy{
					Type:   "LXMLWebScrapingStrategy",
					Params: map[string]interface{}{},
				},
				TableExtraction: models.TableExtraction{
					Type:   "DefaultTableExtraction",
					Params: map[string]interface{}{},
				},
				ExcludeSocialMediaDomains: []string{
					"facebook.com",
					"twitter.com",
					"x.com",
					"linkedin.com",
					"instagram.com",
					"pinterest.com",
					"tiktok.com",
					"snapchat.com",
					"reddit.com",
				},
				ExcludeTags:             []string{"scripts", "style"},
				DelayBeforeReturnHTML:   10.0,
				ExcludeExternalLinks:    true,
				ExcludeSocialMediaLinks: true,
			},
		},
	}

	body, err := json.Marshal(crawlReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("http://%s:11235/crawl", podIP)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crawl4ai returned status: %d", resp.StatusCode)
	}

	var crawlResp models.Crawl4AIResponse
	if err := json.NewDecoder(resp.Body).Decode(&crawlResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &crawlResp, nil
}

func isValidModelFormat(model string) bool {
	// Check for vendor/model-name format
	for i, c := range model {
		if c == '/' {
			// Must have content before and after the slash
			return i > 0 && i < len(model)-1
		}
	}
	return false
}
