package gm

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mhetem/DH-Companion/internal/cards"
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
	Encounters         []LibraryEncounter  `json:"encounters"`
	Campaigns          []LibraryCampaign   `json:"campaigns"`
	Countdowns         []LibraryCountdown  `json:"countdowns"`
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

type LibraryCountdown struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
	Max   int    `json:"max"`
	Kind  string `json:"kind"`
}

type ImportReport struct {
	Parties            int      `json:"parties"`
	CustomAdversaries  int      `json:"customAdversaries"`
	CustomEnvironments int      `json:"customEnvironments"`
	Encounters         int      `json:"encounters"`
	Campaigns          int      `json:"campaigns"`
	Sessions           int      `json:"sessions"`
	Notes              int      `json:"notes"`
	Countdowns         int      `json:"countdowns"`
	Renamed            []string `json:"renamed"`
	Skipped            []string `json:"skipped"`
}

func (s *Service) buildLibrary() (Library, error) {
	lib := Library{
		Format:             LibraryFormat,
		App:                libraryApp,
		ExportedAt:         time.Now().UTC().Format(time.RFC3339),
		Parties:            []LibraryParty{},
		CustomAdversaries:  []cards.Adversary{},
		CustomEnvironments: []cards.Environment{},
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

		clocks, err := s.ListCountdownsForCampaign(c.ID)
		if err != nil {
			return Library{}, err
		}
		for _, cd := range clocks {
			entry.Countdowns = append(entry.Countdowns, LibraryCountdown{Name: cd.Name, Value: cd.Value, Max: cd.Max, Kind: cd.Kind})
		}

		lib.Campaigns = append(lib.Campaigns, entry)
	}

	loose, err := s.ListUnassignedCountdowns()
	if err != nil {
		return Library{}, err
	}
	for _, cd := range loose {
		lib.Countdowns = append(lib.Countdowns, LibraryCountdown{Name: cd.Name, Value: cd.Value, Max: cd.Max, Kind: cd.Kind})
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

	adversarySlugs := map[string]string{}
	for _, card := range lib.CustomAdversaries {
		original, oldSlug := card.Name, card.Slug
		name, renamed, err := s.freeName(original, s.adversarySlugTaken)
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("adversary %q: %v", original, err))
			continue
		}
		card.Name, card.Slug = name, ""
		saved, err := s.CreateCustomAdversary(card)
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("adversary %q: %v", name, err))
			continue
		}
		if oldSlug != "" {
			adversarySlugs[oldSlug] = saved.Slug
		}
		if renamed {
			report.Renamed = append(report.Renamed, fmt.Sprintf("%s → %s", original, name))
		}
		report.CustomAdversaries++
	}

	environmentSlugs := map[string]string{}
	for _, card := range lib.CustomEnvironments {
		original, oldSlug := card.Name, card.Slug
		name, renamed, err := s.freeName(original, s.environmentSlugTaken)
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("environment %q: %v", original, err))
			continue
		}
		card.Name, card.Slug = name, ""
		saved, err := s.CreateCustomEnvironment(card)
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("environment %q: %v", name, err))
			continue
		}
		if oldSlug != "" {
			environmentSlugs[oldSlug] = saved.Slug
		}
		if renamed {
			report.Renamed = append(report.Renamed, fmt.Sprintf("%s → %s", original, name))
		}
		report.CustomEnvironments++
	}

	existingParties, err := s.ListParties()
	if err != nil {
		return report, err
	}
	partyIDs := map[string]int64{}
	for _, p := range existingParties {
		partyIDs[p.Name] = p.ID
	}
	for _, p := range lib.Parties {
		if _, ok := partyIDs[p.Name]; ok {
			continue
		}
		saved, err := s.CreateParty(PartyInput{Name: p.Name, Size: p.Size, Tier: p.Tier})
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("party %q: %v", p.Name, err))
			continue
		}
		partyIDs[p.Name] = saved.ID
		report.Parties++
	}

	encounterIDs := map[string]int64{}
	for _, e := range lib.Encounters {
		in := EncounterInput{
			Name:              e.Name,
			Adversaries:       e.Adversaries,
			CustomAdversaries: remapPicks(e.CustomAdversaries, adversarySlugs),
		}
		if e.PartyName != nil {
			if id, ok := partyIDs[*e.PartyName]; ok {
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
		saved, err := s.SaveEncounter(in)
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("encounter %q: %v", e.Name, err))
			continue
		}
		encounterIDs[e.Name] = saved.ID
		report.Encounters++
	}

	for _, c := range lib.Campaigns {
		campaign, err := s.SaveCampaign(CampaignInput{Name: c.Name, Description: c.Description})
		if err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("campaign %q: %v", c.Name, err))
			continue
		}
		report.Campaigns++

		if c.CurrentFear > 0 {
			if _, err := s.SetCampaignFear(campaign.ID, c.CurrentFear); err != nil {
				report.Skipped = append(report.Skipped, fmt.Sprintf("campaign %q fear: %v", c.Name, err))
			}
		}

		if c.MasterNote != "" {
			if _, err := s.SaveMasterNote(campaign.ID, c.MasterNote); err != nil {
				report.Skipped = append(report.Skipped, fmt.Sprintf("campaign %q master note: %v", c.Name, err))
			}
		}

		for _, sess := range c.Sessions {
			view, err := s.SaveSession(SessionInput{
				CampaignID: campaign.ID,
				Number:     sess.Number,
				Title:      sess.Title,
				Date:       sess.Date,
				Recap:      sess.Recap,
			})
			if err != nil {
				report.Skipped = append(report.Skipped, fmt.Sprintf("session %q: %v", sess.Title, err))
				continue
			}
			report.Sessions++
			for _, name := range sess.Encounters {
				id, ok := encounterIDs[name]
				if !ok {
					continue
				}
				if _, err := s.LinkEncounter(view.ID, id); err != nil {
					report.Skipped = append(report.Skipped, fmt.Sprintf("linking %q to session %d: %v", name, view.Number, err))
				}
			}
		}

		for _, n := range c.Notes {
			if _, err := s.SaveNote(NoteInput{CampaignID: campaign.ID, Kind: n.Kind, Title: n.Title, Body: n.Body}); err != nil {
				report.Skipped = append(report.Skipped, fmt.Sprintf("note %q: %v", n.Title, err))
				continue
			}
			report.Notes++
		}

		for _, cd := range c.Countdowns {
			id := campaign.ID
			if _, err := s.SaveCountdown(CountdownInput{CampaignID: &id, Name: cd.Name, Value: cd.Value, Max: cd.Max, Kind: cd.Kind}); err != nil {
				report.Skipped = append(report.Skipped, fmt.Sprintf("countdown %q: %v", cd.Name, err))
				continue
			}
			report.Countdowns++
		}
	}

	for _, cd := range lib.Countdowns {
		if _, err := s.SaveCountdown(CountdownInput{Name: cd.Name, Value: cd.Value, Max: cd.Max, Kind: cd.Kind}); err != nil {
			report.Skipped = append(report.Skipped, fmt.Sprintf("countdown %q: %v", cd.Name, err))
			continue
		}
		report.Countdowns++
	}

	return report, nil
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
