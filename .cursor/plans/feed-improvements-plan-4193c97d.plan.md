<!-- 4193c97d-83d0-4560-af95-95883dfbf40e af0db38e-db33-4ead-aae7-7065ae23c49f -->
# Feed-Based Crawler Improvements

## Overview

Enhance the feed-based Sora video downloader with three key improvements: use post_id for folder naming, add feed file operations (fetch/download/sync subcommands), and enable JSON export of downloaded videos from the database.

## Changes Required

### 1. Use post_id for Folder Naming (Feed Downloads Only)

**File: `sora/downloader.go`**

Modify `MediaDownloader` to accept an optional `folderName` parameter:

- Add new method `DownloadMediaForFeed(mediaURL, folderName, mediaType)` 
- Keep existing `DownloadMedia()` using task_id extraction (for scroll-based crawler)
- Feed downloader will call the new method with post_id

**File: `sora/feed_downloader.go`**

Update `downloadVideos()` method around lines 176-182:

- Pass `item.Post.ID` as folder name to the new download method
- Change from: `fd.mediaDownloader.DownloadMedia(videoAttachment.DownloadableURL, MediaTypeVideo)`
- Change to: `fd.mediaDownloader.DownloadMediaForFeed(videoAttachment.DownloadableURL, item.Post.ID, MediaTypeVideo)`

### 2. Feed Subcommands Architecture

Create subcommands under `feed`:

- `feed fetch` - Fetch feed and save to file (no video downloads)
- `feed download` - Download videos from saved feed file
- `feed sync` - Fetch feed and download videos (current behavior, becomes default)

**File: `main.go`**

Restructure the feed command (lines 72-134):

- Change `newFeedCmd()` to return a parent command with subcommands
- Create `newFeedFetchCmd()` - fetch only, saves to `--output feed.json`
- Create `newFeedDownloadCmd()` - replay from `--input feed.json`
- Create `newFeedSyncCmd()` - current behavior (fetch + download)
- Make `sync` the default when no subcommand specified

**File: `sora/feed_downloader.go`**

Add new methods:

- `FetchFeedToFile(ctx, outputPath)` - fetch and save feed JSON
- `DownloadFromFile(ctx, feedPath, req)` - load feed from file and download

Update existing `Download()` method to optionally accept pre-loaded feed.

### 3. JSON Export from Database

**File: `sora/database.go`**

Add new query methods:

- `ExportVideosAsJSON(limit int)` - query all/limited videos
- Return data matching feed.json structure with items/post/profile

**File: `main.go`**

Add new subcommand:

- `feed export` - Export downloaded videos as JSON
- Flags: `--output` (default stdout), `--limit` (optional)
- Reads from database, formats as FeedResponse JSON structure

**File: `sora/types.go`**

Add helper method to convert `VideoRecord` to `FeedItem` for export.

## Implementation Details

### Folder Structure After Changes

```
downloads/sora/
├── s_68efcb25841c8191b72bf56904ccc7d6/  # post_id (from feed)
│   ├── video.mp4
│   └── thumbnail.webp
├── task_01k7ma2amre7jsqsx60kpex36s/     # task_id (from scroll crawler)
│   ├── video.mp4
│   └── thumbnail.webp
```

### CLI Usage Examples

```bash
# Fetch feed only (save for later)
./socrawler feed fetch --output feed.json

# Download from saved feed (testing/debugging)
./socrawler feed download --input feed.json --limit 10

# Sync (fetch + download, default behavior)
./socrawler feed sync --limit 50
./socrawler feed --limit 50  # same as sync

# Export downloaded videos as JSON
./socrawler feed export --output my_videos.json
./socrawler feed export --limit 100 > latest_100.json
```

### Database Schema (No Changes Required)

The existing `sora_videos` table already stores `post_id` as primary key and `local_video_path` which will now contain post_id-based paths for feed downloads.

## Files to Modify

1. `sora/downloader.go` - Add `DownloadMedia`
2. `()` method
3. `sora/feed_downloader.go` - Update to use post_id, add file-based methods
4. `sora/database.go` - Add JSON export query methods
5. `sora/types.go` - Add `VideoRecord` to `FeedItem` conversion helper
6. `main.go` - Restructure feed command with subcommands and add export

## Backwards Compatibility

- Scroll-based crawler (`sora/crawler.go`) unchanged, continues using task_id
- Existing feed command behavior preserved as `feed sync` subcommand
- Default `./socrawler feed` maps to `feed sync` for compatibility
- Old downloaded videos (task_id folders) remain accessible

### To-dos

- [ ] Add DownloadMediaForFeed() method to MediaDownloader in sora/downloader.go
- [ ] Update feed_downloader.go to use post_id for folder names and add FetchFeedToFile/DownloadFromFile methods
- [ ] Add ExportVideosAsJSON() and helper methods to sora/database.go
- [ ] Add VideoRecord to FeedItem conversion helper in sora/types.go
- [ ] Restructure main.go to add feed subcommands (fetch/download/sync/export)
- [ ] Test complete workflow: fetch -> download from file -> export to JSON