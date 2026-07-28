<script>
  import Modal from '../Modal.svelte'
  import { ShareAdversary, ShareEnvironment, errorMessage } from './api.js'

  let { kind, slug, name } = $props()

  let open = $state(false)
  let code = $state('')
  let error = $state('')
  let copied = $state(false)

  async function show() {
    open = true
    code = ''
    error = ''
    copied = false
    try {
      code = await (kind === 'adversary' ? ShareAdversary(slug) : ShareEnvironment(slug))
    } catch (e) {
      error = errorMessage(e)
    }
  }

  // navigator.clipboard is unavailable on some WebKitGTK builds, so the textarea
  // stays selectable and the button degrades to telling you to copy by hand.
  async function copy() {
    try {
      await navigator.clipboard.writeText(code)
      copied = true
      setTimeout(() => (copied = false), 1800)
    } catch {
      error = 'Copying failed — select the code and copy it by hand.'
    }
  }
</script>

<button class="btn" onclick={show}>Share</button>

<Modal bind:open label="Share code" width="34rem">
  <header>
    <h2>Share “{name}”</h2>
    <button class="btn ghost" onclick={() => (open = false)}>Close</button>
  </header>

  {#if error}<p class="error">{error}</p>{/if}

  <p class="note">
    Anyone running Hilt can paste this back in to get a copy of this card. It carries the
    card only — not your encounters or campaigns.
  </p>

  <textarea readonly rows="5" value={code} onclick={(e) => e.currentTarget.select()}></textarea>

  <div class="row">
    <button class="btn primary" onclick={copy} disabled={!code}>
      {copied ? 'Copied' : 'Copy code'}
    </button>
    <span class="len">{code.length} characters</span>
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

  .row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .len {
    font-size: 0.72rem;
    color: var(--muted);
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
