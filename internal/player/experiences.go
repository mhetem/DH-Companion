package player

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mhetem/DH-Companion/internal/db"
)

type Experience struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Modifier int    `json:"modifier"`
}

type ExperienceInput struct {
	ID          *int64 `json:"id"`
	CharacterID int64  `json:"characterId"`
	Name        string `json:"name"`
	Modifier    int    `json:"modifier"`
}

func experienceView(r db.Experience) Experience {
	return Experience{ID: r.ID, Name: r.Name, Modifier: int(r.Modifier)}
}

func (s *Service) ListExperiences(characterID int64) ([]Experience, error) {
	rows, err := s.q.ListExperiences(s.ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("listing experiences: %w", err)
	}
	out := make([]Experience, 0, len(rows))
	for _, r := range rows {
		out = append(out, experienceView(r))
	}
	return out, nil
}

func (s *Service) SaveExperience(in ExperienceInput) ([]Experience, error) {
	name, err := validateName(in.Name)
	if err != nil {
		return nil, err
	}
	modifier := clamp(in.Modifier, 0, 9)

	if in.ID == nil {
		if _, err := s.character(in.CharacterID); err != nil {
			return nil, err
		}
		if in.Modifier == 0 {
			modifier = 2
		}
		if _, err := s.q.CreateExperience(s.ctx, db.CreateExperienceParams{
			CharacterID: in.CharacterID,
			Name:        name,
			Modifier:    int64(modifier),
		}); err != nil {
			return nil, fmt.Errorf("creating experience: %w", err)
		}
		return s.ListExperiences(in.CharacterID)
	}

	existing, err := s.q.GetExperience(s.ctx, *in.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("experience", fmt.Sprint(*in.ID))
	}
	if err != nil {
		return nil, fmt.Errorf("loading experience: %w", err)
	}
	if _, err := s.q.UpdateExperience(s.ctx, db.UpdateExperienceParams{
		Name:     name,
		Modifier: int64(modifier),
		ID:       *in.ID,
	}); err != nil {
		return nil, fmt.Errorf("updating experience: %w", err)
	}
	return s.ListExperiences(existing.CharacterID)
}

func (s *Service) DeleteExperience(id int64) ([]Experience, error) {
	existing, err := s.q.GetExperience(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("experience", fmt.Sprint(id))
	}
	if err != nil {
		return nil, fmt.Errorf("loading experience: %w", err)
	}
	if err := s.q.DeleteExperience(s.ctx, id); err != nil {
		return nil, fmt.Errorf("deleting experience: %w", err)
	}
	return s.ListExperiences(existing.CharacterID)
}
