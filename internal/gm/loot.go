package gm

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/dice"
	"github.com/mhetem/DH-Companion/internal/rules"
	"github.com/mhetem/DH-Companion/internal/srd"
)

// homebrewTable is the label homebrew loot carries in place of a source book.
// It isn't a real table: homebrew entries have no roll number, because a roll is
// an index into a full 60-row table.
const homebrewTable = "Homebrew"

// LootFilter browses the item and consumable tables. They carry no tier, so the
// axes are rarity and which set the row came from. Empty means "any".
type LootFilter struct {
	Rarity string `json:"rarity"`
	Table  string `json:"table"`
}

func (f LootFilter) matches(l cards.Loot) bool {
	if r := strings.TrimSpace(f.Rarity); r != "" && !strings.EqualFold(r, l.Rarity) {
		return false
	}
	if t := strings.TrimSpace(f.Table); t != "" && t != l.Table {
		return false
	}
	return true
}

type BrowseLoot struct {
	cards.Loot
	Source string `json:"source"`
}

func toLoot(r db.CustomLoot) cards.Loot {
	kind := cards.KindItem
	typ := "Item"
	if r.Kind == string(cards.KindConsumable) {
		kind, typ = cards.KindConsumable, "Consumable"
	}
	return cards.Loot{
		Meta: cards.Meta{
			Kind:        kind,
			Slug:        r.Slug,
			Name:        r.Name,
			Type:        typ,
			Description: r.Description,
		},
		Table:  homebrewTable,
		Rarity: r.Rarity,
	}
}

func validateLootCard(l cards.Loot) (name, rarity string, err error) {
	if name, err = validateName(l.Name); err != nil {
		return "", "", err
	}
	band, ok := rules.Rarity(l.Rarity)
	if !ok {
		return "", "", fmt.Errorf("rarity must be one of Common, Uncommon, Rare, Legendary, got %q", l.Rarity)
	}
	return name, band.Name, nil
}

func (s *Service) listCustomLoot(kind, rarity string) ([]cards.Loot, error) {
	rows, err := s.q.ShowAllCustomLoot(s.ctx, db.ShowAllCustomLootParams{
		Kind: kind, Rarity: strings.TrimSpace(rarity),
	})
	if err != nil {
		return nil, fmt.Errorf("listing custom %ss: %w", kind, err)
	}
	out := make([]cards.Loot, 0, len(rows))
	for _, r := range rows {
		out = append(out, toLoot(r))
	}
	return out, nil
}

func (s *Service) ListCustomItems(filter LootFilter) ([]cards.Loot, error) {
	return s.listCustomLoot(string(cards.KindItem), filter.Rarity)
}

func (s *Service) ListCustomConsumables(filter LootFilter) ([]cards.Loot, error) {
	return s.listCustomLoot(string(cards.KindConsumable), filter.Rarity)
}

func (s *Service) getCustomLoot(kind, slug string) (cards.Loot, error) {
	r, err := s.q.GetCustomLootBySlug(s.ctx, db.GetCustomLootBySlugParams{Kind: kind, Slug: slug})
	if errors.Is(err, sql.ErrNoRows) {
		return cards.Loot{}, notFound("custom "+kind, slug)
	}
	if err != nil {
		return cards.Loot{}, fmt.Errorf("loading custom %s: %w", kind, err)
	}
	return toLoot(r), nil
}

func (s *Service) GetCustomItem(slug string) (cards.Loot, error) {
	return s.getCustomLoot(string(cards.KindItem), slug)
}

func (s *Service) GetCustomConsumable(slug string) (cards.Loot, error) {
	return s.getCustomLoot(string(cards.KindConsumable), slug)
}

func (s *Service) createCustomLoot(kind string, l cards.Loot) (cards.Loot, error) {
	name, rarity, err := validateLootCard(l)
	if err != nil {
		return cards.Loot{}, err
	}
	slug := slugify(name)
	if slug == "" {
		return cards.Loot{}, fmt.Errorf("name %q does not produce a usable slug", name)
	}

	r, err := s.q.CreateCustomLoot(s.ctx, db.CreateCustomLootParams{
		Kind:        kind,
		Slug:        slug,
		Name:        name,
		Rarity:      rarity,
		Description: l.Description,
	})
	if isUniqueViolation(err) {
		return cards.Loot{}, fmt.Errorf("a custom %s named %q already exists", kind, name)
	}
	if err != nil {
		return cards.Loot{}, fmt.Errorf("creating custom %s: %w", kind, err)
	}
	return toLoot(r), nil
}

func (s *Service) updateCustomLoot(kind string, l cards.Loot) (cards.Loot, error) {
	if l.Slug == "" {
		return cards.Loot{}, fmt.Errorf("slug is required to update a custom %s", kind)
	}
	name, rarity, err := validateLootCard(l)
	if err != nil {
		return cards.Loot{}, err
	}

	r, err := s.q.UpdateCustomLoot(s.ctx, db.UpdateCustomLootParams{
		Name:        name,
		Rarity:      rarity,
		Description: l.Description,
		Kind:        kind,
		Slug:        l.Slug,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return cards.Loot{}, notFound("custom "+kind, l.Slug)
	}
	if err != nil {
		return cards.Loot{}, fmt.Errorf("updating custom %s: %w", kind, err)
	}
	return toLoot(r), nil
}

func (s *Service) CreateCustomItem(l cards.Loot) (cards.Loot, error) {
	return s.createCustomLoot(string(cards.KindItem), l)
}

func (s *Service) CreateCustomConsumable(l cards.Loot) (cards.Loot, error) {
	return s.createCustomLoot(string(cards.KindConsumable), l)
}

func (s *Service) UpdateCustomItem(l cards.Loot) (cards.Loot, error) {
	return s.updateCustomLoot(string(cards.KindItem), l)
}

func (s *Service) UpdateCustomConsumable(l cards.Loot) (cards.Loot, error) {
	return s.updateCustomLoot(string(cards.KindConsumable), l)
}

func (s *Service) deleteCustomLoot(kind, slug string) error {
	if err := s.q.DeleteCustomLoot(s.ctx, db.DeleteCustomLootParams{Kind: kind, Slug: slug}); err != nil {
		return fmt.Errorf("deleting custom %s: %w", kind, err)
	}
	return nil
}

func (s *Service) DeleteCustomItem(slug string) error {
	return s.deleteCustomLoot(string(cards.KindItem), slug)
}

func (s *Service) DeleteCustomConsumable(slug string) error {
	return s.deleteCustomLoot(string(cards.KindConsumable), slug)
}

func (s *Service) BrowseItems(filter LootFilter) ([]BrowseLoot, error) {
	return s.browseLoot(s.catalog.ListItems(), s.ListCustomItems, filter)
}

func (s *Service) BrowseConsumables(filter LootFilter) ([]BrowseLoot, error) {
	return s.browseLoot(s.catalog.ListConsumables(), s.ListCustomConsumables, filter)
}

// browseLoot merges a catalog table with the GM's homebrew. Homebrew is listed
// first: it has no roll number to sort by, and it's what the GM came looking for.
func (s *Service) browseLoot(
	catalog []cards.Loot,
	custom func(LootFilter) ([]cards.Loot, error),
	filter LootFilter,
) ([]BrowseLoot, error) {
	out := []BrowseLoot{}

	if t := strings.TrimSpace(filter.Table); t == "" || t == homebrewTable {
		rows, err := custom(filter)
		if err != nil {
			return nil, err
		}
		for _, l := range rows {
			if filter.matches(l) {
				out = append(out, BrowseLoot{Loot: l, Source: rules.SourceCustom})
			}
		}
	}

	for _, l := range catalog {
		if filter.matches(l) {
			out = append(out, BrowseLoot{Loot: l, Source: rules.SourceSRD})
		}
	}
	return out, nil
}

func (s *Service) GetItem(slug string) (BrowseLoot, error) {
	return s.getLoot(string(cards.KindItem), slug, s.catalog.Item)
}

func (s *Service) GetConsumable(slug string) (BrowseLoot, error) {
	return s.getLoot(string(cards.KindConsumable), slug, s.catalog.Consumable)
}

func (s *Service) getLoot(kind, slug string, fromCatalog func(string) (cards.Loot, bool)) (BrowseLoot, error) {
	if l, err := s.getCustomLoot(kind, slug); err == nil {
		return BrowseLoot{Loot: l, Source: rules.SourceCustom}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return BrowseLoot{}, err
	}
	if l, ok := fromCatalog(slug); ok {
		return BrowseLoot{Loot: l, Source: rules.SourceSRD}, nil
	}
	return BrowseLoot{}, notFound(kind, slug)
}

// lootTables are the source books present in the catalog. The frontend mirrors
// this set the way it mirrors validate.go's — see LOOT_TABLES in api.js.
func (s *Service) lootTables() []string { return srd.LootTables(s.catalog.ListItems()) }

// maxLootPerCategory keeps a mistyped count from generating an unbounded haul;
// no table hands out more than a handful at once in play.
const maxLootPerCategory = 20

// LootRequest is one "what does the party find?" roll. Tier and rarity are
// separate axes on purpose: weapons and armor are drawn from a tier's tables,
// while items and consumables have no tier at all and are drawn by rolling the
// d12s their rarity calls for. A category with a count of 0 is skipped, so the
// same request covers "just a consumable" and "a full hoard".
//
// IncludeHomebrew is off by default so a haul is reproducible from the SRD alone
// unless the GM asks for their own cards to be in the mix.
type LootRequest struct {
	Tier            string `json:"tier"`
	Rarity          string `json:"rarity"`
	Dice            int    `json:"dice"`
	Table           string `json:"table"`
	IncludeHomebrew bool   `json:"includeHomebrew"`
	Weapons         int    `json:"weapons"`
	Armor           int    `json:"armor"`
	Items           int    `json:"items"`
	Consumables     int    `json:"consumables"`
}

// LootRoll records how an entry was arrived at, not just which entry it was, so
// the GM can read the roll back to the table the way they would at the table.
// Homebrew entries have no roll number to land on, so they arrive with Dice empty
// and Total zero — Source is how the UI tells the two apart.
type LootRoll struct {
	Dice   []int      `json:"dice"`
	Total  int        `json:"total"`
	Source string     `json:"source"`
	Entry  cards.Loot `json:"entry"`
}

type LootHaul struct {
	Tier        string             `json:"tier"`
	Rarity      string             `json:"rarity"`
	Dice        int                `json:"dice"`
	Weapons     []BrowseWeapon     `json:"weapons"`
	Armor       []BrowseArmorPiece `json:"armor"`
	Items       []LootRoll         `json:"items"`
	Consumables []LootRoll         `json:"consumables"`
}

// LootLookupRequest resolves a total the players rolled with real dice back to a
// table entry, for tables that would rather throw their own d12s than press Roll.
//
// Table is required and has no "both books" option the way a rolled request does:
// the same total names a different entry in each book, so there is nothing to
// resolve until the GM says which one was rolled against.
type LootLookupRequest struct {
	Kind  string `json:"kind"`
	Table string `json:"table"`
	Total int    `json:"total"`
}

// LookupLoot turns a hand-entered total into the entry it lands on. The result is
// the same shape a rolled one has, so the UI lists the two together — but Dice
// comes back empty, because the dice were on the table rather than in the app.
//
// Any row from 1 to 60 is accepted. Whether the total is reachable with the dice
// the chosen rarity calls for is the caller's business to point out: a GM reading
// a physical roll off the table is the authority on what was rolled, and refusing
// it here would only get in their way.
func (s *Service) LookupLoot(req LootLookupRequest) (LootRoll, error) {
	kind, err := validateLootKind(req.Kind)
	if err != nil {
		return LootRoll{}, err
	}

	table := strings.TrimSpace(req.Table)
	if table == "" {
		return LootRoll{}, fmt.Errorf("choose which table the roll was made against")
	}
	if !containsTable(s.lootTables(), table) {
		return LootRoll{}, fmt.Errorf("unknown loot table %q", table)
	}
	if req.Total < 1 || req.Total > rules.LootTableSize {
		return LootRoll{}, fmt.Errorf("a roll is between 1 and %d, got %d", rules.LootTableSize, req.Total)
	}

	rows := s.catalog.ListItems()
	if kind == string(cards.KindConsumable) {
		rows = s.catalog.ListConsumables()
	}
	entry, ok := srd.LootByRoll(rows, table)[req.Total]
	if !ok {
		return LootRoll{}, fmt.Errorf("the %s %s table has no row %d", table, kind, req.Total)
	}
	return LootRoll{Dice: []int{}, Total: req.Total, Source: rules.SourceSRD, Entry: entry}, nil
}

// RollLoot generates a haul. Equipment is sampled without replacement, because
// handing the party the same sword twice is noise rather than a result. Items
// and consumables keep duplicates: a repeated total is a real outcome of rolling
// the table, and two of the same potion is a normal reward.
func (s *Service) RollLoot(req LootRequest) (LootHaul, error) {
	haul := LootHaul{
		Weapons:     []BrowseWeapon{},
		Armor:       []BrowseArmorPiece{},
		Items:       []LootRoll{},
		Consumables: []LootRoll{},
	}

	wantEquipment := req.Weapons > 0 || req.Armor > 0
	wantLoot := req.Items > 0 || req.Consumables > 0
	if !wantEquipment && !wantLoot {
		return LootHaul{}, fmt.Errorf("choose at least one thing to roll for")
	}

	for label, n := range map[string]int{
		"weapons": req.Weapons, "armor": req.Armor,
		"items": req.Items, "consumables": req.Consumables,
	} {
		if n < 0 {
			return LootHaul{}, fmt.Errorf("%s count can't be negative, got %d", label, n)
		}
		if n > maxLootPerCategory {
			return LootHaul{}, fmt.Errorf("can't roll more than %d %s at once, got %d", maxLootPerCategory, label, n)
		}
	}

	if wantEquipment {
		if err := validateTier(req.Tier); err != nil {
			return LootHaul{}, err
		}
		haul.Tier = strings.TrimSpace(req.Tier)

		weapons, err := s.BrowseWeapons(EquipmentFilter{Tier: haul.Tier})
		if err != nil {
			return LootHaul{}, err
		}
		haul.Weapons = sample(dropHomebrew(weapons, req.IncludeHomebrew,
			func(w BrowseWeapon) string { return w.Source }), req.Weapons)

		armor, err := s.BrowseArmor(EquipmentFilter{Tier: haul.Tier})
		if err != nil {
			return LootHaul{}, err
		}
		haul.Armor = sample(dropHomebrew(armor, req.IncludeHomebrew,
			func(a BrowseArmorPiece) string { return a.Source }), req.Armor)
	}

	if wantLoot {
		rarity, ok := rules.Rarity(req.Rarity)
		if !ok {
			return LootHaul{}, fmt.Errorf("rarity must be one of Common, Uncommon, Rare, Legendary, got %q", req.Rarity)
		}
		count := rarity.DiceFor(req.Dice)
		haul.Rarity, haul.Dice = rarity.Name, count

		table := strings.TrimSpace(req.Table)
		if table != "" && !containsTable(s.lootTables(), table) {
			return LootHaul{}, fmt.Errorf("unknown loot table %q", table)
		}

		var err error
		if haul.Items, err = s.rollLootTable(s.catalog.ListItems(), s.ListCustomItems, table, rarity.Name, count, req.Items, req.IncludeHomebrew); err != nil {
			return LootHaul{}, err
		}
		if haul.Consumables, err = s.rollLootTable(s.catalog.ListConsumables(), s.ListCustomConsumables, table, rarity.Name, count, req.Consumables, req.IncludeHomebrew); err != nil {
			return LootHaul{}, err
		}
	}

	return haul, nil
}

// rollLootTable rolls n entries. When no book is named, each draw picks a source
// first, weighted by how many entries that source actually offers — an SRD table
// contributes only the rows its dice can reach (11n+1 totals for n d12s), and the
// homebrew pool contributes however many entries match the rarity. Weighting this
// way stops two homebrew potions from outweighing a whole 60-row book.
func (s *Service) rollLootTable(
	catalog []cards.Loot,
	custom func(LootFilter) ([]cards.Loot, error),
	table, rarity string,
	diceCount, n int,
	includeHomebrew bool,
) ([]LootRoll, error) {
	out := make([]LootRoll, 0, n)
	if n <= 0 {
		return out, nil
	}

	tables := []string{table}
	if table == "" {
		tables = srd.LootTables(catalog)
	} else if table == homebrewTable {
		tables = nil
	}

	var homebrew []cards.Loot
	if includeHomebrew || table == homebrewTable {
		rows, err := custom(LootFilter{Rarity: rarity})
		if err != nil {
			return nil, err
		}
		homebrew = rows
	}

	byTable := make(map[string]map[int]cards.Loot, len(tables))
	for _, t := range tables {
		byTable[t] = srd.LootByRoll(catalog, t)
	}

	reachable := 11*diceCount + 1
	total := len(tables)*reachable + len(homebrew)
	if total == 0 {
		return out, nil
	}

	for i := 0; i < n; i++ {
		pick := rand.Intn(total)
		if pick >= len(tables)*reachable {
			out = append(out, LootRoll{
				Dice:   []int{},
				Source: rules.SourceCustom,
				Entry:  homebrew[pick-len(tables)*reachable],
			})
			continue
		}

		chosen := tables[pick/reachable]
		rolled := make([]int, 0, diceCount)
		sum := 0
		for d := 0; d < diceCount; d++ {
			v := dice.RollDice(dice.D12)
			rolled = append(rolled, v)
			sum += v
		}
		entry, ok := byTable[chosen][sum]
		if !ok {
			continue
		}
		out = append(out, LootRoll{Dice: rolled, Total: sum, Source: rules.SourceSRD, Entry: entry})
	}
	return out, nil
}

// dropHomebrew filters a merged equipment list back down to the SRD when the GM
// hasn't opted their own cards into the roll.
func dropHomebrew[T any](pool []T, include bool, source func(T) string) []T {
	if include {
		return pool
	}
	out := make([]T, 0, len(pool))
	for _, item := range pool {
		if source(item) == rules.SourceSRD {
			out = append(out, item)
		}
	}
	return out
}

func containsTable(tables []string, want string) bool {
	for _, t := range tables {
		if t == want {
			return true
		}
	}
	return false
}

// sample draws n distinct entries without disturbing the caller's slice.
func sample[T any](pool []T, n int) []T {
	out := make([]T, 0, n)
	if n <= 0 || len(pool) == 0 {
		return out
	}
	if n > len(pool) {
		n = len(pool)
	}
	for _, i := range rand.Perm(len(pool))[:n] {
		out = append(out, pool[i])
	}
	return out
}
