package player

import (
	"fmt"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

type Bonus struct {
	Source string `json:"source"`
	Effect string `json:"effect"`
}

type CreationBonus struct {
	Evasion   int     `json:"evasion"`
	HPMax     int     `json:"hpMax"`
	StressMax int     `json:"stressMax"`
	Major     int     `json:"major"`
	Severe    int     `json:"severe"`
	Sources   []Bonus `json:"sources"`
}

type abilityContext struct {
	armored     bool
	armed       bool
	proficiency int
	agility     int
	domains     map[string]int
}

var classAbilities = map[string]func(abilityContext) statMods{
	"brawler/I Am the Weapon": func(c abilityContext) statMods {
		if c.armed {
			return statMods{}
		}
		return statMods{evasion: 1}
	},
}

var subclassAbilities = map[string]statMods{
	"stalwart/Unwavering":         {major: 1, severe: 1},
	"stalwart/Unrelenting":        {major: 2, severe: 2},
	"stalwart/Undaunted":          {major: 3, severe: 3},
	"vengeance/At Ease":           {stressMax: 1},
	"nightwalker/Fleeting Shadow": {evasion: 1},
	"wingedSentinel/Ascendant":    {severe: 4},
	"schoolOfWar/Battlemage":      {hpMax: 1},
	"juggernaut/Rugged":           {severe: 3},
}

var ancestryAbilities = map[string]func(abilityContext) statMods{
	"galapa/Shell": func(c abilityContext) statMods {
		return statMods{major: c.proficiency, severe: c.proficiency}
	},
}

var bareBonesThresholds = map[int][2]int{
	1: {9, 19},
	2: {11, 24},
	3: {13, 31},
	4: {15, 38},
}

var creationAbilities = map[string]statMods{
	"simiah/Nimble":      {evasion: 1},
	"giant/Endurance":    {hpMax: 1},
	"human/High Stamina": {stressMax: 1},
}

var cardAbilities = map[string]func(abilityContext) statMods{
	"fortifiedArmor": func(c abilityContext) statMods {
		if !c.armored {
			return statMods{}
		}
		return statMods{major: 2, severe: 2}
	},
	"armorer": func(c abilityContext) statMods {
		if !c.armored {
			return statMods{}
		}
		return statMods{armorScore: 1}
	},
	"riseUp": func(c abilityContext) statMods {
		return statMods{severe: c.proficiency}
	},
	"untouchable": func(c abilityContext) statMods {
		return statMods{evasion: c.agility / 2}
	},
	"bladeTouched": func(c abilityContext) statMods {
		if c.domains["Blade"] < 4 {
			return statMods{}
		}
		return statMods{severe: 4}
	},
	"valorTouched": func(c abilityContext) statMods {
		if c.domains["Valor"] < 4 {
			return statMods{}
		}
		return statMods{armorScore: 1}
	},
	"splendorTouched": func(c abilityContext) statMods {
		if c.domains["Splendor"] < 4 {
			return statMods{}
		}
		return statMods{severe: 3}
	},
}

func masteryNeeded(featureType string) int {
	switch featureType {
	case cards.FeatureSpecialization:
		return 2
	case cards.FeatureMastery:
		return 3
	default:
		return 1
	}
}

func signedMod(n int) string {
	if n < 0 {
		return fmt.Sprintf("−%d", -n)
	}
	return fmt.Sprintf("+%d", n)
}

func describeMods(m statMods) string {
	parts := []string{}
	if m.armorScore != 0 {
		parts = append(parts, signedMod(m.armorScore)+" Armor Score")
	}
	if m.evasion != 0 {
		parts = append(parts, signedMod(m.evasion)+" Evasion")
	}
	switch {
	case m.major != 0 && m.major == m.severe:
		parts = append(parts, signedMod(m.major)+" to damage thresholds")
	default:
		if m.major != 0 {
			parts = append(parts, signedMod(m.major)+" Major threshold")
		}
		if m.severe != 0 {
			parts = append(parts, signedMod(m.severe)+" Severe threshold")
		}
	}
	if m.hpMax != 0 {
		parts = append(parts, signedMod(m.hpMax)+" Hit Point slot")
	}
	if m.stressMax != 0 {
		parts = append(parts, signedMod(m.stressMax)+" Stress slot")
	}
	return strings.Join(parts, ", ")
}

func (snap *snapshot) apply(source string, m statMods) {
	if m == (statMods{}) {
		return
	}
	snap.mods.add(m)
	snap.bonuses = append(snap.bonuses, Bonus{Source: source, Effect: describeMods(m)})
}

func (s *Service) classBonuses(snap *snapshot, classSlug string, c abilityContext) {
	class, ok := s.catalog.Class(classSlug)
	if !ok {
		return
	}
	for _, f := range class.Features {
		if fn, ok := classAbilities[class.Slug+"/"+f.Title]; ok {
			snap.apply(f.Title, fn(c))
		}
	}
}

func (s *Service) subclassBonuses(snap *snapshot, classSlug, subclassSlug string, mastery int) {
	class, ok := s.catalog.Class(classSlug)
	if !ok {
		return
	}
	sub, ok := class.Subclass(subclassSlug)
	if !ok {
		return
	}
	for _, f := range sub.Features {
		if masteryNeeded(f.Type) > mastery {
			continue
		}
		if m, ok := subclassAbilities[sub.Slug+"/"+f.Title]; ok {
			snap.apply(f.Title, m)
		}
	}
}

func (s *Service) abilityBonuses(snap *snapshot, row db.Character, held []db.CharacterDomainCard) {
	c := abilityContext{
		armored:     snap.hasArmor,
		armed:       snap.armed,
		proficiency: int(row.Proficiency),
		agility:     int(row.Agility),
		domains:     map[string]int{},
	}
	loadout := []cards.DomainCard{}
	for _, h := range held {
		if h.Location != "loadout" {
			continue
		}
		if card, ok := s.catalog.DomainCard(h.CardSlug); ok {
			loadout = append(loadout, card)
			c.domains[card.Domain]++
		}
	}

	s.classBonuses(snap, row.ClassSlug, c)
	s.subclassBonuses(snap, row.ClassSlug, row.SubclassSlug, int(row.SubclassMastery))
	if row.MulticlassSlug != "" {
		s.subclassBonuses(snap, row.MulticlassSlug, row.MulticlassSubclassSlug, 1)
	}
	if a, ok := s.catalog.Ancestry(row.AncestrySlug); ok {
		for _, f := range a.Features {
			if fn, ok := ancestryAbilities[row.AncestrySlug+"/"+f.Title]; ok {
				snap.apply(f.Title, fn(c))
			}
		}
	}
	for _, card := range loadout {
		if fn, ok := cardAbilities[card.Slug]; ok {
			snap.apply(card.Name, fn(c))
		}
		if card.Slug == "bareBones" && !snap.hasArmor {
			tier := tierForLevel(int(row.Level))
			t := bareBonesThresholds[tier]
			score := max(3+int(row.Strength), 0)
			snap.base = defenseBase{
				key:    fmt.Sprintf("bareBones:%d:%d", tier, row.Strength),
				known:  true,
				score:  score,
				major:  t[0],
				severe: t[1],
			}
			snap.bonuses = append(snap.bonuses, Bonus{
				Source: card.Name,
				Effect: fmt.Sprintf("base Armor Score %d, base thresholds %d/%d", score, t[0], t[1]),
			})
		}
	}
}

func (s *Service) CreationBonuses(classSlug, subclassSlug, ancestrySlug string) CreationBonus {
	var snap snapshot
	c := abilityContext{proficiency: 1, domains: map[string]int{}}
	s.classBonuses(&snap, classSlug, c)
	s.subclassBonuses(&snap, classSlug, subclassSlug, 1)
	if a, ok := s.catalog.Ancestry(ancestrySlug); ok {
		for _, f := range a.Features {
			key := ancestrySlug + "/" + f.Title
			if fn, ok := ancestryAbilities[key]; ok {
				snap.apply(f.Title, fn(c))
			}
			if m, ok := creationAbilities[key]; ok {
				snap.apply(f.Title, m)
			}
		}
	}
	sources := snap.bonuses
	if sources == nil {
		sources = []Bonus{}
	}
	return CreationBonus{
		Evasion:   snap.mods.evasion,
		HPMax:     snap.mods.hpMax,
		StressMax: snap.mods.stressMax,
		Major:     snap.mods.major,
		Severe:    snap.mods.severe,
		Sources:   sources,
	}
}
