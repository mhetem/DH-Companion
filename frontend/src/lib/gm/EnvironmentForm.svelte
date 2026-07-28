<script>
  import { untrack } from 'svelte'
  import CardBrowser from './CardBrowser.svelte'
  import FeatureEditor from './FeatureEditor.svelte'
  import EnvironmentDetail from './EnvironmentDetail.svelte'
  import {
    BrowseEnvironments,
    CreateCustomEnvironment,
    ENVIRONMENT_TYPES,
    TIERS,
    UpdateCustomEnvironment,
    errorMessage
  } from './api.js'

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
          potentialAdversaries: [...(original.potentialAdversaries ?? [])],
          features: (original.features ?? []).map((f) => ({
            ...f,
            questions: [...(f.questions ?? [])]
          }))
        }
      : {
          kind: 'environment',
          slug: '',
          name: '',
          tier: '1',
          type: 'Exploration',
          description: '',
          difficulty: '',
          impulses: '',
          potentialAdversaries: [],
          features: []
        }
  )

  let saving = $state(false)
  let error = $state('')
  let pane = $state('preview')

  const preview = $derived({ ...form, source: 'custom' })

  // Name and slug are left alone — the slug is derived from the name on create and
  // immutable after, so copying one in would collide with the card it came from.
  function useAsTemplate(card) {
    if (!confirm(`Fill this form from “${card.name}”? Everything but the name is replaced.`)) return
    form = {
      ...form,
      tier: card.tier,
      type: card.type,
      description: card.description,
      difficulty: card.difficulty,
      impulses: card.impulses,
      potentialAdversaries: [...(card.potentialAdversaries ?? [])],
      features: (card.features ?? []).map((f) => ({ ...f, questions: [...(f.questions ?? [])] }))
    }
    if (!form.name.trim()) form.name = `${card.name} (copy)`
    pane = 'preview'
  }

  function addAdversaryGroup() {
    form.potentialAdversaries.push('')
  }

  function removeAdversaryGroup(index) {
    form.potentialAdversaries.splice(index, 1)
  }

  async function submit(event) {
    event.preventDefault()
    saving = true
    try {
      // Blank rows are an artefact of the repeater, not data worth storing.
      const payload = {
        ...form,
        potentialAdversaries: form.potentialAdversaries.map((s) => s.trim()).filter(Boolean)
      }
      const saved = editing
        ? await UpdateCustomEnvironment(payload)
        : await CreateCustomEnvironment(payload)
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
    <h2>{editing ? `Edit ${original.name}` : 'New homebrew environment'}</h2>
    <button class="btn ghost" type="button" onclick={oncancel}>Cancel</button>
    <button class="btn primary" type="submit" form="environment-form" disabled={saving || !form.name.trim()}>
      {editing ? 'Save changes' : 'Create'}
    </button>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="split">
    <form id="environment-form" onsubmit={submit}>
      <section>
        <div class="row">
          <label class="grow">
            <span>Name</span>
            <input bind:value={form.name} placeholder="Abandoned Grove" required />
            {#if editing}
              <small>The slug stays <code>{original.slug}</code> — renaming won't detach it from saved encounters.</small>
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
              {#each ENVIRONMENT_TYPES as type (type)}
                <option value={type}>{type}</option>
              {/each}
            </select>
          </label>
          <label class="narrow">
            <span>Difficulty</span>
            <input bind:value={form.difficulty} placeholder="11" />
          </label>
        </div>

        <label>
          <span>Description</span>
          <textarea rows="2" bind:value={form.description} placeholder="A former druidic grove lying fallow and fully reclaimed by nature."></textarea>
        </label>

        <label>
          <span>Impulses</span>
          <input bind:value={form.impulses} placeholder="Draw in the curious, echo the past" />
        </label>
      </section>

      <section>
        <h3>Potential adversaries</h3>
        {#each form.potentialAdversaries as _, i (i)}
          <div class="repeat">
            <input
              bind:value={form.potentialAdversaries[i]}
              placeholder="Beasts (Bear, Dire Wolf, Glass Snake)"
            />
            <button class="btn danger" type="button" onclick={() => removeAdversaryGroup(i)} aria-label="Remove">×</button>
          </div>
        {/each}
        <button class="btn ghost start" type="button" onclick={addAdversaryGroup}>+ Add group</button>
      </section>

      <section>
        <h3>Features</h3>
        <FeatureEditor features={form.features} withQuestions />
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
          <EnvironmentDetail card={preview} />
        </div>
      {:else}
        <div class="reference">
          <CardBrowser
            compact
            types={ENVIRONMENT_TYPES}
            load={BrowseEnvironments}
            emptyLabel="No environments match these filters."
          >
            {#snippet row(item)}
              <span class="rname">{item.name}</span>
              <span class="rmeta">
                Tier {item.tier} · {item.type}
                {#if item.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
              </span>
            {/snippet}

            {#snippet detail(item)}
              <EnvironmentDetail card={item}>
                {#snippet actions()}
                  <button class="btn" type="button" onclick={() => useAsTemplate(item)}>Use as template</button>
                {/snippet}
              </EnvironmentDetail>
            {/snippet}
          </CardBrowser>
        </div>
      {/if}
    </aside>
  </div>
</div>

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
    align-items: stretch;
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

  .repeat {
    display: flex;
    gap: 0.3rem;
  }

  .repeat input { flex: 1; }

  .start { align-self: flex-start; }

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
