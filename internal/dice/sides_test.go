package dice

import "testing"

func TestAllowedSidesAreValid(t *testing.T) {
	for _, s := range Allowed() {
		if !s.Valid() {
			t.Errorf("%s is in Allowed() but reports invalid", s)
		}
		if s < 1 {
			t.Errorf("%s would panic rand.Intn", s)
		}
	}
}

func TestParseSidesAcceptsOnlyTheAllowedSet(t *testing.T) {
	for _, want := range Allowed() {
		got, err := ParseSides(int(want))
		if err != nil {
			t.Errorf("ParseSides(%d) = %v, want %s", int(want), err, want)
		}
		if got != want {
			t.Errorf("ParseSides(%d) = %s, want %s", int(want), got, want)
		}
	}

	for _, n := range []int{-5, 0, 1, 2, 3, 5, 7, 9, 11, 14, 16, 30, 99, 101} {
		if _, err := ParseSides(n); err == nil {
			t.Errorf("ParseSides(%d) succeeded, want an error", n)
		}
	}
}

// The guard exists so a Sides produced by a raw conversion rather than by a
// constant or ParseSides can't reach rand.Intn and panic.
func TestRollDiceDoesNotPanicOnAnyValue(t *testing.T) {
	for n := -100; n <= 100; n++ {
		got := RollDice(Sides(n))
		s := Sides(n)
		if s.Valid() {
			if got < 1 || got > n {
				t.Errorf("RollDice(%s) = %d, want in [1, %d]", s, got, n)
			}
			continue
		}
		if got != 0 {
			t.Errorf("RollDice(%s) = %d, want 0 for an unrollable die", s, got)
		}
	}
}

func TestSidesString(t *testing.T) {
	if got := D20.String(); got != "d20" {
		t.Errorf("D20.String() = %q, want %q", got, "d20")
	}
}
