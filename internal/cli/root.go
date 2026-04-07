// Package cli implements the bt command-line interface.
package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tesserabox/bentotask/internal/style"
)

// version is set at build time via -ldflags.
var version = "dev"

// rootCmd is the top-level command for bt.
var rootCmd = &cobra.Command{
	Use:   "bt",
	Short: "BentoTask — task, habit & routine manager with smart scheduling",
	Long: `BentoTask (bt) is a local-first task, habit, and routine manager.

It stores tasks as plain Markdown files, tracks habits with streaks,
groups actions into routines, and uses a smart scheduling algorithm
to answer: "What should I do now?"

Get started:
  bt add "Buy groceries"          Create a task
  bt list                         List your tasks
  bt done <id>                    Mark a task complete
  bt now                          Get a smart suggestion`,
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
			style.DisableColor()
		}
	},
}

func init() {
	// Cobra defaults cmd.Print/Println to stderr. Force stdout so shell
	// piping (bt list | grep ...) works correctly.
	rootCmd.SetOut(os.Stdout)

	// Global flags (available on every command)
	rootCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Output only IDs (for piping)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")
	rootCmd.PersistentFlags().StringP("data-dir", "d", "", "Path to data directory (default: ~/.bentotask/data)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose/debug output")

	// Add the version command
	rootCmd.AddCommand(versionCmd)
}

// versionCmd prints the version information.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Printf("bt (BentoTask) %s\n", version)
	},
}

// Execute runs the root command. Called from main.
func Execute() error {
	return rootCmd.Execute()
}
