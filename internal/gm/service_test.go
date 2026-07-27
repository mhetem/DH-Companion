package gm

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/mhetem/DH-Companion/internal/cards"
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

func sampleAdversary() cards.Adversary {
	return cards.Adversary{
		Meta: cards.Meta{
			Name:        "Rust Wraith",
			Tier:        "2",
			Type:        "Skulk",
			Description: "A haunt of the scrapyard.",
		},
		Motives:        "Corrode, linger",
		Experiences:    "Ambush +2",
		Difficulty:     "14",
		ThresholdMinor: "8",
		ThresholdMajor: "15",
		Hp:             "5",
		Stress:         "4",
		StandardAttack: cards.Attack{Name: "Touch", Modifier: "+2", Range: "Melee", Damage: "1d8+2", DamageType: "phy"},
		Features: []cards.Feature{
			{Title: "Corrosive", Type: "Passive", Description: "Armor slots burn away."},
		},
	}
}

func TestPartyCRUD(t *testing.T) {
	s := newTestService(t)

	p, err := s.CreateParty(PartyInput{Name: "The Wayfarers", Size: 4, Tier: "2"})
	if err != nil {
		t.Fatalf("CreateParty: %v", err)
	}
	if p.ID == 0 || p.Size != 4 || p.Tier != "2" {
		t.Fatalf("unexpected party %+v", p)
	}

	got, err := s.GetParty(p.ID)
	if err != nil || got.Name != "The Wayfarers" {
		t.Fatalf("GetParty = %+v, %v", got, err)
	}

	updated, err := s.UpdateParty(p.ID, PartyInput{Name: "The Wayfarers", Size: 5, Tier: "3"})
	if err != nil {
		t.Fatalf("UpdateParty: %v", err)
	}
	if updated.Size != 5 || updated.Tier != "3" {
		t.Fatalf("update did not stick: %+v", updated)
	}

	list, err := s.ListParties()
	if err != nil || len(list) != 1 {
		t.Fatalf("ListParties = %v, %v", list, err)
	}

	if err := s.DeleteParty(p.ID); err != nil {
		t.Fatalf("DeleteParty: %v", err)
	}
	if _, err := s.GetParty(p.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestPartyValidation(t *testing.T) {
	s := newTestService(t)

	cases := map[string]PartyInput{
		"empty name": {Name: "  ", Size: 4, Tier: "1"},
		"zero size":  {Name: "Party", Size: 0, Tier: "1"},
		"bad tier":   {Name: "Party", Size: 4, Tier: "7"},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := s.CreateParty(in); err == nil {
				t.Fatalf("expected a validation error for %+v", in)
			}
		})
	}
}

func TestCustomAdversaryRoundTrip(t *testing.T) {
	s := newTestService(t)

	created, err := s.CreateCustomAdversary(sampleAdversary())
	if err != nil {
		t.Fatalf("CreateCustomAdversary: %v", err)
	}
	if created.Slug != "rust-wraith" {
		t.Fatalf("slug = %q, want rust-wraith", created.Slug)
	}
	if created.Kind != cards.KindAdversary {
		t.Fatalf("kind = %q, want adversary", created.Kind)
	}
	// The JSON-in-TEXT columns must survive the round trip decoded, not raw.
	if created.StandardAttack.Damage != "1d8+2" {
		t.Fatalf("standard attack lost: %+v", created.StandardAttack)
	}
	if len(created.Features) != 1 || created.Features[0].Title != "Corrosive" {
		t.Fatalf("features lost: %+v", created.Features)
	}

	edited := created
	edited.Name = "Rust Wraith Elder"
	edited.Tier = "3"
	updated, err := s.UpdateCustomAdversary(edited)
	if err != nil {
		t.Fatalf("UpdateCustomAdversary: %v", err)
	}
	if updated.Slug != "rust-wraith" {
		t.Fatalf("slug changed on rename: %q", updated.Slug)
	}
	if updated.Name != "Rust Wraith Elder" || updated.Tier != "3" {
		t.Fatalf("update did not stick: %+v", updated)
	}

	if err := s.DeleteCustomAdversary("rust-wraith"); err != nil {
		t.Fatalf("DeleteCustomAdversary: %v", err)
	}
	if _, err := s.GetCustomAdversary("rust-wraith"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCustomAdversaryDuplicateName(t *testing.T) {
	s := newTestService(t)

	if _, err := s.CreateCustomAdversary(sampleAdversary()); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := s.CreateCustomAdversary(sampleAdversary())
	if err == nil {
		t.Fatal("expected a duplicate-name error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error should be readable in the UI, got %q", err)
	}
	if strings.Contains(err.Error(), "UNIQUE") {
		t.Fatalf("raw SQL error leaked to the frontend: %q", err)
	}
}

func TestCustomAdversaryRejectsUnknownType(t *testing.T) {
	s := newTestService(t)

	a := sampleAdversary()
	a.Type = "Skirmisher" // not a Daggerheart type; would silently cost 2
	if _, err := s.CreateCustomAdversary(a); err == nil {
		t.Fatal("expected an unknown-type error")
	}
}

func TestBrowseAdversariesMergesSRDAndCustom(t *testing.T) {
	s := newTestService(t)

	if _, err := s.CreateCustomAdversary(sampleAdversary()); err != nil {
		t.Fatalf("create: %v", err)
	}

	all, err := s.BrowseAdversaries(Filter{})
	if err != nil {
		t.Fatalf("BrowseAdversaries: %v", err)
	}
	var srdCount, customCount int
	for _, a := range all {
		switch a.Source {
		case "srd":
			srdCount++
		case "custom":
			customCount++
		default:
			t.Fatalf("card %q has no source", a.Slug)
		}
	}
	if srdCount == 0 {
		t.Fatal("no SRD adversaries in the merged browse")
	}
	if customCount != 1 {
		t.Fatalf("custom count = %d, want 1", customCount)
	}
	// Sorted by name, case-insensitively.
	for i := 1; i < len(all); i++ {
		if strings.ToLower(all[i-1].Name) > strings.ToLower(all[i].Name) {
			t.Fatalf("browse is not sorted: %q before %q", all[i-1].Name, all[i].Name)
		}
	}

	filtered, err := s.BrowseAdversaries(Filter{Tier: "2", Type: "Skulk"})
	if err != nil {
		t.Fatalf("filtered browse: %v", err)
	}
	for _, a := range filtered {
		if a.Tier != "2" || a.Type != "Skulk" {
			t.Fatalf("filter leaked %+v", a.Meta)
		}
	}
}

func TestEncounterHydratesPicksEnvironmentAndBudget(t *testing.T) {
	s := newTestService(t)

	party, err := s.CreateParty(PartyInput{Name: "Party", Size: 4, Tier: "1"})
	if err != nil {
		t.Fatalf("CreateParty: %v", err)
	}
	custom, err := s.CreateCustomAdversary(sampleAdversary())
	if err != nil {
		t.Fatalf("CreateCustomAdversary: %v", err)
	}

	envSlug := "abandonedGrove"
	view, err := s.SaveEncounter(EncounterInput{
		Name:              "Grove Ambush",
		PartyID:           &party.ID,
		EnvironmentSlug:   &envSlug,
		Adversaries:       []Pick{{Slug: "acidBurrower", Count: 1}, {Slug: "dropped", Count: 0}},
		CustomAdversaries: []Pick{{Slug: custom.Slug, Count: 2}},
	})
	if err != nil {
		t.Fatalf("SaveEncounter: %v", err)
	}
	if view.ID == 0 {
		t.Fatal("create did not return an id")
	}
	if len(view.Adversaries) != 2 {
		t.Fatalf("want 2 picks (zero-count dropped), got %d", len(view.Adversaries))
	}
	if view.TotalCount != 3 {
		t.Fatalf("TotalCount = %d, want 3", view.TotalCount)
	}
	if view.Adversaries[0].Source != "srd" || view.Adversaries[0].Name == "" {
		t.Fatalf("SRD pick not resolved: %+v", view.Adversaries[0].Meta)
	}
	if view.Adversaries[1].Source != "custom" || view.Adversaries[1].Difficulty != "14" {
		t.Fatalf("custom pick not resolved: %+v", view.Adversaries[1])
	}
	if view.Environment == nil || view.Environment.Name == "" {
		t.Fatalf("environment not resolved: %+v", view.Environment)
	}
	if view.Budget == nil {
		t.Fatal("budget not computed for an attached party")
	}
	// The exact total is encounter_math's business and is covered by its own
	// tests; here it only matters that the party fed the settings.
	if view.Budget.PartySize != 4 || view.Budget.Spent == 0 {
		t.Fatalf("budget looks wrong: %+v", *view.Budget)
	}

	// Update through the same entry point.
	view.Name = "Grove Ambush, Redux"
	updated, err := s.SaveEncounter(EncounterInput{
		ID:                &view.ID,
		Name:              "Grove Ambush, Redux",
		PartyID:           &party.ID,
		EnvironmentSlug:   &envSlug,
		Adversaries:       []Pick{{Slug: "acidBurrower", Count: 1}},
		CustomAdversaries: []Pick{{Slug: custom.Slug, Count: 2}},
	})
	if err != nil {
		t.Fatalf("SaveEncounter update: %v", err)
	}
	if updated.ID != view.ID || updated.Name != "Grove Ambush, Redux" {
		t.Fatalf("update created a new row or lost the name: %+v", updated)
	}

	list, err := s.ListEncounters()
	if err != nil || len(list) != 1 || list[0].TotalCount != 3 {
		t.Fatalf("ListEncounters = %+v, %v", list, err)
	}
}

func TestEncounterSurvivesDeletedCustomAdversary(t *testing.T) {
	s := newTestService(t)

	custom, err := s.CreateCustomAdversary(sampleAdversary())
	if err != nil {
		t.Fatalf("CreateCustomAdversary: %v", err)
	}
	view, err := s.SaveEncounter(EncounterInput{
		Name:              "Doomed Reference",
		CustomAdversaries: []Pick{{Slug: custom.Slug, Count: 1}},
	})
	if err != nil {
		t.Fatalf("SaveEncounter: %v", err)
	}

	if err := s.DeleteCustomAdversary(custom.Slug); err != nil {
		t.Fatalf("DeleteCustomAdversary: %v", err)
	}

	reopened, err := s.GetEncounter(view.ID)
	if err != nil {
		t.Fatalf("a deleted card must not stop the encounter from opening: %v", err)
	}
	if len(reopened.Adversaries) != 1 || !reopened.Adversaries[0].Unresolved {
		t.Fatalf("pick should be flagged unresolved: %+v", reopened.Adversaries)
	}
	if reopened.Adversaries[0].Slug != custom.Slug {
		t.Fatalf("unresolved pick lost its slug: %+v", reopened.Adversaries[0].Meta)
	}
}

func TestDeletingPartyDetachesEncounter(t *testing.T) {
	s := newTestService(t)

	party, err := s.CreateParty(PartyInput{Name: "Party", Size: 4, Tier: "1"})
	if err != nil {
		t.Fatalf("CreateParty: %v", err)
	}
	view, err := s.SaveEncounter(EncounterInput{
		Name:        "Fight",
		PartyID:     &party.ID,
		Adversaries: []Pick{{Slug: "acidBurrower", Count: 1}},
	})
	if err != nil {
		t.Fatalf("SaveEncounter: %v", err)
	}

	if err := s.DeleteParty(party.ID); err != nil {
		t.Fatalf("DeleteParty: %v", err)
	}

	reopened, err := s.GetEncounter(view.ID)
	if err != nil {
		t.Fatalf("GetEncounter after party delete: %v", err)
	}
	if reopened.PartyID != nil {
		t.Fatalf("party_id should be nulled by ON DELETE SET NULL, got %v", *reopened.PartyID)
	}
	if reopened.Budget != nil {
		t.Fatal("budget should be nil with no party attached")
	}
}

func TestCustomEnvironmentRoundTrip(t *testing.T) {
	s := newTestService(t)

	created, err := s.CreateCustomEnvironment(cards.Environment{
		Meta: cards.Meta{
			Name:        "Sunken Archive",
			Tier:        "3",
			Type:        "Exploration",
			Description: "Shelves half-drowned in black water.",
		},
		Difficulty:           "16",
		Impulses:             "Drown knowledge, hoard secrets",
		PotentialAdversaries: []string{"acidBurrower"},
		Features: []cards.Feature{
			{Title: "Rising Water", Type: "Action", Description: "The tide climbs.", Questions: []string{"What does it uncover?"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateCustomEnvironment: %v", err)
	}
	if created.Slug != "sunken-archive" || created.Kind != cards.KindEnvironment {
		t.Fatalf("unexpected meta: %+v", created.Meta)
	}
	if len(created.PotentialAdversaries) != 1 || len(created.Features) != 1 {
		t.Fatalf("json columns lost: %+v", created)
	}
	if len(created.Features[0].Questions) != 1 {
		t.Fatalf("feature questions lost: %+v", created.Features[0])
	}

	got, err := s.GetCustomEnvironment("sunken-archive")
	if err != nil || got.Impulses != "Drown knowledge, hoard secrets" {
		t.Fatalf("GetCustomEnvironment = %+v, %v", got, err)
	}

	if _, err := s.BrowseEnvironments(Filter{Tier: "3"}); err != nil {
		t.Fatalf("BrowseEnvironments: %v", err)
	}
}
