package rules

import "strings"

// Loot rarity in the SRD is a property of the roll, not of the item: each rarity
// names a number of d12s ("Common (1d12 or 2d12)"), you sum them, and the total
// indexes a 60-row table. Both dice counts are offered, so a rarity is really a
// pair of options — the smaller keeps results clustered low, the larger spreads
// them over the whole band.
const (
	RarityCommon    = "Common"
	RarityUncommon  = "Uncommon"
	RarityRare      = "Rare"
	RarityLegendary = "Legendary"
)

// LootRarity is one rarity band and the d12 counts it can be rolled with.
type LootRarity struct {
	Name     string `json:"name"`
	MinDice  int    `json:"minDice"`
	MaxDice  int    `json:"maxDice"`
	MinRoll  int    `json:"minRoll"`
	MaxRoll  int    `json:"maxRoll"`
	Examples string `json:"examples"`
}

// LootTableSize is the number of rows in every SRD item and consumable table.
const LootTableSize = 60

const lootDie = 12

var lootRarities = []LootRarity{
	{RarityCommon, 1, 2, 1, 24, "Found at an abandoned camp or readily available at a local shop."},
	{RarityUncommon, 2, 3, 2, 36, "Found in limited supply in a shop, kept in a protected place in a camp, or offered as part of a reward for a job."},
	{RarityRare, 3, 4, 3, 48, "Kept under lock and key in a shop, offered as the sole reward for a job, or discovered among a powerful NPC's possessions."},
	{RarityLegendary, 4, 5, 4, 60, "The only item of their kind, a reward for an incredibly difficult or dangerous job, or a powerful adversary's most guarded treasure."},
}

// LootRarities lists the bands in ascending order, for the frontend's picker and
// for validation, so the set is defined only here.
func LootRarities() []LootRarity {
	out := make([]LootRarity, len(lootRarities))
	copy(out, lootRarities)
	return out
}

// Rarity resolves a rarity by name, case-insensitively. An empty name is not a
// rarity: callers that mean "any" check for that before calling.
func Rarity(name string) (LootRarity, bool) {
	for _, r := range lootRarities {
		if strings.EqualFold(r.Name, name) {
			return r, true
		}
	}
	return LootRarity{}, false
}

// RarityFor reports the cheapest rarity that can roll the given table row. A
// band's reach is MaxDice*12, so row 25 is out of Common's reach (2d12 caps at
// 24) and becomes Uncommon.
func RarityFor(roll int) string {
	for _, r := range lootRarities {
		if roll <= r.MaxDice*lootDie {
			return r.Name
		}
	}
	return RarityLegendary
}

// DiceFor picks how many d12s to roll for a rarity. count selects one of the
// band's two options; anything else falls back to the smaller, which is the one
// the SRD prints first.
func (r LootRarity) DiceFor(count int) int {
	if count == r.MaxDice {
		return r.MaxDice
	}
	return r.MinDice
}
