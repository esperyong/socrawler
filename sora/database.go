package sora

import (
	"database/sql"
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
		local_video_path, local_thumbnail_path, width, height
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
	)

	if err != nil {
		return errors.Wrap(err, "failed to insert video record")
	}

	logrus.Debugf("Inserted video record: post_id=%s, username=%s", record.PostID, record.Username)
	return nil
}

// GetVideoByPostID retrieves a video record by post ID
func (vdb *VideoDatabase) GetVideoByPostID(postID string) (*VideoRecord, error) {
	query := `
	SELECT post_id, generation_id, video_url, thumbnail_url, text,
		   username, user_id, posted_at, downloaded_at,
		   local_video_path, local_thumbnail_path, width, height
	FROM sora_videos
	WHERE post_id = ?
	`

	record := &VideoRecord{}
	err := vdb.db.QueryRow(query, postID).Scan(
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
	)

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
	query := `
	SELECT post_id, generation_id, video_url, thumbnail_url, text,
		   username, user_id, posted_at, downloaded_at,
		   local_video_path, local_thumbnail_path, width, height
	FROM sora_videos
	ORDER BY posted_at DESC
	LIMIT ?
	`

	rows, err := vdb.db.Query(query, limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query recent videos")
	}
	defer rows.Close()

	var records []*VideoRecord
	for rows.Next() {
		record := &VideoRecord{}
		err := rows.Scan(
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
		)
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

// Close closes the database connection
func (vdb *VideoDatabase) Close() error {
	if vdb.db != nil {
		return vdb.db.Close()
	}
	return nil
}
