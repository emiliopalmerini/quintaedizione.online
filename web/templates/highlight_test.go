package templates

import "testing"

func TestHighlightQuery(t *testing.T) {
	tests := []struct {
		name  string
		title string
		query string
		want  string
	}{
		{
			name:  "no match returns escaped title",
			title: "Palla di Fuoco",
			query: "xyz",
			want:  "Palla di Fuoco",
		},
		{
			name:  "exact match highlighted",
			title: "Palla di Fuoco",
			query: "Palla",
			want:  "<mark>Palla</mark> di Fuoco",
		},
		{
			name:  "case insensitive match preserves original case",
			title: "Palla di Fuoco",
			query: "palla",
			want:  "<mark>Palla</mark> di Fuoco",
		},
		{
			name:  "HTML special chars in title escaped",
			title: "<script>alert('xss')</script>",
			query: "script",
			want:  "&lt;<mark>script</mark>&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:  "empty query returns escaped title",
			title: "Palla di Fuoco",
			query: "",
			want:  "Palla di Fuoco",
		},
		{
			name:  "only first occurrence highlighted",
			title: "Fuoco di Fuoco",
			query: "Fuoco",
			want:  "<mark>Fuoco</mark> di Fuoco",
		},
		{
			name:  "empty title returns empty",
			title: "",
			query: "test",
			want:  "",
		},
		{
			name:  "match at end of string",
			title: "Dardo Incantato",
			query: "Incantato",
			want:  "Dardo <mark>Incantato</mark>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HighlightQuery(tt.title, tt.query)
			if got != tt.want {
				t.Errorf("HighlightQuery(%q, %q) =\n  %q\nwant:\n  %q", tt.title, tt.query, got, tt.want)
			}
		})
	}
}
