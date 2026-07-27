<script>
  import { CreateParty, DeleteParty, ListParties, TIERS, UpdateParty, errorMessage } from './api.js'

  let parties = $state([])
  let loading = $state(true)
  let error = $state('')
  let saving = $state(false)

  // editingId is null when the form is creating a new party.
  let editingId = $state(null)
  let form = $state(blank())

  function blank() {
    return { name: '', size: 4, tier: '1' }
  }

  async function refresh() {
    try {
      parties = (await ListParties()) ?? []
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      loading = false
    }
  }

  refresh()

  function edit(party) {
    editingId = party.id
    form = { name: party.name, size: party.size, tier: party.tier }
  }

  function reset() {
    editingId = null
    form = blank()
  }

  async function submit(event) {
    event.preventDefault()
    saving = true
    try {
      const input = { name: form.name, size: Number(form.size), tier: form.tier }
      if (editingId === null) {
        await CreateParty(input)
      } else {
        await UpdateParty(editingId, input)
      }
      reset()
      await refresh()
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }

  async function remove(party) {
    // Encounters keep their party_id; a deleted party just drops the budget.
    if (!confirm(`Delete party “${party.name}”?`)) return
    try {
      await DeleteParty(party.id)
      if (editingId === party.id) reset()
      await refresh()
    } catch (e) {
      error = errorMessage(e)
    }
  }
</script>

<div class="parties">
  <header>
    <h2>Parties</h2>
    <p class="blurb">A party's size and tier set the battle-point budget for every encounter built against it.</p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <form onsubmit={submit}>
    <label>
      <span>Name</span>
      <input bind:value={form.name} placeholder="The Sablewood Six" required />
    </label>
    <label class="narrow">
      <span>Size</span>
      <input type="number" min="1" max="12" bind:value={form.size} required />
    </label>
    <label class="narrow">
      <span>Tier</span>
      <select bind:value={form.tier}>
        {#each TIERS as tier (tier)}
          <option value={tier}>Tier {tier}</option>
        {/each}
      </select>
    </label>
    <button class="btn primary" type="submit" disabled={saving || !form.name.trim()}>
      {editingId === null ? 'Add party' : 'Save changes'}
    </button>
    {#if editingId !== null}
      <button class="btn ghost" type="button" onclick={reset}>Cancel</button>
    {/if}
  </form>

  <ul class="list">
    {#if loading}
      <li class="empty">Loading…</li>
    {:else if !parties.length}
      <li class="empty">No parties yet. Add one above to unlock the budget meter.</li>
    {:else}
      {#each parties as party (party.id)}
        <li class:editing={editingId === party.id}>
          <div class="who">
            <span class="name">{party.name}</span>
            <span class="meta">{party.size} {party.size === 1 ? 'player' : 'players'} · Tier {party.tier}</span>
          </div>
          <span class="budget">{3 * party.size + 2} pts</span>
          <button class="btn ghost" onclick={() => edit(party)}>Edit</button>
          <button class="btn danger" onclick={() => remove(party)}>Delete</button>
        </li>
      {/each}
    {/if}
  </ul>
</div>

<style>
  .parties {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  header { margin-bottom: 1rem; }

  h2 {
    margin: 0;
    font-size: 1.25rem;
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

  form {
    display: flex;
    align-items: flex-end;
    gap: 0.6rem;
    padding: 0.85rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  label {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.25rem;
  }

  label.narrow { flex: 0 0 6.5rem; }

  label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .list {
    margin: 1rem 0 0;
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

  .list li.editing { border-color: var(--fear); }

  .who {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }

  .name { font-size: 0.9rem; }

  .meta {
    font-size: 0.75rem;
    color: var(--muted);
  }

  .budget {
    font-size: 0.75rem;
    color: var(--muted);
  }

  .empty {
    display: block;
    padding: 1rem 0;
    font-size: 0.85rem;
    color: var(--muted);
    border: none;
    background: none;
  }
</style>
