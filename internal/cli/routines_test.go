package cli

import (
	"testing"
)

func TestParseStepFlagsBasic(t *testing.T) {
	steps, err := parseStepFlags([]string{"Shower:5", "Breakfast:15", "Review inbox:10"})
	if err != nil {
		t.Fatalf("parseStepFlags() error: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("steps count = %d, want 3", len(steps))
	}
	if steps[0].Title != "Shower" || steps[0].Duration != 5 {
		t.Errorf("step[0] = %+v, want Shower:5", steps[0])
	}
	if steps[1].Title != "Breakfast" || steps[1].Duration != 15 {
		t.Errorf("step[1] = %+v, want Breakfast:15", steps[1])
	}
	if steps[2].Title != "Review inbox" || steps[2].Duration != 10 {
		t.Errorf("step[2] = %+v, want Review inbox:10", steps[2])
	}
}

func TestParseStepFlagsNoDuration(t *testing.T) {
	steps, err := parseStepFlags([]string{"Meditate", "Stretch"})
	if err != nil {
		t.Fatalf("parseStepFlags() error: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("steps count = %d, want 2", len(steps))
	}
	if steps[0].Title != "Meditate" || steps[0].Duration != 0 {
		t.Errorf("step[0] = %+v, want Meditate:0", steps[0])
	}
}

func TestParseStepFlagsOptional(t *testing.T) {
	steps, err := parseStepFlags([]string{"Required:5", "Optional:10?"})
	if err != nil {
		t.Fatalf("parseStepFlags() error: %v", err)
	}
	if steps[0].Optional {
		t.Error("step[0] should not be optional")
	}
	if !steps[1].Optional {
		t.Error("step[1] should be optional")
	}
	if steps[1].Title != "Optional" || steps[1].Duration != 10 {
		t.Errorf("step[1] = %+v, want Optional:10", steps[1])
	}
}

func TestParseStepFlagsEmpty(t *testing.T) {
	_, err := parseStepFlags([]string{""})
	if err == nil {
		t.Error("empty step should return error")
	}
}

func TestParseStepFlagsTitleWithColon(t *testing.T) {
	// Title contains a colon but no valid duration after it
	steps, err := parseStepFlags([]string{"Read: chapter 3"})
	if err != nil {
		t.Fatalf("parseStepFlags() error: %v", err)
	}
	if steps[0].Title != "Read: chapter 3" {
		t.Errorf("title = %q, want 'Read: chapter 3'", steps[0].Title)
	}
	if steps[0].Duration != 0 {
		t.Errorf("duration = %d, want 0 (colon not followed by number)", steps[0].Duration)
	}
}
