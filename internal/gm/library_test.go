package gm

import (
	"testing"
)

type seededLibrary struct {
	adversarySlug string
	encounterID   int64
	campaignID    int64
	sessionID     int64
	noteID        int64
	regionID      int64
	countdownID   int64
}

func seedLibrary(t *testing.T, s *Service) seededLibrary {
	t.Helper()

	party, err := s.CreateParty(PartyInput{Name: "The Table", Size: 4, Tier: "1"})
	if err != nil {
		t.Fatalf("CreateParty: %v", err)
	}
	card, err := s.CreateCustomAdversary(sampleAdversary())
	if err != nil {
		t.Fatalf("CreateCustomAdversary: %v", err)
	}
	enc, err := s.SaveEncounter(EncounterInput{
		Name:              "Scrapyard Ambush",
		PartyID:           &party.ID,
		CustomAdversaries: []Pick{{Slug: card.Slug, Count: 2}},
	})
	if err != nil {
		t.Fatalf("SaveEncounter: %v", err)
	}

	campaign, err := s.SaveCampaign(CampaignInput{Name: "Ashes", Description: "A war in the west."})
	if err != nil {
		t.Fatalf("SaveCampaign: %v", err)
	}
	if _, err := s.SetCampaignFear(campaign.ID, 3); err != nil {
		t.Fatalf("SetCampaignFear: %v", err)
	}
	if _, err := s.SaveMasterNote(campaign.ID, "Day 1: the bridge burns."); err != nil {
		t.Fatalf("SaveMasterNote: %v", err)
	}
	session, err := s.SaveSession(SessionInput{CampaignID: campaign.ID, Title: "Arrival", Recap: "They arrive."})
	if err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
	if _, err := s.LinkEncounter(session.ID, enc.ID); err != nil {
		t.Fatalf("LinkEncounter: %v", err)
	}
	note, err := s.SaveNote(NoteInput{CampaignID: campaign.ID, Kind: "npc", Title: "Sabine", Body: "Runs the docks."})
	if err != nil {
		t.Fatalf("SaveNote: %v", err)
	}
	region, err := s.SaveWorldNote(WorldNoteInput{CampaignID: campaign.ID, Kind: "region", Title: "Westlands", Body: "Dry."})
	if err != nil {
		t.Fatalf("SaveWorldNote (region): %v", err)
	}
	if _, err := s.SaveWorldNote(WorldNoteInput{CampaignID: campaign.ID, ParentID: &region.ID, Kind: "city", Title: "Port Vell"}); err != nil {
		t.Fatalf("SaveWorldNote (city): %v", err)
	}
	clock, err := s.SaveCountdown(CountdownInput{CampaignID: &campaign.ID, Name: "The siege", Max: 4})
	if err != nil {
		t.Fatalf("SaveCountdown (campaign): %v", err)
	}
	if _, err := s.SaveCountdown(CountdownInput{Name: "Collapse", Max: 3}); err != nil {
		t.Fatalf("SaveCountdown (loose): %v", err)
	}

	return seededLibrary{
		adversarySlug: card.Slug,
		encounterID:   enc.ID,
		campaignID:    campaign.ID,
		sessionID:     session.ID,
		noteID:        note.ID,
		regionID:      region.ID,
		countdownID:   clock.ID,
	}
}

func exportLibrary(t *testing.T, s *Service) string {
	t.Helper()
	raw, err := s.ExportLibraryJSON()
	if err != nil {
		t.Fatalf("ExportLibraryJSON: %v", err)
	}
	return raw
}

func importLibrary(t *testing.T, s *Service, raw string) ImportReport {
	t.Helper()
	report, err := s.ImportLibraryJSON(raw)
	if err != nil {
		t.Fatalf("ImportLibraryJSON: %v", err)
	}
	if len(report.Skipped) != 0 || len(report.Renamed) != 0 {
		t.Fatalf("import skipped %v, renamed %v", report.Skipped, report.Renamed)
	}
	return report
}

func onlyCampaign(t *testing.T, s *Service) Campaign {
	t.Helper()
	campaigns, err := s.ListCampaigns()
	if err != nil {
		t.Fatalf("ListCampaigns: %v", err)
	}
	if len(campaigns) != 1 {
		t.Fatalf("want exactly one campaign, got %d: %+v", len(campaigns), campaigns)
	}
	return campaigns[0]
}

func TestLibraryReimportAddsNothing(t *testing.T) {
	s := newTestService(t)
	seedLibrary(t, s)
	raw := exportLibrary(t, s)

	fresh := newTestService(t)
	first := importLibrary(t, fresh, raw)
	wantFirst := ImportCounts{
		Parties: 1, CustomAdversaries: 1, Encounters: 1, Campaigns: 1,
		Sessions: 1, Notes: 1, WorldNotes: 2, Countdowns: 2,
	}
	if first.ImportCounts != wantFirst || first.Updated != (ImportCounts{}) {
		t.Fatalf("first import: added %+v updated %+v", first.ImportCounts, first.Updated)
	}

	second := importLibrary(t, fresh, raw)
	if second.ImportCounts != (ImportCounts{}) || second.Updated != (ImportCounts{}) {
		t.Fatalf("re-importing the same file changed something: added %+v updated %+v", second.ImportCounts, second.Updated)
	}

	campaign := onlyCampaign(t, fresh)
	parties, _ := fresh.ListParties()
	adversaries, _ := fresh.ListCustomAdversaries(Filter{})
	encounters, _ := fresh.ListEncounters()
	sessions, _ := fresh.ListSessions(campaign.ID)
	notes, _ := fresh.ListNotes(campaign.ID)
	world, _ := fresh.ListWorldNotes(campaign.ID)
	clocks, _ := fresh.ListCountdownsForCampaign(campaign.ID)
	loose, _ := fresh.ListUnassignedCountdowns()
	got := []int{len(parties), len(adversaries), len(encounters), len(sessions), len(notes), len(world), len(clocks), len(loose)}
	want := []int{1, 1, 1, 1, 1, 2, 1, 1}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("record counts after two imports = %v, want %v", got, want)
		}
	}

	view, err := fresh.GetSession(sessions[0].ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if len(view.Encounters) != 1 {
		t.Fatalf("session links duplicated or lost: %+v", view.Encounters)
	}
}

func TestLibraryImportOverwritesMatchesAndKeepsTheRest(t *testing.T) {
	s := newTestService(t)
	seed := seedLibrary(t, s)

	fresh := newTestService(t)
	importLibrary(t, fresh, exportLibrary(t, s))
	freshCampaign := onlyCampaign(t, fresh)
	if _, err := fresh.SaveNote(NoteInput{CampaignID: freshCampaign.ID, Kind: "lore", Title: "Only here", Body: "Written on the laptop."}); err != nil {
		t.Fatalf("SaveNote (fresh only): %v", err)
	}

	card, err := s.GetCustomAdversary(seed.adversarySlug)
	if err != nil {
		t.Fatalf("GetCustomAdversary: %v", err)
	}
	card.Name, card.Difficulty = "Rust Wraith Elder", "16"
	if _, err := s.UpdateCustomAdversary(card); err != nil {
		t.Fatalf("UpdateCustomAdversary: %v", err)
	}
	if _, err := s.SaveEncounter(EncounterInput{
		ID:                &seed.encounterID,
		Name:              "Scrapyard Ambush",
		CustomAdversaries: []Pick{{Slug: seed.adversarySlug, Count: 3}},
	}); err != nil {
		t.Fatalf("SaveEncounter: %v", err)
	}
	if _, err := s.SaveCampaign(CampaignInput{ID: &seed.campaignID, Name: "Ashes", Description: "The war reaches the coast."}); err != nil {
		t.Fatalf("SaveCampaign: %v", err)
	}
	if _, err := s.SetCampaignFear(seed.campaignID, 5); err != nil {
		t.Fatalf("SetCampaignFear: %v", err)
	}
	if _, err := s.SaveSession(SessionInput{ID: &seed.sessionID, CampaignID: seed.campaignID, Number: 1, Title: "Arrival", Recap: "They arrive, and the bridge burns."}); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
	if _, err := s.SaveNote(NoteInput{ID: &seed.noteID, CampaignID: seed.campaignID, Kind: "npc", Title: "Sabine", Body: "Runs the docks. Owes the party."}); err != nil {
		t.Fatalf("SaveNote: %v", err)
	}
	if _, err := s.SaveNote(NoteInput{CampaignID: seed.campaignID, Kind: "npc", Title: "Marius", Body: "The warlord."}); err != nil {
		t.Fatalf("SaveNote (new): %v", err)
	}
	if _, err := s.SaveWorldNote(WorldNoteInput{ID: &seed.regionID, CampaignID: seed.campaignID, Kind: "region", Title: "Westlands", Body: "Dry, and burning."}); err != nil {
		t.Fatalf("SaveWorldNote: %v", err)
	}
	if _, err := s.AdvanceCountdown(seed.countdownID, 2); err != nil {
		t.Fatalf("AdvanceCountdown: %v", err)
	}

	report := importLibrary(t, fresh, exportLibrary(t, s))
	wantAdded := ImportCounts{Notes: 1}
	wantUpdated := ImportCounts{
		CustomAdversaries: 1, Encounters: 1, Campaigns: 1,
		Sessions: 1, Notes: 1, WorldNotes: 1, Countdowns: 1,
	}
	if report.ImportCounts != wantAdded || report.Updated != wantUpdated {
		t.Fatalf("added %+v (want %+v), updated %+v (want %+v)", report.ImportCounts, wantAdded, report.Updated, wantUpdated)
	}

	adversaries, err := fresh.ListCustomAdversaries(Filter{})
	if err != nil {
		t.Fatalf("ListCustomAdversaries: %v", err)
	}
	if len(adversaries) != 1 || adversaries[0].Name != "Rust Wraith Elder" || adversaries[0].Difficulty != "16" {
		t.Fatalf("renamed card should overwrite its match by slug, got %+v", adversaries)
	}

	encounters, err := fresh.ListEncounters()
	if err != nil || len(encounters) != 1 {
		t.Fatalf("ListEncounters = %+v, %v", encounters, err)
	}
	enc, err := fresh.GetEncounter(encounters[0].ID)
	if err != nil {
		t.Fatalf("GetEncounter: %v", err)
	}
	if len(enc.Adversaries) != 1 || enc.Adversaries[0].Count != 3 || enc.Adversaries[0].Unresolved {
		t.Fatalf("encounter picks not overwritten: %+v", enc.Adversaries)
	}

	campaign := onlyCampaign(t, fresh)
	if campaign.Description != "The war reaches the coast." || campaign.CurrentFear != 5 {
		t.Fatalf("campaign not overwritten: %+v", campaign)
	}

	sessions, err := fresh.ListSessions(campaign.ID)
	if err != nil || len(sessions) != 1 || sessions[0].Recap != "They arrive, and the bridge burns." {
		t.Fatalf("session not overwritten: %+v, %v", sessions, err)
	}

	notes, err := fresh.ListNotes(campaign.ID)
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	bodies := map[string]string{}
	for _, n := range notes {
		bodies[n.Title] = n.Body
	}
	if len(notes) != 3 || bodies["Sabine"] != "Runs the docks. Owes the party." || bodies["Marius"] != "The warlord." || bodies["Only here"] == "" {
		t.Fatalf("notes after merge = %+v", notes)
	}

	world, err := fresh.ListWorldNotes(campaign.ID)
	if err != nil {
		t.Fatalf("ListWorldNotes: %v", err)
	}
	if len(world) != 2 {
		t.Fatalf("world pages duplicated: %+v", world)
	}
	for _, page := range world {
		if page.Title == "Westlands" && page.Body != "Dry, and burning." {
			t.Fatalf("world page not overwritten: %+v", page)
		}
	}

	clocks, err := fresh.ListCountdownsForCampaign(campaign.ID)
	if err != nil || len(clocks) != 1 || clocks[0].Value != 2 {
		t.Fatalf("countdown not overwritten: %+v, %v", clocks, err)
	}
}

func TestLibraryImportMergesIntoSameNamedLocalRecords(t *testing.T) {
	s := newTestService(t)
	seedLibrary(t, s)
	raw := exportLibrary(t, s)

	fresh := newTestService(t)
	local := sampleAdversary()
	local.Difficulty = "10"
	if _, err := fresh.CreateCustomAdversary(local); err != nil {
		t.Fatalf("CreateCustomAdversary: %v", err)
	}
	if _, err := fresh.SaveCampaign(CampaignInput{Name: "  ashes "}); err != nil {
		t.Fatalf("SaveCampaign: %v", err)
	}

	report := importLibrary(t, fresh, raw)
	if report.CustomAdversaries != 0 || report.Updated.CustomAdversaries != 1 {
		t.Fatalf("same-named adversary should be overwritten, added %d updated %d", report.CustomAdversaries, report.Updated.CustomAdversaries)
	}
	if report.Campaigns != 0 || report.Updated.Campaigns != 1 {
		t.Fatalf("same-named campaign should be overwritten, added %d updated %d", report.Campaigns, report.Updated.Campaigns)
	}

	adversaries, err := fresh.ListCustomAdversaries(Filter{})
	if err != nil || len(adversaries) != 1 || adversaries[0].Difficulty != "14" {
		t.Fatalf("adversaries after merge = %+v, %v", adversaries, err)
	}
	campaign := onlyCampaign(t, fresh)
	if campaign.Name != "Ashes" || campaign.Description != "A war in the west." {
		t.Fatalf("campaign after merge = %+v", campaign)
	}
}
