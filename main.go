package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "gomcp",
		Short:   "A Go-based MCP (Model Context Protocol) server scaffold",
		Long:    `A scaffold for building MCP servers in Go with built-in HTTP and JSON-RPC support.`,
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

	cmd := &cobra.Command{
		Use:   "runserver",
		Short: "Start the MCP server",
		Long:  `Start the MCP server with HTTP and JSON-RPC support`,
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}

			logrus.Infof("Starting MCP server on port %s", port)

			// Create and start the server
			server := NewAppServer()
			if err := server.Start(":" + port); err != nil {
				logrus.Fatalf("Failed to start server: %v", err)
			}
		},
	}

	cmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
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
