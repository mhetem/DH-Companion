package gm

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mhetem/DH-Companion/internal/db"
)

type Campaign struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CurrentFear int    `json:"currentFear"`
	FearMax     int    `json:"fearMax"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type CampaignInput struct {
	ID          *int64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (in CampaignInput) validate() (CampaignInput, error) {
	name, err := validateName(in.Name)
	if err != nil {
		return in, err
	}
	in.Name = name
	in.Description = strings.TrimSpace(in.Description)
	return in, nil
}

func campaignView(r db.Campaign) Campaign {
	return Campaign{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		CurrentFear: int(r.CurrentFear),
		FearMax:     FearMax,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func (s *Service) ListCampaigns() ([]Campaign, error) {
	rows, err := s.q.ListCampaigns(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing campaigns: %w", err)
	}
	out := make([]Campaign, 0, len(rows))
	for _, r := range rows {
		out = append(out, campaignView(r))
	}
	return out, nil
}

func (s *Service) GetCampaign(id int64) (Campaign, error) {
	row, err := s.q.GetCampaign(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Campaign{}, notFound("campaign", fmt.Sprint(id))
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("loading campaign: %w", err)
	}
	return campaignView(row), nil
}

func (s *Service) SaveCampaign(in CampaignInput) (Campaign, error) {
	in, err := in.validate()
	if err != nil {
		return Campaign{}, err
	}

	var row db.Campaign
	if in.ID == nil {
		row, err = s.q.CreateCampaign(s.ctx, db.CreateCampaignParams{
			Name:        in.Name,
			Description: in.Description,
		})
		if err != nil {
			return Campaign{}, fmt.Errorf("creating campaign: %w", err)
		}
	} else {
		row, err = s.q.UpdateCampaign(s.ctx, db.UpdateCampaignParams{
			Name:        in.Name,
			Description: in.Description,
			ID:          *in.ID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return Campaign{}, notFound("campaign", fmt.Sprint(*in.ID))
		}
		if err != nil {
			return Campaign{}, fmt.Errorf("updating campaign: %w", err)
		}
	}
	return campaignView(row), nil
}

type MasterNote struct {
	CampaignID int64  `json:"campaignId"`
	Body       string `json:"body"`
	UpdatedAt  string `json:"updatedAt"`
}

func (s *Service) GetMasterNote(campaignID int64) (MasterNote, error) {
	row, err := s.q.GetCampaignMasterNote(s.ctx, campaignID)
	if errors.Is(err, sql.ErrNoRows) {
		return MasterNote{}, notFound("campaign", fmt.Sprint(campaignID))
	}
	if err != nil {
		return MasterNote{}, fmt.Errorf("loading master note: %w", err)
	}
	return MasterNote{
		CampaignID: campaignID,
		Body:       row.MasterNote,
		UpdatedAt:  row.MasterNoteUpdatedAt,
	}, nil
}

func (s *Service) SaveMasterNote(campaignID int64, body string) (MasterNote, error) {
	if campaignID <= 0 {
		return MasterNote{}, fmt.Errorf("a master note needs a campaign")
	}
	row, err := s.q.SetCampaignMasterNote(s.ctx, db.SetCampaignMasterNoteParams{
		MasterNote: body,
		ID:         campaignID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return MasterNote{}, notFound("campaign", fmt.Sprint(campaignID))
	}
	if err != nil {
		return MasterNote{}, fmt.Errorf("saving master note: %w", err)
	}
	return MasterNote{
		CampaignID: campaignID,
		Body:       row.MasterNote,
		UpdatedAt:  row.MasterNoteUpdatedAt,
	}, nil
}

func (s *Service) DeleteCampaign(id int64) error {
	if err := s.q.DeleteCampaign(s.ctx, id); err != nil {
		return fmt.Errorf("deleting campaign: %w", err)
	}
	return nil
}

func (s *Service) AdjustCampaignFear(campaignID int64, delta int) (int, error) {
	row, err := s.q.AdjustCampaignFear(s.ctx, db.AdjustCampaignFearParams{
		Delta: int64(delta), ID: campaignID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, notFound("campaign", fmt.Sprint(campaignID))
	}
	if err != nil {
		return 0, fmt.Errorf("adjusting campaign fear: %w", err)
	}
	return int(row.CurrentFear), nil
}

func (s *Service) SetCampaignFear(campaignID int64, value int) (int, error) {
	row, err := s.q.SetCampaignFear(s.ctx, db.SetCampaignFearParams{
		Fear: int64(value), ID: campaignID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, notFound("campaign", fmt.Sprint(campaignID))
	}
	if err != nil {
		return 0, fmt.Errorf("setting campaign fear: %w", err)
	}
	return int(row.CurrentFear), nil
}
