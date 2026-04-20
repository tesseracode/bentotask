package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/style"
)

func init() {
	habitCmd.AddCommand(habitAddCmd)
	habitCmd.AddCommand(habitLogCmd)
	habitCmd.AddCommand(habitStatsCmd)
	habitCmd.AddCommand(habitListCmd)
	rootCmd.AddCommand(habitCmd)

	// bt habit add flags
	habitAddCmd.Flags().String("freq", "daily", "Frequency type: daily, weekly")
	habitAddCmd.Flags().Int("target", 1, "Target completions per period")
	habitAddCmd.Flags().Int("max", 0, "Maximum completions per period (0 = unlimited)")
	habitAddCmd.Flags().String("rrule", "", "RRULE string (auto-generated from --freq if omitted)")
	habitAddCmd.Flags().StringP("priority", "p", "", "Priority: none, low, medium, high, urgent")
	habitAddCmd.Flags().StringP("energy", "e", "", "Energy: low, medium, high")
	habitAddCmd.Flags().StringSlice("tag", nil, "Tags (repeatable)")
	habitAddCmd.Flags().StringP("context", "c", "", "Context: home, office, errands, anywhere")

	// bt habit log flags
	habitLogCmd.Flags().Int("duration", 0, "Duration in minutes")
	habitLogCmd.Flags().StringP("note", "n", "", "Completion note")

	// Register completions
	habitAddCmd.RegisterFlagCompletionFunc("freq", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) { //nolint:errcheck
		return []string{"daily\tEvery day", "weekly\tEvery week"}, cobra.ShellCompDirectiveNoFileComp
	})
	habitAddCmd.RegisterFlagCompletionFunc("priority", completePriority) //nolint:errcheck
	habitAddCmd.RegisterFlagCompletionFunc("energy", completeEnergy)     //nolint:errcheck

	// Task ID completion for log and stats
	habitLogCmd.ValidArgsFunction = completeHabitIDs
	habitStatsCmd.ValidArgsFunction = completeHabitIDs
}

// completeHabitIDs provides completion for habit-type tasks only.
func completeHabitIDs(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	a, err := openApp(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer func() { _ = a.Close() }()

	habits, err := a.ListHabits()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var comps []string
	for _, h := range habits {
		comps = append(comps, h.ID+"\t"+h.Title)
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

var habitCmd = &cobra.Command{
	Use:     "habit",
	Aliases: []string{"h", "habits"},
	Short:   "Manage habits",
}

// --- bt habit add ---

var habitAddCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Create a new habit",
	Long: `Create a habit with a frequency target and recurrence rule.

Examples:
  bt habit add "Read 30 minutes" --freq daily
  bt habit add "Exercise" --freq weekly --target 3
  bt habit add "Meditate" --rrule "FREQ=DAILY" -p high`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		title := strings.Join(args, " ")
		opts := app.HabitOptions{}

		opts.FreqType, _ = cmd.Flags().GetString("freq")
		opts.FreqTarget, _ = cmd.Flags().GetInt("target")
		opts.MaxPerPeriod, _ = cmd.Flags().GetInt("max")

		// Auto-generate RRULE from frequency if not explicitly provided
		opts.Recurrence, _ = cmd.Flags().GetString("rrule")
		if opts.Recurrence == "" {
			switch opts.FreqType {
			case "daily":
				opts.Recurrence = "FREQ=DAILY"
			case "weekly":
				opts.Recurrence = "FREQ=WEEKLY"
			default:
				return fmt.Errorf("unknown frequency type %q (use daily or weekly)", opts.FreqType)
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
		if v, _ := cmd.Flags().GetString("context"); v != "" {
			opts.Context = []string{v}
		}

		task, err := a.AddHabit(title, opts)
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

		cmd.Printf("%s %s\n  %s (%s, %dx/%s)\n",
			style.Success("Created habit"), style.Bold(task.ShortID(8)),
			task.Title, task.Frequency.Type, task.Frequency.Target, task.Frequency.Type)
		return nil
	},
}

// --- bt habit log ---

var habitLogCmd = &cobra.Command{
	Use:   "log <id>",
	Short: "Log a habit completion",
	Long: `Record that you completed a habit today.

Examples:
  bt habit log 01JQX                          Log completion
  bt habit log 01JQX --duration 35            With duration
  bt habit log 01JQX -n "Read DDIA ch.7"      With note
  bt habit log 01JQX --duration 30 -n "Great session"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		duration, _ := cmd.Flags().GetInt("duration")
		note, _ := cmd.Flags().GetString("note")

		task, err := a.LogHabit(args[0], duration, note)
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

		cmd.Printf("%s %s\n", style.Success("Logged:"), task.Title)
		if task.StreakCurrent > 0 {
			cmd.Printf("  Streak: %s %d days\n", style.Bold("🔥"), task.StreakCurrent)
		}
		return nil
	},
}

// --- bt habit stats ---

var habitStatsCmd = &cobra.Command{
	Use:   "stats <id>",
	Short: "Show habit statistics",
	Long: `Display streak information and completion rates for a habit.

Examples:
  bt habit stats 01JQX
  bt habit stats reading`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		task, stats, err := a.HabitStats(args[0])
		if err != nil {
			return err
		}

		if isJSON(cmd) {
			return writeJSON(cmd.OutOrStdout(), map[string]any{
				"id":                task.ID,
				"title":             task.Title,
				"current_streak":    stats.CurrentStreak,
				"longest_streak":    stats.LongestStreak,
				"total_completions": stats.TotalCompletions,
				"completion_rate":   stats.CompletionRate,
				"rate_period_days":  stats.RatePeriodDays,
			})
		}

		cmd.Printf("%s  %s\n\n", style.Bold(task.Title), style.Dim(task.ShortID(8)))

		// Streaks
		streakIcon := "🔥"
		if stats.CurrentStreak == 0 {
			streakIcon = "💤"
		}
		cmd.Printf("  Current streak:  %s %d\n", streakIcon, stats.CurrentStreak)
		cmd.Printf("  Longest streak:  🏆 %d\n", stats.LongestStreak)
		cmd.Printf("  Total completions: %d\n", stats.TotalCompletions)
		cmd.Printf("  Completion rate:   %.0f%% (last %d days)\n",
			stats.CompletionRate*100, stats.RatePeriodDays)

		// Frequency info
		if task.Frequency != nil {
			cmd.Printf("\n  Frequency: %dx/%s\n", task.Frequency.Target, task.Frequency.Type)
		}

		return nil
	},
}

// --- bt habit list ---

var habitListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all habits",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a, err := openApp(cmd)
		if err != nil {
			return err
		}
		defer func() { _ = a.Close() }()

		habits, err := a.ListHabits()
		if err != nil {
			return err
		}

		if isJSON(cmd) {
			items := make([]TaskJSON, len(habits))
			for i, h := range habits {
				items[i] = indexedToJSON(h)
			}
			return writeJSON(cmd.OutOrStdout(), items)
		}

		quiet, _ := cmd.Flags().GetBool("quiet")
		if quiet {
			for _, h := range habits {
				cmd.Println(h.ID)
			}
			return nil
		}

		if len(habits) == 0 {
			cmd.Println(style.Dim("No habits found. Create one with: bt habit add \"My habit\""))
			return nil
		}

		cmd.Printf("%-10s %-30s %-10s %s\n",
			style.Bold("ID"), style.Bold("HABIT"), style.Bold("STATUS"), style.Bold("TAGS"))
		for _, h := range habits {
			shortID := h.ID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}

			title := h.Title
			if len(title) > 28 {
				title = title[:27] + "…"
			}

			tagStr := ""
			if len(h.Tags) > 0 {
				styledTags := make([]string, len(h.Tags))
				for i, tag := range h.Tags {
					styledTags[i] = style.Tag(tag)
				}
				tagStr = strings.Join(styledTags, " ")
			}

			cmd.Printf("%-10s %-30s %-10s %s\n",
				style.Dim(shortID), title, style.Status(h.Status), tagStr)
		}

		return nil
	},
}
