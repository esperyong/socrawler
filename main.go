package main

import (
	"context"
	"fmt"

	"github.com/esperyong/socrawler/configs"
	"github.com/esperyong/socrawler/sora"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "socrawler",
		Short:   "A Sora video crawler with MCP support",
		Long:    `A service for crawling Sora videos with REST API and MCP (Model Context Protocol) support.`,
		Version: version,
	}

	// Add subcommands
	rootCmd.AddCommand(newRunServerCmd())
	rootCmd.AddCommand(newFeedCmd())
	rootCmd.AddCommand(newVersionCmd())

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalf("Command execution failed: %v", err)
	}
}

// newRunServerCmd creates the runserver command
func newRunServerCmd() *cobra.Command {
	var port string
	var debug bool
	var headless bool

	cmd := &cobra.Command{
		Use:   "runserver",
		Short: "Start the server",
		Long:  `Start the server with HTTP API and MCP support`,
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}

			// Initialize headless mode
			configs.InitHeadless(headless)
			logrus.Infof("Browser headless mode: %v", headless)

			logrus.Infof("Starting server on port %s", port)

			// Create and start the server
			server := NewAppServer()
			if err := server.Start(":" + port); err != nil {
				logrus.Fatalf("Failed to start server: %v", err)
			}
		},
	}

	cmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	cmd.Flags().BoolVar(&headless, "headless", true, "Run browser in headless mode")

	return cmd
}

// newFeedCmd creates the feed command with subcommands
func newFeedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feed",
		Short: "Manage Sora feed videos",
		Long:  `Fetch, download, sync, and export Sora videos from the public feed endpoint.`,
	}

	// Add subcommands
	cmd.AddCommand(newFeedFetchCmd())
	cmd.AddCommand(newFeedDownloadCmd())
	cmd.AddCommand(newFeedSyncCmd())
	cmd.AddCommand(newFeedExportCmd())
	cmd.AddCommand(newFeedUploadGoldcastCmd())

	// Set default subcommand to sync
	cmd.Run = func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, run sync
		syncCmd := newFeedSyncCmd()
		syncCmd.Run(cmd, args)
	}

	return cmd
}

// newFeedFetchCmd creates the feed fetch subcommand
func newFeedFetchCmd() *cobra.Command {
	var output string
	var debug bool
	var headless bool

	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch feed and save to file",
		Long:  `Fetch the Sora feed from the API and save it to a JSON file without downloading videos.`,
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			} else {
				logrus.SetLevel(logrus.InfoLevel)
			}

			logrus.Infof("Fetching feed to: %s", output)

			ctx := context.Background()
			if err := sora.FetchFeedToFile(ctx, output, headless); err != nil {
				logrus.Fatalf("Feed fetch failed: %v", err)
			}

			fmt.Printf("\nFeed saved to: %s\n", output)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "feed.json", "Output file path for feed JSON")
	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	cmd.Flags().BoolVar(&headless, "headless", true, "Run browser in headless mode")

	return cmd
}

// newFeedDownloadCmd creates the feed download subcommand
func newFeedDownloadCmd() *cobra.Command {
	var input string
	var savePath string
	var dbPath string
	var limit int
	var debug bool

	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download videos from a saved feed file",
		Long:  `Download videos from a previously fetched and saved feed JSON file.`,
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			} else {
				logrus.SetLevel(logrus.InfoLevel)
			}

			logrus.Infof("Downloading from feed file: %s", input)

			req := &sora.FeedDownloadRequest{
				SavePath: savePath,
				DBPath:   dbPath,
				Limit:    limit,
				Headless: false, // No browser needed when loading from file
			}

			ctx := context.Background()
			result, err := sora.DownloadFromFile(ctx, input, req)
			if err != nil {
				logrus.Fatalf("Download from file failed: %v", err)
			}

			// Print results
			fmt.Println("\n========================================")
			fmt.Println("  Feed Download Results")
			fmt.Println("========================================")
			fmt.Printf("Total items in feed:     %d\n", result.TotalFetched)
			fmt.Printf("New videos found:        %d\n", result.NewVideos)
			fmt.Printf("Successfully downloaded: %d\n", result.Downloaded)
			fmt.Printf("Skipped:                 %d\n", result.Skipped)
			fmt.Printf("Failed:                  %d\n", result.Failed)
			fmt.Printf("Duration:                %d seconds\n", result.DurationSeconds)
			fmt.Println("========================================")

			if result.Downloaded > 0 {
				fmt.Printf("\nVideos saved to: %s\n", savePath)
				fmt.Printf("Database: %s\n", dbPath)
			}
		},
	}

	cmd.Flags().StringVarP(&input, "input", "i", "feed.json", "Input feed JSON file")
	cmd.Flags().StringVar(&savePath, "save-path", "./downloads/sora", "Directory to save downloaded videos")
	cmd.Flags().StringVar(&dbPath, "db-path", "./sora.db", "Path to SQLite database for tracking downloads")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of videos to download")
	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")

	return cmd
}

// newFeedSyncCmd creates the feed sync subcommand (default behavior)
func newFeedSyncCmd() *cobra.Command {
	var savePath string
	var dbPath string
	var limit int
	var debug bool
	var headless bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Fetch feed and download videos (default)",
		Long:  `Fetch the Sora feed and download new videos in one operation. This is the default behavior.`,
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			} else {
				logrus.SetLevel(logrus.InfoLevel)
			}

			logrus.Infof("Starting feed sync: save_path=%s, db_path=%s, limit=%d, headless=%v",
				savePath, dbPath, limit, headless)

			// Create request
			req := &sora.FeedDownloadRequest{
				SavePath: savePath,
				DBPath:   dbPath,
				Limit:    limit,
				Headless: headless,
			}

			// Execute download
			ctx := context.Background()
			result, err := sora.DownloadFromFeed(ctx, req)
			if err != nil {
				logrus.Fatalf("Feed sync failed: %v", err)
			}

			// Print results
			fmt.Println("\n========================================")
			fmt.Println("  Feed Sync Results")
			fmt.Println("========================================")
			fmt.Printf("Total items fetched:     %d\n", result.TotalFetched)
			fmt.Printf("New videos found:        %d\n", result.NewVideos)
			fmt.Printf("Successfully downloaded: %d\n", result.Downloaded)
			fmt.Printf("Skipped:                 %d\n", result.Skipped)
			fmt.Printf("Failed:                  %d\n", result.Failed)
			fmt.Printf("Duration:                %d seconds\n", result.DurationSeconds)
			fmt.Println("========================================")

			if result.Downloaded > 0 {
				fmt.Printf("\nVideos saved to: %s\n", savePath)
				fmt.Printf("Database: %s\n", dbPath)
			}
		},
	}

	cmd.Flags().StringVar(&savePath, "save-path", "./downloads/sora", "Directory to save downloaded videos")
	cmd.Flags().StringVar(&dbPath, "db-path", "./sora.db", "Path to SQLite database for tracking downloads")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of videos to download per run")
	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	cmd.Flags().BoolVar(&headless, "headless", true, "Run browser in headless mode")

	return cmd
}

// newFeedExportCmd creates the feed export subcommand
func newFeedExportCmd() *cobra.Command {
	var output string
	var dbPath string
	var limit int
	var debug bool

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export downloaded videos as JSON",
		Long:  `Export videos from the database in feed JSON format. Use "-" or omit --output for stdout.`,
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			} else {
				logrus.SetLevel(logrus.InfoLevel)
			}

			// Open database
			db, err := sora.NewVideoDatabase(dbPath)
			if err != nil {
				logrus.Fatalf("Failed to open database: %v", err)
			}
			defer db.Close()

			// Export to JSON
			if err := db.ExportVideosAsJSON(limit, output); err != nil {
				logrus.Fatalf("Export failed: %v", err)
			}

			if output != "" && output != "-" {
				fmt.Printf("\nExported videos to: %s\n", output)
			}
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (use '-' or omit for stdout)")
	cmd.Flags().StringVar(&dbPath, "db-path", "./sora.db", "Path to SQLite database")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of videos to export (0 = all)")
	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")

	return cmd
}

// newFeedUploadGoldcastCmd creates the feed uploadgoldcast subcommand
func newFeedUploadGoldcastCmd() *cobra.Command {
	var apiKey string
	var apiURL string
	var dbPath string
	var limit int
	var debug bool
	var ossAccessKeyID string
	var ossAccessKeySecret string
	var ossBucketName string
	var ossEndpoint string
	var ossRegion string

	cmd := &cobra.Command{
		Use:   "uploadgoldcast",
		Short: "Upload videos to Goldcast media CMS",
		Long:  `Upload downloaded videos to Goldcast media CMS via API. Only uploads videos that haven't been uploaded yet.`,
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			} else {
				logrus.SetLevel(logrus.InfoLevel)
			}

			logrus.Infof("Starting Goldcast upload: db_path=%s, limit=%d", dbPath, limit)

			// Create Goldcast config (with environment variable fallback)
			config := sora.NewGoldcastConfigFromEnv(apiKey, apiURL)

			// Create OSS config from command-line flags or environment variables
			var ossConfig *sora.OSSConfig
			var err error

			// Check if command-line flags were provided
			if ossAccessKeyID != "" || ossAccessKeySecret != "" {
				// Use command-line flags (with defaults)
				if ossAccessKeyID == "" {
					logrus.Fatal("OSS Access Key ID is required (use --oss-access-key-id or set OSS_ACCESS_KEY_ID)")
				}
				if ossAccessKeySecret == "" {
					logrus.Fatal("OSS Access Key Secret is required (use --oss-access-key-secret or set OSS_ACCESS_KEY_SECRET)")
				}

				ossConfig = &sora.OSSConfig{
					AccessKeyID:     ossAccessKeyID,
					AccessKeySecret: ossAccessKeySecret,
					BucketName:      ossBucketName,
					Endpoint:        ossEndpoint,
					Region:          ossRegion,
				}
			} else {
				// Fall back to environment variables
				ossConfig, err = sora.NewOSSConfigFromEnv()
				if err != nil {
					logrus.Fatalf("Failed to load OSS configuration: %v\n\nPlease provide OSS credentials via:\n  1. Command-line flags: --oss-access-key-id and --oss-access-key-secret\n  2. Environment variables: OSS_ACCESS_KEY_ID and OSS_ACCESS_KEY_SECRET\n\nOptional (have defaults):\n  - OSS_BUCKET_NAME (default: dreammedias)\n  - OSS_ENDPOINT (default: oss-cn-beijing.aliyuncs.com)\n  - OSS_REGION (default: cn-beijing)", err)
				}
			}

			// Open database
			db, err := sora.NewVideoDatabase(dbPath)
			if err != nil {
				logrus.Fatalf("Failed to open database: %v", err)
			}
			defer db.Close()

			// Get upload stats before upload
			uploadedBefore, unuploadedBefore, err := db.GetUploadStats()
			if err != nil {
				logrus.Fatalf("Failed to get upload stats: %v", err)
			}

			logrus.Infof("Current status: %d uploaded, %d not uploaded", uploadedBefore, unuploadedBefore)

			// Create uploader
			uploader, err := sora.NewGoldcastUploader(config, db, ossConfig)
			if err != nil {
				logrus.Fatalf("Failed to create uploader: %v", err)
			}

			// Upload videos
			ctx := context.Background()
			result, err := uploader.UploadUnuploadedVideos(ctx, limit)
			if err != nil {
				logrus.Fatalf("Upload failed: %v", err)
			}

			// Print results
			fmt.Println("\n========================================")
			fmt.Println("  Goldcast Upload Results")
			fmt.Println("========================================")
			fmt.Printf("Total unuploaded videos: %d\n", result.TotalUnuploaded)
			fmt.Printf("Attempted:               %d\n", result.Attempted)
			fmt.Printf("Successfully uploaded:   %d\n", result.Succeeded)
			fmt.Printf("Failed:                  %d\n", result.Failed)
			fmt.Printf("Duration:                %d seconds\n", result.DurationSeconds)

			if len(result.FailedPostIDs) > 0 {
				fmt.Println("\nFailed post IDs:")
				for _, postID := range result.FailedPostIDs {
					fmt.Printf("  - %s\n", postID)
				}
			}

			fmt.Println("========================================")

			if result.Succeeded > 0 {
				fmt.Printf("\n✓ Successfully uploaded %d videos to Goldcast\n", result.Succeeded)
			}

			if result.Failed > 0 {
				fmt.Printf("\n✗ %d videos failed to upload\n", result.Failed)
			}
		},
	}

	cmd.Flags().StringVar(&apiKey, "api-key", "", "Goldcast API key (env: GOLDCAST_API_KEY, default: ucHZBRJ1.w8njpEorJlDgjp0ESnw0qSyOkN6V6VUe)")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "Goldcast API URL (env: GOLDCAST_API_URL, default: https://financial.xiaoyequ9.com/api/v1/external/media/upload)")
	cmd.Flags().StringVar(&dbPath, "db-path", "./sora.db", "Path to SQLite database")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of videos to upload (0 = all unuploaded)")
	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")

	// OSS configuration flags
	cmd.Flags().StringVar(&ossAccessKeyID, "oss-access-key-id", "", "OSS Access Key ID (env: OSS_ACCESS_KEY_ID, required)")
	cmd.Flags().StringVar(&ossAccessKeySecret, "oss-access-key-secret", "", "OSS Access Key Secret (env: OSS_ACCESS_KEY_SECRET, required)")
	cmd.Flags().StringVar(&ossBucketName, "oss-bucket-name", "dreammedias", "OSS Bucket Name (env: OSS_BUCKET_NAME)")
	cmd.Flags().StringVar(&ossEndpoint, "oss-endpoint", "oss-cn-beijing.aliyuncs.com", "OSS Endpoint (env: OSS_ENDPOINT)")
	cmd.Flags().StringVar(&ossRegion, "oss-region", "cn-beijing", "OSS Region (env: OSS_REGION)")

	return cmd
}

// newVersionCmd creates the version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gomcp version %s\n", version)
		},
	}
}
