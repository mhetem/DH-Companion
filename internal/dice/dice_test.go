package dice

import "testing"

// RollDice and DualityDiceRoll draw from the package-level math/rand source, so
// these tests assert invariants over many rolls rather than fixing a seed.
const rolls = 20000

func TestRollDiceRange(t *testing.T) {
	for _, sides := range Allowed() {
		lowest, highest := int(sides)+1, 0
		for i := 0; i < rolls; i++ {
			got := RollDice(sides)
			if got < 1 || got > int(sides) {
				t.Fatalf("RollDice(%s) = %d, want in [1, %d]", sides, got, int(sides))
			}
			if got < lowest {
				lowest = got
			}
			if got > highest {
				highest = got
			}
		}
		if lowest != 1 {
			t.Errorf("RollDice(%s): lowest value seen was %d, want 1", sides, lowest)
		}
		if highest != int(sides) {
			t.Errorf("RollDice(%s): highest value seen was %d, want %d", sides, highest, int(sides))
		}
	}
}

func TestRollDiceCoversEveryFace(t *testing.T) {
	const sides = D12
	seen := make(map[int]int, sides)
	for i := 0; i < rolls; i++ {
		seen[RollDice(sides)]++
	}
	for face := 1; face <= int(sides); face++ {
		if seen[face] == 0 {
			t.Errorf("face %d never rolled in %d rolls", face, rolls)
		}
	}
	if len(seen) != int(sides) {
		t.Errorf("saw %d distinct faces over %d rolls, want %d", len(seen), rolls, int(sides))
	}
}

func TestDualityDiceRollDiceAreInRange(t *testing.T) {
	for i := 0; i < rolls; i++ {
		got := DualityDiceRoll(false, false, 0)
		if got.Hope < 1 || got.Hope > 12 {
			t.Fatalf("Hope = %d, want in [1, 12]", got.Hope)
		}
		if got.Fear < 1 || got.Fear > 12 {
			t.Fatalf("Fear = %d, want in [1, 12]", got.Fear)
		}
	}
}

func TestDualityDiceRollResult(t *testing.T) {
	tests := []struct {
		name         string
		advantage    bool
		disadvantage bool
		modifier     int
		lowOffset    int
		highOffset   int
	}{
		{name: "plain roll", modifier: 0},
		{name: "plain roll with positive modifier", modifier: 3},
		{name: "plain roll with negative modifier", modifier: -4},
		{name: "advantage", advantage: true, lowOffset: 1, highOffset: 6},
		{name: "advantage with modifier", advantage: true, modifier: 2, lowOffset: 1, highOffset: 6},
		{name: "disadvantage", disadvantage: true, lowOffset: -6, highOffset: -1},
		{name: "disadvantage with modifier", disadvantage: true, modifier: 2, lowOffset: -6, highOffset: -1},
		{name: "advantage and disadvantage cancel", advantage: true, disadvantage: true, modifier: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sawCrit := false
			offsets := make(map[int]bool)

			for i := 0; i < rolls; i++ {
				got := DualityDiceRoll(tt.advantage, tt.disadvantage, tt.modifier)
				base := got.Hope + got.Fear

				if got.Hope == got.Fear {
					sawCrit = true
					if got.Result != base {
						t.Fatalf("critical %+v: Result = %d, want %d", got, got.Result, base)
					}
					if got.Msg != "Critical Success!" {
						t.Fatalf("critical %+v: Msg = %q, want %q", got, got.Msg, "Critical Success!")
					}
					continue
				}

				if got.Msg != "" {
					t.Fatalf("non-critical %+v: Msg = %q, want empty", got, got.Msg)
				}

				offset := got.Result - (base + tt.modifier)
				if offset < tt.lowOffset || offset > tt.highOffset {
					t.Fatalf("%+v (modifier %d): Result = %d, want hope+fear+modifier offset in [%d, %d], got %d",
						got, tt.modifier, got.Result, tt.lowOffset, tt.highOffset, offset)
				}
				offsets[offset] = true
			}

			if !sawCrit {
				t.Errorf("no matching hope/fear in %d rolls; expected roughly 1 in 12", rolls)
			}
			want := tt.highOffset - tt.lowOffset + 1
			if got := len(offsets); got != want {
				t.Errorf("extra die took %d distinct values over %d rolls, want %d", got, rolls, want)
			}
		})
	}
}

func TestDualityDiceRollHopeAndFearBothVary(t *testing.T) {
	hopeSeen := make(map[int]bool)
	fearSeen := make(map[int]bool)
	for i := 0; i < rolls; i++ {
		got := DualityDiceRoll(false, false, 0)
		hopeSeen[got.Hope] = true
		fearSeen[got.Fear] = true
	}
	if len(hopeSeen) != 12 {
		t.Errorf("Hope took %d distinct values, want 12", len(hopeSeen))
	}
	if len(fearSeen) != 12 {
		t.Errorf("Fear took %d distinct values, want 12", len(fearSeen))
	}
}

func TestDualityDiceRollHopeAndFearAreSeparateDice(t *testing.T) {
	matches := 0
	for i := 0; i < rolls; i++ {
		got := DualityDiceRoll(false, false, 0)
		if got.Hope == got.Fear {
			matches++
		}
	}
	if matches < rolls/24 || matches > rolls/6 {
		t.Errorf("hope == fear on %d of %d rolls, want roughly %d", matches, rolls, rolls/12)
	}
}
