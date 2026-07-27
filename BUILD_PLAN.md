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

- [x] Port `encounter_math.go` → `internal/rules/`. Add unit tests (they run in-memory now — drop testcontainers).
- [x] `internal/dice/`: **Duality Dice** roller — 2d12 as Hope/Fear, critical on matching dice, advantage/disadvantage (±d6), flat modifiers. Pure functions, table-tested.
- [x] `internal/dice/`: **GM d20** (`RollGMDice`) and damage rolls — only players roll duality dice.
- [x] Port `adv.go` structs → `internal/cards/`. Generalize to a `Card` shape covering adversary + environment + domain + ability variants.
      *(`Meta`/`Card` + `Adversary` and `Environment` variants are in place; `domain` and `ability`
      exist as `Kind` constants and gain their structs with the Phase 5 data work.)*
- [x] `internal/srd/`: loader for static json in `data/`. Reuse the adversary json you already have; add an environment dataset alongside it (same loader).
- [x] Convert schema files (see cheatsheet below), strip `user_id` everywhere.

**Done when:** `go test ./internal/rules/... ./internal/dice/...` passes; SRD loader returns adversaries. ✅

Notes on what landed:
- `ComputeBudget` covers spend per adversary type, party-sized minion batching, and the
  easy/hard, multi-solo, below-tier and no-big-adversary budget adjustments.
- The duality roller returns `{Hope, Fear, Result, Msg}`. A critical (matching dice) is a
  flat `Hope + Fear` — the modifier and advantage die are intentionally skipped, since a
  critical succeeds regardless. Whether a roll came up *with Hope* or *with Fear* is phrased
  by the caller from the two fields, not by the roller.
- Tests are statistical over the package-level `math/rand` source rather than seeded.
- `dice.Roller` is a stateless struct bound as its own Wails target (`window.go.dice.Roller.*`),
  matching how `gm.Service` is bound — the roll functions themselves stay pure and
  package-level.
- **Die sizes are a closed set** — d4, d6, d8, d10, d12, d20, d100. `Sides` is a named type
  holding one constant per die, and
  `RollDice` takes a `Sides` rather than an `int`, so no arbitrary number can reach
  `rand.Intn` (which panics below 1). `ParseSides` is the boundary check for values arriving
  from JS; `Roller.Damage` returns its error rather than rolling something nonsensical, and
  `RollDice` keeps a `Valid()` backstop for a `Sides` made by a raw conversion.
- `Roller.Sizes()` serves that list to the frontend, so the picker can't offer a die the
  backend would reject and the set is defined in exactly one place.
- `DualityDice` and `GMDice` carry no json tags, so they cross the bridge with capitalised
  field names (`Hope`, `Result`, `Msg`) unlike everything else. Worth tagging at some point.
- The GM shell has a **Dice** section (d20 with advantage/disadvantage/modifier, a damage
  roller, and a short roll log). It stays a GM-side utility — duality dice are player-only,
  so nothing here feeds the Fear pool.



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

- [x] Migrate `parties`, `custom_adversaries`, `custom_environments`, `encounters` schema + queries (minus auth)
- [x] Bound methods: CRUD for parties, custom adversaries, **custom environments**, encounters
- [x] `ComputeBudget(settings, picks) -> BudgetSummary` bound straight to `internal/rules`
- [x] Frontend: adversary browser (SRD + custom), **environment browser (SRD + custom)**, encounter builder, live budget meter
- [x] Attach an environment to an encounter (optional, one per encounter); show its impulses/features/potential-adversaries inline
- [x] Homebrew editors — create/edit/delete forms for custom adversaries and environments,
      reached from the browsers

Notes on what landed:
- `internal/gm.Service` is bound as its own Wails struct, so the frontend calls
  `window.go.gm.Service.*`. `gm.Attach` is a package function, not a method — Wails binds
  every exported method on a bound struct, and an `Attach` method would publish
  `context.Context`/`db.Queries`/`sql.DB` into the generated TypeScript models.
- Nothing returns a `db.*` row. Bridge types are `gm.Party`, `cards.Adversary`,
  `cards.Environment`, `rules.EncounterView`, with `*T` for nullable columns — the frontend
  never sees `sql.NullString`, and the JSON-in-TEXT columns arrive decoded.
- Custom cards are addressed by **slug** everywhere (get/update/delete). Slugs are derived
  from the name on create and are immutable, since encounters reference them with no FK.
- `BrowseAdversaries`/`BrowseEnvironments` return the SRD ∪ custom union, sorted, each card
  tagged `source: "srd" | "custom"`; a custom card shadows an SRD one on slug collision.
- Deleting a custom card leaves referencing encounters intact — the pick comes back
  `unresolved: true` rather than failing `GetEncounter`.

Notes on the frontend:
- `Shell.svelte` now renders a section's `component` when one is set and falls back to the
  "coming in phase N" stub otherwise, so the Player shell is untouched and phases 3–5 slot
  in by adding a component to the section list.
- Everything GM lives in `frontend/src/lib/gm/`; `api.js` is the single import surface over
  the generated bindings and mirrors `validate.go`'s tier/type lists and `adversaryCost`.
- **Parties got a section of its own** — it wasn't in the original checklist, but party size
  and tier are what the budget is computed from, so there was no way to reach the meter
  without it.
- The two browsers share `CardBrowser.svelte`. Tier and type go down to the Go `Filter`;
  the name search stays client-side so typing doesn't re-cross the bridge.
- The builder is a full-width view rather than a third column — the shell nav plus a
  picker plus a summary rail already fills the 1024px default window.
- The budget meter re-calls `ComputeBudget` on every edit (party, difficulty, roster) rather
  than reimplementing the math in JS. `difficulty` is a builder-local knob: `EncounterInput`
  has no column for it, so it shifts the live meter but isn't saved.
- Feature descriptions carry inline `<strong>`/`<em>` and hard newlines in the SRD json, so
  `FeatureList.svelte` renders them as HTML inside a `pre-wrap` block.
- Unresolved picks render in the roster outlined in red with a "missing card" note and can
  be removed; on save they ride along in the SRD bucket, since lookup checks both lists.
- The homebrew forms deep-copy the card into local `$state` on mount, so typing never
  mutates the row still sitting in the browser's list, and Cancel really cancels. Both show
  a live preview using the same detail component the browser renders.
- Editing a card **keeps its slug** — `UpdateCustom*` addresses by slug and only writes the
  name, so a rename can't orphan an encounter. The forms say so inline.
- Deleting leaves the browser mounted, so `CardBrowser` takes a `reloadToken` the parent
  bumps; saving unmounts the browser instead, and that path reloads for free.

**Done when:** you can build an encounter against a party, attach an environment, see the budget update live, and save it — feature parity with the web app plus environments. ✅

---

## Phase 3 — GM: Combat Runner (new)

**Goal:** the reason this is a desktop app — run the encounter live.

Schema (island A, cont.):
```sql
combats(id, encounter_id, fear INTEGER DEFAULT 0, active INTEGER, created_at, updated_at)
combatants(id, combat_id, adversary_slug, display_name,
           hp_max, hp_marked, stress_max, stress_marked, spotlight INTEGER)
countdowns(id, name, value INTEGER, max INTEGER, kind TEXT)
```

- [x] Start a combat from a saved encounter (spawns combatants from picks)
- [x] Per-combatant HP/Stress steppers, spotlight toggle, quick-view of card features
- [x] **Fear tracker** (0–12) at the combat level — the GM adjusts it by hand
- [x] **Countdowns / progress clocks** — add, tick up/down, delete
- [x] Surface the encounter's **environment** in the runner — impulses + features on hand as GM prompts during the scene
- [x] Auto-save state (so closing the app mid-fight is safe)

**Only players roll duality dice**, so the runner has no duality widget and nothing bumps
Fear automatically — the GM gains and spends it manually. The Hope/Fear roller belongs to
the Player module in Phase 5.

Notes on the schema:
- Tables are **plural** (`combats`, `combatants`, `countdowns`) to match everything else in
  the database, and every column is `NOT NULL` with a default — the frontend never sees a
  null where zero is meant. `adversary_slug` is the one genuinely nullable column.
- `combatants.adversary_slug` is **plain TEXT, not a foreign key**. SRD adversaries live in
  `data/` json and homebrew lives in `custom_adversaries`, so there is no single table to
  point at, and an ad-hoc combatant has no source card at all.
- `combatants.combat_id` is `ON DELETE CASCADE`; deleting a fight takes its roster with it.
  `combats.encounter_id` stays `ON DELETE SET NULL` so a combat outlives the encounter it
  came from, matching how Phase 2 keeps encounters that reference deleted cards.
- **Clamping happens in SQL**, not Go: Fear, HP, Stress and countdown values are all written
  as `max(0, min(ceiling, current + delta))`. Holding a stepper at either end is a no-op
  rather than an error to handle mid-scene; the `CHECK` constraints are the backstop.

Notes on what landed:
- Nothing returns a `db.*` row, same as Phase 2 — `gm.CombatView`, `gm.CombatantView` and
  `gm.Countdown` are the bridge types. `CombatantView` embeds the resolved `cards.Adversary`
  so the feature quick-view needs no second call.
- `StartCombat` runs in one transaction (the first in the codebase — `Service.tx` is the
  helper), stands down any active fight, and spawns one combatant per adversary per pick.
  Several of the same card get numbered; a lone one keeps the card's name.
- Card `hp`/`stress` are **text**, combatant maxima are integers. `parseStat` takes the
  leading integer run, so `"8"`, `"8 (5)"` and `"8-10"` all give 8, falling back to 1.
  All 129 SRD adversaries parse cleanly; homebrew hp/stress is unvalidated free text, so a
  bad card can't block a fight from starting — the GM corrects the combatant inline.
- A pick whose custom card was deleted still spawns, named after its slug with fallback
  stats and flagged `unresolved`, rather than silently dropping adversaries.
- **Spotlight is not exclusive** — `SetSpotlight` toggles one combatant and
  `ClearSpotlights` wipes the fight, so the GM can hold several at once.
- Auto-save needed no separate path: every mutator is `:one`, touches `updated_at`, and
  returns the updated row, so the UI renders what the database actually holds.
- Countdowns are still **global** until Phase 4 gives them a `campaign_id`.

Notes on the frontend:
- `CombatRunner.svelte` is the section. With no fight running it lists saved encounters to
  start from and past fights to resume; it calls `GetActiveCombat` on mount, so it reopens
  mid-combat where you left off.
- **HP counts up, not down** — `0/8` at full health, ticking to `8/8`, matching the Stress
  track beside it. The database always stored marked-ascending; only the display differed.
- The **Dice** section is reused inside the runner's rail via a `compact` prop rather than a
  second copy of the roll logic, and is opt-in behind a Show/Hide toggle remembered in
  `localStorage`. Die sizes are chips rather than a `<select>`, still served by
  `Roller.Sizes()` so the picker can't offer a die the backend rejects.
- `RollResult.svelte` tumbles through plausible values before settling on the real one —
  the roll is already resolved server-side, so the animation is theatre over a known
  outcome. Results also flash centre-screen for ~2s and then fade; the log below is the
  catch-up if you miss it. Everything honours `prefers-reduced-motion`.

Open: three sqlc params (`SetFear`, `SetCombatantVitals`, `UpdateCountdown`) infer as
`interface{}` because the arg sits inside `max(0, min(…))`, and surface as `any` in the
generated TypeScript. A `CAST(… AS INTEGER)` inside the clamp is the likely fix.

**Done when:** you can run a saved encounter end to end — mark damage, spend/gain Fear,
advance a countdown — and reopen exactly where you left off. ✅

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
      *(Linux needs `wails build -tags webkit2_41` on distros that ship webkit2gtk 4.1 and
      no 4.0 compat package — plain `wails build` fails at pkg-config. Bake the tag into CI.)*
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
