package player

import (
	"fmt"
	"strings"

	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/dice"
)

type ModifierPart struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type RollRequest struct {
	CharacterID   int64   `json:"characterId"`
	Trait         string  `json:"trait"`
	ExperienceIDs []int64 `json:"experienceIds"`
	Bonus         int     `json:"bonus"`
	Advantage     bool    `json:"advantage"`
	Disadvantage  bool    `json:"disadvantage"`
	SpendHope     int     `json:"spendHope"`
	Label         string  `json:"label"`
}

type Roll struct {
	Label         string         `json:"label"`
	Hope          int            `json:"hope"`
	Fear          int            `json:"fear"`
	Result        int            `json:"result"`
	Msg           string         `json:"msg"`
	Critical      bool           `json:"critical"`
	WithHope      bool           `json:"withHope"`
	WithFear      bool           `json:"withFear"`
	Modifier      int            `json:"modifier"`
	ModifierParts []ModifierPart `json:"modifierParts"`
	HopeSpent     int            `json:"hopeSpent"`
	HopeGained    int            `json:"hopeGained"`
	StressCleared int            `json:"stressCleared"`
	Advantage     bool           `json:"advantage"`
	Disadvantage  bool           `json:"disadvantage"`
	Character     Character      `json:"character"`
}

type DamageResult struct {
	Label    string `json:"label"`
	Count    int    `json:"count"`
	Sides    int    `json:"sides"`
	Modifier int    `json:"modifier"`
	Total    int    `json:"total"`
	Critical bool   `json:"critical"`
}

func (s *Service) RollDuality(req RollRequest) (Roll, error) {
	row, err := s.character(req.CharacterID)
	if err != nil {
		return Roll{}, err
	}
	if req.SpendHope < 0 {
		req.SpendHope = 0
	}
	if req.SpendHope > int(row.Hope) {
		return Roll{}, fmt.Errorf("you have %d Hope, not %d", row.Hope, req.SpendHope)
	}

	view := s.characterView(row)
	parts := []ModifierPart{}
	modifier := 0

	if trait := strings.TrimSpace(req.Trait); trait != "" {
		key, err := validateTraitKey(trait)
		if err != nil {
			return Roll{}, err
		}
		value := view.Traits[key]
		modifier += value
		parts = append(parts, ModifierPart{Label: titleCase(key), Value: value})
	}

	if len(req.ExperienceIDs) > 0 {
		all, err := s.ListExperiences(req.CharacterID)
		if err != nil {
			return Roll{}, err
		}
		byID := map[int64]Experience{}
		for _, e := range all {
			byID[e.ID] = e
		}
		for _, id := range req.ExperienceIDs {
			e, ok := byID[id]
			if !ok {
				return Roll{}, notFound("experience", fmt.Sprint(id))
			}
			modifier += e.Modifier
			parts = append(parts, ModifierPart{Label: e.Name, Value: e.Modifier})
		}
	}

	if req.Bonus != 0 {
		modifier += req.Bonus
		parts = append(parts, ModifierPart{Label: "Bonus", Value: req.Bonus})
	}

	d := dice.DualityDiceRoll(req.Advantage, req.Disadvantage, modifier)

	roll := Roll{
		Label:         strings.TrimSpace(req.Label),
		Hope:          d.Hope,
		Fear:          d.Fear,
		Result:        d.Result,
		Msg:           d.Msg,
		Critical:      d.Hope == d.Fear,
		Modifier:      modifier,
		ModifierParts: parts,
		HopeSpent:     req.SpendHope,
		Advantage:     req.Advantage,
		Disadvantage:  req.Disadvantage,
	}
	switch {
	case roll.Critical:
		roll.HopeGained = 1
		roll.StressCleared = 1
	case d.Hope > d.Fear:
		roll.WithHope = true
		roll.HopeGained = 1
	default:
		roll.WithFear = true
	}

	hope := clamp(int(row.Hope)-req.SpendHope+roll.HopeGained, 0, HopeMax)
	stress := clamp(int(row.StressMarked)-roll.StressCleared, 0, int(row.StressMax))

	updated, err := s.q.SetCharacterVitals(s.ctx, db.SetCharacterVitalsParams{
		HpMarked:     row.HpMarked,
		StressMarked: int64(stress),
		Hope:         int64(hope),
		ArmorMarked:  row.ArmorMarked,
		ID:           req.CharacterID,
	})
	if err != nil {
		return Roll{}, fmt.Errorf("applying roll outcome: %w", err)
	}
	roll.Character = s.characterView(updated)
	return roll, nil
}

func (s *Service) RollDamage(characterID int64, sides int, bonus int, critical bool) (DamageResult, error) {
	row, err := s.character(characterID)
	if err != nil {
		return DamageResult{}, err
	}
	die, err := dice.ParseSides(sides)
	if err != nil {
		return DamageResult{}, err
	}
	count := clamp(int(row.Proficiency), 1, 6)
	modifier := bonus
	if critical {
		modifier += count * int(die)
	}
	return DamageResult{
		Label:    fmt.Sprintf("%dd%d", count, int(die)),
		Count:    count,
		Sides:    int(die),
		Modifier: modifier,
		Total:    dice.RollDamage(count, die, modifier),
		Critical: critical,
	}, nil
}
