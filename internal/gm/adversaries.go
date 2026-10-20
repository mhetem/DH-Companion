package gm

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

type adversaryColumns struct {
	name     string
	attack   sql.NullString
	features sql.NullString
	size     string
	segments string
}

func isColossus(a cards.Adversary) bool {
	return strings.TrimSpace(a.Type) == cards.TypeColossus
}

func shapeForType(a cards.Adversary) cards.Adversary {
	if isColossus(a) {
		a.Difficulty, a.Hp, a.HordeNumber = "", "", ""
		a.StandardAttack = cards.Attack{}
		return a
	}
	a.Size, a.Segments = "", nil
	return a
}

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
		Size:           r.Size,
		Segments:       []cards.Segment{},
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
	if r.Segments != "" {
		if err := json.Unmarshal([]byte(r.Segments), &a.Segments); err != nil {
			return a, fmt.Errorf("adversary %q has unreadable segments: %w", r.Slug, err)
		}
	}
	return a, nil
}

func validateAdversary(a cards.Adversary) (adversaryColumns, error) {
	var cols adversaryColumns
	name, err := validateName(a.Name)
	if err != nil {
		return cols, err
	}
	if err := validateTier(a.Tier); err != nil {
		return cols, err
	}
	if err := validateAdversaryType(a.Type); err != nil {
		return cols, err
	}

	segments := []cards.Segment{}
	if isColossus(a) {
		if segments, err = cleanSegments(a.Segments); err != nil {
			return cols, err
		}
	}

	attackJSON, err := json.Marshal(a.StandardAttack)
	if err != nil {
		return cols, fmt.Errorf("encoding standard attack: %w", err)
	}
	if a.Features == nil {
		a.Features = []cards.Feature{}
	}
	featuresJSON, err := json.Marshal(a.Features)
	if err != nil {
		return cols, fmt.Errorf("encoding features: %w", err)
	}
	segmentsJSON, err := json.Marshal(segments)
	if err != nil {
		return cols, fmt.Errorf("encoding segments: %w", err)
	}
	return adversaryColumns{
		name:     name,
		attack:   sql.NullString{String: string(attackJSON), Valid: true},
		features: sql.NullString{String: string(featuresJSON), Valid: true},
		size:     strings.TrimSpace(a.Size),
		segments: string(segmentsJSON),
	}, nil
}

func cleanSegments(in []cards.Segment) ([]cards.Segment, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("a colossus needs at least one segment")
	}
	out := make([]cards.Segment, 0, len(in))
	seen := map[string]bool{}
	for i, seg := range in {
		seg.Name = strings.TrimSpace(seg.Name)
		if seg.Name == "" {
			return nil, fmt.Errorf("segment %d needs a name", i+1)
		}
		key := strings.ToLower(seg.Name)
		if seen[key] {
			return nil, fmt.Errorf("two segments are named %q — the combat runner tells segments apart by name", seg.Name)
		}
		seen[key] = true
		if seg.Count < 1 {
			seg.Count = 1
		}
		if seg.Count > maxSegmentCount {
			return nil, fmt.Errorf("segment %q: a colossus can have at most %d of one segment", seg.Name, maxSegmentCount)
		}
		if seg.Features == nil {
			seg.Features = []cards.Feature{}
		}
		out = append(out, seg)
	}
	return out, nil
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
	a = shapeForType(a)
	cols, err := validateAdversary(a)
	if err != nil {
		return cards.Adversary{}, err
	}
	name := cols.name
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
		StandardAttack: cols.attack,
		Features:       cols.features,
		Size:           cols.size,
		Segments:       cols.segments,
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
	a = shapeForType(a)
	cols, err := validateAdversary(a)
	if err != nil {
		return cards.Adversary{}, err
	}

	r, err := s.q.UpdateCustomAdversary(s.ctx, db.UpdateCustomAdversaryParams{
		Name:           cols.name,
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
		StandardAttack: cols.attack,
		Features:       cols.features,
		Size:           cols.size,
		Segments:       cols.segments,
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
