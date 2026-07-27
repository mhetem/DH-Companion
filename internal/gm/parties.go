package gm

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mhetem/DH-Companion/internal/db"
)

type Party struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Size      int    `json:"size"`
	Tier      string `json:"tier"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type PartyInput struct {
	Name string `json:"name"`
	Size int    `json:"size"`
	Tier string `json:"tier"`
}

func (in PartyInput) validate() (PartyInput, error) {
	name, err := validateName(in.Name)
	if err != nil {
		return in, err
	}
	if in.Size < 1 {
		return in, fmt.Errorf("party size must be at least 1, got %d", in.Size)
	}
	if err := validateTier(in.Tier); err != nil {
		return in, err
	}
	in.Name = name
	return in, nil
}

func toParty(r db.Party) Party {
	return Party{
		ID:        r.ID,
		Name:      r.Name,
		Size:      int(r.Size),
		Tier:      r.Tier,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func (s *Service) ListParties() ([]Party, error) {
	rows, err := s.q.ListParties(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing parties: %w", err)
	}
	out := make([]Party, 0, len(rows))
	for _, r := range rows {
		out = append(out, toParty(r))
	}
	return out, nil
}

func (s *Service) GetParty(id int64) (Party, error) {
	r, err := s.q.GetParty(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Party{}, notFound("party", fmt.Sprint(id))
	}
	if err != nil {
		return Party{}, fmt.Errorf("loading party: %w", err)
	}
	return toParty(r), nil
}

func (s *Service) CreateParty(in PartyInput) (Party, error) {
	in, err := in.validate()
	if err != nil {
		return Party{}, err
	}
	r, err := s.q.CreateParty(s.ctx, db.CreatePartyParams{
		Name: in.Name,
		Size: int64(in.Size),
		Tier: in.Tier,
	})
	if err != nil {
		return Party{}, fmt.Errorf("creating party: %w", err)
	}
	return toParty(r), nil
}

func (s *Service) UpdateParty(id int64, in PartyInput) (Party, error) {
	in, err := in.validate()
	if err != nil {
		return Party{}, err
	}
	r, err := s.q.UpdateParty(s.ctx, db.UpdatePartyParams{
		Name: in.Name,
		Size: int64(in.Size),
		Tier: in.Tier,
		ID:   id,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Party{}, notFound("party", fmt.Sprint(id))
	}
	if err != nil {
		return Party{}, fmt.Errorf("updating party: %w", err)
	}
	return toParty(r), nil
}

func (s *Service) DeleteParty(id int64) error {
	if err := s.q.DeleteParty(s.ctx, id); err != nil {
		return fmt.Errorf("deleting party: %w", err)
	}
	return nil
}
