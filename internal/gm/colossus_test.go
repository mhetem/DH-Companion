package gm

import (
	"strings"
	"testing"

	"github.com/mhetem/DH-Companion/internal/cards"
)

func sampleColossus() cards.Adversary {
	return cards.Adversary{
		Meta: cards.Meta{
			Name:        "Basalt Strider",
			Tier:        "1",
			Type:        cards.TypeColossus,
			Description: "A walking cliff.",
		},
		Motives:        "Stomp, shrug off, loom",
		Experiences:    "Huge +2",
		ThresholdMinor: "11",
		ThresholdMajor: "22",
		Stress:         "6",
		Size:           "30m tall",
		Features: []cards.Feature{
			{Title: "Rockfall", Type: "Reaction", Description: "When it misses, gain a Fear."},
		},
		Segments: []cards.Segment{
			{
				Name:           " Head ",
				Count:          1,
				Adjacent:       "Torso",
				Difficulty:     "16",
				Hp:             "5",
				StandardAttack: cards.Attack{Name: "Bite", Modifier: "+2", Range: "Melee", Damage: "1d10+1", DamageType: "phy"},
			},
			{Name: "Arm", Count: 2, Adjacent: "Torso", Difficulty: "13", Hp: "3 (each)"},
			{Name: "Core", Adjacent: "Head, Arm", Difficulty: "15", Hp: "none"},
		},
	}
}

func TestColossusRoundTrip(t *testing.T) {
	s := newTestService(t)

	draft := sampleColossus()
	draft.Difficulty, draft.Hp = "14", "8"
	draft.StandardAttack = cards.Attack{Name: "Claws", Damage: "1d12+2"}
	created, err := s.CreateCustomAdversary(draft)
	if err != nil {
		t.Fatalf("CreateCustomAdversary: %v", err)
	}
	if created.Size != "30m tall" {
		t.Fatalf("size lost: %q", created.Size)
	}
	if created.Difficulty != "" || created.Hp != "" || created.StandardAttack.Name != "" {
		t.Fatalf("a colossus kept card-level stats its segments own: %q %q %+v", created.Difficulty, created.Hp, created.StandardAttack)
	}
	if len(created.Segments) != 3 {
		t.Fatalf("segments lost: %+v", created.Segments)
	}
	if created.Segments[0].Name != "Head" {
		t.Fatalf("segment name not trimmed: %q", created.Segments[0].Name)
	}
	if created.Segments[0].StandardAttack.Damage != "1d10+1" {
		t.Fatalf("segment attack lost: %+v", created.Segments[0].StandardAttack)
	}
	if created.Segments[2].Count != 1 {
		t.Fatalf("missing count should default to 1, got %d", created.Segments[2].Count)
	}
	if created.Segments[1].Features == nil {
		t.Fatal("segment features should decode as an empty list, not null")
	}

	fetched, err := s.GetCustomAdversary(created.Slug)
	if err != nil {
		t.Fatalf("GetCustomAdversary: %v", err)
	}
	if len(fetched.Segments) != 3 || fetched.Segments[1].Count != 2 {
		t.Fatalf("segments did not survive a read: %+v", fetched.Segments)
	}

	demoted := fetched
	demoted.Type = "Solo"
	updated, err := s.UpdateCustomAdversary(demoted)
	if err != nil {
		t.Fatalf("UpdateCustomAdversary: %v", err)
	}
	if len(updated.Segments) != 0 || updated.Size != "" {
		t.Fatalf("a non-colossus kept its segments: %+v / %q", updated.Segments, updated.Size)
	}
}

func TestColossusValidation(t *testing.T) {
	s := newTestService(t)

	tests := []struct {
		name   string
		mutate func(a *cards.Adversary)
		want   string
	}{
		{
			name:   "no segments",
			mutate: func(a *cards.Adversary) { a.Segments = nil },
			want:   "at least one segment",
		},
		{
			name:   "blank segment name",
			mutate: func(a *cards.Adversary) { a.Segments[1].Name = "  " },
			want:   "segment 2 needs a name",
		},
		{
			name:   "duplicate segment names",
			mutate: func(a *cards.Adversary) { a.Segments[2].Name = "arm" },
			want:   "two segments are named",
		},
		{
			name:   "too many of one segment",
			mutate: func(a *cards.Adversary) { a.Segments[1].Count = maxSegmentCount + 1 },
			want:   "at most",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := sampleColossus()
			tt.mutate(&a)
			_, err := s.CreateCustomAdversary(a)
			if err == nil {
				t.Fatal("expected a validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %q should mention %q", err, tt.want)
			}
		})
	}
}

func assertColossusIntact(t *testing.T, card cards.Adversary) {
	t.Helper()
	if card.Type != cards.TypeColossus || card.Size != "30m tall" {
		t.Fatalf("framework lost: type %q size %q", card.Type, card.Size)
	}
	if card.ThresholdMinor != "11" || card.ThresholdMajor != "22" || card.Stress != "6" {
		t.Fatalf("framework stats lost: %q/%q stress %q", card.ThresholdMinor, card.ThresholdMajor, card.Stress)
	}
	if len(card.Features) != 1 || card.Features[0].Title != "Rockfall" {
		t.Fatalf("colossus features lost: %+v", card.Features)
	}
	if len(card.Segments) != 3 {
		t.Fatalf("segments lost: %+v", card.Segments)
	}
	head, arm, core := card.Segments[0], card.Segments[1], card.Segments[2]
	if head.Name != "Head" || head.Difficulty != "16" || head.Hp != "5" || head.StandardAttack.Damage != "1d10+1" {
		t.Fatalf("head segment lost: %+v", head)
	}
	if arm.Name != "Arm" || arm.Count != 2 || arm.Adjacent != "Torso" || arm.Hp != "3 (each)" {
		t.Fatalf("arm segment lost: %+v", arm)
	}
	if core.Name != "Core" || core.Hp != "none" {
		t.Fatalf("core segment lost: %+v", core)
	}
}

func TestColossusLibraryRoundTrip(t *testing.T) {
	s := newTestService(t)

	card, err := s.CreateCustomAdversary(sampleColossus())
	if err != nil {
		t.Fatalf("CreateCustomAdversary: %v", err)
	}
	if _, err := s.SaveEncounter(EncounterInput{
		Name:              "The Strider",
		CustomAdversaries: []Pick{{Slug: card.Slug, Count: 1}},
	}); err != nil {
		t.Fatalf("SaveEncounter: %v", err)
	}

	raw, err := s.ExportLibraryJSON()
	if err != nil {
		t.Fatalf("ExportLibraryJSON: %v", err)
	}

	fresh := newTestService(t)
	report, err := fresh.ImportLibraryJSON(raw)
	if err != nil {
		t.Fatalf("ImportLibraryJSON: %v", err)
	}
	if report.CustomAdversaries != 1 || report.Encounters != 1 || len(report.Skipped) != 0 {
		t.Fatalf("unexpected report %+v", report)
	}

	imported, err := fresh.GetCustomAdversary(card.Slug)
	if err != nil {
		t.Fatalf("GetCustomAdversary: %v", err)
	}
	assertColossusIntact(t, imported)

	encounters, err := fresh.ListEncounters()
	if err != nil || len(encounters) != 1 {
		t.Fatalf("ListEncounters = %+v, %v", encounters, err)
	}
	enc, err := fresh.GetEncounter(encounters[0].ID)
	if err != nil {
		t.Fatalf("GetEncounter: %v", err)
	}
	if len(enc.Adversaries) != 1 || enc.Adversaries[0].Unresolved || enc.Adversaries[0].Type != cards.TypeColossus {
		t.Fatalf("encounter pick did not follow the colossus: %+v", enc.Adversaries)
	}

	combat, err := fresh.StartCombat(enc.ID, nil, nil)
	if err != nil {
		t.Fatalf("StartCombat: %v", err)
	}
	if len(combat.Combatants) != 5 {
		t.Fatalf("imported colossus should spawn itself plus 4 segments, got %d", len(combat.Combatants))
	}
}

func TestColossusShareCodeRoundTrip(t *testing.T) {
	s := newTestService(t)

	card, err := s.CreateCustomAdversary(sampleColossus())
	if err != nil {
		t.Fatalf("CreateCustomAdversary: %v", err)
	}
	code, err := s.ShareAdversary(card.Slug)
	if err != nil {
		t.Fatalf("ShareAdversary: %v", err)
	}

	fresh := newTestService(t)
	preview, err := fresh.PreviewShareCode(code)
	if err != nil {
		t.Fatalf("PreviewShareCode: %v", err)
	}
	if preview.Type != cards.TypeColossus {
		t.Fatalf("preview type = %q, want Colossus", preview.Type)
	}

	saved, err := fresh.ImportShareCode(code)
	if err != nil {
		t.Fatalf("ImportShareCode: %v", err)
	}
	imported, err := fresh.GetCustomAdversary(saved.Slug)
	if err != nil {
		t.Fatalf("GetCustomAdversary: %v", err)
	}
	assertColossusIntact(t, imported)
}

func TestStartCombatSpawnsColossusSegments(t *testing.T) {
	s := newTestService(t)

	card, err := s.CreateCustomAdversary(sampleColossus())
	if err != nil {
		t.Fatalf("CreateCustomAdversary: %v", err)
	}
	enc, err := s.SaveEncounter(EncounterInput{
		Name:              "The Strider",
		CustomAdversaries: []Pick{{Slug: card.Slug, Count: 1}},
	})
	if err != nil {
		t.Fatalf("SaveEncounter: %v", err)
	}

	combat, err := s.StartCombat(enc.ID, nil, nil)
	if err != nil {
		t.Fatalf("StartCombat: %v", err)
	}
	if len(combat.Combatants) != 5 {
		t.Fatalf("want the colossus plus 4 segments, got %d: %+v", len(combat.Combatants), combat.Combatants)
	}

	frame := combat.Combatants[0]
	if frame.ColossusID != nil || frame.DisplayName != "Basalt Strider" {
		t.Fatalf("first row should be the colossus itself: %+v", frame)
	}
	if frame.HpMax != 0 || frame.StressMax != 6 {
		t.Fatalf("the colossus holds Stress, not HP: hp %d stress %d", frame.HpMax, frame.StressMax)
	}

	wantSegments := []struct {
		display string
		segment string
		hp      int
	}{
		{"Head", "Head", 5},
		{"Arm 1", "Arm", 3},
		{"Arm 2", "Arm", 3},
		{"Core", "Core", 0},
	}
	for i, want := range wantSegments {
		got := combat.Combatants[i+1]
		if got.ColossusID == nil || *got.ColossusID != frame.ID {
			t.Fatalf("%s is not attached to its colossus: %+v", want.display, got.ColossusID)
		}
		if got.DisplayName != want.display || got.SegmentName != want.segment {
			t.Fatalf("segment %d = %q (%q), want %q (%q)", i, got.DisplayName, got.SegmentName, want.display, want.segment)
		}
		if got.HpMax != want.hp || got.StressMax != 0 {
			t.Fatalf("%s: hp %d stress %d, want hp %d stress 0", want.display, got.HpMax, got.StressMax, want.hp)
		}
		if got.Segment == nil || got.Segment.Name != want.segment {
			t.Fatalf("%s did not resolve its segment card: %+v", want.display, got.Segment)
		}
	}

	arm := combat.Combatants[2]
	broken, err := s.SetBroken(arm.ID, true)
	if err != nil {
		t.Fatalf("SetBroken: %v", err)
	}
	if !broken.Broken {
		t.Fatal("segment did not break")
	}
	mended, err := s.SetBroken(arm.ID, false)
	if err != nil {
		t.Fatalf("SetBroken: %v", err)
	}
	if mended.Broken {
		t.Fatal("segment did not mend")
	}

	if err := s.RemoveCombatant(frame.ID); err != nil {
		t.Fatalf("RemoveCombatant: %v", err)
	}
	after, err := s.GetCombat(combat.ID)
	if err != nil {
		t.Fatalf("GetCombat: %v", err)
	}
	if len(after.Combatants) != 0 {
		t.Fatalf("removing the colossus should take its segments along, %d left", len(after.Combatants))
	}
}
