<script>
  import { untrack } from 'svelte'
  import CardBrowser from './CardBrowser.svelte'
  import FeatureEditor from './FeatureEditor.svelte'
  import AdversaryDetail from './AdversaryDetail.svelte'
  import {
    ADVERSARY_TYPES,
    BrowseAdversaries,
    CreateCustomAdversary,
    TIERS,
    UpdateCustomAdversary,
    costLabel,
    errorMessage
  } from './api.js'

  // card is null when creating. Every stat is a string on the Go side — they
  // hold things like "1d12+2" and "+3", not just numbers.
  let { card = null, onsaved, oncancel } = $props()

  // The parent unmounts this form to leave it, so an instance always edits the
  // card it mounted with — snapshot it once and deep-copy into the form so
  // typing never mutates the row still sitting in the browser's list.
  const original = untrack(() => card)
  const editing = original !== null

  let form = $state(
    original
      ? {
          ...original,
          standardAttack: { ...(original.standardAttack ?? {}) },
          features: (original.features ?? []).map((f) => ({
            ...f,
            questions: [...(f.questions ?? [])]
          }))
        }
      : {
          kind: 'adversary',
          slug: '',
          name: '',
          tier: '1',
          type: 'Standard',
          description: '',
          hordeNumber: '1',
          motives: '',
          experiences: '',
          difficulty: '',
          thresholdMinor: '',
          thresholdMajor: '',
          hp: '',
          stress: '',
          standardAttack: { modifier: '', name: '', range: 'Melee', damage: '', damageType: 'phy' },
          features: []
        }
  )

  let saving = $state(false)
  let error = $state('')
  let pane = $state('preview')

  // The card the preview renders — the same shape the browser would show.
  const preview = $derived({ ...form, source: 'custom' })

  // Name and slug are deliberately left alone: the slug is derived from the name on
  // create and is immutable after, so copying one in would either collide with the
  // source card or silently orphan encounters pointing at this one.
  function useAsTemplate(card) {
    if (!confirm(`Fill this form from “${card.name}”? Everything but the name is replaced.`)) return
    form = {
      ...form,
      tier: card.tier,
      type: card.type,
      description: card.description,
      hordeNumber: card.hordeNumber,
      motives: card.motives,
      experiences: card.experiences,
      difficulty: card.difficulty,
      thresholdMinor: card.thresholdMinor,
      thresholdMajor: card.thresholdMajor,
      hp: card.hp,
      stress: card.stress,
      standardAttack: { ...(card.standardAttack ?? {}) },
      features: (card.features ?? []).map((f) => ({ ...f, questions: [...(f.questions ?? [])] }))
    }
    if (!form.name.trim()) form.name = `${card.name} (copy)`
    pane = 'preview'
  }

  async function submit(event) {
    event.preventDefault()
    saving = true
    try {
      const saved = editing ? await UpdateCustomAdversary(form) : await CreateCustomAdversary(form)
      error = ''
      onsaved?.(saved)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }
</script>

<div class="form-pane">
  <header>
    <h2>{editing ? `Edit ${original.name}` : 'New homebrew adversary'}</h2>
    <button class="btn ghost" type="button" onclick={oncancel}>Cancel</button>
    <button class="btn primary" type="submit" form="adversary-form" disabled={saving || !form.name.trim()}>
      {editing ? 'Save changes' : 'Create'}
    </button>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="split">
    <form id="adversary-form" onsubmit={submit}>
      <section>
        <div class="row">
          <label class="grow">
            <span>Name</span>
            <input bind:value={form.name} placeholder="Acid Burrower" required />
            {#if editing}
              <small>The slug stays <code>{original.slug}</code> — renaming won't break saved encounters.</small>
            {/if}
          </label>
          <label class="narrow">
            <span>Tier</span>
            <select bind:value={form.tier}>
              {#each TIERS as tier (tier)}
                <option value={tier}>Tier {tier}</option>
              {/each}
            </select>
          </label>
          <label class="narrow">
            <span>Type</span>
            <select bind:value={form.type}>
              {#each ADVERSARY_TYPES as type (type)}
                <option value={type}>{type}</option>
              {/each}
            </select>
          </label>
        </div>

        <label>
          <span>Description</span>
          <textarea rows="2" bind:value={form.description} placeholder="A horse-sized insect with digging claws and acidic blood."></textarea>
        </label>

        <label>
          <span>Motives &amp; Tactics</span>
          <input bind:value={form.motives} placeholder="Burrow, drag away, feed, reposition" />
        </label>

        <label>
          <span>Experience</span>
          <input bind:value={form.experiences} placeholder="Tremor Sense +2" />
        </label>
      </section>

      <section>
        <h3>Statistics</h3>
        <div class="row">
          <label><span>Difficulty</span><input bind:value={form.difficulty} placeholder="14" /></label>
          <label><span>Threshold minor</span><input bind:value={form.thresholdMinor} placeholder="8" /></label>
          <label><span>Threshold major</span><input bind:value={form.thresholdMajor} placeholder="15" /></label>
        </div>
        <div class="row">
          <label><span>HP</span><input bind:value={form.hp} placeholder="8" /></label>
          <label><span>Stress</span><input bind:value={form.stress} placeholder="3" /></label>
          <label>
            <span>Horde number</span>
            <input bind:value={form.hordeNumber} placeholder="1" />
          </label>
        </div>
      </section>

      <section>
        <h3>Standard attack</h3>
        <div class="row">
          <label class="grow"><span>Name</span><input bind:value={form.standardAttack.name} placeholder="Claws" /></label>
          <label class="narrow"><span>Modifier</span><input bind:value={form.standardAttack.modifier} placeholder="+3" /></label>
        </div>
        <div class="row">
          <label>
            <span>Range</span>
            <input bind:value={form.standardAttack.range} list="attack-ranges" placeholder="Melee" />
          </label>
          <label><span>Damage</span><input bind:value={form.standardAttack.damage} placeholder="1d12+2" /></label>
          <label>
            <span>Damage type</span>
            <input bind:value={form.standardAttack.damageType} list="damage-types" placeholder="phy" />
          </label>
        </div>
      </section>

      <section>
        <h3>Features</h3>
        <FeatureEditor features={form.features} withCommon />
      </section>
    </form>

    <aside>
      <div class="panetabs">
        <button class="tabbtn" class:on={pane === 'preview'} type="button" onclick={() => (pane = 'preview')}>
          Preview
        </button>
        <button class="tabbtn" class:on={pane === 'reference'} type="button" onclick={() => (pane = 'reference')}>
          Reference
        </button>
      </div>

      {#if pane === 'preview'}
        <div class="preview">
          <AdversaryDetail card={preview} />
        </div>
      {:else}
        <div class="reference">
          <CardBrowser
            compact
            types={ADVERSARY_TYPES}
            load={BrowseAdversaries}
            emptyLabel="No adversaries match these filters."
          >
            {#snippet row(item)}
              <span class="rname">{item.name}</span>
              <span class="rmeta">
                Tier {item.tier} · {item.type} · {costLabel(item.type)}
                {#if item.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
              </span>
            {/snippet}

            {#snippet detail(item)}
              <AdversaryDetail card={item}>
                {#snippet actions()}
                  <button class="btn" type="button" onclick={() => useAsTemplate(item)}>Use as template</button>
                {/snippet}
              </AdversaryDetail>
            {/snippet}
          </CardBrowser>
        </div>
      {/if}
    </aside>
  </div>
</div>

<datalist id="attack-ranges">
  <option value="Melee"></option>
  <option value="Very Close"></option>
  <option value="Close"></option>
  <option value="Far"></option>
  <option value="Very Far"></option>
</datalist>
<datalist id="damage-types">
  <option value="phy"></option>
  <option value="mag"></option>
  <option value="phy/mag"></option>
</datalist>

<style>
  .form-pane {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--line);
  }

  h2 {
    flex: 1;
    margin: 0;
    font-size: 1.05rem;
  }

  .error {
    margin: 0;
    padding: 0.5rem 1rem;
    border-bottom: 1px solid var(--danger);
    font-size: 0.8rem;
    color: var(--danger);
  }

  .split {
    display: flex;
    flex: 1;
    min-height: 0;
  }

  form {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 1.25rem;
    min-width: 0;
    padding: 1rem;
    border-right: 1px solid var(--line);
    overflow-y: auto;
  }

  section {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  h3 {
    margin: 0;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--muted);
  }

  .row {
    display: flex;
    gap: 0.5rem;
  }

  label {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.25rem;
    min-width: 0;
  }

  label.grow { flex: 2; }
  label.narrow { flex: 0 0 7rem; }

  label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  small {
    font-size: 0.7rem;
    color: var(--muted);
    opacity: 0.8;
  }

  code {
    padding: 0 0.2rem;
    border-radius: 3px;
    background: var(--panel-2);
  }

  aside {
    display: flex;
    flex-direction: column;
    width: 23rem;
    flex-shrink: 0;
    min-height: 0;
    padding: 1rem;
    overflow: hidden;
  }

  .panetabs {
    display: flex;
    gap: 0.25rem;
    flex-shrink: 0;
  }

  .tabbtn {
    padding: 0.25rem 0.7rem;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: var(--panel);
    color: var(--muted);
    font: inherit;
    font-size: 0.75rem;
    cursor: pointer;
  }

  .tabbtn:hover { color: var(--text); }

  .tabbtn.on {
    border-color: var(--fear);
    color: var(--fear);
  }

  .reference {
    display: flex;
    flex: 1;
    min-height: 0;
    margin-top: 0.5rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
    overflow: hidden;
  }

  .rname {
    display: block;
    font-size: 0.85rem;
  }

  .rmeta {
    display: flex;
    align-items: center;
    gap: 0.3rem;
    font-size: 0.7rem;
    color: var(--muted);
  }

  .preview {
    flex: 1;
    min-height: 0;
    margin-top: 0.5rem;
    padding: 0.85rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
    overflow-y: auto;
  }
</style>
