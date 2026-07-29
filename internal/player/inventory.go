package player

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/mhetem/DH-Companion/internal/db"
)

type Item struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Qty      int    `json:"qty"`
	Equipped bool   `json:"equipped"`
	Detail   string `json:"detail"`
}

type ItemInput struct {
	ID          *int64 `json:"id"`
	CharacterID int64  `json:"characterId"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Qty         int    `json:"qty"`
	Equipped    bool   `json:"equipped"`
	Detail      string `json:"detail"`
}

func itemView(r db.InventoryItem) Item {
	return Item{
		ID:       r.ID,
		Name:     r.Name,
		Kind:     r.Kind,
		Qty:      int(r.Qty),
		Equipped: r.Equipped == 1,
		Detail:   r.Detail,
	}
}

func (s *Service) ListInventory(characterID int64) ([]Item, error) {
	rows, err := s.q.ListInventoryItems(s.ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("listing inventory: %w", err)
	}
	out := make([]Item, 0, len(rows))
	for _, r := range rows {
		out = append(out, itemView(r))
	}
	return out, nil
}

func (s *Service) SaveItem(in ItemInput) ([]Item, error) {
	name, err := validateName(in.Name)
	if err != nil {
		return nil, err
	}
	kind, err := validateItemKind(in.Kind)
	if err != nil {
		return nil, err
	}
	qty := clamp(in.Qty, 0, 999)
	if in.ID == nil && in.Qty == 0 {
		qty = 1
	}
	equipped := int64(0)
	if in.Equipped {
		equipped = 1
	}

	characterID := in.CharacterID
	if in.ID == nil {
		if _, err := s.character(characterID); err != nil {
			return nil, err
		}
		if equipped == 1 && slices.Contains(equipExclusiveKinds, kind) {
			if err := s.unequipKind(characterID, kind); err != nil {
				return nil, err
			}
		}
		if _, err := s.q.CreateInventoryItem(s.ctx, db.CreateInventoryItemParams{
			CharacterID: characterID,
			Name:        name,
			Kind:        kind,
			Qty:         int64(qty),
			Equipped:    equipped,
			Detail:      strings.TrimSpace(in.Detail),
		}); err != nil {
			return nil, fmt.Errorf("creating item: %w", err)
		}
		return s.ListInventory(characterID)
	}

	existing, err := s.q.GetInventoryItem(s.ctx, *in.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("item", fmt.Sprint(*in.ID))
	}
	if err != nil {
		return nil, fmt.Errorf("loading item: %w", err)
	}
	characterID = existing.CharacterID

	if equipped == 1 && slices.Contains(equipExclusiveKinds, kind) {
		if err := s.unequipKind(characterID, kind); err != nil {
			return nil, err
		}
	}
	if _, err := s.q.UpdateInventoryItem(s.ctx, db.UpdateInventoryItemParams{
		Name:     name,
		Kind:     kind,
		Qty:      int64(qty),
		Equipped: equipped,
		Detail:   strings.TrimSpace(in.Detail),
		ID:       *in.ID,
	}); err != nil {
		return nil, fmt.Errorf("updating item: %w", err)
	}
	return s.ListInventory(characterID)
}

func (s *Service) AdjustItemQty(id int64, delta int) ([]Item, error) {
	row, err := s.q.AdjustInventoryQty(s.ctx, db.AdjustInventoryQtyParams{Delta: int64(delta), ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("item", fmt.Sprint(id))
	}
	if err != nil {
		return nil, fmt.Errorf("adjusting quantity: %w", err)
	}
	return s.ListInventory(row.CharacterID)
}

func (s *Service) SetItemEquipped(id int64, equipped bool) ([]Item, error) {
	existing, err := s.q.GetInventoryItem(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("item", fmt.Sprint(id))
	}
	if err != nil {
		return nil, fmt.Errorf("loading item: %w", err)
	}
	if equipped && slices.Contains(equipExclusiveKinds, existing.Kind) {
		if err := s.unequipKind(existing.CharacterID, existing.Kind); err != nil {
			return nil, err
		}
	}
	flag := int64(0)
	if equipped {
		flag = 1
	}
	if _, err := s.q.SetInventoryEquipped(s.ctx, db.SetInventoryEquippedParams{Equipped: flag, ID: id}); err != nil {
		return nil, fmt.Errorf("equipping item: %w", err)
	}
	return s.ListInventory(existing.CharacterID)
}

func (s *Service) DeleteItem(id int64) ([]Item, error) {
	existing, err := s.q.GetInventoryItem(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("item", fmt.Sprint(id))
	}
	if err != nil {
		return nil, fmt.Errorf("loading item: %w", err)
	}
	if err := s.q.DeleteInventoryItem(s.ctx, id); err != nil {
		return nil, fmt.Errorf("deleting item: %w", err)
	}
	return s.ListInventory(existing.CharacterID)
}

func (s *Service) AddClassItems(characterID int64) ([]Item, error) {
	row, err := s.character(characterID)
	if err != nil {
		return nil, err
	}
	class, ok := s.catalog.Class(row.ClassSlug)
	if !ok {
		return nil, notFound("class", row.ClassSlug)
	}
	for _, name := range class.ClassItems {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, err := s.q.CreateInventoryItem(s.ctx, db.CreateInventoryItemParams{
			CharacterID: characterID,
			Name:        name,
			Kind:        "item",
			Qty:         1,
			Equipped:    0,
			Detail:      class.Name + " class item",
		}); err != nil {
			return nil, fmt.Errorf("adding class items: %w", err)
		}
	}
	return s.ListInventory(characterID)
}

func (s *Service) unequipKind(characterID int64, kind string) error {
	if err := s.q.UnequipInventoryKind(s.ctx, db.UnequipInventoryKindParams{
		CharacterID: characterID,
		Kind:        kind,
	}); err != nil {
		return fmt.Errorf("unequipping %s: %w", kind, err)
	}
	return nil
}
