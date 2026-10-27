package player

import (
	"testing"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/dice"
)

func addGear(t *testing.T, s *Service, characterID int64, kind, slug, name string, equipped bool) Item {
	t.Helper()
	items, err := s.SaveItem(ItemInput{CharacterID: characterID, Name: name, Kind: kind, SrdSlug: slug, Equipped: equipped})
	if err != nil {
		t.Fatalf("SaveItem(%s): %v", name, err)
	}
	for _, i := range items {
		if i.Name == name {
			return i
		}
	}
	t.Fatalf("%s isn't in the inventory after saving", name)
	return Item{}
}

func defenses(t *testing.T, s *Service, id int64) (evasion, score, major, severe int) {
	t.Helper()
	c, err := s.GetCharacter(id)
	if err != nil {
		t.Fatalf("GetCharacter: %v", err)
	}
	return c.Evasion, c.ArmorScore, c.ThresholdMajor, c.ThresholdSevere
}

func expectDefenses(t *testing.T, s *Service, id int64, evasion, score, major, severe int) {
	t.Helper()
	e, a, m, v := defenses(t, s, id)
	if e != evasion || a != score || m != major || v != severe {
		t.Fatalf("evasion %d, armor %d, thresholds %d/%d; want %d, %d, %d/%d", e, a, m, v, evasion, score, major, severe)
	}
}

func TestFeatureMods(t *testing.T) {
	for _, tc := range []struct {
		text string
		want statMods
	}{
		{"+1 to Armor Score", statMods{armorScore: 1}},
		{"+1 to your Armor Score", statMods{armorScore: 1}},
		{"+2 to Armor Score; −1 to Evasion", statMods{armorScore: 2, evasion: -1}},
		{"−2 to Evasion; −1 to Agility", statMods{evasion: -2}},
		{"−1 to all character traits and Evasion", statMods{evasion: -1}},
		{"+1 Evasion; you can walk on walls as easily as on the ground.", statMods{evasion: 1}},
		{"−1 to Evasion; +3 to Severe damage threshold", statMods{evasion: -1, severe: 3}},
		{"+2 to damage thresholds", statMods{major: 2, severe: 2}},
		{"Gain a bonus to your Armor Score equal to your Presence.", statMods{}},
		{"When you mark your last Armor Slot, increase your damage thresholds by +2 until you clear at least 1 Armor Slot.", statMods{}},
		{"If you do, you gain a −5 penalty to your damage thresholds until you choose to repair your armor.", statMods{}},
		{"−1 to Finesse", statMods{}},
	} {
		if got := featureMods(&cards.Feature{Description: tc.text}); got != tc.want {
			t.Errorf("featureMods(%q) = %+v, want %+v", tc.text, got, tc.want)
		}
	}
	if got := featureMods(nil); got != (statMods{}) {
		t.Errorf("featureMods(nil) = %+v", got)
	}
}

func TestParseWeaponDamage(t *testing.T) {
	for _, tc := range []struct {
		raw         string
		sides, flat int
		ok          bool
	}{
		{"d8+3", 8, 3, true},
		{"d10", 10, 0, true},
		{"d6−1", 6, -1, true},
		{"2d6", 0, 0, false},
		{"", 0, 0, false},
	} {
		sides, flat, err := parseWeaponDamage(tc.raw)
		if (err == nil) != tc.ok || sides != tc.sides || flat != tc.flat {
			t.Errorf("parseWeaponDamage(%q) = %d, %d, %v", tc.raw, sides, flat, err)
		}
	}
}

func TestEverySRDWeaponCanRollDamage(t *testing.T) {
	s := newTestService(t)
	for _, w := range s.catalog.ListWeapons() {
		sides, _, err := parseWeaponDamage(w.Damage)
		if err != nil {
			t.Errorf("%s: %v", w.Name, err)
			continue
		}
		if _, err := dice.ParseSides(sides); err != nil {
			t.Errorf("%s: %v", w.Name, err)
		}
	}
}

func TestArmorSetsDefensesAndUnarmoredResets(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 3, 7)

	chain := addGear(t, s, c.ID, "armor", "chainmailArmor", "Chainmail Armor", true)
	expectDefenses(t, s, c.ID, 9, 4, 8, 16)

	gambeson := addGear(t, s, c.ID, "armor", "gambesonArmor", "Gambeson Armor", true)
	expectDefenses(t, s, c.ID, 11, 3, 6, 12)

	items, _ := s.ListInventory(c.ID)
	for _, i := range items {
		if i.ID == chain.ID && i.Equipped {
			t.Fatal("equipping the gambeson should have taken the chainmail off")
		}
	}

	if _, err := s.SetItemEquipped(gambeson.ID, false); err != nil {
		t.Fatalf("SetItemEquipped: %v", err)
	}
	expectDefenses(t, s, c.ID, 10, 0, 1, 2)
}

func TestShieldAdjustsHandEnteredDefenses(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 3, 6, 13)

	addGear(t, s, c.ID, "secondary-weapon", "roundShield", "Round Shield", true)
	expectDefenses(t, s, c.ID, 10, 4, 6, 13)

	tower := addGear(t, s, c.ID, "secondary-weapon", "towerShield", "Tower Shield", true)
	expectDefenses(t, s, c.ID, 9, 5, 6, 13)

	if _, err := s.DeleteItem(tower.ID); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	expectDefenses(t, s, c.ID, 10, 3, 6, 13)
}

func TestFreeTextArmorLeavesDefensesToThePlayer(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 3, 6, 13)

	mail := addGear(t, s, c.ID, "armor", "", "Grandfather's mail", true)
	expectDefenses(t, s, c.ID, 10, 3, 6, 13)

	if _, err := s.SetItemEquipped(mail.ID, false); err != nil {
		t.Fatalf("SetItemEquipped: %v", err)
	}
	expectDefenses(t, s, c.ID, 10, 0, 1, 2)
}

func TestTwoHandedPrimaryFreesTheOffHand(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 1, 2)

	shield := addGear(t, s, c.ID, "secondary-weapon", "roundShield", "Round Shield", true)
	sword := addGear(t, s, c.ID, "primary-weapon", "greatsword", "Greatsword", true)

	equipped := func() map[int64]bool {
		items, _ := s.ListInventory(c.ID)
		out := map[int64]bool{}
		for _, i := range items {
			out[i.ID] = i.Equipped
		}
		return out
	}
	if on := equipped(); on[shield.ID] || !on[sword.ID] {
		t.Fatalf("a two-handed greatsword should leave no hand for the shield: %v", on)
	}
	expectDefenses(t, s, c.ID, 9, 0, 1, 2)

	if _, err := s.SetItemEquipped(shield.ID, true); err != nil {
		t.Fatalf("SetItemEquipped: %v", err)
	}
	if on := equipped(); !on[shield.ID] || on[sword.ID] {
		t.Fatalf("taking the shield up should put the greatsword down: %v", on)
	}
	expectDefenses(t, s, c.ID, 10, 1, 1, 2)
}

func TestArmorMarksFollowASmallerScore(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 1, 2)

	addGear(t, s, c.ID, "armor", "chainmailArmor", "Chainmail Armor", true)
	if _, err := s.MarkArmor(c.ID, 4); err != nil {
		t.Fatalf("MarkArmor: %v", err)
	}
	addGear(t, s, c.ID, "armor", "gambesonArmor", "Gambeson Armor", true)
	got, _ := s.GetCharacter(c.ID)
	if got.ArmorScore != 3 || got.ArmorMarked != 3 {
		t.Fatalf("armor %d with %d marked; want 3 with 3", got.ArmorScore, got.ArmorMarked)
	}
}

func TestItemSlugMustMatchItsKind(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 1, 2)

	for _, in := range []ItemInput{
		{CharacterID: c.ID, Name: "Greatsword", Kind: "secondary-weapon", SrdSlug: "greatsword"},
		{CharacterID: c.ID, Name: "Greatsword", Kind: "armor", SrdSlug: "greatsword"},
		{CharacterID: c.ID, Name: "Nothing", Kind: "item", SrdSlug: "noSuchThing"},
	} {
		if _, err := s.SaveItem(in); err == nil {
			t.Errorf("SaveItem(%s as %s) should have been refused", in.SrdSlug, in.Kind)
		}
	}

	items, err := s.SaveItem(ItemInput{CharacterID: c.ID, Name: "Shortsword", Kind: "secondary-weapon", SrdSlug: "shortsword"})
	if err != nil {
		t.Fatalf("SaveItem: %v", err)
	}
	if len(items) != 1 || items[0].Weapon == nil || items[0].Weapon.Damage != "d8" {
		t.Fatalf("the saved shortsword should carry its SRD card: %+v", items)
	}
}

func TestRollWeaponDamage(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 1, 2)
	other := newTestCharacter(t, s, 0, 1, 2)

	sword := addGear(t, s, c.ID, "primary-weapon", "greatsword", "Greatsword", true)
	stick := addGear(t, s, c.ID, "primary-weapon", "", "A stout stick", false)

	hit, err := s.RollWeaponDamage(c.ID, sword.ID, false)
	if err != nil {
		t.Fatalf("RollWeaponDamage: %v", err)
	}
	if hit.Count != 1 || hit.Sides != 10 || hit.Modifier != 3 || hit.Total < 4 || hit.Total > 13 {
		t.Fatalf("unexpected greatsword damage %+v", hit)
	}
	crit, err := s.RollWeaponDamage(c.ID, sword.ID, true)
	if err != nil {
		t.Fatalf("RollWeaponDamage critical: %v", err)
	}
	if crit.Modifier != 13 || crit.Total < 14 || crit.Total > 23 {
		t.Fatalf("a critical should add the dice's maximum: %+v", crit)
	}

	if _, err := s.RollWeaponDamage(c.ID, stick.ID, false); err == nil {
		t.Error("a free-text weapon has no damage to roll")
	}
	if _, err := s.RollWeaponDamage(other.ID, sword.ID, false); err == nil {
		t.Error("another character's weapon shouldn't roll")
	}
}

func TestTakeDamage(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 1, 2)
	addGear(t, s, c.ID, "armor", "chainmailArmor", "Chainmail Armor", true)

	for _, tc := range []struct {
		amount   int
		useArmor bool
		severity string
		reduced  string
		hp       int
	}{
		{0, false, "none", "none", 0},
		{5, false, "minor", "minor", 1},
		{8, false, "major", "major", 2},
		{16, false, "severe", "severe", 3},
		{16, true, "severe", "major", 2},
		{5, true, "minor", "none", 0},
	} {
		before, _ := s.GetCharacter(c.ID)
		got, err := s.TakeDamage(DamageInput{CharacterID: c.ID, Amount: tc.amount, UseArmor: tc.useArmor})
		if err != nil {
			t.Fatalf("TakeDamage(%d): %v", tc.amount, err)
		}
		if got.Severity != tc.severity || got.Reduced != tc.reduced || got.HPMarked != min(tc.hp, before.HPMax-before.HPMarked) {
			t.Errorf("TakeDamage(%d, armor %v) = %+v", tc.amount, tc.useArmor, got)
		}
		if got.ArmorUsed != (tc.useArmor && tc.amount > 0) {
			t.Errorf("TakeDamage(%d, armor %v) used armor = %v", tc.amount, tc.useArmor, got.ArmorUsed)
		}
		if _, err := s.SetVitals(c.ID, 0, 0, got.Character.Hope, 0); err != nil {
			t.Fatalf("SetVitals: %v", err)
		}
	}

	if _, err := s.MarkArmor(c.ID, 4); err != nil {
		t.Fatalf("MarkArmor: %v", err)
	}
	if _, err := s.TakeDamage(DamageInput{CharacterID: c.ID, Amount: 9, UseArmor: true}); err == nil {
		t.Error("marking an Armor Slot with none left should be refused")
	}

	full, _ := s.GetCharacter(c.ID)
	if _, err := s.SetVitals(c.ID, full.HPMax-1, 0, full.Hope, 0); err != nil {
		t.Fatalf("SetVitals: %v", err)
	}
	last, err := s.TakeDamage(DamageInput{CharacterID: c.ID, Amount: 20})
	if err != nil {
		t.Fatalf("TakeDamage: %v", err)
	}
	if !last.Downed || last.HPMarked != 1 || last.Character.HPMarked != full.HPMax {
		t.Fatalf("marking the last Hit Point should say so: %+v", last)
	}

	bare := newTestCharacter(t, s, 0, 0, 0)
	if _, err := s.TakeDamage(DamageInput{CharacterID: bare.ID, Amount: 3}); err == nil {
		t.Error("damage against unset thresholds should be refused")
	}
}

func TestLevelUpPlanOffersMulticlassCards(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 1, 2)

	plan, err := s.PlanLevelUp(c.ID)
	if err != nil {
		t.Fatalf("PlanLevelUp: %v", err)
	}
	if len(plan.MulticlassCards) == 0 {
		t.Fatal("a character who hasn't multiclassed should be offered the other domains' cards")
	}
	for _, card := range plan.MulticlassCards {
		if contains(c.Domains, card.Domain) {
			t.Errorf("%s is from %s, which the character already has", card.Name, card.Domain)
		}
		if cardLevel(card) > plan.ToLevel {
			t.Errorf("%s is level %d, above %d", card.Name, cardLevel(card), plan.ToLevel)
		}
	}
}
