package monster

import (
	"math"
	"testing"
)

func TestParseCR_AcceptsFractionalRatings(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"0", 0},
		{"1/8", 0.125},
		{"1/4", 0.25},
		{"1/2", 0.5},
		{"1", 1},
		{"5", 5},
		{"30", 30},
	}
	for _, tc := range cases {
		got, err := ParseCR(tc.in)
		if err != nil {
			t.Errorf("ParseCR(%q) err = %v, want nil", tc.in, err)
			continue
		}
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("ParseCR(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseCR_StripsDetailSuffix(t *testing.T) {
	// Real SRD strings keep the XP detail in the same field; the parser must
	// ignore everything past the first whitespace.
	cases := []struct {
		in   string
		want float64
	}{
		{"10 (PE 5.900; BC +4)", 10},
		{"1/4 (PE 50; BC +2)", 0.25},
		{"21 (PE 33.000; BC +7)", 21},
		{"0 (10 PE)", 0},
	}
	for _, tc := range cases {
		got, err := ParseCR(tc.in)
		if err != nil {
			t.Errorf("ParseCR(%q) err = %v", tc.in, err)
			continue
		}
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("ParseCR(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseCR_RejectsInvalidInput(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"abc",
		"1/abc",
		"abc/2",
		"1/0",
		"//",
	}
	for _, in := range cases {
		if _, err := ParseCR(in); err == nil {
			t.Errorf("ParseCR(%q) err = nil, want error", in)
		}
	}
}

func TestParseCR_TrimsWhitespace(t *testing.T) {
	got, err := ParseCR("  1/2  ")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != 0.5 {
		t.Errorf("got %v, want 0.5", got)
	}
}
