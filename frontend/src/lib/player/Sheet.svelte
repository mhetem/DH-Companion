<script>
  import EmptyState from '../EmptyState.svelte'
  import SrdText from '../SrdText.svelte'
  import FeatureList from '../gm/FeatureList.svelte'
  import DualityDice from './DualityDice.svelte'
  import Experiences from './Experiences.svelte'
  import LevelUp from './LevelUp.svelte'
  import LoadoutCards from './LoadoutCards.svelte'
  import ResourceTrack from './ResourceTrack.svelte'
  import RestSheet from './RestSheet.svelte'
  import { active } from './active.svelte.js'
  import {
    DropStance,
    GetSheet,
    LearnStance,
    ListTransformations,
    Refocus,
    SetClassTracks,
    SetEffect,
    SetTransformation,
    SetVitals,
    ShiftStance,
    TakeDamage,
    MASTERY_LABELS,
    TRAITS,
    errorMessage,
    featuresUpTo,
    gearLine,
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

  // Mirrors FavorMax and FocusMax in internal/player.
  const FAVOR_MAX = 6
  const FOCUS_MAX = 6

  // A character carries which timed effects are on; the sheet's toggle list is
  // rebuilt from it, since a hit or a rest can end some without the list asking.
  function merge(character) {
    const on = character.activeEffects ?? []
    sheet = {
      ...sheet,
      ...character,
      effects: (sheet.effects ?? []).map((fx) => ({ ...fx, active: on.includes(fx.id) }))
    }
  }

  // Every vitals mutator returns the whole character, so the sheet renders what the
  // database actually holds rather than a guess made in the UI.
  async function apply(fn) {
    if (busy || !sheet) return
    busy = true
    try {
      merge(await fn())
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
  const setFavor = (value) => apply(() => SetClassTracks(sheet.id, value, sheet.focus))
  const setFocus = (value) => apply(() => SetClassTracks(sheet.id, sheet.favor, value))

  const toggleEffect = (fx) => apply(() => SetEffect(sheet.id, fx.id, !fx.active))
  const learnStance = (stance, known) => apply(() => LearnStance(sheet.id, stance.title, known))
  const shiftStance = (stance, payWith) => apply(() => ShiftStance(sheet.id, stance.title, payWith))
  const dropStance = () => apply(() => DropStance(sheet.id))

  async function refocus() {
    if (busy || !sheet) return
    busy = true
    try {
      const result = await Refocus(sheet.id)
      merge(result.character)
      restReport = {
        title: 'Refocus',
        outcomes: [
          result.rolls.length
            ? `Rolled ${result.rolls.join(', ')} — Focus is now ${result.focus}`
            : 'With no Instinct to roll, Focus clears to 0'
        ]
      }
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  function openRest(long) {
    restLong = long
    restReport = null
    resting = true
  }

  function rested(result) {
    merge(result.character)
    restReport = result
  }

  // A swap returns the character and the whole loadout, so the sheet takes both
  // rather than guessing which card moved where. The outcome line reuses the rest
  // report's slot — a swap and a rest are the same kind of "here's what happened".
  function swapped(result) {
    merge(result.character)
    sheet = { ...sheet, loadout: result.loadout.loadout, vault: result.loadout.vault }
    restReport = { swap: true, outcomes: [result.outcome] }
    refreshBonuses()
  }

  // A card moving in or out of the loadout can change what's counted into the
  // defenses, and only the sheet carries that list.
  function refreshBonuses() {
    if (!sheet) return
    GetSheet(sheet.id)
      .then((s) => (sheet = { ...sheet, bonuses: s.bonuses }))
      .catch(() => {})
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

  // Damage is typed in, not rolled here: the GM rolls it and says the number. The
  // backend picks the severity off the thresholds, so the sheet and the rule
  // can't disagree about what "Major" means.
  let hitAmount = $state(null)
  let hitArmor = $state(false)
  let hit = $state(null)

  const slotsLeft = $derived(sheet ? sheet.armorScore - sheet.armorMarked : 0)
  const worn = $derived(
    (sheet?.inventory ?? []).filter((i) => i.equipped && i.kind !== 'item' && i.kind !== 'consumable')
  )

  async function takeDamage(event) {
    event.preventDefault()
    if (busy || !sheet || hitAmount === null || hitAmount === '') return
    busy = true
    try {
      const result = await TakeDamage({ characterId: sheet.id, amount: Number(hitAmount), useArmor: hitArmor })
      merge(result.character)
      hit = result
      hitAmount = null
      hitArmor = false
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  function describeHit(r) {
    if (r.severity === 'none') return 'No damage, so nothing to mark.'
    const severity = r.severity[0].toUpperCase() + r.severity.slice(1)
    const armor = r.armorUsed ? `, down to ${r.reduced === 'none' ? 'nothing' : r.reduced} with an Armor Slot` : ''
    const marked = `${r.hpMarked} Hit Point${r.hpMarked === 1 ? '' : 's'} marked`
    const last = r.downed ? " That's your last Hit Point — time for a death move." : ''
    const ended = r.ended?.length ? ` Ended: ${r.ended.join(', ')}.` : ''
    return `${severity} damage${armor}: ${marked}.${last}${ended}`
  }

  function feature(card) {
    return card?.feature ? `<strong>${card.feature.title}</strong>: ${card.feature.description}` : ''
  }

  // Transformations are the GM's to hand out, during play or at creation, so the
  // sheet takes one rather than the wizard asking for it. A character has one at
  // most; taking another replaces it.
  let transformations = $state([])
  let transformPick = $state('')

  ListTransformations()
    .then((list) => (transformations = list ?? []))
    .catch(() => {})

  const transformQuestions = $derived(
    sheet?.transformation ? [{ title: 'Questions', description: '', questions: sheet.transformation.questions }] : []
  )

  const transform = (slug) =>
    apply(async () => {
      const character = await SetTransformation(sheet.id, slug)
      const fresh = await GetSheet(sheet.id)
      transformPick = ''
      return { ...character, transformation: fresh.transformation }
    })

  const stanceTiers = [1, 2, 3, 4]

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
          {#if sheet.transformationName}· {sheet.transformationName}{/if}
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
          {restReport.title ?? (restReport.swap ? 'Domain cards' : restReport.long ? 'Long rest' : 'Short rest')}
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
          onupdate={merge}
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
          {#if sheet.usesFavor}
            <ResourceTrack label="Favor" tone="hope" marked={sheet.favor} max={FAVOR_MAX} {busy} onchange={setFavor} />
          {/if}
          {#if sheet.usesFocus}
            <ResourceTrack label="Focus" tone="plain" marked={sheet.focus} max={FOCUS_MAX} {busy} onchange={setFocus} />
            <button class="btn ghost refocus" disabled={busy} onclick={refocus}>
              Refocus — clear and roll {Math.max(sheet.traits.instinct, 0)}d6
            </button>
          {/if}
        </div>
      </section>

      <section class="card">
        <h3>Defenses</h3>
        <div class="stats">
          <div class="stat">
            <span class="n">{sheet.effectiveEvasion}</span><span class="l">Evasion</span>
            {#if sheet.effectiveEvasion !== sheet.evasion}<span class="base">base {sheet.evasion}</span>{/if}
          </div>
          <div class="stat"><span class="n">{sheet.armorScore}</span><span class="l">Armor Score</span></div>
          <div class="stat">
            <span class="n">{sheet.effectiveMajor}</span><span class="l">Major</span>
            {#if sheet.effectiveMajor !== sheet.thresholdMajor}<span class="base">base {sheet.thresholdMajor}</span>{/if}
          </div>
          <div class="stat">
            <span class="n">{sheet.effectiveSevere}</span><span class="l">Severe</span>
            {#if sheet.effectiveSevere !== sheet.thresholdSevere}<span class="base">base {sheet.thresholdSevere}</span>{/if}
          </div>
        </div>
        {#if sheet.live?.length}
          <div class="bonuses">
            <span class="counted">Right now</span>
            {#each sheet.live as bonus, i (i)}
              <span class="bonus"><strong>{bonus.source}</strong> {bonus.effect}</span>
            {/each}
          </div>
        {/if}
        {#if sheet.effects?.length}
          <div class="effects">
            {#each sheet.effects as fx (fx.id)}
              <button
                class="effect"
                class:on={fx.active}
                disabled={busy}
                title="{fx.source}: {fx.effect}. Lasts until {fx.ends}."
                onclick={() => toggleEffect(fx)}
              >
                {fx.label}<span class="fx">{fx.effect}</span>
              </button>
            {/each}
          </div>
        {/if}
        {#if sheet.bonuses?.length}
          <div class="bonuses">
            <span class="counted">Counted in</span>
            {#each sheet.bonuses as bonus, i (i)}
              <span class="bonus"><strong>{bonus.source}</strong> {bonus.effect}</span>
            {/each}
          </div>
        {/if}
        {#if worn.length}
          <ul class="worn">
            {#each worn as item (item.id)}
              <li>
                <span class="wname">{item.name}</span>
                <span class="wline">{item.weapon || item.armor ? gearLine(item.weapon ?? item.armor) : item.detail}</span>
                <SrdText text={feature(item.weapon ?? item.armor)} />
              </li>
            {/each}
          </ul>
        {/if}
        <form class="hit" onsubmit={takeDamage}>
          <label class="took">
            <span>Took</span>
            <input type="number" min="0" max="999" bind:value={hitAmount} placeholder="0" aria-label="Damage taken" />
          </label>
          <label class="check">
            <input type="checkbox" bind:checked={hitArmor} disabled={slotsLeft <= 0} />
            <span>Mark an Armor Slot ({slotsLeft} left)</span>
          </label>
          <button class="btn" type="submit" disabled={busy || hitAmount === null || hitAmount === ''}>Take damage</button>
        </form>
        {#if hit}
          <p class="hit-report" class:downed={hit.downed}>{describeHit(hit)}</p>
        {/if}
        <p class="hint">
          Damage at or above Major marks 2 HP, at or above Severe marks 3. Below Major marks 1.
          Marking an Armor Slot drops it one step. Gear, subclass and ancestry features and loadout
          cards with a fixed bonus are counted in as they change; situational ones are still yours
          to apply.
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
          {#if sheet.classSlug === 'brawler'}
            <p class="hint">Your Combo Die is a <strong>d{sheet.comboDie}</strong>.</p>
          {/if}
          {#if sheet.usesFavor}
            <p class="hint">Your Patron Die is a <strong>d{sheet.level >= 5 ? 8 : 6}</strong>.</p>
          {/if}
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
        {#if sheet.transformation}
          <h4>{sheet.transformationName} <span class="aside">transformation</span></h4>
          <FeatureList features={[...sheet.transformation.features, ...transformQuestions]} />
          <button class="btn ghost" disabled={busy} onclick={() => transform('')}>Remove the transformation</button>
        {:else if transformations.length}
          <h4>Transformation</h4>
          <div class="transform">
            <select bind:value={transformPick} aria-label="Transformation">
              <option value="">None</option>
              {#each transformations as t (t.slug)}
                <option value={t.slug}>{t.name}</option>
              {/each}
            </select>
            <button class="btn" disabled={busy || !transformPick} onclick={() => transform(transformPick)}>Take it</button>
          </div>
          {#if transformPick}
            {@const preview = transformations.find((t) => t.slug === transformPick)}
            <SrdText text={preview?.description} />
            <FeatureList features={preview?.features ?? []} />
          {/if}
          <p class="hint">
            Optional, and your GM's to give — during play or at character creation. A character can
            have one, and it doesn't count toward your loadout.
          </p>
        {/if}
      </section>

      {#if sheet.stances}
        <section class="card wide">
          <h3>
            {sheet.stances.name}
            <span class="aside">
              {sheet.knownStances.length} of {sheet.stanceAllowance} known · Focus {sheet.focus}/{FOCUS_MAX}
            </span>
          </h3>
          {#if sheet.activeStance}
            <p class="in-stance">
              In the <strong>{sheet.activeStance}</strong> stance — it ends at the end of the scene, on
              Severe damage, or when you mark your last Hit Point.
              <button class="btn ghost" disabled={busy} onclick={dropStance}>Drop out</button>
            </p>
          {/if}
          <details class="rules">
            <summary>How stances and Focus work</summary>
            <FeatureList features={sheet.stances.rules} />
          </details>
          {#each stanceTiers as tier (tier)}
            <h4 class:locked={tier > sheet.tier}>Tier {tier}</h4>
            <ul class="stances">
              {#each sheet.stances.stances.filter((st) => st.tier === tier) as st (st.title)}
                {@const known = sheet.knownStances.includes(st.title)}
                {@const full = sheet.knownStances.length >= sheet.stanceAllowance}
                <li class:known class:active={sheet.activeStance === st.title}>
                  <label class="learn">
                    <input
                      type="checkbox"
                      checked={known}
                      disabled={busy || (!known && (tier > sheet.tier || full))}
                      onchange={() => learnStance(st, !known)}
                    />
                    <strong>{st.title}</strong>
                    {#if sheet.activeStance === st.title}<span class="chip">active</span>{/if}
                  </label>
                  <SrdText text={st.description} />
                  {#if known && sheet.activeStance !== st.title}
                    <div class="shift">
                      <button class="btn ghost" disabled={busy || sheet.focus < 1} onclick={() => shiftStance(st, 'focus')}>
                        Shift in — spend a Focus
                      </button>
                      {#if sheet.flowState}
                        <button class="btn ghost" disabled={busy} onclick={() => shiftStance(st, 'stress')}>
                          Shift in — mark a Stress
                        </button>
                      {/if}
                    </div>
                  {/if}
                </li>
              {/each}
            </ul>
          {/each}
        </section>
      {/if}

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
    onswap={(result) => {
      merge(result.character)
      sheet = { ...sheet, loadout: result.loadout.loadout, vault: result.loadout.vault }
      refreshBonuses()
    }}
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

  .transform {
    display: flex;
    gap: 0.4rem;
    margin-bottom: 0.5rem;
  }

  .stat .base {
    font-size: 0.65rem;
    color: var(--muted);
  }

  .refocus {
    align-self: flex-start;
    font-size: 0.75rem;
  }

  .effects {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
    margin-top: 0.6rem;
  }

  .effect {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    padding: 0.25rem 0.55rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 0.75rem;
    cursor: pointer;
  }

  .effect .fx { font-size: 0.65rem; }

  .effect.on {
    border-color: var(--hope);
    color: var(--text);
  }

  .in-stance {
    margin: 0 0 0.5rem;
    font-size: 0.8rem;
  }

  .stances {
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .stances li {
    padding: 0.4rem 0.55rem;
    border: 1px solid var(--line);
    border-radius: 8px;
  }

  .stances li.known { border-color: var(--gold-deep); }

  .stances li.active { border-color: var(--hope); }

  .learn {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.85rem;
  }

  .shift {
    display: flex;
    gap: 0.3rem;
    margin-top: 0.3rem;
  }

  h4.locked {
    color: var(--muted);
    opacity: 0.7;
  }

  .rules {
    margin-bottom: 0.4rem;
    font-size: 0.8rem;
  }

  .rules summary {
    color: var(--muted);
    cursor: pointer;
  }

  .bonuses {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.25rem 0.6rem;
    margin-top: 0.6rem;
    font-size: 0.72rem;
    color: var(--muted);
  }

  .counted {
    text-transform: uppercase;
    letter-spacing: 0.07em;
  }

  .bonus strong {
    font-weight: 600;
    color: var(--text);
  }

  .worn {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    margin: 0.7rem 0 0;
    padding: 0;
    list-style: none;
  }

  .worn li {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
  }

  .wname { font-size: 0.82rem; }

  .wline {
    font-size: 0.72rem;
    color: var(--muted);
  }

  .worn :global(.srd) { font-size: 0.72rem; }

  .hit {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: 0.5rem;
    margin-top: 0.8rem;
  }

  .hit label {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
  }

  .hit label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .hit label.check {
    flex-direction: row;
    align-items: center;
    gap: 0.35rem;
  }

  .hit label.check span {
    font-size: 0.75rem;
    text-transform: none;
    letter-spacing: 0;
  }

  .took input { width: 4.5rem; }

  .hit-report {
    margin: 0.5rem 0 0;
    font-size: 0.8rem;
  }

  .hit-report.downed { color: var(--danger); }

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
