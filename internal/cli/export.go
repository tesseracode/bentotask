package cli

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/store"
)

func init() {
	exportCmd.AddCommand(exportJSONCmd)
	exportCmd.AddCommand(exportCSVCmd)
	rootCmd.AddCommand(exportCmd)

	for _, cmd := range []*cobra.Command{exportJSONCmd, exportCSVCmd} {
		cmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
		cmd.Flags().StringP("status", "s", "", "Filter by status")
		cmd.Flags().StringP("priority", "p", "", "Filter by priority")
		cmd.Flags().StringP("energy", "e", "", "Filter by energy")
		cmd.Flags().String("tag", "", "Filter by tag")
		cmd.Flags().StringP("box", "b", "", "Filter by box")
	}
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export tasks in various formats",
}

var exportJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Export tasks as JSON",
	Long: `Export tasks as a JSON array. Supports the same filters as 'bt list'.

Examples:
  bt export json                         Export all tasks to stdout
  bt export json -o tasks.json           Export to file
  bt export json --status pending        Export only pending tasks
  bt export json --tag work -p high      Export high-priority work tasks`,
	RunE: runExportJSON,
}

var exportCSVCmd = &cobra.Command{
	Use:   "csv",
	Short: "Export tasks as CSV",
	Long: `Export tasks as CSV with headers. Supports the same filters as 'bt list'.

Examples:
  bt export csv                          Export all tasks to stdout
  bt export csv -o tasks.csv             Export to file
  bt export csv --status pending         Export only pending tasks`,
	RunE: runExportCSV,
}

func buildExportFilter(cmd *cobra.Command) *store.TaskFilter {
	f := &store.TaskFilter{}
	if v, _ := cmd.Flags().GetString("status"); v != "" {
		f.Status = model.Status(v)
	}
	if v, _ := cmd.Flags().GetString("priority"); v != "" {
		f.Priority = model.Priority(v)
	}
	if v, _ := cmd.Flags().GetString("energy"); v != "" {
		f.Energy = model.Energy(v)
	}
	if v, _ := cmd.Flags().GetString("tag"); v != "" {
		f.Tag = v
	}
	if v, _ := cmd.Flags().GetString("box"); v != "" {
		f.Box = v
	}
	return f
}

func runExportJSON(cmd *cobra.Command, _ []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

	tasks, err := a.ListTasks(buildExportFilter(cmd))
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	items := make([]TaskJSON, len(tasks))
	for i, t := range tasks {
		items[i] = indexedToJSON(t)
	}

	outFile, _ := cmd.Flags().GetString("output")
	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer func() { _ = f.Close() }()
		return writeJSON(f, items)
	}

	return writeJSON(cmd.OutOrStdout(), items)
}

func runExportCSV(cmd *cobra.Command, _ []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

	tasks, err := a.ListTasks(buildExportFilter(cmd))
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	outFile, _ := cmd.Flags().GetString("output")
	var w *csv.Writer
	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer func() { _ = f.Close() }()
		w = csv.NewWriter(f)
	} else {
		w = csv.NewWriter(cmd.OutOrStdout())
	}
	defer w.Flush()

	// Header
	header := []string{"id", "title", "type", "status", "priority", "energy", "estimated_duration", "due_date", "tags", "contexts", "created_at", "updated_at", "completed_at"}
	if err := w.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, t := range tasks {
		dur := ""
		if t.EstimatedDuration != nil {
			dur = fmt.Sprintf("%d", *t.EstimatedDuration)
		}
		priority := ""
		if t.Priority != nil {
			priority = *t.Priority
		}
		energy := ""
		if t.Energy != nil {
			energy = *t.Energy
		}
		dueDate := ""
		if t.DueDate != nil {
			dueDate = *t.DueDate
		}
		completedAt := ""
		if t.CompletedAt != nil {
			completedAt = *t.CompletedAt
		}

		row := []string{
			t.ID,
			t.Title,
			t.Type,
			t.Status,
			priority,
			energy,
			dur,
			dueDate,
			strings.Join(t.Tags, ";"),
			strings.Join(t.Contexts, ";"),
			t.CreatedAt,
			t.UpdatedAt,
			completedAt,
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}

	return nil
}
