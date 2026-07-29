<img src="frontend/src/assets/images/hilt-app-lockup.svg" alt="Hilt" height="90">

[![CI](https://github.com/mhetem/DH-Companion/actions/workflows/ci.yml/badge.svg)](https://github.com/mhetem/DH-Companion/actions/workflows/ci.yml)

> 🚧 **Work in progress.** Both sides of the table are complete and usable. What's left is
> the optional LAN session view — see [Status](#status).

A **[Daggerheart](https://www.daggerheart.com/)™ Compatible** desktop companion for both
sides of the table. Build encounters against a live battle-point budget, run the fight with
per-adversary HP, Stress and Fear tracking, and keep the campaign between sessions —
sessions, typed notes, countdown clocks, and full-text search across all of it.

Local-first: everything lives in one SQLite file on your machine. No account, no server, no
network access unless you explicitly press "check for updates".

## Motivation

Running *Daggerheart* means holding a lot of state at once. Prep is a battle-point budget
weighed against your party's size and tier. The fight itself is HP and Stress on a dozen
adversaries, a Fear pool, and a couple of countdown clocks. Between sessions it's who the
party met, what they owe, and which thread is about to come due.

That usually ends up spread across a spreadsheet, a notes app, and a stack of index cards,
none of which know about each other. Hilt is one tool where the encounter you built is the
fight you run, the fight you run is logged to the session you played, and the NPC you wrote
up three weeks ago is one search away while you're running it.

It's a desktop app because the live combat runner wants to be there when the Wi-Fi isn't,
and because your campaign notes should be a file you own rather than a row in someone's
database.

## Features

### GM

- **Adversary and environment browsers** over the SRD catalog (129 adversaries, 19
  environments) merged with your own homebrew, filterable by tier and type. A custom card
  shadows an SRD one on slug collision.
- **Homebrew editors** for custom adversaries and environments, with a live preview and a
  reference pane that is the real browser — so you can crib from an existing card while you
  write, or copy one wholesale with *Use as template*.
- **Parties** — size and tier, which is what the budget is computed from.
- **Encounter builder** with a live difficulty-budget meter: cost per adversary type,
  party-sized minion batching, and the easy/hard, multi-solo, below-tier and
  no-big-adversary adjustments. Attach an optional environment and see its impulses and
  features inline.
- **Combat runner** — start a fight from a saved encounter and its adversaries spawn as
  combatants with their own HP and Stress. Per-combatant steppers, non-exclusive spotlight,
  card quick-view, a 0–12 Fear tracker, countdown clocks, the environment on hand as GM
  prompts, and the dice rollers in the rail. State autosaves as you go, so closing the app
  mid-fight is safe.
- **Campaigns** — the connective tissue. A campaign owns its Fear, its numbered session log
  (with recaps and the encounters prepped for and actually run that night), its typed notes
  (NPC / location / faction / lore / plot, markdown body), and its countdown clocks.
- **Full-text search** over notes, adversaries and environments in one SQLite FTS5 index,
  with highlighted excerpts.
- **Dice** — a GM d20 with advantage/disadvantage and modifiers, plus a damage roller.
  Duality dice are deliberately absent: only players roll Hope and Fear.

### Player

- **Character creation** — a seven-step wizard through class, subclass, ancestry, community
  and traits, with the SRD text for every option shown beside the picker.
- **The sheet** — traits, live-tracked Hit Points, Stress, Hope and armor slots, Evasion and
  both damage thresholds, gold in handfuls/bags/chests, and your class, subclass, ancestry
  and community features on the page.
- **Loadout and vault** — you know two domain cards at level 1 and one more per level; five
  of them can sit in your loadout at a time, the rest in the vault. The pool only ever offers
  cards in your domains at or below your level.
- **Inventory** — gear and consumables, with one primary weapon, one secondary and one set of
  armor equipped at a time, plus a one-press way to add your class's starting items.
- **Levelling** — the tier advancement table with its slot boxes, remembering what you spent
  at earlier levels in the same tier, and applying the tier achievements at 2, 5 and 8.
- **Duality dice** — 2d12 read as Hope and Fear with your trait and Experience modifiers.
  A roll with Hope grants you a Hope and a critical also clears a Stress, written to your
  sheet as part of the roll. Damage rolls use your Proficiency as the dice count. The roller
  sits in a collapsible block on the sheet as well as in its own section, and **clicking any
  trait rolls it**.
- **Rests** — choose your downtime moves (two on a short rest, three on a long one) rather
  than clearing everything by default, and each one reports the roll behind it.
- **Beastform** for druids — every form at your tier or lower with its trait and Evasion
  bonuses, attack and features. Transforming marks a Stress and shows your effective Evasion
  while you're in it.
- **Companion sheet** for Beastbound rangers — name, Evasion, attack, damage die and range,
  a Stress track, Companion Experiences, and the level-up options as a checklist. Damage
  rolls use your Proficiency and their die.
- **Gold** lives on the Inventory page, counted in handfuls, bags and chests.

### Both

- **Role is a view, not a login.** Switch GM ↔ Player at any time from the header. Nothing
  is locked to a role, and data built in either survives the switch and relaunch.
- **Keyboard-driven combat** — see [Keyboard shortcuts](#keyboard-shortcuts).
- **Backup and restore**, whole-library JSON export/import, and paste-able
  [share codes](#share-codes) for homebrew cards.
- **Window size and UI scale** are separate, remembered settings (80–200% scale, HD → 4K).

## Install

Grab the latest build from **[Releases](https://github.com/mhetem/DH-Companion/releases/latest)**
— no toolchain, no terminal.

| Platform | Download | How |
|---|---|---|
| **Windows** | `Hilt-…-windows-amd64-setup.exe` | Run the installer. Pick a Start-menu shortcut, a desktop shortcut, both, or neither. |
| **macOS** | `Hilt-…-macos-universal.dmg` | Open it and drag Hilt to Applications. Universal — Intel and Apple Silicon. |
| **Linux** (Debian/Ubuntu) | `Hilt-…-amd64.deb` | `sudo apt install ./Hilt-*.deb`, then launch it from your app menu. |
| **Linux** (any distro) | `Hilt-…-x86_64.AppImage` | `chmod +x Hilt-*.AppImage && ./Hilt-*.AppImage` — nothing to install. |

Plain `.zip` / `.tar.gz` archives of the bare binary are attached too.

### First launch

The builds are **unsigned**, so each OS will object once:

- **macOS** — right-click the app and choose *Open* (not double-click), or
  `xattr -dr com.apple.quarantine /Applications/Hilt.app`
- **Windows** — SmartScreen: *More info* → *Run anyway*
- **Linux** — both packages need `libgtk-3-0` and `libwebkit2gtk-4.1-0` present; the `.deb`
  declares them, the AppImage assumes them

Then pick a side of the table. The app creates its database at `~/DH-Companion/data.db`,
runs its migrations, and you're in.

Upgrading is install-over-the-top; your data is untouched. **Uninstalling keeps it too** —
the Windows uninstaller has a *Campaigns, encounters and homebrew* tick-box that deletes
`%USERPROFILE%\DH-Companion` if you want a clean sweep, and it is unticked unless you ask.
Removing the `.deb` or deleting the AppImage never touches your data. There's a backup
button in **Settings & data** regardless.

> If you moved the database with `DH_DATA_DIR`, the uninstaller doesn't know about it —
> delete that folder yourself.

## Building from source

Requires [Go 1.25+](https://go.dev/dl/), [Node 22+](https://nodejs.org/), and the
[Wails CLI](https://wails.io/docs/gettingstarted/installation).

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0
git clone https://github.com/mhetem/DH-Companion.git
cd DH-Companion
wails dev
```

On Linux distros shipping webkit2gtk 4.1 without a 4.0 compat package (Ubuntu 24.04+,
recent Fedora), add the build tag — otherwise the build fails at `pkg-config`:

```bash
wails dev -tags webkit2_41
```

To produce the installable packages yourself:

```bash
wails build -tags webkit2_41            # Linux binary
build/linux/package.sh v1.0.0 dist      # -> dist/*.deb and dist/*.AppImage

wails build -nsis                       # Windows: build/bin/DH-Companion-amd64-installer.exe
```

## Usage

1. **Pick a side of the table.** Game Master or Player. Every launch asks, and the header
   switches at any time.
2. **Create a party** (Parties) — size and tier. Nothing in the budget meter means anything
   until one exists.
3. **Browse adversaries and environments**, and write your own homebrew from the same
   browsers.
4. **Build an encounter** (Encounters): pick a party, add adversaries and quantities, and
   watch the budget meter tell you whether the fight is under, at, or over. Optionally
   attach an environment.
5. **Create a campaign** (Campaigns) and log sessions against it. Fear lives here between
   fights.
6. **Run the fight** (Combat Runner): start from a saved encounter, pick the campaign whose
   Fear pool it spends, then mark damage and Stress, hold the spotlight, spend Fear, and
   advance clocks. Reopen exactly where you left off.
7. **Write it up.** Log the session with a recap, link the encounters you ran, and keep
   typed notes on the NPCs and places the party met. Search finds them later.

## Keyboard shortcuts

Press <kbd>?</kbd> anywhere for the full list. Shortcuts stand down while you are typing in
a field, and while a dialog is open.

### Combat runner

| Key | Action |
|---|---|
| <kbd>↑</kbd> <kbd>↓</kbd> / <kbd>k</kbd> <kbd>j</kbd> | Move the selection through the roster |
| <kbd>h</kbd> / <kbd>H</kbd> | Mark / clear 1 HP on the selected combatant |
| <kbd>s</kbd> / <kbd>S</kbd> | Mark / clear 1 Stress |
| <kbd>x</kbd> | Toggle spotlight |
| <kbd>c</kbd> | Clear every spotlight |
| <kbd>f</kbd> / <kbd>F</kbd> | Gain / spend a Fear |
| <kbd>Esc</kbd> | Drop the selection |

### Dice

| Key | Action |
|---|---|
| <kbd>r</kbd> | Roll the d20 |
| <kbd>d</kbd> | Roll damage |
| <kbd>a</kbd> / <kbd>z</kbd> | Toggle advantage / disadvantage |
| <kbd>1</kbd>…<kbd>7</kbd> | Pick the damage die, smallest to largest |
| <kbd>Esc</kbd> | Dismiss the result flash |

Selection is the keyboard's cursor over the roster and is deliberately *not* spotlight —
spotlight is a rules state you set on purpose and can hold on several combatants at once.

## Tech stack

- **Go 1.25** with **[Wails v2](https://wails.io/)** — native desktop shell, WebKit-based
  webview, no Electron.
- **Svelte 5** (runes) + **Vite** for the frontend, built into the binary via `//go:embed`.
- **SQLite** through **[modernc.org/sqlite](https://gitlab.com/cznic/sqlite)** — pure Go, so
  the build needs no CGO and cross-compiles cleanly.
- **[sqlc](https://sqlc.dev/)** generates type-safe Go from hand-written SQL.
- **[goose](https://github.com/pressly/goose)** for migrations, embedded and run at startup.
- **SQLite FTS5** for search.
- No frontend dependencies beyond Svelte itself — the markdown renderer is ~90 lines in the
  repo rather than a package.

## Architecture

### The bridge

Wails binds Go structs into JavaScript. Each module is bound separately rather than piling
every method onto one object:

```
window.go.main.App.*        role, window size, UI scale, backup/restore, library, updates
window.go.gm.Service.*      everything GM — cards, parties, encounters, combat, campaigns
window.go.dice.Roller.*     the pure roll functions
```

Wails publishes *every* exported method on a bound struct, which shapes the code: `gm.Attach`
is a package function rather than a method, because an `Attach` method would push
`context.Context`, `*db.Queries` and `*sql.DB` into the generated TypeScript models.

### Data flow

```
Svelte component
  -> frontend/src/lib/gm/api.js        single import surface over the generated bindings
    -> frontend/wailsjs/go/…           generated wrappers (regenerated by wails build/dev)
      -> internal/gm.Service           validation, bridge types, transactions
        -> internal/db (sqlc)          type-safe queries
          -> SQLite
```

Two rules hold throughout:

- **Nothing returns a `db.*` row.** The bridge types are `gm.Party`, `cards.Adversary`,
  `rules.EncounterView`, `gm.CombatView` and so on, with `*T` for nullable columns — the
  frontend never sees a `sql.NullString`, and JSON-in-TEXT columns arrive decoded.
- **Mutators return the updated row.** There is no separate save step to fall out of sync
  with; the UI renders what the database actually holds.

### Project layout

```
main.go                 Wails entry, window options, bound structs
app.go                  App struct: role, window size, UI scale, DB open + migrations
portable.go             backup/restore, library export/import, update check (file dialogs)

internal/
  rules/                encounter budget math + the encounter view model
  dice/                 duality dice, GM d20, damage; die sizes as a closed set
  cards/                shared card model (adversary, environment, domain, class, …)
  srd/                  loader for the embedded SRD json
  db/                   sqlc-generated queries and models
  gm/                   the GM service — one file per area, plus:
    browse.go             SRD ∪ custom union, custom-shadows-SRD
    search.go             FTS5 query (hand-written; see Key decisions)
    library.go            whole-library JSON export/import
    share.go              homebrew share codes
    validate.go           tier/type/kind lists, name and slug rules
  share/                the versioned share-code codec
  update/               GitHub release feed check

sql/
  schema/               goose migrations (17, embedded)
  queries/              sqlc sources

data/                   SRD json + //go:embed
build/                  appicon.svg and the rasterized platform icons
.github/workflows/      ci.yml, release.yml

frontend/src/
  App.svelte            role gate, error toast, reload key
  lib/
    Header.svelte         lockup, role badge, window/scale, settings, shortcuts
    Shell.svelte          nav + pane switching
    Modal.svelte          the scrim/sheet dialog shell (owns focus + Escape)
    Settings.svelte       backup, library, version, data location
    ShortcutHelp.svelte   the ? sheet
    EmptyState.svelte     full-pane "nothing here yet"
    keys.js               keyboard mechanism (see Key decisions)
    gm/                   every GM pane; api.js is the single import surface
```

## Key decisions

- **Role is a preference, not a gate.** Both schema islands exist regardless of the current
  role, and switching touches no data. There is no auth anywhere in the app.
- **sqlc, not an ORM.** Queries live in `sql/queries` as SQL and compile to type-safe Go.
- **Clamping happens in SQL.** Fear, HP, Stress and countdown values are written as
  `max(0, min(ceiling, current + delta))`, so holding a stepper at either end is a no-op
  rather than an error to handle mid-scene. `CHECK` constraints are the backstop.
- **Custom cards are addressed by slug, and slugs are immutable.** Slugs derive from the
  name on create; editing a card keeps its slug, so a rename can't orphan an encounter that
  references it. Deleting leaves referencing encounters intact — the pick comes back
  `unresolved` rather than failing the load.
- **The FTS5 read is hand-written Go, not sqlc.** sqlc 1.31.1 rejects every spelling of the
  fts5 `MATCH` idiom because its SQLite parser has no notion of a virtual table's hidden
  name column. Only the index *writes* go through sqlc. See the Phase 4 notes in
  [BUILD_PLAN.md](BUILD_PLAN.md) for the full list of what was tried.
- **Search excerpts are escaped, then re-highlighted.** `snippet()` returns surrounding text
  exactly as indexed, and note bodies are indexed raw, so the renderer escapes everything and
  lets only the `<mark>` markers back through.
- **UI scale is separate from window size.** A 4K window on a 4K screen makes everything
  physically *smaller*, because CSS pixels don't grow with resolution. Scale is applied to
  the root font size, which works because every length in the app is in `rem`.
- **`appearance: none` on native controls.** WebKitGTK paints `select`, number spinners and
  checkboxes as system widgets and ignores `background-color`, so they rendered as bright
  controls carrying the app's light text. Reclaiming them means drawing the chevron and tick
  as data URIs — which is why those two restate palette colors as literals.
- **Only players roll duality dice.** The runner has no Hope/Fear widget and nothing bumps
  Fear automatically; the GM gains and spends it by hand.
- **No auto-update.** Wails v2 ships no updater, and a background check would quietly undo
  the "no server" promise. The version check is manual and opens the download page.

## Data

The database is created at `~/DH-Companion/data.db`. Override the location with
`DH_DATA_DIR`. Everything below is in **Settings & data** (the gear in the header).

| Variable | Used by | Description |
|---|---|---|
| `DH_DATA_DIR` | app | Directory holding `data.db`. Defaults to `~/DH-Companion`. |
| `GOOSE_DRIVER` | migrations | `sqlite3`, for running goose by hand. |
| `GOOSE_DBSTRING` | migrations | Path to the database file. |
| `GOOSE_MIGRATION_DIR` | migrations | `sql/schema`. |

### Backup and restore

A backup is a `VACUUM INTO` snapshot — a consistent copy of the live database taken without
closing it, which a plain file copy of a WAL-mode database is not. Restoring validates the
incoming file first (it must carry both `settings` and `goose_db_version`), snapshots your
current database to `data-replaced-<timestamp>.db`, then swaps and reopens.

### Library export/import

Readable JSON covering parties, homebrew cards, encounters, and campaigns with their
sessions, notes and clocks. Import **adds** rather than replacing, and renames anything
whose name is taken — a second `Gutter Wraith` becomes `Gutter Wraith (2)`, and imported
encounters are remapped so their picks still resolve.

### Share codes

Any custom adversary or environment has a *Share* button producing a `HILT1:…` string;
*Import code* in the browser takes one back. The format is
`HILT<version>:<base64url(zlib(json))>` — versioned, so a code from a newer build is refused
with a message that says so rather than failing as corrupt. A code carries **one card and
nothing else**: no encounters, no campaign, nothing personal. Imports are previewed before
they are written.

## Running

```bash
wails dev      # hot reload; regenerates the JS bindings
wails build    # binary in build/bin
go test ./...  # backend unit tests
```

Add `-tags webkit2_41` to `wails dev`/`wails build` on webkit2gtk 4.1 distros.

After changing anything in `sql/schema/` or `sql/queries/`, regenerate the DB layer:

```bash
sqlc generate
```

To run migrations by hand (the app runs them at startup anyway), source the `.env` in the
repo root and use `goose up`.

## Testing

```bash
go test ./...        # everything
go test ./internal/rules/... ./internal/dice/...   # the pure logic
```

Tests cover the encounter budget math, the dice rollers and die-size validation, the GM
service, the share-code codec, and the update check. There is no container dependency — the
service tests run against SQLite in place, which is the same engine as production, and the
update check runs against an `httptest` server rather than the network.

## CI and releases

Two workflows:

- **`ci.yml`** — on every push and PR: `gofmt -l` (failing on any unformatted file),
  `go vet`, `go test ./... -race`, then a full `wails build` on Linux, Windows and macOS.
  `fail-fast: false`, so one platform breaking still reports the others.
- **`release.yml`** — on a `v*` tag: builds `linux/amd64`, `windows/amd64` and
  `darwin/universal`, packages each with `LICENSE`, `NOTICE.md` and `README.md`, publishes
  `SHA256SUMS.txt`, and attaches everything to a GitHub Release. The tag is stamped into the
  binary via `-ldflags -X main.version=...`; untagged builds report `dev`.

Release binaries are **unsigned** — macOS Gatekeeper and Windows SmartScreen will both warn
on first run.

## Status

| Phase | | |
|---|---|---|
| 0 | Foundations — launches, opens its database, remembers your role | ✅ |
| 1 | Shared core — budget math, dice, card model, SRD loader | ✅ |
| 2 | GM: encounter builder, browsers, homebrew editors, budget meter | ✅ |
| 3 | GM: live combat runner, Fear, countdowns, autosave | ✅ |
| 4 | GM: campaigns, session log, typed notes, FTS5 search | ✅ |
| 5 | Player: character sheet, loadout/vault, leveling, duality dice | ✅ |
| 6 | CI, packaging, releases, data portability, share codes | ✅ |
| 7 | Stretch: LAN session sharing (read-only player view) | ⬜ |

[BUILD_PLAN.md](BUILD_PLAN.md) has the full roadmap and, per phase, a "notes on what
landed" section recording why things are the way they are.

## Contributing

Contributions are welcome. A rough workflow:

1. Fork the repository and create a feature branch.
2. `wails dev` (add `-tags webkit2_41` on Linux if needed) and make your change.
3. If the change touches the database, add a goose migration under `sql/schema/` and the
   matching query in `sql/queries/`, then run `sqlc generate`.
4. Run `gofmt`, `go vet` and `go test ./...` — CI enforces all three.
5. Record *why*, not just what, in the relevant phase's notes in `BUILD_PLAN.md`. That file
   is where this project keeps its rationale, which is why the Go and SQL are comment-free.
6. Open a pull request describing the change and how you verified it.

Please open an issue first for larger changes so the approach can be discussed.

## License

The **software** — Go, Svelte, CSS, SQL, build config, and the Hilt name and logo — is
[MIT](LICENSE).

The **game content in `data/`** is not. It comes from the Daggerheart System Reference
Document 1.0, © Critical Role, LLC, used under the
[Darrington Press Community Gaming License](https://darringtonpress.com/license/), and stays
the property of its owner. The required attribution, and what applies to you if you fork
this, are in [NOTICE.md](NOTICE.md).

Daggerheart™ is a trademark of Critical Role, LLC. This is an independent, unofficial tool,
not sponsored by or affiliated with Darrington Press or Critical Role.

### If you fork or redistribute this

`data/embed.go` compiles the SRD json into the binary, so **shipping a build means
redistributing Public Game Content**. The DPCGL's Permitted Formats (§1.9) are print, video,
podcasts and DRP-whitelisted virtual tabletops; general desktop applications are not on the
list, and it expressly excludes "video games, and any other audiovisual medium not expressly
permitted."

Note that giving it away free does not change this. §1.8 defines Sharing as making content
available to the public by any means, and §2.1(b) applies the format limit to Sharing and
selling alike. The exemption that does apply needs *private* **and** non-commercial —
"private, non-commercial play among friends, family, or gaming groups in a personal
setting." Running this yourself or handing it to your table is squarely outside the
license's reach; a public repo is not.

This project ships in that position knowingly. If you build on it, make your own call. This
is not legal advice.
