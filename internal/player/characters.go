package player

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mhetem/DH-Companion/internal/cards"
	"github.com/mhetem/DH-Companion/internal/db"
)

type Gold struct {
	Handfuls int `json:"handfuls"`
	Bags     int `json:"bags"`
	Chests   int `json:"chests"`
}

type Character struct {
	ID                     int64          `json:"id"`
	Name                   string         `json:"name"`
	Pronouns               string         `json:"pronouns"`
	ClassSlug              string         `json:"classSlug"`
	ClassName              string         `json:"className"`
	SubclassSlug           string         `json:"subclassSlug"`
	SubclassName           string         `json:"subclassName"`
	SubclassMastery        int            `json:"subclassMastery"`
	MulticlassSlug         string         `json:"multiclassSlug"`
	MulticlassName         string         `json:"multiclassName"`
	MulticlassSubclassSlug string         `json:"multiclassSubclassSlug"`
	MulticlassSubclassName string         `json:"multiclassSubclassName"`
	AncestrySlug           string         `json:"ancestrySlug"`
	AncestryName           string         `json:"ancestryName"`
	CommunitySlug          string         `json:"communitySlug"`
	CommunityName          string         `json:"communityName"`
	Domains                []string       `json:"domains"`
	Level                  int            `json:"level"`
	Tier                   int            `json:"tier"`
	Proficiency            int            `json:"proficiency"`
	Traits                 map[string]int `json:"traits"`
	MarkedTraits           []string       `json:"markedTraits"`
	SpellcastTrait         string         `json:"spellcastTrait"`
	BeastformSlug          string         `json:"beastformSlug"`
	BeastformName          string         `json:"beastformName"`
	HPMax                  int            `json:"hpMax"`
	HPMarked               int            `json:"hpMarked"`
	StressMax              int            `json:"stressMax"`
	StressMarked           int            `json:"stressMarked"`
	Hope                   int            `json:"hope"`
	HopeMax                int            `json:"hopeMax"`
	Evasion                int            `json:"evasion"`
	ArmorScore             int            `json:"armorScore"`
	ArmorMarked            int            `json:"armorMarked"`
	ThresholdMajor         int            `json:"thresholdMajor"`
	ThresholdSevere        int            `json:"thresholdSevere"`
	Gold                   Gold           `json:"gold"`
	Background             string         `json:"background"`
	Connections            string         `json:"connections"`
	Notes                  string         `json:"notes"`
	CreatedAt              string         `json:"createdAt"`
	UpdatedAt              string         `json:"updatedAt"`
}

type Sheet struct {
	Character
	Class              *cards.CharacterClass `json:"class"`
	Subclass           *cards.Subclass       `json:"subclass"`
	Multiclass         *cards.CharacterClass `json:"multiclass"`
	MulticlassSubclass *cards.Subclass       `json:"multiclassSubclass"`
	Ancestry           *cards.Ancestry       `json:"ancestry"`
	Community          *cards.Community      `json:"community"`
	Loadout            []DomainCard          `json:"loadout"`
	Vault              []DomainCard          `json:"vault"`
	Inventory          []Item                `json:"inventory"`
	Experiences        []Experience          `json:"experiences"`
	LoadoutMax         int                   `json:"loadoutMax"`
	LevelUps           []LevelUpRecord       `json:"levelUps"`
	CanLevelUp         bool                  `json:"canLevelUp"`
}

type CharacterInput struct {
	ID              *int64         `json:"id"`
	Name            string         `json:"name"`
	Pronouns        string         `json:"pronouns"`
	ClassSlug       string         `json:"classSlug"`
	SubclassSlug    string         `json:"subclassSlug"`
	AncestrySlug    string         `json:"ancestrySlug"`
	CommunitySlug   string         `json:"communitySlug"`
	Traits          map[string]int `json:"traits"`
	HPMax           int            `json:"hpMax"`
	StressMax       int            `json:"stressMax"`
	Evasion         int            `json:"evasion"`
	ArmorScore      int            `json:"armorScore"`
	ThresholdMajor  int            `json:"thresholdMajor"`
	ThresholdSevere int            `json:"thresholdSevere"`
	Background      string         `json:"background"`
	Connections     string         `json:"connections"`
	Notes           string         `json:"notes"`
}

func (s *Service) characterView(r db.Character) Character {
	c := Character{
		ID:                     r.ID,
		Name:                   r.Name,
		Pronouns:               r.Pronouns,
		ClassSlug:              r.ClassSlug,
		SubclassSlug:           r.SubclassSlug,
		SubclassMastery:        int(r.SubclassMastery),
		MulticlassSlug:         r.MulticlassSlug,
		MulticlassSubclassSlug: r.MulticlassSubclassSlug,
		AncestrySlug:           r.AncestrySlug,
		CommunitySlug:          r.CommunitySlug,
		Domains:                []string{},
		Level:                  int(r.Level),
		Tier:                   tierForLevel(int(r.Level)),
		Proficiency:            int(r.Proficiency),
		Traits: map[string]int{
			"agility":   int(r.Agility),
			"strength":  int(r.Strength),
			"finesse":   int(r.Finesse),
			"instinct":  int(r.Instinct),
			"presence":  int(r.Presence),
			"knowledge": int(r.Knowledge),
		},
		MarkedTraits:    decodeTraitList(r.MarkedTraits),
		BeastformSlug:   r.BeastformSlug,
		HPMax:           int(r.HpMax),
		HPMarked:        int(r.HpMarked),
		StressMax:       int(r.StressMax),
		StressMarked:    int(r.StressMarked),
		Hope:            int(r.Hope),
		HopeMax:         HopeMax,
		Evasion:         int(r.Evasion),
		ArmorScore:      int(r.ArmorScore),
		ArmorMarked:     int(r.ArmorMarked),
		ThresholdMajor:  int(r.ThresholdMajor),
		ThresholdSevere: int(r.ThresholdSevere),
		Gold: Gold{
			Handfuls: int(r.GoldHandfuls),
			Bags:     int(r.GoldBags),
			Chests:   int(r.GoldChests),
		},
		Background:  r.Background,
		Connections: r.Connections,
		Notes:       r.Notes,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	if class, ok := s.catalog.Class(r.ClassSlug); ok {
		c.ClassName = class.Name
		c.Domains = append(c.Domains, class.Domains...)
		if sub, ok := class.Subclass(r.SubclassSlug); ok {
			c.SubclassName = sub.Name
			c.SpellcastTrait = sub.SpellcastTrait
		}
	}
	if r.MulticlassSlug != "" {
		if class, ok := s.catalog.Class(r.MulticlassSlug); ok {
			c.MulticlassName = class.Name
			for _, d := range class.Domains {
				if !contains(c.Domains, d) {
					c.Domains = append(c.Domains, d)
				}
			}
			if sub, ok := class.Subclass(r.MulticlassSubclassSlug); ok {
				c.MulticlassSubclassName = sub.Name
			}
		}
	}
	if r.BeastformSlug != "" {
		if form, ok := s.catalog.Beastform(r.BeastformSlug); ok {
			c.BeastformName = form.Name
		}
	}
	if a, ok := s.catalog.Ancestry(r.AncestrySlug); ok {
		c.AncestryName = a.Name
	}
	if m, ok := s.catalog.Community(r.CommunitySlug); ok {
		c.CommunityName = m.Name
	}
	return c
}

func (s *Service) ListCharacters() ([]Character, error) {
	rows, err := s.q.ListCharacters(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing characters: %w", err)
	}
	out := make([]Character, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.characterView(r))
	}
	return out, nil
}

func (s *Service) GetCharacter(id int64) (Character, error) {
	row, err := s.character(id)
	if err != nil {
		return Character{}, err
	}
	return s.characterView(row), nil
}

func (s *Service) GetSheet(id int64) (Sheet, error) {
	row, err := s.character(id)
	if err != nil {
		return Sheet{}, err
	}

	sheet := Sheet{Character: s.characterView(row), LoadoutMax: LoadoutMax}

	if class, ok := s.catalog.Class(row.ClassSlug); ok {
		sheet.Class = &class
		if sub, ok := class.Subclass(row.SubclassSlug); ok {
			sheet.Subclass = &sub
		}
	}
	if row.MulticlassSlug != "" {
		if class, ok := s.catalog.Class(row.MulticlassSlug); ok {
			sheet.Multiclass = &class
			if sub, ok := class.Subclass(row.MulticlassSubclassSlug); ok {
				sheet.MulticlassSubclass = &sub
			}
		}
	}
	if a, ok := s.catalog.Ancestry(row.AncestrySlug); ok {
		sheet.Ancestry = &a
	}
	if m, ok := s.catalog.Community(row.CommunitySlug); ok {
		sheet.Community = &m
	}

	loadout, vault, err := s.domainCards(id)
	if err != nil {
		return Sheet{}, err
	}
	sheet.Loadout, sheet.Vault = loadout, vault

	if sheet.Inventory, err = s.ListInventory(id); err != nil {
		return Sheet{}, err
	}
	if sheet.Experiences, err = s.ListExperiences(id); err != nil {
		return Sheet{}, err
	}
	if sheet.LevelUps, err = s.ListLevelUps(id); err != nil {
		return Sheet{}, err
	}
	sheet.CanLevelUp = sheet.Level < MaxLevel

	return sheet, nil
}

func (s *Service) SaveCharacter(in CharacterInput) (Character, error) {
	name, err := validateName(in.Name)
	if err != nil {
		return Character{}, err
	}
	if err := s.validateBuild(in.ClassSlug, in.SubclassSlug, in.AncestrySlug, in.CommunitySlug); err != nil {
		return Character{}, err
	}

	traits := normalizeTraits(in.Traits)
	class, _ := s.catalog.Class(strings.TrimSpace(in.ClassSlug))

	hpMax := in.HPMax
	if hpMax <= 0 {
		hpMax = leadingInt(class.StartingHitPoints, 6)
	}
	stressMax := in.StressMax
	if stressMax <= 0 {
		stressMax = 6
	}
	evasion := in.Evasion
	if evasion <= 0 {
		evasion = leadingInt(class.StartingEvasion, 10)
	}

	var row db.Character
	if in.ID == nil {
		row, err = s.q.CreateCharacter(s.ctx, db.CreateCharacterParams{
			Name:            name,
			Pronouns:        strings.TrimSpace(in.Pronouns),
			ClassSlug:       strings.TrimSpace(in.ClassSlug),
			SubclassSlug:    strings.TrimSpace(in.SubclassSlug),
			AncestrySlug:    strings.TrimSpace(in.AncestrySlug),
			CommunitySlug:   strings.TrimSpace(in.CommunitySlug),
			Agility:         int64(traits["agility"]),
			Strength:        int64(traits["strength"]),
			Finesse:         int64(traits["finesse"]),
			Instinct:        int64(traits["instinct"]),
			Presence:        int64(traits["presence"]),
			Knowledge:       int64(traits["knowledge"]),
			HpMax:           int64(clamp(hpMax, 1, 20)),
			StressMax:       int64(clamp(stressMax, 1, 20)),
			Hope:            int64(StartingHope),
			Evasion:         int64(clamp(evasion, 0, 30)),
			ArmorScore:      int64(clamp(in.ArmorScore, 0, 20)),
			ThresholdMajor:  int64(clamp(in.ThresholdMajor, 0, 99)),
			ThresholdSevere: int64(clamp(in.ThresholdSevere, 0, 99)),
			Background:      strings.TrimSpace(in.Background),
			Connections:     strings.TrimSpace(in.Connections),
		})
		if err != nil {
			return Character{}, fmt.Errorf("creating character: %w", err)
		}
	} else {
		row, err = s.q.UpdateCharacter(s.ctx, db.UpdateCharacterParams{
			Name:            name,
			Pronouns:        strings.TrimSpace(in.Pronouns),
			ClassSlug:       strings.TrimSpace(in.ClassSlug),
			SubclassSlug:    strings.TrimSpace(in.SubclassSlug),
			AncestrySlug:    strings.TrimSpace(in.AncestrySlug),
			CommunitySlug:   strings.TrimSpace(in.CommunitySlug),
			Agility:         int64(traits["agility"]),
			Strength:        int64(traits["strength"]),
			Finesse:         int64(traits["finesse"]),
			Instinct:        int64(traits["instinct"]),
			Presence:        int64(traits["presence"]),
			Knowledge:       int64(traits["knowledge"]),
			HpMax:           int64(clamp(hpMax, 1, 20)),
			StressMax:       int64(clamp(stressMax, 1, 20)),
			Evasion:         int64(clamp(evasion, 0, 30)),
			ArmorScore:      int64(clamp(in.ArmorScore, 0, 20)),
			ThresholdMajor:  int64(clamp(in.ThresholdMajor, 0, 99)),
			ThresholdSevere: int64(clamp(in.ThresholdSevere, 0, 99)),
			Background:      strings.TrimSpace(in.Background),
			Connections:     strings.TrimSpace(in.Connections),
			Notes:           strings.TrimSpace(in.Notes),
			ID:              *in.ID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return Character{}, notFound("character", fmt.Sprint(*in.ID))
		}
		if err != nil {
			return Character{}, fmt.Errorf("updating character: %w", err)
		}
	}
	return s.characterView(row), nil
}

func (s *Service) DeleteCharacter(id int64) error {
	if err := s.q.DeleteCharacter(s.ctx, id); err != nil {
		return fmt.Errorf("deleting character: %w", err)
	}
	return nil
}

func (s *Service) MarkHP(id int64, delta int) (Character, error) {
	row, err := s.q.AdjustCharacterHP(s.ctx, db.AdjustCharacterHPParams{Delta: int64(delta), ID: id})
	return s.vitalsResult(id, row, err, "marking hit points")
}

func (s *Service) MarkStress(id int64, delta int) (Character, error) {
	row, err := s.q.AdjustCharacterStress(s.ctx, db.AdjustCharacterStressParams{Delta: int64(delta), ID: id})
	return s.vitalsResult(id, row, err, "marking stress")
}

func (s *Service) AdjustHope(id int64, delta int) (Character, error) {
	row, err := s.q.AdjustCharacterHope(s.ctx, db.AdjustCharacterHopeParams{Delta: int64(delta), ID: id})
	return s.vitalsResult(id, row, err, "adjusting hope")
}

func (s *Service) MarkArmor(id int64, delta int) (Character, error) {
	row, err := s.q.AdjustCharacterArmor(s.ctx, db.AdjustCharacterArmorParams{Delta: int64(delta), ID: id})
	return s.vitalsResult(id, row, err, "marking armor")
}

func (s *Service) SetVitals(id int64, hpMarked, stressMarked, hope, armorMarked int) (Character, error) {
	row, err := s.q.SetCharacterVitals(s.ctx, db.SetCharacterVitalsParams{
		HpMarked:     int64(hpMarked),
		StressMarked: int64(stressMarked),
		Hope:         int64(hope),
		ArmorMarked:  int64(armorMarked),
		ID:           id,
	})
	return s.vitalsResult(id, row, err, "setting vitals")
}

func (s *Service) SetGold(id int64, handfuls, bags, chests int) (Character, error) {
	row, err := s.q.SetCharacterGold(s.ctx, db.SetCharacterGoldParams{
		GoldHandfuls: int64(handfuls),
		GoldBags:     int64(bags),
		GoldChests:   int64(chests),
		ID:           id,
	})
	return s.vitalsResult(id, row, err, "setting gold")
}

func (s *Service) vitalsResult(id int64, row db.Character, err error, what string) (Character, error) {
	if errors.Is(err, sql.ErrNoRows) {
		return Character{}, notFound("character", fmt.Sprint(id))
	}
	if err != nil {
		return Character{}, fmt.Errorf("%s: %w", what, err)
	}
	return s.characterView(row), nil
}

func (s *Service) character(id int64) (db.Character, error) {
	row, err := s.q.GetCharacter(s.ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return db.Character{}, notFound("character", fmt.Sprint(id))
	}
	if err != nil {
		return db.Character{}, fmt.Errorf("loading character: %w", err)
	}
	return row, nil
}

func normalizeTraits(in map[string]int) map[string]int {
	out := map[string]int{}
	for _, k := range TraitKeys {
		out[k] = clamp(in[k], -3, 12)
	}
	return out
}

func decodeTraitList(raw string) []string {
	out := []string{}
	if strings.TrimSpace(raw) == "" {
		return out
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return []string{}
	}
	return out
}

func encodeTraitList(in []string) string {
	if in == nil {
		in = []string{}
	}
	b, err := json.Marshal(in)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func leadingInt(raw string, fallback int) int {
	digits := strings.Builder{}
	for _, r := range strings.TrimSpace(raw) {
		if r < '0' || r > '9' {
			break
		}
		digits.WriteRune(r)
	}
	if digits.Len() == 0 {
		return fallback
	}
	n, err := strconv.Atoi(digits.String())
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
