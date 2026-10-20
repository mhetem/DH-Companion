package gm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

const (
	LibraryFormat = 1
	libraryApp    = "hilt"
)

type Library struct {
	Format             int                 `json:"format"`
	App                string              `json:"app"`
	ExportedAt         string              `json:"exportedAt"`
	Parties            []LibraryParty      `json:"parties"`
	CustomAdversaries  []cards.Adversary   `json:"customAdversaries"`
	CustomEnvironments []cards.Environment `json:"customEnvironments"`
	// The equipment and loot homebrew arrived after format 1 and are additive:
	// an older export simply has none, and an older build ignores them, so the
	// format number doesn't move.
	CustomWeapons     []cards.Weapon     `json:"customWeapons"`
	CustomArmor       []cards.Armor      `json:"customArmor"`
	CustomItems       []cards.Loot       `json:"customItems"`
	CustomConsumables []cards.Loot       `json:"customConsumables"`
	Encounters        []LibraryEncounter `json:"encounters"`
	Campaigns         []LibraryCampaign  `json:"campaigns"`
	Countdowns        []LibraryCountdown `json:"countdowns"`
}

type LibraryParty struct {
	Name string `json:"name"`
	Size int    `json:"size"`
	Tier string `json:"tier"`
}

type LibraryEncounter struct {
	Name              string  `json:"name"`
	PartyName         *string `json:"partyName"`
	EnvironmentSlug   *string `json:"environmentSlug"`
	Adversaries       []Pick  `json:"adversaries"`
	CustomAdversaries []Pick  `json:"customAdversaries"`
}

type LibraryCampaign struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CurrentFear int                `json:"currentFear"`
	MasterNote  string             `json:"masterNote"`
	Sessions    []LibrarySession   `json:"sessions"`
	Notes       []LibraryNote      `json:"notes"`
	World       []LibraryWorldNote `json:"world"`
	Countdowns  []LibraryCountdown `json:"countdowns"`
}

type LibrarySession struct {
	Number     int      `json:"number"`
	Title      string   `json:"title"`
	Date       string   `json:"date"`
	Recap      string   `json:"recap"`
	Encounters []string `json:"encounters"`
}

type LibraryNote struct {
	Kind  string `json:"kind"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type LibraryWorldNote struct {
	Kind      string             `json:"kind"`
	Title     string             `json:"title"`
	WorldDate string             `json:"worldDate"`
	Body      string             `json:"body"`
	Children  []LibraryWorldNote `json:"children"`
}

type LibraryCountdown struct {
	Name  string          `json:"name"`
	Value int             `json:"value"`
	Max   int             `json:"max"`
	Kind  string          `json:"kind"`
	Beats []CountdownBeat `json:"beats"`
}

type ImportCounts struct {
	Parties            int `json:"parties"`
	CustomAdversaries  int `json:"customAdversaries"`
	CustomEnvironments int `json:"customEnvironments"`
	CustomWeapons      int `json:"customWeapons"`
	CustomArmor        int `json:"customArmor"`
	CustomItems        int `json:"customItems"`
	CustomConsumables  int `json:"customConsumables"`
	Encounters         int `json:"encounters"`
	Campaigns          int `json:"campaigns"`
	Sessions           int `json:"sessions"`
	Notes              int `json:"notes"`
	WorldNotes         int `json:"worldNotes"`
	Countdowns         int `json:"countdowns"`
}

type ImportReport struct {
	ImportCounts
	Updated ImportCounts `json:"updated"`
	Renamed []string     `json:"renamed"`
	Skipped []string     `json:"skipped"`
}

func (s *Service) buildLibrary() (Library, error) {
	lib := Library{
		Format:             LibraryFormat,
		App:                libraryApp,
		ExportedAt:         time.Now().UTC().Format(time.RFC3339),
		Parties:            []LibraryParty{},
		CustomAdversaries:  []cards.Adversary{},
		CustomEnvironments: []cards.Environment{},
		CustomWeapons:      []cards.Weapon{},
		CustomArmor:        []cards.Armor{},
		CustomItems:        []cards.Loot{},
		CustomConsumables:  []cards.Loot{},
		Encounters:         []LibraryEncounter{},
		Campaigns:          []LibraryCampaign{},
		Countdowns:         []LibraryCountdown{},
	}

	parties, err := s.ListParties()
	if err != nil {
		return Library{}, err
	}
	partyNames := make(map[int64]string, len(parties))
	for _, p := range parties {
		partyNames[p.ID] = p.Name
		lib.Parties = append(lib.Parties, LibraryParty{Name: p.Name, Size: p.Size, Tier: p.Tier})
	}

	adversaries, err := s.ListCustomAdversaries(Filter{})
	if err != nil {
		return Library{}, err
	}
	lib.CustomAdversaries = append(lib.CustomAdversaries, adversaries...)

	environments, err := s.ListCustomEnvironments(Filter{})
	if err != nil {
		return Library{}, err
	}
	lib.CustomEnvironments = append(lib.CustomEnvironments, environments...)

	weapons, err := s.ListCustomWeapons(EquipmentFilter{})
	if err != nil {
		return Library{}, err
	}
	lib.CustomWeapons = append(lib.CustomWeapons, weapons...)

	armor, err := s.ListCustomArmor(EquipmentFilter{})
	if err != nil {
		return Library{}, err
	}
	lib.CustomArmor = append(lib.CustomArmor, armor...)

	items, err := s.ListCustomItems(LootFilter{})
	if err != nil {
		return Library{}, err
	}
	lib.CustomItems = append(lib.CustomItems, items...)

	consumables, err := s.ListCustomConsumables(LootFilter{})
	if err != nil {
		return Library{}, err
	}
	lib.CustomConsumables = append(lib.CustomConsumables, consumables...)

	rows, err := s.q.ShowAllEncounters(s.ctx)
	if err != nil {
		return Library{}, fmt.Errorf("listing encounters: %w", err)
	}
	for _, r := range rows {
		srdPicks, err := decodePicks(r.Adversaries)
		if err != nil {
			return Library{}, err
		}
		customPicks, err := decodePicks(r.CustomAdversaries)
		if err != nil {
			return Library{}, err
		}
		enc := LibraryEncounter{
			Name:              r.EncounterName,
			EnvironmentSlug:   stringPtr(r.EnvironmentSlug),
			Adversaries:       srdPicks,
			CustomAdversaries: customPicks,
		}
		if id := int64Ptr(r.PartyID); id != nil {
			if name, ok := partyNames[*id]; ok {
				enc.PartyName = &name
			}
		}
		lib.Encounters = append(lib.Encounters, enc)
	}

	campaigns, err := s.ListCampaigns()
	if err != nil {
		return Library{}, err
	}
	for _, c := range campaigns {
		master, err := s.GetMasterNote(c.ID)
		if err != nil {
			return Library{}, err
		}
		entry := LibraryCampaign{
			Name:        c.Name,
			Description: c.Description,
			CurrentFear: c.CurrentFear,
			MasterNote:  master.Body,
			Sessions:    []LibrarySession{},
			Notes:       []LibraryNote{},
			World:       []LibraryWorldNote{},
			Countdowns:  []LibraryCountdown{},
		}

		sessions, err := s.ListSessions(c.ID)
		if err != nil {
			return Library{}, err
		}
		for _, sum := range sessions {
			view, err := s.GetSession(sum.ID)
			if err != nil {
				return Library{}, err
			}
			names := make([]string, 0, len(view.Encounters))
			for _, e := range view.Encounters {
				names = append(names, e.Name)
			}
			entry.Sessions = append(entry.Sessions, LibrarySession{
				Number:     sum.Number,
				Title:      sum.Title,
				Date:       sum.Date,
				Recap:      sum.Recap,
				Encounters: names,
			})
		}

		notes, err := s.ListNotes(c.ID)
		if err != nil {
			return Library{}, err
		}
		for _, n := range notes {
			entry.Notes = append(entry.Notes, LibraryNote{Kind: n.Kind, Title: n.Title, Body: n.Body})
		}

		world, err := s.ListWorldNotes(c.ID)
		if err != nil {
			return Library{}, err
		}
		entry.World = libraryWorld(world)

		clocks, err := s.ListCountdownsForCampaign(c.ID)
		if err != nil {
			return Library{}, err
		}
		for _, cd := range clocks {
			entry.Countdowns = append(entry.Countdowns, LibraryCountdown{Name: cd.Name, Value: cd.Value, Max: cd.Max, Kind: cd.Kind, Beats: cd.Beats})
		}

		lib.Campaigns = append(lib.Campaigns, entry)
	}

	loose, err := s.ListUnassignedCountdowns()
	if err != nil {
		return Library{}, err
	}
	for _, cd := range loose {
		lib.Countdowns = append(lib.Countdowns, LibraryCountdown{Name: cd.Name, Value: cd.Value, Max: cd.Max, Kind: cd.Kind, Beats: cd.Beats})
	}

	return lib, nil
}

func (s *Service) ExportLibraryJSON() (string, error) {
	lib, err := s.buildLibrary()
	if err != nil {
		return "", err
	}
	out, err := json.MarshalIndent(lib, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encoding library: %w", err)
	}
	return string(out), nil
}

func (s *Service) ImportLibraryJSON(raw string) (ImportReport, error) {
	var lib Library
	if err := json.Unmarshal([]byte(raw), &lib); err != nil {
		return ImportReport{}, fmt.Errorf("this file isn't a Hilt library export: %w", err)
	}
	if lib.App != "" && lib.App != libraryApp {
		return ImportReport{}, fmt.Errorf("this export came from %q, not Hilt", lib.App)
	}
	if lib.Format > LibraryFormat {
		return ImportReport{}, fmt.Errorf("this export is format %d and this build reads %d", lib.Format, LibraryFormat)
	}
	return s.importLibrary(lib)
}

func (s *Service) importLibrary(lib Library) (ImportReport, error) {
	report := ImportReport{Renamed: []string{}, Skipped: []string{}}

	localAdversaries, err := s.ListCustomAdversaries(Filter{})
	if err != nil {
		return report, err
	}
	adversarySlugs := mergeCards(s, &report, "adversary", lib.CustomAdversaries, localAdversaries,
		func(c cards.Adversary, name, slug string) cards.Adversary { c.Name, c.Slug = name, slug; return c },
		s.adversarySlugTaken, s.CreateCustomAdversary, s.UpdateCustomAdversary,
		&report.CustomAdversaries, &report.Updated.CustomAdversaries)

	localEnvironments, err := s.ListCustomEnvironments(Filter{})
	if err != nil {
		return report, err
	}
	environmentSlugs := mergeCards(s, &report, "environment", lib.CustomEnvironments, localEnvironments,
		func(c cards.Environment, name, slug string) cards.Environment { c.Name, c.Slug = name, slug; return c },
		s.environmentSlugTaken, s.CreateCustomEnvironment, s.UpdateCustomEnvironment,
		&report.CustomEnvironments, &report.Updated.CustomEnvironments)

	localWeapons, err := s.ListCustomWeapons(EquipmentFilter{})
	if err != nil {
		return report, err
	}
	mergeCards(s, &report, "weapon", lib.CustomWeapons, localWeapons,
		func(c cards.Weapon, name, slug string) cards.Weapon { c.Name, c.Slug = name, slug; return c },
		s.slugTaken(func(slug string) error { _, err := s.GetCustomWeapon(slug); return err }),
		s.CreateCustomWeapon, s.UpdateCustomWeapon,
		&report.CustomWeapons, &report.Updated.CustomWeapons)

	localArmor, err := s.ListCustomArmor(EquipmentFilter{})
	if err != nil {
		return report, err
	}
	mergeCards(s, &report, "armor", lib.CustomArmor, localArmor,
		func(c cards.Armor, name, slug string) cards.Armor { c.Name, c.Slug = name, slug; return c },
		s.slugTaken(func(slug string) error { _, err := s.GetCustomArmor(slug); return err }),
		s.CreateCustomArmor, s.UpdateCustomArmor,
		&report.CustomArmor, &report.Updated.CustomArmor)

	localItems, err := s.ListCustomItems(LootFilter{})
	if err != nil {
		return report, err
	}
	mergeCards(s, &report, "item", lib.CustomItems, localItems,
		func(c cards.Loot, name, slug string) cards.Loot { c.Name, c.Slug = name, slug; return c },
		s.slugTaken(func(slug string) error { _, err := s.GetCustomItem(slug); return err }),
		s.CreateCustomItem, s.UpdateCustomItem,
		&report.CustomItems, &report.Updated.CustomItems)

	localConsumables, err := s.ListCustomConsumables(LootFilter{})
	if err != nil {
		return report, err
	}
	mergeCards(s, &report, "consumable", lib.CustomConsumables, localConsumables,
		func(c cards.Loot, name, slug string) cards.Loot { c.Name, c.Slug = name, slug; return c },
		s.slugTaken(func(slug string) error { _, err := s.GetCustomConsumable(slug); return err }),
		s.CreateCustomConsumable, s.UpdateCustomConsumable,
		&report.CustomConsumables, &report.Updated.CustomConsumables)

	partyIDs, err := s.mergeParties(&report, lib.Parties)
	if err != nil {
		return report, err
	}

	encounterIDs, err := s.mergeEncounters(&report, lib.Encounters, partyIDs, adversarySlugs, environmentSlugs)
	if err != nil {
		return report, err
	}

	campaigns, err := s.ListCampaigns()
	if err != nil {
		return report, err
	}
	campaignsByID := make(map[int64]Campaign, len(campaigns))
	for _, c := range campaigns {
		campaignsByID[c.ID] = c
	}
	campaignClaims := claimsOf(campaigns, func(c Campaign) (string, int64) { return c.Name, c.ID })
	for _, c := range lib.Campaigns {
		id, matched := campaignClaims.take(c.Name)
		campaign, ok := s.mergeCampaign(&report, c, campaignsByID[id], matched)
		if !ok {
			continue
		}
		s.mergeSessions(&report, campaign.ID, c.Sessions, encounterIDs)
		s.mergeNotes(&report, campaign.ID, c.Notes)
		s.mergeWorld(&report, campaign.ID, c.World)
		clocks, err := s.ListCountdownsForCampaign(campaign.ID)
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("countdowns of campaign %q: %v", c.Name, err))
			continue
		}
		s.mergeCountdowns(&report, &campaign.ID, c.Countdowns, clocks)
	}

	loose, err := s.ListUnassignedCountdowns()
	if err != nil {
		return report, err
	}
	s.mergeCountdowns(&report, nil, lib.Countdowns, loose)

	return report, nil
}

type claims map[string][]int64

func importKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func claimsOf[T any](items []T, key func(T) (string, int64)) claims {
	c := claims{}
	for _, item := range items {
		name, id := key(item)
		k := importKey(name)
		c[k] = append(c[k], id)
	}
	for _, ids := range c {
		slices.Sort(ids)
	}
	return c
}

func (c claims) take(name string) (int64, bool) {
	k := importKey(name)
	ids := c[k]
	if len(ids) == 0 {
		return 0, false
	}
	c[k] = ids[1:]
	return ids[0], true
}

func tally(matched bool, added, updated *int) {
	if matched {
		*updated++
		return
	}
	*added++
}

func sameJSON(a, b any) bool {
	x, err := json.Marshal(a)
	if err != nil {
		return false
	}
	y, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return bytes.Equal(x, y)
}

func mergeCards[T cards.Card](
	s *Service,
	report *ImportReport,
	label string,
	rows []T,
	local []T,
	identify func(T, string, string) T,
	taken func(string) (bool, error),
	create func(T) (T, error),
	update func(T) (T, error),
	added, updated *int,
) map[string]string {
	bySlug := make(map[string]T, len(local))
	byName := map[string][]string{}
	for _, c := range local {
		m := c.CardMeta()
		bySlug[m.Slug] = c
		byName[importKey(m.Name)] = append(byName[importKey(m.Name)], m.Slug)
	}
	claimed := map[string]bool{}
	match := func(m cards.Meta) (T, bool) {
		if c, ok := bySlug[m.Slug]; ok && m.Slug != "" && !claimed[m.Slug] {
			return c, true
		}
		for _, slug := range byName[importKey(m.Name)] {
			if !claimed[slug] {
				return bySlug[slug], true
			}
		}
		var none T
		return none, false
	}

	slugs := map[string]string{}
	for _, card := range rows {
		m := card.CardMeta()
		if existing, ok := match(m); ok {
			slug := existing.CardMeta().Slug
			claimed[slug] = true
			if m.Slug != "" {
				slugs[m.Slug] = slug
			}
			card = identify(card, m.Name, slug)
			if sameJSON(card, existing) {
				continue
			}
			if _, err := update(card); err != nil {
				report.Skipped = append(report.Skipped, fmt.Sprintf("%s %q: %v", label, m.Name, err))
				continue
			}
			*updated++
			continue
		}

		name, renamed, err := s.freeName(m.Name, taken)
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("%s %q: %v", label, m.Name, err))
			continue
		}
		saved, err := create(identify(card, name, ""))
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("%s %q: %v", label, name, err))
			continue
		}
		if m.Slug != "" {
			slugs[m.Slug] = saved.CardMeta().Slug
		}
		if renamed {
			report.Renamed = append(report.Renamed, fmt.Sprintf("%s → %s", m.Name, name))
		}
		*added++
	}
	return slugs
}

func (s *Service) mergeParties(report *ImportReport, parties []LibraryParty) (map[string]int64, error) {
	local, err := s.ListParties()
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]Party, len(local))
	for _, p := range local {
		byID[p.ID] = p
	}
	partyClaims := claimsOf(local, func(p Party) (string, int64) { return p.Name, p.ID })

	ids := map[string]int64{}
	for _, p := range parties {
		in := PartyInput{Name: p.Name, Size: p.Size, Tier: p.Tier}
		id, matched := partyClaims.take(p.Name)
		if matched {
			ids[importKey(p.Name)] = id
			old := byID[id]
			if norm, err := in.validate(); err == nil && norm.Name == old.Name && norm.Size == old.Size && norm.Tier == old.Tier {
				continue
			}
			if _, err := s.UpdateParty(id, in); err != nil {
				report.Skipped = append(report.Skipped, fmt.Sprintf("party %q: %v", p.Name, err))
				continue
			}
			report.Updated.Parties++
			continue
		}
		saved, err := s.CreateParty(in)
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("party %q: %v", p.Name, err))
			continue
		}
		ids[importKey(p.Name)] = saved.ID
		report.Parties++
	}
	return ids, nil
}

func (s *Service) mergeEncounters(
	report *ImportReport,
	encounters []LibraryEncounter,
	partyIDs map[string]int64,
	adversarySlugs, environmentSlugs map[string]string,
) (map[string]int64, error) {
	rows, err := s.q.ShowAllEncounters(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing encounters: %w", err)
	}
	byID := make(map[int64]db.Encounter, len(rows))
	for _, r := range rows {
		byID[r.ID] = r
	}
	encounterClaims := claimsOf(rows, func(r db.Encounter) (string, int64) { return r.EncounterName, r.ID })

	ids := map[string]int64{}
	for _, e := range encounters {
		in := EncounterInput{
			Name:              e.Name,
			Adversaries:       e.Adversaries,
			CustomAdversaries: remapPicks(e.CustomAdversaries, adversarySlugs),
		}
		if e.PartyName != nil {
			if id, ok := partyIDs[importKey(*e.PartyName)]; ok {
				in.PartyID = &id
			}
		}
		if e.EnvironmentSlug != nil {
			slug := *e.EnvironmentSlug
			if mapped, ok := environmentSlugs[slug]; ok {
				slug = mapped
			}
			in.EnvironmentSlug = &slug
		}

		id, matched := encounterClaims.take(e.Name)
		if matched {
			ids[e.Name] = id
			if encounterUnchanged(byID[id], in) {
				continue
			}
			in.ID = &id
		}
		saved, err := s.SaveEncounter(in)
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("encounter %q: %v", e.Name, err))
			continue
		}
		ids[e.Name] = saved.ID
		tally(matched, &report.Encounters, &report.Updated.Encounters)
	}
	return ids, nil
}

func encounterUnchanged(r db.Encounter, in EncounterInput) bool {
	norm, err := in.validate()
	if err != nil || r.EncounterName != norm.Name {
		return false
	}
	if nullInt64(norm.PartyID) != r.PartyID || nullString(norm.EnvironmentSlug) != r.EnvironmentSlug {
		return false
	}
	srd, err := encodePicks(norm.Adversaries)
	if err != nil {
		return false
	}
	custom, err := encodePicks(norm.CustomAdversaries)
	if err != nil {
		return false
	}
	return srd == r.Adversaries && custom == r.CustomAdversaries
}

func (s *Service) mergeCampaign(report *ImportReport, c LibraryCampaign, local Campaign, matched bool) (Campaign, bool) {
	in := CampaignInput{Name: c.Name, Description: c.Description}
	campaign, changed := local, false
	if !matched {
		saved, err := s.SaveCampaign(in)
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("campaign %q: %v", c.Name, err))
			return Campaign{}, false
		}
		campaign = saved
	} else if norm, err := in.validate(); err != nil || norm.Name != local.Name || norm.Description != local.Description {
		in.ID = &local.ID
		saved, err := s.SaveCampaign(in)
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("campaign %q: %v", c.Name, err))
			return Campaign{}, false
		}
		campaign, changed = saved, true
	}

	if c.CurrentFear != campaign.CurrentFear {
		if _, err := s.SetCampaignFear(campaign.ID, c.CurrentFear); err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("campaign %q fear: %v", c.Name, err))
		} else {
			changed = true
		}
	}

	if c.MasterNote != "" {
		master, err := s.GetMasterNote(campaign.ID)
		if err != nil || master.Body != c.MasterNote {
			if _, err := s.SaveMasterNote(campaign.ID, c.MasterNote); err != nil {
				report.Skipped = append(report.Skipped, fmt.Sprintf("campaign %q master note: %v", c.Name, err))
			} else {
				changed = true
			}
		}
	}

	if !matched {
		report.Campaigns++
	} else if changed {
		report.Updated.Campaigns++
	}
	return campaign, true
}

func (s *Service) mergeSessions(report *ImportReport, campaignID int64, sessions []LibrarySession, encounterIDs map[string]int64) {
	local, err := s.ListSessions(campaignID)
	if err != nil {
		report.Skipped = append(report.Skipped, fmt.Sprintf("sessions of campaign %d: %v", campaignID, err))
		return
	}
	byNumber := make(map[int]SessionSummary, len(local))
	for _, sum := range local {
		byNumber[sum.Number] = sum
	}

	for _, sess := range sessions {
		in := SessionInput{CampaignID: campaignID, Number: sess.Number, Title: sess.Title, Date: sess.Date, Recap: sess.Recap}
		existing, matched := byNumber[sess.Number]
		if sess.Number <= 0 {
			matched = false
		}
		changed := false
		linked := map[int64]bool{}

		var id int64
		if matched {
			id = existing.ID
			norm, err := in.validate()
			if err != nil || norm.Title != existing.Title || norm.Date != existing.Date || norm.Recap != existing.Recap {
				in.ID = &id
				if _, err := s.SaveSession(in); err != nil {
					report.Skipped = append(report.Skipped, fmt.Sprintf("session %q: %v", sess.Title, err))
					continue
				}
				changed = true
			}
			view, err := s.GetSession(id)
			if err != nil {
				report.Skipped = append(report.Skipped, fmt.Sprintf("session %q: %v", sess.Title, err))
				continue
			}
			for _, e := range view.Encounters {
				linked[e.ID] = true
			}
		} else {
			view, err := s.SaveSession(in)
			if err != nil {
				report.Skipped = append(report.Skipped, fmt.Sprintf("session %q: %v", sess.Title, err))
				continue
			}
			id = view.ID
		}

		for _, name := range sess.Encounters {
			encounterID, ok := encounterIDs[name]
			if !ok || linked[encounterID] {
				continue
			}
			if _, err := s.LinkEncounter(id, encounterID); err != nil {
				report.Skipped = append(report.Skipped, fmt.Sprintf("linking %q to session %d: %v", name, sess.Number, err))
				continue
			}
			linked[encounterID] = true
			changed = true
		}

		if !matched {
			report.Sessions++
		} else if changed {
			report.Updated.Sessions++
		}
	}
}

func (s *Service) mergeNotes(report *ImportReport, campaignID int64, notes []LibraryNote) {
	local, err := s.ListNotes(campaignID)
	if err != nil {
		report.Skipped = append(report.Skipped, fmt.Sprintf("notes of campaign %d: %v", campaignID, err))
		return
	}
	byID := make(map[int64]Note, len(local))
	for _, n := range local {
		byID[n.ID] = n
	}
	noteClaims := claimsOf(local, func(n Note) (string, int64) { return n.Title, n.ID })

	for _, n := range notes {
		in := NoteInput{CampaignID: campaignID, Kind: n.Kind, Title: n.Title, Body: n.Body}
		id, matched := noteClaims.take(n.Title)
		if matched {
			old := byID[id]
			if norm, err := in.validate(); err == nil && norm.Kind == old.Kind && norm.Title == old.Title && norm.Body == old.Body {
				continue
			}
			in.ID = &id
		}
		if _, err := s.SaveNote(in); err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("note %q: %v", n.Title, err))
			continue
		}
		tally(matched, &report.Notes, &report.Updated.Notes)
	}
}

func (s *Service) mergeCountdowns(report *ImportReport, campaignID *int64, clocks []LibraryCountdown, local []Countdown) {
	byID := make(map[int64]Countdown, len(local))
	for _, c := range local {
		byID[c.ID] = c
	}
	clockClaims := claimsOf(local, func(c Countdown) (string, int64) { return c.Name, c.ID })

	for _, cd := range clocks {
		in := CountdownInput{CampaignID: campaignID, Name: cd.Name, Value: cd.Value, Max: cd.Max, Kind: cd.Kind, Beats: cd.Beats}
		id, matched := clockClaims.take(cd.Name)
		if matched {
			old := byID[id]
			norm, err := in.validate()
			if err == nil && norm.Name == old.Name && norm.Value == old.Value && norm.Max == old.Max &&
				norm.Kind == old.Kind && slices.Equal(norm.Beats, old.Beats) {
				continue
			}
			in.ID = &id
		}
		if _, err := s.SaveCountdown(in); err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("countdown %q: %v", cd.Name, err))
			continue
		}
		tally(matched, &report.Countdowns, &report.Updated.Countdowns)
	}
}

func libraryWorld(notes []WorldNote) []LibraryWorldNote {
	children := map[int64][]WorldNote{}
	for _, n := range notes {
		var parent int64
		if n.ParentID != nil {
			parent = *n.ParentID
		}
		children[parent] = append(children[parent], n)
	}

	var build func(parent int64) []LibraryWorldNote
	build = func(parent int64) []LibraryWorldNote {
		out := []LibraryWorldNote{}
		for _, n := range children[parent] {
			out = append(out, LibraryWorldNote{
				Kind:      n.Kind,
				Title:     n.Title,
				WorldDate: n.WorldDate,
				Body:      n.Body,
				Children:  build(n.ID),
			})
		}
		return out
	}
	return build(0)
}

func (s *Service) mergeWorld(report *ImportReport, campaignID int64, pages []LibraryWorldNote) {
	local, err := s.ListWorldNotes(campaignID)
	if err != nil {
		report.Skipped = append(report.Skipped, fmt.Sprintf("world of campaign %d: %v", campaignID, err))
		return
	}
	byID := make(map[int64]WorldNote, len(local))
	siblings := map[int64][]WorldNote{}
	for _, n := range local {
		byID[n.ID] = n
		var parent int64
		if n.ParentID != nil {
			parent = *n.ParentID
		}
		siblings[parent] = append(siblings[parent], n)
	}
	levels := make(map[int64]claims, len(siblings))
	for parent, notes := range siblings {
		levels[parent] = claimsOf(notes, func(n WorldNote) (string, int64) { return n.Title, n.ID })
	}
	s.mergeWorldLevel(report, campaignID, nil, pages, levels, byID)
}

func (s *Service) mergeWorldLevel(
	report *ImportReport,
	campaignID int64,
	parentID *int64,
	pages []LibraryWorldNote,
	levels map[int64]claims,
	byID map[int64]WorldNote,
) {
	var parent int64
	if parentID != nil {
		parent = *parentID
	}
	for _, p := range pages {
		if _, err := validateWorldKind(p.Kind); err != nil {
			p.Kind = defaultWorldKind
		}
		in := WorldNoteInput{
			CampaignID: campaignID,
			ParentID:   parentID,
			Kind:       p.Kind,
			Title:      p.Title,
			WorldDate:  p.WorldDate,
			Body:       p.Body,
		}

		id, matched := levels[parent].take(p.Title)
		if matched {
			old := byID[id]
			norm, err := in.validate()
			if err != nil || norm.Kind != old.Kind || norm.Title != old.Title || norm.WorldDate != old.WorldDate || norm.Body != old.Body {
				in.ID = &id
				if _, err := s.SaveWorldNote(in); err != nil {
					report.Skipped = append(report.Skipped, fmt.Sprintf("world page %q: %v", p.Title, err))
				} else {
					report.Updated.WorldNotes++
				}
			}
		} else {
			saved, err := s.SaveWorldNote(in)
			if err != nil {
				if nested := worldPageCount(p.Children); nested > 0 {
					report.Skipped = append(report.Skipped, fmt.Sprintf("world page %q and the %d inside it: %v", p.Title, nested, err))
				} else {
					report.Skipped = append(report.Skipped, fmt.Sprintf("world page %q: %v", p.Title, err))
				}
				continue
			}
			id = saved.ID
			report.WorldNotes++
		}
		s.mergeWorldLevel(report, campaignID, &id, p.Children, levels, byID)
	}
}

func worldPageCount(pages []LibraryWorldNote) int {
	n := len(pages)
	for _, p := range pages {
		n += worldPageCount(p.Children)
	}
	return n
}

func remapPicks(picks []Pick, slugs map[string]string) []Pick {
	if len(slugs) == 0 {
		return picks
	}
	out := make([]Pick, 0, len(picks))
	for _, p := range picks {
		if mapped, ok := slugs[p.Slug]; ok {
			p.Slug = mapped
		}
		out = append(out, p)
	}
	return out
}
