package sora

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// MediaDownloader 媒体下载器
type MediaDownloader struct {
	savePath   string
	httpClient *http.Client
}

// NewMediaDownloader 创建媒体下载器
func NewMediaDownloader(savePath string) *MediaDownloader {
	// 确保保存目录存在
	if err := os.MkdirAll(savePath, 0755); err != nil {
		panic(fmt.Sprintf("failed to create save path: %v", err))
	}

	return &MediaDownloader{
		savePath: savePath,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // 视频文件可能较大，增加超时时间
		},
	}
}

// DownloadMedia 下载媒体文件
// 返回本地文件路径
func (d *MediaDownloader) DownloadMedia(mediaURL string, mediaType MediaType) (string, error) {
	// 生成文件夹和文件路径（优先使用task ID）
	folderName := d.generateFolderName(mediaURL)
	folderPath := filepath.Join(d.savePath, folderName)

	// 确保文件夹存在
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create folder")
	}

	// 根据媒体类型确定文件名
	var fileName string
	if mediaType == MediaTypeVideo {
		fileName = "video.mp4"
	} else {
		fileName = "thumbnail.webp"
	}

	filePath := filepath.Join(folderPath, fileName)

	// 去重：如果文件已存在，直接返回路径
	if _, err := os.Stat(filePath); err == nil {
		logrus.Infof("File already exists, skipping download: %s", filePath)
		return filePath, nil
	}

	logrus.Debugf("Downloading %s: %s", mediaType, mediaURL)

	// 下载媒体数据
	resp, err := d.httpClient.Get(mediaURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to download media")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 读取媒体数据
	mediaData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read media data")
	}

	// 保存到文件
	if err := os.WriteFile(filePath, mediaData, 0644); err != nil {
		return "", errors.Wrap(err, "failed to save media")
	}

	logrus.Infof("Downloaded %s to: %s (size: %d bytes)", mediaType, filePath, len(mediaData))
	return filePath, nil
}

// DownloadMediaBatch 批量下载媒体文件
func (d *MediaDownloader) DownloadMediaBatch(urls []string, mediaType MediaType) ([]string, error) {
	var localPaths []string
	var errs []error

	for _, url := range urls {
		localPath, err := d.DownloadMedia(url, mediaType)
		if err != nil {
			logrus.Warnf("Failed to download %s: %v", url, err)
			errs = append(errs, fmt.Errorf("failed to download %s: %w", url, err))
			continue
		}
		localPaths = append(localPaths, localPath)
	}

	if len(errs) > 0 {
		logrus.Warnf("Some downloads failed: %d errors", len(errs))
		// 不返回错误，只记录日志，因为部分下载成功也是有用的
	}

	return localPaths, nil
}

// extractTaskID 从URL中提取task ID
func extractTaskID(mediaURL string) (string, bool) {
	// URL解码
	decoded, err := url.QueryUnescape(mediaURL)
	if err != nil {
		decoded = mediaURL // 如果解码失败，使用原始URL
	}

	// 匹配 task_[a-z0-9]{26} 模式
	re := regexp.MustCompile(`task_[a-z0-9]{26}`)
	matches := re.FindString(decoded)

	if matches != "" {
		return matches, true
	}
	return "", false
}

// generateFolderName 生成文件夹名（优先使用task ID，否则使用URL哈希）
func (d *MediaDownloader) generateFolderName(mediaURL string) string {
	// 优先尝试提取task ID
	if taskID, ok := extractTaskID(mediaURL); ok {
		logrus.Debugf("Extracted task ID: %s from URL", taskID)
		return taskID
	}

	// 回退方案：使用URL的SHA256哈希
	logrus.Warnf("Could not extract task ID from URL, using hash fallback")
	hash := sha256.Sum256([]byte(mediaURL))
	hashStr := fmt.Sprintf("%x", hash)

	// 取前12位哈希值作为文件夹名
	return hashStr[:12]
}

// IsVideoURL 判断 URL 是否为视频
func IsVideoURL(url string) bool {
	return strings.Contains(url, ".mp4")
}

// IsThumbnailURL 判断 URL 是否为缩略图
func IsThumbnailURL(url string) bool {
	return strings.Contains(url, ".webp") || strings.Contains(url, "thumbnail")
}
