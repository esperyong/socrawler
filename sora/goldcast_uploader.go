package sora

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// GoldcastConfig contains configuration for Goldcast API
type GoldcastConfig struct {
	APIKey  string
	BaseURL string
}

// GoldcastUser represents user information for upload
type GoldcastUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// GoldcastUploadRequest represents the request payload for Goldcast upload API
type GoldcastUploadRequest struct {
	MediaURL    string       `json:"media_url"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	User        GoldcastUser `json:"user"`
}

// GoldcastUploadResponse represents the response from Goldcast upload API
type GoldcastUploadResponse struct {
	Success       bool        `json:"success"`
	TaskID        string      `json:"task_id,omitempty"`
	FriendlyToken string      `json:"friendly_token,omitempty"`
	StatusURL     string      `json:"status_url,omitempty"`
	MediaURL      string      `json:"media_url,omitempty"`
	Error         interface{} `json:"error,omitempty"` // Can be string or object
	Message       string      `json:"message,omitempty"`
}

// GoldcastUploader handles uploading videos to Goldcast CMS
type GoldcastUploader struct {
	config      *GoldcastConfig
	db          *VideoDatabase
	client      *http.Client
	ossUploader *OSSUploader
}

// NewGoldcastUploader creates a new Goldcast uploader
func NewGoldcastUploader(config *GoldcastConfig, db *VideoDatabase, ossConfig *OSSConfig) (*GoldcastUploader, error) {
	// Initialize OSS uploader
	ossUploader, err := NewOSSUploader(ossConfig, db)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize OSS uploader")
	}

	return &GoldcastUploader{
		config:      config,
		db:          db,
		ossUploader: ossUploader,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// NewGoldcastConfigFromEnv creates a GoldcastConfig from environment variables with fallback to defaults
func NewGoldcastConfigFromEnv(apiKey, apiURL string) *GoldcastConfig {
	// Check environment variables first
	if envKey := os.Getenv("GOLDCAST_API_KEY"); envKey != "" && apiKey == "" {
		apiKey = envKey
	}
	if envURL := os.Getenv("GOLDCAST_API_URL"); envURL != "" && apiURL == "" {
		apiURL = envURL
	}

	// Use provided values or defaults
	if apiKey == "" {
		apiKey = "ucHZBRJ1.w8njpEorJlDgjp0ESnw0qSyOkN6V6VUe"
	}
	if apiURL == "" {
		apiURL = "https://financial.xiaoyequ9.com/api/v1/external/media/upload"
	}

	return &GoldcastConfig{
		APIKey:  apiKey,
		BaseURL: apiURL,
	}
}

// generateTitle creates a title from text, truncating to 100 chars if necessary
func generateTitle(text, postID string) string {
	// Trim whitespace
	text = strings.TrimSpace(text)

	if text == "" {
		return fmt.Sprintf("Sora Video - %s", postID)
	}

	// Truncate to 100 characters
	if len(text) <= 100 {
		return text
	}

	// Find a good truncation point (try to break at word boundary)
	truncated := text[:97]
	return truncated + "..."
}

// UploadToGoldcast uploads a single video to Goldcast
func (gu *GoldcastUploader) UploadToGoldcast(ctx context.Context, record *VideoRecord) error {
	// Ensure video has OSS URL (upload if necessary)
	ossURL, err := gu.ossUploader.EnsureOSSURL(ctx, record)
	if err != nil {
		return errors.Wrap(err, "failed to ensure OSS URL")
	}

	// Prepare request payload using OSS URL
	title := generateTitle(record.Text, record.PostID)
	description := record.Text

	// If description is empty, use a default message
	if description == "" {
		description = fmt.Sprintf("Sora generated video - %s", record.PostID)
	}

	req := &GoldcastUploadRequest{
		MediaURL:    ossURL, // Use OSS URL instead of original Sora URL
		Title:       title,
		Description: description,
		User: GoldcastUser{
			Username: "UID75203008801597",
			Email:    "api@example.com",
			Name:     "API User",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request")
	}

	// Log the request payload for debugging
	logrus.Debugf("Goldcast upload request: post_id=%s, title='%s', description='%s', oss_url=%s",
		record.PostID, title, description, ossURL)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", gu.config.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", gu.config.APIKey))

	logrus.Debugf("Uploading video to Goldcast: post_id=%s, oss_url=%s, title=%s",
		record.PostID, ossURL, title)

	// Send request
	resp, err := gu.client.Do(httpReq)
	if err != nil {
		return errors.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()

	// Read response body for logging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	// Log response for debugging
	logrus.Debugf("Goldcast API response: status=%d, body=%s", resp.StatusCode, string(bodyBytes))

	// Parse response
	var uploadResp GoldcastUploadResponse
	if err := json.Unmarshal(bodyBytes, &uploadResp); err != nil {
		logrus.Errorf("Failed to parse Goldcast response: %v, raw response: %s", err, string(bodyBytes))
		return errors.Wrap(err, "failed to decode response")
	}

	// Check for errors in response
	if !uploadResp.Success {
		var errorMsg string

		// Handle error field (can be string or object)
		switch v := uploadResp.Error.(type) {
		case string:
			errorMsg = v
		case map[string]interface{}:
			// Try to extract error message from object
			if msg, ok := v["message"].(string); ok {
				errorMsg = msg
			} else {
				// Convert entire error object to string
				errorBytes, _ := json.Marshal(v)
				errorMsg = string(errorBytes)
			}
		default:
			if uploadResp.Error != nil {
				errorMsg = fmt.Sprintf("%v", uploadResp.Error)
			}
		}

		// Fallback to Message field if error is empty
		if errorMsg == "" {
			errorMsg = uploadResp.Message
		}

		// Log detailed error information
		logrus.Errorf("Goldcast upload failed: success=%v, error=%v, message=%s, response=%s",
			uploadResp.Success, uploadResp.Error, uploadResp.Message, string(bodyBytes))

		return errors.Errorf("upload failed: %s", errorMsg)
	}

	// Save Goldcast token if available
	if uploadResp.FriendlyToken != "" {
		if err := gu.db.UpdateGoldcastToken(record.PostID, uploadResp.FriendlyToken); err != nil {
			logrus.Warnf("Failed to save Goldcast token: %v", err)
			// Don't fail the upload if we can't save the token
		} else {
			logrus.Infof("Successfully uploaded video to Goldcast: post_id=%s, token=%s, url=/view?m=%s",
				record.PostID, uploadResp.FriendlyToken, uploadResp.FriendlyToken)
		}
	} else {
		logrus.Infof("Successfully uploaded video to Goldcast: post_id=%s, title=%s (no token returned)",
			record.PostID, title)
	}

	return nil
}

// UploadUnuploadedVideos uploads all videos that haven't been uploaded yet
func (gu *GoldcastUploader) UploadUnuploadedVideos(ctx context.Context, limit int) (*GoldcastUploadResult, error) {
	startTime := time.Now()

	// Get unuploaded videos
	videos, err := gu.db.GetUnuploadedVideos(limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get unuploaded videos")
	}

	result := &GoldcastUploadResult{
		TotalUnuploaded: len(videos),
		Attempted:       0,
		Succeeded:       0,
		Failed:          0,
		FailedPostIDs:   []string{},
	}

	if len(videos) == 0 {
		logrus.Info("No unuploaded videos found")
		return result, nil
	}

	logrus.Infof("Found %d unuploaded videos, starting upload...", len(videos))

	// Upload each video
	for i, video := range videos {
		result.Attempted++

		logrus.Infof("Uploading video %d/%d: post_id=%s, text=%s",
			i+1, len(videos), video.PostID, truncateString(video.Text, 50))

		// Upload to Goldcast
		err := gu.UploadToGoldcast(ctx, video)
		if err != nil {
			logrus.Errorf("Failed to upload video %s: %v", video.PostID, err)
			result.Failed++
			result.FailedPostIDs = append(result.FailedPostIDs, video.PostID)
			continue
		}

		// Mark as uploaded in database
		if err := gu.db.MarkVideoAsUploaded(video.PostID); err != nil {
			logrus.Errorf("Failed to mark video as uploaded %s: %v", video.PostID, err)
			result.Failed++
			result.FailedPostIDs = append(result.FailedPostIDs, video.PostID)
			continue
		}

		result.Succeeded++

		// Small delay to avoid overwhelming the API
		if i < len(videos)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	result.DurationSeconds = int(time.Since(startTime).Seconds())

	logrus.Infof("Upload completed: succeeded=%d, failed=%d, duration=%ds",
		result.Succeeded, result.Failed, result.DurationSeconds)

	return result, nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
