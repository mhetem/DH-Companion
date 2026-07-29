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
| Auth | **None.** Role is a switchable view (GM / Player), not a login. Every launch opens on the picker; switch anytime without losing data |
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
- [x] `settings` table (`key TEXT PRIMARY KEY, value TEXT`) — window size and UI scale
- [x] Frontend: on launch, always show the role picker
- [x] **Persistent "switch role" affordance** (header) available anytime — flips the view, touches no data
- [x] Bound methods: `GetRole()`, `SetRole(role)` — *removed; see below*

**Done when:** you can switch GM↔Player at will; data built in either mode survives the switch and relaunch. ✅

Notes on the role, revised:
- **The role is no longer remembered.** Every launch opens on the picker. `last_role`,
  `GetRole` and `SetRole` are gone, and switching from the header is now purely a frontend
  state change — nothing about a role reaches the database. The `settings` table stays for
  `window_size` and `ui_scale`.
- The reasoning that made it a stored preference — "returning users shouldn't see the picker
  flash" — assumed one person per install. Two people sharing a machine, or one person who
  GMs one night and plays the next, get the choice put in front of them instead of landing in
  whichever shell they closed last. A pre-existing `last_role` row is simply ignored.

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
  `ui_scale`. The header picker offers HD → 4K, and `shutdown` records whatever size the
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
characters(id, name, pronouns, class_slug, subclass_slug, subclass_mastery,
           multiclass_slug, multiclass_subclass_slug, ancestry_slug, community_slug,
           level, proficiency,
           agility, strength, finesse, instinct, presence, knowledge, marked_traits,
           hp_max, hp_marked, stress_max, stress_marked, hope,
           evasion, armor_score, armor_marked, threshold_major, threshold_severe,
           gold_handfuls, gold_bags, gold_chests, beastform_slug,
           background, connections, notes, created_at, updated_at)
character_domain_cards(id, character_id, card_slug, location)  -- 'loadout' | 'vault'
inventory_items(id, character_id, name, kind, qty, equipped, detail)
experiences(id, character_id, name, modifier)
character_levelups(id, character_id, level, tier, choices TEXT, summary)
companions(id, character_id UNIQUE, name, evasion, damage_die, attack_range, attack,
           stress_max, stress_marked, experiences TEXT, upgrades TEXT, notes)
```

- [x] SRD data: domain cards, class/subclass features, ancestries, communities (source + load into `data/`)
- [x] Character CRUD + create wizard (class → subclass → ancestry → community → traits)
- [x] Sheet view: traits, live-tracked HP/Stress/Hope, evasion/armor/thresholds, gold
- [x] **Loadout (max 5) vs. Vault** domain-card management (drag/move, enforce cap)
- [x] Inventory + equipment
- [x] Experiences (name + modifier)
- [x] Leveling flow across tiers (1, 2–4, 5–7, 8–10)
- [x] **Duality dice** widget wired to Hope — a roll with Hope grants a Hope; apply Experience/trait modifiers
- [x] **Beastform** (druid) and the **companion sheet** (Beastbound ranger) — the two
      subclass features that need pages of their own

Notes on the schema:
- Tables are **plural** and every column is `NOT NULL` with a default, same two rules as
  Phase 3. There is not a single nullable column in island B — an unbuilt character has
  `class_slug = ''`, not `NULL`, so the frontend never sees a null where a blank is meant.
- **Five columns the original sketch didn't have, each because a rules feature needs it.**
  `proficiency` is the damage-dice count and rises on tier entry. `subclass_mastery` (1–3)
  is which of Foundation/Specialization/Mastery you've unlocked, so the sheet can show only
  the cards you actually have. `marked_traits` is a json array — the trait advancement marks
  what it raises and the marks clear on entering a new tier, which is the whole reason a
  trait can't be raised twice in a tier. `multiclass_slug` / `multiclass_subclass_slug` carry
  the tier 3+ multiclass advancement, and they're what widens the domain-card pool.
- **Gold is three columns, not one.** The sheet counts handfuls, bags and chests, ten to the
  next, so `CHECK (gold_handfuls BETWEEN 0 AND 9)` is the carry rule written down.
- `hope` has no `hope_max` column. Six is a constant of the rules, not a per-character value,
  so it's `player.HopeMax` and rides out on the view as `hopeMax` — exactly how `FearMax`
  works on the GM side.
- **Clamping happens in SQL**, continuing Phase 3: HP, Stress, Hope, armor slots, gold and
  item quantities are all written as `max(lo, min(hi, …))`, with `CAST(… AS INTEGER)` around
  the parameter from the start — the Phase 3 "Open" note's fix, applied on the way in rather
  than after the fact. The `CHECK` constraints are the backstop.
- Every child table is `ON DELETE CASCADE` on `character_id`: unlike a combat, which outlives
  its encounter, a loadout has no meaning without the character holding it.
- `character_domain_cards` is `UNIQUE (character_id, card_slug)` — one copy of a card per
  character, and `location` is the only thing that moves.
- `character_levelups` is `UNIQUE (character_id, level)` and stores the raw choices as json.
  **That json is not a log, it's the source of truth for the slot counts** — an advancement
  can be taken a limited number of times *per tier*, so answering "can I take this again"
  means reading back what was taken at every earlier level in the same tier.

Notes on the leveling rules:
- **The advancement table is data, not code**: `data/leveling.json`, loaded by the same
  `internal/srd` loader as everything else in `data/`, into `srd.Leveling` / `srd.Tier` /
  `srd.Advancement`. It's the only reference dataset that isn't card-shaped, which is why the
  types live in `srd` rather than `cards`.
- An advancement is `{slots, cost, effect}`. `slots` is how many times it can be taken in the
  tier; `cost` is how many of the level's two advancement picks it consumes (multiclass is
  the only one that costs two). `effect` is a flat struct of mechanical deltas, so adding an
  option is a data edit and the apply code doesn't change.
- **The numbers in that file are transcribed from the SRD and worth a second pair of eyes.**
  The structure — two advancements a level, +1 to both thresholds, a domain card at or below
  your level, and the tier-entry achievements at 2/5/8 — is what the code enforces; the
  per-option slot counts are just values in the file, so correcting one is a one-line data
  change with no code to touch.
- `ApplyLevelUp` runs in one transaction and applies things **in rules order**: tier
  achievements first (clear marked traits, +1 Proficiency, new Experience), then the two
  advancements, then thresholds, then the domain card. The order matters — clearing marks
  first is what lets a trait raised in the last tier be raised again in this one.
- **The plan is computed, not remembered.** `PlanLevelUp` returns each option with
  `used`/`remaining`/`available` plus a `reason` when it's off, so the UI greys out
  "upgraded subclass card" at mastery and "multiclass" once you've multiclassed without
  restating any of that logic in JS. Everything it reports is re-validated on apply — the
  frontend is not trusted with the caps.
- **Known gap:** the plan's `availableCards` is computed from the domains you have *now*, so
  picking multiclass and a level card in the same level-up won't offer the new class's
  domain that level. The backend accepts it (it validates against the post-multiclass
  domains); only the picker is a level behind.

Notes on the two subclass pages:
- **Beastform and Companion are nav sections that only exist for the characters that have
  them.** `PlayerShell` reads the open character and splices them in — a bard has no
  Beastform page rather than an empty one. Everything else in the nav is fixed.
- **`beastform_slug` is a column on `characters`, not a table.** You are in one form or
  none, and the form itself is SRD data — so the character stores which one and the
  catalog supplies the rest. Transforming marks a Stress (the class feature's cost) through
  the existing clamped `AdjustCharacterStress`; dropping out is free, which is why it isn't
  a straight toggle.
- Effective Evasion while transformed is computed, not stored: base + the form's
  `evasionBonus`. The trait bonus and attack line are shown rather than folded into the
  sheet's numbers — they apply to specific rolls, and quietly rewriting a trait would be
  worse than showing what to add.
- The beastform list stays visible for non-druids as reference; only the Transform buttons
  are gated, and `Transform` re-checks the class and the tier server-side.
- **The companion is its own table, keyed `character_id UNIQUE`** — one companion per
  ranger, cascading on delete. Its Experiences and taken level-up options are JSON TEXT
  columns rather than two more tables: both are short lists edited as a unit, and the
  project already stores card features and encounter picks this way.
- **The level-up options are a checklist, not an engine.** Ticking *Vicious* records that
  you took it; it doesn't step the damage die, because "increase by one step" is a choice
  between die and range, and *Aware*/*Resilient* would then need un-applying on untick.
  The pane says so and the fields are editable.
- `RollCompanionDamage` is its own method because the dice count is the **ranger's**
  Proficiency while the die is the **companion's** — the rule that makes commanding them
  worth doing, and easy to get wrong by hand.
- **Gold moved to Inventory.** It is what you carry, it sits next to the gear it buys, and
  the sheet's stat cards were the wrong shape for three steppers — `Gold.svelte` is the
  extracted component so it renders identically wherever it lands.

Notes on the Wails binding warnings:
- **`cards.Class` is `cards.CharacterClass`.** Wails' TypeScript generator lowercases every
  type name and compares it against the JS reserved-word list, so `Class` tripped
  *"Usage of reserved keyword found and not supported"*. It only ever warned — the generated
  `export class Class` is legal TypeScript — but the rename costs nothing and the build is
  quiet. Nothing user-visible changed; `Catalog.Class(slug)` keeps its name, since the check
  is on type names only.
- **`player.PlanOption` no longer embeds `srd.Advancement`.** Wails flattened the embedded
  fields correctly but couldn't resolve `srd.AdvancementEffect` — no bound method returns an
  `srd` type directly, so it isn't in `KnownStructs` — and emitted `effect: any` with a
  *"Not found"* line. The option now spells its fields out and carries a `player.Effect`
  mirroring the SRD one, converted by `effectView`. Runtime JSON is identical; the generated
  types are real instead of `any`.

Notes on what landed (service layer):
- `internal/player.Service` is bound as its own Wails struct alongside `gm.Service`, so the
  frontend calls `window.go.player.Service.*`. `player.Attach` is a package function for the
  same reason `gm.Attach` is — a method would publish `context.Context` and `*sql.DB` into
  the generated TypeScript. Both `startup` and `portable.go`'s `reopen` attach it, so a
  database restore re-points the Player module too.
- Nothing returns a `db.*` row, same as every phase before it. `player.Character`,
  `player.Sheet`, `player.Loadout`, `player.Item`, `player.Experience`, `player.LevelUpPlan`
  and `player.Roll` are the bridge types. `Sheet` embeds `Character` and adds the resolved
  class/subclass/ancestry/community cards plus loadout, vault, gear and Experiences — the
  same summary/detail split as `CombatSummary`/`CombatView`.
- **A character knows two domain cards at level 1**, not five. The loadout cap (5) and how
  many cards you own at all are different limits, and only the first was enforced at first.
  `Loadout` now carries `held` and `allowance`, where allowance is `2 + one per level-up +
  one per extra-card advancement taken` — derived from the level-up records rather than
  stored, so it can't drift from the history. Removing a card frees the slot back up, and a
  character who somehow holds more than their allowance simply can't add another.
- **Slugs are resolved, never stored twice.** A character holds `class_slug`; the name, the
  domains and the spellcast trait are looked up in the catalog on the way out. A domain card
  whose slug no longer resolves comes back `unresolved: true` rather than failing the sheet,
  matching how Phase 2 handles a deleted custom card.
- HP and Evasion default from the class card (`startingHitPoints`, `startingEvasion`) when
  the wizard leaves them at zero. Those are text in the SRD json, so `leadingInt` takes the
  leading digit run — the same problem `parseStat` solves for adversaries in Phase 3.
- **Armor Score and both thresholds have no defaults**, because in the rules they come from
  the armor you're wearing and `data/` has no armor dataset. The wizard asks for them
  outright and says where to read them off.
- **Only one primary weapon, secondary weapon or piece of armor can be equipped at a time.**
  `UnequipInventoryKind` runs before the equip, so the exclusivity is enforced on the way in
  rather than validated afterwards; consumables and plain items are unlimited.
- `AddClassItems` pulls the class's starting kit straight off the card, so a new character
  isn't typing in two items the SRD already lists.
- **Rest is a choice of downtime moves, not a button that clears things.** `RestMoves(long)`
  serves the list — Tend to Wounds, Clear Stress, Repair Armor, Prepare on a short rest; the
  "all" versions plus Work on a Project on a long one — and `RestAllowance(long)` the budget
  (2 short, 3 long). `Rest(RestInput)` applies exactly the moves you picked and nothing else,
  so a short rest spent on Repair Armor and Prepare leaves your Hit Points where they were.
  A move can be taken **more than once**: two goes at Tend to Wounds is a legal short rest, so
  `moves` is a list with repeats rather than a set.
  Each move returns its own prose outcome with the roll that produced it (`1d4 + tier`), which
  is why `RestResult.Outcomes` is a list and not one sentence. An earlier pass had `Rest(id,
  long)` clear HP and Stress automatically; that invented a rule the SRD doesn't have.
- **`RollDuality` is the one method that both rolls and writes.** It resolves the roll, then
  writes the Hope and Stress it earned in the same call and returns the updated character, so
  "a roll with Hope grants a Hope" can't be forgotten by a caller. A critical grants a Hope
  *and* clears a Stress; a roll with Fear grants nothing, because the Fear pool is the GM's
  and this app never reaches across the two islands.
- `Roll` carries `modifierParts`, so the UI can show *why* the modifier is +4 rather than
  just that it is. The roller in `internal/dice` stays pure and untouched — this is a wrapper
  around `DualityDiceRoll`, not a second implementation.
- `RollDamage` uses the character's Proficiency as the dice count, and a critical adds the
  maximum value of those dice on top of the roll.

Notes on the frontend:
- **Which character is open is shared state, not a prop.** Five nav sections work on one
  character, so `player/active.svelte.js` exports a `$state` object backed by `localStorage`
  — the same "remember the last thing" trick Campaigns uses, but as a rune so switching
  sections doesn't drop the selection. Characters is the first nav entry because everything
  after it is empty until something is picked there.
- The wizard is seven steps (Identity → Class → Subclass → Ancestry → Community → Traits →
  Defenses) and doubles as the edit form. Each step shows the SRD text for what you're
  choosing next to the picker, through the **same `FeatureList` the GM browsers use** — class
  features, subclass cards, ancestry and community features are all `cards.Feature`, so there
  was nothing to write.
- Picking a class clears the subclass. A subclass belongs to exactly one class, so leaving it
  set would silently produce an invalid pair the backend then rejects.
- The Traits step checks the **spread, not the individual values**: sort the six and compare
  against sorted `[2,1,1,0,0,-1]`. That accepts any assignment of the array and rejects
  spending it twice, without caring which trait got what.
- `ResourceTrack` is the pip row behind HP, Stress, Hope and armor slots. **Marked counts up**,
  the same direction the GM runner settled on in Phase 3, and clicking the pip you're already
  on clears down to it rather than being a no-op.
- Every vitals mutator returns the whole character and the pane assigns it, so the sheet
  renders what the database holds — Phase 3's auto-save property, inherited for free.
- **Domain Cards is three columns, not drag-and-drop**: loadout, vault, and the pool of cards
  you can still take. Buttons move a card between the first two; the cap is shown as a tally
  that turns gold when full, and Equip is disabled rather than failing. The pool comes from
  `AvailableDomainCards`, which already applies "your domains, at or below your level, not
  already held", so the UI never offers a card the backend would refuse.
- **A disabled "Level up" button now says why.** `blockers` is a list of what's still
  missing in plain words ("Choose 1 more trait", "Name your new Experience"), rendered above
  the footer and, per pick, inside the pick itself; `complete` is just "no blockers". The
  gate was correct before but silent, and a two-trait advancement with one trait chosen
  looked finished.
- **The two domain-card selects render every card and `disabled` the ones already spoken
  for**, rather than filtering them out of the list. Filtering made each select's option list
  depend on its own bound value, which is the classic way to desync a `<select>` binding —
  options being removed and re-added under a live selection. A stable list can't.
- `LevelUp` is a `Modal`, and it renders the advancement table as its printed slot boxes —
  filled for what earlier levels used, filled again for what you're picking now. Options that
  need a follow-up choice (which traits, which Experiences, which card, which class) grow that
  choice inline under "Your picks" rather than opening another dialog.
- The Dice pane takes the `compact` prop convention that `Dice`, `Notes` and `CardBrowser`
  established. Results flash centre-screen and fade, reusing `RollResult` from the GM side —
  the tumble is theatre over an already-resolved roll, and it honours `prefers-reduced-motion`
  because that component already did.
- **The sheet hosts the dice panel in a collapsible block**, remembered in `localStorage` the
  same way the runner remembers its own. Standalone it loads its own character; embedded it
  takes `character` and `experiences` as props and reports writes back through `onupdate`, so
  the sheet and the panel can't drift apart when a roll spends Hope or clears Stress. The
  panel hides its own Hope/Stress tracks when embedded, since Resources is already on the page.
- **Every trait on the sheet is a roll button.** Clicking one opens the dice block if it's
  closed and calls the panel's exported `rollTrait`, so the modifier breakdown, the flash and
  the log are the same ones the dice section shows rather than a second, thinner roll path.
  The block is rendered whether open or closed — collapsed is a `hidden` attribute, not an
  unmount — so `bind:this` is always live and there is no tick to wait on.
- **`SrdText` is the one place SRD prose is rendered.** Catalog descriptions carry inline
  `<strong>`/`<em>` exactly like feature text does, and rendering them through normal
  interpolation printed the tags on the page — visible on Mixed Ancestry and on 151 of the
  domain cards. It is deliberately *not* used for anything the user typed: homebrew
  descriptions, a character's background and note bodies all stay escaped.

**Done when:** you can build a character, manage loadout/vault, roll with your traits, and track resources in play. ✅

---

## Phase 6 — CI, Packaging & Release

> Set up the CI workflow back at Phase 0 and let it grow with the project — it's listed
> here only because release automation depends on the app being buildable. Don't leave CI
> to the end.

- [x] **CI (GitHub Actions):** `gofmt`, `go vet`, `go test ./... -race`, and `wails build` on
      all three platforms for every push/PR; status badge in README
      *(Linux needs `wails build -tags webkit2_41` on distros that ship webkit2gtk 4.1 and
      no 4.0 compat package — plain `wails build` fails at pkg-config. Baked into CI.)*
- [x] **Cross-platform release automation:** on tag, build Win/macOS/Linux binaries via Wails and attach to a GitHub Release
- [x] **Update check** — *not* auto-update; see the note below on why that bullet was wrong
- [x] **Data portability:** DB backup/export + import, plus **whole-library JSON export/import**
- [x] **Shareable homebrew codes:** compress+base64 a custom adversary/environment into a paste-able string others can import (versioned format)
- [x] App icon + name
- [x] Empty states
- [x] Keyboard shortcuts for combat/dice
- [x] README + first-run notes

Notes on CI and release:
- Two workflows. `ci.yml` runs `gofmt -l` (failing on any unformatted file), `go vet`,
  `go test ./... -race`, then a full `wails build` on Linux, Windows and macOS —
  `fail-fast: false`, so one platform breaking still reports the other two.
- `release.yml` fires on a `v*` tag, builds `linux/amd64`, `windows/amd64` and
  `darwin/universal`, packages each with `LICENSE`/`NOTICE.md`/`README.md`, publishes
  `SHA256SUMS.txt`, and attaches everything to a GitHub Release.
- **The Linux webkit tag is baked into both.** Ubuntu runners ship webkit2gtk 4.1 with no
  4.0 compat package, so a plain `wails build` fails at pkg-config; the Linux branch passes
  `-tags webkit2_41` and the others do not.
- The tag is stamped in with `-ldflags "-X main.version=$GITHUB_REF_NAME"`. Untagged builds
  report `dev`, which the update check treats as uncomparable rather than as "out of date".
- Release notes say the builds are unsigned. There are no signing certificates, so macOS
  Gatekeeper and Windows SmartScreen will both warn on first run, and that is better said
  than discovered.

Notes on installers:
- Every platform gets a **real installable artifact**, not just a binary in an archive:
  an NSIS `setup.exe`, a drag-to-Applications `.dmg`, and both a `.deb` and an AppImage.
  The plain archives are still attached for anyone who would rather not install.
- **Windows is Wails' own** — `wails build -nsis` uses the already-scaffolded
  `build/windows/installer/`, emitting `build/bin/DH-Companion-amd64-installer.exe`. The
  installer's display name is `info.productName` ("Hilt") while the file name comes from
  `name` ("DH-Companion"), which is why the workflow renames it on the way to `dist/`.
  Its uninstaller clears the WebView2 data path, *not* `~/DH-Companion` — campaigns survive
  an uninstall.
- **The Windows installer has a components page.** Both shortcuts are their own optional
  section, so ticking neither is the "no shortcut" case and needs no third option. The app
  section is `SectionIn RO` — not optional — and `wails.setShellContext` moved into each
  shortcut section so they still land in the all-users locations under a machine install.
- **The uninstaller offers to delete your data, unticked by default (`/o`).** It is a
  separate `un.` section on `MUI_UNPAGE_COMPONENTS` rather than a message box, so it reads
  as a deliberate choice rather than a dialog to dismiss. Losing a campaign log to an
  absent-minded uninstall is far worse than leaving a database behind, so the safe option is
  the default. `APP_DATA_DIR` is defined once at the top of `project.nsi` and mirrors
  `os.UserHomeDir() + "DH-Companion"` in `app.go` — **if the data directory ever moves, that
  define has to move with it.** A `DH_DATA_DIR` override is invisible to the installer.
- `project.nsi` is a Wails *template*: it is only regenerated when missing, so these edits
  persist — but deleting the file to "reset to defaults" silently throws them away.
- **`wails_tools.nsh` is the opposite and must stay in its `{{.Name}}` placeholder form in
  git.** Every `-nsis` build rewrites it with the current ProjectInfo values, so a careless
  `git add -A` commits a hardcoded `INFO_PRODUCTVERSION`. Since it guards each define with
  `!ifndef`, a stale committed value would win over whatever CI stamps into `wails.json` and
  every release would carry the wrong version. Check `git diff` on this file before
  committing; the answer is almost always to restore it.
- `-nsis` also downloads `installer/tmp/MicrosoftEdgeWebview2Setup.exe` (1.8 MB) to embed
  the WebView2 bootstrapper. It is gitignored — it is a fetched third-party artifact, not
  source.
- **macOS is `hdiutil`**, no third-party tooling: stage the `.app` next to an `/Applications`
  symlink and compress with `-format UDZO`. `ditto -c -k --keepParent` makes the zip, since
  plain `zip` mangles bundle metadata.
- **Linux is `build/linux/package.sh`**, deliberately a script rather than inline YAML so it
  can be run and verified locally — which it was: the `.deb` was built and inspected with
  `dpkg-deb`, and the AppImage was extracted *and launched*, confirming it runs its
  migrations and creates a database.
- **The version has to be stamped into `wails.json` before the build**, not just into
  ldflags. NSIS metadata and the macOS bundle read `info.productVersion` from there;
  `main.version` (ldflags) is only what the in-app update check compares.
- **`info.productVersion` must be a bare numeric `X.Y.Z`.** `project.nsi` builds
  `VIProductVersion "${INFO_PRODUCTVERSION}.0"`, and NSIS rejects anything that isn't four
  numeric parts — so `v0.1.0-rc1` fails with *invalid VIFileVersion format*. The first
  stamping step only stripped the leading `v`, which meant **every prerelease tag would have
  failed the release**; it now pulls out the numeric triple and drops any suffix. The full
  tag still reaches the app through `main.version`, so the UI and the update check show
  `v0.1.0-rc1` while the Windows resources carry `0.1.0`.
- **`wails build -nsis` does not reliably fail when the installer isn't produced.** With
  NSIS present but the script broken it exits 1; with `makensis` *absent* it prints
  `Warning: Cannot create installer: makensis not found` and **exits 0**. The v0.1.0-rc1
  release hit exactly that: `windows-latest` ships no NSIS, the build step went green, and
  the job only fell over later trying to rename a file that was never created. The workflow
  now installs NSIS with choco, asserts `makensis` is on PATH before building, and tests for
  the installer immediately after — the exit code is not trusted.
- **`hdiutil create` intermittently returns "Resource busy" on the macOS runners.** It took
  down the same release. The disk image is now staged inside the workspace rather than
  `/var/folders`, built as HFS+ rather than the APFS default, and retried four times before
  the job gives up.
- Both of those were the two steps that could not be exercised locally, and both failed on
  the first real tag while the locally-verified Linux job passed. Worth remembering the next
  time something "should be fine".
- **Windows cross-compiles from Linux.** Wails' Windows target is pure Go (the WebView2
  loader is `winloader`, no CGO), so `wails build -platform windows/amd64 -nsis` works on a
  Linux box with `makensis` installed, and produces a genuine PE32+ binary plus a working
  NSIS installer. Useful for testing without waiting on a tag; the Windows runner is still
  the canonical build.
- The `.deb` puts the SVG in `hicolor/scalable/apps` and the PNG in `pixmaps`. The first
  attempt filed the 1024x1024 `appicon.png` under `512x512/apps`, which claims a size it
  isn't; `pixmaps` is the legacy fallback and carries no size contract.
- `package.sh` **skips** the AppImage if appimagetool can't be downloaded, so a machine with
  no network still gets a `.deb`. A release must not be that forgiving, so the workflow has
  a separate step asserting every expected artifact exists and is non-empty.

Notes on the update check:
- **The plan said "wire Wails' updater to the release feed". There is no such thing.**
  Wails v2 ships no updater package — the only matches for "updater" under `pkg/` are React
  and Preact boilerplate in the project templates. So this is an update *check*, not
  auto-update: `internal/update` asks the GitHub releases API for the latest tag, compares
  it as semver, and offers a button that opens the download page. Nothing replaces a binary.
- **It only runs when you press the button.** No check on startup, no background poll. The
  README calls this app local-first with no server, and a phone-home on launch would quietly
  make that untrue for the sake of a version string.
- Real self-update would mean a third-party library replacing the running executable, plus
  code signing to make that safe to do. Both are real work and neither is free; the button
  is the honest version until they exist.

Notes on data portability:
- **Backup is `VACUUM INTO ?`, not a file copy.** It takes a consistent snapshot of a live
  database without closing it, which a `cp` of a WAL-mode database does not — you would get
  the main file without the log. Verified that the driver accepts it as a *bound parameter*
  (including paths with spaces and quotes) rather than needing the path spliced into SQL.
- Restore validates before it destroys: the candidate is opened read-only and must carry
  both `settings` and `goose_db_version` or it is refused as "not made by Hilt". The current
  database is then snapshotted to `data-replaced-<stamp>.db` *before* the swap, the WAL and
  SHM sidecars are cleared, and the app reopens, re-migrates and reindexes in place.
- The library export is deliberately **not** the database — it is readable JSON of the things
  you authored, and it **merges**. Cards whose name is taken are renamed (`Gutter Wraith (2)`),
  and because slugs derive from names, imported encounters are remapped through a
  slug-old→slug-new table so their picks still resolve. Parties are matched by name and
  reused rather than duplicated.
- Import reports counts, renames and skips instead of failing whole. One bad card should not
  cost you the other forty.
- `ExportLibrary`/`ImportLibrary` are built **entirely on existing exported `gm.Service`
  methods** — no new SQL, so none of this needed `sqlc generate`. The one exception is
  reading raw encounter rows, which `library.go` can do directly because it lives in package
  `gm` and can reach `s.q` and `decodePicks`.

Notes on share codes:
- Format is `HILT<version>:<base64url(zlib(json))>` — `HILT1:` today. The version is checked
  before anything is decoded, so a code from a future build is refused with a message that
  says so rather than failing as corrupt.
- **Decoding is defensive**, because a code arrives from a stranger: raw input is stripped of
  all whitespace first (codes get wrapped by chat clients), the inflate is capped at 1 MiB
  against a zip bomb, and every failure mode returns the same plain-language error rather
  than a stack of wrapped internals. Verified round-trip fidelity, tolerance of line-wrapped
  paste, and rejection of empty / wrong-prefix / future-version / truncated / garbage input.
- base64**url** with no padding, so a code survives being pasted into a URL or a chat client
  that eats `+`, `/` and `=`.
- A code carries **one card and nothing else** — no encounters, no campaign, nothing personal.
- Sharing is offered on homebrew only. An SRD card is one that everybody already has, and
  re-encoding SRD text into shareable strings is not a thing this app needs to do.
- Import previews before it writes, so you see what a code holds before it lands.
- `internal/share` and `internal/update` are table-tested. **Two of those tests were initially
  worthless and a mutation check caught it** — deleting the zip-bomb cap and deleting the
  future-version gate both left the suite green, because a truncated inflate and a
  non-zlib body fail for their own reasons anyway. Asserting only "an error happened" proved
  nothing; both now assert the *specific* rejection, and a forged code differing from a valid
  one only in its version digit is what exercises the version gate.
- `Check` delegates to an unexported `checkAt(ctx, url, current)` so the HTTP paths —
  status codes, the read cap, header shape, cancellation — are testable against an
  `httptest` server. The exported API is unchanged and the feed URL is still a constant.

Notes on the frontend:
- `Modal.svelte` is the scrim-and-sheet shell, extracted once there were going to be four
  copies of it. It owns focus, the `Escape` binding, and the capture-phase modal gate from
  the keyboard work, so no caller re-implements any of that. `ShortcutHelp` and `Settings`
  were retrofitted onto it rather than left as near-duplicates.
- `.iconbtn` in `style.css` is the shared chrome for the two header openers, for the same
  reason `.empty` and `kbd` live there.
- **Settings is a header sheet, not a nav section.** It is app-wide — backup, library,
  version, data location — and the GM nav is already eight entries deep and entirely about
  running a game.
- Restoring a database or importing a library bumps a `reloadToken` in `App.svelte` that
  keys the shell, so every pane remounts and re-reads. Without it the panes would keep
  rendering whatever they loaded before the data underneath changed.
- The generated Wails bindings under `frontend/wailsjs/` are committed, so the new methods
  were added there by hand to keep the tree building. **`wails build` and `wails dev`
  regenerate that directory** — the hand-written entries match what the generator emits, and
  will simply be rewritten identically.

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
