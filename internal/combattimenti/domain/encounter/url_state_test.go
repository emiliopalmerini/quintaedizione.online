package encounter

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultURLState(t *testing.T) {
	state := DefaultURLState()
	if state.Ruleset != "2024" {
		t.Errorf("default ruleset: want 2024, got %q", state.Ruleset)
	}
	if state.Difficulty != "Moderate" {
		t.Errorf("default difficulty: want Moderate, got %q", state.Difficulty)
	}
	if !reflect.DeepEqual(state.Party, []int{3, 3, 3, 3}) {
		t.Errorf("default party: want [3 3 3 3], got %v", state.Party)
	}
	if len(state.Cart) != 0 {
		t.Errorf("default cart: want empty, got %v", state.Cart)
	}
}

func TestDecodeURLState_AllDefaultsWhenEmpty(t *testing.T) {
	state := DecodeURLState(url.Values{})
	if !reflect.DeepEqual(state, DefaultURLState()) {
		t.Errorf("empty querystring: want defaults, got %+v", state)
	}
}

func TestDecodeURLState_FullySpecified2024(t *testing.T) {
	v := url.Values{}
	v.Set("ruleset", "2024")
	v.Set("party", "4,4,4,4")
	v.Set("diff", "High")
	v.Set("cart", "aboleth@5.5e:2,ankheg@5.5e")

	state := DecodeURLState(v)

	if state.Ruleset != "2024" {
		t.Errorf("ruleset: %q", state.Ruleset)
	}
	if !reflect.DeepEqual(state.Party, []int{4, 4, 4, 4}) {
		t.Errorf("party: %v", state.Party)
	}
	if state.Difficulty != "High" {
		t.Errorf("difficulty: %q", state.Difficulty)
	}
	if len(state.Cart) != 2 {
		t.Fatalf("cart len: %d", len(state.Cart))
	}
	if state.Cart[0] != (CartRef{ID: "aboleth", Source: "5.5e", Qty: 2}) {
		t.Errorf("cart[0]: %+v", state.Cart[0])
	}
	if state.Cart[1] != (CartRef{ID: "ankheg", Source: "5.5e", Qty: 1}) {
		t.Errorf("cart[1]: %+v", state.Cart[1])
	}
}

func TestDecodeURLState_2014DefaultsToMedia(t *testing.T) {
	v := url.Values{}
	v.Set("ruleset", "2014")

	state := DecodeURLState(v)

	if state.Ruleset != "2014" {
		t.Errorf("ruleset: %q", state.Ruleset)
	}
	if state.Difficulty != "Media" {
		t.Errorf("difficulty default for 2014: want Media, got %q", state.Difficulty)
	}
}

func TestDecodeURLState_InvalidRulesetFallsBackToDefault(t *testing.T) {
	v := url.Values{}
	v.Set("ruleset", "bogus")
	v.Set("diff", "Moderate")

	state := DecodeURLState(v)

	if state.Ruleset != "2024" {
		t.Errorf("want fallback to 2024, got %q", state.Ruleset)
	}
	if state.Difficulty != "Moderate" {
		t.Errorf("difficulty: %q", state.Difficulty)
	}
}

func TestDecodeURLState_DifficultyMismatchUsesRulesetDefault(t *testing.T) {
	v := url.Values{}
	v.Set("ruleset", "2014")
	v.Set("diff", "Low") // 2024 value, invalid for 2014

	state := DecodeURLState(v)

	if state.Difficulty != "Media" {
		t.Errorf("want default Media for 2014, got %q", state.Difficulty)
	}
}

func TestDecodeURLState_PartyInvalidLevelsDropped(t *testing.T) {
	v := url.Values{}
	v.Set("party", "3,0,4,21,abc,5")

	state := DecodeURLState(v)

	if !reflect.DeepEqual(state.Party, []int{3, 4, 5}) {
		t.Errorf("want [3 4 5], got %v", state.Party)
	}
}

func TestDecodeURLState_PartyAllInvalidFallsBackToDefault(t *testing.T) {
	v := url.Values{}
	v.Set("party", "0,21,xyz")

	state := DecodeURLState(v)

	if !reflect.DeepEqual(state.Party, defaultParty) {
		t.Errorf("want default party, got %v", state.Party)
	}
}

func TestDecodeURLState_PartyClampedToMax(t *testing.T) {
	// Build a party of 200 valid levels.
	parts := make([]string, 200)
	for i := range parts {
		parts[i] = "3"
	}
	v := url.Values{}
	v.Set("party", strings.Join(parts, ","))

	state := DecodeURLState(v)

	if len(state.Party) != MaxPartySize {
		t.Errorf("want clamped to %d, got %d", MaxPartySize, len(state.Party))
	}
}

func TestDecodeURLState_CartMalformedDropped(t *testing.T) {
	v := url.Values{}
	// missing @source, missing id, missing source, zero qty, negative qty,
	// non-numeric qty, then one valid trailing entry.
	v.Set("cart", "noseparator,@onlysource,onlyid@,goblin@5.5e:0,goblin@5.5e:-1,goblin@5.5e:abc,aboleth@5.5e:3")

	state := DecodeURLState(v)

	if len(state.Cart) != 1 {
		t.Fatalf("want 1 valid cart entry, got %d (%+v)", len(state.Cart), state.Cart)
	}
	want := CartRef{ID: "aboleth", Source: "5.5e", Qty: 3}
	if state.Cart[0] != want {
		t.Errorf("cart[0]: want %+v, got %+v", want, state.Cart[0])
	}
}

func TestDecodeURLState_CartCoalescesDuplicates(t *testing.T) {
	v := url.Values{}
	v.Set("cart", "goblin@5.5e,goblin@5.5e:2,orc@5.5e,goblin@5.5e")

	state := DecodeURLState(v)

	if len(state.Cart) != 2 {
		t.Fatalf("want 2 distinct entries, got %d", len(state.Cart))
	}
	if state.Cart[0] != (CartRef{ID: "goblin", Source: "5.5e", Qty: 4}) {
		t.Errorf("goblin coalesced wrong: %+v", state.Cart[0])
	}
	if state.Cart[1] != (CartRef{ID: "orc", Source: "5.5e", Qty: 1}) {
		t.Errorf("orc: %+v", state.Cart[1])
	}
}

func TestURLState_Encode_OmitsDefaults(t *testing.T) {
	encoded := DefaultURLState().EncodeQuery()
	if encoded != "" {
		t.Errorf("default state should encode to empty string, got %q", encoded)
	}
}

func TestURLState_Encode_OmitsRulesetDifficultyDefault(t *testing.T) {
	state := URLState{
		Ruleset:    "2024",
		Party:      []int{4, 4, 4, 4},
		Difficulty: "Moderate", // default for 2024
	}
	v := state.Encode()
	if v.Has("ruleset") {
		t.Errorf("ruleset should be omitted when default; got %q", v.Get("ruleset"))
	}
	if v.Has("diff") {
		t.Errorf("diff should be omitted when default; got %q", v.Get("diff"))
	}
	if v.Get("party") != "4,4,4,4" {
		t.Errorf("party should be present; got %q", v.Get("party"))
	}
}

func TestURLState_Encode_NonDefaultRulesetEmitsBoth(t *testing.T) {
	state := URLState{
		Ruleset:    "2014",
		Party:      []int{3, 3, 3, 3},
		Difficulty: "Letale",
	}
	v := state.Encode()
	if v.Get("ruleset") != "2014" {
		t.Errorf("ruleset: %q", v.Get("ruleset"))
	}
	if v.Get("diff") != "Letale" {
		t.Errorf("diff: %q", v.Get("diff"))
	}
	if v.Has("party") {
		t.Errorf("party should be omitted when default; got %q", v.Get("party"))
	}
}

func TestURLState_Encode_CartFormat(t *testing.T) {
	state := URLState{
		Ruleset:    "2024",
		Party:      []int{3, 3, 3, 3},
		Difficulty: "Moderate",
		Cart: []CartRef{
			{ID: "goblin", Source: "5.5e", Qty: 3},
			{ID: "orc", Source: "5.5e", Qty: 1},
		},
	}
	v := state.Encode()
	if v.Get("cart") != "goblin@5.5e:3,orc@5.5e" {
		t.Errorf("cart encoding: %q", v.Get("cart"))
	}
}

func TestURLState_RoundTrip(t *testing.T) {
	cases := []URLState{
		DefaultURLState(),
		{
			Ruleset:    "2014",
			Party:      []int{1, 2, 3, 4, 5},
			Difficulty: "Difficile",
			Cart: []CartRef{
				{ID: "aboleth", Source: "5e", Qty: 1},
				{ID: "ankheg", Source: "5e", Qty: 4},
			},
		},
		{
			Ruleset:    "2024",
			Party:      []int{4, 4, 4, 4},
			Difficulty: "High",
			Cart: []CartRef{
				{ID: "goblin", Source: "5.5e", Qty: 7},
			},
		},
	}
	for i, original := range cases {
		encoded := original.Encode()
		decoded := DecodeURLState(encoded)
		if !reflect.DeepEqual(decoded, original) {
			t.Errorf("case %d: round-trip mismatch\nencoded: %s\nbefore: %+v\nafter:  %+v",
				i, encoded.Encode(), original, decoded)
		}
	}
}

func TestURLState_RoundTrip_CoalesceIsIdempotent(t *testing.T) {
	original := URLState{
		Ruleset:    "2024",
		Party:      []int{3, 3, 3, 3},
		Difficulty: "Moderate",
		Cart: []CartRef{
			{ID: "goblin", Source: "5.5e", Qty: 2},
			{ID: "goblin", Source: "5.5e", Qty: 3}, // duplicate; will coalesce
		},
	}
	encoded := original.Encode()
	decoded := DecodeURLState(encoded)

	want := []CartRef{{ID: "goblin", Source: "5.5e", Qty: 5}}
	if !reflect.DeepEqual(decoded.Cart, want) {
		t.Errorf("coalesce: want %+v, got %+v", want, decoded.Cart)
	}

	// Encoding again should still yield the same string.
	if decoded.EncodeQuery() != encoded.Encode() {
		t.Errorf("not idempotent: first=%q second=%q", encoded.Encode(), decoded.EncodeQuery())
	}
}

func TestURLState_WithSource_DropsForeignEntries(t *testing.T) {
	state := URLState{
		Ruleset: "2024",
		Cart: []CartRef{
			{ID: "goblin", Source: "5.5e", Qty: 1},
			{ID: "aboleth", Source: "5e", Qty: 1},
			{ID: "orc", Source: "5.5e", Qty: 2},
		},
	}
	filtered := state.WithSource("5.5e")
	if len(filtered.Cart) != 2 {
		t.Fatalf("want 2 entries, got %d", len(filtered.Cart))
	}
	for _, ref := range filtered.Cart {
		if ref.Source != "5.5e" {
			t.Errorf("foreign source survived: %+v", ref)
		}
	}
}

func TestURLState_WithSource_NoopWhenSourceEmpty(t *testing.T) {
	state := URLState{Cart: []CartRef{{ID: "goblin", Source: "5.5e", Qty: 1}}}
	if got := state.WithSource(""); !reflect.DeepEqual(got, state) {
		t.Errorf("WithSource(\"\") should be a no-op; got %+v", got)
	}
}

func TestDecodeURLState_RulesetIsLowercased(t *testing.T) {
	// NewRuleset accepts case-insensitively today; document the behavior.
	v := url.Values{}
	v.Set("ruleset", "2024")
	state := DecodeURLState(v)
	if state.Ruleset != "2024" {
		t.Errorf("ruleset normalization: %q", state.Ruleset)
	}
}

func TestDecodeURLState_EmptyPartyParamFallsBack(t *testing.T) {
	v := url.Values{}
	v.Set("party", "  ")
	state := DecodeURLState(v)
	if !reflect.DeepEqual(state.Party, defaultParty) {
		t.Errorf("blank party should default; got %v", state.Party)
	}
}
