package sora

// MediaType 媒体类型
type MediaType string

const (
	MediaTypeVideo     MediaType = "video"
	MediaTypeThumbnail MediaType = "thumbnail"
)

// MediaFile 媒体文件信息
type MediaFile struct {
	URL       string    `json:"url"`
	LocalPath string    `json:"local_path"`
	Type      MediaType `json:"type"`
	Size      int64     `json:"size,omitempty"`
}

// CrawlRequest 爬取请求参数（内部使用）
type CrawlRequest struct {
	TotalDurationSeconds  int    `json:"total_duration_seconds"`
	ScrollIntervalSeconds int    `json:"scroll_interval_seconds"`
	SavePath              string `json:"save_path"`
}

// CrawlResult 爬取结果（内部使用）
type CrawlResult struct {
	Videos          []string `json:"videos"`
	Thumbnails      []string `json:"thumbnails"`
	TotalVideos     int      `json:"total_videos"`
	TotalThumbnails int      `json:"total_thumbnails"`
	DurationSeconds int      `json:"duration_seconds"`
}

// Feed-based types for new feed downloader

// FeedResponse represents the response from the feed endpoint
type FeedResponse struct {
	Items []FeedItem `json:"items"`
}

// FeedItem represents a single item in the feed
type FeedItem struct {
	Post    Post    `json:"post"`
	Profile Profile `json:"profile"`
}

// Post represents a Sora post
type Post struct {
	ID          string       `json:"id"`
	SharedBy    string       `json:"shared_by"`
	PostedAt    float64      `json:"posted_at"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
	Permalink   string       `json:"permalink"`
	LikeCount   int          `json:"like_count"`
	ViewCount   int          `json:"view_count"`
}

// Attachment represents a media attachment (video)
type Attachment struct {
	ID              string    `json:"id"`
	Kind            string    `json:"kind"`
	GenerationID    string    `json:"generation_id"`
	GenerationType  string    `json:"generation_type"`
	URL             string    `json:"url"`
	DownloadableURL string    `json:"downloadable_url"`
	Width           int       `json:"width"`
	Height          int       `json:"height"`
	Encodings       Encodings `json:"encodings"`
}

// Encodings represents different media encodings
type Encodings struct {
	Source    Encoding `json:"source"`
	SourceWM  Encoding `json:"source_wm"`
	Thumbnail Encoding `json:"thumbnail"`
	MD        Encoding `json:"md"`
	GIF       Encoding `json:"gif"`
}

// Encoding represents a single encoding path
type Encoding struct {
	Path string `json:"path"`
}

// Profile represents a user profile
type Profile struct {
	UserID            string `json:"user_id"`
	Username          string `json:"username"`
	DisplayName       string `json:"display_name"`
	ProfilePictureURL string `json:"profile_picture_url"`
	FollowerCount     int    `json:"follower_count"`
	PostCount         int    `json:"post_count"`
	Verified          bool   `json:"verified"`
	Location          string `json:"location"`
	Description       string `json:"description"`
	Permalink         string `json:"permalink"`
}

// FeedDownloadRequest represents a request to download from feed
type FeedDownloadRequest struct {
	SavePath string `json:"save_path"`
	DBPath   string `json:"db_path"`
	Limit    int    `json:"limit"`
	Headless bool   `json:"headless"`
}

// FeedDownloadResult represents the result of feed download
type FeedDownloadResult struct {
	TotalFetched    int      `json:"total_fetched"`
	NewVideos       int      `json:"new_videos"`
	Downloaded      int      `json:"downloaded"`
	Skipped         int      `json:"skipped"`
	Failed          int      `json:"failed"`
	VideoPaths      []string `json:"video_paths"`
	ThumbnailPaths  []string `json:"thumbnail_paths"`
	DurationSeconds int      `json:"duration_seconds"`
}

// ToFeedItem converts a VideoRecord to a FeedItem
func (vr *VideoRecord) ToFeedItem() FeedItem {
	return FeedItem{
		Post: Post{
			ID:        vr.PostID,
			PostedAt:  vr.PostedAt,
			Text:      vr.Text,
			Permalink: "https://sora.chatgpt.com/p/" + vr.PostID,
			Attachments: []Attachment{
				{
					ID:              vr.PostID + "-attachment-0",
					Kind:            "sora",
					GenerationID:    vr.GenerationID,
					GenerationType:  "video_gen",
					URL:             vr.VideoURL,
					DownloadableURL: vr.VideoURL,
					Width:           vr.Width,
					Height:          vr.Height,
					Encodings: Encodings{
						Source: Encoding{
							Path: vr.VideoURL,
						},
						SourceWM: Encoding{
							Path: vr.VideoURL,
						},
						Thumbnail: Encoding{
							Path: vr.ThumbnailURL,
						},
					},
				},
			},
		},
		Profile: Profile{
			UserID:   vr.UserID,
			Username: vr.Username,
		},
	}
}

// VideoRecordsToFeedResponse converts a slice of VideoRecords to a FeedResponse
func VideoRecordsToFeedResponse(records []*VideoRecord) *FeedResponse {
	items := make([]FeedItem, len(records))
	for i, record := range records {
		items[i] = record.ToFeedItem()
	}
	return &FeedResponse{
		Items: items,
	}
}
