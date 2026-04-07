package cli

import (
	"bufio"
	"fmt"
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
	routineCmd.AddCommand(routineCreateCmd)
	routineCmd.AddCommand(routineListCmd)
	routineCmd.AddCommand(routineShowCmd)
	routineCmd.AddCommand(routinePlayCmd)
	rootCmd.AddCommand(routineCmd)

	// bt routine create flags
	routineCreateCmd.Flags().StringSlice("step", nil, `Steps as "title:duration" (repeatable, e.g., --step "Shower:5" --step "Breakfast:15")`)
	routineCreateCmd.Flags().String("schedule-time", "", "Schedule time (HH:MM)")
	routineCreateCmd.Flags().StringSlice("schedule-days", nil, "Schedule days (e.g., --schedule-days mon,wed,fri)")
	routineCreateCmd.Flags().StringP("priority", "p", "", "Priority: none, low, medium, high, urgent")
	routineCreateCmd.Flags().StringP("energy", "e", "", "Energy: low, medium, high")
	routineCreateCmd.Flags().StringSlice("tag", nil, "Tags (repeatable)")

	// Register completions
	routineCreateCmd.RegisterFlagCompletionFunc("priority", completePriority) //nolint:errcheck
	routineCreateCmd.RegisterFlagCompletionFunc("energy", completeEnergy)     //nolint:errcheck

	// Task ID completion for show and play
	routineShowCmd.ValidArgsFunction = completeRoutineIDs
	routinePlayCmd.ValidArgsFunction = completeRoutineIDs
}

// completeRoutineIDs provides completion for routine-type tasks only.
func completeRoutineIDs(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	a, err := openApp(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer func() { _ = a.Close() }()

	routines, err := a.ListRoutines()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var comps []string
	for _, r := range routines {
		comps = append(comps, r.ID+"\t"+r.Title)
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

var routineCmd = &cobra.Command{
	Use:     "routine",
	Aliases: []string{"r", "routines"},
	Short:   "Manage routines",
}

// --- bt routine create ---

var routineCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new routine",
	Long: `Create a routine with an ordered sequence of steps.

Steps are specified as "title:duration_in_minutes". Duration is optional.

Examples:
  bt routine create "Morning Routine" --step "Shower:5" --step "Breakfast:15" --step "Review inbox:10"
  bt routine create "Evening Wind-down" --step "Journal:10" --step "Read:30"
  bt routine create "Weekly Review" --step "Review goals" --step "Plan week:30" --schedule-time 09:00 --schedule-days mon`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		title := strings.Join(args, " ")

		// Parse steps from --step flags
		stepStrs, _ := cmd.Flags().GetStringSlice("step")
		if len(stepStrs) == 0 {
			return fmt.Errorf("at least one --step is required (e.g., --step \"Shower:5\")")
		}

		steps, err := parseStepFlags(stepStrs)
		if err != nil {
			return err
		}

		opts := app.RoutineOptions{
			Steps: steps,
		}

		// Parse schedule
		schedTime, _ := cmd.Flags().GetString("schedule-time")
		schedDays, _ := cmd.Flags().GetStringSlice("schedule-days")
		if schedTime != "" || len(schedDays) > 0 {
			opts.Schedule = &model.RoutineSchedule{
				Time: schedTime,
				Days: schedDays,
			}
		}

		if v, _ := cmd.Flags().GetString("priority"); v != "" {
			opts.Priority = model.Priority(v)
		}
		if v, _ := cmd.Flags().GetString("energy"); v != "" {
			opts.Energy = model.Energy(v)
		}
		if v, _ := cmd.Flags().GetStringSlice("tag"); len(v) > 0 {
			opts.Tags = v
		}

		task, err := a.AddRoutine(title, opts)
		if err != nil {
			return err
		}

		if isJSON(cmd) {
			_, relPath, _ := a.GetTask(task.ID)
			return writeJSON(cmd.OutOrStdout(), taskToJSON(task, relPath))
		}

		quiet, _ := cmd.Flags().GetBool("quiet")
		if quiet {
			cmd.Println(task.ID)
			return nil
		}

		totalDur := 0
		for _, s := range task.Steps {
			totalDur += s.Duration
		}
		durStr := ""
		if totalDur > 0 {
			durStr = fmt.Sprintf(" (~%d min)", totalDur)
		}
		cmd.Printf("%s %s\n  %s — %d steps%s\n",
			style.Success("Created routine"), style.Bold(task.ShortID(8)),
			task.Title, len(task.Steps), durStr)
		return nil
	},
}

// --- bt routine list ---

var routineListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all routines",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		routines, err := a.ListRoutines()
		if err != nil {
			return err
		}

		if isJSON(cmd) {
			items := make([]TaskJSON, len(routines))
			for i, r := range routines {
				items[i] = indexedToJSON(r)
			}
			return writeJSON(cmd.OutOrStdout(), items)
		}

		quiet, _ := cmd.Flags().GetBool("quiet")
		if quiet {
			for _, r := range routines {
				cmd.Println(r.ID)
			}
			return nil
		}

		if len(routines) == 0 {
			cmd.Println(style.Dim("No routines found. Create one with: bt routine create \"My routine\" --step \"Step 1:5\""))
			return nil
		}

		cmd.Printf("%-10s %-30s %-10s %s\n",
			style.Bold("ID"), style.Bold("ROUTINE"), style.Bold("STATUS"), style.Bold("TAGS"))
		for _, r := range routines {
			shortID := r.ID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}

			title := r.Title
			if len(title) > 28 {
				title = title[:27] + "…"
			}

			tagStr := ""
			if len(r.Tags) > 0 {
				styledTags := make([]string, len(r.Tags))
				for i, tag := range r.Tags {
					styledTags[i] = style.Tag(tag)
				}
				tagStr = strings.Join(styledTags, " ")
			}

			cmd.Printf("%-10s %-30s %-10s %s\n",
				style.Dim(shortID), title, style.Status(r.Status), tagStr)
		}

		return nil
	},
}

// --- bt routine show ---

var routineShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show routine details",
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
		if task.Type != model.TaskTypeRoutine {
			return fmt.Errorf("task %q is not a routine (type: %s)", task.Title, task.Type)
		}

		if isJSON(cmd) {
			return writeJSON(cmd.OutOrStdout(), taskToJSON(task, relPath))
		}

		cmd.Printf("ID:       %s\n", style.Dim(task.ID))
		cmd.Printf("Title:    %s\n", style.Bold(task.Title))
		cmd.Printf("Type:     routine\n")
		cmd.Printf("Status:   %s\n", style.Status(string(task.Status)))

		if task.Priority != "" {
			cmd.Printf("Priority: %s\n", style.Priority(string(task.Priority)))
		}
		if task.Energy != "" {
			cmd.Printf("Energy:   %s\n", style.Energy(string(task.Energy)))
		}
		if task.EstimatedDuration > 0 {
			cmd.Printf("Duration: ~%d min (total)\n", task.EstimatedDuration)
		}

		// Steps
		cmd.Printf("\n%s\n", style.Bold("Steps:"))
		for i, step := range task.Steps {
			opt := ""
			if step.Optional {
				opt = style.Dim(" (optional)")
			}
			dur := ""
			if step.Duration > 0 {
				dur = fmt.Sprintf(" %s", style.Dim(fmt.Sprintf("~%dmin", step.Duration)))
			}
			cmd.Printf("  %d. %s%s%s\n", i+1, step.Title, dur, opt)
		}

		// Schedule
		if task.Schedule != nil {
			cmd.Printf("\n%s\n", style.Bold("Schedule:"))
			if task.Schedule.Time != "" {
				cmd.Printf("  Time: %s\n", task.Schedule.Time)
			}
			if len(task.Schedule.Days) > 0 {
				cmd.Printf("  Days: %s\n", strings.Join(task.Schedule.Days, ", "))
			}
		}

		if len(task.Tags) > 0 {
			styledTags := make([]string, len(task.Tags))
			for i, tag := range task.Tags {
				styledTags[i] = style.Tag(tag)
			}
			cmd.Printf("\nTags:     %s\n", strings.Join(styledTags, " "))
		}

		cmd.Printf("\nFile:     %s\n", style.Dim(relPath))
		cmd.Printf("Created:  %s\n", style.Dim(task.Created.Format("2006-01-02 15:04")))
		cmd.Printf("Updated:  %s\n", style.Dim(task.Updated.Format("2006-01-02 15:04")))

		return nil
	},
}

// --- bt routine play ---

var routinePlayCmd = &cobra.Command{
	Use:   "play <id>",
	Short: "Play a routine step by step",
	Long: `Enter play mode for a routine — step through each item one at a time.

Press Enter to complete a step and advance, or 's' to skip optional steps.
A timer shows elapsed time for each step.

Examples:
  bt routine play 01JQX
  bt routine play morning`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		task, _, err := a.GetTask(args[0])
		if err != nil {
			return err
		}
		if task.Type != model.TaskTypeRoutine {
			return fmt.Errorf("task %q is not a routine (type: %s)", task.Title, task.Type)
		}

		if isJSON(cmd) {
			return runRoutinePlayJSON(cmd, task)
		}

		return runRoutinePlayInteractive(cmd, task)
	},
}

// runRoutinePlayJSON outputs each step result as a JSON array.
func runRoutinePlayJSON(cmd *cobra.Command, task *model.Task) error {
	type stepResult struct {
		Step     int    `json:"step"`
		Title    string `json:"title"`
		Status   string `json:"status"`
		Optional bool   `json:"optional,omitempty"`
		Duration int    `json:"estimated_duration,omitempty"`
	}
	results := make([]stepResult, len(task.Steps))
	for i, step := range task.Steps {
		results[i] = stepResult{
			Step:     i + 1,
			Title:    step.Title,
			Status:   "pending",
			Optional: step.Optional,
			Duration: step.Duration,
		}
	}
	return writeJSON(cmd.OutOrStdout(), map[string]any{
		"id":    task.ID,
		"title": task.Title,
		"steps": results,
	})
}

// runRoutinePlayInteractive runs the routine in interactive terminal mode.
func runRoutinePlayInteractive(cmd *cobra.Command, task *model.Task) error {
	totalSteps := len(task.Steps)
	completed := 0
	skipped := 0
	routineStart := time.Now()
	reader := bufio.NewReader(os.Stdin)

	cmd.Printf("\n%s  %s\n", style.Bold("▶ "+task.Title), style.Dim(fmt.Sprintf("(%d steps)", totalSteps)))
	cmd.Println(style.Dim(strings.Repeat("─", 40)))

	for i, step := range task.Steps {
		dur := ""
		if step.Duration > 0 {
			dur = fmt.Sprintf(" ~%dmin", step.Duration)
		}
		opt := ""
		if step.Optional {
			opt = " " + style.Dim("(optional)")
		}

		cmd.Printf("\n  %s  %s%s%s\n",
			style.Bold(fmt.Sprintf("[%d/%d]", i+1, totalSteps)),
			step.Title, style.Dim(dur), opt)

		stepStart := time.Now()

		// Prompt
		prompt := style.Dim("  Press Enter to complete")
		if step.Optional {
			prompt = style.Dim("  Enter=done, s=skip")
		}
		cmd.Printf("%s ", prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			// EOF or error — treat remaining steps as skipped
			break
		}
		input = strings.TrimSpace(strings.ToLower(input))

		elapsed := time.Since(stepStart).Round(time.Second)

		if input == "s" && step.Optional {
			skipped++
			cmd.Printf("  %s  %s\n", style.Dim("⊘ Skipped"), step.Title)
		} else {
			completed++
			cmd.Printf("  %s  %s (%s)\n", style.Success("✓"), step.Title, elapsed)
		}
	}

	routineElapsed := time.Since(routineStart).Round(time.Second)
	cmd.Println(style.Dim(strings.Repeat("─", 40)))
	cmd.Printf("\n%s %s completed in %s\n", style.Success("■"), task.Title, routineElapsed)
	cmd.Printf("  Steps: %d completed", completed)
	if skipped > 0 {
		cmd.Printf(", %d skipped", skipped)
	}
	cmd.Println()

	return nil
}

// --- Helpers ---

// parseStepFlags parses step flag strings like "Title:Duration" into RoutineStep slices.
// Duration is optional: "Shower:5" → 5 min, "Meditate" → 0 min (untimed).
func parseStepFlags(stepStrs []string) ([]model.RoutineStep, error) {
	var steps []model.RoutineStep
	for i, s := range stepStrs {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil, fmt.Errorf("step %d is empty", i+1)
		}

		step := model.RoutineStep{}

		// Check for optional suffix
		if strings.HasSuffix(s, "?") {
			step.Optional = true
			s = strings.TrimSuffix(s, "?")
		}

		// Split on last colon to separate title from duration
		if idx := strings.LastIndex(s, ":"); idx > 0 {
			maybeDur := strings.TrimSpace(s[idx+1:])
			if dur, err := strconv.Atoi(maybeDur); err == nil && dur > 0 {
				step.Title = strings.TrimSpace(s[:idx])
				step.Duration = dur
			} else {
				// Colon is part of the title (no valid duration after it)
				step.Title = s
			}
		} else {
			step.Title = s
		}

		if step.Title == "" {
			return nil, fmt.Errorf("step %d has empty title", i+1)
		}

		steps = append(steps, step)
	}
	return steps, nil
}
