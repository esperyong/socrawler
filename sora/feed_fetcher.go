package sora

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	FeedURL = "https://sora.chatgpt.com/backend/public/nf2/feed"
)

// FeedFetcher fetches and parses the Sora feed
type FeedFetcher struct {
	page *rod.Page
}

// NewFeedFetcher creates a new feed fetcher
func NewFeedFetcher(page *rod.Page) *FeedFetcher {
	return &FeedFetcher{page: page}
}

// FetchFeed fetches the feed from the endpoint using browser
func (f *FeedFetcher) FetchFeed(ctx context.Context) (*FeedResponse, error) {
	logrus.Infof("Fetching feed from: %s", FeedURL)

	// Create context with timeout
	fetchCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	page := f.page.Context(fetchCtx)

	// Navigate to feed URL
	if err := page.Navigate(FeedURL); err != nil {
		return nil, errors.Wrap(err, "failed to navigate to feed URL")
	}

	// Wait for page to load
	if err := page.WaitLoad(); err != nil {
		return nil, errors.Wrap(err, "failed to wait for page load")
	}

	// Wait a moment for content to render
	time.Sleep(2 * time.Second)

	// Get the page content
	html, err := page.HTML()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get page HTML")
	}

	logrus.Debugf("Received HTML content, length: %d bytes", len(html))

	// Try to extract JSON from <pre> tag or body
	jsonContent, err := f.extractJSON(page)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract JSON from page")
	}

	logrus.Debugf("Extracted JSON content, length: %d bytes", len(jsonContent))

	// Parse JSON into FeedResponse
	var feedResponse FeedResponse
	if err := json.Unmarshal([]byte(jsonContent), &feedResponse); err != nil {
		return nil, errors.Wrap(err, "failed to parse feed JSON")
	}

	logrus.Infof("Successfully fetched feed: %d items", len(feedResponse.Items))

	return &feedResponse, nil
}

// extractJSON extracts JSON content from the page
func (f *FeedFetcher) extractJSON(page *rod.Page) (string, error) {
	// Try to get content from <pre> tag first (common for JSON endpoints)
	preContent, err := page.Eval(`() => {
		const pre = document.querySelector('pre');
		if (pre) {
			return pre.textContent || pre.innerText;
		}
		return null;
	}`)

	if err == nil && preContent.Value.Str() != "" {
		logrus.Debug("Found JSON in <pre> tag")
		return preContent.Value.Str(), nil
	}

	// Try to get from body
	bodyContent, err := page.Eval(`() => {
		return document.body.textContent || document.body.innerText;
	}`)

	if err != nil {
		return "", errors.Wrap(err, "failed to extract body content")
	}

	bodyText := bodyContent.Value.Str()
	if bodyText == "" {
		return "", errors.New("page body is empty")
	}

	logrus.Debug("Extracted JSON from body")
	return bodyText, nil
}

// ValidateFeedResponse validates the feed response
func ValidateFeedResponse(feed *FeedResponse) error {
	if feed == nil {
		return errors.New("feed response is nil")
	}

	if len(feed.Items) == 0 {
		return errors.New("feed has no items")
	}

	// Check first few items for validity
	validItems := 0
	for i, item := range feed.Items {
		if i >= 5 { // Only check first 5 items
			break
		}

		if item.Post.ID == "" {
			logrus.Warnf("Item %d has empty post ID", i)
			continue
		}

		if len(item.Post.Attachments) == 0 {
			logrus.Debugf("Item %d (post_id=%s) has no attachments", i, item.Post.ID)
			continue
		}

		validItems++
	}

	if validItems == 0 {
		return errors.New("no valid items found in feed")
	}

	logrus.Debugf("Feed validation passed: %d/%d items checked are valid", validItems, len(feed.Items))
	return nil
}

// SaveFeedToFile saves a feed response to a JSON file
func SaveFeedToFile(feed *FeedResponse, filePath string) error {
	// Marshal feed to JSON with indentation
	data, err := json.MarshalIndent(feed, "", "    ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal feed to JSON")
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return errors.Wrap(err, "failed to write feed to file")
	}

	logrus.Infof("Feed saved to file: %s (%d bytes)", filePath, len(data))
	return nil
}

// LoadFeedFromFile loads a feed response from a JSON file
func LoadFeedFromFile(filePath string) (*FeedResponse, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read feed file")
	}

	// Unmarshal JSON
	var feed FeedResponse
	if err := json.Unmarshal(data, &feed); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal feed JSON")
	}

	logrus.Infof("Feed loaded from file: %s (%d items)", filePath, len(feed.Items))
	return &feed, nil
}
