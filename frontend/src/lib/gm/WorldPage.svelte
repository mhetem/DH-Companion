<script>
  import { SaveWorldNote, WORLD_DATED_KINDS, WORLD_KINDS, errorMessage } from './api.js'
  import { renderMarkdown } from './markdown.js'

  // One world page's own content — title, kind, date and body — autosaved the way
  // MasterNote is. Where the page sits in the tree is World's business, not this
  // component's, so a move can never race an autosave over the same fields.
  //
  // World keys this on the page id, so note.id is fixed for the component's life.
  // That is what lets the unmount flush save against the page that was open rather
  // than the one being switched to. The note prop is only read once: the form is
  // the source of truth while it's mounted, so a list refresh can't clobber edits
  // that haven't gone out yet.
  let { note, onsaved } = $props()

  const AUTOSAVE_MS = 600

  const PLACEHOLDERS = {
    region: 'Climate, borders, who holds it and who wants it.\n\n## Settlements\nAdd cities inside this page — they nest under it.',
    city: '**Population** about 12,000 · **Ruled by** the Tide Council\n\n## Districts\n- The drowned market\n- Lantern Row',
    place: 'What it looks like, who is usually here, and what the party can find.',
    timeline: 'The shape of this age in a line or two.\n\nAdd events inside this page and they are laid out in order below.',
    event: 'What happened, who was there, and what it changed.',
    culture: 'Customs, language, what they eat, what they fear.',
    religion: 'The gods, the clergy, the heresies.',
    page: 'Markdown — headings, **bold**, lists, > quotes.'
  }

  let form = $state(fields(note))
  let saved = $state(fields(note))
  let updatedAt = $state(note.updatedAt)
  let saving = $state(false)
  let error = $state('')
  let preview = $state(false)

  let timer = null

  const dirty = $derived(!same(form, saved))
  const dated = $derived(WORLD_DATED_KINDS.includes(form.kind))

  const status = $derived(saving ? 'Saving…' : dirty ? 'Unsaved' : savedAt(updatedAt))

  function fields(n) {
    return { kind: n.kind, title: n.title, worldDate: n.worldDate, body: n.body }
  }

  function same(a, b) {
    return a.kind === b.kind && a.title === b.title && a.worldDate === b.worldDate && a.body === b.body
  }

  const input = (pending) => ({ id: note.id, campaignId: note.campaignId, parentId: note.parentId, ...pending })

  // One write in flight at a time, with a trailing call for whatever was typed
  // while it was away — the same chain as MasterNote, for the same reasons.
  async function persist() {
    if (saving) return
    const pending = $state.snapshot(form)
    if (same(pending, saved)) return
    saving = true
    try {
      const next = await SaveWorldNote(input(pending))
      saved = pending
      updatedAt = next.updatedAt
      error = ''
      onsaved?.(next)
      saving = false
      if (!same(form, saved)) persist()
    } catch (e) {
      error = errorMessage(e)
      saving = false
    }
  }

  function schedule() {
    clearTimeout(timer)
    timer = setTimeout(persist, AUTOSAVE_MS)
  }

  function flush() {
    clearTimeout(timer)
    persist()
  }

  // Picking another page, leaving the tab, or deleting this page unmounts it. The
  // last keystrokes still go out; a failure has nowhere to be shown, and a page that
  // was just deleted fails harmlessly — the update matches no row.
  $effect(() => () => {
    clearTimeout(timer)
    const pending = $state.snapshot(form)
    if (!same(pending, saved)) {
      SaveWorldNote(input(pending))
        .then((next) => onsaved?.(next))
        .catch(() => {})
    }
  })

  function keydown(event) {
    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') {
      event.preventDefault()
      flush()
    }
  }

  function savedAt(iso) {
    const at = new Date(iso)
    if (!iso || Number.isNaN(at.getTime())) return 'Saved'
    const today = at.toDateString() === new Date().toDateString()
    return today
      ? `Saved ${at.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
      : `Saved ${at.toLocaleDateString()}`
  }
</script>

<div class="editor">
  <div class="head">
    <input
      class="title"
      bind:value={form.title}
      oninput={schedule}
      onblur={flush}
      onkeydown={keydown}
      placeholder="Untitled"
      aria-label="Title"
    />
    <div class="tools">
      <span class="status" class:dirty class:saving>{status}</span>
      <button class="chip" class:on={preview} onclick={() => (preview = !preview)}>
        {preview ? 'Edit' : 'Preview'}
      </button>
    </div>
  </div>

  <div class="meta">
    <label class="kind">
      <span>Kind</span>
      <select bind:value={form.kind} onchange={flush}>
        {#each WORLD_KINDS as kind (kind.value)}
          <option value={kind.value}>{kind.label}</option>
        {/each}
      </select>
    </label>
    {#if dated}
      <label class="when">
        <span>When</span>
        <input
          bind:value={form.worldDate}
          oninput={schedule}
          onblur={flush}
          onkeydown={keydown}
          placeholder={form.kind === 'timeline' ? 'Year 0 – 812 of the Second Age' : '312 of the Second Age'}
        />
      </label>
    {/if}
  </div>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if preview}
    <div class="body">
      {#if form.body.trim()}
        {@html renderMarkdown(form.body)}
      {:else}
        <p class="empty">Nothing written yet.</p>
      {/if}
    </div>
  {:else}
    <textarea
      bind:value={form.body}
      oninput={schedule}
      onblur={flush}
      onkeydown={keydown}
      rows="14"
      placeholder={PLACEHOLDERS[form.kind] ?? PLACEHOLDERS.page}
      aria-label="Page body"
    ></textarea>
  {/if}
</div>

<style>
  .editor {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .head {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  .title {
    flex: 1;
    min-width: 0;
    padding: 0.35rem 0.55rem;
    font-size: 1.05rem;
    font-weight: 600;
  }

  .tools {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .status {
    font-size: 0.7rem;
    color: var(--muted);
    white-space: nowrap;
  }

  .status.dirty,
  .status.saving { color: var(--gold); }

  .chip {
    background: transparent;
    font: inherit;
    font-size: 0.7rem;
    cursor: pointer;
  }

  .chip.on {
    border-color: var(--fear);
    color: var(--fear);
  }

  .meta {
    display: flex;
    gap: 0.6rem;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  label.kind { flex: 0 0 9rem; }
  label.when { flex: 1; }

  label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .error {
    margin: 0;
    font-size: 0.75rem;
    color: var(--danger);
  }

  textarea {
    width: 100%;
    resize: vertical;
    font: inherit;
    font-size: 0.9rem;
    line-height: 1.55;
  }

  .body {
    padding: 0.6rem 0.8rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
    font-size: 0.9rem;
    line-height: 1.55;
    color: var(--text);
  }

  .body :global(h3),
  .body :global(h4),
  .body :global(h5) {
    margin: 0.7rem 0 0.3rem;
    font-size: 0.9rem;
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
</style>
