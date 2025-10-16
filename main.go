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
