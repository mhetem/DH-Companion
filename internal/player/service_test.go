package player

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/srd"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func newTestService(t *testing.T) *Service {
	t.Helper()

	conn, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	goose.SetBaseFS(os.DirFS("../../sql/schema"))
	t.Cleanup(func() { goose.SetBaseFS(nil) })
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("setting goose dialect: %v", err)
	}
	goose.SetLogger(goose.NopLogger())
	if err := goose.Up(conn, "."); err != nil {
		t.Fatalf("migrating test db: %v", err)
	}

	catalog, err := srd.Default()
	if err != nil {
		t.Fatalf("loading srd: %v", err)
	}

	s := New()
	Attach(s, context.Background(), db.New(conn), conn, catalog)
	return s
}

func newTestCharacter(t *testing.T, s *Service, armorScore, major, severe int) Character {
	t.Helper()
	return buildCharacter(t, s, "bard", "troubadour", "clank", armorScore, major, severe)
}

func buildCharacter(t *testing.T, s *Service, class, subclass, ancestry string, armorScore, major, severe int) Character {
	t.Helper()
	c, err := s.SaveCharacter(CharacterInput{
		Name:            "Sabine",
		ClassSlug:       class,
		SubclassSlug:    subclass,
		AncestrySlug:    ancestry,
		CommunitySlug:   s.catalog.ListCommunities()[0].Slug,
		Traits:          map[string]int{"agility": 2, "strength": 1, "finesse": 1, "instinct": 0, "presence": 0, "knowledge": -1},
		Evasion:         10,
		ArmorScore:      armorScore,
		ThresholdMajor:  major,
		ThresholdSevere: severe,
	})
	if err != nil {
		t.Fatalf("SaveCharacter: %v", err)
	}
	return c
}
