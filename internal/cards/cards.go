package cards

type Kind string

const (
	KindAdversary   Kind = "adversary"
	KindEnvironment Kind = "environment"
	KindDomain      Kind = "domain"
	KindAbility     Kind = "ability"
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
	Common      string `json:"common"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	// Questions are the italic GM prompts printed beneath a feature. Environment
	// cards carry them; adversary cards don't.
	Questions []string `json:"questions,omitempty"`
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
