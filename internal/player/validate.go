package player

import (
	"fmt"
	"slices"
	"strings"
)

const (
	HopeMax             = 6
	LoadoutMax          = 5
	MaxLevel            = 10
	MaxProficiency      = 6
	StartingHope        = 2
	StartingArmor       = 0
	StartingDomainCards = 2
)

var TraitKeys = []string{"agility", "strength", "finesse", "instinct", "presence", "knowledge"}

var TraitArray = []int{2, 1, 1, 0, 0, -1}

var validLocations = []string{"loadout", "vault"}

var validItemKinds = []string{"primary-weapon", "secondary-weapon", "armor", "consumable", "item"}

var equipExclusiveKinds = []string{"primary-weapon", "secondary-weapon", "armor"}

func (s *Service) Traits() []string { return slices.Clone(TraitKeys) }

func (s *Service) ItemKinds() []string { return slices.Clone(validItemKinds) }

func (s *Service) Limits() map[string]int {
	return map[string]int{
		"hopeMax":        HopeMax,
		"loadoutMax":     LoadoutMax,
		"maxLevel":       MaxLevel,
		"maxProficiency": MaxProficiency,
	}
}

func validateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	return name, nil
}

func validateTraitKey(key string) (string, error) {
	k := strings.ToLower(strings.TrimSpace(key))
	if !slices.Contains(TraitKeys, k) {
		return "", fmt.Errorf("trait must be one of %s, got %q", strings.Join(TraitKeys, ", "), key)
	}
	return k, nil
}

func validateLocation(loc string) (string, error) {
	l := strings.ToLower(strings.TrimSpace(loc))
	if l == "" {
		return "loadout", nil
	}
	if !slices.Contains(validLocations, l) {
		return "", fmt.Errorf("location must be one of %s, got %q", strings.Join(validLocations, ", "), loc)
	}
	return l, nil
}

func validateItemKind(kind string) (string, error) {
	k := strings.ToLower(strings.TrimSpace(kind))
	if k == "" {
		return "item", nil
	}
	if !slices.Contains(validItemKinds, k) {
		return "", fmt.Errorf("item kind must be one of %s, got %q", strings.Join(validItemKinds, ", "), kind)
	}
	return k, nil
}

func (s *Service) validateBuild(classSlug, subclassSlug, ancestrySlug, communitySlug string) error {
	classSlug = strings.TrimSpace(classSlug)
	if classSlug == "" {
		return fmt.Errorf("class is required")
	}
	class, ok := s.catalog.Class(classSlug)
	if !ok {
		return notFound("class", classSlug)
	}
	subclassSlug = strings.TrimSpace(subclassSlug)
	if subclassSlug == "" {
		return fmt.Errorf("subclass is required")
	}
	if _, ok := class.Subclass(subclassSlug); !ok {
		return fmt.Errorf("%s is not a subclass of %s", subclassSlug, class.Name)
	}
	ancestrySlug = strings.TrimSpace(ancestrySlug)
	if ancestrySlug == "" {
		return fmt.Errorf("ancestry is required")
	}
	if _, ok := s.catalog.Ancestry(ancestrySlug); !ok {
		return notFound("ancestry", ancestrySlug)
	}
	communitySlug = strings.TrimSpace(communitySlug)
	if communitySlug == "" {
		return fmt.Errorf("community is required")
	}
	if _, ok := s.catalog.Community(communitySlug); !ok {
		return notFound("community", communitySlug)
	}
	return nil
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func tierForLevel(level int) int {
	switch {
	case level <= 1:
		return 1
	case level <= 4:
		return 2
	case level <= 7:
		return 3
	default:
		return 4
	}
}

func (s *Service) TierForLevel(level int) int { return tierForLevel(level) }
