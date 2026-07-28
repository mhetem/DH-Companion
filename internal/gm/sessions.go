package gm

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mhetem/DH-Companion/internal/db"
)

const defaultSessionTitle = "Untitled"

type SessionSummary struct {
	ID         int64  `json:"id"`
	CampaignID int64  `json:"campaignId"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	Date       string `json:"date"`
	Recap      string `json:"recap"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type SessionView struct {
	SessionSummary
	Encounters []EncounterSummary `json:"encounters"`
	Combats    []CombatSummary    `json:"combats"`
}

type SessionInput struct {
	ID         *int64 `json:"id"`
	CampaignID int64  `json:"campaignId"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	Date       string `json:"date"`
	Recap      string `json:"recap"`
}

func (in SessionInput) validate() (SessionInput, error) {
	if in.CampaignID <= 0 {
		return in, fmt.Errorf("a session needs a campaign")
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		in.Title = defaultSessionTitle
	}
	in.Date = strings.TrimSpace(in.Date)
	if in.Date == "" {
		in.Date = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	}
	in.Recap = strings.TrimSpace(in.Recap)
	if in.Number < 0 {
		return in, fmt.Errorf("session number can't be negative, got %d", in.Number)
	}
	return in, nil
}

func sessionSummary(r db.Session) SessionSummary {
	return SessionSummary{
		ID:         r.ID,
		CampaignID: r.CampaignID,
		Number:     int(r.Number),
		Title:      r.Title,
		Date:       r.Date,
		Recap:      r.Recap,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func (s *Service) ListSessions(campaignID int64) ([]SessionSummary, error) {
	rows, err := s.q.ListSessionsForCampaign(s.ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}
	out := make([]SessionSummary, 0, len(rows))
	for _, r := range rows {
		out = append(out, sessionSummary(r))
	}
	return out, nil
}

func (s *Service) GetSession(id int64) (SessionView, error) {
	row, err := s.q.GetSession(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return SessionView{}, notFound("session", fmt.Sprint(id))
	}
	if err != nil {
		return SessionView{}, fmt.Errorf("loading session: %w", err)
	}
	return s.hydrateSession(row)
}

func (s *Service) SaveSession(in SessionInput) (SessionView, error) {
	in, err := in.validate()
	if err != nil {
		return SessionView{}, err
	}

	var row db.Session
	if in.ID == nil {
		number := int64(in.Number)
		if number == 0 {
			number, err = s.q.NextSessionNumber(s.ctx, in.CampaignID)
			if err != nil {
				return SessionView{}, fmt.Errorf("numbering session: %w", err)
			}
		}
		row, err = s.q.CreateSession(s.ctx, db.CreateSessionParams{
			CampaignID: in.CampaignID,
			Number:     number,
			Title:      in.Title,
			Date:       in.Date,
			Recap:      in.Recap,
		})
		if isUniqueViolation(err) {
			return SessionView{}, fmt.Errorf("session %d already exists in this campaign", number)
		}
		if err != nil {
			return SessionView{}, fmt.Errorf("creating session: %w", err)
		}
	} else {
		row, err = s.q.UpdateSession(s.ctx, db.UpdateSessionParams{
			Number: int64(in.Number),
			Title:  in.Title,
			Date:   in.Date,
			Recap:  in.Recap,
			ID:     *in.ID,
		})
		if isUniqueViolation(err) {
			return SessionView{}, fmt.Errorf("session %d already exists in this campaign", in.Number)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return SessionView{}, notFound("session", fmt.Sprint(*in.ID))
		}
		if err != nil {
			return SessionView{}, fmt.Errorf("updating session: %w", err)
		}
	}
	return s.hydrateSession(row)
}

func (s *Service) DeleteSession(id int64) error {
	if err := s.q.DeleteSession(s.ctx, id); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}

func (s *Service) LinkEncounter(sessionID, encounterID int64) (SessionView, error) {
	err := s.q.LinkSessionEncounter(s.ctx, db.LinkSessionEncounterParams{
		SessionID: sessionID, EncounterID: encounterID,
	})
	if err != nil {
		return SessionView{}, fmt.Errorf("linking encounter to session: %w", err)
	}
	return s.GetSession(sessionID)
}

func (s *Service) UnlinkEncounter(sessionID, encounterID int64) (SessionView, error) {
	err := s.q.UnlinkSessionEncounter(s.ctx, db.UnlinkSessionEncounterParams{
		SessionID: sessionID, EncounterID: encounterID,
	})
	if err != nil {
		return SessionView{}, fmt.Errorf("unlinking encounter from session: %w", err)
	}
	return s.GetSession(sessionID)
}

func (s *Service) SessionsForEncounter(encounterID int64) ([]SessionSummary, error) {
	rows, err := s.q.ListSessionsForEncounter(s.ctx, encounterID)
	if err != nil {
		return nil, fmt.Errorf("listing sessions for encounter: %w", err)
	}
	out := make([]SessionSummary, 0, len(rows))
	for _, r := range rows {
		out = append(out, sessionSummary(r))
	}
	return out, nil
}

func (s *Service) hydrateSession(row db.Session) (SessionView, error) {
	rows, err := s.q.ListEncountersForSession(s.ctx, row.ID)
	if err != nil {
		return SessionView{}, fmt.Errorf("loading session encounters: %w", err)
	}
	view := SessionView{
		SessionSummary: sessionSummary(row),
		Encounters:     make([]EncounterSummary, 0, len(rows)),
	}
	for _, r := range rows {
		summary, err := encounterSummary(r)
		if err != nil {
			return SessionView{}, err
		}
		view.Encounters = append(view.Encounters, summary)
	}

	combats, err := s.CombatsForSession(row.ID)
	if err != nil {
		return SessionView{}, err
	}
	view.Combats = combats
	return view, nil
}
