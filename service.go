package main

import (
	"context"

	"github.com/esperyong/socrawler/browser"
	"github.com/esperyong/socrawler/configs"
	"github.com/esperyong/socrawler/sora"
)

// SoraService handles Sora video crawling operations
type SoraService struct{}

// NewSoraService creates a new SoraService instance
func NewSoraService() *SoraService {
	return &SoraService{}
}

// CrawlSoraVideos crawls Sora videos and thumbnails
func (s *SoraService) CrawlSoraVideos(ctx context.Context, req *SoraCrawlRequest) (*SoraCrawlResponse, error) {
	// Set default values
	if req.TotalDurationSeconds <= 0 {
		req.TotalDurationSeconds = 300 // Default 5 minutes
	}
	if req.ScrollIntervalSeconds <= 0 {
		req.ScrollIntervalSeconds = 20 // Default 20 seconds
	}
	if req.SavePath == "" {
		req.SavePath = "./downloads/sora"
	}

	// Create a clean browser instance (no cookies needed for Sora)
	b := browser.NewCleanBrowser(configs.IsHeadless())
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	// Create crawler instance
	crawler := sora.NewCrawlerAction(page)

	// Convert request parameters to internal format
	crawlReq := &sora.CrawlRequest{
		TotalDurationSeconds:  req.TotalDurationSeconds,
		ScrollIntervalSeconds: req.ScrollIntervalSeconds,
		SavePath:              req.SavePath,
	}

	// Start crawling
	result, err := crawler.StartCrawl(ctx, crawlReq)
	if err != nil {
		return nil, err
	}

	// Convert response to API format
	response := &SoraCrawlResponse{
		Videos:          result.Videos,
		Thumbnails:      result.Thumbnails,
		TotalVideos:     result.TotalVideos,
		TotalThumbnails: result.TotalThumbnails,
		DurationSeconds: result.DurationSeconds,
	}

	return response, nil
}
