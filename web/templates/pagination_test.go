package templates

import (
	"reflect"
	"testing"
)

func TestPaginationPages(t *testing.T) {
	tests := []struct {
		name      string
		current   int
		total     int
		wantPages []int
		wantEllip bool // at least one ellipsis expected
	}{
		{
			name:      "small total shows all pages",
			current:   1,
			total:     5,
			wantPages: []int{1, 2, 3, 4, 5},
		},
		{
			name:      "beginning of large range",
			current:   1,
			total:     33,
			wantPages: []int{1, 2, 3, -1, 33}, // -1 = ellipsis
			wantEllip: true,
		},
		{
			name:      "middle of large range",
			current:   15,
			total:     33,
			wantPages: []int{1, -1, 13, 14, 15, 16, 17, -1, 33},
			wantEllip: true,
		},
		{
			name:      "end of large range",
			current:   33,
			total:     33,
			wantPages: []int{1, -1, 31, 32, 33},
			wantEllip: true,
		},
		{
			name:      "near beginning",
			current:   3,
			total:     33,
			wantPages: []int{1, 2, 3, 4, 5, -1, 33},
			wantEllip: true,
		},
		{
			name:      "near end",
			current:   31,
			total:     33,
			wantPages: []int{1, -1, 29, 30, 31, 32, 33},
			wantEllip: true,
		},
		{
			name:      "single page",
			current:   1,
			total:     1,
			wantPages: []int{1},
		},
		{
			name:      "two pages",
			current:   1,
			total:     2,
			wantPages: []int{1, 2},
		},
		{
			name:      "seven pages no ellipsis needed",
			current:   4,
			total:     7,
			wantPages: []int{1, 2, 3, 4, 5, 6, 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := paginationPages(tt.current, tt.total)
			if !reflect.DeepEqual(got, tt.wantPages) {
				t.Errorf("paginationPages(%d, %d) = %v, want %v", tt.current, tt.total, got, tt.wantPages)
			}
		})
	}
}
