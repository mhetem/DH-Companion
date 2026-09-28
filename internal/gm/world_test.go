package gm

import (
	"strings"
	"testing"
)

func newTestCampaign(t *testing.T, s *Service, name string) Campaign {
	t.Helper()
	c, err := s.SaveCampaign(CampaignInput{Name: name})
	if err != nil {
		t.Fatalf("SaveCampaign: %v", err)
	}
	return c
}

func newWorldPage(t *testing.T, s *Service, campaignID int64, parent *WorldNote, kind, title string) WorldNote {
	t.Helper()
	in := WorldNoteInput{CampaignID: campaignID, Kind: kind, Title: title}
	if parent != nil {
		in.ParentID = &parent.ID
	}
	n, err := s.SaveWorldNote(in)
	if err != nil {
		t.Fatalf("SaveWorldNote(%q): %v", title, err)
	}
	return n
}

func worldTitles(t *testing.T, s *Service, campaignID int64, parentID *int64) []string {
	t.Helper()
	notes, err := s.ListWorldNotes(campaignID)
	if err != nil {
		t.Fatalf("ListWorldNotes: %v", err)
	}
	out := []string{}
	for _, n := range notes {
		if (n.ParentID == nil && parentID == nil) || (n.ParentID != nil && parentID != nil && *n.ParentID == *parentID) {
			out = append(out, n.Title)
		}
	}
	return out
}

func TestWorldNotesNest(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	region := newWorldPage(t, s, c.ID, nil, "Region", "The Drowned South")
	city := newWorldPage(t, s, c.ID, &region, "city", "Ashfall")
	place := newWorldPage(t, s, c.ID, &city, "place", "The Ferryman's Rest")

	if region.Kind != "region" || region.ParentID != nil {
		t.Fatalf("unexpected root %+v", region)
	}
	if city.ParentID == nil || *city.ParentID != region.ID {
		t.Fatalf("city not under region: %+v", city)
	}
	if place.ParentID == nil || *place.ParentID != city.ID {
		t.Fatalf("place not under city: %+v", place)
	}

	blank, err := s.SaveWorldNote(WorldNoteInput{CampaignID: c.ID, Title: "  ", WorldDate: "  312 AS "})
	if err != nil {
		t.Fatalf("SaveWorldNote blank: %v", err)
	}
	if blank.Title != defaultWorldTitle || blank.Kind != defaultWorldKind || blank.WorldDate != "312 AS" {
		t.Fatalf("blank page not normalised: %+v", blank)
	}

	if _, err := s.SaveWorldNote(WorldNoteInput{CampaignID: c.ID, Kind: "planet", Title: "Nope"}); err == nil {
		t.Fatal("expected an unknown kind to be rejected")
	}
}

func TestWorldNoteUpdateKeepsParent(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	timeline := newWorldPage(t, s, c.ID, nil, "timeline", "The Second Age")
	event := newWorldPage(t, s, c.ID, &timeline, "event", "The Burning")

	updated, err := s.SaveWorldNote(WorldNoteInput{
		ID:         &event.ID,
		CampaignID: c.ID,
		Kind:       "event",
		Title:      "The Burning of Varn",
		WorldDate:  "312 AS",
		Body:       "The bridge goes first.\n",
	})
	if err != nil {
		t.Fatalf("SaveWorldNote update: %v", err)
	}
	if updated.ParentID == nil || *updated.ParentID != timeline.ID {
		t.Fatalf("update moved the page: %+v", updated)
	}
	if updated.Body != "The bridge goes first.\n" || updated.WorldDate != "312 AS" {
		t.Fatalf("update did not stick: %+v", updated)
	}
}

func TestWorldNoteMoveRejectsCycles(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	region := newWorldPage(t, s, c.ID, nil, "region", "The Drowned South")
	city := newWorldPage(t, s, c.ID, &region, "city", "Ashfall")
	place := newWorldPage(t, s, c.ID, &city, "place", "The Ferryman's Rest")

	if _, err := s.MoveWorldNote(region.ID, &region.ID); err == nil {
		t.Fatal("expected a page moved into itself to be rejected")
	}
	if _, err := s.MoveWorldNote(region.ID, &place.ID); err == nil {
		t.Fatal("expected a page moved into its own grandchild to be rejected")
	}

	moved, err := s.MoveWorldNote(place.ID, nil)
	if err != nil {
		t.Fatalf("MoveWorldNote to top level: %v", err)
	}
	if moved.ParentID != nil {
		t.Fatalf("page still has a parent: %+v", moved)
	}
	if got := worldTitles(t, s, c.ID, nil); len(got) != 2 || got[1] != "The Ferryman's Rest" {
		t.Fatalf("moved page should land last at the top level, got %v", got)
	}

	back, err := s.MoveWorldNote(region.ID, &place.ID)
	if err != nil {
		t.Fatalf("MoveWorldNote under a former descendant: %v", err)
	}
	if back.ParentID == nil || *back.ParentID != place.ID {
		t.Fatalf("region not under the place: %+v", back)
	}
}

func TestWorldNoteParentMustShareCampaign(t *testing.T) {
	s := newTestService(t)
	a := newTestCampaign(t, s, "Age of Umbra")
	b := newTestCampaign(t, s, "The Long Winter")

	region := newWorldPage(t, s, a.ID, nil, "region", "The Drowned South")
	other := newWorldPage(t, s, b.ID, nil, "region", "The Frost March")

	if _, err := s.SaveWorldNote(WorldNoteInput{CampaignID: b.ID, ParentID: &region.ID, Title: "Stray"}); err == nil {
		t.Fatal("expected a parent from another campaign to be rejected on create")
	}
	if _, err := s.MoveWorldNote(other.ID, &region.ID); err == nil {
		t.Fatal("expected a parent from another campaign to be rejected on move")
	}
}

func TestWorldNoteShiftReorders(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	timeline := newWorldPage(t, s, c.ID, nil, "timeline", "The Second Age")
	first := newWorldPage(t, s, c.ID, &timeline, "event", "Founding")
	newWorldPage(t, s, c.ID, &timeline, "event", "Schism")
	third := newWorldPage(t, s, c.ID, &timeline, "event", "Burning")
	newWorldPage(t, s, c.ID, nil, "region", "Elsewhere")

	steps := []struct {
		id    int64
		delta int
		want  string
	}{
		{third.ID, -1, "Founding,Burning,Schism"},
		{first.ID, -5, "Founding,Burning,Schism"},
		{first.ID, 2, "Burning,Schism,Founding"},
		{first.ID, 1, "Burning,Schism,Founding"},
	}
	for _, step := range steps {
		if err := s.ShiftWorldNote(step.id, step.delta); err != nil {
			t.Fatalf("ShiftWorldNote(%d, %d): %v", step.id, step.delta, err)
		}
		if got := strings.Join(worldTitles(t, s, c.ID, &timeline.ID), ","); got != step.want {
			t.Fatalf("after shifting %d by %d: got %s, want %s", step.id, step.delta, got, step.want)
		}
	}
	if got := strings.Join(worldTitles(t, s, c.ID, nil), ","); got != "The Second Age,Elsewhere" {
		t.Fatalf("shifting children disturbed the top level: %s", got)
	}
}

func TestWorldNoteDeleteTakesSubtreeAndSearchRows(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	region := newWorldPage(t, s, c.ID, nil, "region", "The Umbermarch")
	city := newWorldPage(t, s, c.ID, &region, "city", "Ashfall")
	newWorldPage(t, s, c.ID, &city, "place", "Mirewatch Tavern")

	hits, err := s.Search("mirewatch", c.ID, 0)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 || hits[0].Entity != entityWorld || hits[0].CampaignID != c.ID {
		t.Fatalf("expected one world hit, got %+v", hits)
	}

	other := newTestCampaign(t, s, "The Long Winter")
	if hits, err := s.Search("mirewatch", other.ID, 0); err != nil || len(hits) != 0 {
		t.Fatalf("world page leaked into another campaign's search: %+v, %v", hits, err)
	}

	if err := s.DeleteWorldNote(city.ID); err != nil {
		t.Fatalf("DeleteWorldNote: %v", err)
	}
	if got := worldTitles(t, s, c.ID, &city.ID); len(got) != 0 {
		t.Fatalf("children survived their parent: %v", got)
	}
	notes, err := s.ListWorldNotes(c.ID)
	if err != nil || len(notes) != 1 {
		t.Fatalf("expected only the region left, got %+v, %v", notes, err)
	}
	if hits, err := s.Search("mirewatch", 0, 0); err != nil || len(hits) != 0 {
		t.Fatalf("a cascaded page left a search row: %+v, %v", hits, err)
	}

	if err := s.DeleteCampaign(c.ID); err != nil {
		t.Fatalf("DeleteCampaign: %v", err)
	}
	if hits, err := s.Search("umbermarch", 0, 0); err != nil || len(hits) != 0 {
		t.Fatalf("a deleted campaign left world search rows: %+v, %v", hits, err)
	}
}

func TestWorldNotesLibraryRoundTrip(t *testing.T) {
	s := newTestService(t)
	c := newTestCampaign(t, s, "Age of Umbra")

	region := newWorldPage(t, s, c.ID, nil, "region", "The Drowned South")
	city := newWorldPage(t, s, c.ID, &region, "city", "Ashfall")
	newWorldPage(t, s, c.ID, &city, "place", "Mirewatch Tavern")
	newWorldPage(t, s, c.ID, &region, "city", "Varn")
	timeline := newWorldPage(t, s, c.ID, nil, "timeline", "The Second Age")
	if _, err := s.SaveWorldNote(WorldNoteInput{CampaignID: c.ID, ParentID: &timeline.ID, Kind: "event", Title: "The Burning", WorldDate: "312 AS"}); err != nil {
		t.Fatalf("SaveWorldNote: %v", err)
	}

	raw, err := s.ExportLibraryJSON()
	if err != nil {
		t.Fatalf("ExportLibraryJSON: %v", err)
	}
	raw = strings.Replace(raw, `"kind": "timeline"`, `"kind": "almanac"`, 1)

	fresh := newTestService(t)
	report, err := fresh.ImportLibraryJSON(raw)
	if err != nil {
		t.Fatalf("ImportLibraryJSON: %v", err)
	}
	if report.WorldNotes != 6 || len(report.Skipped) != 0 {
		t.Fatalf("unexpected report %+v", report)
	}

	campaigns, err := fresh.ListCampaigns()
	if err != nil || len(campaigns) != 1 {
		t.Fatalf("ListCampaigns = %+v, %v", campaigns, err)
	}
	notes, err := fresh.ListWorldNotes(campaigns[0].ID)
	if err != nil {
		t.Fatalf("ListWorldNotes: %v", err)
	}
	byTitle := map[string]WorldNote{}
	for _, n := range notes {
		byTitle[n.Title] = n
	}
	parentOf := func(title string) string {
		n := byTitle[title]
		if n.ParentID == nil {
			return ""
		}
		for _, p := range notes {
			if p.ID == *n.ParentID {
				return p.Title
			}
		}
		return "?"
	}

	wants := map[string]string{
		"The Drowned South": "",
		"Ashfall":           "The Drowned South",
		"Mirewatch Tavern":  "Ashfall",
		"Varn":              "The Drowned South",
		"The Second Age":    "",
		"The Burning":       "The Second Age",
	}
	for title, parent := range wants {
		if got := parentOf(title); got != parent {
			t.Fatalf("%q imported under %q, want %q", title, got, parent)
		}
	}
	southID := byTitle["The Drowned South"].ID
	if got := strings.Join(worldTitles(t, fresh, campaigns[0].ID, &southID), ","); got != "Ashfall,Varn" {
		t.Fatalf("sibling order lost on import: %s", got)
	}
	if byTitle["The Second Age"].Kind != defaultWorldKind {
		t.Fatalf("an unknown kind should import as %q, got %q", defaultWorldKind, byTitle["The Second Age"].Kind)
	}
	if byTitle["The Burning"].WorldDate != "312 AS" {
		t.Fatalf("world date lost on import: %+v", byTitle["The Burning"])
	}
}
