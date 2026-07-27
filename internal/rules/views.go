package rules

import "github.com/mhetem/DH-Companion/internal/cards"

// Source tells the frontend which library a pick was resolved against.
const (
	SourceSRD    = "srd"
	SourceCustom = "custom"
)

type EncounterAdversary struct {
	cards.Adversary
	Count  int    `json:"count"`
	Source string `json:"source"`
	ID     int64  `json:"id"`
	// Unresolved marks a pick whose slug no longer matches anything — a custom
	// adversary that was deleted after the encounter referenced it. The pick is
	// kept so the encounter still opens; only Slug and Count are meaningful.
	Unresolved bool `json:"unresolved,omitempty"`
}

type EncounterView struct {
	ID              int64                `json:"id"`
	Name            string               `json:"name"`
	PartyID         *int64               `json:"partyId"`
	Adversaries     []EncounterAdversary `json:"adversaries"`
	TotalCount      int                  `json:"totalCount"`
	EnvironmentSlug *string              `json:"environmentSlug"`
	// Environment is the resolved card for EnvironmentSlug, nil when none is
	// attached or when the slug no longer resolves.
	Environment *cards.Environment `json:"environment"`
	// Budget is computed from the attached party at standard difficulty, nil
	// when no party is attached. The live meter re-computes via ComputeBudget
	// as the builder is edited.
	Budget    *BudgetSummary `json:"budget"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}
