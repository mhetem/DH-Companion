<script>
  import { GetMasterNote, SaveMasterNote, errorMessage } from './api.js'
  import { renderMarkdown } from './markdown.js'

  // One note per campaign — the timeline you keep open all session. Unlike the
  // typed notes it has no title, no kind and no id: it *is* the campaign's note,
  // so it is addressed by campaign and can never be created or deleted.
  //
  // Both hosts key this component on the campaign id, so campaignId is fixed for
  // the component's life. That is what lets the unmount flush below save against
  // it without re-reading the prop.
  //
  // compact is the runner's rail. Unlike Notes it stays editable there — jotting
  // "they burned the bridge" mid-fight is the whole reason it's on hand.
  let { campaignId, compact = false } = $props()

  const AUTOSAVE_MS = 600

  const PLACEHOLDER = `## Session 4 — the drowned market

- **Day 12** the party burns the bridge at Ashfall
- **Day 14** Sabine turns; the Flood clock starts ticking

### Threads
- Who paid the ferryman?`

  let body = $state('')
  let savedBody = $state('')
  let updatedAt = $state('')
  let loading = $state(true)
  let saving = $state(false)
  let error = $state('')
  let preview = $state(false)

  let timer = null

  const dirty = $derived(!loading && body !== savedBody)

  const status = $derived(
    loading
      ? ''
      : saving
        ? 'Saving…'
        : dirty
          ? 'Unsaved'
          : updatedAt
            ? savedAt(updatedAt)
            : 'Not saved yet'
  )

  GetMasterNote(campaignId)
    .then((note) => {
      body = note.body
      savedBody = note.body
      updatedAt = note.updatedAt
    })
    .catch((e) => (error = errorMessage(e)))
    .finally(() => (loading = false))

  // One write in flight at a time. Anything typed while it was away is caught by
  // the trailing call rather than racing a second write against the same row.
  // A failure stops the chain instead of retrying itself into a loop — the next
  // keystroke reschedules, which is also how you recover from it.
  async function persist() {
    if (saving || loading) return
    const pending = body
    if (pending === savedBody) return
    saving = true
    try {
      const note = await SaveMasterNote(campaignId, pending)
      savedBody = pending
      updatedAt = note.updatedAt
      error = ''
      saving = false
      if (body !== savedBody) persist()
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

  // Switching campaigns or leaving the section unmounts this, and the pending
  // debounce would go with it and take the last few keystrokes along. There is no
  // UI left to report a failure to, so this one is deliberately silent.
  $effect(() => () => {
    clearTimeout(timer)
    if (body !== savedBody) SaveMasterNote(campaignId, body).catch(() => {})
  })

  // keys.js guards on isTyping, so a textarea never reaches the global map — the
  // save chord has to be bound here.
  function keydown(event) {
    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') {
      event.preventDefault()
      flush()
    }
  }

  function savedAt(iso) {
    const at = new Date(iso)
    if (Number.isNaN(at.getTime())) return 'Saved'
    const today = at.toDateString() === new Date().toDateString()
    return today
      ? `Saved ${at.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
      : `Saved ${at.toLocaleDateString()}`
  }
</script>

<section class="master" class:compact>
  <header>
    <h3>Master note</h3>
    <div class="tools">
      <span class="status" class:dirty class:saving>{status}</span>
      <button class="chip" class:on={preview} onclick={() => (preview = !preview)}>
        {preview ? 'Edit' : 'Preview'}
      </button>
    </div>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if loading}
    <p class="empty">Loading…</p>
  {:else if preview}
    <div class="body">
      {#if body.trim()}
        {@html renderMarkdown(body)}
      {:else}
        <p class="empty">Nothing written yet.</p>
      {/if}
    </div>
  {:else}
    <textarea
      bind:value={body}
      oninput={schedule}
      onblur={flush}
      onkeydown={keydown}
      rows={compact ? 8 : 22}
      placeholder={PLACEHOLDER}
      aria-label="Master note"
    ></textarea>
  {/if}

  {#if !compact}
    <p class="hint">Markdown, saved as you type. One per campaign — the running timeline, not a filed note.</p>
  {/if}
</section>

<style>
  .master {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .master.compact {
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

  .tools {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .status {
    font-size: 0.7rem;
    color: var(--muted);
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

  .compact textarea {
    font-size: 0.8rem;
    line-height: 1.45;
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

  .compact .body {
    padding: 0.4rem 0.5rem;
    background: var(--bg);
    font-size: 0.8rem;
    color: var(--muted);
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

  .hint {
    margin: 0;
    font-size: 0.75rem;
    color: var(--muted);
  }
</style>
