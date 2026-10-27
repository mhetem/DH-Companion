package player

import "testing"

func TestNewSRDContentLoads(t *testing.T) {
	s := newTestService(t)

	for _, slug := range []string{"assassin", "brawler", "warlock", "witch"} {
		class, ok := s.catalog.Class(slug)
		if !ok || len(class.Subclasses) != 2 || class.HopeFeature.Title == "" {
			t.Errorf("%s should load with its hope feature and two subclasses: %+v", slug, class.Name)
		}
	}
	dread := 0
	for _, card := range s.catalog.ListDomainCards() {
		if card.Domain == "Dread" {
			dread++
		}
	}
	if dread != 21 {
		t.Errorf("the Dread domain has %d cards, want 21", dread)
	}
	if !contains(s.DomainList(), "Dread") {
		t.Error("Dread should be one of the domains")
	}
	if got := len(s.ListTransformations()); got != 6 {
		t.Errorf("%d transformations, want 6", got)
	}
	if got := len(s.ListAncestries()); got != 25 {
		t.Errorf("%d ancestries, want 25", got)
	}
	if got := len(s.ListCommunities()); got != 15 {
		t.Errorf("%d communities, want 15", got)
	}
}

func TestTransformationOnTheSheet(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 1, 2)

	if _, err := s.SetTransformation(c.ID, "lich"); err == nil {
		t.Fatal("an unknown transformation should be refused")
	}
	updated, err := s.SetTransformation(c.ID, "werewolf")
	if err != nil {
		t.Fatalf("SetTransformation: %v", err)
	}
	if updated.TransformationSlug != "werewolf" || updated.TransformationName != "Werewolf" {
		t.Fatalf("unexpected character %+v", updated)
	}

	sheet, err := s.GetSheet(c.ID)
	if err != nil {
		t.Fatalf("GetSheet: %v", err)
	}
	if sheet.Transformation == nil || len(sheet.Transformation.Features) != 2 || len(sheet.Transformation.Questions) != 6 {
		t.Fatalf("the sheet should carry the werewolf card: %+v", sheet.Transformation)
	}

	if _, err := s.SetTransformation(c.ID, ""); err != nil {
		t.Fatalf("clearing the transformation: %v", err)
	}
	sheet, _ = s.GetSheet(c.ID)
	if sheet.Transformation != nil || sheet.TransformationName != "" {
		t.Fatalf("the transformation should be gone: %+v", sheet.Transformation)
	}
}

func TestMartialArtistGetsTheStanceSheet(t *testing.T) {
	s := newTestService(t)

	artist := buildCharacter(t, s, "brawler", "martialArtist", "clank", 0, 1, 2)
	sheet, err := s.GetSheet(artist.ID)
	if err != nil {
		t.Fatalf("GetSheet: %v", err)
	}
	if sheet.Stances == nil || len(sheet.Stances.Stances) != 16 || len(sheet.Stances.Rules) != 4 {
		t.Fatalf("a Martial Artist should carry the stance sheet: %+v", sheet.Stances)
	}

	juggernaut := buildCharacter(t, s, "brawler", "juggernaut", "clank", 0, 1, 2)
	if sheet, _ := s.GetSheet(juggernaut.ID); sheet.Stances != nil {
		t.Fatal("a Juggernaut has no stances")
	}
}
