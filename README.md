<img src="frontend/src/assets/images/hilt-app-lockup.svg" alt="Hilt" height="90">

[![CI](https://github.com/mhetem/DH-Companion/actions/workflows/ci.yml/badge.svg)](https://github.com/mhetem/DH-Companion/actions/workflows/ci.yml)

> 🚧 **Work in progress** — early development, nothing is stable yet.

A **[Daggerheart](https://www.daggerheart.com/)™ Compatible** desktop companion app, for
both sides of the table: a GM tool (encounter builder, live combat runner, campaign notes)
and a player tool (character sheet, loadout/vault, dice). Local-first — everything lives in
a SQLite file on your machine, no account, no server.

**Hilt** is the app's name; the repo, the Go module and the data directory are still
`DH-Companion`, so an existing database keeps working untouched.

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

**Phase 4 done** — campaigns as the connective tissue: campaign CRUD, a numbered session
log with recaps and the encounters linked to each night, typed notes (NPC / location /
faction / lore / plot) in markdown, and SQLite FTS5 full-text search across notes,
adversaries and environments. Fear graduates from per-combat to per-campaign, so it
persists between fights.

The Player module is still navigation-only — it starts at Phase 5. See
[BUILD_PLAN.md](BUILD_PLAN.md) for the full roadmap.

## First run

Pick a side of the table — Game Master or Player. It is a preference, not a login: the
header switches roles at any time, nothing is locked to one, and data built in either mode
survives the switch. The choice is remembered, so subsequent launches open straight into it.

The window opens at 1920×1080, clamped down if your screen is smaller. Both the window size
and a separate UI scale (80–200%) live in the header and are remembered — reach for the
scale rather than the size if the app is merely hard to read.

Press <kbd>?</kbd> for the keyboard shortcuts. The combat runner is built to be run from the
keyboard mid-fight — <kbd>↓</kbd> to pick a combatant, <kbd>h</kbd> and <kbd>s</kbd> to mark
HP and Stress, <kbd>f</kbd> to gain a Fear — and the dice roll with <kbd>r</kbd> and
<kbd>d</kbd>. Shortcuts stay out of the way while you are typing in a field.

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
`DH_DATA_DIR` environment variable. The gear button in the header opens **Settings & data**,
which is where the following live:

- **Back up / restore.** A backup is a `VACUUM INTO` snapshot — a consistent copy of the
  live database taken without closing it. Restoring backs up what you have first, then
  swaps the file and reopens.
- **Export / import library.** Readable JSON covering parties, homebrew cards, encounters
  and campaigns (with their sessions, notes and clocks). Import **adds** rather than
  replacing, and renames anything whose name is taken — `Gutter Wraith` arriving beside an
  existing one becomes `Gutter Wraith (2)`, and encounters that referenced it follow the
  rename.
- **Update check.** Manual only. Nothing is sent, and nothing is checked until you press
  the button — the app makes no network request otherwise.

**Share codes** are on the homebrew cards themselves: any custom adversary or environment
has a *Share* button that produces a `HILT1:…` string, and *Import code* in the browser
takes one back. The code is a compressed, versioned payload carrying that one card and
nothing else — no encounters, no campaign, no personal data.

After changing anything in `sql/schema/` or `sql/queries/`, regenerate the DB layer:

```sh
sqlc generate
```

## License

The **software** — Go, Svelte, CSS, SQL, build config, and the Hilt name and logo — is
[MIT](LICENSE).

The **game content in `data/`** is not. It comes from the Daggerheart System Reference
Document 1.0, © Critical Role, LLC, used under the
[Darrington Press Community Gaming License](https://darringtonpress.com/license/), and
stays the property of its owner. The required attribution, and what applies to you if you
fork this, are in [NOTICE.md](NOTICE.md).

Daggerheart™ is a trademark of Critical Role, LLC. This is an independent, unofficial
tool, not sponsored by or affiliated with Darrington Press or Critical Role.

### If you fork or redistribute this

`data/embed.go` compiles the SRD json into the binary, so **shipping a build means
redistributing Public Game Content**. The DPCGL's Permitted Formats (§1.9) are print,
video, podcasts and DRP-whitelisted virtual tabletops; general desktop applications are not
on the list, and it expressly excludes "video games, and any other audiovisual medium not
expressly permitted."

Note that giving it away free does not change this. §1.8 defines Sharing as making content
available to the public by any means, and §2.1(b) applies the format limit to Sharing and
selling alike. The exemption that does apply needs *private* **and** non-commercial —
"private, non-commercial play among friends, family, or gaming groups in a personal
setting." Running this yourself or handing it to your table is squarely outside the
license's reach; a public repo is not.

This project ships in that position knowingly. If you build on it, make your own call.
This is not legal advice.
