package cli

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/tesserabox/bentotask/internal/notion"
	"github.com/tesserabox/bentotask/internal/style"
)

func init() {
	notionCmd.AddCommand(notionImportCmd)
	rootCmd.AddCommand(notionCmd)

	notionImportCmd.Flags().String("token", "", "Notion integration token (required)")
	notionImportCmd.Flags().String("database", "", "Notion database ID (required)")
	notionImportCmd.Flags().Bool("dry-run", false, "Preview import without creating tasks")
	_ = notionImportCmd.MarkFlagRequired("token")
	_ = notionImportCmd.MarkFlagRequired("database")
}

var notionCmd = &cobra.Command{
	Use:   "notion",
	Short: "Notion integration",
}

var notionImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import tasks from a Notion database",
	Long: `Import tasks from a Notion database using the Notion API.

Requires:
  - A Notion internal integration token (create at https://www.notion.so/my-integrations)
  - The integration must be connected to the database
  - The database ID (from the Notion URL)

Property mapping:
  Title property → task title
  Priority (select) → priority
  Date/Due/Deadline → due_date
  Tags/Labels (multi-select) → tags
  Energy/Effort (select) → energy
  Duration/Estimate (number) → estimated_duration

Examples:
  bt notion import --token ntn_xxx --database abc123
  bt notion import --token ntn_xxx --database abc123 --dry-run`,
	RunE: runNotionImport,
}

func runNotionImport(cmd *cobra.Command, _ []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

	token, _ := cmd.Flags().GetString("token")
	databaseID, _ := cmd.Flags().GetString("database")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	client := notion.NewClientWithHTTP(token, http.DefaultClient)

	result, err := notion.ImportDatabase(client, databaseID, a, dryRun)
	if err != nil {
		return fmt.Errorf("notion import: %w", err)
	}

	if isJSON(cmd) {
		return writeJSON(cmd.OutOrStdout(), map[string]any{
			"imported": result.Imported,
			"skipped":  result.Skipped,
			"errors":   result.Errors,
			"dry_run":  dryRun,
		})
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	if quiet {
		cmd.Printf("%d\n", result.Imported)
		return nil
	}

	action := "Imported"
	if dryRun {
		action = "Would import"
	}

	cmd.Printf("%s %s %d tasks from Notion", style.Success("✓"), action, result.Imported)
	if result.Skipped > 0 {
		cmd.Printf(" (%d skipped)", result.Skipped)
	}
	cmd.Println()

	for _, e := range result.Errors {
		cmd.Printf("  %s %s\n", style.ErrorMsg("✗"), e)
	}

	return nil
}
