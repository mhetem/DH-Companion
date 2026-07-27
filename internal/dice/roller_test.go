package dice

import "testing"

func TestRollerDamageAppliesModifierOnce(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		sides    int
		modifier int
	}{
		{name: "single die", count: 1, sides: 12, modifier: 0},
		{name: "single die with modifier", count: 1, sides: 12, modifier: 3},
		{name: "two dice", count: 2, sides: 6, modifier: 0},
		{name: "two dice with modifier", count: 2, sides: 6, modifier: 3},
		{name: "many dice with negative modifier", count: 5, sides: 8, modifier: -2},
	}

	const samples = 20000

	var r Roller
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			low := tt.count + tt.modifier           // every die shows 1
			high := tt.count*tt.sides + tt.modifier // every die shows its max
			sum := 0

			for i := 0; i < samples; i++ {
				got, err := r.Damage(tt.count, tt.sides, tt.modifier)
				if err != nil {
					t.Fatalf("Damage(%d, %d, %d) = %v", tt.count, tt.sides, tt.modifier, err)
				}
				if got.Total < low || got.Total > high {
					t.Fatalf("Total = %d, want in [%d, %d]", got.Total, low, high)
				}
				if got.Count != tt.count || got.Sides != tt.sides || got.Modifier != tt.modifier {
					t.Fatalf("echoed %+v, want count %d sides %d modifier %d",
						got, tt.count, tt.sides, tt.modifier)
				}
				sum += got.Total
			}

			// The mean pins both halves of the contract, and unlike watching for
			// the extremes it stays reliable as the dice count grows: skipping
			// dice drags the average down toward the modifier, and applying the
			// modifier per-die pushes it up by modifier*(count-1).
			want := float64(tt.count)*(float64(tt.sides)+1)/2 + float64(tt.modifier)
			got := float64(sum) / samples
			if diff := got - want; diff < -0.5 || diff > 0.5 {
				t.Errorf("mean total over %d rolls = %.2f, want %.2f", samples, got, want)
			}
		})
	}
}

// The multi-dice branch of RollDamage is easy to get wrong in a way that still
// returns a plausible number, so pin it against the single-die branch: 2d6 has
// to beat 6 sometimes, which one die never can.
func TestRollDamageRollsEveryDie(t *testing.T) {
	aboveOneDie := 0
	for i := 0; i < 20000; i++ {
		if got := RollDamage(2, D6, 0); got > 6 {
			aboveOneDie++
		}
		if got := RollDamage(2, D6, 0); got < 2 {
			t.Fatalf("RollDamage(2, D6, 0) = %d, want at least 2", got)
		}
	}
	if aboveOneDie == 0 {
		t.Fatal("2d6 never exceeded 6 — the second die is not being rolled")
	}
}

func TestRollerDamageRejectsUnrollableDice(t *testing.T) {
	var r Roller
	for _, sides := range []int{0, -1, 1, 3, 7, 13, 99, 101} {
		if _, err := r.Damage(2, sides, 0); err == nil {
			t.Errorf("Damage(2, %d, 0) succeeded, want an error", sides)
		}
	}
}

func TestRollerDamageClampsCount(t *testing.T) {
	var r Roller
	for _, count := range []int{0, -3} {
		got, err := r.Damage(count, 6, 0)
		if err != nil {
			t.Fatalf("Damage(%d, 6, 0) = %v", count, err)
		}
		if got.Count != 1 || got.Total < 1 || got.Total > 6 {
			t.Errorf("Damage(%d, 6, 0) = %+v, want a single d6", count, got)
		}
	}
}

func TestRollerSizesMatchesAllowed(t *testing.T) {
	var r Roller
	got := r.Sizes()
	want := Allowed()
	if len(got) != len(want) {
		t.Fatalf("Sizes() = %v, want %d entries", got, len(want))
	}
	for i, s := range want {
		if got[i] != int(s) {
			t.Errorf("Sizes()[%d] = %d, want %d", i, got[i], int(s))
		}
	}
}

func TestRollerGMStaysInRange(t *testing.T) {
	var r Roller
	for _, tt := range []struct {
		name         string
		advantage    bool
		disadvantage bool
		modifier     int
	}{
		{name: "plain"},
		{name: "advantage", advantage: true},
		{name: "disadvantage", disadvantage: true},
		{name: "both cancel", advantage: true, disadvantage: true},
		{name: "with modifier", modifier: 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 5000; i++ {
				got := r.GM(tt.advantage, tt.disadvantage, tt.modifier)
				// A critical returns a flat 20 with the modifier dropped, the
				// same way DualityDiceRoll drops it on a matching pair.
				if got.Msg == "Critical!" {
					if got.Result != 20 {
						t.Fatalf("critical %+v: Result = %d, want 20", got, got.Result)
					}
					continue
				}
				if got.Result < 1+tt.modifier || got.Result > 20+tt.modifier {
					t.Fatalf("%+v: Result = %d, want in [%d, %d]",
						got, got.Result, 1+tt.modifier, 20+tt.modifier)
				}
			}
		})
	}
}

func TestRollerDualityMatchesPackageFunction(t *testing.T) {
	var r Roller
	for i := 0; i < 2000; i++ {
		got := r.Duality(false, false, 2)
		if got.Hope < 1 || got.Hope > 12 || got.Fear < 1 || got.Fear > 12 {
			t.Fatalf("%+v: dice out of range", got)
		}
		want := got.Hope + got.Fear + 2
		if got.Hope == got.Fear {
			want = got.Hope + got.Fear
		}
		if got.Result != want {
			t.Fatalf("%+v: Result = %d, want %d", got, got.Result, want)
		}
	}
}
