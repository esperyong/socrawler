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
