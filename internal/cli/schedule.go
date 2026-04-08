package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/engine"
	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/style"
)

func init() {
	// bt now
	nowCmd.Flags().IntP("time", "t", 60, "Available time in minutes")
	nowCmd.Flags().StringP("energy", "e", "medium", "Current energy level: low, medium, high")
	nowCmd.Flags().StringP("context", "c", "", "Current context: home, office, errands")
	nowCmd.Flags().IntP("count", "n", 5, "Number of suggestions to show")
	rootCmd.AddCommand(nowCmd)

	// bt plan today
	planTodayCmd.Flags().IntP("time", "t", 480, "Total available time in minutes (default: 8 hours)")
	planTodayCmd.Flags().StringP("energy", "e", "medium", "Current energy level: low, medium, high")
	planTodayCmd.Flags().StringP("context", "c", "", "Current context: home, office, errands")
	planCmd.AddCommand(planTodayCmd)
	rootCmd.AddCommand(planCmd)

	// Dynamic completions for energy and context flags
	for _, cmd := range []*cobra.Command{nowCmd, planTodayCmd} {
		_ = cmd.RegisterFlagCompletionFunc("energy", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"low", "medium", "high"}, cobra.ShellCompDirectiveNoFileComp
		})
		_ = cmd.RegisterFlagCompletionFunc("context", func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			a, err := openApp(cmd)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			defer func() { _ = a.Close() }()
			ctxs, _ := a.CompleteContexts()
			return ctxs, cobra.ShellCompDirectiveNoFileComp
		})
	}
}

// --- bt now ---

var nowCmd = &cobra.Command{
	Use:   "now",
	Short: "What should I do now? Get smart task suggestions",
	Long: `Suggests what to work on based on urgency, priority, energy,
streak risk, and task dependencies.

Examples:
  bt now                              Default: 60 min, medium energy
  bt now --time 30 --energy low       Quick low-energy tasks
  bt now -t 120 -e high -c office     Deep work at the office
  bt now --count 3                    Show only top 3 suggestions
  bt now --json                       JSON output for scripting`,
	RunE: runNow,
}

func runNow(cmd *cobra.Command, _ []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

	availTime, _ := cmd.Flags().GetInt("time")
	energyStr, _ := cmd.Flags().GetString("energy")
	context, _ := cmd.Flags().GetString("context")
	count, _ := cmd.Flags().GetInt("count")

	opts := app.SuggestOptions{
		AvailableTime: availTime,
		Energy:        model.Energy(energyStr),
		Context:       context,
	}

	suggestions, err := a.Suggest(opts, count)
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return writeJSON(cmd.OutOrStdout(), suggestionsToJSON(suggestions))
	}

	if len(suggestions) == 0 {
		cmd.Println(style.Dim("No tasks match your current filters. Try relaxing energy or context."))
		return nil
	}

	// Header
	header := fmt.Sprintf("What to do now — %d min, %s energy",
		availTime, energyStr)
	if context != "" {
		header += ", " + context
	}
	cmd.Println(style.Header(header))
	cmd.Println()

	for i, s := range suggestions {
		renderSuggestion(cmd, i+1, s)
	}

	return nil
}

// --- bt plan today ---

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan your time",
}

var planTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Generate today's task schedule",
	Long: `Packs tasks into your available time using the Bento algorithm.
Shows a time-blocked plan ordered by score.

Examples:
  bt plan today                       Default: 8h, medium energy
  bt plan today --time 240            Plan for 4 hours
  bt plan today -e high -c office     High-energy office day
  bt plan today --json                JSON output for scripting`,
	RunE: runPlanToday,
}

func runPlanToday(cmd *cobra.Command, _ []string) error {
	a, err := openApp(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = a.Close() }()

	availTime, _ := cmd.Flags().GetInt("time")
	energyStr, _ := cmd.Flags().GetString("energy")
	context, _ := cmd.Flags().GetString("context")

	opts := app.SuggestOptions{
		AvailableTime: availTime,
		Energy:        model.Energy(energyStr),
		Context:       context,
	}

	result, err := a.PlanDay(opts)
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return writeJSON(cmd.OutOrStdout(), planToJSON(result, availTime))
	}

	if len(result.Suggestions) == 0 {
		cmd.Println(style.Dim("No tasks to plan. Try relaxing energy or context filters."))
		return nil
	}

	// Header
	header := fmt.Sprintf("Today's Plan — %s available", formatDuration(availTime))
	if energyStr != "" {
		header += ", " + energyStr + " energy"
	}
	if context != "" {
		header += ", " + context
	}
	cmd.Println(style.Header(header))
	cmd.Println()

	// Time-blocked view
	elapsed := 0
	for i, s := range result.Suggestions {
		startMin := elapsed
		endMin := elapsed + s.Duration
		timeSlot := fmt.Sprintf("%s – %s", formatClock(startMin), formatClock(endMin))

		shortID := s.Task.ID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		// Score bar
		scoreBar := renderScoreBar(s.Score.Total)

		cmd.Printf("  %s  %s  %s  %s  %s\n",
			style.Dim(fmt.Sprintf("%d.", i+1)),
			style.Bold(timeSlot),
			s.Task.Title,
			style.Dim(fmt.Sprintf("[%s]", shortID)),
			scoreBar,
		)

		elapsed = endMin
	}

	cmd.Println()
	cmd.Printf("  %s %s packed, %s free\n",
		style.Bold("Total:"),
		formatDuration(result.TotalDuration),
		formatDuration(result.TimeRemaining),
	)

	return nil
}

// --- Rendering helpers ---

func renderSuggestion(cmd *cobra.Command, rank int, s engine.Suggestion) {
	shortID := s.Task.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	// Title line
	cmd.Printf("  %s %s  %s  ~%d min\n",
		style.Bold(fmt.Sprintf("%d.", rank)),
		s.Task.Title,
		style.Dim(fmt.Sprintf("[%s]", shortID)),
		s.Duration,
	)

	// Score breakdown
	parts := []string{}
	if s.Score.Urgency > 0 {
		parts = append(parts, fmt.Sprintf("urgency=%.1f", s.Score.Urgency))
	}
	if s.Score.Priority > 0 {
		parts = append(parts, fmt.Sprintf("priority=%.1f", s.Score.Priority))
	}
	if s.Score.EnergyMatch > 0 {
		parts = append(parts, fmt.Sprintf("energy=%.1f", s.Score.EnergyMatch))
	}
	if s.Score.StreakRisk > 0 {
		parts = append(parts, fmt.Sprintf("streak=%.1f", s.Score.StreakRisk))
	}
	if s.Score.AgeBoost > 0.05 {
		parts = append(parts, fmt.Sprintf("age=%.2f", s.Score.AgeBoost))
	}
	if s.Score.DependencyUnlock > 0 {
		parts = append(parts, fmt.Sprintf("unlock=%.1f", s.Score.DependencyUnlock))
	}

	scoreBar := renderScoreBar(s.Score.Total)
	cmd.Printf("     %s %s  %s\n",
		scoreBar,
		style.Dim(fmt.Sprintf("%.2f", s.Score.Total)),
		style.Dim(strings.Join(parts, " ")),
	)

	// Extra info line
	var extras []string
	if s.Task.Priority != "" && s.Task.Priority != model.PriorityNone {
		extras = append(extras, style.Priority(string(s.Task.Priority)))
	}
	if s.Task.Energy != "" {
		extras = append(extras, style.Energy(string(s.Task.Energy)))
	}
	if s.Task.DueDate != "" {
		extras = append(extras, "due "+s.Task.DueDate)
	}
	if len(s.Task.Tags) > 0 {
		for _, tag := range s.Task.Tags {
			extras = append(extras, style.Tag(tag))
		}
	}
	if len(extras) > 0 {
		cmd.Printf("     %s\n", strings.Join(extras, "  "))
	}
	cmd.Println()
}

func renderScoreBar(score float64) string {
	// 10-char bar: ████████░░
	filled := min(int(score*10/0.95), 10) // normalize to max possible score (0.95)
	filled = max(filled, 0)
	return style.Bold(strings.Repeat("█", filled)) + style.Dim(strings.Repeat("░", 10-filled))
}

func formatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	h := minutes / 60
	m := minutes % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}

func formatClock(minutesFromStart int) string {
	h := minutesFromStart / 60
	m := minutesFromStart % 60
	return fmt.Sprintf("%d:%02d", h, m)
}
