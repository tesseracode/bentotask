package notion

// DatabaseQueryResponse is the response from POST /databases/{id}/query.
type DatabaseQueryResponse struct {
	Results    []Page `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

// Page represents a Notion page (row in a database).
type Page struct {
	ID         string              `json:"id"`
	Properties map[string]Property `json:"properties"`
}

// Property represents a Notion property value.
type Property struct {
	Type        string        `json:"type"`
	Title       []RichText    `json:"title,omitempty"`
	RichText    []RichText    `json:"rich_text,omitempty"`
	Select      *SelectValue  `json:"select,omitempty"`
	MultiSelect []SelectValue `json:"multi_select,omitempty"`
	Date        *DateValue    `json:"date,omitempty"`
	Status      *StatusValue  `json:"status,omitempty"`
	Number      *float64      `json:"number,omitempty"`
}

// RichText represents a Notion rich text segment.
type RichText struct {
	PlainText string `json:"plain_text"`
}

// SelectValue represents a Notion select option.
type SelectValue struct {
	Name string `json:"name"`
}

// DateValue represents a Notion date property.
type DateValue struct {
	Start string `json:"start"`
	End   string `json:"end,omitempty"`
}

// StatusValue represents a Notion status property.
type StatusValue struct {
	Name string `json:"name"`
}

// PlainText extracts plain text from a title or rich_text property.
func (p *Property) PlainText() string {
	texts := p.Title
	if len(texts) == 0 {
		texts = p.RichText
	}
	if len(texts) == 0 {
		return ""
	}
	var sb []string
	for _, t := range texts {
		sb = append(sb, t.PlainText)
	}
	result := ""
	for _, s := range sb {
		result += s
	}
	return result
}

// SelectName returns the select option name, or empty string.
func (p *Property) SelectName() string {
	if p.Select != nil {
		return p.Select.Name
	}
	if p.Status != nil {
		return p.Status.Name
	}
	return ""
}

// MultiSelectNames returns all multi-select option names.
func (p *Property) MultiSelectNames() []string {
	names := make([]string, len(p.MultiSelect))
	for i, s := range p.MultiSelect {
		names[i] = s.Name
	}
	return names
}

// DateStart returns the date start string (YYYY-MM-DD or datetime).
func (p *Property) DateStart() string {
	if p.Date == nil {
		return ""
	}
	// Notion dates can be "2026-04-10" or "2026-04-10T10:00:00.000+00:00"
	// We only need the date portion
	s := p.Date.Start
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}
