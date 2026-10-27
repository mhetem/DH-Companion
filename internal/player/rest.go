package player

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/dice"
)

const (
	ShortRestMoves = 2
	LongRestMoves  = 3
)

type RestMove struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type RestInput struct {
	CharacterID int64    `json:"characterId"`
	Long        bool     `json:"long"`
	Moves       []string `json:"moves"`
}

type RestResult struct {
	Character Character `json:"character"`
	Long      bool      `json:"long"`
	Allowed   int       `json:"allowed"`
	Outcomes  []string  `json:"outcomes"`
}

var shortRestMoves = []RestMove{
	{ID: "tend-wounds", Label: "Tend to Wounds", Description: "Clear 1d4 + your tier Hit Points."},
	{ID: "clear-stress", Label: "Clear Stress", Description: "Clear 1d4 + your tier Stress."},
	{ID: "repair-armor", Label: "Repair Armor", Description: "Clear 1d4 + your tier Armor Slots."},
	{ID: "prepare", Label: "Prepare", Description: "Gain a Hope. Preparing alongside the party is worth 2 Hope each."},
}

var longRestMoves = []RestMove{
	{ID: "tend-all-wounds", Label: "Tend to All Wounds", Description: "Clear all Hit Points."},
	{ID: "clear-all-stress", Label: "Clear All Stress", Description: "Clear all Stress."},
	{ID: "repair-all-armor", Label: "Repair All Armor", Description: "Clear all Armor Slots."},
	{ID: "prepare", Label: "Prepare", Description: "Gain a Hope. Preparing alongside the party is worth 2 Hope each."},
	{ID: "work-on-project", Label: "Work on a Project", Description: "Advance a long-term project. No change to the sheet."},
}

var tributeMove = RestMove{
	ID:          "tribute",
	Label:       "Show Tribute",
	Description: "Show tribute to your patron and describe how. Gain Favor equal to your Spellcast trait.",
}

func (s *Service) RestMoves(characterID int64, long bool) ([]RestMove, error) {
	row, err := s.character(characterID)
	if err != nil {
		return nil, err
	}
	src := shortRestMoves
	if long {
		src = longRestMoves
	}
	out := slices.Clone(src)
	if usesFavor(row) {
		out = append(out, tributeMove)
	}
	return out, nil
}

func (s *Service) RestAllowance(long bool) int {
	if long {
		return LongRestMoves
	}
	return ShortRestMoves
}

func (s *Service) Rest(in RestInput) (RestResult, error) {
	row, err := s.character(in.CharacterID)
	if err != nil {
		return RestResult{}, err
	}

	allowed := s.RestAllowance(in.Long)
	kind := "short"
	if in.Long {
		kind = "long"
	}
	if len(in.Moves) == 0 {
		return RestResult{}, fmt.Errorf("choose at least one downtime move")
	}
	if len(in.Moves) > allowed {
		return RestResult{}, fmt.Errorf("a %s rest is %d downtime moves — you picked %d", kind, allowed, len(in.Moves))
	}

	tier := tierForLevel(int(row.Level))
	hp := int(row.HpMarked)
	stress := int(row.StressMarked)
	hope := int(row.Hope)
	armor := int(row.ArmorMarked)
	favor := int(row.Favor)
	outcomes := []string{}

	for _, raw := range in.Moves {
		move, ok := findRestMove(in.Long, raw)
		if !ok {
			return RestResult{}, fmt.Errorf("%q is not a %s rest move", raw, kind)
		}
		switch move.ID {
		case "tend-wounds":
			rolled := dice.RollDice(dice.D4) + tier
			cleared := hp - clamp(hp-rolled, 0, int(row.HpMax))
			hp -= cleared
			outcomes = append(outcomes, fmt.Sprintf("%s — cleared %s (rolled %d)", move.Label, count(cleared, "Hit Point"), rolled))
		case "clear-stress":
			rolled := dice.RollDice(dice.D4) + tier
			cleared := stress - clamp(stress-rolled, 0, int(row.StressMax))
			stress -= cleared
			outcomes = append(outcomes, fmt.Sprintf("%s — cleared %d Stress (rolled %d)", move.Label, cleared, rolled))
		case "repair-armor":
			rolled := dice.RollDice(dice.D4) + tier
			cleared := armor - clamp(armor-rolled, 0, int(row.ArmorScore))
			armor -= cleared
			outcomes = append(outcomes, fmt.Sprintf("%s — cleared %s (rolled %d)", move.Label, count(cleared, "Armor Slot"), rolled))
		case "tend-all-wounds":
			outcomes = append(outcomes, fmt.Sprintf("%s — cleared %s", move.Label, count(hp, "Hit Point")))
			hp = 0
		case "clear-all-stress":
			outcomes = append(outcomes, fmt.Sprintf("%s — cleared %d Stress", move.Label, stress))
			stress = 0
		case "repair-all-armor":
			outcomes = append(outcomes, fmt.Sprintf("%s — cleared %s", move.Label, count(armor, "Armor Slot")))
			armor = 0
		case "prepare":
			before := hope
			hope = clamp(hope+1, 0, HopeMax)
			outcomes = append(outcomes, fmt.Sprintf("%s — gained %d Hope", move.Label, hope-before))
		case "work-on-project":
			outcomes = append(outcomes, move.Label+" — no change to the sheet")
		case "tribute":
			if !usesFavor(row) {
				return RestResult{}, fmt.Errorf("only a warlock has a patron to show tribute to")
			}
			before := favor
			favor = min(favor+s.favorTrait(row), FavorMax)
			outcomes = append(outcomes, fmt.Sprintf("%s — gained %d Favor", move.Label, favor-before))
		}
	}

	kept, ended := endedEffects(decodeTraitList(row.ActiveEffects), false, false, true)
	if row.ActiveStance != "" {
		ended = append(ended, row.ActiveStance+" stance")
	}
	if len(ended) > 0 {
		outcomes = append(outcomes, "Ended with the rest: "+strings.Join(ended, ", "))
	}

	err = s.tx(func(q *db.Queries) error {
		if _, err := q.SetCharacterVitals(s.ctx, db.SetCharacterVitalsParams{
			HpMarked:     int64(hp),
			StressMarked: int64(stress),
			Hope:         int64(hope),
			ArmorMarked:  int64(armor),
			ID:           in.CharacterID,
		}); err != nil {
			return fmt.Errorf("setting vitals: %w", err)
		}
		if _, err := q.SetCharacterTracks(s.ctx, db.SetCharacterTracksParams{
			Favor: int64(favor),
			Focus: row.Focus,
			ID:    in.CharacterID,
		}); err != nil {
			return fmt.Errorf("gaining Favor: %w", err)
		}
		if _, err := s.writeEffects(q, in.CharacterID, kept); err != nil {
			return err
		}
		_, err := s.writeStances(q, in.CharacterID, decodeTraitList(row.KnownStances), "")
		return err
	})
	if err != nil {
		return RestResult{}, err
	}
	character, err := s.GetCharacter(in.CharacterID)
	if err != nil {
		return RestResult{}, err
	}
	return RestResult{Character: character, Long: in.Long, Allowed: allowed, Outcomes: outcomes}, nil
}

func count(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func findRestMove(long bool, id string) (RestMove, bool) {
	src := shortRestMoves
	if long {
		src = longRestMoves
	}
	id = strings.TrimSpace(id)
	for _, m := range append(slices.Clone(src), tributeMove) {
		if m.ID == id {
			return m, true
		}
	}
	return RestMove{}, false
}
