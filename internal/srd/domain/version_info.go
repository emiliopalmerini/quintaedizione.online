package domain

// VersionInfo describes one available edition of a document.
type VersionInfo struct {
	SourceShort string // e.g. "5.5e"
	CompositeID string // e.g. "5.5e/palla-di-fuoco"
}
