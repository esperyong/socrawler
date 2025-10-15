package main

import (
	"fmt"

	"github.com/esperyong/socrawler/configs"
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
