<script>
  import EmptyState from '../EmptyState.svelte'
  import RollResult from '../gm/RollResult.svelte'
  import ResourceTrack from './ResourceTrack.svelte'
  import { active } from './active.svelte.js'
  import {
    GetCharacter,
    ListExperiences,
    RollDamage,
    RollDuality,
    SetVitals,
    Sizes,
    TRAITS,
    errorMessage,
    signed
  } from './api.js'

  // Standalone (the Dice nav section) this loads its own character. Embedded in the
  // sheet the parent already holds one, and passes it down with `onupdate` so both
  // copies can't drift apart when a roll writes Hope or Stress.
  let {
    compact = false,
    character: provided = null,
    experiences: providedExperiences = null,
    onupdate
  } = $props()

  let owned = $state(null)
  let ownedExperiences = $state([])
  let selfLoading = $state(true)
  let error = $state('')
  let rolling = $state(false)

  const embedded = $derived(provided !== null)
  const loading = $derived(embedded ? false : selfLoading)
  const character = $derived(provided ?? owned)
  const experiences = $derived(providedExperiences ?? ownedExperiences)

  let trait = $state('')
  let picked = $state([])
  let bonus = $state(0)
  let advantage = $state(false)
  let disadvantage = $state(false)
  let spendHope = $state(0)

  let sizes = $state([])
  let damageSides = $state(8)
  let damageBonus = $state(0)
  let damageCrit = $state(false)

  let log = $state([])
  let flash = $state(null)
  let flashTimer = null

  const id = $derived(active.id)

  const preview = $derived(
    (trait ? (character?.traits?.[trait] ?? 0) : 0) +
      picked.reduce((total, e) => total + (experiences.find((x) => x.id === e)?.modifier ?? 0), 0) +
      Number(bonus || 0)
  )

  Sizes()
    .then((s) => (sizes = s ?? []))
    .catch((e) => (error = errorMessage(e)))

  $effect(() => {
    if (provided) return
    const target = id
    if (!target) {
      owned = null
      selfLoading = false
      return
    }
    selfLoading = true
    Promise.all([GetCharacter(target), ListExperiences(target)])
      .then(([c, x]) => {
        owned = c
        ownedExperiences = x ?? []
        error = ''
      })
      .catch((e) => (error = errorMessage(e)))
      .finally(() => (selfLoading = false))
  })

  $effect(() => () => clearTimeout(flashTimer))

  // Both copies are written, and `character` prefers the prop — so an embedded panel
  // stays correct even if the parent chooses not to handle onupdate.
  function applyCharacter(next) {
    owned = next
    onupdate?.(next)
  }

  // Called by the sheet when a trait on the character sheet is clicked.
  export function rollTrait(key) {
    trait = key
    return roll()
  }

  function toggleExperience(experienceId) {
    picked = picked.includes(experienceId)
      ? picked.filter((e) => e !== experienceId)
      : [...picked, experienceId]
  }

  function show(entry) {
    log = [entry, ...log].slice(0, 12)
    flash = entry
    clearTimeout(flashTimer)
    flashTimer = setTimeout(() => (flash = null), 2200)
  }

  async function roll() {
    if (rolling || !character) return
    rolling = true
    try {
      const result = await RollDuality({
        characterId: character.id,
        trait,
        experienceIds: picked,
        bonus: Number(bonus || 0),
        advantage,
        disadvantage,
        spendHope: Number(spendHope || 0),
        label: trait ? TRAITS.find((t) => t.key === trait)?.label : ''
      })
      // The roll returns the character it just wrote, so Hope and Stress on screen
      // are the row's values rather than a local guess.
      applyCharacter(result.character)
      spendHope = 0
      show({ kind: 'duality', ...result })
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      rolling = false
    }
  }

  async function damage() {
    if (rolling || !character) return
    rolling = true
    try {
      const result = await RollDamage(character.id, Number(damageSides), Number(damageBonus || 0), damageCrit)
      show({ kind: 'damage', ...result })
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      rolling = false
    }
  }

  async function setHope(value) {
    try {
      applyCharacter(
        await SetVitals(character.id, character.hpMarked, character.stressMarked, value, character.armorMarked)
      )
    } catch (e) {
      error = errorMessage(e)
    }
  }

  async function setStress(value) {
    try {
      applyCharacter(
        await SetVitals(character.id, character.hpMarked, value, character.hope, character.armorMarked)
      )
    } catch (e) {
      error = errorMessage(e)
    }
  }
</script>

{#if loading}
  <div class="pane" class:compact><p class="empty">Loading…</p></div>
{:else if !character}
  <div class="pane" class:compact>
    {#if error}
      <p class="error">{error}</p>
    {:else}
      <EmptyState title="No character open">Pick one in Characters first.</EmptyState>
    {/if}
  </div>
{:else}
  <div class="pane" class:compact>
    {#if !compact}
      <header>
        <h2>Duality dice</h2>
        <p class="blurb">
          Two d12 — Hope and Fear. Matching dice are a critical success. A roll with Hope
          gains you a Hope; a critical gains a Hope and clears a Stress.
        </p>
      </header>
    {/if}

    {#if error}<p class="error">{error}</p>{/if}

    {#if !embedded}
      <section class="vitals">
        <ResourceTrack label="Hope" tone="hope" marked={character.hope} max={character.hopeMax} onchange={setHope} />
        <ResourceTrack
          label="Stress"
          tone="stress"
          marked={character.stressMarked}
          max={character.stressMax}
          onchange={setStress}
        />
      </section>
    {/if}

    <section class="build">
      <h3>Trait</h3>
      <div class="chips">
        <button class="chip" class:on={trait === ''} onclick={() => (trait = '')}>None</button>
        {#each TRAITS as t (t.key)}
          <button class="chip" class:on={trait === t.key} onclick={() => (trait = t.key)}>
            {t.label} {signed(character.traits[t.key])}
          </button>
        {/each}
      </div>

      {#if experiences.length}
        <h3>Experiences</h3>
        <div class="chips">
          {#each experiences as experience (experience.id)}
            <button
              class="chip"
              class:on={picked.includes(experience.id)}
              onclick={() => toggleExperience(experience.id)}
            >
              {experience.name} {signed(experience.modifier)}
            </button>
          {/each}
        </div>
        <p class="hint">Spend a Hope to bring an Experience to bear — tick it here and spend below.</p>
      {/if}

      <div class="row">
        <label class="narrow">
          <span>Bonus</span>
          <input type="number" min="-10" max="10" bind:value={bonus} />
        </label>
        <label class="narrow">
          <span>Spend Hope</span>
          <input type="number" min="0" max={character.hope} bind:value={spendHope} />
        </label>
        <label class="check">
          <input type="checkbox" bind:checked={advantage} onchange={() => advantage && (disadvantage = false)} />
          <span>Advantage</span>
        </label>
        <label class="check">
          <input type="checkbox" bind:checked={disadvantage} onchange={() => disadvantage && (advantage = false)} />
          <span>Disadvantage</span>
        </label>
        <span class="total">{signed(preview)}</span>
        <button class="btn primary" disabled={rolling} onclick={roll}>Roll duality</button>
      </div>
    </section>

    <section class="build">
      <h3>Damage</h3>
      <div class="row">
        <div class="chips">
          {#each sizes as size (size)}
            <button class="chip" class:on={damageSides === size} onclick={() => (damageSides = size)}>d{size}</button>
          {/each}
        </div>
        <label class="narrow">
          <span>Bonus</span>
          <input type="number" min="-20" max="40" bind:value={damageBonus} />
        </label>
        <label class="check">
          <input type="checkbox" bind:checked={damageCrit} />
          <span>Critical</span>
        </label>
        <button class="btn ghost" disabled={rolling} onclick={damage}>
          Roll {character.proficiency}d{damageSides}
        </button>
      </div>
      <p class="hint">Damage rolls use your Proficiency ({character.proficiency}) as the dice count.</p>
    </section>

    {#if log.length}
      <section class="log">
        <h3>Recent rolls</h3>
        <ul>
          {#each log as entry, i (i)}
            <li>
              {#if entry.kind === 'duality'}
                <span class="pair">
                  <span class="hope">{entry.hope}</span>
                  <span class="sep">/</span>
                  <span class="fear">{entry.fear}</span>
                </span>
                <span class="tag" class:crit={entry.critical} class:withhope={entry.withHope}>
                  {entry.critical ? 'Critical!' : entry.withHope ? 'with Hope' : 'with Fear'}
                </span>
                <span class="what">
                  {entry.label || 'Duality'}
                  {#if entry.modifier}{signed(entry.modifier)}{/if}
                  {#if entry.advantage}· adv{/if}{#if entry.disadvantage}· dis{/if}
                </span>
                <span class="gains">
                  {#if entry.hopeSpent}−{entry.hopeSpent} Hope{/if}
                  {#if entry.hopeGained}+{entry.hopeGained} Hope{/if}
                  {#if entry.stressCleared}· −{entry.stressCleared} Stress{/if}
                </span>
                <RollResult value={entry.result} max={24} crit={entry.critical} compact />
              {:else}
                <span class="what">{entry.label}{#if entry.modifier}{signed(entry.modifier)}{/if} damage</span>
                <span class="gains">{entry.critical ? 'critical' : ''}</span>
                <RollResult value={entry.total} max={entry.sides * entry.count} compact />
              {/if}
            </li>
          {/each}
        </ul>
      </section>
    {/if}
  </div>

  {#if flash}
    <div class="flash" class:crit={flash.critical}>
      {#if flash.kind === 'duality'}
        <RollResult value={flash.result} max={24} crit={flash.critical} big />
        <p class="flash-tag">
          {flash.critical ? 'Critical success' : flash.withHope ? 'with Hope' : 'with Fear'}
        </p>
        <p class="flash-sub">Hope {flash.hope} · Fear {flash.fear}</p>
      {:else}
        <RollResult value={flash.total} max={flash.sides * flash.count} big />
        <p class="flash-tag">{flash.label} damage</p>
      {/if}
    </div>
  {/if}
{/if}

<style>
  .pane {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.85rem;
    min-height: 0;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  .pane.compact {
    padding: 0;
    gap: 0.6rem;
  }

  header { margin-bottom: -0.2rem; }

  h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  h3 {
    margin: 0 0 0.4rem;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .blurb {
    margin: 0.25rem 0 0;
    max-width: 40rem;
    font-size: 0.85rem;
    color: var(--muted);
  }

  .error {
    margin: 0;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }

  .vitals {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
    gap: 0.9rem;
    max-width: 40rem;
  }

  .build {
    max-width: 48rem;
    padding: 0.85rem 0.95rem;
    border: 1px solid var(--line);
    border-radius: 10px;
    background: var(--panel);
  }

  .build h3:not(:first-child) { margin-top: 0.8rem; }

  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
  }

  .chip {
    padding: 0.25rem 0.55rem;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: var(--panel-2);
    color: var(--muted);
    font: inherit;
    font-size: 0.75rem;
    cursor: pointer;
  }

  .chip:hover { border-color: var(--gold-deep); }

  .chip.on {
    border-color: var(--hope);
    color: var(--text);
  }

  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: 0.6rem;
    margin-top: 0.7rem;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  label.narrow { flex: 0 0 6rem; }

  label.check {
    flex-direction: row;
    align-items: center;
    gap: 0.35rem;
  }

  label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .total {
    padding: 0.3rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    font-size: 0.85rem;
    color: var(--hope);
  }

  .hint {
    margin: 0.5rem 0 0;
    font-size: 0.72rem;
    color: var(--muted);
  }

  .log ul {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .log li {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.35rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--panel);
    font-size: 0.78rem;
  }

  .pair {
    display: flex;
    gap: 0.2rem;
    font-variant-numeric: tabular-nums;
  }

  .hope { color: var(--hope); }
  .fear { color: var(--fear); }
  .sep { color: var(--muted); }

  .tag {
    padding: 0.1rem 0.4rem;
    border-radius: 999px;
    border: 1px solid var(--fear);
    font-size: 0.68rem;
    color: var(--fear);
  }

  .tag.withhope {
    border-color: var(--hope);
    color: var(--hope);
  }

  .tag.crit {
    border-color: var(--gold);
    color: var(--gold);
  }

  .what {
    flex: 1;
    min-width: 0;
    color: var(--muted);
  }

  .gains {
    font-size: 0.72rem;
    color: var(--gold);
  }

  .flash {
    position: fixed;
    inset: 0;
    z-index: 30;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.2rem;
    pointer-events: none;
    background: rgb(0 0 0 / 35%);
    animation: fade 2200ms ease-out forwards;
  }

  .flash-tag {
    margin: 0;
    font-size: 1rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--hope);
  }

  .flash.crit .flash-tag { color: var(--gold); }

  .flash-sub {
    margin: 0;
    font-size: 0.85rem;
    color: var(--muted);
  }

  @keyframes fade {
    0% { opacity: 0; }
    10% { opacity: 1; }
    75% { opacity: 1; }
    100% { opacity: 0; }
  }

  @media (prefers-reduced-motion: reduce) {
    .flash { animation: none; }
  }
</style>
