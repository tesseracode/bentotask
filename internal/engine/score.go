// Package engine implements the BentoTask smart scheduling algorithm.
//
// The core feature is the "Bento Packing Algorithm" — a weighted knapsack
// variant that answers "What should I do now?" by scoring tasks on urgency,
// priority, energy match, streak risk, age, and dependency impact.
//
// See SPEC.md §6 for the full algorithm specification.
package engine

import (
	"math"
	"time"

	"github.com/tesserabox/bentotask/internal/model"
)

// DefaultWeights are the scoring weights from SPEC.md §6.2.
// user_preference (w7=0.05) is deferred — requires accept/skip history.
// The remaining 0.95 is distributed across the 6 active factors.
var DefaultWeights = Weights{
	Urgency:         0.25,
	Priority:        0.20,
	EnergyMatch:     0.15,
	StreakRisk:       0.15,
	AgeBoost:        0.10,
	DependencyUnlock: 0.10,
}

// Weights holds the scoring weights for the Bento Packing Algorithm.
// All values should be in [0, 1] and ideally sum to ~1.0.
type Weights struct {
	Urgency         float64
	Priority        float64
	EnergyMatch     float64
	StreakRisk       float64
	AgeBoost        float64
	DependencyUnlock float64
}

// ScoreBreakdown holds the individual factor scores and the final
// weighted score for a task. Useful for debugging and display.
type ScoreBreakdown struct {
	Urgency         float64 `json:"urgency"`
	Priority        float64 `json:"priority"`
	EnergyMatch     float64 `json:"energy_match"`
	StreakRisk       float64 `json:"streak_risk"`
	AgeBoost        float64 `json:"age_boost"`
	DependencyUnlock float64 `json:"dependency_unlock"`
	Total           float64 `json:"total"`
}

// HabitInfo carries habit-specific data needed for streak risk scoring.
// Populated from the habit package before calling ScoreTask.
type HabitInfo struct {
	FreqType          string // "daily" or "weekly"
	FreqTarget        int    // target completions per period
	CompletedToday    bool   // has the habit been completed today?
	CompletionsThisWeek int  // number of completions in the current ISO week
	CurrentStreak     int    // current streak length
}

// TaskContext provides all the context needed to score a single task.
type TaskContext struct {
	Task          *model.Task
	Now           time.Time
	UserEnergy    model.Energy        // user's current energy level
	HabitInfo     *HabitInfo          // nil for non-habit tasks
	BlockedCount  int                 // how many tasks are blocked by this one
	TotalTasks    int                 // total eligible tasks (for normalization)
}

// ScoreTask computes the full score breakdown for a task using the given weights.
func ScoreTask(ctx TaskContext, w Weights) ScoreBreakdown {
	bd := ScoreBreakdown{
		Urgency:         Urgency(ctx.Task, ctx.Now),
		Priority:        PriorityScore(ctx.Task.Priority),
		EnergyMatch:     EnergyMatch(ctx.Task.Energy, ctx.UserEnergy),
		StreakRisk:       StreakRisk(ctx.HabitInfo, ctx.Now),
		AgeBoost:        AgeBoost(ctx.Task.Created, ctx.Now),
		DependencyUnlock: DependencyUnlock(ctx.BlockedCount, ctx.TotalTasks),
	}

	bd.Total = w.Urgency*bd.Urgency +
		w.Priority*bd.Priority +
		w.EnergyMatch*bd.EnergyMatch +
		w.StreakRisk*bd.StreakRisk +
		w.AgeBoost*bd.AgeBoost +
		w.DependencyUnlock*bd.DependencyUnlock

	return bd
}

// --- Individual scoring functions ---

// Urgency scores a task based on how close its due date is.
// Returns a value in [0, 1] per SPEC.md §6.3:
//
//	due today       → 1.0
//	due tomorrow    → 0.8
//	due within 3d   → 0.6
//	due within 7d   → 0.4
//	due within 30d  → 0.2
//	floating task   → 0.1 + age_factor (capped at 0.5)
//	no due date     → 0.0
func Urgency(t *model.Task, now time.Time) float64 {
	dueStr := t.DueDate
	if dueStr == "" {
		// Ranged tasks: use end date for urgency
		dueStr = t.DueEnd
	}

	if dueStr == "" {
		// Floating tasks get a small urgency boost based on age
		if t.Type == model.TaskTypeFloating {
			ageFactor := ageFactorLinear(t.Created, now)
			score := 0.1 + ageFactor*0.4 // caps at 0.5
			if score > 0.5 {
				score = 0.5
			}
			return score
		}
		return 0.0
	}

	due, err := time.Parse("2006-01-02", dueStr)
	if err != nil {
		return 0.0
	}

	// Calculate days until due date (using calendar days, not 24h periods)
	todayDate := truncateToDay(now)
	dueDate := truncateToDay(due)
	daysUntil := int(dueDate.Sub(todayDate).Hours() / 24)

	// Overdue tasks are maximally urgent
	if daysUntil <= 0 {
		return 1.0
	}
	if daysUntil == 1 {
		return 0.8
	}
	if daysUntil <= 3 {
		return 0.6
	}
	if daysUntil <= 7 {
		return 0.4
	}
	if daysUntil <= 30 {
		return 0.2
	}
	return 0.0
}

// PriorityScore maps a Priority level to a score in [0, 1].
func PriorityScore(p model.Priority) float64 {
	switch p {
	case model.PriorityUrgent:
		return 1.0
	case model.PriorityHigh:
		return 0.75
	case model.PriorityMedium:
		return 0.5
	case model.PriorityLow:
		return 0.25
	default:
		return 0.0
	}
}

// EnergyMatch scores how well a task's energy requirement matches the
// user's current energy level. Returns a value in [0, 1]:
//
//	exact match        → 1.0
//	one level below    → 0.5
//	two levels below   → 0.2
//
// Tasks requiring MORE energy than the user has are still scored here
// (they'll be filtered out during packing, not scoring).
func EnergyMatch(taskEnergy, userEnergy model.Energy) float64 {
	taskLevel := energyLevel(taskEnergy)
	userLevel := energyLevel(userEnergy)

	diff := userLevel - taskLevel
	switch {
	case diff == 0:
		return 1.0
	case diff == 1:
		return 0.5
	case diff >= 2:
		return 0.2
	default:
		// Task requires more energy than user has (diff < 0)
		// Still give a small score — filtering happens at packing stage
		return 0.1
	}
}

// StreakRisk scores the risk of breaking a habit's streak.
// Returns a value in [0, 1]:
//
//	Daily habit not completed today             → 1.0
//	Daily habit completed today                  → 0.0
//	Weekly habit: (target - completions) / target, boosted near deadline
//	Non-habits                                   → 0.0
func StreakRisk(info *HabitInfo, now time.Time) float64 {
	if info == nil {
		return 0.0
	}

	switch info.FreqType {
	case "daily":
		if info.CompletedToday {
			return 0.0
		}
		// Not completed today — streak at risk
		// Boost if there's an active streak that would break
		if info.CurrentStreak > 0 {
			return 1.0
		}
		// No active streak, still somewhat urgent to build one
		return 0.7

	case "weekly":
		if info.FreqTarget <= 0 {
			return 0.0
		}
		remaining := info.FreqTarget - info.CompletionsThisWeek
		if remaining <= 0 {
			return 0.0 // target met for this week
		}

		// Base risk: proportion of remaining completions needed
		baseRisk := float64(remaining) / float64(info.FreqTarget)

		// Boost risk as the week progresses (more urgent near end of week)
		weekday := now.Weekday() // Sunday=0, Monday=1, ...
		daysLeft := 7 - int(weekday)
		if daysLeft == 0 {
			daysLeft = 1 // Sunday: treat as 1 day left
		}

		// If remaining completions >= days left, it's very urgent
		if remaining >= daysLeft {
			return math.Min(1.0, baseRisk*1.5)
		}

		return baseRisk

	default:
		return 0.0
	}
}

// AgeBoost gives older tasks a small score boost to prevent them from
// being permanently buried. Uses logarithmic growth, capped at 1.0.
//
// Formula: min(1.0, log2(1 + days_since_creation) / log2(91))
//
// This means a task reaches max boost after ~90 days.
func AgeBoost(created time.Time, now time.Time) float64 {
	days := now.Sub(created).Hours() / 24
	if days <= 0 {
		return 0.0
	}

	// Logarithmic growth: reaches 1.0 at ~90 days
	score := math.Log2(1+days) / math.Log2(91)
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// DependencyUnlock scores how many tasks are blocked by this task.
// Tasks that unblock more downstream work get higher scores.
//
// Formula: min(1.0, blocked_count / max(1, total_tasks * 0.1))
//
// If a task blocks 10% or more of all eligible tasks, it gets 1.0.
func DependencyUnlock(blockedCount, totalTasks int) float64 {
	if blockedCount == 0 {
		return 0.0
	}

	denominator := math.Max(1, float64(totalTasks)*0.1)
	score := float64(blockedCount) / denominator
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// --- Helpers ---

// energyLevel converts an Energy enum to a numeric level.
func energyLevel(e model.Energy) int {
	switch e {
	case model.EnergyHigh:
		return 3
	case model.EnergyMedium:
		return 2
	case model.EnergyLow:
		return 1
	default:
		return 2 // default to medium
	}
}

// ageFactorLinear returns a [0, 1] factor based on task age.
// Reaches 1.0 at 90 days. Used for floating task urgency.
func ageFactorLinear(created time.Time, now time.Time) float64 {
	days := now.Sub(created).Hours() / 24
	if days <= 0 {
		return 0.0
	}
	factor := days / 90.0
	if factor > 1.0 {
		factor = 1.0
	}
	return factor
}

// truncateToDay returns midnight of the given time's date.
func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
