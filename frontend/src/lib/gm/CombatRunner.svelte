<script>
  import {
    AdjustFear,
    EndCombat,
    GetActiveCombat,
    LinkCombat,
    ListCampaigns,
    ListCombats,
    ListEncounters,
    MarkHP,
    MarkStress,
    RemoveCombatant,
    ResumeCombat,
    SetFear,
    SetSpotlight,
    ClearSpotlights,
    StartCombat,
    errorMessage
  } from './api.js'
  import CombatLinks from './CombatLinks.svelte'
  import CombatantRow from './CombatantRow.svelte'
  import Countdowns from './Countdowns.svelte'
  import Dice from './Dice.svelte'
  import FearTracker from './FearTracker.svelte'
  import FeatureList from './FeatureList.svelte'
  import Notes from './Notes.svelte'

  // The rollers are opt-in — the Dice section still stands on its own, this just
  // saves a trip out of the runner mid-fight. Remembered so it stays as you left it.
  const DICE_KEY = 'gm.runner.dice'
  let showDice = $state(localStorage.getItem(DICE_KEY) === '1')

  function toggleDice() {
    showDice = !showDice
    localStorage.setItem(DICE_KEY, showDice ? '1' : '0')
  }

  // A fight started against a campaign spends that campaign's Fear pool instead of
  // its own, so the pick has to happen before the fight starts. Remembered, since a
  // GM runs the same campaign week after week.
  const CAMPAIGN_KEY = 'gm.runner.campaign'

  let combat = $state(null)
  let encounters = $state([])
  let campaigns = $state([])
  let campaignPick = $state(localStorage.getItem(CAMPAIGN_KEY) ?? '')
  let sessionPick = $state('')
  let past = $state([])
  let loading = $state(true)
  let busy = $state(false)
  let error = $state('')

  const spotlit = $derived(combat?.combatants.filter((c) => c.spotlight).length ?? 0)

  async function load() {
    try {
      combat = await GetActiveCombat()
      if (!combat) await loadPickers()
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      loading = false
    }
  }

  async function loadPickers() {
    const [enc, combats, camps] = await Promise.all([ListEncounters(), ListCombats(), ListCampaigns()])
    encounters = enc ?? []
    past = (combats ?? []).filter((c) => !c.active)
    campaigns = camps ?? []
    if (campaignPick && !campaigns.some((c) => String(c.id) === campaignPick)) campaignPick = ''
  }

  load()

  // Every mutator returns the updated row, so the runner renders what the database
  // actually holds — there's no separate save step to get out of sync with.
  async function run(fn) {
    busy = true
    try {
      await fn()
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  function replace(updated) {
    combat.combatants = combat.combatants.map((c) => (c.id === updated.id ? updated : c))
  }

  const start = (id) =>
    run(async () => {
      localStorage.setItem(CAMPAIGN_KEY, campaignPick)
      combat = await StartCombat(
        id,
        campaignPick ? Number(campaignPick) : null,
        sessionPick ? Number(sessionPick) : null
      )
    })

  function pickLinks(campaignId, sessionId) {
    campaignPick = campaignId === null ? '' : String(campaignId)
    sessionPick = sessionId === null ? '' : String(sessionId)
  }

  const relink = (campaignId, sessionId) =>
    run(async () => {
      combat = await LinkCombat(combat.id, campaignId, sessionId)
      localStorage.setItem(CAMPAIGN_KEY, combat.campaignId === null ? '' : String(combat.campaignId))
    })
  const resume = (id) => run(async () => { combat = await ResumeCombat(id) })

  const end = () =>
    run(async () => {
      await EndCombat(combat.id)
      combat = null
      loading = true
      await loadPickers()
      loading = false
    })

  const adjustFear = (delta) => run(async () => { combat.fear = await AdjustFear(combat.id, delta) })
  const setFear = (value) => run(async () => { combat.fear = await SetFear(combat.id, value) })

  const markHp = (c, delta) => run(async () => replace(await MarkHP(c.id, delta)))
  const markStress = (c, delta) => run(async () => replace(await MarkStress(c.id, delta)))
  const spotlight = (c, on) => run(async () => replace(await SetSpotlight(c.id, on)))

  const clearSpotlights = () =>
    run(async () => {
      await ClearSpotlights(combat.id)
      combat.combatants = combat.combatants.map((c) => ({ ...c, spotlight: false }))
    })

  const remove = (c) =>
    run(async () => {
      if (!confirm(`Remove ${c.displayName} from the fight?`)) return
      await RemoveCombatant(c.id)
      combat.combatants = combat.combatants.filter((x) => x.id !== c.id)
    })
</script>

<div class="runner">
  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if loading}
    <p class="empty">Loading…</p>
  {:else if !combat}
    <header>
      <h2>Combat Runner</h2>
      <p class="blurb">Start a fight from a saved encounter. Its adversaries spawn as combatants with their own HP and Stress.</p>
    </header>

    {#if campaigns.length}
      <section class="pick">
        <h3>Run this fight for</h3>
        <CombatLinks
          {campaigns}
          campaignId={campaignPick ? Number(campaignPick) : null}
          sessionId={sessionPick ? Number(sessionPick) : null}
          onchange={pickLinks}
        />
        <p class="note">
          {#if campaignPick}
            Fear comes out of the campaign's pool and carries on after the fight.
          {:else}
            Fear stays with this fight only.
          {/if}
        </p>
      </section>
    {/if}

    <section class="pick">
      <h3>Start from an encounter</h3>
      {#if !encounters.length}
        <p class="empty">No saved encounters yet. Build one first.</p>
      {:else}
        <ul class="list">
          {#each encounters as enc (enc.id)}
            <li>
              <div class="who">
                <span class="name">{enc.name || 'Untitled encounter'}</span>
                <span class="meta">{enc.totalCount} {enc.totalCount === 1 ? 'adversary' : 'adversaries'}</span>
              </div>
              <button class="btn primary" onclick={() => start(enc.id)} disabled={busy}>Start combat</button>
            </li>
          {/each}
        </ul>
      {/if}
    </section>

    {#if past.length}
      <section class="pick">
        <h3>Reopen a past fight</h3>
        <ul class="list">
          {#each past as c (c.id)}
            <li>
              <div class="who">
                <span class="name">{c.encounterName || 'Untitled encounter'}</span>
                <span class="meta">
                  {c.createdAt.slice(0, 10)} · Fear {c.fear}
                  {#if c.campaignName} · {c.campaignName}{/if}
                  {#if c.sessionLabel} · {c.sessionLabel}{/if}
                </span>
              </div>
              <button class="btn ghost" onclick={() => resume(c.id)} disabled={busy}>Resume</button>
            </li>
          {/each}
        </ul>
      </section>
    {/if}
  {:else}
    <header class="live">
      <div>
        <h2>{combat.encounterName || 'Untitled encounter'}</h2>
        <p class="blurb">
          {combat.combatants.length} in the fight · autosaved as you go
          {#if combat.campaignName} · {combat.campaignName}{/if}
          {#if combat.sessionLabel} · {combat.sessionLabel}{/if}
        </p>
      </div>
      <button class="btn ghost" onclick={end} disabled={busy}>End combat</button>
    </header>

    {#if campaigns.length}
      <details class="linkbox" open={!combat.campaignId}>
        <summary>
          {#if combat.sessionLabel}
            Logged to {combat.campaignName} · {combat.sessionLabel}
          {:else if combat.campaignName}
            Part of {combat.campaignName} — not logged to a session
          {:else}
            Not linked to a campaign
          {/if}
        </summary>
        <CombatLinks
          {campaigns}
          campaignId={combat.campaignId}
          sessionId={combat.sessionId}
          {busy}
          onchange={relink}
        />
      </details>
    {/if}

    <div class="grid">
      <section class="roster">
        <div class="rhead">
          <h3>Roster</h3>
          {#if spotlit > 0}
            <button class="btn ghost" onclick={clearSpotlights} disabled={busy}>Clear spotlight ({spotlit})</button>
          {/if}
        </div>
        {#if !combat.combatants.length}
          <p class="empty">Nothing left standing.</p>
        {:else}
          <ul class="combatants">
            {#each combat.combatants as c (c.id)}
              <CombatantRow
                combatant={c}
                {busy}
                onmarkhp={(d) => markHp(c, d)}
                onmarkstress={(d) => markStress(c, d)}
                onspotlight={(on) => spotlight(c, on)}
                onremove={() => remove(c)}
              />
            {/each}
          </ul>
        {/if}
      </section>

      <aside class="rail">
        <FearTracker fear={combat.fear} max={combat.fearMax} {busy} onadjust={adjustFear} onset={setFear} />

        {#if combat.environment}
          <section class="env">
            <h3>{combat.environment.name}</h3>
            <p class="meta">Tier {combat.environment.tier} · {combat.environment.type} · Difficulty {combat.environment.difficulty}</p>
            {#if combat.environment.impulses}
              <p class="impulses"><span class="tag">Impulses</span> {combat.environment.impulses}</p>
            {/if}
            <FeatureList features={combat.environment.features ?? []} />
          </section>
        {/if}

        <Countdowns campaignId={combat.campaignId} />

        {#if combat.campaignId}
          <Notes campaignId={combat.campaignId} compact />
        {/if}

        <section class="dicepanel">
          <header class="phead">
            <h3>Dice</h3>
            <button class="btn ghost" onclick={toggleDice}>{showDice ? 'Hide' : 'Show'}</button>
          </header>
          {#if showDice}
            <Dice compact />
          {/if}
        </section>
      </aside>
    </div>
  {/if}
</div>

<style>
  .runner {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  header { margin-bottom: 1rem; }

  header.live {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
  }

  h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  h3 {
    margin: 0;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .blurb {
    margin: 0.25rem 0 0;
    font-size: 0.85rem;
    color: var(--muted);
  }

  .error {
    margin: 0 0 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }

  /* The rail is allowed to give ground rather than hold 20rem and squeeze the
     roster — at 200% UI scale that column is twice as wide in real pixels. */
  .grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(14rem, 20rem);
    gap: 1rem;
    align-items: start;
  }

  .rhead {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .combatants {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .rail {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .env {
    padding: 0.7rem 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .env h3 {
    font-size: 0.9rem;
    text-transform: none;
    letter-spacing: normal;
    color: var(--text);
  }

  .meta {
    margin: 0.2rem 0 0.5rem;
    font-size: 0.75rem;
    color: var(--muted);
  }

  .impulses {
    margin: 0 0 0.6rem;
    font-size: 0.8rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .tag {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--fear);
  }

  .dicepanel {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.7rem 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .phead {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    margin: 0;
  }

  .pick { margin-bottom: 1.25rem; }

  .note {
    margin: 0.3rem 0 0;
    font-size: 0.75rem;
    color: var(--muted);
  }

  .linkbox {
    margin-bottom: 1rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .linkbox summary {
    font-size: 0.8rem;
    color: var(--muted);
    cursor: pointer;
  }

  .linkbox[open] summary { margin-bottom: 0.6rem; }

  .list {
    margin: 0.5rem 0 0;
    padding: 0;
    list-style: none;
  }

  .list li {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.6rem 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    margin-bottom: 0.4rem;
    background: var(--panel);
  }

  .who {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }

  .name { font-size: 0.9rem; }

  .who .meta {
    margin: 0;
    font-size: 0.75rem;
  }

  .empty {
    margin: 0;
    padding: 0.75rem 0;
    font-size: 0.85rem;
    color: var(--muted);
  }

  @media (max-width: 900px) {
    .grid { grid-template-columns: minmax(0, 1fr); }
  }
</style>
