<!-- b2763fdd-7080-49c9-8a11-91f5c920f7ed 31acbf32-2f04-4cb7-bc52-f382428b9ce0 -->
# Feed-based Sora Video Downloader

## Overview

Create a new command-line tool that fetches Sora videos from the public feed endpoint (`https://sora.chatgpt.com/backend/public/nf2/feed`) using browser stealth mode, parses the feed JSON, and downloads new videos with metadata tracking.

## Implementation Plan

### 1. Database Schema (SQLite)

Create `sora/database.go` to manage video metadata:

- Table: `sora_videos` with columns:
  - `post_id` (primary key) - unique identifier from feed
  - `generation_id` - video generation ID
  - `video_url` - download URL
  - `thumbnail_url` - thumbnail URL
  - `text` - prompt/description text
  - `username` - author username
  - `user_id` - author user ID
  - `posted_at` - timestamp
  - `downloaded_at` - when we downloaded it
  - `local_video_path` - where video is saved
  - `local_thumbnail_path` - where thumbnail is saved
  - `width`, `height` - video dimensions

### 2. Feed Fetcher (sora/feed_fetcher.go)

Create a new module to fetch and parse the feed:

- Use browser with stealth mode to access feed endpoint
- Navigate to `https://sora.chatgpt.com/backend/public/nf2/feed`
- Extract response body (JSON)
- Parse into Go structs matching the feed structure
- Key structs:
  - `FeedResponse` with `Items []FeedItem`
  - `FeedItem` with `Post` and `Profile`
  - `Post` with `ID`, `Text`, `PostedAt`, `Attachments`
  - `Attachment` with `URL`, `DownloadableURL`, `Width`, `Height`, `Encodings`

### 3. Feed-based Downloader (sora/feed_downloader.go)

Create new downloader that:

- Queries database for existing post IDs to avoid duplicates
- Filters feed items to only new videos (not in DB)
- Downloads videos and thumbnails using existing `MediaDownloader`
- Saves metadata to database after successful download
- Uses task_id from URL for folder structure (existing logic)

### 4. New CLI Command (main.go)

Add new subcommand `feed`:

```bash
./socrawler feed --save-path ./downloads/sora --db-path ./sora.db [--limit 50]
```

- Flags:
  - `--save-path`: where to save videos (default: ./downloads/sora)
  - `--db-path`: SQLite database path (default: ./sora.db)
  - `--limit`: max videos to download per run (default: 50)
- Can be called by cron or external scheduler

### 5. File Structure

```
sora/
├── crawler.go          (existing - unchanged)
├── downloader.go       (existing - reuse MediaDownloader)
├── types.go            (existing - extend with feed types)
├── database.go         (NEW - SQLite operations)
├── feed_fetcher.go     (NEW - fetch & parse feed)
├── feed_downloader.go  (NEW - download from feed)
```

### 6. Key Design Decisions

- **Browser access**: Use existing `browser.NewCleanBrowser()` with stealth mode
- **Deduplication**: SQLite DB tracks post_id to prevent re-downloads
- **Metadata**: Store author info, timestamps, prompts with each video
- **Backwards compatible**: Old scroll-based crawler remains unchanged
- **Standalone**: New command works independently, can be scheduled externally

### 7. Example Usage

```bash
# One-time download
./socrawler feed --save-path ./downloads/sora --limit 100

# Schedule with cron (every 6 hours)
0 */6 * * * cd /path/to/socrawler && ./socrawler feed --save-path ./downloads/sora
```

### 8. Integration with run_service.sh

Add new commands to `run_service.sh`:

- `./run_service.sh feed` - Run feed downloader once
- `./run_service.sh feed --limit=100` - Download up to 100 videos
- `./run_service.sh feed --headless=false` - Run with visible browser for debugging

This provides:

- Same convenient interface as existing commands
- Consistent parameter handling (headless mode, paths)
- Easy testing and scheduling integration

### 9. Enhanced Example Usage

```bash
# Direct command-line usage
./socrawler feed --save-path ./downloads/sora --db-path ./sora.db --limit 100

# Using run_service.sh wrapper (recommended)
./run_service.sh feed
./run_service.sh feed --limit=50
./run_service.sh feed --headless=false  # For debugging

# Schedule with cron (every 6 hours)
0 */6 * * * cd /path/to/socrawler && ./run_service.sh feed --limit=100
```

### 10. Implementation Order

1. Define feed JSON structs in `sora/types.go`
2. Implement SQLite database layer in `sora/database.go`
3. Create feed fetcher in `sora/feed_fetcher.go`
4. Build feed downloader in `sora/feed_downloader.go`
5. Add CLI command in `main.go`
6. Integrate feed command into `run_service.sh`
7. Test end-to-end flow

### To-dos

- [ ] Define feed JSON structs in sora/types.go (FeedResponse, FeedItem, Post, Profile, Attachment, Encodings)
- [ ] Implement SQLite database layer in sora/database.go with video metadata table and CRUD operations
- [ ] Create feed fetcher in sora/feed_fetcher.go to fetch and parse feed using browser stealth mode
- [ ] Build feed downloader in sora/feed_downloader.go integrating database, fetcher, and existing MediaDownloader
- [ ] Add new 'feed' subcommand in main.go with flags for save-path, db-path, and limit
- [ ] Test complete workflow: fetch feed → filter new videos → download → save metadata