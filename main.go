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

// newFeedCmd creates the feed command
func newFeedCmd() *cobra.Command {
	var savePath string
	var dbPath string
	var limit int
	var debug bool
	var headless bool

	cmd := &cobra.Command{
		Use:   "feed",
		Short: "Download videos from Sora feed",
		Long:  `Download new Sora videos from the public feed endpoint. Uses SQLite database for deduplication.`,
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			} else {
				logrus.SetLevel(logrus.InfoLevel)
			}

			logrus.Infof("Starting feed download: save_path=%s, db_path=%s, limit=%d, headless=%v",
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
				logrus.Fatalf("Feed download failed: %v", err)
			}

			// Print results
			fmt.Println("\n========================================")
			fmt.Println("  Feed Download Results")
			fmt.Println("========================================")
			fmt.Printf("Total items fetched:    %d\n", result.TotalFetched)
			fmt.Printf("New videos found:       %d\n", result.NewVideos)
			fmt.Printf("Successfully downloaded: %d\n", result.Downloaded)
			fmt.Printf("Skipped:                %d\n", result.Skipped)
			fmt.Printf("Failed:                 %d\n", result.Failed)
			fmt.Printf("Duration:               %d seconds\n", result.DurationSeconds)
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
