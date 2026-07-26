# DH Companion — Build Plan

A Daggerheart GM + Player desktop companion. **Go + Wails + SQLite.**
Reuses the domain logic from `DHEncounters` (encounter math, adversary model,
the sqlc workflow) and drops everything web/multi-user.

---

## Decisions locked

| Decision | Choice |
|---|---|
| Framework | Wails v2, Go backend + Svelte frontend |
| Storage | SQLite (local file), `sqlc` for queries, `goose` for migrations |
| Auth | **None.** Role is a switchable view (GM / Player), not a login. Last-used role is remembered; switch anytime without losing data |
| GM ↔ Player data | **Fully separate.** Two schema islands in one DB, no shared records |
| Shipping | Two modules, independently buildable. GM first (mostly a port), Player second (greenfield) |

### What carries over from DHEncounters
- `encounter_math.go` → `internal/rules/` **verbatim** (pure, no deps)
- `adv.go` structs → `internal/cards/`
- The `.sql`-as-source + sqlc codegen pattern
- Adversary reference data in `data/`

### What gets dropped
- All HTTP handlers, `decode_helper.go`, `views.go`, HTTP error plumbing
- `auth_handlers.go`, `session.go`, `ratelimit.go`, `internal/auth/`, users + sessions tables
- `templates/`, `static/`
- pgx / pgtype / Postgres, testcontainers, Docker/compose
- `user_id` on every table and query

---

## Target structure

```
dh-companion/
├── main.go                 # Wails entry
├── app.go                  # App struct + bound methods (the JS<->Go bridge)
├── wails.json
├── internal/
│   ├── rules/              # encounter_math.go (ported verbatim)
│   ├── dice/               # duality dice (Hope/Fear) roller
│   ├── cards/              # shared card model + variants (adversary|environment|domain|ability)
│   ├── srd/                # static reference data loader
│   ├── db/                 # sqlc-generated SQLite code
│   ├── gm/                 # encounters, combat runner, campaigns, sessions, notes
│   └── player/             # characters, loadout/vault, inventory, leveling
├── sql/
│   ├── schema/             # goose migrations (SQLite dialect)
│   └── queries/            # sqlc source
├── data/                   # SRD json (adversaries, environments, domain cards, ancestries…)
├── .github/workflows/      # CI (test/vet/build) + tagged release automation
└── frontend/               # Svelte: /gm and /player route trees
```

---

## Phase 0 — Foundations (skeleton that launches)

**Goal:** empty app opens, shows a role, opens a SQLite file. Role is switchable.

- [x] `wails init -n dh-companion -t svelte`
- [x] Add deps: SQLite driver (`modernc.org/sqlite`, pure-Go = no CGO), `pressly/goose`, `sqlc`
- [x] `sqlc.yaml`: engine `sqlite`, point at `sql/schema` + `sql/queries`
- [x] DB bootstrap: open `~/DH-Companion/data.db` (override with `DH_DATA_DIR`), run goose migrations on startup.
      **Both schema islands (GM + Player) always exist regardless of current role** — nothing is role-scoped.
- [x] `settings` table (`key TEXT PRIMARY KEY, value TEXT`) — stores `last_role` (a preference, not a gate)
- [x] Frontend: on launch, open the last-used role's shell; first run shows the picker
- [x] **Persistent "switch role" affordance** (header) available anytime — flips the view, updates `last_role`, touches no data
- [x] Bound methods: `GetRole()`, `SetRole(role)`

**Done when:** app launches into your last role; you can switch GM↔Player at will; data built in either mode survives the switch and relaunch.

---

## Phase 1 — Shared core

**Goal:** the pieces both modules depend on, no UI yet.

- [ ] Port `encounter_math.go` → `internal/rules/`. Add unit tests (they run in-memory now — drop testcontainers).
- [ ] `internal/dice/`: **Duality Dice** roller — 2d12 as Hope/Fear, critical on matching dice, advantage/disadvantage (±d6), flat modifiers. Pure functions, table-tested.
- [ ] Port `adv.go` structs → `internal/cards/`. Generalize to a `Card` shape covering adversary + environment + domain + ability variants.
- [ ] `internal/srd/`: loader for static json in `data/`. Reuse the adversary json you already have; add an environment dataset alongside it (same loader).
- [ ] Convert schema files (see cheatsheet below), strip `user_id` everywhere.

**Done when:** `go test ./internal/rules/... ./internal/dice/...` passes; SRD loader returns adversaries.

### pgx → SQLite porting cheatsheet
| Postgres | SQLite |
|---|---|
| `SERIAL PRIMARY KEY` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| `pgtype.Timestamp` | `TEXT` (RFC3339) or `INTEGER` (unix) |
| `pgtype.Int4` | `sql.NullInt64` |
| `[]byte` JSON columns | keep as `TEXT` |
| `$1, $2` placeholders | `?` (sqlc handles via engine setting) |
| testcontainers Postgres | in-memory `:memory:` SQLite |
| — | **remove** every `user_id` column + `WHERE user_id = ?` clause |

---

## Phase 2 — GM: Encounter Builder (port)

**Goal:** rebuild what DHEncounters already does, as a desktop module.

Schema (island A):
```sql
party(id, name, size, tier, created_at, updated_at)
custom_adversary(id, slug, name, tier, type, ..., standard_attack TEXT, features TEXT)
custom_environment(id, slug, name, tier, type, difficulty, description,
                   impulses TEXT, potential_adversaries TEXT, features TEXT)
encounter(id, name, party_id, adversaries TEXT, environment_slug TEXT,
          created_at, updated_at)   -- environment_slug: SRD or custom, nullable
```

- [ ] Migrate `parties`, `custom_adversaries`, `custom_environments`, `encounters` schema + queries (minus auth)
- [ ] Bound methods: CRUD for parties, custom adversaries, **custom environments**, encounters
- [ ] `ComputeBudget(settings, picks) -> BudgetSummary` bound straight to `internal/rules`
- [ ] Frontend: adversary browser (SRD + custom), **environment browser (SRD + custom)**, encounter builder, live budget meter
- [ ] Attach an environment to an encounter (optional, one per encounter); show its impulses/features/potential-adversaries inline

**Done when:** you can build an encounter against a party, attach an environment, see the budget update live, and save it — feature parity with the web app plus environments.

---

## Phase 3 — GM: Combat Runner (new)

**Goal:** the reason this is a desktop app — run the encounter live.

Schema (island A, cont.):
```sql
combat(id, encounter_id, fear INTEGER DEFAULT 0, active INTEGER, created_at)
combatant(id, combat_id, adversary_slug, display_name,
          hp_max, hp_marked, stress_max, stress_marked, spotlight INTEGER)
countdown(id, name, value INTEGER, max INTEGER, kind TEXT)
```

- [ ] Start a combat from a saved encounter (spawns combatants from picks)
- [ ] Per-combatant HP/Stress steppers, spotlight toggle, quick-view of card features
- [ ] **Fear tracker** (0–12) at the combat level
- [ ] **Countdowns / progress clocks** — add, tick up/down, delete
- [ ] Surface the encounter's **environment** in the runner — impulses + features on hand as GM prompts during the scene
- [ ] **Duality dice** widget wired to Fear — a roll with Fear bumps the pool
- [ ] Auto-save state (so closing the app mid-fight is safe)

**Done when:** you can run a saved encounter end to end — mark damage, roll duality dice, spend/gain Fear, advance a countdown — and reopen exactly where you left off.

---

## Phase 4 — GM: Campaign & Session Notes (new)

**Goal:** the connective tissue — where a campaign lives between fights. (This was the
original motivation for the app; it ties the builder, runner, and notes into one tool.)

Schema (island A, cont.):
```sql
campaign(id, name, description, current_fear INTEGER DEFAULT 0, created_at, updated_at)
session(id, campaign_id, number, title, date, recap TEXT, created_at)
note(id, campaign_id, kind TEXT, title, body TEXT, created_at, updated_at) -- kind: npc|location|faction|lore|plot
session_encounter(session_id, encounter_id)   -- link fights you ran to a session
-- backfill: add campaign_id to `countdown` (clocks belong to a campaign)
```
**Fear graduates here** from per-combat (Phase 3) to per-campaign — it persists between
fights, matching the rules. The runner reads/writes `campaign.current_fear` when a combat
belongs to a campaign.

- [ ] Campaign CRUD; a campaign owns its Fear, countdowns, sessions, and notes
- [ ] Session log: numbered sessions with title/date/recap; link the encounters run that session
- [ ] Notes: typed entries (NPC / location / faction / lore / plot thread), markdown body
- [ ] Cross-link — jump from a running combat to the campaign's NPCs/notes; from a session, open its encounters
- [ ] **SQLite FTS5 full-text search** across notes + adversaries + environments (one virtual table)

**Done when:** you can create a campaign, log a session with a recap linked to an encounter
you ran, keep typed NPC/location notes, and full-text search across everything.

---

## Phase 5 — Player: Character Companion (greenfield)

**Goal:** the second persona. Full Daggerheart character sheet.

Schema (island B):
```sql
character(id, name, class, subclass, ancestry, community, level,
          agility, strength, finesse, instinct, presence, knowledge,
          hp_max, hp_marked, stress_max, stress_marked, hope,
          evasion, armor_score, threshold_major, threshold_severe, gold,
          created_at, updated_at)
character_domain_card(id, character_id, card_slug, location)  -- 'loadout' | 'vault'
inventory_item(id, character_id, name, qty, kind, equipped)
experience(id, character_id, name, modifier)
```

- [ ] SRD data: domain cards, class/subclass features, ancestries, communities (source + load into `data/`)
- [ ] Character CRUD + create wizard (class → subclass → ancestry → community → traits)
- [ ] Sheet view: traits, live-tracked HP/Stress/Hope, evasion/armor/thresholds, gold
- [ ] **Loadout (max 5) vs. Vault** domain-card management (drag/move, enforce cap)
- [ ] Inventory + equipment
- [ ] Experiences (name + modifier)
- [ ] Leveling flow across tiers (1, 2–4, 5–7, 8–10)
- [ ] **Duality dice** widget wired to Hope — a roll with Hope grants a Hope; apply Experience/trait modifiers

**Done when:** you can build a character, manage loadout/vault, roll with your traits, and track resources in play.

---

## Phase 6 — CI, Packaging & Release

> Set up the CI workflow back at Phase 0 and let it grow with the project — it's listed
> here only because release automation depends on the app being buildable. Don't leave CI
> to the end.

- [ ] **CI (GitHub Actions):** `go vet`, `go test ./...`, `wails build` on every push/PR; status badge in README
- [ ] **Cross-platform release automation:** on tag, build Win/macOS/Linux binaries via Wails and attach to a GitHub Release
- [ ] **Auto-update:** wire Wails' updater to the release feed
- [ ] **Data portability:** DB backup/export + import (single-file copy), plus **whole-library JSON export/import**
- [ ] **Shareable homebrew codes:** compress+base64 a custom adversary/environment/character into a paste-able string others can import (versioned format)
- [ ] App icon + name; empty states; keyboard shortcuts for combat/dice; README + first-run notes

---

## Phase 7 — Stretch: LAN session sharing

**Goal:** the senior-signal feature — real-time, read-only session view for players.

- [ ] GM's app runs a lightweight WebSocket server (opt-in, bound to the local network)
- [ ] Broadcasts live **Fear, countdowns, and spotlight** as they change
- [ ] Players on the same Wi-Fi connect (URL/QR) to a read-only session view
- [ ] Graceful degrade — everything still works fully offline if nobody connects

**Done when:** a second device on the LAN watches Fear and countdowns update live as the GM runs a fight. Clearly optional; nothing else depends on it.

---

## Data sourcing note
Adversaries you already have. **Environments** (GM side), plus domain cards, class
features, ancestries, and communities (Player side) need to be assembled into `data/`
json (same loader pattern as adversaries). This is content work, not code — it can proceed
in parallel with Phases 2–5. Check the Darrington Press Community Gaming License for what
you can redistribute.

## Suggested order
Phase 0 → 1 → 2 → 3 → 4 = a complete, usable **GM tool** (build, run, and record a
campaign). Then Phase 5 (Player) slots in independently, and Phase 6 (CI/release) rides
along from Phase 0 onward. Phase 7 (LAN sharing) is an optional capstone.
GM side reaches "better than the web app" by the end of Phase 4.
