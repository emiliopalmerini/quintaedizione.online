package domain

// Column represents a single rollable column within a table.
type Column struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

// Table represents a random generator table.
// A table has either a flat Items list (single column) or multiple Columns.
type Table struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Die         string   `json:"die"`
	Order       int      `json:"order"`
	Items       []string `json:"items,omitempty"`
	Columns     []Column `json:"columns,omitempty"`
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
}
