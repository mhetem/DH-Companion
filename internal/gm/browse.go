package gm

import (
	"errors"
	"sort"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/rules"
)

type BrowseAdversary struct {
	cards.Adversary
	Source string `json:"source"`
}

type BrowseEnvironment struct {
	cards.Environment
	Source string `json:"source"`
}

func matchesFilter(tier, typ, cardTier, cardType string) bool {
	if tier != "" && cardTier != tier {
		return false
	}
	if typ != "All" && cardType != typ {
		return false
	}
	return true
}

func (s *Service) BrowseAdversaries(filter Filter) ([]BrowseAdversary, error) {
	tier, typ := filter.normalized()

	bySlug := map[string]BrowseAdversary{}
	for slug, a := range s.catalog.Adversaries {
		if matchesFilter(tier, typ, a.Tier, a.Type) {
			bySlug[slug] = BrowseAdversary{Adversary: a, Source: rules.SourceSRD}
		}
	}
	custom, err := s.ListCustomAdversaries(filter)
	if err != nil {
		return nil, err
	}
	for _, a := range custom {
		bySlug[a.Slug] = BrowseAdversary{Adversary: a, Source: rules.SourceCustom}
	}

	out := make([]BrowseAdversary, 0, len(bySlug))
	for _, a := range bySlug {
		out = append(out, a)
	}
	sortByName(out, func(a BrowseAdversary) string { return a.Name })
	return out, nil
}

func (s *Service) BrowseEnvironments(filter Filter) ([]BrowseEnvironment, error) {
	tier, typ := filter.normalized()

	bySlug := map[string]BrowseEnvironment{}
	for slug, e := range s.catalog.Environments {
		if matchesFilter(tier, typ, e.Tier, e.Type) {
			bySlug[slug] = BrowseEnvironment{Environment: e, Source: rules.SourceSRD}
		}
	}
	custom, err := s.ListCustomEnvironments(filter)
	if err != nil {
		return nil, err
	}
	for _, e := range custom {
		bySlug[e.Slug] = BrowseEnvironment{Environment: e, Source: rules.SourceCustom}
	}

	out := make([]BrowseEnvironment, 0, len(bySlug))
	for _, e := range bySlug {
		out = append(out, e)
	}
	sortByName(out, func(e BrowseEnvironment) string { return e.Name })
	return out, nil
}

func sortByName[T any](items []T, name func(T) string) {
	sort.SliceStable(items, func(i, j int) bool {
		return strings.ToLower(name(items[i])) < strings.ToLower(name(items[j]))
	})
}

func (s *Service) lookupAdversary(slug string) (cards.Adversary, string, bool) {
	if a, err := s.GetCustomAdversary(slug); err == nil {
		return a, rules.SourceCustom, true
	} else if !errors.Is(err, ErrNotFound) {
		return cards.Adversary{}, "", false
	}
	if a, ok := s.catalog.Adversaries[slug]; ok {
		return a, rules.SourceSRD, true
	}
	return cards.Adversary{}, "", false
}

func (s *Service) lookupEnvironment(slug string) (cards.Environment, string, bool) {
	if e, err := s.GetCustomEnvironment(slug); err == nil {
		return e, rules.SourceCustom, true
	} else if !errors.Is(err, ErrNotFound) {
		return cards.Environment{}, "", false
	}
	if e, ok := s.catalog.Environments[slug]; ok {
		return e, rules.SourceSRD, true
	}
	return cards.Environment{}, "", false
}

func cardStub(slug string) cards.Adversary {
	return cards.Adversary{
		Meta:     cards.Meta{Kind: cards.KindAdversary, Slug: slug, Name: slug},
		Features: []cards.Feature{},
	}
}

func (s *Service) GetAdversary(slug string) (BrowseAdversary, error) {
	a, source, ok := s.lookupAdversary(slug)
	if !ok {
		return BrowseAdversary{}, notFound("adversary", slug)
	}
	return BrowseAdversary{Adversary: a, Source: source}, nil
}

func (s *Service) GetEnvironment(slug string) (BrowseEnvironment, error) {
	e, source, ok := s.lookupEnvironment(slug)
	if !ok {
		return BrowseEnvironment{}, notFound("environment", slug)
	}
	return BrowseEnvironment{Environment: e, Source: source}, nil
}

func (s *Service) ComputeBudget(settings rules.EncounterSettings, picks []rules.EncounterAdversary) rules.BudgetSummary {
	return rules.ComputeBudget(settings, picks)
}
