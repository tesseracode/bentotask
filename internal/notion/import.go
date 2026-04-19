package notion

import (
	"fmt"
	"strings"

	"github.com/tesserabox/bentotask/internal/app"
	"github.com/tesserabox/bentotask/internal/model"
)

// ImportResult holds the result of a database import.
type ImportResult struct {
	Imported int
	Skipped  int
	Errors   []string
}

// ImportDatabase imports all pages from a Notion database as BentoTask tasks.
func ImportDatabase(client *Client, databaseID string, a *app.App, dryRun bool) (*ImportResult, error) {
	resp, err := client.QueryDatabase(databaseID)
	if err != nil {
		return nil, fmt.Errorf("query database: %w", err)
	}

	return importFromPages(resp.Results, a, dryRun)
}

// importFromPages creates tasks from a list of Notion pages.
func importFromPages(pages []Page, a *app.App, dryRun bool) (*ImportResult, error) {
	result := &ImportResult{}

	for _, page := range pages {
		title := extractTitle(page)
		if title == "" {
			result.Skipped++
			continue
		}

		opts := mapProperties(page)

		if dryRun {
			result.Imported++
			continue
		}

		if _, err := a.AddTask(title, opts); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", title, err))
			continue
		}
		result.Imported++
	}

	return result, nil
}

// extractTitle finds the title property in a Notion page.
func extractTitle(page Page) string {
	for _, prop := range page.Properties {
		if prop.Type == "title" {
			return prop.PlainText()
		}
	}
	return ""
}

// mapProperties maps Notion page properties to BentoTask TaskOptions.
func mapProperties(page Page) app.TaskOptions {
	opts := app.TaskOptions{}

	for name, prop := range page.Properties {
		lname := strings.ToLower(name)

		switch {
		case prop.Type == "select" && (lname == "priority" || lname == "p"):
			opts.Priority = mapPriority(prop.SelectName())

		case prop.Type == "status" && (lname == "status" || lname == "state"):
			// Notion statuses vary; common ones: "Not started", "In progress", "Done"
			// We don't set status on creation — tasks default to pending

		case prop.Type == "date" && (lname == "date" || lname == "due" || lname == "due date" || lname == "deadline"):
			if d := prop.DateStart(); d != "" {
				opts.DueDate = d
			}

		case prop.Type == "multi_select" && (lname == "tags" || lname == "labels" || lname == "categories"):
			opts.Tags = prop.MultiSelectNames()

		case prop.Type == "select" && (lname == "energy" || lname == "effort"):
			opts.Energy = mapEnergy(prop.SelectName())

		case prop.Type == "number" && (lname == "duration" || lname == "estimate" || lname == "time"):
			if prop.Number != nil {
				opts.Duration = int(*prop.Number)
			}
		}
	}

	return opts
}

func mapPriority(s string) model.Priority {
	switch strings.ToLower(s) {
	case "urgent", "critical", "p0":
		return model.PriorityUrgent
	case "high", "p1":
		return model.PriorityHigh
	case "medium", "normal", "p2":
		return model.PriorityMedium
	case "low", "p3":
		return model.PriorityLow
	default:
		return ""
	}
}

func mapEnergy(s string) model.Energy {
	switch strings.ToLower(s) {
	case "high":
		return model.EnergyHigh
	case "medium", "normal":
		return model.EnergyMedium
	case "low":
		return model.EnergyLow
	default:
		return ""
	}
}
