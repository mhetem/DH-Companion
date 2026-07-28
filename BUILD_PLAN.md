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
  picker plus a summary rail already filled the window it was designed against.
  *(The window opened at 1024×768 then; the built-in default is 1920×1080 now, with a
  1100×700 floor. Prose-heavy panes — card detail, the campaign page — carry a `max-width`
  so lines stay readable rather than running the full width.)*
- **Window size is a user setting**, stored in `settings` as `window_size` = `"WxH"` next to
  `last_role`. The header picker offers HD → 4K, and `shutdown` records whatever size the
  window was left at, so dragging the frame is remembered the same way picking a preset is.
  Sizes are clamped against `ScreenGetAll` — a preset larger than the display shrinks to fit
  and is marked unavailable in the picker rather than opening off-screen. `Screen.Size` is
  the logical-pixel space `WindowSetSize` works in; `Screen.Width`/`Height` are physical and
  deprecated. A malformed or too-small stored value is ignored so the app still opens.
- **UI scale is the separate knob for readability** (`ui_scale`, 80–200%). Window size and
  legibility aren't the same thing: a 4K window on a 4K screen makes everything physically
  *smaller*, because CSS pixels don't grow with the resolution. Scale is applied as a
  percentage on the root font size, which works because every length in the app is in `rem`
  — controls, gutters and the prose `max-width`s all grow with the text instead of it
  swelling inside fixed chrome. Only two `px` lengths existed and both are gone.
  The header applies it optimistically and rolls back if the write fails; a bad stored value
  falls back to 100% rather than rendering the app at some unusable size.
  *(Known edge: `@media` breakpoints resolve against the viewport and ignore root font size —
  `em` in a media query means the initial 16px, not ours — so a breakpoint can't respond to
  scale. The runner's grid gives the rail a `minmax` range instead of a fixed width, which
  covers the case that mattered.)*
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
campaigns(id, name, description, current_fear INTEGER DEFAULT 0, created_at, updated_at)
sessions(id, campaign_id, number, title, date, recap TEXT, created_at, updated_at)
notes(id, campaign_id, kind TEXT, title, body TEXT, created_at, updated_at) -- kind: npc|location|faction|lore|plot
session_encounters(session_id, encounter_id, created_at)  -- encounters prepped for a session
search(entity, entity_id, campaign_id, slug, title, body) -- fts5 virtual table
-- backfill: campaign_id on `countdowns` (clocks belong to a campaign)
-- backfill: campaign_id + session_id on `combats` (whose Fear, and which night it was run)
```
**Fear graduates here** from per-combat (Phase 3) to per-campaign — it persists between
fights, matching the rules. The runner reads/writes `campaign.current_fear` when a combat
belongs to a campaign.

- [x] Campaign CRUD; a campaign owns its Fear, countdowns, sessions, and notes
- [x] Session log: numbered sessions with title/date/recap; link the encounters run that session
- [x] Notes: typed entries (NPC / location / faction / lore / plot thread), markdown body
- [x] Cross-link — jump from a running combat to the campaign's NPCs/notes; from a session, open its encounters
- [x] **SQLite FTS5 full-text search** across notes + adversaries + environments (one virtual table)

Notes on the schema:
- **A campaign owns its sessions, notes, and countdowns** — those FKs are `ON DELETE CASCADE`
  and `campaign_id` is `NOT NULL` on `sessions`/`notes`. `combats.campaign_id` is the
  exception: nullable and `ON DELETE SET NULL`, so a fight outlives the campaign it was run
  in, matching how Phase 3 keeps a combat whose encounter was deleted.
- `countdowns.campaign_id` is nullable — clocks created before this phase have no campaign,
  and `ListUnassignedCountdowns` is how they stay reachable.
- `session_encounters` is keyed `PRIMARY KEY (session_id, encounter_id)`, so linking the
  same fight twice is a no-op (`INSERT OR IGNORE`) rather than a duplicate row. Both sides
  cascade — a link row with a deleted end is unreachable, not history worth keeping.
- Sessions are `UNIQUE (campaign_id, number)`; `NextSessionNumber` hands out the next one
  rather than the frontend guessing.
- `notes.kind` is constrained in SQL (`CHECK (kind IN (...))`), unlike the Phase 2 tier/type
  lists that live in `validate.go` — the set is closed by the rules and never came from SRD
  data, so there is nothing for Go to disagree with.
- Every new column is `NOT NULL` with a default, same rule as Phase 3. `campaigns.current_fear`
  mirrors `combats.fear` exactly (`NOT NULL DEFAULT 0 CHECK (… BETWEEN 0 AND 12)`).
- **Search is one fts5 table over mixed sources.** `notes` stay in sync through three
  triggers; adversaries and environments are *not* trigger-fed, because SRD cards live in
  `data/` json rather than a table. Go reindexes them (`ClearCardIndex` + `IndexCard`) from
  the `BrowseAdversaries`/`BrowseEnvironments` union, which already resolves the
  custom-shadows-SRD rule, so the index inherits it for free.
- `entity`/`entity_id`/`campaign_id`/`slug` are `UNINDEXED` — they are payload for the
  caller to route a hit back to its record, not search terms. `campaign_id` is what lets a
  search scope notes to one campaign while still matching every card.
- **The search read is hand-written Go, not sqlc.** sqlc 1.31.1 rejects every spelling of
  the fts5 idiom — `search MATCH ?`, the qualified and aliased forms, the table-valued
  `search(?)`, and a rowid subquery all fail with `column "search" does not exist`, because
  its SQLite parser has no notion of a virtual table's hidden name column. Only
  `title MATCH ?` parses, and that silently restricts the search to one column, which is
  worse than useless for note bodies. So `gm.Search` in `internal/gm/search.go` issues the
  query through `Service.conn` directly; only the index writes (`IndexCard`,
  `ClearCardIndex`) go through sqlc, and those generate fine.
- `matchExpr` quotes every token as its own phrase before it reaches `MATCH`. Raw input is
  not a valid FTS5 expression — `sabine-the` parses as a column filter and errors with
  `no such column: the`, and a stray `"` gives `unterminated string`. The last token also
  gets a `*`, so hits narrow while the GM is still typing.
- `CAST(… AS INTEGER)` around the clamped params in `AdjustCampaignFear`/`SetCampaignFear`
  is the Phase 3 "Open" note's proposed fix for those inferring as `interface{}`.

Notes on linking a fight:
- A combat carries **both** `campaign_id` and `session_id`. They answer different
  questions — whose Fear pool this fight spends, and which night it was actually run — and
  a fight can have the first without the second (a skirmish between sessions).
- **`session_encounters` and `combats.session_id` are not redundant.** The first is prep:
  encounters you lined up for a session. The second is history: fights that really happened.
  A session shows both lists, and they routinely differ — prepped fights the party walked
  around, improvised fights that were never an encounter.
- **A session links as many encounters as you like.** It always could — `session_encounters`
  is a join table and `LinkSessionEncounter` is an `INSERT OR IGNORE` — but the UI only
  offered a one-at-a-time dropdown, so linking four fights was eight clicks. The picker is a
  scrolling checkbox list of the not-yet-linked encounters now, committed in one press
  ("Link 3 encounters"). It stays one `LinkEncounter` call per pick rather than a bulk
  method: each insert is independent, so a failure part-way leaves the earlier ones linked
  and on screen instead of rolling them back.
- Picks are cleared when a different session is expanded, so they never ride along.
- Picking a session **implies its campaign**, and `resolveLinks` overrides whatever campaign
  the caller passed rather than letting the two disagree. `StartCombat` and `LinkCombat`
  share it, so starting a linked fight and linking one afterwards can't drift apart. The
  frontend's campaign select follows the session rather than fighting it.
- `combats.session_id` is `ON DELETE SET NULL`, same reasoning as `encounter_id`: deleting a
  session is editing your log, not erasing the fight. The combat stays in its campaign.
- Links are editable **mid-fight**, not just at the start — the runner's header carries a
  collapsed control that opens by default while the fight has no campaign.

Notes on the homebrew reference pane:
- The homebrew forms' right rail is now tabbed **Preview / Reference**, and Reference is
  `CardBrowser` with a `compact` prop that stacks the list above the detail. That is the
  third use of the compact-prop convention (`Dice`, `Notes`) and means the reference pane
  filters and searches exactly like the real browser, because it *is* the real browser.
- **"Use as template" copies everything except the name and slug.** Slugs derive from the
  name on create and are immutable after, so copying one in would either collide with the
  source card or, when editing, silently orphan encounters pointing at this one. The name is
  only filled in — as "X (copy)" — when the field is still empty.

Notes on the frontend:
- **WebKitGTK paints `select` and number spinners as native widgets** and ignores
  `background-color` on them, so they rendered as light system controls carrying the app's
  light text — unreadable, and the one thing on screen that didn't look like the app.
  `appearance: none` in `style.css` is what hands the painting back to CSS, at the cost of
  having to draw the dropdown arrow (an inline SVG data URI, since data URIs can't read CSS
  vars — that chevron is the one place `--muted` is hardcoded). The *open* dropdown is still
  the platform's, which is why `option` / `option:checked` / `option:hover` are styled
  separately. Worth knowing before the Player sheet grows its own selects.
- **Campaigns is the first nav section**, and it owns the campaign picker plus three tabs
  (Sessions / Notes / Countdowns) rather than three more nav entries — none of them mean
  anything without a campaign selected. The last campaign opened is remembered in
  `localStorage`, the same trick the runner uses for its dice panel.
- The campaign's `FearTracker` is **the same component the runner uses**, pointed at
  `AdjustCampaignFear`/`SetCampaignFear`. Between fights the GM adjusts Fear here; during a
  fight the runner's tracker writes the same row.
- `Countdowns.svelte` took a `campaignId` prop instead of being duplicated. With one it shows
  and creates that campaign's clocks; without one it shows `ListUnassignedCountdowns` — so
  clocks made before this phase, and fights run outside a campaign, still have a home.
- `Notes.svelte` does the same with a `compact` prop, mirroring how `Dice` is reused in the
  runner's rail. Compact is read-only and appears in the rail only when the fight belongs to
  a campaign — that is the "jump from a running combat to the campaign's NPCs" cross-link.
- The runner asks which campaign a fight is for **before** it starts, since that is what
  decides whose Fear pool gets spent, and it says so inline. The pick is remembered.
- Note bodies render through `markdown.js`, a ~90-line subset (headings, emphasis, code,
  lists, blockquotes) written rather than pulled in as a dependency. It escapes first and
  formats after, so a note containing `<script>` renders as text. Emphasis requires a word
  boundary and a non-space after the marker, so `3 * 4` and `snake_case_word` survive.
- **Search excerpts are escaped, not trusted.** `snippet()` wraps matches in `<mark>` but
  returns the surrounding text exactly as indexed, and note bodies are indexed raw — so
  `renderExcerpt` escapes everything and then lets only the highlight markers back through.
  Card text is stripped at index time; note text is not, which is why this matters.
- Search is its own nav section rather than a box inside Campaigns: it spans cards as well as
  notes, and the cards belong to no campaign. The campaign selector there narrows *notes*
  only, matching what `SearchInCampaign` does on the Go side.

Notes on what landed (service layer):
- Bridge types follow Phase 2/3: nothing returns a `db.*` row. `gm.Campaign`, `gm.Note`,
  `gm.SessionSummary`/`gm.SessionView` and `gm.SearchHit` are the surface, with `*T` for
  nullable columns. `SessionView` embeds `SessionSummary` and adds the linked encounters, so
  lists stay one query and only the detail view fans out — the same split as
  `CombatSummary`/`CombatView`.
- **`StartCombat` now takes a `campaignID *int64`.** Passing nil keeps the Phase 3 behaviour
  exactly: an unattached fight owns its own Fear.
- **Fear graduation is a routing decision inside `AdjustFear`/`SetFear`.** They look up the
  combat's `campaign_id` and, when set, write `campaigns.current_fear` instead of
  `combats.fear`; `hydrateCombat` reads it back the same way. The frontend keeps calling the
  same two methods with a combat id and never learns which table answered.
- `SaveSession` calls `NextSessionNumber` when the input has no number, so the GM never picks
  one. A collision on `UNIQUE (campaign_id, number)` comes back as "session N already exists
  in this campaign" rather than a constraint error.
- Note kinds are validated *and* normalised (`"NPC"` → `"npc"`) in `validate.go`, and
  `Service.NoteKinds()` serves the list to the frontend — same one-definition rule as
  `Roller.Sizes()`.
- `ReindexCards` runs at startup and after every custom-card create/update/delete, rebuilding
  all 148 cards in one transaction. That is cheap enough to beat tracking deltas, and it
  reuses `BrowseAdversaries`/`BrowseEnvironments`, so the index inherits the
  custom-shadows-SRD rule instead of restating it.
- Indexed card text strips HTML before it reaches fts5. Feature descriptions carry inline
  `<strong>`/`<em>`, and without stripping, every card would match a search for "strong".
- `encounterSummary` was lifted out of `ListEncounters` so sessions can reuse it.
- The three sqlc params that inferred as `interface{}` in Phase 3 (`SetFear`,
  `SetCombatantVitals`, `UpdateCountdown.Value`) are `int64` now — the `CAST(… AS INTEGER)`
  fix that the Phase 3 "Open" note proposed works, and it is applied to all of them.

**Done when:** you can create a campaign, log a session with a recap linked to an encounter
you ran, keep typed NPC/location notes, and full-text search across everything. ✅

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
- [x] App icon + name
- [x] Empty states
- [x] Keyboard shortcuts for combat/dice
- [x] README + first-run notes

Notes on the name:
- **The app is called Hilt.** The rename is display-only: the window title, the page title,
  `wails.json`'s `info.productName`, the header, the role picker and the README all say
  Hilt, while the repo, the Go module path (`github.com/mhetem/DH-Companion`), the binary
  and the data directory (`~/DH-Companion/data.db`) are unchanged — renaming the data
  directory would strand every existing database, and none of the rest is user-visible.
- `frontend/src/assets/images/` holds the two source SVGs: `hilt-logo.svg` (the mark) and
  `hilt-app-lockup.svg` (mark + wordmark). The header and the role picker both `import` the
  lockup, so there is one copy of the artwork rather than a component-inlined duplicate.
  Vite inlines it as a data URI at build time — it is well under the 4 kB threshold.
- `build/appicon.svg` is the icon source: the same mark on a tighter viewBox, since the
  logo carries an 18-unit transparent margin that would waste an icon's canvas. It is
  rasterised to `build/appicon.png` at 1024² and to `build/windows/icon.ico` as a
  seven-entry PNG-payload ICO (16→256). Regenerate both from it rather than editing the
  bitmaps.

Notes on the keyboard:
- `lib/keys.js` is the whole mechanism: an `isTyping` guard on the event target, a `combo`
  normaliser (`h`, `H` for shift, DOM names for the rest, and ctrl/meta/alt chords left to
  the platform), and `onKeys(read)`. **`read` is a function, called per keypress**, so a map
  built out of runes sees current state instead of what it closed over on mount.
- **Selection is a new concept in the runner** — the keyboard's cursor over the roster,
  deliberately *not* spotlight. Spotlight is a rules state the GM sets on purpose and can
  hold on several combatants at once; overloading it as the keyboard target would have
  conflated the two. The cursor is drawn in `--gold` so it reads as chrome and can sit on a
  spotlit (`--fear`) row without the borders fighting. Selection is keyboard-only: mouse
  users click the steppers directly, and a hint line above the roster says which keys act
  on what.
- **`--gold` earned its keep here.** It was added for the brand frame and turned out to be
  the only accent not already spoken for by the rules colours.
- The runner gates its steppers on `busy`, mirroring the buttons being `disabled` mid-write
  rather than racing concurrent marks against the same row.
- **A modal counter alone could not make the shortcut sheet exclusive.** A microtask
  checkpoint runs between two window listeners, so Svelte flushed the state change that
  closed the sheet — and with it `modal.close()` — before the next listener was called, and
  one Escape both dismissed the sheet and dropped the roster selection behind it. The fix is
  that modal owners listen in the **capture** phase and `stopPropagation()` there, which is
  what actually keeps the event from the handlers underneath. The counter stays as the
  second line of defence.
- Dice claims `Escape` only while a flash is up, so it stays free for the runner's "drop the
  selection" when both are mounted and nothing has been rolled. That is the only key the two
  maps both want; `s` is Stress in the runner, which is why disadvantage is `z` and not `s`.
- `SHORTCUTS` in `keys.js` feeds the `?` sheet, but the components bind the key strings by
  hand — **nothing checks that the two agree.** Key caps also sit inline on the controls
  themselves, which is the discoverability that matters; the sheet is the reference.

Notes on the palette:
- The custom properties in `style.css` are retuned to the logo: `--bg`/`--panel` are its
  purple-tinted near-black field, `--text` is its parchment `#ece3d0`, and `--gold` /
  `--gold-deep` are its two golds, added as brand-frame tokens distinct from `--hope`.
  `--hope` and `--fear` stay the rules colours they were.
- **Every foreground clears WCAG AA (4.5:1) against all three surfaces** — `--bg`,
  `--panel` and `--panel-2`. That is what set the exact values: the logo's crimson
  `#b6412e` only reaches 2.8:1 as `--danger` on `--panel-2`, so `--danger` is a lightened
  `#dd7161` and the deep crimson stays in the artwork. `--fear` was lightened from
  `#9b6bd6` for the same reason (4.06 → 4.83).
- `main.go`'s `BackgroundColour` matches `--bg`. It was still the Wails template's slate,
  which flashed on launch before the frontend painted.
- The `select` chevron data URI has to restate `--muted` as a literal, so it moved with it.

Notes on the frontend:
- **The GM nav follows the order of use**, not the order the phases were built in:
  Campaigns → Parties → Encounters → Combat Runner, then the reference sections
  (Adversaries, Environments), then the two utilities (Dice, Search). Prep runs top to
  bottom — you need a party before the budget meter means anything, and an encounter before
  there is a fight to run.
- **The role picker's emoji were tofu.** 🎲 and 🛡️ render as boxes on any machine without
  an emoji font — this one has none (`fc-match emoji` falls through to DejaVu Sans), and a
  GTK build can't assume one. They are inline SVG now, tinted `--fear` and `--hope` so the
  two sides read as the two sides. The `✕` buttons elsewhere are Dingbats, which DejaVu
  covers, so they stay.
- **Checkboxes joined `select` and the number spinners** in the WebKitGTK `appearance: none`
  block. Phase 4 caught the other two; the Dice pane's advantage/disadvantage boxes were
  still painting as bright system squares on a dark panel. Same data-URI caveat — the tick
  restates `--bg` as a literal, because a data URI can't read CSS vars.
- `EmptyState.svelte` is the full-pane "nothing here yet" block — centred, with a hint and
  a small gold hilt. It carries the hilt as inline SVG rather than the logo, because the
  full mark is a dark tile with a gold frame and turns to mush at 3rem. It is in
  `lib/` rather than `lib/gm/` since the Player panes will want it in Phase 5.
- The inline one-liner **`.empty` is now a single rule in `style.css`**. Ten components had
  each grown their own copy at 0.8rem or 0.85rem; they keep only their own spacing and
  structural resets now. The two cases that were a bare line in an otherwise empty pane —
  Campaigns and Encounters — use `EmptyState` instead; the rest sit under a heading in a
  section that is already doing the explaining, where a line is right.

---

## Phase 7 — Stretch: LAN session sharing

**Goal:** the senior-signal feature — real-time, read-only session view for players.

- [ ] GM's app runs a lightweight WebSocket server (opt-in, bound to the local network)
- [ ] Broadcasts live **Fear, countdowns, and spotlight** as they change
- [ ] Players on the same Wi-Fi connect (URL/QR) to a read-only session view
- [ ] Graceful degrade — everything still works fully offline if nobody connects

**Done when:** a second device on the LAN watches Fear and countdowns update live as the GM runs a fight. Clearly optional; nothing else depends on it.

---

## Licensing note

`LICENSE` is MIT **for the software only**, with the scope spelled out at the top of the
file; `NOTICE.md` carries the DPCGL attribution and the per-directory breakdown. The split
matters because `data/` is not the author's to relicense — DPCGL §3 keeps ownership with
Critical Role, and an unqualified MIT file would purport to grant rights over it.

**Known and accepted:** `data/embed.go` compiles the SRD json into the binary, so cutting a
public release Shares Public Game Content. DPCGL §1.9 Permitted Formats are print,
live-stream/video, podcasts and DRP-whitelisted VTTs, and it expressly excludes video games
and other unlisted media — a general desktop app is not on the list.

**Staying free does not resolve that**, which is the easy thing to get wrong here. §2.1(b)
grants the right to "produce, reproduce, Share *and sell*... solely in the Permitted
Formats" — the format limit sits on Sharing, and §1.8 defines Sharing as making content
available to the public by any means, price irrelevant. The exemption that does apply is
§1.8's, and it needs *private* **and** non-commercial: "private, non-commercial play among
friends, family, or gaming groups in a personal setting." Running this yourself, or handing
it to your table, is untouched. A public repo with release binaries is not a personal
setting even when it is free.

What staying non-commercial *does* remove is §4.2: no front-cover Darrington Press
community logo, no title-page trademark statement. Only §4.1's attribution applies, and
`NOTICE.md` carries it, including the §4.1(e) statement of the modification we made —
transcribing the SRD into json.

The decision is to ship anyway and deal with it if DRP ever objects. The two ways out, if
that changes: ask DRP to whitelist the app, or move the SRD out of the binary so the app
reads a `data/` directory the user supplies and the release ships code only. The second is
a real change to `internal/srd` and `data/embed.go`, not a docs fix.

## Data sourcing note
Adversaries you already have. **Environments** (GM side), plus domain cards, class
features, ancestries, and communities (Player side) need to be assembled into `data/`
json (same loader pattern as adversaries). This is content work, not code — it can proceed
in parallel with Phases 2–5. All of it has landed in `data/` ahead of the Phase 5 code.
For what may be redistributed, see the licensing note above — the constraint is the DPCGL's
Permitted Formats, not the sourcing.

## Suggested order
Phase 0 → 1 → 2 → 3 → 4 = a complete, usable **GM tool** (build, run, and record a
campaign). Then Phase 5 (Player) slots in independently, and Phase 6 (CI/release) rides
along from Phase 0 onward. Phase 7 (LAN sharing) is an optional capstone.
GM side reaches "better than the web app" by the end of Phase 4.
