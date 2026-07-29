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
