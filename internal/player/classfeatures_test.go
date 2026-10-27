package player

import (
	"testing"
)

func characterWith(t *testing.T, s *Service, class, subclass string, traits map[string]int, major, severe int) Character {
	t.Helper()
	c, err := s.SaveCharacter(CharacterInput{
		Name:            "Sabine",
		ClassSlug:       class,
		SubclassSlug:    subclass,
		AncestrySlug:    "clank",
		CommunitySlug:   s.catalog.ListCommunities()[0].Slug,
		Traits:          traits,
		Evasion:         10,
		ThresholdMajor:  major,
		ThresholdSevere: severe,
	})
	if err != nil {
		t.Fatalf("SaveCharacter: %v", err)
	}
	return c
}

func setColumn(t *testing.T, s *Service, id int64, column string, value any) {
	t.Helper()
	if _, err := s.conn.ExecContext(s.ctx, "UPDATE characters SET "+column+" = ? WHERE id = ?", value, id); err != nil {
		t.Fatalf("setting %s: %v", column, err)
	}
}

func TestBareBonesIsABaselineWithoutArmor(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 1, 2)

	if _, err := s.AddDomainCard(c.ID, "bareBones", "loadout"); err != nil {
		t.Fatalf("AddDomainCard: %v", err)
	}
	expectDefenses(t, s, c.ID, 10, 4, 10, 20)

	chain := addGear(t, s, c.ID, "armor", "chainmailArmor", "Chainmail Armor", true)
	expectDefenses(t, s, c.ID, 9, 4, 8, 16)

	if _, err := s.SetItemEquipped(chain.ID, false); err != nil {
		t.Fatalf("SetItemEquipped: %v", err)
	}
	expectDefenses(t, s, c.ID, 10, 4, 10, 20)

	if _, err := s.MoveDomainCard(c.ID, "bareBones", "vault"); err != nil {
		t.Fatalf("MoveDomainCard: %v", err)
	}
	expectDefenses(t, s, c.ID, 10, 0, 1, 2)
}

func TestWarlockFavor(t *testing.T) {
	s := newTestService(t)
	traits := map[string]int{"agility": 0, "strength": 0, "finesse": 1, "instinct": 1, "presence": 2, "knowledge": -1}
	c := characterWith(t, s, "warlock", "pactOfTheEndless", traits, 5, 10)
	if c.Favor != StartingFavor {
		t.Fatalf("a warlock starts with %d Favor, got %d", StartingFavor, c.Favor)
	}

	traded, err := s.FavorInsteadOfHope(c.ID)
	if err != nil {
		t.Fatalf("FavorInsteadOfHope: %v", err)
	}
	if traded.Favor != StartingFavor+1 || traded.Hope != c.Hope-1 {
		t.Fatalf("trading should move a Hope to Favor: %d Favor, %d Hope", traded.Favor, traded.Hope)
	}

	rested, err := s.Rest(RestInput{CharacterID: c.ID, Moves: []string{"tribute"}})
	if err != nil {
		t.Fatalf("Rest: %v", err)
	}
	if rested.Character.Favor != FavorMax {
		t.Fatalf("tribute adds Presence (2) Favor, capped at %d: got %d", FavorMax, rested.Character.Favor)
	}

	bard := newTestCharacter(t, s, 0, 1, 2)
	if _, err := s.FavorInsteadOfHope(bard.ID); err == nil {
		t.Error("only a warlock has Favor")
	}
	if _, err := s.Rest(RestInput{CharacterID: bard.ID, Moves: []string{"tribute"}}); err == nil {
		t.Error("only a warlock can show tribute")
	}
}

func TestStancesAndFocus(t *testing.T) {
	s := newTestService(t)
	traits := map[string]int{"agility": 1, "strength": 2, "finesse": 0, "instinct": 1, "presence": 0, "knowledge": -1}
	c := characterWith(t, s, "brawler", "martialArtist", traits, 5, 10)

	for _, title := range []string{"Reliable", "Favored"} {
		if _, err := s.LearnStance(c.ID, title, true); err != nil {
			t.Fatalf("LearnStance(%s): %v", title, err)
		}
	}
	if _, err := s.LearnStance(c.ID, "Quick", true); err == nil {
		t.Error("a level 1 Martial Artist knows two stances")
	}
	if _, err := s.LearnStance(c.ID, "Favored", false); err != nil {
		t.Fatalf("forgetting a stance: %v", err)
	}
	if _, err := s.LearnStance(c.ID, "Anchored", true); err == nil {
		t.Error("Anchored is a tier 2 stance")
	}

	if _, err := s.ShiftStance(c.ID, "Reliable", "focus"); err == nil {
		t.Error("shifting with no Focus should be refused")
	}
	focused, err := s.Refocus(c.ID)
	if err != nil {
		t.Fatalf("Refocus: %v", err)
	}
	if len(focused.Rolls) != 1 || focused.Focus < 1 || focused.Focus > 6 {
		t.Fatalf("Instinct 1 rolls one d6: %+v", focused)
	}
	if _, err := s.ShiftStance(c.ID, "Reliable", "stress"); err == nil {
		t.Error("marking a Stress to shift takes Flow State")
	}
	shifted, err := s.ShiftStance(c.ID, "Reliable", "focus")
	if err != nil {
		t.Fatalf("ShiftStance: %v", err)
	}
	if shifted.ActiveStance != "Reliable" || shifted.Focus != focused.Focus-1 {
		t.Fatalf("shifting should spend a Focus: %+v", shifted)
	}

	hit, err := s.TakeDamage(DamageInput{CharacterID: c.ID, Amount: 20})
	if err != nil {
		t.Fatalf("TakeDamage: %v", err)
	}
	if hit.Character.ActiveStance != "" || len(hit.Ended) != 1 {
		t.Fatalf("Severe damage drops the stance: %+v", hit)
	}
}

func TestAnchoredStanceAddsToThresholdsWhileActive(t *testing.T) {
	s := newTestService(t)
	c := characterWith(t, s, "brawler", "martialArtist", nil, 5, 10)
	setColumn(t, s, c.ID, "level", 2)
	setColumn(t, s, c.ID, "focus", 2)

	if _, err := s.LearnStance(c.ID, "Anchored", true); err != nil {
		t.Fatalf("LearnStance: %v", err)
	}
	in, err := s.ShiftStance(c.ID, "Anchored", "focus")
	if err != nil {
		t.Fatalf("ShiftStance: %v", err)
	}
	if in.ThresholdMajor != 5 || in.EffectiveMajor != 7 || in.EffectiveSevere != 12 {
		t.Fatalf("Anchored adds +2 on top of the stored thresholds: %+v", in)
	}
	hit, err := s.TakeDamage(DamageInput{CharacterID: c.ID, Amount: 6})
	if err != nil {
		t.Fatalf("TakeDamage: %v", err)
	}
	if hit.Severity != "minor" {
		t.Fatalf("6 is under the anchored Major of 7, so it's Minor: %+v", hit)
	}
}

func TestToggledEffects(t *testing.T) {
	s := newTestService(t)
	rogue := characterWith(t, s, "rogue", "nightwalker", nil, 5, 10)

	if _, err := s.SetEffect(rogue.ID, "patronsMantle", true); err == nil {
		t.Error("a rogue has no Patron's Mantle")
	}
	dodging, err := s.SetEffect(rogue.ID, "roguesDodge", true)
	if err != nil {
		t.Fatalf("SetEffect: %v", err)
	}
	if dodging.EffectiveEvasion != dodging.Evasion+2 {
		t.Fatalf("Rogue's Dodge is +2 Evasion: %+v", dodging)
	}
	hit, err := s.TakeDamage(DamageInput{CharacterID: rogue.ID, Amount: 1})
	if err != nil {
		t.Fatalf("TakeDamage: %v", err)
	}
	if len(hit.Character.ActiveEffects) != 0 || hit.Character.EffectiveEvasion != hit.Character.Evasion {
		t.Fatalf("an attack that lands ends the dodge: %+v", hit.Character)
	}

	warlock := characterWith(t, s, "warlock", "pactOfTheEndless", nil, 5, 10)
	mantled, err := s.SetEffect(warlock.ID, "patronsMantle", true)
	if err != nil {
		t.Fatalf("SetEffect: %v", err)
	}
	if mantled.EffectiveMajor != 6 || mantled.EffectiveSevere != 11 {
		t.Fatalf("Patron's Mantle adds the tier: %+v", mantled)
	}
	minor, _ := s.TakeDamage(DamageInput{CharacterID: warlock.ID, Amount: 3})
	if len(minor.Character.ActiveEffects) != 1 {
		t.Fatal("Minor damage leaves the mantle up")
	}
	severe, _ := s.TakeDamage(DamageInput{CharacterID: warlock.ID, Amount: 11})
	if severe.Severity != "severe" || len(severe.Character.ActiveEffects) != 0 {
		t.Fatalf("Severe damage ends the mantle: %+v", severe)
	}
}

func TestEffectGroupsAndArmorBlock(t *testing.T) {
	s := newTestService(t)
	druid := characterWith(t, s, "druid", "wardenOfTheElements", nil, 5, 10)
	setColumn(t, s, druid.ID, "subclass_mastery", 3)

	if _, err := s.SetEffect(druid.ID, "channelEarth", true); err != nil {
		t.Fatalf("SetEffect: %v", err)
	}
	air, err := s.SetEffect(druid.ID, "channelAir", true)
	if err != nil {
		t.Fatalf("SetEffect: %v", err)
	}
	if len(air.ActiveEffects) != 1 || air.ActiveEffects[0] != "channelAir" {
		t.Fatalf("channeling one element stops channeling the other: %v", air.ActiveEffects)
	}

	c := newTestCharacter(t, s, 3, 5, 10)
	if _, err := s.AddDomainCard(c.ID, "frenzy", "loadout"); err != nil {
		t.Fatalf("AddDomainCard: %v", err)
	}
	if _, err := s.SetEffect(c.ID, "frenzy", true); err != nil {
		t.Fatalf("SetEffect: %v", err)
	}
	if _, err := s.TakeDamage(DamageInput{CharacterID: c.ID, Amount: 6, UseArmor: true}); err == nil {
		t.Error("a Frenzied character can't use Armor Slots")
	}
	rested, err := s.Rest(RestInput{CharacterID: c.ID, Moves: []string{"prepare"}})
	if err != nil {
		t.Fatalf("Rest: %v", err)
	}
	if len(rested.Character.ActiveEffects) != 0 {
		t.Fatal("a rest ends the Frenzy")
	}
}

func TestLiveBonusesFollowStressAndHope(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 5, 10)
	if _, err := s.AddDomainCard(c.ID, "eldritchFlesh", "loadout"); err != nil {
		t.Fatalf("AddDomainCard: %v", err)
	}
	stressed, err := s.SetVitals(c.ID, 0, 3, c.Hope, 0)
	if err != nil {
		t.Fatalf("SetVitals: %v", err)
	}
	if stressed.ThresholdMajor != 5 || stressed.EffectiveMajor != 8 || stressed.EffectiveSevere != 13 {
		t.Fatalf("Eldritch Flesh adds marked Stress to the thresholds: %+v", stressed)
	}

	wizard := characterWith(t, s, "wizard", "schoolOfWar", nil, 5, 10)
	setColumn(t, s, wizard.ID, "subclass_mastery", 2)
	setColumn(t, s, wizard.ID, "proficiency", 2)
	shielded, _ := s.SetVitals(wizard.ID, 0, 0, 2, 0)
	if shielded.EffectiveEvasion != shielded.Evasion+2 {
		t.Fatalf("Conjure Shield adds Proficiency with 2 Hope: %+v", shielded)
	}
	bare, _ := s.SetVitals(wizard.ID, 0, 0, 1, 0)
	if bare.EffectiveEvasion != bare.Evasion {
		t.Fatalf("Conjure Shield drops below 2 Hope: %+v", bare)
	}
}

func TestBrawlerCombat(t *testing.T) {
	s := newTestService(t)
	c := characterWith(t, s, "brawler", "juggernaut", nil, 5, 10)

	strike, err := s.RollBrawlersStrike(c.ID, false)
	if err != nil {
		t.Fatalf("RollBrawlersStrike: %v", err)
	}
	if strike.Total < 2 || strike.Total > 14 {
		t.Fatalf("1d8+1d6 at Proficiency 1: %+v", strike)
	}
	crit, _ := s.RollBrawlersStrike(c.ID, true)
	if crit.Total < 16 || crit.Total > 28 {
		t.Fatalf("a critical adds 8+6: %+v", crit)
	}

	combo, err := s.RollCombo(c.ID, "stress")
	if err != nil {
		t.Fatalf("RollCombo: %v", err)
	}
	n := len(combo.Rolls)
	if combo.Die != 4 || n < 2 || combo.Rolls[n-1] >= combo.Rolls[n-2] || combo.Character.StressMarked != 1 {
		t.Fatalf("a combo stops on a lower roll and costs a Stress: %+v", combo)
	}
	sum := 0
	for _, r := range combo.Rolls {
		sum += r
	}
	if combo.Total != sum {
		t.Fatalf("total %d isn't the sum of %v", combo.Total, combo.Rolls)
	}
	if _, err := s.RollCombo(c.ID, "focus"); err == nil {
		t.Error("spending Focus on a combo takes Flow State")
	}

	addGear(t, s, c.ID, "primary-weapon", "", "Club", true)
	if _, err := s.RollBrawlersStrike(c.ID, false); err == nil {
		t.Error("Brawler's Strike is put away while a weapon is equipped")
	}

	plan, err := s.PlanLevelUp(c.ID)
	if err != nil {
		t.Fatalf("PlanLevelUp: %v", err)
	}
	found := false
	for _, opt := range plan.Options {
		if opt.ID == "comboDie" && opt.Available && opt.Effect.ComboDie == 1 {
			found = true
		}
	}
	if !found {
		t.Fatal("a brawler's level-up offers the Combo Die")
	}
	bard := newTestCharacter(t, s, 0, 1, 2)
	bardPlan, _ := s.PlanLevelUp(bard.ID)
	for _, opt := range bardPlan.Options {
		if opt.ID == "comboDie" {
			t.Fatal("only a brawler has a Combo Die")
		}
	}
}
