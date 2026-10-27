package player

import (
	"testing"

	"github.com/mhetem/DH-Companion/internal/db"
)

func TestSubclassAbilitiesFollowMastery(t *testing.T) {
	s := newTestService(t)
	for _, tc := range []struct {
		mastery int64
		want    statMods
	}{
		{1, statMods{major: 1, severe: 1}},
		{2, statMods{major: 3, severe: 3}},
		{3, statMods{major: 6, severe: 6}},
	} {
		row := db.Character{ClassSlug: "guardian", SubclassSlug: "stalwart", SubclassMastery: tc.mastery, Proficiency: 1}
		if got := s.snapshotOf(row, nil, nil).mods; got != tc.want {
			t.Errorf("stalwart at mastery %d: %+v, want %+v", tc.mastery, got, tc.want)
		}
	}

	multi := db.Character{
		ClassSlug: "bard", SubclassSlug: "troubadour", SubclassMastery: 3, Proficiency: 1,
		MulticlassSlug: "guardian", MulticlassSubclassSlug: "stalwart",
	}
	if got := s.snapshotOf(multi, nil, nil).mods; got != (statMods{major: 1, severe: 1}) {
		t.Errorf("a multiclass subclass should only grant its foundation: %+v", got)
	}
}

func TestLoadoutCardAbilities(t *testing.T) {
	s := newTestService(t)
	row := db.Character{ClassSlug: "bard", SubclassSlug: "troubadour", SubclassMastery: 1, Proficiency: 2, Agility: 3}
	loadout := func(slugs ...string) []db.CharacterDomainCard {
		out := []db.CharacterDomainCard{}
		for _, slug := range slugs {
			out = append(out, db.CharacterDomainCard{CardSlug: slug, Location: "loadout"})
		}
		return out
	}
	armor := []db.InventoryItem{{ID: 1, Kind: "armor", Name: "Mail", Equipped: 1}}

	for _, tc := range []struct {
		name  string
		items []db.InventoryItem
		held  []db.CharacterDomainCard
		want  statMods
	}{
		{"fortified armor unarmored", nil, loadout("fortifiedArmor"), statMods{}},
		{"fortified armor armored", armor, loadout("fortifiedArmor"), statMods{major: 2, severe: 2}},
		{"armorer armored", armor, loadout("armorer"), statMods{armorScore: 1}},
		{"rise up", nil, loadout("riseUp"), statMods{severe: 2}},
		{"untouchable", nil, loadout("untouchable"), statMods{evasion: 1}},
		{"blade-touched with three", nil, loadout("bladeTouched", "getBackUp", "whirlwind"), statMods{}},
		{"blade-touched with four", nil, loadout("bladeTouched", "getBackUp", "whirlwind", "reckless"), statMods{severe: 4}},
		{"vaulted cards do nothing", nil, []db.CharacterDomainCard{{CardSlug: "untouchable", Location: "vault"}}, statMods{}},
	} {
		if got := s.snapshotOf(row, tc.items, tc.held).mods; got != tc.want {
			t.Errorf("%s: %+v, want %+v", tc.name, got, tc.want)
		}
	}
}

func TestShellFollowsProficiency(t *testing.T) {
	s := newTestService(t)
	row := db.Character{ClassSlug: "bard", SubclassSlug: "troubadour", SubclassMastery: 1, Proficiency: 3, AncestrySlug: "galapa"}
	snap := s.snapshotOf(row, nil, nil)
	if snap.mods != (statMods{major: 3, severe: 3}) {
		t.Fatalf("Shell at Proficiency 3: %+v", snap.mods)
	}
	if len(snap.bonuses) != 1 || snap.bonuses[0].Source != "Shell" || snap.bonuses[0].Effect != "+3 to damage thresholds" {
		t.Fatalf("unexpected bonus list %+v", snap.bonuses)
	}
}

func TestArmorKeepsAbilityBonuses(t *testing.T) {
	s := newTestService(t)
	c := buildCharacter(t, s, "guardian", "stalwart", "clank", 0, 1, 2)

	addGear(t, s, c.ID, "armor", "chainmailArmor", "Chainmail Armor", true)
	expectDefenses(t, s, c.ID, 9, 4, 9, 17)

	sheet, err := s.GetSheet(c.ID)
	if err != nil {
		t.Fatalf("GetSheet: %v", err)
	}
	sources := map[string]string{}
	for _, b := range sheet.Bonuses {
		sources[b.Source] = b.Effect
	}
	if sources["Unwavering"] != "+1 to damage thresholds" || sources["Chainmail Armor"] != "−1 Evasion" {
		t.Fatalf("the sheet should list what it applied: %+v", sheet.Bonuses)
	}
}

func TestCardsMoveStatsAsTheyComeAndGo(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 1, 2)

	if _, err := s.AddDomainCard(c.ID, "untouchable", "loadout"); err != nil {
		t.Fatalf("AddDomainCard: %v", err)
	}
	expectDefenses(t, s, c.ID, 11, 0, 1, 2)

	if _, err := s.AddDomainCard(c.ID, "fortifiedArmor", "loadout"); err != nil {
		t.Fatalf("AddDomainCard: %v", err)
	}
	expectDefenses(t, s, c.ID, 11, 0, 1, 2)

	addGear(t, s, c.ID, "armor", "chainmailArmor", "Chainmail Armor", true)
	expectDefenses(t, s, c.ID, 10, 4, 10, 18)

	if _, err := s.MoveDomainCard(c.ID, "fortifiedArmor", "vault"); err != nil {
		t.Fatalf("MoveDomainCard: %v", err)
	}
	expectDefenses(t, s, c.ID, 10, 4, 8, 16)

	if _, err := s.RemoveDomainCard(c.ID, "untouchable"); err != nil {
		t.Fatalf("RemoveDomainCard: %v", err)
	}
	expectDefenses(t, s, c.ID, 9, 4, 8, 16)
}

func TestSlotChangesClampMarks(t *testing.T) {
	s := newTestService(t)
	c := newTestCharacter(t, s, 0, 1, 2)

	if err := s.syncStats(s.q, c.ID, snapshot{}, snapshot{mods: statMods{hpMax: 1, stressMax: 1}}); err != nil {
		t.Fatalf("syncStats: %v", err)
	}
	grown, _ := s.GetCharacter(c.ID)
	if grown.HPMax != c.HPMax+1 || grown.StressMax != c.StressMax+1 {
		t.Fatalf("slots should grow by one: %d/%d from %d/%d", grown.HPMax, grown.StressMax, c.HPMax, c.StressMax)
	}

	if _, err := s.SetVitals(c.ID, grown.HPMax, grown.StressMax, grown.Hope, 0); err != nil {
		t.Fatalf("SetVitals: %v", err)
	}
	if err := s.syncStats(s.q, c.ID, snapshot{mods: statMods{hpMax: 1, stressMax: 1}}, snapshot{}); err != nil {
		t.Fatalf("syncStats: %v", err)
	}
	shrunk, _ := s.GetCharacter(c.ID)
	if shrunk.HPMarked != c.HPMax || shrunk.StressMarked != c.StressMax {
		t.Fatalf("marks should follow the slots down: %d/%d marked of %d/%d", shrunk.HPMarked, shrunk.StressMarked, shrunk.HPMax, shrunk.StressMax)
	}
}

func TestBrawlerFightsBarehanded(t *testing.T) {
	s := newTestService(t)
	row := db.Character{ClassSlug: "brawler", SubclassSlug: "juggernaut", SubclassMastery: 1, Proficiency: 1}
	sword := []db.InventoryItem{{ID: 1, Kind: "primary-weapon", Name: "Sword", Equipped: 1}}

	if got := s.snapshotOf(row, nil, nil).mods; got != (statMods{evasion: 1, severe: 3}) {
		t.Errorf("an unarmed Juggernaut has I Am the Weapon and Rugged: %+v", got)
	}
	if got := s.snapshotOf(row, sword, nil).mods; got != (statMods{severe: 3}) {
		t.Errorf("a weapon in hand puts Brawler's Strike away: %+v", got)
	}

	c := buildCharacter(t, s, "brawler", "martialArtist", "clank", 0, 1, 2)
	addGear(t, s, c.ID, "primary-weapon", "", "Sword", true)
	expectDefenses(t, s, c.ID, 9, 0, 1, 2)

	creation := s.CreationBonuses("brawler", "martialArtist", "clank")
	if creation.Evasion != 1 {
		t.Errorf("a new Brawler starts barehanded, with the Evasion that brings: %+v", creation)
	}
}

func TestCreationBonuses(t *testing.T) {
	s := newTestService(t)

	simiah := s.CreationBonuses("guardian", "stalwart", "simiah")
	if simiah.Evasion != 1 || simiah.Major != 1 || simiah.Severe != 1 || len(simiah.Sources) != 2 {
		t.Errorf("a Simiah Stalwart starts with Nimble and Unwavering: %+v", simiah)
	}
	human := s.CreationBonuses("wizard", "schoolOfWar", "human")
	if human.HPMax != 1 || human.StressMax != 1 {
		t.Errorf("a Human Battlemage starts with a Hit Point and a Stress slot more: %+v", human)
	}
	plain := s.CreationBonuses("bard", "troubadour", "clank")
	if plain.Sources == nil || len(plain.Sources) != 0 || plain.Evasion != 0 {
		t.Errorf("a Clank Troubadour has nothing to add: %+v", plain)
	}
}
