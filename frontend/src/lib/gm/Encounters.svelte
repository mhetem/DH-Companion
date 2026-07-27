<script>
  import EncounterBuilder from './EncounterBuilder.svelte'
  import { ListEncounters, ListParties, errorMessage } from './api.js'

  // The builder is wide enough on its own that the list and the builder swap
  // places rather than sitting side by side next to the shell's nav.
  let editing = $state(null) // null = list view, otherwise an id or 'new'
  let encounters = $state([])
  let parties = $state([])
  let loading = $state(true)
  let error = $state('')

  async function refresh() {
    try {
      const [e, p] = await Promise.all([ListEncounters(), ListParties()])
      encounters = e ?? []
      parties = p ?? []
      error = ''
    } catch (err) {
      error = errorMessage(err)
    } finally {
      loading = false
    }
  }

  refresh()

  function partyName(id) {
    return parties.find((p) => p.id === id)?.name ?? 'Unknown party'
  }

  function onsaved() {
    refresh()
  }

  function ondeleted() {
    editing = null
    refresh()
  }

  function back() {
    editing = null
    refresh()
  }
</script>

{#if editing !== null}
  <div class="edit">
    <div class="toolbar">
      <button class="btn ghost" onclick={back}>← All encounters</button>
    </div>
    {#key editing}
      <EncounterBuilder
        encounterId={editing === 'new' ? null : editing}
        {parties}
        {onsaved}
        {ondeleted}
      />
    {/key}
  </div>
{:else}
  <div class="list">
    <header>
      <div>
        <h2>Encounters</h2>
        <p class="blurb">Build a roster against a party, attach an environment, and watch the budget.</p>
      </div>
      <button class="btn primary" onclick={() => (editing = 'new')}>New encounter</button>
    </header>

    {#if error}
      <p class="error">{error}</p>
    {/if}

    {#if loading}
      <p class="empty">Loading…</p>
    {:else if !encounters.length}
      <p class="empty">No encounters yet. Start one with <em>New encounter</em>.</p>
    {:else}
      <ul>
        {#each encounters as encounter (encounter.id)}
          <li>
            <button onclick={() => (editing = encounter.id)}>
              <span class="name">{encounter.name}</span>
              <span class="meta">
                {encounter.totalCount}
                {encounter.totalCount === 1 ? 'adversary' : 'adversaries'}
                {#if encounter.partyId != null}· {partyName(encounter.partyId)}{/if}
                {#if encounter.environmentSlug}· {encounter.environmentSlug}{/if}
              </span>
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}

<style>
  .edit {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  .toolbar {
    padding: 0.5rem 1rem 0;
  }

  .list {
    flex: 1;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  header {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  header div { flex: 1; }

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

  ul {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li { margin-bottom: 0.4rem; }

  li button {
    display: block;
    width: 100%;
    padding: 0.6rem 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  li button:hover { border-color: var(--muted); }

  .name {
    display: block;
    font-size: 0.9rem;
  }

  .meta {
    font-size: 0.75rem;
    color: var(--muted);
  }

  .empty {
    font-size: 0.85rem;
    color: var(--muted);
  }
</style>
