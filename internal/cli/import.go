package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/style"
)

func init() {
	importCmd.AddCommand(importTodoistCmd)
	importCmd.AddCommand(importTaskwarriorCmd)
	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import tasks from external sources",
}

// --- Todoist CSV import ---

var importTodoistCmd = &cobra.Command{
	Use:   "todoist <file.csv>",
	Short: "Import tasks from Todoist CSV export",
	Long: `Import tasks from a Todoist CSV export file.

Todoist CSV columns: TYPE, CONTENT, DESCRIPTION, PRIORITY, INDENT,
AUTHOR, RESPONSIBLE, DATE, DATE_LANG, TIMEZONE

Priority mapping: 1→urgent, 2→high, 3→medium, 4→low

Examples:
  bt import todoist ~/Downloads/todoist-export.csv`,
	Args: cobra.ExactArgs(1),
	RunE: runImportTodoist,
}

func runImportTodoist(cmd *cobra.Command, args []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

	f, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	reader := csv.NewReader(f)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	colIndex := make(map[string]int)
	for i, h := range header {
		colIndex[strings.ToUpper(strings.TrimSpace(h))] = i
	}

	imported := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read row: %w", err)
		}

		// Skip section rows
		typeCol := getCol(row, colIndex, "TYPE")
		if typeCol == "section" || typeCol == "" {
			continue
		}

		title := getCol(row, colIndex, "CONTENT")
		if title == "" {
			continue
		}

		opts := app.TaskOptions{}

		// Description → body (stored in task creation isn't supported via opts,
		// but we can use it as the title suffix for now)
		// Priority mapping: Todoist 1=urgent, 2=high, 3=medium, 4=low
		if p := getCol(row, colIndex, "PRIORITY"); p != "" {
			if n, err := strconv.Atoi(p); err == nil {
				switch n {
				case 1:
					opts.Priority = model.PriorityUrgent
				case 2:
					opts.Priority = model.PriorityHigh
				case 3:
					opts.Priority = model.PriorityMedium
				case 4:
					opts.Priority = model.PriorityLow
				}
			}
		}

		// Date parsing
		if dateStr := getCol(row, colIndex, "DATE"); dateStr != "" {
			if parsed := parseTodoistDate(dateStr); parsed != "" {
				opts.DueDate = parsed
			}
		}

		// Description → body
		if desc := getCol(row, colIndex, "DESCRIPTION"); desc != "" {
			opts.Body = desc
		}

		if _, err := a.AddTask(title, opts); err != nil {
			cmd.PrintErrf("Warning: failed to import %q: %v\n", title, err)
			continue
		}
		imported++
	}

	if isJSON(cmd) {
		return writeJSON(cmd.OutOrStdout(), map[string]int{"imported": imported})
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	if quiet {
		cmd.Printf("%d\n", imported)
		return nil
	}

	cmd.Printf("%s Imported %d tasks from Todoist\n", style.Success("✓"), imported)
	return nil
}

func getCol(row []string, index map[string]int, col string) string {
	i, ok := index[col]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func parseTodoistDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// Try YYYY-MM-DD
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return s
	}

	// Try "Jan 2 2006" / "January 2 2006"
	formats := []string{
		"Jan 2 2006",
		"January 2 2006",
		"Jan 2, 2006",
		"January 2, 2006",
		"2 Jan 2006",
	}
	for _, fmt := range formats {
		if t, err := time.Parse(fmt, s); err == nil {
			return t.Format("2006-01-02")
		}
	}

	return "" // unparseable
}

// --- Taskwarrior JSON import ---

var importTaskwarriorCmd = &cobra.Command{
	Use:   "taskwarrior <file.json>",
	Short: "Import tasks from Taskwarrior JSON export",
	Long: `Import tasks from a Taskwarrior JSON export file.

Run 'task export' in Taskwarrior to generate the JSON file.

Field mapping:
  description → title
  priority (H/M/L) → priority (high/medium/low)
  due → due_date
  tags → tags
  project → box
  annotations → body

Examples:
  task export > tasks.json
  bt import taskwarrior tasks.json`,
	Args: cobra.ExactArgs(1),
	RunE: runImportTaskwarrior,
}

type taskwarriorTask struct {
	Description string                  `json:"description"`
	Priority    string                  `json:"priority"`
	Due         string                  `json:"due"`
	Tags        []string                `json:"tags"`
	Project     string                  `json:"project"`
	Status      string                  `json:"status"`
	Annotations []taskwarriorAnnotation `json:"annotations"`
}

type taskwarriorAnnotation struct {
	Description string `json:"description"`
}

func runImportTaskwarrior(cmd *cobra.Command, args []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var twTasks []taskwarriorTask
	if err := json.Unmarshal(data, &twTasks); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	imported := 0
	for _, tw := range twTasks {
		if tw.Description == "" {
			continue
		}

		opts := app.TaskOptions{}

		// Priority mapping
		switch strings.ToUpper(tw.Priority) {
		case "H":
			opts.Priority = model.PriorityHigh
		case "M":
			opts.Priority = model.PriorityMedium
		case "L":
			opts.Priority = model.PriorityLow
		}

		// Due date (ISO8601 → YYYY-MM-DD)
		if tw.Due != "" {
			if t, err := time.Parse("20060102T150405Z", tw.Due); err == nil {
				opts.DueDate = t.Format("2006-01-02")
			}
		}

		// Tags
		if len(tw.Tags) > 0 {
			opts.Tags = tw.Tags
		}

		// Project → box
		if tw.Project != "" {
			opts.Box = tw.Project
		}

		// Annotations → body
		if len(tw.Annotations) > 0 {
			var lines []string
			for _, ann := range tw.Annotations {
				lines = append(lines, ann.Description)
			}
			opts.Body = strings.Join(lines, "\n")
		}

		task, err := a.AddTask(tw.Description, opts)
		if err != nil {
			cmd.PrintErrf("Warning: failed to import %q: %v\n", tw.Description, err)
			continue
		}

		// If completed/deleted, mark accordingly
		if tw.Status == "completed" {
			_, _ = a.CompleteTask(task.ID)
		}

		imported++
	}

	if isJSON(cmd) {
		return writeJSON(cmd.OutOrStdout(), map[string]int{"imported": imported})
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	if quiet {
		cmd.Printf("%d\n", imported)
		return nil
	}

	cmd.Printf("%s Imported %d tasks from Taskwarrior\n", style.Success("✓"), imported)
	return nil
}
