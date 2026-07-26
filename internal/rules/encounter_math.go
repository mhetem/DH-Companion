package rules

import (
	"strconv"
	"strings"
)

func getBudgetForEncounter(numPlayers int) int {
	return (3 * numPlayers) + 2
}

func tierValue(tier string) int {
	n, err := strconv.Atoi(strings.TrimSpace(tier))
	if err != nil {
		return 0
	}
	return n
}

func adversaryCost(advType string) int {
	switch advType {
	case "Social", "Support":
		return 1
	case "Standard", "Horde", "Skulk", "Ranged":
		return 2
	case "Leader":
		return 3
	case "Bruiser":
		return 4
	case "Solo":
		return 5
	default:
		return 2
	}
}

type EncounterSettings struct {
	PartySize  int
	PartyTier  string
	Difficulty string
}

type BudgetSummary struct {
	PartySize   int
	Budget      int
	Spent       int
	Remaining   int
	Over        bool
	Adjustments []string
}

func isBigType(advType string) bool {
	switch advType {
	case "Bruiser", "Horde", "Leader", "Solo":
		return true
	default:
		return false
	}
}

func ComputeBudget(s EncounterSettings, picks []EncounterAdversary) BudgetSummary {
	partySize := s.PartySize
	if partySize < 1 {
		partySize = 1
	}
	partyTier := tierValue(s.PartyTier)

	spent := 0
	minions := 0
	solos := 0
	anyPicked := false
	hasBig := false
	belowTier := false
	for _, p := range picks {
		if p.Count <= 0 {
			continue
		}
		anyPicked = true

		if p.Type == "Minion" {
			minions += p.Count
		} else {
			spent += adversaryCost(p.Type) * p.Count
		}
		if p.Type == "Solo" {
			solos += p.Count
		}
		if isBigType(p.Type) {
			hasBig = true
		}
		if pickTier := tierValue(p.Tier); partyTier > 0 && pickTier > 0 && pickTier < partyTier {
			belowTier = true
		}
	}
	if minions > 0 {
		spent += (minions + partySize - 1) / partySize
	}

	budget := getBudgetForEncounter(partySize)
	var adjustments []string

	switch s.Difficulty {
	case "easy":
		budget--
		adjustments = append(adjustments, "Easy encounter: -1")
	case "hard":
		budget += 2
		adjustments = append(adjustments, "Hard encounter: +2")
	}
	if solos > 1 {
		budget -= 2
		adjustments = append(adjustments, "More than one solo: -2")
	}
	if belowTier {
		budget++
		adjustments = append(adjustments, "Adversary below party tier: +1")
	}
	if anyPicked && !hasBig {
		budget++
		adjustments = append(adjustments, "No bruiser, horde, leader, or solo: +1")
	}

	return BudgetSummary{
		PartySize:   partySize,
		Budget:      budget,
		Spent:       spent,
		Remaining:   budget - spent,
		Over:        spent > budget,
		Adjustments: adjustments,
	}
}
