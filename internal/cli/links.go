package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/style"
)

func init() {
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(unlinkCmd)

	linkCmd.Flags().StringP("type", "t", "related-to", "Link type: depends-on, blocks, related-to")
	unlinkCmd.Flags().StringP("type", "t", "related-to", "Link type: depends-on, blocks, related-to")

	// Flag completions
	_ = linkCmd.RegisterFlagCompletionFunc("type", completeLinkType)
	_ = unlinkCmd.RegisterFlagCompletionFunc("type", completeLinkType)

	// Task ID completions for positional args
	linkCmd.ValidArgsFunction = completeTaskIDs
	unlinkCmd.ValidArgsFunction = completeTaskIDs
}

// completeLinkType returns valid link type values for shell completion.
func completeLinkType(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"depends-on\tThis task depends on the target",
		"blocks\tThis task blocks the target",
		"related-to\tInformational relationship",
	}, cobra.ShellCompDirectiveNoFileComp
}

// --- bt link ---

var linkCmd = &cobra.Command{
	Use:   "link <source-id> <target-id>",
	Short: "Link two tasks together",
	Long: `Create a relationship between two tasks.

Link types:
  depends-on  — Source depends on target (target must be done first)
  blocks      — Source blocks target (target can't start until source is done)
  related-to  — Informational link (no scheduling impact)

Cycle detection prevents circular depends-on and blocks chains.

Examples:
  bt link 01JQX 01JQY                       Related-to (default)
  bt link 01JQX 01JQY -t depends-on         Source depends on target
  bt link 01JQX 01JQY -t blocks             Source blocks target`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		linkType, _ := cmd.Flags().GetString("type")
		lt := model.LinkType(linkType)

		source, target, err := a.LinkTasks(args[0], args[1], lt)
		if err != nil {
			return err
		}

		if isJSON(cmd) {
			return writeJSON(cmd.OutOrStdout(), map[string]any{
				"source":    source.ShortID(8),
				"target":    target.ShortID(8),
				"link_type": string(lt),
				"source_id": source.ID,
				"target_id": target.ID,
			})
		}

		quiet, _ := cmd.Flags().GetBool("quiet")
		if quiet {
			cmd.Printf("%s %s\n", source.ID, target.ID)
			return nil
		}

		cmd.Printf("%s %s %s %s\n",
			style.Success("Linked:"),
			source.Title,
			style.Dim(fmt.Sprintf("—[%s]→", lt)),
			target.Title)

		return nil
	},
}

// --- bt unlink ---

var unlinkCmd = &cobra.Command{
	Use:   "unlink <source-id> <target-id>",
	Short: "Remove a link between two tasks",
	Long: `Remove an existing relationship between two tasks.

Examples:
  bt unlink 01JQX 01JQY                     Remove related-to link (default)
  bt unlink 01JQX 01JQY -t depends-on       Remove depends-on link`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		linkType, _ := cmd.Flags().GetString("type")
		lt := model.LinkType(linkType)

		source, target, err := a.UnlinkTasks(args[0], args[1], lt)
		if err != nil {
			return err
		}

		if isJSON(cmd) {
			return writeJSON(cmd.OutOrStdout(), map[string]any{
				"source":    source.ShortID(8),
				"target":    target.ShortID(8),
				"link_type": string(lt),
				"source_id": source.ID,
				"target_id": target.ID,
				"removed":   true,
			})
		}

		quiet, _ := cmd.Flags().GetBool("quiet")
		if quiet {
			cmd.Printf("%s %s\n", source.ID, target.ID)
			return nil
		}

		cmd.Printf("%s %s %s %s\n",
			style.Success("Unlinked:"),
			source.Title,
			style.Dim(fmt.Sprintf("—[%s]→", lt)),
			target.Title)

		return nil
	},
}

// --- Link display helpers ---

// LinkDisplay holds the data needed to render a link in CLI output.
type LinkDisplay struct {
	Type      string
	Direction string
	TaskID    string
	TaskTitle string
}

// renderLinks formats task links for display in bt task show / bt routine show.
func renderLinks(cmd *cobra.Command, links []LinkDisplay) {
	if len(links) == 0 {
		return
	}

	cmd.Printf("\n%s\n", style.Bold("Links:"))
	for _, l := range links {
		arrow := "→"
		direction := ""
		if l.Direction == "incoming" {
			arrow = "←"
			direction = " " + style.Dim("(incoming)")
		}

		shortID := l.TaskID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		cmd.Printf("  %s %s %s %s%s\n",
			style.Dim(fmt.Sprintf("[%s]", l.Type)),
			arrow,
			style.Dim(shortID),
			l.TaskTitle,
			direction)
	}
}

// linksToJSON converts link info to JSON-serializable format.
func linksToJSON(links []LinkDisplay) []map[string]string {
	result := make([]map[string]string, len(links))
	for i, l := range links {
		result[i] = map[string]string{
			"type":       l.Type,
			"direction":  l.Direction,
			"task_id":    l.TaskID,
			"task_title": l.TaskTitle,
		}
	}
	return result
}
