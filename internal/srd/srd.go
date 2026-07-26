package srd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"

	"github.com/mhetem/DH-Companion/data"
	"github.com/mhetem/DH-Companion/internal/cards"
)

const (
	adversariesFile  = "adversaries.json"
	environmentsFile = "environments.json"
)

type Catalog struct {
	Adversaries  map[string]cards.Adversary
	Environments map[string]cards.Environment
}

func Load(fsys fs.FS) (*Catalog, error) {
	c := &Catalog{
		Adversaries:  map[string]cards.Adversary{},
		Environments: map[string]cards.Environment{},
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

func (c *Catalog) EnvironmentsByTier(tier string) []cards.Environment {
	var out []cards.Environment
	for _, e := range c.ListEnvironments() {
		if e.Tier == tier {
			out = append(out, e)
		}
	}
	return out
}
