package engine

import (
	"slices"
	"sort"
	"time"

	"github.com/tesserabox/bentotask/internal/model"
)

// Suggestion is a single task recommendation from the packing algorithm.
type Suggestion struct {
	Task     *model.Task    `json:"task"`
	Score    ScoreBreakdown `json:"score"`
	Duration int            `json:"duration"` // effective duration in minutes
}

// PackRequest holds all inputs for the Bento Packing Algorithm.
type PackRequest struct {
	// Available time in minutes.
	AvailableTime int

	// User's current energy level.
	UserEnergy model.Energy

	// Current context (e.g., "home", "office"). Empty means any context.
	Context string

	// Current time (for scoring). Use time.Now() in production.
	Now time.Time

	// Scoring weights.
	Weights Weights

	// All candidate tasks (pre-filtered to status=pending/active).
	Tasks []*model.Task

	// HabitInfoMap maps task ID → HabitInfo for habit tasks.
	// Non-habit tasks should not have entries.
	HabitInfoMap map[string]*HabitInfo

	// BlockedByMap maps task ID → count of tasks that are blocked by it.
	// Built from the dependency graph.
	BlockedByMap map[string]int

	// UnmetDependencies is the set of task IDs whose dependencies are NOT met.
	// Tasks in this set are excluded from packing.
	UnmetDependencies map[string]bool
}

// PackResult holds the output of the Bento Packing Algorithm.
type PackResult struct {
	Suggestions   []Suggestion `json:"suggestions"`
	TotalDuration int          `json:"total_duration"` // sum of packed task durations
	TimeRemaining int          `json:"time_remaining"` // unused time
}

// DefaultDuration is the fallback task duration (in minutes) when
// a task has no estimated_duration set.
const DefaultDuration = 15

// Pack runs the Bento Packing Algorithm per SPEC.md §6.1:
//
//  1. Filter: context match, energy <= E, duration <= T, dependencies met
//  2. Score each eligible task
//  3. Greedy knapsack: sort by score/duration ratio, pack until full
//  4. First Fit Decreasing for remaining gaps
//  5. Return ordered suggestion list
func Pack(req PackRequest) PackResult {
	// Step 1: Filter eligible tasks
	eligible := filterTasks(req)

	if len(eligible) == 0 {
		return PackResult{
			TimeRemaining: req.AvailableTime,
		}
	}

	totalTasks := len(eligible)

	// Step 2: Score each task
	type scored struct {
		task     *model.Task
		score    ScoreBreakdown
		duration int
		ratio    float64 // score / duration ratio for knapsack
	}

	var items []scored
	for _, task := range eligible {
		dur := effectiveDuration(task)

		ctx := TaskContext{
			Task:         task,
			Now:          req.Now,
			UserEnergy:   req.UserEnergy,
			HabitInfo:    req.HabitInfoMap[task.ID],
			BlockedCount: req.BlockedByMap[task.ID],
			TotalTasks:   totalTasks,
		}

		bd := ScoreTask(ctx, req.Weights)

		ratio := bd.Total
		if dur > 0 {
			ratio = bd.Total / float64(dur)
		}

		items = append(items, scored{
			task:     task,
			score:    bd,
			duration: dur,
			ratio:    ratio,
		})
	}

	// Step 3: Greedy knapsack — sort by score/duration ratio (descending)
	sort.Slice(items, func(i, j int) bool {
		if items[i].ratio != items[j].ratio {
			return items[i].ratio > items[j].ratio
		}
		// Tie-break: prefer higher total score
		return items[i].score.Total > items[j].score.Total
	})

	remaining := req.AvailableTime
	packed := make(map[int]bool) // index → packed
	var suggestions []Suggestion

	// First pass: greedy pack by ratio
	for i, item := range items {
		if item.duration <= remaining {
			suggestions = append(suggestions, Suggestion{
				Task:     item.task,
				Score:    item.score,
				Duration: item.duration,
			})
			remaining -= item.duration
			packed[i] = true
		}
	}

	// Step 4: First Fit Decreasing for remaining gaps
	// Sort unpacked items by duration descending, try to fit them
	if remaining > 0 {
		var unpacked []int
		for i := range items {
			if !packed[i] {
				unpacked = append(unpacked, i)
			}
		}

		// Sort by duration descending (FFD)
		sort.Slice(unpacked, func(a, b int) bool {
			return items[unpacked[a]].duration > items[unpacked[b]].duration
		})

		for _, idx := range unpacked {
			item := items[idx]
			if item.duration <= remaining {
				suggestions = append(suggestions, Suggestion{
					Task:     item.task,
					Score:    item.score,
					Duration: item.duration,
				})
				remaining -= item.duration
				packed[idx] = true
			}
		}
	}

	// Calculate total duration
	totalDuration := 0
	for _, s := range suggestions {
		totalDuration += s.Duration
	}

	return PackResult{
		Suggestions:   suggestions,
		TotalDuration: totalDuration,
		TimeRemaining: remaining,
	}
}

// TopN returns the top N suggestions without packing constraints
// (no time limit). Useful for "bt now" which wants ranked suggestions.
func TopN(req PackRequest, n int) []Suggestion {
	eligible := filterTasks(req)

	if len(eligible) == 0 {
		return nil
	}

	totalTasks := len(eligible)

	type scored struct {
		task  *model.Task
		score ScoreBreakdown
		dur   int
	}

	var items []scored
	for _, task := range eligible {
		dur := effectiveDuration(task)

		ctx := TaskContext{
			Task:         task,
			Now:          req.Now,
			UserEnergy:   req.UserEnergy,
			HabitInfo:    req.HabitInfoMap[task.ID],
			BlockedCount: req.BlockedByMap[task.ID],
			TotalTasks:   totalTasks,
		}

		bd := ScoreTask(ctx, req.Weights)
		items = append(items, scored{task: task, score: bd, dur: dur})
	}

	// Sort by total score descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].score.Total > items[j].score.Total
	})

	if n > len(items) {
		n = len(items)
	}

	suggestions := make([]Suggestion, n)
	for i := 0; i < n; i++ {
		suggestions[i] = Suggestion{
			Task:     items[i].task,
			Score:    items[i].score,
			Duration: items[i].dur,
		}
	}

	return suggestions
}

// filterTasks returns only tasks eligible for scheduling.
func filterTasks(req PackRequest) []*model.Task {
	var eligible []*model.Task

	for _, task := range req.Tasks {
		// Skip non-actionable statuses
		if task.Status != model.StatusPending && task.Status != model.StatusActive {
			continue
		}

		// Skip routines (they're containers, not actionable)
		if task.Type == model.TaskTypeRoutine {
			continue
		}

		// Skip tasks with unmet dependencies
		if req.UnmetDependencies != nil && req.UnmetDependencies[task.ID] {
			continue
		}

		// Filter by context: task must match the user's context (or have no context set)
		if req.Context != "" && !matchesContext(task, req.Context) {
			continue
		}

		// Filter by energy: task must not require more energy than user has
		if !energyFits(task.Energy, req.UserEnergy) {
			continue
		}

		// Filter by duration: task must fit within available time
		dur := effectiveDuration(task)
		if req.AvailableTime > 0 && dur > req.AvailableTime {
			continue
		}

		eligible = append(eligible, task)
	}

	return eligible
}

// matchesContext checks if a task matches the user's context.
// A task matches if:
//   - The task has no contexts set (available anywhere)
//   - The task's contexts include the user's context
func matchesContext(task *model.Task, userContext string) bool {
	if len(task.Context) == 0 {
		return true // no context restriction
	}
	return slices.Contains(task.Context, userContext)
}

// energyFits checks if a task's energy requirement fits the user's level.
// A task fits if its energy is <= the user's energy.
func energyFits(taskEnergy, userEnergy model.Energy) bool {
	return energyLevel(taskEnergy) <= energyLevel(userEnergy)
}

// effectiveDuration returns a task's duration, defaulting to DefaultDuration
// if the task has no estimated_duration.
func effectiveDuration(task *model.Task) int {
	if task.EstimatedDuration > 0 {
		return task.EstimatedDuration
	}
	return DefaultDuration
}
