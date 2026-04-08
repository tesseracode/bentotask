// Package app provides the core application logic for BentoTask.
//
// It sits between the CLI (or API) and the store layer, coordinating
// operations that touch both markdown files and the SQLite index.
// Commands should call App methods rather than using store directly.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tesserabox/bentotask/internal/engine"
	"github.com/tesserabox/bentotask/internal/habit"
	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/recurrence"
	"github.com/tesserabox/bentotask/internal/store"
)

// App is the core application. It manages the data directory, index,
// and provides high-level CRUD operations for tasks.
type App struct {
	DataDir  string
	Index    *store.Index
	indexDir string
}

// Open initializes the application with the given data directory.
// It creates the directory structure and opens the SQLite index.
func Open(dataDir string) (*App, error) {
	// Expand ~ if needed
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		dataDir = filepath.Join(home, ".bentotask", "data")
	}

	// Create the standard directory structure
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "inbox"),
		filepath.Join(dataDir, ".bentotask"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("create %s: %w", d, err)
		}
	}

	// Open the index
	indexPath := filepath.Join(dataDir, ".bentotask", "index.db")
	idx, err := store.OpenIndex(indexPath)
	if err != nil {
		return nil, fmt.Errorf("open index: %w", err)
	}

	return &App{
		DataDir:  dataDir,
		Index:    idx,
		indexDir: filepath.Join(dataDir, ".bentotask"),
	}, nil
}

// Close shuts down the application cleanly.
func (a *App) Close() error {
	return a.Index.Close()
}

// AddTask creates a new task, writes it to disk, and indexes it.
func (a *App) AddTask(title string, opts TaskOptions) (*model.Task, error) {
	now := time.Now().UTC()
	task := &model.Task{
		ID:      model.NewID(),
		Title:   title,
		Type:    model.TaskTypeOneShot,
		Status:  model.StatusPending,
		Created: now,
		Updated: now,
	}

	// Apply options
	if opts.Type != "" {
		task.Type = opts.Type
	}
	if opts.Priority != "" {
		task.Priority = opts.Priority
	}
	if opts.Energy != "" {
		task.Energy = opts.Energy
	}
	if opts.Duration > 0 {
		task.EstimatedDuration = opts.Duration
	}
	if opts.DueDate != "" {
		task.DueDate = opts.DueDate
		if task.Type == model.TaskTypeOneShot {
			task.Type = model.TaskTypeDated
		}
	}
	if opts.DueStart != "" {
		task.DueStart = opts.DueStart
	}
	if opts.DueEnd != "" {
		task.DueEnd = opts.DueEnd
		if task.DueStart != "" && task.Type == model.TaskTypeOneShot {
			task.Type = model.TaskTypeRanged
		}
	}
	if len(opts.Tags) > 0 {
		task.Tags = opts.Tags
	}
	if len(opts.Context) > 0 {
		task.Context = opts.Context
	}
	if opts.Box != "" {
		task.Box = opts.Box
	}

	// Validate
	if errs := task.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("validation failed: %s", errs[0])
	}

	// Determine file path based on box
	relPath := taskFilePath(task)
	absPath := filepath.Join(a.DataDir, relPath)

	// Write to disk
	if err := store.WriteFile(absPath, task); err != nil {
		return nil, fmt.Errorf("write task: %w", err)
	}

	// Index
	if err := a.Index.UpsertTask(task, relPath); err != nil {
		return nil, fmt.Errorf("index task: %w", err)
	}

	return task, nil
}

// ListTasks returns tasks matching the given filter.
func (a *App) ListTasks(f *store.TaskFilter) ([]*store.IndexedTask, error) {
	return a.Index.ListTasks(f)
}

// SearchTasks performs a full-text search across task titles and bodies.
func (a *App) SearchTasks(query string) ([]*store.IndexedTask, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}
	return a.Index.Search(query)
}

// GetTask retrieves a single task by exact ID or prefix.
// If a prefix matches exactly one task, it returns that task.
func (a *App) GetTask(idOrPrefix string) (*model.Task, string, error) {
	// Try exact match first
	indexed, err := a.Index.GetTask(idOrPrefix)
	if err == nil {
		return a.loadTask(indexed)
	}

	// Try prefix match
	matches, err := a.Index.FindByPrefix(idOrPrefix)
	if err != nil {
		return nil, "", fmt.Errorf("search: %w", err)
	}

	switch len(matches) {
	case 0:
		return nil, "", fmt.Errorf("no task matching %q", idOrPrefix)
	case 1:
		return a.loadTask(matches[0])
	default:
		return nil, "", fmt.Errorf("ambiguous prefix %q matches %d tasks", idOrPrefix, len(matches))
	}
}

// UpdateTask applies edits to a task, saves to disk, and re-indexes.
// The apply function receives the task and can modify any fields.
// Updated timestamp is set automatically.
func (a *App) UpdateTask(idOrPrefix string, apply func(*model.Task)) (*model.Task, error) {
	task, relPath, err := a.GetTask(idOrPrefix)
	if err != nil {
		return nil, err
	}

	apply(task)
	task.Updated = time.Now().UTC()

	if errs := task.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("validation failed: %s", errs[0])
	}

	if err := a.saveTask(task, relPath); err != nil {
		return nil, err
	}

	return task, nil
}

// EditTaskFile returns the absolute path to a task's .md file.
// Used by the CLI to open the file in $EDITOR.
func (a *App) EditTaskFile(idOrPrefix string) (absPath string, err error) {
	_, relPath, err := a.GetTask(idOrPrefix)
	if err != nil {
		return "", err
	}
	return filepath.Join(a.DataDir, relPath), nil
}

// ReloadTask re-reads a task from disk and updates the index.
// Called after $EDITOR closes to pick up any changes.
func (a *App) ReloadTask(idOrPrefix string) (*model.Task, error) {
	// We need to find the file path from the index first
	task, relPath, err := a.GetTask(idOrPrefix)
	if err != nil {
		return nil, err
	}

	// Re-read from disk (editor may have changed it)
	absPath := filepath.Join(a.DataDir, relPath)
	updated, err := store.ParseFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("reload %s: %w", relPath, err)
	}

	// Re-index
	if err := a.Index.UpsertTask(updated, relPath); err != nil {
		return nil, fmt.Errorf("re-index %s: %w", relPath, err)
	}

	_ = task // silence unused warning in case original was needed
	return updated, nil
}

// CompleteTask marks a task as done.
func (a *App) CompleteTask(idOrPrefix string) (*model.Task, error) {
	task, relPath, err := a.GetTask(idOrPrefix)
	if err != nil {
		return nil, err
	}

	if task.IsDone() {
		return nil, fmt.Errorf("task %q is already %s", task.Title, task.Status)
	}

	now := time.Now().UTC()
	task.Status = model.StatusDone
	task.CompletedAt = &now
	task.Updated = now

	return task, a.saveTask(task, relPath)
}

// DeleteTask removes a task from disk and the index.
func (a *App) DeleteTask(idOrPrefix string) (*model.Task, error) {
	task, relPath, err := a.GetTask(idOrPrefix)
	if err != nil {
		return nil, err
	}

	absPath := filepath.Join(a.DataDir, relPath)
	if err := os.Remove(absPath); err != nil {
		return nil, fmt.Errorf("delete file: %w", err)
	}

	if err := a.Index.DeleteTask(task.ID); err != nil {
		return nil, fmt.Errorf("remove from index: %w", err)
	}

	return task, nil
}

// RebuildIndex drops the index and rebuilds it from all .md files.
func (a *App) RebuildIndex() (int, error) {
	return a.Index.RebuildIndex(a.DataDir)
}

// --- Task Linking ---

// LinkTasks creates a link between two tasks. It validates that both tasks exist,
// the link type is valid, and the link doesn't create a dependency cycle.
func (a *App) LinkTasks(sourceIDOrPrefix, targetIDOrPrefix string, linkType model.LinkType) (*model.Task, *model.Task, error) {
	// Validate link type
	if !model.IsValidLinkType(linkType) {
		return nil, nil, fmt.Errorf("invalid link type: %q (valid: depends-on, blocks, related-to)", linkType)
	}

	// Resolve both tasks
	source, sourceRel, err := a.GetTask(sourceIDOrPrefix)
	if err != nil {
		return nil, nil, fmt.Errorf("source task: %w", err)
	}
	target, _, err := a.GetTask(targetIDOrPrefix)
	if err != nil {
		return nil, nil, fmt.Errorf("target task: %w", err)
	}

	// Cannot link a task to itself
	if source.ID == target.ID {
		return nil, nil, fmt.Errorf("cannot link a task to itself")
	}

	// Check for duplicate link
	for _, existing := range source.Links {
		if existing.Target == target.ID && existing.Type == linkType {
			return nil, nil, fmt.Errorf("link already exists: %s %s %s",
				source.ShortID(8), linkType, target.ShortID(8))
		}
	}

	// For depends-on and blocks, check for cycles
	if linkType == model.LinkDependsOn || linkType == model.LinkBlocks {
		graph, err := a.Index.DependencyGraph()
		if err != nil {
			return nil, nil, fmt.Errorf("load dependency graph: %w", err)
		}
		// Add the proposed edge and check for cycles
		graph[source.ID] = append(graph[source.ID], target.ID)
		if hasCycle(graph, source.ID) {
			return nil, nil, fmt.Errorf("link would create a dependency cycle")
		}
	}

	// Add the link to source task
	source.Links = append(source.Links, model.Link{
		Type:   linkType,
		Target: target.ID,
	})
	source.Updated = time.Now().UTC()

	// Save source
	if err := a.saveTask(source, sourceRel); err != nil {
		return nil, nil, fmt.Errorf("save source: %w", err)
	}

	return source, target, nil
}

// UnlinkTasks removes a link between two tasks.
func (a *App) UnlinkTasks(sourceIDOrPrefix, targetIDOrPrefix string, linkType model.LinkType) (*model.Task, *model.Task, error) {
	source, sourceRel, err := a.GetTask(sourceIDOrPrefix)
	if err != nil {
		return nil, nil, fmt.Errorf("source task: %w", err)
	}
	target, _, err := a.GetTask(targetIDOrPrefix)
	if err != nil {
		return nil, nil, fmt.Errorf("target task: %w", err)
	}

	// Find and remove the link
	found := false
	var remaining []model.Link
	for _, link := range source.Links {
		if link.Target == target.ID && link.Type == linkType {
			found = true
			continue
		}
		remaining = append(remaining, link)
	}
	if !found {
		return nil, nil, fmt.Errorf("no %s link from %s to %s",
			linkType, source.ShortID(8), target.ShortID(8))
	}

	source.Links = remaining
	source.Updated = time.Now().UTC()

	if err := a.saveTask(source, sourceRel); err != nil {
		return nil, nil, fmt.Errorf("save source: %w", err)
	}

	return source, target, nil
}

// TaskLinkInfo holds the resolved data for a single task link relationship.
type TaskLinkInfo struct {
	Type      model.LinkType
	Direction string // "outgoing" or "incoming"
	TaskID    string
	TaskTitle string
}

// GetTaskLinks returns all links (outgoing + incoming) for a task with titles.
func (a *App) GetTaskLinks(taskID string) ([]TaskLinkInfo, error) {
	outgoing, err := a.Index.LoadLinks(taskID)
	if err != nil {
		return nil, err
	}
	incoming, err := a.Index.LoadBacklinks(taskID)
	if err != nil {
		return nil, err
	}

	var result []TaskLinkInfo

	for _, l := range outgoing {
		info := TaskLinkInfo{
			Type:      model.LinkType(l.LinkType),
			Direction: "outgoing",
			TaskID:    l.TargetID,
		}
		if indexed, err := a.Index.GetTask(l.TargetID); err == nil {
			info.TaskTitle = indexed.Title
		} else {
			info.TaskTitle = "(unknown)"
		}
		result = append(result, info)
	}

	for _, l := range incoming {
		info := TaskLinkInfo{
			Type:      model.LinkType(l.LinkType),
			Direction: "incoming",
			TaskID:    l.SourceID,
		}
		if indexed, err := a.Index.GetTask(l.SourceID); err == nil {
			info.TaskTitle = indexed.Title
		} else {
			info.TaskTitle = "(unknown)"
		}
		result = append(result, info)
	}

	return result, nil
}

// hasCycle performs a DFS to detect cycles in the dependency graph starting from start.
func hasCycle(graph map[string][]string, start string) bool {
	const (
		white = 0 // unvisited
		gray  = 1 // in current DFS path
		black = 2 // fully explored
	)

	color := make(map[string]int)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		color[node] = gray
		for _, neighbor := range graph[node] {
			switch color[neighbor] {
			case gray:
				return true // back edge → cycle
			case white:
				if dfs(neighbor) {
					return true
				}
			}
		}
		color[node] = black
		return false
	}

	return dfs(start)
}

// --- Routines ---

// AddRoutine creates a new routine with the given steps and optional schedule.
func (a *App) AddRoutine(title string, opts RoutineOptions) (*model.Task, error) {
	now := time.Now().UTC()

	task := &model.Task{
		ID:      model.NewID(),
		Title:   title,
		Type:    model.TaskTypeRoutine,
		Status:  model.StatusActive,
		Created: now,
		Updated: now,
		Steps:   opts.Steps,
	}

	if opts.Schedule != nil {
		task.Schedule = opts.Schedule
	}
	if opts.Priority != "" {
		task.Priority = opts.Priority
	}
	if opts.Energy != "" {
		task.Energy = opts.Energy
	}
	if len(opts.Tags) > 0 {
		task.Tags = opts.Tags
	}

	// Compute estimated duration from step durations
	totalDur := 0
	for _, s := range opts.Steps {
		totalDur += s.Duration
	}
	if totalDur > 0 {
		task.EstimatedDuration = totalDur
	}

	if errs := task.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("validation failed: %s", errs[0])
	}

	relPath := taskFilePath(task)
	absPath := filepath.Join(a.DataDir, relPath)

	if err := store.WriteFile(absPath, task); err != nil {
		return nil, fmt.Errorf("write routine: %w", err)
	}
	if err := a.Index.UpsertTask(task, relPath); err != nil {
		return nil, fmt.Errorf("index routine: %w", err)
	}

	return task, nil
}

// ListRoutines returns all routine tasks from the index.
func (a *App) ListRoutines() ([]*store.IndexedTask, error) {
	return a.Index.ListTasks(&store.TaskFilter{Type: model.TaskTypeRoutine})
}

// RoutineOptions holds options for creating a new routine.
type RoutineOptions struct {
	Steps    []model.RoutineStep
	Schedule *model.RoutineSchedule
	Priority model.Priority
	Energy   model.Energy
	Tags     []string
}

// --- Habits ---

// AddHabit creates a new habit task with the given frequency and recurrence.
func (a *App) AddHabit(title string, opts HabitOptions) (*model.Task, error) {
	now := time.Now().UTC()

	// Validate the recurrence rule
	if err := recurrence.Validate(opts.Recurrence); err != nil {
		return nil, fmt.Errorf("invalid recurrence: %w", err)
	}

	task := &model.Task{
		ID:      model.NewID(),
		Title:   title,
		Type:    model.TaskTypeHabit,
		Status:  model.StatusActive,
		Created: now,
		Updated: now,
		Frequency: &model.HabitFrequency{
			Type:   opts.FreqType,
			Target: opts.FreqTarget,
		},
		Recurrence: opts.Recurrence,
	}

	if opts.Priority != "" {
		task.Priority = opts.Priority
	}
	if opts.Energy != "" {
		task.Energy = opts.Energy
	}
	if len(opts.Tags) > 0 {
		task.Tags = opts.Tags
	}
	if len(opts.Context) > 0 {
		task.Context = opts.Context
	}

	if errs := task.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("validation failed: %s", errs[0])
	}

	relPath := taskFilePath(task)
	absPath := filepath.Join(a.DataDir, relPath)

	if err := store.WriteFile(absPath, task); err != nil {
		return nil, fmt.Errorf("write habit: %w", err)
	}
	if err := a.Index.UpsertTask(task, relPath); err != nil {
		return nil, fmt.Errorf("index habit: %w", err)
	}

	return task, nil
}

// LogHabit records a completion for a habit. It updates both the SQLite index
// and the markdown body (source of truth).
func (a *App) LogHabit(idOrPrefix string, duration int, note string) (*model.Task, error) {
	task, relPath, err := a.GetTask(idOrPrefix)
	if err != nil {
		return nil, err
	}
	if task.Type != model.TaskTypeHabit {
		return nil, fmt.Errorf("task %q is not a habit (type: %s)", task.Title, task.Type)
	}

	now := time.Now().UTC()

	// Append to markdown body
	c := habit.Completion{
		CompletedAt: now,
		Duration:    duration,
		Note:        note,
	}
	task.Body = habit.AppendCompletionToBody(task.Body, c)
	task.Updated = now

	// Recalculate streaks from all completions in the body
	completions := habit.ParseCompletionsFromBody(task.Body)
	freqType := "daily"
	if task.Frequency != nil {
		freqType = task.Frequency.Type
	}
	stats := habit.CalculateStreak(completions, freqType)
	task.StreakCurrent = stats.CurrentStreak
	task.StreakLongest = stats.LongestStreak

	// Write to disk
	absPath := filepath.Join(a.DataDir, relPath)
	if err := store.WriteFile(absPath, task); err != nil {
		return nil, fmt.Errorf("write habit: %w", err)
	}

	// Update index
	if err := a.Index.UpsertTask(task, relPath); err != nil {
		return nil, fmt.Errorf("index habit: %w", err)
	}

	// Log to SQLite completions table
	if err := a.Index.LogHabitCompletion(task.ID, now, duration, note); err != nil {
		return nil, fmt.Errorf("log completion: %w", err)
	}

	return task, nil
}

// HabitStats returns statistics for a habit including streaks and completion rate.
func (a *App) HabitStats(idOrPrefix string) (*model.Task, *habit.Stats, error) {
	task, _, err := a.GetTask(idOrPrefix)
	if err != nil {
		return nil, nil, err
	}
	if task.Type != model.TaskTypeHabit {
		return nil, nil, fmt.Errorf("task %q is not a habit (type: %s)", task.Title, task.Type)
	}

	completions := habit.ParseCompletionsFromBody(task.Body)

	freqType := "daily"
	target := 1
	if task.Frequency != nil {
		freqType = task.Frequency.Type
		if task.Frequency.Target > 0 {
			target = task.Frequency.Target
		}
	}

	stats := habit.CalculateStreak(completions, freqType)
	stats.CompletionRate = habit.CompletionRate(completions, freqType, target, 30)
	stats.RatePeriodDays = 30

	return task, &stats, nil
}

// ListHabits returns all habit tasks from the index.
func (a *App) ListHabits() ([]*store.IndexedTask, error) {
	return a.Index.ListTasks(&store.TaskFilter{Type: model.TaskTypeHabit})
}

// HabitOptions holds options for creating a new habit.
type HabitOptions struct {
	FreqType   string // "daily" or "weekly"
	FreqTarget int    // how many times per period
	Recurrence string // RRULE string
	Priority   model.Priority
	Energy     model.Energy
	Tags       []string
	Context    []string
}

// --- Shell Completions ---

// CompleteTasks returns task ID+title pairs for shell completion.
// Only returns non-done tasks for a better completion experience.
func (a *App) CompleteTasks() ([]string, error) {
	tasks, err := a.Index.ListTasks(nil)
	if err != nil {
		return nil, err
	}
	var comps []string
	for _, t := range tasks {
		if t.Status != "done" && t.Status != "cancelled" {
			comps = append(comps, t.ID+"\t"+t.Title)
		}
	}
	return comps, nil
}

// CompleteTags returns all distinct tags for shell completion.
func (a *App) CompleteTags() ([]string, error) {
	return a.Index.DistinctTags()
}

// CompleteBoxes returns all distinct boxes for shell completion.
func (a *App) CompleteBoxes() ([]string, error) {
	return a.Index.DistinctBoxes()
}

// CompleteContexts returns all distinct contexts for shell completion.
func (a *App) CompleteContexts() ([]string, error) {
	return a.Index.DistinctContexts()
}

// --- Smart Scheduling ---

// SuggestOptions holds options for the bt now / bt plan today commands.
type SuggestOptions struct {
	AvailableTime int          // available minutes (0 = no time limit)
	Energy        model.Energy // current energy level
	Context       string       // current context (empty = any)
}

// Suggest returns the top N task suggestions using the Bento Packing Algorithm.
// It loads all pending/active tasks, builds habit info and dependency data,
// then runs the scoring engine.
func (a *App) Suggest(opts SuggestOptions, n int) ([]engine.Suggestion, error) {
	req, err := a.buildPackRequest(opts)
	if err != nil {
		return nil, err
	}

	return engine.TopN(req, n), nil
}

// PlanDay packs tasks into the available time using the Bento Packing Algorithm.
// Returns the packed suggestions and metadata.
func (a *App) PlanDay(opts SuggestOptions) (*engine.PackResult, error) {
	req, err := a.buildPackRequest(opts)
	if err != nil {
		return nil, err
	}

	result := engine.Pack(req)
	return &result, nil
}

// buildPackRequest constructs a PackRequest from the app's data.
func (a *App) buildPackRequest(opts SuggestOptions) (engine.PackRequest, error) {
	now := time.Now().UTC()

	// Load all non-done tasks from the index
	indexed, err := a.Index.ListTasks(nil)
	if err != nil {
		return engine.PackRequest{}, fmt.Errorf("list tasks: %w", err)
	}

	// Load full tasks from disk (need body for habit completions)
	var tasks []*model.Task
	for _, it := range indexed {
		if it.Status == "done" || it.Status == "cancelled" {
			continue
		}
		task, _, loadErr := a.loadTask(it)
		if loadErr != nil {
			continue // skip unreadable tasks
		}
		tasks = append(tasks, task)
	}

	// Build habit info map
	habitInfoMap := make(map[string]*engine.HabitInfo)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for _, task := range tasks {
		if task.Type != model.TaskTypeHabit {
			continue
		}
		completions := habit.ParseCompletionsFromBody(task.Body)

		freqType := "daily"
		freqTarget := 1
		if task.Frequency != nil {
			freqType = task.Frequency.Type
			if task.Frequency.Target > 0 {
				freqTarget = task.Frequency.Target
			}
		}

		// Check if completed today
		completedToday := false
		completionsThisWeek := 0
		_, thisWeek := now.ISOWeek()
		thisYear := now.Year()
		for _, c := range completions {
			cDate := time.Date(c.CompletedAt.Year(), c.CompletedAt.Month(), c.CompletedAt.Day(), 0, 0, 0, 0, time.UTC)
			if cDate.Equal(today) {
				completedToday = true
			}
			cYear, cWeek := c.CompletedAt.ISOWeek()
			if cYear == thisYear && cWeek == thisWeek {
				completionsThisWeek++
			}
		}

		stats := habit.CalculateStreak(completions, freqType)

		habitInfoMap[task.ID] = &engine.HabitInfo{
			FreqType:            freqType,
			FreqTarget:          freqTarget,
			CompletedToday:      completedToday,
			CompletionsThisWeek: completionsThisWeek,
			CurrentStreak:       stats.CurrentStreak,
		}
	}

	// Build blocked-by map from dependency graph.
	// depGraph stores depends-on edges: source depends on target.
	// So if A depends-on B, completing B unblocks A.
	// We count: for each task T, how many tasks depend on T?
	blockedByMap := make(map[string]int)
	depGraph, err := a.Index.DependencyGraph()
	if err == nil {
		for _, targets := range depGraph {
			for _, target := range targets {
				blockedByMap[target]++
			}
		}
	}

	// Build unmet dependencies set.
	// A task is blocked if:
	//   1. It has a depends-on link to a task that isn't done, OR
	//   2. Another task has a blocks link pointing to it and isn't done.
	unmetDeps := make(map[string]bool)

	// Check depends-on links (outgoing from the task)
	for _, task := range tasks {
		for _, link := range task.Links {
			if link.Type == model.LinkDependsOn {
				depTask, depErr := a.Index.GetTask(link.Target)
				if depErr != nil || depTask.Status != "done" {
					unmetDeps[task.ID] = true
					break
				}
			}
		}
	}

	// Check blocks links (if A blocks B and A isn't done, B is blocked)
	for _, task := range tasks {
		for _, link := range task.Links {
			if link.Type == model.LinkBlocks {
				// This task blocks link.Target — if this task isn't done,
				// the target has an unmet dependency.
				if task.Status != model.StatusDone {
					unmetDeps[link.Target] = true
				}
			}
		}
	}

	energy := opts.Energy
	if energy == "" {
		energy = model.EnergyMedium
	}

	return engine.PackRequest{
		AvailableTime:     opts.AvailableTime,
		UserEnergy:        energy,
		Context:           opts.Context,
		Now:               now,
		Weights:           engine.DefaultWeights,
		Tasks:             tasks,
		HabitInfoMap:      habitInfoMap,
		BlockedByMap:      blockedByMap,
		UnmetDependencies: unmetDeps,
	}, nil
}

// --- Helpers ---

// loadTask reads the full task from disk given an indexed reference.
func (a *App) loadTask(indexed *store.IndexedTask) (*model.Task, string, error) {
	absPath := filepath.Join(a.DataDir, indexed.FilePath)
	task, err := store.ParseFile(absPath)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", indexed.FilePath, err)
	}
	return task, indexed.FilePath, nil
}

// saveTask writes a task to disk and updates the index.
func (a *App) saveTask(task *model.Task, relPath string) error {
	absPath := filepath.Join(a.DataDir, relPath)
	if err := store.WriteFile(absPath, task); err != nil {
		return fmt.Errorf("write task: %w", err)
	}
	if err := a.Index.UpsertTask(task, relPath); err != nil {
		return fmt.Errorf("index task: %w", err)
	}
	return nil
}

// taskFilePath determines the relative file path for a task based on its box.
func taskFilePath(task *model.Task) string {
	dir := "inbox"
	if task.Box != "" {
		dir = task.Box
	}
	return filepath.Join(dir, task.ID+".md")
}

// TaskOptions holds optional fields for creating a task.
type TaskOptions struct {
	Type     model.TaskType
	Priority model.Priority
	Energy   model.Energy
	Duration int
	DueDate  string
	DueStart string
	DueEnd   string
	Tags     []string
	Context  []string
	Box      string
}
