package srd

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/repositories"
)

const monsterCollection = "mostri"

// xpPatterns match the two XP formats used across editions:
//   - 5.5e: "10 (PE 5.900; BC +4)" → "PE 5.900"
//   - 5e:   "1/4 (50 PE)"          → "50 PE"
//
// Italian uses dots as thousands separators ("5.900" == 5900).
var xpPatterns = []*regexp.Regexp{
	regexp.MustCompile(`PE\s+([\d.]+)`),
	regexp.MustCompile(`([\d.]+)\s+PE`),
}

// MonsterReader adapts the SRD DocumentRepository to the combattimenti monster.Reader port.
type MonsterReader struct {
	docs repositories.DocumentReader
}

// NewMonsterReader wires a combattimenti monster.Reader on top of the SRD document repo.
func NewMonsterReader(docs repositories.DocumentReader) *MonsterReader {
	return &MonsterReader{docs: docs}
}

// Search returns monsters matching the query, filtered by source edition,
// max XP, CR range, and type.
func (r *MonsterReader) Search(ctx context.Context, q monster.SearchQuery) ([]monster.Monster, error) {
	if q.Source == "" {
		return nil, nil
	}
	needle := strings.ToLower(strings.TrimSpace(q.Query))
	typeFilter := strings.ToLower(strings.TrimSpace(q.Type))
	limit := q.Limit
	if limit <= 0 {
		limit = 100
	}

	predicate := func(d map[string]any) bool {
		if src, _ := d["_source_short"].(string); src != q.Source {
			return false
		}
		if needle != "" {
			title, _ := d["title"].(string)
			if !strings.Contains(strings.ToLower(title), needle) {
				return false
			}
		}
		if typeFilter != "" {
			t, _ := d["tipo"].(string)
			if !strings.EqualFold(strings.TrimSpace(t), typeFilter) {
				return false
			}
		}
		return true
	}

	// -1 limit to pull everything, then price-filter and trim. Monster collections
	// are small (~330 docs), so post-filtering is cheap and avoids pushing XP
	// parsing into the store.
	rawDocs, _, err := r.docs.FindByPredicate(ctx, monsterCollection, predicate, 0, -1)
	if err != nil {
		return nil, err
	}

	result := make([]monster.Monster, 0, len(rawDocs))
	for _, doc := range rawDocs {
		m := fromDocument(doc)
		if q.MinCR > 0 || q.MaxCR > 0 {
			// Parse the bare CR (grado_sfida) when present; fall back to
			// the detail string. Skip the doc on parse failure so junk data
			// does not bypass the filter.
			cr, parseErr := monster.ParseCR(m.CR)
			if parseErr != nil {
				continue
			}
			if q.MinCR > 0 && cr < q.MinCR {
				continue
			}
			if q.MaxCR > 0 && cr > q.MaxCR {
				continue
			}
		}
		result = append(result, m)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].XP != result[j].XP {
			return result[i].XP < result[j].XP
		}
		return result[i].Name < result[j].Name
	})

	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// FindByID loads a single monster by its SRD id, scoped to the given source.
// Store keys are composite ("{source}/{id}"); the combattimenti cart uses
// (source, bare id) pairs so we build the composite here.
func (r *MonsterReader) FindByID(ctx context.Context, source, id string) (monster.Monster, error) {
	if source == "" || id == "" {
		return monster.Monster{}, errors.New("monster: source and id required")
	}

	doc, err := r.docs.FindByID(ctx, monsterCollection, source+"/"+id)
	if err != nil {
		return monster.Monster{}, err
	}
	if doc == nil || doc.Source != source {
		return monster.Monster{}, errors.New("monster: not found")
	}
	return fromDocument(doc), nil
}

// Facets enumerates the distinct creature types found in the given source's
// monster collection. Used to populate the picker's type dropdown.
//
// Result is sorted alphabetically. Empty/whitespace types are skipped.
func (r *MonsterReader) Facets(ctx context.Context, source string) (monster.FacetSet, error) {
	if source == "" {
		return monster.FacetSet{}, nil
	}
	predicate := func(d map[string]any) bool {
		src, _ := d["_source_short"].(string)
		return src == source
	}
	rawDocs, _, err := r.docs.FindByPredicate(ctx, monsterCollection, predicate, 0, -1)
	if err != nil {
		return monster.FacetSet{}, err
	}
	seen := make(map[string]struct{}, 16)
	for _, doc := range rawDocs {
		t := strings.TrimSpace(doc.GetFieldString("tipo"))
		if t == "" {
			continue
		}
		seen[t] = struct{}{}
	}
	types := make([]string, 0, len(seen))
	for t := range seen {
		types = append(types, t)
	}
	sort.Strings(types)
	return monster.FacetSet{Types: types}, nil
}

func fromDocument(doc *domain.Document) monster.Monster {
	// cr holds the detail string used for XP parsing (format varies by edition);
	// grado_sfida holds the bare challenge rating ("10", "1/4") for display.
	// 5e cr often lacks the XP marker, so we also scan the raw markdown which
	// consistently contains "(N PE)" across both editions.
	display := doc.GetFieldString("grado_sfida")
	if display == "" {
		display = doc.GetFieldString("cr")
	}
	xp := parseXP(doc.GetFieldString("cr"))
	if xp == 0 {
		xp = parseXP(doc.RawContent.String())
	}
	return monster.Monster{
		ID:     string(doc.ID),
		Source: doc.Source,
		Name:   doc.Title,
		Type:   doc.GetFieldString("tipo"),
		Size:   doc.GetFieldString("size"),
		CR:     display,
		XP:     xp,
		HP:     doc.GetFieldString("hp"),
		AC:     doc.GetFieldString("ac"),
		Speed:  doc.GetFieldString("speed"),
	}
}

// parseXP pulls the XP integer from a string, trying both edition formats.
// Returns 0 when no "PE N" / "N PE" token is present.
func parseXP(text string) int {
	for _, re := range xpPatterns {
		matches := re.FindStringSubmatch(text)
		if len(matches) >= 2 {
			clean := strings.ReplaceAll(matches[1], ".", "")
			if xp, err := strconv.Atoi(clean); err == nil && xp > 0 {
				return xp
			}
		}
	}
	return 0
}
