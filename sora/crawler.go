package sora

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	SoraURL = "https://sora.chatgpt.com/"
)

// CrawlerAction Sora 爬虫
type CrawlerAction struct {
	page *rod.Page
}

// NewCrawlerAction 创建爬虫实例
func NewCrawlerAction(page *rod.Page) *CrawlerAction {
	return &CrawlerAction{page: page}
}

// StartCrawl 开始爬取
func (c *CrawlerAction) StartCrawl(ctx context.Context, req *CrawlRequest) (*CrawlResult, error) {
	logrus.Infof("Starting Sora crawl: duration=%ds, scroll_interval=%ds, save_path=%s",
		req.TotalDurationSeconds, req.ScrollIntervalSeconds, req.SavePath)

	// 创建带超时的上下文
	crawlCtx, cancel := context.WithTimeout(ctx, time.Duration(req.TotalDurationSeconds)*time.Second)
	defer cancel()

	// 使用上下文的 page
	page := c.page.Context(crawlCtx)

	// 收集的媒体 URL
	videoURLs := make(map[string]bool)
	thumbnailURLs := make(map[string]bool)
	var mu sync.Mutex

	// 网络请求统计
	var totalRequests int
	var openaiRequests int
	var mediaRequests int

	// 创建 HTTP 客户端用于网络请求
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 设置网络拦截
	router := page.HijackRequests()
	defer router.Stop()

	router.MustAdd("*", func(ctx *rod.Hijack) {
		mu.Lock()
		totalRequests++
		mu.Unlock()
		// 继续请求
		if err := ctx.LoadResponse(httpClient, true); err != nil {
			logrus.Debugf("Failed to load response: %v", err)
			return
		}

		url := ctx.Request.URL().String()

		// 统计 OpenAI 相关请求
		if containsString(url, "openai.com") {
			mu.Lock()
			openaiRequests++
			mu.Unlock()
		}

		// 记录所有媒体相关的请求
		if containsString(url, "openai.com") &&
			(containsString(url, ".mp4") || containsString(url, ".webp") ||
				containsString(url, "video") || containsString(url, "thumbnail")) {
			mu.Lock()
			mediaRequests++
			mu.Unlock()
			logrus.Debugf("Detected media-related URL: %s", url)
		}

		// 只捕获 videos.openai.com 的媒体文件
		if !containsString(url, "videos.openai.com") {
			return
		}

		mu.Lock()
		defer mu.Unlock()

		// 判断是视频还是缩略图
		if IsVideoURL(url) {
			if !videoURLs[url] {
				logrus.Infof("Found video: %s", url)
				videoURLs[url] = true
			}
		} else if IsThumbnailURL(url) {
			if !thumbnailURLs[url] {
				logrus.Infof("Found thumbnail: %s", url)
				thumbnailURLs[url] = true
			}
		} else {
			logrus.Debugf("Found videos.openai.com URL but not video/thumbnail: %s", url)
		}
	})

	go router.Run()

	// 导航到 Sora 页面
	logrus.Info("Navigating to Sora page...")
	if err := page.Navigate(SoraURL); err != nil {
		return nil, errors.Wrap(err, "failed to navigate to Sora page")
	}

	// 等待页面加载
	if err := page.WaitLoad(); err != nil {
		return nil, errors.Wrap(err, "failed to wait for page load")
	}

	logrus.Info("Page loaded, waiting for initial content...")

	// 检查页面标题和URL
	pageTitle, _ := page.Info()
	logrus.Infof("Page title: %s, URL: %s", pageTitle.Title, pageTitle.URL)

	// 检查页面是否需要登录
	html, err := page.HTML()
	if err == nil {
		if containsString(html, "Sign in") || containsString(html, "Log in") || containsString(html, "login") {
			logrus.Warn("Page may require login - detected login-related content")
		}
		logrus.Debugf("Page HTML length: %d bytes", len(html))

		// 保存初始页面截图用于调试
		screenshotPath := req.SavePath + "/debug_initial_page.png"
		if screenshotData, err := page.Screenshot(true, nil); err == nil {
			if err := os.WriteFile(screenshotPath, screenshotData, 0644); err == nil {
				logrus.Infof("Saved initial page screenshot to: %s", screenshotPath)
			}
		}
	}

	time.Sleep(5 * time.Second) // 等待初始内容加载

	// 定时滚动页面
	ticker := time.NewTicker(time.Duration(req.ScrollIntervalSeconds) * time.Second)
	defer ticker.Stop()

	scrollCount := 0
	startTime := time.Now()

	for {
		select {
		case <-crawlCtx.Done():
			// 超时或取消
			logrus.Info("Crawl finished due to timeout or cancellation")
			goto DOWNLOAD

		case <-ticker.C:
			scrollCount++
			logrus.Infof("Scrolling page (count: %d)...", scrollCount)

			// 滚动页面
			if err := c.scrollPage(page); err != nil {
				logrus.Warnf("Failed to scroll page: %v", err)
			}

			// 等待新内容加载
			time.Sleep(2 * time.Second)

			// 定期报告当前收集的媒体数量和网络请求统计
			mu.Lock()
			currentVideos := len(videoURLs)
			currentThumbnails := len(thumbnailURLs)
			currentTotal := totalRequests
			currentOpenAI := openaiRequests
			currentMedia := mediaRequests
			mu.Unlock()
			logrus.Infof("Current status: videos=%d, thumbnails=%d, total_requests=%d, openai_requests=%d, media_requests=%d",
				currentVideos, currentThumbnails, currentTotal, currentOpenAI, currentMedia)
		}
	}

DOWNLOAD:
	elapsed := time.Since(startTime)
	logrus.Infof("Crawl completed: duration=%v, videos=%d, thumbnails=%d, total_requests=%d, openai_requests=%d, media_requests=%d",
		elapsed, len(videoURLs), len(thumbnailURLs), totalRequests, openaiRequests, mediaRequests)

	// 下载媒体文件
	result, err := c.downloadMedia(videoURLs, thumbnailURLs, req.SavePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to download media")
	}

	result.DurationSeconds = int(elapsed.Seconds())

	return result, nil
}

// scrollPage 滚动页面
func (c *CrawlerAction) scrollPage(page *rod.Page) error {
	// 滚动一个视口高度
	_, err := page.Eval(`() => {
		window.scrollBy({
			top: window.innerHeight,
			behavior: 'smooth'
		});
	}`)
	return err
}

// downloadMedia 下载媒体文件
func (c *CrawlerAction) downloadMedia(videoURLs, thumbnailURLs map[string]bool, savePath string) (*CrawlResult, error) {
	downloader := NewMediaDownloader(savePath)

	// 转换 map 到 slice
	videos := make([]string, 0, len(videoURLs))
	for url := range videoURLs {
		videos = append(videos, url)
	}

	thumbnails := make([]string, 0, len(thumbnailURLs))
	for url := range thumbnailURLs {
		thumbnails = append(thumbnails, url)
	}

	logrus.Infof("Starting download: %d videos, %d thumbnails", len(videos), len(thumbnails))

	// 下载视频
	var videoPaths []string
	var err error
	if len(videos) > 0 {
		videoPaths, err = downloader.DownloadMediaBatch(videos, MediaTypeVideo)
		if err != nil {
			logrus.Warnf("Some video downloads failed: %v", err)
		}
	}

	// 下载缩略图
	var thumbnailPaths []string
	if len(thumbnails) > 0 {
		thumbnailPaths, err = downloader.DownloadMediaBatch(thumbnails, MediaTypeThumbnail)
		if err != nil {
			logrus.Warnf("Some thumbnail downloads failed: %v", err)
		}
	}

	return &CrawlResult{
		Videos:          videoPaths,
		Thumbnails:      thumbnailPaths,
		TotalVideos:     len(videoPaths),
		TotalThumbnails: len(thumbnailPaths),
	}, nil
}

// containsString 检查字符串是否包含子串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

// findSubstring 在字符串中查找子串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
