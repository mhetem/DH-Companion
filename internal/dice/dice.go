package dice

import (
	"math/rand"
)

type DualityDice struct {
	Hope   int
	Fear   int
	Result int
	Msg    string
}

func RollDice(sides int) int {
	return rand.Intn(sides) + 1
}

func DualityDiceRoll(hasAdvantage bool, hasDisadvantage bool, modifier int) DualityDice {
	hope := RollDice(12)
	fear := RollDice(12)
	if hope == fear {
		return DualityDice{Hope: hope, Fear: fear, Result: hope + fear, Msg: "Critical Success!"}
	}
	if hasAdvantage && !hasDisadvantage {
		advDice := RollDice(6)
		return DualityDice{Hope: hope, Fear: fear, Result: hope + fear + advDice + modifier, Msg: ""}
	}
	if hasDisadvantage && !hasAdvantage {
		disDice := RollDice(6)
		return DualityDice{Hope: hope, Fear: fear, Result: hope + fear - disDice + modifier, Msg: ""}
	}
	return DualityDice{Hope: hope, Fear: fear, Result: hope + fear + modifier, Msg: ""}
}
