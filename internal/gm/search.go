package gm

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

const (
	entityNote        = "note"
	entityMaster      = "master"
	entityAdversary   = "adversary"
	entityEnvironment = "environment"
)

const searchSQL = `
SELECT entity, entity_id, campaign_id, slug, title,
       snippet(search, 5, '<mark>', '</mark>', '…', 12),
       bm25(search)
FROM search
WHERE search MATCH ?
  AND (? = 0 OR entity NOT IN ('note', 'master') OR campaign_id = ?)
ORDER BY rank
LIMIT ?`

const defaultSearchLimit = 50

type SearchHit struct {
	Entity     string  `json:"entity"`
	EntityID   int64   `json:"entityId"`
	CampaignID int64   `json:"campaignId"`
	Slug       string  `json:"slug"`
	Title      string  `json:"title"`
	Excerpt    string  `json:"excerpt"`
	Score      float64 `json:"score"`
}

func (s *Service) Search(query string, campaignID int64, limit int) ([]SearchHit, error) {
	match := matchExpr(query)
	if match == "" {
		return []SearchHit{}, nil
	}
	if limit <= 0 || limit > defaultSearchLimit {
		limit = defaultSearchLimit
	}

	rows, err := s.conn.QueryContext(s.ctx, searchSQL, match, campaignID, campaignID, limit)
	if err != nil {
		return nil, fmt.Errorf("searching: %w", err)
	}
	defer rows.Close()

	hits := []SearchHit{}
	for rows.Next() {
		var h SearchHit
		if err := rows.Scan(&h.Entity, &h.EntityID, &h.CampaignID, &h.Slug, &h.Title, &h.Excerpt, &h.Score); err != nil {
			return nil, fmt.Errorf("scanning search hit: %w", err)
		}
		hits = append(hits, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading search hits: %w", err)
	}
	return hits, nil
}

func (s *Service) ReindexCards() error {
	adversaries, err := s.BrowseAdversaries(Filter{})
	if err != nil {
		return err
	}
	environments, err := s.BrowseEnvironments(Filter{})
	if err != nil {
		return err
	}

	return s.tx(func(q *db.Queries) error {
		if err := q.ClearCardIndex(s.ctx); err != nil {
			return fmt.Errorf("clearing card index: %w", err)
		}
		for _, a := range adversaries {
			if err := q.IndexCard(s.ctx, db.IndexCardParams{
				Entity: entityAdversary,
				Slug:   a.Slug,
				Title:  a.Name,
				Body:   adversaryText(a.Adversary),
			}); err != nil {
				return fmt.Errorf("indexing adversary %q: %w", a.Slug, err)
			}
		}
		for _, e := range environments {
			if err := q.IndexCard(s.ctx, db.IndexCardParams{
				Entity: entityEnvironment,
				Slug:   e.Slug,
				Title:  e.Name,
				Body:   environmentText(e.Environment),
			}); err != nil {
				return fmt.Errorf("indexing environment %q: %w", e.Slug, err)
			}
		}
		return nil
	})
}

func adversaryText(a cards.Adversary) string {
	parts := []string{a.Type, a.Description, a.Motives, a.Experiences}
	parts = append(parts, featureText(a.Features)...)
	return joinText(parts)
}

func environmentText(e cards.Environment) string {
	parts := []string{e.Type, e.Description, e.Impulses}
	parts = append(parts, e.PotentialAdversaries...)
	parts = append(parts, featureText(e.Features)...)
	return joinText(parts)
}

func featureText(features []cards.Feature) []string {
	out := make([]string, 0, len(features)*2)
	for _, f := range features {
		out = append(out, f.Title, f.Description)
		out = append(out, f.Questions...)
	}
	return out
}

var htmlTag = regexp.MustCompile(`<[^>]*>`)

func joinText(parts []string) string {
	kept := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(htmlTag.ReplaceAllString(p, " "))
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, "\n")
}

func matchExpr(query string) string {
	tokens := strings.FieldsFunc(query, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	if len(tokens) == 0 {
		return ""
	}

	var b strings.Builder
	for i, tok := range tokens {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteByte('"')
		b.WriteString(tok)
		b.WriteByte('"')
		if i == len(tokens)-1 {
			b.WriteByte('*')
		}
	}
	return b.String()
}
