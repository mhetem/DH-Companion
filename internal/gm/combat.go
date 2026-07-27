package gm

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

const FearMax = 12

const statFallback = 1

type CombatantView struct {
	ID            int64            `json:"id"`
	CombatID      int64            `json:"combatId"`
	DisplayName   string           `json:"displayName"`
	AdversarySlug *string          `json:"adversarySlug"`
	HpMax         int              `json:"hpMax"`
	HpMarked      int              `json:"hpMarked"`
	StressMax     int              `json:"stressMax"`
	StressMarked  int              `json:"stressMarked"`
	Spotlight     bool             `json:"spotlight"`
	Adversary     *cards.Adversary `json:"adversary"`
	Source        string           `json:"source"`
	Unresolved    bool             `json:"unresolved,omitempty"`
	CreatedAt     string           `json:"createdAt"`
	UpdatedAt     string           `json:"updatedAt"`
}

type CombatView struct {
	ID            int64              `json:"id"`
	EncounterID   *int64             `json:"encounterId"`
	EncounterName string             `json:"encounterName"`
	Fear          int                `json:"fear"`
	FearMax       int                `json:"fearMax"`
	Active        bool               `json:"active"`
	Combatants    []CombatantView    `json:"combatants"`
	Environment   *cards.Environment `json:"environment"`
	CreatedAt     string             `json:"createdAt"`
	UpdatedAt     string             `json:"updatedAt"`
}

type CombatSummary struct {
	ID            int64  `json:"id"`
	EncounterID   *int64 `json:"encounterId"`
	EncounterName string `json:"encounterName"`
	Fear          int    `json:"fear"`
	Active        bool   `json:"active"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type CombatantInput struct {
	ID            *int64  `json:"id"`
	CombatID      int64   `json:"combatId"`
	DisplayName   string  `json:"displayName"`
	AdversarySlug *string `json:"adversarySlug"`
	HpMax         int     `json:"hpMax"`
	StressMax     int     `json:"stressMax"`
}

func (in CombatantInput) validate() (CombatantInput, error) {
	name, err := validateName(in.DisplayName)
	if err != nil {
		return in, err
	}
	in.DisplayName = name
	if in.HpMax < 0 || in.StressMax < 0 {
		return in, fmt.Errorf("hp and stress maximums can't be negative")
	}
	return in, nil
}

func parseStat(raw string) int64 {
	raw = strings.TrimSpace(raw)
	end := 0
	for end < len(raw) && raw[end] >= '0' && raw[end] <= '9' {
		end++
	}
	if end == 0 {
		return statFallback
	}
	n := int64(0)
	for i := 0; i < end; i++ {
		n = n*10 + int64(raw[i]-'0')
	}
	if n < 0 {
		return statFallback
	}
	return n
}

func spawnNames(name string, count int) []string {
	if count <= 1 {
		return []string{name}
	}
	out := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		out = append(out, fmt.Sprintf("%s %d", name, i))
	}
	return out
}

func (s *Service) StartCombat(encounterID int64) (CombatView, error) {
	view, err := s.GetEncounter(encounterID)
	if err != nil {
		return CombatView{}, err
	}

	var row db.Combat
	err = s.tx(func(q *db.Queries) error {
		if err := q.DeactivateAllCombats(s.ctx); err != nil {
			return fmt.Errorf("standing down the previous combat: %w", err)
		}
		var err error
		row, err = q.CreateCombat(s.ctx, sql.NullInt64{Int64: encounterID, Valid: true})
		if err != nil {
			return fmt.Errorf("starting combat: %w", err)
		}
		for _, pick := range view.Adversaries {
			slug := pick.Slug
			hp, stress := parseStat(pick.Hp), parseStat(pick.Stress)
			for _, name := range spawnNames(pick.Name, pick.Count) {
				if _, err := q.CreateCombatant(s.ctx, db.CreateCombatantParams{
					CombatID:      row.ID,
					AdversarySlug: sql.NullString{String: slug, Valid: slug != ""},
					DisplayName:   name,
					HpMax:         hp,
					StressMax:     stress,
				}); err != nil {
					return fmt.Errorf("spawning %q: %w", name, err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return CombatView{}, err
	}
	return s.hydrateCombat(row)
}

func (s *Service) GetCombat(id int64) (CombatView, error) {
	row, err := s.q.GetCombat(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return CombatView{}, notFound("combat", fmt.Sprint(id))
	}
	if err != nil {
		return CombatView{}, fmt.Errorf("loading combat: %w", err)
	}
	return s.hydrateCombat(row)
}

func (s *Service) GetActiveCombat() (*CombatView, error) {
	row, err := s.q.GetActiveCombat(s.ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("loading active combat: %w", err)
	}
	view, err := s.hydrateCombat(row)
	if err != nil {
		return nil, err
	}
	return &view, nil
}

func (s *Service) ListCombats() ([]CombatSummary, error) {
	rows, err := s.q.ListCombats(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing combats: %w", err)
	}
	out := make([]CombatSummary, 0, len(rows))
	for _, r := range rows {
		out = append(out, CombatSummary{
			ID:            r.ID,
			EncounterID:   int64Ptr(r.EncounterID),
			EncounterName: s.encounterName(r.EncounterID),
			Fear:          int(r.Fear),
			Active:        r.Active == 1,
			CreatedAt:     r.CreatedAt,
			UpdatedAt:     r.UpdatedAt,
		})
	}
	return out, nil
}

func (s *Service) EndCombat(id int64) (CombatView, error) {
	row, err := s.q.SetCombatActive(s.ctx, db.SetCombatActiveParams{Active: 0, ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		return CombatView{}, notFound("combat", fmt.Sprint(id))
	}
	if err != nil {
		return CombatView{}, fmt.Errorf("ending combat: %w", err)
	}
	return s.hydrateCombat(row)
}

func (s *Service) ResumeCombat(id int64) (CombatView, error) {
	var row db.Combat
	err := s.tx(func(q *db.Queries) error {
		if err := q.DeactivateAllCombats(s.ctx); err != nil {
			return fmt.Errorf("standing down the previous combat: %w", err)
		}
		var err error
		row, err = q.SetCombatActive(s.ctx, db.SetCombatActiveParams{Active: 1, ID: id})
		if errors.Is(err, sql.ErrNoRows) {
			return notFound("combat", fmt.Sprint(id))
		}
		if err != nil {
			return fmt.Errorf("resuming combat: %w", err)
		}
		return nil
	})
	if err != nil {
		return CombatView{}, err
	}
	return s.hydrateCombat(row)
}

func (s *Service) DeleteCombat(id int64) error {
	if err := s.q.DeleteCombat(s.ctx, id); err != nil {
		return fmt.Errorf("deleting combat: %w", err)
	}
	return nil
}

func (s *Service) AdjustFear(combatID int64, delta int) (int, error) {
	row, err := s.q.AdjustFear(s.ctx, db.AdjustFearParams{Delta: int64(delta), ID: combatID})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, notFound("combat", fmt.Sprint(combatID))
	}
	if err != nil {
		return 0, fmt.Errorf("adjusting fear: %w", err)
	}
	return int(row.Fear), nil
}

func (s *Service) SetFear(combatID int64, value int) (int, error) {
	row, err := s.q.SetFear(s.ctx, db.SetFearParams{Fear: int64(value), ID: combatID})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, notFound("combat", fmt.Sprint(combatID))
	}
	if err != nil {
		return 0, fmt.Errorf("setting fear: %w", err)
	}
	return int(row.Fear), nil
}

func (s *Service) SaveCombatant(in CombatantInput) (CombatantView, error) {
	in, err := in.validate()
	if err != nil {
		return CombatantView{}, err
	}

	var row db.Combatant
	if in.ID == nil {
		row, err = s.q.CreateCombatant(s.ctx, db.CreateCombatantParams{
			CombatID:      in.CombatID,
			AdversarySlug: nullString(in.AdversarySlug),
			DisplayName:   in.DisplayName,
			HpMax:         int64(in.HpMax),
			StressMax:     int64(in.StressMax),
		})
		if err != nil {
			return CombatantView{}, fmt.Errorf("adding combatant: %w", err)
		}
	} else {
		row, err = s.q.UpdateCombatant(s.ctx, db.UpdateCombatantParams{
			DisplayName: in.DisplayName,
			HpMax:       int64(in.HpMax),
			StressMax:   int64(in.StressMax),
			ID:          *in.ID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return CombatantView{}, notFound("combatant", fmt.Sprint(*in.ID))
		}
		if err != nil {
			return CombatantView{}, fmt.Errorf("updating combatant: %w", err)
		}
	}
	return s.resolveCombatant(row), nil
}

func (s *Service) MarkHP(combatantID int64, delta int) (CombatantView, error) {
	row, err := s.q.AdjustCombatantHP(s.ctx, db.AdjustCombatantHPParams{
		Delta: int64(delta), ID: combatantID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return CombatantView{}, notFound("combatant", fmt.Sprint(combatantID))
	}
	if err != nil {
		return CombatantView{}, fmt.Errorf("marking hp: %w", err)
	}
	return s.resolveCombatant(row), nil
}

func (s *Service) MarkStress(combatantID int64, delta int) (CombatantView, error) {
	row, err := s.q.AdjustCombatantStress(s.ctx, db.AdjustCombatantStressParams{
		Delta: int64(delta), ID: combatantID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return CombatantView{}, notFound("combatant", fmt.Sprint(combatantID))
	}
	if err != nil {
		return CombatantView{}, fmt.Errorf("marking stress: %w", err)
	}
	return s.resolveCombatant(row), nil
}

func (s *Service) SetVitals(combatantID int64, hpMarked, stressMarked int) (CombatantView, error) {
	row, err := s.q.SetCombatantVitals(s.ctx, db.SetCombatantVitalsParams{
		HpMarked: int64(hpMarked), StressMarked: int64(stressMarked), ID: combatantID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return CombatantView{}, notFound("combatant", fmt.Sprint(combatantID))
	}
	if err != nil {
		return CombatantView{}, fmt.Errorf("setting vitals: %w", err)
	}
	return s.resolveCombatant(row), nil
}

func (s *Service) SetSpotlight(combatantID int64, on bool) (CombatantView, error) {
	var flag int64
	if on {
		flag = 1
	}
	row, err := s.q.SetCombatantSpotlight(s.ctx, db.SetCombatantSpotlightParams{
		Spotlight: flag, ID: combatantID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return CombatantView{}, notFound("combatant", fmt.Sprint(combatantID))
	}
	if err != nil {
		return CombatantView{}, fmt.Errorf("setting spotlight: %w", err)
	}
	return s.resolveCombatant(row), nil
}

func (s *Service) ClearSpotlights(combatID int64) error {
	if err := s.q.ClearCombatantSpotlights(s.ctx, combatID); err != nil {
		return fmt.Errorf("clearing spotlights: %w", err)
	}
	return nil
}

func (s *Service) RemoveCombatant(id int64) error {
	if err := s.q.DeleteCombatant(s.ctx, id); err != nil {
		return fmt.Errorf("removing combatant: %w", err)
	}
	return nil
}

func (s *Service) hydrateCombat(row db.Combat) (CombatView, error) {
	rows, err := s.q.ListCombatants(s.ctx, row.ID)
	if err != nil {
		return CombatView{}, fmt.Errorf("loading combatants: %w", err)
	}

	view := CombatView{
		ID:            row.ID,
		EncounterID:   int64Ptr(row.EncounterID),
		EncounterName: s.encounterName(row.EncounterID),
		Fear:          int(row.Fear),
		FearMax:       FearMax,
		Active:        row.Active == 1,
		Combatants:    make([]CombatantView, 0, len(rows)),
		Environment:   s.combatEnvironment(row.EncounterID),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	for _, r := range rows {
		view.Combatants = append(view.Combatants, s.resolveCombatant(r))
	}
	return view, nil
}

func (s *Service) resolveCombatant(r db.Combatant) CombatantView {
	view := CombatantView{
		ID:            r.ID,
		CombatID:      r.CombatID,
		DisplayName:   r.DisplayName,
		AdversarySlug: stringPtr(r.AdversarySlug),
		HpMax:         int(r.HpMax),
		HpMarked:      int(r.HpMarked),
		StressMax:     int(r.StressMax),
		StressMarked:  int(r.StressMarked),
		Spotlight:     r.Spotlight == 1,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
	if !r.AdversarySlug.Valid || r.AdversarySlug.String == "" {
		return view
	}
	card, source, ok := s.lookupAdversary(r.AdversarySlug.String)
	if !ok {
		view.Unresolved = true
		return view
	}
	view.Adversary, view.Source = &card, source
	return view
}

func (s *Service) encounterName(id sql.NullInt64) string {
	if !id.Valid {
		return ""
	}
	row, err := s.q.GetEncounter(s.ctx, id.Int64)
	if err != nil {
		return ""
	}
	return row.EncounterName
}

func (s *Service) combatEnvironment(id sql.NullInt64) *cards.Environment {
	if !id.Valid {
		return nil
	}
	row, err := s.q.GetEncounter(s.ctx, id.Int64)
	if err != nil || !row.EnvironmentSlug.Valid {
		return nil
	}
	env, _, ok := s.lookupEnvironment(row.EnvironmentSlug.String)
	if !ok {
		return nil
	}
	return &env
}
