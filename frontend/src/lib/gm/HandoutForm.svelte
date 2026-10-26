<script>
  import { untrack } from 'svelte'
  import { ChooseHandoutMedia, HANDOUT_KINDS, errorMessage } from './api.js'
  import { renderMarkdown } from './markdown.js'

  // Adds or edits one map, scene or text. A chosen file stays a path until Save —
  // that's when it's copied into the media folder — so backing out of the form
  // never leaves a stray copy behind. The host keys this on what's being edited,
  // so the props are only read once.
  let { handout = null, kind = 'scene', onsave, oncancel } = $props()

  const initial = untrack(() => handout)

  let form = $state({
    kind: initial?.kind ?? untrack(() => kind),
    title: initial?.title ?? '',
    body: initial?.body ?? ''
  })
  let source = $state('')
  let dropped = $state(false)
  let saving = $state(false)
  let error = $state('')

  const text = $derived(form.kind === 'text')
  const hint = $derived(HANDOUT_KINDS.find((k) => k.value === form.kind)?.hint ?? '')
  const fileName = $derived(source ? source.split(/[\\/]/).pop() : dropped ? '' : (initial?.mediaName ?? ''))
  const ready = $derived(text ? !!(form.body.trim() || fileName) : !!fileName)

  async function choose() {
    try {
      const path = await ChooseHandoutMedia()
      if (path) {
        source = path
        dropped = false
      }
      error = ''
    } catch (e) {
      error = errorMessage(e)
    }
  }

  function dropBackdrop() {
    source = ''
    dropped = true
  }

  async function submit(event) {
    event.preventDefault()
    saving = true
    try {
      await onsave({
        id: initial?.id ?? null,
        kind: form.kind,
        title: form.title,
        body: form.body,
        source,
        clearMedia: dropped && !source
      })
    } catch (e) {
      error = errorMessage(e)
      saving = false
    }
  }
</script>

<form onsubmit={submit}>
  <div class="kinds">
    {#each HANDOUT_KINDS as k (k.value)}
      <button type="button" class="tabbtn" class:on={form.kind === k.value} onclick={() => (form.kind = k.value)}>
        {k.label}
      </button>
    {/each}
    <span class="hint">{hint}</span>
  </div>

  <label>
    <span>Title — only you see it</span>
    <input bind:value={form.title} placeholder={text ? 'A letter from Sabine' : 'Left blank, it takes the file name'} />
  </label>

  <div class="file">
    <span class="label">{text ? 'Backdrop — optional' : 'Image or video'}</span>
    <div class="filerow">
      <span class="name" class:none={!fileName}>{fileName || 'No file chosen'}</span>
      <button type="button" class="btn" onclick={choose} disabled={saving}>{fileName ? 'Replace…' : 'Choose…'}</button>
      {#if text && fileName}
        <button type="button" class="btn ghost" onclick={dropBackdrop} disabled={saving}>Remove</button>
      {/if}
    </div>
    <p class="hint">
      PNG, JPEG, WebP or GIF images, MP4 or WebM video. Saving copies it into Hilt's media folder, so
      the original can move or go without breaking anything here.
    </p>
  </div>

  <label>
    <span>{text ? 'Text — shown on the table, markdown' : 'Notes — only you see these'}</span>
    <textarea
      bind:value={form.body}
      rows={text ? 6 : 3}
      placeholder={text
        ? '# The Drowned Market\n\nThe tide is out, and the stalls stand in a foot of black water.'
        : 'The secret door is on the north wall.'}
    ></textarea>
  </label>
  {#if text && form.body.trim()}
    <div class="preview">{@html renderMarkdown(form.body)}</div>
  {/if}

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="actions">
    <button class="btn primary" type="submit" disabled={saving || !ready}>
      {#if saving}{source ? 'Copying…' : 'Saving…'}{:else}{initial ? 'Save changes' : 'Add'}{/if}
    </button>
    <button class="btn ghost" type="button" onclick={oncancel} disabled={saving}>Cancel</button>
  </div>
</form>

<style>
  form {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding: 0.85rem;
    border: 1px solid var(--fear);
    border-radius: 8px;
    background: var(--panel);
  }

  .kinds {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    flex-wrap: wrap;
  }

  .kinds .hint { margin-left: 0.4rem; }

  .tabbtn {
    padding: 0.3rem 0.7rem;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: var(--panel);
    color: var(--muted);
    font: inherit;
    font-size: 0.8rem;
    cursor: pointer;
  }

  .tabbtn:hover { color: var(--text); }

  .tabbtn.on {
    border-color: var(--fear);
    color: var(--fear);
  }

  label,
  .file {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  label span,
  .label {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .filerow {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    font-size: 0.85rem;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .name.none { color: var(--muted); }

  .hint {
    margin: 0;
    font-size: 0.75rem;
    color: var(--muted);
  }

  .preview {
    padding: 0.5rem 0.7rem;
    border: 1px dashed var(--line);
    border-radius: 6px;
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .error {
    margin: 0;
    font-size: 0.75rem;
    color: var(--danger);
  }

  .actions {
    display: flex;
    gap: 0.5rem;
  }
</style>
