package gm

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/mhetem/DH-Companion/internal/db"
)

const defaultWorldTitle = "Untitled"

type WorldNote struct {
	ID         int64  `json:"id"`
	CampaignID int64  `json:"campaignId"`
	ParentID   *int64 `json:"parentId"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	WorldDate  string `json:"worldDate"`
	Body       string `json:"body"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type WorldNoteInput struct {
	ID         *int64 `json:"id"`
	CampaignID int64  `json:"campaignId"`
	ParentID   *int64 `json:"parentId"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	WorldDate  string `json:"worldDate"`
	Body       string `json:"body"`
}

func (in WorldNoteInput) validate() (WorldNoteInput, error) {
	if in.CampaignID <= 0 {
		return in, fmt.Errorf("a world page needs a campaign")
	}
	kind, err := validateWorldKind(in.Kind)
	if err != nil {
		return in, err
	}
	in.Kind = kind
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		in.Title = defaultWorldTitle
	}
	in.WorldDate = strings.TrimSpace(in.WorldDate)
	return in, nil
}

func worldNoteView(r db.WorldNote) WorldNote {
	return WorldNote{
		ID:         r.ID,
		CampaignID: r.CampaignID,
		ParentID:   int64Ptr(r.ParentID),
		Kind:       r.Kind,
		Title:      r.Title,
		WorldDate:  r.WorldDate,
		Body:       r.Body,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func (s *Service) ListWorldNotes(campaignID int64) ([]WorldNote, error) {
	rows, err := s.q.ListWorldNotesForCampaign(s.ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("listing world pages: %w", err)
	}
	out := make([]WorldNote, 0, len(rows))
	for _, r := range rows {
		out = append(out, worldNoteView(r))
	}
	return out, nil
}

func (s *Service) GetWorldNote(id int64) (WorldNote, error) {
	row, err := s.q.GetWorldNote(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return WorldNote{}, notFound("world page", fmt.Sprint(id))
	}
	if err != nil {
		return WorldNote{}, fmt.Errorf("loading world page: %w", err)
	}
	return worldNoteView(row), nil
}

func (s *Service) SaveWorldNote(in WorldNoteInput) (WorldNote, error) {
	in, err := in.validate()
	if err != nil {
		return WorldNote{}, err
	}

	var row db.WorldNote
	if in.ID == nil {
		if in.ParentID != nil {
			if err := s.checkWorldParent(in.CampaignID, *in.ParentID, 0); err != nil {
				return WorldNote{}, err
			}
		}
		position, err := s.q.NextWorldNotePosition(s.ctx, in.CampaignID)
		if err != nil {
			return WorldNote{}, fmt.Errorf("placing world page: %w", err)
		}
		row, err = s.q.CreateWorldNote(s.ctx, db.CreateWorldNoteParams{
			CampaignID: in.CampaignID,
			ParentID:   nullInt64(in.ParentID),
			Kind:       in.Kind,
			Title:      in.Title,
			WorldDate:  in.WorldDate,
			Body:       in.Body,
			Position:   position,
		})
		if err != nil {
			return WorldNote{}, fmt.Errorf("creating world page: %w", err)
		}
	} else {
		row, err = s.q.UpdateWorldNote(s.ctx, db.UpdateWorldNoteParams{
			Kind:      in.Kind,
			Title:     in.Title,
			WorldDate: in.WorldDate,
			Body:      in.Body,
			ID:        *in.ID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return WorldNote{}, notFound("world page", fmt.Sprint(*in.ID))
		}
		if err != nil {
			return WorldNote{}, fmt.Errorf("updating world page: %w", err)
		}
	}
	return worldNoteView(row), nil
}

func (s *Service) MoveWorldNote(id int64, parentID *int64) (WorldNote, error) {
	note, err := s.GetWorldNote(id)
	if err != nil {
		return WorldNote{}, err
	}
	if parentID != nil {
		if err := s.checkWorldParent(note.CampaignID, *parentID, id); err != nil {
			return WorldNote{}, err
		}
	}
	position, err := s.q.NextWorldNotePosition(s.ctx, note.CampaignID)
	if err != nil {
		return WorldNote{}, fmt.Errorf("placing world page: %w", err)
	}
	row, err := s.q.MoveWorldNote(s.ctx, db.MoveWorldNoteParams{
		ParentID: nullInt64(parentID),
		Position: position,
		ID:       id,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return WorldNote{}, notFound("world page", fmt.Sprint(id))
	}
	if err != nil {
		return WorldNote{}, fmt.Errorf("moving world page: %w", err)
	}
	return worldNoteView(row), nil
}

func (s *Service) ShiftWorldNote(id int64, delta int) error {
	note, err := s.q.GetWorldNote(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return notFound("world page", fmt.Sprint(id))
	}
	if err != nil {
		return fmt.Errorf("loading world page: %w", err)
	}
	rows, err := s.q.ListWorldNotesForCampaign(s.ctx, note.CampaignID)
	if err != nil {
		return fmt.Errorf("listing world pages: %w", err)
	}

	siblings := make([]db.WorldNote, 0, len(rows))
	from := -1
	for _, r := range rows {
		if r.ParentID != note.ParentID {
			continue
		}
		if r.ID == id {
			from = len(siblings)
		}
		siblings = append(siblings, r)
	}
	to := min(max(from+delta, 0), len(siblings)-1)
	if from < 0 || to == from {
		return nil
	}

	positions := make([]int64, len(siblings))
	for i, r := range siblings {
		positions[i] = r.Position
	}
	moved := siblings[from]
	siblings = slices.Insert(slices.Delete(siblings, from, from+1), to, moved)

	return s.tx(func(q *db.Queries) error {
		for i, r := range siblings {
			if r.Position == positions[i] {
				continue
			}
			if err := q.SetWorldNotePosition(s.ctx, db.SetWorldNotePositionParams{
				Position: positions[i],
				ID:       r.ID,
			}); err != nil {
				return fmt.Errorf("reordering world pages: %w", err)
			}
		}
		return nil
	})
}

func (s *Service) DeleteWorldNote(id int64) error {
	if err := s.q.DeleteWorldNote(s.ctx, id); err != nil {
		return fmt.Errorf("deleting world page: %w", err)
	}
	return nil
}

func (s *Service) checkWorldParent(campaignID, parentID, childID int64) error {
	for at := &parentID; at != nil; {
		if *at == childID {
			return fmt.Errorf("a page can't be moved inside itself or one of its own pages")
		}
		row, err := s.q.GetWorldNote(s.ctx, *at)
		if errors.Is(err, sql.ErrNoRows) {
			return notFound("world page", fmt.Sprint(*at))
		}
		if err != nil {
			return fmt.Errorf("loading world page: %w", err)
		}
		if row.CampaignID != campaignID {
			return fmt.Errorf("a page can only sit inside a page from the same campaign")
		}
		at = int64Ptr(row.ParentID)
	}
	return nil
}
