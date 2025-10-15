package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// crawlSoraHandler handles the Sora video crawling request
func (s *AppServer) crawlSoraHandler(c *gin.Context) {
	var req SoraCrawlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST",
			"Invalid request parameters", err.Error())
		return
	}

	logrus.Infof("Received Sora crawl request: duration=%ds, scroll_interval=%ds, save_path=%s",
		req.TotalDurationSeconds, req.ScrollIntervalSeconds, req.SavePath)

	// Create service and execute crawl
	service := NewSoraService()
	result, err := service.CrawlSoraVideos(c.Request.Context(), &req)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "CRAWL_FAILED",
			"Failed to crawl Sora videos", err.Error())
		return
	}

	respondSuccess(c, result, "Crawl completed successfully")
}
