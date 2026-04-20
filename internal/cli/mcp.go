package cli

import (
	"github.com/spf13/cobra"

	"github.com/tesserabox/bentotask/internal/mcp"
)

func init() {
	rootCmd.AddCommand(mcpCmd)
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server (for AI assistant integration)",
	Long: `Start a Model Context Protocol (MCP) server on stdio.

This allows AI assistants like Claude to interact with BentoTask
by creating tasks, checking schedules, logging habits, etc.

Configure in Claude Desktop's settings:
  {
    "mcpServers": {
      "bentotask": {
        "command": "bt",
        "args": ["mcp", "--data-dir", "/path/to/tasks"]
      }
    }
  }

The server reads JSON-RPC from stdin and writes responses to stdout.
All logging goes to stderr.`,
	RunE: runMCP,
}

func runMCP(cmd *cobra.Command, _ []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

	srv := mcp.NewServer(a)
	return srv.Run()
}
