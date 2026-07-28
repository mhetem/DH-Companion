package srd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strconv"

	"github.com/mhetem/DH-Companion/data"
	"github.com/mhetem/DH-Companion/internal/cards"
)

const (
	adversariesFile  = "adversaries.json"
	environmentsFile = "environments.json"
	domainCardsFile  = "domainCards.json"
	ancestriesFile   = "ancestries.json"
	communitiesFile  = "communities.json"
	classesFile      = "classes.json"
	beastformsFile   = "beastforms.json"
	companionsFile   = "rangerCompanion.json"
)

type Catalog struct {
	Adversaries  map[string]cards.Adversary
	Environments map[string]cards.Environment
	DomainCards  map[string]cards.DomainCard
	Ancestries   map[string]cards.Ancestry
	Communities  map[string]cards.Community
	Classes      map[string]cards.Class
	Beastforms   map[string]cards.Beastform
	Companions   map[string]cards.Companion
}

func Load(fsys fs.FS) (*Catalog, error) {
	c := &Catalog{
		Adversaries:  map[string]cards.Adversary{},
		Environments: map[string]cards.Environment{},
		DomainCards:  map[string]cards.DomainCard{},
		Ancestries:   map[string]cards.Ancestry{},
		Communities:  map[string]cards.Community{},
		Classes:      map[string]cards.Class{},
		Beastforms:   map[string]cards.Beastform{},
		Companions:   map[string]cards.Companion{},
	}

	if err := decodeFile(fsys, adversariesFile, &c.Adversaries); err != nil {
		return nil, err
	}
	for slug, a := range c.Adversaries {
		a.Kind = cards.KindAdversary
		a.Slug = slug
		c.Adversaries[slug] = a
	}

	if err := decodeFile(fsys, environmentsFile, &c.Environments); err != nil {
		return nil, err
	}
	for slug, e := range c.Environments {
		e.Kind = cards.KindEnvironment
		e.Slug = slug
		c.Environments[slug] = e
	}

	if err := decodeFile(fsys, domainCardsFile, &c.DomainCards); err != nil {
		return nil, err
	}
	for slug, d := range c.DomainCards {
		d.Kind = cards.KindDomain
		d.Slug = slug
		c.DomainCards[slug] = d
	}

	if err := decodeFile(fsys, ancestriesFile, &c.Ancestries); err != nil {
		return nil, err
	}
	for slug, a := range c.Ancestries {
		a.Kind = cards.KindAncestry
		a.Slug = slug
		c.Ancestries[slug] = a
	}

	if err := decodeFile(fsys, communitiesFile, &c.Communities); err != nil {
		return nil, err
	}
	for slug, m := range c.Communities {
		m.Kind = cards.KindCommunity
		m.Slug = slug
		c.Communities[slug] = m
	}

	if err := decodeFile(fsys, classesFile, &c.Classes); err != nil {
		return nil, err
	}
	for slug, k := range c.Classes {
		k.Kind = cards.KindClass
		k.Slug = slug
		c.Classes[slug] = k
	}

	if err := decodeFile(fsys, beastformsFile, &c.Beastforms); err != nil {
		return nil, err
	}
	for slug, b := range c.Beastforms {
		b.Kind = cards.KindBeastform
		b.Slug = slug
		c.Beastforms[slug] = b
	}

	if err := decodeFile(fsys, companionsFile, &c.Companions); err != nil {
		return nil, err
	}
	for slug, p := range c.Companions {
		p.Kind = cards.KindCompanion
		p.Slug = slug
		c.Companions[slug] = p
	}

	return c, nil
}

func Default() (*Catalog, error) { return Load(data.FS) }

func decodeFile(fsys fs.FS, name string, dst any) error {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("srd: reading %s: %w", name, err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return fmt.Errorf("srd: parsing %s: %w", name, err)
	}
	return nil
}

func (c *Catalog) Adversary(slug string) (cards.Adversary, bool) {
	a, ok := c.Adversaries[slug]
	return a, ok
}

func (c *Catalog) Environment(slug string) (cards.Environment, bool) {
	e, ok := c.Environments[slug]
	return e, ok
}

func (c *Catalog) ListAdversaries() []cards.Adversary {
	out := make([]cards.Adversary, 0, len(c.Adversaries))
	for _, a := range c.Adversaries {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (c *Catalog) ListEnvironments() []cards.Environment {
	out := make([]cards.Environment, 0, len(c.Environments))
	for _, e := range c.Environments {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Tier != out[j].Tier {
			return out[i].Tier < out[j].Tier
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func (c *Catalog) DomainCard(slug string) (cards.DomainCard, bool) {
	d, ok := c.DomainCards[slug]
	return d, ok
}

// ListDomainCards returns every domain card sorted the way they're printed:
// by domain, then by level, then by name.
func (c *Catalog) ListDomainCards() []cards.DomainCard {
	out := make([]cards.DomainCard, 0, len(c.DomainCards))
	for _, d := range c.DomainCards {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Domain != out[j].Domain {
			return out[i].Domain < out[j].Domain
		}
		li, _ := strconv.Atoi(out[i].Level)
		lj, _ := strconv.Atoi(out[j].Level)
		if li != lj {
			return li < lj
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func (c *Catalog) DomainCardsByDomain(domain string) []cards.DomainCard {
	var out []cards.DomainCard
	for _, d := range c.ListDomainCards() {
		if d.Domain == domain {
			out = append(out, d)
		}
	}
	return out
}

func (c *Catalog) Ancestry(slug string) (cards.Ancestry, bool) {
	a, ok := c.Ancestries[slug]
	return a, ok
}

func (c *Catalog) ListAncestries() []cards.Ancestry {
	out := make([]cards.Ancestry, 0, len(c.Ancestries))
	for _, a := range c.Ancestries {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (c *Catalog) Community(slug string) (cards.Community, bool) {
	m, ok := c.Communities[slug]
	return m, ok
}

func (c *Catalog) ListCommunities() []cards.Community {
	out := make([]cards.Community, 0, len(c.Communities))
	for _, m := range c.Communities {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (c *Catalog) Class(slug string) (cards.Class, bool) {
	k, ok := c.Classes[slug]
	return k, ok
}

func (c *Catalog) ListClasses() []cards.Class {
	out := make([]cards.Class, 0, len(c.Classes))
	for _, k := range c.Classes {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ClassesByDomain returns the classes that draw on the given domain. Every domain
// backs exactly two classes.
func (c *Catalog) ClassesByDomain(domain string) []cards.Class {
	var out []cards.Class
	for _, k := range c.ListClasses() {
		for _, d := range k.Domains {
			if d == domain {
				out = append(out, k)
				break
			}
		}
	}
	return out
}

// Subclass finds a subclass by its own slug, across every class.
func (c *Catalog) Subclass(slug string) (cards.Class, cards.Subclass, bool) {
	for _, k := range c.ListClasses() {
		if s, ok := k.Subclass(slug); ok {
			return k, s, true
		}
	}
	return cards.Class{}, cards.Subclass{}, false
}

func (c *Catalog) Beastform(slug string) (cards.Beastform, bool) {
	b, ok := c.Beastforms[slug]
	return b, ok
}

// ListBeastforms returns every Beastform option sorted by tier, then by name.
func (c *Catalog) ListBeastforms() []cards.Beastform {
	out := make([]cards.Beastform, 0, len(c.Beastforms))
	for _, b := range c.Beastforms {
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Tier != out[j].Tier {
			return out[i].Tier < out[j].Tier
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// BeastformsUpToTier returns the forms a druid of the given tier can assume:
// their tier or lower, per the Beastform class feature.
func (c *Catalog) BeastformsUpToTier(tier string) []cards.Beastform {
	var out []cards.Beastform
	for _, b := range c.ListBeastforms() {
		if b.Tier <= tier {
			out = append(out, b)
		}
	}
	return out
}

func (c *Catalog) Companion(slug string) (cards.Companion, bool) {
	p, ok := c.Companions[slug]
	return p, ok
}

func (c *Catalog) EnvironmentsByTier(tier string) []cards.Environment {
	var out []cards.Environment
	for _, e := range c.ListEnvironments() {
		if e.Tier == tier {
			out = append(out, e)
		}
	}
	return out
}
