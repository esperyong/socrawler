# Sora Crawler Deduplication System

## Overview

The Sora crawler now includes a deduplication system that prevents re-downloading the same videos and thumbnails on subsequent crawls. Files are organized into unique folders based on their URL hash for better organization.

## How It Works

### URL-Based Deduplication

1. **Hash Generation**: Each media URL is hashed using SHA256, and the first 12 characters are used as a unique identifier
2. **Folder Organization**: Each unique video/thumbnail gets its own folder named with the hash
3. **File Check**: Before downloading, the system checks if the file already exists
4. **Skip Download**: If the file exists, download is skipped, saving bandwidth and time

### File Structure

**Before (old system):**
```
downloads/sora/
├── sora_video_1760521880_8bfb1a7bdd74.mp4
├── sora_video_1760521881_8bfb1a7bdd74.mp4  (duplicate!)
├── sora_thumb_1760521926_3892fb6b23d4.webp
└── sora_thumb_1760521927_3892fb6b23d4.webp  (duplicate!)
```

**After (new system):**
```
downloads/sora/
├── 8bfb1a7bdd74/
│   ├── video.mp4
│   └── thumbnail.webp
├── 3892fb6b23d4/
│   ├── video.mp4
│   └── thumbnail.webp
└── debug_initial_page.png
```

### Benefits

1. **No Duplicates**: Same URL always maps to same folder/file
2. **Bandwidth Savings**: Skip downloads for already-fetched content
3. **Better Organization**: Each video-thumbnail pair in its own folder
4. **Faster Subsequent Crawls**: Only download new content
5. **Deterministic**: Same URL always produces same hash/folder name

## Implementation Details

### Key Changes in `sora/downloader.go`

#### 1. `generateURLHash()` Method
```go
func (d *MediaDownloader) generateURLHash(mediaURL string) string {
    hash := sha256.Sum256([]byte(mediaURL))
    hashStr := fmt.Sprintf("%x", hash)
    return hashStr[:12]  // First 12 chars of hash
}
```

#### 2. Updated `DownloadMedia()` Method
- Generates folder name from URL hash
- Creates folder structure: `{savePath}/{hash}/`
- Saves files with simple names: `video.mp4` or `thumbnail.webp`
- Checks file existence before downloading
- Skips download if file exists

#### 3. Deduplication Check
```go
if _, err := os.Stat(filePath); err == nil {
    logrus.Infof("File already exists, skipping download: %s", filePath)
    return filePath, nil
}
```

## Usage

No changes needed in your API calls! The system works automatically:

```bash
# First crawl - downloads new content
curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 60,
    "scroll_interval_seconds": 5,
    "save_path": "./downloads/sora"
  }'

# Second crawl - skips duplicates, only downloads new content
curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 60,
    "scroll_interval_seconds": 5,
    "save_path": "./downloads/sora"
  }'
```

## Testing

Use the included test script to verify deduplication:

```bash
./test_deduplication.sh
```

This script:
1. Runs a crawl for 30 seconds
2. Waits 5 seconds
3. Runs another crawl for 30 seconds
4. Shows file structure
5. Highlights "File already exists, skipping download" messages in logs

## Log Messages

### New Download
```
time="..." level=debug msg="Downloading video: https://videos.openai.com/..."
time="..." level=info msg="Downloaded video to: downloads/sora/8bfb1a7bdd74/video.mp4 (size: 1756155 bytes)"
```

### Duplicate Skipped
```
time="..." level=info msg="File already exists, skipping download: downloads/sora/8bfb1a7bdd74/video.mp4"
```

## Migration from Old System

If you have existing files from the old naming system:

1. **Option 1 - Clean Start**: Delete old files and start fresh
   ```bash
   rm -rf downloads/sora/sora_video_*
   rm -rf downloads/sora/sora_thumb_*
   ```

2. **Option 2 - Keep Both**: Old files will remain, new system creates folders
   - Old files: `sora_video_1760521880_8bfb1a7bdd74.mp4`
   - New files: `8bfb1a7bdd74/video.mp4`
   - You can manually clean up old files later

## Technical Notes

- **Hash Algorithm**: SHA256 ensures consistent hashing across runs
- **Hash Length**: 12 characters provides good uniqueness (16^12 combinations)
- **Collision Risk**: Extremely low with SHA256
- **File Permissions**: Folders created with 0755, files with 0644
- **Atomic Check**: File existence check is atomic via `os.Stat()`

## Future Enhancements

Potential improvements:
- Add metadata file tracking download timestamps
- Support for cleaning up orphaned files
- Statistics on duplicates skipped
- Configurable hash length
- Optional URL-to-hash mapping file for debugging

