package mcp

import (
	"fmt"
	"strings"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/store"
)

func (s *Server) registerResources() {
	s.registerResource(Resource{
		URI:         "bentotask://tasks/pending",
		Name:        "Pending Tasks",
		Description: "All pending tasks with priority and due dates",
		MimeType:    "text/plain",
		handler: func() (string, error) {
			tasks, err := s.app.ListTasks(&store.TaskFilter{Status: model.StatusPending})
			if err != nil {
				return "", fmt.Errorf("list pending tasks: %w", err)
			}
			if len(tasks) == 0 {
				return "No pending tasks.", nil
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "%d pending tasks:\n", len(tasks))
			for i, t := range tasks {
				shortID := t.ID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}
				line := fmt.Sprintf("%d. [%s] %s", i+1, shortID, t.Title)
				if t.Priority != nil {
					line += fmt.Sprintf(" (%s", *t.Priority)
					if t.DueDate != nil {
						line += fmt.Sprintf(", due %s", *t.DueDate)
					}
					line += ")"
				} else if t.DueDate != nil {
					line += fmt.Sprintf(" (due %s)", *t.DueDate)
				}
				if len(t.Tags) > 0 {
					line += " #" + strings.Join(t.Tags, " #")
				}
				fmt.Fprintln(&sb, line)
			}
			return sb.String(), nil
		},
	})

	s.registerResource(Resource{
		URI:         "bentotask://tasks/today",
		Name:        "Today's Suggestions",
		Description: "Top task suggestions ranked by scheduling score",
		MimeType:    "text/plain",
		handler: func() (string, error) {
			opts := app.SuggestOptions{AvailableTime: 480, Energy: model.EnergyMedium}
			suggestions, err := s.app.Suggest(opts, 10)
			if err != nil {
				return "", fmt.Errorf("suggest: %w", err)
			}
			if len(suggestions) == 0 {
				return "No suggestions available.", nil
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "Top %d suggestions:\n", len(suggestions))
			for i, sg := range suggestions {
				fmt.Fprintf(&sb, "%d. %s (score: %.2f, %dm)\n", i+1, sg.Task.Title, sg.Score.Total, sg.Duration)
			}
			return sb.String(), nil
		},
	})

	s.registerResource(Resource{
		URI:         "bentotask://habits/status",
		Name:        "Habit Streaks",
		Description: "Current streak and completion status for all habits",
		MimeType:    "text/plain",
		handler: func() (string, error) {
			habits, err := s.app.ListHabits()
			if err != nil {
				return "", fmt.Errorf("list habits: %w", err)
			}
			if len(habits) == 0 {
				return "No habits tracked.", nil
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "%d habits:\n", len(habits))
			for _, h := range habits {
				_, stats, err := s.app.HabitStats(h.ID)
				if err != nil {
					fmt.Fprintf(&sb, "- %s (stats unavailable)\n", h.Title)
					continue
				}
				done := "❌"
				if stats.CompletedToday {
					done = "✅"
				}
				fmt.Fprintf(&sb, "- %s: 🔥 %d streak, %d%% rate, today: %s\n",
					h.Title, stats.CurrentStreak, int(stats.CompletionRate*100), done)
			}
			return sb.String(), nil
		},
	})

	s.registerResource(Resource{
		URI:         "bentotask://plan/today",
		Name:        "Today's Plan",
		Description: "Time-blocked plan for the day",
		MimeType:    "text/plain",
		handler: func() (string, error) {
			opts := app.SuggestOptions{AvailableTime: 480, Energy: model.EnergyMedium}
			result, err := s.app.PlanDay(opts)
			if err != nil {
				return "", fmt.Errorf("plan today: %w", err)
			}
			if len(result.Suggestions) == 0 {
				return "Nothing to plan.", nil
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "Day plan (%dm packed, %dm free):\n", result.TotalDuration, result.TimeRemaining)
			elapsed := 0
			for i, sg := range result.Suggestions {
				start := elapsed
				end := elapsed + sg.Duration
				fmt.Fprintf(&sb, "%d. %d:%02d–%d:%02d  %s (%dm)\n",
					i+1, start/60, start%60, end/60, end%60, sg.Task.Title, sg.Duration)
				elapsed = end
			}
			return sb.String(), nil
		},
	})

	s.registerResource(Resource{
		URI:         "bentotask://meta/summary",
		Name:        "Task Summary",
		Description: "Overview of all tasks by status, plus habit and routine counts",
		MimeType:    "text/plain",
		handler: func() (string, error) {
			all, err := s.app.ListTasks(nil)
			if err != nil {
				return "", fmt.Errorf("list tasks: %w", err)
			}
			counts := map[string]int{}
			for _, t := range all {
				counts[t.Status]++
				counts["total"]++
			}
			habits, _ := s.app.ListHabits()
			routines, _ := s.app.ListRoutines()

			var sb strings.Builder
			fmt.Fprintf(&sb, "BentoTask Summary:\n")
			fmt.Fprintf(&sb, "  Total tasks: %d\n", counts["total"])
			for _, status := range []string{"pending", "active", "paused", "waiting", "done", "cancelled"} {
				if c := counts[status]; c > 0 {
					fmt.Fprintf(&sb, "  %s: %d\n", status, c)
				}
			}
			fmt.Fprintf(&sb, "  Habits: %d\n", len(habits))
			fmt.Fprintf(&sb, "  Routines: %d\n", len(routines))
			return sb.String(), nil
		},
	})
}
