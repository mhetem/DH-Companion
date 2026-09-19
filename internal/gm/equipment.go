package gm

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/rules"
)

// EquipmentFilter is the weapon/armor equivalent of Filter. Weapons browse on
// three axes rather than two, so this doesn't reuse Filter: Category splits
// primary from secondary, Type splits physical from magic. An empty field means
// "any", which is why there's no "All" sentinel the way Filter has one.
type EquipmentFilter struct {
	Tier     string `json:"tier"`
	Type     string `json:"type"`
	Category string `json:"category"`
}

func (f EquipmentFilter) normalized() (tier, typ, category string) {
	return strings.TrimSpace(f.Tier), strings.TrimSpace(f.Type), strings.TrimSpace(f.Category)
}

func (f EquipmentFilter) matches(tier, typ, category string) bool {
	ft, fy, fc := f.normalized()
	if ft != "" && ft != tier {
		return false
	}
	if fy != "" && fy != typ {
		return false
	}
	if fc != "" && fc != category {
		return false
	}
	return true
}

type BrowseWeapon struct {
	cards.Weapon
	Source string `json:"source"`
}

type BrowseArmorPiece struct {
	cards.Armor
	Source string `json:"source"`
}

// A weapon or armor carries at most one feature, so it's stored as a single JSON
// object rather than the array the adversary and environment tables use. An empty
// column is a card with no feature.
func decodeFeature(raw, kind, slug string) (*cards.Feature, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var f cards.Feature
	if err := json.Unmarshal([]byte(raw), &f); err != nil {
		return nil, fmt.Errorf("%s %q has an unreadable feature: %w", kind, slug, err)
	}
	return &f, nil
}

func encodeFeature(f *cards.Feature) (string, error) {
	if f == nil || strings.TrimSpace(f.Title) == "" {
		return "", nil
	}
	b, err := json.Marshal(f)
	if err != nil {
		return "", fmt.Errorf("encoding feature: %w", err)
	}
	return string(b), nil
}

func toWeapon(r db.CustomWeapon) (cards.Weapon, error) {
	feature, err := decodeFeature(r.Feature, "weapon", r.Slug)
	if err != nil {
		return cards.Weapon{}, err
	}
	return cards.Weapon{
		Meta: cards.Meta{
			Kind:        cards.KindWeapon,
			Slug:        r.Slug,
			Name:        r.Name,
			Tier:        r.Tier,
			Type:        r.Type,
			Description: r.Description,
		},
		Category:   r.Category,
		Trait:      r.Trait,
		Range:      r.WeaponRange,
		Damage:     r.Damage,
		DamageType: r.DamageType,
		Burden:     r.Burden,
		Feature:    feature,
	}, nil
}

func toArmor(r db.CustomArmor) (cards.Armor, error) {
	feature, err := decodeFeature(r.Feature, "armor", r.Slug)
	if err != nil {
		return cards.Armor{}, err
	}
	return cards.Armor{
		Meta: cards.Meta{
			Kind:        cards.KindArmor,
			Slug:        r.Slug,
			Name:        r.Name,
			Tier:        r.Tier,
			Type:        armorType,
			Description: r.Description,
		},
		ThresholdMajor:  int(r.ThresholdMajor),
		ThresholdSevere: int(r.ThresholdSevere),
		BaseScore:       int(r.BaseScore),
		Feature:         feature,
	}, nil
}

func validateWeapon(w cards.Weapon) (name, feature string, err error) {
	if name, err = validateName(w.Name); err != nil {
		return "", "", err
	}
	if err = validateTier(w.Tier); err != nil {
		return "", "", err
	}
	if err = validateWeaponType(w.Type); err != nil {
		return "", "", err
	}
	if err = validateWeaponCategory(w.Category); err != nil {
		return "", "", err
	}
	if feature, err = encodeFeature(w.Feature); err != nil {
		return "", "", err
	}
	return name, feature, nil
}

func validateArmorCard(a cards.Armor) (name, feature string, err error) {
	if name, err = validateName(a.Name); err != nil {
		return "", "", err
	}
	if err = validateTier(a.Tier); err != nil {
		return "", "", err
	}
	// The SRD caps a PC's Armor Score at 12, so a base score above that can never
	// be legal no matter what the rest of the build adds.
	if a.BaseScore < 0 || a.BaseScore > maxArmorScore {
		return "", "", fmt.Errorf("base Armor Score must be between 0 and %d, got %d", maxArmorScore, a.BaseScore)
	}
	if a.ThresholdMajor < 0 || a.ThresholdSevere < 0 {
		return "", "", fmt.Errorf("damage thresholds can't be negative")
	}
	if a.ThresholdSevere < a.ThresholdMajor {
		return "", "", fmt.Errorf("the Severe threshold (%d) must be at least the Major threshold (%d)", a.ThresholdSevere, a.ThresholdMajor)
	}
	if feature, err = encodeFeature(a.Feature); err != nil {
		return "", "", err
	}
	return name, feature, nil
}

func (s *Service) ListCustomWeapons(filter EquipmentFilter) ([]cards.Weapon, error) {
	tier, typ, category := filter.normalized()
	rows, err := s.q.ShowAllCustomWeapons(s.ctx, db.ShowAllCustomWeaponsParams{
		Tier: tier, Type: typ, Category: category,
	})
	if err != nil {
		return nil, fmt.Errorf("listing custom weapons: %w", err)
	}
	out := make([]cards.Weapon, 0, len(rows))
	for _, r := range rows {
		w, err := toWeapon(r)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, nil
}

func (s *Service) ListCustomArmor(filter EquipmentFilter) ([]cards.Armor, error) {
	tier, _, _ := filter.normalized()
	rows, err := s.q.ShowAllCustomArmor(s.ctx, tier)
	if err != nil {
		return nil, fmt.Errorf("listing custom armor: %w", err)
	}
	out := make([]cards.Armor, 0, len(rows))
	for _, r := range rows {
		a, err := toArmor(r)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

func (s *Service) GetCustomWeapon(slug string) (cards.Weapon, error) {
	r, err := s.q.GetCustomWeaponBySlug(s.ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return cards.Weapon{}, notFound("custom weapon", slug)
	}
	if err != nil {
		return cards.Weapon{}, fmt.Errorf("loading custom weapon: %w", err)
	}
	return toWeapon(r)
}

func (s *Service) GetCustomArmor(slug string) (cards.Armor, error) {
	r, err := s.q.GetCustomArmorBySlug(s.ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return cards.Armor{}, notFound("custom armor", slug)
	}
	if err != nil {
		return cards.Armor{}, fmt.Errorf("loading custom armor: %w", err)
	}
	return toArmor(r)
}

func (s *Service) CreateCustomWeapon(w cards.Weapon) (cards.Weapon, error) {
	name, feature, err := validateWeapon(w)
	if err != nil {
		return cards.Weapon{}, err
	}
	slug := slugify(name)
	if slug == "" {
		return cards.Weapon{}, fmt.Errorf("name %q does not produce a usable slug", name)
	}

	r, err := s.q.CreateCustomWeapon(s.ctx, db.CreateCustomWeaponParams{
		Slug:        slug,
		Name:        name,
		Tier:        w.Tier,
		Type:        w.Type,
		Description: w.Description,
		Category:    w.Category,
		Trait:       w.Trait,
		WeaponRange: w.Range,
		Damage:      w.Damage,
		DamageType:  w.DamageType,
		Burden:      w.Burden,
		Feature:     feature,
	})
	if isUniqueViolation(err) {
		return cards.Weapon{}, fmt.Errorf("a custom weapon named %q already exists", name)
	}
	if err != nil {
		return cards.Weapon{}, fmt.Errorf("creating custom weapon: %w", err)
	}
	return toWeapon(r)
}

func (s *Service) UpdateCustomWeapon(w cards.Weapon) (cards.Weapon, error) {
	if w.Slug == "" {
		return cards.Weapon{}, fmt.Errorf("slug is required to update a custom weapon")
	}
	name, feature, err := validateWeapon(w)
	if err != nil {
		return cards.Weapon{}, err
	}

	r, err := s.q.UpdateCustomWeapon(s.ctx, db.UpdateCustomWeaponParams{
		Name:        name,
		Tier:        w.Tier,
		Type:        w.Type,
		Description: w.Description,
		Category:    w.Category,
		Trait:       w.Trait,
		WeaponRange: w.Range,
		Damage:      w.Damage,
		DamageType:  w.DamageType,
		Burden:      w.Burden,
		Feature:     feature,
		Slug:        w.Slug,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return cards.Weapon{}, notFound("custom weapon", w.Slug)
	}
	if err != nil {
		return cards.Weapon{}, fmt.Errorf("updating custom weapon: %w", err)
	}
	return toWeapon(r)
}

func (s *Service) DeleteCustomWeapon(slug string) error {
	if err := s.q.DeleteCustomWeapon(s.ctx, slug); err != nil {
		return fmt.Errorf("deleting custom weapon: %w", err)
	}
	return nil
}

func (s *Service) CreateCustomArmor(a cards.Armor) (cards.Armor, error) {
	name, feature, err := validateArmorCard(a)
	if err != nil {
		return cards.Armor{}, err
	}
	slug := slugify(name)
	if slug == "" {
		return cards.Armor{}, fmt.Errorf("name %q does not produce a usable slug", name)
	}

	r, err := s.q.CreateCustomArmor(s.ctx, db.CreateCustomArmorParams{
		Slug:            slug,
		Name:            name,
		Tier:            a.Tier,
		Description:     a.Description,
		ThresholdMajor:  int64(a.ThresholdMajor),
		ThresholdSevere: int64(a.ThresholdSevere),
		BaseScore:       int64(a.BaseScore),
		Feature:         feature,
	})
	if isUniqueViolation(err) {
		return cards.Armor{}, fmt.Errorf("a custom armor named %q already exists", name)
	}
	if err != nil {
		return cards.Armor{}, fmt.Errorf("creating custom armor: %w", err)
	}
	return toArmor(r)
}

func (s *Service) UpdateCustomArmor(a cards.Armor) (cards.Armor, error) {
	if a.Slug == "" {
		return cards.Armor{}, fmt.Errorf("slug is required to update a custom armor")
	}
	name, feature, err := validateArmorCard(a)
	if err != nil {
		return cards.Armor{}, err
	}

	r, err := s.q.UpdateCustomArmor(s.ctx, db.UpdateCustomArmorParams{
		Name:            name,
		Tier:            a.Tier,
		Description:     a.Description,
		ThresholdMajor:  int64(a.ThresholdMajor),
		ThresholdSevere: int64(a.ThresholdSevere),
		BaseScore:       int64(a.BaseScore),
		Feature:         feature,
		Slug:            a.Slug,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return cards.Armor{}, notFound("custom armor", a.Slug)
	}
	if err != nil {
		return cards.Armor{}, fmt.Errorf("updating custom armor: %w", err)
	}
	return toArmor(r)
}

func (s *Service) DeleteCustomArmor(slug string) error {
	if err := s.q.DeleteCustomArmor(s.ctx, slug); err != nil {
		return fmt.Errorf("deleting custom armor: %w", err)
	}
	return nil
}

// BrowseWeapons and BrowseArmor merge the SRD catalog with the GM's homebrew,
// letting a custom card win on a slug clash the same way the adversary and
// environment browsers do.
func (s *Service) BrowseWeapons(filter EquipmentFilter) ([]BrowseWeapon, error) {
	bySlug := map[string]BrowseWeapon{}
	for _, w := range s.catalog.ListWeapons() {
		if filter.matches(w.Tier, w.Type, w.Category) {
			bySlug[w.Slug] = BrowseWeapon{Weapon: w, Source: rules.SourceSRD}
		}
	}
	custom, err := s.ListCustomWeapons(filter)
	if err != nil {
		return nil, err
	}
	for _, w := range custom {
		bySlug[w.Slug] = BrowseWeapon{Weapon: w, Source: rules.SourceCustom}
	}

	out := make([]BrowseWeapon, 0, len(bySlug))
	for _, w := range bySlug {
		out = append(out, w)
	}
	sortEquipment(out, func(w BrowseWeapon) (string, string) { return w.Tier, w.Name })
	return out, nil
}

func (s *Service) BrowseArmor(filter EquipmentFilter) ([]BrowseArmorPiece, error) {
	bySlug := map[string]BrowseArmorPiece{}
	for _, a := range s.catalog.ListArmor() {
		if filter.matches(a.Tier, a.Type, "") {
			bySlug[a.Slug] = BrowseArmorPiece{Armor: a, Source: rules.SourceSRD}
		}
	}
	custom, err := s.ListCustomArmor(filter)
	if err != nil {
		return nil, err
	}
	for _, a := range custom {
		if filter.matches(a.Tier, a.Type, "") {
			bySlug[a.Slug] = BrowseArmorPiece{Armor: a, Source: rules.SourceCustom}
		}
	}

	out := make([]BrowseArmorPiece, 0, len(bySlug))
	for _, a := range bySlug {
		out = append(out, a)
	}
	sortEquipment(out, func(a BrowseArmorPiece) (string, string) { return a.Tier, a.Name })
	return out, nil
}

// lookupWeapon and lookupArmor resolve a slug against homebrew first, mirroring
// how the browsers let a custom card shadow an SRD one.
func (s *Service) lookupWeapon(slug string) (cards.Weapon, string, bool) {
	if w, err := s.GetCustomWeapon(slug); err == nil {
		return w, rules.SourceCustom, true
	} else if !errors.Is(err, ErrNotFound) {
		return cards.Weapon{}, "", false
	}
	if w, ok := s.catalog.Weapon(slug); ok {
		return w, rules.SourceSRD, true
	}
	return cards.Weapon{}, "", false
}

func (s *Service) lookupArmor(slug string) (cards.Armor, string, bool) {
	if a, err := s.GetCustomArmor(slug); err == nil {
		return a, rules.SourceCustom, true
	} else if !errors.Is(err, ErrNotFound) {
		return cards.Armor{}, "", false
	}
	if a, ok := s.catalog.ArmorPiece(slug); ok {
		return a, rules.SourceSRD, true
	}
	return cards.Armor{}, "", false
}

func (s *Service) GetWeapon(slug string) (BrowseWeapon, error) {
	w, source, ok := s.lookupWeapon(slug)
	if !ok {
		return BrowseWeapon{}, notFound("weapon", slug)
	}
	return BrowseWeapon{Weapon: w, Source: source}, nil
}

func (s *Service) GetArmor(slug string) (BrowseArmorPiece, error) {
	a, source, ok := s.lookupArmor(slug)
	if !ok {
		return BrowseArmorPiece{}, notFound("armor", slug)
	}
	return BrowseArmorPiece{Armor: a, Source: source}, nil
}
