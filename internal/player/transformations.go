package player

import (
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

const (
	martialArtistSubclass = "martialArtist"
	martialStancesSlug    = "martialStances"
)

func (s *Service) ListTransformations() []cards.Transformation {
	return s.catalog.ListTransformations()
}

func (s *Service) SetTransformation(characterID int64, slug string) (Character, error) {
	slug = strings.TrimSpace(slug)
	if slug != "" {
		if _, ok := s.catalog.Transformation(slug); !ok {
			return Character{}, notFound("transformation", slug)
		}
	}
	row, err := s.q.SetCharacterTransformation(s.ctx, db.SetCharacterTransformationParams{
		TransformationSlug: slug,
		ID:                 characterID,
	})
	return s.vitalsResult(characterID, row, err, "setting transformation")
}

func (s *Service) stanceSheet(row db.Character) *cards.Stances {
	if row.SubclassSlug != martialArtistSubclass && row.MulticlassSubclassSlug != martialArtistSubclass {
		return nil
	}
	sheet, ok := s.catalog.StanceSheet(martialStancesSlug)
	if !ok {
		return nil
	}
	return &sheet
}
