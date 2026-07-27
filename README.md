# DH Companion

> 🚧 **Work in progress** — early development, nothing is stable yet.

A desktop companion app for [Daggerheart](https://www.daggerheart.com/), for both sides of
the table: a GM tool (encounter builder, live combat runner, campaign notes) and a player
tool (character sheet, loadout/vault, dice). Local-first — everything lives in a SQLite
file on your machine, no account, no server.

## Stack

- **Go** + **[Wails v2](https://wails.io/)** — native desktop shell
- **Svelte 5** + Vite — frontend
- **SQLite** (`modernc.org/sqlite`, pure Go, no CGO) with **sqlc** for queries and
  **goose** for migrations

## Status

**Phase 0 done** — the app launches, creates its database and runs migrations on startup,
asks for a role on first run (GM or Player), remembers it, and lets you switch at any time
from the header.

**Phase 1 done** — the shared core both modules build on, all of it headless:

- `internal/rules` — encounter budget math: cost per adversary type, minion batching, and
  the difficulty / multi-solo / below-tier / no-big-adversary adjustments
- `internal/dice` — duality dice (2d12 Hope + Fear) for the player side, a GM d20 with
  advantage/disadvantage, and damage rolls
- `internal/cards` — shared card model (adversary and environment variants)
- `internal/srd` — loader for the SRD json embedded in `data/`
- `sql/schema` + `sql/queries` — SQLite schema and sqlc queries for parties, custom
  adversaries, custom environments and encounters

**Phase 2 done** — the GM encounter builder:

- browsers for adversaries and environments over the SRD and your homebrew, merged and
  filterable by tier and type
- homebrew editors — create, edit and delete custom adversaries and environments
- parties, and an encounter builder with a live battle-point budget meter
- an optional environment per encounter, with its impulses and features shown inline

**Phase 3 done** — the live combat runner: start a fight from a saved encounter and its
adversaries spawn as combatants with their own HP and Stress, a Fear tracker, countdown
clocks, the environment on hand as GM prompts, and the dice rollers available without
leaving the fight. State saves as you go, so closing mid-combat is safe.

The Player module is still navigation-only — it starts at Phase 5. See
[BUILD_PLAN.md](BUILD_PLAN.md) for the full roadmap.

## Running it

```sh
wails dev      # dev mode with hot reload
wails build    # produce a binary in build/bin
go test ./...  # backend unit tests
```

Requires Go 1.25+, Node, and the [Wails CLI](https://wails.io/docs/gettingstarted/installation).

On Linux distros that only ship webkit2gtk 4.1, add `-tags webkit2_41` to both commands.

## Data

The database is created at `~/DH-Companion/data.db`. Override the location with the
`DH_DATA_DIR` environment variable.

After changing anything in `sql/schema/` or `sql/queries/`, regenerate the DB layer:

```sh
sqlc generate
```

## License

TBD. Daggerheart content belongs to Darrington Press — see the Community Gaming License.
