package gm

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mhetem/DH-Companion/internal/db"
)

const defaultNoteTitle = "Untitled"

type Note struct {
	ID         int64  `json:"id"`
	CampaignID int64  `json:"campaignId"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type NoteInput struct {
	ID         *int64 `json:"id"`
	CampaignID int64  `json:"campaignId"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	Body       string `json:"body"`
}

func (in NoteInput) validate() (NoteInput, error) {
	if in.CampaignID <= 0 {
		return in, fmt.Errorf("a note needs a campaign")
	}
	kind, err := validateNoteKind(in.Kind)
	if err != nil {
		return in, err
	}
	in.Kind = kind
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		in.Title = defaultNoteTitle
	}
	return in, nil
}

func noteView(r db.Note) Note {
	return Note{
		ID:         r.ID,
		CampaignID: r.CampaignID,
		Kind:       r.Kind,
		Title:      r.Title,
		Body:       r.Body,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func (s *Service) ListNotes(campaignID int64) ([]Note, error) {
	rows, err := s.q.ListNotesForCampaign(s.ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("listing notes: %w", err)
	}
	out := make([]Note, 0, len(rows))
	for _, r := range rows {
		out = append(out, noteView(r))
	}
	return out, nil
}

func (s *Service) ListNotesByKind(campaignID int64, kind string) ([]Note, error) {
	k, err := validateNoteKind(kind)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.ListNotesForCampaignByKind(s.ctx, db.ListNotesForCampaignByKindParams{
		CampaignID: campaignID, Kind: k,
	})
	if err != nil {
		return nil, fmt.Errorf("listing notes: %w", err)
	}
	out := make([]Note, 0, len(rows))
	for _, r := range rows {
		out = append(out, noteView(r))
	}
	return out, nil
}

func (s *Service) GetNote(id int64) (Note, error) {
	row, err := s.q.GetNote(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Note{}, notFound("note", fmt.Sprint(id))
	}
	if err != nil {
		return Note{}, fmt.Errorf("loading note: %w", err)
	}
	return noteView(row), nil
}

func (s *Service) SaveNote(in NoteInput) (Note, error) {
	in, err := in.validate()
	if err != nil {
		return Note{}, err
	}

	var row db.Note
	if in.ID == nil {
		row, err = s.q.CreateNote(s.ctx, db.CreateNoteParams{
			CampaignID: in.CampaignID,
			Kind:       in.Kind,
			Title:      in.Title,
			Body:       in.Body,
		})
		if err != nil {
			return Note{}, fmt.Errorf("creating note: %w", err)
		}
	} else {
		row, err = s.q.UpdateNote(s.ctx, db.UpdateNoteParams{
			Kind:  in.Kind,
			Title: in.Title,
			Body:  in.Body,
			ID:    *in.ID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return Note{}, notFound("note", fmt.Sprint(*in.ID))
		}
		if err != nil {
			return Note{}, fmt.Errorf("updating note: %w", err)
		}
	}
	return noteView(row), nil
}

func (s *Service) DeleteNote(id int64) error {
	if err := s.q.DeleteNote(s.ctx, id); err != nil {
		return fmt.Errorf("deleting note: %w", err)
	}
	return nil
}
