package gm

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

func toEnvironment(r db.CustomEnvironment) (cards.Environment, error) {
	e := cards.Environment{
		Meta: cards.Meta{
			Kind:        cards.KindEnvironment,
			Slug:        r.Slug,
			Name:        r.Name,
			Tier:        r.Tier,
			Type:        r.Type,
			Description: r.Description,
		},
		Difficulty:           r.Difficulty,
		Impulses:             r.Impulses,
		PotentialAdversaries: []string{},
		Features:             []cards.Feature{},
	}
	if r.PotentialAdversaries != "" {
		if err := json.Unmarshal([]byte(r.PotentialAdversaries), &e.PotentialAdversaries); err != nil {
			return e, fmt.Errorf("environment %q has an unreadable adversary list: %w", r.Slug, err)
		}
	}
	if r.Features != "" {
		if err := json.Unmarshal([]byte(r.Features), &e.Features); err != nil {
			return e, fmt.Errorf("environment %q has unreadable features: %w", r.Slug, err)
		}
	}
	return e, nil
}

func validateEnvironment(e cards.Environment) (name, potential, features string, err error) {
	if name, err = validateName(e.Name); err != nil {
		return "", "", "", err
	}
	if err = validateTier(e.Tier); err != nil {
		return "", "", "", err
	}
	if err = validateEnvironmentType(e.Type); err != nil {
		return "", "", "", err
	}

	if e.PotentialAdversaries == nil {
		e.PotentialAdversaries = []string{}
	}
	potentialJSON, err := json.Marshal(e.PotentialAdversaries)
	if err != nil {
		return "", "", "", fmt.Errorf("encoding potential adversaries: %w", err)
	}
	if e.Features == nil {
		e.Features = []cards.Feature{}
	}
	featuresJSON, err := json.Marshal(e.Features)
	if err != nil {
		return "", "", "", fmt.Errorf("encoding features: %w", err)
	}
	return name, string(potentialJSON), string(featuresJSON), nil
}

func (s *Service) ListCustomEnvironments(filter Filter) ([]cards.Environment, error) {
	tier, typ := filter.normalized()
	rows, err := s.q.ShowAllCustomEnvironments(s.ctx, db.ShowAllCustomEnvironmentsParams{Tier: tier, Type: typ})
	if err != nil {
		return nil, fmt.Errorf("listing custom environments: %w", err)
	}
	out := make([]cards.Environment, 0, len(rows))
	for _, r := range rows {
		e, err := toEnvironment(r)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (s *Service) GetCustomEnvironment(slug string) (cards.Environment, error) {
	r, err := s.q.GetCustomEnvironmentBySlug(s.ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return cards.Environment{}, notFound("custom environment", slug)
	}
	if err != nil {
		return cards.Environment{}, fmt.Errorf("loading custom environment: %w", err)
	}
	card, err := toEnvironment(r)
	if err != nil {
		return card, err
	}
	return card, s.ReindexCards()
}

func (s *Service) CreateCustomEnvironment(e cards.Environment) (cards.Environment, error) {
	name, potential, features, err := validateEnvironment(e)
	if err != nil {
		return cards.Environment{}, err
	}
	slug := slugify(name)
	if slug == "" {
		return cards.Environment{}, fmt.Errorf("name %q does not produce a usable slug", name)
	}

	r, err := s.q.CreateCustomEnvironment(s.ctx, db.CreateCustomEnvironmentParams{
		Slug:                 slug,
		Name:                 name,
		Tier:                 e.Tier,
		Type:                 e.Type,
		Description:          e.Description,
		Impulses:             e.Impulses,
		Difficulty:           e.Difficulty,
		PotentialAdversaries: potential,
		Features:             features,
	})
	if isUniqueViolation(err) {
		return cards.Environment{}, fmt.Errorf("a custom environment named %q already exists", name)
	}
	if err != nil {
		return cards.Environment{}, fmt.Errorf("creating custom environment: %w", err)
	}
	card, err := toEnvironment(r)
	if err != nil {
		return card, err
	}
	return card, s.ReindexCards()
}

func (s *Service) UpdateCustomEnvironment(e cards.Environment) (cards.Environment, error) {
	if e.Slug == "" {
		return cards.Environment{}, fmt.Errorf("slug is required to update a custom environment")
	}
	name, potential, features, err := validateEnvironment(e)
	if err != nil {
		return cards.Environment{}, err
	}

	r, err := s.q.UpdateCustomEnvironment(s.ctx, db.UpdateCustomEnvironmentParams{
		Name:                 name,
		Tier:                 e.Tier,
		Type:                 e.Type,
		Description:          e.Description,
		Impulses:             e.Impulses,
		Difficulty:           e.Difficulty,
		PotentialAdversaries: potential,
		Features:             features,
		Slug:                 e.Slug,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return cards.Environment{}, notFound("custom environment", e.Slug)
	}
	if err != nil {
		return cards.Environment{}, fmt.Errorf("updating custom environment: %w", err)
	}
	card, err := toEnvironment(r)
	if err != nil {
		return card, err
	}
	return card, s.ReindexCards()
}

func (s *Service) DeleteCustomEnvironment(slug string) error {
	if err := s.q.DeleteCustomEnvironment(s.ctx, slug); err != nil {
		return fmt.Errorf("deleting custom environment: %w", err)
	}
	return s.ReindexCards()
}
