<script>
  import EmptyState from '../EmptyState.svelte'
  import FeatureList from '../gm/FeatureList.svelte'
  import RollResult from '../gm/RollResult.svelte'
  import ResourceTrack from './ResourceTrack.svelte'
  import { active } from './active.svelte.js'
  import {
    DeleteCompanion,
    GetCompanion,
    MarkCompanionStress,
    RollCompanionDamage,
    SaveCompanion,
    SetCompanionStress,
    ToggleCompanionUpgrade,
    errorMessage,
    signed
  } from './api.js'

  let view = $state(null)
  let loading = $state(true)
  let error = $state('')
  let busy = $state(false)
  let editing = $state(false)
  let form = $state(blank())
  let lastRoll = $state(null)
  let showSetup = $state(false)

  const id = $derived(active.id)
  const companion = $derived(view?.companion ?? null)

  function blank() {
    return {
      name: '',
      evasion: 10,
      damageDie: 'd6',
      attackRange: 'Melee',
      attack: '',
      stressMax: 5,
      // The sheet starts with two Experiences at +2; the rules say a new one for you
      // is a new one for them, so this list grows by hand alongside yours.
      experiences: [
        { name: '', modifier: 2 },
        { name: '', modifier: 2 }
      ],
      upgrades: [],
      notes: ''
    }
  }

  $effect(() => {
    const target = id
    if (!target) {
      view = null
      loading = false
      return
    }
    loading = true
    GetCompanion(target)
      .then((v) => {
        view = v
        error = ''
      })
      .catch((e) => {
        error = errorMessage(e)
        view = null
      })
      .finally(() => (loading = false))
  })

  async function run(fn) {
    if (busy) return
    busy = true
    try {
      view = await fn()
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  function startEdit() {
    form = companion
      ? {
          ...companion,
          experiences: companion.experiences.map((e) => ({ ...e })),
          upgrades: [...companion.upgrades]
        }
      : blank()
    editing = true
  }

  async function save() {
    await run(() =>
      SaveCompanion({
        characterId: id,
        name: form.name,
        evasion: Number(form.evasion),
        damageDie: form.damageDie,
        attackRange: form.attackRange,
        attack: form.attack,
        stressMax: Number(form.stressMax),
        experiences: form.experiences
          .filter((e) => e.name.trim())
          .map((e) => ({ name: e.name, modifier: Number(e.modifier) })),
        upgrades: form.upgrades,
        notes: form.notes
      })
    )
    if (!error) editing = false
  }

  function addExperience() {
    form.experiences = [...form.experiences, { name: '', modifier: 2 }]
  }

  function dropExperience(index) {
    form.experiences = form.experiences.filter((_, i) => i !== index)
  }

  async function rollDamage(critical) {
    if (busy) return
    busy = true
    try {
      lastRoll = await RollCompanionDamage(id, 0, critical)
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  function remove() {
    if (!confirm(`Delete ${companion.name || 'this companion'}? Their Experiences go with them.`)) return
    return run(() => DeleteCompanion(id))
  }
</script>

{#if !id}
  <div class="pane">
    <EmptyState title="No character open">Pick one in Characters first.</EmptyState>
  </div>
{:else if loading}
  <div class="pane"><p class="empty">Loading companion sheet…</p></div>
{:else if !view}
  <div class="pane"><p class="error">{error}</p></div>
{:else if !view.eligible}
  <div class="pane">
    <EmptyState title="No companion">
      The companion sheet belongs to the Beastbound ranger subclass.
    </EmptyState>
  </div>
{:else}
  <div class="pane">
    <header>
      <div class="intro">
        <h2>Companion</h2>
        <p class="blurb">
          Command them with a Spellcast Roll. Their damage roll uses <strong>your</strong>
          Proficiency ({view.proficiency}) and their damage die.
        </p>
      </div>
      {#if companion && !editing}
        <div class="header-actions">
          <button class="btn ghost" onclick={startEdit}>Edit</button>
          <button class="btn danger" disabled={busy} onclick={remove}>Delete</button>
        </div>
      {/if}
    </header>

    {#if error}<p class="error">{error}</p>{/if}

    {#if editing || !companion}
      {#if !companion && !editing}
        <EmptyState title="No companion sheet yet">
          Work with your GM to decide what kind of animal they are, then build their sheet.
        </EmptyState>
        <div class="centre">
          <button class="btn primary" onclick={startEdit}>Build the companion sheet</button>
        </div>
      {:else}
        <section class="card form">
          <h3>{companion ? 'Edit companion' : 'New companion'}</h3>
          <div class="grid">
            <label class="grow">
              <span>Name</span>
              <input bind:value={form.name} placeholder="Bramble" required />
            </label>
            <label class="narrow">
              <span>Evasion</span>
              <input type="number" min="0" max="30" bind:value={form.evasion} />
            </label>
            <label class="narrow">
              <span>Stress slots</span>
              <input type="number" min="1" max="12" bind:value={form.stressMax} />
            </label>
            <label class="narrow">
              <span>Damage die</span>
              <select bind:value={form.damageDie}>
                {#each view.damageDice as die (die)}<option value={die}>{die}</option>{/each}
              </select>
            </label>
            <label class="narrow">
              <span>Range</span>
              <select bind:value={form.attackRange}>
                {#each view.ranges as range (range)}<option value={range}>{range}</option>{/each}
              </select>
            </label>
            <label class="grow">
              <span>Attack</span>
              <input bind:value={form.attack} placeholder="Raking claws" />
            </label>
          </div>

          <h4>Companion Experiences</h4>
          <ul class="exp-edit">
            {#each form.experiences as experience, index (index)}
              <li>
                <input bind:value={experience.name} placeholder="Expert Climber" />
                <input class="narrow" type="number" min="0" max="9" bind:value={experience.modifier} />
                <button class="btn danger" onclick={() => dropExperience(index)}>✕</button>
              </li>
            {/each}
          </ul>
          <button class="btn ghost" onclick={addExperience}>Add an Experience</button>

          <label class="full">
            <span>Notes</span>
            <textarea rows="3" bind:value={form.notes} placeholder="What they look like, how you met."></textarea>
          </label>

          <div class="foot">
            <button class="btn ghost" onclick={() => (editing = false)} disabled={!companion}>Cancel</button>
            <span class="spacer"></span>
            <button class="btn primary" disabled={busy || !form.name.trim()} onclick={save}>
              {companion ? 'Save' : 'Create companion'}
            </button>
          </div>
        </section>
      {/if}
    {:else}
      <div class="columns">
        <section class="card">
          <h3>{companion.name}</h3>
          <div class="stats">
            <div class="stat"><span class="n">{companion.evasion}</span><span class="l">Evasion</span></div>
            <div class="stat"><span class="n">{view.proficiency}{companion.damageDie}</span><span class="l">Damage</span></div>
            <div class="stat"><span class="n">{companion.attackRange}</span><span class="l">Range</span></div>
          </div>
          {#if companion.attack}<p class="attack">{companion.attack}</p>{/if}
          <div class="roll">
            <button class="btn primary" disabled={busy} onclick={() => rollDamage(false)}>Roll damage</button>
            <button class="btn ghost" disabled={busy} onclick={() => rollDamage(true)}>Critical</button>
            {#if lastRoll}
              <span class="result">
                <RollResult value={lastRoll.total} max={lastRoll.sides * lastRoll.count} crit={lastRoll.critical} compact />
                <span class="label">{lastRoll.label}</span>
              </span>
            {/if}
          </div>
        </section>

        <section class="card">
          <h3>Stress</h3>
          <ResourceTrack
            label="Stress"
            tone="stress"
            marked={companion.stressMarked}
            max={companion.stressMax}
            {busy}
            onchange={(value) => run(() => SetCompanionStress(id, value))}
          />
          <p class="hint">
            Any amount of damage marks a Stress. On their last one they drop out of the scene and
            return at your next long rest with one Stress cleared.
          </p>
          <button class="btn ghost" disabled={busy} onclick={() => run(() => MarkCompanionStress(id, 1))}>
            Take damage
          </button>
        </section>

        <section class="card">
          <h3>Experiences</h3>
          <ul class="exp">
            {#each companion.experiences as experience (experience.name)}
              <li>
                <span class="name">{experience.name}</span>
                <span class="mod">{signed(experience.modifier)}</span>
              </li>
            {/each}
            {#if !companion.experiences.length}
              <li class="empty">No Experiences yet — edit the sheet to add them.</li>
            {/if}
          </ul>
          <p class="hint">Spend a Hope to bring one to bear on a Spellcast Roll commanding them.</p>
        </section>

        {#if companion.notes}
          <section class="card">
            <h3>Notes</h3>
            <p class="prose">{companion.notes}</p>
          </section>
        {/if}

        {#if view.reference}
          <section class="card wide">
            <h3>Level-up options</h3>
            <p class="hint">
              A checklist of what you've taken. Options that change a number — Vicious, Resilient,
              Aware — aren't applied automatically; edit the sheet above to match.
            </p>
            <ul class="upgrades">
              {#each view.reference.levelUpOptions as option (option.title)}
                {@const taken = companion.upgrades.includes(option.title)}
                <li class:taken>
                  <button disabled={busy} onclick={() => run(() => ToggleCompanionUpgrade(id, option.title))}>
                    <span class="box" class:on={taken}></span>
                    <span class="what">
                      <span class="title">{option.title}</span>
                      <span class="desc">{@html option.description}</span>
                    </span>
                  </button>
                </li>
              {/each}
            </ul>
          </section>

          <section class="card wide">
            <button class="reveal" onclick={() => (showSetup = !showSetup)} aria-expanded={showSetup}>
              <span class="caret" class:open={showSetup}>▸</span>
              <span>Rules and setup</span>
            </button>
            {#if showSetup}
              <h4>Rules</h4>
              <FeatureList features={view.reference.rules} />
              <h4>Setup</h4>
              <FeatureList features={view.reference.setup} />
              <h4>Example Experiences</h4>
              <p class="prose">{view.reference.exampleExperiences.join(' · ')}</p>
            {/if}
          </section>
        {/if}
      </div>
    {/if}
  </div>
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

  header .intro { flex: 1; }

  .header-actions {
    display: flex;
    gap: 0.4rem;
  }

  h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  h3 {
    margin: 0 0 0.6rem;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  h4 {
    margin: 1rem 0 0.4rem;
    font-size: 0.8rem;
    color: var(--gold);
  }

  .blurb {
    margin: 0.25rem 0 0;
    max-width: 44rem;
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

  .centre {
    display: flex;
    justify-content: center;
  }

  .columns {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(17rem, 1fr));
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

  .form {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    max-width: 44rem;
  }

  .grid {
    display: flex;
    flex-wrap: wrap;
    gap: 0.6rem;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  label.grow { flex: 1 1 11rem; }
  label.narrow { flex: 0 0 7rem; }
  label.full { width: 100%; }

  label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .exp-edit {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .exp-edit li {
    display: flex;
    gap: 0.4rem;
  }

  .exp-edit input { flex: 1; }
  .exp-edit input.narrow { flex: 0 0 4.5rem; }

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
    font-size: 1.15rem;
    color: var(--gold);
  }

  .stat .l {
    font-size: 0.62rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--muted);
  }

  .attack {
    margin: 0.6rem 0 0;
    font-size: 0.85rem;
    color: var(--muted);
  }

  .roll {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem;
    margin-top: 0.75rem;
  }

  .result {
    display: flex;
    align-items: center;
    gap: 0.35rem;
  }

  .result .label {
    font-size: 0.72rem;
    color: var(--muted);
  }

  .exp {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .exp li {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.35rem 0.5rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--panel-2);
  }

  .exp li.empty {
    border: none;
    background: none;
    padding: 0.2rem 0;
  }

  .exp .name {
    flex: 1;
    min-width: 0;
    font-size: 0.85rem;
  }

  .exp .mod {
    font-size: 0.85rem;
    color: var(--hope);
  }

  .upgrades {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .upgrades button {
    display: flex;
    align-items: flex-start;
    gap: 0.6rem;
    width: 100%;
    padding: 0.5rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel-2);
    color: var(--text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .upgrades li.taken button { border-color: var(--hope); }

  .box {
    flex: none;
    width: 0.85rem;
    height: 0.85rem;
    margin-top: 0.2rem;
    border: 1px solid var(--line);
    border-radius: 3px;
  }

  .box.on {
    border-color: var(--hope);
    background: var(--hope);
  }

  .what {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
    min-width: 0;
  }

  .title { font-size: 0.85rem; }

  .desc {
    font-size: 0.75rem;
    line-height: 1.5;
    color: var(--muted);
    white-space: pre-wrap;
  }

  .reveal {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
    padding: 0;
    border: none;
    background: none;
    color: var(--text);
    font: inherit;
    font-size: 0.85rem;
    text-align: left;
    cursor: pointer;
  }

  .caret {
    color: var(--gold);
    transition: transform 120ms ease-out;
  }

  .caret.open { transform: rotate(90deg); }

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

  .foot {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .spacer { flex: 1; }

  @media (prefers-reduced-motion: reduce) {
    .caret { transition: none; }
  }
</style>
