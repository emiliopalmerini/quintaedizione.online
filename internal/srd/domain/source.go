package domain

// Source represents a loaded edition/dataset with its metadata.
type Source struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	Year      int    `json:"year"`
	Ruleset   string `json:"ruleset"`
	XPSystem  string `json:"xp_system"`
	Default   bool   `json:"default"`
}
