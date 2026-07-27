package dice

import (
	"fmt"
	"strconv"
	"strings"
)

type Sides int

const (
	D4   Sides = 4
	D6   Sides = 6
	D8   Sides = 8
	D10  Sides = 10
	D12  Sides = 12
	D20  Sides = 20
	D100 Sides = 100
)

func Allowed() []Sides {
	return []Sides{D4, D6, D8, D10, D12, D20, D100}
}

func (s Sides) Valid() bool {
	for _, allowed := range Allowed() {
		if s == allowed {
			return true
		}
	}
	return false
}

func (s Sides) String() string { return "d" + strconv.Itoa(int(s)) }

func ParseSides(n int) (Sides, error) {
	s := Sides(n)
	if !s.Valid() {
		names := make([]string, 0, len(Allowed()))
		for _, allowed := range Allowed() {
			names = append(names, allowed.String())
		}
		return 0, fmt.Errorf("d%d is not a rollable die; use one of %s", n, strings.Join(names, ", "))
	}
	return s, nil
}
