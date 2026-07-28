<script>
  import { DeleteNote, ListNotes, NOTE_KINDS, SaveNote, noteKindLabel, errorMessage } from './api.js'
  import { renderMarkdown } from './markdown.js'

  // compact is the runner's rail: read-only, so a GM can pull up an NPC mid-scene
  // without leaving the fight.
  let { campaignId, compact = false } = $props()

  let notes = $state([])
  let loading = $state(true)
  let error = $state('')
  let saving = $state(false)
  let kindFilter = $state('')
  let openId = $state(null)
  let editingId = $state(null)
  let form = $state(blank())

  const shown = $derived(kindFilter ? notes.filter((n) => n.kind === kindFilter) : notes)

  function blank() {
    return { kind: 'npc', title: '', body: '' }
  }

  async function refresh() {
    try {
      notes = (await ListNotes(campaignId)) ?? []
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      loading = false
    }
  }

  refresh()

  function edit(note) {
    editingId = note.id
    form = { kind: note.kind, title: note.title, body: note.body }
  }

  function reset() {
    editingId = null
    form = blank()
  }

  async function submit(event) {
    event.preventDefault()
    saving = true
    try {
      await SaveNote({
        id: editingId,
        campaignId,
        kind: form.kind,
        title: form.title,
        body: form.body
      })
      reset()
      await refresh()
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }

  async function remove(note) {
    if (!confirm(`Delete note “${note.title}”?`)) return
    try {
      await DeleteNote(note.id)
      if (editingId === note.id) reset()
      if (openId === note.id) openId = null
      await refresh()
    } catch (e) {
      error = errorMessage(e)
    }
  }

  const toggle = (note) => (openId = openId === note.id ? null : note.id)
</script>

<section class="notes" class:compact>
  <header>
    <h3>Notes</h3>
    <div class="filters">
      <button class="chip" class:on={kindFilter === ''} onclick={() => (kindFilter = '')}>All</button>
      {#each NOTE_KINDS as kind (kind.value)}
        <button class="chip" class:on={kindFilter === kind.value} onclick={() => (kindFilter = kind.value)}>
          {kind.label}
        </button>
      {/each}
    </div>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if !compact}
    <form onsubmit={submit}>
      <div class="row">
        <label class="narrow">
          <span>Kind</span>
          <select bind:value={form.kind}>
            {#each NOTE_KINDS as kind (kind.value)}
              <option value={kind.value}>{kind.label}</option>
            {/each}
          </select>
        </label>
        <label>
          <span>Title</span>
          <input bind:value={form.title} placeholder="Sabine the Grey" required />
        </label>
      </div>
      <label>
        <span>Body — markdown</span>
        <textarea bind:value={form.body} rows="5" placeholder="A wandering merchant who **lurks** near the mire.&#10;&#10;- Owes the party a favour&#10;- Knows the swamp roads"></textarea>
      </label>
      {#if form.body.trim()}
        <div class="preview">{@html renderMarkdown(form.body)}</div>
      {/if}
      <div class="actions">
        <button class="btn primary" type="submit" disabled={saving || !form.title.trim()}>
          {editingId === null ? 'Add note' : 'Save changes'}
        </button>
        {#if editingId !== null}
          <button class="btn ghost" type="button" onclick={reset}>Cancel</button>
        {/if}
      </div>
    </form>
  {/if}

  {#if loading}
    <p class="empty">Loading…</p>
  {:else if !shown.length}
    <p class="empty">
      {#if kindFilter}No {noteKindLabel(kindFilter).toLowerCase()} notes yet.{:else}No notes yet. Keep NPCs, locations, factions, lore and plot threads here.{/if}
    </p>
  {:else}
    <ul class="list">
      {#each shown as note (note.id)}
        <li class:editing={editingId === note.id}>
          <div class="top">
            <button class="disclose" onclick={() => toggle(note)}>
              <span class="caret" class:open={openId === note.id}>›</span>
              <span class="name">{note.title}</span>
            </button>
            <span class="chip">{noteKindLabel(note.kind)}</span>
            {#if !compact}
              <button class="btn ghost" onclick={() => edit(note)}>Edit</button>
              <button class="btn danger" onclick={() => remove(note)} title="Delete">✕</button>
            {/if}
          </div>
          {#if openId === note.id}
            <div class="body">
              {#if note.body.trim()}
                {@html renderMarkdown(note.body)}
              {:else}
                <p class="empty">No body yet.</p>
              {/if}
            </div>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .notes {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  .notes.compact {
    gap: 0.4rem;
    padding: 0.7rem 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  h3 {
    margin: 0;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .filters {
    display: flex;
    gap: 0.25rem;
    flex-wrap: wrap;
  }

  .filters .chip {
    background: transparent;
    font: inherit;
    font-size: 0.7rem;
    cursor: pointer;
  }

  .filters .chip.on {
    border-color: var(--fear);
    color: var(--fear);
  }

  .error {
    margin: 0;
    font-size: 0.75rem;
    color: var(--danger);
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.85rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .row {
    display: flex;
    gap: 0.6rem;
  }

  label {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.25rem;
  }

  label.narrow { flex: 0 0 8rem; }

  label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .actions {
    display: flex;
    gap: 0.5rem;
  }

  .preview {
    padding: 0.5rem 0.7rem;
    border: 1px dashed var(--line);
    border-radius: 6px;
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .list li {
    padding: 0.5rem 0.7rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .compact .list li { background: var(--bg); }

  .list li.editing { border-color: var(--fear); }

  .top {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .disclose {
    display: flex;
    flex: 1;
    align-items: center;
    gap: 0.35rem;
    min-width: 0;
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
    display: inline-block;
    color: var(--muted);
    transition: transform 0.12s ease;
  }

  .caret.open { transform: rotate(90deg); }

  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .body {
    margin-top: 0.4rem;
    padding-top: 0.4rem;
    border-top: 1px solid var(--line);
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .body :global(h3),
  .body :global(h4),
  .body :global(h5) {
    margin: 0.5rem 0 0.25rem;
    font-size: 0.85rem;
    color: var(--text);
  }

  .body :global(p) { margin: 0 0 0.5rem; }
  .body :global(ul),
  .body :global(ol) { margin: 0 0 0.5rem; padding-left: 1.1rem; }
  .body :global(code) {
    padding: 0.05rem 0.25rem;
    border-radius: 4px;
    background: var(--panel-2);
    font-size: 0.8rem;
  }
  .body :global(blockquote) {
    margin: 0 0 0.5rem;
    padding-left: 0.6rem;
    border-left: 2px solid var(--line);
    font-style: italic;
  }

  @media (prefers-reduced-motion: reduce) {
    .caret { transition: none; }
  }
</style>
