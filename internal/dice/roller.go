package dice

// Roller is the Wails-bound face of this package. The rolls themselves stay
// pure package-level functions; this type exists only to give them a home the
// frontend can reach at window.go.dice.Roller.*, matching how gm.Service is
// bound.
type Roller struct{}

func NewRoller() *Roller { return &Roller{} }

// DamageRoll echoes what was asked for alongside the total, so the UI can label
// a result "2d6+3" without holding onto the request itself.
type DamageRoll struct {
	Count    int `json:"count"`
	Sides    int `json:"sides"`
	Modifier int `json:"modifier"`
	Total    int `json:"total"`
}

// Duality is the player-side roll: 2d12 read as Hope and Fear.
func (Roller) Duality(hasAdvantage bool, hasDisadvantage bool, modifier int) DualityDice {
	return DualityDiceRoll(hasAdvantage, hasDisadvantage, modifier)
}

// GM is the GM-side roll: a single d20, since only players roll duality dice.
func (Roller) GM(hasAdvantage bool, hasDisadvantage bool, modifier int) GMDice {
	return RollGMDice(hasAdvantage, hasDisadvantage, modifier)
}

// Sizes lists the rollable die sizes for the frontend's picker, so the set is
// defined once here rather than duplicated in JS.
func (Roller) Sizes() []int {
	allowed := Allowed()
	out := make([]int, 0, len(allowed))
	for _, s := range allowed {
		out = append(out, int(s))
	}
	return out
}

// Damage rolls count dice of the given size. Arguments arrive from JS as plain
// numbers, so this is where they're checked: an unrollable die is an error the
// UI can show, and count is clamped to at least one.
func (Roller) Damage(count int, sides int, modifier int) (DamageRoll, error) {
	die, err := ParseSides(sides)
	if err != nil {
		return DamageRoll{}, err
	}
	if count < 1 {
		count = 1
	}
	return DamageRoll{
		Count:    count,
		Sides:    int(die),
		Modifier: modifier,
		Total:    RollDamage(count, die, modifier),
	}, nil
}
