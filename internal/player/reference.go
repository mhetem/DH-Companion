package player

import (
	"sort"
	"strconv"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
)

var Domains = []string{"Arcana", "Blade", "Bone", "Codex", "Grace", "Midnight", "Sage", "Splendor", "Valor"}

type DomainFilter struct {
	Domain   string `json:"domain"`
	MaxLevel int    `json:"maxLevel"`
	Level    string `json:"level"`
}

func (s *Service) ListClasses() []cards.CharacterClass { return s.catalog.ListClasses() }

func (s *Service) GetClass(slug string) (cards.CharacterClass, error) {
	c, ok := s.catalog.Class(strings.TrimSpace(slug))
	if !ok {
		return cards.CharacterClass{}, notFound("class", slug)
	}
	return c, nil
}

func (s *Service) ListAncestries() []cards.Ancestry { return s.catalog.ListAncestries() }

func (s *Service) GetAncestry(slug string) (cards.Ancestry, error) {
	a, ok := s.catalog.Ancestry(strings.TrimSpace(slug))
	if !ok {
		return cards.Ancestry{}, notFound("ancestry", slug)
	}
	return a, nil
}

func (s *Service) ListCommunities() []cards.Community { return s.catalog.ListCommunities() }

func (s *Service) GetCommunity(slug string) (cards.Community, error) {
	m, ok := s.catalog.Community(strings.TrimSpace(slug))
	if !ok {
		return cards.Community{}, notFound("community", slug)
	}
	return m, nil
}

func (s *Service) DomainList() []string {
	out := make([]string, len(Domains))
	copy(out, Domains)
	return out
}

func (s *Service) GetDomainCard(slug string) (cards.DomainCard, error) {
	d, ok := s.catalog.DomainCard(strings.TrimSpace(slug))
	if !ok {
		return cards.DomainCard{}, notFound("domain card", slug)
	}
	return d, nil
}

func (s *Service) BrowseDomainCards(f DomainFilter) []cards.DomainCard {
	domain := strings.TrimSpace(f.Domain)
	level := strings.TrimSpace(f.Level)

	out := make([]cards.DomainCard, 0, len(s.catalog.DomainCards))
	for _, d := range s.catalog.ListDomainCards() {
		if domain != "" && domain != "All" && d.Domain != domain {
			continue
		}
		if level != "" && level != "All" && d.Level != level {
			continue
		}
		if f.MaxLevel > 0 && cardLevel(d) > f.MaxLevel {
			continue
		}
		out = append(out, d)
	}
	return out
}

func (s *Service) AvailableDomainCards(characterID int64) ([]cards.DomainCard, error) {
	row, err := s.character(characterID)
	if err != nil {
		return nil, err
	}
	owned, err := s.q.ListCharacterDomainCards(s.ctx, characterID)
	if err != nil {
		return nil, err
	}
	held := map[string]bool{}
	for _, o := range owned {
		held[o.CardSlug] = true
	}

	view := s.characterView(row)
	out := []cards.DomainCard{}
	for _, d := range s.catalog.ListDomainCards() {
		if held[d.Slug] {
			continue
		}
		if cardLevel(d) > int(row.Level) {
			continue
		}
		if len(view.Domains) > 0 && !contains(view.Domains, d.Domain) {
			continue
		}
		out = append(out, d)
	}
	return out, nil
}

func (s *Service) ListBeastforms(tier string) []cards.Beastform {
	tier = strings.TrimSpace(tier)
	if tier == "" || tier == "All" {
		return s.catalog.ListBeastforms()
	}
	return s.catalog.BeastformsUpToTier(tier)
}

func (s *Service) ListCompanions() []cards.Companion {
	out := make([]cards.Companion, 0, len(s.catalog.Companions))
	for _, c := range s.catalog.Companions {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func cardLevel(d cards.DomainCard) int {
	n, err := strconv.Atoi(strings.TrimSpace(d.Level))
	if err != nil {
		return 1
	}
	return n
}
