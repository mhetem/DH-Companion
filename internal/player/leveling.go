package player

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/srd"
)

type LevelUpRecord struct {
	Level     int      `json:"level"`
	Tier      int      `json:"tier"`
	Choices   []string `json:"choices"`
	Summary   string   `json:"summary"`
	CreatedAt string   `json:"createdAt"`
}

type Effect struct {
	Traits          int  `json:"traits"`
	Evasion         int  `json:"evasion"`
	HPSlots         int  `json:"hpSlots"`
	StressSlots     int  `json:"stressSlots"`
	Experiences     int  `json:"experiences"`
	ExperienceBonus int  `json:"experienceBonus"`
	DomainCards     int  `json:"domainCards"`
	Proficiency     int  `json:"proficiency"`
	SubclassUpgrade int  `json:"subclassUpgrade"`
	Multiclass      bool `json:"multiclass"`
}

type PlanOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Slots       int    `json:"slots"`
	Cost        int    `json:"cost"`
	Effect      Effect `json:"effect"`
	Used        int    `json:"used"`
	Remaining   int    `json:"remaining"`
	Available   bool   `json:"available"`
	Reason      string `json:"reason"`
}

func effectView(e srd.AdvancementEffect) Effect {
	return Effect{
		Traits:          e.Traits,
		Evasion:         e.Evasion,
		HPSlots:         e.HPSlots,
		StressSlots:     e.StressSlots,
		Experiences:     e.Experiences,
		ExperienceBonus: e.ExperienceBonus,
		DomainCards:     e.DomainCards,
		Proficiency:     e.Proficiency,
		SubclassUpgrade: e.SubclassUpgrade,
		Multiclass:      e.Multiclass,
	}
}

type LevelUpPlan struct {
	CharacterID          int64                  `json:"characterId"`
	FromLevel            int                    `json:"fromLevel"`
	ToLevel              int                    `json:"toLevel"`
	FromTier             int                    `json:"fromTier"`
	Tier                 int                    `json:"tier"`
	TierName             string                 `json:"tierName"`
	NewTier              bool                   `json:"newTier"`
	Achievements         []string               `json:"achievements"`
	NeedsNewExperience   bool                   `json:"needsNewExperience"`
	ProficiencyBonus     int                    `json:"proficiencyBonus"`
	ClearsMarkedTraits   bool                   `json:"clearsMarkedTraits"`
	AdvancementsRequired int                    `json:"advancementsRequired"`
	ThresholdIncrease    int                    `json:"thresholdIncrease"`
	DomainCardsPerLevel  int                    `json:"domainCardsPerLevel"`
	Options              []PlanOption           `json:"options"`
	AvailableCards       []cards.DomainCard     `json:"availableCards"`
	Experiences          []Experience           `json:"experiences"`
	Traits               map[string]int         `json:"traits"`
	MarkedTraits         []string               `json:"markedTraits"`
	Classes              []cards.CharacterClass `json:"classes"`
}

type AdvancementChoice struct {
	ID                     string   `json:"id"`
	Traits                 []string `json:"traits"`
	ExperienceIDs          []int64  `json:"experienceIds"`
	DomainCardSlug         string   `json:"domainCardSlug"`
	MulticlassSlug         string   `json:"multiclassSlug"`
	MulticlassSubclassSlug string   `json:"multiclassSubclassSlug"`
}

type LevelUpInput struct {
	CharacterID       int64               `json:"characterId"`
	Choices           []AdvancementChoice `json:"choices"`
	NewExperienceName string              `json:"newExperienceName"`
	DomainCardSlug    string              `json:"domainCardSlug"`
}

func levelUpView(r db.CharacterLevelup) LevelUpRecord {
	rec := LevelUpRecord{
		Level:     int(r.Level),
		Tier:      int(r.Tier),
		Choices:   []string{},
		Summary:   r.Summary,
		CreatedAt: r.CreatedAt,
	}
	var choices []AdvancementChoice
	if err := json.Unmarshal([]byte(r.Choices), &choices); err == nil {
		for _, c := range choices {
			rec.Choices = append(rec.Choices, c.ID)
		}
	}
	return rec
}

func (s *Service) ListLevelUps(characterID int64) ([]LevelUpRecord, error) {
	rows, err := s.q.ListLevelUps(s.ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("listing level-ups: %w", err)
	}
	out := make([]LevelUpRecord, 0, len(rows))
	for _, r := range rows {
		out = append(out, levelUpView(r))
	}
	return out, nil
}

func (s *Service) PlanLevelUp(characterID int64) (LevelUpPlan, error) {
	row, err := s.character(characterID)
	if err != nil {
		return LevelUpPlan{}, err
	}
	fromLevel := int(row.Level)
	if fromLevel >= MaxLevel {
		return LevelUpPlan{}, fmt.Errorf("level %d is the cap", MaxLevel)
	}

	toLevel := fromLevel + 1
	tier := tierForLevel(toLevel)
	rules, ok := s.catalog.Tier(tier)
	if !ok {
		return LevelUpPlan{}, fmt.Errorf("no advancement table for tier %d", tier)
	}
	newTier := tierForLevel(fromLevel) != tier

	used, err := s.tierUsage(characterID, tier)
	if err != nil {
		return LevelUpPlan{}, err
	}

	view := s.characterView(row)
	marked := view.MarkedTraits
	if newTier && rules.ClearsMarkedTraits {
		marked = []string{}
	}

	plan := LevelUpPlan{
		CharacterID:          characterID,
		FromLevel:            fromLevel,
		ToLevel:              toLevel,
		FromTier:             tierForLevel(fromLevel),
		Tier:                 tier,
		TierName:             rules.Name,
		NewTier:              newTier,
		Achievements:         []string{},
		NeedsNewExperience:   newTier && rules.NewExperience,
		ClearsMarkedTraits:   newTier && rules.ClearsMarkedTraits,
		AdvancementsRequired: s.catalog.Leveling.AdvancementsPerLevel,
		ThresholdIncrease:    s.catalog.Leveling.ThresholdIncreasePerLevel,
		DomainCardsPerLevel:  s.catalog.Leveling.DomainCardsPerLevel,
		Options:              []PlanOption{},
		Traits:               view.Traits,
		MarkedTraits:         marked,
		Classes:              s.catalog.ListClasses(),
	}
	if newTier {
		plan.Achievements = rules.Achievements
		plan.ProficiencyBonus = rules.ProficiencyBonus
	}

	unmarked := 0
	for _, k := range TraitKeys {
		if !slices.Contains(marked, k) {
			unmarked++
		}
	}

	for _, a := range rules.Advancements {
		opt := PlanOption{
			ID:          a.ID,
			Label:       a.Label,
			Description: a.Description,
			Slots:       a.Slots,
			Cost:        a.Cost,
			Effect:      effectView(a.Effect),
			Used:        used[a.ID],
		}
		opt.Remaining = a.Slots - opt.Used
		opt.Available = opt.Remaining > 0
		if !opt.Available {
			opt.Reason = "no slots left this tier"
		}
		if opt.Available && a.Cost > plan.AdvancementsRequired {
			opt.Available = false
			opt.Reason = fmt.Sprintf("costs %d advancements", a.Cost)
		}
		if opt.Available && a.Effect.Traits > 0 && unmarked < a.Effect.Traits {
			opt.Available = false
			opt.Reason = "not enough unmarked traits"
		}
		if opt.Available && a.Effect.SubclassUpgrade > 0 && int(row.SubclassMastery) >= 3 {
			opt.Available = false
			opt.Reason = "subclass is already at mastery"
		}
		if opt.Available && a.Effect.Multiclass && row.MulticlassSlug != "" {
			opt.Available = false
			opt.Reason = "already multiclassed"
		}
		plan.Options = append(plan.Options, opt)
	}

	if plan.Experiences, err = s.ListExperiences(characterID); err != nil {
		return LevelUpPlan{}, err
	}
	if plan.AvailableCards, err = s.cardsUpToLevel(characterID, toLevel, view.Domains); err != nil {
		return LevelUpPlan{}, err
	}
	return plan, nil
}

func (s *Service) ApplyLevelUp(in LevelUpInput) (Sheet, error) {
	row, err := s.character(in.CharacterID)
	if err != nil {
		return Sheet{}, err
	}
	fromLevel := int(row.Level)
	if fromLevel >= MaxLevel {
		return Sheet{}, fmt.Errorf("level %d is the cap", MaxLevel)
	}

	toLevel := fromLevel + 1
	tier := tierForLevel(toLevel)
	rules, ok := s.catalog.Tier(tier)
	if !ok {
		return Sheet{}, fmt.Errorf("no advancement table for tier %d", tier)
	}
	newTier := tierForLevel(fromLevel) != tier

	used, err := s.tierUsage(in.CharacterID, tier)
	if err != nil {
		return Sheet{}, err
	}

	view := s.characterView(row)
	traits := view.Traits
	marked := append([]string{}, view.MarkedTraits...)
	proficiency := int(row.Proficiency)
	mastery := int(row.SubclassMastery)
	hpMax := int(row.HpMax)
	stressMax := int(row.StressMax)
	evasion := int(row.Evasion)
	major := int(row.ThresholdMajor)
	severe := int(row.ThresholdSevere)
	multiclass := row.MulticlassSlug
	multiSubclass := row.MulticlassSubclassSlug
	domains := append([]string{}, view.Domains...)

	if newTier {
		if rules.ClearsMarkedTraits {
			marked = []string{}
		}
		proficiency += rules.ProficiencyBonus
	}

	spent := 0
	taken := map[string]int{}
	summary := []string{}
	experienceBumps := map[int64]int{}
	extraCards := []string{}

	for _, choice := range in.Choices {
		adv, ok := findAdvancement(rules, choice.ID)
		if !ok {
			return Sheet{}, fmt.Errorf("%q is not a tier %d advancement", choice.ID, tier)
		}
		taken[adv.ID]++
		if used[adv.ID]+taken[adv.ID] > adv.Slots {
			return Sheet{}, fmt.Errorf("%q has no slots left this tier", adv.Label)
		}
		cost := adv.Cost
		if cost < 1 {
			cost = 1
		}
		spent += cost

		e := adv.Effect
		if e.Traits > 0 {
			if len(choice.Traits) != e.Traits {
				return Sheet{}, fmt.Errorf("%q needs %d traits, got %d", adv.Label, e.Traits, len(choice.Traits))
			}
			for _, raw := range choice.Traits {
				key, err := validateTraitKey(raw)
				if err != nil {
					return Sheet{}, err
				}
				if slices.Contains(marked, key) {
					return Sheet{}, fmt.Errorf("%s is already marked this tier", key)
				}
				traits[key] = clamp(traits[key]+1, -3, 12)
				marked = append(marked, key)
			}
			summary = append(summary, fmt.Sprintf("+1 %s", strings.Join(choice.Traits, ", +1 ")))
		}
		if e.Evasion > 0 {
			evasion = clamp(evasion+e.Evasion, 0, 30)
			summary = append(summary, fmt.Sprintf("+%d Evasion", e.Evasion))
		}
		if e.HPSlots > 0 {
			hpMax = clamp(hpMax+e.HPSlots, 1, 20)
			summary = append(summary, fmt.Sprintf("+%d HP slot", e.HPSlots))
		}
		if e.StressSlots > 0 {
			stressMax = clamp(stressMax+e.StressSlots, 1, 20)
			summary = append(summary, fmt.Sprintf("+%d Stress slot", e.StressSlots))
		}
		if e.Experiences > 0 {
			if len(choice.ExperienceIDs) != e.Experiences {
				return Sheet{}, fmt.Errorf("%q needs %d Experiences, got %d", adv.Label, e.Experiences, len(choice.ExperienceIDs))
			}
			bonus := e.ExperienceBonus
			if bonus < 1 {
				bonus = 1
			}
			for _, id := range choice.ExperienceIDs {
				experienceBumps[id] += bonus
			}
			summary = append(summary, fmt.Sprintf("+%d to %d Experiences", bonus, e.Experiences))
		}
		if e.DomainCards > 0 {
			slug := strings.TrimSpace(choice.DomainCardSlug)
			if slug == "" {
				return Sheet{}, fmt.Errorf("%q needs a domain card", adv.Label)
			}
			extraCards = append(extraCards, slug)
			summary = append(summary, "extra domain card")
		}
		if e.Proficiency > 0 {
			proficiency = clamp(proficiency+e.Proficiency, 1, 6)
			summary = append(summary, fmt.Sprintf("+%d Proficiency", e.Proficiency))
		}
		if e.SubclassUpgrade > 0 {
			if mastery >= 3 {
				return Sheet{}, fmt.Errorf("your subclass is already at mastery")
			}
			mastery = clamp(mastery+e.SubclassUpgrade, 1, 3)
			summary = append(summary, "subclass upgrade")
		}
		if e.Multiclass {
			if multiclass != "" {
				return Sheet{}, fmt.Errorf("you have already multiclassed")
			}
			slug := strings.TrimSpace(choice.MulticlassSlug)
			subSlug := strings.TrimSpace(choice.MulticlassSubclassSlug)
			if slug == "" || subSlug == "" {
				return Sheet{}, fmt.Errorf("multiclassing needs a class and a subclass")
			}
			if slug == row.ClassSlug {
				return Sheet{}, fmt.Errorf("pick a class other than your own")
			}
			class, ok := s.catalog.Class(slug)
			if !ok {
				return Sheet{}, notFound("class", slug)
			}
			if _, ok := class.Subclass(subSlug); !ok {
				return Sheet{}, fmt.Errorf("%s is not a subclass of %s", subSlug, class.Name)
			}
			multiclass, multiSubclass = slug, subSlug
			for _, d := range class.Domains {
				if !contains(domains, d) {
					domains = append(domains, d)
				}
			}
			summary = append(summary, "multiclass into "+class.Name)
		}
	}

	if spent != s.catalog.Leveling.AdvancementsPerLevel {
		return Sheet{}, fmt.Errorf("pick %d advancements — you picked %d", s.catalog.Leveling.AdvancementsPerLevel, spent)
	}

	levelCard := strings.TrimSpace(in.DomainCardSlug)
	if s.catalog.Leveling.DomainCardsPerLevel > 0 && levelCard == "" {
		return Sheet{}, fmt.Errorf("choose a domain card for level %d", toLevel)
	}

	newCards := []string{}
	if levelCard != "" {
		newCards = append(newCards, levelCard)
	}
	newCards = append(newCards, extraCards...)
	for _, slug := range newCards {
		if err := s.assertCardEligible(in.CharacterID, slug, toLevel, domains); err != nil {
			return Sheet{}, err
		}
	}
	if err := assertDistinct(newCards); err != nil {
		return Sheet{}, err
	}

	newExperience := strings.TrimSpace(in.NewExperienceName)
	if newTier && rules.NewExperience && newExperience == "" {
		return Sheet{}, fmt.Errorf("entering tier %d grants a new Experience — name it", tier)
	}

	major += s.catalog.Leveling.ThresholdIncreasePerLevel
	severe += s.catalog.Leveling.ThresholdIncreasePerLevel
	major = clamp(major, 0, 99)
	severe = clamp(severe, 0, 99)

	choicesJSON, err := json.Marshal(in.Choices)
	if err != nil {
		return Sheet{}, fmt.Errorf("recording level-up: %w", err)
	}
	if newTier {
		summary = append([]string{rules.Name + " achievements"}, summary...)
	}

	err = s.tx(func(q *db.Queries) error {
		if _, err := q.UpdateCharacterProgress(s.ctx, db.UpdateCharacterProgressParams{
			Level:                  int64(toLevel),
			Proficiency:            int64(proficiency),
			SubclassMastery:        int64(mastery),
			MulticlassSlug:         multiclass,
			MulticlassSubclassSlug: multiSubclass,
			Agility:                int64(traits["agility"]),
			Strength:               int64(traits["strength"]),
			Finesse:                int64(traits["finesse"]),
			Instinct:               int64(traits["instinct"]),
			Presence:               int64(traits["presence"]),
			Knowledge:              int64(traits["knowledge"]),
			MarkedTraits:           encodeTraitList(marked),
			HpMax:                  int64(hpMax),
			StressMax:              int64(stressMax),
			Evasion:                int64(evasion),
			ThresholdMajor:         int64(major),
			ThresholdSevere:        int64(severe),
			ID:                     in.CharacterID,
		}); err != nil {
			return fmt.Errorf("applying level-up: %w", err)
		}

		for id, bonus := range experienceBumps {
			if _, err := q.AdjustExperienceModifier(s.ctx, db.AdjustExperienceModifierParams{
				Delta: int64(bonus), ID: id,
			}); err != nil {
				return fmt.Errorf("raising experience: %w", err)
			}
		}

		if newExperience != "" {
			if _, err := q.CreateExperience(s.ctx, db.CreateExperienceParams{
				CharacterID: in.CharacterID,
				Name:        newExperience,
				Modifier:    2,
			}); err != nil {
				return fmt.Errorf("adding experience: %w", err)
			}
		}

		held, err := q.CountCharacterLoadout(s.ctx, in.CharacterID)
		if err != nil {
			return fmt.Errorf("counting loadout: %w", err)
		}
		for _, slug := range newCards {
			location := "vault"
			if held < LoadoutMax {
				location = "loadout"
				held++
			}
			if _, err := q.AddCharacterDomainCard(s.ctx, db.AddCharacterDomainCardParams{
				CharacterID: in.CharacterID,
				CardSlug:    slug,
				Location:    location,
			}); err != nil {
				return fmt.Errorf("adding domain card: %w", err)
			}
		}

		if _, err := q.CreateLevelUp(s.ctx, db.CreateLevelUpParams{
			CharacterID: in.CharacterID,
			Level:       int64(toLevel),
			Tier:        int64(tier),
			Choices:     string(choicesJSON),
			Summary:     strings.Join(summary, " · "),
		}); err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("level %d is already recorded", toLevel)
			}
			return fmt.Errorf("recording level-up: %w", err)
		}
		return nil
	})
	if err != nil {
		return Sheet{}, err
	}

	return s.GetSheet(in.CharacterID)
}

func (s *Service) tierUsage(characterID int64, tier int) (map[string]int, error) {
	rows, err := s.q.ListLevelUpsForTier(s.ctx, db.ListLevelUpsForTierParams{
		CharacterID: characterID,
		Tier:        int64(tier),
	})
	if err != nil {
		return nil, fmt.Errorf("loading level history: %w", err)
	}
	used := map[string]int{}
	for _, r := range rows {
		var choices []AdvancementChoice
		if err := json.Unmarshal([]byte(r.Choices), &choices); err != nil {
			continue
		}
		for _, c := range choices {
			used[c.ID]++
		}
	}
	return used, nil
}

func (s *Service) cardsUpToLevel(characterID int64, level int, domains []string) ([]cards.DomainCard, error) {
	owned, err := s.q.ListCharacterDomainCards(s.ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("listing domain cards: %w", err)
	}
	held := map[string]bool{}
	for _, o := range owned {
		held[o.CardSlug] = true
	}
	out := []cards.DomainCard{}
	for _, d := range s.catalog.ListDomainCards() {
		if held[d.Slug] || cardLevel(d) > level {
			continue
		}
		if len(domains) > 0 && !contains(domains, d.Domain) {
			continue
		}
		out = append(out, d)
	}
	return out, nil
}

func (s *Service) assertCardEligible(characterID int64, slug string, level int, domains []string) error {
	card, ok := s.catalog.DomainCard(slug)
	if !ok {
		return notFound("domain card", slug)
	}
	if cardLevel(card) > level {
		return fmt.Errorf("%s is a level %s card", card.Name, card.Level)
	}
	if len(domains) > 0 && !contains(domains, card.Domain) {
		return fmt.Errorf("%s is not one of your domains", card.Domain)
	}
	owned, err := s.q.ListCharacterDomainCards(s.ctx, characterID)
	if err != nil {
		return fmt.Errorf("listing domain cards: %w", err)
	}
	for _, o := range owned {
		if o.CardSlug == slug {
			return fmt.Errorf("you already have %s", card.Name)
		}
	}
	return nil
}

func assertDistinct(slugs []string) error {
	seen := map[string]bool{}
	for _, s := range slugs {
		if seen[s] {
			return fmt.Errorf("you picked %q twice", s)
		}
		seen[s] = true
	}
	return nil
}

func findAdvancement(t srd.Tier, id string) (srd.Advancement, bool) {
	for _, a := range t.Advancements {
		if a.ID == id {
			return a, true
		}
	}
	return srd.Advancement{}, false
}
