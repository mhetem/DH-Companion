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

type GMDice struct {
	Result int
	Msg    string
}

func RollDice(sides Sides) int {
	if !sides.Valid() {
		return 0
	}
	return rand.Intn(int(sides)) + 1
}

func DualityDiceRoll(hasAdvantage bool, hasDisadvantage bool, modifier int) DualityDice {
	hope := RollDice(D12)
	fear := RollDice(D12)
	if hope == fear {
		return DualityDice{Hope: hope, Fear: fear, Result: hope + fear, Msg: "Critical Success!"}
	}
	if hasAdvantage && !hasDisadvantage {
		advDice := RollDice(D6)
		return DualityDice{Hope: hope, Fear: fear, Result: hope + fear + advDice + modifier, Msg: ""}
	}
	if hasDisadvantage && !hasAdvantage {
		disDice := RollDice(D6)
		return DualityDice{Hope: hope, Fear: fear, Result: hope + fear - disDice + modifier, Msg: ""}
	}
	return DualityDice{Hope: hope, Fear: fear, Result: hope + fear + modifier, Msg: ""}
}

func RollDamage(numDice int, diceType Sides, modifier int) int {
	if numDice > 1 {
		totalDmgRolled := 0
		for i := 0; i < numDice; i++ {
			roll := RollDice(diceType)
			totalDmgRolled += roll
		}
		return totalDmgRolled + modifier
	}
	return RollDice(diceType) + modifier
}

func RollGMDice(hasAdvantage bool, hasDisadvantage bool, modifier int) GMDice {
	regularRoll := RollDice(D20)
	if hasAdvantage && !hasDisadvantage {
		advDice := RollDice(D20)
		if advDice == 20 || regularRoll == 20 {
			return GMDice{Result: 20, Msg: "Critical!"}
		}
		if advDice >= regularRoll {
			return GMDice{Result: advDice + modifier, Msg: ""}
		}
		if regularRoll > advDice {
			return GMDice{Result: regularRoll + modifier, Msg: ""}
		}
	}
	if hasDisadvantage && !hasAdvantage {
		disDice := RollDice(D20)
		if disDice == 20 && regularRoll == 20 {
			return GMDice{Result: 20, Msg: "Critical!"}
		}
		if disDice <= regularRoll {
			return GMDice{Result: disDice + modifier, Msg: ""}
		}
		if regularRoll < disDice {
			return GMDice{Result: regularRoll + modifier, Msg: ""}
		}
	}
	if regularRoll == 20 {
		return GMDice{Result: 20, Msg: "Critical!"}
	}
	return GMDice{Result: regularRoll + modifier, Msg: ""}
}
