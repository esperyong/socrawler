# Goldcast Upload Integration - Implementation Summary

## Overview

Successfully implemented Goldcast media CMS integration that allows automatic uploading of downloaded Sora videos to the Goldcast system via API.

## Implementation Date

October 17, 2025

## Changes Made

### 1. Database Schema Updates

**File**: `sora/database.go`, `sora/types.go`

- Added `uploaded` boolean field to `sora_videos` table
- Implemented automatic migration for existing databases
- Added index on `uploaded` column for query performance
- New database methods:
  - `GetUnuploadedVideos(limit int)` - Query videos not yet uploaded
  - `MarkVideoAsUploaded(postID string)` - Mark video as uploaded
  - `GetUploadStats()` - Get statistics on upload status

**Migration**: Existing databases automatically get the `uploaded` column added on first use with the new code.

### 2. Goldcast Upload Module

**New File**: `sora/goldcast_uploader.go`

Core functionality:
- `GoldcastUploader` - Main uploader struct
- `UploadToGoldcast()` - Upload single video to Goldcast API
- `UploadUnuploadedVideos()` - Batch upload all unuploaded videos
- Automatic title truncation to 100 characters (Goldcast limit)
- Uses Sora's original video URL for upload
- Configurable API key and URL via flags or environment variables

**Key Features**:
- Environment variable support (`GOLDCAST_API_KEY`, `GOLDCAST_API_URL`)
- Default API credentials built-in
- Proper error handling and retry logic
- Progress tracking and detailed logging
- 500ms delay between uploads to avoid API rate limits

### 3. CLI Command

**File**: `main.go`

Added new subcommand: `feed uploadgoldcast`

**Flags**:
- `--api-key` - Goldcast API key
- `--api-url` - Goldcast API URL
- `--db-path` - Database path (default: ./sora.db)
- `--limit` - Upload limit (0 = all unuploaded)
- `--debug` - Enable debug logging

**Usage**:
```bash
./socrawler feed uploadgoldcast --limit 10
./socrawler feed uploadgoldcast --api-key YOUR_KEY
```

### 4. Shell Script Wrapper

**File**: `run_service.sh`

Added `feed-uploadgoldcast` command with:
- Environment variable defaults
- Parameter validation
- Colored output for better UX
- Integration with existing feed commands

**Usage**:
```bash
./run_service.sh feed-uploadgoldcast
./run_service.sh feed-uploadgoldcast --upload-limit=10
```

### 5. Documentation

**File**: `FEED_DOWNLOADER.md`

Added comprehensive "Goldcast Upload Integration" section covering:
- Basic usage examples
- Command parameters
- Environment variable configuration
- Typical workflows (sync + upload)
- Cron job integration examples
- Upload logic and behavior
- Video field mapping
- Idempotency guarantees
- SQL queries for monitoring

## Technical Details

### Video Information Mapping

| Sora Field | Goldcast Field | Processing |
|------------|----------------|------------|
| `video_url` | `media_url` | Direct mapping |
| `text` | `title` | Truncated to 100 chars |
| `text` | `description` | Full text |
| Fixed | `user` | {"username": "UID75203008801597", "email": "api@example.com", "name": "API User"} |

### Title Processing Logic

```go
// If text is empty, use default
if text == "" {
    return "Sora Video - {post_id}"
}

// Truncate to 100 characters
if len(text) <= 100 {
    return text
}

// Truncate with ellipsis
return text[:97] + "..."
```

### Upload Flow

1. Query database for videos where `uploaded = 0`
2. Apply limit if specified
3. For each video:
   - Upload to Goldcast via HTTP POST
   - If successful, mark `uploaded = 1` in database
   - If failed, leave `uploaded = 0` for retry
4. Return statistics (attempted, succeeded, failed)

### Idempotency

The system is designed to be idempotent:
- Only uploads videos with `uploaded = 0`
- Can be safely run multiple times
- Failed uploads remain retryable
- Perfect for cron jobs

## Testing Results

✅ Database migration works correctly on existing databases
✅ Command help and parameters display properly
✅ Shell script integration functional
✅ Upload process handles API errors gracefully
✅ Statistics and logging work as expected

**Test Database**: 317 videos, 0 uploaded initially

**Sample Output**:
```
time="2025-10-17T10:36:25+08:00" level=info msg="Current status: 0 uploaded, 317 not uploaded"
time="2025-10-17T10:36:25+08:00" level=info msg="Found 317 unuploaded videos, starting upload..."
time="2025-10-17T10:36:25+08:00" level=info msg="Uploading video 1/317: post_id=s_68f..., text=Make him smack the camera"
```

## Files Modified

1. ✅ `sora/types.go` - Added GoldcastUploadResult type
2. ✅ `sora/database.go` - Schema migration + new methods
3. ✅ `sora/goldcast_uploader.go` - NEW FILE - Core upload logic
4. ✅ `main.go` - Added feed uploadgoldcast subcommand
5. ✅ `run_service.sh` - Added shell wrapper for upload command
6. ✅ `FEED_DOWNLOADER.md` - Added documentation section

## Usage Examples

### Basic Upload

```bash
# Upload all unuploaded videos
./run_service.sh feed-uploadgoldcast

# Upload with limit
./run_service.sh feed-uploadgoldcast --upload-limit=10
```

### Automated Workflow

```bash
# Step 1: Download latest videos
./run_service.sh feed-sync --limit=1000

# Step 2: Upload to Goldcast
./run_service.sh feed-uploadgoldcast
```

### Cron Job Setup

```bash
# Download at 3 AM, upload at 3:30 AM daily
0 3 * * * cd /path/to/socrawler && ./run_service.sh feed-sync --limit=1000
30 3 * * * cd /path/to/socrawler && ./run_service.sh feed-uploadgoldcast
```

### Monitor Upload Status

```bash
# Check upload statistics
sqlite3 sora.db "SELECT 
    COUNT(*) as total,
    SUM(uploaded) as uploaded,
    COUNT(*) - SUM(uploaded) as pending
FROM sora_videos;"
```

## Environment Variables

### Required OSS Configuration (Updated: October 17, 2025)

As of the latest update, **OSS credentials are now required** and must be provided via environment variables for security reasons:

```bash
# Required OSS configuration
export OSS_ACCESS_KEY_ID="your-aliyun-access-key-id"
export OSS_ACCESS_KEY_SECRET="your-aliyun-access-key-secret"
export OSS_BUCKET_NAME="your-bucket-name"
export OSS_ENDPOINT="oss-cn-beijing.aliyuncs.com"
export OSS_REGION="cn-beijing"  # Optional, defaults to cn-beijing
```

### Optional Goldcast Configuration

```bash
# Optional: Set Goldcast API configuration
export GOLDCAST_API_KEY="your-api-key"
export GOLDCAST_API_URL="https://your-goldcast-instance.com/api/upload"

# Then run without flags
./run_service.sh feed-uploadgoldcast
```

**Important**: The upload command will fail with a clear error message if OSS environment variables are not set.

## Error Handling

The system handles various error scenarios:

1. **Expired Video URLs** (403): Logged as failed, can be re-downloaded in next feed-sync
2. **Network Errors**: Logged, video remains unuploaded for retry
3. **API Errors**: Full error message logged, video not marked as uploaded
4. **Database Errors**: Transaction rolled back, no data corruption

## Build and Deploy

```bash
# Build the binary
./run_service.sh build

# Or manually
go build -o socrawler .

# Verify the new command
./socrawler feed uploadgoldcast --help
```

## Performance Considerations

- **Batch Size**: Use `--limit` to control batch size
- **Rate Limiting**: 500ms delay between uploads to avoid overwhelming API
- **Database Queries**: Indexed on `uploaded` field for fast queries
- **Memory**: Processes videos sequentially to minimize memory usage

## Future Enhancements

Possible improvements:
1. Configurable retry logic with exponential backoff
2. Parallel upload support (with concurrency limit)
3. Upload status dashboard/web UI
4. Webhook notifications on upload completion
5. Support for uploading from local files (not just URLs)
6. Batch API support if Goldcast adds it

## Troubleshooting

### Video URLs expired (403 errors)

**Solution**: Run `feed-sync` again to get fresh URLs:
```bash
./run_service.sh feed-sync --limit=1000
./run_service.sh feed-uploadgoldcast
```

### Database locked error

**Solution**: Ensure no other process is using the database
```bash
lsof sora.db  # Check what's using it
```

### Upload failures

**Solution**: Check logs for specific error messages
```bash
./run_service.sh feed-uploadgoldcast --debug
```

## Security Notes

- **OSS Credentials**: Must be provided via environment variables (no longer stored in code)
- **Goldcast API Key**: Passed in Authorization header: `Api-Key {key}`
- **Goldcast API Key Default**: Embedded in code (acceptable for internal tool)
- Environment variables recommended for production deployments
- No sensitive data stored in database (only URLs and metadata)

## Compliance

- Videos uploaded using original Sora URLs (Goldcast downloads them)
- User information is hardcoded as specified
- Upload tracking prevents duplicates
- Respects API rate limits with delays

## Success Criteria Met

✅ Uploads videos to Goldcast CMS via API
✅ Tracks upload status in database
✅ Prevents duplicate uploads
✅ CLI command and shell wrapper implemented
✅ Environment variable support
✅ Comprehensive documentation
✅ Backward compatible with existing database
✅ Idempotent operation
✅ Error handling and logging
✅ Easy to integrate with cron/automation

## Contact

For questions or issues regarding this integration, please refer to the main project documentation or contact the development team.


