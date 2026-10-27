package player

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/dice"
)

const (
	FavorMax      = 6
	FocusMax      = 6
	StartingFavor = 3

	warlockClass = "warlock"
	brawlerClass = "brawler"

	maxComboDie   = 12
	maxComboRolls = 100
)

type FocusRoll struct {
	Rolls     []int     `json:"rolls"`
	Focus     int       `json:"focus"`
	Character Character `json:"character"`
}

type ComboResult struct {
	Die       int       `json:"die"`
	Rolls     []int     `json:"rolls"`
	Total     int       `json:"total"`
	Paid      string    `json:"paid"`
	Character Character `json:"character"`
}

func usesFavor(row db.Character) bool {
	return row.ClassSlug == warlockClass || row.MulticlassSlug == warlockClass
}

func usesFocus(row db.Character) bool {
	return row.SubclassSlug == martialArtistSubclass || row.MulticlassSubclassSlug == martialArtistSubclass
}

func hasFlowState(row db.Character) bool {
	return row.SubclassSlug == martialArtistSubclass && row.SubclassMastery >= 3
}

func (s *Service) favorTrait(row db.Character) int {
	classSlug, subSlug := row.ClassSlug, row.SubclassSlug
	if row.ClassSlug != warlockClass {
		classSlug, subSlug = row.MulticlassSlug, row.MulticlassSubclassSlug
	}
	class, ok := s.catalog.Class(classSlug)
	if !ok {
		return 0
	}
	sub, ok := class.Subclass(subSlug)
	if !ok {
		return 0
	}
	key := strings.ToLower(sub.SpellcastTrait)
	return max(s.characterView(row).Traits[key], 0)
}

func (s *Service) SetClassTracks(characterID int64, favor, focus int) (Character, error) {
	row, err := s.q.SetCharacterTracks(s.ctx, db.SetCharacterTracksParams{
		Favor: int64(favor),
		Focus: int64(focus),
		ID:    characterID,
	})
	return s.vitalsResult(characterID, row, err, "setting Favor and Focus")
}

func (s *Service) FavorInsteadOfHope(characterID int64) (Character, error) {
	row, err := s.character(characterID)
	if err != nil {
		return Character{}, err
	}
	if !usesFavor(row) {
		return Character{}, fmt.Errorf("only a warlock gains Favor")
	}
	if row.Hope < 1 {
		return Character{}, fmt.Errorf("there's no Hope to trade for Favor")
	}
	if row.Favor >= FavorMax {
		return Character{}, fmt.Errorf("you already hold the most Favor you can, %d", FavorMax)
	}
	var updated db.Character
	err = s.tx(func(q *db.Queries) error {
		if _, err := q.SetCharacterVitals(s.ctx, db.SetCharacterVitalsParams{
			HpMarked:     row.HpMarked,
			StressMarked: row.StressMarked,
			Hope:         row.Hope - 1,
			ArmorMarked:  row.ArmorMarked,
			ID:           characterID,
		}); err != nil {
			return fmt.Errorf("trading Hope: %w", err)
		}
		updated, err = q.SetCharacterTracks(s.ctx, db.SetCharacterTracksParams{
			Favor: row.Favor + 1,
			Focus: row.Focus,
			ID:    characterID,
		})
		if err != nil {
			return fmt.Errorf("gaining Favor: %w", err)
		}
		return nil
	})
	if err != nil {
		return Character{}, err
	}
	return s.characterView(updated), nil
}

func (s *Service) Refocus(characterID int64) (FocusRoll, error) {
	row, err := s.character(characterID)
	if err != nil {
		return FocusRoll{}, err
	}
	if !usesFocus(row) {
		return FocusRoll{}, fmt.Errorf("only a Martial Artist has Focus")
	}
	rolls := []int{}
	best := 0
	for range max(int(row.Instinct), 0) {
		r := dice.RollDice(dice.D6)
		rolls = append(rolls, r)
		best = max(best, r)
	}
	updated, err := s.q.SetCharacterTracks(s.ctx, db.SetCharacterTracksParams{
		Favor: row.Favor,
		Focus: int64(best),
		ID:    characterID,
	})
	if err != nil {
		return FocusRoll{}, fmt.Errorf("refocusing: %w", err)
	}
	return FocusRoll{Rolls: rolls, Focus: best, Character: s.characterView(updated)}, nil
}

func (s *Service) stanceAllowance(row db.Character) int {
	switch {
	case row.SubclassSlug == martialArtistSubclass:
		return 2 + int(row.Level) - 1
	case row.MulticlassSubclassSlug == martialArtistSubclass:
		from := int(row.Level)
		if records, err := s.ListLevelUps(row.ID); err == nil {
			for _, rec := range records {
				if slices.Contains(rec.Choices, "multiclass") {
					from = rec.Level
				}
			}
		}
		return 2 + int(row.Level) - from
	}
	return 0
}

func findStance(sheet *cards.Stances, title string) (cards.Stance, bool) {
	if sheet == nil {
		return cards.Stance{}, false
	}
	for _, st := range sheet.Stances {
		if strings.EqualFold(st.Title, strings.TrimSpace(title)) {
			return st, true
		}
	}
	return cards.Stance{}, false
}

func (s *Service) writeStances(q *db.Queries, characterID int64, known []string, active string) (db.Character, error) {
	if known == nil {
		known = []string{}
	}
	raw, err := json.Marshal(known)
	if err != nil {
		return db.Character{}, fmt.Errorf("encoding stances: %w", err)
	}
	row, err := q.SetCharacterStances(s.ctx, db.SetCharacterStancesParams{
		KnownStances: string(raw),
		ActiveStance: active,
		ID:           characterID,
	})
	if err != nil {
		return db.Character{}, fmt.Errorf("saving stances: %w", err)
	}
	return row, nil
}

func (s *Service) LearnStance(characterID int64, title string, known bool) (Character, error) {
	row, err := s.character(characterID)
	if err != nil {
		return Character{}, err
	}
	stance, ok := findStance(s.stanceSheet(row), title)
	if !ok {
		return Character{}, notFound("stance", title)
	}
	list := decodeTraitList(row.KnownStances)
	active := row.ActiveStance
	if known {
		if slices.Contains(list, stance.Title) {
			return s.characterView(row), nil
		}
		if tier := tierForLevel(int(row.Level)); stance.Tier > tier {
			return Character{}, fmt.Errorf("%s is a tier %d stance and you are tier %d", stance.Title, stance.Tier, tier)
		}
		if allowance := s.stanceAllowance(row); len(list) >= allowance {
			return Character{}, fmt.Errorf("you know %d stances, which is all you can at level %d", len(list), row.Level)
		}
		list = append(list, stance.Title)
	} else {
		list = slices.DeleteFunc(list, func(t string) bool { return t == stance.Title })
		if active == stance.Title {
			active = ""
		}
	}
	updated, err := s.writeStances(s.q, characterID, list, active)
	if err != nil {
		return Character{}, err
	}
	return s.characterView(updated), nil
}

func (s *Service) ShiftStance(characterID int64, title, payWith string) (Character, error) {
	row, err := s.character(characterID)
	if err != nil {
		return Character{}, err
	}
	stance, ok := findStance(s.stanceSheet(row), title)
	if !ok {
		return Character{}, notFound("stance", title)
	}
	if !slices.Contains(decodeTraitList(row.KnownStances), stance.Title) {
		return Character{}, fmt.Errorf("you can only shift into a stance you know")
	}
	if row.ActiveStance == stance.Title {
		return s.characterView(row), nil
	}

	focus, stress := row.Focus, row.StressMarked
	switch payWith {
	case "focus":
		if focus < 1 {
			return Character{}, fmt.Errorf("you have no Focus to spend")
		}
		focus--
	case "stress":
		if !hasFlowState(row) {
			return Character{}, fmt.Errorf("marking a Stress to shift takes Flow State")
		}
		if stress >= row.StressMax {
			return Character{}, fmt.Errorf("no Stress left to mark")
		}
		stress++
	default:
		return Character{}, fmt.Errorf("shifting into a stance costs a Focus, or a Stress with Flow State")
	}

	var updated db.Character
	err = s.tx(func(q *db.Queries) error {
		if _, err := q.SetCharacterTracks(s.ctx, db.SetCharacterTracksParams{Favor: row.Favor, Focus: focus, ID: characterID}); err != nil {
			return fmt.Errorf("spending Focus: %w", err)
		}
		if _, err := q.SetCharacterVitals(s.ctx, db.SetCharacterVitalsParams{
			HpMarked:     row.HpMarked,
			StressMarked: stress,
			Hope:         row.Hope,
			ArmorMarked:  row.ArmorMarked,
			ID:           characterID,
		}); err != nil {
			return fmt.Errorf("marking stress: %w", err)
		}
		updated, err = s.writeStances(q, characterID, decodeTraitList(row.KnownStances), stance.Title)
		return err
	})
	if err != nil {
		return Character{}, err
	}
	return s.characterView(updated), nil
}

func (s *Service) DropStance(characterID int64) (Character, error) {
	row, err := s.character(characterID)
	if err != nil {
		return Character{}, err
	}
	updated, err := s.writeStances(s.q, characterID, decodeTraitList(row.KnownStances), "")
	if err != nil {
		return Character{}, err
	}
	return s.characterView(updated), nil
}

func (s *Service) RollCombo(characterID int64, payWith string) (ComboResult, error) {
	row, err := s.character(characterID)
	if err != nil {
		return ComboResult{}, err
	}
	if row.ClassSlug != brawlerClass {
		return ComboResult{}, fmt.Errorf("only a brawler makes combo strikes")
	}
	die, err := dice.ParseSides(int(row.ComboDie))
	if err != nil {
		return ComboResult{}, err
	}

	focus, stress := row.Focus, row.StressMarked
	paid := ""
	switch payWith {
	case "focus":
		if !hasFlowState(row) {
			return ComboResult{}, fmt.Errorf("spending a Focus on a combo strike takes Flow State")
		}
		if focus < 1 {
			return ComboResult{}, fmt.Errorf("you have no Focus to spend")
		}
		focus--
		paid = "spent a Focus"
	default:
		if stress >= row.StressMax {
			return ComboResult{}, fmt.Errorf("no Stress left to mark")
		}
		stress++
		paid = "marked a Stress"
	}

	rolls := []int{dice.RollDice(die)}
	total := rolls[0]
	for len(rolls) < maxComboRolls {
		next := dice.RollDice(die)
		rolls = append(rolls, next)
		total += next
		if next < rolls[len(rolls)-2] {
			break
		}
	}

	var updated db.Character
	err = s.tx(func(q *db.Queries) error {
		if _, err := q.SetCharacterTracks(s.ctx, db.SetCharacterTracksParams{Favor: row.Favor, Focus: focus, ID: characterID}); err != nil {
			return fmt.Errorf("spending Focus: %w", err)
		}
		updated, err = q.SetCharacterVitals(s.ctx, db.SetCharacterVitalsParams{
			HpMarked:     row.HpMarked,
			StressMarked: stress,
			Hope:         row.Hope,
			ArmorMarked:  row.ArmorMarked,
			ID:           characterID,
		})
		if err != nil {
			return fmt.Errorf("marking stress: %w", err)
		}
		return nil
	})
	if err != nil {
		return ComboResult{}, err
	}
	return ComboResult{Die: int(die), Rolls: rolls, Total: total, Paid: paid, Character: s.characterView(updated)}, nil
}

func (s *Service) RollBrawlersStrike(characterID int64, critical bool) (DamageResult, error) {
	row, err := s.character(characterID)
	if err != nil {
		return DamageResult{}, err
	}
	if row.ClassSlug != brawlerClass {
		return DamageResult{}, fmt.Errorf("only a brawler has Brawler's Strike")
	}
	items, err := s.q.ListInventoryItems(s.ctx, characterID)
	if err != nil {
		return DamageResult{}, fmt.Errorf("listing inventory: %w", err)
	}
	for _, it := range items {
		if it.Equipped == 1 && (it.Kind == "primary-weapon" || it.Kind == "secondary-weapon") {
			return DamageResult{}, fmt.Errorf("Brawler's Strike is only active while no other weapon is equipped - %s is", it.Name)
		}
	}
	count := clamp(int(row.Proficiency), 1, MaxProficiency)
	modifier := 0
	if critical {
		modifier = count * (int(dice.D8) + int(dice.D6))
	}
	return DamageResult{
		Label:    fmt.Sprintf("Brawler's Strike %dd8+%dd6", count, count),
		Count:    count,
		Sides:    int(dice.D8) + int(dice.D6),
		Modifier: modifier,
		Total:    dice.RollDamage(count, dice.D8, 0) + dice.RollDamage(count, dice.D6, modifier),
		Critical: critical,
	}, nil
}
