package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/nlp"
	"github.com/tesserabox/bentotask/internal/store"
)

func (s *Server) registerTools() {
	s.registerTaskTools()
	s.registerHabitTools()
	s.registerRoutineTools()
	s.registerLinkTools()
	s.registerSchedulingTools()
	s.registerMetaTools()
	s.registerNLPTools()
}

func (s *Server) registerTaskTools() {
	s.register(Tool{
		Name:        "add_task",
		Description: "Create a new task in BentoTask",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title":    map[string]any{"type": "string", "description": "Task title (required)"},
				"priority": map[string]any{"type": "string", "enum": []string{"urgent", "high", "medium", "low"}, "description": "Priority level"},
				"energy":   map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}, "description": "Energy required"},
				"duration": map[string]any{"type": "integer", "description": "Estimated duration in minutes"},
				"due_date": map[string]any{"type": "string", "description": "Due date (YYYY-MM-DD)"},
				"tags":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Tags"},
				"box":      map[string]any{"type": "string", "description": "Box/folder for organization"},
				"body":     map[string]any{"type": "string", "description": "Markdown body content"},
			},
			"required": []string{"title"},
		},
		handler: func(params map[string]any) (string, error) {
			title := getString(params, "title")
			if title == "" {
				return "", fmt.Errorf("title is required")
			}
			opts := app.TaskOptions{
				Priority: model.Priority(getString(params, "priority")),
				Energy:   model.Energy(getString(params, "energy")),
				Duration: getInt(params, "duration"),
				DueDate:  getString(params, "due_date"),
				Tags:     getStringSlice(params, "tags"),
				Box:      getString(params, "box"),
				Body:     getString(params, "body"),
			}
			task, err := s.app.AddTask(title, opts)
			if err != nil {
				return "", fmt.Errorf("add task: %w", err)
			}
			result := fmt.Sprintf("Created task '%s' (ID: %s", task.Title, task.ShortID(8))
			if task.Priority != "" {
				result += fmt.Sprintf(", priority: %s", task.Priority)
			}
			if task.DueDate != "" {
				result += fmt.Sprintf(", due: %s", task.DueDate)
			}
			result += ")"
			return result, nil
		},
	})

	s.register(Tool{
		Name:        "list_tasks",
		Description: "List tasks with optional filters",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status":   map[string]any{"type": "string", "enum": []string{"pending", "active", "done", "cancelled", "paused", "waiting"}},
				"priority": map[string]any{"type": "string", "enum": []string{"urgent", "high", "medium", "low"}},
				"energy":   map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}},
				"tag":      map[string]any{"type": "string", "description": "Filter by tag"},
				"box":      map[string]any{"type": "string", "description": "Filter by box"},
				"limit":    map[string]any{"type": "integer", "description": "Max results"},
			},
		},
		handler: func(params map[string]any) (string, error) {
			f := &store.TaskFilter{
				Status:   model.Status(getString(params, "status")),
				Priority: model.Priority(getString(params, "priority")),
				Energy:   model.Energy(getString(params, "energy")),
				Tag:      getString(params, "tag"),
				Box:      getString(params, "box"),
				Limit:    getInt(params, "limit"),
			}
			tasks, err := s.app.ListTasks(f)
			if err != nil {
				return "", fmt.Errorf("list tasks: %w", err)
			}
			if len(tasks) == 0 {
				return "No tasks found matching the filters.", nil
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "Found %d tasks:\n", len(tasks))
			for i, t := range tasks {
				shortID := t.ID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}
				line := fmt.Sprintf("%d. [%s] %s (%s", i+1, shortID, t.Title, t.Status)
				if t.Priority != nil {
					line += fmt.Sprintf(", %s priority", *t.Priority)
				}
				if t.DueDate != nil {
					line += fmt.Sprintf(", due %s", *t.DueDate)
				}
				line += ")"
				fmt.Fprintln(&sb, line)
			}
			return sb.String(), nil
		},
	})

	s.register(Tool{
		Name:        "get_task",
		Description: "Get full details of a task by ID or prefix",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"id": map[string]any{"type": "string", "description": "Task ID or prefix"}},
			"required":   []string{"id"},
		},
		handler: func(params map[string]any) (string, error) {
			id := getString(params, "id")
			task, _, err := s.app.GetTask(id)
			if err != nil {
				return "", fmt.Errorf("get task: %w", err)
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "Task: %s\n", task.Title)
			fmt.Fprintf(&sb, "ID: %s\n", task.ID)
			fmt.Fprintf(&sb, "Type: %s | Status: %s\n", task.Type, task.Status)
			if task.Priority != "" {
				fmt.Fprintf(&sb, "Priority: %s\n", task.Priority)
			}
			if task.Energy != "" {
				fmt.Fprintf(&sb, "Energy: %s\n", task.Energy)
			}
			if task.DueDate != "" {
				fmt.Fprintf(&sb, "Due: %s\n", task.DueDate)
			}
			if len(task.Tags) > 0 {
				fmt.Fprintf(&sb, "Tags: %s\n", strings.Join(task.Tags, ", "))
			}
			if task.Body != "" {
				fmt.Fprintf(&sb, "Body:\n%s\n", task.Body)
			}
			return sb.String(), nil
		},
	})

	s.register(Tool{
		Name:        "update_task",
		Description: "Update task fields (title, priority, energy, due_date, tags, status)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":       map[string]any{"type": "string", "description": "Task ID or prefix"},
				"title":    map[string]any{"type": "string"},
				"priority": map[string]any{"type": "string", "enum": []string{"urgent", "high", "medium", "low", "none"}},
				"energy":   map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}},
				"due_date": map[string]any{"type": "string"},
				"tags":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"status":   map[string]any{"type": "string", "enum": []string{"pending", "active", "done", "cancelled", "paused", "waiting"}},
			},
			"required": []string{"id"},
		},
		handler: func(params map[string]any) (string, error) {
			id := getString(params, "id")
			task, err := s.app.UpdateTask(id, func(t *model.Task) {
				if v := getString(params, "title"); v != "" {
					t.Title = v
				}
				if v := getString(params, "priority"); v != "" {
					t.Priority = model.Priority(v)
				}
				if v := getString(params, "energy"); v != "" {
					t.Energy = model.Energy(v)
				}
				if v := getString(params, "due_date"); v != "" {
					t.DueDate = v
				}
				if v := getStringSlice(params, "tags"); v != nil {
					t.Tags = v
				}
				if v := getString(params, "status"); v != "" {
					t.Status = model.Status(v)
				}
			})
			if err != nil {
				return "", fmt.Errorf("update task: %w", err)
			}
			return fmt.Sprintf("Updated task '%s' (%s)", task.Title, task.ShortID(8)), nil
		},
	})

	s.register(Tool{
		Name:        "complete_task",
		Description: "Mark a task as done",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"id": map[string]any{"type": "string", "description": "Task ID or prefix"}},
			"required":   []string{"id"},
		},
		handler: func(params map[string]any) (string, error) {
			task, err := s.app.CompleteTask(getString(params, "id"))
			if err != nil {
				return "", fmt.Errorf("complete task: %w", err)
			}
			return fmt.Sprintf("Completed: %s (%s)", task.Title, task.ShortID(8)), nil
		},
	})

	s.register(Tool{
		Name:        "delete_task",
		Description: "Delete a task permanently",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"id": map[string]any{"type": "string", "description": "Task ID or prefix"}},
			"required":   []string{"id"},
		},
		handler: func(params map[string]any) (string, error) {
			task, err := s.app.DeleteTask(getString(params, "id"))
			if err != nil {
				return "", fmt.Errorf("delete task: %w", err)
			}
			return fmt.Sprintf("Deleted: %s", task.Title), nil
		},
	})

	s.register(Tool{
		Name:        "search_tasks",
		Description: "Full-text search across task titles and body content",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"query": map[string]any{"type": "string", "description": "Search query"}},
			"required":   []string{"query"},
		},
		handler: func(params map[string]any) (string, error) {
			query := getString(params, "query")
			tasks, err := s.app.SearchTasks(query)
			if err != nil {
				return "", fmt.Errorf("search: %w", err)
			}
			if len(tasks) == 0 {
				return fmt.Sprintf("No results for %q", query), nil
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "Found %d results for %q:\n", len(tasks), query)
			for i, t := range tasks {
				shortID := t.ID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}
				fmt.Fprintf(&sb, "%d. [%s] %s (%s)\n", i+1, shortID, t.Title, t.Status)
			}
			return sb.String(), nil
		},
	})
}

func (s *Server) registerHabitTools() {
	s.register(Tool{
		Name:        "add_habit",
		Description: "Create a new habit to track",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title":          map[string]any{"type": "string", "description": "Habit name"},
				"freq_type":      map[string]any{"type": "string", "enum": []string{"daily", "weekly"}, "description": "Frequency type"},
				"freq_target":    map[string]any{"type": "integer", "description": "Target completions per period"},
				"max_per_period": map[string]any{"type": "integer", "description": "Max completions per period (0 = unlimited)"},
				"priority":       map[string]any{"type": "string", "enum": []string{"urgent", "high", "medium", "low"}},
				"energy":         map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}},
			},
			"required": []string{"title"},
		},
		handler: func(params map[string]any) (string, error) {
			title := getString(params, "title")
			freqType := getString(params, "freq_type")
			if freqType == "" {
				freqType = "daily"
			}
			freqTarget := getInt(params, "freq_target")
			if freqTarget == 0 {
				freqTarget = 1
			}
			recurrence := "FREQ=DAILY"
			if freqType == "weekly" {
				recurrence = "FREQ=WEEKLY"
			}
			opts := app.HabitOptions{
				FreqType:     freqType,
				FreqTarget:   freqTarget,
				MaxPerPeriod: getInt(params, "max_per_period"),
				Recurrence:   recurrence,
				Priority:     model.Priority(getString(params, "priority")),
				Energy:       model.Energy(getString(params, "energy")),
			}
			task, err := s.app.AddHabit(title, opts)
			if err != nil {
				return "", fmt.Errorf("add habit: %w", err)
			}
			return fmt.Sprintf("Created habit '%s' (%s, %s, target: %d)", task.Title, task.ShortID(8), freqType, freqTarget), nil
		},
	})

	s.register(Tool{
		Name:        "list_habits",
		Description: "List all tracked habits",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		handler: func(_ map[string]any) (string, error) {
			habits, err := s.app.ListHabits()
			if err != nil {
				return "", fmt.Errorf("list habits: %w", err)
			}
			if len(habits) == 0 {
				return "No habits tracked yet.", nil
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "%d habits:\n", len(habits))
			for i, h := range habits {
				shortID := h.ID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}
				fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, shortID, h.Title)
			}
			return sb.String(), nil
		},
	})

	s.register(Tool{
		Name:        "log_habit",
		Description: "Log a habit completion for today",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":       map[string]any{"type": "string", "description": "Habit ID or prefix"},
				"duration": map[string]any{"type": "integer", "description": "Duration in minutes"},
				"note":     map[string]any{"type": "string", "description": "Completion note"},
			},
			"required": []string{"id"},
		},
		handler: func(params map[string]any) (string, error) {
			task, err := s.app.LogHabit(getString(params, "id"), getInt(params, "duration"), getString(params, "note"))
			if err != nil {
				return "", fmt.Errorf("log habit: %w", err)
			}
			return fmt.Sprintf("Logged completion for '%s'", task.Title), nil
		},
	})

	s.register(Tool{
		Name:        "habit_stats",
		Description: "Get streak and completion statistics for a habit",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"id": map[string]any{"type": "string", "description": "Habit ID or prefix"}},
			"required":   []string{"id"},
		},
		handler: func(params map[string]any) (string, error) {
			task, stats, err := s.app.HabitStats(getString(params, "id"))
			if err != nil {
				return "", fmt.Errorf("habit stats: %w", err)
			}
			return fmt.Sprintf("%s — 🔥 %d day streak (longest: %d), %d total completions, %d%% rate (%dd)",
				task.Title, stats.CurrentStreak, stats.LongestStreak,
				stats.TotalCompletions, int(stats.CompletionRate*100), stats.RatePeriodDays), nil
		},
	})
}

func (s *Server) registerRoutineTools() {
	s.register(Tool{
		Name:        "list_routines",
		Description: "List all routines",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		handler: func(_ map[string]any) (string, error) {
			routines, err := s.app.ListRoutines()
			if err != nil {
				return "", fmt.Errorf("list routines: %w", err)
			}
			if len(routines) == 0 {
				return "No routines defined.", nil
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "%d routines:\n", len(routines))
			for i, r := range routines {
				shortID := r.ID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}
				fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, shortID, r.Title)
			}
			return sb.String(), nil
		},
	})

	s.register(Tool{
		Name:        "get_routine",
		Description: "Get routine details including steps",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"id": map[string]any{"type": "string", "description": "Routine ID or prefix"}},
			"required":   []string{"id"},
		},
		handler: func(params map[string]any) (string, error) {
			task, _, err := s.app.GetTask(getString(params, "id"))
			if err != nil {
				return "", fmt.Errorf("get routine: %w", err)
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "Routine: %s (%s)\n", task.Title, task.ShortID(8))
			if len(task.Steps) == 0 {
				sb.WriteString("No steps defined.\n")
			} else {
				fmt.Fprintf(&sb, "%d steps:\n", len(task.Steps))
				for i, step := range task.Steps {
					line := fmt.Sprintf("  %d. %s", i+1, step.Title)
					if step.Duration > 0 {
						line += fmt.Sprintf(" (~%dm)", step.Duration)
					}
					if step.Optional {
						line += " [optional]"
					}
					fmt.Fprintln(&sb, line)
				}
			}
			return sb.String(), nil
		},
	})
}

func (s *Server) registerLinkTools() {
	s.register(Tool{
		Name:        "link_tasks",
		Description: "Create a relationship between two tasks (depends-on, blocks, related-to)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"source_id": map[string]any{"type": "string", "description": "Source task ID"},
				"target_id": map[string]any{"type": "string", "description": "Target task ID"},
				"type":      map[string]any{"type": "string", "enum": []string{"depends-on", "blocks", "related-to"}, "description": "Link type"},
			},
			"required": []string{"source_id", "target_id"},
		},
		handler: func(params map[string]any) (string, error) {
			lt := model.LinkType(getString(params, "type"))
			if lt == "" {
				lt = model.LinkRelatedTo
			}
			source, target, err := s.app.LinkTasks(getString(params, "source_id"), getString(params, "target_id"), lt)
			if err != nil {
				return "", fmt.Errorf("link tasks: %w", err)
			}
			return fmt.Sprintf("Linked: %s —[%s]→ %s", source.Title, lt, target.Title), nil
		},
	})

	s.register(Tool{
		Name:        "unlink_tasks",
		Description: "Remove a link between two tasks",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"source_id": map[string]any{"type": "string", "description": "Source task ID"},
				"target_id": map[string]any{"type": "string", "description": "Target task ID"},
				"type":      map[string]any{"type": "string", "enum": []string{"depends-on", "blocks", "related-to"}},
			},
			"required": []string{"source_id", "target_id"},
		},
		handler: func(params map[string]any) (string, error) {
			lt := model.LinkType(getString(params, "type"))
			if lt == "" {
				lt = model.LinkRelatedTo
			}
			source, target, err := s.app.UnlinkTasks(getString(params, "source_id"), getString(params, "target_id"), lt)
			if err != nil {
				return "", fmt.Errorf("unlink tasks: %w", err)
			}
			return fmt.Sprintf("Unlinked: %s —[%s]→ %s", source.Title, lt, target.Title), nil
		},
	})
}

func (s *Server) registerSchedulingTools() {
	s.register(Tool{
		Name:        "suggest",
		Description: "Get smart task suggestions based on available time, energy, and context",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"time":    map[string]any{"type": "integer", "description": "Available time in minutes (default: 60)"},
				"energy":  map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}, "description": "Current energy level"},
				"context": map[string]any{"type": "string", "description": "Current context (home, office, etc.)"},
				"count":   map[string]any{"type": "integer", "description": "Number of suggestions (default: 5)"},
			},
		},
		handler: func(params map[string]any) (string, error) {
			availTime := getInt(params, "time")
			if availTime == 0 {
				availTime = 60
			}
			energy := model.Energy(getString(params, "energy"))
			if energy == "" {
				energy = model.EnergyMedium
			}
			count := getInt(params, "count")
			if count == 0 {
				count = 5
			}
			opts := app.SuggestOptions{
				AvailableTime: availTime,
				Energy:        energy,
				Context:       getString(params, "context"),
			}
			suggestions, err := s.app.Suggest(opts, count)
			if err != nil {
				return "", fmt.Errorf("suggest: %w", err)
			}
			if len(suggestions) == 0 {
				return "No suggestions available. Try adjusting filters or adding more tasks.", nil
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "Top %d suggestions for %dmin/%s energy:\n", len(suggestions), availTime, energy)
			for i, s := range suggestions {
				fmt.Fprintf(&sb, "%d. %s (score: %.2f, %dm)", i+1, s.Task.Title, s.Score.Total, s.Duration)
				parts := []string{}
				if s.Score.Urgency > 0 {
					parts = append(parts, fmt.Sprintf("urgency: %.1f", s.Score.Urgency))
				}
				if s.Score.Priority > 0 {
					parts = append(parts, fmt.Sprintf("priority: %.1f", s.Score.Priority))
				}
				if len(parts) > 0 {
					fmt.Fprintf(&sb, " — %s", strings.Join(parts, ", "))
				}
				fmt.Fprintln(&sb)
			}
			return sb.String(), nil
		},
	})

	s.register(Tool{
		Name:        "plan_today",
		Description: "Generate a packed daily plan optimized for available time and energy",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"time":    map[string]any{"type": "integer", "description": "Total available time in minutes (default: 480)"},
				"energy":  map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}},
				"context": map[string]any{"type": "string", "description": "Current context"},
			},
		},
		handler: func(params map[string]any) (string, error) {
			availTime := getInt(params, "time")
			if availTime == 0 {
				availTime = 480
			}
			energy := model.Energy(getString(params, "energy"))
			if energy == "" {
				energy = model.EnergyMedium
			}
			opts := app.SuggestOptions{
				AvailableTime: availTime,
				Energy:        energy,
				Context:       getString(params, "context"),
			}
			result, err := s.app.PlanDay(opts)
			if err != nil {
				return "", fmt.Errorf("plan today: %w", err)
			}
			if len(result.Suggestions) == 0 {
				return "Nothing to plan. Add some tasks first!", nil
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "Day plan (%dm available, %dm packed, %dm free):\n",
				availTime, result.TotalDuration, result.TimeRemaining)
			elapsed := 0
			for i, s := range result.Suggestions {
				start := elapsed
				end := elapsed + s.Duration
				fmt.Fprintf(&sb, "%d. %d:%02d–%d:%02d  %s (%dm)\n",
					i+1, start/60, start%60, end/60, end%60, s.Task.Title, s.Duration)
				elapsed = end
			}
			return sb.String(), nil
		},
	})
}

func (s *Server) registerMetaTools() {
	s.register(Tool{
		Name:        "list_tags",
		Description: "List all unique tags used across tasks",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		handler: func(_ map[string]any) (string, error) {
			tags, err := s.app.CompleteTags()
			if err != nil {
				return "", fmt.Errorf("list tags: %w", err)
			}
			if len(tags) == 0 {
				return "No tags in use.", nil
			}
			return fmt.Sprintf("Tags: %s", strings.Join(tags, ", ")), nil
		},
	})
}

func (s *Server) registerNLPTools() {
	s.register(Tool{
		Name:        "parse_natural",
		Description: "Parse natural language into a structured task. Returns extracted fields without creating the task — use add_task to create it.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text": map[string]any{"type": "string", "description": "Natural language task description"},
			},
			"required": []string{"text"},
		},
		handler: func(params map[string]any) (string, error) {
			text := getString(params, "text")
			if text == "" {
				return "", fmt.Errorf("text is required")
			}
			parsed := nlp.Parse(text, time.Now())
			var sb strings.Builder
			fmt.Fprintf(&sb, "Parsed task:\n")
			fmt.Fprintf(&sb, "  Title: %s\n", parsed.Title)
			if parsed.Priority != "" {
				fmt.Fprintf(&sb, "  Priority: %s\n", parsed.Priority)
			}
			if parsed.Energy != "" {
				fmt.Fprintf(&sb, "  Energy: %s\n", parsed.Energy)
			}
			if parsed.DueDate != "" {
				fmt.Fprintf(&sb, "  Due: %s\n", parsed.DueDate)
			}
			if parsed.Duration > 0 {
				fmt.Fprintf(&sb, "  Duration: %dm\n", parsed.Duration)
			}
			if len(parsed.Tags) > 0 {
				fmt.Fprintf(&sb, "  Tags: %s\n", strings.Join(parsed.Tags, ", "))
			}
			if parsed.Context != "" {
				fmt.Fprintf(&sb, "  Context: %s\n", parsed.Context)
			}
			return sb.String(), nil
		},
	})

	s.register(Tool{
		Name:        "quick_add",
		Description: "Parse natural language and create a task in one step. Extracts dates, priority, tags, duration from the text.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text": map[string]any{"type": "string", "description": "Natural language task description (e.g., 'buy groceries tomorrow #errands')"},
			},
			"required": []string{"text"},
		},
		handler: func(params map[string]any) (string, error) {
			text := getString(params, "text")
			if text == "" {
				return "", fmt.Errorf("text is required")
			}
			parsed := nlp.Parse(text, time.Now())
			if parsed.Title == "" {
				return "", fmt.Errorf("could not extract a task title from: %q", text)
			}
			opts := app.TaskOptions{
				Priority: model.Priority(parsed.Priority),
				Energy:   model.Energy(parsed.Energy),
				DueDate:  parsed.DueDate,
				Duration: parsed.Duration,
				Tags:     parsed.Tags,
			}
			if parsed.Context != "" {
				opts.Context = []string{parsed.Context}
			}
			task, err := s.app.AddTask(parsed.Title, opts)
			if err != nil {
				return "", fmt.Errorf("create task: %w", err)
			}
			result := fmt.Sprintf("Created: '%s' (ID: %s", task.Title, task.ShortID(8))
			if parsed.Priority != "" {
				result += ", " + parsed.Priority
			}
			if parsed.DueDate != "" {
				result += ", due " + parsed.DueDate
			}
			if parsed.Duration > 0 {
				result += fmt.Sprintf(", %dm", parsed.Duration)
			}
			if len(parsed.Tags) > 0 {
				result += ", #" + strings.Join(parsed.Tags, " #")
			}
			result += ")"
			return result, nil
		},
	})
}
