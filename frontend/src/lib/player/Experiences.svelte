<script>
  import { DeleteExperience, SaveExperience, errorMessage, signed } from './api.js'

  let { characterId, experiences = [], compact = false, onchange } = $props()

  let error = $state('')
  let saving = $state(false)
  let editingId = $state(null)
  let form = $state({ name: '', modifier: 2 })

  function reset() {
    editingId = null
    form = { name: '', modifier: 2 }
  }

  function edit(experience) {
    editingId = experience.id
    form = { name: experience.name, modifier: experience.modifier }
  }

  async function submit(event) {
    event.preventDefault()
    saving = true
    error = ''
    try {
      const list = await SaveExperience({
        id: editingId,
        characterId,
        name: form.name,
        modifier: Number(form.modifier)
      })
      reset()
      onchange?.(list ?? [])
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }

  async function remove(experience) {
    if (!confirm(`Delete the Experience “${experience.name}”?`)) return
    try {
      const list = await DeleteExperience(experience.id)
      if (editingId === experience.id) reset()
      onchange?.(list ?? [])
    } catch (e) {
      error = errorMessage(e)
    }
  }
</script>

<div class="experiences" class:compact>
  {#if error}
    <p class="error">{error}</p>
  {/if}

  <ul>
    {#each experiences as experience (experience.id)}
      <li>
        <span class="name">{experience.name}</span>
        <span class="mod">{signed(experience.modifier)}</span>
        {#if !compact}
          <button class="btn ghost" onclick={() => edit(experience)}>Edit</button>
          <button class="btn danger" onclick={() => remove(experience)}>✕</button>
        {/if}
      </li>
    {/each}
    {#if !experiences.length}
      <li class="empty">No Experiences yet. You start with two at +2.</li>
    {/if}
  </ul>

  {#if !compact}
    <form onsubmit={submit}>
      <input bind:value={form.name} placeholder="Grew up on the docks" required />
      <input class="narrow" type="number" min="0" max="9" bind:value={form.modifier} />
      <button class="btn primary" type="submit" disabled={saving || !form.name.trim()}>
        {editingId === null ? 'Add' : 'Save'}
      </button>
      {#if editingId !== null}
        <button class="btn ghost" type="button" onclick={reset}>Cancel</button>
      {/if}
    </form>
  {/if}
</div>

<style>
  .experiences {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  ul {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.35rem 0.5rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--panel-2);
  }

  li.empty {
    border: none;
    background: none;
    padding: 0.2rem 0;
  }

  .name {
    flex: 1;
    min-width: 0;
    font-size: 0.85rem;
  }

  .mod {
    font-size: 0.85rem;
    color: var(--hope);
  }

  form {
    display: flex;
    gap: 0.4rem;
  }

  form input { flex: 1; }
  form input.narrow { flex: 0 0 4.5rem; }

  .error {
    margin: 0;
    font-size: 0.78rem;
    color: var(--danger);
  }

  .compact li { background: none; }
</style>
