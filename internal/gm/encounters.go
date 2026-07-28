package gm

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/rules"
)

type Pick struct {
	Slug  string `json:"slug"`
	Count int    `json:"count"`
}

type EncounterInput struct {
	ID                *int64  `json:"id"`
	Name              string  `json:"name"`
	PartyID           *int64  `json:"partyId"`
	EnvironmentSlug   *string `json:"environmentSlug"`
	Adversaries       []Pick  `json:"adversaries"`
	CustomAdversaries []Pick  `json:"customAdversaries"`
}

type EncounterSummary struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	PartyID         *int64  `json:"partyId"`
	EnvironmentSlug *string `json:"environmentSlug"`
	TotalCount      int     `json:"totalCount"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

func decodePicks(raw string) ([]Pick, error) {
	if raw == "" {
		return nil, nil
	}
	var picks []Pick
	if err := json.Unmarshal([]byte(raw), &picks); err != nil {
		return nil, fmt.Errorf("unreadable adversary picks: %w", err)
	}
	return picks, nil
}

func encodePicks(picks []Pick) (string, error) {
	clean := make([]Pick, 0, len(picks))
	for _, p := range picks {
		if p.Slug == "" || p.Count <= 0 {
			continue
		}
		clean = append(clean, p)
	}
	b, err := json.Marshal(clean)
	if err != nil {
		return "", fmt.Errorf("encoding adversary picks: %w", err)
	}
	return string(b), nil
}

func (in EncounterInput) validate() (EncounterInput, error) {
	name, err := validateName(in.Name)
	if err != nil {
		return in, err
	}
	in.Name = name
	return in, nil
}

func (s *Service) ListEncounters() ([]EncounterSummary, error) {
	rows, err := s.q.ShowAllEncounters(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing encounters: %w", err)
	}
	out := make([]EncounterSummary, 0, len(rows))
	for _, r := range rows {
		summary, err := encounterSummary(r)
		if err != nil {
			return nil, err
		}
		out = append(out, summary)
	}
	return out, nil
}

func encounterSummary(r db.Encounter) (EncounterSummary, error) {
	srdPicks, err := decodePicks(r.Adversaries)
	if err != nil {
		return EncounterSummary{}, err
	}
	customPicks, err := decodePicks(r.CustomAdversaries)
	if err != nil {
		return EncounterSummary{}, err
	}
	total := 0
	for _, p := range slices.Concat(srdPicks, customPicks) {
		total += p.Count
	}
	return EncounterSummary{
		ID:              r.ID,
		Name:            r.EncounterName,
		PartyID:         int64Ptr(r.PartyID),
		EnvironmentSlug: stringPtr(r.EnvironmentSlug),
		TotalCount:      total,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}, nil
}

func (s *Service) GetEncounter(id int64) (rules.EncounterView, error) {
	r, err := s.q.GetEncounter(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return rules.EncounterView{}, notFound("encounter", fmt.Sprint(id))
	}
	if err != nil {
		return rules.EncounterView{}, fmt.Errorf("loading encounter: %w", err)
	}
	return s.hydrate(r)
}

func (s *Service) hydrate(r db.Encounter) (rules.EncounterView, error) {
	srdPicks, err := decodePicks(r.Adversaries)
	if err != nil {
		return rules.EncounterView{}, err
	}
	customPicks, err := decodePicks(r.CustomAdversaries)
	if err != nil {
		return rules.EncounterView{}, err
	}

	view := rules.EncounterView{
		ID:              r.ID,
		Name:            r.EncounterName,
		PartyID:         int64Ptr(r.PartyID),
		Adversaries:     []rules.EncounterAdversary{},
		EnvironmentSlug: stringPtr(r.EnvironmentSlug),
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}

	for _, p := range slices.Concat(srdPicks, customPicks) {
		view.Adversaries = append(view.Adversaries, s.resolvePick(p))
		view.TotalCount += p.Count
	}

	if slug := view.EnvironmentSlug; slug != nil {
		if e, _, ok := s.lookupEnvironment(*slug); ok {
			view.Environment = &e
		}
	}

	if view.PartyID != nil {
		party, err := s.GetParty(*view.PartyID)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return rules.EncounterView{}, err
		}
		if err == nil {
			budget := rules.ComputeBudget(
				rules.EncounterSettings{PartySize: party.Size, PartyTier: party.Tier},
				view.Adversaries,
			)
			view.Budget = &budget
		}
	}

	return view, nil
}

func (s *Service) resolvePick(p Pick) rules.EncounterAdversary {
	card, source, ok := s.lookupAdversary(p.Slug)
	if !ok {
		return rules.EncounterAdversary{
			Adversary:  cardStub(p.Slug),
			Count:      p.Count,
			Unresolved: true,
		}
	}
	return rules.EncounterAdversary{Adversary: card, Count: p.Count, Source: source}
}

func (s *Service) SaveEncounter(in EncounterInput) (rules.EncounterView, error) {
	in, err := in.validate()
	if err != nil {
		return rules.EncounterView{}, err
	}
	srdJSON, err := encodePicks(in.Adversaries)
	if err != nil {
		return rules.EncounterView{}, err
	}
	customJSON, err := encodePicks(in.CustomAdversaries)
	if err != nil {
		return rules.EncounterView{}, err
	}

	var row db.Encounter
	if in.ID == nil {
		row, err = s.q.CreateEncounter(s.ctx, db.CreateEncounterParams{
			EncounterName:     in.Name,
			Adversaries:       srdJSON,
			CustomAdversaries: customJSON,
			EnvironmentSlug:   nullString(in.EnvironmentSlug),
			PartyID:           nullInt64(in.PartyID),
		})
		if err != nil {
			return rules.EncounterView{}, fmt.Errorf("creating encounter: %w", err)
		}
	} else {
		row, err = s.q.UpdateEncounter(s.ctx, db.UpdateEncounterParams{
			EncounterName:     in.Name,
			Adversaries:       srdJSON,
			CustomAdversaries: customJSON,
			EnvironmentSlug:   nullString(in.EnvironmentSlug),
			PartyID:           nullInt64(in.PartyID),
			ID:                *in.ID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return rules.EncounterView{}, notFound("encounter", fmt.Sprint(*in.ID))
		}
		if err != nil {
			return rules.EncounterView{}, fmt.Errorf("updating encounter: %w", err)
		}
	}
	return s.hydrate(row)
}

func (s *Service) DeleteEncounter(id int64) error {
	if err := s.q.DeleteEncounter(s.ctx, id); err != nil {
		return fmt.Errorf("deleting encounter: %w", err)
	}
	return nil
}
