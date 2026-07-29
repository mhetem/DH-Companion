<script>
  import EmptyState from '../EmptyState.svelte'
  import FeatureList from '../gm/FeatureList.svelte'
  import DualityDice from './DualityDice.svelte'
  import Experiences from './Experiences.svelte'
  import LevelUp from './LevelUp.svelte'
  import LoadoutCards from './LoadoutCards.svelte'
  import ResourceTrack from './ResourceTrack.svelte'
  import RestSheet from './RestSheet.svelte'
  import { active } from './active.svelte.js'
  import {
    GetSheet,
    SetVitals,
    MASTERY_LABELS,
    TRAITS,
    errorMessage,
    featuresUpTo,
    signed
  } from './api.js'

  const DICE_KEY = 'hilt.player.sheetDice'

  let sheet = $state(null)
  let loading = $state(true)
  let error = $state('')
  let busy = $state(false)
  let levelling = $state(false)
  let restReport = $state(null)
  let resting = $state(false)
  let restLong = $state(false)

  // Rendered whether or not it's open, so clicking a trait always has a panel to
  // roll through — the collapse is a hidden attribute, not a mount.
  let dicePanel = $state(null)
  let diceOpen = $state(localStorage.getItem(DICE_KEY) === 'open')

  function toggleDice() {
    diceOpen = !diceOpen
    localStorage.setItem(DICE_KEY, diceOpen ? 'open' : 'closed')
  }

  const id = $derived(active.id)

  $effect(() => {
    const target = id
    if (!target) {
      sheet = null
      loading = false
      return
    }
    loading = true
    GetSheet(target)
      .then((s) => {
        sheet = s
        error = ''
      })
      .catch((e) => {
        error = errorMessage(e)
        sheet = null
      })
      .finally(() => (loading = false))
  })

  // Every vitals mutator returns the whole character, so the sheet renders what the
  // database actually holds rather than a guess made in the UI.
  async function apply(fn) {
    if (busy || !sheet) return
    busy = true
    try {
      const character = await fn()
      sheet = { ...sheet, ...character }
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  const setHP = (value) => apply(() => SetVitals(sheet.id, value, sheet.stressMarked, sheet.hope, sheet.armorMarked))
  const setStress = (value) => apply(() => SetVitals(sheet.id, sheet.hpMarked, value, sheet.hope, sheet.armorMarked))
  const setHope = (value) => apply(() => SetVitals(sheet.id, sheet.hpMarked, sheet.stressMarked, value, sheet.armorMarked))
  const setArmor = (value) => apply(() => SetVitals(sheet.id, sheet.hpMarked, sheet.stressMarked, sheet.hope, value))

  function openRest(long) {
    restLong = long
    restReport = null
    resting = true
  }

  function rested(result) {
    sheet = { ...sheet, ...result.character }
    restReport = result
  }

  // A swap returns the character and the whole loadout, so the sheet takes both
  // rather than guessing which card moved where. The outcome line reuses the rest
  // report's slot — a swap and a rest are the same kind of "here's what happened".
  function swapped(result) {
    sheet = { ...sheet, ...result.character, loadout: result.loadout.loadout, vault: result.loadout.vault }
    restReport = { swap: true, outcomes: [result.outcome] }
  }

  function levelled(updated) {
    sheet = updated
    levelling = false
  }

  // Clicking a trait rolls it. The panel owns the roll so the modifier breakdown,
  // the flash and the log are the same ones the dice section shows.
  function rollTrait(key) {
    if (!diceOpen) toggleDice()
    dicePanel?.rollTrait(key)
  }

  const subclassFeatures = $derived(featuresUpTo(sheet?.subclass?.features ?? [], sheet?.subclassMastery ?? 1))
  const multiclassFeatures = $derived(featuresUpTo(sheet?.multiclassSubclass?.features ?? [], 1))
</script>

{#if !id}
  <div class="pane">
    <EmptyState title="No character open">
      Pick one in Characters — the sheet follows whatever is selected there.
    </EmptyState>
  </div>
{:else if loading}
  <div class="pane"><p class="empty">Loading sheet…</p></div>
{:else if !sheet}
  <div class="pane"><p class="error">{error}</p></div>
{:else}
  <div class="pane">
    <header>
      <div class="who">
        <h2>{sheet.name}{#if sheet.pronouns}<span class="pronouns">{sheet.pronouns}</span>{/if}</h2>
        <p class="meta">
          Level {sheet.level} · Tier {sheet.tier} · {sheet.className} ({sheet.subclassName})
          {#if sheet.multiclassName}· {sheet.multiclassName} ({sheet.multiclassSubclassName}){/if}
          · {sheet.ancestryName} · {sheet.communityName}
        </p>
        <p class="meta">
          Proficiency {sheet.proficiency} · {MASTERY_LABELS[sheet.subclassMastery - 1]}
          {#if sheet.domains?.length}· {sheet.domains.join(' & ')}{/if}
          {#if sheet.spellcastTrait}· Spellcast: {sheet.spellcastTrait}{/if}
        </p>
      </div>
      <div class="header-actions">
        <button class="btn ghost" disabled={busy} onclick={() => openRest(false)}>Short rest</button>
        <button class="btn ghost" disabled={busy} onclick={() => openRest(true)}>Long rest</button>
        <button class="btn primary" disabled={!sheet.canLevelUp} onclick={() => (levelling = true)}>
          {sheet.canLevelUp ? 'Level up' : 'Max level'}
        </button>
      </div>
    </header>

    {#if error}
      <p class="error">{error}</p>
    {/if}

    {#if restReport}
      <div class="rest">
        <span class="rest-title">
          {restReport.swap ? 'Domain cards' : restReport.long ? 'Long rest' : 'Short rest'}
        </span>
        <ul>
          {#each restReport.outcomes as outcome, i (i)}<li>{outcome}</li>{/each}
        </ul>
        <button class="btn ghost" onclick={() => (restReport = null)}>✕</button>
      </div>
    {/if}

    <section class="dice-block">
      <button class="dice-toggle" onclick={toggleDice} aria-expanded={diceOpen}>
        <span class="caret" class:open={diceOpen}>▸</span>
        <span>Dice</span>
        <span class="dice-hint">Duality and damage rolls — or click a trait below</span>
      </button>
      <div class="dice-body" hidden={!diceOpen}>
        <DualityDice
          bind:this={dicePanel}
          compact
          character={sheet}
          experiences={sheet.experiences}
          onupdate={(character) => (sheet = { ...sheet, ...character })}
        />
      </div>
    </section>

    <div class="columns">
      <section class="card">
        <h3>Traits <span class="aside">click to roll</span></h3>
        <div class="traits">
          {#each TRAITS as trait (trait.key)}
            <button
              class="trait"
              class:marked={sheet.markedTraits.includes(trait.key)}
              onclick={() => rollTrait(trait.key)}
              title="Roll duality with {trait.label} {signed(sheet.traits[trait.key])}"
            >
              <span class="value">{signed(sheet.traits[trait.key])}</span>
              <span class="tname">{trait.label}</span>
              <span class="tblurb">{trait.blurb}</span>
              {#if sheet.markedTraits.includes(trait.key)}<span class="mark">marked</span>{/if}
            </button>
          {/each}
        </div>
      </section>

      <section class="card">
        <h3>Resources</h3>
        <div class="tracks">
          <ResourceTrack label="Hit Points" tone="danger" marked={sheet.hpMarked} max={sheet.hpMax} {busy} onchange={setHP} />
          <ResourceTrack label="Stress" tone="stress" marked={sheet.stressMarked} max={sheet.stressMax} {busy} onchange={setStress} />
          <ResourceTrack label="Hope" tone="hope" marked={sheet.hope} max={sheet.hopeMax} {busy} onchange={setHope} />
          <ResourceTrack label="Armor slots" tone="armor" marked={sheet.armorMarked} max={sheet.armorScore} {busy} onchange={setArmor} />
        </div>
      </section>

      <section class="card">
        <h3>Defenses</h3>
        <div class="stats">
          <div class="stat"><span class="n">{sheet.evasion}</span><span class="l">Evasion</span></div>
          <div class="stat"><span class="n">{sheet.armorScore}</span><span class="l">Armor Score</span></div>
          <div class="stat"><span class="n">{sheet.thresholdMajor}</span><span class="l">Major</span></div>
          <div class="stat"><span class="n">{sheet.thresholdSevere}</span><span class="l">Severe</span></div>
        </div>
        <p class="hint">
          Damage at or above Major marks 2 HP, at or above Severe marks 3. Below Major marks 1.
        </p>
      </section>

      <section class="card wide">
        <h3>
          Loadout
          <span class="aside">{sheet.loadout.length}/{sheet.loadoutMax} active · {sheet.vault.length} vaulted</span>
        </h3>
        <LoadoutCards
          characterId={sheet.id}
          loadout={sheet.loadout}
          vault={sheet.vault}
          loadoutMax={sheet.loadoutMax}
          stressRoom={sheet.stressMax - sheet.stressMarked}
          onswap={swapped}
        />
      </section>

      <section class="card wide">
        <h3>Experiences</h3>
        <Experiences
          characterId={sheet.id}
          experiences={sheet.experiences}
          onchange={(list) => (sheet = { ...sheet, experiences: list })}
        />
      </section>

      <section class="card wide">
        <h3>{sheet.className} features</h3>
        {#if sheet.class}
          <h4>Hope feature</h4>
          <FeatureList features={[sheet.class.hopeFeature]} />
          <h4>Class features</h4>
          <FeatureList features={sheet.class.features} />
        {/if}
        <h4>{sheet.subclassName}</h4>
        <FeatureList features={subclassFeatures} />
        {#if sheet.subclassMastery < 3}
          <p class="hint">
            {MASTERY_LABELS[sheet.subclassMastery]} unlocks with the "upgraded subclass card" advancement.
          </p>
        {/if}
        {#if sheet.multiclassSubclass}
          <h4>{sheet.multiclassName} — {sheet.multiclassSubclassName}</h4>
          <FeatureList features={multiclassFeatures} />
        {/if}
      </section>

      <section class="card wide">
        <h3>Ancestry &amp; community</h3>
        <h4>{sheet.ancestryName}</h4>
        <FeatureList features={sheet.ancestry?.features ?? []} />
        <h4>{sheet.communityName}</h4>
        <FeatureList features={sheet.community?.features ?? []} />
      </section>

      {#if sheet.background || sheet.connections}
        <section class="card wide">
          <h3>Background</h3>
          {#if sheet.background}<p class="prose">{sheet.background}</p>{/if}
          {#if sheet.connections}
            <h4>Connections</h4>
            <p class="prose">{sheet.connections}</p>
          {/if}
        </section>
      {/if}

      {#if sheet.levelUps?.length}
        <section class="card wide">
          <h3>Advancement log</h3>
          <ul class="log">
            {#each sheet.levelUps as record (record.level)}
              <li>
                <span class="lvl">Level {record.level}</span>
                <span class="summary">{record.summary || record.choices.join(', ')}</span>
              </li>
            {/each}
          </ul>
        </section>
      {/if}
    </div>
  </div>

  <LevelUp
    bind:open={levelling}
    characterId={sheet.id}
    onapplied={levelled}
    oncancel={() => (levelling = false)}
  />

  <RestSheet
    bind:open={resting}
    long={restLong}
    characterId={sheet.id}
    onrested={rested}
    onswap={(result) => (sheet = { ...sheet, ...result.character, loadout: result.loadout.loadout, vault: result.loadout.vault })}
  />
{/if}

<style>
  .pane {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  header {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .who { flex: 1; min-width: 0; }

  h2 {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    margin: 0;
    font-size: 1.35rem;
  }

  .pronouns {
    font-size: 0.75rem;
    font-weight: 400;
    color: var(--muted);
  }

  h3 {
    margin: 0 0 0.6rem;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  h4 {
    margin: 0.9rem 0 0.4rem;
    font-size: 0.8rem;
    color: var(--gold);
  }

  .meta {
    margin: 0.25rem 0 0;
    font-size: 0.8rem;
    color: var(--muted);
  }

  .header-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
  }

  .error {
    margin: 0 0 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }

  .rest {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    max-width: 40rem;
    margin-bottom: 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--gold-deep);
    border-radius: 8px;
    background: var(--panel);
  }

  .rest-title {
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--gold);
  }

  .rest ul {
    flex: 1;
    margin: 0;
    padding-left: 1rem;
    font-size: 0.78rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .dice-block {
    max-width: 52rem;
    margin-bottom: 0.75rem;
    border: 1px solid var(--line);
    border-radius: 10px;
    background: var(--panel);
  }

  .dice-toggle {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    width: 100%;
    padding: 0.55rem 0.75rem;
    border: none;
    border-radius: 10px;
    background: none;
    color: var(--text);
    font: inherit;
    font-size: 0.85rem;
    text-align: left;
    cursor: pointer;
  }

  .caret {
    display: inline-block;
    color: var(--gold);
    transition: transform 120ms ease-out;
  }

  .caret.open { transform: rotate(90deg); }

  .dice-hint {
    flex: 1;
    font-size: 0.72rem;
    color: var(--muted);
  }

  .dice-body {
    padding: 0 0.75rem 0.75rem;
  }

  .dice-body[hidden] { display: none; }

  @media (prefers-reduced-motion: reduce) {
    .caret { transition: none; }
  }

  .columns {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr));
    gap: 0.75rem;
    align-items: start;
  }

  .card {
    padding: 0.85rem 0.95rem;
    border: 1px solid var(--line);
    border-radius: 10px;
    background: var(--panel);
  }

  .card.wide {
    grid-column: 1 / -1;
    max-width: 52rem;
  }

  .traits {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(7rem, 1fr));
    gap: 0.5rem;
  }

  .trait {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.1rem;
    padding: 0.5rem 0.3rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel-2);
    color: var(--text);
    font: inherit;
    text-align: center;
    cursor: pointer;
  }

  .trait:hover {
    border-color: var(--hope);
    background: var(--panel);
  }

  .trait.marked { border-color: var(--gold-deep); }

  .trait.marked:hover { border-color: var(--hope); }

  .aside {
    font-size: 0.65rem;
    text-transform: none;
    letter-spacing: 0;
    opacity: 0.75;
  }

  .value {
    font-size: 1.3rem;
    color: var(--hope);
  }

  .tname { font-size: 0.8rem; }

  .tblurb {
    font-size: 0.65rem;
    color: var(--muted);
  }

  .mark {
    font-size: 0.6rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--gold);
  }

  .tracks {
    display: flex;
    flex-direction: column;
    gap: 0.7rem;
  }

  .stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(5rem, 1fr));
    gap: 0.5rem;
  }

  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 0.5rem 0.3rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel-2);
  }

  .stat .n {
    font-size: 1.3rem;
    color: var(--gold);
  }

  .stat .l {
    font-size: 0.68rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--muted);
  }


  .hint {
    margin: 0.6rem 0 0;
    font-size: 0.72rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .prose {
    margin: 0;
    font-size: 0.85rem;
    line-height: 1.55;
    color: var(--muted);
    white-space: pre-wrap;
  }

  .log {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .log li {
    display: flex;
    gap: 0.6rem;
    font-size: 0.8rem;
  }

  .lvl {
    flex: 0 0 4.5rem;
    color: var(--gold);
  }

  .summary { color: var(--muted); }
</style>
