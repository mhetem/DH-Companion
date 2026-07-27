package gm

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mhetem/DH-Companion/internal/db"
)

type Countdown struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Value     int    `json:"value"`
	Max       int    `json:"max"`
	Kind      string `json:"kind"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type CountdownInput struct {
	ID    *int64 `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
	Max   int    `json:"max"`
	Kind  string `json:"kind"`
}

func (in CountdownInput) validate() (CountdownInput, error) {
	name, err := validateName(in.Name)
	if err != nil {
		return in, err
	}
	in.Name = name
	if in.Max < 1 {
		return in, fmt.Errorf("a countdown needs at least one segment, got %d", in.Max)
	}
	in.Kind = strings.TrimSpace(in.Kind)
	if in.Value < 0 {
		in.Value = 0
	}
	if in.Value > in.Max {
		in.Value = in.Max
	}
	return in, nil
}

func countdownView(r db.Countdown) Countdown {
	return Countdown{
		ID:        r.ID,
		Name:      r.Name,
		Value:     int(r.Value),
		Max:       int(r.Max),
		Kind:      r.Kind,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func (s *Service) ListCountdowns() ([]Countdown, error) {
	rows, err := s.q.ListCountdowns(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing countdowns: %w", err)
	}
	out := make([]Countdown, 0, len(rows))
	for _, r := range rows {
		out = append(out, countdownView(r))
	}
	return out, nil
}

func (s *Service) GetCountdown(id int64) (Countdown, error) {
	row, err := s.q.GetCountdown(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Countdown{}, notFound("countdown", fmt.Sprint(id))
	}
	if err != nil {
		return Countdown{}, fmt.Errorf("loading countdown: %w", err)
	}
	return countdownView(row), nil
}

func (s *Service) SaveCountdown(in CountdownInput) (Countdown, error) {
	in, err := in.validate()
	if err != nil {
		return Countdown{}, err
	}

	var row db.Countdown
	if in.ID == nil {
		row, err = s.q.CreateCountdown(s.ctx, db.CreateCountdownParams{
			Name:  in.Name,
			Value: int64(in.Value),
			Max:   int64(in.Max),
			Kind:  in.Kind,
		})
		if err != nil {
			return Countdown{}, fmt.Errorf("creating countdown: %w", err)
		}
	} else {
		row, err = s.q.UpdateCountdown(s.ctx, db.UpdateCountdownParams{
			Name:  in.Name,
			Max:   int64(in.Max),
			Value: int64(in.Value),
			Kind:  in.Kind,
			ID:    *in.ID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return Countdown{}, notFound("countdown", fmt.Sprint(*in.ID))
		}
		if err != nil {
			return Countdown{}, fmt.Errorf("updating countdown: %w", err)
		}
	}
	return countdownView(row), nil
}

func (s *Service) AdvanceCountdown(id int64, delta int) (Countdown, error) {
	row, err := s.q.AdjustCountdown(s.ctx, db.AdjustCountdownParams{
		Delta: int64(delta), ID: id,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Countdown{}, notFound("countdown", fmt.Sprint(id))
	}
	if err != nil {
		return Countdown{}, fmt.Errorf("advancing countdown: %w", err)
	}
	return countdownView(row), nil
}

func (s *Service) DeleteCountdown(id int64) error {
	if err := s.q.DeleteCountdown(s.ctx, id); err != nil {
		return fmt.Errorf("deleting countdown: %w", err)
	}
	return nil
}
