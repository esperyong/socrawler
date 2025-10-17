package sora

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// OSSConfig contains configuration for OSS
type OSSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	Endpoint        string
	Region          string
}

// OSSUploader handles uploading videos to Aliyun OSS
type OSSUploader struct {
	config *OSSConfig
	client *oss.Client
	db     *VideoDatabase
}

// NewOSSUploader creates a new OSS uploader with provided configuration
func NewOSSUploader(config *OSSConfig, db *VideoDatabase) (*OSSUploader, error) {
	// Validate required fields
	if config.AccessKeyID == "" {
		return nil, errors.New("OSS AccessKeyID is required")
	}
	if config.AccessKeySecret == "" {
		return nil, errors.New("OSS AccessKeySecret is required")
	}

	// Set defaults for optional fields
	if config.BucketName == "" {
		config.BucketName = "dreammedias" // default bucket
	}
	if config.Endpoint == "" {
		config.Endpoint = "oss-cn-beijing.aliyuncs.com" // default endpoint
	}
	if config.Region == "" {
		config.Region = "cn-beijing" // default region
	}

	// Initialize OSS client
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.AccessKeyID,
			config.AccessKeySecret,
		)).
		WithRegion(config.Region)

	client := oss.NewClient(cfg)

	return &OSSUploader{
		config: config,
		client: client,
		db:     db,
	}, nil
}

// NewOSSConfigFromEnv creates an OSSConfig from environment variables
// Only AccessKeyID and AccessKeySecret are required, others have defaults
func NewOSSConfigFromEnv() (*OSSConfig, error) {
	accessKeyID := os.Getenv("OSS_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("OSS_ACCESS_KEY_SECRET")
	bucketName := os.Getenv("OSS_BUCKET_NAME")
	endpoint := os.Getenv("OSS_ENDPOINT")
	region := os.Getenv("OSS_REGION")

	// Validate required fields
	if accessKeyID == "" {
		return nil, errors.New("OSS_ACCESS_KEY_ID environment variable is required")
	}
	if accessKeySecret == "" {
		return nil, errors.New("OSS_ACCESS_KEY_SECRET environment variable is required")
	}

	// Set defaults for optional fields
	if bucketName == "" {
		bucketName = "dreammedias"
	}
	if endpoint == "" {
		endpoint = "oss-cn-beijing.aliyuncs.com"
	}
	if region == "" {
		region = "cn-beijing"
	}

	return &OSSConfig{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		BucketName:      bucketName,
		Endpoint:        endpoint,
		Region:          region,
	}, nil
}

// UploadVideoToOSS uploads a single video file to OSS
func (ou *OSSUploader) UploadVideoToOSS(ctx context.Context, localPath, postID string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return "", errors.Errorf("local video file does not exist: %s", localPath)
	}

	// Read file
	fileData, err := os.ReadFile(localPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to read video file")
	}

	// Get file extension
	ext := filepath.Ext(localPath)
	if ext == "" {
		ext = ".mp4" // default extension
	}

	// Generate OSS key: sora_videos/{postID}.mp4
	ossKey := fmt.Sprintf("sora_videos/%s%s", postID, ext)

	logrus.Debugf("Uploading video to OSS: local_path=%s, oss_key=%s, size=%d bytes",
		localPath, ossKey, len(fileData))

	// Create upload request
	putRequest := &oss.PutObjectRequest{
		Bucket: oss.Ptr(ou.config.BucketName),
		Key:    oss.Ptr(ossKey),
		Body:   bytes.NewReader(fileData),
	}

	// Upload file
	result, err := ou.client.PutObject(ctx, putRequest)
	if err != nil {
		return "", errors.Wrap(err, "OSS upload failed")
	}

	// Construct public URL
	ossURL := fmt.Sprintf("https://%s.%s/%s",
		ou.config.BucketName,
		ou.config.Endpoint,
		ossKey)

	logrus.Infof("Successfully uploaded video to OSS: post_id=%s, url=%s, etag=%s",
		postID, ossURL, oss.ToString(result.ETag))

	return ossURL, nil
}

// EnsureOSSURL ensures a video has an OSS URL, uploading if necessary
func (ou *OSSUploader) EnsureOSSURL(ctx context.Context, record *VideoRecord) (string, error) {
	// If already has OSS URL, return it
	if record.OSSVideoURL.Valid && record.OSSVideoURL.String != "" {
		logrus.Debugf("Video already has OSS URL: post_id=%s, url=%s",
			record.PostID, record.OSSVideoURL.String)
		return record.OSSVideoURL.String, nil
	}

	// Check if local video file exists
	if record.LocalVideoPath == "" {
		return "", errors.New("video has no local path recorded")
	}

	if _, err := os.Stat(record.LocalVideoPath); os.IsNotExist(err) {
		return "", errors.Errorf("local video file not found: %s", record.LocalVideoPath)
	}

	// Upload to OSS
	logrus.Infof("Uploading video to OSS: post_id=%s, local_path=%s",
		record.PostID, record.LocalVideoPath)

	ossURL, err := ou.UploadVideoToOSS(ctx, record.LocalVideoPath, record.PostID)
	if err != nil {
		return "", errors.Wrap(err, "failed to upload video to OSS")
	}

	// Update database
	if err := ou.db.UpdateOSSVideoURL(record.PostID, ossURL); err != nil {
		return "", errors.Wrap(err, "failed to update OSS URL in database")
	}

	// Update the record in memory
	record.OSSVideoURL = sql.NullString{String: ossURL, Valid: true}

	return ossURL, nil
}

// UploadUnuploadedVideosToOSS uploads all videos that need OSS URLs
func (ou *OSSUploader) UploadUnuploadedVideosToOSS(ctx context.Context, limit int) (int, int, error) {
	// Get unuploaded videos (these need OSS URLs before Goldcast upload)
	videos, err := ou.db.GetUnuploadedVideos(limit)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to get unuploaded videos")
	}

	if len(videos) == 0 {
		logrus.Info("No videos need OSS upload")
		return 0, 0, nil
	}

	successCount := 0
	failureCount := 0

	for i, video := range videos {
		// Skip if already has OSS URL
		if video.OSSVideoURL.Valid && video.OSSVideoURL.String != "" {
			logrus.Debugf("[%d/%d] Video already has OSS URL, skipping: post_id=%s",
				i+1, len(videos), video.PostID)
			continue
		}

		logrus.Infof("[%d/%d] Uploading video to OSS: post_id=%s",
			i+1, len(videos), video.PostID)

		_, err := ou.EnsureOSSURL(ctx, video)
		if err != nil {
			logrus.Errorf("Failed to upload video to OSS: post_id=%s, error=%v",
				video.PostID, err)
			failureCount++
			continue
		}

		successCount++
	}

	logrus.Infof("OSS upload batch complete: success=%d, failed=%d",
		successCount, failureCount)

	return successCount, failureCount, nil
}
