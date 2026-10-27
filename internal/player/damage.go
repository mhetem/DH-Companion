package player

import (
	"fmt"

	"github.com/mhetem/DH-Companion/internal/db"
)

var severities = []string{"none", "minor", "major", "severe"}

type DamageInput struct {
	CharacterID int64 `json:"characterId"`
	Amount      int   `json:"amount"`
	UseArmor    bool  `json:"useArmor"`
}

type DamageTaken struct {
	Amount    int       `json:"amount"`
	Severity  string    `json:"severity"`
	Reduced   string    `json:"reduced"`
	HPMarked  int       `json:"hpMarked"`
	ArmorUsed bool      `json:"armorUsed"`
	Downed    bool      `json:"downed"`
	Ended     []string  `json:"ended"`
	Character Character `json:"character"`
}

func damageSeverity(amount, major, severe int) int {
	switch {
	case amount <= 0:
		return 0
	case amount >= severe:
		return 3
	case amount >= major:
		return 2
	default:
		return 1
	}
}

func (s *Service) TakeDamage(in DamageInput) (DamageTaken, error) {
	row, err := s.character(in.CharacterID)
	if err != nil {
		return DamageTaken{}, err
	}
	if in.Amount < 0 {
		return DamageTaken{}, fmt.Errorf("damage can't be negative, got %d", in.Amount)
	}
	if int(row.ThresholdMajor) <= 0 || int(row.ThresholdSevere) <= 0 {
		return DamageTaken{}, fmt.Errorf("your damage thresholds aren't set — equip armor in Inventory or set them on the character")
	}
	held, err := s.q.ListCharacterDomainCards(s.ctx, in.CharacterID)
	if err != nil {
		return DamageTaken{}, fmt.Errorf("listing domain cards: %w", err)
	}
	live, _ := s.liveBonuses(row, held)
	major := max(int(row.ThresholdMajor)+live.major, 1)
	severe := max(int(row.ThresholdSevere)+live.severe, major)

	dealt := damageSeverity(in.Amount, major, severe)
	marks := dealt
	armorMarked := int(row.ArmorMarked)
	used := false
	if in.UseArmor && marks > 0 {
		if source, blocked := armorBlocked(row); blocked {
			return DamageTaken{}, fmt.Errorf("%s is active, so you can't use Armor Slots", source)
		}
		if armorMarked >= int(row.ArmorScore) {
			return DamageTaken{}, fmt.Errorf("you have no Armor Slots left to mark")
		}
		armorMarked++
		marks--
		used = true
	}

	hpMax := int(row.HpMax)
	hp := min(int(row.HpMarked)+marks, hpMax)
	downed := marks > 0 && hp >= hpMax

	kept, ended := endedEffects(decodeTraitList(row.ActiveEffects), dealt == 3, in.Amount > 0, false)
	stance := row.ActiveStance
	if stance != "" && (dealt == 3 || downed) {
		ended = append(ended, stance+" stance")
		stance = ""
	}

	var updated db.Character
	err = s.tx(func(q *db.Queries) error {
		if _, err := q.SetCharacterVitals(s.ctx, db.SetCharacterVitalsParams{
			HpMarked:     int64(hp),
			StressMarked: row.StressMarked,
			Hope:         row.Hope,
			ArmorMarked:  int64(armorMarked),
			ID:           in.CharacterID,
		}); err != nil {
			return fmt.Errorf("marking damage: %w", err)
		}
		if _, err := s.writeEffects(q, in.CharacterID, kept); err != nil {
			return err
		}
		updated, err = s.writeStances(q, in.CharacterID, decodeTraitList(row.KnownStances), stance)
		return err
	})
	if err != nil {
		return DamageTaken{}, err
	}
	if ended == nil {
		ended = []string{}
	}
	return DamageTaken{
		Amount:    in.Amount,
		Severity:  severities[dealt],
		Reduced:   severities[marks],
		HPMarked:  hp - int(row.HpMarked),
		ArmorUsed: used,
		Downed:    downed,
		Ended:     ended,
		Character: s.characterView(updated),
	}, nil
}
