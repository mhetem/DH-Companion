package player

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/dice"
)

const twoHanded = "Two-Handed"

type Gear struct {
	Weapons     []cards.Weapon `json:"weapons"`
	Armor       []cards.Armor  `json:"armor"`
	Items       []cards.Loot   `json:"items"`
	Consumables []cards.Loot   `json:"consumables"`
}

func (s *Service) ListGear() Gear {
	return Gear{
		Weapons:     s.catalog.ListWeapons(),
		Armor:       s.catalog.ListArmor(),
		Items:       s.catalog.ListItems(),
		Consumables: s.catalog.ListConsumables(),
	}
}

type statMods struct {
	armorScore int
	evasion    int
	major      int
	severe     int
	hpMax      int
	stressMax  int
}

func (m *statMods) add(o statMods) {
	m.armorScore += o.armorScore
	m.evasion += o.evasion
	m.major += o.major
	m.severe += o.severe
	m.hpMax += o.hpMax
	m.stressMax += o.stressMax
}

var (
	armorScoreMod = regexp.MustCompile(`([+\-−]\d+) to (?:your )?Armor Score`)
	evasionMod    = regexp.MustCompile(`([+\-−]\d+) (?:to (?:all character traits and )?)?Evasion`)
	thresholdsMod = regexp.MustCompile(`([+\-−]\d+) to damage thresholds`)
	majorMod      = regexp.MustCompile(`([+\-−]\d+) to Major damage threshold`)
	severeMod     = regexp.MustCompile(`([+\-−]\d+) to Severe damage threshold`)
	weaponDamage  = regexp.MustCompile(`^d(\d+)(?:([+\-−])(\d+))?$`)
)

func sumMods(re *regexp.Regexp, text string) int {
	total := 0
	for _, m := range re.FindAllStringSubmatch(text, -1) {
		n, err := strconv.Atoi(strings.ReplaceAll(m[1], "−", "-"))
		if err == nil {
			total += n
		}
	}
	return total
}

func featureMods(f *cards.Feature) statMods {
	if f == nil {
		return statMods{}
	}
	text := f.Description
	both := sumMods(thresholdsMod, text)
	return statMods{
		armorScore: sumMods(armorScoreMod, text),
		evasion:    sumMods(evasionMod, text),
		major:      both + sumMods(majorMod, text),
		severe:     both + sumMods(severeMod, text),
	}
}

func parseWeaponDamage(raw string) (int, int, error) {
	m := weaponDamage.FindStringSubmatch(strings.TrimSpace(raw))
	if m == nil {
		return 0, 0, fmt.Errorf("can't read the damage %q", raw)
	}
	sides, _ := strconv.Atoi(m[1])
	flat := 0
	if m[3] != "" {
		flat, _ = strconv.Atoi(m[3])
		if m[2] != "+" {
			flat = -flat
		}
	}
	return sides, flat, nil
}

type defenseBase struct {
	key       string
	known     bool
	unarmored bool
	score     int
	major     int
	severe    int
}

type snapshot struct {
	hasArmor     bool
	armed        bool
	armorName    string
	armorFeature *cards.Feature
	base         defenseBase
	mods         statMods
	bonuses      []Bonus
}

func (s *Service) snapshotOf(row db.Character, items []db.InventoryItem, held []db.CharacterDomainCard) snapshot {
	snap := snapshot{base: defenseBase{key: "unarmored", known: true, unarmored: true}}
	for _, r := range items {
		if r.Equipped != 1 {
			continue
		}
		switch r.Kind {
		case "armor":
			snap.hasArmor = true
			snap.base = defenseBase{key: fmt.Sprintf("armor:%d:%s", r.ID, r.SrdSlug)}
			if a, ok := s.catalog.ArmorPiece(r.SrdSlug); ok && r.SrdSlug != "" {
				snap.base.known = true
				snap.base.score, snap.base.major, snap.base.severe = a.BaseScore, a.ThresholdMajor, a.ThresholdSevere
				snap.armorName, snap.armorFeature = r.Name, a.Feature
				snap.apply(r.Name, featureMods(a.Feature))
			}
		case "primary-weapon", "secondary-weapon":
			snap.armed = true
			if w, ok := s.catalog.Weapon(r.SrdSlug); ok && r.SrdSlug != "" {
				snap.apply(r.Name, featureMods(w.Feature))
			}
		}
	}
	s.abilityBonuses(&snap, row, held)
	return snap
}

func (s *Service) snapshot(q *db.Queries, characterID int64) (snapshot, error) {
	row, err := q.GetCharacter(s.ctx, characterID)
	if err != nil {
		return snapshot{}, fmt.Errorf("loading character: %w", err)
	}
	items, err := q.ListInventoryItems(s.ctx, characterID)
	if err != nil {
		return snapshot{}, fmt.Errorf("listing inventory: %w", err)
	}
	held, err := q.ListCharacterDomainCards(s.ctx, characterID)
	if err != nil {
		return snapshot{}, fmt.Errorf("listing domain cards: %w", err)
	}
	return s.snapshotOf(row, items, held), nil
}

func (s *Service) syncStats(q *db.Queries, characterID int64, before, after snapshot) error {
	same := before.base.key == after.base.key
	if same && before.mods == after.mods {
		return nil
	}
	row, err := q.GetCharacter(s.ctx, characterID)
	if err != nil {
		return fmt.Errorf("loading character: %w", err)
	}

	level := int(row.Level)
	evasion := int(row.Evasion) + after.mods.evasion - before.mods.evasion
	hpMax := clamp(int(row.HpMax)+after.mods.hpMax-before.mods.hpMax, 1, 20)
	stressMax := clamp(int(row.StressMax)+after.mods.stressMax-before.mods.stressMax, 1, 20)
	score, major, severe := int(row.ArmorScore), int(row.ThresholdMajor), int(row.ThresholdSevere)
	switch {
	case same || !after.base.known:
		score += after.mods.armorScore - before.mods.armorScore
		major += after.mods.major - before.mods.major
		severe += after.mods.severe - before.mods.severe
	case after.base.unarmored:
		score = after.mods.armorScore
		major = level + after.mods.major
		severe = 2*level + after.mods.severe
	default:
		score = after.base.score + after.mods.armorScore
		major = after.base.major + level + after.mods.major
		severe = after.base.severe + level + after.mods.severe
	}

	score = clamp(score, 0, 20)
	if _, err := q.SetCharacterStats(s.ctx, db.SetCharacterStatsParams{
		Evasion:         int64(clamp(evasion, 0, 30)),
		ArmorScore:      int64(score),
		ArmorMarked:     int64(min(int(row.ArmorMarked), score)),
		ThresholdMajor:  int64(clamp(major, 0, 99)),
		ThresholdSevere: int64(clamp(severe, 0, 99)),
		HpMax:           int64(hpMax),
		HpMarked:        int64(min(int(row.HpMarked), hpMax)),
		StressMax:       int64(stressMax),
		StressMarked:    int64(min(int(row.StressMarked), stressMax)),
		ID:              characterID,
	}); err != nil {
		return fmt.Errorf("updating defenses: %w", err)
	}
	return nil
}

func (s *Service) keepStats(q *db.Queries, characterID int64, fn func() error) error {
	before, err := s.snapshot(q, characterID)
	if err != nil {
		return err
	}
	if err := fn(); err != nil {
		return err
	}
	after, err := s.snapshot(q, characterID)
	if err != nil {
		return err
	}
	return s.syncStats(q, characterID, before, after)
}

func (s *Service) changeEquipment(characterID int64, fn func(q *db.Queries) error) ([]Item, error) {
	err := s.tx(func(q *db.Queries) error {
		return s.keepStats(q, characterID, func() error { return fn(q) })
	})
	if err != nil {
		return nil, err
	}
	return s.ListInventory(characterID)
}

func (s *Service) makeRoom(q *db.Queries, characterID int64, kind, slug string) error {
	if !slices.Contains(equipExclusiveKinds, kind) {
		return nil
	}
	if err := s.unequipKind(q, characterID, kind); err != nil {
		return err
	}
	switch kind {
	case "primary-weapon":
		if w, ok := s.catalog.Weapon(slug); ok && slug != "" && w.Burden == twoHanded {
			return s.unequipKind(q, characterID, "secondary-weapon")
		}
	case "secondary-weapon":
		rows, err := q.ListInventoryItems(s.ctx, characterID)
		if err != nil {
			return fmt.Errorf("listing inventory: %w", err)
		}
		for _, r := range rows {
			if r.Equipped != 1 || r.Kind != "primary-weapon" {
				continue
			}
			if w, ok := s.catalog.Weapon(r.SrdSlug); ok && r.SrdSlug != "" && w.Burden == twoHanded {
				return s.unequipKind(q, characterID, "primary-weapon")
			}
		}
	}
	return nil
}

func (s *Service) validateItemSlug(kind, slug string) (string, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "", nil
	}
	switch kind {
	case "primary-weapon", "secondary-weapon":
		w, ok := s.catalog.Weapon(slug)
		if !ok {
			return "", notFound("weapon", slug)
		}
		if want := weaponCategory(kind); w.Category != want {
			return "", fmt.Errorf("%s is a %s weapon, not a %s one", w.Name, strings.ToLower(w.Category), strings.ToLower(want))
		}
	case "armor":
		if _, ok := s.catalog.ArmorPiece(slug); !ok {
			return "", notFound("armor", slug)
		}
	case "item":
		if _, ok := s.catalog.Item(slug); !ok {
			return "", notFound("item", slug)
		}
	case "consumable":
		if _, ok := s.catalog.Consumable(slug); !ok {
			return "", notFound("consumable", slug)
		}
	}
	return slug, nil
}

func weaponCategory(kind string) string {
	if kind == "secondary-weapon" {
		return cards.CategorySecondary
	}
	return cards.CategoryPrimary
}

func (s *Service) RollWeaponDamage(characterID, itemID int64, critical bool) (DamageResult, error) {
	row, err := s.character(characterID)
	if err != nil {
		return DamageResult{}, err
	}
	item, err := s.inventoryItem(itemID)
	if err != nil {
		return DamageResult{}, err
	}
	if item.CharacterID != characterID {
		return DamageResult{}, fmt.Errorf("%s isn't in this character's inventory", item.Name)
	}
	w, ok := s.catalog.Weapon(item.SrdSlug)
	if !ok || item.SrdSlug == "" {
		return DamageResult{}, fmt.Errorf("%s isn't from the SRD weapon tables, so roll its damage by hand", item.Name)
	}
	sides, flat, err := parseWeaponDamage(w.Damage)
	if err != nil {
		return DamageResult{}, err
	}
	die, err := dice.ParseSides(sides)
	if err != nil {
		return DamageResult{}, err
	}
	count := clamp(int(row.Proficiency), 1, MaxProficiency)
	modifier := flat
	if critical {
		modifier += count * int(die)
	}
	return DamageResult{
		Label:    fmt.Sprintf("%s %dd%d", item.Name, count, int(die)),
		Count:    count,
		Sides:    int(die),
		Modifier: modifier,
		Total:    dice.RollDamage(count, die, modifier),
		Critical: critical,
	}, nil
}
