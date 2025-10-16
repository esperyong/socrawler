package sora

import (
	"context"
	"time"

	"github.com/esperyong/socrawler/browser"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// FeedDownloader downloads videos from the feed
type FeedDownloader struct {
	db              *VideoDatabase
	mediaDownloader *MediaDownloader
	savePath        string
}

// NewFeedDownloader creates a new feed downloader
func NewFeedDownloader(dbPath string, savePath string) (*FeedDownloader, error) {
	// Initialize database
	db, err := NewVideoDatabase(dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize database")
	}

	// Initialize media downloader
	mediaDownloader := NewMediaDownloader(savePath)

	return &FeedDownloader{
		db:              db,
		mediaDownloader: mediaDownloader,
		savePath:        savePath,
	}, nil
}

// Download executes the feed download process
func (fd *FeedDownloader) Download(ctx context.Context, req *FeedDownloadRequest) (*FeedDownloadResult, error) {
	return fd.DownloadWithFeed(ctx, req, nil)
}

// DownloadWithFeed executes the feed download process with optional pre-loaded feed
func (fd *FeedDownloader) DownloadWithFeed(ctx context.Context, req *FeedDownloadRequest, feed *FeedResponse) (*FeedDownloadResult, error) {
	startTime := time.Now()
	logrus.Infof("Starting feed download: save_path=%s, db_path=%s, limit=%d, headless=%v",
		req.SavePath, req.DBPath, req.Limit, req.Headless)

	// Fetch feed if not provided
	if feed == nil {
		// Create browser instance
		b := browser.NewCleanBrowser(req.Headless)
		defer b.Close()

		page := b.NewPage()
		defer page.Close()

		// Fetch feed
		fetcher := NewFeedFetcher(page)
		var err error
		feed, err = fetcher.FetchFeed(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch feed")
		}

		// Validate feed
		if err := ValidateFeedResponse(feed); err != nil {
			return nil, errors.Wrap(err, "feed validation failed")
		}
	}

	logrus.Infof("Fetched %d items from feed", len(feed.Items))

	// Get existing post IDs from database
	existingPostIDs, err := fd.db.GetExistingPostIDs()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get existing post IDs")
	}

	logrus.Infof("Found %d existing videos in database", len(existingPostIDs))

	// Filter new items
	newItems := fd.filterNewItems(feed.Items, existingPostIDs, req.Limit)
	logrus.Infof("Found %d new videos to download", len(newItems))

	if len(newItems) == 0 {
		logrus.Info("No new videos to download")
		elapsed := time.Since(startTime)
		return &FeedDownloadResult{
			TotalFetched:    len(feed.Items),
			NewVideos:       0,
			Downloaded:      0,
			Skipped:         0,
			Failed:          0,
			VideoPaths:      []string{},
			ThumbnailPaths:  []string{},
			DurationSeconds: int(elapsed.Seconds()),
		}, nil
	}

	// Download videos
	result := fd.downloadVideos(newItems)

	elapsed := time.Since(startTime)
	result.TotalFetched = len(feed.Items)
	result.DurationSeconds = int(elapsed.Seconds())

	logrus.Infof("Download completed: fetched=%d, new=%d, downloaded=%d, skipped=%d, failed=%d, duration=%ds",
		result.TotalFetched, result.NewVideos, result.Downloaded, result.Skipped, result.Failed, result.DurationSeconds)

	return result, nil
}

// filterNewItems filters feed items to only include new videos
func (fd *FeedDownloader) filterNewItems(items []FeedItem, existingPostIDs map[string]bool, limit int) []FeedItem {
	var newItems []FeedItem

	for _, item := range items {
		// Skip if already exists
		if existingPostIDs[item.Post.ID] {
			continue
		}

		// Skip if no attachments
		if len(item.Post.Attachments) == 0 {
			logrus.Debugf("Skipping post %s: no attachments", item.Post.ID)
			continue
		}

		// Only process "sora" type attachments
		hasSoraVideo := false
		for _, attachment := range item.Post.Attachments {
			if attachment.Kind == "sora" && attachment.DownloadableURL != "" {
				hasSoraVideo = true
				break
			}
		}

		if !hasSoraVideo {
			logrus.Debugf("Skipping post %s: no sora video attachment", item.Post.ID)
			continue
		}

		newItems = append(newItems, item)

		// Check limit
		if limit > 0 && len(newItems) >= limit {
			logrus.Infof("Reached download limit of %d videos", limit)
			break
		}
	}

	return newItems
}

// downloadVideos downloads videos and saves metadata
func (fd *FeedDownloader) downloadVideos(items []FeedItem) *FeedDownloadResult {
	result := &FeedDownloadResult{
		NewVideos:      len(items),
		Downloaded:     0,
		Skipped:        0,
		Failed:         0,
		VideoPaths:     []string{},
		ThumbnailPaths: []string{},
	}

	for i, item := range items {
		logrus.Infof("Processing video %d/%d: post_id=%s, username=%s",
			i+1, len(items), item.Post.ID, item.Profile.Username)

		// Find the first sora video attachment
		var videoAttachment *Attachment
		for j := range item.Post.Attachments {
			if item.Post.Attachments[j].Kind == "sora" && item.Post.Attachments[j].DownloadableURL != "" {
				videoAttachment = &item.Post.Attachments[j]
				break
			}
		}

		if videoAttachment == nil {
			logrus.Warnf("No valid video attachment found for post %s", item.Post.ID)
			result.Skipped++
			continue
		}

		// Download video using post_id as folder name
		videoPath, err := fd.mediaDownloader.DownloadMediaForFeed(videoAttachment.DownloadableURL, item.Post.ID, MediaTypeVideo)
		if err != nil {
			logrus.Errorf("Failed to download video for post %s: %v", item.Post.ID, err)
			result.Failed++
			continue
		}

		result.VideoPaths = append(result.VideoPaths, videoPath)

		// Download thumbnail using post_id as folder name
		thumbnailPath := ""
		if videoAttachment.Encodings.Thumbnail.Path != "" {
			thumbnailPath, err = fd.mediaDownloader.DownloadMediaForFeed(videoAttachment.Encodings.Thumbnail.Path, item.Post.ID, MediaTypeThumbnail)
			if err != nil {
				logrus.Warnf("Failed to download thumbnail for post %s: %v", item.Post.ID, err)
				// Don't fail the entire download if thumbnail fails
			} else {
				result.ThumbnailPaths = append(result.ThumbnailPaths, thumbnailPath)
			}
		}

		// Save metadata to database
		record := &VideoRecord{
			PostID:             item.Post.ID,
			GenerationID:       videoAttachment.GenerationID,
			VideoURL:           videoAttachment.DownloadableURL,
			ThumbnailURL:       videoAttachment.Encodings.Thumbnail.Path,
			Text:               item.Post.Text,
			Username:           item.Profile.Username,
			UserID:             item.Profile.UserID,
			PostedAt:           item.Post.PostedAt,
			DownloadedAt:       time.Now(),
			LocalVideoPath:     videoPath,
			LocalThumbnailPath: thumbnailPath,
			Width:              videoAttachment.Width,
			Height:             videoAttachment.Height,
		}

		if err := fd.db.InsertVideo(record); err != nil {
			logrus.Errorf("Failed to save metadata for post %s: %v", item.Post.ID, err)
			// Don't fail the download, just log the error
		}

		result.Downloaded++
		logrus.Infof("Successfully downloaded video %d/%d: %s", i+1, len(items), videoPath)
	}

	return result
}

// Close closes the downloader and releases resources
func (fd *FeedDownloader) Close() error {
	if fd.db != nil {
		return fd.db.Close()
	}
	return nil
}

// DownloadFromFeed is a convenience function to download from feed
func DownloadFromFeed(ctx context.Context, req *FeedDownloadRequest) (*FeedDownloadResult, error) {
	// Set defaults
	if req.SavePath == "" {
		req.SavePath = "./downloads/sora"
	}
	if req.DBPath == "" {
		req.DBPath = "./sora.db"
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	// Create downloader
	downloader, err := NewFeedDownloader(req.DBPath, req.SavePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create feed downloader")
	}
	defer downloader.Close()

	// Execute download
	result, err := downloader.Download(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "download failed")
	}

	return result, nil
}

// FetchFeedToFile fetches the feed and saves it to a file
func FetchFeedToFile(ctx context.Context, outputPath string, headless bool) error {
	logrus.Infof("Fetching feed to file: %s", outputPath)

	// Create browser instance
	b := browser.NewCleanBrowser(headless)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	// Fetch feed
	fetcher := NewFeedFetcher(page)
	feed, err := fetcher.FetchFeed(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to fetch feed")
	}

	// Validate feed
	if err := ValidateFeedResponse(feed); err != nil {
		return errors.Wrap(err, "feed validation failed")
	}

	logrus.Infof("Fetched %d items from feed", len(feed.Items))

	// Save to file
	if err := SaveFeedToFile(feed, outputPath); err != nil {
		return errors.Wrap(err, "failed to save feed to file")
	}

	logrus.Infof("Feed saved to: %s", outputPath)
	return nil
}

// DownloadFromFile downloads videos from a saved feed file
func DownloadFromFile(ctx context.Context, feedPath string, req *FeedDownloadRequest) (*FeedDownloadResult, error) {
	logrus.Infof("Loading feed from file: %s", feedPath)

	// Load feed from file
	feed, err := LoadFeedFromFile(feedPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load feed from file")
	}

	logrus.Infof("Loaded %d items from feed file", len(feed.Items))

	// Set defaults
	if req.SavePath == "" {
		req.SavePath = "./downloads/sora"
	}
	if req.DBPath == "" {
		req.DBPath = "./sora.db"
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	// Create downloader
	downloader, err := NewFeedDownloader(req.DBPath, req.SavePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create feed downloader")
	}
	defer downloader.Close()

	// Execute download with pre-loaded feed
	result, err := downloader.DownloadWithFeed(ctx, req, feed)
	if err != nil {
		return nil, errors.Wrap(err, "download failed")
	}

	return result, nil
}
