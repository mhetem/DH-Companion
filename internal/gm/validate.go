package gm

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var (
	validTiers            = []string{"1", "2", "3", "4"}
	validAdversaryTypes   = []string{"Bruiser", "Horde", "Leader", "Minion", "Ranged", "Skulk", "Social", "Solo", "Standard", "Support"}
	validEnvironmentTypes = []string{"Event", "Exploration", "Social", "Traversal"}
	validNoteKinds        = []string{"npc", "location", "faction", "lore", "plot"}
	validWeaponTypes      = []string{"Physical", "Magic"}
	validWeaponCategories = []string{"Primary", "Secondary"}
	validLootKinds        = []string{"item", "consumable"}
)

const defaultNoteKind = "npc"

// armorType is the single Type every armor carries: armor has no sub-types the
// way adversaries and weapons do, but Meta.Type is what the browsers filter on.
const armorType = "Armor"

// maxArmorScore mirrors the SRD's cap on a PC's Armor Score.
const maxArmorScore = 12

func validateWeaponType(t string) error {
	if !slices.Contains(validWeaponTypes, strings.TrimSpace(t)) {
		return fmt.Errorf("weapon damage must be one of %s, got %q", strings.Join(validWeaponTypes, ", "), t)
	}
	return nil
}

func validateWeaponCategory(c string) error {
	if !slices.Contains(validWeaponCategories, strings.TrimSpace(c)) {
		return fmt.Errorf("weapon category must be one of %s, got %q", strings.Join(validWeaponCategories, ", "), c)
	}
	return nil
}

func validateLootKind(kind string) (string, error) {
	k := strings.ToLower(strings.TrimSpace(kind))
	if !slices.Contains(validLootKinds, k) {
		return "", fmt.Errorf("loot kind must be one of %s, got %q", strings.Join(validLootKinds, ", "), kind)
	}
	return k, nil
}

func validateTier(tier string) error {
	if !slices.Contains(validTiers, strings.TrimSpace(tier)) {
		return fmt.Errorf("tier must be one of %s, got %q", strings.Join(validTiers, ", "), tier)
	}
	return nil
}

func validateAdversaryType(t string) error {
	if !slices.Contains(validAdversaryTypes, strings.TrimSpace(t)) {
		return fmt.Errorf("adversary type must be one of %s, got %q", strings.Join(validAdversaryTypes, ", "), t)
	}
	return nil
}

func validateEnvironmentType(t string) error {
	if !slices.Contains(validEnvironmentTypes, strings.TrimSpace(t)) {
		return fmt.Errorf("environment type must be one of %s, got %q", strings.Join(validEnvironmentTypes, ", "), t)
	}
	return nil
}

func validateNoteKind(kind string) (string, error) {
	k := strings.ToLower(strings.TrimSpace(kind))
	if k == "" {
		return defaultNoteKind, nil
	}
	if !slices.Contains(validNoteKinds, k) {
		return "", fmt.Errorf("note kind must be one of %s, got %q", strings.Join(validNoteKinds, ", "), kind)
	}
	return k, nil
}

func (s *Service) NoteKinds() []string {
	return slices.Clone(validNoteKinds)
}

func validateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	return name, nil
}

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := nonSlugChars.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "-")
	return strings.Trim(s, "-")
}
