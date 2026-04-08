package engine

import (
	"testing"
	"time"

	"github.com/tesserabox/bentotask/internal/model"
)

// makeTask is a helper to create a minimal task for packing tests.
func makeTask(id string, dur int, priority model.Priority, energy model.Energy) *model.Task {
	return &model.Task{
		ID:                id,
		Title:             "Task " + id,
		Type:              model.TaskTypeOneShot,
		Status:            model.StatusPending,
		Priority:          priority,
		Energy:            energy,
		EstimatedDuration: dur,
		Created:           refTime.AddDate(0, 0, -7),
	}
}

func baseRequest(tasks []*model.Task) PackRequest {
	return PackRequest{
		AvailableTime:     60,
		UserEnergy:        model.EnergyMedium,
		Now:               refTime,
		Weights:           DefaultWeights,
		Tasks:             tasks,
		HabitInfoMap:      make(map[string]*HabitInfo),
		BlockedByMap:      make(map[string]int),
		UnmetDependencies: make(map[string]bool),
	}
}

// --- Filter Tests ---

func TestFilterExcludesDoneTasks(t *testing.T) {
	done := makeTask("done", 10, model.PriorityMedium, model.EnergyMedium)
	done.Status = model.StatusDone

	pending := makeTask("pending", 10, model.PriorityMedium, model.EnergyMedium)

	req := baseRequest([]*model.Task{done, pending})
	result := Pack(req)

	if len(result.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(result.Suggestions))
	}
	if result.Suggestions[0].Task.ID != "pending" {
		t.Errorf("expected pending task, got %s", result.Suggestions[0].Task.ID)
	}
}

func TestFilterExcludesCancelledTasks(t *testing.T) {
	cancelled := makeTask("cancelled", 10, model.PriorityMedium, model.EnergyMedium)
	cancelled.Status = model.StatusCancelled

	req := baseRequest([]*model.Task{cancelled})
	result := Pack(req)

	if len(result.Suggestions) != 0 {
		t.Fatalf("expected 0 suggestions for cancelled tasks, got %d", len(result.Suggestions))
	}
}

func TestFilterIncludesActiveTasks(t *testing.T) {
	active := makeTask("active", 10, model.PriorityMedium, model.EnergyMedium)
	active.Status = model.StatusActive

	req := baseRequest([]*model.Task{active})
	result := Pack(req)

	if len(result.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion for active task, got %d", len(result.Suggestions))
	}
}

func TestFilterExcludesRoutines(t *testing.T) {
	routine := makeTask("routine", 10, model.PriorityMedium, model.EnergyMedium)
	routine.Type = model.TaskTypeRoutine

	req := baseRequest([]*model.Task{routine})
	result := Pack(req)

	if len(result.Suggestions) != 0 {
		t.Fatalf("expected 0 suggestions for routines, got %d", len(result.Suggestions))
	}
}

func TestFilterContextMatch(t *testing.T) {
	home := makeTask("home-task", 10, model.PriorityMedium, model.EnergyMedium)
	home.Context = []string{"home"}

	office := makeTask("office-task", 10, model.PriorityMedium, model.EnergyMedium)
	office.Context = []string{"office"}

	anywhere := makeTask("any-task", 10, model.PriorityMedium, model.EnergyMedium)
	// no context = available anywhere

	req := baseRequest([]*model.Task{home, office, anywhere})
	req.Context = "home"
	result := Pack(req)

	// Should include home-task and any-task, but not office-task
	if len(result.Suggestions) != 2 {
		t.Fatalf("expected 2 suggestions for home context, got %d", len(result.Suggestions))
	}

	ids := map[string]bool{}
	for _, s := range result.Suggestions {
		ids[s.Task.ID] = true
	}
	if !ids["home-task"] {
		t.Error("expected home-task in suggestions")
	}
	if !ids["any-task"] {
		t.Error("expected any-task in suggestions")
	}
	if ids["office-task"] {
		t.Error("office-task should not be in suggestions")
	}
}

func TestFilterNoContextMatchesAll(t *testing.T) {
	home := makeTask("home-task", 10, model.PriorityMedium, model.EnergyMedium)
	home.Context = []string{"home"}

	office := makeTask("office-task", 10, model.PriorityMedium, model.EnergyMedium)
	office.Context = []string{"office"}

	req := baseRequest([]*model.Task{home, office})
	req.Context = "" // any context
	result := Pack(req)

	if len(result.Suggestions) != 2 {
		t.Fatalf("expected 2 suggestions with no context filter, got %d", len(result.Suggestions))
	}
}

func TestFilterEnergyExclusion(t *testing.T) {
	high := makeTask("high-energy", 10, model.PriorityMedium, model.EnergyHigh)
	low := makeTask("low-energy", 10, model.PriorityMedium, model.EnergyLow)

	req := baseRequest([]*model.Task{high, low})
	req.UserEnergy = model.EnergyLow
	result := Pack(req)

	// Only low-energy task should be eligible
	if len(result.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion for low energy user, got %d", len(result.Suggestions))
	}
	if result.Suggestions[0].Task.ID != "low-energy" {
		t.Errorf("expected low-energy task, got %s", result.Suggestions[0].Task.ID)
	}
}

func TestFilterDurationExclusion(t *testing.T) {
	long := makeTask("long", 120, model.PriorityMedium, model.EnergyMedium)
	short := makeTask("short", 15, model.PriorityMedium, model.EnergyMedium)

	req := baseRequest([]*model.Task{long, short})
	req.AvailableTime = 60
	result := Pack(req)

	if len(result.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion (short only), got %d", len(result.Suggestions))
	}
	if result.Suggestions[0].Task.ID != "short" {
		t.Errorf("expected short task, got %s", result.Suggestions[0].Task.ID)
	}
}

func TestFilterUnmetDependencies(t *testing.T) {
	blocked := makeTask("blocked", 10, model.PriorityUrgent, model.EnergyMedium)
	free := makeTask("free", 10, model.PriorityLow, model.EnergyMedium)

	req := baseRequest([]*model.Task{blocked, free})
	req.UnmetDependencies = map[string]bool{"blocked": true}
	result := Pack(req)

	if len(result.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion (free only), got %d", len(result.Suggestions))
	}
	if result.Suggestions[0].Task.ID != "free" {
		t.Errorf("expected free task, got %s", result.Suggestions[0].Task.ID)
	}
}

// --- Packing Tests ---

func TestPackEmptyTasks(t *testing.T) {
	req := baseRequest(nil)
	result := Pack(req)

	if len(result.Suggestions) != 0 {
		t.Fatalf("expected 0 suggestions, got %d", len(result.Suggestions))
	}
	if result.TimeRemaining != 60 {
		t.Errorf("expected 60 min remaining, got %d", result.TimeRemaining)
	}
}

func TestPackSingleTask(t *testing.T) {
	task := makeTask("single", 30, model.PriorityMedium, model.EnergyMedium)
	req := baseRequest([]*model.Task{task})
	result := Pack(req)

	if len(result.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(result.Suggestions))
	}
	if result.TotalDuration != 30 {
		t.Errorf("expected 30 min total, got %d", result.TotalDuration)
	}
	if result.TimeRemaining != 30 {
		t.Errorf("expected 30 min remaining, got %d", result.TimeRemaining)
	}
}

func TestPackMultipleTasksFitExactly(t *testing.T) {
	tasks := []*model.Task{
		makeTask("a", 20, model.PriorityMedium, model.EnergyMedium),
		makeTask("b", 20, model.PriorityMedium, model.EnergyMedium),
		makeTask("c", 20, model.PriorityMedium, model.EnergyMedium),
	}

	req := baseRequest(tasks)
	req.AvailableTime = 60
	result := Pack(req)

	if len(result.Suggestions) != 3 {
		t.Fatalf("expected 3 suggestions, got %d", len(result.Suggestions))
	}
	if result.TotalDuration != 60 {
		t.Errorf("expected 60 min total, got %d", result.TotalDuration)
	}
	if result.TimeRemaining != 0 {
		t.Errorf("expected 0 remaining, got %d", result.TimeRemaining)
	}
}

func TestPackPrefersHigherScorePerMinute(t *testing.T) {
	// High priority but long → lower ratio
	longHigh := makeTask("long-high", 60, model.PriorityHigh, model.EnergyMedium)
	// Medium priority but short → higher ratio
	shortMed := makeTask("short-med", 10, model.PriorityMedium, model.EnergyMedium)

	req := baseRequest([]*model.Task{longHigh, shortMed})
	req.AvailableTime = 60
	result := Pack(req)

	// Both should fit (10 + ... but let's check ordering)
	if len(result.Suggestions) == 0 {
		t.Fatal("expected at least 1 suggestion")
	}

	// The shorter task should be packed first (higher ratio)
	if result.Suggestions[0].Task.ID != "short-med" {
		t.Errorf("expected short-med first (higher score/min ratio), got %s", result.Suggestions[0].Task.ID)
	}
}

func TestPackKnapsackDoesntOverflow(t *testing.T) {
	tasks := []*model.Task{
		makeTask("a", 40, model.PriorityHigh, model.EnergyMedium),
		makeTask("b", 40, model.PriorityMedium, model.EnergyMedium),
		makeTask("c", 10, model.PriorityLow, model.EnergyMedium),
	}

	req := baseRequest(tasks)
	req.AvailableTime = 60
	result := Pack(req)

	if result.TotalDuration > 60 {
		t.Errorf("packed %d min into 60 min slot — overflow!", result.TotalDuration)
	}
	if result.TimeRemaining < 0 {
		t.Errorf("negative time remaining: %d", result.TimeRemaining)
	}
}

func TestPackFFDFillsGaps(t *testing.T) {
	// Setup: two high-ratio tasks that greedy packs first (25 + 25 = 50 min),
	// leaving a 10-min gap. The low-priority 10-min task should be filled
	// by FFD even though its ratio was too low for the greedy pass.
	highA := makeTask("highA", 25, model.PriorityUrgent, model.EnergyMedium)
	highA.DueDate = "2026-04-08" // due today → max urgency

	highB := makeTask("highB", 25, model.PriorityUrgent, model.EnergyMedium)
	highB.DueDate = "2026-04-08"

	// This 30-min medium task won't fit after the greedy pass packs 50 min
	medium := makeTask("medium", 30, model.PriorityMedium, model.EnergyMedium)

	// This 10-min low task should be filled by FFD into the remaining gap
	filler := makeTask("filler", 10, model.PriorityLow, model.EnergyMedium)

	req := baseRequest([]*model.Task{highA, highB, medium, filler})
	req.AvailableTime = 60
	result := Pack(req)

	ids := map[string]bool{}
	for _, s := range result.Suggestions {
		ids[s.Task.ID] = true
	}

	if !ids["highA"] || !ids["highB"] {
		t.Error("expected both high-priority tasks to be packed by greedy pass")
	}
	if !ids["filler"] {
		t.Error("expected filler task to be packed by FFD into remaining gap")
	}
	if result.TotalDuration > 60 {
		t.Errorf("total duration %d exceeds available 60", result.TotalDuration)
	}
	if result.TimeRemaining < 0 {
		t.Errorf("negative time remaining: %d", result.TimeRemaining)
	}
}

func TestPackDefaultDuration(t *testing.T) {
	task := makeTask("no-dur", 0, model.PriorityMedium, model.EnergyMedium)
	task.EstimatedDuration = 0 // no duration set

	req := baseRequest([]*model.Task{task})
	result := Pack(req)

	if len(result.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(result.Suggestions))
	}
	if result.Suggestions[0].Duration != DefaultDuration {
		t.Errorf("expected default duration %d, got %d", DefaultDuration, result.Suggestions[0].Duration)
	}
}

func TestPackScoreBreakdownPopulated(t *testing.T) {
	task := makeTask("scored", 15, model.PriorityHigh, model.EnergyMedium)
	task.DueDate = "2026-04-09" // tomorrow

	req := baseRequest([]*model.Task{task})
	result := Pack(req)

	if len(result.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(result.Suggestions))
	}

	bd := result.Suggestions[0].Score
	if bd.Urgency != 0.8 {
		t.Errorf("expected urgency 0.8, got %v", bd.Urgency)
	}
	if bd.Priority != 0.75 {
		t.Errorf("expected priority 0.75, got %v", bd.Priority)
	}
	if bd.Total <= 0 {
		t.Error("expected positive total score")
	}
}

func TestPackDependencyBoost(t *testing.T) {
	// Two tasks with same priority/energy/duration
	// One blocks 5 other tasks — it should rank higher
	blocker := makeTask("blocker", 20, model.PriorityMedium, model.EnergyMedium)
	normal := makeTask("normal", 20, model.PriorityMedium, model.EnergyMedium)

	req := baseRequest([]*model.Task{blocker, normal})
	req.BlockedByMap = map[string]int{"blocker": 5}

	result := Pack(req)

	if len(result.Suggestions) < 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(result.Suggestions))
	}

	// Blocker should rank first
	if result.Suggestions[0].Task.ID != "blocker" {
		t.Errorf("expected blocker first due to dependency boost, got %s", result.Suggestions[0].Task.ID)
	}
}

func TestPackHabitStreakRiskBoost(t *testing.T) {
	habit := makeTask("habit", 10, model.PriorityLow, model.EnergyLow)
	habit.Type = model.TaskTypeHabit

	regular := makeTask("regular", 10, model.PriorityMedium, model.EnergyMedium)

	req := baseRequest([]*model.Task{habit, regular})
	req.HabitInfoMap = map[string]*HabitInfo{
		"habit": {
			FreqType:       "daily",
			CompletedToday: false,
			CurrentStreak:  20, // long streak at risk!
		},
	}

	result := Pack(req)

	if len(result.Suggestions) < 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(result.Suggestions))
	}

	// Habit with long streak at risk should rank first despite lower priority
	if result.Suggestions[0].Task.ID != "habit" {
		t.Errorf("expected habit first due to streak risk, got %s", result.Suggestions[0].Task.ID)
	}
}

// --- TopN Tests ---

func TestTopNBasic(t *testing.T) {
	tasks := []*model.Task{
		makeTask("a", 10, model.PriorityLow, model.EnergyMedium),
		makeTask("b", 10, model.PriorityHigh, model.EnergyMedium),
		makeTask("c", 10, model.PriorityUrgent, model.EnergyMedium),
	}

	req := baseRequest(tasks)
	results := TopN(req, 2)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Urgent should be first
	if results[0].Task.ID != "c" {
		t.Errorf("expected urgent task first, got %s", results[0].Task.ID)
	}
	// High should be second
	if results[1].Task.ID != "b" {
		t.Errorf("expected high-priority task second, got %s", results[1].Task.ID)
	}
}

func TestTopNMoreThanAvailable(t *testing.T) {
	tasks := []*model.Task{
		makeTask("a", 10, model.PriorityMedium, model.EnergyMedium),
		makeTask("b", 10, model.PriorityHigh, model.EnergyMedium),
	}

	req := baseRequest(tasks)
	results := TopN(req, 10) // ask for 10, only 2 available

	if len(results) != 2 {
		t.Fatalf("expected 2 results (capped), got %d", len(results))
	}
}

func TestTopNEmpty(t *testing.T) {
	req := baseRequest(nil)
	results := TopN(req, 5)

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestTopNRespectsFilters(t *testing.T) {
	high := makeTask("high-energy", 10, model.PriorityUrgent, model.EnergyHigh)
	low := makeTask("low-energy", 10, model.PriorityLow, model.EnergyLow)

	req := baseRequest([]*model.Task{high, low})
	req.UserEnergy = model.EnergyLow
	results := TopN(req, 5)

	if len(results) != 1 {
		t.Fatalf("expected 1 result (low energy only), got %d", len(results))
	}
	if results[0].Task.ID != "low-energy" {
		t.Errorf("expected low-energy task, got %s", results[0].Task.ID)
	}
}

// --- Helper Tests ---

func TestMatchesContext(t *testing.T) {
	tests := []struct {
		name        string
		taskCtx     []string
		userCtx     string
		shouldMatch bool
	}{
		{"no context on task", nil, "home", true},
		{"empty context on task", []string{}, "home", true},
		{"exact match", []string{"home"}, "home", true},
		{"multi-context match", []string{"home", "office"}, "home", true},
		{"no match", []string{"office"}, "home", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &model.Task{Context: tt.taskCtx}
			got := matchesContext(task, tt.userCtx)
			if got != tt.shouldMatch {
				t.Errorf("matchesContext(%v, %q) = %v, want %v", tt.taskCtx, tt.userCtx, got, tt.shouldMatch)
			}
		})
	}
}

func TestEnergyFits(t *testing.T) {
	tests := []struct {
		task, user model.Energy
		fits       bool
	}{
		{model.EnergyLow, model.EnergyLow, true},
		{model.EnergyLow, model.EnergyMedium, true},
		{model.EnergyLow, model.EnergyHigh, true},
		{model.EnergyMedium, model.EnergyLow, false},
		{model.EnergyMedium, model.EnergyMedium, true},
		{model.EnergyMedium, model.EnergyHigh, true},
		{model.EnergyHigh, model.EnergyLow, false},
		{model.EnergyHigh, model.EnergyMedium, false},
		{model.EnergyHigh, model.EnergyHigh, true},
		{"", "", true},               // both default to medium
		{"", model.EnergyLow, false}, // default medium > low
	}

	for _, tt := range tests {
		got := energyFits(tt.task, tt.user)
		if got != tt.fits {
			t.Errorf("energyFits(%q, %q) = %v, want %v", tt.task, tt.user, got, tt.fits)
		}
	}
}

func TestEffectiveDuration(t *testing.T) {
	task := &model.Task{EstimatedDuration: 45}
	if effectiveDuration(task) != 45 {
		t.Errorf("expected 45, got %d", effectiveDuration(task))
	}

	task = &model.Task{EstimatedDuration: 0}
	if effectiveDuration(task) != DefaultDuration {
		t.Errorf("expected default %d, got %d", DefaultDuration, effectiveDuration(task))
	}
}

// --- Realistic scenario ---

func TestPackRealisticMorningPlan(t *testing.T) {
	// Simulate a morning planning session:
	// - 90 minutes available, medium energy, home context
	// - Mix of habits, deadlines, and floating tasks
	now := time.Date(2026, 4, 8, 7, 0, 0, 0, time.UTC)

	meditate := &model.Task{
		ID: "meditate", Title: "Meditate", Type: model.TaskTypeHabit,
		Status: model.StatusPending, Priority: model.PriorityMedium,
		Energy: model.EnergyLow, EstimatedDuration: 10,
		Context: []string{"home"}, Created: now.AddDate(0, -2, 0),
	}

	exercise := &model.Task{
		ID: "exercise", Title: "Exercise", Type: model.TaskTypeHabit,
		Status: model.StatusPending, Priority: model.PriorityHigh,
		Energy: model.EnergyHigh, EstimatedDuration: 30,
		Context: []string{"home"}, Created: now.AddDate(0, -2, 0),
	}

	deployV2 := &model.Task{
		ID: "deploy", Title: "Deploy v2.1", Type: model.TaskTypeDated,
		Status: model.StatusPending, Priority: model.PriorityUrgent,
		Energy: model.EnergyHigh, EstimatedDuration: 45,
		DueDate: "2026-04-08", Created: now.AddDate(0, 0, -3),
	}

	groceries := &model.Task{
		ID: "groceries", Title: "Buy groceries", Type: model.TaskTypeFloating,
		Status: model.StatusPending, Priority: model.PriorityLow,
		Energy: model.EnergyLow, EstimatedDuration: 30,
		Context: []string{"errands"}, Created: now.AddDate(0, 0, -14),
	}

	readBook := &model.Task{
		ID: "read", Title: "Read book chapter", Type: model.TaskTypeFloating,
		Status: model.StatusPending, Priority: model.PriorityLow,
		Energy: model.EnergyLow, EstimatedDuration: 20,
		Context: []string{"home"}, Created: now.AddDate(0, 0, -30),
	}

	tasks := []*model.Task{meditate, exercise, deployV2, groceries, readBook}

	req := PackRequest{
		AvailableTime: 90,
		UserEnergy:    model.EnergyMedium,
		Context:       "home",
		Now:           now,
		Weights:       DefaultWeights,
		Tasks:         tasks,
		HabitInfoMap: map[string]*HabitInfo{
			"meditate": {FreqType: "daily", CompletedToday: false, CurrentStreak: 30},
			"exercise": {FreqType: "daily", CompletedToday: false, CurrentStreak: 15},
		},
		BlockedByMap:      map[string]int{"deploy": 2},
		UnmetDependencies: map[string]bool{},
	}

	result := Pack(req)

	// Basic sanity checks
	if result.TotalDuration > 90 {
		t.Errorf("overpacked: %d > 90", result.TotalDuration)
	}
	if result.TimeRemaining < 0 {
		t.Errorf("negative remaining: %d", result.TimeRemaining)
	}
	if result.TotalDuration+result.TimeRemaining != 90 {
		t.Errorf("duration + remaining = %d, want 90", result.TotalDuration+result.TimeRemaining)
	}

	// Exercise and deploy require high energy but user has medium → filtered out
	for _, s := range result.Suggestions {
		if s.Task.ID == "exercise" {
			t.Error("exercise requires high energy, should be filtered (user has medium)")
		}
		if s.Task.ID == "deploy" {
			t.Error("deploy requires high energy, should be filtered (user has medium)")
		}
		if s.Task.ID == "groceries" {
			t.Error("groceries is errands context, should be filtered (user is home)")
		}
	}

	// Meditate should definitely be there (streak risk + home context + low energy)
	found := false
	for _, s := range result.Suggestions {
		if s.Task.ID == "meditate" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected meditate in suggestions (high streak risk, home context)")
	}

	t.Logf("Packed %d tasks, %d min used, %d min remaining",
		len(result.Suggestions), result.TotalDuration, result.TimeRemaining)
	for _, s := range result.Suggestions {
		t.Logf("  %s (%d min) — score=%.3f", s.Task.Title, s.Duration, s.Score.Total)
	}
}

// --- Benchmark ---

func BenchmarkPack(b *testing.B) {
	// Create 100 tasks
	tasks := make([]*model.Task, 100)
	for i := range tasks {
		p := model.PriorityLow
		if i%4 == 0 {
			p = model.PriorityHigh
		}
		tasks[i] = makeTask(
			"task-"+string(rune('A'+i%26)),
			10+i%50,
			p,
			model.EnergyMedium,
		)
		tasks[i].Created = refTime.AddDate(0, 0, -i)
	}

	req := baseRequest(tasks)
	req.AvailableTime = 240

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Pack(req)
	}
}
