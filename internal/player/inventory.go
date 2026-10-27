package player

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

type Item struct {
	ID       int64         `json:"id"`
	Name     string        `json:"name"`
	Kind     string        `json:"kind"`
	Qty      int           `json:"qty"`
	Equipped bool          `json:"equipped"`
	Detail   string        `json:"detail"`
	SrdSlug  string        `json:"srdSlug"`
	Weapon   *cards.Weapon `json:"weapon"`
	Armor    *cards.Armor  `json:"armor"`
	Loot     *cards.Loot   `json:"loot"`
}

type ItemInput struct {
	ID          *int64 `json:"id"`
	CharacterID int64  `json:"characterId"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Qty         int    `json:"qty"`
	Equipped    bool   `json:"equipped"`
	Detail      string `json:"detail"`
	SrdSlug     string `json:"srdSlug"`
}

func (s *Service) itemView(r db.InventoryItem) Item {
	item := Item{
		ID:       r.ID,
		Name:     r.Name,
		Kind:     r.Kind,
		Qty:      int(r.Qty),
		Equipped: r.Equipped == 1,
		Detail:   r.Detail,
		SrdSlug:  r.SrdSlug,
	}
	if r.SrdSlug == "" {
		return item
	}
	switch r.Kind {
	case "primary-weapon", "secondary-weapon":
		if w, ok := s.catalog.Weapon(r.SrdSlug); ok {
			item.Weapon = &w
		}
	case "armor":
		if a, ok := s.catalog.ArmorPiece(r.SrdSlug); ok {
			item.Armor = &a
		}
	case "item":
		if l, ok := s.catalog.Item(r.SrdSlug); ok {
			item.Loot = &l
		}
	case "consumable":
		if l, ok := s.catalog.Consumable(r.SrdSlug); ok {
			item.Loot = &l
		}
	}
	return item
}

func (s *Service) ListInventory(characterID int64) ([]Item, error) {
	rows, err := s.q.ListInventoryItems(s.ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("listing inventory: %w", err)
	}
	out := make([]Item, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.itemView(r))
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
	slug, err := s.validateItemSlug(kind, in.SrdSlug)
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
	detail := strings.TrimSpace(in.Detail)

	characterID := in.CharacterID
	if in.ID == nil {
		if _, err := s.character(characterID); err != nil {
			return nil, err
		}
	} else {
		existing, err := s.inventoryItem(*in.ID)
		if err != nil {
			return nil, err
		}
		characterID = existing.CharacterID
	}

	return s.changeEquipment(characterID, func(q *db.Queries) error {
		if equipped == 1 {
			if err := s.makeRoom(q, characterID, kind, slug); err != nil {
				return err
			}
		}
		if in.ID == nil {
			if _, err := q.CreateInventoryItem(s.ctx, db.CreateInventoryItemParams{
				CharacterID: characterID,
				Name:        name,
				Kind:        kind,
				Qty:         int64(qty),
				Equipped:    equipped,
				Detail:      detail,
				SrdSlug:     slug,
			}); err != nil {
				return fmt.Errorf("creating item: %w", err)
			}
			return nil
		}
		if _, err := q.UpdateInventoryItem(s.ctx, db.UpdateInventoryItemParams{
			Name:     name,
			Kind:     kind,
			Qty:      int64(qty),
			Equipped: equipped,
			Detail:   detail,
			SrdSlug:  slug,
			ID:       *in.ID,
		}); err != nil {
			return fmt.Errorf("updating item: %w", err)
		}
		return nil
	})
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
	existing, err := s.inventoryItem(id)
	if err != nil {
		return nil, err
	}
	return s.changeEquipment(existing.CharacterID, func(q *db.Queries) error {
		flag := int64(0)
		if equipped {
			flag = 1
			if err := s.makeRoom(q, existing.CharacterID, existing.Kind, existing.SrdSlug); err != nil {
				return err
			}
		}
		if _, err := q.SetInventoryEquipped(s.ctx, db.SetInventoryEquippedParams{Equipped: flag, ID: id}); err != nil {
			return fmt.Errorf("equipping item: %w", err)
		}
		return nil
	})
}

func (s *Service) DeleteItem(id int64) ([]Item, error) {
	existing, err := s.inventoryItem(id)
	if err != nil {
		return nil, err
	}
	return s.changeEquipment(existing.CharacterID, func(q *db.Queries) error {
		if err := q.DeleteInventoryItem(s.ctx, id); err != nil {
			return fmt.Errorf("deleting item: %w", err)
		}
		return nil
	})
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

func (s *Service) inventoryItem(id int64) (db.InventoryItem, error) {
	row, err := s.q.GetInventoryItem(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return db.InventoryItem{}, notFound("item", fmt.Sprint(id))
	}
	if err != nil {
		return db.InventoryItem{}, fmt.Errorf("loading item: %w", err)
	}
	return row, nil
}

func (s *Service) unequipKind(q *db.Queries, characterID int64, kind string) error {
	if err := q.UnequipInventoryKind(s.ctx, db.UnequipInventoryKindParams{
		CharacterID: characterID,
		Kind:        kind,
	}); err != nil {
		return fmt.Errorf("unequipping %s: %w", kind, err)
	}
	return nil
}
