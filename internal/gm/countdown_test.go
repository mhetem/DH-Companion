package gm

import (
	"slices"
	"testing"
)

func TestCountdownBeatsAreCleanedAndOrdered(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "The Marius Contract")

	clock, err := s.SaveCountdown(CountdownInput{
		CampaignID: &c.ID,
		Name:       "The invasion",
		Max:        8,
		Beats: []CountdownBeat{
			{At: 0, Text: "Open war."},
			{At: 8, Text: "  The deal is struck.  "},
			{At: 5, Text: "   "},
			{At: 9, Text: "Past the top."},
			{At: -1, Text: "Below zero."},
			{At: 4, Text: "First draft."},
			{At: 4, Text: "A bridgehead in the captured city."},
		},
	})
	if err != nil {
		t.Fatalf("SaveCountdown: %v", err)
	}

	want := []CountdownBeat{
		{At: 8, Text: "The deal is struck."},
		{At: 4, Text: "A bridgehead in the captured city."},
		{At: 0, Text: "Open war."},
	}
	if !slices.Equal(clock.Beats, want) {
		t.Fatalf("saved beats = %+v, want %+v", clock.Beats, want)
	}

	got, err := s.GetCountdown(clock.ID)
	if err != nil {
		t.Fatalf("GetCountdown: %v", err)
	}
	if !slices.Equal(got.Beats, want) {
		t.Fatalf("reloaded beats = %+v, want %+v", got.Beats, want)
	}
}

func TestCountdownWithoutBeatsHasEmptyTrail(t *testing.T) {
	s := newTestService(t)

	clock, err := s.SaveCountdown(CountdownInput{Name: "Reinforcements", Max: 4})
	if err != nil {
		t.Fatalf("SaveCountdown: %v", err)
	}
	if clock.Beats == nil || len(clock.Beats) != 0 {
		t.Fatalf("beats = %#v, want an empty non-nil slice", clock.Beats)
	}

	clocks, err := s.ListUnassignedCountdowns()
	if err != nil {
		t.Fatalf("ListUnassignedCountdowns: %v", err)
	}
	if len(clocks) != 1 || clocks[0].Beats == nil {
		t.Fatalf("listed clocks = %#v", clocks)
	}
}

func TestAdoptUnassignedCountdowns(t *testing.T) {
	s := newTestService(t)
	home := newTestCampaign(t, s, "The Marius Contract")
	other := newTestCampaign(t, s, "Age of Umbra")

	for _, in := range []CountdownInput{
		{Name: "Reinforcements", Max: 4},
		{Name: "The tide", Max: 6},
		{CampaignID: &other.ID, Name: "Already home", Max: 3},
	} {
		if _, err := s.SaveCountdown(in); err != nil {
			t.Fatalf("SaveCountdown(%q): %v", in.Name, err)
		}
	}

	n, err := s.AdoptUnassignedCountdowns(home.ID)
	if err != nil {
		t.Fatalf("AdoptUnassignedCountdowns: %v", err)
	}
	if n != 2 {
		t.Fatalf("moved %d countdowns, want 2", n)
	}
	if loose, _ := s.ListUnassignedCountdowns(); len(loose) != 0 {
		t.Fatalf("still unassigned: %+v", loose)
	}
	adopted, _ := s.ListCountdownsForCampaign(home.ID)
	if len(adopted) != 2 {
		t.Fatalf("campaign holds %d countdowns, want 2", len(adopted))
	}
	kept, _ := s.ListCountdownsForCampaign(other.ID)
	if len(kept) != 1 || kept[0].Name != "Already home" {
		t.Fatalf("the other campaign's clock moved: %+v", kept)
	}

	if _, err := s.AdoptUnassignedCountdowns(9999); err == nil {
		t.Fatal("adopting into a missing campaign should fail")
	}
}

func TestCountdownBeatsSurviveTicksAndShrink(t *testing.T) {
	s := newTestService(t)

	clock, err := s.SaveCountdown(CountdownInput{
		Name: "Ritual",
		Max:  6,
		Beats: []CountdownBeat{
			{At: 6, Text: "The chanting begins."},
			{At: 3, Text: "The sky darkens."},
			{At: 0, Text: "The gate opens."},
		},
	})
	if err != nil {
		t.Fatalf("SaveCountdown: %v", err)
	}

	ticked, err := s.AdvanceCountdown(clock.ID, 3)
	if err != nil {
		t.Fatalf("AdvanceCountdown: %v", err)
	}
	if ticked.Value != 3 || !slices.Equal(ticked.Beats, clock.Beats) {
		t.Fatalf("after ticking: value %d, beats %+v", ticked.Value, ticked.Beats)
	}

	shrunk, err := s.SaveCountdown(CountdownInput{
		ID:    &clock.ID,
		Name:  clock.Name,
		Value: ticked.Value,
		Max:   4,
		Beats: ticked.Beats,
	})
	if err != nil {
		t.Fatalf("SaveCountdown (shrink): %v", err)
	}
	want := []CountdownBeat{
		{At: 3, Text: "The sky darkens."},
		{At: 0, Text: "The gate opens."},
	}
	if shrunk.Value != 3 || !slices.Equal(shrunk.Beats, want) {
		t.Fatalf("after shrinking: value %d, beats %+v, want %+v", shrunk.Value, shrunk.Beats, want)
	}
}

func TestCountdownBeatsLibraryRoundTrip(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "The Marius Contract")

	campaignBeats := []CountdownBeat{
		{At: 2, Text: "Marius presses the realm's army."},
		{At: 0, Text: "The invasion becomes open war."},
	}
	if _, err := s.SaveCountdown(CountdownInput{CampaignID: &c.ID, Name: "The invasion", Max: 2, Beats: campaignBeats}); err != nil {
		t.Fatalf("SaveCountdown (campaign): %v", err)
	}
	looseBeats := []CountdownBeat{{At: 3, Text: "The bridge starts to give."}}
	if _, err := s.SaveCountdown(CountdownInput{Name: "Collapse", Max: 3, Beats: looseBeats}); err != nil {
		t.Fatalf("SaveCountdown (loose): %v", err)
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
	if report.Countdowns != 2 || len(report.Skipped) != 0 {
		t.Fatalf("unexpected report %+v", report)
	}

	campaigns, err := fresh.ListCampaigns()
	if err != nil || len(campaigns) != 1 {
		t.Fatalf("ListCampaigns = %+v, %v", campaigns, err)
	}
	clocks, err := fresh.ListCountdownsForCampaign(campaigns[0].ID)
	if err != nil {
		t.Fatalf("ListCountdownsForCampaign: %v", err)
	}
	if len(clocks) != 1 || !slices.Equal(clocks[0].Beats, campaignBeats) {
		t.Fatalf("campaign clocks = %+v", clocks)
	}

	loose, err := fresh.ListUnassignedCountdowns()
	if err != nil {
		t.Fatalf("ListUnassignedCountdowns: %v", err)
	}
	if len(loose) != 1 || !slices.Equal(loose[0].Beats, looseBeats) {
		t.Fatalf("loose clocks = %+v", loose)
	}
}
