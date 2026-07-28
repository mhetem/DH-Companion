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
