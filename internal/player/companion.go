package player

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/dice"
)

const (
	CompanionSubclass = "beastbound"
	companionSlug     = "rangerCompanion"
)

var companionDamageDice = []string{"d4", "d6", "d8", "d10", "d12"}

var companionRanges = []string{"Melee", "Very Close", "Close", "Far", "Very Far"}

type CompanionExperience struct {
	Name     string `json:"name"`
	Modifier int    `json:"modifier"`
}

type Companion struct {
	ID           int64                 `json:"id"`
	CharacterID  int64                 `json:"characterId"`
	Name         string                `json:"name"`
	Evasion      int                   `json:"evasion"`
	DamageDie    string                `json:"damageDie"`
	AttackRange  string                `json:"attackRange"`
	Attack       string                `json:"attack"`
	StressMax    int                   `json:"stressMax"`
	StressMarked int                   `json:"stressMarked"`
	Experiences  []CompanionExperience `json:"experiences"`
	Upgrades     []string              `json:"upgrades"`
	Notes        string                `json:"notes"`
}

type CompanionView struct {
	Eligible    bool             `json:"eligible"`
	Companion   *Companion       `json:"companion"`
	Reference   *cards.Companion `json:"reference"`
	Proficiency int              `json:"proficiency"`
	DamageDice  []string         `json:"damageDice"`
	Ranges      []string         `json:"ranges"`
}

type CompanionInput struct {
	CharacterID int64                 `json:"characterId"`
	Name        string                `json:"name"`
	Evasion     int                   `json:"evasion"`
	DamageDie   string                `json:"damageDie"`
	AttackRange string                `json:"attackRange"`
	Attack      string                `json:"attack"`
	StressMax   int                   `json:"stressMax"`
	Experiences []CompanionExperience `json:"experiences"`
	Upgrades    []string              `json:"upgrades"`
	Notes       string                `json:"notes"`
}

func companionView(r db.Companion) Companion {
	c := Companion{
		ID:           r.ID,
		CharacterID:  r.CharacterID,
		Name:         r.Name,
		Evasion:      int(r.Evasion),
		DamageDie:    r.DamageDie,
		AttackRange:  r.AttackRange,
		Attack:       r.Attack,
		StressMax:    int(r.StressMax),
		StressMarked: int(r.StressMarked),
		Experiences:  []CompanionExperience{},
		Upgrades:     []string{},
		Notes:        r.Notes,
	}
	if err := json.Unmarshal([]byte(r.Experiences), &c.Experiences); err != nil || c.Experiences == nil {
		c.Experiences = []CompanionExperience{}
	}
	if err := json.Unmarshal([]byte(r.Upgrades), &c.Upgrades); err != nil || c.Upgrades == nil {
		c.Upgrades = []string{}
	}
	return c
}

func (s *Service) GetCompanion(characterID int64) (CompanionView, error) {
	row, err := s.character(characterID)
	if err != nil {
		return CompanionView{}, err
	}

	view := CompanionView{
		Eligible:    hasCompanion(row),
		Proficiency: int(row.Proficiency),
		DamageDice:  slices.Clone(companionDamageDice),
		Ranges:      slices.Clone(companionRanges),
	}
	if ref, ok := s.catalog.Companion(companionSlug); ok {
		view.Reference = &ref
	}

	companion, err := s.q.GetCompanion(s.ctx, characterID)
	if errors.Is(err, sql.ErrNoRows) {
		return view, nil
	}
	if err != nil {
		return CompanionView{}, fmt.Errorf("loading companion: %w", err)
	}
	c := companionView(companion)
	view.Companion = &c
	return view, nil
}

func (s *Service) SaveCompanion(in CompanionInput) (CompanionView, error) {
	row, err := s.character(in.CharacterID)
	if err != nil {
		return CompanionView{}, err
	}
	if !hasCompanion(row) {
		return CompanionView{}, fmt.Errorf("only a Beastbound ranger takes a companion sheet")
	}

	name, err := validateName(in.Name)
	if err != nil {
		return CompanionView{}, err
	}
	die := strings.TrimSpace(in.DamageDie)
	if !slices.Contains(companionDamageDice, die) {
		return CompanionView{}, fmt.Errorf("damage die must be one of %s, got %q", strings.Join(companionDamageDice, ", "), in.DamageDie)
	}
	rng := strings.TrimSpace(in.AttackRange)
	if !slices.Contains(companionRanges, rng) {
		return CompanionView{}, fmt.Errorf("range must be one of %s, got %q", strings.Join(companionRanges, ", "), in.AttackRange)
	}

	experiences := []CompanionExperience{}
	for _, e := range in.Experiences {
		e.Name = strings.TrimSpace(e.Name)
		if e.Name == "" {
			continue
		}
		e.Modifier = clamp(e.Modifier, 0, 9)
		experiences = append(experiences, e)
	}
	upgrades := []string{}
	for _, u := range in.Upgrades {
		u = strings.TrimSpace(u)
		if u != "" && !slices.Contains(upgrades, u) {
			upgrades = append(upgrades, u)
		}
	}

	experiencesJSON, err := json.Marshal(experiences)
	if err != nil {
		return CompanionView{}, fmt.Errorf("saving companion: %w", err)
	}
	upgradesJSON, err := json.Marshal(upgrades)
	if err != nil {
		return CompanionView{}, fmt.Errorf("saving companion: %w", err)
	}

	evasion := int64(clamp(in.Evasion, 0, 30))
	stressMax := int64(clamp(in.StressMax, 1, 12))
	if in.StressMax == 0 {
		stressMax = 5
	}

	_, err = s.q.UpdateCompanion(s.ctx, db.UpdateCompanionParams{
		Name:        name,
		Evasion:     evasion,
		DamageDie:   die,
		AttackRange: rng,
		Attack:      strings.TrimSpace(in.Attack),
		StressMax:   stressMax,
		Experiences: string(experiencesJSON),
		Upgrades:    string(upgradesJSON),
		Notes:       strings.TrimSpace(in.Notes),
		CharacterID: in.CharacterID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		_, err = s.q.CreateCompanion(s.ctx, db.CreateCompanionParams{
			CharacterID: in.CharacterID,
			Name:        name,
			Evasion:     evasion,
			DamageDie:   die,
			AttackRange: rng,
			Attack:      strings.TrimSpace(in.Attack),
			StressMax:   stressMax,
			Experiences: string(experiencesJSON),
			Upgrades:    string(upgradesJSON),
			Notes:       strings.TrimSpace(in.Notes),
		})
	}
	if err != nil {
		return CompanionView{}, fmt.Errorf("saving companion: %w", err)
	}
	return s.GetCompanion(in.CharacterID)
}

func (s *Service) MarkCompanionStress(characterID int64, delta int) (CompanionView, error) {
	_, err := s.q.AdjustCompanionStress(s.ctx, db.AdjustCompanionStressParams{
		Delta: int64(delta), CharacterID: characterID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return CompanionView{}, notFound("companion", fmt.Sprint(characterID))
	}
	if err != nil {
		return CompanionView{}, fmt.Errorf("marking companion stress: %w", err)
	}
	return s.GetCompanion(characterID)
}

func (s *Service) SetCompanionStress(characterID int64, value int) (CompanionView, error) {
	_, err := s.q.SetCompanionStress(s.ctx, db.SetCompanionStressParams{
		StressMarked: int64(value), CharacterID: characterID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return CompanionView{}, notFound("companion", fmt.Sprint(characterID))
	}
	if err != nil {
		return CompanionView{}, fmt.Errorf("setting companion stress: %w", err)
	}
	return s.GetCompanion(characterID)
}

func (s *Service) ToggleCompanionUpgrade(characterID int64, title string) (CompanionView, error) {
	view, err := s.GetCompanion(characterID)
	if err != nil {
		return CompanionView{}, err
	}
	if view.Companion == nil {
		return CompanionView{}, fmt.Errorf("build the companion sheet first")
	}

	title = strings.TrimSpace(title)
	upgrades := view.Companion.Upgrades
	if slices.Contains(upgrades, title) {
		upgrades = slices.DeleteFunc(slices.Clone(upgrades), func(u string) bool { return u == title })
	} else {
		upgrades = append(slices.Clone(upgrades), title)
	}

	return s.SaveCompanion(CompanionInput{
		CharacterID: characterID,
		Name:        view.Companion.Name,
		Evasion:     view.Companion.Evasion,
		DamageDie:   view.Companion.DamageDie,
		AttackRange: view.Companion.AttackRange,
		Attack:      view.Companion.Attack,
		StressMax:   view.Companion.StressMax,
		Experiences: view.Companion.Experiences,
		Upgrades:    upgrades,
		Notes:       view.Companion.Notes,
	})
}

func (s *Service) DeleteCompanion(characterID int64) (CompanionView, error) {
	if err := s.q.DeleteCompanion(s.ctx, characterID); err != nil {
		return CompanionView{}, fmt.Errorf("deleting companion: %w", err)
	}
	return s.GetCompanion(characterID)
}

func (s *Service) RollCompanionDamage(characterID int64, bonus int, critical bool) (DamageResult, error) {
	view, err := s.GetCompanion(characterID)
	if err != nil {
		return DamageResult{}, err
	}
	if view.Companion == nil {
		return DamageResult{}, fmt.Errorf("build the companion sheet first")
	}
	sides, err := dice.ParseSides(leadingInt(strings.TrimPrefix(view.Companion.DamageDie, "d"), 6))
	if err != nil {
		return DamageResult{}, err
	}
	count := clamp(view.Proficiency, 1, 6)
	modifier := bonus
	if critical {
		modifier += count * int(sides)
	}
	return DamageResult{
		Label:    fmt.Sprintf("%dd%d", count, int(sides)),
		Count:    count,
		Sides:    int(sides),
		Modifier: modifier,
		Total:    dice.RollDamage(count, sides, modifier),
		Critical: critical,
	}, nil
}

func hasCompanion(row db.Character) bool {
	return row.SubclassSlug == CompanionSubclass || row.MulticlassSubclassSlug == CompanionSubclass
}
