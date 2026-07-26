package rules

import "github.com/mhetem/DH-Companion/internal/cards"

type EncounterAdversary struct {
	cards.Adversary
	Count  int    `json:"count"`
	Source string `json:"source"`
	ID     int64  `json:"id"`
}

type EncounterView struct {
	ID          int64                `json:"id"`
	Name        string               `json:"name"`
	Adversaries []EncounterAdversary `json:"adversaries"`
	TotalCount  int                  `json:"totalCount"`
}
