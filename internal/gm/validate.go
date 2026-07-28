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
)

const defaultNoteKind = "npc"

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
