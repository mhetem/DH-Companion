package rules

import (
	"reflect"
	"testing"

	"github.com/mhetem/DH-Companion/internal/cards"
)

func adv(advType, tier string, count int) EncounterAdversary {
	return EncounterAdversary{
		Adversary: cards.Adversary{
			Meta: cards.Meta{
				Kind: cards.KindAdversary,
				Name: advType,
				Type: advType,
				Tier: tier,
			},
		},
		Count: count,
	}
}

func TestGetBudgetForEncounter(t *testing.T) {
	tests := []struct {
		players int
		want    int
	}{
		{1, 5},
		{2, 8},
		{3, 11},
		{4, 14},
		{5, 17},
		{6, 20},
	}
	for _, tt := range tests {
		if got := getBudgetForEncounter(tt.players); got != tt.want {
			t.Errorf("getBudgetForEncounter(%d) = %d, want %d", tt.players, got, tt.want)
		}
	}
}

func TestTierValue(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"1", 1},
		{"4", 4},
		{"  3  ", 3},
		{"", 0},
		{"Tier 2", 0},
		{"two", 0},
		{"0", 0},
	}
	for _, tt := range tests {
		if got := tierValue(tt.in); got != tt.want {
			t.Errorf("tierValue(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestAdversaryCost(t *testing.T) {
	tests := []struct {
		advType string
		want    int
	}{
		{"Social", 1},
		{"Support", 1},
		{"Standard", 2},
		{"Horde", 2},
		{"Skulk", 2},
		{"Ranged", 2},
		{"Leader", 3},
		{"Bruiser", 4},
		{"Solo", 5},
		{"Minion", 2},
		{"", 2},
		{"standard", 2},
	}
	for _, tt := range tests {
		if got := adversaryCost(tt.advType); got != tt.want {
			t.Errorf("adversaryCost(%q) = %d, want %d", tt.advType, got, tt.want)
		}
	}
}

func TestIsBigType(t *testing.T) {
	big := []string{"Bruiser", "Horde", "Leader", "Solo"}
	small := []string{"Social", "Support", "Standard", "Skulk", "Ranged", "Minion", "", "solo"}

	for _, tt := range big {
		if !isBigType(tt) {
			t.Errorf("isBigType(%q) = false, want true", tt)
		}
	}
	for _, tt := range small {
		if isBigType(tt) {
			t.Errorf("isBigType(%q) = true, want false", tt)
		}
	}
}

func TestComputeBudgetSpend(t *testing.T) {
	tests := []struct {
		name      string
		settings  EncounterSettings
		picks     []EncounterAdversary
		wantSpent int
	}{
		{
			name:      "no picks spends nothing",
			settings:  EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks:     nil,
			wantSpent: 0,
		},
		{
			name:     "costs are multiplied by count",
			settings: EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks: []EncounterAdversary{
				adv("Standard", "1", 3),
				adv("Support", "1", 2),
			},
			wantSpent: 8,
		},
		{
			name:     "solo and bruiser",
			settings: EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks: []EncounterAdversary{
				adv("Solo", "1", 1),
				adv("Bruiser", "1", 1),
			},
			wantSpent: 9,
		},
		{
			name:     "zero and negative counts are ignored",
			settings: EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks: []EncounterAdversary{
				adv("Solo", "1", 0),
				adv("Bruiser", "1", -2),
				adv("Standard", "1", 1),
			},
			wantSpent: 2,
		},
		{
			name:     "unknown type costs 2",
			settings: EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks: []EncounterAdversary{
				adv("Mystery", "1", 2),
			},
			wantSpent: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeBudget(tt.settings, tt.picks)
			if got.Spent != tt.wantSpent {
				t.Errorf("Spent = %d, want %d", got.Spent, tt.wantSpent)
			}
		})
	}
}

func TestComputeBudgetMinionBatching(t *testing.T) {
	tests := []struct {
		name      string
		partySize int
		minions   int
		wantSpent int
	}{
		{"none", 4, 0, 0},
		{"one minion is one batch", 4, 1, 1},
		{"exactly one party-sized batch", 4, 4, 1},
		{"one over rounds up", 4, 5, 2},
		{"two full batches", 4, 8, 2},
		{"seven with party of three", 3, 7, 3},
		{"party of one bills per minion", 1, 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			picks := []EncounterAdversary{adv("Minion", "1", tt.minions)}
			got := ComputeBudget(EncounterSettings{PartySize: tt.partySize, PartyTier: "1"}, picks)
			if got.Spent != tt.wantSpent {
				t.Errorf("Spent = %d, want %d", got.Spent, tt.wantSpent)
			}
		})
	}
}

func TestComputeBudgetMinionsSplitAcrossPicksSharePartyDivisor(t *testing.T) {
	picks := []EncounterAdversary{
		adv("Minion", "1", 5),
		adv("Minion", "1", 5),
	}
	got := ComputeBudget(EncounterSettings{PartySize: 4, PartyTier: "1"}, picks)
	if got.Spent != 3 {
		t.Errorf("Spent = %d, want 3", got.Spent)
	}

	picks = []EncounterAdversary{
		adv("Minion", "1", 1),
		adv("Minion", "1", 1),
		adv("Minion", "1", 1),
		adv("Minion", "1", 1),
	}
	got = ComputeBudget(EncounterSettings{PartySize: 4, PartyTier: "1"}, picks)
	if got.Spent != 1 {
		t.Errorf("pooled minions: Spent = %d, want 1", got.Spent)
	}
}

func TestComputeBudgetAdjustments(t *testing.T) {
	tests := []struct {
		name            string
		settings        EncounterSettings
		picks           []EncounterAdversary
		wantBudget      int
		wantAdjustments []string
	}{
		{
			name:            "empty encounter gets the base budget only",
			settings:        EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks:           nil,
			wantBudget:      14,
			wantAdjustments: nil,
		},
		{
			name:       "easy subtracts one",
			settings:   EncounterSettings{PartySize: 4, PartyTier: "1", Difficulty: "easy"},
			picks:      []EncounterAdversary{adv("Solo", "1", 1)},
			wantBudget: 13,
			wantAdjustments: []string{
				"Easy encounter: -1",
			},
		},
		{
			name:       "hard adds two",
			settings:   EncounterSettings{PartySize: 4, PartyTier: "1", Difficulty: "hard"},
			picks:      []EncounterAdversary{adv("Solo", "1", 1)},
			wantBudget: 16,
			wantAdjustments: []string{
				"Hard encounter: +2",
			},
		},
		{
			name:            "unknown difficulty is treated as standard",
			settings:        EncounterSettings{PartySize: 4, PartyTier: "1", Difficulty: "brutal"},
			picks:           []EncounterAdversary{adv("Solo", "1", 1)},
			wantBudget:      14,
			wantAdjustments: nil,
		},
		{
			name:            "one solo does not trigger the multi-solo penalty",
			settings:        EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks:           []EncounterAdversary{adv("Solo", "1", 1)},
			wantBudget:      14,
			wantAdjustments: nil,
		},
		{
			name:       "two solos subtract two",
			settings:   EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks:      []EncounterAdversary{adv("Solo", "1", 2)},
			wantBudget: 12,
			wantAdjustments: []string{
				"More than one solo: -2",
			},
		},
		{
			name:     "two solos across separate picks still subtract two once",
			settings: EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks: []EncounterAdversary{
				adv("Solo", "1", 1),
				adv("Solo", "1", 1),
			},
			wantBudget: 12,
			wantAdjustments: []string{
				"More than one solo: -2",
			},
		},
		{
			name:       "adversary below party tier adds one",
			settings:   EncounterSettings{PartySize: 4, PartyTier: "2"},
			picks:      []EncounterAdversary{adv("Solo", "1", 1)},
			wantBudget: 15,
			wantAdjustments: []string{
				"Adversary below party tier: +1",
			},
		},
		{
			name:       "several below-tier adversaries add one total",
			settings:   EncounterSettings{PartySize: 4, PartyTier: "3"},
			picks:      []EncounterAdversary{adv("Solo", "1", 1), adv("Bruiser", "2", 1)},
			wantBudget: 15,
			wantAdjustments: []string{
				"Adversary below party tier: +1",
			},
		},
		{
			name:            "adversary at or above party tier adds nothing",
			settings:        EncounterSettings{PartySize: 4, PartyTier: "2"},
			picks:           []EncounterAdversary{adv("Solo", "2", 1), adv("Bruiser", "3", 1)},
			wantBudget:      14,
			wantAdjustments: nil,
		},
		{
			name:            "unparseable party tier disables the tier check",
			settings:        EncounterSettings{PartySize: 4, PartyTier: ""},
			picks:           []EncounterAdversary{adv("Solo", "1", 1)},
			wantBudget:      14,
			wantAdjustments: nil,
		},
		{
			name:            "unparseable adversary tier disables the tier check",
			settings:        EncounterSettings{PartySize: 4, PartyTier: "3"},
			picks:           []EncounterAdversary{adv("Solo", "n/a", 1)},
			wantBudget:      14,
			wantAdjustments: nil,
		},
		{
			name:       "no big adversary adds one",
			settings:   EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks:      []EncounterAdversary{adv("Standard", "1", 2), adv("Support", "1", 1)},
			wantBudget: 15,
			wantAdjustments: []string{
				"No bruiser, horde, leader, or solo: +1",
			},
		},
		{
			name:       "minions alone count as no big adversary",
			settings:   EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks:      []EncounterAdversary{adv("Minion", "1", 4)},
			wantBudget: 15,
			wantAdjustments: []string{
				"No bruiser, horde, leader, or solo: +1",
			},
		},
		{
			name:            "a horde counts as a big adversary",
			settings:        EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks:           []EncounterAdversary{adv("Horde", "1", 1)},
			wantBudget:      14,
			wantAdjustments: nil,
		},
		{
			name:            "a leader counts as a big adversary",
			settings:        EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks:           []EncounterAdversary{adv("Leader", "1", 1)},
			wantBudget:      14,
			wantAdjustments: nil,
		},
		{
			name:            "picks with no count do not trigger the no-big bonus",
			settings:        EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks:           []EncounterAdversary{adv("Standard", "1", 0)},
			wantBudget:      14,
			wantAdjustments: nil,
		},
		{
			name:     "adjustments stack in order",
			settings: EncounterSettings{PartySize: 4, PartyTier: "3", Difficulty: "hard"},
			picks: []EncounterAdversary{
				adv("Solo", "1", 2),
				adv("Standard", "3", 1),
			},

			wantBudget: 15,
			wantAdjustments: []string{
				"Hard encounter: +2",
				"More than one solo: -2",
				"Adversary below party tier: +1",
			},
		},
		{
			name:       "easy plus below tier plus no big",
			settings:   EncounterSettings{PartySize: 3, PartyTier: "2", Difficulty: "easy"},
			picks:      []EncounterAdversary{adv("Skulk", "1", 2)},
			wantBudget: 12,
			wantAdjustments: []string{
				"Easy encounter: -1",
				"Adversary below party tier: +1",
				"No bruiser, horde, leader, or solo: +1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeBudget(tt.settings, tt.picks)
			if got.Budget != tt.wantBudget {
				t.Errorf("Budget = %d, want %d", got.Budget, tt.wantBudget)
			}
			if !reflect.DeepEqual(got.Adjustments, tt.wantAdjustments) {
				t.Errorf("Adjustments = %#v, want %#v", got.Adjustments, tt.wantAdjustments)
			}
		})
	}
}

func TestComputeBudgetClampsPartySize(t *testing.T) {
	for _, size := range []int{0, -1, -10} {
		got := ComputeBudget(EncounterSettings{PartySize: size, PartyTier: "1"}, nil)
		if got.PartySize != 1 {
			t.Errorf("PartySize %d: got %d, want 1", size, got.PartySize)
		}
		if got.Budget != 5 {
			t.Errorf("PartySize %d: Budget = %d, want 5", size, got.Budget)
		}
	}

	got := ComputeBudget(EncounterSettings{PartySize: 0, PartyTier: "1"}, []EncounterAdversary{adv("Minion", "1", 3)})
	if got.Spent != 3 {
		t.Errorf("Spent = %d, want 3", got.Spent)
	}
}

func TestComputeBudgetRemainingAndOver(t *testing.T) {
	tests := []struct {
		name          string
		settings      EncounterSettings
		picks         []EncounterAdversary
		wantRemaining int
		wantOver      bool
	}{
		{
			name:          "under budget",
			settings:      EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks:         []EncounterAdversary{adv("Solo", "1", 1)}, // 5 of 14
			wantRemaining: 9,
			wantOver:      false,
		},
		{
			name:          "exactly on budget is not over",
			settings:      EncounterSettings{PartySize: 4, PartyTier: "1"},
			picks:         []EncounterAdversary{adv("Standard", "1", 7), adv("Support", "1", 1)},
			wantRemaining: 0,
			wantOver:      false,
		},
		{
			name:          "over budget",
			settings:      EncounterSettings{PartySize: 2, PartyTier: "1"},
			picks:         []EncounterAdversary{adv("Solo", "1", 2)},
			wantRemaining: -4,
			wantOver:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeBudget(tt.settings, tt.picks)
			if got.Remaining != tt.wantRemaining {
				t.Errorf("Remaining = %d, want %d (budget %d, spent %d)", got.Remaining, tt.wantRemaining, got.Budget, got.Spent)
			}
			if got.Over != tt.wantOver {
				t.Errorf("Over = %v, want %v (budget %d, spent %d)", got.Over, tt.wantOver, got.Budget, got.Spent)
			}
			if got.Remaining != got.Budget-got.Spent {
				t.Errorf("Remaining %d is not Budget %d - Spent %d", got.Remaining, got.Budget, got.Spent)
			}
		})
	}
}
