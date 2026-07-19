// Package encounter — URL state encoding/decoding.
//
// URLState is the share-link payload for the encounter calculator. It maps to
// /combattimenti's querystring so a user can bookmark, share, or refresh the
// page without losing party / ruleset / difficulty / cart selections.
//
// The querystring scheme:
//
//	ruleset=2024|2014                 (default: "2024")
//	party=3,3,3,3                     (comma-separated levels 1..20, default 4× level 3)
//	diff=Low|Moderate|High|Facile|Media|Difficile|Letale
//	                                  (default: Moderate for 2024, Media for 2014)
//	cart=id@source[:qty][,id@source[:qty]...]
//	                                  (qty omitted when 1)
//
// All params are optional. Malformed cart entries are dropped silently,
// invalid party levels are dropped, party is clamped to MaxPartySize. The
// decoder never errors — it always returns a usable state.
package encounter

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// MaxPartySize caps the party at the same upper bound the UI enforces.
const MaxPartySize = 8

// Default values applied when the querystring is empty.
var (
	defaultParty      = []int{3, 3, 3, 3}
	defaultRuleset    = string(Ruleset2024)
	defaultDifficulty = string(DifficultyModerate)
)

// CartRef is a single cart line item in a URLState. Qty is at least 1.
type CartRef struct {
	ID     string
	Source string
	Qty    int
}

// URLState is the full snapshot of an encounter screen, suitable for
// round-tripping through a URL.
type URLState struct {
	Ruleset    string
	Party      []int
	Difficulty string
	Cart       []CartRef
}

// DefaultURLState returns the state shown when the page is loaded without
// any query parameters. Mirrors the form's hard-coded defaults so a bare
// /combattimenti URL stays equivalent to a fully-defaulted one.
func DefaultURLState() URLState {
	party := make([]int, len(defaultParty))
	copy(party, defaultParty)
	return URLState{
		Ruleset:    defaultRuleset,
		Party:      party,
		Difficulty: defaultDifficulty,
		Cart:       nil,
	}
}

// DecodeURLState reads a querystring into a URLState, applying defaults for
// missing values and dropping malformed pieces. It never returns an error;
// the worst-case is the all-defaults state.
//
// Normalization rules:
//   - Ruleset is lowercased; unknown values fall back to the default.
//   - Difficulty is kept verbatim if valid for the (possibly defaulted)
//     ruleset; otherwise the ruleset's default difficulty is used.
//   - Party levels outside [1, 20] are dropped; the result is clamped to
//     MaxPartySize entries. Empty result falls back to the default party.
//   - Cart entries missing id/source are dropped. Qty defaults to 1 and is
//     clamped to [1, 999]; entries with explicit qty<=0 are dropped.
func DecodeURLState(values url.Values) URLState {
	state := DefaultURLState()

	if raw := strings.TrimSpace(values.Get("ruleset")); raw != "" {
		if rs, err := NewRuleset(raw); err == nil {
			state.Ruleset = string(rs)
		}
	}

	if raw := strings.TrimSpace(values.Get("party")); raw != "" {
		if party := parseParty(raw); len(party) > 0 {
			state.Party = party
		}
	}

	// Difficulty is parsed after ruleset because it must validate against it.
	state.Difficulty = defaultDifficultyFor(state.Ruleset)
	if raw := strings.TrimSpace(values.Get("diff")); raw != "" {
		if _, err := NewDifficulty(raw, Ruleset(state.Ruleset)); err == nil {
			state.Difficulty = raw
		}
	}

	if raw := strings.TrimSpace(values.Get("cart")); raw != "" {
		state.Cart = parseCart(raw)
	}

	return state
}

// Encode serializes the state to a url.Values. Defaults are omitted so a
// bookmarked link with only non-default values stays short.
//
// Encoding is deterministic: cart entries keep input order, but identical
// (id,source) pairs collapse to a single entry whose Qty is the sum.
func (s URLState) Encode() url.Values {
	out := url.Values{}

	if s.Ruleset != "" && s.Ruleset != defaultRuleset {
		out.Set("ruleset", s.Ruleset)
	}

	if !partyMatchesDefault(s.Party) {
		out.Set("party", joinInts(s.Party))
	}

	// Only emit diff when it deviates from the ruleset default.
	if s.Difficulty != "" && s.Difficulty != defaultDifficultyFor(s.Ruleset) {
		out.Set("diff", s.Difficulty)
	}

	if cart := encodeCart(s.Cart); cart != "" {
		out.Set("cart", cart)
	}

	return out
}

// EncodeQuery is a convenience that returns the encoded values as a raw
// querystring (without leading "?"). Empty when the state is fully default.
func (s URLState) EncodeQuery() string {
	return s.Encode().Encode()
}

// WithSource drops cart entries whose source does not match the given short
// name. Used by the web layer to clear the cart when the user switches
// ruleset (per ADR-024: monsters from different editions can't share a
// budget).
func (s URLState) WithSource(source string) URLState {
	if source == "" || len(s.Cart) == 0 {
		return s
	}
	filtered := make([]CartRef, 0, len(s.Cart))
	for _, ref := range s.Cart {
		if ref.Source == source {
			filtered = append(filtered, ref)
		}
	}
	out := s
	out.Cart = filtered
	return out
}

// --- Helpers ----------------------------------------------------------------

func defaultDifficultyFor(ruleset string) string {
	switch Ruleset(ruleset) {
	case Ruleset2014:
		return string(DifficultyMedium)
	default:
		return string(DifficultyModerate)
	}
}

func partyMatchesDefault(party []int) bool {
	if len(party) != len(defaultParty) {
		return false
	}
	for i, v := range party {
		if v != defaultParty[i] {
			return false
		}
	}
	return true
}

func parseParty(raw string) []int {
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		if n < 1 || n > 20 {
			continue
		}
		out = append(out, n)
		if len(out) >= MaxPartySize {
			break
		}
	}
	return out
}

func parseCart(raw string) []CartRef {
	parts := strings.Split(raw, ",")
	// Preserve first-seen order while accumulating qty across duplicates.
	indexByKey := make(map[string]int, len(parts))
	out := make([]CartRef, 0, len(parts))
	for _, p := range parts {
		ref, ok := parseCartEntry(p)
		if !ok {
			continue
		}
		key := ref.ID + "@" + ref.Source
		if idx, exists := indexByKey[key]; exists {
			out[idx].Qty += ref.Qty
			continue
		}
		indexByKey[key] = len(out)
		out = append(out, ref)
	}
	return out
}

// parseCartEntry handles "id@source" and "id@source:qty". Returns ok=false
// for empty, missing-source, missing-id, or non-positive-qty entries.
func parseCartEntry(raw string) (CartRef, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return CartRef{}, false
	}

	idSource, qtyStr, hasQty := strings.Cut(raw, ":")
	id, src, hasAt := strings.Cut(idSource, "@")
	id = strings.TrimSpace(id)
	src = strings.TrimSpace(src)
	if !hasAt || id == "" || src == "" {
		return CartRef{}, false
	}

	qty := 1
	if hasQty {
		n, err := strconv.Atoi(strings.TrimSpace(qtyStr))
		if err != nil || n <= 0 {
			return CartRef{}, false
		}
		if n > 999 {
			n = 999
		}
		qty = n
	}
	return CartRef{ID: id, Source: src, Qty: qty}, true
}

func encodeCart(cart []CartRef) string {
	if len(cart) == 0 {
		return ""
	}
	// Coalesce duplicates the same way DecodeURLState does, so Encode→Decode
	// is a fixed point.
	type slot struct {
		ref CartRef
		pos int
	}
	merged := make(map[string]*slot, len(cart))
	order := make([]string, 0, len(cart))
	for _, ref := range cart {
		if ref.ID == "" || ref.Source == "" {
			continue
		}
		qty := ref.Qty
		if qty <= 0 {
			qty = 1
		}
		if qty > 999 {
			qty = 999
		}
		key := ref.ID + "@" + ref.Source
		if s, ok := merged[key]; ok {
			s.ref.Qty += qty
			continue
		}
		merged[key] = &slot{ref: CartRef{ID: ref.ID, Source: ref.Source, Qty: qty}, pos: len(order)}
		order = append(order, key)
	}
	if len(order) == 0 {
		return ""
	}
	// Stable: insertion order
	sort.SliceStable(order, func(i, j int) bool {
		return merged[order[i]].pos < merged[order[j]].pos
	})
	parts := make([]string, 0, len(order))
	for _, key := range order {
		ref := merged[key].ref
		if ref.Qty == 1 {
			parts = append(parts, fmt.Sprintf("%s@%s", ref.ID, ref.Source))
		} else {
			parts = append(parts, fmt.Sprintf("%s@%s:%d", ref.ID, ref.Source, ref.Qty))
		}
	}
	return strings.Join(parts, ",")
}

func joinInts(ints []int) string {
	parts := make([]string, len(ints))
	for i, n := range ints {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ",")
}
