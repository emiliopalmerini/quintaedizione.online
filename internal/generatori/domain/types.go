package domain

import "encoding/json"

// Item represents a single entry in a generator table.
// JSON supports both plain strings and {"text":"...", "link":"..."} objects.
type Item struct {
	Text string `json:"text"`
	Link string `json:"link,omitempty"`
}

func (i *Item) UnmarshalJSON(data []byte) error {
	// Try plain string first.
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		i.Text = s
		return nil
	}
	// Otherwise parse as object.
	type alias Item
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*i = Item(a)
	return nil
}

func (i Item) MarshalJSON() ([]byte, error) {
	if i.Link == "" {
		return json.Marshal(i.Text)
	}
	type alias Item
	return json.Marshal(alias(i))
}

// Column represents a single rollable column within a table.
type Column struct {
	Name  string `json:"name"`
	Items []Item `json:"items"`
}

// Source represents the original author and URL of a generator table.
type Source struct {
	Author string `json:"author"`
	URL    string `json:"url"`
}

// Table represents a random generator table.
// A table has either a flat Items list (single column) or multiple Columns.
type Table struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Die         string   `json:"die"`
	Order       int      `json:"order"`
	Group       string   `json:"group"`
	Source      Source   `json:"source"`
	Items       []Item   `json:"items,omitempty"`
	Columns     []Column `json:"columns,omitempty"`
}

// Group represents a collection of related generator tables.
type Group struct {
	ID          string
	Label       string
	Description string
	Tables      []Table
}

// IsMultiColumn returns true if the table has multiple independent columns.
func (t Table) IsMultiColumn() bool {
	return len(t.Columns) > 0
}

// RollResult holds the outcome of rolling on a table.
type RollResult struct {
	Entries []RollEntry
}

// RollEntry is one column name + rolled value.
type RollEntry struct {
	Column string
	Value  string
	Link   string
}
