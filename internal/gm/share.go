package gm

import (
	"errors"
	"fmt"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/share"
)

const maxImportRenames = 50

type SharePreview struct {
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Tier        string `json:"tier"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Renamed     bool   `json:"renamed"`
}

func (s *Service) ShareAdversary(slug string) (string, error) {
	card, err := s.GetCustomAdversary(slug)
	if err != nil {
		return "", err
	}
	return share.Encode(share.KindAdversary, card)
}

func (s *Service) ShareEnvironment(slug string) (string, error) {
	card, err := s.GetCustomEnvironment(slug)
	if err != nil {
		return "", err
	}
	return share.Encode(share.KindEnvironment, card)
}

func (s *Service) ShareWeapon(slug string) (string, error) {
	card, err := s.GetCustomWeapon(slug)
	if err != nil {
		return "", err
	}
	return share.Encode(share.KindWeapon, card)
}

func (s *Service) ShareArmor(slug string) (string, error) {
	card, err := s.GetCustomArmor(slug)
	if err != nil {
		return "", err
	}
	return share.Encode(share.KindArmor, card)
}

func (s *Service) ShareItem(slug string) (string, error) {
	card, err := s.GetCustomItem(slug)
	if err != nil {
		return "", err
	}
	return share.Encode(share.KindItem, card)
}

func (s *Service) ShareConsumable(slug string) (string, error) {
	card, err := s.GetCustomConsumable(slug)
	if err != nil {
		return "", err
	}
	return share.Encode(share.KindConsumable, card)
}

func (s *Service) PreviewShareCode(code string) (SharePreview, error) {
	payload, err := share.Decode(code)
	if err != nil {
		return SharePreview{}, err
	}

	switch payload.Kind {
	case share.KindAdversary:
		var card cards.Adversary
		if err := payload.Into(&card); err != nil {
			return SharePreview{}, err
		}
		return previewOf(payload.Kind, card.Meta), nil
	case share.KindEnvironment:
		var card cards.Environment
		if err := payload.Into(&card); err != nil {
			return SharePreview{}, err
		}
		return previewOf(payload.Kind, card.Meta), nil
	case share.KindWeapon:
		var card cards.Weapon
		if err := payload.Into(&card); err != nil {
			return SharePreview{}, err
		}
		return previewOf(payload.Kind, card.Meta), nil
	case share.KindArmor:
		var card cards.Armor
		if err := payload.Into(&card); err != nil {
			return SharePreview{}, err
		}
		return previewOf(payload.Kind, card.Meta), nil
	case share.KindItem, share.KindConsumable:
		var card cards.Loot
		if err := payload.Into(&card); err != nil {
			return SharePreview{}, err
		}
		return previewOf(payload.Kind, card.Meta), nil
	}
	return SharePreview{}, fmt.Errorf("unknown share kind %q", payload.Kind)
}

func (s *Service) ImportShareCode(code string) (SharePreview, error) {
	payload, err := share.Decode(code)
	if err != nil {
		return SharePreview{}, err
	}

	switch payload.Kind {
	case share.KindAdversary:
		var card cards.Adversary
		if err := payload.Into(&card); err != nil {
			return SharePreview{}, err
		}
		name, renamed, err := s.freeName(card.Name, s.adversarySlugTaken)
		if err != nil {
			return SharePreview{}, err
		}
		card.Name, card.Slug = name, ""
		saved, err := s.CreateCustomAdversary(card)
		if err != nil {
			return SharePreview{}, err
		}
		out := previewOf(payload.Kind, saved.Meta)
		out.Renamed = renamed
		return out, nil

	case share.KindEnvironment:
		var card cards.Environment
		if err := payload.Into(&card); err != nil {
			return SharePreview{}, err
		}
		name, renamed, err := s.freeName(card.Name, s.environmentSlugTaken)
		if err != nil {
			return SharePreview{}, err
		}
		card.Name, card.Slug = name, ""
		saved, err := s.CreateCustomEnvironment(card)
		if err != nil {
			return SharePreview{}, err
		}
		out := previewOf(payload.Kind, saved.Meta)
		out.Renamed = renamed
		return out, nil

	case share.KindWeapon:
		var card cards.Weapon
		if err := payload.Into(&card); err != nil {
			return SharePreview{}, err
		}
		name, renamed, err := s.freeName(card.Name, s.slugTaken(func(slug string) error {
			_, err := s.GetCustomWeapon(slug)
			return err
		}))
		if err != nil {
			return SharePreview{}, err
		}
		card.Name, card.Slug = name, ""
		saved, err := s.CreateCustomWeapon(card)
		if err != nil {
			return SharePreview{}, err
		}
		out := previewOf(payload.Kind, saved.Meta)
		out.Renamed = renamed
		return out, nil

	case share.KindArmor:
		var card cards.Armor
		if err := payload.Into(&card); err != nil {
			return SharePreview{}, err
		}
		name, renamed, err := s.freeName(card.Name, s.slugTaken(func(slug string) error {
			_, err := s.GetCustomArmor(slug)
			return err
		}))
		if err != nil {
			return SharePreview{}, err
		}
		card.Name, card.Slug = name, ""
		saved, err := s.CreateCustomArmor(card)
		if err != nil {
			return SharePreview{}, err
		}
		out := previewOf(payload.Kind, saved.Meta)
		out.Renamed = renamed
		return out, nil

	case share.KindItem, share.KindConsumable:
		var card cards.Loot
		if err := payload.Into(&card); err != nil {
			return SharePreview{}, err
		}
		get, create := s.GetCustomItem, s.CreateCustomItem
		if payload.Kind == share.KindConsumable {
			get, create = s.GetCustomConsumable, s.CreateCustomConsumable
		}
		name, renamed, err := s.freeName(card.Name, s.slugTaken(func(slug string) error {
			_, err := get(slug)
			return err
		}))
		if err != nil {
			return SharePreview{}, err
		}
		card.Name, card.Slug = name, ""
		saved, err := create(card)
		if err != nil {
			return SharePreview{}, err
		}
		out := previewOf(payload.Kind, saved.Meta)
		out.Renamed = renamed
		return out, nil
	}
	return SharePreview{}, fmt.Errorf("unknown share kind %q", payload.Kind)
}

func previewOf(kind string, m cards.Meta) SharePreview {
	return SharePreview{
		Kind:        kind,
		Name:        m.Name,
		Slug:        m.Slug,
		Tier:        m.Tier,
		Type:        m.Type,
		Description: m.Description,
	}
}

func (s *Service) adversarySlugTaken(slug string) (bool, error) {
	_, err := s.GetCustomAdversary(slug)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (s *Service) environmentSlugTaken(slug string) (bool, error) {
	_, err := s.GetCustomEnvironment(slug)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return false, err
}

// slugTaken adapts any "get by slug" lookup into the predicate freeName wants,
// so each new homebrew kind doesn't need its own copy of the ErrNotFound dance.
func (s *Service) slugTaken(get func(string) error) func(string) (bool, error) {
	return func(slug string) (bool, error) {
		err := get(slug)
		if err == nil {
			return true, nil
		}
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, err
	}
}

func (s *Service) freeName(name string, taken func(string) (bool, error)) (string, bool, error) {
	base, err := validateName(name)
	if err != nil {
		return "", false, err
	}
	if slugify(base) == "" {
		return "", false, fmt.Errorf("name %q does not produce a usable slug", base)
	}

	candidate := base
	for n := 2; n < maxImportRenames+2; n++ {
		busy, err := taken(slugify(candidate))
		if err != nil {
			return "", false, err
		}
		if !busy {
			return candidate, candidate != base, nil
		}
		candidate = fmt.Sprintf("%s (%d)", base, n)
	}
	return "", false, fmt.Errorf("you already have %d copies of %q", maxImportRenames, base)
}
