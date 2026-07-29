package player

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

type DomainCard struct {
	ID         int64            `json:"id"`
	CardSlug   string           `json:"cardSlug"`
	Location   string           `json:"location"`
	Unresolved bool             `json:"unresolved"`
	Card       cards.DomainCard `json:"card"`
}

type Loadout struct {
	Loadout    []DomainCard `json:"loadout"`
	Vault      []DomainCard `json:"vault"`
	LoadoutMax int          `json:"loadoutMax"`
	Held       int          `json:"held"`
	Allowance  int          `json:"allowance"`
}

type SwapInput struct {
	CharacterID int64  `json:"characterId"`
	Recall      string `json:"recall"`
	Vault       string `json:"vault"`
	Resting     bool   `json:"resting"`
}

type SwapResult struct {
	Character    Character `json:"character"`
	Loadout      Loadout   `json:"loadout"`
	RecallCost   int       `json:"recallCost"`
	StressMarked int       `json:"stressMarked"`
	Outcome      string    `json:"outcome"`
}

func (s *Service) domainCardView(r db.CharacterDomainCard) DomainCard {
	out := DomainCard{ID: r.ID, CardSlug: r.CardSlug, Location: r.Location}
	card, ok := s.catalog.DomainCard(r.CardSlug)
	if !ok {
		out.Unresolved = true
		out.Card = cards.DomainCard{Meta: cards.Meta{
			Kind: cards.KindDomain,
			Slug: r.CardSlug,
			Name: r.CardSlug,
			Type: "Unknown",
		}}
		return out
	}
	out.Card = card
	return out
}

func (s *Service) domainCards(characterID int64) ([]DomainCard, []DomainCard, error) {
	rows, err := s.q.ListCharacterDomainCards(s.ctx, characterID)
	if err != nil {
		return nil, nil, fmt.Errorf("listing domain cards: %w", err)
	}
	loadout, vault := []DomainCard{}, []DomainCard{}
	for _, r := range rows {
		card := s.domainCardView(r)
		if card.Location == "vault" {
			vault = append(vault, card)
			continue
		}
		loadout = append(loadout, card)
	}
	return loadout, vault, nil
}

func (s *Service) GetLoadout(characterID int64) (Loadout, error) {
	if _, err := s.character(characterID); err != nil {
		return Loadout{}, err
	}
	loadout, vault, err := s.domainCards(characterID)
	if err != nil {
		return Loadout{}, err
	}
	allowance, err := s.domainCardAllowance(characterID)
	if err != nil {
		return Loadout{}, err
	}
	return Loadout{
		Loadout:    loadout,
		Vault:      vault,
		LoadoutMax: LoadoutMax,
		Held:       len(loadout) + len(vault),
		Allowance:  allowance,
	}, nil
}

func (s *Service) domainCardAllowance(characterID int64) (int, error) {
	rows, err := s.q.ListLevelUps(s.ctx, characterID)
	if err != nil {
		return 0, fmt.Errorf("loading level history: %w", err)
	}
	allowance := StartingDomainCards
	for _, r := range rows {
		allowance += s.catalog.Leveling.DomainCardsPerLevel
		rules, ok := s.catalog.Tier(int(r.Tier))
		if !ok {
			continue
		}
		var choices []AdvancementChoice
		if err := json.Unmarshal([]byte(r.Choices), &choices); err != nil {
			continue
		}
		for _, c := range choices {
			if adv, found := findAdvancement(rules, c.ID); found {
				allowance += adv.Effect.DomainCards
			}
		}
	}
	return allowance, nil
}

func (s *Service) assertCardAllowance(characterID int64) error {
	current, err := s.GetLoadout(characterID)
	if err != nil {
		return err
	}
	if current.Held >= current.Allowance {
		return fmt.Errorf(
			"you hold %d of %d domain cards — levelling up is what earns another",
			current.Held, current.Allowance,
		)
	}
	return nil
}

func (s *Service) AddDomainCard(characterID int64, slug string, location string) (Loadout, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return Loadout{}, fmt.Errorf("a card is required")
	}
	if _, err := s.character(characterID); err != nil {
		return Loadout{}, err
	}
	if _, ok := s.catalog.DomainCard(slug); !ok {
		return Loadout{}, notFound("domain card", slug)
	}
	loc, err := validateLocation(location)
	if err != nil {
		return Loadout{}, err
	}
	if err := s.assertCardAllowance(characterID); err != nil {
		return Loadout{}, err
	}
	if loc == "loadout" {
		if err := s.assertLoadoutRoom(characterID); err != nil {
			return Loadout{}, err
		}
	}

	_, err = s.q.AddCharacterDomainCard(s.ctx, db.AddCharacterDomainCardParams{
		CharacterID: characterID,
		CardSlug:    slug,
		Location:    loc,
	})
	if isUniqueViolation(err) {
		return Loadout{}, fmt.Errorf("this character already has %q", slug)
	}
	if err != nil {
		return Loadout{}, fmt.Errorf("adding domain card: %w", err)
	}
	return s.GetLoadout(characterID)
}

func (s *Service) MoveDomainCard(characterID int64, slug string, location string) (Loadout, error) {
	slug = strings.TrimSpace(slug)
	loc, err := validateLocation(location)
	if err != nil {
		return Loadout{}, err
	}

	current, err := s.q.GetCharacterDomainCard(s.ctx, db.GetCharacterDomainCardParams{
		CharacterID: characterID,
		CardSlug:    slug,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Loadout{}, notFound("domain card", slug)
	}
	if err != nil {
		return Loadout{}, fmt.Errorf("loading domain card: %w", err)
	}
	if current.Location == loc {
		return s.GetLoadout(characterID)
	}
	if loc == "loadout" {
		if err := s.assertLoadoutRoom(characterID); err != nil {
			return Loadout{}, err
		}
	}

	if _, err := s.q.SetDomainCardLocation(s.ctx, db.SetDomainCardLocationParams{
		Location:    loc,
		CharacterID: characterID,
		CardSlug:    slug,
	}); err != nil {
		return Loadout{}, fmt.Errorf("moving domain card: %w", err)
	}
	return s.GetLoadout(characterID)
}

func (s *Service) RemoveDomainCard(characterID int64, slug string) (Loadout, error) {
	if err := s.q.RemoveCharacterDomainCard(s.ctx, db.RemoveCharacterDomainCardParams{
		CharacterID: characterID,
		CardSlug:    strings.TrimSpace(slug),
	}); err != nil {
		return Loadout{}, fmt.Errorf("removing domain card: %w", err)
	}
	return s.GetLoadout(characterID)
}

func (s *Service) SwapDomainCard(in SwapInput) (SwapResult, error) {
	row, err := s.character(in.CharacterID)
	if err != nil {
		return SwapResult{}, err
	}

	inSlug := strings.TrimSpace(in.Recall)
	outSlug := strings.TrimSpace(in.Vault)
	if inSlug == "" {
		return SwapResult{}, fmt.Errorf("choose a vault card to bring in")
	}
	if inSlug == outSlug {
		return SwapResult{}, fmt.Errorf("that is the same card")
	}

	incoming, err := s.q.GetCharacterDomainCard(s.ctx, db.GetCharacterDomainCardParams{
		CharacterID: in.CharacterID,
		CardSlug:    inSlug,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return SwapResult{}, notFound("domain card", inSlug)
	}
	if err != nil {
		return SwapResult{}, fmt.Errorf("loading domain card: %w", err)
	}
	if incoming.Location != "vault" {
		return SwapResult{}, fmt.Errorf("%s is already in your loadout", s.domainCardName(inSlug))
	}

	if outSlug != "" {
		outgoing, err := s.q.GetCharacterDomainCard(s.ctx, db.GetCharacterDomainCardParams{
			CharacterID: in.CharacterID,
			CardSlug:    outSlug,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return SwapResult{}, notFound("domain card", outSlug)
		}
		if err != nil {
			return SwapResult{}, fmt.Errorf("loading domain card: %w", err)
		}
		if outgoing.Location != "loadout" {
			return SwapResult{}, fmt.Errorf("%s is not in your loadout", s.domainCardName(outSlug))
		}
	} else {
		held, err := s.q.CountCharacterLoadout(s.ctx, in.CharacterID)
		if err != nil {
			return SwapResult{}, fmt.Errorf("counting loadout: %w", err)
		}
		if held >= LoadoutMax {
			return SwapResult{}, fmt.Errorf("your loadout already holds %d cards — choose one to vault", LoadoutMax)
		}
	}

	recall := s.recallCost(inSlug)
	cost := recall
	if in.Resting {
		cost = 0
	}
	if room := int(row.StressMax) - int(row.StressMarked); cost > room {
		return SwapResult{}, fmt.Errorf(
			"recalling %s costs %d Stress and you have %d left — swap it on a rest instead",
			s.domainCardName(inSlug), cost, room,
		)
	}

	err = s.tx(func(q *db.Queries) error {
		if outSlug != "" {
			if _, err := q.SetDomainCardLocation(s.ctx, db.SetDomainCardLocationParams{
				Location:    "vault",
				CharacterID: in.CharacterID,
				CardSlug:    outSlug,
			}); err != nil {
				return fmt.Errorf("vaulting domain card: %w", err)
			}
		}
		if _, err := q.SetDomainCardLocation(s.ctx, db.SetDomainCardLocationParams{
			Location:    "loadout",
			CharacterID: in.CharacterID,
			CardSlug:    inSlug,
		}); err != nil {
			return fmt.Errorf("recalling domain card: %w", err)
		}
		if cost > 0 {
			if _, err := q.AdjustCharacterStress(s.ctx, db.AdjustCharacterStressParams{
				Delta: int64(cost),
				ID:    in.CharacterID,
			}); err != nil {
				return fmt.Errorf("marking stress: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return SwapResult{}, err
	}

	character, err := s.GetCharacter(in.CharacterID)
	if err != nil {
		return SwapResult{}, err
	}
	loadout, err := s.GetLoadout(in.CharacterID)
	if err != nil {
		return SwapResult{}, err
	}

	outcome := "Recalled " + s.domainCardName(inSlug)
	if outSlug != "" {
		outcome += ", vaulting " + s.domainCardName(outSlug)
	}
	switch {
	case in.Resting && recall > 0:
		outcome += fmt.Sprintf(" — free on a rest (Recall Cost %d)", recall)
	case cost > 0:
		outcome += fmt.Sprintf(" — marked %d Stress", cost)
	default:
		outcome += " — no Recall Cost"
	}

	return SwapResult{
		Character:    character,
		Loadout:      loadout,
		RecallCost:   recall,
		StressMarked: cost,
		Outcome:      outcome,
	}, nil
}

func (s *Service) domainCardName(slug string) string {
	if card, ok := s.catalog.DomainCard(slug); ok {
		return card.Name
	}
	return slug
}

func (s *Service) recallCost(slug string) int {
	card, ok := s.catalog.DomainCard(slug)
	if !ok {
		return 0
	}
	return clamp(leadingInt(card.RecallCost, 0), 0, 20)
}

func (s *Service) assertLoadoutRoom(characterID int64) error {
	n, err := s.q.CountCharacterLoadout(s.ctx, characterID)
	if err != nil {
		return fmt.Errorf("counting loadout: %w", err)
	}
	if n >= LoadoutMax {
		return fmt.Errorf("your loadout already holds %d cards — vault one first", LoadoutMax)
	}
	return nil
}
