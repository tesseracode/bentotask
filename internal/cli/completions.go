package cli

import (
	"github.com/spf13/cobra"
)

// registerCompletions sets up dynamic shell completions for task IDs,
// tags, boxes, contexts, and enum flags. Called during init().
//
// Per ADR-003 §6, completions show "ID\tTitle" for task arguments,
// and valid values for enum flags like --status, --priority, --energy.
func registerCompletions() {
	// --- Task ID completions ---
	// Commands that take a task ID as a positional argument
	taskIDCommands := []*cobra.Command{
		taskDoneCmd, doneCmd,
		taskShowCmd,
		taskDeleteCmd,
		taskEditCmd,
	}
	for _, cmd := range taskIDCommands {
		cmd.ValidArgsFunction = completeTaskIDs
	}

	// --- Flag completions ---
	// Register for both paired commands (task add + add, task list + list, etc.)
	for _, cmd := range []*cobra.Command{taskAddCmd, addCmd} {
		registerAddCompletions(cmd)
	}
	for _, cmd := range []*cobra.Command{taskListCmd, listCmd} {
		registerListCompletions(cmd)
	}
	registerEditCompletions(taskEditCmd)
}

// completeTaskIDs provides dynamic completion for task ID arguments.
// Shows non-done tasks as "ID<TAB>Title" for a rich completion experience.
func completeTaskIDs(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	a, err := openApp(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer func() { _ = a.Close() }()

	comps, err := a.CompleteTasks()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

// completeTags provides dynamic completion for --tag flag values.
func completeTags(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	a, err := openApp(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer func() { _ = a.Close() }()

	tags, err := a.CompleteTags()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return tags, cobra.ShellCompDirectiveNoFileComp
}

// completeBoxes provides dynamic completion for --box flag values.
func completeBoxes(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	a, err := openApp(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer func() { _ = a.Close() }()

	boxes, err := a.CompleteBoxes()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return boxes, cobra.ShellCompDirectiveNoFileComp
}

// registerAddCompletions sets up flag completions for add commands.
func registerAddCompletions(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc("priority", completePriority)
	_ = cmd.RegisterFlagCompletionFunc("energy", completeEnergy)
	_ = cmd.RegisterFlagCompletionFunc("tag", completeTagFlag)
	_ = cmd.RegisterFlagCompletionFunc("context", completeContextFixed)
	_ = cmd.RegisterFlagCompletionFunc("box", completeBoxFlag)
}

// registerListCompletions sets up flag completions for list commands.
func registerListCompletions(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc("status", completeStatus)
	_ = cmd.RegisterFlagCompletionFunc("priority", completePriority)
	_ = cmd.RegisterFlagCompletionFunc("energy", completeEnergy)
	_ = cmd.RegisterFlagCompletionFunc("tag", completeTagFlag)
	_ = cmd.RegisterFlagCompletionFunc("context", completeContextFixed)
	_ = cmd.RegisterFlagCompletionFunc("box", completeBoxFlag)
}

// registerEditCompletions sets up flag completions for edit command.
func registerEditCompletions(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc("status", completeStatus)
	_ = cmd.RegisterFlagCompletionFunc("priority", completePriority)
	_ = cmd.RegisterFlagCompletionFunc("energy", completeEnergy)
	_ = cmd.RegisterFlagCompletionFunc("tag", completeTagFlag)
	_ = cmd.RegisterFlagCompletionFunc("context", completeContextFixed)
	_ = cmd.RegisterFlagCompletionFunc("box", completeBoxFlag)
}

// --- Static enum completions ---

// completeStatus returns valid status values.
func completeStatus(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"pending\tNot yet started",
		"active\tCurrently in progress",
		"paused\tTemporarily paused",
		"done\tCompleted",
		"cancelled\tCancelled",
		"waiting\tWaiting on something",
	}, cobra.ShellCompDirectiveNoFileComp
}

// completePriority returns valid priority values.
func completePriority(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"none\tNo priority",
		"low\tLow priority",
		"medium\tMedium priority",
		"high\tHigh priority",
		"urgent\tUrgent priority",
	}, cobra.ShellCompDirectiveNoFileComp
}

// completeEnergy returns valid energy values.
func completeEnergy(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"low\tLow energy task",
		"medium\tMedium energy task",
		"high\tHigh energy task",
	}, cobra.ShellCompDirectiveNoFileComp
}

// completeContextFixed returns the fixed set of context values.
func completeContextFixed(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"home\tAt home",
		"office\tAt the office",
		"errands\tOut running errands",
		"anywhere\tCan be done anywhere",
	}, cobra.ShellCompDirectiveNoFileComp
}

// --- Dynamic flag completions (wrappers) ---

// completeTagFlag wraps completeTags for flag completion signature.
func completeTagFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeTags(cmd, args, toComplete)
}

// completeBoxFlag wraps completeBoxes for flag completion signature.
func completeBoxFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeBoxes(cmd, args, toComplete)
}
