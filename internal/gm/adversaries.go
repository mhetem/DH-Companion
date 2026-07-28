package gm

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

func toAdversary(r db.CustomAdversary) (cards.Adversary, error) {
	a := cards.Adversary{
		Meta: cards.Meta{
			Kind:        cards.KindAdversary,
			Slug:        r.Slug,
			Name:        r.Name,
			Tier:        r.Tier,
			Type:        r.Type,
			Description: r.Description,
		},
		HordeNumber:    r.HordeNumber,
		Motives:        r.Motives,
		Experiences:    r.Experiences,
		Difficulty:     r.Difficulty,
		ThresholdMinor: r.ThresholdMinor,
		ThresholdMajor: r.ThresholdMajor,
		Hp:             r.Hp,
		Stress:         r.Stress,
		Features:       []cards.Feature{},
	}
	if r.StandardAttack.Valid && r.StandardAttack.String != "" {
		if err := json.Unmarshal([]byte(r.StandardAttack.String), &a.StandardAttack); err != nil {
			return a, fmt.Errorf("adversary %q has an unreadable standard attack: %w", r.Slug, err)
		}
	}
	if r.Features.Valid && r.Features.String != "" {
		if err := json.Unmarshal([]byte(r.Features.String), &a.Features); err != nil {
			return a, fmt.Errorf("adversary %q has unreadable features: %w", r.Slug, err)
		}
	}
	return a, nil
}

func validateAdversary(a cards.Adversary) (name string, attack, features sql.NullString, err error) {
	if name, err = validateName(a.Name); err != nil {
		return "", attack, features, err
	}
	if err = validateTier(a.Tier); err != nil {
		return "", attack, features, err
	}
	if err = validateAdversaryType(a.Type); err != nil {
		return "", attack, features, err
	}

	attackJSON, err := json.Marshal(a.StandardAttack)
	if err != nil {
		return "", attack, features, fmt.Errorf("encoding standard attack: %w", err)
	}
	if a.Features == nil {
		a.Features = []cards.Feature{}
	}
	featuresJSON, err := json.Marshal(a.Features)
	if err != nil {
		return "", attack, features, fmt.Errorf("encoding features: %w", err)
	}
	attack = sql.NullString{String: string(attackJSON), Valid: true}
	features = sql.NullString{String: string(featuresJSON), Valid: true}
	return name, attack, features, nil
}

func (s *Service) ListCustomAdversaries(filter Filter) ([]cards.Adversary, error) {
	tier, typ := filter.normalized()
	rows, err := s.q.ShowAllCustomAdversaries(s.ctx, db.ShowAllCustomAdversariesParams{Tier: tier, Type: typ})
	if err != nil {
		return nil, fmt.Errorf("listing custom adversaries: %w", err)
	}
	out := make([]cards.Adversary, 0, len(rows))
	for _, r := range rows {
		a, err := toAdversary(r)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

func (s *Service) GetCustomAdversary(slug string) (cards.Adversary, error) {
	r, err := s.q.GetCustomBySlug(s.ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return cards.Adversary{}, notFound("custom adversary", slug)
	}
	if err != nil {
		return cards.Adversary{}, fmt.Errorf("loading custom adversary: %w", err)
	}
	return toAdversary(r)
}

func (s *Service) CreateCustomAdversary(a cards.Adversary) (cards.Adversary, error) {
	name, attack, features, err := validateAdversary(a)
	if err != nil {
		return cards.Adversary{}, err
	}
	slug := slugify(name)
	if slug == "" {
		return cards.Adversary{}, fmt.Errorf("name %q does not produce a usable slug", name)
	}

	r, err := s.q.CreateCustomAdversary(s.ctx, db.CreateCustomAdversaryParams{
		Slug:           slug,
		Name:           name,
		Tier:           a.Tier,
		Type:           a.Type,
		Description:    a.Description,
		HordeNumber:    a.HordeNumber,
		Motives:        a.Motives,
		Experiences:    a.Experiences,
		Difficulty:     a.Difficulty,
		ThresholdMinor: a.ThresholdMinor,
		ThresholdMajor: a.ThresholdMajor,
		Hp:             a.Hp,
		Stress:         a.Stress,
		StandardAttack: attack,
		Features:       features,
	})
	if isUniqueViolation(err) {
		return cards.Adversary{}, fmt.Errorf("a custom adversary named %q already exists", name)
	}
	if err != nil {
		return cards.Adversary{}, fmt.Errorf("creating custom adversary: %w", err)
	}
	card, err := toAdversary(r)
	if err != nil {
		return card, err
	}
	return card, s.ReindexCards()
}

func (s *Service) UpdateCustomAdversary(a cards.Adversary) (cards.Adversary, error) {
	if a.Slug == "" {
		return cards.Adversary{}, fmt.Errorf("slug is required to update a custom adversary")
	}
	name, attack, features, err := validateAdversary(a)
	if err != nil {
		return cards.Adversary{}, err
	}

	r, err := s.q.UpdateCustomAdversary(s.ctx, db.UpdateCustomAdversaryParams{
		Name:           name,
		Tier:           a.Tier,
		Type:           a.Type,
		Description:    a.Description,
		HordeNumber:    a.HordeNumber,
		Motives:        a.Motives,
		Experiences:    a.Experiences,
		Difficulty:     a.Difficulty,
		ThresholdMinor: a.ThresholdMinor,
		ThresholdMajor: a.ThresholdMajor,
		Hp:             a.Hp,
		Stress:         a.Stress,
		StandardAttack: attack,
		Features:       features,
		Slug:           a.Slug,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return cards.Adversary{}, notFound("custom adversary", a.Slug)
	}
	if err != nil {
		return cards.Adversary{}, fmt.Errorf("updating custom adversary: %w", err)
	}
	card, err := toAdversary(r)
	if err != nil {
		return card, err
	}
	return card, s.ReindexCards()
}

func (s *Service) DeleteCustomAdversary(slug string) error {
	if err := s.q.DeleteCustomAdversary(s.ctx, slug); err != nil {
		return fmt.Errorf("deleting custom adversary: %w", err)
	}
	return s.ReindexCards()
}
