package cards

type Kind string

const (
	KindAdversary   Kind = "adversary"
	KindEnvironment Kind = "environment"
	KindDomain      Kind = "domain"
	KindAbility     Kind = "ability"
	KindAncestry    Kind = "ancestry"
	KindCommunity   Kind = "community"
	KindClass       Kind = "class"
	KindBeastform   Kind = "beastform"
	KindCompanion   Kind = "companion"
	KindWeapon      Kind = "weapon"
	KindArmor       Kind = "armor"
	KindItem        Kind = "item"
	KindConsumable  Kind = "consumable"
)

// Weapon.Category values. A character equips at most one of each at a time.
const (
	CategoryPrimary   = "Primary"
	CategorySecondary = "Secondary"
)

// Feature.Type values used by class data. Subclass features are printed in these
// three groups, unlocked as the character advances.
const (
	FeatureFoundation     = "Foundation"
	FeatureSpecialization = "Specialization"
	FeatureMastery        = "Mastery"
)

type Meta struct {
	Kind        Kind   `json:"kind"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Tier        string `json:"tier"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func (m Meta) CardMeta() Meta { return m }

type Card interface {
	CardMeta() Meta
}

type Feature struct {
	Common      string   `json:"common"`
	Title       string   `json:"title"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Questions   []string `json:"questions,omitempty"`
}

type Attack struct {
	Modifier   string `json:"modifier"`
	Name       string `json:"name"`
	Range      string `json:"range"`
	Damage     string `json:"damage"`
	DamageType string `json:"damageType"`
}

type Adversary struct {
	Meta
	HordeNumber    string    `json:"hordeNumber"`
	Motives        string    `json:"motives"`
	Experiences    string    `json:"experiences"`
	Difficulty     string    `json:"difficulty"`
	ThresholdMinor string    `json:"thresholdMinor"`
	ThresholdMajor string    `json:"thresholdMajor"`
	Hp             string    `json:"hp"`
	Stress         string    `json:"stress"`
	StandardAttack Attack    `json:"standardAttack"`
	Features       []Feature `json:"features"`
}

type Environment struct {
	Meta
	Difficulty           string    `json:"difficulty"`
	Impulses             string    `json:"impulses"`
	PotentialAdversaries []string  `json:"potentialAdversaries"`
	Features             []Feature `json:"features"`
}

type DomainCard struct {
	Meta
	Domain     string `json:"domain"`
	Level      string `json:"level"`
	RecallCost string `json:"recallCost"`
}

type Ancestry struct {
	Meta
	Features []Feature `json:"features"`
}

type Community struct {
	Meta
	Adjectives []string  `json:"adjectives"`
	Features   []Feature `json:"features"`
}

type Subclass struct {
	Slug           string    `json:"slug"`
	Name           string    `json:"name"`
	Tagline        string    `json:"tagline"`
	SpellcastTrait string    `json:"spellcastTrait"`
	Features       []Feature `json:"features"`
}

type CharacterClass struct {
	Meta
	Domains             []string   `json:"domains"`
	StartingEvasion     string     `json:"startingEvasion"`
	StartingHitPoints   string     `json:"startingHitPoints"`
	ClassItems          []string   `json:"classItems"`
	HopeFeature         Feature    `json:"hopeFeature"`
	Features            []Feature  `json:"features"`
	Subclasses          []Subclass `json:"subclasses"`
	BackgroundQuestions []string   `json:"backgroundQuestions"`
	Connections         []string   `json:"connections"`
}

func (c CharacterClass) Subclass(slug string) (Subclass, bool) {
	for _, s := range c.Subclasses {
		if s.Slug == slug {
			return s, true
		}
	}
	return Subclass{}, false
}

type BeastformAttack struct {
	Range      string `json:"range"`
	Trait      string `json:"trait"`
	Damage     string `json:"damage"`
	DamageType string `json:"damageType"`
}

type Beastform struct {
	Meta
	Examples     []string        `json:"examples"`
	Trait        string          `json:"trait"`
	TraitBonus   string          `json:"traitBonus"`
	EvasionBonus string          `json:"evasionBonus"`
	Attack       BeastformAttack `json:"attack"`
	Advantages   []string        `json:"advantages"`
	Features     []Feature       `json:"features"`
}

type Companion struct {
	Meta
	StartingEvasion            string    `json:"startingEvasion"`
	StartingDamageDie          string    `json:"startingDamageDie"`
	StartingRange              string    `json:"startingRange"`
	StartingExperienceModifier string    `json:"startingExperienceModifier"`
	Setup                      []Feature `json:"setup"`
	ExampleExperiences         []string  `json:"exampleExperiences"`
	Rules                      []Feature `json:"rules"`
	LevelUpOptions             []Feature `json:"levelUpOptions"`
}

// Weapon covers the primary, secondary and combat wheelchair tables. Type is
// "Physical" or "Magic" so it lines up with the browse filter; DamageType is the
// SRD's own abbreviation ("phy", "mag", "phy/mag"), which is what the tables print.
type Weapon struct {
	Meta
	Category   string   `json:"category"`
	Trait      string   `json:"trait"`
	Range      string   `json:"range"`
	Damage     string   `json:"damage"`
	DamageType string   `json:"damageType"`
	Burden     string   `json:"burden"`
	Feature    *Feature `json:"feature"`
}

// Armor's thresholds are the base Major/Severe pair before the wearer's level is
// added, and BaseScore is how many Armor Slots it grants before bonuses.
type Armor struct {
	Meta
	ThresholdMajor  int      `json:"thresholdMajor"`
	ThresholdSevere int      `json:"thresholdSevere"`
	BaseScore       int      `json:"baseScore"`
	Feature         *Feature `json:"feature"`
}

// Loot is one row of an item or consumable table. Items and consumables have the
// same shape and differ only in Kind and Type, so they share this type.
//
// Roll is the row's number in its own 60-row table, which is what the d12 sum
// indexes. Rarity is derived from Roll rather than printed by the SRD: rarity
// there describes how many d12s you roll, so an entry's rarity is the cheapest
// one whose dice can still reach it. See rules.RarityFor.
type Loot struct {
	Meta
	Roll   int    `json:"roll"`
	Table  string `json:"table"`
	Rarity string `json:"rarity"`
}
