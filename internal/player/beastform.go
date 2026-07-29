package player

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

const BeastformClass = "druid"

type BeastformView struct {
	Eligible    bool              `json:"eligible"`
	Tier        int               `json:"tier"`
	Available   []cards.Beastform `json:"available"`
	Active      *cards.Beastform  `json:"active"`
	BaseEvasion int               `json:"baseEvasion"`
	Evasion     int               `json:"evasion"`
	TraitBonus  string            `json:"traitBonus"`
	Character   Character         `json:"character"`
}

func (s *Service) Beastforms(characterID int64) (BeastformView, error) {
	row, err := s.character(characterID)
	if err != nil {
		return BeastformView{}, err
	}
	return s.beastformView(row), nil
}

func (s *Service) Transform(characterID int64, slug string, markStress bool) (BeastformView, error) {
	row, err := s.character(characterID)
	if err != nil {
		return BeastformView{}, err
	}
	if !isBeastformer(row) {
		return BeastformView{}, fmt.Errorf("only a druid can Beastform")
	}

	slug = strings.TrimSpace(slug)
	form, ok := s.catalog.Beastform(slug)
	if !ok {
		return BeastformView{}, notFound("beastform", slug)
	}
	tier := tierForLevel(int(row.Level))
	if leadingInt(form.Tier, 1) > tier {
		return BeastformView{}, fmt.Errorf("%s is a tier %s form and you are tier %d", form.Name, form.Tier, tier)
	}

	if markStress {
		if int(row.StressMarked) >= int(row.StressMax) {
			return BeastformView{}, fmt.Errorf("no Stress left to mark")
		}
		if _, err := s.q.AdjustCharacterStress(s.ctx, db.AdjustCharacterStressParams{
			Delta: 1, ID: characterID,
		}); err != nil {
			return BeastformView{}, fmt.Errorf("marking stress: %w", err)
		}
	}

	updated, err := s.q.SetCharacterBeastform(s.ctx, db.SetCharacterBeastformParams{
		BeastformSlug: slug, ID: characterID,
	})
	if err != nil {
		return BeastformView{}, fmt.Errorf("transforming: %w", err)
	}
	return s.beastformView(updated), nil
}

func (s *Service) DropBeastform(characterID int64) (BeastformView, error) {
	if _, err := s.character(characterID); err != nil {
		return BeastformView{}, err
	}
	updated, err := s.q.SetCharacterBeastform(s.ctx, db.SetCharacterBeastformParams{
		BeastformSlug: "", ID: characterID,
	})
	if err != nil {
		return BeastformView{}, fmt.Errorf("dropping out of Beastform: %w", err)
	}
	return s.beastformView(updated), nil
}

func (s *Service) beastformView(row db.Character) BeastformView {
	tier := tierForLevel(int(row.Level))
	view := BeastformView{
		Eligible:    isBeastformer(row),
		Tier:        tier,
		Available:   s.catalog.BeastformsUpToTier(strconv.Itoa(tier)),
		BaseEvasion: int(row.Evasion),
		Evasion:     int(row.Evasion),
		Character:   s.characterView(row),
	}
	if view.Available == nil {
		view.Available = []cards.Beastform{}
	}
	if row.BeastformSlug != "" {
		if form, ok := s.catalog.Beastform(row.BeastformSlug); ok {
			view.Active = &form
			view.Evasion = clamp(int(row.Evasion)+signedInt(form.EvasionBonus), 0, 60)
			view.TraitBonus = strings.TrimSpace(form.Trait + " " + form.TraitBonus)
		}
	}
	return view
}

func isBeastformer(row db.Character) bool {
	return row.ClassSlug == BeastformClass || row.MulticlassSlug == BeastformClass
}

func signedInt(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	sign := 1
	switch raw[0] {
	case '+':
		raw = raw[1:]
	case '-':
		sign = -1
		raw = raw[1:]
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	return sign * n
}
