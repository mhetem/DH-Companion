<script>
  import Modal from '../Modal.svelte'
  import { ImportShareCode, PreviewShareCode, errorMessage } from './api.js'

  let { expect, onimported } = $props()

  let open = $state(false)
  let code = $state('')
  let preview = $state(null)
  let error = $state('')
  let busy = $state(false)

  function show() {
    open = true
    code = ''
    preview = null
    error = ''
  }

  // Previewed before it is written, so a code from a stranger says what it is and
  // what kind it is before anything lands in your homebrew.
  async function look() {
    preview = null
    error = ''
    if (!code.trim()) return
    try {
      const found = await PreviewShareCode(code)
      if (expect && found.kind !== expect) {
        error = `That code holds ${aOrAn(found.kind)}, not ${aOrAn(expect)}. Open the ${found.kind} browser to import it.`
        return
      }
      preview = found
    } catch (e) {
      error = errorMessage(e)
    }
  }

  const aOrAn = (word) => (/^[aeiou]/i.test(word) ? `an ${word}` : `a ${word}`)

  async function accept() {
    busy = true
    error = ''
    try {
      const saved = await ImportShareCode(code)
      open = false
      onimported?.(saved)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }
</script>

<button class="btn" onclick={show}>Import code</button>

<Modal bind:open label="Import a share code" width="34rem">
  <header>
    <h2>Import a share code</h2>
    <button class="btn ghost" onclick={() => (open = false)}>Close</button>
  </header>

  {#if error}<p class="error">{error}</p>{/if}

  <p class="note">
    Paste a code someone shared with you. It is added to your homebrew, and renamed if you
    already have a card by that name — nothing you have is overwritten.
  </p>

  <textarea
    rows="5"
    bind:value={code}
    oninput={() => (preview = null)}
    placeholder="HILT1:…"
  ></textarea>

  {#if preview}
    <div class="preview">
      <span class="pkind">{preview.kind}</span>
      <strong>{preview.name}</strong>
      <span class="pmeta">Tier {preview.tier} · {preview.type}</span>
      {#if preview.description}<p class="pdesc">{preview.description}</p>{/if}
    </div>
  {/if}

  <div class="row">
    {#if preview}
      <button class="btn primary" onclick={accept} disabled={busy}>
        {busy ? 'Importing…' : `Add “${preview.name}”`}
      </button>
      <button class="btn ghost" onclick={() => (preview = null)}>Back</button>
    {:else}
      <button class="btn primary" onclick={look} disabled={!code.trim()}>Check code</button>
    {/if}
  </div>
</Modal>

<style>
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  h2 {
    margin: 0;
    font-size: 1.05rem;
  }

  .note {
    margin: 0;
    font-size: 0.78rem;
    line-height: 1.5;
    color: var(--muted);
  }

  textarea {
    width: 100%;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.72rem;
    word-break: break-all;
  }

  .preview {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding: 0.7rem 0.8rem;
    border: 1px solid var(--gold);
    border-radius: 8px;
    background: var(--bg);
  }

  .pkind {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--gold);
  }

  .pmeta {
    font-size: 0.72rem;
    color: var(--muted);
  }

  .pdesc {
    margin: 0.3rem 0 0;
    font-size: 0.78rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .row {
    display: flex;
    gap: 0.5rem;
  }

  .error {
    margin: 0;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }
</style>
