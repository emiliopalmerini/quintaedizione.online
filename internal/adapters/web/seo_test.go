package web

import "testing"

func TestTruncateDescription(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		maxLen int
		want   string
	}{
		{
			name:   "empty input",
			raw:    "",
			maxLen: 160,
			want:   "",
		},
		{
			name:   "short input unchanged",
			raw:    "A simple spell description.",
			maxLen: 160,
			want:   "A simple spell description.",
		},
		{
			name:   "long input truncated at word boundary",
			raw:    "This is a very long description that exceeds the maximum length limit and should be truncated at a word boundary to avoid cutting words in half",
			maxLen: 80,
			want:   "This is a very long description that exceeds the maximum length limit and...",
		},
		{
			name:   "markdown bold stripped",
			raw:    "**Palla di Fuoco** è un incantesimo potente.",
			maxLen: 160,
			want:   "Palla di Fuoco è un incantesimo potente.",
		},
		{
			name:   "markdown italic stripped",
			raw:    "*Evocazione di livello 3*",
			maxLen: 160,
			want:   "Evocazione di livello 3",
		},
		{
			name:   "multi-line uses first paragraph only",
			raw:    "First paragraph here.\n\nSecond paragraph ignored.",
			maxLen: 160,
			want:   "First paragraph here.",
		},
		{
			name:   "markdown links stripped",
			raw:    "Vedi [regole](/regole) per dettagli.",
			maxLen: 160,
			want:   "Vedi regole per dettagli.",
		},
		{
			name:   "markdown headings stripped",
			raw:    "## Titolo\n\nContenuto della sezione.",
			maxLen: 160,
			want:   "Contenuto della sezione.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateDescription(tt.raw, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateDescription(%q, %d) = %q, want %q", tt.raw, tt.maxLen, got, tt.want)
			}
		})
	}
}
