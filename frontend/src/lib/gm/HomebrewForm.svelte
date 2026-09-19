<script>
  // Shared chrome for the equipment and loot homebrew forms: the header, the error
  // line, and the form/preview split. The adversary and environment forms predate
  // this and carry their own copy — they also have a reference browser and a
  // template picker this doesn't, so they were left alone rather than bent to fit.
  //
  // `fields` is the form body, `preview` the live card beside it.
  let { title, saving = false, canSave = true, error = '', formId, onsubmit, oncancel, fields, preview } =
    $props()
</script>

<div class="form-pane">
  <header>
    <h2>{title}</h2>
    <button class="btn ghost" type="button" onclick={oncancel}>Cancel</button>
    <button class="btn primary" type="submit" form={formId} disabled={saving || !canSave}>
      {saving ? 'Saving…' : 'Save'}
    </button>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="split">
    <form id={formId} {onsubmit}>
      {@render fields()}
    </form>
    <aside>
      <span class="label">Preview</span>
      <div class="preview">{@render preview()}</div>
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

  aside {
    display: flex;
    flex-direction: column;
    width: 23rem;
    flex-shrink: 0;
    min-height: 0;
    padding: 1rem;
    overflow: hidden;
  }

  .label {
    flex-shrink: 0;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
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

  /* The field styles live here so each form body is markup only. :global is how a
     parent's styles reach the snippet it passes in, which Svelte scopes to the
     component that declared it. */
  form :global(section) {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    align-items: stretch;
  }

  form :global(h3) {
    margin: 0;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--muted);
  }

  form :global(.row) {
    display: flex;
    gap: 0.5rem;
  }

  form :global(label) {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.25rem;
    min-width: 0;
  }

  form :global(label.grow) { flex: 2; }
  form :global(label.narrow) { flex: 0 0 7rem; }

  form :global(label span) {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  form :global(small) {
    font-size: 0.7rem;
    color: var(--muted);
    opacity: 0.8;
  }

  form :global(code) {
    padding: 0 0.2rem;
    border-radius: 3px;
    background: var(--panel-2);
  }
</style>
