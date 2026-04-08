package engine

import (
	"math"
	"testing"
	"time"

	"github.com/tesserabox/bentotask/internal/model"
)

// refTime is a fixed reference time for deterministic tests.
// Wednesday, April 8, 2026 at 10:00 UTC.
var refTime = time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC)

func assertFloat(t *testing.T, expected, actual float64, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

func assertInDelta(t *testing.T, expected, actual, delta float64, msg string) {
	t.Helper()
	if math.Abs(expected-actual) > delta {
		t.Errorf("%s: expected %v ± %v, got %v", msg, expected, delta, actual)
	}
}

func assertGreater(t *testing.T, a, b float64, msg string) {
	t.Helper()
	if a <= b {
		t.Errorf("%s: expected %v > %v", msg, a, b)
	}
}

func assertLessOrEqual(t *testing.T, a, b float64, msg string) {
	t.Helper()
	if a > b {
		t.Errorf("%s: expected %v <= %v", msg, a, b)
	}
}

func assertGreaterOrEqual(t *testing.T, a, b float64, msg string) {
	t.Helper()
	if a < b {
		t.Errorf("%s: expected %v >= %v", msg, a, b)
	}
}

// --- Urgency Tests ---

func TestUrgencyDueToday(t *testing.T) {
	task := &model.Task{DueDate: "2026-04-08"}
	assertFloat(t, 1.0, Urgency(task, refTime), "due today")
}

func TestUrgencyOverdue(t *testing.T) {
	task := &model.Task{DueDate: "2026-04-07"}
	assertFloat(t, 1.0, Urgency(task, refTime), "overdue tasks are maximally urgent")
}

func TestUrgencyDueTomorrow(t *testing.T) {
	task := &model.Task{DueDate: "2026-04-09"}
	assertFloat(t, 0.8, Urgency(task, refTime), "due tomorrow")
}

func TestUrgencyDueWithin3Days(t *testing.T) {
	// 2 days away
	task := &model.Task{DueDate: "2026-04-10"}
	assertFloat(t, 0.6, Urgency(task, refTime), "due in 2 days")

	// 3 days away
	task = &model.Task{DueDate: "2026-04-11"}
	assertFloat(t, 0.6, Urgency(task, refTime), "due in 3 days")
}

func TestUrgencyDueWithin7Days(t *testing.T) {
	// 5 days away
	task := &model.Task{DueDate: "2026-04-13"}
	assertFloat(t, 0.4, Urgency(task, refTime), "due in 5 days")

	// 7 days away
	task = &model.Task{DueDate: "2026-04-15"}
	assertFloat(t, 0.4, Urgency(task, refTime), "due in 7 days")
}

func TestUrgencyDueWithin30Days(t *testing.T) {
	task := &model.Task{DueDate: "2026-05-01"}
	assertFloat(t, 0.2, Urgency(task, refTime), "due in ~23 days")
}

func TestUrgencyDueFarFuture(t *testing.T) {
	task := &model.Task{DueDate: "2026-12-31"}
	assertFloat(t, 0.0, Urgency(task, refTime), "due in far future")
}

func TestUrgencyNoDueDate(t *testing.T) {
	task := &model.Task{Type: model.TaskTypeOneShot}
	assertFloat(t, 0.0, Urgency(task, refTime), "no due date, one-shot")
}

func TestUrgencyFloatingTask(t *testing.T) {
	task := &model.Task{
		Type:    model.TaskTypeFloating,
		Created: refTime.AddDate(0, 0, -30), // 30 days old
	}
	score := Urgency(task, refTime)
	assertGreater(t, score, 0.1, "floating tasks get more than base 0.1")
	assertLessOrEqual(t, score, 0.5, "floating tasks are capped at 0.5")
}

func TestUrgencyFloatingTaskNewlyCreated(t *testing.T) {
	task := &model.Task{
		Type:    model.TaskTypeFloating,
		Created: refTime, // just created
	}
	score := Urgency(task, refTime)
	assertInDelta(t, 0.1, score, 0.01, "new floating task gets near-base urgency")
}

func TestUrgencyRangedTaskUsesEndDate(t *testing.T) {
	task := &model.Task{
		Type:     model.TaskTypeRanged,
		DueStart: "2026-04-01",
		DueEnd:   "2026-04-09", // end is tomorrow
	}
	assertFloat(t, 0.8, Urgency(task, refTime), "ranged task end date is tomorrow")
}

func TestUrgencyInvalidDateFormat(t *testing.T) {
	task := &model.Task{DueDate: "not-a-date"}
	assertFloat(t, 0.0, Urgency(task, refTime), "invalid date format")
}

// --- Priority Tests ---

func TestPriorityScoreAll(t *testing.T) {
	tests := []struct {
		priority model.Priority
		expected float64
	}{
		{model.PriorityUrgent, 1.0},
		{model.PriorityHigh, 0.75},
		{model.PriorityMedium, 0.5},
		{model.PriorityLow, 0.25},
		{model.PriorityNone, 0.0},
		{"", 0.0},
	}
	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			assertFloat(t, tt.expected, PriorityScore(tt.priority), string(tt.priority))
		})
	}
}

// --- Energy Match Tests ---

func TestEnergyMatchExact(t *testing.T) {
	assertFloat(t, 1.0, EnergyMatch(model.EnergyLow, model.EnergyLow), "low-low")
	assertFloat(t, 1.0, EnergyMatch(model.EnergyMedium, model.EnergyMedium), "med-med")
	assertFloat(t, 1.0, EnergyMatch(model.EnergyHigh, model.EnergyHigh), "high-high")
}

func TestEnergyMatchOneLevelBelow(t *testing.T) {
	assertFloat(t, 0.5, EnergyMatch(model.EnergyLow, model.EnergyMedium), "low task, medium user")
	assertFloat(t, 0.5, EnergyMatch(model.EnergyMedium, model.EnergyHigh), "medium task, high user")
}

func TestEnergyMatchTwoLevelsBelow(t *testing.T) {
	assertFloat(t, 0.2, EnergyMatch(model.EnergyLow, model.EnergyHigh), "low task, high user")
}

func TestEnergyMatchTaskRequiresMore(t *testing.T) {
	assertFloat(t, 0.1, EnergyMatch(model.EnergyHigh, model.EnergyLow), "high task, low user")
}

func TestEnergyMatchEmptyDefaults(t *testing.T) {
	assertFloat(t, 1.0, EnergyMatch("", ""), "empty defaults to medium match")
	assertFloat(t, 1.0, EnergyMatch(model.EnergyMedium, ""), "medium task, empty user")
}

// --- Streak Risk Tests ---

func TestStreakRiskNonHabit(t *testing.T) {
	assertFloat(t, 0.0, StreakRisk(nil, refTime), "non-habit")
}

func TestStreakRiskDailyNotCompleted(t *testing.T) {
	info := &HabitInfo{
		FreqType:       "daily",
		CompletedToday: false,
		CurrentStreak:  5,
	}
	assertFloat(t, 1.0, StreakRisk(info, refTime), "daily habit not completed with active streak")
}

func TestStreakRiskDailyNotCompletedNoStreak(t *testing.T) {
	info := &HabitInfo{
		FreqType:       "daily",
		CompletedToday: false,
		CurrentStreak:  0,
	}
	assertFloat(t, 0.7, StreakRisk(info, refTime), "daily habit not completed, no streak")
}

func TestStreakRiskDailyCompleted(t *testing.T) {
	info := &HabitInfo{
		FreqType:       "daily",
		CompletedToday: true,
		CurrentStreak:  5,
	}
	assertFloat(t, 0.0, StreakRisk(info, refTime), "daily habit completed")
}

func TestStreakRiskWeeklyTargetMet(t *testing.T) {
	info := &HabitInfo{
		FreqType:            "weekly",
		FreqTarget:          3,
		CompletionsThisWeek: 3,
	}
	assertFloat(t, 0.0, StreakRisk(info, refTime), "weekly target met")
}

func TestStreakRiskWeeklyPartiallyDone(t *testing.T) {
	// Wednesday (refTime is Wednesday), target 5, completed 2
	info := &HabitInfo{
		FreqType:            "weekly",
		FreqTarget:          5,
		CompletionsThisWeek: 2,
	}
	score := StreakRisk(info, refTime)
	assertGreater(t, score, 0.0, "weekly partially done should be > 0")
	assertLessOrEqual(t, score, 1.0, "weekly risk should be <= 1")
}

func TestStreakRiskWeeklyUrgentEndOfWeek(t *testing.T) {
	// Saturday: 1 day left (Sunday), need 3 more completions
	saturday := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	info := &HabitInfo{
		FreqType:            "weekly",
		FreqTarget:          5,
		CompletionsThisWeek: 2,
	}
	score := StreakRisk(info, saturday)
	assertGreater(t, score, 0.5, "should be highly urgent near end of week")
}

func TestStreakRiskWeeklyZeroTarget(t *testing.T) {
	info := &HabitInfo{
		FreqType:   "weekly",
		FreqTarget: 0,
	}
	assertFloat(t, 0.0, StreakRisk(info, refTime), "zero target")
}

// --- Age Boost Tests ---

func TestAgeBoostNewTask(t *testing.T) {
	score := AgeBoost(refTime, refTime)
	assertFloat(t, 0.0, score, "brand new task")
}

func TestAgeBoostOneDay(t *testing.T) {
	created := refTime.AddDate(0, 0, -1)
	score := AgeBoost(created, refTime)
	assertGreater(t, score, 0.0, "1 day old should be > 0")
	if score >= 0.3 {
		t.Errorf("1 day old should be < 0.3, got %v", score)
	}
}

func TestAgeBoost30Days(t *testing.T) {
	created := refTime.AddDate(0, 0, -30)
	score := AgeBoost(created, refTime)
	assertGreater(t, score, 0.5, "30 days should be > 0.5")
	if score >= 1.0 {
		t.Errorf("30 days should be < 1.0, got %v", score)
	}
}

func TestAgeBoost90Days(t *testing.T) {
	created := refTime.AddDate(0, 0, -90)
	score := AgeBoost(created, refTime)
	assertInDelta(t, 1.0, score, 0.05, "should be near 1.0 at 90 days")
}

func TestAgeBoost180Days(t *testing.T) {
	created := refTime.AddDate(0, 0, -180)
	score := AgeBoost(created, refTime)
	assertFloat(t, 1.0, score, "should be capped at 1.0")
}

func TestAgeBoostMonotonicallyIncreasing(t *testing.T) {
	var prev float64
	for days := 1; days <= 90; days++ {
		created := refTime.AddDate(0, 0, -days)
		score := AgeBoost(created, refTime)
		assertGreaterOrEqual(t, score, prev, "age boost must be monotonically increasing")
		prev = score
	}
}

// --- Dependency Unlock Tests ---

func TestDependencyUnlockNoBlocked(t *testing.T) {
	assertFloat(t, 0.0, DependencyUnlock(0, 100), "no blocked tasks")
}

func TestDependencyUnlockOneBlocked(t *testing.T) {
	score := DependencyUnlock(1, 100)
	assertGreater(t, score, 0.0, "blocking 1 task should be > 0")
	if score >= 1.0 {
		t.Errorf("blocking 1 of 100 should be < 1.0, got %v", score)
	}
}

func TestDependencyUnlockManyBlocked(t *testing.T) {
	// 10% of 100 tasks blocked → max score
	score := DependencyUnlock(10, 100)
	assertFloat(t, 1.0, score, "blocking 10% of tasks should max out")
}

func TestDependencyUnlockCappedAtOne(t *testing.T) {
	score := DependencyUnlock(50, 100)
	assertFloat(t, 1.0, score, "should be capped at 1.0")
}

func TestDependencyUnlockSmallPool(t *testing.T) {
	score := DependencyUnlock(1, 5)
	assertGreater(t, score, 0.0, "blocking 1 of 5 should be > 0")
}

func TestDependencyUnlockZeroTotal(t *testing.T) {
	score := DependencyUnlock(1, 0)
	assertFloat(t, 1.0, score, "blocking anything with zero total is max")
}

// --- ScoreTask Integration Tests ---

func TestScoreTaskBasic(t *testing.T) {
	task := &model.Task{
		ID:       "test-task",
		Title:    "Test task",
		Type:     model.TaskTypeDated,
		Status:   model.StatusPending,
		Priority: model.PriorityHigh,
		Energy:   model.EnergyMedium,
		DueDate:  "2026-04-09", // tomorrow
		Created:  refTime.AddDate(0, 0, -7),
	}

	ctx := TaskContext{
		Task:         task,
		Now:          refTime,
		UserEnergy:   model.EnergyMedium,
		HabitInfo:    nil,
		BlockedCount: 2,
		TotalTasks:   50,
	}

	bd := ScoreTask(ctx, DefaultWeights)

	assertFloat(t, 0.8, bd.Urgency, "due tomorrow")
	assertFloat(t, 0.75, bd.Priority, "high priority")
	assertFloat(t, 1.0, bd.EnergyMatch, "exact energy match")
	assertFloat(t, 0.0, bd.StreakRisk, "not a habit")
	assertGreater(t, bd.AgeBoost, 0.0, "7 days old")
	assertGreater(t, bd.DependencyUnlock, 0.0, "blocks 2 tasks")
	assertGreater(t, bd.Total, 0.0, "total score should be positive")
}

func TestScoreTaskHabitWithStreakRisk(t *testing.T) {
	task := &model.Task{
		ID:       "habit-task",
		Title:    "Daily meditation",
		Type:     model.TaskTypeHabit,
		Status:   model.StatusPending,
		Priority: model.PriorityMedium,
		Energy:   model.EnergyLow,
		Created:  refTime.AddDate(0, -1, 0), // 1 month old
	}

	ctx := TaskContext{
		Task:       task,
		Now:        refTime,
		UserEnergy: model.EnergyMedium,
		HabitInfo: &HabitInfo{
			FreqType:       "daily",
			CompletedToday: false,
			CurrentStreak:  15,
		},
		BlockedCount: 0,
		TotalTasks:   30,
	}

	bd := ScoreTask(ctx, DefaultWeights)

	assertFloat(t, 1.0, bd.StreakRisk, "daily habit not completed with active streak")
	assertFloat(t, 0.5, bd.EnergyMatch, "low task for medium user = one level below")
	assertGreater(t, bd.Total, 0.3, "should score well due to streak risk")
}

func TestScoreTaskAllZero(t *testing.T) {
	task := &model.Task{
		ID:      "zero-task",
		Title:   "Zero score task",
		Type:    model.TaskTypeOneShot,
		Status:  model.StatusPending,
		Created: refTime, // just created
	}

	ctx := TaskContext{
		Task:         task,
		Now:          refTime,
		UserEnergy:   model.EnergyMedium,
		HabitInfo:    nil,
		BlockedCount: 0,
		TotalTasks:   10,
	}

	bd := ScoreTask(ctx, DefaultWeights)
	assertFloat(t, 0.0, bd.Urgency, "no due date")
	assertFloat(t, 0.0, bd.Priority, "no priority")
	assertFloat(t, 0.0, bd.StreakRisk, "not a habit")
	assertFloat(t, 0.0, bd.AgeBoost, "just created")
	assertFloat(t, 0.0, bd.DependencyUnlock, "blocks nothing")
}

// --- Weights Tests ---

func TestDefaultWeightsSum(t *testing.T) {
	w := DefaultWeights
	sum := w.Urgency + w.Priority + w.EnergyMatch +
		w.StreakRisk + w.AgeBoost + w.DependencyUnlock
	// Should sum to 0.95 (user_preference=0.05 is deferred)
	assertInDelta(t, 0.95, sum, 0.001, "default weights sum")
}

func TestScoreTaskMaximumScore(t *testing.T) {
	task := &model.Task{
		ID:       "max-task",
		Title:    "Maximum urgency task",
		Type:     model.TaskTypeDated,
		Status:   model.StatusPending,
		Priority: model.PriorityUrgent,
		Energy:   model.EnergyMedium,
		DueDate:  "2026-04-08", // due today
		Created:  refTime.AddDate(0, 0, -120),
	}

	ctx := TaskContext{
		Task:       task,
		Now:        refTime,
		UserEnergy: model.EnergyMedium,
		HabitInfo: &HabitInfo{
			FreqType:       "daily",
			CompletedToday: false,
			CurrentStreak:  10,
		},
		BlockedCount: 50,
		TotalTasks:   100,
	}

	bd := ScoreTask(ctx, DefaultWeights)

	assertFloat(t, 1.0, bd.Urgency, "due today")
	assertFloat(t, 1.0, bd.Priority, "urgent priority")
	assertFloat(t, 1.0, bd.EnergyMatch, "exact match")
	assertFloat(t, 1.0, bd.StreakRisk, "streak at risk")
	assertFloat(t, 1.0, bd.AgeBoost, "120 days old")
	assertFloat(t, 1.0, bd.DependencyUnlock, "blocks many")
	assertInDelta(t, 0.95, bd.Total, 0.001, "max total = sum of weights")
}

// --- Helper Tests ---

func TestEnergyLevel(t *testing.T) {
	if energyLevel(model.EnergyLow) != 1 {
		t.Error("EnergyLow should be 1")
	}
	if energyLevel(model.EnergyMedium) != 2 {
		t.Error("EnergyMedium should be 2")
	}
	if energyLevel(model.EnergyHigh) != 3 {
		t.Error("EnergyHigh should be 3")
	}
	if energyLevel("") != 2 {
		t.Error("empty energy should default to 2 (medium)")
	}
}

func TestTruncateToDay(t *testing.T) {
	input := time.Date(2026, 4, 8, 15, 30, 45, 123, time.UTC)
	expected := time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC)
	got := truncateToDay(input)
	if !got.Equal(expected) {
		t.Errorf("truncateToDay: expected %v, got %v", expected, got)
	}
}

// Verify score components are in [0,1] range for various inputs
func TestScoreComponentsRange(t *testing.T) {
	tasks := []*model.Task{
		{Type: model.TaskTypeOneShot, Priority: model.PriorityNone, Created: refTime},
		{Type: model.TaskTypeDated, Priority: model.PriorityUrgent, DueDate: "2026-04-08", Energy: model.EnergyHigh, Created: refTime.AddDate(0, -6, 0)},
		{Type: model.TaskTypeFloating, Created: refTime.AddDate(-1, 0, 0)},
		{Type: model.TaskTypeHabit, Priority: model.PriorityMedium, Energy: model.EnergyLow, Created: refTime.AddDate(0, 0, -45)},
	}

	habitInfos := []*HabitInfo{
		nil,
		{FreqType: "daily", CompletedToday: true, CurrentStreak: 10},
		{FreqType: "weekly", FreqTarget: 5, CompletionsThisWeek: 2},
	}

	for i, task := range tasks {
		for j, hi := range habitInfos {
			ctx := TaskContext{
				Task:         task,
				Now:          refTime,
				UserEnergy:   model.EnergyMedium,
				HabitInfo:    hi,
				BlockedCount: i,
				TotalTasks:   10,
			}

			bd := ScoreTask(ctx, DefaultWeights)
			fields := []struct {
				name  string
				value float64
			}{
				{"urgency", bd.Urgency},
				{"priority", bd.Priority},
				{"energy_match", bd.EnergyMatch},
				{"streak_risk", bd.StreakRisk},
				{"age_boost", bd.AgeBoost},
				{"dependency_unlock", bd.DependencyUnlock},
			}

			for _, f := range fields {
				if f.value < 0 || f.value > 1.0 {
					t.Errorf("task %d, habit %d: %s = %v, want [0,1]", i, j, f.name, f.value)
				}
			}
			if math.IsNaN(bd.Total) {
				t.Errorf("task %d, habit %d: total is NaN", i, j)
			}
		}
	}
}

// --- Benchmarks ---

func BenchmarkScoreTask(b *testing.B) {
	task := &model.Task{
		ID:       "bench-task",
		Title:    "Benchmark task",
		Type:     model.TaskTypeDated,
		Status:   model.StatusPending,
		Priority: model.PriorityHigh,
		Energy:   model.EnergyMedium,
		DueDate:  "2026-04-09",
		Created:  refTime.AddDate(0, 0, -7),
	}

	ctx := TaskContext{
		Task:       task,
		Now:        refTime,
		UserEnergy: model.EnergyMedium,
		HabitInfo: &HabitInfo{
			FreqType:       "daily",
			CompletedToday: false,
			CurrentStreak:  5,
		},
		BlockedCount: 3,
		TotalTasks:   100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScoreTask(ctx, DefaultWeights)
	}
}

func BenchmarkUrgency(b *testing.B) {
	task := &model.Task{DueDate: "2026-04-09"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Urgency(task, refTime)
	}
}
