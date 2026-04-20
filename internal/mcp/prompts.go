package mcp

import (
	"fmt"
	"strings"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/store"
)

func (s *Server) registerPrompts() {
	s.registerPrompt(Prompt{
		Name:        "daily-review",
		Description: "Review today's tasks, suggestions, and habit status with AI guidance",
		handler: func(_ map[string]string) ([]PromptMessage, error) {
			var sb strings.Builder
			sb.WriteString("Here's my current BentoTask status:\n\n")

			// Pending tasks
			tasks, _ := s.app.ListTasks(&store.TaskFilter{Status: model.StatusPending})
			sb.WriteString("PENDING TASKS:\n")
			if len(tasks) == 0 {
				sb.WriteString("  (none)\n")
			}
			for _, t := range tasks {
				line := fmt.Sprintf("  - %s", t.Title)
				if t.Priority != nil {
					line += fmt.Sprintf(" [%s]", *t.Priority)
				}
				if t.DueDate != nil {
					line += fmt.Sprintf(" (due %s)", *t.DueDate)
				}
				fmt.Fprintln(&sb, line)
			}

			// Suggestions
			opts := app.SuggestOptions{AvailableTime: 480, Energy: model.EnergyMedium}
			suggestions, _ := s.app.Suggest(opts, 5)
			sb.WriteString("\nTODAY'S SUGGESTIONS:\n")
			if len(suggestions) == 0 {
				sb.WriteString("  (none)\n")
			}
			for i, sg := range suggestions {
				fmt.Fprintf(&sb, "  %d. %s (score: %.2f, %dm)\n", i+1, sg.Task.Title, sg.Score.Total, sg.Duration)
			}

			// Habits
			habits, _ := s.app.ListHabits()
			sb.WriteString("\nHABIT STATUS:\n")
			if len(habits) == 0 {
				sb.WriteString("  (none)\n")
			}
			for _, h := range habits {
				_, stats, err := s.app.HabitStats(h.ID)
				if err != nil {
					fmt.Fprintf(&sb, "  - %s (stats unavailable)\n", h.Title)
					continue
				}
				done := "not done"
				if stats.CompletedToday {
					done = "done today"
				}
				fmt.Fprintf(&sb, "  - %s: 🔥 %d streak, %s\n", h.Title, stats.CurrentStreak, done)
			}

			sb.WriteString("\nPlease review this and help me:\n")
			sb.WriteString("1. Are there any overdue or at-risk items I should address first?\n")
			sb.WriteString("2. Does the suggested plan make sense for my priorities?\n")
			sb.WriteString("3. Which habits are at risk of breaking their streak?\n")
			sb.WriteString("4. Any tasks I should consider deferring or delegating?\n")

			return []PromptMessage{{
				Role:    "user",
				Content: PromptContent{Type: "text", Text: sb.String()},
			}}, nil
		},
	})

	s.registerPrompt(Prompt{
		Name:        "inbox-triage",
		Description: "Categorize uncategorized inbox tasks with AI help",
		handler: func(_ map[string]string) ([]PromptMessage, error) {
			tasks, _ := s.app.ListTasks(&store.TaskFilter{Status: model.StatusPending})

			// Filter to tasks without priority
			var uncategorized []*store.IndexedTask
			for _, t := range tasks {
				if t.Priority == nil || *t.Priority == "" || *t.Priority == "none" {
					uncategorized = append(uncategorized, t)
				}
			}

			var sb strings.Builder
			sb.WriteString("Here are my uncategorized inbox tasks:\n\n")
			if len(uncategorized) == 0 {
				sb.WriteString("  (all tasks are already categorized!)\n")
			}
			for _, t := range uncategorized {
				shortID := t.ID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}
				fmt.Fprintf(&sb, "  - [%s] %s\n", shortID, t.Title)
			}

			sb.WriteString("\nFor each task, please suggest:\n")
			sb.WriteString("- Priority level (urgent/high/medium/low)\n")
			sb.WriteString("- Energy required (low/medium/high)\n")
			sb.WriteString("- Appropriate tags\n")
			sb.WriteString("- Whether it needs a due date\n")
			sb.WriteString("\nThen I'll update them using the update_task tool.\n")

			return []PromptMessage{{
				Role:    "user",
				Content: PromptContent{Type: "text", Text: sb.String()},
			}}, nil
		},
	})

	s.registerPrompt(Prompt{
		Name:        "weekly-plan",
		Description: "Plan the week ahead with AI assistance",
		Arguments: []PromptArgument{
			{Name: "available_hours", Description: "Available hours this week (default: 40)", Required: false},
		},
		handler: func(args map[string]string) ([]PromptMessage, error) {
			hours := args["available_hours"]
			if hours == "" {
				hours = "40"
			}

			tasks, _ := s.app.ListTasks(&store.TaskFilter{Status: model.StatusPending})

			groups := map[string][]*store.IndexedTask{
				"urgent": {},
				"high":   {},
				"medium": {},
				"low":    {},
				"none":   {},
			}
			for _, t := range tasks {
				p := "none"
				if t.Priority != nil && *t.Priority != "" {
					p = *t.Priority
				}
				if _, ok := groups[p]; ok {
					groups[p] = append(groups[p], t)
				} else {
					groups["none"] = append(groups["none"], t)
				}
			}

			var sb strings.Builder
			sb.WriteString("Here are all my pending tasks for weekly planning:\n\n")
			for _, level := range []string{"urgent", "high", "medium", "low", "none"} {
				if len(groups[level]) == 0 {
					continue
				}
				fmt.Fprintf(&sb, "%s:\n", strings.ToUpper(level))
				for _, t := range groups[level] {
					line := fmt.Sprintf("  - %s", t.Title)
					if t.DueDate != nil {
						line += fmt.Sprintf(" (due %s)", *t.DueDate)
					}
					fmt.Fprintln(&sb, line)
				}
				sb.WriteString("\n")
			}

			fmt.Fprintf(&sb, "Available hours this week: %s\n\n", hours)
			sb.WriteString("Please help me:\n")
			sb.WriteString("1. Which tasks should I commit to this week?\n")
			sb.WriteString("2. Create a rough day-by-day breakdown\n")
			sb.WriteString("3. Flag any tasks that have been pending too long\n")
			sb.WriteString("4. Suggest which tasks to defer to next week\n")

			return []PromptMessage{{
				Role:    "user",
				Content: PromptContent{Type: "text", Text: sb.String()},
			}}, nil
		},
	})

	s.registerPrompt(Prompt{
		Name:        "habit-check",
		Description: "Check in on habit streaks and progress with AI coaching",
		handler: func(_ map[string]string) ([]PromptMessage, error) {
			habits, _ := s.app.ListHabits()

			var sb strings.Builder
			sb.WriteString("Here's my habit tracking status:\n\n")
			if len(habits) == 0 {
				sb.WriteString("  (no habits tracked yet)\n")
			}
			for _, h := range habits {
				_, stats, err := s.app.HabitStats(h.ID)
				if err != nil {
					fmt.Fprintf(&sb, "- %s (stats unavailable)\n", h.Title)
					continue
				}
				done := "❌ not done"
				if stats.CompletedToday {
					done = "✅ done"
				}
				fmt.Fprintf(&sb, "- %s: 🔥 %d streak (best: %d), %d total, %d%% rate, today: %s\n",
					h.Title, stats.CurrentStreak, stats.LongestStreak,
					stats.TotalCompletions, int(stats.CompletionRate*100), done)
			}

			sb.WriteString("\nPlease help me:\n")
			sb.WriteString("1. Which habits are at risk of breaking their streak?\n")
			sb.WriteString("2. Am I on track for my weekly targets?\n")
			sb.WriteString("3. Any habits I should consider adjusting the target for?\n")
			sb.WriteString("4. Celebrate any milestones or improvements!\n")

			return []PromptMessage{{
				Role:    "user",
				Content: PromptContent{Type: "text", Text: sb.String()},
			}}, nil
		},
	})
}
