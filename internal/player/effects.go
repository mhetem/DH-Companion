package player

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/mhetem/DH-Companion/internal/db"
)

type EffectOption struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Label  string `json:"label"`
	Effect string `json:"effect"`
	Ends   string `json:"ends"`
	Active bool   `json:"active"`
}

type liveContext struct {
	row     db.Character
	tier    int
	loadout map[string]bool
}

func (c liveContext) has(subclass string, mastery int) bool {
	if c.row.SubclassSlug == subclass && int(c.row.SubclassMastery) >= mastery {
		return true
	}
	return mastery <= 1 && c.row.MulticlassSubclassSlug == subclass
}

type effectDef struct {
	id           string
	source       string
	label        string
	ends         string
	group        string
	endsOnSevere bool
	endsOnHit    bool
	endsOnRest   bool
	blocksArmor  bool
	available    func(liveContext) bool
	mods         func(liveContext) statMods
}

func proficiencyThresholds(c liveContext) statMods {
	return statMods{major: int(c.row.Proficiency), severe: int(c.row.Proficiency)}
}

var effectDefs = []effectDef{
	{
		id:           "patronsMantle",
		source:       "Patron's Mantle",
		label:        "Cloaked in your Patron",
		ends:         "Severe damage or the end of the scene",
		endsOnSevere: true,
		endsOnRest:   true,
		available:    func(c liveContext) bool { return c.has("pactOfTheEndless", 1) },
		mods:         func(c liveContext) statMods { return statMods{major: c.tier, severe: c.tier} },
	},
	{
		id:         "roguesDodge",
		source:     "Rogue's Dodge",
		label:      "Dodging",
		ends:       "the next attack that succeeds against you, or your next rest",
		endsOnHit:  true,
		endsOnRest: true,
		available:  func(c liveContext) bool { return c.row.ClassSlug == "rogue" },
		mods:       func(liveContext) statMods { return statMods{evasion: 2} },
	},
	{
		id:           "channelEarth",
		source:       "Elemental Incarnation",
		label:        "Channeling Earth",
		group:        "channel",
		ends:         "Severe damage or your next rest",
		endsOnSevere: true,
		endsOnRest:   true,
		available:    func(c liveContext) bool { return c.has("wardenOfTheElements", 1) },
		mods:         proficiencyThresholds,
	},
	{
		id:           "channelAir",
		source:       "Elemental Dominion",
		label:        "Channeling Air",
		group:        "channel",
		ends:         "Severe damage or your next rest",
		endsOnSevere: true,
		endsOnRest:   true,
		available:    func(c liveContext) bool { return c.has("wardenOfTheElements", 3) },
		mods:         func(liveContext) statMods { return statMods{evasion: 1} },
	},
	{
		id:         "transcendenceSevere",
		source:     "Transcendence",
		label:      "Transcended, +4 Severe",
		ends:       "your next rest",
		endsOnRest: true,
		available:  func(c liveContext) bool { return c.has("elementalOrigin", 3) },
		mods:       func(liveContext) statMods { return statMods{severe: 4} },
	},
	{
		id:         "transcendenceEvasion",
		source:     "Transcendence",
		label:      "Transcended, +2 Evasion",
		ends:       "your next rest",
		endsOnRest: true,
		available:  func(c liveContext) bool { return c.has("elementalOrigin", 3) },
		mods:       func(liveContext) statMods { return statMods{evasion: 2} },
	},
	{
		id:         "circleOfPower",
		source:     "Circle of Power",
		label:      "Inside your circle",
		ends:       "the last token, or leaving the circle",
		endsOnRest: true,
		available:  func(c liveContext) bool { return c.has("hedge", 3) },
		mods:       func(liveContext) statMods { return statMods{evasion: 2, major: 2, severe: 2} },
	},
	{
		id:        "lunarFull",
		source:    "Lunar Phases",
		label:     "Full moon",
		group:     "lunar",
		ends:      "the end of the session",
		available: func(c liveContext) bool { return c.has("moon", 3) },
		mods:      func(liveContext) statMods { return statMods{major: 3, severe: 3} },
	},
	{
		id:        "lunarWaning",
		source:    "Lunar Phases",
		label:     "Waning moon",
		group:     "lunar",
		ends:      "the end of the session",
		available: func(c liveContext) bool { return c.has("moon", 3) },
		mods:      func(liveContext) statMods { return statMods{evasion: 1} },
	},
	{
		id:        "shadowhunter",
		source:    "Shadowhunter",
		label:     "In low light or darkness",
		ends:      "stepping out of the shadows",
		available: func(c liveContext) bool { return c.loadout["shadowhunter"] },
		mods:      func(liveContext) statMods { return statMods{evasion: 1} },
	},
	{
		id:          "frenzy",
		source:      "Frenzy",
		label:       "Frenzied",
		ends:        "no adversaries left in sight - and you can't use Armor Slots while it lasts",
		endsOnRest:  true,
		blocksArmor: true,
		available:   func(c liveContext) bool { return c.loadout["frenzy"] },
		mods:        func(liveContext) statMods { return statMods{severe: 8} },
	},
}

var beastformAbilities = map[string]statMods{
	"powerfulBeast/Thick Hide":      {major: 2, severe: 2},
	"wingedBeast/Hollow Bones":      {major: -2, severe: -2},
	"mightyLizard/Physical Defense": {major: 3, severe: 3},
	"massiveBehemoth/Undaunted":     {major: 2, severe: 2},
}

func findEffect(id string) (effectDef, bool) {
	for _, def := range effectDefs {
		if def.id == id {
			return def, true
		}
	}
	return effectDef{}, false
}

func (s *Service) liveContextOf(row db.Character, held []db.CharacterDomainCard) liveContext {
	c := liveContext{row: row, tier: tierForLevel(int(row.Level)), loadout: map[string]bool{}}
	for _, h := range held {
		if h.Location == "loadout" {
			c.loadout[h.CardSlug] = true
		}
	}
	return c
}

func (s *Service) liveBonuses(row db.Character, held []db.CharacterDomainCard) (statMods, []Bonus) {
	c := s.liveContextOf(row, held)
	var live snapshot

	if c.loadout["eldritchFlesh"] && row.StressMarked > 0 {
		stress := int(row.StressMarked)
		live.apply("Eldritch Flesh", statMods{major: stress, severe: stress})
	}
	if c.has("schoolOfWar", 2) && row.Hope >= 2 {
		live.apply("Conjure Shield", statMods{evasion: int(row.Proficiency)})
	}
	switch row.ActiveStance {
	case "Anchored":
		live.apply("Anchored stance", statMods{major: 2, severe: 2})
	case "Aggressive":
		live.apply("Aggressive stance", statMods{evasion: -1})
	}
	if row.BeastformSlug != "" {
		if form, ok := s.catalog.Beastform(row.BeastformSlug); ok {
			m := statMods{evasion: signedInt(form.EvasionBonus)}
			for _, f := range form.Features {
				if fm, ok := beastformAbilities[form.Slug+"/"+f.Title]; ok {
					m.add(fm)
				}
			}
			live.apply(form.Name, m)
		}
	}
	for _, id := range decodeTraitList(row.ActiveEffects) {
		if def, ok := findEffect(id); ok && def.available(c) {
			live.apply(def.source, def.mods(c))
		}
	}
	return live.mods, live.bonuses
}

func (s *Service) effectOptions(row db.Character, held []db.CharacterDomainCard) []EffectOption {
	c := s.liveContextOf(row, held)
	active := decodeTraitList(row.ActiveEffects)
	out := []EffectOption{}
	for _, def := range effectDefs {
		if !def.available(c) {
			continue
		}
		out = append(out, EffectOption{
			ID:     def.id,
			Source: def.source,
			Label:  def.label,
			Effect: describeMods(def.mods(c)),
			Ends:   def.ends,
			Active: slices.Contains(active, def.id),
		})
	}
	return out
}

func (s *Service) writeEffects(q *db.Queries, characterID int64, active []string) (db.Character, error) {
	if active == nil {
		active = []string{}
	}
	raw, err := json.Marshal(active)
	if err != nil {
		return db.Character{}, fmt.Errorf("encoding effects: %w", err)
	}
	row, err := q.SetCharacterEffects(s.ctx, db.SetCharacterEffectsParams{ActiveEffects: string(raw), ID: characterID})
	if err != nil {
		return db.Character{}, fmt.Errorf("saving effects: %w", err)
	}
	return row, nil
}

func (s *Service) SetEffect(characterID int64, id string, on bool) (Character, error) {
	row, err := s.character(characterID)
	if err != nil {
		return Character{}, err
	}
	def, ok := findEffect(id)
	if !ok {
		return Character{}, notFound("effect", id)
	}
	held, err := s.q.ListCharacterDomainCards(s.ctx, characterID)
	if err != nil {
		return Character{}, fmt.Errorf("listing domain cards: %w", err)
	}
	if on && !def.available(s.liveContextOf(row, held)) {
		return Character{}, fmt.Errorf("%s isn't something this character has", def.source)
	}

	active := slices.DeleteFunc(decodeTraitList(row.ActiveEffects), func(other string) bool {
		if other == id {
			return true
		}
		o, ok := findEffect(other)
		return on && ok && def.group != "" && o.group == def.group
	})
	if on {
		active = append(active, id)
	}
	updated, err := s.writeEffects(s.q, characterID, active)
	if err != nil {
		return Character{}, err
	}
	return s.characterView(updated), nil
}

func endedEffects(active []string, severe, hit, rest bool) (kept []string, ended []string) {
	for _, id := range active {
		def, ok := findEffect(id)
		if ok && ((severe && def.endsOnSevere) || (hit && def.endsOnHit) || (rest && def.endsOnRest)) {
			ended = append(ended, def.source)
			continue
		}
		kept = append(kept, id)
	}
	return kept, ended
}

func armorBlocked(row db.Character) (string, bool) {
	for _, id := range decodeTraitList(row.ActiveEffects) {
		if def, ok := findEffect(id); ok && def.blocksArmor {
			return def.source, true
		}
	}
	return "", false
}
