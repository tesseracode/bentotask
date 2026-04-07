package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/store"
)

func init() {
	// Register task subcommands
	taskCmd.AddCommand(taskAddCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskDoneCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskEditCmd)

	rootCmd.AddCommand(taskCmd)

	// Top-level aliases per ADR-003
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(doneCmd)

	// --- Flags ---

	// bt task add / bt add
	for _, cmd := range []*cobra.Command{taskAddCmd, addCmd} {
		cmd.Flags().StringP("priority", "p", "", "Priority: none, low, medium, high, urgent")
		cmd.Flags().StringP("energy", "e", "", "Energy: low, medium, high")
		cmd.Flags().Int("duration", 0, "Estimated duration in minutes")
		cmd.Flags().String("due", "", "Due date (YYYY-MM-DD)")
		cmd.Flags().String("due-start", "", "Due window start (YYYY-MM-DD)")
		cmd.Flags().String("due-end", "", "Due window end (YYYY-MM-DD)")
		cmd.Flags().StringSlice("tag", nil, "Tags (repeatable)")
		cmd.Flags().StringP("context", "c", "", "Context: home, office, errands, anywhere")
		cmd.Flags().StringP("box", "b", "", "Box/project path")
	}

	// bt task edit
	taskEditCmd.Flags().String("title", "", "New title")
	taskEditCmd.Flags().StringP("priority", "p", "", "Priority: none, low, medium, high, urgent")
	taskEditCmd.Flags().StringP("energy", "e", "", "Energy: low, medium, high")
	taskEditCmd.Flags().Int("duration", 0, "Estimated duration in minutes")
	taskEditCmd.Flags().String("due", "", "Due date (YYYY-MM-DD)")
	taskEditCmd.Flags().String("due-start", "", "Due window start (YYYY-MM-DD)")
	taskEditCmd.Flags().String("due-end", "", "Due window end (YYYY-MM-DD)")
	taskEditCmd.Flags().StringSlice("tag", nil, "Replace tags (repeatable)")
	taskEditCmd.Flags().StringP("context", "c", "", "Context: home, office, errands, anywhere")
	taskEditCmd.Flags().StringP("box", "b", "", "Box/project path")
	taskEditCmd.Flags().StringP("status", "s", "", "Status: pending, active, paused, done, cancelled, waiting")

	// bt task list / bt list
	for _, cmd := range []*cobra.Command{taskListCmd, listCmd} {
		cmd.Flags().StringP("status", "s", "", "Filter by status")
		cmd.Flags().StringP("priority", "p", "", "Filter by priority")
		cmd.Flags().StringP("energy", "e", "", "Filter by energy")
		cmd.Flags().String("tag", "", "Filter by tag")
		cmd.Flags().StringP("box", "b", "", "Filter by box")
		cmd.Flags().StringP("context", "c", "", "Filter by context")
		cmd.Flags().IntP("limit", "n", 0, "Limit number of results")
	}
}

// taskCmd is the parent command: `bt task <subcommand>`.
var taskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"t", "tasks"},
	Short:   "Manage tasks",
}

// --- bt task add / bt add ---

var taskAddCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Create a new task",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAdd,
}

var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Create a new task (shortcut for 'bt task add')",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

	title := strings.Join(args, " ")
	opts := app.TaskOptions{}

	// Read flags
	if v, _ := cmd.Flags().GetString("priority"); v != "" {
		opts.Priority = model.Priority(v)
	}
	if v, _ := cmd.Flags().GetString("energy"); v != "" {
		opts.Energy = model.Energy(v)
	}
	if v, _ := cmd.Flags().GetInt("duration"); v > 0 {
		opts.Duration = v
	}
	if v, _ := cmd.Flags().GetString("due"); v != "" {
		opts.DueDate = v
	}
	if v, _ := cmd.Flags().GetString("due-start"); v != "" {
		opts.DueStart = v
	}
	if v, _ := cmd.Flags().GetString("due-end"); v != "" {
		opts.DueEnd = v
	}
	if v, _ := cmd.Flags().GetStringSlice("tag"); len(v) > 0 {
		opts.Tags = v
	}
	if v, _ := cmd.Flags().GetString("context"); v != "" {
		opts.Context = []string{v}
	}
	if v, _ := cmd.Flags().GetString("box"); v != "" {
		opts.Box = v
	}

	task, err := a.AddTask(title, opts)
	if err != nil {
		return err
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	if quiet {
		cmd.Println(task.ID)
		return nil
	}

	cmd.Printf("✓ Created task %s\n  %s\n", task.ShortID(8), task.Title)
	return nil
}

// --- bt task list / bt list ---

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE:  runList,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks (shortcut for 'bt task list')",
	RunE:  runList,
}

func runList(cmd *cobra.Command, _ []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

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
	if v, _ := cmd.Flags().GetString("context"); v != "" {
		f.Context = v
	}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		f.Limit = v
	}

	tasks, err := a.ListTasks(f)
	if err != nil {
		return err
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	if quiet {
		for _, t := range tasks {
			cmd.Println(t.ID)
		}
		return nil
	}

	if len(tasks) == 0 {
		cmd.Println("No tasks found.")
		return nil
	}

	// Simple table output
	cmd.Printf("%-10s %-30s %-8s %-8s %s\n", "ID", "TITLE", "STATUS", "PRIORITY", "DUE")
	for _, t := range tasks {
		title := t.Title
		if len(title) > 28 {
			title = title[:27] + "…"
		}

		priority := "-"
		if t.Priority != nil {
			priority = *t.Priority
		}

		due := "-"
		if t.DueDate != nil {
			due = *t.DueDate
		} else if t.DueEnd != nil {
			due = "by " + *t.DueEnd
		}

		shortID := t.ID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		cmd.Printf("%-10s %-30s %-8s %-8s %s\n", shortID, title, t.Status, priority, due)
	}

	return nil
}

// --- bt task done / bt done ---

var taskDoneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a task as complete",
	Args:  cobra.ExactArgs(1),
	RunE:  runDone,
}

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a task as complete (shortcut for 'bt task done')",
	Args:  cobra.ExactArgs(1),
	RunE:  runDone,
}

func runDone(cmd *cobra.Command, args []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

	task, err := a.CompleteTask(args[0])
	if err != nil {
		return err
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	if quiet {
		cmd.Println(task.ID)
		return nil
	}

	cmd.Printf("✓ Completed: %s\n", task.Title)
	return nil
}

// --- bt task show ---

var taskShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show task details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		task, relPath, err := a.GetTask(args[0])
		if err != nil {
			return err
		}

		cmd.Printf("ID:       %s\n", task.ID)
		cmd.Printf("Title:    %s\n", task.Title)
		cmd.Printf("Type:     %s\n", task.Type)
		cmd.Printf("Status:   %s\n", task.Status)
		if task.Priority != "" {
			cmd.Printf("Priority: %s\n", task.Priority)
		}
		if task.Energy != "" {
			cmd.Printf("Energy:   %s\n", task.Energy)
		}
		if task.EstimatedDuration > 0 {
			cmd.Printf("Duration: ~%d min\n", task.EstimatedDuration)
		}
		if task.DueDate != "" {
			cmd.Printf("Due:      %s\n", task.DueDate)
		}
		if task.DueStart != "" || task.DueEnd != "" {
			cmd.Printf("Due:      %s – %s\n", task.DueStart, task.DueEnd)
		}
		if len(task.Tags) > 0 {
			cmd.Printf("Tags:     %s\n", strings.Join(task.Tags, ", "))
		}
		if len(task.Context) > 0 {
			cmd.Printf("Context:  %s\n", strings.Join(task.Context, ", "))
		}
		if task.Box != "" {
			cmd.Printf("Box:      %s\n", task.Box)
		}
		cmd.Printf("File:     %s\n", relPath)
		cmd.Printf("Created:  %s\n", task.Created.Format("2006-01-02 15:04"))
		cmd.Printf("Updated:  %s\n", task.Updated.Format("2006-01-02 15:04"))
		if task.CompletedAt != nil {
			cmd.Printf("Done:     %s\n", task.CompletedAt.Format("2006-01-02 15:04"))
		}

		if task.Body != "" {
			cmd.Printf("\n%s\n", task.Body)
		}

		return nil
	},
}

// --- bt task edit ---

var taskEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a task (via flags or $EDITOR)",
	Long: `Edit a task. If flags are provided, applies them directly.
If no flags are given, opens the task file in $EDITOR.

Examples:
  bt task edit 01JQX --title "New title"     Update title
  bt task edit 01JQX -p high -s active       Update priority and status
  bt task edit 01JQX --tag errands --tag home Replace tags
  bt task edit 01JQX                         Open in $EDITOR`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

	idOrPrefix := args[0]

	// Check if any edit flags were provided
	hasFlags := cmd.Flags().Changed("title") ||
		cmd.Flags().Changed("priority") ||
		cmd.Flags().Changed("energy") ||
		cmd.Flags().Changed("duration") ||
		cmd.Flags().Changed("due") ||
		cmd.Flags().Changed("due-start") ||
		cmd.Flags().Changed("due-end") ||
		cmd.Flags().Changed("tag") ||
		cmd.Flags().Changed("context") ||
		cmd.Flags().Changed("box") ||
		cmd.Flags().Changed("status")

	if hasFlags {
		return editWithFlags(cmd, a, idOrPrefix)
	}

	return editWithEditor(cmd, a, idOrPrefix)
}

// editWithFlags applies flag-based edits directly.
func editWithFlags(cmd *cobra.Command, a *app.App, idOrPrefix string) error {
	task, err := a.UpdateTask(idOrPrefix, func(t *model.Task) {
		if v, _ := cmd.Flags().GetString("title"); cmd.Flags().Changed("title") {
			t.Title = v
		}
		if v, _ := cmd.Flags().GetString("priority"); cmd.Flags().Changed("priority") {
			t.Priority = model.Priority(v)
		}
		if v, _ := cmd.Flags().GetString("energy"); cmd.Flags().Changed("energy") {
			t.Energy = model.Energy(v)
		}
		if v, _ := cmd.Flags().GetInt("duration"); cmd.Flags().Changed("duration") {
			t.EstimatedDuration = v
		}
		if v, _ := cmd.Flags().GetString("due"); cmd.Flags().Changed("due") {
			t.DueDate = v
		}
		if v, _ := cmd.Flags().GetString("due-start"); cmd.Flags().Changed("due-start") {
			t.DueStart = v
		}
		if v, _ := cmd.Flags().GetString("due-end"); cmd.Flags().Changed("due-end") {
			t.DueEnd = v
		}
		if v, _ := cmd.Flags().GetStringSlice("tag"); cmd.Flags().Changed("tag") {
			t.Tags = v
		}
		if v, _ := cmd.Flags().GetString("context"); cmd.Flags().Changed("context") {
			t.Context = []string{v}
		}
		if v, _ := cmd.Flags().GetString("box"); cmd.Flags().Changed("box") {
			t.Box = v
		}
		if v, _ := cmd.Flags().GetString("status"); cmd.Flags().Changed("status") {
			t.Status = model.Status(v)
		}
	})
	if err != nil {
		return err
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	if quiet {
		cmd.Println(task.ID)
		return nil
	}

	cmd.Printf("✓ Updated: %s\n", task.Title)
	return nil
}

// editWithEditor opens the task's .md file in $EDITOR.
func editWithEditor(cmd *cobra.Command, a *app.App, idOrPrefix string) error {
	filePath, err := a.EditTaskFile(idOrPrefix)
	if err != nil {
		return err
	}

	editor := firstNonEmpty(os.Getenv("EDITOR"), os.Getenv("VISUAL"), "vi")

	// Split editor command in case it has args (e.g., "code --wait")
	parts := strings.Fields(editor)
	editorCmd := exec.Command(parts[0], append(parts[1:], filePath)...)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Reload the task from disk (editor may have changed it)
	task, err := a.ReloadTask(idOrPrefix)
	if err != nil {
		return fmt.Errorf("reload after edit: %w", err)
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	if quiet {
		cmd.Println(task.ID)
		return nil
	}

	cmd.Printf("✓ Updated: %s\n", task.Title)
	return nil
}

// firstNonEmpty returns the first non-empty string, or the last argument as fallback.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return values[len(values)-1]
}

// --- bt task delete ---

var taskDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		task, err := a.DeleteTask(args[0])
		if err != nil {
			return err
		}

		quiet, _ := cmd.Flags().GetBool("quiet")
		if quiet {
			cmd.Println(task.ID)
			return nil
		}

		cmd.Printf("✓ Deleted: %s\n", task.Title)
		return nil
	},
}

// --- bt index rebuild ---

func init() {
	indexCmd.AddCommand(indexRebuildCmd)
	rootCmd.AddCommand(indexCmd)
}

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index management",
}

var indexRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild the SQLite index from markdown files",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		count, err := a.RebuildIndex()
		if err != nil {
			return err
		}

		cmd.Printf("✓ Rebuilt index: %d tasks indexed\n", count)
		return nil
	},
}

// --- Helpers ---

// openApp creates an App instance using the --data-dir flag or default.
func openApp(cmd *cobra.Command) (*app.App, error) {
	dataDir, _ := cmd.Flags().GetString("data-dir")

	// Resolve relative paths
	if dataDir != "" {
		abs, err := filepath.Abs(dataDir)
		if err != nil {
			return nil, fmt.Errorf("resolve data-dir: %w", err)
		}
		dataDir = abs
	}

	return app.Open(dataDir)
}
