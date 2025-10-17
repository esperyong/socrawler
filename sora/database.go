package sora

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// VideoDatabase manages video metadata in SQLite
type VideoDatabase struct {
	db *sql.DB
}

// VideoRecord represents a video record in the database
type VideoRecord struct {
	PostID             string
	GenerationID       string
	VideoURL           string
	ThumbnailURL       string
	Text               string
	Username           string
	UserID             string
	PostedAt           float64
	DownloadedAt       time.Time
	LocalVideoPath     string
	LocalThumbnailPath string
	Width              int
	Height             int
	Uploaded           bool
	OSSVideoURL        sql.NullString
	GoldcastToken      sql.NullString
}

// NewVideoDatabase creates or opens a SQLite database
func NewVideoDatabase(dbPath string) (*VideoDatabase, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping database")
	}

	vdb := &VideoDatabase{db: db}

	// Create tables
	if err := vdb.createTables(); err != nil {
		return nil, errors.Wrap(err, "failed to create tables")
	}

	logrus.Infof("Database initialized: %s", dbPath)
	return vdb, nil
}

// createTables creates the necessary database tables
func (vdb *VideoDatabase) createTables() error {
	// Create table without uploaded column first (for backward compatibility)
	schema := `
	CREATE TABLE IF NOT EXISTS sora_videos (
		post_id TEXT PRIMARY KEY,
		generation_id TEXT,
		video_url TEXT,
		thumbnail_url TEXT,
		text TEXT,
		username TEXT,
		user_id TEXT,
		posted_at REAL,
		downloaded_at DATETIME,
		local_video_path TEXT,
		local_thumbnail_path TEXT,
		width INTEGER,
		height INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_posted_at ON sora_videos(posted_at);
	CREATE INDEX IF NOT EXISTS idx_username ON sora_videos(username);
	CREATE INDEX IF NOT EXISTS idx_downloaded_at ON sora_videos(downloaded_at);
	`

	_, err := vdb.db.Exec(schema)
	if err != nil {
		return errors.Wrap(err, "failed to execute schema")
	}

	// Add uploaded column to existing tables (migration)
	// Check if column exists first
	var uploadedColumnExists int
	row := vdb.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('sora_videos') WHERE name='uploaded'")
	if err := row.Scan(&uploadedColumnExists); err == nil && uploadedColumnExists == 0 {
		migration := `ALTER TABLE sora_videos ADD COLUMN uploaded INTEGER DEFAULT 0;`
		if _, err := vdb.db.Exec(migration); err != nil {
			logrus.Warnf("Failed to add uploaded column (may already exist): %v", err)
		} else {
			logrus.Info("Added uploaded column to existing sora_videos table")
		}
	}

	// Add oss_video_url column to existing tables (migration)
	var ossColumnExists int
	row2 := vdb.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('sora_videos') WHERE name='oss_video_url'")
	if err := row2.Scan(&ossColumnExists); err == nil && ossColumnExists == 0 {
		migration := `ALTER TABLE sora_videos ADD COLUMN oss_video_url TEXT;`
		if _, err := vdb.db.Exec(migration); err != nil {
			logrus.Warnf("Failed to add oss_video_url column (may already exist): %v", err)
		} else {
			logrus.Info("Added oss_video_url column to existing sora_videos table")
		}
	}

	// Add goldcast_token column to existing tables (migration)
	var goldcastTokenExists int
	row3 := vdb.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('sora_videos') WHERE name='goldcast_token'")
	if err := row3.Scan(&goldcastTokenExists); err == nil && goldcastTokenExists == 0 {
		migration := `ALTER TABLE sora_videos ADD COLUMN goldcast_token TEXT;`
		if _, err := vdb.db.Exec(migration); err != nil {
			logrus.Warnf("Failed to add goldcast_token column (may already exist): %v", err)
		} else {
			logrus.Info("Added goldcast_token column to existing sora_videos table")
		}
	}

	// Create index for uploaded column if it exists
	indexSchema := `CREATE INDEX IF NOT EXISTS idx_uploaded ON sora_videos(uploaded);`
	vdb.db.Exec(indexSchema) // Ignore errors if column doesn't exist

	return nil
}

// VideoExists checks if a video with the given post_id exists
func (vdb *VideoDatabase) VideoExists(postID string) (bool, error) {
	var count int
	err := vdb.db.QueryRow("SELECT COUNT(*) FROM sora_videos WHERE post_id = ?", postID).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "failed to check video existence")
	}
	return count > 0, nil
}

// GetExistingPostIDs returns a set of existing post IDs for quick lookup
func (vdb *VideoDatabase) GetExistingPostIDs() (map[string]bool, error) {
	rows, err := vdb.db.Query("SELECT post_id FROM sora_videos")
	if err != nil {
		return nil, errors.Wrap(err, "failed to query post IDs")
	}
	defer rows.Close()

	postIDs := make(map[string]bool)
	for rows.Next() {
		var postID string
		if err := rows.Scan(&postID); err != nil {
			return nil, errors.Wrap(err, "failed to scan post ID")
		}
		postIDs[postID] = true
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating rows")
	}

	return postIDs, nil
}

// InsertVideo inserts a new video record
func (vdb *VideoDatabase) InsertVideo(record *VideoRecord) error {
	query := `
	INSERT INTO sora_videos (
		post_id, generation_id, video_url, thumbnail_url, text,
		username, user_id, posted_at, downloaded_at,
		local_video_path, local_thumbnail_path, width, height, uploaded, oss_video_url, goldcast_token
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := vdb.db.Exec(query,
		record.PostID,
		record.GenerationID,
		record.VideoURL,
		record.ThumbnailURL,
		record.Text,
		record.Username,
		record.UserID,
		record.PostedAt,
		record.DownloadedAt,
		record.LocalVideoPath,
		record.LocalThumbnailPath,
		record.Width,
		record.Height,
		record.Uploaded,
		record.OSSVideoURL,
		record.GoldcastToken,
	)

	if err != nil {
		return errors.Wrap(err, "failed to insert video record")
	}

	logrus.Debugf("Inserted video record: post_id=%s, username=%s", record.PostID, record.Username)
	return nil
}

// scanVideoRecord scans a row into a VideoRecord
func scanVideoRecord(scanner interface {
	Scan(dest ...interface{}) error
}) (*VideoRecord, error) {
	record := &VideoRecord{}
	err := scanner.Scan(
		&record.PostID,
		&record.GenerationID,
		&record.VideoURL,
		&record.ThumbnailURL,
		&record.Text,
		&record.Username,
		&record.UserID,
		&record.PostedAt,
		&record.DownloadedAt,
		&record.LocalVideoPath,
		&record.LocalThumbnailPath,
		&record.Width,
		&record.Height,
		&record.Uploaded,
		&record.OSSVideoURL,
		&record.GoldcastToken,
	)
	return record, err
}

// selectVideoFields returns the field list for SELECT queries
func selectVideoFields() string {
	return `post_id, generation_id, video_url, thumbnail_url, text,
		   username, user_id, posted_at, downloaded_at,
		   local_video_path, local_thumbnail_path, width, height, uploaded, oss_video_url, goldcast_token`
}

// GetVideoByPostID retrieves a video record by post ID
func (vdb *VideoDatabase) GetVideoByPostID(postID string) (*VideoRecord, error) {
	query := fmt.Sprintf(`SELECT %s FROM sora_videos WHERE post_id = ?`, selectVideoFields())

	record, err := scanVideoRecord(vdb.db.QueryRow(query, postID))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get video record")
	}

	return record, nil
}

// GetVideoCount returns the total number of videos in the database
func (vdb *VideoDatabase) GetVideoCount() (int, error) {
	var count int
	err := vdb.db.QueryRow("SELECT COUNT(*) FROM sora_videos").Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count videos")
	}
	return count, nil
}

// GetRecentVideos returns the most recent N videos
func (vdb *VideoDatabase) GetRecentVideos(limit int) ([]*VideoRecord, error) {
	query := fmt.Sprintf(`SELECT %s FROM sora_videos ORDER BY posted_at DESC LIMIT ?`, selectVideoFields())

	rows, err := vdb.db.Query(query, limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query recent videos")
	}
	defer rows.Close()

	var records []*VideoRecord
	for rows.Next() {
		record, err := scanVideoRecord(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan video record")
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating rows")
	}

	return records, nil
}

// GetAllVideos returns all videos from the database
func (vdb *VideoDatabase) GetAllVideos(limit int) ([]*VideoRecord, error) {
	var query string
	var rows *sql.Rows
	var err error

	fields := selectVideoFields()
	if limit > 0 {
		query = fmt.Sprintf(`SELECT %s FROM sora_videos ORDER BY posted_at DESC LIMIT ?`, fields)
		rows, err = vdb.db.Query(query, limit)
	} else {
		query = fmt.Sprintf(`SELECT %s FROM sora_videos ORDER BY posted_at DESC`, fields)
		rows, err = vdb.db.Query(query)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to query videos")
	}
	defer rows.Close()

	var records []*VideoRecord
	for rows.Next() {
		record, err := scanVideoRecord(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan video record")
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating rows")
	}

	return records, nil
}

// ExportVideosAsJSON exports videos as JSON in FeedResponse format
func (vdb *VideoDatabase) ExportVideosAsJSON(limit int, outputPath string) error {
	// Get videos from database
	records, err := vdb.GetAllVideos(limit)
	if err != nil {
		return errors.Wrap(err, "failed to get videos")
	}

	if len(records) == 0 {
		logrus.Warn("No videos found in database")
		return errors.New("no videos found in database")
	}

	// Convert to FeedResponse
	feedResponse := VideoRecordsToFeedResponse(records)

	// Marshal to JSON
	var data []byte
	if outputPath == "" || outputPath == "-" {
		// Output to stdout without indentation for piping
		data, err = json.Marshal(feedResponse)
	} else {
		// Output to file with indentation for readability
		data, err = json.MarshalIndent(feedResponse, "", "    ")
	}

	if err != nil {
		return errors.Wrap(err, "failed to marshal JSON")
	}

	// Write output
	if outputPath == "" || outputPath == "-" {
		// Write to stdout
		os.Stdout.Write(data)
		os.Stdout.Write([]byte("\n"))
	} else {
		// Write to file
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return errors.Wrap(err, "failed to write output file")
		}
		logrus.Infof("Exported %d videos to: %s", len(records), outputPath)
	}

	return nil
}

// GetUnuploadedVideos returns videos that haven't been uploaded to Goldcast
func (vdb *VideoDatabase) GetUnuploadedVideos(limit int) ([]*VideoRecord, error) {
	var query string
	var rows *sql.Rows
	var err error

	fields := selectVideoFields()
	if limit > 0 {
		query = fmt.Sprintf(`SELECT %s FROM sora_videos WHERE uploaded = 0 ORDER BY posted_at DESC LIMIT ?`, fields)
		rows, err = vdb.db.Query(query, limit)
	} else {
		query = fmt.Sprintf(`SELECT %s FROM sora_videos WHERE uploaded = 0 ORDER BY posted_at DESC`, fields)
		rows, err = vdb.db.Query(query)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to query unuploaded videos")
	}
	defer rows.Close()

	var records []*VideoRecord
	for rows.Next() {
		record, err := scanVideoRecord(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan video record")
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating rows")
	}

	return records, nil
}

// MarkVideoAsUploaded marks a video as uploaded to Goldcast
func (vdb *VideoDatabase) MarkVideoAsUploaded(postID string) error {
	query := `UPDATE sora_videos SET uploaded = 1 WHERE post_id = ?`

	result, err := vdb.db.Exec(query, postID)
	if err != nil {
		return errors.Wrap(err, "failed to mark video as uploaded")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.Errorf("no video found with post_id: %s", postID)
	}

	logrus.Debugf("Marked video as uploaded: post_id=%s", postID)
	return nil
}

// GetUploadStats returns statistics about uploaded vs unuploaded videos
func (vdb *VideoDatabase) GetUploadStats() (uploaded int, unuploaded int, err error) {
	// Get uploaded count
	err = vdb.db.QueryRow("SELECT COUNT(*) FROM sora_videos WHERE uploaded = 1").Scan(&uploaded)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to count uploaded videos")
	}

	// Get unuploaded count
	err = vdb.db.QueryRow("SELECT COUNT(*) FROM sora_videos WHERE uploaded = 0").Scan(&unuploaded)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to count unuploaded videos")
	}

	return uploaded, unuploaded, nil
}

// UpdateOSSVideoURL updates the OSS video URL for a video
func (vdb *VideoDatabase) UpdateOSSVideoURL(postID, ossURL string) error {
	query := `UPDATE sora_videos SET oss_video_url = ? WHERE post_id = ?`

	result, err := vdb.db.Exec(query, ossURL, postID)
	if err != nil {
		return errors.Wrap(err, "failed to update OSS video URL")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.Errorf("no video found with post_id: %s", postID)
	}

	logrus.Debugf("Updated OSS video URL: post_id=%s, url=%s", postID, ossURL)
	return nil
}

// UpdateGoldcastToken updates the Goldcast token for a video
func (vdb *VideoDatabase) UpdateGoldcastToken(postID, token string) error {
	query := `UPDATE sora_videos SET goldcast_token = ? WHERE post_id = ?`

	result, err := vdb.db.Exec(query, token, postID)
	if err != nil {
		return errors.Wrap(err, "failed to update Goldcast token")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.Errorf("no video found with post_id: %s", postID)
	}

	logrus.Debugf("Updated Goldcast token: post_id=%s, token=%s", postID, token)
	return nil
}

// Close closes the database connection
func (vdb *VideoDatabase) Close() error {
	if vdb.db != nil {
		return vdb.db.Close()
	}
	return nil
}
