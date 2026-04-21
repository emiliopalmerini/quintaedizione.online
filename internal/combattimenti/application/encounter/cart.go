package encounter

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/encounter"
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
)

// CartItemRef is a raw cart entry as posted by the browser: "{id}@{source}".
type CartItemRef struct {
	ID     string
	Source string
}

// ParseCartRefs decodes the repeated "monsters[]" form values into structured
// refs. Malformed values (missing "@source" separator or empty fields) are
// silently dropped; callers can log them via the logger hook in the pricer.
func ParseCartRefs(raw []string) []CartItemRef {
	refs := make([]CartItemRef, 0, len(raw))
	for _, v := range raw {
		id, src, found := strings.Cut(strings.TrimSpace(v), "@")
		if !found || id == "" || src == "" {
			continue
		}
		refs = append(refs, CartItemRef{ID: id, Source: src})
	}
	return refs
}

// PriceCartRequest is the cart-pricing input.
type PriceCartRequest struct {
	Ruleset string
	Refs    []CartItemRef
	// Budget is the encounter's total XP budget. Used to compute remaining.
	Budget int
}

// PriceCartResponse holds the priced cart.
type PriceCartResponse struct {
	Entries       []monster.CartEntry
	Subtotal      int
	EffectiveCost int
	Remaining     int
}

// CartPricer resolves cart refs to monsters and prices them under the active
// ruleset's multiplier rules.
type CartPricer struct {
	logger *slog.Logger
	reader monster.Reader
	repo   encounter.Repository
}

// NewCartPricer builds a CartPricer with the given monster reader and
// encounter repository (for 2014 multiplier lookup).
func NewCartPricer(logger *slog.Logger, reader monster.Reader, repo encounter.Repository) *CartPricer {
	return &CartPricer{logger: logger, reader: reader, repo: repo}
}

// Price resolves cart refs to monster details and returns the fully priced
// cart (subtotal, effective cost, remaining budget). Unknown or
// different-source entries are dropped with a debug log; the cart is still
// returned with whatever resolved.
//
// Ruleset switch also acts as a cart-clear trigger: entries whose source
// does not share a ruleset with the request are silently dropped. This keeps
// the cart consistent when the user toggles between editions.
func (p *CartPricer) Price(ctx context.Context, req PriceCartRequest) (*PriceCartResponse, error) {
	ruleset, err := encounter.NewRuleset(req.Ruleset)
	if err != nil {
		return nil, fmt.Errorf("invalid ruleset: %w", err)
	}

	cart := monster.Cart{Entries: make([]monster.CartEntry, 0, len(req.Refs))}
	for _, ref := range req.Refs {
		m, err := p.reader.FindByID(ctx, ref.Source, ref.ID)
		if err != nil {
			p.logger.Debug("cart entry dropped", "id", ref.ID, "source", ref.Source, "error", err)
			continue
		}
		cart.Entries = append(cart.Entries, monster.CartEntry{
			ID:     m.ID,
			Source: m.Source,
			Name:   m.Name,
			CR:     m.CR,
			UnitXP: m.XP,
		})
	}

	multiplier := p.multiplierFor(ruleset)
	effective := cart.EffectiveCost(multiplier)
	return &PriceCartResponse{
		Entries:       cart.Entries,
		Subtotal:      cart.Subtotal(),
		EffectiveCost: effective,
		Remaining:     req.Budget - effective,
	}, nil
}

func (p *CartPricer) multiplierFor(ruleset encounter.Ruleset) func(int) float64 {
	if ruleset == encounter.Ruleset2024 {
		return func(int) float64 { return 1.0 }
	}
	// 2014: delegate to the repository's multiplier table.
	return func(count int) float64 {
		if count < 1 {
			return 1.0
		}
		m, err := p.repo.GetMultiplierFor2014(count)
		if err != nil {
			p.logger.Warn("multiplier lookup failed", "count", count, "error", err)
			return 1.0
		}
		return m
	}
}
