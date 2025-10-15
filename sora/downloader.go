package sora

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

	// 生成唯一文件名
	fileName := d.generateFileName(mediaURL, mediaType)
	filePath := filepath.Join(d.savePath, fileName)

	// 如果文件已存在，直接返回路径
	if _, err := os.Stat(filePath); err == nil {
		logrus.Debugf("File already exists: %s", filePath)
		return filePath, nil
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

// generateFileName 生成唯一的文件名
func (d *MediaDownloader) generateFileName(mediaURL string, mediaType MediaType) string {
	// 使用URL的SHA256哈希作为文件名，确保唯一性
	hash := sha256.Sum256([]byte(mediaURL))
	hashStr := fmt.Sprintf("%x", hash)

	// 取前12位哈希值作为文件名
	shortHash := hashStr[:12]

	// 添加时间戳确保更好的唯一性
	timestamp := time.Now().Unix()

	// 根据媒体类型确定扩展名
	var extension string
	var prefix string

	if mediaType == MediaTypeVideo {
		extension = "mp4"
		prefix = "sora_video"
	} else {
		extension = "webp"
		prefix = "sora_thumb"
	}

	return fmt.Sprintf("%s_%d_%s.%s", prefix, timestamp, shortHash, extension)
}

// IsVideoURL 判断 URL 是否为视频
func IsVideoURL(url string) bool {
	return strings.Contains(url, ".mp4")
}

// IsThumbnailURL 判断 URL 是否为缩略图
func IsThumbnailURL(url string) bool {
	return strings.Contains(url, ".webp") || strings.Contains(url, "thumbnail")
}
