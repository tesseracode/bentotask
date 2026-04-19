package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tesserabox/bentotask/internal/style"
)

func init() {
	obsidianCmd.AddCommand(obsidianInitCmd)
	rootCmd.AddCommand(obsidianCmd)
}

var obsidianCmd = &cobra.Command{
	Use:     "obsidian",
	Aliases: []string{"obs"},
	Short:   "Obsidian vault integration",
}

var obsidianInitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Set up BentoTask inside an Obsidian vault",
	Long: `Initialize a BentoTask data directory suitable for use as an Obsidian vault.

Creates the folder structure, SQLite index location, and a README file.
After init, run 'bt serve --data-dir <path>' to start using BentoTask.

Examples:
  bt obsidian init ~/Notes/BentoTasks
  bt obs init ./vault/tasks`,
	Args: cobra.ExactArgs(1),
	RunE: runObsidianInit,
}

const obsidianReadme = `# BentoTask

This folder is managed by BentoTask. Tasks are stored as markdown files
with YAML frontmatter.

You can edit these files directly in Obsidian — changes are automatically
picked up by BentoTask.

## Usage

Start the BentoTask server:

` + "```" + `
bt serve --data-dir /path/to/this/folder
` + "```" + `

## File Structure

- ` + "`inbox/`" + ` — default task location
- Other folders act as "boxes" for organizing tasks
- ` + "`.bentotask/`" + ` — BentoTask index (hidden from Obsidian)
`

func runObsidianInit(cmd *cobra.Command, args []string) error {
	target, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Create directories
	dirs := []string{
		target,
		filepath.Join(target, "inbox"),
		filepath.Join(target, ".bentotask"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", d, err)
		}
	}

	// Write README (don't overwrite if exists)
	readmePath := filepath.Join(target, "_README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		if err := os.WriteFile(readmePath, []byte(obsidianReadme), 0o644); err != nil {
			return fmt.Errorf("write readme: %w", err)
		}
	}

	if isJSON(cmd) {
		return writeJSON(cmd.OutOrStdout(), map[string]string{
			"path":    target,
			"inbox":   filepath.Join(target, "inbox"),
			"index":   filepath.Join(target, ".bentotask"),
			"readme":  readmePath,
			"message": "Obsidian vault initialized",
		})
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	if quiet {
		cmd.Println(target)
		return nil
	}

	cmd.Printf("%s Obsidian vault initialized at %s\n\n", style.Success("✓"), style.Bold(target))
	cmd.Printf("  Created: %s\n", style.Dim("inbox/"))
	cmd.Printf("  Created: %s\n", style.Dim(".bentotask/"))
	cmd.Printf("  Created: %s\n\n", style.Dim("_README.md"))
	cmd.Printf("Run: %s\n", style.Bold(fmt.Sprintf("bt serve --data-dir %s", target)))

	return nil
}
